package handlers

import (
	"time"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"
	"openpenpal-backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

type RefreshTokenHandler struct {
	userService *services.UserService
	jwtSecret   string
}

// NewRefreshTokenHandler 创建新的RefreshTokenHandler实例
func NewRefreshTokenHandler(userService *services.UserService, jwtSecret string) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

// RefreshToken 刷新访问令牌
func (h *RefreshTokenHandler) RefreshToken(c *gin.Context) {
	// 获取当前用户ID和角色（从认证中间件设置的上下文中）
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "User not authenticated", "success": false})
		return
	}

	userRole, exists := middleware.GetUserRole(c)
	if !exists {
		userRole = "user" // 默认角色
	}

	// 生成新的访问令牌（使用固定24小时过期时间）
	expiresAt := time.Now().Add(24 * time.Hour)

	newToken, err := auth.GenerateJWT(userID, models.UserRole(userRole), h.jwtSecret, expiresAt)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate new token", "success": false})
		return
	}

	// 返回新的访问令牌
	c.JSON(200, gin.H{
		"success": true,
		"message": "Token refreshed successfully",
		"data": gin.H{
			"token":     newToken,
			"expiresAt": expiresAt.Format(time.RFC3339),
		},
	})
}

// CheckTokenExpiry 检查令牌过期时间
func (h *RefreshTokenHandler) CheckTokenExpiry(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// 从Authorization header获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 7 {
		utils.BadRequestResponse(c, "No token provided", nil)
		return
	}

	token := authHeader[7:] // 移除 "Bearer " 前缀

	// 解析令牌获取过期时间
	claims, err := auth.ValidateJWT(token, h.jwtSecret)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid token")
		return
	}

	// 计算剩余时间
	now := time.Now()
	expiresAt := claims.ExpiresAt.Time
	expiresIn := expiresAt.Sub(now)

	// 判断是否已过期
	isExpired := expiresIn <= 0

	// 判断是否应该刷新（剩余时间少于2小时）
	shouldRefresh := expiresIn > 0 && expiresIn <= 2*time.Hour

	utils.SuccessResponse(c, 200, "Token expiry checked successfully", gin.H{
		"userId":        userID,
		"isExpired":     isExpired,
		"expiresAt":     expiresAt.Format(time.RFC3339),
		"expiresIn":     int(expiresIn.Seconds()),
		"shouldRefresh": shouldRefresh,
	})
}
