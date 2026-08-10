package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
)

// JWT密钥安全性工具

// GenerateJWTSecret 生成安全的JWT密钥（32字节随机字符串）
func GenerateJWTSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成随机密钥失败: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateJWTSecret 验证JWT密钥强度
// 返回: (是否有效, 错误消息)
func ValidateJWTSecret(secret string) (bool, string) {
	// 检查是否为空
	if secret == "" {
		return false, "JWT密钥不能为空"
	}

	// 检查长度（至少32字节）
	if len(secret) < 32 {
		return false, fmt.Sprintf("JWT密钥长度不足: 当前%d字节，至少需要32字节", len(secret))
	}

	// 检查是否使用默认值或弱密钥
	weakSecrets := []string{
		"secret",
		"change-in-production",
		"123456",
		"password",
		"jwtsecret",
		"oneops",
		"admin",
		"test",
		"demo",
	}

	lowerSecret := strings.ToLower(secret)
	for _, weak := range weakSecrets {
		if lowerSecret == weak {
			return false, fmt.Sprintf("JWT密钥使用了弱密钥或默认值: %s", weak)
		}
		if strings.Contains(lowerSecret, weak) {
			return false, fmt.Sprintf("JWT密钥包含弱密钥: %s", weak)
		}
	}

	// 检查熵值（简单检查：不同字符的数量）
	if !hasEnoughEntropy(secret) {
		return false, "JWT密钥熵值不足，建议包含大小写字母、数字和特殊字符"
	}

	return true, ""
}

// hasEnoughEntropy 检查密钥是否具有足够的熵值
func hasEnoughEntropy(secret string) bool {
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, c := range secret {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			// 特殊字符
			hasSpecial = true
		}
	}

	// 至少满足3种字符类型
	score := 0
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}

	return score >= 3
}

// HashJWTSecret 对JWT密钥进行哈希（用于验证）
func HashJWTSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return fmt.Sprintf("%x", hash)
}

// GenerateSecureToken 生成安全令牌（用于临时访问）
func GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成令牌失败: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("长度必须大于0")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("生成随机字符串失败: %w", err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

// GenerateAPIKey 生成API密钥
func GenerateAPIKey() (string, error) {
	// 格式: oneops_32位随机字符
	randomPart, err := GenerateRandomString(32)
	if err != nil {
		return "", err
	}
	return "oneops_" + randomPart, nil
}

// MaskSecret 脱敏密钥（用于日志）
func MaskSecret(secret string, showPrefix int) string {
	if len(secret) <= showPrefix {
		return "***"
	}
	if showPrefix <= 0 {
		return "*****"
	}
	return secret[:showPrefix] + "*****"
}
