package models

import "time"

// APIResponse 统一API响应格式
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"msg"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
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

// ErrorResponse 错误响应
func ErrorResponse(code int, message string, err interface{}) APIResponse {
	return APIResponse{
		Code:      code,
		Message:   message,
		Error:     err,
		Timestamp: time.Now(),
	}
}

// 状态码常量
const (
	CodeSuccess           = 0   // 成功
	CodeParamError       = 1   // 参数错误
	CodeUnauthorized     = 2   // 无权限
	CodeNotFound         = 3   // 资源不存在
	CodeConflict         = 4   // 冲突
	CodeInternalError    = 500 // 服务器内部错误
)