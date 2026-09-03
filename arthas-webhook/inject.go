// MsrePilot：Arthas 注入 admission webhook（独立服务，集群内部署）
//
// 职责（纯数据面，无 DB/无 K8s 客户端）：
//
//	apiserver 创建 Pod → MutatingWebhookConfiguration（Service 型 clientConfig）→
//	本服务校验 label → 返回 JSON Patch（initContainer 拷贝 agent + 主容器 javaagent）
//
// 配置全部来自环境变量（ConfigMap envSource 注入，治理页保存时由平台同步并滚动重启）。
//
// 与 backend3 的关系：
//   - 管理面（MWC 创建/同步、ConfigMap 同步、治理页）在 backend3/service/k8s/arthas_webhook.go
//   - 本文件是注入核心的独立副本，修改注入逻辑时必须同步两份
//     （另一份：backend3/service/k8s/arthas_webhook.go BuildInjectionPatch）
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ========== 常量（与 backend3/service/k8s/arthas_webhook.go 保持一致） ==========

const (
	// webhookHandlerPath admission 回调路径
	webhookHandlerPath = "/webhook/msre-pilot"
	// injectLabel 触发注入的 Pod label
	injectLabel = "oneops-arthas-injection"
	// injectDisabledAnnotation 显式排除注入的 annotation
	injectDisabledAnnotation = "oneops-arthas.injection/disabled"
	// injectedMarker 已注入标记（幂等）
	injectedMarker = "oneops-arthas.io/injected"

	injectVolumeName = "oneops-arthas"
	injectAgentMount = "/oneops/arthas"
	injectInitName   = "oneops-arthas-agent"

	defaultInitImage = "registry.cn-hangzhou.aliyuncs.com/oneops/arthas-agent:3.7.2"
)

// ========== 注入判定与 Patch 构建 ==========

// jsonPatchOp RFC6902 JSON Patch 操作
type jsonPatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// shouldInject 判断 Pod 是否应注入：label 触发 + annotation 排除 + 幂等
func shouldInject(pod *corev1.Pod) bool {
	if pod.Labels[injectLabel] != "enabled" {
		return false
	}
	if pod.Annotations[injectDisabledAnnotation] == "true" {
		return false
	}
	if pod.Annotations[injectedMarker] == "true" {
		return false
	}
	return true
}

// buildInjectionPatch 生成注入 JSON Patch
// tunnelWS 形如 ws://arthas-tunnel.oneops:7777/ws；agentId 由 initContainer 经
// 主容器 JAVA_TOOL_OPTIONS 的 -Darthas.agentId=$(POD_NAME)-$(POD_NAMESPACE) 注入
func buildInjectionPatch(pod *corev1.Pod, tunnelWS, initImage string) ([]jsonPatchOp, error) {
	if tunnelWS == "" {
		return nil, fmt.Errorf("tunnel-server WS 地址未配置（ARTHAS_TUNNEL_WS）")
	}
	patches := []jsonPatchOp{}

	// 1) 标记已注入（幂等 + 可观测）
	if pod.Annotations == nil {
		patches = append(patches, jsonPatchOp{Op: "add", Path: "/metadata/annotations", Value: map[string]string{}})
	}
	patches = append(patches, jsonPatchOp{
		Op: "add", Path: "/metadata/annotations/" + escapeJSONPointer(injectedMarker), Value: "true",
	})

	// 2) 共享卷（emptyDir）
	hasVolume := false
	for _, v := range pod.Spec.Volumes {
		if v.Name == injectVolumeName {
			hasVolume = true
			break
		}
	}
	if !hasVolume {
		vol := corev1.Volume{
			Name: injectVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: resource.NewScaledQuantity(128, resource.Mega)},
			},
		}
		if pod.Spec.Volumes == nil {
			patches = append(patches, jsonPatchOp{Op: "add", Path: "/spec/volumes", Value: []corev1.Volume{vol}})
		} else {
			patches = append(patches, jsonPatchOp{Op: "add", Path: "/spec/volumes/-", Value: vol})
		}
	}

	// 3) initContainer：拷贝全量 arthas 文件到共享卷
	//
	// 设计依据：Arthas 4.x 的 AgentBootstrap.args 协议是 <arthas-core.jar>;<Configure 串>，
	// tunnel 配置走 arthas.properties 文件（镜像内预置），agentId/tunnelServer 通过
	// 主容器 JAVA_TOOL_OPTIONS 注入 -D 系统属性覆盖（优先级 System Properties > 文件）。
	initCtr := corev1.Container{
		Name:            injectInitName,
		Image:           initImage,
		ImagePullPolicy: corev1.PullAlways,
		Command: []string{"sh", "-c",
			"set -e; cp -rf /opt/arthas/. /oneops-agent/ ; " +
				"test -f /oneops-agent/arthas.properties || { echo 'arthas.properties not found in init image' >&2; exit 1; }"},
		VolumeMounts: []corev1.VolumeMount{{Name: injectVolumeName, MountPath: "/oneops-agent"}},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},
	}
	if pod.Spec.InitContainers == nil {
		patches = append(patches, jsonPatchOp{Op: "add", Path: "/spec/initContainers", Value: []corev1.Container{initCtr}})
	} else {
		patches = append(patches, jsonPatchOp{Op: "add", Path: "/spec/initContainers/-", Value: initCtr})
	}

	// 4) 主容器：挂载 agent + 注入 env（ARTHAS_TUNNEL_SERVER + downward API + JAVA_TOOL_OPTIONS）
	// K8s 按 env 出现顺序展开 $(VAR)，故 ARTHAS_TUNNEL_SERVER/POD_NAME/POD_NAMESPACE 必须在 JAVA_TOOL_OPTIONS 之前。
	javaagentArg := fmt.Sprintf(
		"-javaagent:%s/arthas-agent.jar -Darthas.tunnelServer=$(ARTHAS_TUNNEL_SERVER) -Darthas.agentId=$(POD_NAME)-$(POD_NAMESPACE)",
		injectAgentMount,
	)
	for i, ctr := range pod.Spec.Containers {
		base := fmt.Sprintf("/spec/containers/%d", i)

		// volumeMounts
		hasMount := false
		for _, m := range ctr.VolumeMounts {
			if m.Name == injectVolumeName {
				hasMount = true
				break
			}
		}
		if !hasMount {
			mount := corev1.VolumeMount{Name: injectVolumeName, MountPath: injectAgentMount, ReadOnly: true}
			if ctr.VolumeMounts == nil {
				patches = append(patches, jsonPatchOp{Op: "add", Path: base + "/volumeMounts", Value: []corev1.VolumeMount{mount}})
			} else {
				patches = append(patches, jsonPatchOp{Op: "add", Path: base + "/volumeMounts/-", Value: mount})
			}
		}

		// env：先确保 downward API env 存在（缺失时 $(POD_NAME) 会以字面量进入 agentId）
		envNil := ctr.Env == nil
		hasTunnelSrv, hasPodName, hasPodNS := false, false, false
		existingOpts, optsIdx := "", -1
		for idx, e := range ctr.Env {
			switch e.Name {
			case "ARTHAS_TUNNEL_SERVER":
				hasTunnelSrv = true
			case "POD_NAME":
				hasPodName = true
			case "POD_NAMESPACE":
				hasPodNS = true
			case "JAVA_TOOL_OPTIONS":
				existingOpts, optsIdx = e.Value, idx
			}
		}
		addEnv := func(e corev1.EnvVar) {
			if envNil {
				patches = append(patches, jsonPatchOp{Op: "add", Path: base + "/env", Value: []corev1.EnvVar{e}})
				envNil = false
			} else {
				patches = append(patches, jsonPatchOp{Op: "add", Path: base + "/env/-", Value: e})
			}
		}
		// ARTHAS_TUNNEL_SERVER 必须先于 JAVA_TOOL_OPTIONS 注入（K8s 按顺序展开 $(VAR)）
		if !hasTunnelSrv {
			addEnv(corev1.EnvVar{Name: "ARTHAS_TUNNEL_SERVER", Value: tunnelWS})
		}
		if !hasPodName {
			addEnv(corev1.EnvVar{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}}})
		}
		if !hasPodNS {
			addEnv(corev1.EnvVar{Name: "POD_NAMESPACE", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}})
		}

		// JAVA_TOOL_OPTIONS：已有则合并（保留应用既有参数），无则新增
		if optsIdx >= 0 {
			if !strings.Contains(existingOpts, "arthas-agent.jar") {
				merged := corev1.EnvVar{Name: "JAVA_TOOL_OPTIONS", Value: strings.TrimSpace(existingOpts + " " + javaagentArg)}
				patches = append(patches, jsonPatchOp{
					Op: "replace", Path: fmt.Sprintf("%s/env/%d", base, optsIdx), Value: merged,
				})
			}
		} else {
			addEnv(corev1.EnvVar{Name: "JAVA_TOOL_OPTIONS", Value: javaagentArg})
		}
	}

	return patches, nil
}

// escapeJSONPointer JSON Pointer 转义（~ → ~0，/ → ~1）
func escapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	return strings.ReplaceAll(s, "/", "~1")
}

// admissionPodName admission 阶段 Deployment 创建的 Pod 尚未分配随机后缀名
// （apiserver 在全部 admission 通过后才由 generateName 生成 name），
// 记 generateName 前缀（含所属 workload 线索，如 test-java2-5f598c85d4-*）
func admissionPodName(pod *corev1.Pod) string {
	if pod.GenerateName != "" {
		return pod.GenerateName + "*"
	}
	return pod.Name
}

// ========== AdmissionReview 处理 ==========

// handleAdmission 处理 AdmissionReview：label 命中则返回注入 patch
func handleAdmission(c *gin.Context) {
	var review admissionv1.AdmissionReview
	if err := c.ShouldBindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid admission review: %v", err)})
		return
	}
	resp := &admissionv1.AdmissionResponse{Allowed: true}
	if review.Request != nil {
		resp.UID = review.Request.UID
		pod := &corev1.Pod{}
		if err := json.Unmarshal(review.Request.Object.Raw, pod); err != nil {
			resp.Allowed = false
			resp.Result = &metav1.Status{Message: fmt.Sprintf("decode pod failed: %v", err)}
		} else {
			logger.Info("webhook 收到 admission 请求",
				zap.String("namespace", pod.Namespace), zap.String("pod", admissionPodName(pod)),
				zap.String("injectLabel", pod.Labels[injectLabel]),
				zap.Bool("shouldInject", shouldInject(pod)))
			if shouldInject(pod) {
				if patches, err := buildInjectionPatch(pod, tunnelWSConfig(), initImageConfig()); err != nil {
					resp.Allowed = false
					resp.Result = &metav1.Status{Message: err.Error()}
				} else if len(patches) > 0 {
					patchBytes, err := json.Marshal(patches)
					if err != nil {
						resp.Allowed = false
						resp.Result = &metav1.Status{Message: fmt.Sprintf("marshal patch failed: %v", err)}
					} else {
						logger.Info("admission 注入 patch 已返回",
							zap.String("namespace", pod.Namespace), zap.String("pod", admissionPodName(pod)),
							zap.Int("patches", len(patches)))
						patchType := admissionv1.PatchTypeJSONPatch
						resp.PatchType = &patchType
						// Patch 为 []byte：JSON 序列化时自动 base64（wire 协议自带）。
						// 此处必须放原始 JSON Patch 字节，手动再编一次会双重编码，
						// apiserver 解码后拿到 base64 字符串而非 patch，注入被静默丢弃
						resp.Patch = patchBytes
					}
				}
			}
		}
	}
	review.Response = resp
	review.Request = nil
	c.JSON(http.StatusOK, review)
}
