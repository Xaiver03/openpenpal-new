package handlers

import (
	"net/http"
	"net/url"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type PrivacyHandler struct {
	privacyService *services.PrivacyService
}

func NewPrivacyHandler(privacyService *services.PrivacyService) *PrivacyHandler {
	return &PrivacyHandler{
		privacyService: privacyService,
	}
}

// GetPrivacySettings 获取当前用户的隐私设置
// @Summary 获取隐私设置
// @Description 获取当前用户的隐私设置
// @Tags Privacy
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=models.PrivacySettings}
// @Router /api/v1/privacy/settings [get]
func (h *PrivacyHandler) GetPrivacySettings(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	settings, err := h.privacyService.GetPrivacySettings(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get privacy settings", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", settings)
}

// UpdatePrivacySettings 更新隐私设置
// @Summary 更新隐私设置
// @Description 更新当前用户的隐私设置
// @Tags Privacy
// @Accept json
// @Produce json
// @Param request body models.UpdatePrivacySettingsRequest true "更新请求"
// @Success 200 {object} utils.Response{data=models.PrivacySettings}
// @Router /api/v1/privacy/settings [put]
func (h *PrivacyHandler) UpdatePrivacySettings(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req models.UpdatePrivacySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	settings, err := h.privacyService.UpdatePrivacySettings(userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update privacy settings", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Privacy settings updated successfully", settings)
}

// ResetPrivacySettings 重置隐私设置为默认值
// @Summary 重置隐私设置
// @Description 重置当前用户的隐私设置为默认值
// @Tags Privacy
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=models.PrivacySettings}
// @Router /api/v1/privacy/settings/reset [post]
func (h *PrivacyHandler) ResetPrivacySettings(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	settings, err := h.privacyService.ResetPrivacySettings(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to reset privacy settings", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Privacy settings reset to defaults", settings)
}

// CheckPrivacy 检查隐私权限
// @Summary 检查隐私权限
// @Description 检查当前用户是否可以访问目标用户的特定内容
// @Tags Privacy
// @Accept json
// @Produce json
// @Param user_id path string true "目标用户ID"
// @Param action query string false "检查的行为" default(view_profile)
// @Success 200 {object} utils.Response{data=models.PrivacyCheckResult}
// @Router /api/v1/privacy/check/{user_id} [get]
func (h *PrivacyHandler) CheckPrivacy(c *gin.Context) {
	viewerID := c.GetString("user_id")
	if viewerID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Target user ID is required", nil)
		return
	}

	action := c.DefaultQuery("action", "view_profile")

	result, err := h.privacyService.CheckPrivacy(viewerID, targetUserID, action)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to check privacy", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", result)
}

// BatchCheckPrivacy 批量检查隐私权限
// @Summary 批量检查隐私权限
// @Description 批量检查当前用户对目标用户的多种行为权限
// @Tags Privacy
// @Accept json
// @Produce json
// @Param user_id path string true "目标用户ID"
// @Param request body models.BatchPrivacyCheckRequest true "批量检查请求"
// @Success 200 {object} utils.Response{data=map[string]models.PrivacyCheckResult}
// @Router /api/v1/privacy/check/{user_id}/batch [post]
func (h *PrivacyHandler) BatchCheckPrivacy(c *gin.Context) {
	viewerID := c.GetString("user_id")
	if viewerID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Target user ID is required", nil)
		return
	}

	var req models.BatchPrivacyCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	results, err := h.privacyService.BatchCheckPrivacy(viewerID, targetUserID, req.Actions)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to batch check privacy", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", results)
}

// BlockUser 屏蔽用户
// @Summary 屏蔽用户
// @Description 将指定用户添加到屏蔽列表
// @Tags Privacy
// @Accept json
// @Produce json
// @Param request body models.BlockUserRequest true "屏蔽用户请求"
// @Success 200 {object} utils.Response{data=object{success=bool}}
// @Router /api/v1/privacy/block [post]
func (h *PrivacyHandler) BlockUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req models.BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if req.UserID == userID {
		utils.ErrorResponse(c, http.StatusBadRequest, "Cannot block yourself", nil)
		return
	}

	err := h.privacyService.BlockUser(userID, req.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to block user", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User blocked successfully", map[string]interface{}{
		"success": true,
	})
}

// UnblockUser 取消屏蔽用户
// @Summary 取消屏蔽用户
// @Description 从屏蔽列表中移除指定用户
// @Tags Privacy
// @Accept json
// @Produce json
// @Param user_id path string true "要取消屏蔽的用户ID"
// @Success 200 {object} utils.Response{data=object{success=bool}}
// @Router /api/v1/privacy/block/{user_id} [delete]
func (h *PrivacyHandler) UnblockUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Target user ID is required", nil)
		return
	}

	err := h.privacyService.UnblockUser(userID, targetUserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to unblock user", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User unblocked successfully", map[string]interface{}{
		"success": true,
	})
}

// MuteUser 静音用户
// @Summary 静音用户
// @Description 将指定用户添加到静音列表
// @Tags Privacy
// @Accept json
// @Produce json
// @Param request body models.MuteUserRequest true "静音用户请求"
// @Success 200 {object} utils.Response{data=object{success=bool}}
// @Router /api/v1/privacy/mute [post]
func (h *PrivacyHandler) MuteUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req models.MuteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	if req.UserID == userID {
		utils.ErrorResponse(c, http.StatusBadRequest, "Cannot mute yourself", nil)
		return
	}

	err := h.privacyService.MuteUser(userID, req.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to mute user", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User muted successfully", map[string]interface{}{
		"success": true,
	})
}

// UnmuteUser 取消静音用户
// @Summary 取消静音用户
// @Description 从静音列表中移除指定用户
// @Tags Privacy
// @Accept json
// @Produce json
// @Param user_id path string true "要取消静音的用户ID"
// @Success 200 {object} utils.Response{data=object{success=bool}}
// @Router /api/v1/privacy/mute/{user_id} [delete]
func (h *PrivacyHandler) UnmuteUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Target user ID is required", nil)
		return
	}

	err := h.privacyService.UnmuteUser(userID, targetUserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to unmute user", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User unmuted successfully", map[string]interface{}{
		"success": true,
	})
}

// GetBlockedUsers 获取屏蔽用户列表
// @Summary 获取屏蔽用户列表
// @Description 获取当前用户的屏蔽用户列表
// @Tags Privacy
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=object{blocked_users=[]string}}
// @Router /api/v1/privacy/blocked [get]
func (h *PrivacyHandler) GetBlockedUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	blockedUsers, err := h.privacyService.GetBlockedUsers(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get blocked users", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", map[string]interface{}{
		"blocked_users": blockedUsers,
	})
}

// GetMutedUsers 获取静音用户列表
// @Summary 获取静音用户列表
// @Description 获取当前用户的静音用户列表
// @Tags Privacy
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=object{muted_users=[]string}}
// @Router /api/v1/privacy/muted [get]
func (h *PrivacyHandler) GetMutedUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	mutedUsers, err := h.privacyService.GetMutedUsers(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get muted users", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", map[string]interface{}{
		"muted_users": mutedUsers,
	})
}

// AddBlockedKeyword 添加屏蔽关键词
// @Summary 添加屏蔽关键词
// @Description 添加关键词到屏蔽列表
// @Tags Privacy
// @Accept json
// @Produce json
// @Param request body models.AddBlockedKeywordRequest true "添加屏蔽关键词请求"
// @Success 200 {object} utils.Response{data=object{success=bool}}
// @Router /api/v1/privacy/keywords/block [post]
func (h *PrivacyHandler) AddBlockedKeyword(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req models.AddBlockedKeywordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	err := h.privacyService.AddBlockedKeyword(userID, req.Keyword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to add blocked keyword", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Blocked keyword added successfully", map[string]interface{}{
		"success": true,
	})
}

// RemoveBlockedKeyword 移除屏蔽关键词
// @Summary 移除屏蔽关键词
// @Description 从屏蔽列表中移除指定关键词
// @Tags Privacy
// @Accept json
// @Produce json
// @Param keyword path string true "要移除的关键词"
// @Success 200 {object} utils.Response{data=object{success=bool}}
// @Router /api/v1/privacy/keywords/block/{keyword} [delete]
func (h *PrivacyHandler) RemoveBlockedKeyword(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	keyword := c.Param("keyword")
	if keyword == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Keyword is required", nil)
		return
	}

	// URL decode the keyword
	decodedKeyword, err := url.QueryUnescape(keyword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid keyword format", err)
		return
	}

	err = h.privacyService.RemoveBlockedKeyword(userID, decodedKeyword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove blocked keyword", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Blocked keyword removed successfully", map[string]interface{}{
		"success": true,
	})
}

// GetBlockedKeywords 获取屏蔽关键词列表
// @Summary 获取屏蔽关键词列表
// @Description 获取当前用户的屏蔽关键词列表
// @Tags Privacy
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=object{keywords=[]string}}
// @Router /api/v1/privacy/keywords/blocked [get]
func (h *PrivacyHandler) GetBlockedKeywords(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	keywords, err := h.privacyService.GetBlockedKeywords(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get blocked keywords", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", map[string]interface{}{
		"keywords": keywords,
	})
}