// Package response 提供轻量级HTTP响应处理
// 替代过度复杂的shared/pkg/response
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StandardResponse 标准响应结构 - 简化版本
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	Code      int         `json:"code"`
	Timestamp int64       `json:"timestamp"`
}

// Meta 分页元数据
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"page_size,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	StandardResponse
	Meta *Meta `json:"meta,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}, message ...string) {
	msg := "操作成功"
	if len(message) > 0 {
		msg = message[0]
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success:   true,
		Data:      data,
		Message:   msg,
		Code:      http.StatusOK,
		Timestamp: time.Now().Unix(),
	})
}

// Error 错误响应
func Error(c *gin.Context, statusCode int, err string, message ...string) {
	msg := "操作失败"
	if len(message) > 0 {
		msg = message[0]
	}

	c.JSON(statusCode, StandardResponse{
		Success:   false,
		Error:     err,
		Message:   msg,
		Code:      statusCode,
		Timestamp: time.Now().Unix(),
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, err string, message ...string) {
	Error(c, http.StatusBadRequest, err, message...)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message ...string) {
	msg := "未授权访问"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(c, http.StatusUnauthorized, "unauthorized", msg)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message ...string) {
	msg := "访问被禁止"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(c, http.StatusForbidden, "forbidden", msg)
}

// NotFound 404错误
func NotFound(c *gin.Context, message ...string) {
	msg := "资源未找到"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(c, http.StatusNotFound, "not_found", msg)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, err string, message ...string) {
	msg := "服务器内部错误"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(c, http.StatusInternalServerError, err, msg)
}

// Paginated 分页响应
func Paginated(c *gin.Context, data interface{}, page, pageSize int, total int64, message ...string) {
	msg := "获取成功"
	if len(message) > 0 {
		msg = message[0]
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, PaginatedResponse{
		StandardResponse: StandardResponse{
			Success:   true,
			Data:      data,
			Message:   msg,
			Code:      http.StatusOK,
			Timestamp: time.Now().Unix(),
		},
		Meta: &Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// JSON 自定义JSON响应
func JSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, StandardResponse{
		Success:   statusCode >= 200 && statusCode < 300,
		Data:      data,
		Code:      statusCode,
		Timestamp: time.Now().Unix(),
	})
}

// GinResponse 兼容性响应结构 - 遵循SOTA原则统一接口
type GinResponse struct{}

// NewGinResponse 创建GinResponse实例 - 兼容现有代码
func NewGinResponse() *GinResponse {
	return &GinResponse{}
}

// Success 成功响应方法
func (r *GinResponse) Success(c *gin.Context, data interface{}) {
	Success(c, data)
}

// SuccessWithMessage 带消息的成功响应
func (r *GinResponse) SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	Success(c, data, message)
}

// Created 创建成功响应
func (r *GinResponse) Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, StandardResponse{
		Success:   true,
		Data:      data,
		Message:   "创建成功",
		Code:      http.StatusCreated,
		Timestamp: time.Now().Unix(),
	})
}

// CreatedWithMessage 带自定义消息的创建成功响应
func (r *GinResponse) CreatedWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, StandardResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Code:      http.StatusCreated,
		Timestamp: time.Now().Unix(),
	})
}

// Error 错误响应方法
func (r *GinResponse) Error(c *gin.Context, status int, message string) {
	Error(c, status, message)
}

// BadRequest 400错误方法
func (r *GinResponse) BadRequest(c *gin.Context, message string) {
	BadRequest(c, message)
}

// Unauthorized 401错误方法
func (r *GinResponse) Unauthorized(c *gin.Context, message string) {
	Unauthorized(c, message)
}

// NotFound 404错误方法
func (r *GinResponse) NotFound(c *gin.Context, message string) {
	NotFound(c, message)
}

// InternalServerError 500错误方法
func (r *GinResponse) InternalServerError(c *gin.Context, message string) {
	InternalServerError(c, message)
}

// ValidationError 验证错误响应
func (r *GinResponse) ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusUnprocessableEntity, message, "验证失败")
}

// OK 简单OK响应
func (r *GinResponse) OK(c *gin.Context, message string) {
	Success(c, nil, message)
}