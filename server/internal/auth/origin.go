package auth

import (
	"regexp"
)

// 允許的 Origin 規則
var allowedOriginPatterns = []*regexp.Regexp{
	// Cloudflare Tunnel: https://random-name.trycloudflare.com
	regexp.MustCompile(`^https://[a-z0-9-]+\.trycloudflare\.com$`),

	// 生產環境: https://*.ganhua.wang
	regexp.MustCompile(`^https://[a-z0-9-]+\.ganhua\.wang$`),

	// 本地開發: http://localhost:任意埠號
	regexp.MustCompile(`^http://localhost:\d+$`),

	// 本地開發: http://127.0.0.1:任意埠號
	regexp.MustCompile(`^http://127\.0\.0\.1:\d+$`),
}

// ValidateOrigin 驗證 Origin 是否在白名單中
func ValidateOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	for _, pattern := range allowedOriginPatterns {
		if pattern.MatchString(origin) {
			return true
		}
	}

	return false
}

// AddAllowedOriginPattern 動態新增允許的 Origin 規則
func AddAllowedOriginPattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	allowedOriginPatterns = append(allowedOriginPatterns, re)
	return nil
}
