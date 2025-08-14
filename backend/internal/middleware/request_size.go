package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestSizeLimitMiddleware 请求大小限制中间件
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查Content-Length头
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success":  false,
				"error":    "请求体过大",
				"message":  "请求体大小超过限制",
				"max_size": maxSize,
			})
			c.Abort()
			return
		}

		// 限制请求体读取大小
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		// 处理可能的错误
		c.Next()

		// 检查是否因为请求体过大而出错
		if c.Errors != nil {
			for _, err := range c.Errors {
				if err.Error() == "http: request body too large" {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{
						"success":  false,
						"error":    "请求体过大",
						"message":  "上传的数据超过最大限制",
						"max_size": maxSize,
					})
					return
				}
			}
		}
	}
}

// 预定义的大小限制
const (
	// DefaultMaxRequestSize 默认最大请求大小 - 10MB
	DefaultMaxRequestSize = 10 * 1024 * 1024

	// MaxUploadSize 文件上传最大大小 - 50MB
	MaxUploadSize = 50 * 1024 * 1024

	// MaxJSONRequestSize JSON请求最大大小 - 1MB
	MaxJSONRequestSize = 1 * 1024 * 1024

	// MaxLetterContentSize 信件内容最大大小 - 100KB
	MaxLetterContentSize = 100 * 1024
)
