package errors

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// PanicRecovery panic恢复处理
type PanicRecovery struct {
	Logger Logger
}

// Logger 定义日志接口
type Logger interface {
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewPanicRecovery 创建panic恢复处理器
func NewPanicRecovery(logger Logger) *PanicRecovery {
	return &PanicRecovery{
		Logger: logger,
	}
}

// Recover 恢复panic并转换为错误
func (pr *PanicRecovery) Recover() error {
	if r := recover(); r != nil {
		// 获取堆栈信息
		stack := debug.Stack()
		
		// 获取panic发生的位置
		pc, file, line, ok := runtime.Caller(2)
		location := "unknown"
		if ok {
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				location = fmt.Sprintf("%s:%d %s", getShortFilename(file), line, fn.Name())
			} else {
				location = fmt.Sprintf("%s:%d", getShortFilename(file), line)
			}
		}
		
		// 记录panic日志
		pr.Logger.Error("Panic recovered",
			"panic", r,
			"location", location,
			"stack", string(stack),
		)
		
		// 转换为自定义错误
		return &CourierServiceError{
			Code:       CodeInternalError,
			Message:    fmt.Sprintf("Panic recovered: %v", r),
			Type:       TypeSystem,
			Context:    map[string]interface{}{"panic_location": location},
			StackTrace: string(stack),
			Retryable:  false,
			HTTPStatus: 500,
		}
	}
	return nil
}

// RecoverWithCallback 带回调的panic恢复
func (pr *PanicRecovery) RecoverWithCallback(callback func(error)) error {
	if err := pr.Recover(); err != nil {
		if callback != nil {
			callback(err)
		}
		return err
	}
	return nil
}

// SafeGo 安全的goroutine启动
func (pr *PanicRecovery) SafeGo(fn func()) {
	go func() {
		defer func() {
			if err := pr.Recover(); err != nil {
				pr.Logger.Error("Goroutine panic recovered", "error", err)
			}
		}()
		fn()
	}()
}

// SafeGoWithContext 带上下文的安全goroutine启动
func (pr *PanicRecovery) SafeGoWithContext(fn func(), context map[string]interface{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				pr.Logger.Error("Goroutine panic recovered",
					"panic", r,
					"context", context,
					"stack", string(stack),
				)
			}
		}()
		fn()
	}()
}

// getShortFilename 获取短文件名
func getShortFilename(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return path
}

// RecoveryWrapper 通用恢复包装器
func RecoveryWrapper(logger Logger, operation string) func() error {
	return func() error {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logger.Error("Panic recovered in operation",
					"operation", operation,
					"panic", r,
					"stack", string(stack),
				)
			}
		}()
		return nil
	}
}

// DatabaseRecovery 数据库操作恢复
func DatabaseRecovery(logger Logger, operation string) func(func() error) error {
	return func(fn func() error) error {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logger.Error("Database operation panic recovered",
					"operation", operation,
					"panic", r,
					"stack", string(stack),
				)
			}
		}()
		return fn()
	}
}

// ServiceRecovery 服务层恢复
func ServiceRecovery(logger Logger, service string, method string) func(func() error) error {
	return func(fn func() error) error {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logger.Error("Service method panic recovered",
					"service", service,
					"method", method,
					"panic", r,
					"stack", string(stack),
				)
			}
		}()
		return fn()
	}
}

// ErrorRecoveryStrategy 错误恢复策略
type ErrorRecoveryStrategy interface {
	ShouldRecover(error) bool
	Recover(error) error
}

// DefaultRecoveryStrategy 默认恢复策略
type DefaultRecoveryStrategy struct {
	MaxRetries int
	Logger     Logger
}

// ShouldRecover 判断是否应该恢复
func (s *DefaultRecoveryStrategy) ShouldRecover(err error) bool {
	var courierErr *CourierServiceError
	if As(err, &courierErr) {
		return courierErr.IsRetryable()
	}
	return false
}

// Recover 执行恢复操作
func (s *DefaultRecoveryStrategy) Recover(err error) error {
	var courierErr *CourierServiceError
	if As(err, &courierErr) {
		if courierErr.IsRetryable() {
			s.Logger.Info("Attempting error recovery", "error", err)
			// 这里可以实现具体的恢复逻辑
			// 例如重试、降级、缓存fallback等
		}
	}
	return err
}

// As 类型断言辅助函数
func As(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	
	switch v := target.(type) {
	case **CourierServiceError:
		if courierErr, ok := err.(*CourierServiceError); ok {
			*v = courierErr
			return true
		}
	}
	return false
}