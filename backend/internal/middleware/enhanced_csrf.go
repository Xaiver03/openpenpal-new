package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"openpenpal-backend/internal/utils"
)

// EnhancedCSRF 增强版CSRF防护中间件
type EnhancedCSRF struct {
	TokenLength    int
	CookieName     string
	HeaderName     string
	CookieSecure   bool
	CookieSameSite http.SameSite
	SkipPaths      []string
	HighRiskPaths  []string // 高风险路径需要额外验证
}

// NewEnhancedCSRF 创建增强版CSRF中间件
func NewEnhancedCSRF() *EnhancedCSRF {
	return &EnhancedCSRF{
		TokenLength:    32,
		CookieName:     "csrf-token",
		HeaderName:     "X-CSRF-Token",
		CookieSecure:   true,
		CookieSameSite: http.SameSiteStrictMode,
		SkipPaths: []string{
			"/health",
			"/api/v1/auth/csrf",
			"/api/v1/letters/public",
			"/api/v1/ai/personas", // AI功能允许无认证访问
		},
		HighRiskPaths: []string{
			"/api/v1/auth/change-password",
			"/api/v1/users/me/delete",
			"/api/v1/admin/",
		},
	}
}

// generateCSRFToken 生成安全的CSRF Token
func (c *EnhancedCSRF) generateCSRFToken() (string, error) {
	bytes := make([]byte, c.TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// isPathSkipped 检查路径是否跳过CSRF验证
func (c *EnhancedCSRF) isPathSkipped(path string) bool {
	for _, skipPath := range c.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// isHighRiskPath 检查是否为高风险路径
func (c *EnhancedCSRF) isHighRiskPath(path string) bool {
	for _, riskPath := range c.HighRiskPaths {
		if strings.HasPrefix(path, riskPath) {
			return true
		}
	}
	return false
}

// validateOrigin 验证请求来源
func (c *EnhancedCSRF) validateOrigin(ctx *gin.Context) bool {
	origin := ctx.GetHeader("Origin")
	referer := ctx.GetHeader("Referer")
	
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://localhost:3000",
		"http://127.0.0.1:3000",
		"https://127.0.0.1:3000",
	}
	
	// 检查Origin header
	if origin != "" {
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
	}
	
	// 检查Referer header
	if referer != "" {
		for _, allowed := range allowedOrigins {
			if strings.HasPrefix(referer, allowed) {
				return true
			}
		}
	}
	
	return false
}

// Middleware CSRF中间件主函数
func (c *EnhancedCSRF) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		
		// 跳过GET请求和指定路径
		if method == "GET" || c.isPathSkipped(path) {
			ctx.Next()
			return
		}
		
		// 检查自定义安全header（简单防护）
		customHeader := ctx.GetHeader("X-Requested-With")
		openpenpalHeader := ctx.GetHeader("X-OpenPenPal-Auth")
		
		hasCustomHeader := customHeader == "XMLHttpRequest" || openpenpalHeader == "frontend-client"
		
		// Origin验证
		originValid := c.validateOrigin(ctx)
		
		// 对于JWT认证的API，如果有自定义header且Origin验证通过，可以放行
		authHeader := ctx.GetHeader("Authorization")
		hasJWT := strings.HasPrefix(authHeader, "Bearer ")
		
		if hasJWT && hasCustomHeader && originValid {
			// JWT + 自定义Header + Origin验证 = 基础安全级别
			if !c.isHighRiskPath(path) {
				ctx.Next()
				return
			}
		}
		
		// 高风险操作需要完整CSRF验证
		if c.isHighRiskPath(path) || (!hasCustomHeader && !originValid) {
			// 获取Cookie中的token
			cookie, err := ctx.Cookie(c.CookieName)
			if err != nil {
				utils.BadRequestResponse(ctx, "CSRF token missing in cookie", nil)
				ctx.Abort()
				return
			}
			
			// 获取Header中的token
			headerToken := ctx.GetHeader(c.HeaderName)
			if headerToken == "" {
				utils.BadRequestResponse(ctx, "CSRF token missing in header", nil)
				ctx.Abort()
				return
			}
			
			// 使用constant-time比较防止时序攻击
			if subtle.ConstantTimeCompare([]byte(cookie), []byte(headerToken)) != 1 {
				utils.UnauthorizedResponse(ctx, "CSRF token mismatch")
				ctx.Abort()
				return
			}
		}
		
		ctx.Next()
	}
}

// GenerateToken 生成CSRF Token的处理函数
func (c *EnhancedCSRF) GenerateToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := c.generateCSRFToken()
		if err != nil {
			utils.InternalServerErrorResponse(ctx, "Failed to generate CSRF token", err)
			return
		}
		
		// 在Cookie中设置token
		ctx.SetSameSite(c.CookieSameSite)
		ctx.SetCookie(
			c.CookieName,
			token,
			int((24 * time.Hour).Seconds()), // 24小时过期
			"/",
			"",
			c.CookieSecure,
			true, // HttpOnly
		)
		
		// 在响应中也返回token供客户端使用
		utils.SuccessResponse(ctx, http.StatusOK, "CSRF token generated", gin.H{
			"token": token,
		})
	}
}

// SmartCSRF 智能CSRF中间件 - 根据风险级别选择验证策略
func SmartCSRF() gin.HandlerFunc {
	enhanced := NewEnhancedCSRF()
	
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		
		// GET请求和公开API跳过
		if method == "GET" || enhanced.isPathSkipped(path) {
			ctx.Next()
			return
		}
		
		// 检查是否有JWT认证
		authHeader := ctx.GetHeader("Authorization") 
		hasJWT := strings.HasPrefix(authHeader, "Bearer ")
		
		// 检查Origin/Referer
		originValid := enhanced.validateOrigin(ctx)
		
		// 检查自定义安全header
		customHeader := ctx.GetHeader("X-OpenPenPal-Auth")
		hasCustomHeader := customHeader == "frontend-client"
		
		// 低风险操作：JWT + Origin + 自定义Header 即可
		if hasJWT && originValid && hasCustomHeader && !enhanced.isHighRiskPath(path) {
			ctx.Header("X-CSRF-Protection", "jwt-origin-header")
			ctx.Next()
			return
		}
		
		// 高风险操作：需要完整CSRF验证
		if enhanced.isHighRiskPath(path) {
			ctx.Header("X-CSRF-Protection", "full-token")
			// 执行完整CSRF验证逻辑
			enhanced.Middleware()(ctx)
			return
		}
		
		// 默认：如果没有足够的安全措施，拒绝请求
		utils.UnauthorizedResponse(ctx, "Insufficient security headers for this operation")
		ctx.Abort()
	}
}