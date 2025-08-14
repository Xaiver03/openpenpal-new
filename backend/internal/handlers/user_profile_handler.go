package handlers

import (
	"net/http"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserProfileHandler struct {
	profileService *services.UserProfileService
}

func NewUserProfileHandler(db *gorm.DB) *UserProfileHandler {
	return &UserProfileHandler{
		profileService: services.NewUserProfileService(db),
	}
}

// GetUserProfile 获取用户档案
// GET /api/users/:username/profile
func (h *UserProfileHandler) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.BadRequestResponse(c, "用户名不能为空", nil)
		return
	}

	// 获取当前登录用户ID（如果已登录）
	requestingUserID := ""
	if userClaims, exists := c.Get("userClaims"); exists {
		if claims, ok := userClaims.(map[string]interface{}); ok {
			if userID, ok := claims["userID"].(string); ok {
				requestingUserID = userID
			}
		}
	}

	// 获取用户档案
	profile, err := h.profileService.GetUserProfile(username, requestingUserID)
	if err != nil {
		if err.Error() == "用户不存在" {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		if err.Error() == "该用户的资料未公开" {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, "获取用户档案失败", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取用户资料成功", profile)
}

// GetUserLetters 获取用户信件列表
// GET /api/users/:username/letters
func (h *UserProfileHandler) GetUserLetters(c *gin.Context) {
	username := c.Param("username")
	publicOnly := c.Query("public") == "true"

	if username == "" {
		utils.BadRequestResponse(c, "用户名不能为空", nil)
		return
	}

	// 获取当前登录用户ID
	requestingUserID := ""
	if userClaims, exists := c.Get("userClaims"); exists {
		if claims, ok := userClaims.(map[string]interface{}); ok {
			if userID, ok := claims["userID"].(string); ok {
				requestingUserID = userID
			}
		}
	}

	// 获取信件列表
	letters, err := h.profileService.GetUserLetters(username, publicOnly, requestingUserID)
	if err != nil {
		if err.Error() == "用户不存在" {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, "获取用户信件失败", err)
		return
	}

	// 构造响应数据
	letterResponses := make([]gin.H, len(letters))
	for i, letter := range letters {
		letterResponses[i] = gin.H{
			"id":              letter.ID,
			"title":           letter.Title,
			"content_preview": truncateContent(letter.Content, 100),
			"created_at":      letter.CreatedAt,
			"status":          letter.Status,
			"visibility":      letter.Visibility,
			"like_count":      letter.LikeCount,
			"sender_username": username, // 简化处理
			"is_public":       letter.Visibility == models.VisibilityPublic,
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "获取用户信件成功", gin.H{
		"letters": letterResponses,
		"count":   len(letterResponses),
	})
}

// UpdateUserProfile 更新用户档案
// PUT /api/users/profile
func (h *UserProfileHandler) UpdateUserProfile(c *gin.Context) {
	// 获取当前登录用户
	userClaims, exists := c.Get("userClaims")
	if !exists {
		utils.UnauthorizedResponse(c, "请先登录")
		return
	}

	claims, ok := userClaims.(map[string]interface{})
	if !ok {
		utils.InternalServerErrorResponse(c, "用户信息格式错误", nil)
		return
	}

	userID, ok := claims["userID"].(string)
	if !ok {
		utils.InternalServerErrorResponse(c, "用户ID格式错误", nil)
		return
	}

	// 解析请求
	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Bio      string `json:"bio"`
		School   string `json:"school"`
		OPCode   string `json:"op_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效", err)
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.School != "" {
		updates["school"] = req.School
	}
	if req.OPCode != "" {
		// 验证OP Code格式
		if len(req.OPCode) != 6 {
			utils.BadRequestResponse(c, "OP Code必须是6位字符", nil)
			return
		}
		updates["op_code"] = req.OPCode
	}

	// 更新档案
	if err := h.profileService.UpdateUserProfile(userID, updates); err != nil {
		utils.InternalServerErrorResponse(c, "更新用户档案失败", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "更新成功", nil)
}

// UpdateUserPrivacy 更新用户隐私设置
// PUT /api/users/privacy
func (h *UserProfileHandler) UpdateUserPrivacy(c *gin.Context) {
	// 获取当前登录用户
	userClaims, exists := c.Get("userClaims")
	if !exists {
		utils.UnauthorizedResponse(c, "请先登录")
		return
	}

	claims, ok := userClaims.(map[string]interface{})
	if !ok {
		utils.InternalServerErrorResponse(c, "用户信息格式错误", nil)
		return
	}

	userID, ok := claims["userID"].(string)
	if !ok {
		utils.InternalServerErrorResponse(c, "用户ID格式错误", nil)
		return
	}

	// 解析请求
	var privacy models.UserPrivacy
	if err := c.ShouldBindJSON(&privacy); err != nil {
		utils.BadRequestResponse(c, "请求参数无效", err)
		return
	}

	// 验证隐私级别
	validPrivacyLevels := map[models.OPCodePrivacyLevel]bool{
		models.OPCodePrivacyFull:    true,
		models.OPCodePrivacyPartial: true,
		models.OPCodePrivacyHidden:  true,
	}

	if !validPrivacyLevels[privacy.OPCodePrivacy] {
		privacy.OPCodePrivacy = models.OPCodePrivacyPartial
	}

	// 更新隐私设置
	if err := h.profileService.UpdateUserPrivacy(userID, privacy); err != nil {
		utils.InternalServerErrorResponse(c, "更新隐私设置失败", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "更新成功", privacy)
}

// 辅助函数：截断内容
func truncateContent(content string, maxLength int) string {
	runes := []rune(content)
	if len(runes) <= maxLength {
		return content
	}
	return string(runes[:maxLength]) + "..."
}
