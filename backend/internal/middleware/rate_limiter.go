package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	go i.cleanupRoutine()

	return i
}

func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter
	return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		return i.AddIP(ip)
	}

	return limiter
}

func (i *IPRateLimiter) cleanupRoutine() {
	for {
		time.Sleep(time.Hour)
		i.mu.Lock()
		for ip, limiter := range i.ips {
			if limiter.Tokens() == float64(i.b) {
				delete(i.ips, ip)
			}
		}
		i.mu.Unlock()
	}
}

var (
	generalLimiter *IPRateLimiter
	authLimiter    *IPRateLimiter
)

func init() {
	// Check if we're in test mode for more lenient rate limiting
	if isTestMode() {
		// Test mode: Very lenient rate limits for integration testing
		generalLimiter = NewIPRateLimiter(rate.Every(time.Millisecond*50), 200) // 20 req/sec, burst 200
		authLimiter = NewIPRateLimiter(rate.Every(time.Millisecond*500), 100)   // 2 req/sec, burst 100
		log.Printf("[RATE_LIMITER] TEST_MODE enabled - using lenient rate limits")
	} else {
		// Production mode: More lenient rate limits for development
		generalLimiter = NewIPRateLimiter(rate.Every(time.Millisecond*100), 100) // 10 req/sec, burst 100
		authLimiter = NewIPRateLimiter(rate.Every(time.Second*10), 20)           // 6 req/min, burst 20
		log.Printf("[RATE_LIMITER] Production mode - using moderate rate limits for development")
	}
}

// isTestMode checks if we're running in test mode
func isTestMode() bool {
	env := os.Getenv("ENVIRONMENT")
	testMode := os.Getenv("TEST_MODE")
	return env == "test" || env == "testing" || testMode == "true" || testMode == "1"
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := generalLimiter.GetLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Header("Retry-After", "60")
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", generalLimiter.b))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int(limiter.Tokens())))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
			c.Abort()
			return
		}
		c.Next()
	}
}

func AuthRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := authLimiter.GetLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many authentication attempts",
				"message": "You have made too many authentication attempts. Please try again later.",
			})
			c.Header("Retry-After", "300")
			c.Abort()
			return
		}
		c.Next()
	}
}

func APIKeyRateLimitMiddleware(limiter *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "API rate limit exceeded",
				"message": "Your API key has exceeded the rate limit.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserRateLimiter 用户级别的频率限制器
type UserRateLimiter struct {
	users map[string]*rate.Limiter
	mu    *sync.RWMutex
	r     rate.Limit
	b     int
}

// NewUserRateLimiter 创建用户级别的频率限制器
func NewUserRateLimiter(r rate.Limit, b int) *UserRateLimiter {
	u := &UserRateLimiter{
		users: make(map[string]*rate.Limiter),
		mu:    &sync.RWMutex{},
		r:     r,
		b:     b,
	}

	go u.cleanupRoutine()

	return u
}

// GetLimiter 获取特定用户的限制器
func (u *UserRateLimiter) GetLimiter(userID string) *rate.Limiter {
	u.mu.RLock()
	limiter, exists := u.users[userID]
	u.mu.RUnlock()

	if !exists {
		return u.AddUser(userID)
	}

	return limiter
}

// AddUser 添加用户限制器
func (u *UserRateLimiter) AddUser(userID string) *rate.Limiter {
	u.mu.Lock()
	defer u.mu.Unlock()

	limiter := rate.NewLimiter(u.r, u.b)
	u.users[userID] = limiter
	return limiter
}

// cleanupRoutine 定期清理不活跃的用户限制器
func (u *UserRateLimiter) cleanupRoutine() {
	for {
		time.Sleep(time.Hour)
		u.mu.Lock()
		for userID, limiter := range u.users {
			// 如果限制器的令牌桶已满，说明用户不活跃
			if limiter.Tokens() == float64(u.b) {
				delete(u.users, userID)
			}
		}
		u.mu.Unlock()
	}
}

var (
	userGeneralLimiter *UserRateLimiter
	userAuthLimiter    *UserRateLimiter
)

func init() {
	// 初始化用户级别的限制器
	if isTestMode() {
		// 测试模式：宽松限制
		userGeneralLimiter = NewUserRateLimiter(rate.Every(time.Millisecond*20), 500) // 50 req/sec per user
		userAuthLimiter = NewUserRateLimiter(rate.Every(time.Second), 10)             // 1 req/sec per user
	} else {
		// 生产模式：适中限制
		userGeneralLimiter = NewUserRateLimiter(rate.Every(time.Millisecond*50), 200) // 20 req/sec per user
		userAuthLimiter = NewUserRateLimiter(rate.Every(time.Second*5), 5)            // 12 req/min per user
	}
	log.Printf("[USER_RATE_LIMITER] Initialized user-level rate limiters")
}

// UserRateLimitMiddleware 用户级别的频率限制中间件
func UserRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			// 如果没有用户ID，回退到IP限制
			RateLimitMiddleware()(c)
			return
		}

		limiter := userGeneralLimiter.GetLimiter(userID.(string))
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "请求过于频繁",
				"message": "您的请求频率超过限制，请稍后重试",
			})
			c.Header("Retry-After", "60")
			c.Header("X-RateLimit-Type", "user")
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", userGeneralLimiter.b))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int(limiter.Tokens())))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserAuthRateLimitMiddleware 用户级别的认证频率限制
func UserAuthRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于登录请求，使用用户名作为标识
		var identifier string

		if c.Request.URL.Path == "/api/v1/auth/login" {
			// 尝试从请求体获取用户名
			var loginReq struct {
				Username string `json:"username"`
			}
			if c.ShouldBindJSON(&loginReq) == nil {
				identifier = "login:" + loginReq.Username
			}
		} else {
			// 其他认证请求使用用户ID
			userID, exists := c.Get("user_id")
			if exists {
				identifier = userID.(string)
			}
		}

		// 如果无法识别用户，使用IP限制
		if identifier == "" {
			AuthRateLimitMiddleware()(c)
			return
		}

		limiter := userAuthLimiter.GetLimiter(identifier)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "认证请求过于频繁",
				"message": "您的认证请求次数过多，请5分钟后重试",
			})
			c.Header("Retry-After", "300")
			c.Header("X-RateLimit-Type", "user-auth")
			c.Abort()
			return
		}
		c.Next()
	}
}
