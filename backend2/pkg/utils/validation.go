package utils

import (
	"strings"
)

// SanitizeLikeInput 清理 LIKE 查询的输入，防止通配符注入
func SanitizeLikeInput(input string) string {
	// 转义特殊字符
	input = strings.ReplaceAll(input, "\\", "\\\\")
	input = strings.ReplaceAll(input, "%", "\\%")
	input = strings.ReplaceAll(input, "_", "\\_")
	return input
}
