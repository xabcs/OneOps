package utils

import (
	"regexp"
	"strings"
)

// UserAgentInfo 用户代理信息
type UserAgentInfo struct {
	Browser string
	OS      string
}

// ParseUserAgent 解析用户代理字符串
func ParseUserAgent(userAgent string) UserAgentInfo {
	info := UserAgentInfo{
		Browser: "Unknown",
		OS:      "Unknown",
	}

	// 解析操作系统
	osPatterns := map[string]string{
		"Windows":       `Windows NT (\d+\.\d+)`,
		"Macintosh":     `Macintosh`,
		"Linux":         `Linux`,
		"Android":       `Android (\d+\.\d+)`,
		"iOS":           `iPhone|iPad|iPod`,
		"Chrome OS":     `CrOS`,
	}

	for osName, pattern := range osPatterns {
		if regexp.MustCompile(pattern).MatchString(userAgent) {
			info.OS = osName
			break
		}
	}

	// 解析浏览器
	browserPatterns := map[string]string{
		"Chrome":     `Chrome/(\d+\.\d+)`,
		"Firefox":    `Firefox/(\d+\.\d+)`,
		"Safari":     `Safari/(\d+\.\d+)`,
		"Edge":       `Edge/(\d+\.\d+)`,
		"Opera":      `Opera/(\d+\.\d+)`,
		"IE":         `MSIE (\d+\.\d+)`,
		"Chrome Mobile": `CriOS/(\d+\.\d+)`,
	}

	for browserName, pattern := range browserPatterns {
		if regexp.MustCompile(pattern).MatchString(userAgent) {
			info.Browser = browserName
			break
		}
	}

	// 特殊处理：Safari 在 Chrome 之后检查，因为 Chrome 也包含 Safari 字符串
	if strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome") {
		info.Browser = "Safari"
	}

	return info
}
