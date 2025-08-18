package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 启用共享响应包集成
func init() {
	utils.BeginResponseMigration()
}

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
		utils.SharedBadRequestResponse(c, "Invalid request data", err)
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		utils.SharedBadRequestResponse(c, "Registration failed", err)
		return
	}

	utils.SharedSuccessResponse(c, http.StatusCreated, "User registered successfully", user)
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

// UploadAvatar 上传用户头像
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.BadRequestResponse(c, "No file uploaded", err)
		return
	}
	defer file.Close()

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		utils.BadRequestResponse(c, "Invalid file type. Only JPG, PNG, GIF and WEBP are allowed", nil)
		return
	}

	// 限制文件大小 (5MB)
	if header.Size > 5*1024*1024 {
		utils.BadRequestResponse(c, "File size too large. Maximum 5MB allowed", nil)
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%s_%s%s", userID, uuid.New().String(), ext)

	// 保存文件
	avatarURL, err := h.userService.SaveAvatar(userID, file, filename)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to save avatar", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Avatar uploaded successfully", gin.H{
		"avatar_url": avatarURL,
	})
}

// RemoveAvatar 移除用户头像
func (h *UserHandler) RemoveAvatar(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	if err := h.userService.RemoveAvatar(userID); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to remove avatar", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Avatar removed successfully", nil)
}
