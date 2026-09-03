// Package controller 实现组件自愈能力：Webhook TLS 证书自动签发/轮转、
// MutatingWebhookConfiguration 校正，以及探针配置热同步。
package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"otelpilot/internal/certs"
)

const (
	// certRotateThreshold 证书剩余有效期低于该阈值时自动轮转。
	certRotateThreshold = 30 * 24 * time.Hour
	// certResyncInterval 定期巡检周期（兜底到期检查）。
	certResyncInterval = 12 * time.Hour

	// WebhookEntryName webhook 条目名（需为合法域名格式）。
	WebhookEntryName = "mpod.otel-pilot.io"
	// WebhookPath webhook 服务路径。
	WebhookPath = "/mutate--v1-pod"
)

// CertOptions 证书管理参数。
type CertOptions struct {
	Namespace     string
	ServiceName   string
	ClusterDomain string
	SecretName    string
	WebhookName   string
}

// CertReconciler 证书与 Webhook 配置自愈控制器（仅 Leader 执行写入）。
type CertReconciler struct {
	client.Client
	Options CertOptions
	Logger  logr.Logger
}

// DNSNames 返回 webhook 服务端证书需要覆盖的 SAN。
func (o CertOptions) DNSNames() []string {
	return []string{
		o.ServiceName,
		fmt.Sprintf("%s.%s", o.ServiceName, o.Namespace),
		fmt.Sprintf("%s.%s.svc", o.ServiceName, o.Namespace),
		fmt.Sprintf("%s.%s.svc.%s", o.ServiceName, o.Namespace, o.ClusterDomain),
		"localhost",
	}
}

// Reconcile 确保证书有效且 Webhook 配置指向正确 CA。
func (r *CertReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	changed, err := EnsureWebhookCert(ctx, r.Client, r.Options)
	if err != nil {
		return ctrl.Result{}, err
	}
	if changed {
		r.Logger.Info("webhook cert or configuration updated")
	}
	// 周期巡检，兜底处理证书到期。
	return ctrl.Result{RequeueAfter: certResyncInterval}, nil
}

// SetupCertController 注册证书控制器：监听证书 Secret 与 MutatingWebhookConfiguration 变化。
func SetupCertController(mgr ctrl.Manager, opts CertOptions, logger logr.Logger) error {
	r := &CertReconciler{Client: mgr.GetClient(), Options: opts, Logger: logger}
	secretPredicate := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetName() == opts.SecretName && o.GetNamespace() == opts.Namespace
	})
	webhookPredicate := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetName() == opts.WebhookName
	})
	return ctrl.NewControllerManagedBy(mgr).
		Named("otel-pilot-cert").
		For(&corev1.Secret{}, builder.WithPredicates(secretPredicate)).
		Watches(&admissionregistrationv1.MutatingWebhookConfiguration{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: opts.WebhookName}}}
			}),
			builder.WithPredicates(webhookPredicate)).
		Complete(r)
}

// EnsureWebhookCert 确保证书 Secret 存在且有效（缺失/临期/SAN 变化时重签），
// 并将 CA 同步至 MutatingWebhookConfiguration（不存在时自愈重建）。
// 该函数同时服务于常驻控制器与安装期 bootstrap Job。
func EnsureWebhookCert(ctx context.Context, c client.Client, opts CertOptions) (bool, error) {
	changed := false

	// ---------- 1. 证书 Secret ----------
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: opts.Namespace, Name: opts.SecretName},
	}
	err := c.Get(ctx, client.ObjectKeyFromObject(secret), secret)
	switch {
	case apierrors.IsNotFound(err):
		secret, err = buildCertSecret(nil, nil, nil, opts)
		if err != nil {
			return false, err
		}
		if err := c.Create(ctx, secret); err != nil {
			return false, fmt.Errorf("create cert secret: %w", err)
		}
		changed = true
	case err != nil:
		return false, fmt.Errorf("get cert secret: %w", err)
	default:
		needRotate, err := certNeedsRotate(secret, opts)
		if err != nil {
			return false, err
		}
		if needRotate {
			// CA 有效则复用，仅轮转叶子证书。
			caPEM, caKeyPEM := secret.Data["ca.crt"], secret.Data["ca.key"]
			if !caValid(caPEM, caKeyPEM) {
				caPEM, caKeyPEM, err = certs.NewCA()
				if err != nil {
					return false, err
				}
			}
			secret, err = buildCertSecret(secret, caPEM, caKeyPEM, opts)
			if err != nil {
				return false, err
			}
			if err := c.Update(ctx, secret); err != nil {
				return false, fmt.Errorf("rotate cert secret: %w", err)
			}
			changed = true
		}
	}
	caPEM := secret.Data["ca.crt"]

	// ---------- 2. MutatingWebhookConfiguration ----------
	caBundle := []byte(base64.StdEncoding.EncodeToString(caPEM))
	mwc := &admissionregistrationv1.MutatingWebhookConfiguration{}
	err = c.Get(ctx, types.NamespacedName{Name: opts.WebhookName}, mwc)
	switch {
	case apierrors.IsNotFound(err):
		// 自愈：配置被删除时按内置模板重建。
		mwc = desiredWebhookConfig(opts, caBundle)
		if err := c.Create(ctx, mwc); err != nil && !apierrors.IsAlreadyExists(err) {
			return false, fmt.Errorf("create mutatingwebhookconfiguration: %w", err)
		}
		changed = true
	case err != nil:
		return false, fmt.Errorf("get mutatingwebhookconfiguration: %w", err)
	default:
		desired := desiredWebhookConfig(opts, caBundle)
		idx := -1
		for i := range mwc.Webhooks {
			if mwc.Webhooks[i].Name == WebhookEntryName {
				idx = i
				break
			}
		}
		if idx == -1 {
			mwc.Webhooks = append(mwc.Webhooks, desired.Webhooks[0])
		} else if webhookDrifted(&mwc.Webhooks[idx], &desired.Webhooks[0]) {
			mwc.Webhooks[idx] = desired.Webhooks[0]
		} else {
			return changed, nil
		}
		if err := c.Update(ctx, mwc); err != nil {
			return false, fmt.Errorf("update mutatingwebhookconfiguration: %w", err)
		}
		changed = true
	}
	return changed, nil
}

// certNeedsRotate 判断现有证书是否缺失、临期或 SAN 不匹配。
func certNeedsRotate(secret *corev1.Secret, opts CertOptions) (bool, error) {
	leaf, err := certs.ParseCertificate(secret.Data["tls.crt"])
	if err != nil {
		return true, nil
	}
	if _, err := certs.ParsePrivateKey(secret.Data["tls.key"]); err != nil {
		return true, nil
	}
	if time.Until(leaf.NotAfter) < certRotateThreshold {
		return true, nil
	}
	want := opts.DNSNames()
	if len(leaf.DNSNames) != len(want) {
		return true, nil
	}
	got := map[string]struct{}{}
	for _, d := range leaf.DNSNames {
		got[d] = struct{}{}
	}
	for _, d := range want {
		if _, ok := got[d]; !ok {
			return true, nil
		}
	}
	return false, nil
}

func caValid(caPEM, caKeyPEM []byte) bool {
	ca, err := certs.ParseCertificate(caPEM)
	if err != nil || !ca.IsCA || time.Until(ca.NotAfter) < 2*certRotateThreshold {
		return false
	}
	_, err = certs.ParsePrivateKey(caKeyPEM)
	return err == nil
}

// buildCertSecret 生成证书 Secret（CA 无效时重建 CA，总是签发新叶子证书）；
// base 非 nil 时在其基础上轮转（保留 ResourceVersion 用于 Update）。
func buildCertSecret(base *corev1.Secret, caPEM, caKeyPEM []byte, opts CertOptions) (*corev1.Secret, error) {
	if !caValid(caPEM, caKeyPEM) {
		var err error
		caPEM, caKeyPEM, err = certs.NewCA()
		if err != nil {
			return nil, fmt.Errorf("generate ca: %w", err)
		}
	}
	leafPEM, leafKeyPEM, err := certs.NewLeafCert(caPEM, caKeyPEM, opts.ServiceName+"."+opts.Namespace, opts.DNSNames())
	if err != nil {
		return nil, fmt.Errorf("generate leaf cert: %w", err)
	}
	labels := map[string]string{"app.kubernetes.io/managed-by": "otel-pilot"}
	resourceVersion := ""
	if base != nil {
		resourceVersion = base.ResourceVersion
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       opts.Namespace,
			Name:            opts.SecretName,
			Labels:          labels,
			ResourceVersion: resourceVersion,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": leafPEM,
			"tls.key": leafKeyPEM,
			"ca.crt":  caPEM,
			"ca.key":  caKeyPEM,
		},
	}, nil
}

// desiredWebhookConfig 内置的期望 Webhook 配置模板（failurePolicy=Ignore，异常不阻塞 Pod 创建）。
func desiredWebhookConfig(opts CertOptions, caBundle []byte) *admissionregistrationv1.MutatingWebhookConfiguration {
	ignore := admissionregistrationv1.Ignore
	none := admissionregistrationv1.SideEffectClassNone
	ifNeeded := admissionregistrationv1.IfNeededReinvocationPolicy
	namespaced := admissionregistrationv1.NamespacedScope
	return &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.WebhookName,
			Labels: map[string]string{"app.kubernetes.io/managed-by": "otel-pilot"},
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name:                    WebhookEntryName,
			AdmissionReviewVersions: []string{"v1"},
			SideEffects:             &none,
			FailurePolicy:           &ignore,
			ReinvocationPolicy:      &ifNeeded,
			TimeoutSeconds:          ptrTo(int32(5)),
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Name:      opts.ServiceName,
					Namespace: opts.Namespace,
					Port:      ptrTo(int32(443)),
					Path:      ptrTo(WebhookPath),
				},
				CABundle: caBundle,
			},
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "kubernetes.io/metadata.name",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{"kube-system", "kube-public", "kube-node-lease", opts.Namespace},
				}},
			},
			Rules: []admissionregistrationv1.RuleWithOperations{{
				Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Resources:   []string{"pods"},
					Scope:       &namespaced,
				},
			}},
		}},
	}
}

// webhookDrifted 仅比较需要自愈的字段（clientConfig / 策略 / 规则）。
func webhookDrifted(current, desired *admissionregistrationv1.MutatingWebhook) bool {
	c, d := current.ClientConfig.Service, desired.ClientConfig.Service
	if c == nil || d == nil {
		return true
	}
	if c.Name != d.Name || c.Namespace != d.Namespace || val(c.Path) != val(d.Path) || deref(c.Port) != deref(d.Port) {
		return true
	}
	if string(current.ClientConfig.CABundle) != string(desired.ClientConfig.CABundle) {
		return true
	}
	if deref(current.FailurePolicy) != deref(desired.FailurePolicy) ||
		len(current.Rules) != len(desired.Rules) {
		return true
	}
	return false
}

func ptrTo[T any](v T) *T { return &v }

func deref[T comparable](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func val[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
