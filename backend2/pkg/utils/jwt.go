package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("oneops-jwt-secret-key-2024") // 默认密钥，应该被环境变量覆盖

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
