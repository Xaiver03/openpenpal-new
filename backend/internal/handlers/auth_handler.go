package handlers

import (
	"fmt"
	"net/http"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"
	"openpenpal-backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器 - 专门处理认证相关请求
type AuthHandler struct {
	userService *services.UserService
	config      *config.Config
	csrfHandler *middleware.CSRFHandler
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService *services.UserService, config *config.Config) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		config:      config,
		csrfHandler: middleware.NewCSRFHandler(),
	}
}

// GetCSRFToken 获取CSRF令牌
func (h *AuthHandler) GetCSRFToken(c *gin.Context) {
	h.csrfHandler.GetCSRFToken(c)
}

// Register 用户注册 - 增强版本
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求格式不正确", err)
		return
	}

	// 验证请求字段
	if req.Username == "" {
		utils.BadRequestResponse(c, "用户名不能为空", fmt.Errorf("username is required"))
		return
	}
	if req.Password == "" {
		utils.BadRequestResponse(c, "密码不能为空", fmt.Errorf("password is required"))
		return
	}
	if req.Email == "" {
		utils.BadRequestResponse(c, "邮箱不能为空", fmt.Errorf("email is required"))
		return
	}

	// 密码强度验证
	if len(req.Password) < 8 {
		utils.BadRequestResponse(c, "密码长度至少8位", fmt.Errorf("password too short"))
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		if err.Error() == "username or email already exists" {
			utils.BadRequestResponse(c, "用户名或邮箱已存在", err)
			return
		}
		utils.BadRequestResponse(c, "注册失败", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "注册成功", gin.H{
		"user": gin.H{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"nickname":    user.Nickname,
			"role":        user.Role,
			"school_code": user.SchoolCode,
			"created_at":  user.CreatedAt,
		},
	})
}

// Login 用户登录 - 完全重构版本
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求格式不正确", err)
		return
	}

	// 验证必填字段
	if req.Username == "" {
		utils.BadRequestResponse(c, "用户名不能为空", fmt.Errorf("username is required"))
		return
	}
	if req.Password == "" {
		utils.BadRequestResponse(c, "密码不能为空", fmt.Errorf("password is required"))
		return
	}

	// 调用用户服务登录
	loginResponse, err := h.userService.Login(&req)
	if err != nil {
		// 详细的错误处理
		switch err.Error() {
		case "invalid username or password":
			utils.UnauthorizedResponse(c, "用户名或密码错误")
		case "user account is disabled":
			utils.UnauthorizedResponse(c, "账户已被禁用")
		default:
			utils.UnauthorizedResponse(c, "登录失败")
		}
		return
	}

	// 设置JWT Cookie（可选）
	if h.config.Environment == "production" {
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie(
			"auth_token",
			loginResponse.Token,
			int(24*time.Hour.Seconds()), // 24小时
			"/",
			"",
			true,  // Secure
			true,  // HttpOnly
		)
	}

	// 构建用户响应数据
	userData := gin.H{
		"id":            loginResponse.User.ID,
		"username":      loginResponse.User.Username,
		"email":         loginResponse.User.Email,
		"nickname":      loginResponse.User.Nickname,
		"role":          loginResponse.User.Role,
		"school_code":   loginResponse.User.SchoolCode,
		"is_active":     loginResponse.User.IsActive,
		"last_login_at": loginResponse.User.LastLoginAt,
		"created_at":    loginResponse.User.CreatedAt,
		"updated_at":    loginResponse.User.UpdatedAt,
	}

	// 如果用户是信使，获取信使信息
	if isCourierRole(string(loginResponse.User.Role)) {
		courierService := services.NewCourierService(h.userService.GetDB())
		courierInfo, err := courierService.GetCourierByUserID(loginResponse.User.ID)
		if err == nil && courierInfo != nil {
			userData["courierInfo"] = gin.H{
				"level":          courierInfo.Level,
				"zoneCode":       courierInfo.Zone,
				"zoneType":       getZoneTypeFromLevel(courierInfo.Level),
				"status":         courierInfo.Status,
				"points":         courierInfo.Points,
				"taskCount":      courierInfo.TaskCount,
				"completedTasks": courierInfo.TaskCount,
				"averageRating":  4.5, // Default rating as the model doesn't have rating field
				"lastActiveAt":   courierInfo.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	// 返回成功响应
	utils.SuccessResponse(c, http.StatusOK, "登录成功", gin.H{
		"token":      loginResponse.Token,
		"expires_at": loginResponse.ExpiresAt,
		"user":       userData,
	})
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "用户未认证")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, "用户不存在")
		return
	}

	// 构建用户响应数据
	userData := gin.H{
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"nickname":      user.Nickname,
		"role":          user.Role,
		"school_code":   user.SchoolCode,
		"is_active":     user.IsActive,
		"last_login_at": user.LastLoginAt,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
	}

	// 如果用户是信使，获取信使信息
	if isCourierRole(string(user.Role)) {
		courierService := services.NewCourierService(h.userService.GetDB())
		courierInfo, err := courierService.GetCourierByUserID(user.ID)
		if err == nil && courierInfo != nil {
			userData["courierInfo"] = gin.H{
				"level":          courierInfo.Level,
				"zoneCode":       courierInfo.Zone,
				"zoneType":       getZoneTypeFromLevel(courierInfo.Level),
				"status":         courierInfo.Status,
				"points":         courierInfo.Points,
				"taskCount":      courierInfo.TaskCount,
				"completedTasks": courierInfo.TaskCount,
				"averageRating":  4.5, // Default rating as the model doesn't have rating field
				"lastActiveAt":   courierInfo.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "获取用户信息成功", userData)
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 清除cookie（如果设置了）
	c.SetCookie(
		"auth_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	utils.SuccessResponse(c, http.StatusOK, "登出成功", nil)
}

// Helper functions
func isCourierRole(role string) bool {
	courierRoles := []string{
		"courier", "senior_courier", "courier_coordinator",
		"courier_level1", "courier_level2", "courier_level3", "courier_level4",
	}
	for _, r := range courierRoles {
		if role == r {
			return true
		}
	}
	return false
}

func getZoneTypeFromLevel(level int) string {
	switch level {
	case 4:
		return "city"
	case 3:
		return "school"
	case 2:
		return "zone"
	case 1:
		return "building"
	default:
		return "building"
	}
}

// RefreshToken 刷新JWT令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "用户未认证")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		utils.UnauthorizedResponse(c, "用户不存在")
		return
	}

	// 生成新的JWT令牌
	expiresAt := time.Now().Add(24 * time.Hour)
	token, err := auth.GenerateJWT(user.ID, user.Role, h.config.JWTSecret, expiresAt)
	if err != nil {
		utils.InternalServerErrorResponse(c, "令牌生成失败", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "令牌刷新成功", gin.H{
		"token":      token,
		"expires_at": expiresAt,
	})
}

// CheckTokenExpiry 检查令牌过期时间
func (h *AuthHandler) CheckTokenExpiry(c *gin.Context) {
	// 从Authorization header获取token
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		utils.UnauthorizedResponse(c, "未提供认证令牌")
		return
	}

	// 移除 "Bearer " 前缀
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// 验证并解析token
	claims, err := auth.ValidateJWT(tokenString, h.config.JWTSecret)
	if err != nil {
		utils.UnauthorizedResponse(c, "令牌无效")
		return
	}

	// 计算剩余时间
	expiryTime := claims.ExpiresAt.Time
	remainingTime := time.Until(expiryTime)

	utils.SuccessResponse(c, http.StatusOK, "令牌状态检查成功", gin.H{
		"expires_at":     expiryTime,
		"remaining_time": remainingTime.Seconds(),
		"is_expired":     remainingTime <= 0,
		"user_id":        claims.UserID,
		"role":           claims.Role,
	})
}