package models

import (
	"encoding/json"
	"time"
)

// APIResponse 统一API响应格式
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"msg"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorResponse 错误响应格式
type ErrorResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"msg"`
	Details   string      `json:"details,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Path      string      `json:"path,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Code:      0,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string, details string) ErrorResponse {
	return ErrorResponse{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// ToJSON 转换为JSON
func (e ErrorResponse) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]interface{} `json:"services,omitempty"`
	Uptime    string                 `json:"uptime,omitempty"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	HealthyCount int       `json:"healthy_count"`
	TotalCount   int       `json:"total_count"`
	LastCheck    time.Time `json:"last_check"`
	ResponseTime string    `json:"response_time,omitempty"`
	ErrorCount   int       `json:"error_count"`
	SuccessRate  float64   `json:"success_rate"`
}

// MetricsResponse 监控指标响应
type MetricsResponse struct {
	Gateway struct {
		Uptime         string  `json:"uptime"`
		TotalRequests  int64   `json:"total_requests"`
		ErrorRate      float64 `json:"error_rate"`
		AverageLatency string  `json:"average_latency"`
		RequestsPerSec float64 `json:"requests_per_sec"`
	} `json:"gateway"`
	Services map[string]ServiceMetrics `json:"services"`
}

// ServiceMetrics 服务指标
type ServiceMetrics struct {
	TotalRequests  int64   `json:"total_requests"`
	ErrorCount     int64   `json:"error_count"`
	ErrorRate      float64 `json:"error_rate"`
	AverageLatency string  `json:"average_latency"`
	P95Latency     string  `json:"p95_latency"`
	P99Latency     string  `json:"p99_latency"`
}

// LoadBalanceInfo 负载均衡信息
type LoadBalanceInfo struct {
	ServiceName      string `json:"service_name"`
	SelectedHost     string `json:"selected_host"`
	Algorithm        string `json:"algorithm"`
	TotalInstances   int    `json:"total_instances"`
	HealthyInstances int    `json:"healthy_instances"`
}

// RateLimitInfo 限流信息
type RateLimitInfo struct {
	Limit      int   `json:"limit"`
	Remaining  int   `json:"remaining"`
	ResetTime  int64 `json:"reset_time"`
	RetryAfter int   `json:"retry_after,omitempty"`
}

// CircuitBreakerStatus 熔断器状态
type CircuitBreakerStatus struct {
	ServiceName  string    `json:"service_name"`
	State        string    `json:"state"` // closed, open, half-open
	FailureRate  float64   `json:"failure_rate"`
	FailureCount int       `json:"failure_count"`
	LastFailure  time.Time `json:"last_failure"`
	NextRetry    time.Time `json:"next_retry,omitempty"`
}

// 状态码常量
const (
	// 成功状态码
	CodeSuccess = 0

	// 客户端错误状态码
	CodeBadRequest      = 400
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeTimeout         = 408
	CodeConflict        = 409
	CodeTooManyRequests = 429

	// 服务端错误状态码
	CodeInternalError      = 500
	CodeBadGateway         = 502
	CodeServiceUnavailable = 503
	CodeGatewayTimeout     = 504
)

// 错误消息常量
const (
	MsgSuccess            = "success"
	MsgBadRequest         = "Bad request"
	MsgUnauthorized       = "Unauthorized"
	MsgForbidden          = "Forbidden"
	MsgNotFound           = "Not found"
	MsgTimeout            = "Request timeout"
	MsgTooManyRequests    = "Too many requests"
	MsgInternalError      = "Internal server error"
	MsgBadGateway         = "Bad gateway"
	MsgServiceUnavailable = "Service unavailable"
	MsgGatewayTimeout     = "Gateway timeout"
)

// GetStatusMessage 根据状态码获取消息
func GetStatusMessage(code int) string {
	switch code {
	case CodeSuccess:
		return MsgSuccess
	case CodeBadRequest:
		return MsgBadRequest
	case CodeUnauthorized:
		return MsgUnauthorized
	case CodeForbidden:
		return MsgForbidden
	case CodeNotFound:
		return MsgNotFound
	case CodeTimeout:
		return MsgTimeout
	case CodeTooManyRequests:
		return MsgTooManyRequests
	case CodeInternalError:
		return MsgInternalError
	case CodeBadGateway:
		return MsgBadGateway
	case CodeServiceUnavailable:
		return MsgServiceUnavailable
	case CodeGatewayTimeout:
		return MsgGatewayTimeout
	default:
		return "Unknown error"
	}
}
