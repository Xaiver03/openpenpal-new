package handlers

import (
	"net/http"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"shared/pkg/response"
)

type LetterHandler struct {
	letterService   *services.LetterService
	envelopeService *services.EnvelopeService
}

func NewLetterHandler(letterService *services.LetterService, envelopeService *services.EnvelopeService) *LetterHandler {
	return &LetterHandler{
		letterService:   letterService,
		envelopeService: envelopeService,
	}
}

// CreateDraft 创建草稿
func (h *LetterHandler) CreateDraft(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.CreateLetterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.LetterValidationMsg)
		return
	}

	letter, err := h.letterService.CreateDraft(userID, &req)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "Draft created successfully", letter)
}

// GenerateCode 生成信件编号和二维码
func (h *LetterHandler) GenerateCode(c *gin.Context) {
	resp := response.NewGinResponse()

	_, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	letterID := c.Param("id")
	if letterID == "" {
		resp.BadRequest(c, "Letter ID is required")
		return
	}

	// 验证信件所有权
	// 这里可以添加额外的验证逻辑

	letterCode, err := h.letterService.GenerateCode(letterID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "Letter code generated successfully", gin.H{
		"letter_code": letterCode.Code,
		"qr_code_url": letterCode.QRCodeURL,
		"read_url":    letterCode.Letter.ID, // 这里应该生成正确的读取URL
	})
}

// GetLetterByCode 通过编号获取信件
func (h *LetterHandler) GetLetterByCode(c *gin.Context) {
	resp := response.NewGinResponse()

	code := c.Param("code")
	if code == "" {
		resp.BadRequest(c, "Letter code is required")
		return
	}

	letter, err := h.letterService.GetLetterByCode(code)
	if err != nil {
		resp.NotFound(c, "Letter not found")
		return
	}

	resp.Success(c, letter)
}

// UpdateStatus 更新信件状态
func (h *LetterHandler) UpdateStatus(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	code := c.Param("code")
	if code == "" {
		resp.BadRequest(c, "Letter code is required")
		return
	}

	var req models.UpdateLetterStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.letterService.UpdateStatus(code, &req, userID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "Letter status updated successfully")
}

// GetUserLetters 获取用户信件列表
func (h *LetterHandler) GetUserLetters(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	var params models.LetterListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	letters, total, err := h.letterService.GetUserLetters(userID, &params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	pagination := utils.CalculatePagination(params.Page, params.Limit, total)

	resp.Success(c, gin.H{
		"data":       letters,
		"pagination": pagination,
	})
}

// GetUserStats 获取用户信件统计
func (h *LetterHandler) GetUserStats(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	stats, err := h.letterService.GetUserStats(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// MarkAsRead 标记信件为已读
func (h *LetterHandler) MarkAsRead(c *gin.Context) {
	resp := response.NewGinResponse()

	letterID := c.Param("id")
	if letterID == "" {
		resp.BadRequest(c, "Letter ID is required")
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	// 先获取信件的code
	letter, err := h.letterService.GetLetterByID(letterID, userID)
	if err != nil {
		if err.Error() == "letter not found" || err.Error() == "unauthorized to view this letter" {
			resp.NotFound(c, err.Error())
		} else {
			resp.InternalServerError(c, err.Error())
		}
		return
	}

	// 获取信件的code
	if letter.Code == nil {
		resp.BadRequest(c, "Letter code not generated yet")
		return
	}

	if err := h.letterService.MarkAsRead(letter.Code.Code, userID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "Letter marked as read successfully")
}

// GetLetter 获取单个信件详情
func (h *LetterHandler) GetLetter(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	letterID := c.Param("id")
	if letterID == "" {
		resp.BadRequest(c, "Letter ID is required")
		return
	}

	// 调用服务层获取信件详情
	letter, err := h.letterService.GetLetterByID(letterID, userID)
	if err != nil {
		if err.Error() == "letter not found" || err.Error() == "unauthorized to view this letter" {
			resp.NotFound(c, err.Error())
		} else {
			resp.InternalServerError(c, err.Error())
		}
		return
	}

	resp.Success(c, letter)
}

// UpdateLetter 更新信件内容
func (h *LetterHandler) UpdateLetter(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	letterID := c.Param("id")
	if letterID == "" {
		resp.BadRequest(c, "Letter ID is required")
		return
	}

	var req models.UpdateLetterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 调用服务层更新信件
	err := h.letterService.UpdateLetter(letterID, userID, &req)
	if err != nil {
		if err.Error() == "letter not found or unauthorized" {
			resp.NotFound(c, err.Error())
		} else if err.Error() == "only draft letters can be edited" {
			resp.BadRequest(c, err.Error())
		} else {
			resp.InternalServerError(c, err.Error())
		}
		return
	}

	// 获取更新后的信件数据
	updatedLetter, err := h.letterService.GetLetterByID(letterID, userID)
	if err != nil {
		resp.InternalServerError(c, "Failed to retrieve updated letter")
		return
	}

	resp.SuccessWithMessage(c, "Letter updated successfully", updatedLetter)
}

// GetPublicLetters 获取广场公开信件列表
func (h *LetterHandler) GetPublicLetters(c *gin.Context) {
	var params models.LetterListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequestResponse(c, "Invalid query parameters", err)
		return
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	letters, total, err := h.letterService.GetPublicLetters(&params)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get public letters", err)
		return
	}

	pagination := utils.CalculatePagination(params.Page, params.Limit, total)

	utils.SuccessResponse(c, 200, "Public letters retrieved successfully", gin.H{
		"data":       letters,
		"pagination": pagination,
	})
}

// DeleteLetter 删除信件
func (h *LetterHandler) DeleteLetter(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	letterID := c.Param("id")
	if letterID == "" {
		resp.BadRequest(c, "Letter ID is required")
		return
	}

	// 调用服务层删除信件
	err := h.letterService.DeleteLetter(letterID, userID)
	if err != nil {
		if err.Error() == "letter not found or unauthorized" {
			resp.NotFound(c, err.Error())
		} else if err.Error() == "only draft letters can be deleted" {
			resp.BadRequest(c, err.Error())
		} else {
			resp.InternalServerError(c, err.Error())
		}
		return
	}

	resp.OK(c, "Letter deleted successfully")
}

// ========================= Reply/Thread Handlers =========================

// GetReplyInfoByCode 扫码获取回信信息
func (h *LetterHandler) GetReplyInfoByCode(c *gin.Context) {
	resp := response.NewGinResponse()

	code := c.Param("code")
	if code == "" {
		resp.BadRequest(c, "信件代码不能为空")
		return
	}

	letterInfo, err := h.letterService.GetReplyInfoByCode(code)
	if err != nil {
		if err.Error() == "信件不存在或二维码无效" {
			resp.NotFound(c, err.Error())
		} else if err.Error() == "只有已送达或已读的信件才能回信" {
			resp.BadRequest(c, err.Error())
		} else {
			resp.InternalServerError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, letterInfo)
}

// CreateReply 创建回信
func (h *LetterHandler) CreateReply(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户未登录")
		return
	}

	var req models.CreateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	reply, err := h.letterService.CreateReply(userID, &req)
	if err != nil {
		if err.Error() == "原始信件不存在" {
			resp.NotFound(c, err.Error())
		} else {
			resp.InternalServerError(c, err.Error())
		}
		return
	}

	resp.Created(c, reply)
}

// GetUserThreads 获取用户线程列表
func (h *LetterHandler) GetUserThreads(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户未登录")
		return
	}

	// 解析分页参数
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	threads, err := h.letterService.GetUserThreads(userID, page, limit)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"threads": threads,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(threads), // 这里简化处理，实际应该返回总数
		},
	})
}

// GetThreadByID 获取指定线程详情
func (h *LetterHandler) GetThreadByID(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户未登录")
		return
	}

	threadID := c.Param("id")
	if threadID == "" {
		resp.BadRequest(c, "线程ID不能为空")
		return
	}

	thread, err := h.letterService.GetThreadByID(userID, threadID)
	if err != nil {
		if err.Error() == "线程不存在" {
			resp.NotFound(c, err.Error())
		} else if err.Error() == "无权限访问此线程" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			resp.InternalServerError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, thread)
}

// ========================= AI-Powered Letter Endpoints =========================

// GetWritingInspiration 获取AI写作灵感
func (h *LetterHandler) GetWritingInspiration(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户未登录")
		return
	}

	var req struct {
		Theme string   `json:"theme"`
		Style string   `json:"style"`
		Tags  []string `json:"tags"`
		Count int      `json:"count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Count <= 0 || req.Count > 10 {
		req.Count = 3
	}

	inspiration, err := h.letterService.GetWritingInspiration(userID, req.Theme, req.Style, req.Tags, req.Count)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, inspiration)
}

// GetAIReplyAssistance 获取AI回信助手建议
func (h *LetterHandler) GetAIReplyAssistance(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户未登录")
		return
	}

	letterID := c.Param("letter_id")
	if letterID == "" {
		resp.BadRequest(c, "信件ID不能为空")
		return
	}

	var req struct {
		Persona models.AIPersona `json:"persona"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	reply, err := h.letterService.GetAIReplyAssistance(userID, letterID, req.Persona)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, reply)
}

// MatchPenPalForLetter 为信件匹配笔友
func (h *LetterHandler) MatchPenPalForLetter(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户未登录")
		return
	}

	letterID := c.Param("letter_id")
	if letterID == "" {
		resp.BadRequest(c, "信件ID不能为空")
		return
	}

	matches, err := h.letterService.MatchPenPalForLetter(userID, letterID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, matches)
}

// AutoCurateLetterForMuseum 自动策展信件到博物馆
func (h *LetterHandler) AutoCurateLetterForMuseum(c *gin.Context) {
	resp := response.NewGinResponse()

	_, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户未登录")
		return
	}

	letterID := c.Param("letter_id")
	if letterID == "" {
		resp.BadRequest(c, "信件ID不能为空")
		return
	}

	var req struct {
		ExhibitionID string `json:"exhibition_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.letterService.AutoCurateLetterForMuseum(letterID, req.ExhibitionID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "信件已成功策展到博物馆")
}

// GetDrafts 获取草稿列表
func (h *LetterHandler) GetDrafts(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sortBy := c.DefaultQuery("sort_by", "updated_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	drafts, total, err := h.letterService.GetUserDrafts(c.Request.Context(), userID, page, limit, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get drafts",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"drafts": drafts,
			"total":  total,
			"page":   page,
			"limit":  limit,
		},
	})
}

// PublishLetter 发布信件
func (h *LetterHandler) PublishLetter(c *gin.Context) {
	letterID := c.Param("id")
	if letterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Letter ID is required",
		})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		ScheduledAt *time.Time `json:"scheduled_at"`
		Visibility  string     `json:"visibility"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空请求体
		req.Visibility = "private"
	}

	letter, err := h.letterService.PublishLetter(c.Request.Context(), letterID, userID, req.ScheduledAt, req.Visibility)
	if err != nil {
		if err.Error() == "letter not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Letter not found",
			})
			return
		}
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "You can only publish your own letters",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to publish letter",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    letter,
		"message": "Letter published successfully",
	})
}

// LikeLetter 点赞信件
func (h *LetterHandler) LikeLetter(c *gin.Context) {
	letterID := c.Param("id")
	if letterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Letter ID is required",
		})
		return
	}

	userID, _ := middleware.GetUserID(c)
	userIDStr := userID

	err := h.letterService.LikeLetter(c.Request.Context(), letterID, userIDStr)
	if err != nil {
		if err.Error() == "letter not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Letter not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to like letter",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Letter liked successfully",
	})
}

// ShareLetter 分享信件
func (h *LetterHandler) ShareLetter(c *gin.Context) {
	letterID := c.Param("id")
	if letterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Letter ID is required",
		})
		return
	}

	var req struct {
		Platform string `json:"platform"` // wechat, weibo, twitter, etc.
		Message  string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	userID, _ := middleware.GetUserID(c)
	userIDStr := userID

	shareInfo, err := h.letterService.ShareLetter(c.Request.Context(), letterID, userIDStr, req.Platform)
	if err != nil {
		if err.Error() == "letter not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Letter not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to share letter",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    shareInfo,
		"message": "Letter shared successfully",
	})
}

// GetLetterTemplates 获取信件模板
func (h *LetterHandler) GetLetterTemplates(c *gin.Context) {
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	templates, total, err := h.letterService.GetLetterTemplates(c.Request.Context(), category, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get templates",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"templates": templates,
			"total":     total,
			"page":      page,
			"limit":     limit,
		},
	})
}

// GetLetterTemplateByID 获取模板详情
func (h *LetterHandler) GetLetterTemplateByID(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Template ID is required",
		})
		return
	}

	template, err := h.letterService.GetTemplateByID(c.Request.Context(), templateID)
	if err != nil {
		if err.Error() == "template not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Template not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get template",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    template,
	})
}

// SearchLetters 搜索信件
func (h *LetterHandler) SearchLetters(c *gin.Context) {
	var req struct {
		Query      string   `json:"query"`
		Tags       []string `json:"tags"`
		DateFrom   string   `json:"date_from"`
		DateTo     string   `json:"date_to"`
		Visibility string   `json:"visibility"`
		Page       int      `json:"page"`
		Limit      int      `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	userID, _ := middleware.GetUserID(c)
	userIDStr := userID

	results, total, err := h.letterService.SearchLetters(c.Request.Context(), userIDStr, req.Query, req.Tags, req.DateFrom, req.DateTo, req.Visibility, req.Page, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to search letters",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"letters": results,
			"total":   total,
			"page":    req.Page,
			"limit":   req.Limit,
		},
	})
}

// GetPopularLetters 获取热门信件
func (h *LetterHandler) GetPopularLetters(c *gin.Context) {
	timeRange := c.DefaultQuery("time_range", "week") // day, week, month, all
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	letters, total, err := h.letterService.GetPopularLetters(c.Request.Context(), timeRange, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get popular letters",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"letters": letters,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
	})
}

// GetRecommendedLetters 获取推荐信件
func (h *LetterHandler) GetRecommendedLetters(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	userIDStr := userID

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	letters, total, err := h.letterService.GetRecommendedLetters(c.Request.Context(), userIDStr, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get recommended letters",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"letters": letters,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
	})
}

// BatchOperateLetters 批量操作信件
func (h *LetterHandler) BatchOperateLetters(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		LetterIDs []string               `json:"letter_ids" binding:"required"`
		Operation string                 `json:"operation" binding:"required,oneof=delete archive publish"`
		Data      map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	results, err := h.letterService.BatchOperate(c.Request.Context(), userID, req.LetterIDs, req.Operation, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to perform batch operation",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"message": "Batch operation completed",
	})
}

// ExportLetters 导出信件
func (h *LetterHandler) ExportLetters(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		LetterIDs          []string `json:"letter_ids"`
		Format             string   `json:"format" binding:"required,oneof=pdf txt json"`
		IncludeAttachments bool     `json:"include_attachments"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	exportData, err := h.letterService.ExportLetters(c.Request.Context(), userID, req.LetterIDs, req.Format, req.IncludeAttachments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to export letters",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    exportData,
		"message": "Letters exported successfully",
	})
}

// AutoSaveDraft 自动保存草稿
func (h *LetterHandler) AutoSaveDraft(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var req models.Letter
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	req.UserID = userID

	letter, err := h.letterService.AutoSaveDraft(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to auto-save draft",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    letter,
		"message": "Draft auto-saved",
	})
}

// GetWritingSuggestions 获取写作建议
func (h *LetterHandler) GetWritingSuggestions(c *gin.Context) {
	var req struct {
		Content string   `json:"content"`
		Style   string   `json:"style"`
		Tags    []string `json:"tags"`
		Mood    string   `json:"mood"`
		Purpose string   `json:"purpose"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	suggestions, err := h.letterService.GetWritingSuggestions(c.Request.Context(), req.Content, req.Style, req.Tags, req.Mood, req.Purpose)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get suggestions",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    suggestions,
	})
}
