package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TracingConfig 追踪配置
type TracingConfig struct {
	ServiceName     string
	EnableLogging   bool
	EnableMetrics   bool
	LogRequestBody  bool
	LogResponseBody bool
	MaxBodySize     int64 // 最大记录的请求/响应体大小
}

// RequestTracingMiddleware 增强的请求追踪中间件
func RequestTracingMiddleware(config TracingConfig) gin.HandlerFunc {
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 1024 * 10 // 默认10KB
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		
		// 获取或生成请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		// 设置到上下文
		c.Set("request_id", requestID)
		c.Set("service_name", config.ServiceName)
		c.Header("X-Request-ID", requestID)
		
		// 记录请求信息
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil && c.Request.ContentLength <= config.MaxBodySize {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应写入器以捕获响应
		responseWriter := &tracingResponseWriter{
			ResponseWriter: c.Writer,
			body:          bytes.NewBuffer(nil),
			maxBodySize:   config.MaxBodySize,
		}
		c.Writer = responseWriter

		// 记录开始日志
		if config.EnableLogging {
			logRequestStart(c, requestID, requestBody, config)
		}

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)
		
		// 记录结束日志
		if config.EnableLogging {
			logRequestEnd(c, requestID, responseWriter, duration, config)
		}

		// 记录指标
		if config.EnableMetrics {
			recordMetrics(c, duration, config)
		}
	})
}

// tracingResponseWriter 追踪响应写入器
type tracingResponseWriter struct {
	gin.ResponseWriter
	body        *bytes.Buffer
	maxBodySize int64
	bodySize    int64
}

func (w *tracingResponseWriter) Write(data []byte) (int, error) {
	// 限制记录的响应体大小
	if w.bodySize+int64(len(data)) <= w.maxBodySize {
		w.body.Write(data)
	}
	w.bodySize += int64(len(data))
	
	return w.ResponseWriter.Write(data)
}

// logRequestStart 记录请求开始日志
func logRequestStart(c *gin.Context, requestID string, requestBody []byte, config TracingConfig) {
	logData := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"request_id":   requestID,
		"service":      config.ServiceName,
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"remote_addr":  c.ClientIP(),
		"user_agent":   c.Request.UserAgent(),
		"content_type": c.Request.Header.Get("Content-Type"),
		"event":        "request_start",
	}

	// 添加用户信息（如果已认证）
	if userID, exists := c.Get("user_id"); exists {
		logData["user_id"] = userID
	}

	// 添加请求体
	if len(requestBody) > 0 && config.LogRequestBody {
		// 尝试解析JSON，失败则记录原始字符串
		var jsonBody interface{}
		if err := json.Unmarshal(requestBody, &jsonBody); err == nil {
			logData["request_body"] = jsonBody
		} else {
			logData["request_body"] = string(requestBody)
		}
	}

	logJSON(logData)
}

// logRequestEnd 记录请求结束日志  
func logRequestEnd(c *gin.Context, requestID string, responseWriter *tracingResponseWriter, duration time.Duration, config TracingConfig) {
	logData := map[string]interface{}{
		"timestamp":       time.Now().Format(time.RFC3339),
		"request_id":      requestID,
		"service":         config.ServiceName,
		"method":          c.Request.Method,
		"path":            c.Request.URL.Path,
		"status_code":     responseWriter.Status(),
		"response_size":   responseWriter.bodySize,
		"duration_ms":     float64(duration.Nanoseconds()) / 1e6,
		"event":           "request_end",
	}

	// 添加错误信息
	if len(c.Errors) > 0 {
		errors := make([]string, len(c.Errors))
		for i, err := range c.Errors {
			errors[i] = err.Error()
		}
		logData["errors"] = errors
	}

	// 添加响应体
	if responseWriter.body.Len() > 0 && config.LogResponseBody {
		responseBody := responseWriter.body.String()
		
		// 尝试解析JSON
		var jsonBody interface{}
		if err := json.Unmarshal([]byte(responseBody), &jsonBody); err == nil {
			logData["response_body"] = jsonBody
		} else {
			logData["response_body"] = responseBody
		}
	}

	logJSON(logData)
}

// recordMetrics 记录指标
func recordMetrics(c *gin.Context, duration time.Duration, config TracingConfig) {
	// 这里可以集成到Prometheus或其他指标系统
	// 暂时使用日志记录指标
	logData := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"request_id":  c.GetString("request_id"),
		"service":     config.ServiceName,
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"status_code": c.Writer.Status(),
		"duration_ms": float64(duration.Nanoseconds()) / 1e6,
		"event":       "metrics",
	}
	
	logJSON(logData)
}

// logJSON 输出JSON格式日志
func logJSON(data map[string]interface{}) {
	if jsonBytes, err := json.Marshal(data); err == nil {
		log.Println(string(jsonBytes))
	}
}

// PropagateRequestID 传播请求ID到下游服务
func PropagateRequestID(req *http.Request, c *gin.Context) {
	if requestID := c.GetString("request_id"); requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}
}

// CreateHTTPClientWithTracing 创建带追踪的HTTP客户端
func CreateHTTPClientWithTracing(serviceName string) *http.Client {
	transport := &tracingTransport{
		base:        http.DefaultTransport,
		serviceName: serviceName,
	}
	
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// tracingTransport 追踪传输层
type tracingTransport struct {
	base        http.RoundTripper
	serviceName string
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	
	// 记录出站请求
	logData := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"service":      t.serviceName,
		"event":        "outbound_request",
		"method":       req.Method,
		"url":          req.URL.String(),
		"request_id":   req.Header.Get("X-Request-ID"),
	}
	logJSON(logData)
	
	// 执行请求
	resp, err := t.base.RoundTrip(req)
	duration := time.Since(start)
	
	// 记录响应
	logData = map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"service":     t.serviceName,
		"event":       "outbound_response",
		"method":      req.Method,
		"url":         req.URL.String(),
		"request_id":  req.Header.Get("X-Request-ID"),
		"duration_ms": float64(duration.Nanoseconds()) / 1e6,
	}
	
	if resp != nil {
		logData["status_code"] = resp.StatusCode
	}
	if err != nil {
		logData["error"] = err.Error()
	}
	
	logJSON(logData)
	
	return resp, err
}

// GetRequestID 从上下文获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// GetRequestContext 获取带请求ID的上下文
func GetRequestContext(c *gin.Context) context.Context {
	ctx := context.Background()
	if requestID := GetRequestID(c); requestID != "" {
		ctx = context.WithValue(ctx, "request_id", requestID)
	}
	return ctx
}

// ExtractRequestIDFromContext 从上下文提取请求ID
func ExtractRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// LogWithRequestID 带请求ID的日志记录
func LogWithRequestID(c *gin.Context, level string, message string, fields map[string]interface{}) {
	logData := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"level":      strings.ToUpper(level),
		"message":    message,
		"request_id": GetRequestID(c),
		"service":    c.GetString("service_name"),
	}
	
	// 合并额外字段
	for k, v := range fields {
		logData[k] = v
	}
	
	logJSON(logData)
}