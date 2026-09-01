package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrJWTSecretNotSet 未设置 JWT 密钥（须通过 SetJWTSecret 或 JWT_SECRET 环境变量配置）
var ErrJWTSecretNotSet = errors.New("JWT secret not set, refusing to sign/verify with empty key")

// Claims JWT 声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// jwtSecret 默认为空：绝不内置弱默认密钥——否则任何绕过主入口校验的路径
// （测试二进制、新增 cmd 入口）都会用公开的默认密钥签发/接受 token，属伪造凭证风险
var jwtSecret []byte

// SetJWTSecret 设置 JWT 密钥
func SetJWTSecret(secret string) {
	if secret == "" {
		panic("JWT secret cannot be empty")
	}
	if len(secret) < 32 {
		panic("JWT secret must be at least 32 characters for security")
	}
	jwtSecret = []byte(secret)
}

// init 初始化 JWT 密钥，优先使用环境变量
func init() {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		SetJWTSecret(secret)
	}
}

// GenerateToken 生成 JWT token
func GenerateToken(userID uint, username string, expireHours int) (string, error) {
	if len(jwtSecret) == 0 {
		return "", ErrJWTSecretNotSet
	}
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expireHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "oneops",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析 JWT token
func ParseToken(tokenString string) (*Claims, error) {
	if len(jwtSecret) == 0 {
		return nil, ErrJWTSecretNotSet
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
