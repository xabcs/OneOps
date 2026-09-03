// Package certs 提供 Webhook TLS 证书的自签与轮转能力，
// 仅依赖 Go 标准库，避免引入 cert-manager 等外部组件。
package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"
)

const (
	// CAValidity CA 自签证书有效期（10 年）。
	CAValidity = 10 * 365 * 24 * time.Hour
	// LeafValidity 服务端叶子证书有效期（1 年）。
	LeafValidity = 365 * 24 * time.Hour

	rsaKeyBits = 2048
)

// NewCA 生成自签名 CA 证书与私钥（PEM 编码）。
func NewCA() (caPEM []byte, caKeyPEM []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ca key: %w", err)
	}
	tpl := &x509.Certificate{
		SerialNumber:          randomSerial(),
		Subject:               pkix.Name{CommonName: "otel-pilot-ca", Organization: []string{"otel-pilot"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(CAValidity),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("create ca cert: %w", err)
	}
	return pemEncode("CERTIFICATE", der), pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key)), nil
}

// NewLeafCert 使用给定 CA 签发服务端证书（PEM 编码），SAN 覆盖 dnsNames 与 127.0.0.1。
func NewLeafCert(caPEM, caKeyPEM []byte, commonName string, dnsNames []string) (certPEM []byte, keyPEM []byte, err error) {
	caCert, err := ParseCertificate(caPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parse ca cert: %w", err)
	}
	caKey, err := ParsePrivateKey(caKeyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parse ca key: %w", err)
	}
	key, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		return nil, nil, fmt.Errorf("generate leaf key: %w", err)
	}
	tpl := &x509.Certificate{
		SerialNumber: randomSerial(),
		Subject:      pkix.Name{CommonName: commonName, Organization: []string{"otel-pilot"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(LeafValidity),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     dnsNames,
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create leaf cert: %w", err)
	}
	return pemEncode("CERTIFICATE", der), pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key)), nil
}

// ParseCertificate 解析 PEM 编码的证书。
func ParseCertificate(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("no valid PEM certificate found")
	}
	return x509.ParseCertificate(block.Bytes)
}

// ParsePrivateKey 解析 PEM 编码的 RSA 私钥。
func ParsePrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no valid PEM private key found")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	key, ok := keyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}
	return key, nil
}

func randomSerial() *big.Int {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return big.NewInt(time.Now().UnixNano())
	}
	return serial
}

func pemEncode(typ string, der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der})
}
