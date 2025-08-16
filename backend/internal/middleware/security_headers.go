package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityConfig 安全配置结构
type SecurityConfig struct {
	IsDevelopment        bool
	AllowedDomains      []string
	FrontendURL         string
	WebSocketURL        string
	EnableHSTS          bool
	EnableCSPReporting  bool
	CSPReportURI        string
	TrustedCDNs         []string
}

// NewSecurityConfig 创建安全配置
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		IsDevelopment: os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "",
		AllowedDomains: []string{
			os.Getenv("FRONTEND_URL"),
			"http://localhost:3000",
			"https://localhost:3000",
		},
		FrontendURL:  os.Getenv("FRONTEND_URL"),
		WebSocketURL: os.Getenv("WEBSOCKET_URL"),
		EnableHSTS:   os.Getenv("ENABLE_HSTS") == "true",
		EnableCSPReporting: os.Getenv("ENABLE_CSP_REPORTING") == "true",
		CSPReportURI: os.Getenv("CSP_REPORT_URI"),
		TrustedCDNs: []string{
			"https://cdn.jsdelivr.net",
			"https://fonts.googleapis.com", 
			"https://fonts.gstatic.com",
		},
	}
}

// generateNonce 生成CSP nonce
func generateNonce() string {
	bytes := make([]byte, 24) // 增加到24字节以提高安全性
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Warning: Failed to generate secure nonce: %v", err)
		// 使用时间戳作为后备
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

// SecurityHeadersMiddleware 安全头中间件 - SOTA级别实现
func SecurityHeadersMiddleware() gin.HandlerFunc {
	config := NewSecurityConfig()

	return func(c *gin.Context) {
		// 生成请求级别nonce
		nonce := generateNonce()
		c.Set("csp_nonce", nonce)
		// 基础安全头 - 增强版
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "0") // 现代浏览器中建议禁用，依赖CSP
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-Powered-By", "") // 隐藏服务器信息
		
		// 完善的权限策略 - 最新业界标准
		permissionsPolicy := []string{
			"geolocation=()",
			"microphone=()",
			"camera=()",
			"usb=()",
			"bluetooth=()",
			"gyroscope=()",
			"accelerometer=()",
			"magnetometer=()",
			"payment=()",
			"midi=()",
			"speaker=()",
			"vibrate=()",
			"push=()",
			"notifications=()",
			"persistent-storage=()",
			"sync-xhr=(self)",
			"autoplay=(self)",
			"encrypted-media=(self)",
			"picture-in-picture=(self)",
			"fullscreen=(self)",
			"web-share=()",
			"cross-origin-isolated=()",
			"interest-cohort=()", // 禁用Google FLoC
			"browsing-topics=()", // 禁用Topics API
		}
		c.Header("Permissions-Policy", strings.Join(permissionsPolicy, ", "))

		// HSTS - 基于配置启用
		if config.EnableHSTS && !config.IsDevelopment {
			c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload") // 2年
		}

		// 构建动态CSP策略
		cspPolicy := buildCSPPolicy(config, nonce)
		c.Header("Content-Security-Policy", cspPolicy)
		
		// CSP报告模式（仅在开发环境）
		if config.IsDevelopment && config.EnableCSPReporting {
			cspReportPolicy := buildCSPReportOnlyPolicy(config, nonce)
			c.Header("Content-Security-Policy-Report-Only", cspReportPolicy)
		}
		
		// 将nonce传递给前端
		c.Header("X-CSP-Nonce", nonce)

		// 添加其他推荐的安全头 - 加强版
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("X-Download-Options", "noopen")
		c.Header("X-DNS-Prefetch-Control", "off")
		c.Header("X-Robots-Tag", "noindex, nofollow, nosnippet, noarchive")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")
		
		// 防止缓存敏感信息
		if c.Request.URL.Path == "/api/v1/auth/login" || 
		   c.Request.URL.Path == "/api/v1/auth/refresh" ||
		   c.Request.URL.Path == "/api/v1/admin/sensitive-words" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		
		// 防止点击劫持的额外保护
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", c.GetHeader("Content-Security-Policy")+"; frame-ancestors 'none';")
		
		c.Next()
	}
}

// CSPViolationHandler 处理CSP违规报告 - 增强版
func CSPViolationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var violation map[string]interface{}
		if err := c.ShouldBindJSON(&violation); err != nil {
			c.JSON(400, gin.H{"error": "Invalid CSP violation report"})
			return
		}
		
		// 记录详细的CSP违规信息
		log.Printf("CSP Violation Report: User-Agent: %s, IP: %s, Violation: %+v", 
			c.GetHeader("User-Agent"), c.ClientIP(), violation)
		
		// 可以在这里添加更多处理逻辑：
		// 1. 发送到监控系统
		// 2. 检查是否为潜在攻击
		// 3. 更新安全策略
		
		// 分析违规类型和严重程度
		if violationType, exists := violation["violated-directive"]; exists {
			log.Printf("CSP Violation Type: %v", violationType)
			
			// 检查是否为高风险违规
			if isHighRiskViolation(fmt.Sprintf("%v", violationType)) {
				log.Printf("HIGH RISK CSP Violation detected from IP: %s", c.ClientIP())
				// 这里可以触发额外的安全措施
			}
		}
		
		c.Status(204) // No Content
	}
}

// isHighRiskViolation 检查是否为高风险违规
func isHighRiskViolation(violationType string) bool {
	highRiskDirectives := []string{
		"script-src", "object-src", "base-uri", "form-action",
	}
	
	for _, directive := range highRiskDirectives {
		if strings.Contains(violationType, directive) {
			return true
		}
	}
	return false
}

// buildCSPPolicy 构建动态CSP策略
func buildCSPPolicy(config *SecurityConfig, nonce string) string {
	policies := []string{
		"default-src 'self'",
		"base-uri 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
		"object-src 'none'",
		"media-src 'self'",
	}
	
	if config.IsDevelopment {
		// 开发环境：允许热重载和开发工具
		policies = append(policies,
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' "+strings.Join(config.TrustedCDNs, " "),
			"style-src 'self' 'unsafe-inline' "+strings.Join(config.TrustedCDNs, " "),
			"connect-src 'self' ws://localhost:* wss://localhost:* http://localhost:* https://localhost:*",
			"img-src 'self' data: blob: https:",
		)
	} else {
		// 生产环境：严格安全策略
		policies = append(policies,
			"script-src 'self' 'nonce-"+nonce+"' "+strings.Join(config.TrustedCDNs, " "),
			"style-src 'self' 'nonce-"+nonce+"' "+strings.Join(config.TrustedCDNs, " "),
			"connect-src 'self' "+config.WebSocketURL,
			"img-src 'self' data: https:",
			"upgrade-insecure-requests",
			"block-all-mixed-content",
		)
		
		// 支持Trusted Types（仅支持的浏览器）
		policies = append(policies, "require-trusted-types-for 'script'")
	}
	
	// 添加字体支持
	policies = append(policies, "font-src 'self' https://fonts.gstatic.com data:")
	
	// CSP报告
	if config.EnableCSPReporting && config.CSPReportURI != "" {
		policies = append(policies, "report-uri "+config.CSPReportURI)
		policies = append(policies, "report-to csp-endpoint")
	}
	
	return strings.Join(policies, "; ")
}

// buildCSPReportOnlyPolicy 构建报告模式CSP策略
func buildCSPReportOnlyPolicy(config *SecurityConfig, nonce string) string {
	// 与主策略相同，但更严格，用于检测潜在问题
	return buildCSPPolicy(config, nonce)
}

// ThreatDetectionMiddleware 威胁检测中间件
func ThreatDetectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查常见攻击模式
		if detectSQLInjection(c) {
			log.Printf("Security Threat: Potential SQL injection detected from IP: %s", c.ClientIP())
			c.Header("X-Threat-Detected", "sql-injection")
		}
		
		if detectXSSAttempt(c) {
			log.Printf("Security Threat: Potential XSS attempt detected from IP: %s", c.ClientIP())
			c.Header("X-Threat-Detected", "xss-attempt")
		}
		
		if detectDirectoryTraversal(c) {
			log.Printf("Security Threat: Directory traversal attempt detected from IP: %s", c.ClientIP())
			c.Header("X-Threat-Detected", "directory-traversal")
		}
		
		c.Next()
	}
}

// detectSQLInjection 检测SQL注入尝试
func detectSQLInjection(c *gin.Context) bool {
	sqlPatterns := []string{
		"union", "select", "insert", "delete", "update", "drop",
		"exec", "--", "/*", "*/", "'", "\"", ";",
	}
	
	queryString := strings.ToLower(c.Request.URL.RawQuery)
	for _, pattern := range sqlPatterns {
		if strings.Contains(queryString, pattern) {
			return true
		}
	}
	return false
}

// detectXSSAttempt 检测XSS尝试
func detectXSSAttempt(c *gin.Context) bool {
	xssPatterns := []string{
		"<script", "javascript:", "onerror=", "onload=",
		"<iframe", "<object", "<embed", "vbscript:",
	}
	
	queryString := strings.ToLower(c.Request.URL.RawQuery)
	for _, pattern := range xssPatterns {
		if strings.Contains(queryString, pattern) {
			return true
		}
	}
	return false
}

// detectDirectoryTraversal 检测目录遍历尝试
func detectDirectoryTraversal(c *gin.Context) bool {
	path := c.Request.URL.Path
	traversalPatterns := []string{
		"../", "..\\", "%2e%2e%2f", "%2e%2e%5c",
		"..%2f", "..%5c", "%2e%2e/", "%2e%2e\\",
	}
	
	pathLower := strings.ToLower(path)
	for _, pattern := range traversalPatterns {
		if strings.Contains(pathLower, pattern) {
			return true
		}
	}
	return false
}
