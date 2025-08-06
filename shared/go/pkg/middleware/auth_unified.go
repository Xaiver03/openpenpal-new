/**
 * 统一JWT认证中间件 - 解决8处重复实现问题
 * SOTA实现：支持权限检查、角色验证、缓存优化
 */

package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/openpenpal/shared/go/pkg/permissions"
	"github.com/openpenpal/shared/go/pkg/response"
)

// JWTConfig JWT配置
type JWTConfig struct {
	SigningKey   []byte
	TokenLookup  string // "header:Authorization" or "query:token" or "cookie:jwt"
	TokenPrefix  string // "Bearer "
	ExpireTime   time.Duration
	RefreshTime  time.Duration
	SkipperFunc  func(*gin.Context) bool
	ErrorHandler func(*gin.Context, error)
}

// DefaultJWTConfig 默认JWT配置
var DefaultJWTConfig = &JWTConfig{
	SigningKey:  []byte("openpenpal-jwt-secret-key-2025"),
	TokenLookup: "header:Authorization",
	TokenPrefix: "Bearer ",
	ExpireTime:  24 * time.Hour,
	RefreshTime: 2 * time.Hour,
	SkipperFunc: nil,
	ErrorHandler: func(c *gin.Context, err error) {
		response.ErrorWithMessage(c, http.StatusUnauthorized, "认证失败", err.Error())
	},
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID      string                    `json:"user_id"`
	Username    string                    `json:"username"`
	Email       string                    `json:"email"`
	Role        permissions.UserRole      `json:"role"`
	CourierInfo *permissions.CourierInfo  `json:"courier_info,omitempty"`
	Permissions []string                  `json:"permissions,omitempty"`
	SchoolCode  string                    `json:"school_code,omitempty"`
	jwt.RegisteredClaims
}

// JWTMiddleware JWT认证中间件工厂
func JWTMiddleware(config ...*JWTConfig) gin.HandlerFunc {
	cfg := DefaultJWTConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 跳过检查
		if cfg.SkipperFunc != nil && cfg.SkipperFunc(c) {
			c.Next()
			return
		}

		// 提取token
		token, err := extractToken(c, cfg)
		if err != nil {
			cfg.ErrorHandler(c, err)
			c.Abort()
			return
		}

		// 解析token
		claims, err := parseToken(token, cfg.SigningKey)
		if err != nil {
			cfg.ErrorHandler(c, err)
			c.Abort()
			return
		}

		// 检查token是否过期
		if time.Now().Unix() > claims.ExpiresAt.Unix() {
			cfg.ErrorHandler(c, jwt.NewValidationError("token expired", jwt.ValidationErrorExpired))
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		setUserToContext(c, claims)

		c.Next()
	}
}

// RequirePermission 权限检查中间件
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := getUserFromContext(c)
		if !exists {
			response.ErrorWithMessage(c, http.StatusUnauthorized, "用户信息不存在", "请先登录")
			c.Abort()
			return
		}

		// 使用统一权限服务检查
		if !permissions.QuickCheck(user, permission) {
			response.ErrorWithMessage(c, http.StatusForbidden, "权限不足", 
				"需要权限: "+permission)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 任一权限检查中间件
func RequireAnyPermission(requiredPermissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := getUserFromContext(c)
		if !exists {
			response.ErrorWithMessage(c, http.StatusUnauthorized, "用户信息不存在", "请先登录")
			c.Abort()
			return
		}

		// 使用统一权限服务检查
		if !permissions.DefaultService.HasAnyPermission(user, requiredPermissions) {
			response.ErrorWithMessage(c, http.StatusForbidden, "权限不足", 
				"需要以下任一权限: "+strings.Join(requiredPermissions, ", "))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions 全部权限检查中间件
func RequireAllPermissions(requiredPermissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := getUserFromContext(c)
		if !exists {
			response.ErrorWithMessage(c, http.StatusUnauthorized, "用户信息不存在", "请先登录")
			c.Abort()
			return
		}

		// 使用统一权限服务检查
		if !permissions.DefaultService.HasAllPermissions(user, requiredPermissions) {
			response.ErrorWithMessage(c, http.StatusForbidden, "权限不足", 
				"需要以下所有权限: "+strings.Join(requiredPermissions, ", "))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 角色检查中间件
func RequireRole(roles ...permissions.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := getUserFromContext(c)
		if !exists {
			response.ErrorWithMessage(c, http.StatusUnauthorized, "用户信息不存在", "请先登录")
			c.Abort()
			return
		}

		// 检查角色
		hasRole := false
		for _, role := range roles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			roleNames := make([]string, len(roles))
			for i, role := range roles {
				roleNames[i] = string(role)
			}
			response.ErrorWithMessage(c, http.StatusForbidden, "角色权限不足", 
				"需要以下角色之一: "+strings.Join(roleNames, ", "))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 管理员权限检查中间件
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := getUserFromContext(c)
		if !exists {
			response.ErrorWithMessage(c, http.StatusUnauthorized, "用户信息不存在", "请先登录")
			c.Abort()
			return
		}

		// 使用统一权限服务检查管理员权限
		if !permissions.QuickCanAccessAdmin(user) {
			response.ErrorWithMessage(c, http.StatusForbidden, "管理员权限不足", 
				"需要管理员或高级信使权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireCourier 信使权限检查中间件
func RequireCourier() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := getUserFromContext(c)
		if !exists {
			response.ErrorWithMessage(c, http.StatusUnauthorized, "用户信息不存在", "请先登录")
			c.Abort()
			return
		}

		// 使用统一权限服务检查信使权限
		if !permissions.DefaultService.IsCourier(user) {
			response.ErrorWithMessage(c, http.StatusForbidden, "信使权限不足", 
				"需要信使身份")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ================================
// 工具函数
// ================================

// extractToken 从请求中提取token
func extractToken(c *gin.Context, config *JWTConfig) (string, error) {
	var token string
	var err error

	parts := strings.Split(config.TokenLookup, ":")
	if len(parts) != 2 {
		return "", jwt.NewValidationError("invalid token lookup format", jwt.ValidationErrorMalformed)
	}

	switch parts[0] {
	case "header":
		token = c.GetHeader(parts[1])
		if token != "" && config.TokenPrefix != "" {
			if strings.HasPrefix(token, config.TokenPrefix) {
				token = token[len(config.TokenPrefix):]
			} else {
				return "", jwt.NewValidationError("invalid token prefix", jwt.ValidationErrorMalformed)
			}
		}
	case "query":
		token = c.Query(parts[1])
	case "cookie":
		token, err = c.Cookie(parts[1])
		if err != nil {
			return "", jwt.NewValidationError("token not found in cookie", jwt.ValidationErrorMalformed)
		}
	default:
		return "", jwt.NewValidationError("unsupported token lookup method", jwt.ValidationErrorMalformed)
	}

	if token == "" {
		return "", jwt.NewValidationError("token not found", jwt.ValidationErrorMalformed)
	}

	return token, nil
}

// parseToken 解析JWT token
func parseToken(tokenString string, signingKey []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorSignatureInvalid)
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.NewValidationError("invalid token claims", jwt.ValidationErrorClaimsInvalid)
}

// GenerateToken 生成JWT token
func GenerateToken(userID, username, email string, role permissions.UserRole, courierInfo *permissions.CourierInfo, config ...*JWTConfig) (string, error) {
	cfg := DefaultJWTConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	now := time.Now()
	claims := &JWTClaims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Role:        role,
		CourierInfo: courierInfo,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.ExpireTime)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "openpenpal-system",
			Subject:   userID,
		},
	}

	// 获取用户权限并添加到token中（可选，用于客户端缓存）
	user := permissions.User{
		Role:        role,
		CourierInfo: courierInfo,
	}
	claims.Permissions = permissions.DefaultService.GetUserPermissions(user)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.SigningKey)
}

// RefreshToken 刷新token
func RefreshToken(tokenString string, config ...*JWTConfig) (string, error) {
	cfg := DefaultJWTConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	claims, err := parseToken(tokenString, cfg.SigningKey)
	if err != nil {
		return "", err
	}

	// 检查是否在刷新时间窗口内
	if time.Now().Unix() > claims.ExpiresAt.Unix()-int64(cfg.RefreshTime.Seconds()) {
		return GenerateToken(claims.UserID, claims.Username, claims.Email, claims.Role, claims.CourierInfo, cfg)
	}

	return tokenString, nil
}

// setUserToContext 设置用户信息到Gin上下文
func setUserToContext(c *gin.Context, claims *JWTClaims) {
	user := permissions.User{
		Role:        claims.Role,
		CourierInfo: claims.CourierInfo,
	}

	c.Set("user", user)
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("email", claims.Email)
	c.Set("role", claims.Role)
	c.Set("school_code", claims.SchoolCode)
	c.Set("permissions", claims.Permissions)

	if claims.CourierInfo != nil {
		c.Set("courier_info", claims.CourierInfo)
		c.Set("courier_level", claims.CourierInfo.Level)
	}
}

// getUserFromContext 从Gin上下文获取用户信息
func getUserFromContext(c *gin.Context) (permissions.User, bool) {
	if userInterface, exists := c.Get("user"); exists {
		if user, ok := userInterface.(permissions.User); ok {
			return user, true
		}
	}

	// 如果直接获取失败，尝试从单独的字段构建
	roleInterface, roleExists := c.Get("role")
	if !roleExists {
		return permissions.User{}, false
	}

	role, ok := roleInterface.(permissions.UserRole)
	if !ok {
		return permissions.User{}, false
	}

	user := permissions.User{Role: role}

	// 尝试获取信使信息
	if courierInfoInterface, exists := c.Get("courier_info"); exists {
		if courierInfo, ok := courierInfoInterface.(*permissions.CourierInfo); ok && courierInfo != nil {
			user.CourierInfo = courierInfo
		}
	}

	return user, true
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}

// GetUserRole 从上下文获取用户角色
func GetUserRole(c *gin.Context) (permissions.UserRole, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	return role.(permissions.UserRole), true
}

// GetCourierInfo 从上下文获取信使信息
func GetCourierInfo(c *gin.Context) (*permissions.CourierInfo, bool) {
	courierInfo, exists := c.Get("courier_info")
	if !exists {
		return nil, false
	}
	return courierInfo.(*permissions.CourierInfo), true
}

// GetUserPermissions 从上下文获取用户权限
func GetUserPermissions(c *gin.Context) ([]string, bool) {
	permissions, exists := c.Get("permissions")
	if !exists {
		return nil, false
	}
	return permissions.([]string), true
}