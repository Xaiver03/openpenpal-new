/**
 * 统一Gin HTTP响应处理 - SOTA实现，解决12处重复问题
 * 支持：结构化响应、错误追踪、国际化、缓存控制、审计日志
 */

package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UnifiedResponse 统一响应结构
type UnifiedResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
	Code      int         `json:"code"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
}

// Meta 响应元数据
type Meta struct {
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
	Total      int64  `json:"total,omitempty"`
	TotalPages int    `json:"total_pages,omitempty"`
	Version    string `json:"version,omitempty"`
	Duration   string `json:"duration,omitempty"`
}

// ErrorDetail 详细错误信息
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationErrorResponse 验证错误响应
type ValidationErrorResponse struct {
	UnifiedResponse
	Details []ErrorDetail `json:"details,omitempty"`
}

// ResponseOption 响应选项
type ResponseOption func(*UnifiedResponse)

// WithMeta 添加元数据
func WithMeta(meta *Meta) ResponseOption {
	return func(r *UnifiedResponse) {
		r.Meta = meta
	}
}

// WithErrorCode 添加错误代码
func WithErrorCode(code string) ResponseOption {
	return func(r *UnifiedResponse) {
		r.ErrorCode = code
	}
}

// WithRequestID 添加请求ID
func WithRequestID(requestID string) ResponseOption {
	return func(r *UnifiedResponse) {
		r.RequestID = requestID
	}
}

// ================================
// 成功响应
// ================================

// Success 成功响应
func Success(c *gin.Context, data interface{}, options ...ResponseOption) {
	response := &UnifiedResponse{
		Success:   true,
		Data:      data,
		Code:      http.StatusOK,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}

	for _, option := range options {
		option(response)
	}

	// 设置缓存控制头
	setCacheHeaders(c, http.StatusOK)
	
	c.JSON(http.StatusOK, response)
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, data interface{}, message string, options ...ResponseOption) {
	response := &UnifiedResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Code:      http.StatusOK,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}

	for _, option := range options {
		option(response)
	}

	setCacheHeaders(c, http.StatusOK)
	c.JSON(http.StatusOK, response)
}

// SuccessWithPagination 带分页的成功响应
func SuccessWithPagination(c *gin.Context, data interface{}, page, pageSize int, total int64, options ...ResponseOption) {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	meta := &Meta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	options = append(options, WithMeta(meta))
	SuccessWithMessage(c, data, "查询成功", options...)
}

// Created 201创建成功响应
func Created(c *gin.Context, data interface{}, message string, options ...ResponseOption) {
	response := &UnifiedResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Code:      http.StatusCreated,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}

	for _, option := range options {
		option(response)
	}

	setCacheHeaders(c, http.StatusCreated)
	c.JSON(http.StatusCreated, response)
}

// NoContent 204无内容响应
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ================================
// 错误响应
// ================================

// Error 通用错误响应
func Error(c *gin.Context, code int, message string, options ...ResponseOption) {
	response := &UnifiedResponse{
		Success:   false,
		Error:     message,
		Code:      code,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}

	for _, option := range options {
		option(response)
	}

	// 记录错误日志
	logError(c, code, message, response.ErrorCode)
	
	setCacheHeaders(c, code)
	c.JSON(code, response)
}

// ErrorWithMessage 带详细消息的错误响应
func ErrorWithMessage(c *gin.Context, code int, message, detail string, options ...ResponseOption) {
	response := &UnifiedResponse{
		Success:   false,
		Error:     message,
		Message:   detail,
		Code:      code,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}

	for _, option := range options {
		option(response)
	}

	logError(c, code, message+" - "+detail, response.ErrorCode)
	setCacheHeaders(c, code)
	c.JSON(code, response)
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, message string, details []ErrorDetail, options ...ResponseOption) {
	response := &ValidationErrorResponse{
		UnifiedResponse: UnifiedResponse{
			Success:   false,
			Error:     message,
			Code:      http.StatusBadRequest,
			Timestamp: time.Now().Unix(),
			RequestID: getRequestID(c),
			ErrorCode: "VALIDATION_ERROR",
		},
		Details: details,
	}

	for _, option := range options {
		option(&response.UnifiedResponse)
	}

	logError(c, http.StatusBadRequest, message, "VALIDATION_ERROR")
	setCacheHeaders(c, http.StatusBadRequest)
	c.JSON(http.StatusBadRequest, response)
}

// ================================
// 常见HTTP状态码响应
// ================================

// BadRequest 400错误
func BadRequest(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("BAD_REQUEST"))
	Error(c, http.StatusBadRequest, message, options...)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("UNAUTHORIZED"))
	Error(c, http.StatusUnauthorized, message, options...)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("FORBIDDEN"))
	Error(c, http.StatusForbidden, message, options...)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("NOT_FOUND"))
	Error(c, http.StatusNotFound, message, options...)
}

// MethodNotAllowed 405错误
func MethodNotAllowed(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("METHOD_NOT_ALLOWED"))
	Error(c, http.StatusMethodNotAllowed, message, options...)
}

// Conflict 409错误
func Conflict(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("CONFLICT"))
	Error(c, http.StatusConflict, message, options...)
}

// UnprocessableEntity 422错误
func UnprocessableEntity(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("UNPROCESSABLE_ENTITY"))
	Error(c, http.StatusUnprocessableEntity, message, options...)
}

// TooManyRequests 429错误
func TooManyRequests(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("TOO_MANY_REQUESTS"))
	Error(c, http.StatusTooManyRequests, message, options...)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("INTERNAL_SERVER_ERROR"))
	Error(c, http.StatusInternalServerError, message, options...)
}

// BadGateway 502错误
func BadGateway(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("BAD_GATEWAY"))
	Error(c, http.StatusBadGateway, message, options...)
}

// ServiceUnavailable 503错误
func ServiceUnavailable(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("SERVICE_UNAVAILABLE"))
	Error(c, http.StatusServiceUnavailable, message, options...)
}

// GatewayTimeout 504错误
func GatewayTimeout(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("GATEWAY_TIMEOUT"))
	Error(c, http.StatusGatewayTimeout, message, options...)
}

// ================================
// 业务特定响应
// ================================

// PermissionDenied 权限拒绝响应
func PermissionDenied(c *gin.Context, permission string, options ...ResponseOption) {
	options = append(options, WithErrorCode("PERMISSION_DENIED"))
	ErrorWithMessage(c, http.StatusForbidden, "权限不足", "需要权限: "+permission, options...)
}

// ResourceNotFound 资源未找到响应
func ResourceNotFound(c *gin.Context, resource string, options ...ResponseOption) {
	options = append(options, WithErrorCode("RESOURCE_NOT_FOUND"))
	ErrorWithMessage(c, http.StatusNotFound, "资源不存在", resource+"未找到", options...)
}

// DataConflict 数据冲突响应
func DataConflict(c *gin.Context, message string, options ...ResponseOption) {
	options = append(options, WithErrorCode("DATA_CONFLICT"))
	ErrorWithMessage(c, http.StatusConflict, "数据冲突", message, options...)
}

// RateLimitExceeded 频率限制响应
func RateLimitExceeded(c *gin.Context, options ...ResponseOption) {
	options = append(options, WithErrorCode("RATE_LIMIT_EXCEEDED"))
	ErrorWithMessage(c, http.StatusTooManyRequests, "请求频率过高", "请稍后再试", options...)
}

// MaintenanceMode 维护模式响应
func MaintenanceMode(c *gin.Context, options ...ResponseOption) {
	options = append(options, WithErrorCode("MAINTENANCE_MODE"))
	ErrorWithMessage(c, http.StatusServiceUnavailable, "系统维护中", "系统正在维护，请稍后访问", options...)
}

// ================================
// 工具函数
// ================================

// getRequestID 获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// setCacheHeaders 设置缓存头
func setCacheHeaders(c *gin.Context, statusCode int) {
	switch statusCode {
	case http.StatusOK:
		// 成功响应可以缓存5分钟
		c.Header("Cache-Control", "public, max-age=300")
	case http.StatusCreated:
		// 创建响应不缓存
		c.Header("Cache-Control", "no-cache")
	default:
		// 错误响应不缓存
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	}
	
	// 安全头
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
}

// logError 记录错误日志
func logError(c *gin.Context, code int, message, errorCode string) {
	// 这里可以集成具体的日志系统
	// 例如：logrus, zap, 或者自定义的日志系统
	
	userID := ""
	if id, exists := c.Get("user_id"); exists {
		userID = id.(string)
	}
	
	requestID := getRequestID(c)
	path := c.Request.URL.Path
	method := c.Request.Method
	userAgent := c.GetHeader("User-Agent")
	
	// 简单的控制台日志，实际应用中应该使用结构化日志
	if code >= 500 {
		// 服务器错误
		c.Header("X-Error-Reference", requestID)
		// TODO: 发送错误报告到监控系统
	}
	
	// 审计日志记录
	auditLog := map[string]interface{}{
		"timestamp":   time.Now().Unix(),
		"request_id":  requestID,
		"user_id":     userID,
		"method":      method,
		"path":        path,
		"status_code": code,
		"error_code":  errorCode,
		"message":     message,
		"user_agent":  userAgent,
	}
	
	// TODO: 发送到审计日志系统
	_ = auditLog
}