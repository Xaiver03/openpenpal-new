package services

import (
	"context"
	"fmt"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================
// SOTA增强功能 - 多目标评论系统扩展
// ================================

// ValidateTarget 验证目标对象是否存在及权限
func (s *CommentService) ValidateTarget(ctx context.Context, userID, targetID string, targetType models.CommentType) error {
	switch targetType {
	case models.CommentTypeLetter:
		return s.validateLetterTarget(ctx, userID, targetID)
	case models.CommentTypeProfile:
		return s.validateProfileTarget(ctx, userID, targetID)
	case models.CommentTypeMuseum:
		return s.validateMuseumTarget(ctx, userID, targetID)
	default:
		return fmt.Errorf("unsupported comment target type: %s", targetType)
	}
}

// validateLetterTarget 验证信件目标
func (s *CommentService) validateLetterTarget(ctx context.Context, userID, letterID string) error {
	var letter models.Letter
	if err := s.db.Where("id = ?", letterID).First(&letter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("letter not found")
		}
		return fmt.Errorf("failed to verify letter: %w", err)
	}

	// 检查可见性权限
	switch letter.Visibility {
	case "public":
		return nil
	case "school":
		// TODO: 实现学校权限检查
		return nil
	case "private":
		// For private letters, only the sender/author can comment
		if letter.UserID != userID && letter.AuthorID != userID {
			return fmt.Errorf("letter not accessible")
		}
		return nil
	default:
		return fmt.Errorf("letter not accessible")
	}
}

// validateProfileTarget 验证个人资料目标
func (s *CommentService) validateProfileTarget(ctx context.Context, userID, profileUserID string) error {
	var user models.User
	if err := s.db.Where("id = ?", profileUserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user profile not found")
		}
		return fmt.Errorf("failed to verify user profile: %w", err)
	}

	// 检查隐私设置
	if s.privacySvc != nil {
		// TODO: 实现隐私权限检查
	}

	return nil
}

// validateMuseumTarget 验证博物馆目标
func (s *CommentService) validateMuseumTarget(ctx context.Context, userID, museumItemID string) error {
	// TODO: 实现博物馆项目验证
	return nil
}

// validateParentComment 验证父评论
func (s *CommentService) validateParentComment(ctx context.Context, parentID, targetID string, targetType models.CommentType) error {
	var parentComment models.Comment
	if err := s.db.Where("id = ?", parentID).First(&parentComment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("parent comment not found")
		}
		return fmt.Errorf("failed to verify parent comment: %w", err)
	}

	// 检查父评论是否属于同一目标
	if parentComment.TargetID != targetID || parentComment.TargetType != targetType {
		return fmt.Errorf("parent comment does not belong to the same target")
	}

	// 检查是否可以添加回复
	if !parentComment.CanAddReply() {
		return fmt.Errorf("cannot add reply to this comment")
	}

	return nil
}

// isSpamContent SOTA垃圾内容检测
func (s *CommentService) isSpamContent(content string) bool {
	// 使用模型中的垃圾检测逻辑
	comment := &models.Comment{Content: content}
	return comment.IsSpamLikely()
}

// detectLanguage 检测评论语言
func (s *CommentService) detectLanguage(content string) string {
	// 简单的中英文检测
	chineseCount := 0
	englishCount := 0
	
	for _, r := range content {
		if r >= '\u4e00' && r <= '\u9fff' {
			chineseCount++
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			englishCount++
		}
	}
	
	if chineseCount > englishCount {
		return "zh"
	} else if englishCount > 0 {
		return "en"
	}
	return "unknown"
}

// calculateCommentPoints SOTA积分计算
func (s *CommentService) calculateCommentPoints(targetType models.CommentType, isReply bool) int {
	basePoints := map[models.CommentType]int{
		models.CommentTypeLetter:  2,
		models.CommentTypeProfile: 1,
		models.CommentTypeMuseum:  3,
	}
	
	points := basePoints[targetType]
	if isReply {
		points = points / 2 // 回复的积分减半
	}
	
	return points
}

// getCommentAction 获取积分动作描述
func (s *CommentService) getCommentAction(targetType models.CommentType, isReply bool) string {
	if isReply {
		return fmt.Sprintf("回复%s评论", s.getTargetTypeName(targetType))
	}
	return fmt.Sprintf("评论%s", s.getTargetTypeName(targetType))
}

// getTargetTypeName 获取目标类型名称
func (s *CommentService) getTargetTypeName(targetType models.CommentType) string {
	names := map[models.CommentType]string{
		models.CommentTypeLetter:  "信件",
		models.CommentTypeProfile: "个人资料",
		models.CommentTypeMuseum:  "博物馆展品",
	}
	return names[targetType]
}

// ================================
// SOTA多目标评论创建 - 增强版本
// ================================

// CreateCommentSOTA 创建评论 - SOTA多目标评论系统
func (s *CommentService) CreateCommentSOTA(ctx context.Context, userID string, req *models.CommentCreateRequest) (*models.CommentResponse, error) {
	// 兼容性处理：如果提供了旧版LetterID但未提供新版字段，则转换
	if req.LetterID != "" && req.TargetID == "" {
		req.TargetID = req.LetterID
		req.TargetType = models.CommentTypeLetter
	}

	// 验证目标对象是否存在及权限
	if err := s.ValidateTarget(ctx, userID, req.TargetID, req.TargetType); err != nil {
		return nil, err
	}

	// 如果是回复，验证父评论是否存在且属于同一目标
	if req.ParentID != nil {
		if err := s.validateParentComment(ctx, *req.ParentID, req.TargetID, req.TargetType); err != nil {
			return nil, err
		}
	}

	// SOTA垃圾内容检测
	if s.isSpamContent(req.Content) {
		return nil, fmt.Errorf("detected spam content, comment rejected")
	}

	// 内容审核
	commentID := uuid.New().String()
	if s.moderationSvc != nil {
		moderationReq := &models.ModerationRequest{
			UserID:      userID,
			ContentType: models.ContentTypeComment,
			ContentID:   commentID,
			Content:     req.Content,
		}
		response, err := s.moderationSvc.ModerateContent(ctx, moderationReq)
		if err != nil {
			return nil, fmt.Errorf("moderation check failed: %w", err)
		}
		if response.Status == models.ModerationRejected {
			reasons := ""
			if len(response.Reasons) > 0 {
				reasons = response.Reasons[0]
			}
			return nil, fmt.Errorf("comment content rejected: %s", reasons)
		}
	}

	// 创建评论 - SOTA增强版
	comment := &models.Comment{
		ID:          commentID,
		TargetID:    req.TargetID,
		TargetType:  req.TargetType,
		UserID:      userID,
		ParentID:    req.ParentID,
		Content:     req.Content,
		Status:      models.CommentStatusActive,
		IsAnonymous: req.IsAnonymous,
		LikeCount:   0,
		ReplyCount:  0,
		ReportCount: 0,
		Language:    s.detectLanguage(req.Content),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 兼容性：设置旧版字段
	if req.TargetType == models.CommentTypeLetter {
		comment.LetterID = &req.TargetID
	}

	// 数据库事务
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 创建评论（GORM钩子会自动处理层级路径）
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 如果是回复，更新父评论的回复数和根评论的回复计数
		if req.ParentID != nil {
			if err := tx.Model(&models.Comment{}).Where("id = ?", *req.ParentID).
				Update("reply_count", gorm.Expr("reply_count + 1")).Error; err != nil {
				return err
			}

			// 如果有根评论，也更新根评论的统计
			if comment.RootID != nil {
				if err := tx.Model(&models.Comment{}).Where("id = ?", *comment.RootID).
					Update("reply_count", gorm.Expr("reply_count + 1")).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// SOTA积分奖励系统
	if s.creditSvc != nil {
		go func() {
			points := s.calculateCommentPoints(req.TargetType, req.ParentID != nil)
			action := s.getCommentAction(req.TargetType, req.ParentID != nil)
			s.creditSvc.AddPoints(userID, points, action, comment.ID)
		}()
	}

	// 返回完整评论信息
	return s.GetCommentByIDSOTA(ctx, comment.ID, userID)
}

// ================================
// SOTA多目标评论查询
// ================================

// GetCommentsByTarget 获取目标对象的评论列表 - SOTA通用版本
func (s *CommentService) GetCommentsByTarget(ctx context.Context, targetID string, targetType models.CommentType, userID string, query *models.CommentListQuery) (*models.CommentListResponse, error) {
	// 验证目标对象权限
	if err := s.ValidateTarget(ctx, userID, targetID, targetType); err != nil {
		return nil, err
	}

	// 构建查询
	db := s.db.Model(&models.Comment{}).
		Where("target_id = ? AND target_type = ? AND status IN (?)", 
			targetID, targetType, []models.CommentStatus{models.CommentStatusActive})

	// 如果只获取顶级评论
	if query.OnlyTopLevel {
		db = db.Where("parent_id IS NULL")
	}

	// 如果获取特定评论的回复
	if query.ParentID != "" {
		db = db.Where("parent_id = ?", query.ParentID)
	}

	// 如果获取根评论的所有回复
	if query.RootID != "" {
		db = db.Where("root_id = ?", query.RootID)
	}

	// 最大嵌套层级限制
	if query.MaxLevel > 0 {
		db = db.Where("level <= ?", query.MaxLevel)
	}

	// 作者过滤
	if query.AuthorID != "" {
		db = db.Where("user_id = ?", query.AuthorID)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count comments: %w", err)
	}

	// 排序和分页
	orderClause := fmt.Sprintf("%s %s", query.SortBy, query.Order)
	offset := (query.Page - 1) * query.Limit

	var comments []models.Comment
	dbQuery := db.Order(orderClause).Limit(query.Limit).Offset(offset)

	// 预加载关联数据
	if query.IncludeReplies && !query.OnlyTopLevel {
		dbQuery = dbQuery.Preload("User").Preload("Replies.User")
	} else {
		dbQuery = dbQuery.Preload("User")
	}

	if err := dbQuery.Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	// 转换为响应格式
	responses := make([]models.CommentResponse, len(comments))
	for i, comment := range comments {
		response, err := s.buildEnhancedCommentResponse(&comment, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to build comment response: %w", err)
		}
		responses[i] = *response
	}

	// 获取统计信息
	stats, err := s.GetCommentStatsSOTA(ctx, targetID, targetType)
	if err != nil {
		stats = models.CommentStats{} // 使用空统计信息
	}

	return &models.CommentListResponse{
		Comments:   responses,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: int((total + int64(query.Limit) - 1) / int64(query.Limit)),
		Stats:      stats,
	}, nil
}

// GetCommentByIDSOTA 获取单个评论详情 - SOTA增强版
func (s *CommentService) GetCommentByIDSOTA(ctx context.Context, commentID string, userID string) (*models.CommentResponse, error) {
	var comment models.Comment
	if err := s.db.Where("id = ?", commentID).
		Preload("User").
		First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return s.buildEnhancedCommentResponse(&comment, userID)
}

// buildEnhancedCommentResponse 构建增强的评论响应 - SOTA版本
func (s *CommentService) buildEnhancedCommentResponse(comment *models.Comment, userID string) (*models.CommentResponse, error) {
	// 检查当前用户是否点赞
	var isLiked bool
	if userID != "" {
		var like models.CommentLike
		err := s.db.Where("comment_id = ? AND user_id = ? AND is_like = ?", comment.ID, userID, true).First(&like).Error
		isLiked = (err == nil)
	}

	// 构建用户信息（考虑匿名）
	var userInfo *models.UserBasicInfo
	if comment.User != nil {
		userInfo = comment.GetDisplayAuthor(&models.UserBasicInfo{
			ID:       comment.User.ID,
			Username: comment.User.Username,
			Nickname: comment.User.Nickname,
			Avatar:   comment.User.Avatar,
		})
	}

	// 权限检查
	canEdit := comment.CanEdit(userID)
	canDelete := false
	canReport := comment.CanReport(userID)
	
	if s.userSvc != nil && userID != "" {
		// TODO: 获取用户角色进行权限检查
		canDelete = comment.UserID == userID
	}

	response := &models.CommentResponse{
		ID:           comment.ID,
		TargetID:     comment.TargetID,
		TargetType:   comment.TargetType,
		LetterID:     comment.LetterID, // 兼容性
		UserID:       comment.UserID,
		ParentID:     comment.ParentID,
		RootID:       comment.RootID,
		Content:      comment.GetDisplayContent(),
		Status:       comment.Status,
		Level:        comment.Level,
		Path:         comment.Path,
		LikeCount:    comment.LikeCount,
		ReplyCount:   comment.ReplyCount,
		ReportCount:  comment.ReportCount,
		NetLikes:     comment.GetNetLikes(),
		IsTop:        comment.IsTop,
		IsAnonymous:  comment.IsAnonymous,
		IsEdited:     comment.IsEdited,
		CreatedAt:    comment.CreatedAt,
		UpdatedAt:    comment.UpdatedAt,
		User:         userInfo,
		IsLiked:      isLiked,
		CanEdit:      canEdit,
		CanDelete:    canDelete,
		CanReport:    canReport,
	}

	return response, nil
}

// GetCommentStatsSOTA 获取评论统计信息 - SOTA版本
func (s *CommentService) GetCommentStatsSOTA(ctx context.Context, targetID string, targetType models.CommentType) (models.CommentStats, error) {
	stats := models.CommentStats{}

	// 总评论数
	var totalComments int64
	s.db.Model(&models.Comment{}).Where("target_id = ? AND target_type = ?", targetID, targetType).Count(&totalComments)
	stats.TotalComments = int(totalComments)

	// 总回复数
	var totalReplies int64
	s.db.Model(&models.Comment{}).Where("target_id = ? AND target_type = ? AND parent_id IS NOT NULL", targetID, targetType).Count(&totalReplies)
	stats.TotalReplies = int(totalReplies)

	// 活跃评论数
	var activeComments int64
	s.db.Model(&models.Comment{}).Where("target_id = ? AND target_type = ? AND status = ?", targetID, targetType, models.CommentStatusActive).Count(&activeComments)
	stats.ActiveComments = int(activeComments)

	// 待审核评论数
	var pendingComments int64
	s.db.Model(&models.Comment{}).Where("target_id = ? AND target_type = ? AND status = ?", targetID, targetType, models.CommentStatusPending).Count(&pendingComments)
	stats.PendingComments = int(pendingComments)

	// 被举报评论数
	var reportedComments int64
	s.db.Model(&models.Comment{}).Where("target_id = ? AND target_type = ? AND report_count > 0", targetID, targetType).Count(&reportedComments)
	stats.ReportedComments = int(reportedComments)

	// 总点赞数
	s.db.Raw(`
		SELECT COALESCE(SUM(like_count), 0) 
		FROM comments 
		WHERE target_id = ? AND target_type = ? AND status = ?
	`, targetID, targetType, models.CommentStatusActive).Scan(&stats.TotalLikes)

	return stats, nil
}

// ================================
// SOTA举报和审核功能
// ================================

// ReportComment 举报评论
func (s *CommentService) ReportComment(ctx context.Context, commentID string, reporterID string, req *models.CommentReportRequest) error {
	// 检查评论是否存在
	var comment models.Comment
	if err := s.db.Where("id = ? AND status = ?", commentID, models.CommentStatusActive).First(&comment).Error; err != nil {
		return fmt.Errorf("comment not found")
	}

	// 检查是否可以举报
	if !comment.CanReport(reporterID) {
		return fmt.Errorf("cannot report this comment")
	}

	// 检查是否已经举报过
	var existingReport models.CommentReport
	if err := s.db.Where("comment_id = ? AND reporter_id = ?", commentID, reporterID).First(&existingReport).Error; err == nil {
		return fmt.Errorf("comment already reported by this user")
	}

	// 创建举报记录
	report := &models.CommentReport{
		ID:          uuid.New().String(),
		CommentID:   commentID,
		ReporterID:  reporterID,
		Reason:      req.Reason,
		Description: req.Description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 创建举报记录
		if err := tx.Create(report).Error; err != nil {
			return err
		}

		// 更新评论举报计数
		if err := tx.Model(&models.Comment{}).Where("id = ?", commentID).
			Update("report_count", gorm.Expr("report_count + 1")).Error; err != nil {
			return err
		}

		return nil
	})
}

// ModerateComment 审核评论
func (s *CommentService) ModerateComment(ctx context.Context, commentID string, moderatorID string, req *models.CommentModerationRequest) error {
	var comment models.Comment
	if err := s.db.Where("id = ?", commentID).First(&comment).Error; err != nil {
		return fmt.Errorf("comment not found")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":             req.Status,
		"moderated_at":       &now,
		"moderated_by":       &moderatorID,
		"moderation_reason":  req.Reason,
		"updated_at":         now,
	}

	return s.db.Model(&comment).Updates(updates).Error
}