package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

/**
 * 证书自愈（借鉴 OtelPilot internal/controller/cert.go，按单进程无 controller-runtime 精简）：
 *  - 首次：Secret 缺证书/占位非法 → 自动签发全套（CA+leaf）写回 Secret，替代手工 openssl/bootstrap Job
 *  - 巡检：12h 一轮（启动随机抖动 0-5min 降低多副本同触发），leaf 剩余 <30 天自动轮转
 *  - CA 复用：CA 剩余 ≥60 天且 ca.key 在 Secret 内 → 仅重签 leaf；否则重建 CA
 *  - 接管：非本体系证书（无 ca.key 的手工预置）即使仍在有效期也重建全套，证书完全由轮转器管理
 *  - 跟随：签发/轮转后即时更新 MWC caBundle（MWC 生命周期归 reconcile.go 控制器，
 *    此处仅加速证书轮转的 caBundle 收敛，幂等无冲突）
 *  - 生效：TLS GetCertificate 热加载，kubelet 同步卷后新握手即用新证书（无进程重启）
 *
 * 说明：CA 私钥（ca.key）存于集群 Secret（与 OtelPilot 同）——换取全自动轮转，不留手工证书路径。
 */

const (
	certRotateThreshold = 30 * 24 * time.Hour // leaf 剩余 <30 天 → 轮转
	caRotateThreshold   = 60 * 24 * time.Hour // CA 剩余 <60 天 → 重建 CA
	leafValidity        = 10 * 365 * 24 * time.Hour
	caValidity          = 10 * 365 * 24 * time.Hour
	mwcName             = "msre-pilot-inject"
)

// rotatorConfig 轮转器配置（env 推导）
type rotatorConfig struct {
	namespace  string
	secretName string
	certDir    string
	hosts      []string
}

// buildKubeClient 集群内 ServiceAccount；本地调试 fallback kubeconfig
func buildKubeClient() (*kubernetes.Clientset, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		return kubernetes.NewForConfig(cfg)
	}
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadCfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("无 in-cluster 凭据且读取 kubeconfig 失败: %w", err)
	}
	return kubernetes.NewForConfig(loadCfg)
}

// bootstrapCert 启动期证书保障：挂载目录已有合法证书直接返回；
// 否则同步修复 Secret（首次自动签发/轮转）并等待 kubelet 同步挂载卷（≤5min）。
func bootstrapCert(cfg rotatorConfig) error {
	if err := verifyCert(cfg.certDir, cfg.hosts); err == nil {
		return nil
	}
	client, err := buildKubeClient()
	if err != nil {
		return fmt.Errorf("挂载证书不可用且无法自动签发（无 K8s 凭据）: %w", err)
	}
	if _, err := ensureSecretCert(client, cfg); err != nil {
		return fmt.Errorf("证书签发失败: %w", err)
	}
	logger.Info("证书已签发至 Secret，等待 kubelet 同步挂载卷", zap.String("secret", cfg.secretName))
	if err := waitCertReady(cfg.certDir, cfg.hosts, 5*time.Minute); err != nil {
		return err
	}
	// 首次签发时 MWC 可能已存在旧 caBundle（手工证书期创建的），同步跟随
	if err := syncMWCCaBundle(client, cfg); err != nil {
		logger.Warn("MWC caBundle 同步失败（平台启用时会读最新 ca.crt）", zap.Error(err))
	}
	return nil
}

// startCertRotator 后台巡检循环（12h + 启动抖动）
func startCertRotator(cfg rotatorConfig) {
	client, err := buildKubeClient()
	if err != nil {
		logger.Warn("证书轮转器禁用（无 K8s 凭据，证书需手工维护）", zap.Error(err))
		return
	}
	go func() {
		jitter := time.Duration(randInt63n(int64(5 * time.Minute)))
		logger.Info("证书轮转器已启动", zap.Duration("firstCheckAfter", jitter),
			zap.Duration("interval", 12*time.Hour))
		time.Sleep(jitter)
		runRotateOnce(client, cfg)
		t := time.NewTicker(12 * time.Hour)
		defer t.Stop()
		for range t.C {
			runRotateOnce(client, cfg)
		}
	}()
}

func runRotateOnce(client *kubernetes.Clientset, cfg rotatorConfig) {
	rotated, err := ensureSecretCert(client, cfg)
	if err != nil {
		logger.Error("证书巡检失败（下轮重试）", zap.Error(err))
		return
	}
	if !rotated {
		return
	}
	// MWC caBundle 跟随（MWC 不存在则跳过，平台启用时会读 Secret 最新 ca.crt）
	if err := syncMWCCaBundle(client, cfg); err != nil {
		logger.Error("MWC caBundle 同步失败", zap.Error(err))
	}
	// TLS GetCertificate 热加载：kubelet 同步 Secret 卷（≤1min）后新握手自动用新证书，
	// 不再退出进程重启（重启方案会形成 Exit(0)→Back-off 循环）
	logger.Info("证书已轮转，热加载生效（无需重启）")
}

// ensureSecretCert 确保证书 Secret 合法：需轮转则签发写回（resourceVersion 乐观锁）。
// 返回是否发生了轮转。合法（剩余充足 + SAN 匹配）时不动（不覆盖手工预置证书）。
func ensureSecretCert(client *kubernetes.Clientset, cfg rotatorConfig) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secret, err := client.CoreV1().Secrets(cfg.namespace).Get(ctx, cfg.secretName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// Secret 丢失自愈：重建空壳走签发（与清单 data:{} 语义一致），
		// 避免"Secret 被删 → 永久 Fatal → CrashLoop"无出口循环
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: cfg.secretName, Namespace: cfg.namespace},
			Type:       corev1.SecretTypeOpaque,
		}
		err = nil
	}
	if err != nil {
		return false, err
	}

	caPEM, caKey, caOK := usableCA(secret)
	leafOK := usableLeaf(secret, cfg.hosts)
	if leafOK && caOK {
		return false, nil // 本体系证书健康，不动
	}
	// leaf 健康但 CA 不可复用（无 ca.key 等非本体系证书）→ 重建全套接管

	logger.Info("证书需要签发/轮转", zap.Bool("leafUsable", leafOK), zap.Bool("caUsable", caOK))
	caKeyPEMBytes := secret.Data["ca.key"] // CA 复用时保留原值
	if !caOK {
		newPEM, newKey, err := newCAPEM()
		if err != nil {
			return false, err
		}
		caPEM, caKey, caKeyPEMBytes = newPEM, newKey, caKeyPEM(newKey)
	}
	tlsCrt, tlsKey, err := newLeafPEM(caPEM, caKey, cfg.hosts)
	if err != nil {
		return false, err
	}

	// 乐观锁写回：Conflict 说明并发副本已轮转，重读校验
	secret.Data = map[string][]byte{
		"tls.crt": tlsCrt, "tls.key": tlsKey,
		"ca.crt": caPEM, "ca.key": caKeyPEMBytes,
	}
	if secret.ResourceVersion == "" {
		// 丢失重建路径：无 resourceVersion 走 Create
		_, err = client.CoreV1().Secrets(cfg.namespace).Create(ctx, secret, metav1.CreateOptions{})
		return err == nil, err
	}
	_, err = client.CoreV1().Secrets(cfg.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if apierrors.IsConflict(err) {
		fresh, gerr := client.CoreV1().Secrets(cfg.namespace).Get(ctx, cfg.secretName, metav1.GetOptions{})
		if gerr == nil && usableLeaf(fresh, cfg.hosts) {
			return false, nil // 另一副本已完成
		}
	}
	return true, err
}

// usableCA Secret 内 CA 可复用：ca.crt+ca.key 齐全、解析合法、剩余 ≥ 阈值
// （ca.key 为轮转器自签的 EC SEC1 格式，无手工 openssl 兼容路径）
func usableCA(secret *corev1.Secret) (caPEM []byte, caKey any, ok bool) {
	crt, key := secret.Data["ca.crt"], secret.Data["ca.key"]
	if len(crt) == 0 || len(key) == 0 {
		return nil, nil, false
	}
	block, _ := pem.Decode(crt)
	if block == nil {
		return nil, nil, false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil || !cert.IsCA || time.Until(cert.NotAfter) < caRotateThreshold {
		return nil, nil, false
	}
	if k, e := x509.ParseECPrivateKey(key); e == nil {
		return crt, k, true
	}
	return nil, nil, false
}

// usableLeaf leaf 合法：齐全、解析成功、剩余 ≥ 阈值、SAN 覆盖全部 hosts
func usableLeaf(secret *corev1.Secret, hosts []string) bool {
	crt, key := secret.Data["tls.crt"], secret.Data["tls.key"]
	if len(crt) == 0 || len(key) == 0 {
		return false
	}
	block, _ := pem.Decode(crt)
	if block == nil {
		return false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}
	if time.Until(cert.NotAfter) < certRotateThreshold {
		return false
	}
	for _, h := range hosts {
		if !coversHost(cert, h) {
			return false
		}
	}
	return true
}

// syncMWCCaBundle 把 Secret 的 ca.crt 写入 MWC 所有 webhook 条目的 caBundle（漂移自愈）
func syncMWCCaBundle(client *kubernetes.Clientset, cfg rotatorConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	secret, err := client.CoreV1().Secrets(cfg.namespace).Get(ctx, cfg.secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	caPEM := secret.Data["ca.crt"]
	if len(caPEM) == 0 {
		return fmt.Errorf("Secret 缺少 ca.crt")
	}
	want := base64.StdEncoding.EncodeToString(caPEM)

	mwc, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, mwcName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		logger.Info("MWC 未创建，跳过 caBundle 同步（reconciler 会在意图 enabled 时创建）")
		return nil
	}
	if err != nil {
		return err
	}
	changed := false
	for i := range mwc.Webhooks {
		if string(mwc.Webhooks[i].ClientConfig.CABundle) != want {
			mwc.Webhooks[i].ClientConfig.CABundle = []byte(want)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	_, err = client.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, mwc, metav1.UpdateOptions{})
	if err == nil {
		logger.Info("MWC caBundle 已同步（证书轮转跟随）", zap.String("mwc", mwcName))
	}
	return err
}

// waitCertReady 等待挂载目录出现合法证书（首次签发后 kubelet 同步 Secret 卷，≤1min）
func waitCertReady(certDir string, hosts []string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := verifyCert(certDir, hosts); err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("等待证书就绪超时（%s），检查轮转器日志与 Secret 卷同步", timeout)
}

// ========== 自签（ECDSA P256，与 backend3 generateSelfSignedCert 同构） ==========

func newCAPEM() ([]byte, *ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber:          randomSerial(),
		Subject:               pkix.Name{CommonName: "msre-pilot-ca", Organization: []string{"OneOps"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(caValidity),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	return pemEncode("CERTIFICATE", der), key, nil
}

func newLeafPEM(caPEM []byte, caKey any, hosts []string) ([]byte, []byte, error) {
	caBlock, _ := pem.Decode(caPEM)
	if caBlock == nil {
		return nil, nil, fmt.Errorf("CA PEM 非法")
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	dnsNames, ipAddrs := []string{"localhost"}, []net.IP{net.ParseIP("127.0.0.1")}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			ipAddrs = append(ipAddrs, ip)
		} else {
			dnsNames = append(dnsNames, h)
		}
	}
	tmpl := &x509.Certificate{
		SerialNumber: randomSerial(),
		Subject:      pkix.Name{CommonName: "msre-pilot", Organization: []string{"OneOps"}},
		DNSNames:     dnsNames,
		IPAddresses:  ipAddrs,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(leafValidity),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	return pemEncode("CERTIFICATE", der), pemEncode("EC PRIVATE KEY", keyDER), nil
}

func caKeyPEM(key *ecdsa.PrivateKey) []byte {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil
	}
	return pemEncode("EC PRIVATE KEY", der)
}

func pemEncode(blockType string, der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
}

func randomSerial() *big.Int {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return big.NewInt(time.Now().UnixNano())
	}
	return n
}

func randInt63n(n int64) int64 {
	v, err := rand.Int(rand.Reader, big.NewInt(n))
	if err != nil {
		return 0
	}
	return v.Int64()
}
