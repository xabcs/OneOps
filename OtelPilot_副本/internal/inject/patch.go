// Package inject 实现基于自研标签判定的 OpenTelemetry-JavaAgent 无侵入注入
// patch 构建（纯函数，便于单测）。
package inject

import (
	"fmt"
	"sort"
	"strings"

	jsonpatch "gomodules.xyz/jsonpatch/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"otelpilot/internal/config"
)

// 自研标签体系。
const (
	// LabelEnable 注入开关，仅识别字符串 on / off，其余值（含布尔、数字）一律视为关闭。
	LabelEnable = "otel.pilot.auto.enable"
	// LabelAppName 应用名称，映射为 OTEL_SERVICE_NAME。
	LabelAppName = "otel.pilot.app.name"
	// LabelAgentVersion 探针版本覆盖。
	LabelAgentVersion = "otel.pilot.agent.version"
	// LabelJDKVersion JDK 版本（记录到 Annotation，供观测与排障）。
	LabelJDKVersion = "otel.pilot.jdk.version"

	// EnableOn 开启注入的唯一合法取值。
	EnableOn = "on"
)

// 防重注入 Annotation。
const (
	AnnotationStatus       = "otel-pilot.io/status"
	AnnotationAgentVersion = "otel-pilot.io/agent-version"
	AnnotationJDKVersion   = "otel-pilot.io/jdk-version"
	StatusInjected         = "injected"
)

// 注入资源命名。
const (
	VolumeName        = "otel-pilot-agent"
	AgentMountPath    = "/otel-pilot"
	AgentJarName      = "opentelemetry-javaagent.jar"
	InitContainerName = "otel-pilot-init"
	JavaToolOptions   = "JAVA_TOOL_OPTIONS"
)

// ShouldInject 标签判定：开关为 on 且未注入过才执行注入，其余情况直接放行。
func ShouldInject(pod *corev1.Pod) bool {
	if pod.Labels[LabelEnable] != EnableOn {
		return false
	}
	// 防重注入：已处理过的 Pod 直接放行，避免循环修改。
	return pod.Annotations[AnnotationStatus] != StatusInjected
}

// BuildPatches 基于当前配置快照构建 Pod 的 RFC6902 JSON Patch。
func BuildPatches(pod *corev1.Pod, snap *config.Snapshot) ([]jsonpatch.JsonPatchOperation, error) {
	agentVersion := resolveAgentVersion(pod, snap)

	var ops []jsonpatch.JsonPatchOperation

	// 1. 防重注入 Annotation。
	annotations := make(map[string]string, len(pod.Annotations)+3)
	for k, v := range pod.Annotations {
		annotations[k] = v
	}
	annotations[AnnotationStatus] = StatusInjected
	annotations[AnnotationAgentVersion] = agentVersion
	if v := pod.Labels[LabelJDKVersion]; v != "" {
		annotations[AnnotationJDKVersion] = v
	}
	ops = append(ops, add("/metadata/annotations", annotations))

	// 2. Init 容器与业务容器共享的 emptyDir 临时卷。
	if !hasVolume(pod.Spec.Volumes, VolumeName) {
		vol := corev1.Volume{
			Name: VolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
		ops = append(ops, appendList("/spec/volumes", len(pod.Spec.Volumes), vol))
	}

	// 3. Init 容器：下载探针并修复权限。
	if !hasContainer(pod.Spec.InitContainers, InitContainerName) {
		ic := buildInitContainer(snap, agentVersion)
		ops = append(ops, appendList("/spec/initContainers", len(pod.Spec.InitContainers), ic))
	}

	// 4. 业务容器：挂载探针卷并注入环境变量。
	for i := range pod.Spec.Containers {
		ops = append(ops, containerOps(pod, i, snap)...)
	}
	return ops, nil
}

// containerOps 为第 i 个业务容器生成 volumeMount 与环境变量 patch。
func containerOps(pod *corev1.Pod, i int, snap *config.Snapshot) []jsonpatch.JsonPatchOperation {
	c := &pod.Spec.Containers[i]
	var ops []jsonpatch.JsonPatchOperation
	prefix := fmt.Sprintf("/spec/containers/%d", i)

	// 共享卷挂载（只读）。
	if !hasMount(c.VolumeMounts, VolumeName) {
		vm := corev1.VolumeMount{Name: VolumeName, MountPath: AgentMountPath, ReadOnly: true}
		ops = append(ops, appendList(prefix+"/volumeMounts", len(c.VolumeMounts), vm))
	}

	// 已有同名环境变量时不覆盖，尊重业务自定义。
	existing := make(map[string]struct{}, len(c.Env))
	jtIdx, jtVal := -1, ""
	for j, e := range c.Env {
		existing[e.Name] = struct{}{}
		if e.Name == JavaToolOptions {
			jtIdx, jtVal = j, e.Value
		}
	}

	// $(VAR) 依赖变量必须先于引用它的 OTEL_RESOURCE_ATTRIBUTES 注入。
	var env []corev1.EnvVar
	for _, fe := range []struct{ name, fieldPath string }{
		{"OTEL_POD_NAME", "metadata.name"},
		{"OTEL_POD_NAMESPACE", "metadata.namespace"},
		{"OTEL_NODE_NAME", "spec.nodeName"},
	} {
		if _, ok := existing[fe.name]; ok {
			continue
		}
		env = append(env, corev1.EnvVar{
			Name: fe.name,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: fe.fieldPath},
			},
		})
	}

	// JAVA_TOOL_OPTIONS：保留业务已有值并追加 -javaagent，避免覆盖导致探针失效。
	agentArg := "-javaagent:" + AgentMountPath + "/" + AgentJarName
	if jtIdx == -1 {
		env = append(env, corev1.EnvVar{Name: JavaToolOptions, Value: agentArg})
	} else if !strings.Contains(jtVal, agentArg) {
		ops = append(ops, jsonpatch.JsonPatchOperation{
			Operation: "replace",
			Path:      fmt.Sprintf("%s/env/%d/value", prefix, jtIdx),
			Value:     strings.TrimSpace(jtVal + " " + agentArg),
		})
	}

	// OTel 运行时配置。
	plainEnv := []struct {
		name, value string
	}{
		{"OTEL_SERVICE_NAME", resolveServiceName(pod)},
		{"OTEL_EXPORTER_OTLP_ENDPOINT", snap.ExporterEndpoint},
		{"OTEL_EXPORTER_OTLP_PROTOCOL", snap.ExporterProtocol},
		{"OTEL_TRACES_EXPORTER", snap.TracesExporter},
		{"OTEL_METRICS_EXPORTER", snap.MetricsExporter},
		{"OTEL_LOGS_EXPORTER", snap.LogsExporter},
		{"OTEL_PROPAGATORS", snap.Propagators},
		{"OTEL_RESOURCE_ATTRIBUTES", buildResourceAttributes(snap)},
	}
	if snap.AuthHeaders != "" {
		plainEnv = append(plainEnv, struct{ name, value string }{"OTEL_EXPORTER_OTLP_HEADERS", snap.AuthHeaders})
	}
	for _, kv := range plainEnv {
		if kv.value == "" {
			continue
		}
		if _, ok := existing[kv.name]; ok {
			continue
		}
		env = append(env, corev1.EnvVar{Name: kv.name, Value: kv.value})
	}

	if len(env) > 0 {
		if len(c.Env) == 0 {
			ops = append(ops, add(prefix+"/env", env))
		} else {
			for _, e := range env {
				ops = append(ops, add(prefix+"/env/-", e))
			}
		}
	}
	return ops
}

// buildInitContainer 构建下载探针的 Init 容器，以 root 运行修复文件权限，适配非 Root 业务容器。
func buildInitContainer(snap *config.Snapshot, agentVersion string) corev1.Container {
	url := strings.ReplaceAll(snap.AgentDownloadURL, "{version}", agentVersion)
	script := fmt.Sprintf(`set -e
AGENT_URL=%s
AGENT_PATH=%s/%s
echo "[otel-pilot] downloading OpenTelemetry JavaAgent ${AGENT_URL}"
curl -fsSL --retry 3 --retry-delay 2 --connect-timeout 10 -o "${AGENT_PATH}.tmp" "${AGENT_URL}"
mv "${AGENT_PATH}.tmp" "${AGENT_PATH}"
chmod 0644 "${AGENT_PATH}"
echo "[otel-pilot] agent ${AGENT_PATH} ready"`, url, AgentMountPath, AgentJarName)
	return corev1.Container{
		Name:    InitContainerName,
		Image:   snap.InitImage,
		Command: []string{"sh", "-c"},
		Args:    []string{script},
		VolumeMounts: []corev1.VolumeMount{
			{Name: VolumeName, MountPath: AgentMountPath},
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:                ptr.To[int64](0),
			AllowPrivilegeEscalation: ptr.To(false),
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("16Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
		},
	}
}

// buildResourceAttributes 组装资源属性：K8s 元信息（依赖 $(VAR) 展开）+ 静态配置。
func buildResourceAttributes(snap *config.Snapshot) string {
	pairs := []string{
		"k8s.pod.name=$(OTEL_POD_NAME)",
		"k8s.namespace.name=$(OTEL_POD_NAMESPACE)",
		"k8s.node.name=$(OTEL_NODE_NAME)",
	}
	keys := make([]string, 0, len(snap.ResourceAttributes))
	for k := range snap.ResourceAttributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		pairs = append(pairs, k+"="+snap.ResourceAttributes[k])
	}
	return strings.Join(pairs, ",")
}

// resolveServiceName 解析 OTEL_SERVICE_NAME：应用名标签 > 通用应用标签 > Pod 名。
func resolveServiceName(pod *corev1.Pod) string {
	if v := pod.Labels[LabelAppName]; v != "" {
		return v
	}
	for _, key := range []string{"app.kubernetes.io/name", "app"} {
		if v := pod.Labels[key]; v != "" {
			return v
		}
	}
	if pod.GenerateName != "" {
		return strings.TrimSuffix(pod.GenerateName, "-")
	}
	return pod.Name
}

// resolveAgentVersion 解析探针版本：Pod 标签覆盖 > 全局配置。
func resolveAgentVersion(pod *corev1.Pod, snap *config.Snapshot) string {
	if v := pod.Labels[LabelAgentVersion]; v != "" {
		return v
	}
	return snap.AgentVersion
}

func add(path string, value any) jsonpatch.JsonPatchOperation {
	return jsonpatch.JsonPatchOperation{Operation: "add", Path: path, Value: value}
}

// appendList 向已存在的数组追加单个元素；数组不存在（K8s omitted 字段）时创建。
func appendList[T any](path string, currentLen int, item T) jsonpatch.JsonPatchOperation {
	if currentLen == 0 {
		return add(path, []T{item})
	}
	return add(path+"/-", item)
}

func hasVolume(vols []corev1.Volume, name string) bool {
	for i := range vols {
		if vols[i].Name == name {
			return true
		}
	}
	return false
}

func hasContainer(ctrs []corev1.Container, name string) bool {
	for i := range ctrs {
		if ctrs[i].Name == name {
			return true
		}
	}
	return false
}

func hasMount(mounts []corev1.VolumeMount, name string) bool {
	for i := range mounts {
		if mounts[i].Name == name {
			return true
		}
	}
	return false
}
