package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// SensitiveWordHandler 敏感词管理处理器
type SensitiveWordHandler struct {
	securitySvc *services.ContentSecurityService
	// 速率限制器（针对敏感词管理操作）
	createLimiter *rate.Limiter // 创建敏感词限制
	updateLimiter *rate.Limiter // 更新敏感词限制
	batchLimiter  *rate.Limiter // 批量操作限制
}

// NewSensitiveWordHandler 创建敏感词管理处理器
func NewSensitiveWordHandler(securitySvc *services.ContentSecurityService) *SensitiveWordHandler {
	return &SensitiveWordHandler{
		securitySvc:   securitySvc,
		// 初始化速率限制器
		createLimiter: rate.NewLimiter(rate.Every(time.Second), 10),    // 每秒最多10次创建操作
		updateLimiter: rate.NewLimiter(rate.Every(time.Second), 20),    // 每秒最多20次更新操作
		batchLimiter:  rate.NewLimiter(rate.Every(time.Minute), 5),     // 每分钟最多5次批量操作
	}
}

// ================================
// 敏感词CRUD操作
// ================================

// ListSensitiveWords 获取敏感词列表
// GET /api/v1/admin/sensitive-words
func (h *SensitiveWordHandler) ListSensitiveWords(c *gin.Context) {
	// 权限检查：只有四级信使和平台管理员可以访问
	userRole, _ := middleware.GetUserRole(c)
	if !h.hasPermission(userRole) {
		utils.ForbiddenResponse(c, "Insufficient permissions to manage sensitive words")
		return
	}

	// 查询参数验证
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	category := c.Query("category")
	isActive := c.Query("is_active")

	// 分页参数验证
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // 限制每页最大数量
	}

	// 分类参数验证
	if category != "" {
		if err := h.validateCategory(category); err != nil {
			utils.BadRequestResponse(c, "Invalid category parameter", err)
			return
		}
	}

	words, total, err := h.securitySvc.GetSensitiveWords(page, limit, category, isActive)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch sensitive words", err)
		return
	}

	pagination := utils.CalculatePagination(page, limit, total)
	utils.SuccessResponseWithPagination(c, words, pagination)
}

// CreateSensitiveWord 创建敏感词
// POST /api/v1/admin/sensitive-words
func (h *SensitiveWordHandler) CreateSensitiveWord(c *gin.Context) {
	// 权限检查
	userRole, _ := middleware.GetUserRole(c)
	userID, _ := middleware.GetUserID(c)
	if !h.hasPermission(userRole) {
		utils.ForbiddenResponse(c, "Insufficient permissions to manage sensitive words")
		return
	}

	// 速率限制检查
	if !h.createLimiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   "Rate limit exceeded",
			"message": "Too many create requests. Please try again later.",
		})
		c.Header("Retry-After", "60")
		return
	}

	var req models.SensitiveWordRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 增强输入验证
	if err := h.validateSensitiveWordRequest(&req); err != nil {
		utils.BadRequestResponse(c, "Input validation failed", err)
		return
	}

	// 调用服务创建敏感词
	word := &models.SensitiveWord{
		ID:        uuid.New().String(),
		Word:      req.Word,
		Category:  req.Category,
		Level:     req.Level,
		IsActive:  true,
		CreatedBy: userID,
	}

	if err := h.securitySvc.AddSensitiveWord(word.Word, word.Category, string(word.Level)); err != nil {
		utils.BadRequestResponse(c, "Failed to create sensitive word", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Sensitive word created successfully", word)
}

// UpdateSensitiveWord 更新敏感词
// PUT /api/v1/admin/sensitive-words/:id
func (h *SensitiveWordHandler) UpdateSensitiveWord(c *gin.Context) {
	// 权限检查
	userRole, _ := middleware.GetUserRole(c)
	if !h.hasPermission(userRole) {
		utils.ForbiddenResponse(c, "Insufficient permissions to manage sensitive words")
		return
	}

	// 速率限制检查
	if !h.updateLimiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   "Rate limit exceeded",
			"message": "Too many update requests. Please try again later.",
		})
		c.Header("Retry-After", "30")
		return
	}

	wordID := c.Param("id")
	if err := h.validateWordID(wordID); err != nil {
		utils.BadRequestResponse(c, "Invalid word ID format", err)
		return
	}

	var req models.SensitiveWordRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 增强输入验证
	if err := h.validateSensitiveWordRequest(&req); err != nil {
		utils.BadRequestResponse(c, "Input validation failed", err)
		return
	}

	// 调用服务更新敏感词
	if err := h.securitySvc.UpdateSensitiveWord(wordID, req.Word, req.Category, string(req.Level)); err != nil {
		utils.BadRequestResponse(c, "Failed to update sensitive word", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sensitive word updated successfully", nil)
}

// DeleteSensitiveWord 删除敏感词（软删除）
// DELETE /api/v1/admin/sensitive-words/:id
func (h *SensitiveWordHandler) DeleteSensitiveWord(c *gin.Context) {
	// 权限检查
	userRole, _ := middleware.GetUserRole(c)
	if !h.hasPermission(userRole) {
		utils.ForbiddenResponse(c, "Insufficient permissions to manage sensitive words")
		return
	}

	// 速率限制检查（使用更新限制器）
	if !h.updateLimiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   "Rate limit exceeded",
			"message": "Too many delete requests. Please try again later.",
		})
		c.Header("Retry-After", "30")
		return
	}

	wordID := c.Param("id")
	if err := h.validateWordID(wordID); err != nil {
		utils.BadRequestResponse(c, "Invalid word ID format", err)
		return
	}

	// 调用服务删除敏感词
	if err := h.securitySvc.DeleteSensitiveWord(wordID); err != nil {
		utils.BadRequestResponse(c, "Failed to delete sensitive word", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sensitive word deleted successfully", nil)
}

// ================================
// 批量操作
// ================================

// BatchImportSensitiveWords 批量导入敏感词
// POST /api/v1/admin/sensitive-words/batch-import
func (h *SensitiveWordHandler) BatchImportSensitiveWords(c *gin.Context) {
	// 权限检查
	userRole, _ := middleware.GetUserRole(c)
	if !h.hasPermission(userRole) {
		utils.ForbiddenResponse(c, "Insufficient permissions to manage sensitive words")
		return
	}

	// 批量操作速率限制检查
	if !h.batchLimiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   "Rate limit exceeded",
			"message": "Too many batch operations. Please try again later.",
		})
		c.Header("Retry-After", "300") // 5分钟后重试
		return
	}

	var req struct {
		Words []struct {
			Word     string                 `json:"word" binding:"required"`
			Category string                 `json:"category"`
			Level    models.ModerationLevel `json:"level" binding:"required"`
		} `json:"words" binding:"required"`
	}

	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 批量导入限制检查
	if err := h.validateBatchImportRequest(&req); err != nil {
		utils.BadRequestResponse(c, "Batch import validation failed", err)
		return
	}

	// 批量添加敏感词
	successCount := 0
	failedWords := []string{}

	for _, w := range req.Words {
		err := h.securitySvc.AddSensitiveWord(w.Word, w.Category, string(w.Level))
		if err != nil {
			failedWords = append(failedWords, w.Word)
		} else {
			successCount++
		}
	}

	result := map[string]interface{}{
		"success_count": successCount,
		"failed_count":  len(failedWords),
		"failed_words":  failedWords,
		"total":         len(req.Words),
	}

	if len(failedWords) > 0 {
		utils.SuccessResponse(c, http.StatusPartialContent, "Batch import completed with some failures", result)
	} else {
		utils.SuccessResponse(c, http.StatusOK, "All sensitive words imported successfully", result)
	}
}

// ExportSensitiveWords 导出敏感词
// GET /api/v1/admin/sensitive-words/export
func (h *SensitiveWordHandler) ExportSensitiveWords(c *gin.Context) {
	// 权限检查
	userRole, _ := middleware.GetUserRole(c)
	if !h.hasPermission(userRole) {
		utils.ForbiddenResponse(c, "Insufficient permissions to manage sensitive words")
		return
	}

	// 获取所有敏感词
	words, _, err := h.securitySvc.GetSensitiveWords(1, 10000, "", "true")
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to export sensitive words", err)
		return
	}

	// 转换为导出格式
	exportData := make([]map[string]interface{}, len(words))
	for i, word := range words {
		exportData[i] = map[string]interface{}{
			"word":     word.Word,
			"category": word.Category,
			"level":    word.Level,
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Sensitive words exported successfully", exportData)
}

// ================================
// 敏感词刷新
// ================================

// RefreshSensitiveWords 刷新敏感词库（重新加载到内存）
// POST /api/v1/admin/sensitive-words/refresh
func (h *SensitiveWordHandler) RefreshSensitiveWords(c *gin.Context) {
	// 权限检查
	userRole, _ := middleware.GetUserRole(c)
	if !h.hasPermission(userRole) {
		utils.ForbiddenResponse(c, "Insufficient permissions to manage sensitive words")
		return
	}

	// 刷新敏感词库
	if err := h.securitySvc.RefreshSensitiveWords(); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to refresh sensitive words", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sensitive words refreshed successfully", nil)
}

// ================================
// 敏感词统计
// ================================

// GetSensitiveWordStats 获取敏感词统计信息
// GET /api/v1/admin/sensitive-words/stats
func (h *SensitiveWordHandler) GetSensitiveWordStats(c *gin.Context) {
	// 权限检查
	userRole, _ := middleware.GetUserRole(c)
	if !h.hasPermission(userRole) {
		utils.ForbiddenResponse(c, "Insufficient permissions to view sensitive word stats")
		return
	}

	// 添加缓存头以避免频繁请求
	c.Header("Cache-Control", "public, max-age=300") // 5分钟缓存

	stats, err := h.securitySvc.GetSensitiveWordStats()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get sensitive word stats", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sensitive word stats fetched successfully", stats)
}

// ================================
// 辅助方法
// ================================

// hasPermission 检查用户是否有权限管理敏感词
func (h *SensitiveWordHandler) hasPermission(role string) bool {
	// 只有四级信使（城市总代）和平台管理员有权限
	allowedRoles := map[string]bool{
		string(models.RoleCourierLevel4):  true, // 四级信使
		string(models.RolePlatformAdmin):  true, // 平台管理员
		string(models.RoleSuperAdmin):     true, // 超级管理员
	}
	return allowedRoles[role]
}

// ================================
// 输入验证辅助方法
// ================================

// validateSensitiveWordRequest 验证敏感词请求
func (h *SensitiveWordHandler) validateSensitiveWordRequest(req *models.SensitiveWordRequest) error {
	// 敏感词内容验证
	if err := h.validateWord(req.Word); err != nil {
		return fmt.Errorf("word validation failed: %w", err)
	}

	// 分类验证
	if req.Category != "" {
		if err := h.validateCategory(req.Category); err != nil {
			return fmt.Errorf("category validation failed: %w", err)
		}
	}

	// 等级验证
	if err := h.validateModerationLevel(string(req.Level)); err != nil {
		return fmt.Errorf("level validation failed: %w", err)
	}

	return nil
}

// validateWord 验证敏感词内容
func (h *SensitiveWordHandler) validateWord(word string) error {
	// 长度检查
	if len(strings.TrimSpace(word)) == 0 {
		return fmt.Errorf("word cannot be empty")
	}

	if utf8.RuneCountInString(word) > 100 {
		return fmt.Errorf("word too long (max 100 characters)")
	}

	if utf8.RuneCountInString(word) < 1 {
		return fmt.Errorf("word too short (min 1 character)")
	}

	// 字符验证：只允许中文、英文、数字、常见标点
	validPattern := regexp.MustCompile(`^[\p{Han}\p{L}\p{N}\p{P}\p{S}\s]+$`)
	if !validPattern.MatchString(word) {
		return fmt.Errorf("word contains invalid characters")
	}

	// 防止XSS和注入攻击
	dangerousPatterns := []string{
		"<script", "javascript:", "onclick=", "onerror=",
		"<iframe", "<object", "<embed", "data:text/html",
		"'; drop", "'; delete", "union select",
	}

	wordLower := strings.ToLower(word)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(wordLower, pattern) {
			return fmt.Errorf("word contains potentially dangerous content")
		}
	}

	return nil
}

// validateCategory 验证分类
func (h *SensitiveWordHandler) validateCategory(category string) error {
	if len(strings.TrimSpace(category)) == 0 {
		return nil // 分类可以为空
	}

	if utf8.RuneCountInString(category) > 50 {
		return fmt.Errorf("category too long (max 50 characters)")
	}

	// 预定义的有效分类
	validCategories := map[string]bool{
		"政治":   true,
		"色情":   true,
		"暴力":   true,
		"广告":   true,
		"欺诈":   true,
		"仇恨":   true,
		"谣言":   true,
		"赌博":   true,
		"毒品":   true,
		"其他":   true,
		"违法":   true,
		"不当":   true,
		"politics": true,
		"adult":    true,
		"violence": true,
		"spam":     true,
		"fraud":    true,
		"hate":     true,
		"fake":     true,
		"gambling": true,
		"drugs":    true,
		"other":    true,
		"illegal":  true,
		"inappropriate": true,
	}

	if !validCategories[category] {
		return fmt.Errorf("invalid category: %s", category)
	}

	return nil
}

// validateModerationLevel 验证审核等级
func (h *SensitiveWordHandler) validateModerationLevel(level string) error {
	validLevels := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"block":  true,
	}

	if !validLevels[level] {
		return fmt.Errorf("invalid moderation level: %s", level)
	}

	return nil
}

// validateWordID 验证敏感词ID
func (h *SensitiveWordHandler) validateWordID(wordID string) error {
	if len(strings.TrimSpace(wordID)) == 0 {
		return fmt.Errorf("word ID cannot be empty")
	}

	// UUID格式验证
	uuidPattern := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	if !uuidPattern.MatchString(wordID) {
		return fmt.Errorf("invalid UUID format")
	}

	return nil
}

// validateBatchImportRequest 验证批量导入请求
func (h *SensitiveWordHandler) validateBatchImportRequest(req *struct {
	Words []struct {
		Word     string                 `json:"word" binding:"required"`
		Category string                 `json:"category"`
		Level    models.ModerationLevel `json:"level" binding:"required"`
	} `json:"words" binding:"required"`
}) error {
	// 批量导入数量限制
	if len(req.Words) == 0 {
		return fmt.Errorf("no words provided")
	}

	if len(req.Words) > 1000 {
		return fmt.Errorf("too many words in batch (max 1000)")
	}

	// 验证每个敏感词
	wordSet := make(map[string]bool)
	for i, wordReq := range req.Words {
		// 创建临时请求对象进行验证
		tempReq := &models.SensitiveWordRequest{
			Word:     wordReq.Word,
			Category: wordReq.Category,
			Level:    wordReq.Level,
		}

		if err := h.validateSensitiveWordRequest(tempReq); err != nil {
			return fmt.Errorf("word %d validation failed: %w", i+1, err)
		}

		// 检查重复
		wordLower := strings.ToLower(strings.TrimSpace(wordReq.Word))
		if wordSet[wordLower] {
			return fmt.Errorf("duplicate word in batch: %s", wordReq.Word)
		}
		wordSet[wordLower] = true
	}

	return nil
}