package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	encryptionKey     []byte
	encryptionKeyOnce sync.Once
)

// InitEncryption 初始化加密密钥
func InitEncryption() error {
	var initErr error
	encryptionKeyOnce.Do(func() {
		keyStr := os.Getenv("ENCRYPTION_KEY")
		if keyStr == "" {
			// 如果未设置环境变量，尝试从配置文件读取
			keyStr = os.Getenv("DATA_ENCRYPTION_KEY")
			if keyStr == "" {
				initErr = errors.New("ENCRYPTION_KEY环境变量未设置，请设置加密密钥以保护敏感数据")
				return
			}
		}

		key, err := base64.URLEncoding.DecodeString(keyStr)
		if err != nil {
			initErr = fmt.Errorf("解析加密密钥失败: %w", err)
			return
		}

		if len(key) != 32 {
			initErr = fmt.Errorf("加密密钥长度错误，期望32字节，实际%d字节", len(key))
			return
		}

		encryptionKey = key
	})

	return initErr
}

// InitEncryptionWithKey 使用指定的密钥初始化加密模块
func InitEncryptionWithKey(keyStr string) error {
	var initErr error
	encryptionKeyOnce.Do(func() {
		if keyStr == "" {
			initErr = errors.New("加密密钥不能为空")
			return
		}

		key, err := base64.URLEncoding.DecodeString(keyStr)
		if err != nil {
			initErr = fmt.Errorf("解析加密密钥失败: %w", err)
			return
		}

		if len(key) != 32 {
			initErr = fmt.Errorf("加密密钥长度错误，期望32字节，实际%d字节", len(key))
			return
		}

		encryptionKey = key
	})

	return initErr
}

// MustInitEncryption 初始化加密密钥（panic on error）
func MustInitEncryption() {
	if err := InitEncryption(); err != nil {
		panic(fmt.Sprintf("初始化加密模块失败: %v", err))
	}
}

// GenerateEncryptionKey 生成加密密钥（用于首次部署）
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("生成加密密钥失败: %w", err)
	}
	return base64.URLEncoding.EncodeToString(key), nil
}

// EnsureEncryptionInitialized 确保加密模块已初始化
func EnsureEncryptionInitialized() error {
	if encryptionKey == nil {
		return InitEncryption()
	}
	return nil
}

// EncryptData AES-256-GCM加密
func EncryptData(plaintext []byte) (string, error) {
	if err := EnsureEncryptionInitialized(); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密数据（GCM模式自动添加认证标签）
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 返回 base64 编码的结果
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// EncryptString 加密字符串
func EncryptString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	return EncryptData([]byte(plaintext))
}

// DecryptData AES-256-GCM解密
func DecryptData(ciphertext string) ([]byte, error) {
	if err := EnsureEncryptionInitialized(); err != nil {
		return nil, err
	}

	data, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %w", err)
	}

	if len(data) < 12 { // GCM nonce 最小12字节
		return nil, errors.New("密文长度不足")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM模式失败: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("密文长度不足")
	}

	// 分离nonce和密文
	nonce, cipherData := data[:nonceSize], data[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, fmt.Errorf("解密失败: %w", err)
	}

	return plaintext, nil
}

// DecryptString 解密字符串
func DecryptString(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	decrypted, err := DecryptData(ciphertext)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// IsEncrypted 检查数据是否已加密（简单启发式检查）
func IsEncrypted(data string) bool {
	// 检查是否为Base64格式且长度合理
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		// 如果不是标准Base64，可能是未加密的原始数据
		return false
	}

	// 加密后的数据格式: base64(nonce + ciphertext + tag)
	// Base64解码后至少29字节（12字节nonce + 至少1字节密文 + 16字节tag）
	if len(decoded) >= 29 {
		return true
	}

	return false
}

// HashSecret 用于存储不可逆哈希（如token签名）
func HashSecret(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
