package middleware

import (
	"courier-service/internal/logging"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求ID
		requestID := c.GetHeader(RequestIDHeader)

		// 如果没有请求ID，生成一个新的
		if requestID == "" {
			requestID = logging.GenerateRequestID()
		}

		// 设置到上下文中
		c.Set(RequestIDKey, requestID)

		// 设置响应头
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID 从上下文获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if rid, ok := requestID.(string); ok {
			return rid
		}
	}
	return ""
}
