package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// verifyCert 校验证书目录：tls.crt/tls.key 存在，且 tls.crt 的 SAN 覆盖全部 hosts
// （独立服务不自签：集群内部署证书由清单预置 Secret，本地调试用 deploy.yaml 头部 openssl 命令生成）
func verifyCert(dir string, hosts []string) error {
	certPath := filepath.Join(dir, "tls.crt")
	keyPath := filepath.Join(dir, "tls.key")
	for _, p := range []string{certPath, keyPath} {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("证书文件缺失 %s: %w", p, err)
		}
	}
	pemBytes, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return fmt.Errorf("%s 不是合法 PEM", certPath)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	for _, h := range hosts {
		if !coversHost(cert, h) {
			return fmt.Errorf("证书 SAN 未覆盖 %s（当前 SAN: DNS=%v IP=%v），请重签", h, cert.DNSNames, cert.IPAddresses)
		}
	}
	return nil
}

// loadTLSCert 从目录加载 tls.crt/tls.key（TLS GetCertificate 热加载入口）。
// admission 握手频率极低，每次读文件的代价可忽略；换 mtime 缓存属于过度设计
func loadTLSCert(dir string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(filepath.Join(dir, "tls.crt"), filepath.Join(dir, "tls.key"))
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// coversHost 域名 → DNSNames，IP → IPAddresses
func coversHost(cert *x509.Certificate, host string) bool {
	if ip := net.ParseIP(host); ip != nil {
		for _, cip := range cert.IPAddresses {
			if cip.Equal(ip) {
				return true
			}
		}
		return false
	}
	for _, d := range cert.DNSNames {
		if strings.EqualFold(d, host) {
			return true
		}
	}
	return false
}
