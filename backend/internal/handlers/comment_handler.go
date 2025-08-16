package handlers

import (
	"net/http"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentService *services.CommentService
}

func NewCommentHandler(commentService *services.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// CreateComment 创建评论
// POST /api/v1/comments
func (h *CommentHandler) CreateComment(c *gin.Context) {

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

	comment, err := h.commentService.CreateComment(c.Request.Context(), userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Request failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Comment created successfully", comment)
}

// GetCommentsByLetterID 获取信件的评论列表
// GET /api/v1/letters/:letter_id/comments
func (h *CommentHandler) GetCommentsByLetterID(c *gin.Context) {

	letterID := c.Param("id")
	if letterID == "" {
		utils.BadRequestResponse(c, "Letter ID is required", nil)
		return
	}

	// 获取当前用户ID（可能为空，用于判断是否点赞）
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
	query.LetterID = letterID

	comments, total, err := h.commentService.GetCommentsByLetterID(c.Request.Context(), letterID, userID, &query)
	if err != nil {
		utils.BadRequestResponse(c, "Request failed", err)
		return
	}

	// 构建分页响应
	pagination := map[string]interface{}{
		"page":  query.Page,
		"limit": query.Limit,
		"total": total,
		"pages": (total + int64(query.Limit) - 1) / int64(query.Limit),
	}

	result := map[string]interface{}{
		"comments":   comments,
		"pagination": pagination,
	}

	utils.SuccessResponse(c, http.StatusOK, "Comments retrieved successfully", result)
}

// GetCommentReplies 获取评论的回复列表
// GET /api/v1/comments/:comment_id/replies
func (h *CommentHandler) GetCommentReplies(c *gin.Context) {

	commentID := c.Param("id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required", nil)
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

	// 首先获取父评论信息以获取letter_id
	parentComment, err := h.commentService.GetCommentByID(c.Request.Context(), commentID, userID)
	if err != nil {
		utils.BadRequestResponse(c, "Parent comment not found", nil)
		return
	}

	query.LetterID = parentComment.LetterID

	replies, total, err := h.commentService.GetCommentsByLetterID(c.Request.Context(), parentComment.LetterID, userID, &query)
	if err != nil {
		utils.BadRequestResponse(c, "Request failed", err)
		return
	}

	pagination := map[string]interface{}{
		"page":  query.Page,
		"limit": query.Limit,
		"total": total,
		"pages": (total + int64(query.Limit) - 1) / int64(query.Limit),
	}

	result := map[string]interface{}{
		"replies":    replies,
		"pagination": pagination,
	}

	utils.SuccessResponse(c, http.StatusOK, "Replies retrieved successfully", result)
}

// GetCommentByID 获取单个评论详情
// GET /api/v1/comments/:id
func (h *CommentHandler) GetCommentByID(c *gin.Context) {

	commentID := c.Param("id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required", nil)
		return
	}

	userID, _ := middleware.GetUserID(c)

	comment, err := h.commentService.GetCommentByID(c.Request.Context(), commentID, userID)
	if err != nil {
		utils.NotFoundResponse(c, "Comment not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment retrieved successfully", comment)
}

// UpdateComment 更新评论
// PUT /api/v1/comments/:id
func (h *CommentHandler) UpdateComment(c *gin.Context) {

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

	var req models.CommentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.LetterValidationMsg)
		return
	}

	comment, err := h.commentService.UpdateComment(c.Request.Context(), commentID, userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Request failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment updated successfully", comment)
}

// DeleteComment 删除评论
// DELETE /api/v1/comments/:id
func (h *CommentHandler) DeleteComment(c *gin.Context) {

	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// 获取用户角色
	userRole, exists := middleware.GetUserRole(c)
	if !exists {
		utils.BadRequestResponse(c, "User role not found", nil)
		return
	}

	commentID := c.Param("id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required", nil)
		return
	}

	err := h.commentService.DeleteComment(c.Request.Context(), commentID, userID, models.UserRole(userRole))
	if err != nil {
		utils.BadRequestResponse(c, "Request failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment deleted successfully", nil)
}

// LikeComment 点赞/取消点赞评论
// POST /api/v1/comments/:id/like
func (h *CommentHandler) LikeComment(c *gin.Context) {

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
		utils.BadRequestResponse(c, "Request failed", err)
		return
	}

	result := map[string]interface{}{
		"is_liked":   isLiked,
		"like_count": likeCount,
	}

	message := "Comment unliked successfully"
	if isLiked {
		message = "Comment liked successfully"
	}

	utils.SuccessResponse(c, http.StatusOK, message, result)
}

// GetCommentStats 获取信件的评论统计
// GET /api/v1/letters/:letter_id/comments/stats
func (h *CommentHandler) GetCommentStats(c *gin.Context) {

	letterID := c.Param("id")
	if letterID == "" {
		utils.BadRequestResponse(c, "Letter ID is required", nil)
		return
	}

	count, err := h.commentService.GetCommentStats(c.Request.Context(), letterID)
	if err != nil {
		utils.BadRequestResponse(c, "Request failed", err)
		return
	}

	result := map[string]interface{}{
		"comment_count": count,
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment stats retrieved successfully", result)
}

// BatchOperateComments 批量操作评论
// POST /api/v1/comments/batch
func (h *CommentHandler) BatchOperateComments(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// 获取用户角色
	userRole, exists := middleware.GetUserRole(c)
	if !exists {
		utils.BadRequestResponse(c, "User role not found", nil)
		return
	}

	var req struct {
		CommentIDs []string               `json:"comment_ids" binding:"required"`
		Operation  string                 `json:"operation" binding:"required,oneof=delete approve reject hide show moderate"`
		Data       map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.LetterValidationMsg)
		return
	}

	err := h.commentService.BatchOperateComments(c.Request.Context(), userID, models.UserRole(userRole), req.CommentIDs, req.Operation, req.Data)
	if err != nil {
		utils.BadRequestResponse(c, "Batch operation failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Batch operation completed successfully", nil)
}

// RegisterCommentRoutes 注册评论相关路由
func RegisterCommentRoutes(router *gin.RouterGroup, commentHandler *CommentHandler) {
	comments := router.Group("/comments")
	{
		// 基础CRUD
		comments.POST("", commentHandler.CreateComment)                // 创建评论
		comments.GET("/:id", commentHandler.GetCommentByID)            // 获取评论详情
		comments.PUT("/:id", commentHandler.UpdateComment)             // 更新评论
		comments.DELETE("/:id", commentHandler.DeleteComment)          // 删除评论
		comments.POST("/:id/like", commentHandler.LikeComment)         // 点赞/取消点赞
		comments.GET("/:id/replies", commentHandler.GetCommentReplies) // 获取回复列表
		
		// 批量操作
		comments.POST("/batch", commentHandler.BatchOperateComments)   // 批量操作评论
	}

	letters := router.Group("/letters")
	{
		letters.GET("/:id/comments", commentHandler.GetCommentsByLetterID) // 获取信件评论列表
		letters.GET("/:id/comments/stats", commentHandler.GetCommentStats) // 获取评论统计
	}
}
