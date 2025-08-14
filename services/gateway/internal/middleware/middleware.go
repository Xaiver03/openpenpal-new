package middleware

import (
	"api-gateway/internal/models"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Trace-ID, X-User-ID, X-Client-Version, X-CSRF-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24小时

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Logger 日志中间件
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录请求日志
		duration := time.Since(startTime)

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("Gateway request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("trace_id", c.GetHeader("X-Trace-ID")),
			zap.String("user_id", GetUserID(c)),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}

// Recovery 恢复中间件
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Gateway panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
				)

				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Code:      http.StatusInternalServerError,
					Message:   "Internal server error",
					Timestamp: time.Now(),
					Path:      c.Request.URL.Path,
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// RateLimiter 限流中间件
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
	rate     rate.Limit
	burst    int
}

var rateLimiterInstance *RateLimiter
var rateLimiterOnce sync.Once

// GetRateLimiter 获取限流器实例
func getRateLimiter(requestsPerMinute int) *RateLimiter {
	rateLimiterOnce.Do(func() {
		rateLimiterInstance = &RateLimiter{
			limiters: make(map[string]*rate.Limiter),
			rate:     rate.Limit(requestsPerMinute) / 60, // 转换为每秒
			burst:    requestsPerMinute / 6,              // 突发流量为平均值的1/6
		}
	})
	return rateLimiterInstance
}

// NewRateLimiter 创建限流中间件
func NewRateLimiter(requestsPerMinute int) gin.HandlerFunc {
	limiter := getRateLimiter(requestsPerMinute)

	return func(c *gin.Context) {
		// 使用客户端IP作为限流key
		key := c.ClientIP()

		// 如果是已认证用户，使用用户ID
		if userID := GetUserID(c); userID != "" {
			key = "user:" + userID
		}

		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Code:      http.StatusTooManyRequests,
				Message:   "Rate limit exceeded",
				Details:   fmt.Sprintf("Maximum %d requests per minute", requestsPerMinute),
				Timestamp: time.Now(),
				Path:      c.Request.URL.Path,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow 检查是否允许请求
func (rl *RateLimiter) allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter.Allow()
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// 检查是否超时
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, models.ErrorResponse{
				Code:      http.StatusRequestTimeout,
				Message:   "Request timeout",
				Details:   fmt.Sprintf("Request exceeded %v timeout", timeout),
				Timestamp: time.Now(),
				Path:      c.Request.URL.Path,
			})
			c.Abort()
		}
	}
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取或生成请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置响应头
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}

// Metrics 监控指标中间件
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		// 记录指标（这里会在monitor包中实现）
		duration := time.Since(startTime)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// 更新Prometheus指标（在monitor包中实现）
		recordMetrics(method, path, status, duration)
	}
}

// recordMetrics 记录监控指标（占位符，实际实现在monitor包中）
func recordMetrics(method, path string, status int, duration time.Duration) {
	// 这里会调用monitor包的相关函数
	// monitor.RecordHTTPRequest(method, path, status, duration)
}

// Security 安全头中间件
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req-%d-%s", time.Now().UnixNano(), randomString(8))
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
