package services

import (
	"context"
	"fmt"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentService struct {
	db            *gorm.DB
	config        *config.Config
	letterSvc     *LetterService
	creditSvc     *CreditService
	moderationSvc *ModerationService
	securitySvc   *ContentSecurityService
	userSvc       *UserService
}

func NewCommentService(db *gorm.DB, config *config.Config) *CommentService {
	return &CommentService{
		db:     db,
		config: config,
	}
}

// SetLetterService 设置信件服务（避免循环依赖）
func (s *CommentService) SetLetterService(letterSvc *LetterService) {
	s.letterSvc = letterSvc
}

// SetCreditService 设置积分服务（避免循环依赖）
func (s *CommentService) SetCreditService(creditSvc *CreditService) {
	s.creditSvc = creditSvc
}

// SetModerationService 设置审核服务（避免循环依赖）
func (s *CommentService) SetModerationService(moderationSvc *ModerationService) {
	s.moderationSvc = moderationSvc
}

// SetContentSecurityService 设置内容安全服务（避免循环依赖）
func (s *CommentService) SetContentSecurityService(securitySvc *ContentSecurityService) {
	s.securitySvc = securitySvc
}

// SetUserService 设置用户服务（避免循环依赖）
func (s *CommentService) SetUserService(userSvc *UserService) {
	s.userSvc = userSvc
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(ctx context.Context, userID string, req *models.CommentCreateRequest) (*models.CommentResponse, error) {
	// 验证信件是否存在
	var letter models.Letter
	if err := s.db.Where("id = ? AND visibility IN (?)", req.LetterID, []string{"public", "school"}).First(&letter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("letter not found or not accessible")
		}
		return nil, fmt.Errorf("failed to verify letter: %w", err)
	}

	// 如果是回复，验证父评论是否存在
	if req.ParentID != nil {
		var parentComment models.Comment
		if err := s.db.Where("id = ? AND letter_id = ?", *req.ParentID, req.LetterID).First(&parentComment).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("parent comment not found")
			}
			return nil, fmt.Errorf("failed to verify parent comment: %w", err)
		}
	}

	// 1. 优先进行内容安全检查（XSS防护和内容清理）
	var cleanedContent string = req.Content
	if s.securitySvc != nil {
		securityResult, err := s.securitySvc.ValidateCommentContent(req.Content, userID)
		if err != nil {
			return nil, fmt.Errorf("content security validation failed: %w", err)
		}
		
		// 如果检测到XSS或高风险内容，直接拒绝
		if securityResult.XSSDetected || !securityResult.IsSafe {
			return nil, fmt.Errorf("comment content contains security violations: %s", 
				fmt.Sprintf("Risk level: %s, Violations: %v", securityResult.RiskLevel, securityResult.ViolationType))
		}
		
		// 使用清理后的内容
		if securityResult.HTMLCleaned {
			cleanedContent = securityResult.SanitizedContent
		}
		
		// 如果需要人工审核，记录但允许创建（待审核状态）
		if securityResult.RequiresModeration {
			// 内容将以pending状态创建，等待人工审核
		}
	}

	// 2. 内容审核（传统审核流程）
	var commentStatus models.CommentStatus = models.CommentStatusActive
	if s.moderationSvc != nil {
		moderationReq := &models.ModerationRequest{
			UserID:      userID,
			ContentType: models.ContentTypeComment,
			ContentID:   uuid.New().String(), // 临时ID，将在创建后更新
			Content:     cleanedContent, // 使用安全清理后的内容
		}
		response, err := s.moderationSvc.ModerateContent(ctx, moderationReq)
		if err != nil {
			return nil, fmt.Errorf("moderation check failed: %w", err)
		}
		if response.Status == models.ModerationRejected {
			reasons := ""
			if len(response.Reasons) > 0 {
				reasons = response.Reasons[0] // 使用第一个拒绝原因
			}
			return nil, fmt.Errorf("comment content rejected: %s", reasons)
		}
		
		// 如果审核结果需要等待，设置为pending状态
		if response.NeedReview {
			commentStatus = models.CommentStatusPending
		}
	}

	// 创建评论（使用安全清理后的内容和状态）
	comment := &models.Comment{
		ID:         uuid.New().String(),
		LetterID:   req.LetterID,
		UserID:     userID,
		ParentID:   req.ParentID,
		Content:    cleanedContent, // 使用安全清理后的内容
		Status:     commentStatus,  // 使用安全检查后确定的状态
		LikeCount:  0,
		ReplyCount: 0,
		IsTop:      false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 数据库事务
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 创建评论
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 如果是回复，更新父评论的回复数
		if req.ParentID != nil {
			if err := tx.Model(&models.Comment{}).Where("id = ?", *req.ParentID).
				Update("reply_count", gorm.Expr("reply_count + 1")).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// 奖励积分
	if s.creditSvc != nil {
		go func() {
			s.creditSvc.AddPoints(userID, 2, "发表评论", comment.ID)
		}()
	}

	// 返回完整评论信息
	return s.GetCommentByID(ctx, comment.ID, userID)
}

// GetCommentsByLetterID 获取信件的评论列表
func (s *CommentService) GetCommentsByLetterID(ctx context.Context, letterID string, userID string, query *models.CommentListQuery) ([]models.CommentResponse, int64, error) {
	// 验证信件是否存在且用户有权限访问
	var letter models.Letter
	if err := s.db.Where("id = ?", letterID).First(&letter).Error; err != nil {
		return nil, 0, fmt.Errorf("letter not found")
	}

	// 构建查询
	db := s.db.Model(&models.Comment{}).
		Where("letter_id = ? AND status = ?", letterID, models.CommentStatusActive)

	// 如果只获取顶级评论
	if query.OnlyTopLevel {
		db = db.Where("parent_id IS NULL")
	}

	// 如果获取特定评论的回复
	if query.ParentID != "" {
		db = db.Where("parent_id = ?", query.ParentID)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	// 排序和分页
	orderClause := fmt.Sprintf("%s %s", query.SortBy, query.Order)
	offset := (query.Page - 1) * query.Limit

	var comments []models.Comment
	if err := db.Order(orderClause).
		Limit(query.Limit).
		Offset(offset).
		Preload("User").
		Find(&comments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get comments: %w", err)
	}

	// 转换为响应格式
	responses := make([]models.CommentResponse, len(comments))
	for i, comment := range comments {
		response, err := s.buildCommentResponse(&comment, userID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to build comment response: %w", err)
		}
		responses[i] = *response
	}

	return responses, total, nil
}

// GetCommentByID 获取单个评论详情
func (s *CommentService) GetCommentByID(ctx context.Context, commentID string, userID string) (*models.CommentResponse, error) {
	var comment models.Comment
	if err := s.db.Where("id = ?", commentID).
		Preload("User").
		First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return s.buildCommentResponse(&comment, userID)
}

// UpdateComment 更新评论
func (s *CommentService) UpdateComment(ctx context.Context, commentID string, userID string, req *models.CommentUpdateRequest) (*models.CommentResponse, error) {
	var comment models.Comment
	if err := s.db.Where("id = ?", commentID).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	// 权限检查
	if !comment.CanEdit(userID) {
		return nil, fmt.Errorf("permission denied")
	}

	// 内容审核
	if s.moderationSvc != nil {
		moderationReq := &models.ModerationRequest{
			UserID:      userID,
			ContentType: models.ContentTypeComment,
			ContentID:   uuid.New().String(), // 临时ID，将在创建后更新
			Content:     req.Content,
		}
		response, err := s.moderationSvc.ModerateContent(ctx, moderationReq)
		if err != nil {
			return nil, fmt.Errorf("moderation check failed: %w", err)
		}
		if response.Status == models.ModerationRejected {
			reasons := ""
			if len(response.Reasons) > 0 {
				reasons = response.Reasons[0] // 使用第一个拒绝原因
			}
			return nil, fmt.Errorf("comment content rejected: %s", reasons)
		}
	}

	// 更新评论
	updates := map[string]interface{}{
		"content":    req.Content,
		"updated_at": time.Now(),
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := s.db.Model(&comment).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return s.GetCommentByID(ctx, commentID, userID)
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(ctx context.Context, commentID string, userID string, userRole models.UserRole) error {
	var comment models.Comment
	if err := s.db.Where("id = ?", commentID).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("comment not found")
		}
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// 权限检查
	if !comment.CanDelete(userID, userRole) {
		return fmt.Errorf("permission denied")
	}

	// 软删除评论
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 软删除评论
		if err := tx.Model(&comment).Update("status", models.CommentStatusDeleted).Error; err != nil {
			return err
		}

		// 如果是回复，减少父评论的回复数
		if comment.ParentID != nil {
			if err := tx.Model(&models.Comment{}).Where("id = ?", *comment.ParentID).
				Update("reply_count", gorm.Expr("reply_count - 1")).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// LikeComment 点赞评论
func (s *CommentService) LikeComment(ctx context.Context, commentID string, userID string) (bool, int, error) {
	// 检查评论是否存在
	var comment models.Comment
	if err := s.db.Where("id = ? AND status = ?", commentID, models.CommentStatusActive).First(&comment).Error; err != nil {
		return false, 0, fmt.Errorf("comment not found")
	}

	// 检查是否已点赞
	var existingLike models.CommentLike
	err := s.db.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&existingLike).Error

	if err == nil {
		// 已点赞，取消点赞
		return s.unlikeComment(commentID, userID)
	} else if err == gorm.ErrRecordNotFound {
		// 未点赞，添加点赞
		return s.addCommentLike(commentID, userID)
	} else {
		return false, 0, fmt.Errorf("failed to check like status: %w", err)
	}
}

// addCommentLike 添加点赞
func (s *CommentService) addCommentLike(commentID string, userID string) (bool, int, error) {
	var likeCount int
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 创建点赞记录
		like := &models.CommentLike{
			ID:        uuid.New().String(),
			CommentID: commentID,
			UserID:    userID,
			CreatedAt: time.Now(),
		}

		if err := tx.Create(like).Error; err != nil {
			return err
		}

		// 更新评论点赞数
		if err := tx.Model(&models.Comment{}).Where("id = ?", commentID).
			Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			return err
		}

		// 获取更新后的点赞数
		var comment models.Comment
		if err := tx.Where("id = ?", commentID).First(&comment).Error; err != nil {
			return err
		}
		likeCount = comment.LikeCount

		return nil
	})

	if err != nil {
		return false, 0, fmt.Errorf("failed to add like: %w", err)
	}

	// 奖励积分
	if s.creditSvc != nil {
		go func() {
			s.creditSvc.AddPoints(userID, 1, "评论点赞", commentID)
		}()
	}

	return true, likeCount, nil
}

// unlikeComment 取消点赞
func (s *CommentService) unlikeComment(commentID string, userID string) (bool, int, error) {
	var likeCount int
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 删除点赞记录
		if err := tx.Where("comment_id = ? AND user_id = ?", commentID, userID).
			Delete(&models.CommentLike{}).Error; err != nil {
			return err
		}

		// 更新评论点赞数
		if err := tx.Model(&models.Comment{}).Where("id = ?", commentID).
			Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
			return err
		}

		// 获取更新后的点赞数
		var comment models.Comment
		if err := tx.Where("id = ?", commentID).First(&comment).Error; err != nil {
			return err
		}
		likeCount = comment.LikeCount

		return nil
	})

	if err != nil {
		return false, 0, fmt.Errorf("failed to remove like: %w", err)
	}

	return false, likeCount, nil
}

// buildCommentResponse 构建评论响应
func (s *CommentService) buildCommentResponse(comment *models.Comment, userID string) (*models.CommentResponse, error) {
	// 检查当前用户是否点赞
	var isLiked bool
	if userID != "" {
		var like models.CommentLike
		err := s.db.Where("comment_id = ? AND user_id = ?", comment.ID, userID).First(&like).Error
		isLiked = (err == nil)
	}

	// 构建用户信息
	var userInfo *models.UserBasicInfo
	if comment.User != nil {
		userInfo = &models.UserBasicInfo{
			ID:       comment.User.ID,
			Username: comment.User.Username,
			Nickname: comment.User.Nickname,
			Avatar:   comment.User.Avatar,
		}
	}

	response := &models.CommentResponse{
		ID:         comment.ID,
		LetterID:   comment.LetterID,
		UserID:     comment.UserID,
		ParentID:   comment.ParentID,
		Content:    comment.Content,
		Status:     comment.Status,
		LikeCount:  comment.LikeCount,
		ReplyCount: comment.ReplyCount,
		IsTop:      comment.IsTop,
		CreatedAt:  comment.CreatedAt,
		UpdatedAt:  comment.UpdatedAt,
		User:       userInfo,
		IsLiked:    isLiked,
	}

	return response, nil
}

// GetCommentStats 获取评论统计
func (s *CommentService) GetCommentStats(ctx context.Context, letterID string) (int64, error) {
	var count int64
	if err := s.db.Model(&models.Comment{}).
		Where("letter_id = ? AND status = ?", letterID, models.CommentStatusActive).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}

	return count, nil
}

// BatchOperateComments 批量操作评论
func (s *CommentService) BatchOperateComments(ctx context.Context, userID string, userRole models.UserRole, commentIDs []string, operation string, data map[string]interface{}) error {
	if len(commentIDs) == 0 {
		return fmt.Errorf("no comment IDs provided")
	}

	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	switch operation {
	case "delete":
		// 根据权限删除评论
		query := tx.Model(&models.Comment{}).Where("id IN (?)", commentIDs)
		
		// 非管理员只能删除自己的评论
		if !s.isAdminRole(userRole) {
			query = query.Where("user_id = ?", userID)
		}
		
		// 软删除
		if err := query.Update("status", models.CommentStatusDeleted).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete comments: %w", err)
		}

	case "approve":
		// 只有管理员可以批量审核
		if !s.isAdminRole(userRole) {
			return fmt.Errorf("permission denied: only admins can approve comments")
		}
		
		if err := tx.Model(&models.Comment{}).
			Where("id IN (?)", commentIDs).
			Update("status", models.CommentStatusActive).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to approve comments: %w", err)
		}

	case "reject":
		// 只有管理员可以批量拒绝
		if !s.isAdminRole(userRole) {
			return fmt.Errorf("permission denied: only admins can reject comments")
		}
		
		if err := tx.Model(&models.Comment{}).
			Where("id IN (?)", commentIDs).
			Update("status", models.CommentStatusRejected).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to reject comments: %w", err)
		}

	case "hide":
		// 管理员可以隐藏任何评论，用户只能隐藏自己的评论
		query := tx.Model(&models.Comment{}).Where("id IN (?)", commentIDs)
		
		if !s.isAdminRole(userRole) {
			query = query.Where("user_id = ?", userID)
		}
		
		if err := query.Update("status", models.CommentStatusHidden).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to hide comments: %w", err)
		}

	case "show":
		// 管理员可以显示任何评论，用户只能显示自己的评论
		query := tx.Model(&models.Comment{}).Where("id IN (?)", commentIDs)
		
		if !s.isAdminRole(userRole) {
			query = query.Where("user_id = ?", userID)
		}
		
		if err := query.Update("status", models.CommentStatusActive).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to show comments: %w", err)
		}

	case "moderate":
		// 只有管理员可以批量设为待审核
		if !s.isAdminRole(userRole) {
			return fmt.Errorf("permission denied: only admins can moderate comments")
		}
		
		if err := tx.Model(&models.Comment{}).
			Where("id IN (?)", commentIDs).
			Update("status", models.CommentStatusPending).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to moderate comments: %w", err)
		}

	default:
		return fmt.Errorf("unsupported operation: %s", operation)
	}

	return tx.Commit().Error
}

// isAdminRole 检查是否为管理员角色
func (s *CommentService) isAdminRole(role models.UserRole) bool {
	adminRoles := []models.UserRole{
		models.RolePlatformAdmin,
		models.RoleSuperAdmin,
		models.RoleCourierLevel3,
		models.RoleCourierLevel4,
	}
	
	for _, adminRole := range adminRoles {
		if role == adminRole {
			return true
		}
	}
	
	return false
}
