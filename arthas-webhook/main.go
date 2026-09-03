package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 包级共享（inject.go / cert.go 同包使用）
var logger = zap.Must(zap.NewProduction())

func main() {
	defer logger.Sync()

	port := envOr("ARTHAS_WEBHOOK_PORT", "9443")
	certDir := envOr("ARTHAS_WEBHOOK_CERT_DIR", "/etc/webhook-tls")

	// Service DNS（apiserver 按此名校验证书 SAN）
	svcNS := envOr("ARTHAS_WEBHOOK_SERVICE_NS", "test")
	svcName := envOr("ARTHAS_WEBHOOK_SERVICE_NAME", "msre-pilot")
	hosts := []string{
		fmt.Sprintf("%s.%s.svc", svcName, svcNS),
		fmt.Sprintf("%s.%s.svc.cluster.local", svcName, svcNS),
		svcName,
	}
	if extra := os.Getenv("ARTHAS_WEBHOOK_TLS_DNS"); extra != "" {
		hosts = append(hosts, strings.Split(extra, ",")...)
	}

	// 证书保障：挂载卷有合法证书直接用（手工 openssl 预置）；缺失/非法/SAN 不覆盖
	// 则由轮转器自动签发写回 Secret 并等 kubelet 同步（首次部署零手工证书步骤）
	rotCfg := rotatorConfig{
		namespace:  svcNS,
		secretName: envOr("ARTHAS_WEBHOOK_TLS_SECRET", "msre-pilot-tls"),
		certDir:    certDir,
		hosts:      hosts,
	}
	if err := bootstrapCert(rotCfg); err != nil {
		logger.Fatal("webhook 证书不可用（自动签发失败）",
			zap.String("certDir", certDir), zap.Strings("requiredSAN", hosts), zap.Error(err))
	}

	if tunnelWSConfig() == "" {
		logger.Warn("ARTHAS_TUNNEL_WS 未配置，命中注入 label 的 Pod 将被拒绝创建",
			zap.String("hint", "治理页保存注入参数会同步 ConfigMap，本服务 watch 热加载即时生效"))
	}
	logger.Info("arthas-webhook 独立服务启动",
		zap.String("port", port), zap.String("certDir", certDir),
		zap.String("tunnelWS", tunnelWSConfig()), zap.String("initImage", initImageConfig()),
		zap.Strings("certSAN", hosts),
	)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.POST(webhookHandlerPath, handleAdmission)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			// 证书热加载：每次握手从挂载卷读取证书。轮转器写回 Secret 后，
			// kubelet 同步卷（≤1min）即自动生效，无需退出进程重启加载
			// （重启方案在 K8s 中会形成 Exit(0)→Back-off 循环）
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				return loadTLSCert(certDir)
			},
		},
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			logger.Fatal("webhook TLS server 启动失败", zap.Error(err))
		}
	}()
	logger.Info("arthas-webhook TLS server 已启动", zap.String("addr", srv.Addr))

	// 证书轮转巡检：12h 一轮，剩余 <30 天自动轮转并跟随更新 MWC caBundle
	startCertRotator(rotCfg)

	// MWC 生命周期控制器：治理页意图（ConfigMap WEBHOOK_ENABLED）→ 准入规则收敛
	portNum := int32(9443)
	if p, err := strconv.Atoi(port); err == nil && p > 0 && p < 65536 {
		portNum = int32(p)
	}
	startReconciler(rotCfg, portNum)

	// 优雅退出（SIGTERM 滚动更新时排空在途 admission 请求）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("收到退出信号，正在关闭")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("webhook server 关闭异常", zap.Error(err))
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
