package handlers

import (
	"net/http"
	"strconv"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"
	
	"github.com/gin-gonic/gin"
)

type FollowHandler struct {
	followService *services.FollowService
}

func NewFollowHandler(followService *services.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

// FollowUser 关注用户
// @Summary 关注用户
// @Description 关注指定用户
// @Tags Follow
// @Accept json
// @Produce json
// @Param request body models.FollowActionRequest true "关注请求"
// @Success 200 {object} utils.Response{data=models.FollowActionResponse}
// @Router /api/v1/follow/users [post]
func (h *FollowHandler) FollowUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req models.FollowActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	response, err := h.followService.FollowUser(userID, req.UserID, req.NotificationEnabled)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to follow user", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response)
}

// UnfollowUser 取消关注用户
// @Summary 取消关注用户
// @Description 取消关注指定用户
// @Tags Follow
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} utils.Response{data=models.FollowActionResponse}
// @Router /api/v1/follow/users/{user_id} [delete]
func (h *FollowHandler) UnfollowUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	response, err := h.followService.UnfollowUser(userID, targetUserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to unfollow user", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response)
}

// GetFollowers 获取用户的粉丝列表
// @Summary 获取粉丝列表
// @Description 获取指定用户的粉丝列表
// @Tags Follow
// @Accept json
// @Produce json
// @Param user_id path string false "用户ID（为空则获取当前用户）"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页大小" default(20)
// @Param sort_by query string false "排序字段" Enums(created_at, nickname, username)
// @Param order query string false "排序方式" Enums(asc, desc) default(desc)
// @Param search query string false "搜索关键词"
// @Param school_filter query string false "学校过滤"
// @Success 200 {object} utils.Response{data=models.FollowListResponse}
// @Router /api/v1/follow/followers [get]
// @Router /api/v1/follow/users/{user_id}/followers [get]
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	targetUserID := c.Param("user_id")
	
	// 如果没有指定用户ID，则使用当前用户
	if targetUserID == "" {
		targetUserID = currentUserID
	}

	if targetUserID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	req := h.parseFollowListRequest(c)
	
	response, err := h.followService.GetFollowers(targetUserID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get followers", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response)
}

// GetFollowing 获取用户的关注列表
// @Summary 获取关注列表
// @Description 获取指定用户的关注列表
// @Tags Follow
// @Accept json
// @Produce json
// @Param user_id path string false "用户ID（为空则获取当前用户）"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页大小" default(20)
// @Param sort_by query string false "排序字段" Enums(created_at, nickname, username)
// @Param order query string false "排序方式" Enums(asc, desc) default(desc)
// @Param search query string false "搜索关键词"
// @Param school_filter query string false "学校过滤"
// @Success 200 {object} utils.Response{data=models.FollowListResponse}
// @Router /api/v1/follow/following [get]
// @Router /api/v1/follow/users/{user_id}/following [get]
func (h *FollowHandler) GetFollowing(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	targetUserID := c.Param("user_id")
	
	// 如果没有指定用户ID，则使用当前用户
	if targetUserID == "" {
		targetUserID = currentUserID
	}

	if targetUserID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	req := h.parseFollowListRequest(c)
	
	response, err := h.followService.GetFollowing(targetUserID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get following", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response)
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Description 根据关键词搜索用户
// @Tags Follow
// @Accept json
// @Produce json
// @Param query query string false "搜索关键词"
// @Param school_code query string false "学校代码"
// @Param role query string false "用户角色"
// @Param min_followers query int false "最小粉丝数"
// @Param max_followers query int false "最大粉丝数"
// @Param active_since query string false "活跃时间过滤"
// @Param sort_by query string false "排序字段" Enums(followers, activity, joined, relevance) default(relevance)
// @Param order query string false "排序方式" Enums(asc, desc) default(desc)
// @Param limit query int false "返回数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} utils.Response{data=models.UserSearchResponse}
// @Router /api/v1/follow/users/search [get]
func (h *FollowHandler) SearchUsers(c *gin.Context) {
	currentUserID := c.GetString("user_id")

	req := h.parseUserSearchRequest(c)
	
	response, err := h.followService.SearchUsers(req, currentUserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search users", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response)
}

// GetUserSuggestions 获取用户推荐
// @Summary 获取用户推荐
// @Description 获取推荐的用户列表
// @Tags Follow
// @Accept json
// @Produce json
// @Param limit query int false "返回数量" default(10)
// @Param based_on query string false "推荐基础" Enums(school, mutual_followers, activity, interests) default(school)
// @Param exclude_followed query bool false "排除已关注用户" default(true)
// @Param min_activity_score query number false "最小活跃度分数"
// @Success 200 {object} utils.Response{data=models.UserSuggestionsResponse}
// @Router /api/v1/follow/suggestions [get]
func (h *FollowHandler) GetUserSuggestions(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	req := h.parseUserSuggestionsRequest(c)
	
	response, err := h.followService.GetUserSuggestions(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user suggestions", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response)
}

// GetFollowStatus 获取关注状态
// @Summary 获取关注状态
// @Description 检查当前用户与指定用户的关注关系
// @Tags Follow
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} utils.Response{data=object}
// @Router /api/v1/follow/users/{user_id}/status [get]
func (h *FollowHandler) GetFollowStatus(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	if currentUserID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	status, err := h.followService.GetFollowStatus(currentUserID, targetUserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get follow status", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", status)
}

// RefreshSuggestions 刷新推荐
// @Summary 刷新推荐
// @Description 刷新用户推荐列表
// @Tags Follow
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=models.UserSuggestionsResponse}
// @Router /api/v1/follow/suggestions/refresh [post]
func (h *FollowHandler) RefreshSuggestions(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// 刷新推荐，使用默认参数
	req := &models.UserSuggestionsRequest{
		Limit:            10,
		BasedOn:          "school",
		ExcludeFollowed:  true,
		MinActivityScore: 0.0,
	}
	
	response, err := h.followService.GetUserSuggestions(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to refresh suggestions", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response)
}

// Helper functions for parsing request parameters

func (h *FollowHandler) parseFollowListRequest(c *gin.Context) *models.FollowListRequest {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	// 限制最大每页数量
	if limit > 100 {
		limit = 100
	}
	
	return &models.FollowListRequest{
		Page:         page,
		Limit:        limit,
		SortBy:       c.DefaultQuery("sort_by", "created_at"),
		Order:        c.DefaultQuery("order", "desc"),
		Search:       c.Query("search"),
		SchoolFilter: c.Query("school_filter"),
	}
}

func (h *FollowHandler) parseUserSearchRequest(c *gin.Context) *models.UserSearchRequest {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	minFollowers, _ := strconv.Atoi(c.Query("min_followers"))
	maxFollowers, _ := strconv.Atoi(c.Query("max_followers"))
	
	// 限制最大返回数量
	if limit > 100 {
		limit = 100
	}
	
	return &models.UserSearchRequest{
		Query:        c.Query("query"),
		SchoolCode:   c.Query("school_code"),
		Role:         c.Query("role"),
		MinFollowers: minFollowers,
		MaxFollowers: maxFollowers,
		ActiveSince:  c.Query("active_since"),
		SortBy:       c.DefaultQuery("sort_by", "relevance"),
		Order:        c.DefaultQuery("order", "desc"),
		Limit:        limit,
		Offset:       offset,
	}
}

func (h *FollowHandler) parseUserSuggestionsRequest(c *gin.Context) *models.UserSuggestionsRequest {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	excludeFollowed, _ := strconv.ParseBool(c.DefaultQuery("exclude_followed", "true"))
	minActivityScore, _ := strconv.ParseFloat(c.Query("min_activity_score"), 64)
	
	// 限制最大推荐数量
	if limit > 50 {
		limit = 50
	}
	
	return &models.UserSuggestionsRequest{
		Limit:             limit,
		BasedOn:           c.DefaultQuery("based_on", "school"),
		ExcludeFollowed:   excludeFollowed,
		MinActivityScore:  minActivityScore,
	}
}

// FollowMultipleUsers 批量关注用户
// @Summary 批量关注用户
// @Description 一次关注多个用户
// @Tags Follow
// @Accept json
// @Produce json
// @Param request body object{user_ids=[]string} true "用户ID列表"
// @Success 200 {object} utils.Response{data=object}
// @Router /api/v1/follow/users/batch [post]
func (h *FollowHandler) FollowMultipleUsers(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	successCount := 0
	failedCount := 0
	var errors []string

	for _, targetUserID := range req.UserIDs {
		_, err := h.followService.FollowUser(userID, targetUserID, true)
		if err != nil {
			failedCount++
			errors = append(errors, targetUserID+": "+err.Error())
		} else {
			successCount++
		}
	}

	response := map[string]interface{}{
		"success":       successCount,
		"failed":        failedCount,
		"total":         len(req.UserIDs),
		"errors":        errors,
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response)
}

// RemoveFollower 移除粉丝
// @Summary 移除粉丝
// @Description 从粉丝列表中移除指定用户
// @Tags Follow
// @Accept json
// @Produce json
// @Param user_id path string true "粉丝用户ID"
// @Success 200 {object} utils.Response{data=object}
// @Router /api/v1/follow/followers/{user_id} [delete]
func (h *FollowHandler) RemoveFollower(c *gin.Context) {
	currentUserID := c.GetString("user_id")
	if currentUserID == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	followerUserID := c.Param("user_id")
	if followerUserID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Follower user ID is required", nil)
		return
	}

	// 移除粉丝实际上是让对方取消对当前用户的关注
	response, err := h.followService.UnfollowUser(followerUserID, currentUserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove follower", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", response)
}