package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditEntry 审计日志条目
type AuditEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id"`
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	ClientIP    string                 `json:"client_ip"`
	UserAgent   string                 `json:"user_agent"`
	StatusCode  int                    `json:"status_code"`
	Duration    string                 `json:"duration"`
	RequestBody interface{}            `json:"request_body,omitempty"`
	Response    interface{}            `json:"response,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// AuditLogMiddleware 审计日志中间件
func AuditLogMiddleware() gin.HandlerFunc {
	// 需要审计的路径
	auditPaths := map[string]bool{
		"/api/v1/auth/login":      true,
		"/api/v1/auth/logout":     true,
		"/api/v1/auth/register":   true,
		"/api/v1/users":           true,
		"/api/v1/letters/send":    true,
		"/api/v1/admin":           true,
		"/api/v1/courier/assign":  true,
		"/api/v1/courier/promote": true,
		"/api/v1/settings":        true,
	}

	// 敏感字段，需要脱敏
	sensitiveFields := map[string]bool{
		"password":      true,
		"new_password":  true,
		"old_password":  true,
		"token":         true,
		"access_token":  true,
		"refresh_token": true,
		"api_key":       true,
		"secret":        true,
	}

	return func(c *gin.Context) {
		// 检查是否需要审计
		needAudit := false
		for path := range auditPaths {
			if c.Request.URL.Path == path || c.GetBool("force_audit") {
				needAudit = true
				break
			}
		}

		if !needAudit {
			c.Next()
			return
		}

		// 记录开始时间
		start := time.Now()

		// 生成请求ID（如果没有）
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}

		// 读取请求体（如果是JSON）
		var requestBody interface{}
		if c.Request.Method != "GET" && c.ContentType() == "application/json" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			var body map[string]interface{}
			if json.Unmarshal(bodyBytes, &body) == nil {
				// 脱敏处理
				requestBody = sanitizeData(body, sensitiveFields)
			}
		}

		// 创建响应记录器
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 继续处理请求
		c.Next()

		// 计算请求处理时间
		duration := time.Since(start)

		// 创建审计条目
		entry := AuditEntry{
			Timestamp:   start,
			RequestID:   requestID,
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			ClientIP:    c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			StatusCode:  c.Writer.Status(),
			Duration:    duration.String(),
			RequestBody: requestBody,
		}

		// 添加用户信息
		if userID, exists := c.Get("user_id"); exists {
			entry.UserID = userID.(string)
		}
		if username, exists := c.Get("username"); exists {
			entry.Username = username.(string)
		}

		// 添加自定义操作
		if action, exists := c.Get("audit_action"); exists {
			entry.Action = action.(string)
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// 解析响应（仅记录错误响应）
		if entry.StatusCode >= 400 {
			var response interface{}
			if json.Unmarshal(blw.body.Bytes(), &response) == nil {
				entry.Response = response
			}
		}

		// 添加额外信息
		if extra, exists := c.Get("audit_extra"); exists {
			if extraMap, ok := extra.(map[string]interface{}); ok {
				entry.Extra = extraMap
			}
		}

		// 记录审计日志
		logAuditEntry(entry)
	}
}

// bodyLogWriter 用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// sanitizeData 数据脱敏
func sanitizeData(data map[string]interface{}, sensitiveFields map[string]bool) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		if sensitiveFields[key] {
			result[key] = "***REDACTED***"
		} else {
			switch v := value.(type) {
			case map[string]interface{}:
				result[key] = sanitizeData(v, sensitiveFields)
			default:
				result[key] = value
			}
		}
	}
	return result
}

// logAuditEntry 记录审计日志
func logAuditEntry(entry AuditEntry) {
	// 这里可以根据需要将日志写入数据库、文件或发送到日志服务
	// 当前实现仅打印到标准输出
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("[AUDIT_ERROR] Failed to marshal audit entry: %v", err)
		return
	}

	log.Printf("[AUDIT] %s", string(jsonData))

	// Write audit log to database using the AuditLog model
	// db.Create(&AuditLog{
	//     UserID: entry.UserID,
	//     Action: entry.Action,
	//     ...
	// })
}

// ForceAudit 强制为特定请求启用审计
func ForceAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("force_audit", true)
		c.Next()
	}
}
