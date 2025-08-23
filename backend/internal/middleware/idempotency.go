package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"openpenpal-backend/pkg/cache"
)

// IdempotencyConfig 幂等配置
type IdempotencyConfig struct {
	CacheManager   *cache.EnhancedCacheManager // 使用增强缓存管理器
	RedisClient    *redis.Client               // 保持向后兼容
	TTL            time.Duration               // 幂等键过期时间
	SkipPaths      []string                    // 跳过幂等检查的路径
	AllowedMethods []string                    // 需要幂等检查的HTTP方法
}

// IdempotencyMiddleware 幂等处理中间件
func IdempotencyMiddleware(config IdempotencyConfig) gin.HandlerFunc {
	// 默认配置
	if config.TTL == 0 {
		config.TTL = 24 * time.Hour
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"POST", "PUT", "PATCH"}
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		// 检查是否需要幂等处理
		if !shouldCheckIdempotency(c, config) {
			c.Next()
			return
		}

		// 获取幂等键
		idempotencyKey := getIdempotencyKey(c)
		if idempotencyKey == "" {
			// 没有幂等键，生成一个基于请求内容的键
			idempotencyKey = generateIdempotencyKey(c)
		}

		ctx := context.Background()
		cacheKey := idempotencyKey

		// 检查是否已存在相同请求
		var cachedResponse StoredResponse
		err := config.CacheManager.Get(ctx, cacheKey, &cachedResponse)
		if err == nil {
			// 返回缓存的响应
			c.Header("X-Idempotency-Key", idempotencyKey)
			c.Header("X-Idempotency-Replayed", "true")
			
			c.JSON(cachedResponse.StatusCode, cachedResponse.Body)
			c.Abort()
			return
		}

		// 使用响应写入器捕获响应
		responseWriter := &responseCapture{
			ResponseWriter: c.Writer,
			statusCode:     200,
			body:          make([]byte, 0),
		}
		c.Writer = responseWriter

		// 设置请求处理标记
		c.Set("idempotency_key", idempotencyKey)
		c.Header("X-Idempotency-Key", idempotencyKey)

		// 继续处理请求
		c.Next()

		// 如果请求成功，缓存响应
		if responseWriter.statusCode >= 200 && responseWriter.statusCode < 300 {
			response := StoredResponse{
				StatusCode: responseWriter.statusCode,
				Body:       responseWriter.body,
				Timestamp:  time.Now(),
			}

			// 使用增强缓存管理器存储，自动应用TTL jitter
			config.CacheManager.Set(ctx, cacheKey, response, config.TTL)
		}
	})
}

// shouldCheckIdempotency 检查是否需要幂等处理
func shouldCheckIdempotency(c *gin.Context, config IdempotencyConfig) bool {
	// 检查HTTP方法
	methodAllowed := false
	for _, method := range config.AllowedMethods {
		if c.Request.Method == method {
			methodAllowed = true
			break
		}
	}
	if !methodAllowed {
		return false
	}

	// 检查跳过路径
	for _, path := range config.SkipPaths {
		if strings.HasPrefix(c.Request.URL.Path, path) {
			return false
		}
	}

	return true
}

// getIdempotencyKey 获取客户端提供的幂等键
func getIdempotencyKey(c *gin.Context) string {
	// 从Header获取
	if key := c.GetHeader("Idempotency-Key"); key != "" {
		return key
	}
	if key := c.GetHeader("X-Idempotency-Key"); key != "" {
		return key
	}
	
	// 从查询参数获取
	if key := c.Query("idempotency_key"); key != "" {
		return key
	}

	return ""
}

// generateIdempotencyKey 基于请求内容生成幂等键
func generateIdempotencyKey(c *gin.Context) string {
	// 构建幂等键的组成部分
	var parts []string
	
	// 用户ID (如果已认证)
	if userID, exists := c.Get("user_id"); exists {
		parts = append(parts, fmt.Sprintf("user:%v", userID))
	}
	
	// HTTP方法和路径
	parts = append(parts, c.Request.Method, c.Request.URL.Path)
	
	// 查询参数 (排除幂等键参数)
	query := c.Request.URL.Query()
	query.Del("idempotency_key")
	if len(query) > 0 {
		parts = append(parts, query.Encode())
	}
	
	// 请求体内容 (对于小请求体)
	if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength < 1024*10 { // 10KB
		if body, err := io.ReadAll(c.Request.Body); err == nil {
			parts = append(parts, string(body))
			// 重新设置请求体供后续中间件使用
			c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
		}
	}
	
	// 生成SHA256哈希
	content := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])[:32] // 取前32位
}

// responseCapture 响应捕获器
type responseCapture struct {
	gin.ResponseWriter
	statusCode int
	body       []byte
}

func (w *responseCapture) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseCapture) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

// StoredResponse 存储的响应
type StoredResponse struct {
	StatusCode int         `json:"status_code"`
	Body       []byte      `json:"body"`
	Timestamp  time.Time   `json:"timestamp"`
}

// serializeResponse 序列化响应
func serializeResponse(response StoredResponse) (string, error) {
	// 简单的序列化，实际可以使用JSON
	return fmt.Sprintf("%d|%s|%d", 
		response.StatusCode, 
		string(response.Body), 
		response.Timestamp.Unix()), nil
}

// parseStoredResponse 解析存储的响应
func parseStoredResponse(data string) (*StoredResponse, error) {
	// 简单的反序列化，实际可以使用JSON
	parts := strings.SplitN(data, "|", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid stored response format")
	}
	
	var statusCode int
	var timestamp int64
	
	if _, err := fmt.Sscanf(parts[0], "%d", &statusCode); err != nil {
		return nil, err
	}
	
	if _, err := fmt.Sscanf(parts[2], "%d", &timestamp); err != nil {
		return nil, err
	}
	
	return &StoredResponse{
		StatusCode: statusCode,
		Body:       []byte(parts[1]),
		Timestamp:  time.Unix(timestamp, 0),
	}, nil
}

// GetIdempotencyKey 从上下文获取幂等键
func GetIdempotencyKey(c *gin.Context) string {
	if key, exists := c.Get("idempotency_key"); exists {
		return key.(string)
	}
	return ""
}