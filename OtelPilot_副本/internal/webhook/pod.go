// Package webhook 实现同步注入的 Mutating 准入 Webhook。
package webhook

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	corev1 "k8s.io/api/core/v1"

	"otelpilot/internal/config"
	"otelpilot/internal/inject"
)

// PodMutator 拦截 Pod 创建事件，按自研标签判定是否注入 OpenTelemetry-JavaAgent。
type PodMutator struct {
	Decoder        admission.Decoder
	Store          *config.Store
	Logger         logr.Logger
	SkipNamespaces map[string]struct{}
}

// Handle 处理 AdmissionReview：命中标签则返回注入 Patch，否则直接放行。
func (m *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}
	if err := m.Decoder.Decode(req, pod); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// 系统与组件自身命名空间不注入。
	if _, skip := m.SkipNamespaces[pod.Namespace]; skip {
		return admission.Allowed("")
	}

	// 标签开关判定：非 on 一律放行且不做任何修改。
	if !inject.ShouldInject(pod) {
		return admission.Allowed("otel-pilot: injection disabled")
	}

	snap := m.Store.Get()
	ops, err := inject.BuildPatches(pod, snap)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	m.Logger.Info("inject opentelemetry javaagent",
		"namespace", pod.Namespace,
		"pod", pod.Name,
		"agentVersion", pod.Labels[inject.LabelAgentVersion],
		"defaultAgentVersion", snap.AgentVersion,
	)
	return admission.Patched("otel-pilot: opentelemetry javaagent injected", ops...)
}
