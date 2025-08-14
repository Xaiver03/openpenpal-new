package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"os"

	"github.com/gin-gonic/gin"
)

// generateNonce 生成CSP nonce
func generateNonce() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.StdEncoding.EncodeToString(bytes)
}

// SecurityHeadersMiddleware 安全头中间件 - SOTA级别实现
func SecurityHeadersMiddleware() gin.HandlerFunc {
	// 判断是否为开发环境
	isDev := os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == ""

	return func(c *gin.Context) {
		// 基础安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HSTS - 仅在生产环境启用
		if !isDev {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// CSP策略
		if isDev {
			// 开发环境：较宽松的策略，支持热重载等开发工具
			c.Header("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; "+
					"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
					"font-src 'self' https://fonts.gstatic.com data:; "+
					"img-src 'self' data: https: blob:; "+
					"connect-src 'self' ws://localhost:* wss://* http://localhost:*; "+
					"media-src 'self'; "+
					"object-src 'none'; "+
					"frame-ancestors 'none'; "+
					"base-uri 'self'; "+
					"form-action 'self';")
		} else {
			// 生产环境：严格的CSP策略，使用nonce
			nonce := generateNonce()
			c.Set("csp_nonce", nonce)

			c.Header("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'nonce-"+nonce+"' https://cdn.jsdelivr.net; "+
					"style-src 'self' 'nonce-"+nonce+"' https://fonts.googleapis.com; "+
					"font-src 'self' https://fonts.gstatic.com data:; "+
					"img-src 'self' data: https:; "+
					"connect-src 'self' wss://*; "+
					"media-src 'self'; "+
					"object-src 'none'; "+
					"frame-ancestors 'none'; "+
					"base-uri 'self'; "+
					"form-action 'self'; "+
					"upgrade-insecure-requests;")

			// 将nonce传递给模板引擎（如果使用）
			c.Header("X-CSP-Nonce", nonce)
		}

		// 添加其他推荐的安全头
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("X-Download-Options", "noopen")
		c.Header("X-DNS-Prefetch-Control", "off")

		c.Next()
	}
}

// CORSMiddleware is already defined in auth.go
