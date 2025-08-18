package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

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

// =============== 标签CRUD操作 ===============

// CreateTag 创建标签
func (s *TagService) CreateTag(req *models.TagRequest, createdBy string) (*models.Tag, error) {
	// 检查标签名是否已存在
	var existingTag models.Tag
	if err := s.db.Where("name = ? OR display_name = ?", req.Name, req.DisplayName).First(&existingTag).Error; err == nil {
		return nil, fmt.Errorf("tag with name '%s' already exists", req.Name)
	}

	tag := &models.Tag{
		ID:          uuid.New().String(),
		Name:        strings.ToLower(strings.TrimSpace(req.Name)),
		DisplayName: req.DisplayName,
		Description: req.Description,
		Type:        models.TagTypeUser,
		Status:      models.TagStatusActive,
		Color:       req.Color,
		Icon:        req.Icon,
		CategoryID:  req.CategoryID,
		UsageCount:  0,
		TrendingScore: 0,
		CreatedBy:   createdBy,
	}

	// 如果没有提供显示名称，使用标签名
	if tag.DisplayName == "" {
		tag.DisplayName = req.Name
	}

	// 如果没有提供颜色，生成随机颜色
	if tag.Color == "" {
		tag.Color = s.generateRandomColor()
	}

	if err := s.db.Create(tag).Error; err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

// GetTag 获取标签详情
func (s *TagService) GetTag(tagID string) (*models.Tag, error) {
	var tag models.Tag
	if err := s.db.Preload("Category").First(&tag, "id = ?", tagID).Error; err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// 计算内容数量
	var contentCount int64
	s.db.Model(&models.ContentTag{}).Where("tag_id = ?", tagID).Count(&contentCount)
	tag.ContentCount = contentCount

	return &tag, nil
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(tagID string, req *models.TagRequest) (*models.Tag, error) {
	var tag models.Tag
	if err := s.db.First(&tag, "id = ?", tagID).Error; err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// 检查新名称是否与其他标签冲突
	if req.Name != tag.Name {
		var existingTag models.Tag
		if err := s.db.Where("(name = ? OR display_name = ?) AND id != ?", 
			req.Name, req.DisplayName, tagID).First(&existingTag).Error; err == nil {
			return nil, fmt.Errorf("tag with name '%s' already exists", req.Name)
		}
	}

	// 更新字段
	tag.Name = strings.ToLower(strings.TrimSpace(req.Name))
	tag.DisplayName = req.DisplayName
	tag.Description = req.Description
	tag.Color = req.Color
	tag.Icon = req.Icon
	tag.CategoryID = req.CategoryID

	if tag.DisplayName == "" {
		tag.DisplayName = req.Name
	}

	if err := s.db.Save(&tag).Error; err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return &tag, nil
}

// DeleteTag 删除标签（软删除）
func (s *TagService) DeleteTag(tagID string) error {
	// 检查标签是否存在
	var tag models.Tag
	if err := s.db.First(&tag, "id = ?", tagID).Error; err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	// 将标签状态设为非活跃，而不是直接删除
	tag.Status = models.TagStatusInactive
	if err := s.db.Save(&tag).Error; err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// =============== 标签搜索和发现 ===============

// SearchTags 搜索标签
func (s *TagService) SearchTags(req *models.TagSearchRequest) (*models.TagListResponse, error) {
	query := s.db.Model(&models.Tag{}).Preload("Category")

	// 构建查询条件
	if req.Query != "" {
		searchTerm := "%" + strings.ToLower(req.Query) + "%"
		query = query.Where("(LOWER(name) LIKE ? OR LOWER(display_name) LIKE ? OR LOWER(description) LIKE ?)", 
			searchTerm, searchTerm, searchTerm)
	}

	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	} else {
		// 默认只显示活跃标签
		query = query.Where("status = ?", models.TagStatusActive)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count tags: %w", err)
	}

	// 应用排序
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "usage_count"
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 应用分页
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit
	query = query.Offset(offset).Limit(req.Limit)

	// 执行查询
	var tags []models.Tag
	if err := query.Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to search tags: %w", err)
	}

	// 转换为响应格式
	tagResponses := make([]models.TagResponse, len(tags))
	for i, tag := range tags {
		// 计算内容数量
		var contentCount int64
		s.db.Model(&models.ContentTag{}).Where("tag_id = ?", tag.ID).Count(&contentCount)
		
		tagResponses[i] = models.TagResponse{
			Tag:          tag,
			ContentCount: contentCount,
			IsFollowed:   false, // TODO: 实现用户关注检查
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

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
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var tags []models.Tag
	if err := s.db.Where("status = ?", models.TagStatusActive).
		Order("usage_count DESC, trending_score DESC").
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get popular tags: %w", err)
	}

	return tags, nil
}

// GetTrendingTags 获取趋势标签
func (s *TagService) GetTrendingTags(limit int) ([]models.Tag, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var tags []models.Tag
	if err := s.db.Where("status = ?", models.TagStatusActive).
		Order("trending_score DESC, usage_count DESC").
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get trending tags: %w", err)
	}

	return tags, nil
}

// =============== 内容标签关系管理 ===============

// TagContent 为内容添加标签
func (s *TagService) TagContent(req *models.ContentTagRequest, userID string) error {
	// 验证内容存在性（根据内容类型）
	if err := s.validateContentExists(req.ContentType, req.ContentID); err != nil {
		return err
	}

	// 开始事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有的标签关联
		if err := tx.Where("content_type = ? AND content_id = ?", 
			req.ContentType, req.ContentID).Delete(&models.ContentTag{}).Error; err != nil {
			return fmt.Errorf("failed to remove existing tags: %w", err)
		}

		// 添加新的标签关联
		for _, tagID := range req.TagIDs {
			// 验证标签存在
			var tag models.Tag
			if err := tx.First(&tag, "id = ? AND status = ?", tagID, models.TagStatusActive).Error; err != nil {
				return fmt.Errorf("tag %s not found or inactive", tagID)
			}

			// 创建内容标签关联
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
				return fmt.Errorf("failed to create content tag: %w", err)
			}

			// 更新标签使用次数
			if err := tx.Model(&tag).UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error; err != nil {
				log.Printf("Warning: failed to update tag usage count: %v", err)
			}
		}

		return nil
	})
}

// GetContentTags 获取内容的标签
func (s *TagService) GetContentTags(contentType, contentID string) (*models.ContentTagsResponse, error) {
	var contentTags []models.ContentTag
	if err := s.db.Preload("Tag").Where("content_type = ? AND content_id = ?", 
		contentType, contentID).Find(&contentTags).Error; err != nil {
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

// RemoveContentTag 移除内容标签
func (s *TagService) RemoveContentTag(contentType, contentID, tagID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除标签关联
		result := tx.Where("content_type = ? AND content_id = ? AND tag_id = ?", 
			contentType, contentID, tagID).Delete(&models.ContentTag{})
		
		if result.Error != nil {
			return fmt.Errorf("failed to remove content tag: %w", result.Error)
		}

		if result.RowsAffected > 0 {
			// 更新标签使用次数
			if err := tx.Model(&models.Tag{}).Where("id = ?", tagID).
				UpdateColumn("usage_count", gorm.Expr("usage_count - 1")).Error; err != nil {
				log.Printf("Warning: failed to update tag usage count: %v", err)
			}
		}

		return nil
	})
}

// =============== AI标签建议 ===============

// SuggestTags AI标签建议
func (s *TagService) SuggestTags(req *models.TagSuggestionRequest) (*models.TagSuggestionResponse, error) {
	response := &models.TagSuggestionResponse{
		SuggestedTags:   []models.Tag{},
		AIGeneratedTags: []models.Tag{},
		RelatedTags:     []models.Tag{},
		PopularTags:     []models.Tag{},
	}

	// 获取热门标签
	popularTags, err := s.GetPopularTags(req.Limit)
	if err == nil {
		response.PopularTags = popularTags
	}

	// 如果有内容，使用AI生成标签建议
	if req.Content != "" && s.aiService != nil {
		aiTags, err := s.generateAITags(req.Content, req.Limit)
		if err != nil {
			log.Printf("Warning: AI tag generation failed: %v", err)
		} else {
			response.AIGeneratedTags = aiTags
		}
	}

	// 如果指定了内容ID，获取相关标签
	if req.ContentID != "" {
		relatedTags, err := s.getRelatedTags(req.ContentType, req.ContentID, req.Limit)
		if err == nil {
			response.RelatedTags = relatedTags
		}
	}

	// 基于内容类型的推荐标签
	typeBasedTags, err := s.getTypeBasedTags(req.ContentType, req.Limit)
	if err == nil {
		response.SuggestedTags = typeBasedTags
	}

	return response, nil
}

// generateAITags 使用AI生成标签
func (s *TagService) generateAITags(content string, limit int) ([]models.Tag, error) {
	if s.aiService == nil {
		return []models.Tag{}, fmt.Errorf("AI service not available")
	}

	// 调用AI服务（简化实现）
	// TODO: 实现真正的AI标签生成，当前AI服务没有GenerateContent方法
	log.Printf("AI tag generation requested for content (length: %d), using simplified implementation", len(content))
	
	// 简化的模拟实现，返回JSON格式的标签
	aiResponse := struct {
		Content string
	}{
		Content: `{"tags": ["写作", "日记", "生活", "情感", "思考"]}`, // 模拟AI生成的标签JSON
	}

	// 解析AI响应
	var aiResult struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(aiResponse.Content), &aiResult); err != nil {
		log.Printf("Warning: failed to parse AI tag response: %v", err)
		// 如果JSON解析失败，尝试简单的文本处理
		return s.parseAITagsFromText(aiResponse.Content, limit), nil
	}

	// 处理AI生成的标签
	var aiTags []models.Tag
	for i, tagName := range aiResult.Tags {
		if i >= limit {
			break
		}

		// 检查标签是否已存在
		var existingTag models.Tag
		tagName = strings.ToLower(strings.TrimSpace(tagName))
		if err := s.db.Where("name = ?", tagName).First(&existingTag).Error; err == nil {
			aiTags = append(aiTags, existingTag)
		} else {
			// 创建新的AI标签
			newTag := models.Tag{
				ID:          uuid.New().String(),
				Name:        tagName,
				DisplayName: tagName,
				Type:        models.TagTypeAI,
				Status:      models.TagStatusActive,
				Color:       s.generateRandomColor(),
				UsageCount:  0,
				TrendingScore: 0,
				CreatedBy:   "ai-system",
			}
			
			if err := s.db.Create(&newTag).Error; err == nil {
				aiTags = append(aiTags, newTag)
			}
		}
	}

	return aiTags, nil
}

// parseAITagsFromText 从AI文本响应中提取标签
func (s *TagService) parseAITagsFromText(text string, limit int) []models.Tag {
	// 简单的文本解析逻辑
	lines := strings.Split(text, "\n")
	var tagNames []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 移除常见的标记符号
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimPrefix(line, "•")
		line = strings.TrimSpace(line)
		
		if line != "" && len(tagNames) < limit {
			tagNames = append(tagNames, strings.ToLower(line))
		}
	}

	// 转换为标签对象
	var tags []models.Tag
	for _, tagName := range tagNames {
		var existingTag models.Tag
		if err := s.db.Where("name = ?", tagName).First(&existingTag).Error; err == nil {
			tags = append(tags, existingTag)
		}
	}

	return tags
}

// getRelatedTags 获取相关标签
func (s *TagService) getRelatedTags(contentType, contentID string, limit int) ([]models.Tag, error) {
	// 获取相同类型内容的常用标签
	query := `
		SELECT t.* FROM tags t
		JOIN content_tags ct ON t.id = ct.tag_id
		WHERE ct.content_type = ? AND ct.content_id != ?
		GROUP BY t.id
		ORDER BY COUNT(*) DESC
		LIMIT ?
	`

	var tags []models.Tag
	if err := s.db.Raw(query, contentType, contentID, limit).Scan(&tags).Error; err != nil {
		return []models.Tag{}, fmt.Errorf("failed to get related tags: %w", err)
	}

	return tags, nil
}

// getTypeBasedTags 获取基于内容类型的推荐标签
func (s *TagService) getTypeBasedTags(contentType string, limit int) ([]models.Tag, error) {
	var tags []models.Tag
	
	// 根据内容类型获取常用标签
	query := s.db.Model(&models.Tag{}).
		Joins("JOIN content_tags ON tags.id = content_tags.tag_id").
		Where("content_tags.content_type = ? AND tags.status = ?", contentType, models.TagStatusActive).
		Group("tags.id").
		Order("COUNT(*) DESC").
		Limit(limit)

	if err := query.Find(&tags).Error; err != nil {
		return []models.Tag{}, fmt.Errorf("failed to get type-based tags: %w", err)
	}

	return tags, nil
}

// =============== 标签统计和趋势 ===============

// GetTagStats 获取标签统计
func (s *TagService) GetTagStats() (*models.TagStatsResponse, error) {
	stats := &models.TagStatsResponse{}

	// 总标签数
	s.db.Model(&models.Tag{}).Where("status = ?", models.TagStatusActive).Count(&stats.TotalTags)

	// 分类统计
	s.db.Model(&models.Tag{}).Where("status = ? AND type = ?", models.TagStatusActive, models.TagTypeUser).Count(&stats.UserTags)
	s.db.Model(&models.Tag{}).Where("status = ? AND type = ?", models.TagStatusActive, models.TagTypeSystem).Count(&stats.SystemTags)
	s.db.Model(&models.Tag{}).Where("status = ? AND type = ?", models.TagStatusActive, models.TagTypeAI).Count(&stats.AITags)

	// 总使用次数
	var result struct {
		TotalUsage int64 `json:"total_usage"`
	}
	s.db.Model(&models.Tag{}).Select("SUM(usage_count) as total_usage").
		Where("status = ?", models.TagStatusActive).Scan(&result)
	stats.TotalUsage = result.TotalUsage

	// 平均使用次数
	if stats.TotalTags > 0 {
		stats.AvgUsagePerTag = float64(stats.TotalUsage) / float64(stats.TotalTags)
	}

	// 热门标签
	trendingTags, _ := s.GetTrendingTags(10)
	stats.TrendingTags = trendingTags

	// 流行标签
	popularTags, _ := s.GetPopularTags(10)
	stats.PopularTags = popularTags

	return stats, nil
}

// UpdateTrendingScores 更新趋势分数
func (s *TagService) UpdateTrendingScores() error {
	// 计算过去7天的标签使用趋势
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	
	query := `
		UPDATE tags SET trending_score = (
			SELECT COALESCE(
				(COUNT(*) * 1.0) / (EXTRACT(EPOCH FROM (NOW() - ?)) / 86400.0), 
				0
			)
			FROM content_tags 
			WHERE content_tags.tag_id = tags.id 
			AND content_tags.created_at >= ?
		)
		WHERE status = ?
	`

	if err := s.db.Exec(query, sevenDaysAgo, sevenDaysAgo, models.TagStatusActive).Error; err != nil {
		return fmt.Errorf("failed to update trending scores: %w", err)
	}

	log.Printf("Updated trending scores for all active tags")
	return nil
}

// =============== 辅助方法 ===============

// validateContentExists 验证内容是否存在
func (s *TagService) validateContentExists(contentType, contentID string) error {
	var count int64
	
	switch contentType {
	case "letter":
		s.db.Model(&models.Letter{}).Where("id = ?", contentID).Count(&count)
	case "museum":
		s.db.Model(&models.MuseumItem{}).Where("id = ?", contentID).Count(&count)
	case "comment":
		s.db.Model(&models.Comment{}).Where("id = ?", contentID).Count(&count)
	case "profile":
		s.db.Model(&models.User{}).Where("id = ?", contentID).Count(&count)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	if count == 0 {
		return fmt.Errorf("content not found: %s/%s", contentType, contentID)
	}

	return nil
}

// generateRandomColor 生成随机颜色
func (s *TagService) generateRandomColor() string {
	colors := []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7",
		"#DDA0DD", "#98D8C8", "#F7DC6F", "#BB8FCE", "#85C1E9",
		"#F8C471", "#82E0AA", "#F1948A", "#85C1E9", "#D2B4DE",
	}
	
	return colors[time.Now().UnixNano()%int64(len(colors))]
}

// =============== 标签分类管理 ===============

// CreateTagCategory 创建标签分类
func (s *TagService) CreateTagCategory(req *models.TagCategoryRequest) (*models.TagCategory, error) {
	// 检查分类名是否已存在
	var existingCategory models.TagCategory
	if err := s.db.Where("name = ?", req.Name).First(&existingCategory).Error; err == nil {
		return nil, fmt.Errorf("category with name '%s' already exists", req.Name)
	}

	category := &models.TagCategory{
		ID:          uuid.New().String(),
		Name:        strings.ToLower(strings.TrimSpace(req.Name)),
		DisplayName: req.DisplayName,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		IsActive:    true,
	}

	if category.DisplayName == "" {
		category.DisplayName = req.Name
	}

	if category.Color == "" {
		category.Color = s.generateRandomColor()
	}

	if err := s.db.Create(category).Error; err != nil {
		return nil, fmt.Errorf("failed to create tag category: %w", err)
	}

	return category, nil
}

// GetTagCategories 获取标签分类列表
func (s *TagService) GetTagCategories() ([]models.TagCategory, error) {
	var categories []models.TagCategory
	if err := s.db.Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get tag categories: %w", err)
	}

	// 计算每个分类的标签数量
	for i := range categories {
		var tagCount int64
		s.db.Model(&models.Tag{}).Where("category_id = ? AND status = ?", 
			categories[i].ID, models.TagStatusActive).Count(&tagCount)
		categories[i].TagCount = tagCount
	}

	return categories, nil
}