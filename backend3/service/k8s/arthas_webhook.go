package k8s

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"oneops/backend3/pkg/logger"
	repok8s "oneops/backend3/repository/k8s"
)

/**
 * Arthas Webhook 自动注入服务（Pod label 触发）
 *
 * 架构（纯 Tunnel 接入，无 exec/临时容器兜底）：
 *  1. 应用在 Pod template 打 label `oneops-arthas-injection=enabled`（仅 template，勿动 Deployment selector）
 *  2. apiserver 创建 Pod 时回调本 webhook（objectSelector 命中，Pod 级粒度）
 *  3. 注入 initContainer 将 arthas-agent.jar 放入共享 emptyDir，并向主容器注入
 *     JAVA_TOOL_OPTIONS=-javaagent:...（agent 在目标 JVM 进程内启动并主动反连 tunnel-server）
 *  4. agentId 约定 $(POD_NAME)-$(POD_NAMESPACE)，K8s env 引用展开后由 JVM 原样获得
 *
 * 排除：annotation `oneops-arthas.injection/disabled: "true"`（label 被基础模板继承时二次防御）
 */

// ========== 常量与环境变量 ==========

const (
	// webhookHandlerPath TLS webhook server 的回调路径
	webhookHandlerPath = "/webhook/msre-pilot"
	// InjectLabel 触发注入的 Pod label
	InjectLabel = "oneops-arthas-injection"
	// InjectDisabledAnnotation 显式排除注入的 annotation
	InjectDisabledAnnotation = "oneops-arthas.injection/disabled"
	// InjectedMarker 已注入标记（幂等）
	InjectedMarker = "oneops-arthas.io/injected"

	injectVolumeName = "oneops-arthas"
	injectAgentMount = "/oneops/arthas"
	injectInitName   = "oneops-arthas-agent"

	defaultWebhookPort = "9443"

	// MWCName MutatingWebhookConfiguration 名称（每集群一份）
	MWCName = "msre-pilot-inject"

	// 集群内独立 webhook 服务的工作负载/配置名（与 arthas-webhook/msrepilot-deploy.yaml 对应）
	WebhookDeploymentName = "msre-pilot"
	WebhookConfigMapName  = "msre-pilot-config"

	// WebhookIntentKey 治理页注入开关意图（写入集群 ConfigMap，msre-pilot 控制器消费）
	WebhookIntentKey = "WEBHOOK_ENABLED"

	// 配置 key（k8s_diagnostic_config，clusterID=0 全局默认，可按集群覆盖）
	ConfigKeyTunnelWS    = "arthas.tunnel.ws"
	ConfigKeyInitImage   = "arthas.init.image"
	ConfigKeyWebhookURL  = "arthas.webhook.url"  // apiserver 回调地址（治理页配置，免环境变量）
	ConfigKeyWebhookPort = "arthas.webhook.port" // TLS 监听端口，默认 9443

	defaultInitImage = "registry.cn-hangzhou.aliyuncs.com/oneops/arthas-agent:3.7.2"
)

// ArthasWebhookService Webhook 注入服务
type ArthasWebhookService struct {
	clusterSvc     *K8sClusterService
	diagnosticRepo *repok8s.DiagnosticRepository
}

// NewArthasWebhookService 构造
func NewArthasWebhookService(clusterSvc *K8sClusterService, diagnosticRepo *repok8s.DiagnosticRepository) *ArthasWebhookService {
	return &ArthasWebhookService{clusterSvc: clusterSvc, diagnosticRepo: diagnosticRepo}
}

// ========== 配置存取 ==========

// clusterKey 配置存储键：集群 DB 主键转字符串，"0" 表示全局默认
func clusterKey(clusterID uint) string {
	return strconv.FormatUint(uint64(clusterID), 10)
}

// GetTunnelWS 获取 agent 反连 tunnel-server 的 WS 地址（集群级覆盖 > 全局 > 环境变量 > 默认）
func (s *ArthasWebhookService) GetTunnelWS(clusterID uint) string {
	if v := s.diagnosticRepo.GetConfigValue(clusterKey(clusterID), ConfigKeyTunnelWS, ""); v != "" {
		return v
	}
	if v := s.diagnosticRepo.GetConfigValue(clusterKey(0), ConfigKeyTunnelWS, ""); v != "" {
		return v
	}
	if v := os.Getenv("ARTHAS_TUNNEL_WS"); v != "" {
		return v
	}
	return "ws://arthas-tunnel.test:7777/ws"
}

// GetInitImage 获取 initContainer 镜像
func (s *ArthasWebhookService) GetInitImage(clusterID uint) string {
	if v := s.diagnosticRepo.GetConfigValue(clusterKey(clusterID), ConfigKeyInitImage, ""); v != "" {
		return v
	}
	if v := s.diagnosticRepo.GetConfigValue(clusterKey(0), ConfigKeyInitImage, ""); v != "" {
		return v
	}
	return defaultInitImage
}

// GetWebhookURL apiserver 回调地址（全局 DB > 环境变量兜底）
func (s *ArthasWebhookService) GetWebhookURL() string {
	if v := s.diagnosticRepo.GetConfigValue(clusterKey(0), ConfigKeyWebhookURL, ""); v != "" {
		return v
	}
	return os.Getenv("ARTHAS_WEBHOOK_URL")
}

// webhookPort TLS 监听端口（全局 DB > 环境变量 > 9443）
func (s *ArthasWebhookService) webhookPort() string {
	if v := s.diagnosticRepo.GetConfigValue(clusterKey(0), ConfigKeyWebhookPort, ""); v != "" {
		return v
	}
	if p := os.Getenv("ARTHAS_WEBHOOK_PORT"); p != "" {
		return p
	}
	return defaultWebhookPort
}

// SaveWebhookConfig 保存注入配置（clusterID=0 为全局默认）
// 保存回调地址后自动同步所有已启用集群的 MWC（url/caBundle），改地址保存即生效，无需重新拨开关
func (s *ArthasWebhookService) SaveWebhookConfig(clusterID uint, tunnelWS, initImage, webhookURL, tunnelURL, tunnelSession string) error {
	if tunnelWS != "" {
		if err := s.diagnosticRepo.SetConfigValue(clusterKey(clusterID), ConfigKeyTunnelWS, tunnelWS, "Arthas agent 反连 tunnel-server 地址"); err != nil {
			return err
		}
	}
	if initImage != "" {
		if err := s.diagnosticRepo.SetConfigValue(clusterKey(clusterID), ConfigKeyInitImage, initImage, "Arthas agent initContainer 镜像"); err != nil {
			return err
		}
	}
	if webhookURL != "" {
		if _, err := url.ParseRequestURI(webhookURL); err != nil {
			return fmt.Errorf("回调地址格式错误: %w", err)
		}
		if err := s.diagnosticRepo.SetConfigValue(clusterKey(0), ConfigKeyWebhookURL, webhookURL, "apiserver 回调 webhook 的 HTTPS 地址（全局）"); err != nil {
			return err
		}
	}
	if tunnelURL != "" {
		if err := s.diagnosticRepo.SetConfigValue(clusterKey(clusterID), ConfigKeyTunnelURL, tunnelURL, "平台访问 tunnel-server 的 HTTP 地址（探测 agent 在线）"); err != nil {
			return err
		}
	}
	if tunnelSession != "" {
		if err := s.diagnosticRepo.SetConfigValue(clusterKey(clusterID), ConfigKeyTunnelSessionURL, tunnelSession, "平台访问 tunnel-server 的 WS 会话地址（诊断终端）"); err != nil {
			return err
		}
	}

	// 回调地址变化时同步已启用集群的 MWC，避免"MWC 指向旧地址 + failurePolicy=Ignore 静默失效"
	// 同步失败不影响保存结果（配置已落库，下次启用/保存会再次同步），仅记警告日志
	if webhookURL != "" {
		s.syncEnabledMWCs()
	}
	// 注入参数变化时同步集群内独立 webhook 服务的 ConfigMap 并滚动重启（svc:// 形态）
	if tunnelWS != "" || initImage != "" {
		s.syncWebhookWorkloads()
	}
	return nil
}

// syncWebhookWorkloads 同步集群内独立 webhook 服务（svc:// 形态）的注入参数：
// upsert ConfigMap（ARTHAS_TUNNEL_WS/ARTHAS_INIT_IMAGE）即完成——webhook 服务
// informer watch 此 ConfigMap 热加载，无需滚动重启（无重启窗口、无漏注入风险）。
// URL 型（本进程即 webhook）无需同步——本进程每次请求实时读 DB。
func (s *ArthasWebhookService) syncWebhookWorkloads() {
	ref, ok := parseSvcRef(s.GetWebhookURL())
	if !ok {
		return // URL 型或未配置，无集群内工作负载
	}
	clusterIDs, err := s.clusterSvc.ListAllClusterIDs()
	if err != nil {
		logger.Warn("查询集群列表失败，跳过 webhook 工作负载同步", zap.Error(err))
		return
	}
	for _, id := range clusterIDs {
		client, err := s.clusterSvc.GetClient(id)
		if err != nil {
			continue
		}
		// 按集群读配置（GetXxx(id) 内含优先级链：集群级覆盖 > 全局 > 默认）：
		// svc:// 形态一个集群一套 ConfigMap，固定读全局(0)会导致
		// "集群级保存的 initImage/tunnelWS 永远同步不下去（被全局/默认值覆盖）"
		desired := map[string]string{
			"ARTHAS_TUNNEL_WS":  s.GetTunnelWS(id),
			"ARTHAS_INIT_IMAGE": s.GetInitImage(id),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		if err := s.upsertWebhookConfigMap(ctx, client, ref.namespace, desired); err != nil {
			logger.Warn("同步 webhook ConfigMap 失败", zap.Uint64("clusterID", uint64(id)), zap.Error(err))
		} else {
			logger.Info("webhook ConfigMap 已同步（服务 watch 热加载即时生效）",
				zap.Uint64("clusterID", uint64(id)), zap.String("namespace", ref.namespace))
		}
		cancel()
	}
}

// upsertWebhookConfigMap 创建/更新集群内 webhook 配置 ConfigMap（值变化才更新，避免无谓滚动）
func (s *ArthasWebhookService) upsertWebhookConfigMap(ctx context.Context, client *kubernetes.Clientset, ns string, data map[string]string) error {
	cm, err := client.CoreV1().ConfigMaps(ns).Get(ctx, WebhookConfigMapName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.CoreV1().ConfigMaps(ns).Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: WebhookConfigMapName, Namespace: ns},
			Data:       data,
		}, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	changed := false
	for k, v := range data {
		if cm.Data[k] != v {
			cm.Data[k] = v
			changed = true
		}
	}
	if !changed {
		return nil
	}
	_, err = client.CoreV1().ConfigMaps(ns).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

// syncEnabledMWCs 对所有已创建 MWC 的集群重建 webhook 配置（url/caBundle 同步）
// 仅处理已启用（MWC 存在）的集群，未启用的不代为开启
func (s *ArthasWebhookService) syncEnabledMWCs() {
	clusterIDs, err := s.clusterSvc.ListAllClusterIDs()
	if err != nil {
		logger.Warn("查询集群列表失败，跳过 MWC 同步", zap.Error(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, id := range clusterIDs {
		client, err := s.clusterSvc.GetClient(id)
		if err != nil {
			continue
		}
		if _, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().
			Get(ctx, MWCName, metav1.GetOptions{}); err != nil {
			continue // 未启用（NotFound）或查询失败，跳过
		}
		if err := s.EnableWebhook(id); err != nil {
			logger.Warn("同步 MWC 失败（回调地址已保存，下次启用/保存时重试）",
				zap.Uint64("clusterID", uint64(id)), zap.Error(err))
		} else {
			logger.Info("MWC 已同步（回调地址更新生效）", zap.Uint64("clusterID", uint64(id)))
		}
	}
}

// ========== Patch 构建（核心注入逻辑） ==========

// jsonPatchOp RFC6902 JSON Patch 操作
type jsonPatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// ShouldInject 判断 Pod 是否应注入：label 触发 + annotation 排除 + 幂等
func ShouldInject(pod *corev1.Pod) bool {
	if pod.Labels[InjectLabel] != "enabled" {
		return false
	}
	if pod.Annotations[InjectDisabledAnnotation] == "true" {
		return false
	}
	if pod.Annotations[InjectedMarker] == "true" {
		return false
	}
	return true
}

// BuildInjectionPatch 生成注入 JSON Patch
// tunnelWS 形如 ws://arthas-tunnel.oneops:7777/ws；tunnel/agentId 由 initContainer 通过 sed 写入 arthas.properties
func BuildInjectionPatch(pod *corev1.Pod, tunnelWS, initImage string) ([]jsonPatchOp, error) {
	if tunnelWS == "" {
		return nil, fmt.Errorf("tunnel-server WS 地址未配置（%s）", ConfigKeyTunnelWS)
	}
	patches := []jsonPatchOp{}

	// 1) 标记已注入（幂等 + 可观测）
	if pod.Annotations == nil {
		patches = append(patches, jsonPatchOp{Op: "add", Path: "/metadata/annotations", Value: map[string]string{}})
	}
	patches = append(patches, jsonPatchOp{
		Op: "add", Path: "/metadata/annotations/" + escapeJSONPointer(InjectedMarker), Value: "true",
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
	// 第一个分号前必须是 jar 路径；直接传 tunnelServer=xxx 作为 javaagent 参数会被当作 jar 路径吞掉。
	// 故 tunnel 配置改走 arthas.properties 文件（镜像内预置 tunnelServer，agentId 通过
	// 主容器 JAVA_TOOL_OPTIONS 注入 -Darthas.agentId 系统属性覆盖，因优先级 System Properties > 文件）。
	initCtr := corev1.Container{
		Name:            injectInitName,
		Image:           initImage,
		ImagePullPolicy: corev1.PullAlways, // 总是拉取最新镜像，避免用本地缓存的旧镜像
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
	// 配置优先级（官方）：命令行参数 > System Properties(-D) > arthas.properties
	// -Darthas.tunnelServer=$(ARTHAS_TUNNEL_SERVER)：env 引用 K8s 展开后作为系统属性，
	//   优先级高于镜像内 arthas.properties 默认值；治理页改 tunnelWS 即可覆盖镜像默认。
	// -Darthas.agentId=$(POD_NAME)-$(POD_NAMESPACE)：每 Pod 唯一 agentId
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

		// env：先确保 downward API env 存在（K8s 按出现顺序展开 $(VAR) 引用，
		// 缺失时 $(POD_NAME) 会以字面量进入 agentId，导致所有 Pod agentId 相同）
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

// ========== Admission 处理 ==========

// HandleAdmissionReview 处理 AdmissionReview 请求（TLS server 回调入口，URL 型模式）
// 集群内部署形态（svc:// 回调）由独立工程 arthas-webhook/ 承接，本进程仅服务 URL 型模式；
// 注入 Patch 逻辑改动时须同步两份（arthas-webhook/inject.go buildInjectionPatch）
func (s *ArthasWebhookService) HandleAdmissionReview(c *gin.Context) {
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
				zap.String("namespace", pod.Namespace), zap.String("pod", pod.Name),
				zap.String("injectLabel", pod.Labels[InjectLabel]),
				zap.Bool("shouldInject", ShouldInject(pod)))
			if ShouldInject(pod) {
				if patches, err := BuildInjectionPatch(pod, s.GetTunnelWS(0), s.GetInitImage(0)); err != nil {
					resp.Allowed = false
					resp.Result = &metav1.Status{Message: err.Error()}
				} else if len(patches) > 0 {
					patchBytes, err := json.Marshal(patches)
					if err != nil {
						resp.Allowed = false
						resp.Result = &metav1.Status{Message: fmt.Sprintf("marshal patch failed: %v", err)}
					} else {
						patchType := admissionv1.PatchTypeJSONPatch
						resp.PatchType = &patchType
						// Patch 为 []byte：JSON 序列化时自动 base64（wire 协议自带），
						// 手动再编会双重编码导致 apiserver 静默丢弃注入 patch
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

// ========== TLS Webhook Server（惰性启动，零环境变量） ==========

var (
	tlsServerMu      sync.Mutex
	tlsServerRunning bool
)

// certDirOf webhook 证书目录（env 可覆盖）
func certDirOf() string {
	if d := os.Getenv("ARTHAS_WEBHOOK_CERT_DIR"); d != "" {
		return d
	}
	return "./data/arthas-webhook-tls"
}

// StartTLSServer 进程启动时的入口：已配置回调地址则拉起，否则跳过（治理页启用时再按需拉起）
func (s *ArthasWebhookService) StartTLSServer() {
	if s.GetWebhookURL() == "" && os.Getenv("ARTHAS_WEBHOOK_SERVICE_NAME") == "" {
		logger.Info("Arthas webhook 未配置回调地址，TLS server 暂不启动（治理页配置并启用后自动拉起）")
		return
	}
	if err := s.ensureCertAndServer(); err != nil {
		logger.Error("Arthas webhook TLS server 启动失败", zap.Error(err))
	}
}

// webhookSANHosts 证书 SAN 主机列表：回调地址 host 自动推导 + env 追加
func (s *ArthasWebhookService) webhookSANHosts() []string {
	var hosts []string
	if u := s.GetWebhookURL(); u != "" {
		if parsed, err := url.Parse(u); err == nil && parsed.Hostname() != "" {
			hosts = append(hosts, parsed.Hostname())
		}
	}
	if extra := os.Getenv("ARTHAS_WEBHOOK_TLS_DNS"); extra != "" {
		for _, d := range strings.Split(extra, ",") {
			if d = strings.TrimSpace(d); d != "" {
				hosts = append(hosts, d)
			}
		}
	}
	return hosts
}

// ensureCertAndServer 幂等：证书就绪（SAN 覆盖回调 host，不足则重签）+ TLS server 未运行则拉起
func (s *ArthasWebhookService) ensureCertAndServer() error {
	tlsServerMu.Lock()
	defer tlsServerMu.Unlock()

	certDir := certDirOf()
	certPath := filepath.Join(certDir, "tls.crt")
	keyPath := filepath.Join(certDir, "tls.key")

	if err := ensureCert(certDir, s.webhookSANHosts()); err != nil {
		return err
	}

	if !tlsServerRunning {
		port := s.webhookPort()
		gin.SetMode(gin.ReleaseMode)
		engine := gin.New()
		engine.POST(webhookHandlerPath, s.HandleAdmissionReview)
		// 健康检查（无鉴权，仅探活）
		engine.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
		srv := &http.Server{
			Addr:    ":" + port,
			Handler: engine,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				ClientAuth: tls.NoClientCert,
			},
			ReadHeaderTimeout: 5 * time.Second,
		}
		go func() {
			tlsServerRunning = true
			logger.Info("Arthas webhook TLS server 启动",
				zap.String("port", port), zap.String("path", webhookHandlerPath))
			if err := srv.ListenAndServeTLS(certPath, keyPath); err != nil && err != http.ErrServerClosed {
				tlsServerRunning = false
				logger.Error("Arthas webhook TLS server 退出", zap.Error(err))
			}
		}()
	}
	return nil
}

// ensureCert 确保证书存在且 SAN 覆盖全部 hosts；缺失或 SAN 不足则重新签发
func ensureCert(dir string, hosts []string) error {
	certPath := filepath.Join(dir, "tls.crt")
	if pemBytes, err := os.ReadFile(certPath); err == nil {
		if block, _ := pem.Decode(pemBytes); block != nil {
			if cert, err := x509.ParseCertificate(block.Bytes); err == nil && certCoversHosts(cert, hosts) {
				return nil
			}
		}
		logger.Info("Arthas webhook 证书 SAN 未覆盖回调地址，重新签发", zap.Strings("hosts", hosts))
	}
	return generateSelfSignedCert(dir, hosts)
}

// certCoversHosts 校验证书 SAN 是否覆盖所有 host（域名 → DNSNames，IP → IPAddresses）
func certCoversHosts(cert *x509.Certificate, hosts []string) bool {
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			covered := false
			for _, cip := range cert.IPAddresses {
				if cip.Equal(ip) {
					covered = true
					break
				}
			}
			if !covered {
				return false
			}
			continue
		}
		found := false
		for _, d := range cert.DNSNames {
			if strings.EqualFold(d, h) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ========== MutatingWebhookConfiguration 管理 ==========

// svcWebhookRef Service 型回调地址引用
type svcWebhookRef struct {
	namespace string
	name      string
	port      int32
}

// parseSvcRef 解析 svc://<namespace>/<name>:<port>（Service 型回调，集群内独立 webhook）
// 与 URL 型（https://...）互斥；治理页同一字段按前缀智能识别
func parseSvcRef(u string) (*svcWebhookRef, bool) {
	rest, ok := strings.CutPrefix(u, "svc://")
	if !ok {
		return nil, false
	}
	nsName, portStr, _ := strings.Cut(rest, ":")
	ns, name, ok := strings.Cut(nsName, "/")
	if !ok || ns == "" || name == "" {
		return nil, false
	}
	port := int32(9443)
	if p, err := strconv.ParseInt(portStr, 10, 32); err == nil && p > 0 && p < 65536 {
		port = int32(p)
	}
	return &svcWebhookRef{namespace: ns, name: name, port: port}, true
}

// webhookClientConfig apiserver 回调地址（仅 URL 型使用，svc:// 形态 MWC 由
// 集群内 msre-pilot 控制器构造）：
//   - https://... → URL 型（apiserver 直连平台侧，要求 master 可出网到该地址）
//   - 未配置      → env Service 兜底（in-cluster 部署形态）
func (s *ArthasWebhookService) webhookClientConfig() (*admissionregv1.WebhookClientConfig, error) {
	if u := s.GetWebhookURL(); u != "" {
		return &admissionregv1.WebhookClientConfig{URL: &u}, nil
	}
	ns := os.Getenv("ARTHAS_WEBHOOK_SERVICE_NS")
	if ns == "" {
		ns = "oneops"
	}
	name := os.Getenv("ARTHAS_WEBHOOK_SERVICE_NAME")
	if name == "" {
		name = "oneops"
	}
	port := int32(9443)
	if p := s.webhookPort(); p != "" {
		var v int32
		if _, err := fmt.Sscanf(p, "%d", &v); err == nil && v > 0 && v < 65536 {
			port = v
		}
	}
	path := webhookHandlerPath
	return &admissionregv1.WebhookClientConfig{
		Service: &admissionregv1.ServiceReference{Namespace: ns, Name: name, Path: &path, Port: &port},
	}, nil
}

// caBundleForWebhook 读取 CA 证书（仅 URL 型使用：本地 certDirOf()/ca.crt，
// 本进程 ensureCert 维护；svc:// 形态的 MWC/caBundle 归集群内 msre-pilot 控制器）
func (s *ArthasWebhookService) caBundleForWebhook() ([]byte, error) {
	return os.ReadFile(filepath.Join(certDirOf(), "ca.crt"))
}

// EnableWebhook 在指定集群创建/更新 MutatingWebhookConfiguration（Pod label 触发）
func (s *ArthasWebhookService) EnableWebhook(clusterID uint) error {
	client, err := s.clusterSvc.GetClient(clusterID)
	if err != nil {
		return fmt.Errorf("获取集群客户端失败: %w", err)
	}

	// 前置：配置回调地址（DB/环境变量）
	//   - svc:// 形态：MWC 生命周期归集群内 msre-pilot 控制器管理，
	//     平台只写启用意图（ConfigMap WEBHOOK_ENABLED）并等待收敛
	//   - https:// 形态：admission 服务即本进程，需证书/TLS 就绪，平台直建 MWC
	isSvcMode := false
	var svcRef *svcWebhookRef
	if u := s.GetWebhookURL(); u != "" {
		svcRef, isSvcMode = parseSvcRef(u)
	}
	if s.GetWebhookURL() == "" && os.Getenv("ARTHAS_WEBHOOK_SERVICE_NAME") == "" {
		return fmt.Errorf("请先在治理页保存配置：回调地址（集群内服务 svc://<ns>/<name>:9443，或 https://<OneOps地址>:9443%s）", webhookHandlerPath)
	}
	if isSvcMode {
		return s.setWebhookIntent(clusterID, svcRef, true)
	}
	if err := s.ensureCertAndServer(); err != nil {
		return fmt.Errorf("webhook TLS 服务就绪失败: %w", err)
	}

	caPEM, err := s.caBundleForWebhook()
	if err != nil {
		return err
	}

	clientCfg, err := s.webhookClientConfig()
	if err != nil {
		return err
	}
	clientCfg.CABundle = caPEM

	enabled := "enabled"
	none := admissionregv1.SideEffectClassNone
	ignore := admissionregv1.Ignore
	equivalent := admissionregv1.Equivalent
	timeout := int32(5)
	desired := &admissionregv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: MWCName},
		Webhooks: []admissionregv1.MutatingWebhook{{
			Name:         "msre-pilot.oneops.cn",
			ClientConfig: *clientCfg,
			Rules: []admissionregv1.RuleWithOperations{{
				Operations: []admissionregv1.OperationType{admissionregv1.Create},
				Rule: admissionregv1.Rule{
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Resources:   []string{"pods"},
				},
			}},
			// Pod 级触发：仅 label oneops-arthas-injection=enabled 的 Pod 进入注入逻辑
			ObjectSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{InjectLabel: enabled},
			},
			// 防御性排除系统命名空间
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "kubernetes.io/metadata.name",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{"kube-system", "kube-public", "kube-node-lease"},
				}},
			},
			FailurePolicy:           &ignore, // webhook 不可用时放行，不影响业务 Pod 创建
			MatchPolicy:             &equivalent,
			SideEffects:             &none,
			TimeoutSeconds:          &timeout,
			AdmissionReviewVersions: []string{"v1"},
		}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, MWCName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx, desired, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return fmt.Errorf("查询 MWC 失败: %w", err)
	}
	_, err = client.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, desired, metav1.UpdateOptions{})
	return err
}

// setWebhookIntent svc:// 形态：写启用/禁用意图到集群内 webhook ConfigMap，
// 由 msre-pilot 控制器（15s 一轮）收敛 MWC 创建/删除。返回前轮询确认收敛
// （超时不视为失败——意图已持久化，控制器最终一致）。
func (s *ArthasWebhookService) setWebhookIntent(clusterID uint, ref *svcWebhookRef, enabled bool) error {
	client, err := s.clusterSvc.GetClient(clusterID)
	if err != nil {
		return fmt.Errorf("获取集群客户端失败: %w", err)
	}
	val := "false"
	if enabled {
		val = "true"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.upsertWebhookConfigMap(ctx, client, ref.namespace, map[string]string{WebhookIntentKey: val}); err != nil {
		return fmt.Errorf("写入注入开关意图失败: %w", err)
	}
	// 等待控制器收敛：msre-pilot watch 意图变化为快路径（通常 1s 内），
	// 5s 上限兜底（watch 断线时最坏 15s ticker 收敛，超时仅 Warn 最终一致）
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		_, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, MWCName, metav1.GetOptions{})
		if enabled && err == nil {
			return nil
		}
		if !enabled && apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("查询 MWC 失败: %w", err)
		}
		time.Sleep(1 * time.Second)
	}
	logger.Warn("等待 msre-pilot 控制器收敛超时（意图已写入，将最终收敛）",
		zap.Uint64("clusterID", uint64(clusterID)), zap.Bool("enabled", enabled))
	return nil
}

// DisableWebhook 关闭指定集群注入（已注入的 Pod 不受影响）
// svc:// 形态写意图由 msre-pilot 控制器删 MWC；URL 型（本进程即 webhook）直接删
func (s *ArthasWebhookService) DisableWebhook(clusterID uint) error {
	client, err := s.clusterSvc.GetClient(clusterID)
	if err != nil {
		return fmt.Errorf("获取集群客户端失败: %w", err)
	}
	if ref, ok := parseSvcRef(s.GetWebhookURL()); ok {
		return s.setWebhookIntent(clusterID, ref, false)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err = client.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(ctx, MWCName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// WebhookStatus 注入开关状态
type WebhookStatus struct {
	Enabled    bool   `json:"enabled"`
	TunnelWS   string `json:"tunnelWS"`   // agent 反连地址（集群内）
	TunnelURL  string `json:"tunnelURL"`  // 平台 HTTP 探测地址
	TunnelSess string `json:"tunnelSess"` // 平台 WS 会话地址
	InitImage  string `json:"initImage"`
	WebhookURL string `json:"webhookURL"`
	ObjectRule string `json:"objectRule"`
}

// GetWebhookStatus 查询集群注入状态（MWC 存在即启用）
func (s *ArthasWebhookService) GetWebhookStatus(clusterID uint) (*WebhookStatus, error) {
	status := &WebhookStatus{
		TunnelWS:   s.GetTunnelWS(clusterID),
		TunnelURL:  s.diagnosticRepo.GetConfigValue(clusterKey(clusterID), ConfigKeyTunnelURL, ""),
		TunnelSess: s.diagnosticRepo.GetConfigValue(clusterKey(clusterID), ConfigKeyTunnelSessionURL, ""),
		InitImage:  s.GetInitImage(clusterID),
		WebhookURL: s.GetWebhookURL(),
		ObjectRule: fmt.Sprintf("Pod label %s=enabled（annotation %s=true 可排除）", InjectLabel, InjectDisabledAnnotation),
	}
	client, err := s.clusterSvc.GetClient(clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取集群客户端失败: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// svc:// 形态：状态 = 治理页意图（MWC 由 msre-pilot 控制器最终一致地建/删）
	if ref, ok := parseSvcRef(status.WebhookURL); ok {
		cm, err := client.CoreV1().ConfigMaps(ref.namespace).Get(ctx, WebhookConfigMapName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			status.Enabled = false
			return status, nil
		}
		if err != nil {
			return nil, err
		}
		status.Enabled = cm.Data[WebhookIntentKey] == "true"
		return status, nil
	}
	// URL 型：MWC 存在即启用
	_, err = client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, MWCName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		status.Enabled = false
		return status, nil
	}
	if err != nil {
		return nil, err
	}
	status.Enabled = true
	return status, nil
}

// ========== 自签证书 ==========

// generateSelfSignedCert 生成自签 CA + 服务证书（apiserver 经 MWC caBundle 信任）
// extraHosts 为需覆盖的额外主机（回调地址 host，域名或 IP）
func generateSelfSignedCert(dir string, extraHosts []string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	// CA
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          randomSerial(),
		Subject:               pkix.Name{CommonName: "msre-pilot-ca", Organization: []string{"OneOps"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		return err
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return err
	}

	// 服务证书：SAN 覆盖 in-cluster service 名 + 回调地址 host（域名/IP 自动分类）
	srvKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	dnsNames := []string{
		"oneops", "oneops.oneops", "oneops.oneops.svc", "oneops.oneops.svc.cluster.local",
		"localhost",
	}
	ipAddrs := []net.IP{net.ParseIP("127.0.0.1")}
	for _, h := range extraHosts {
		if ip := net.ParseIP(h); ip != nil {
			ipAddrs = append(ipAddrs, ip)
		} else {
			dnsNames = append(dnsNames, h)
		}
	}
	srvTmpl := &x509.Certificate{
		SerialNumber: randomSerial(),
		Subject:      pkix.Name{CommonName: "msre-pilot", Organization: []string{"OneOps"}},
		DNSNames:     dnsNames,
		IPAddresses:  ipAddrs,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	srvDER, err := x509.CreateCertificate(rand.Reader, srvTmpl, caCert, &srvKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	// 落盘（PEM）
	if err := writePEM(filepath.Join(dir, "ca.crt"), "CERTIFICATE", caDER); err != nil {
		return err
	}
	if err := writePEM(filepath.Join(dir, "tls.crt"), "CERTIFICATE", srvDER); err != nil {
		return err
	}
	srvKeyDER, err := x509.MarshalECPrivateKey(srvKey)
	if err != nil {
		return err
	}
	return writePEM(filepath.Join(dir, "tls.key"), "EC PRIVATE KEY", srvKeyDER)
}

func randomSerial() *big.Int {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return big.NewInt(time.Now().UnixNano())
	}
	return n
}

func writePEM(path, blockType string, der []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: blockType, Bytes: der})
}
