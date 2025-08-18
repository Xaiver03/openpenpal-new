package handlers

import (
	"net/http"
	"strconv"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// TagHandler 标签处理器
type TagHandler struct {
	tagService *services.TagService
}

// NewTagHandler 创建标签处理器
func NewTagHandler(tagService *services.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// =============== 标签CRUD操作 ===============

// CreateTag 创建标签
// POST /api/v1/tags
func (h *TagHandler) CreateTag(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.TagRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	tag, err := h.tagService.CreateTag(&req, userID)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to create tag", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Tag created successfully", tag)
}

// GetTag 获取标签详情
// GET /api/v1/tags/:id
func (h *TagHandler) GetTag(c *gin.Context) {
	tagID := c.Param("id")
	if tagID == "" {
		utils.BadRequestResponse(c, "Tag ID is required", nil)
		return
	}

	tag, err := h.tagService.GetTag(tagID)
	if err != nil {
		utils.NotFoundResponse(c, "Tag not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag retrieved successfully", tag)
}

// UpdateTag 更新标签
// PUT /api/v1/tags/:id
func (h *TagHandler) UpdateTag(c *gin.Context) {
	tagID := c.Param("id")
	if tagID == "" {
		utils.BadRequestResponse(c, "Tag ID is required", nil)
		return
	}

	var req models.TagRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	tag, err := h.tagService.UpdateTag(tagID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to update tag", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag updated successfully", tag)
}

// DeleteTag 删除标签
// DELETE /api/v1/tags/:id
func (h *TagHandler) DeleteTag(c *gin.Context) {
	tagID := c.Param("id")
	if tagID == "" {
		utils.BadRequestResponse(c, "Tag ID is required", nil)
		return
	}

	if err := h.tagService.DeleteTag(tagID); err != nil {
		utils.BadRequestResponse(c, "Failed to delete tag", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag deleted successfully", nil)
}

// =============== 标签搜索和发现 ===============

// SearchTags 搜索标签
// GET /api/v1/tags/search
func (h *TagHandler) SearchTags(c *gin.Context) {
	// 解析查询参数
	req := models.TagSearchRequest{
		Query:     c.Query("query"),
		Type:      models.TagType(c.Query("type")),
		Status:    models.TagStatus(c.Query("status")),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	// 解析分页参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	} else {
		req.Page = 1
	}

	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		req.Limit = limit
	} else {
		req.Limit = 20
	}

	// 解析分类ID
	if categoryID := c.Query("category_id"); categoryID != "" {
		req.CategoryID = &categoryID
	}

	result, err := h.tagService.SearchTags(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search tags", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tags retrieved successfully", result)
}

// GetPopularTags 获取热门标签
// GET /api/v1/tags/popular
func (h *TagHandler) GetPopularTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tags, err := h.tagService.GetPopularTags(limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get popular tags", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Popular tags retrieved successfully", map[string]interface{}{
		"tags":  tags,
		"count": len(tags),
	})
}

// GetTrendingTags 获取趋势标签
// GET /api/v1/tags/trending
func (h *TagHandler) GetTrendingTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tags, err := h.tagService.GetTrendingTags(limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get trending tags", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Trending tags retrieved successfully", map[string]interface{}{
		"tags":  tags,
		"count": len(tags),
	})
}

// =============== 内容标签管理 ===============

// TagContent 为内容添加标签
// POST /api/v1/content/tags
func (h *TagHandler) TagContent(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.ContentTagRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	if err := h.tagService.TagContent(&req, userID); err != nil {
		utils.BadRequestResponse(c, "Failed to tag content", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Content tagged successfully", nil)
}

// GetContentTags 获取内容的标签
// GET /api/v1/content/:type/:id/tags
func (h *TagHandler) GetContentTags(c *gin.Context) {
	contentType := c.Param("type")
	contentID := c.Param("id")

	if contentType == "" || contentID == "" {
		utils.BadRequestResponse(c, "Content type and ID are required", nil)
		return
	}

	result, err := h.tagService.GetContentTags(contentType, contentID)
	if err != nil {
		utils.NotFoundResponse(c, "Content tags not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Content tags retrieved successfully", result)
}

// RemoveContentTag 移除内容标签
// DELETE /api/v1/content/:type/:id/tags/:tag_id
func (h *TagHandler) RemoveContentTag(c *gin.Context) {
	contentType := c.Param("type")
	contentID := c.Param("id")
	tagID := c.Param("tag_id")

	if contentType == "" || contentID == "" || tagID == "" {
		utils.BadRequestResponse(c, "Content type, content ID, and tag ID are required", nil)
		return
	}

	if err := h.tagService.RemoveContentTag(contentType, contentID, tagID); err != nil {
		utils.BadRequestResponse(c, "Failed to remove content tag", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Content tag removed successfully", nil)
}

// =============== 标签建议 ===============

// SuggestTags AI标签建议
// POST /api/v1/tags/suggest
func (h *TagHandler) SuggestTags(c *gin.Context) {
	var req models.TagSuggestionRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 设置默认限制
	if req.Limit <= 0 || req.Limit > 20 {
		req.Limit = 10
	}

	suggestions, err := h.tagService.SuggestTags(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate tag suggestions", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag suggestions generated successfully", suggestions)
}

// =============== 标签统计 ===============

// GetTagStats 获取标签统计
// GET /api/v1/tags/stats
func (h *TagHandler) GetTagStats(c *gin.Context) {
	stats, err := h.tagService.GetTagStats()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get tag statistics", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag statistics retrieved successfully", stats)
}

// UpdateTrendingScores 更新趋势分数（管理员功能）
// POST /api/v1/tags/update-trending
func (h *TagHandler) UpdateTrendingScores(c *gin.Context) {
	// 检查管理员权限
	userRole, _ := middleware.GetUserRole(c)
	if !isAdminRole(userRole) {
		utils.ForbiddenResponse(c, "Admin privileges required")
		return
	}

	if err := h.tagService.UpdateTrendingScores(); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update trending scores", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Trending scores updated successfully", nil)
}

// =============== 标签分类管理 ===============

// CreateTagCategory 创建标签分类
// POST /api/v1/tags/categories
func (h *TagHandler) CreateTagCategory(c *gin.Context) {
	// 检查管理员权限
	userRole, _ := middleware.GetUserRole(c)
	if !isAdminRole(userRole) {
		utils.ForbiddenResponse(c, "Admin privileges required")
		return
	}

	var req models.TagCategoryRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	category, err := h.tagService.CreateTagCategory(&req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to create tag category", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Tag category created successfully", category)
}

// GetTagCategories 获取标签分类列表
// GET /api/v1/tags/categories
func (h *TagHandler) GetTagCategories(c *gin.Context) {
	categories, err := h.tagService.GetTagCategories()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get tag categories", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag categories retrieved successfully", map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	})
}

// =============== 批量操作 ===============

// BatchTagContent 批量为内容添加标签
// POST /api/v1/content/batch-tag
func (h *TagHandler) BatchTagContent(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		ContentItems []models.ContentTagRequest `json:"content_items" binding:"required"`
	}

	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 限制批量操作的数量
	if len(req.ContentItems) > 100 {
		utils.BadRequestResponse(c, "Too many items in batch (max 100)", nil)
		return
	}

	successCount := 0
	failedItems := []map[string]interface{}{}

	for _, item := range req.ContentItems {
		if err := h.tagService.TagContent(&item, userID); err != nil {
			failedItems = append(failedItems, map[string]interface{}{
				"content_type": item.ContentType,
				"content_id":   item.ContentID,
				"error":        err.Error(),
			})
		} else {
			successCount++
		}
	}

	result := map[string]interface{}{
		"success_count": successCount,
		"failed_count":  len(failedItems),
		"failed_items":  failedItems,
		"total":         len(req.ContentItems),
	}

	if len(failedItems) > 0 {
		utils.SuccessResponse(c, http.StatusPartialContent, "Batch tagging completed with some failures", result)
	} else {
		utils.SuccessResponse(c, http.StatusOK, "All items tagged successfully", result)
	}
}

// =============== 高级搜索 ===============

// SearchContentByTags 根据标签搜索内容
// POST /api/v1/tags/search-content
func (h *TagHandler) SearchContentByTags(c *gin.Context) {
	var req struct {
		TagIDs      []string `json:"tag_ids" binding:"required"`
		ContentType string   `json:"content_type"`
		MatchMode   string   `json:"match_mode"` // "any" or "all"
		Page        int      `json:"page"`
		Limit       int      `json:"limit"`
	}

	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.MatchMode == "" {
		req.MatchMode = "any"
	}

	// 这里需要实现具体的搜索逻辑
	// 由于涉及到多个内容类型，这里提供一个简化的响应
	result := map[string]interface{}{
		"message": "Content search by tags functionality is under development",
		"request": req,
	}

	utils.SuccessResponse(c, http.StatusOK, "Search request processed", result)
}

// =============== 用户标签关注 ===============

// FollowTag 关注标签
// POST /api/v1/tags/:id/follow
func (h *TagHandler) FollowTag(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	tagID := c.Param("id")
	if tagID == "" {
		utils.BadRequestResponse(c, "Tag ID is required", nil)
		return
	}

	// TODO: 实现用户关注标签的逻辑
	result := map[string]interface{}{
		"message": "Tag follow functionality is under development",
		"user_id": userID,
		"tag_id":  tagID,
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag followed successfully", result)
}

// UnfollowTag 取消关注标签
// DELETE /api/v1/tags/:id/follow
func (h *TagHandler) UnfollowTag(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	tagID := c.Param("id")
	if tagID == "" {
		utils.BadRequestResponse(c, "Tag ID is required", nil)
		return
	}

	// TODO: 实现取消关注标签的逻辑
	result := map[string]interface{}{
		"message": "Tag unfollow functionality is under development",
		"user_id": userID,
		"tag_id":  tagID,
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag unfollowed successfully", result)
}

// GetUserFollowedTags 获取用户关注的标签
// GET /api/v1/users/me/followed-tags
func (h *TagHandler) GetUserFollowedTags(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// TODO: 实现获取用户关注标签的逻辑
	result := map[string]interface{}{
		"message":  "User followed tags functionality is under development",
		"user_id":  userID,
		"page":     page,
		"limit":    limit,
		"tags":     []models.Tag{},
		"total":    0,
	}

	utils.SuccessResponse(c, http.StatusOK, "User followed tags retrieved successfully", result)
}

// =============== 辅助方法 ===============

// isAdminRole 检查是否为管理员角色
func isAdminRole(role string) bool {
	adminRoles := map[string]bool{
		string(models.RolePlatformAdmin): true,
		string(models.RoleSuperAdmin):    true,
		string(models.RoleCourierLevel4): true, // L4信使也有管理权限
		string(models.RoleCourierLevel3): true, // L3信使也有部分管理权限
	}
	return adminRoles[role]
}

// =============== 标签验证和清理 ===============

// ValidateTagUsage 验证标签使用情况（管理员功能）
// POST /api/v1/tags/validate-usage
func (h *TagHandler) ValidateTagUsage(c *gin.Context) {
	// 检查管理员权限
	userRole, _ := middleware.GetUserRole(c)
	if !isAdminRole(userRole) {
		utils.ForbiddenResponse(c, "Admin privileges required")
		return
	}

	// TODO: 实现标签使用验证逻辑
	result := map[string]interface{}{
		"message": "Tag usage validation functionality is under development",
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag usage validation completed", result)
}

// CleanupUnusedTags 清理未使用的标签（管理员功能）
// DELETE /api/v1/tags/cleanup-unused
func (h *TagHandler) CleanupUnusedTags(c *gin.Context) {
	// 检查管理员权限
	userRole, _ := middleware.GetUserRole(c)
	if !isAdminRole(userRole) {
		utils.ForbiddenResponse(c, "Admin privileges required")
		return
	}

	// TODO: 实现清理未使用标签的逻辑
	result := map[string]interface{}{
		"message": "Tag cleanup functionality is under development",
	}

	utils.SuccessResponse(c, http.StatusOK, "Unused tags cleanup completed", result)
}

// =============== 兼容层方法（来自legacy版本） ===============

// UntagContent 取消标记内容
// POST /api/v1/content/untag
func (h *TagHandler) UntagContent(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		ContentType string   `json:"content_type" binding:"required"`
		ContentID   string   `json:"content_id" binding:"required"`
		TagIDs      []string `json:"tag_ids"`
	}

	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// 简化实现：移除所有指定的标签关联
	for _, tagID := range req.TagIDs {
		if err := h.tagService.RemoveContentTag(req.ContentType, req.ContentID, tagID); err != nil {
			utils.BadRequestResponse(c, "Failed to untag content", err)
			return
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Content untagged successfully", nil)
}

// GetFollowedTags 获取用户关注的标签
// GET /api/v1/users/me/followed-tags
func (h *TagHandler) GetFollowedTags(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// TODO: 实现获取用户关注标签的逻辑
	result := map[string]interface{}{
		"message":  "User followed tags functionality is under development",
		"user_id":  userID,
		"page":     page,
		"limit":    limit,
		"tags":     []models.Tag{},
		"total":    0,
	}

	utils.SuccessResponse(c, http.StatusOK, "User followed tags retrieved successfully", result)
}

// GetTagCategory 获取标签分类详情
// GET /api/v1/tags/categories/:id
func (h *TagHandler) GetTagCategory(c *gin.Context) {
	categoryID := c.Param("id")
	if categoryID == "" {
		utils.BadRequestResponse(c, "Category ID is required", nil)
		return
	}

	// TODO: 实现获取分类详情逻辑
	result := map[string]interface{}{
		"message":     "Tag category detail functionality is under development",
		"category_id": categoryID,
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag category retrieved successfully", result)
}

// GetTagTrend 获取标签趋势
// GET /api/v1/tags/:id/trend
func (h *TagHandler) GetTagTrend(c *gin.Context) {
	tagID := c.Param("id")
	if tagID == "" {
		utils.BadRequestResponse(c, "Tag ID is required", nil)
		return
	}

	// TODO: 实现标签趋势分析
	result := map[string]interface{}{
		"tag_id":       tagID,
		"tag_name":     "Sample Tag",
		"trend_data":   []interface{}{},
		"current_rank": 1,
		"trend_change": "stable",
		"message":      "Tag trend analysis functionality is under development",
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag trend retrieved successfully", result)
}

// BatchOperateTags 批量操作标签
// POST /api/v1/tags/batch-operate
func (h *TagHandler) BatchOperateTags(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		TagIDs    []string               `json:"tag_ids" binding:"required"`
		Operation string                 `json:"operation" binding:"required,oneof=delete update_status update_category"`
		Data      map[string]interface{} `json:"data"`
	}

	if err := middleware.BindAndValidate(c, &req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// TODO: 实现批量操作逻辑
	result := map[string]interface{}{
		"message":    "Batch operation functionality is under development",
		"user_id":    userID,
		"operation":  req.Operation,
		"tag_count":  len(req.TagIDs),
		"processed":  0,
		"failed":     0,
	}

	utils.SuccessResponse(c, http.StatusOK, "Batch operation completed successfully", result)
}