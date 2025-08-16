package handlers

import (
	"net/http"
	"strconv"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

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

// CreateTag 创建标签
func (h *TagHandler) CreateTag(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var req models.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	tag, err := h.tagService.CreateTag(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create tag",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Tag created successfully",
		"data":    tag,
	})
}

// GetTag 获取标签详情
func (h *TagHandler) GetTag(c *gin.Context) {
	tagID := c.Param("id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Tag ID is required",
		})
		return
	}

	userID, _ := middleware.GetUserID(c) // 不要求认证，但如果有认证会提供更多信息

	tagResponse, err := h.tagService.GetTag(tagID, userID)
	if err != nil {
		if err.Error() == "tag not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Tag not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get tag",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tagResponse,
	})
}

// UpdateTag 更新标签
func (h *TagHandler) UpdateTag(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	tagID := c.Param("id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Tag ID is required",
		})
		return
	}

	var req models.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	tag, err := h.tagService.UpdateTag(tagID, userID, &req)
	if err != nil {
		if err.Error() == "tag not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Tag not found",
			})
			return
		}
		
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Permission denied",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update tag",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tag updated successfully",
		"data":    tag,
	})
}

// DeleteTag 删除标签
func (h *TagHandler) DeleteTag(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	tagID := c.Param("id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Tag ID is required",
		})
		return
	}

	err := h.tagService.DeleteTag(tagID, userID)
	if err != nil {
		if err.Error() == "tag not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Tag not found",
			})
			return
		}
		
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Permission denied",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete tag",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tag deleted successfully",
	})
}

// SearchTags 搜索标签
func (h *TagHandler) SearchTags(c *gin.Context) {
	var req models.TagSearchRequest

	// 解析查询参数
	req.Query = c.Query("query")
	req.Type = models.TagType(c.Query("type"))
	req.Status = models.TagStatus(c.Query("status"))
	
	categoryID := c.Query("category_id")
	if categoryID != "" {
		req.CategoryID = &categoryID
	}

	// 分页参数
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	req.Page = page

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	req.Limit = limit

	req.SortBy = c.Query("sort_by")
	req.SortOrder = c.Query("sort_order")

	result, err := h.tagService.SearchTags(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to search tags",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetPopularTags 获取热门标签
func (h *TagHandler) GetPopularTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	tags, err := h.tagService.GetPopularTags(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get popular tags",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tags,
	})
}

// GetTrendingTags 获取趋势标签
func (h *TagHandler) GetTrendingTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	tags, err := h.tagService.GetTrendingTags(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get trending tags",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tags,
	})
}

// SuggestTags 标签建议
func (h *TagHandler) SuggestTags(c *gin.Context) {
	var req models.TagSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	if req.Limit < 1 || req.Limit > 20 {
		req.Limit = 10
	}

	suggestions, err := h.tagService.SuggestTags(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get tag suggestions",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    suggestions,
	})
}

// TagContent 标记内容
func (h *TagHandler) TagContent(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var req models.ContentTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	err := h.tagService.TagContent(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to tag content",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Content tagged successfully",
	})
}

// UntagContent 取消标记
func (h *TagHandler) UntagContent(c *gin.Context) {
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		ContentType string   `json:"content_type" binding:"required"`
		ContentID   string   `json:"content_id" binding:"required"`
		TagIDs      []string `json:"tag_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	err := h.tagService.UntagContent(req.ContentType, req.ContentID, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to untag content",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Content untagged successfully",
	})
}

// GetContentTags 获取内容标签
func (h *TagHandler) GetContentTags(c *gin.Context) {
	contentType := c.Param("content_type")
	contentID := c.Param("content_id")

	if contentType == "" || contentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Content type and ID are required",
		})
		return
	}

	result, err := h.tagService.GetContentTags(contentType, contentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get content tags",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// FollowTag 关注标签
func (h *TagHandler) FollowTag(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	tagID := c.Param("id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Tag ID is required",
		})
		return
	}

	err := h.tagService.FollowTag(userID, tagID)
	if err != nil {
		if err.Error() == "already following this tag" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Already following this tag",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to follow tag",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tag followed successfully",
	})
}

// UnfollowTag 取消关注标签
func (h *TagHandler) UnfollowTag(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	tagID := c.Param("id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Tag ID is required",
		})
		return
	}

	err := h.tagService.UnfollowTag(userID, tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to unfollow tag",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tag unfollowed successfully",
	})
}

// GetFollowedTags 获取关注的标签
func (h *TagHandler) GetFollowedTags(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result, err := h.tagService.GetFollowedTags(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get followed tags",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetTagCategories 获取标签分类
func (h *TagHandler) GetTagCategories(c *gin.Context) {
	// 这里应该调用标签分类服务，暂时返回空数据
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []models.TagCategory{},
		"message": "Tag categories feature coming soon",
	})
}

// CreateTagCategory 创建标签分类
func (h *TagHandler) CreateTagCategory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"message": "Tag categories feature coming soon",
	})
}

// GetTagCategory 获取分类详情
func (h *TagHandler) GetTagCategory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"message": "Tag categories feature coming soon",
	})
}

// GetTagStats 获取标签统计
func (h *TagHandler) GetTagStats(c *gin.Context) {
	stats, err := h.tagService.GetTagStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get tag statistics",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetTagTrend 获取标签趋势
func (h *TagHandler) GetTagTrend(c *gin.Context) {
	tagID := c.Param("id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Tag ID is required",
		})
		return
	}

	// 这里应该实现标签趋势分析，暂时返回占位数据
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tag_id":       tagID,
			"tag_name":     "Sample Tag",
			"trend_data":   []interface{}{},
			"current_rank": 1,
			"trend_change": "stable",
		},
		"message": "Tag trend analysis feature coming soon",
	})
}

// BatchOperateTags 批量操作标签
func (h *TagHandler) BatchOperateTags(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		TagIDs    []string               `json:"tag_ids" binding:"required"`
		Operation string                 `json:"operation" binding:"required,oneof=delete update_status update_category"`
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

	err := h.tagService.BatchOperateTags(userID, req.Operation, req.TagIDs, req.Data)
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
		"message": "Batch operation completed successfully",
	})
}