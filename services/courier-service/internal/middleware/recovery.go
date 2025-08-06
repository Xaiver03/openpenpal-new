package middleware

import (
	"courier-service/internal/errors"
	"courier-service/internal/logging"

	"github.com/gin-gonic/gin"
)

// Recovery 恢复中间件配置
type RecoveryConfig struct {
	Logger       logging.Logger
	DebugMode    bool
	SkipPaths    []string
	RecoveryFunc func(c *gin.Context, err interface{})
}

// Recovery 创建恢复中间件
func Recovery(config RecoveryConfig) gin.HandlerFunc {
	if config.Logger == nil {
		config.Logger = logging.GetDefaultLogger()
	}

	panicRecovery := errors.NewPanicRecovery(config.Logger)

	return func(c *gin.Context) {
		// 检查是否跳过此路径
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		defer func() {
			if err := panicRecovery.Recover(); err != nil {
				// 获取请求信息
				requestID := c.GetString("request_id")
				userID := c.GetString("user_id")
				
				// 创建带上下文的错误
				courierErr := err.(*errors.CourierServiceError)
				courierErr.WithRequestID(requestID).WithUserID(userID)
				
				// 记录详细错误信息
				config.Logger.Error("Request panic recovered",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"query", c.Request.URL.RawQuery,
					"request_id", requestID,
					"user_id", userID,
					"error", err,
				)

				// 如果有自定义恢复函数，调用它
				if config.RecoveryFunc != nil {
					config.RecoveryFunc(c, err)
					return
				}

				// 默认恢复处理
				c.Header("Content-Type", "application/json")
				
				response := gin.H{
					"code":    courierErr.Code,
					"message": courierErr.Message,
					"success": false,
				}

				// 在调试模式下包含更多信息
				if config.DebugMode {
					response["debug"] = gin.H{
						"request_id":  requestID,
						"stack_trace": courierErr.StackTrace,
						"context":     courierErr.Context,
					}
				}

				c.JSON(courierErr.HTTPStatus, response)
				c.Abort()
			}
		}()

		c.Next()
	}
}

// DefaultRecovery 默认恢复中间件
func DefaultRecovery() gin.HandlerFunc {
	logger := logging.GetDefaultLogger()
	
	return Recovery(RecoveryConfig{
		Logger:    logger,
		DebugMode: false,
	})
}

// DebugRecovery 调试模式恢复中间件
func DebugRecovery() gin.HandlerFunc {
	logger := logging.GetDefaultLogger()
	
	return Recovery(RecoveryConfig{
		Logger:    logger,
		DebugMode: true,
	})
}