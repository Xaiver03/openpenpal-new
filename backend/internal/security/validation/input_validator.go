package validation

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// InputValidator provides comprehensive input validation
type InputValidator struct {
	validator        *validator.Validate
	maxFieldLength   map[string]int
	dangerousPatterns []*regexp.Regexp
	sqlPatterns      []*regexp.Regexp
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// NewInputValidator creates a new input validator with security rules
func NewInputValidator() *InputValidator {
	v := validator.New()
	
	// Register custom validators
	v.RegisterValidation("safe_string", validateSafeString)
	v.RegisterValidation("no_sql", validateNoSQL)
	v.RegisterValidation("no_script", validateNoScript)
	v.RegisterValidation("safe_email", validateSafeEmail)
	v.RegisterValidation("safe_username", validateSafeUsername)
	
	return &InputValidator{
		validator: v,
		maxFieldLength: map[string]int{
			"username":    30,
			"email":       100,
			"password":    128,
			"title":       200,
			"content":     10000,
			"description": 1000,
			"name":        100,
			"school_code": 20,
			"zone_code":   20,
			"op_code":     10,
		},
		dangerousPatterns: []*regexp.Regexp{
			regexp.MustCompile(`<script.*?>.*?</script>`),
			regexp.MustCompile(`javascript:`),
			regexp.MustCompile(`on\w+\s*=`),
			regexp.MustCompile(`<iframe.*?>`),
			regexp.MustCompile(`<object.*?>`),
			regexp.MustCompile(`<embed.*?>`),
			regexp.MustCompile(`<link.*?>`),
			regexp.MustCompile(`<meta.*?>`),
		},
		sqlPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(union.*select|select.*from|insert.*into|delete.*from|drop.*table|update.*set)`),
			regexp.MustCompile(`(?i)(exec\s*\(|execute\s|xp_cmdshell|sp_executesql)`),
			regexp.MustCompile(`(/\*.*\*/|--.*$|;.*drop|;.*delete|;.*insert|;.*update)`),
			regexp.MustCompile(`(?i)(script.*>|<.*script|javascript:|vbscript:|onload=|onerror=|onclick=)`),
		},
	}
}

// Middleware returns the Gin middleware for input validation
func (iv *InputValidator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for safe methods
		if c.Request.Method == http.MethodGet || 
		   c.Request.Method == http.MethodHead || 
		   c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Validate content length
		if c.Request.ContentLength > 10*1024*1024 { // 10MB max
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
				"code":  "PAYLOAD_TOO_LARGE",
			})
			c.Abort()
			return
		}

		// Validate query parameters
		if err := iv.validateQueryParams(c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid query parameters",
				"code":  "INVALID_QUERY_PARAMS",
				"details": err,
			})
			c.Abort()
			return
		}

		// Store validator in context for use in handlers
		c.Set("validator", iv)
		
		c.Next()
	}
}

// ValidateStruct validates a struct according to its tags
func (iv *InputValidator) ValidateStruct(s interface{}) []ValidationError {
	var errors []ValidationError
	
	if err := iv.validator.Struct(s); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: iv.getErrorMessage(e),
			})
		}
	}
	
	return errors
}

// ValidateField validates a single field value
func (iv *InputValidator) ValidateField(fieldName string, value string) error {
	// Check field length
	if maxLen, exists := iv.maxFieldLength[fieldName]; exists {
		if utf8.RuneCountInString(value) > maxLen {
			return fmt.Errorf("field %s exceeds maximum length of %d characters", fieldName, maxLen)
		}
	}

	// Check for dangerous patterns
	for _, pattern := range iv.dangerousPatterns {
		if pattern.MatchString(value) {
			return fmt.Errorf("field %s contains potentially dangerous content", fieldName)
		}
	}

	// Check for SQL injection patterns
	for _, pattern := range iv.sqlPatterns {
		if pattern.MatchString(value) {
			return fmt.Errorf("field %s contains suspicious SQL patterns", fieldName)
		}
	}

	return nil
}

// validateQueryParams validates URL query parameters
func (iv *InputValidator) validateQueryParams(c *gin.Context) error {
	for key, values := range c.Request.URL.Query() {
		// Validate key
		if err := iv.ValidateField("query_key", key); err != nil {
			return err
		}
		
		// Validate values
		for _, value := range values {
			if err := iv.ValidateField("query_value", value); err != nil {
				return err
			}
		}
	}
	return nil
}

// SanitizeString removes potentially dangerous content from a string
func (iv *InputValidator) SanitizeString(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")
	
	// Remove dangerous HTML/Script patterns
	for _, pattern := range iv.dangerousPatterns {
		sanitized = pattern.ReplaceAllString(sanitized, "")
	}
	
	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)
	
	return sanitized
}

// getErrorMessage returns a user-friendly error message for validation errors
func (iv *InputValidator) getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", e.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", e.Field(), e.Param())
	case "safe_string":
		return fmt.Sprintf("%s contains invalid characters", e.Field())
	case "no_sql":
		return fmt.Sprintf("%s contains suspicious patterns", e.Field())
	case "no_script":
		return fmt.Sprintf("%s contains script tags which are not allowed", e.Field())
	case "safe_email":
		return fmt.Sprintf("%s is not a valid email format", e.Field())
	case "safe_username":
		return fmt.Sprintf("%s can only contain letters, numbers, and underscores", e.Field())
	default:
		return fmt.Sprintf("%s is invalid", e.Field())
	}
}

// Custom validators

func validateSafeString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Allow letters, numbers, spaces, and common punctuation
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s\-_.,:;!?'"()\[\]{}@#$%^&*+=~` + "`" + `]+$`, value)
	return matched
}

func validateNoSQL(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())
	sqlKeywords := []string{
		"select", "insert", "update", "delete", "drop", "union",
		"exec", "execute", "script", "javascript", "vbscript",
	}
	
	for _, keyword := range sqlKeywords {
		if strings.Contains(value, keyword) {
			return false
		}
	}
	return true
}

func validateNoScript(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())
	return !strings.Contains(value, "<script") && !strings.Contains(value, "</script>")
}

func validateSafeEmail(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Basic email validation with additional safety checks
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		return false
	}
	
	// Check for suspicious patterns
	suspicious := []string{"<", ">", "'", "\"", ";", "javascript:", "script"}
	valueLower := strings.ToLower(value)
	for _, pattern := range suspicious {
		if strings.Contains(valueLower, pattern) {
			return false
		}
	}
	
	return true
}

func validateSafeUsername(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Only allow alphanumeric and underscore
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, value)
	return matched && len(value) >= 3 && len(value) <= 30
}

// ValidateJSON validates JSON request body
func (iv *InputValidator) ValidateJSON(c *gin.Context, v interface{}) error {
	if err := c.ShouldBindJSON(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	
	if errors := iv.ValidateStruct(v); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Validation failed",
			"code":  "VALIDATION_ERROR",
			"errors": errors,
		})
		return fmt.Errorf("validation failed")
	}
	
	return nil
}