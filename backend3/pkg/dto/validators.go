package dto

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// 自定义验证器
var (
	// 主机名验证正则（只允许字母、数字、连字符）
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
)

// RegisterCustomValidators 注册自定义验证器
func RegisterCustomValidators(v *validator.Validate) error {
	// 注册 hostname 验证
	if err := v.RegisterValidation("hostname", validateHostname); err != nil {
		return err
	}

	// 注册 ip 验证
	if err := v.RegisterValidation("ip", validateIP); err != nil {
		return err
	}

	return nil
}

// validateHostname 验证主机名格式
func validateHostname(fl validator.FieldLevel) bool {
	field := fl.Field().String()

	if field == "" {
		return true // 允许为空（配合 required 标签使用）
	}

	return hostnameRegex.MatchString(field)
}

// validateIP 验证 IP 地址格式
func validateIP(fl validator.FieldLevel) bool {
	field := fl.Field().String()

	if field == "" {
		return true // 允许为空
	}

	// 简单的 IP 验证（支持 IPv4）
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if !ipRegex.MatchString(field) {
		return false
	}

	// 验证每个段是否在 0-255 范围内
	parts := regexp.MustCompile(`\.`).Split(field, -1)
	for _, part := range parts {
		num := 0
		for _, c := range part {
			num = num*10 + int(c-'0')
		}
		if num > 255 {
			return false
		}
	}

	return true
}
