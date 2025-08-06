package handlers

import (
	"net/http"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		utils.BadRequestResponse(c, "Registration failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", user)
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	response, err := h.userService.Login(&req)
	if err != nil {
		utils.UnauthorizedResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// GetProfile 获取用户档案
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", user)
}

// UpdateProfile 更新用户档案
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Profile update failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", user)
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	if err := h.userService.ChangePassword(userID, &req); err != nil {
		utils.BadRequestResponse(c, "Password change failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}

// GetUserStats 获取用户统计
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	stats, err := h.userService.GetUserStats(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve user stats", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User stats retrieved successfully", stats)
}

// DeactivateAccount 停用账户
func (h *UserHandler) DeactivateAccount(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	if err := h.userService.DeactivateUser(userID); err != nil {
		utils.InternalServerErrorResponse(c, "Account deactivation failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Account deactivated successfully", nil)
}

// AdminGetUser 管理员获取用户信息
func (h *UserHandler) AdminGetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.BadRequestResponse(c, "User ID is required", nil)
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", user)
}

// AdminDeactivateUser 管理员停用用户
func (h *UserHandler) AdminDeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.BadRequestResponse(c, "User ID is required", nil)
		return
	}

	if err := h.userService.DeactivateUser(userID); err != nil {
		utils.InternalServerErrorResponse(c, "User deactivation failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User deactivated successfully", nil)
}

// AdminReactivateUser 管理员重新激活用户
func (h *UserHandler) AdminReactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.BadRequestResponse(c, "User ID is required", nil)
		return
	}

	if err := h.userService.ReactivateUser(userID); err != nil {
		utils.InternalServerErrorResponse(c, "User reactivation failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User reactivated successfully", nil)
}
