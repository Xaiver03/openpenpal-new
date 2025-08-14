package middleware

import (
	"courier-service/internal/errors"
	"courier-service/internal/logging"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerConfig 错误处理中间件配置
type ErrorHandlerConfig struct {
	Logger            logging.Logger
	EnableStackTrace  bool
	EnableErrorDetail bool
	CustomHandlers    map[errors.ErrorCode]func(*gin.Context, *errors.CourierServiceError)
}

// ErrorHandler 错误处理中间件
func ErrorHandler(config ErrorHandlerConfig) gin.HandlerFunc {
	if config.Logger == nil {
		config.Logger = logging.GetDefaultLogger()
	}

	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) == 0 {
			return
		}

		// 处理最后一个错误
		err := c.Errors.Last().Err

		// 获取请求上下文信息
		requestID := c.GetString("request_id")
		userID := c.GetString("user_id")

		// 转换为自定义错误类型
		var courierErr *errors.CourierServiceError
		if !errors.As(err, &courierErr) {
			// 如果不是自定义错误，包装它
			courierErr = errors.Wrap(err, errors.CodeInternalError, "Internal server error", errors.TypeSystem)
		}

		// 添加上下文信息
		courierErr.WithRequestID(requestID).WithUserID(userID)

		// 记录错误日志
		config.Logger.Error("Request error",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", courierErr.HTTPStatus,
			"error_code", courierErr.Code,
			"error_type", courierErr.Type,
			"request_id", requestID,
			"user_id", userID,
			"error", err,
		)

		// 检查是否有自定义处理器
		if handler, exists := config.CustomHandlers[courierErr.Code]; exists {
			handler(c, courierErr)
			return
		}

		// 构建响应
		response := gin.H{
			"code":      courierErr.Code,
			"message":   courierErr.Message,
			"success":   false,
			"timestamp": time.Now().Unix(),
		}

		// 添加请求ID（用于追踪）
		if requestID != "" {
			response["request_id"] = requestID
		}

		// 在开发环境下添加更多调试信息
		if config.EnableErrorDetail {
			response["details"] = gin.H{
				"error_type": courierErr.Type,
				"retryable":  courierErr.Retryable,
				"context":    courierErr.Context,
			}

			if config.EnableStackTrace && courierErr.StackTrace != "" {
				response["stack_trace"] = courierErr.StackTrace
			}
		}

		// 设置响应头
		c.Header("Content-Type", "application/json")
		c.Header("X-Error-Code", string(courierErr.Code))
		if requestID != "" {
			c.Header("X-Request-ID", requestID)
		}

		// 发送响应
		c.JSON(courierErr.HTTPStatus, response)
		c.Abort()
	}
}

// DefaultErrorHandler 默认错误处理中间件
func DefaultErrorHandler() gin.HandlerFunc {
	return ErrorHandler(ErrorHandlerConfig{
		Logger:            logging.GetDefaultLogger(),
		EnableStackTrace:  false,
		EnableErrorDetail: false,
	})
}

// DebugErrorHandler 调试模式错误处理中间件
func DebugErrorHandler() gin.HandlerFunc {
	return ErrorHandler(ErrorHandlerConfig{
		Logger:            logging.GetDefaultLogger(),
		EnableStackTrace:  true,
		EnableErrorDetail: true,
	})
}

// NotFoundHandler 404处理器
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := errors.NewError(errors.CodeNotFound, "Resource not found", errors.TypePermanent)
		c.Error(err)
	}
}

// MethodNotAllowedHandler 405处理器
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := errors.NewError(errors.CodeInvalidRequest, "Method not allowed", errors.TypePermanent)
		c.Error(err)
	}
}

// TimeoutHandler 超时处理器
func TimeoutHandler(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置请求超时
		if timeout > 0 {
			c.Request = c.Request.WithContext(
				c.Request.Context(),
			)

			// 启动超时检查
			timeoutChan := time.After(timeout)
			doneChan := make(chan bool, 1)

			go func() {
				c.Next()
				doneChan <- true
			}()

			select {
			case <-timeoutChan:
				err := errors.NewError(errors.CodeServiceUnavailable, "Request timeout", errors.TypeTemporary)
				c.Error(err)
				c.Abort()
			case <-doneChan:
				// 请求正常完成
			}
		} else {
			c.Next()
		}
	}
}

// RateLimitHandler 限流处理器
func RateLimitHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以集成具体的限流逻辑
		// 例如使用Redis进行限流控制

		c.Next()
	}
}

// ValidationErrorHandler 验证错误处理器
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 特殊处理验证错误
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				if ginErr.Type == gin.ErrorTypeBind {
					err := errors.NewValidationError("request", ginErr.Error())
					c.Error(err)
					return
				}
			}
		}
	}
}

// CORSErrorHandler CORS错误处理器
func CORSErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查CORS配置
		if origin != "" && !isAllowedOrigin(origin) {
			err := errors.NewError(errors.CodeForbidden, "CORS policy violation", errors.TypePermanent)
			c.Error(err)
			c.Abort()
			return
		}

		c.Next()
	}
}

// isAllowedOrigin 检查是否允许的来源
func isAllowedOrigin(origin string) bool {
	// 这里实现具体的CORS检查逻辑
	// 可以从配置文件或环境变量读取允许的域名
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://openpenpal.com",
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}
