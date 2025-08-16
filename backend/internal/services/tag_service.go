package services

import (
	"fmt"
	"strings"

	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TagService 标签服务
type TagService struct {
	db        *gorm.DB
	aiService *AIService
}

// NewTagService 创建标签服务
func NewTagService(db *gorm.DB) *TagService {
	return &TagService{
		db: db,
	}
}

// SetAIService 设置AI服务依赖
func (s *TagService) SetAIService(aiService *AIService) {
	s.aiService = aiService
}

// CreateTag 创建标签
func (s *TagService) CreateTag(userID string, req *models.TagRequest) (*models.Tag, error) {
	tag := &models.Tag{
		ID:          uuid.New().String(),
		Name:        strings.TrimSpace(req.Name),
		DisplayName: req.DisplayName,
		Description: req.Description,
		Type:        models.TagTypeUser,
		Status:      models.TagStatusActive,
		Color:       req.Color,
		Icon:        req.Icon,
		CategoryID:  req.CategoryID,
		CreatedBy:   userID,
	}

	if err := s.db.Create(tag).Error; err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

// GetTag 获取标签详情
func (s *TagService) GetTag(tagID string, userID string) (*models.TagResponse, error) {
	var tag models.Tag
	err := s.db.Preload("Category").Preload("SubTags").First(&tag, "id = ?", tagID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	// 获取内容数量
	var contentCount int64
	s.db.Model(&models.ContentTag{}).Where("tag_id = ?", tagID).Count(&contentCount)

	// 检查用户是否关注该标签
	var isFollowed bool
	if userID != "" {
		var followCount int64
		s.db.Model(&models.UserTagFollow{}).Where("user_id = ? AND tag_id = ?", userID, tagID).Count(&followCount)
		isFollowed = followCount > 0
	}

	return &models.TagResponse{
		Tag:          tag,
		ContentCount: contentCount,
		IsFollowed:   isFollowed,
	}, nil
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(tagID string, userID string, req *models.TagRequest) (*models.Tag, error) {
	var tag models.Tag
	if err := s.db.First(&tag, "id = ?", tagID).Error; err != nil {
		return nil, fmt.Errorf("tag not found")
	}

	// 权限检查：只有创建者或管理员可以更新
	if tag.CreatedBy != userID {
		return nil, fmt.Errorf("permission denied")
	}

	// 更新字段
	updates := map[string]interface{}{
		"display_name": req.DisplayName,
		"description":  req.Description,
		"color":        req.Color,
		"icon":         req.Icon,
		"category_id":  req.CategoryID,
	}

	if err := s.db.Model(&tag).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return &tag, nil
}

// DeleteTag 删除标签
func (s *TagService) DeleteTag(tagID string, userID string) error {
	var tag models.Tag
	if err := s.db.First(&tag, "id = ?", tagID).Error; err != nil {
		return fmt.Errorf("tag not found")
	}

	// 权限检查
	if tag.CreatedBy != userID {
		return fmt.Errorf("permission denied")
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除相关的内容标签关联
	if err := tx.Where("tag_id = ?", tagID).Delete(&models.ContentTag{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete content tags: %w", err)
	}

	// 删除用户关注关系
	if err := tx.Where("tag_id = ?", tagID).Delete(&models.UserTagFollow{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete tag follows: %w", err)
	}

	// 删除标签
	if err := tx.Delete(&tag).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return tx.Commit().Error
}

// SearchTags 搜索标签
func (s *TagService) SearchTags(req *models.TagSearchRequest) (*models.TagListResponse, error) {
	query := s.db.Model(&models.Tag{}).Preload("Category")

	// 搜索条件
	if req.Query != "" {
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", "%"+req.Query+"%", "%"+req.Query+"%")
	}

	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 分页和排序
	offset := (req.Page - 1) * req.Limit
	if req.SortBy != "" {
		order := req.SortBy
		if req.SortOrder == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("usage_count DESC, created_at DESC")
	}

	var tags []models.Tag
	if err := query.Offset(offset).Limit(req.Limit).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to search tags: %w", err)
	}

	// 构建响应
	tagResponses := make([]models.TagResponse, len(tags))
	for i, tag := range tags {
		var contentCount int64
		s.db.Model(&models.ContentTag{}).Where("tag_id = ?", tag.ID).Count(&contentCount)
		
		tagResponses[i] = models.TagResponse{
			Tag:          tag,
			ContentCount: contentCount,
			IsFollowed:   false, // 在搜索结果中不计算关注状态以提高性能
		}
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &models.TagListResponse{
		Tags:       tagResponses,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetPopularTags 获取热门标签
func (s *TagService) GetPopularTags(limit int) ([]models.Tag, error) {
	var tags []models.Tag
	err := s.db.Where("status = ?", models.TagStatusActive).
		Order("usage_count DESC, trending_score DESC").
		Limit(limit).
		Find(&tags).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get popular tags: %w", err)
	}

	return tags, nil
}

// GetTrendingTags 获取趋势标签
func (s *TagService) GetTrendingTags(limit int) ([]models.Tag, error) {
	var tags []models.Tag
	err := s.db.Where("status = ? AND trending_score > 0", models.TagStatusActive).
		Order("trending_score DESC, usage_count DESC").
		Limit(limit).
		Find(&tags).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get trending tags: %w", err)
	}

	return tags, nil
}

// SuggestTags AI标签建议
func (s *TagService) SuggestTags(req *models.TagSuggestionRequest) (*models.TagSuggestionResponse, error) {
	response := &models.TagSuggestionResponse{
		SuggestedTags:   []models.Tag{},
		AIGeneratedTags: []models.Tag{},
		RelatedTags:     []models.Tag{},
		PopularTags:     []models.Tag{},
	}

	// 获取热门标签
	popularTags, err := s.GetPopularTags(10)
	if err == nil {
		response.PopularTags = popularTags
	}

	// 如果有内容，尝试AI生成标签
	// TODO: Implement AI tag generation when AI service supports it
	/*
	if req.Content != "" && s.aiService != nil {
		aiTags, err := s.generateAITags(req.Content, req.Limit)
		if err == nil {
			response.AIGeneratedTags = aiTags
		}
	}
	*/

	// 基于现有标签的相关建议
	if req.ContentID != "" {
		relatedTags, err := s.getRelatedTags(req.ContentType, req.ContentID, req.Limit)
		if err == nil {
			response.RelatedTags = relatedTags
		}
	}

	return response, nil
}

// generateAITags AI生成标签 - 暂时禁用
func (s *TagService) generateAITags(content string, limit int) ([]models.Tag, error) {
	// TODO: Implement when AI service supports GenerateContent method
	return []models.Tag{}, nil
}

// parseAITagResponse 解析AI标签响应
func (s *TagService) parseAITagResponse(content string) []string {
	// 简化的解析逻辑，实际应该使用JSON解析
	content = strings.TrimSpace(content)
	
	// 移除可能的JSON格式标记
	content = strings.ReplaceAll(content, "{", "")
	content = strings.ReplaceAll(content, "}", "")
	content = strings.ReplaceAll(content, "\"", "")
	content = strings.ReplaceAll(content, "tags:", "")
	content = strings.ReplaceAll(content, "[", "")
	content = strings.ReplaceAll(content, "]", "")
	
	// 按逗号分割
	parts := strings.Split(content, ",")
	var tags []string
	
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" && len(tag) >= 2 && len(tag) <= 10 {
			tags = append(tags, tag)
		}
	}
	
	return tags
}

// getRelatedTags 获取相关标签
func (s *TagService) getRelatedTags(contentType, contentID string, limit int) ([]models.Tag, error) {
	var tags []models.Tag
	
	// 通过子查询获取同一内容类型的相关标签
	subQuery := s.db.Model(&models.ContentTag{}).
		Select("tag_id").
		Where("content_type = ? AND content_id != ?", contentType, contentID)
	
	err := s.db.Where("id IN (?)", subQuery).
		Where("status = ?", models.TagStatusActive).
		Order("usage_count DESC").
		Limit(limit).
		Find(&tags).Error
		
	if err != nil {
		return nil, fmt.Errorf("failed to get related tags: %w", err)
	}
	
	return tags, nil
}

// TagContent 为内容添加标签
func (s *TagService) TagContent(req *models.ContentTagRequest, userID string) error {
	// 验证内容是否存在
	if err := s.validateContentExists(req.ContentType, req.ContentID); err != nil {
		return err
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除现有标签关联
	if err := tx.Where("content_type = ? AND content_id = ?", req.ContentType, req.ContentID).
		Delete(&models.ContentTag{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove existing tags: %w", err)
	}

	// 添加新的标签关联
	for _, tagID := range req.TagIDs {
		contentTag := &models.ContentTag{
			ID:          uuid.New().String(),
			ContentType: req.ContentType,
			ContentID:   req.ContentID,
			TagID:       tagID,
			Source:      "user",
			Confidence:  1.0,
			CreatedBy:   userID,
		}

		if err := tx.Create(contentTag).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create content tag: %w", err)
		}

		// 更新标签使用计数
		if err := tx.Model(&models.Tag{}).Where("id = ?", tagID).
			UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update tag usage count: %w", err)
		}
	}

	return tx.Commit().Error
}

// UntagContent 移除内容标签
func (s *TagService) UntagContent(contentType, contentID string, tagIDs []string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除标签关联
	query := tx.Where("content_type = ? AND content_id = ?", contentType, contentID)
	if len(tagIDs) > 0 {
		query = query.Where("tag_id IN (?)", tagIDs)
	}

	if err := query.Delete(&models.ContentTag{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove content tags: %w", err)
	}

	// 更新标签使用计数
	for _, tagID := range tagIDs {
		if err := tx.Model(&models.Tag{}).Where("id = ?", tagID).
			UpdateColumn("usage_count", gorm.Expr("usage_count - 1")).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update tag usage count: %w", err)
		}
	}

	return tx.Commit().Error
}

// GetContentTags 获取内容的标签
func (s *TagService) GetContentTags(contentType, contentID string) (*models.ContentTagsResponse, error) {
	var contentTags []models.ContentTag
	err := s.db.Preload("Tag").
		Where("content_type = ? AND content_id = ?", contentType, contentID).
		Find(&contentTags).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get content tags: %w", err)
	}

	tags := make([]models.Tag, len(contentTags))
	for i, ct := range contentTags {
		tags[i] = ct.Tag
	}

	return &models.ContentTagsResponse{
		ContentType: contentType,
		ContentID:   contentID,
		Tags:        tags,
		TagCount:    len(tags),
	}, nil
}

// validateContentExists 验证内容是否存在
func (s *TagService) validateContentExists(contentType, contentID string) error {
	var count int64
	
	switch contentType {
	case "letter":
		s.db.Model(&models.Letter{}).Where("id = ?", contentID).Count(&count)
	case "comment":
		s.db.Model(&models.Comment{}).Where("id = ?", contentID).Count(&count)
	case "museum":
		s.db.Model(&models.MuseumItem{}).Where("id = ?", contentID).Count(&count)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
	
	if count == 0 {
		return fmt.Errorf("content not found")
	}
	
	return nil
}

// FollowTag 关注标签
func (s *TagService) FollowTag(userID, tagID string) error {
	// 检查是否已关注
	var count int64
	s.db.Model(&models.UserTagFollow{}).
		Where("user_id = ? AND tag_id = ?", userID, tagID).
		Count(&count)
	
	if count > 0 {
		return fmt.Errorf("already following this tag")
	}

	follow := &models.UserTagFollow{
		ID:     uuid.New().String(),
		UserID: userID,
		TagID:  tagID,
	}

	return s.db.Create(follow).Error
}

// UnfollowTag 取消关注标签
func (s *TagService) UnfollowTag(userID, tagID string) error {
	return s.db.Where("user_id = ? AND tag_id = ?", userID, tagID).
		Delete(&models.UserTagFollow{}).Error
}

// GetFollowedTags 获取用户关注的标签
func (s *TagService) GetFollowedTags(userID string, page, limit int) (*models.TagListResponse, error) {
	var follows []models.UserTagFollow
	
	offset := (page - 1) * limit
	err := s.db.Preload("Tag").
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&follows).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get followed tags: %w", err)
	}

	// 计算总数
	var total int64
	s.db.Model(&models.UserTagFollow{}).Where("user_id = ?", userID).Count(&total)

	// 构建响应
	tagResponses := make([]models.TagResponse, len(follows))
	for i, follow := range follows {
		var contentCount int64
		s.db.Model(&models.ContentTag{}).Where("tag_id = ?", follow.TagID).Count(&contentCount)
		
		tagResponses[i] = models.TagResponse{
			Tag:          follow.Tag,
			ContentCount: contentCount,
			IsFollowed:   true,
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &models.TagListResponse{
		Tags:       tagResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetTagStats 获取标签统计
func (s *TagService) GetTagStats() (*models.TagStatsResponse, error) {
	stats := &models.TagStatsResponse{}

	// 基础统计
	s.db.Model(&models.Tag{}).Count(&stats.TotalTags)
	s.db.Model(&models.Tag{}).Where("type = ?", models.TagTypeUser).Count(&stats.UserTags)
	s.db.Model(&models.Tag{}).Where("type = ?", models.TagTypeSystem).Count(&stats.SystemTags)
	s.db.Model(&models.Tag{}).Where("type = ?", models.TagTypeAI).Count(&stats.AITags)
	
	// 使用统计
	var totalUsage struct {
		Sum float64
		Avg float64
	}
	s.db.Model(&models.Tag{}).Select("SUM(usage_count) as sum, AVG(usage_count) as avg").Scan(&totalUsage)
	stats.TotalUsage = int64(totalUsage.Sum)
	stats.AvgUsagePerTag = totalUsage.Avg

	// 热门标签
	popularTags, _ := s.GetPopularTags(10)
	stats.PopularTags = popularTags

	// 趋势标签
	trendingTags, _ := s.GetTrendingTags(10)
	stats.TrendingTags = trendingTags

	return stats, nil
}

// BatchOperateTags 批量操作标签
func (s *TagService) BatchOperateTags(userID string, operation string, tagIDs []string, data map[string]interface{}) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	switch operation {
	case "delete":
		// 批量删除标签（只能删除自己创建的）
		if err := tx.Where("id IN (?) AND created_by = ?", tagIDs, userID).
			Delete(&models.Tag{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to batch delete tags: %w", err)
		}
		
		// 删除相关关联
		tx.Where("tag_id IN (?)", tagIDs).Delete(&models.ContentTag{})
		tx.Where("tag_id IN (?)", tagIDs).Delete(&models.UserTagFollow{})

	case "update_status":
		status, ok := data["status"].(string)
		if !ok {
			return fmt.Errorf("invalid status")
		}
		
		if err := tx.Model(&models.Tag{}).
			Where("id IN (?) AND created_by = ?", tagIDs, userID).
			Update("status", status).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to batch update tag status: %w", err)
		}

	case "update_category":
		categoryID, ok := data["category_id"].(string)
		if !ok {
			return fmt.Errorf("invalid category_id")
		}
		
		if err := tx.Model(&models.Tag{}).
			Where("id IN (?) AND created_by = ?", tagIDs, userID).
			Update("category_id", categoryID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to batch update tag category: %w", err)
		}

	default:
		return fmt.Errorf("unsupported operation: %s", operation)
	}

	return tx.Commit().Error
}