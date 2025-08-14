package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
}

// DetailedValidationResponse represents a detailed validation error response
type DetailedValidationResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Error     string       `json:"error,omitempty"`
	ErrorCode string       `json:"error_code"`
	Details   []FieldError `json:"details,omitempty"`
	Timestamp int64        `json:"timestamp"`
	RequestID string       `json:"request_id,omitempty"`
}

// Field display name mappings for Chinese UI
var fieldDisplayNames = map[string]string{
	"username":    "用户名",
	"email":       "邮箱",
	"password":    "密码",
	"nickname":    "昵称",
	"school_code": "学校代码",
	"content":     "内容",
	"title":       "标题",
	"recipient":   "收件人",
	"sender":      "发件人",
	"description": "描述",
	"tags":        "标签",
	"name":        "名称",
	"price":       "价格",
	"category":    "分类",
	"status":      "状态",
	"type":        "类型",
	"amount":      "金额",
	"quantity":    "数量",
	"zone":        "区域",
	"level":       "等级",
	"reason":      "原因",
	"comment":     "评论",
	"rating":      "评分",
	"address":     "地址",
	"phone":       "电话",
	"code":        "编码",
	"message":     "消息",
	"data":        "数据",
}

// ParseValidationErrors converts Go validator errors to user-friendly field errors
func ParseValidationErrors(err error) []FieldError {
	var fieldErrors []FieldError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   strings.ToLower(fieldError.Field()),
				Value:   fieldError.Value(),
				Message: getFieldErrorMessage(fieldError),
				Code:    fieldError.Tag(),
			})
		}
	} else {
		// Handle other types of binding errors (JSON parsing, etc.)
		fieldErrors = append(fieldErrors, FieldError{
			Field:   "request",
			Message: "请求格式不正确，请检查JSON格式",
			Code:    "invalid_format",
		})
	}

	return fieldErrors
}

// getFieldErrorMessage generates user-friendly error messages in Chinese
func getFieldErrorMessage(fe validator.FieldError) string {
	fieldName := getFieldDisplayName(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s不能为空", fieldName)
	case "email":
		return "请输入有效的邮箱地址"
	case "min":
		if fe.Kind().String() == "string" {
			return fmt.Sprintf("%s长度不能少于%s个字符", fieldName, fe.Param())
		}
		return fmt.Sprintf("%s不能小于%s", fieldName, fe.Param())
	case "max":
		if fe.Kind().String() == "string" {
			return fmt.Sprintf("%s长度不能超过%s个字符", fieldName, fe.Param())
		}
		return fmt.Sprintf("%s不能大于%s", fieldName, fe.Param())
	case "len":
		return fmt.Sprintf("%s长度必须为%s个字符", fieldName, fe.Param())
	case "alphanum":
		return fmt.Sprintf("%s只能包含字母和数字", fieldName)
	case "numeric":
		return fmt.Sprintf("%s必须为数字", fieldName)
	case "oneof":
		return fmt.Sprintf("%s必须为以下值之一: %s", fieldName, fe.Param())
	case "gte":
		return fmt.Sprintf("%s必须大于或等于%s", fieldName, fe.Param())
	case "lte":
		return fmt.Sprintf("%s必须小于或等于%s", fieldName, fe.Param())
	case "gt":
		return fmt.Sprintf("%s必须大于%s", fieldName, fe.Param())
	case "lt":
		return fmt.Sprintf("%s必须小于%s", fieldName, fe.Param())
	case "uuid":
		return fmt.Sprintf("%s格式不正确，必须为有效的UUID", fieldName)
	case "url":
		return fmt.Sprintf("%s必须为有效的URL地址", fieldName)
	case "contains":
		return fmt.Sprintf("%s必须包含'%s'", fieldName, fe.Param())
	case "excludes":
		return fmt.Sprintf("%s不能包含'%s'", fieldName, fe.Param())
	case "unique":
		return fmt.Sprintf("%s已存在，请选择其他%s", fieldName, fieldName)
	default:
		return fmt.Sprintf("%s格式不正确", fieldName)
	}
}

// getFieldDisplayName returns the Chinese display name for a field
func getFieldDisplayName(field string) string {
	fieldLower := strings.ToLower(field)
	if displayName, exists := fieldDisplayNames[fieldLower]; exists {
		return displayName
	}
	return field // Fallback to original field name
}

// DetailedValidationError responds with detailed validation errors
func DetailedValidationError(c *gin.Context, message string, details []FieldError) {
	response := DetailedValidationResponse{
		Success:   false,
		Message:   message,
		Error:     "VALIDATION_ERROR",
		ErrorCode: "VALIDATION_ERROR",
		Details:   details,
		Timestamp: GetCurrentTimestamp(),
		RequestID: GetRequestID(c),
	}

	c.JSON(400, response)
}

// SimpleValidationError responds with a simple validation error
func SimpleValidationError(c *gin.Context, message string, errorCode string) {
	response := DetailedValidationResponse{
		Success:   false,
		Message:   message,
		Error:     errorCode,
		ErrorCode: errorCode,
		Timestamp: GetCurrentTimestamp(),
		RequestID: GetRequestID(c),
	}

	c.JSON(400, response)
}

// ParseAndRespondValidationError is a helper function for common validation error handling
func ParseAndRespondValidationError(c *gin.Context, err error, contextMessage string) {
	if err == nil {
		return
	}

	details := ParseValidationErrors(err)
	DetailedValidationError(c, contextMessage, details)
}

// Common validation error messages
const (
	UserRegistrationValidationMsg = "用户注册信息验证失败"
	UserLoginValidationMsg        = "用户登录信息验证失败"
	LetterValidationMsg           = "信件信息验证失败"
	MuseumValidationMsg           = "博物馆信息验证失败"
	ShopValidationMsg             = "商店信息验证失败"
	CourierValidationMsg          = "信使信息验证失败"
	AdminValidationMsg            = "管理员操作验证失败"
	AIValidationMsg               = "AI请求信息验证失败"
	EnvelopeValidationMsg         = "信封信息验证失败"
)

// Helper functions

// GetCurrentTimestamp returns the current Unix timestamp
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetRequestID extracts or generates a request ID from the Gin context
func GetRequestID(c *gin.Context) string {
	if requestID := c.GetString("request_id"); requestID != "" {
		return requestID
	}
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return "" // Return empty string if no request ID is available
}
