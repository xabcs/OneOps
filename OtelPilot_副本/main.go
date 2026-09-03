// OtelPilot：自研 K8s OpenTelemetry-JavaAgent 自动注入组件。
// 架构：MutatingWebhook（同步注入）+ Controller（证书轮转 / 配置同步 / 自愈）。
package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	zaplogs "sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	crwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"otelpilot/internal/config"
	"otelpilot/internal/controller"
	podwebhook "otelpilot/internal/webhook"
)

const defaultNamespace = "otel-pilot"

func main() {
	var (
		mode           string
		metricsAddr    string
		probeAddr      string
		leaderElect    bool
		webhookPort    int
		webhookCertDir string

		namespace      string
		serviceName    string
		clusterDomain  string
		certSecretName string
		webhookCfgName string
		configMapName  string
		authSecretName string
	)
	flag.StringVar(&mode, "mode", "run", "运行模式：run（常驻组件）| cert-bootstrap（安装期证书初始化后退出）")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8383", "Metrics 监听地址")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "健康检查监听地址")
	flag.BoolVar(&leaderElect, "leader-elect", true, "多副本 Leader 选举（控制器写入仅 Leader 执行）")
	flag.IntVar(&webhookPort, "webhook-port", 9443, "Webhook TLS 监听端口")
	flag.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/k8s-webhook-server/serving-certs", "Webhook 证书目录（tls.crt/tls.key）")

	flag.StringVar(&namespace, "namespace", os.Getenv("POD_NAMESPACE"), "组件运行命名空间")
	flag.StringVar(&serviceName, "service-name", "otel-pilot", "Webhook Service 名称")
	flag.StringVar(&clusterDomain, "cluster-domain", "cluster.local", "集群 DNS 域")
	flag.StringVar(&certSecretName, "cert-secret-name", "otel-pilot-cert", "证书 Secret 名称")
	flag.StringVar(&webhookCfgName, "webhook-config-name", "otel-pilot", "MutatingWebhookConfiguration 名称")
	flag.StringVar(&configMapName, "configmap-name", "otel-pilot-config", "探针配置 ConfigMap 名称")
	flag.StringVar(&authSecretName, "auth-secret-name", "otel-pilot-auth", "授权 Secret 名称（key: headers，不存在则不注入授权头）")

	opts := zaplogs.Options{Development: false}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zaplogs.New(zaplogs.UseFlagOptions(&opts)))

	ns := resolveNamespace(namespace)
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)

	certOpts := controller.CertOptions{
		Namespace:     ns,
		ServiceName:   serviceName,
		ClusterDomain: clusterDomain,
		SecretName:    certSecretName,
		WebhookName:   webhookCfgName,
	}

	if mode == "cert-bootstrap" {
		runCertBootstrap(scheme, certOpts)
		return
	}

	store := config.NewStore()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		Metrics:                 metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          leaderElect,
		LeaderElectionID:        "otel-pilot-leader-election",
		LeaderElectionNamespace: ns,
		// 命名空间级缓存：Secret/ConfigMap 仅本命名空间，RBAC 只需 Role。
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{ns: {}},
		},
		WebhookServer: crwebhook.NewServer(crwebhook.Options{
			Port:    webhookPort,
			CertDir: webhookCertDir,
		}),
	})
	if err != nil {
		setupLog().Error(err, "unable to start manager")
		os.Exit(1)
	}

	// MutatingWebhook：Pod 创建时按标签注入。
	mutator := &podwebhook.PodMutator{
		Decoder: admission.NewDecoder(scheme),
		Store:   store,
		Logger:  ctrl.Log.WithName("webhook").WithName("pod"),
		SkipNamespaces: map[string]struct{}{
			ns: {}, "kube-system": {}, "kube-public": {}, "kube-node-lease": {},
		},
	}
	mgr.GetWebhookServer().Register(controller.WebhookPath, &admission.Webhook{Handler: mutator})

	// 配置热同步：所有副本各自运行。
	if err := mgr.Add(controller.NewConfigSyncer(
		mgr.GetAPIReader(), store, ns, configMapName, authSecretName,
		ctrl.Log.WithName("controller").WithName("configsync"),
	)); err != nil {
		setupLog().Error(err, "unable to register config syncer")
		os.Exit(1)
	}

	// 证书自愈控制器：仅 Leader 写入。
	if err := controller.SetupCertController(mgr, certOpts, ctrl.Log.WithName("controller").WithName("cert")); err != nil {
		setupLog().Error(err, "unable to create cert controller")
		os.Exit(1)
	}

	_ = mgr.AddHealthzCheck("healthz", healthz.Ping)
	_ = mgr.AddReadyzCheck("readyz", healthz.Ping)

	setupLog().Info("starting otel-pilot",
		"namespace", ns, "webhookPort", webhookPort, "leaderElect", leaderElect)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog().Error(err, "manager exited with error")
		os.Exit(1)
	}
}

// runCertBootstrap 安装期一次性签发证书并同步 Webhook CA 后退出，解决 Webhook 服务依赖证书的启动顺序问题。
func runCertBootstrap(scheme *runtime.Scheme, certOpts controller.CertOptions) {
	cfg := ctrl.GetConfigOrDie()
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		setupLog().Error(err, "bootstrap: create client")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	changed, err := controller.EnsureWebhookCert(ctx, c, certOpts)
	if err != nil {
		setupLog().Error(err, "bootstrap: ensure webhook cert")
		os.Exit(1)
	}
	setupLog().Info("cert bootstrap completed", "changed", changed)
}

func resolveNamespace(ns string) string {
	if ns = strings.TrimSpace(ns); ns != "" {
		return ns
	}
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if v := strings.TrimSpace(string(data)); v != "" {
			return v
		}
	}
	return defaultNamespace
}

func setupLog() logr.Logger {
	return ctrl.Log.WithName("setup")
}
