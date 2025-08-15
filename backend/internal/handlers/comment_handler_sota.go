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

// CommentHandlerSOTA - SOTA级别评论处理器，支持多目标评论系统
type CommentHandlerSOTA struct {
	commentService *services.CommentService
}

// NewCommentHandlerSOTA 创建SOTA评论处理器
func NewCommentHandlerSOTA(commentService *services.CommentService) *CommentHandlerSOTA {
	return &CommentHandlerSOTA{
		commentService: commentService,
	}
}

// ================================
// SOTA多目标评论CRUD操作
// ================================

// CreateCommentSOTA 创建评论 - SOTA多目标支持
// POST /api/v2/comments
func (h *CommentHandlerSOTA) CreateCommentSOTA(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.LetterValidationMsg)
		return
	}

	comment, err := h.commentService.CreateCommentSOTA(c.Request.Context(), userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to create comment", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Comment created successfully", comment)
}

// GetCommentsByTargetSOTA 获取目标对象的评论列表 - SOTA通用版本
// GET /api/v2/targets/:target_type/:target_id/comments
func (h *CommentHandlerSOTA) GetCommentsByTargetSOTA(c *gin.Context) {
	targetType := models.CommentType(c.Param("target_type"))
	targetID := c.Param("target_id")

	if targetID == "" {
		utils.BadRequestResponse(c, "Target ID is required", nil)
		return
	}

	// 验证目标类型
	validTypes := map[models.CommentType]bool{
		models.CommentTypeLetter:  true,
		models.CommentTypeProfile: true,
		models.CommentTypeMuseum:  true,
	}

	if !validTypes[targetType] {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	// 获取当前用户ID
	userID, _ := middleware.GetUserID(c)

	// 解析查询参数
	var query models.CommentListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequestResponse(c, "Invalid query parameters", nil)
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	if query.Order == "" {
		query.Order = "desc"
	}
	if query.MaxLevel == 0 {
		query.MaxLevel = 3 // 默认最大3层嵌套
	}

	response, err := h.commentService.GetCommentsByTarget(c.Request.Context(), targetID, targetType, userID, &query)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to get comments", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comments retrieved successfully", response)
}

// GetCommentByIDSOTA 获取单个评论详情 - SOTA增强版
// GET /api/v2/comments/:id
func (h *CommentHandlerSOTA) GetCommentByIDSOTA(c *gin.Context) {
	commentID := c.Param("id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required", nil)
		return
	}

	userID, _ := middleware.GetUserID(c)

	comment, err := h.commentService.GetCommentByIDSOTA(c.Request.Context(), commentID, userID)
	if err != nil {
		utils.NotFoundResponse(c, "Comment not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment retrieved successfully", comment)
}

// ================================
// 专用路由处理器
// ================================

// GetLetterCommentsSOTA 获取信件评论 - 兼容旧版API
// GET /api/v2/letters/:id/comments
func (h *CommentHandlerSOTA) GetLetterCommentsSOTA(c *gin.Context) {
	letterID := c.Param("id")
	if letterID == "" {
		utils.BadRequestResponse(c, "Letter ID is required", nil)
		return
	}

	// 重定向到通用评论API
	c.Params = gin.Params{
		{Key: "target_type", Value: string(models.CommentTypeLetter)},
		{Key: "target_id", Value: letterID},
	}

	h.GetCommentsByTargetSOTA(c)
}

// GetProfileCommentsSOTA 获取个人资料评论
// GET /api/v2/profiles/:id/comments
func (h *CommentHandlerSOTA) GetProfileCommentsSOTA(c *gin.Context) {
	profileID := c.Param("id")
	if profileID == "" {
		utils.BadRequestResponse(c, "Profile ID is required", nil)
		return
	}

	// 重定向到通用评论API
	c.Params = gin.Params{
		{Key: "target_type", Value: string(models.CommentTypeProfile)},
		{Key: "target_id", Value: profileID},
	}

	h.GetCommentsByTargetSOTA(c)
}

// GetMuseumCommentsSOTA 获取博物馆展品评论
// GET /api/v2/museum/:id/comments
func (h *CommentHandlerSOTA) GetMuseumCommentsSOTA(c *gin.Context) {
	museumItemID := c.Param("id")
	if museumItemID == "" {
		utils.BadRequestResponse(c, "Museum item ID is required", nil)
		return
	}

	// 重定向到通用评论API
	c.Params = gin.Params{
		{Key: "target_type", Value: string(models.CommentTypeMuseum)},
		{Key: "target_id", Value: museumItemID},
	}

	h.GetCommentsByTargetSOTA(c)
}

// ================================
// SOTA举报和审核功能
// ================================

// ReportComment 举报评论
// POST /api/v2/comments/:id/report
func (h *CommentHandlerSOTA) ReportComment(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required", nil)
		return
	}

	var req models.CommentReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.LetterValidationMsg)
		return
	}

	err := h.commentService.ReportComment(c.Request.Context(), commentID, userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to report comment", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment reported successfully", nil)
}

// ModerateComment 审核评论 - 管理员功能
// POST /api/v2/comments/:id/moderate
func (h *CommentHandlerSOTA) ModerateComment(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// 检查管理员权限
	userRole, exists := middleware.GetUserRole(c)
	if !exists || (userRole != string(models.RolePlatformAdmin) && userRole != string(models.RoleSuperAdmin)) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required", nil)
		return
	}

	var req models.CommentModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.LetterValidationMsg)
		return
	}

	err := h.commentService.ModerateComment(c.Request.Context(), commentID, userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to moderate comment", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment moderated successfully", nil)
}

// ================================
// SOTA增强功能
// ================================

// GetCommentRepliesSOTA 获取评论回复 - SOTA增强版
// GET /api/v2/comments/:id/replies
func (h *CommentHandlerSOTA) GetCommentRepliesSOTA(c *gin.Context) {
	commentID := c.Param("id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required", nil)
		return
	}

	userID, _ := middleware.GetUserID(c)

	// 首先获取父评论信息
	parentComment, err := h.commentService.GetCommentByIDSOTA(c.Request.Context(), commentID, userID)
	if err != nil {
		utils.NotFoundResponse(c, "Parent comment not found")
		return
	}

	// 解析查询参数
	var query models.CommentListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequestResponse(c, "Invalid query parameters", nil)
		return
	}

	// 设置默认值和特殊参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	if query.Order == "" {
		query.Order = "asc" // 回复默认按时间正序
	}
	query.ParentID = commentID

	response, err := h.commentService.GetCommentsByTarget(
		c.Request.Context(), 
		parentComment.TargetID, 
		parentComment.TargetType, 
		userID, 
		&query,
	)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to get replies", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Replies retrieved successfully", response)
}

// GetCommentStatsSOTA 获取评论统计 - SOTA增强版
// GET /api/v2/targets/:target_type/:target_id/comments/stats
func (h *CommentHandlerSOTA) GetCommentStatsSOTA(c *gin.Context) {
	targetType := models.CommentType(c.Param("target_type"))
	targetID := c.Param("target_id")

	if targetID == "" {
		utils.BadRequestResponse(c, "Target ID is required", nil)
		return
	}

	// 验证目标类型
	validTypes := map[models.CommentType]bool{
		models.CommentTypeLetter:  true,
		models.CommentTypeProfile: true,
		models.CommentTypeMuseum:  true,
	}

	if !validTypes[targetType] {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	userID, _ := middleware.GetUserID(c)

	// 验证权限
	if err := h.commentService.ValidateTarget(c.Request.Context(), userID, targetID, targetType); err != nil {
		utils.BadRequestResponse(c, "Access denied", err)
		return
	}

	stats, err := h.commentService.GetCommentStatsSOTA(c.Request.Context(), targetID, targetType)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to get comment stats", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment stats retrieved successfully", stats)
}

// ToggleCommentLikeSOTA 点赞/取消点赞评论 - SOTA增强版
// POST /api/v2/comments/:id/like
func (h *CommentHandlerSOTA) ToggleCommentLikeSOTA(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required", nil)
		return
	}

	isLiked, likeCount, err := h.commentService.LikeComment(c.Request.Context(), commentID, userID)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to toggle like", err)
		return
	}

	result := map[string]interface{}{
		"is_liked":   isLiked,
		"like_count": likeCount,
		"action":     map[bool]string{true: "liked", false: "unliked"}[isLiked],
	}

	message := "Comment like toggled successfully"
	utils.SuccessResponse(c, http.StatusOK, message, result)
}

// ================================
// SOTA管理功能
// ================================

// GetReportedComments 获取被举报的评论列表 - 管理员功能
// GET /api/v2/admin/comments/reported
func (h *CommentHandlerSOTA) GetReportedComments(c *gin.Context) {
	// 检查管理员权限
	userRole, exists := middleware.GetUserRole(c)
	if !exists || (userRole != string(models.RolePlatformAdmin) && userRole != string(models.RoleSuperAdmin)) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	_ = c.DefaultQuery("status", "pending") // TODO: Implement status filtering

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// TODO: 实现获取举报评论列表的逻辑
	// 这里应该调用相应的服务方法

	result := map[string]interface{}{
		"reports": []interface{}{},
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": 0,
			"pages": 0,
		},
		"stats": map[string]interface{}{
			"pending":   0,
			"resolved":  0,
			"dismissed": 0,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Reported comments retrieved successfully", result)
}

// ================================
// SOTA路由注册
// ================================

// RegisterCommentRoutesSOTA 注册SOTA评论相关路由
func RegisterCommentRoutesSOTA(router *gin.RouterGroup, commentHandler *CommentHandlerSOTA) {
	// V2 API - SOTA多目标评论系统
	v2 := router.Group("/v2")
	{
		// 通用评论API
		comments := v2.Group("/comments")
		{
			comments.POST("", commentHandler.CreateCommentSOTA)              // 创建评论
			comments.GET("/:id", commentHandler.GetCommentByIDSOTA)          // 获取评论详情
			comments.GET("/:id/replies", commentHandler.GetCommentRepliesSOTA) // 获取回复
			comments.POST("/:id/like", commentHandler.ToggleCommentLikeSOTA)  // 点赞/取消点赞
			comments.POST("/:id/report", commentHandler.ReportComment)        // 举报评论
			comments.POST("/:id/moderate", commentHandler.ModerateComment)    // 审核评论(管理员)
		}

		// 多目标评论API
		targets := v2.Group("/targets")
		{
			targets.GET("/:target_type/:target_id/comments", commentHandler.GetCommentsByTargetSOTA)       // 获取目标对象评论
			targets.GET("/:target_type/:target_id/comments/stats", commentHandler.GetCommentStatsSOTA)    // 获取评论统计
		}

		// 专用API路由
		letters := v2.Group("/letters")
		{
			letters.GET("/:id/comments", commentHandler.GetLetterCommentsSOTA) // 信件评论
		}

		profiles := v2.Group("/profiles") 
		{
			profiles.GET("/:id/comments", commentHandler.GetProfileCommentsSOTA) // 个人资料评论
		}

		museum := v2.Group("/museum")
		{
			museum.GET("/:id/comments", commentHandler.GetMuseumCommentsSOTA) // 博物馆评论
		}

		// 管理员功能
		admin := v2.Group("/admin")
		{
			admin.GET("/comments/reported", commentHandler.GetReportedComments) // 获取举报评论列表
		}
	}
}