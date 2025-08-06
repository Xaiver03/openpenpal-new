package errors

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrorCode 定义错误代码类型
type ErrorCode string

const (
	// 通用错误代码
	CodeSuccess         ErrorCode = "SUCCESS"
	CodeInternalError   ErrorCode = "INTERNAL_ERROR"
	CodeInvalidRequest  ErrorCode = "INVALID_REQUEST"
	CodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	CodeForbidden       ErrorCode = "FORBIDDEN"
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeConflict        ErrorCode = "CONFLICT"
	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// 数据库相关错误
	CodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	CodeRecordNotFound     ErrorCode = "RECORD_NOT_FOUND"
	CodeDuplicateRecord    ErrorCode = "DUPLICATE_RECORD"
	CodeConstraintViolation ErrorCode = "CONSTRAINT_VIOLATION"
	CodeDeadlock           ErrorCode = "DEADLOCK"
	CodeConnectionTimeout  ErrorCode = "CONNECTION_TIMEOUT"

	// 业务逻辑错误
	CodeInvalidCourierStatus ErrorCode = "INVALID_COURIER_STATUS"
	CodeInsufficientPermission ErrorCode = "INSUFFICIENT_PERMISSION"
	CodeTaskNotAssignable    ErrorCode = "TASK_NOT_ASSIGNABLE"
	CodeInvalidZoneCode      ErrorCode = "INVALID_ZONE_CODE"
	CodeLevelMismatch        ErrorCode = "LEVEL_MISMATCH"
	CodeInvalidHierarchy     ErrorCode = "INVALID_HIERARCHY"

	// 外部服务错误
	CodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	CodeCircuitBreakerOpen   ErrorCode = "CIRCUIT_BREAKER_OPEN"
	CodeRetryExhausted      ErrorCode = "RETRY_EXHAUSTED"

	// 队列相关错误
	CodeQueueError         ErrorCode = "QUEUE_ERROR"
	CodeMessageProcessing  ErrorCode = "MESSAGE_PROCESSING_ERROR"
	CodeDeadLetterQueue    ErrorCode = "DEAD_LETTER_QUEUE"

	// WebSocket相关错误
	CodeWebSocketError     ErrorCode = "WEBSOCKET_ERROR"
	CodeConnectionLost     ErrorCode = "CONNECTION_LOST"
	CodeBroadcastFailed    ErrorCode = "BROADCAST_FAILED"

	// 验证相关错误
	CodeValidationError    ErrorCode = "VALIDATION_ERROR"
	CodeMissingField       ErrorCode = "MISSING_FIELD"
	CodeInvalidFormat      ErrorCode = "INVALID_FORMAT"
)

// ErrorType 定义错误类型
type ErrorType string

const (
	TypeTemporary   ErrorType = "TEMPORARY"   // 临时错误，可重试
	TypePermanent   ErrorType = "PERMANENT"   // 永久错误，不可重试
	TypeRetryable   ErrorType = "RETRYABLE"   // 可重试错误
	TypeValidation  ErrorType = "VALIDATION"  // 验证错误
	TypeBusiness    ErrorType = "BUSINESS"    // 业务逻辑错误
	TypeSystem      ErrorType = "SYSTEM"      // 系统错误
)

// CourierServiceError 自定义错误结构
type CourierServiceError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Type        ErrorType              `json:"type"`
	Cause       error                  `json:"-"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Retryable   bool                   `json:"retryable"`
	HTTPStatus  int                    `json:"-"`
}

// Error 实现error接口
func (e *CourierServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 支持错误链
func (e *CourierServiceError) Unwrap() error {
	return e.Cause
}

// IsType 检查错误类型
func (e *CourierServiceError) IsType(errorType ErrorType) bool {
	return e.Type == errorType
}

// IsRetryable 检查是否可重试
func (e *CourierServiceError) IsRetryable() bool {
	return e.Retryable || e.Type == TypeTemporary || e.Type == TypeRetryable
}

// WithContext 添加上下文信息
func (e *CourierServiceError) WithContext(key string, value interface{}) *CourierServiceError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithRequestID 添加请求ID
func (e *CourierServiceError) WithRequestID(requestID string) *CourierServiceError {
	e.RequestID = requestID
	return e
}

// WithUserID 添加用户ID
func (e *CourierServiceError) WithUserID(userID string) *CourierServiceError {
	e.UserID = userID
	return e
}

// NewError 创建新的错误
func NewError(code ErrorCode, message string, errorType ErrorType) *CourierServiceError {
	return &CourierServiceError{
		Code:       code,
		Message:    message,
		Type:       errorType,
		Timestamp:  time.Now(),
		Retryable:  errorType == TypeTemporary || errorType == TypeRetryable,
		HTTPStatus: getHTTPStatus(code),
	}
}

// Wrap 包装现有错误
func Wrap(err error, code ErrorCode, message string, errorType ErrorType) *CourierServiceError {
	return &CourierServiceError{
		Code:       code,
		Message:    message,
		Type:       errorType,
		Cause:      err,
		Timestamp:  time.Now(),
		Retryable:  errorType == TypeTemporary || errorType == TypeRetryable,
		HTTPStatus: getHTTPStatus(code),
	}
}

// getHTTPStatus 根据错误代码返回对应的HTTP状态码
func getHTTPStatus(code ErrorCode) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeInvalidRequest, CodeValidationError, CodeMissingField, CodeInvalidFormat:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden, CodeInsufficientPermission:
		return http.StatusForbidden
	case CodeNotFound, CodeRecordNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeDuplicateRecord:
		return http.StatusConflict
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeInternalError, CodeDatabaseError, CodeExternalServiceError:
		return http.StatusInternalServerError
	case CodeServiceUnavailable, CodeCircuitBreakerOpen:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// 预定义错误构造函数

// NewDatabaseError 数据库错误
func NewDatabaseError(err error, operation string) *CourierServiceError {
	return Wrap(err, CodeDatabaseError, fmt.Sprintf("Database operation failed: %s", operation), TypeSystem).
		WithContext("operation", operation)
}

// NewValidationError 验证错误
func NewValidationError(field string, message string) *CourierServiceError {
	return NewError(CodeValidationError, fmt.Sprintf("Validation failed for field '%s': %s", field, message), TypeValidation).
		WithContext("field", field)
}

// NewBusinessError 业务逻辑错误
func NewBusinessError(code ErrorCode, message string) *CourierServiceError {
	return NewError(code, message, TypeBusiness)
}

// NewNotFoundError 记录不存在错误
func NewNotFoundError(resource string, id interface{}) *CourierServiceError {
	return NewError(CodeRecordNotFound, fmt.Sprintf("%s not found: %v", resource, id), TypePermanent).
		WithContext("resource", resource).
		WithContext("id", id)
}

// NewPermissionError 权限不足错误
func NewPermissionError(operation string) *CourierServiceError {
	return NewError(CodeInsufficientPermission, fmt.Sprintf("Insufficient permission for operation: %s", operation), TypePermanent).
		WithContext("operation", operation)
}

// NewCircuitBreakerError 熔断器错误
func NewCircuitBreakerError(service string) *CourierServiceError {
	return NewError(CodeCircuitBreakerOpen, fmt.Sprintf("Circuit breaker is open for service: %s", service), TypeTemporary).
		WithContext("service", service)
}

// IsError 检查错误是否为特定类型
func IsError(err error, code ErrorCode) bool {
	var courierErr *CourierServiceError
	if errors.As(err, &courierErr) {
		return courierErr.Code == code
	}
	return false
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	var courierErr *CourierServiceError
	if errors.As(err, &courierErr) {
		return courierErr.IsRetryable()
	}
	// 默认情况下，未知错误不可重试
	return false
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) ErrorCode {
	var courierErr *CourierServiceError
	if errors.As(err, &courierErr) {
		return courierErr.Code
	}
	return CodeInternalError
}

// GetHTTPStatus 获取HTTP状态码
func GetHTTPStatus(err error) int {
	var courierErr *CourierServiceError
	if errors.As(err, &courierErr) {
		return courierErr.HTTPStatus
	}
	return http.StatusInternalServerError
}