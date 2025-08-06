# 中间件安全性改进建议

## 高优先级改进

### 1. CSP策略收紧
**当前问题**: CSP允许`unsafe-inline`和`unsafe-eval`
**建议**: 
```go
// 生产环境严格CSP
c.Header("Content-Security-Policy", 
    "default-src 'self'; "+
    "script-src 'self' 'nonce-{nonce}'; "+
    "style-src 'self' 'nonce-{nonce}'; "+
    "img-src 'self' data: https:; "+
    "connect-src 'self' wss:;")
```

### 2. 用户级频率限制
**当前问题**: 只有IP级别限制，同一NAT后用户可能被误限制
**建议**:
```go
func UserRateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("user_id")
        if exists {
            // 基于用户ID的限制
            limiter := getUserLimiter(userID.(string))
            if !limiter.Allow() {
                c.JSON(429, gin.H{"error": "用户请求过于频繁"})
                c.Abort()
                return
            }
        }
        c.Next()
    }
}
```

### 3. Token黑名单机制
**当前问题**: 注销的Token仍然有效直到过期
**建议**:
```go
type TokenBlacklist struct {
    tokens map[string]time.Time
    mu     sync.RWMutex
}

func (tb *TokenBlacklist) Add(tokenID string, expiry time.Time) {
    tb.mu.Lock()
    tb.tokens[tokenID] = expiry
    tb.mu.Unlock()
}

func (tb *TokenBlacklist) IsBlacklisted(tokenID string) bool {
    tb.mu.RLock()
    expiry, exists := tb.tokens[tokenID]
    tb.mu.RUnlock()
    
    if !exists {
        return false
    }
    
    if time.Now().After(expiry) {
        tb.Delete(tokenID)
        return false
    }
    
    return true
}
```

## 中优先级改进

### 4. 请求大小限制
```go
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.ContentLength > maxSize {
            c.JSON(413, gin.H{"error": "请求体过大"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 5. IP白名单/黑名单
```go
func IPFilterMiddleware(allowList, blockList []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        // 检查黑名单
        for _, blockedIP := range blockList {
            if clientIP == blockedIP {
                c.JSON(403, gin.H{"error": "访问被拒绝"})
                c.Abort()
                return
            }
        }
        
        // 如果有白名单，检查是否在白名单中
        if len(allowList) > 0 {
            allowed := false
            for _, allowedIP := range allowList {
                if clientIP == allowedIP {
                    allowed = true
                    break
                }
            }
            if !allowed {
                c.JSON(403, gin.H{"error": "访问被拒绝"})
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}
```

### 6. 审计日志中间件
```go
func AuditLogMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 记录请求信息
        userID, _ := c.Get("user_id")
        requestID := c.GetString("request_id")
        
        c.Next()
        
        // 记录响应信息
        duration := time.Since(start)
        status := c.Writer.Status()
        
        // 记录审计日志
        log.Printf("[AUDIT] RequestID=%s UserID=%v Method=%s Path=%s Status=%d Duration=%v IP=%s",
            requestID, userID, c.Request.Method, c.Request.URL.Path, 
            status, duration, c.ClientIP())
    }
}
```

## 低优先级改进

### 7. 地理位置限制
```go
func GeoLocationMiddleware(allowedCountries []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        country := getCountryByIP(clientIP) // 需要GeoIP数据库
        
        allowed := false
        for _, allowedCountry := range allowedCountries {
            if country == allowedCountry {
                allowed = true
                break
            }
        }
        
        if !allowed {
            c.JSON(403, gin.H{"error": "地理位置访问限制"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 8. 设备指纹验证
```go
func DeviceFingerprintMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userAgent := c.GetHeader("User-Agent")
        fingerprint := c.GetHeader("X-Device-Fingerprint")
        
        if fingerprint != "" {
            // 验证设备指纹的有效性
            if !isValidFingerprint(userAgent, fingerprint) {
                c.JSON(403, gin.H{"error": "设备验证失败"})
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}
```

## 实施建议

### 阶段1（立即实施）：
1. 收紧CSP策略
2. 实现Token黑名单
3. 添加用户级频率限制

### 阶段2（1-2周内）：
1. 添加请求大小限制
2. 实现审计日志
3. 优化错误处理

### 阶段3（1个月内）：
1. IP过滤功能
2. 地理位置限制（如需要）
3. 设备指纹验证（高安全需求场景）

## 配置示例

```go
// 在main.go中的中间件配置顺序很重要
func setupMiddleware(r *gin.Engine, config *config.Config, db *gorm.DB) {
    // 1. 基础中间件
    r.Use(RequestIDMiddleware())
    r.Use(SecurityHeadersMiddleware())
    r.Use(CORSMiddleware())
    
    // 2. 安全中间件
    r.Use(RequestSizeLimitMiddleware(10 * 1024 * 1024)) // 10MB
    r.Use(RateLimitMiddleware())
    
    // 3. 认证中间件（按需）
    authGroup := r.Group("/api/v1")
    authGroup.Use(OptimizedAuthMiddleware(config, db))
    authGroup.Use(AuditLogMiddleware())
    
    // 4. 性能监控
    r.Use(MetricsMiddleware())
}
```