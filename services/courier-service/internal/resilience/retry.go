package resilience

import (
	"context"
	"courier-service/internal/errors"
	"courier-service/internal/logging"
	"math"
	"math/rand"
	"time"
)

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts     int                `json:"max_attempts"`
	InitialInterval time.Duration      `json:"initial_interval"`
	MaxInterval     time.Duration      `json:"max_interval"`
	Multiplier      float64            `json:"multiplier"`
	Jitter          bool               `json:"jitter"`
	RetryableErrors []errors.ErrorCode `json:"retryable_errors"`
}

// DefaultRetryPolicy 默认重试策略
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:     3,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		Jitter:          true,
		RetryableErrors: []errors.ErrorCode{
			errors.CodeDatabaseError,
			errors.CodeConnectionTimeout,
			errors.CodeDeadlock,
			errors.CodeExternalServiceError,
			errors.CodeServiceUnavailable,
		},
	}
}

// DatabaseRetryPolicy 数据库操作重试策略
func DatabaseRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:     5,
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     5 * time.Second,
		Multiplier:      1.5,
		Jitter:          true,
		RetryableErrors: []errors.ErrorCode{
			errors.CodeDatabaseError,
			errors.CodeConnectionTimeout,
			errors.CodeDeadlock,
		},
	}
}

// ExternalServiceRetryPolicy 外部服务重试策略
func ExternalServiceRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:     3,
		InitialInterval: 200 * time.Millisecond,
		MaxInterval:     30 * time.Second,
		Multiplier:      3.0,
		Jitter:          true,
		RetryableErrors: []errors.ErrorCode{
			errors.CodeExternalServiceError,
			errors.CodeServiceUnavailable,
			errors.CodeConnectionTimeout,
		},
	}
}

// RetryableOperation 可重试的操作
type RetryableOperation func() error

// RetryResult 重试结果
type RetryResult struct {
	Success     bool          `json:"success"`
	Attempts    int           `json:"attempts"`
	TotalTime   time.Duration `json:"total_time"`
	LastError   error         `json:"last_error,omitempty"`
	RetryErrors []error       `json:"retry_errors,omitempty"`
}

// Retrier 重试器
type Retrier struct {
	policy RetryPolicy
	logger logging.Logger
}

// NewRetrier 创建重试器
func NewRetrier(policy RetryPolicy, logger logging.Logger) *Retrier {
	if logger == nil {
		logger = logging.GetDefaultLogger()
	}

	return &Retrier{
		policy: policy,
		logger: logger,
	}
}

// Execute 执行重试操作
func (r *Retrier) Execute(ctx context.Context, operation RetryableOperation) *RetryResult {
	return r.ExecuteWithCallback(ctx, operation, nil)
}

// ExecuteWithCallback 执行重试操作（带回调）
func (r *Retrier) ExecuteWithCallback(
	ctx context.Context,
	operation RetryableOperation,
	onRetry func(attempt int, err error),
) *RetryResult {
	result := &RetryResult{
		RetryErrors: make([]error, 0),
	}

	startTime := time.Now()

	for attempt := 1; attempt <= r.policy.MaxAttempts; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			result.LastError = ctx.Err()
			result.TotalTime = time.Since(startTime)
			return result
		default:
		}

		result.Attempts = attempt

		// 执行操作
		err := operation()
		if err == nil {
			// 成功
			result.Success = true
			result.TotalTime = time.Since(startTime)

			if attempt > 1 {
				r.logger.Info("Operation succeeded after retry",
					"attempts", attempt,
					"total_time", result.TotalTime,
				)
			}

			return result
		}

		// 记录错误
		result.RetryErrors = append(result.RetryErrors, err)
		result.LastError = err

		// 检查是否可重试
		if !r.isRetryable(err) || attempt >= r.policy.MaxAttempts {
			break
		}

		// 调用重试回调
		if onRetry != nil {
			onRetry(attempt, err)
		}

		// 记录重试日志
		r.logger.Warn("Operation failed, retrying",
			"attempt", attempt,
			"max_attempts", r.policy.MaxAttempts,
			"error", err,
		)

		// 计算延迟时间
		delay := r.calculateDelay(attempt)

		// 等待重试
		select {
		case <-ctx.Done():
			result.LastError = ctx.Err()
			result.TotalTime = time.Since(startTime)
			return result
		case <-time.After(delay):
			// 继续重试
		}
	}

	result.TotalTime = time.Since(startTime)

	r.logger.Error("Operation failed after all retries",
		"attempts", result.Attempts,
		"total_time", result.TotalTime,
		"last_error", result.LastError,
	)

	return result
}

// isRetryable 检查错误是否可重试
func (r *Retrier) isRetryable(err error) bool {
	// 首先检查是否是自定义错误
	var courierErr *errors.CourierServiceError
	if errors.As(err, &courierErr) {
		// 检查错误类型
		if courierErr.IsRetryable() {
			return true
		}

		// 检查特定错误代码
		for _, retryableCode := range r.policy.RetryableErrors {
			if courierErr.Code == retryableCode {
				return true
			}
		}
	}

	// 检查特定的Go标准错误
	if isNetworkError(err) || isTemporaryError(err) {
		return true
	}

	return false
}

// calculateDelay 计算延迟时间
func (r *Retrier) calculateDelay(attempt int) time.Duration {
	// 计算指数退避延迟
	delay := float64(r.policy.InitialInterval) * math.Pow(r.policy.Multiplier, float64(attempt-1))

	// 限制最大延迟
	if time.Duration(delay) > r.policy.MaxInterval {
		delay = float64(r.policy.MaxInterval)
	}

	// 添加随机抖动
	if r.policy.Jitter {
		jitter := rand.Float64() * 0.1 * delay // 10%的抖动
		delay += jitter
	}

	return time.Duration(delay)
}

// RetryWithPolicy 使用指定策略重试
func RetryWithPolicy(
	ctx context.Context,
	policy RetryPolicy,
	operation RetryableOperation,
	logger logging.Logger,
) *RetryResult {
	retrier := NewRetrier(policy, logger)
	return retrier.Execute(ctx, operation)
}

// RetryDatabase 重试数据库操作
func RetryDatabase(
	ctx context.Context,
	operation RetryableOperation,
	logger logging.Logger,
) *RetryResult {
	return RetryWithPolicy(ctx, DatabaseRetryPolicy(), operation, logger)
}

// RetryExternalService 重试外部服务调用
func RetryExternalService(
	ctx context.Context,
	operation RetryableOperation,
	logger logging.Logger,
) *RetryResult {
	return RetryWithPolicy(ctx, ExternalServiceRetryPolicy(), operation, logger)
}

// 辅助函数

// isNetworkError 检查是否是网络错误
func isNetworkError(err error) bool {
	// 检查常见的网络错误类型
	errStr := err.Error()
	networkErrors := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"network is unreachable",
		"no route to host",
		"temporary failure",
	}

	for _, netErr := range networkErrors {
		if contains(errStr, netErr) {
			return true
		}
	}

	return false
}

// isTemporaryError 检查是否是临时错误
func isTemporaryError(err error) bool {
	type temporary interface {
		Temporary() bool
	}

	if te, ok := err.(temporary); ok {
		return te.Temporary()
	}

	return false
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

// indexOf 查找子字符串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// RetryableFunc 带重试的函数包装器
func RetryableFunc[T any](
	ctx context.Context,
	policy RetryPolicy,
	fn func() (T, error),
	logger logging.Logger,
) (T, error) {
	var result T
	var lastErr error

	operation := func() error {
		var err error
		result, err = fn()
		lastErr = err
		return err
	}

	retryResult := RetryWithPolicy(ctx, policy, operation, logger)

	if retryResult.Success {
		return result, nil
	}

	return result, lastErr
}

// AsyncRetryableOperation 异步可重试操作
type AsyncRetryableOperation struct {
	Operation  RetryableOperation
	Policy     RetryPolicy
	Logger     logging.Logger
	OnComplete func(*RetryResult)
	OnRetry    func(attempt int, err error)
}

// ExecuteAsync 异步执行重试操作
func (aro *AsyncRetryableOperation) ExecuteAsync(ctx context.Context) {
	go func() {
		retrier := NewRetrier(aro.Policy, aro.Logger)
		result := retrier.ExecuteWithCallback(ctx, aro.Operation, aro.OnRetry)

		if aro.OnComplete != nil {
			aro.OnComplete(result)
		}
	}()
}
