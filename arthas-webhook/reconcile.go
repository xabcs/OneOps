package main

import (
	"bytes"
	"context"
	"os"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

/**
 * MWC 生命周期控制器（意图与实现分离）：
 * 治理页"启用/禁用"只写意图——ConfigMap msre-pilot-config 的 WEBHOOK_ENABLED，
 * 本控制器持续收敛集群实际状态（level-triggered，漂移自愈）：
 *   - enabled ：确保 MWC msre-pilot-inject 存在且 spec/caBundle 正确
 *   - disabled：删除 MWC（已注入的 Pod 不受影响）
 *   - 孤儿清理：带所有权 label 但名字非当前名（改名遗留）、历史名（无 label 收编）
 * 收敛触发为双路径：意图 watch 快路径（亚秒）+ 15s ticker 兜底。
 *
 * 所有权 label：按属性识别本体系资源而非按名字，改名后旧资源仍可被识别回收。
 * 平台侧因此不再需要 cluster-scoped MWC 写权限（管理面只下发意图）。
 */

const (
	// intentConfigMap 意图载体（与注入参数同 ConfigMap，治理页写入）
	intentConfigMap = "msre-pilot-config"
	// intentKey 启用意图："true"/"false"，缺失视为 false
	intentKey = "WEBHOOK_ENABLED"

	reconcileInterval = 15 * time.Second

	// legacyMWCName 改名前的历史 MWC 名（无所有权 label，一次性收编清理）
	legacyMWCName = "oneops-arthas-inject"
)

// mwcLabels 所有权标记（识别"我创建的" MWC，与名字解耦）
var mwcLabels = map[string]string{
	"oneops.io/managed-by": "oneops",
	"oneops.io/component":  "arthas-webhook",
}

// ========== 注入参数热缓存（admission 实时读，ConfigMap 变更亚秒生效） ==========

// configCache 缓存 msre-pilot-config 全部 key/value，由 watchIntent 的 informer
// 回调维护（含初次 List 全量同步）。admission 处理实时读取，治理页保存 →
// 平台同步 ConfigMap → watch 推送 → 新值即时生效，全程无 Pod 重启。
// 未填充（informer 未同步/无 K8s 凭据/CM 被删）时读取方降级到启动 env。
var configCache atomic.Pointer[map[string]string]

// storeConfig 快照 ConfigMap data（防御性拷贝，不共享 informer 内部 map）
func storeConfig(data map[string]string) {
	m := make(map[string]string, len(data))
	for k, v := range data {
		m[k] = v
	}
	configCache.Store(&m)
}

// storeConfigFromObj informer 对象转配置快照（断言失败静默跳过，如 tombstone）
func storeConfigFromObj(obj any) {
	if cm, ok := obj.(*corev1.ConfigMap); ok {
		storeConfig(cm.Data)
	}
}

// tunnelWSConfig agent 反连 tunnel-server 地址：informer 缓存 > 启动 env
func tunnelWSConfig() string {
	if p := configCache.Load(); p != nil {
		if v := (*p)["ARTHAS_TUNNEL_WS"]; v != "" {
			return v
		}
	}
	return os.Getenv("ARTHAS_TUNNEL_WS")
}

// initImageConfig initContainer 镜像：informer 缓存 > 启动 env > 默认
func initImageConfig() string {
	if p := configCache.Load(); p != nil {
		if v := (*p)["ARTHAS_INIT_IMAGE"]; v != "" {
			return v
		}
	}
	if v := os.Getenv("ARTHAS_INIT_IMAGE"); v != "" {
		return v
	}
	return defaultInitImage
}

// startReconciler 启动 MWC 控制循环（事件驱动 + level-triggered 兜底）：
//   - 快路径：watch 意图 ConfigMap，治理页写入 → 亚秒级触发收敛
//   - 慢路径：15s ticker 兜底（漂移自愈 + watch 断线/漏事件补偿）
func startReconciler(cfg rotatorConfig, port int32) {
	client, err := buildKubeClient()
	if err != nil {
		logger.Warn("MWC 控制器禁用（无 K8s 凭据，准入规则需手工维护）", zap.Error(err))
		return
	}
	go func() {
		// 容量 1 的唤醒信号：合并事件风暴，reconcile 始终串行执行无竞态
		wake := make(chan struct{}, 1)
		scheduleReconcile := func() {
			select {
			case wake <- struct{}{}:
			default: // 已有待处理唤醒，合并
			}
		}
		go watchIntent(client, cfg.namespace, scheduleReconcile)
		logger.Info("MWC 控制器已启动",
			zap.Duration("interval", reconcileInterval), zap.String("mwc", mwcName))
		runReconcileOnce(client, cfg, port) // 启动即一轮
		t := time.NewTicker(reconcileInterval)
		defer t.Stop()
		for {
			select {
			case <-t.C: // 周期兜底
			case <-wake: // 意图变化快路径
				logger.Info("意图事件触发快路径收敛")
			}
			runReconcileOnce(client, cfg, port)
		}
	}()
}

// watchIntent 快路径：informer 只监听意图 ConfigMap（fieldSelector 限定单对象），
// 意图值变化即唤醒 reconcile。断线由 client-go 自动重连，重连间隙由 ticker 兜底。
func watchIntent(client *kubernetes.Clientset, ns string, wake func()) {
	lw := cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			opts.FieldSelector = "metadata.name=" + intentConfigMap
			return client.CoreV1().ConfigMaps(ns).List(context.Background(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (k8swatch.Interface, error) {
			opts.FieldSelector = "metadata.name=" + intentConfigMap
			return client.CoreV1().ConfigMaps(ns).Watch(context.Background(), opts)
		},
	}
	_, ctrl := cache.NewInformer(&lw, &corev1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
		// 三个回调都先刷新注入参数缓存（含 ARTHAS_TUNNEL_WS/ARTHAS_INIT_IMAGE 等任意 key），
		// 再按需唤醒 reconcile（仅意图值变化才唤醒，参数变化无需收敛 MWC）
		AddFunc: func(obj any) {
			storeConfigFromObj(obj)
			wake()
		},
		UpdateFunc: func(old, cur any) {
			storeConfigFromObj(cur)
			o, ok1 := old.(*corev1.ConfigMap)
			c, ok2 := cur.(*corev1.ConfigMap)
			if ok1 && ok2 && o.Data[intentKey] == c.Data[intentKey] {
				return // 意图值未变（注入参数等其他 key 更新），不触发
			}
			wake()
		},
		DeleteFunc: func(obj any) { // CM 被删 = 意图缺失 = disabled；参数缓存清空降级 env
			configCache.Store(nil)
			wake()
		},
	})
	ctrl.Run(make(chan struct{})) // 进程生命周期内常驻
}

func runReconcileOnce(client *kubernetes.Clientset, cfg rotatorConfig, port int32) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	enabled, err := readIntent(ctx, client, cfg.namespace)
	if err != nil {
		logger.Error("读取注入意图失败（下轮重试）", zap.Error(err))
		return
	}
	cleanupOrphanMWCs(ctx, client)
	if enabled {
		ensureMWC(ctx, client, cfg, port)
	} else {
		deleteMWC(ctx, client)
	}
}

// readIntent 读治理页写入的启用意图（ConfigMap/key 缺失 = false）
func readIntent(ctx context.Context, client *kubernetes.Clientset, ns string) (bool, error) {
	cm, err := client.CoreV1().ConfigMaps(ns).Get(ctx, intentConfigMap, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return cm.Data[intentKey] == "true", nil
}

// cleanupOrphanMWCs 清理孤儿 MWC：历史名（一次性收编）或带所有权 label 的非当前名（改名遗留）
func cleanupOrphanMWCs(ctx context.Context, client *kubernetes.Clientset) {
	list, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Error("列出 MWC 失败（孤儿清理跳过，下轮重试）", zap.Error(err))
		return
	}
	for i := range list.Items {
		m := &list.Items[i]
		if m.Name == mwcName {
			continue
		}
		owned := m.Labels["oneops.io/managed-by"] == "oneops" &&
			m.Labels["oneops.io/component"] == "arthas-webhook"
		if !owned && m.Name != legacyMWCName {
			continue
		}
		if err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().
			Delete(ctx, m.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			logger.Error("清理孤儿 MWC 失败", zap.String("mwc", m.Name), zap.Error(err))
			continue
		}
		logger.Info("已清理孤儿 MWC（历史名/改名遗留）", zap.String("mwc", m.Name))
	}
}

// ensureMWC 确保 MWC 存在且与期望一致（漂移自愈：caBundle/Service 指向/选择器）
func ensureMWC(ctx context.Context, client *kubernetes.Clientset, cfg rotatorConfig, port int32) {
	caPEM, err := os.ReadFile(cfg.certDir + "/ca.crt")
	if err != nil {
		logger.Error("读取 ca.crt 失败（证书未就绪，下轮重试）", zap.Error(err))
		return
	}
	desired := desiredWebhook(cfg, port, caPEM)

	cur, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, mwcName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx,
			&admissionregv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: mwcName, Labels: mwcLabels},
				Webhooks:   []admissionregv1.MutatingWebhook{desired},
			}, metav1.CreateOptions{})
		if err != nil {
			logger.Error("创建 MWC 失败（下轮重试）", zap.String("mwc", mwcName), zap.Error(err))
			return
		}
		logger.Info("MWC 已创建（意图 enabled）", zap.String("mwc", mwcName))
		return
	}
	if err != nil {
		logger.Error("查询 MWC 失败（下轮重试）", zap.String("mwc", mwcName), zap.Error(err))
		return
	}
	if !mwcDrifted(cur, desired) {
		return
	}
	// 以现有对象为底更新（保留 resourceVersion 与 apiserver 填充的默认字段）
	cur.Webhooks = []admissionregv1.MutatingWebhook{desired}
	if cur.Labels == nil {
		cur.Labels = map[string]string{}
	}
	for k, v := range mwcLabels {
		cur.Labels[k] = v
	}
	if _, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, cur, metav1.UpdateOptions{}); err != nil {
		logger.Error("修正 MWC 漂移失败（下轮重试）", zap.String("mwc", mwcName), zap.Error(err))
		return
	}
	logger.Info("MWC 漂移已修正（caBundle/Service/选择器）", zap.String("mwc", mwcName))
}

// desiredWebhook 期望的 webhook 条目（spec 与 backend3 URL 型 ensureMWC 同构）
func desiredWebhook(cfg rotatorConfig, port int32, caPEM []byte) admissionregv1.MutatingWebhook {
	svcName := envOr("ARTHAS_WEBHOOK_SERVICE_NAME", "msre-pilot")
	path := webhookHandlerPath
	ignore := admissionregv1.Ignore
	equivalent := admissionregv1.Equivalent
	none := admissionregv1.SideEffectClassNone
	timeout := int32(5)
	return admissionregv1.MutatingWebhook{
		Name: "msre-pilot.oneops.cn",
		ClientConfig: admissionregv1.WebhookClientConfig{
			// Service 型：apiserver → ClusterIP 必达（istio 同构路径）
			Service: &admissionregv1.ServiceReference{
				Namespace: cfg.namespace,
				Name:      svcName,
				Port:      &port,
				Path:      &path,
			},
			CABundle: caPEM,
		},
		Rules: []admissionregv1.RuleWithOperations{{
			Operations: []admissionregv1.OperationType{admissionregv1.Create},
			Rule: admissionregv1.Rule{
				APIGroups: []string{""}, APIVersions: []string{"v1"}, Resources: []string{"pods"},
			},
		}},
		// Pod 级触发：仅 label oneops-arthas-injection=enabled 的 Pod 进入注入逻辑
		ObjectSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{injectLabel: "enabled"},
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
	}
}

// mwcDrifted 比较关键字段（caBundle/Service 指向/failurePolicy/label 选择器），
// 未变化则跳过 Update，避免每轮 reconcile 写放大
func mwcDrifted(cur *admissionregv1.MutatingWebhookConfiguration, want admissionregv1.MutatingWebhook) bool {
	if len(cur.Webhooks) != 1 {
		return true
	}
	have := cur.Webhooks[0]
	svc, wsvc := have.ClientConfig.Service, want.ClientConfig.Service
	if svc == nil || wsvc == nil {
		return true
	}
	if svc.Namespace != wsvc.Namespace || svc.Name != wsvc.Name {
		return true
	}
	if svc.Port == nil || *svc.Port != *wsvc.Port || svc.Path == nil || *svc.Path != *wsvc.Path {
		return true
	}
	if !bytes.Equal(have.ClientConfig.CABundle, want.ClientConfig.CABundle) {
		return true
	}
	if have.FailurePolicy == nil || *have.FailurePolicy != *want.FailurePolicy {
		return true
	}
	if have.ObjectSelector == nil || have.ObjectSelector.MatchLabels[injectLabel] != "enabled" {
		return true
	}
	if cur.Labels["oneops.io/managed-by"] != mwcLabels["oneops.io/managed-by"] {
		return true
	}
	return false
}

// deleteMWC 删除 MWC（意图 disabled；NotFound 视为已收敛）
func deleteMWC(ctx context.Context, client *kubernetes.Clientset) {
	err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(ctx, mwcName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return
	}
	if err != nil {
		logger.Error("删除 MWC 失败（下轮重试）", zap.String("mwc", mwcName), zap.Error(err))
		return
	}
	logger.Info("MWC 已删除（意图 disabled）", zap.String("mwc", mwcName))
}
