package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// InputValidator 输入验证器
type InputValidator struct {
	validator *validator.Validate
}

// NewInputValidator 创建输入验证器
func NewInputValidator() *InputValidator {
	v := validator.New()
	
	// 注册自定义验证规则
	v.RegisterValidation("no_xss", validateNoXSS)
	v.RegisterValidation("safe_string", validateSafeString)
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("safe_html", validateSafeHTML)
	v.RegisterValidation("no_sql_injection", validateNoSQLInjection)
	v.RegisterValidation("safe_filename", validateSafeFilename)
	v.RegisterValidation("chinese_name", validateChineseName)
	v.RegisterValidation("phone_cn", validatePhoneCN)
	v.RegisterValidation("postcode_cn", validatePostcodeCN)
	
	return &InputValidator{
		validator: v,
	}
}

// 全局输入验证器实例
var globalValidator = NewInputValidator()

// InputValidation 输入验证中间件
func InputValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证请求头
		if err := validateHeaders(c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request headers",
				"details": err.Error(),
			})
			c.Abort()
			return
		}
		
		// 验证查询参数
		if err := validateQueryParams(c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid query parameters",
				"details": err.Error(),
			})
			c.Abort()
			return
		}
		
		// 验证URL路径参数
		if err := validatePathParams(c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid path parameters",
				"details": err.Error(),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// ContentLengthValidation 内容长度验证中间件
func ContentLengthValidation(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("Request body too large. Maximum allowed: %d bytes", maxBytes),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// validateHeaders 验证请求头
func validateHeaders(c *gin.Context) error {
	// 验证User-Agent
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		return fmt.Errorf("User-Agent header is required")
	}
	
	// 防止超长User-Agent
	if len(userAgent) > 500 {
		return fmt.Errorf("User-Agent header too long")
	}
	
	// 验证Content-Type（如果有body）
	if c.Request.ContentLength > 0 {
		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			return fmt.Errorf("Content-Type header is required for requests with body")
		}
		
		// 检查是否是允许的Content-Type
		allowedTypes := []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
		}
		
		valid := false
		for _, allowed := range allowedTypes {
			if strings.HasPrefix(contentType, allowed) {
				valid = true
				break
			}
		}
		
		if !valid {
			return fmt.Errorf("unsupported Content-Type: %s", contentType)
		}
	}
	
	return nil
}

// validateQueryParams 验证查询参数
func validateQueryParams(c *gin.Context) error {
	for key, values := range c.Request.URL.Query() {
		// 验证参数名
		if !isValidParamName(key) {
			return fmt.Errorf("invalid query parameter name: %s", key)
		}
		
		// 验证参数值
		for _, value := range values {
			if !isValidParamValue(value) {
				return fmt.Errorf("invalid query parameter value for %s", key)
			}
		}
	}
	return nil
}

// validatePathParams 验证路径参数
func validatePathParams(c *gin.Context) error {
	for _, param := range c.Params {
		// 验证参数名
		if !isValidParamName(param.Key) {
			return fmt.Errorf("invalid path parameter name: %s", param.Key)
		}
		
		// 验证参数值
		if !isValidParamValue(param.Value) {
			return fmt.Errorf("invalid path parameter value for %s", param.Key)
		}
		
		// 特殊参数验证
		switch param.Key {
		case "id":
			if !isValidID(param.Value) {
				return fmt.Errorf("invalid ID format")
			}
		case "username":
			if !isValidUsername(param.Value) {
				return fmt.Errorf("invalid username format")
			}
		}
	}
	return nil
}

// 验证函数

// validateNoXSS 验证是否包含XSS攻击内容
func validateNoXSS(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// XSS危险模式
	xssPatterns := []string{
		`<script`,
		`javascript:`,
		`onerror=`,
		`onclick=`,
		`onmouseover=`,
		`<iframe`,
		`<object`,
		`<embed`,
		`vbscript:`,
		`data:text/html`,
	}
	
	valueLower := strings.ToLower(value)
	for _, pattern := range xssPatterns {
		if strings.Contains(valueLower, pattern) {
			return false
		}
	}
	
	return true
}

// validateSafeString 验证安全字符串
func validateSafeString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// 检查是否包含控制字符
	for _, r := range value {
		if r < 32 && r != 9 && r != 10 && r != 13 {
			return false
		}
	}
	
	// 检查长度
	if utf8.RuneCountInString(value) > 10000 {
		return false
	}
	
	return true
}

// validateUsername 验证用户名
func validateUsername(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// 用户名规则：3-20个字符，只能包含字母、数字、下划线、中划线
	pattern := `^[a-zA-Z0-9_-]{3,20}$`
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

// validateSafeHTML 验证安全的HTML内容
func validateSafeHTML(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// 禁止的HTML标签
	dangerousTags := []string{
		"script", "iframe", "object", "embed", "form",
		"input", "button", "select", "textarea", "style",
		"link", "meta", "base",
	}
	
	valueLower := strings.ToLower(value)
	for _, tag := range dangerousTags {
		if strings.Contains(valueLower, "<"+tag) {
			return false
		}
	}
	
	return true
}

// validateNoSQLInjection 验证是否包含SQL注入
func validateNoSQLInjection(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// SQL注入关键词
	sqlKeywords := []string{
		"select ", "insert ", "update ", "delete ",
		"drop ", "create ", "alter ", "exec ",
		"union ", "script ", "--", "/*", "*/",
		"xp_", "sp_", "';", "';",
	}
	
	valueLower := strings.ToLower(value)
	for _, keyword := range sqlKeywords {
		if strings.Contains(valueLower, keyword) {
			return false
		}
	}
	
	return true
}

// validateSafeFilename 验证安全的文件名
func validateSafeFilename(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// 文件名规则：只允许字母、数字、点、下划线、中划线
	pattern := `^[a-zA-Z0-9._-]+$`
	matched, _ := regexp.MatchString(pattern, value)
	
	// 防止路径穿越
	if strings.Contains(value, "..") || strings.Contains(value, "/") || strings.Contains(value, "\\") {
		return false
	}
	
	// 检查文件扩展名
	dangerousExts := []string{
		".exe", ".bat", ".cmd", ".com", ".scr",
		".vbs", ".js", ".jar", ".zip", ".rar",
	}
	
	valueLower := strings.ToLower(value)
	for _, ext := range dangerousExts {
		if strings.HasSuffix(valueLower, ext) {
			return false
		}
	}
	
	return matched
}

// validateChineseName 验证中文姓名
func validateChineseName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// 中文姓名：2-10个中文字符
	pattern := `^[\p{Han}]{2,10}$`
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

// validatePhoneCN 验证中国手机号
func validatePhoneCN(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// 中国手机号规则
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

// validatePostcodeCN 验证中国邮编
func validatePostcodeCN(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// 中国邮编：6位数字
	pattern := `^\d{6}$`
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

// 辅助函数

// isValidParamName 验证参数名是否有效
func isValidParamName(name string) bool {
	// 参数名只允许字母、数字、下划线
	pattern := `^[a-zA-Z0-9_]+$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched && len(name) <= 50
}

// isValidParamValue 验证参数值是否有效
func isValidParamValue(value string) bool {
	// 检查长度
	if len(value) > 1000 {
		return false
	}
	
	// 检查是否包含控制字符
	for _, r := range value {
		if r < 32 && r != 9 && r != 10 && r != 13 {
			return false
		}
	}
	
	return true
}

// isValidID 验证ID格式
func isValidID(id string) bool {
	// UUID格式或数字ID
	uuidPattern := `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`
	numericPattern := `^\d{1,20}$`
	
	matchedUUID, _ := regexp.MatchString(uuidPattern, id)
	matchedNumeric, _ := regexp.MatchString(numericPattern, id)
	
	return matchedUUID || matchedNumeric
}

// isValidUsername 验证用户名格式
func isValidUsername(username string) bool {
	pattern := `^[a-zA-Z0-9_-]{3,20}$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
}

// ValidateStruct 验证结构体
func ValidateStruct(s interface{}) error {
	return globalValidator.validator.Struct(s)
}

// BindAndValidate 绑定并验证请求数据
func BindAndValidate(c *gin.Context, obj interface{}) error {
	// 绑定数据
	if err := c.ShouldBindJSON(obj); err != nil {
		return fmt.Errorf("invalid request format: %w", err)
	}
	
	// 验证数据
	if err := ValidateStruct(obj); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	return nil
}