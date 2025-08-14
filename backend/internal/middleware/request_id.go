package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否已有请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 生成新的请求ID
			requestID = generateRequestID()
		}

		// 设置到上下文和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// MetricsMiddleware 性能监控中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// 添加性能头
		c.Header("X-Response-Time", fmt.Sprintf("%v", duration))

		// 记录慢请求
		if duration > time.Second {
			requestID := c.GetString("request_id")
			userID := c.GetString("user_id")

			fmt.Printf("[SLOW_REQUEST] RequestID=%s UserID=%s Method=%s Path=%s Status=%d Duration=%v\n",
				requestID, userID, method, path, statusCode, duration)
		}

		// 记录错误请求
		if statusCode >= 500 {
			requestID := c.GetString("request_id")
			fmt.Printf("[ERROR_REQUEST] RequestID=%s Method=%s Path=%s Status=%d Duration=%v Error=%v\n",
				requestID, method, path, statusCode, duration, c.Errors.String())
		}
	}
}

// generateRequestID 生成唯一的请求ID
func generateRequestID() string {
	// 时间戳部分
	timestamp := time.Now().UnixNano()

	// 随机部分
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("%x-%s", timestamp, randomHex)
}
