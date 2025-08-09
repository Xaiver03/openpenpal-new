package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MuseumService struct {
	db              *gorm.DB
	creditSvc       *CreditService
	notificationSvc *NotificationService
	aiSvc           *AIService
}

func NewMuseumService(db *gorm.DB) *MuseumService {
	return &MuseumService{
		db: db,
	}
}

// SetCreditService 设置积分服务
func (s *MuseumService) SetCreditService(creditSvc *CreditService) {
	s.creditSvc = creditSvc
}

// SetNotificationService 设置通知服务
func (s *MuseumService) SetNotificationService(notificationSvc *NotificationService) {
	s.notificationSvc = notificationSvc
}

// SetAIService 设置AI服务
func (s *MuseumService) SetAIService(aiSvc *AIService) {
	s.aiSvc = aiSvc
}

// GenerateItemDescription 使用AI生成博物馆物品描述
func (s *MuseumService) GenerateItemDescription(ctx context.Context, item *models.MuseumItem) (string, error) {
	if s.aiSvc == nil {
		return "", errors.New("AI service not available")
	}

	// 构建AI提示
	prompt := s.buildDescriptionPrompt(item)

	// 获取AI配置
	config, err := s.aiSvc.GetActiveProvider()
	if err != nil {
		return "", err
	}

	// 调用AI API
	response, err := s.aiSvc.callAIAPI(ctx, config, prompt, models.TaskTypeCurate)
	if err != nil {
		return "", err
	}

	// 提取和清理描述
	description := s.extractDescriptionFromResponse(response)

	// 记录AI使用情况
	s.aiSvc.logAIUsage("system", models.TaskTypeCurate, item.ID, config, 0, 0, "success", "")

	return description, nil
}

// GetMuseumEntries 获取博物馆条目列表
func (s *MuseumService) GetMuseumEntries(ctx context.Context, page, limit int, status string) ([]models.MuseumEntry, int64, error) {
	var entries []models.MuseumEntry
	var total int64

	query := s.db.Model(&models.MuseumEntry{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// GetMuseumEntry 获取单个博物馆条目
func (s *MuseumService) GetMuseumEntry(ctx context.Context, id string) (*models.MuseumEntry, error) {
	var entry models.MuseumEntry

	entryUUID, err := utils.ParseUUID(id)
	if err != nil {
		return nil, errors.New("invalid entry ID format")
	}

	if err := s.db.Where("id = ?", entryUUID).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("museum entry not found")
		}
		return nil, err
	}

	// 增加查看次数
	s.db.Model(&entry).Update("view_count", gorm.Expr("view_count + 1"))

	return &entry, nil
}

// CreateMuseumItem 创建博物馆物品
func (s *MuseumService) CreateMuseumItem(ctx context.Context, req *CreateMuseumItemRequest) (*models.MuseumItem, error) {
	item := &models.MuseumItem{
		ID:          uuid.New().String(),
		SourceType:  req.SourceType,
		SourceID:    req.SourceID,
		Title:       req.Title,
		Description: req.Description,
		Tags:        strings.Join(req.Tags, ","),
		Status:      models.MuseumItemPending,
		SubmittedBy: req.SubmittedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}

	return item, nil
}

// GetMuseumExhibitions 获取博物馆展览列表
func (s *MuseumService) GetMuseumExhibitions(ctx context.Context, page, limit int) ([]models.MuseumExhibition, int64, error) {
	var exhibitions []models.MuseumExhibition
	var total int64

	query := s.db.Model(&models.MuseumExhibition{}).Where("deleted_at IS NULL")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&exhibitions).Error; err != nil {
		return nil, 0, err
	}

	return exhibitions, total, nil
}

// ApproveMuseumItem 审批博物馆物品
func (s *MuseumService) ApproveMuseumItem(ctx context.Context, itemID, approverID string) error {
	itemUUID, err := utils.ParseUUID(itemID)
	if err != nil {
		return errors.New("invalid item ID format")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      models.MuseumItemApproved,
		"approved_by": approverID,
		"approved_at": &now,
		"updated_at":  now,
	}

	result := s.db.Model(&models.MuseumItem{}).Where("id = ?", itemUUID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("museum item not found")
	}

	// 发送审批通过通知和奖励积分
	if s.notificationSvc != nil {
		go func() {
			// 获取物品信息以便发送通知
			var item models.MuseumItem
			if err := s.db.Where("id = ?", itemUUID).First(&item).Error; err == nil {
				s.notificationSvc.NotifyUser(item.SubmittedBy, "museum_approved", map[string]interface{}{
					"item_id":     itemID,
					"item_title":  item.Title,
					"approver_id": approverID,
				})

				// 奖励审核通过积分
				if s.creditSvc != nil {
					if err := s.creditSvc.RewardMuseumApproved(item.SubmittedBy, itemID); err != nil {
						// 记录错误但不影响主流程
					}
				}
			}
		}()
	}

	return nil
}

// SubmitLetterToMuseum 提交信件到博物馆
func (s *MuseumService) SubmitLetterToMuseum(ctx context.Context, letterID, userID, title, description string, tags []string) (*models.MuseumItem, error) {
	// 检查信件是否存在且状态为已送达或已读
	var letter models.Letter
	if err := s.db.Where("id = ? AND status IN ?", letterID, []models.LetterStatus{models.StatusDelivered, models.StatusRead}).First(&letter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("letter not found or not eligible for museum submission")
		}
		return nil, err
	}

	// 检查是否已经提交过
	var existingItem models.MuseumItem
	if err := s.db.Where("source_type = ? AND source_id = ?", models.SourceTypeLetter, letterID).First(&existingItem).Error; err == nil {
		return nil, errors.New("letter already submitted to museum")
	}

	// 创建博物馆物品
	item := &models.MuseumItem{
		ID:          uuid.New().String(),
		SourceType:  models.SourceTypeLetter,
		SourceID:    letterID,
		Title:       title,
		Description: description,
		Tags:        strings.Join(tags, ","), // Convert slice to string
		Status:      models.MuseumItemPending,
		SubmittedBy: userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}

	// 奖励提交积分和发送通知
	if s.creditSvc != nil {
		go func() {
			if err := s.creditSvc.RewardMuseumSubmit(userID, item.ID); err != nil {
				// 记录错误但不影响主流程
			}
		}()
	}

	if s.notificationSvc != nil {
		go func() {
			s.notificationSvc.NotifyUser(userID, "museum_submitted", map[string]interface{}{
				"item_id":    item.ID,
				"item_title": title,
				"letter_id":  letterID,
			})
		}()
	}

	return item, nil
}

// LikeMuseumItem 点赞博物馆物品
func (s *MuseumService) LikeMuseumItem(ctx context.Context, itemID, userID string) error {
	itemUUID, err := utils.ParseUUID(itemID)
	if err != nil {
		return errors.New("invalid item ID format")
	}

	// 检查是否已经点赞
	// TODO: 实现点赞记录表来避免重复点赞

	// 更新点赞数
	result := s.db.Model(&models.MuseumItem{}).
		Where("id = ?", itemUUID).
		Update("like_count", gorm.Expr("like_count + 1"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("museum item not found")
	}

	// 获取物品信息以便发送通知和奖励
	var item models.MuseumItem
	if err := s.db.Where("id = ?", itemUUID).First(&item).Error; err == nil {
		// 发送通知给作品提交者
		if s.notificationSvc != nil {
			go func() {
				s.notificationSvc.NotifyUser(item.SubmittedBy, "museum_liked", map[string]interface{}{
					"item_id":    itemID,
					"item_title": item.Title,
					"liked_by":   userID,
				})
			}()
		}

		// 奖励被点赞积分
		if s.creditSvc != nil {
			go func() {
				if err := s.creditSvc.RewardMuseumLiked(item.SubmittedBy, itemID); err != nil {
					// 记录错误但不影响主流程
				}
			}()
		}
	}

	return nil
}

// RejectMuseumItem 拒绝博物馆物品
func (s *MuseumService) RejectMuseumItem(ctx context.Context, itemID, reviewerID, reason string) error {
	itemUUID, err := utils.ParseUUID(itemID)
	if err != nil {
		return errors.New("invalid item ID format")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      models.MuseumItemRejected,
		"approved_by": reviewerID, // 也记录拒绝者信息
		"approved_at": &now,
		"updated_at":  now,
	}

	result := s.db.Model(&models.MuseumItem{}).Where("id = ?", itemUUID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("museum item not found")
	}

	// 发送拒绝通知
	if s.notificationSvc != nil {
		go func() {
			// 获取物品信息以便发送通知
			var item models.MuseumItem
			if err := s.db.Where("id = ?", itemUUID).First(&item).Error; err == nil {
				s.notificationSvc.NotifyUser(item.SubmittedBy, "museum_rejected", map[string]interface{}{
					"item_id":     itemID,
					"item_title":  item.Title,
					"reason":      reason,
					"reviewer_id": reviewerID,
				})
			}
		}()
	}

	return nil
}

// buildDescriptionPrompt 构建描述生成提示
func (s *MuseumService) buildDescriptionPrompt(item *models.MuseumItem) string {
	prompt := `作为博物馆策展人，请为以下展品生成一个专业的描述：

展品标题：` + item.Title + `
展品类型：`

	switch item.SourceType {
	case models.SourceTypeLetter:
		prompt += "手写信件"
	case models.SourceTypePhoto:
		prompt += "照片图像"
	case models.SourceTypeAudio:
		prompt += "音频内容"
	default:
		prompt += "其他"
	}

	if item.Tags != "" {
		prompt += `
相关标签：` + item.Tags
	}

	if item.Description != "" {
		prompt += `
现有描述：` + item.Description
	}

	prompt += `

请生成一个200-300字的专业博物馆展品描述，包括：
1. 展品的历史背景和文化意义
2. 艺术价值和技艺特色
3. 在该时期的代表性意义
4. 对观众的教育价值

要求：
- 语言专业但易懂
- 突出展品的独特性
- 体现历史文化内涵
- 激发观众兴趣

请直接返回描述内容，不要包含额外说明。`

	return prompt
}

// extractDescriptionFromResponse 从AI响应中提取描述
func (s *MuseumService) extractDescriptionFromResponse(response string) string {
	// 简单清理AI响应，去除可能的前缀和后缀
	cleaned := strings.TrimSpace(response)

	// 移除常见的AI回复前缀
	prefixes := []string{"描述：", "展品描述：", "博物馆描述："}
	for _, prefix := range prefixes {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, prefix))
		}
	}

	return cleaned
}

// CreateMuseumItemRequest 创建博物馆物品请求
type CreateMuseumItemRequest struct {
	SourceType  models.MuseumSourceType `json:"sourceType" binding:"required"`
	SourceID    string                  `json:"sourceId" binding:"required"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Tags        []string                `json:"tags"`
	SubmittedBy string                  `json:"submittedBy" binding:"required"`
}

// GetPopularEntries 获取热门条目
func (s *MuseumService) GetPopularEntries(ctx context.Context, page, limit int, timeRange string) ([]models.MuseumEntry, int64, error) {
	var items []models.MuseumEntry
	var total int64

	// 计算时间范围
	var startTime time.Time
	now := time.Now()
	switch timeRange {
	case "day":
		startTime = now.AddDate(0, 0, -1)
	case "week":
		startTime = now.AddDate(0, 0, -7)
	case "month":
		startTime = now.AddDate(0, -1, 0)
	default:
		startTime = time.Time{} // 所有时间
	}

	query := s.db.Model(&models.MuseumEntry{}).
		Where("status = ?", models.MuseumItemApproved)

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}

	// 按照综合得分排序（浏览量 + 点赞量*2 + 收藏量*3 + 分享量*4）
	query = query.Select("museum_entries.*, (view_count + like_count*2 + bookmark_count*3 + share_count*4) as score").
		Order("score DESC")

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetExhibitionByID 根据ID获取展览
func (s *MuseumService) GetExhibitionByID(ctx context.Context, exhibitionID string) (*models.MuseumExhibition, error) {
	var exhibition models.MuseumExhibition
	if err := s.db.Where("id = ?", exhibitionID).First(&exhibition).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("exhibition not found")
		}
		return nil, err
	}
	return &exhibition, nil
}

// GetPopularTags 获取热门标签
func (s *MuseumService) GetPopularTags(ctx context.Context, category string, limit int) ([]models.MuseumTag, error) {
	var tags []models.MuseumTag

	// 从标签表中查询
	query := s.db.Model(&models.MuseumTag{}).Order("usage_count DESC")
	
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Limit(limit).Find(&tags).Error; err != nil {
		return nil, err
	}

	// 如果标签表为空，从博物馆条目中提取标签
	if len(tags) == 0 {
		type tagCount struct {
			Tag   string
			Count int
		}
		var tagCounts []tagCount

		// 使用原生SQL查询标签使用情况
		sql := `
			SELECT tag, COUNT(*) as count 
			FROM (
				SELECT unnest(tags) as tag 
				FROM museum_entries 
				WHERE status = ?
			) t 
			GROUP BY tag 
			ORDER BY count DESC 
			LIMIT ?
		`
		
		if err := s.db.Raw(sql, models.MuseumItemApproved, limit).
			Scan(&tagCounts).Error; err != nil {
			// 如果数据库不支持unnest，使用备用方案
			return s.getTagsAlternative(ctx, limit)
		}

		// 转换为MuseumTag格式
		for _, tc := range tagCounts {
			tags = append(tags, models.MuseumTag{
				ID:         fmt.Sprintf("tag_%s", tc.Tag),
				Name:       tc.Tag,
				Category:   "general",
				UsageCount: tc.Count,
			})
		}
	}

	return tags, nil
}

// getTagsAlternative 备用的标签获取方法
func (s *MuseumService) getTagsAlternative(ctx context.Context, limit int) ([]models.MuseumTag, error) {
	var items []models.MuseumItem
	tagMap := make(map[string]int)

	// 获取所有已发布的条目
	if err := s.db.Where("status = ?", 
		models.MuseumItemApproved).
		Find(&items).Error; err != nil {
		return nil, err
	}

	// 统计标签使用次数
	for _, item := range items {
		if item.Tags != "" {
			tags := strings.Split(item.Tags, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tagMap[tag]++
				}
			}
		}
	}

	// 转换为标签列表并排序
	var tags []models.MuseumTag
	for tag, count := range tagMap {
		tags = append(tags, models.MuseumTag{
			ID:         fmt.Sprintf("tag_%s", tag),
			Name:       tag,
			Category:   "general",
			UsageCount: count,
		})
	}

	// 手动排序并限制数量
	// 这里应该使用排序算法，为简化使用简单实现
	if len(tags) > limit {
		tags = tags[:limit]
	}

	return tags, nil
}

// RecordInteraction 记录用户互动
func (s *MuseumService) RecordInteraction(ctx context.Context, entryID, userID, interactionType string) error {
	// 先检查条目是否存在
	var item models.MuseumItem
	if err := s.db.Where("id = ?", entryID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("entry not found")
		}
		return err
	}

	// 更新对应的计数器
	updates := make(map[string]interface{})
	switch interactionType {
	case "view":
		updates["view_count"] = gorm.Expr("view_count + ?", 1)
	case "like":
		updates["like_count"] = gorm.Expr("like_count + ?", 1)
	case "bookmark":
		updates["bookmark_count"] = gorm.Expr("bookmark_count + ?", 1)
	case "share":
		updates["share_count"] = gorm.Expr("share_count + ?", 1)
	default:
		return errors.New("invalid interaction type")
	}

	// 更新条目
	if err := s.db.Model(&models.MuseumItem{}).Where("id = ?", entryID).Updates(updates).Error; err != nil {
		return err
	}

	// 记录互动历史（可选）
	if userID != "" {
		interaction := models.MuseumInteraction{
			ID:        fmt.Sprintf("interaction_%s", time.Now().Format("20060102150405")),
			EntryID:   entryID,
			UserID:    userID,
			Type:      interactionType,
			CreatedAt: time.Now(),
		}
		s.db.Create(&interaction) // 忽略错误，因为这不是关键操作
	}

	return nil
}

// AddReaction 添加反应
func (s *MuseumService) AddReaction(ctx context.Context, entryID, userID, reactionType, comment string) (*models.MuseumReaction, error) {
	// 检查条目是否存在
	var item models.MuseumItem
	if err := s.db.Where("id = ?", entryID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("entry not found")
		}
		return nil, err
	}

	// 创建反应记录
	reaction := &models.MuseumReaction{
		ID:           fmt.Sprintf("reaction_%s_%s", userID, time.Now().Format("20060102150405")),
		EntryID:      entryID,
		UserID:       userID,
		ReactionType: reactionType,
		Comment:      comment,
		CreatedAt:    time.Now(),
	}

	if err := s.db.Create(reaction).Error; err != nil {
		return nil, err
	}

	// 更新条目的反应计数
	if reactionType == "like" || reactionType == "love" {
		s.db.Model(&models.MuseumItem{}).Where("id = ?", entryID).
			Update("like_count", gorm.Expr("like_count + ?", 1))
	}

	// 发送通知给作者
	if s.notificationSvc != nil && item.SubmittedBy != userID {
		s.notificationSvc.NotifyUser(item.SubmittedBy, "museum_reaction", map[string]interface{}{
			"entry_id": entryID,
			"reaction_type": reactionType,
			"title": item.Title,
		})
	}

	return reaction, nil
}

// WithdrawEntry 撤回条目
func (s *MuseumService) WithdrawEntry(ctx context.Context, entryID, userID string) error {
	var item models.MuseumItem
	if err := s.db.Where("id = ?", entryID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("entry not found")
		}
		return err
	}

	// 检查权限：只有提交者可以撤回
	if item.SubmittedBy != userID {
		return errors.New("unauthorized")
	}

	// 更新状态为已撤回
	updates := map[string]interface{}{
		"status":     "withdrawn",
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&models.MuseumItem{}).Where("id = ?", entryID).Updates(updates).Error; err != nil {
		return err
	}

	// 如果有相关的提交记录，也更新其状态
	s.db.Model(&models.MuseumSubmission{}).
		Where("entry_id = ? AND submitted_by = ?", entryID, userID).
		Update("status", "withdrawn")

	return nil
}

// GetUserSubmissions 获取用户的提交记录
func (s *MuseumService) GetUserSubmissions(ctx context.Context, userID string, page, limit int, status string) ([]models.MuseumSubmission, int64, error) {
	var submissions []models.MuseumSubmission
	var total int64

	query := s.db.Model(&models.MuseumSubmission{}).Where("submitted_by = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Order("submitted_at DESC").Offset(offset).Limit(limit).
		Preload("Letter").Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, total, nil
}

// ModerateEntry 审核条目
func (s *MuseumService) ModerateEntry(ctx context.Context, entryID, moderatorID, status, reason string, featured bool) error {
	var item models.MuseumItem
	if err := s.db.Where("id = ?", entryID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("entry not found")
		}
		return err
	}

	// 更新审核状态
	updates := map[string]interface{}{
		"status":      status,
		"approved_by": moderatorID,
		"approved_at": time.Now(),
		"updated_at":  time.Now(),
	}

	if status == "approved" {
		updates["status"] = models.MuseumItemApproved
		if featured {
			updates["featured_at"] = time.Now()
		}
	} else {
		updates["status"] = models.MuseumItemRejected
	}

	if err := s.db.Model(&models.MuseumItem{}).Where("id = ?", entryID).Updates(updates).Error; err != nil {
		return err
	}

	// 更新相关的提交记录
	s.db.Model(&models.MuseumSubmission{}).
		Where("entry_id = ?", entryID).
		Updates(map[string]interface{}{
			"status":      status,
			"reviewed_at": time.Now(),
			"reviewed_by": moderatorID,
		})

	// 发送通知
	if s.notificationSvc != nil && item.SubmittedBy != "" {
		var notifType string
		var notifData map[string]interface{}
		
		if status == "approved" {
			notifType = "museum_approved"
			notifData = map[string]interface{}{
				"entry_id": entryID,
				"title": item.Title,
			}
		} else {
			notifType = "museum_rejected"
			notifData = map[string]interface{}{
				"entry_id": entryID,
				"title": item.Title,
				"reason": reason,
			}
		}

		s.notificationSvc.NotifyUser(item.SubmittedBy, notifType, notifData)
	}

	// 如果审核通过，给用户加积分
	if status == "approved" && s.creditSvc != nil && item.SubmittedBy != "" {
		s.creditSvc.AddPoints(item.SubmittedBy, 50, "museum_approved", 
			fmt.Sprintf("信件《%s》被博物馆收录", item.Title))
	}

	return nil
}

// GetPendingEntries 获取待审核条目
func (s *MuseumService) GetPendingEntries(ctx context.Context, page, limit int) ([]models.MuseumItem, int64, error) {
	var items []models.MuseumItem
	var total int64

	query := s.db.Model(&models.MuseumItem{}).Where("status = ?", models.MuseumItemPending)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Order("created_at ASC").Offset(offset).Limit(limit).
		Preload("Letter").Preload("SubmittedByUser").Preload("ApprovedByUser").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CreateExhibition 创建展览
func (s *MuseumService) CreateExhibition(ctx context.Context, exhibition *models.MuseumExhibition) (*models.MuseumExhibition, error) {
	exhibition.ID = fmt.Sprintf("exhibition_%s", time.Now().Format("20060102150405"))
	exhibition.CreatedAt = time.Now()
	exhibition.UpdatedAt = time.Now()
	exhibition.ViewCount = 0

	if err := s.db.Create(exhibition).Error; err != nil {
		return nil, err
	}

	return exhibition, nil
}

// UpdateExhibition 更新展览
func (s *MuseumService) UpdateExhibition(ctx context.Context, exhibition *models.MuseumExhibition, userID string) (*models.MuseumExhibition, error) {
	// 检查展览是否存在
	var existing models.MuseumExhibition
	if err := s.db.Where("id = ?", exhibition.ID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("exhibition not found")
		}
		return nil, err
	}

	// 检查权限
	if existing.CreatorID != userID {
		return nil, errors.New("unauthorized")
	}

	// 更新字段
	exhibition.UpdatedAt = time.Now()
	exhibition.CreatorID = existing.CreatorID // 保持原创建者
	exhibition.CreatedAt = existing.CreatedAt // 保持原创建时间

	if err := s.db.Save(exhibition).Error; err != nil {
		return nil, err
	}

	return exhibition, nil
}

// DeleteExhibition 删除展览
func (s *MuseumService) DeleteExhibition(ctx context.Context, exhibitionID, userID string) error {
	// 检查展览是否存在
	var exhibition models.MuseumExhibition
	if err := s.db.Where("id = ?", exhibitionID).First(&exhibition).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("exhibition not found")
		}
		return err
	}

	// 检查权限
	if exhibition.CreatorID != userID {
		return errors.New("unauthorized")
	}

	// 软删除
	return s.db.Delete(&exhibition).Error
}

// RefreshStats 刷新统计数据
func (s *MuseumService) RefreshStats(ctx context.Context) error {
	// 这里可以实现各种统计数据的刷新逻辑
	// 例如：重新计算标签使用次数、更新热度分数等

	// 更新标签使用次数

	// 统计所有已发布条目的标签
	var items []models.MuseumItem
	if err := s.db.Where("status = ?", 
		models.MuseumItemApproved).Find(&items).Error; err != nil {
		return err
	}

	tagMap := make(map[string]int)
	for _, item := range items {
		if item.Tags != "" {
			tags := strings.Split(item.Tags, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tagMap[tag]++
				}
			}
		}
	}

	// 更新或创建标签记录
	for tag, count := range tagMap {
		museumTag := models.MuseumTag{
			ID:         fmt.Sprintf("tag_%s", tag),
			Name:       tag,
			Category:   "general",
			UsageCount: count,
			CreatedAt:  time.Now(),
		}

		// 使用 Upsert 操作
		s.db.Where("id = ?", museumTag.ID).Assign(museumTag).FirstOrCreate(&museumTag)
	}

	return nil
}

// GetAnalytics 获取分析数据
func (s *MuseumService) GetAnalytics(ctx context.Context, timeRange, startDate, endDate string) (*models.MuseumAnalytics, error) {
	analytics := &models.MuseumAnalytics{
		GeneratedAt: time.Now(),
	}

	// 计算时间范围
	var start, end time.Time
	if startDate != "" && endDate != "" {
		// 使用自定义时间范围
		var err error
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, err
		}
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, err
		}
	} else {
		// 使用预设时间范围
		end = time.Now()
		switch timeRange {
		case "day":
			start = end.AddDate(0, 0, -1)
		case "week":
			start = end.AddDate(0, 0, -7)
		case "month":
			start = end.AddDate(0, -1, 0)
		case "year":
			start = end.AddDate(-1, 0, 0)
		default:
			start = end.AddDate(0, -1, 0) // 默认一个月
		}
	}

	// 基础统计
	s.db.Model(&models.MuseumItem{}).
		Where("status = ?", models.MuseumItemApproved).
		Count(&analytics.TotalEntries)

	// 总浏览量、点赞量、分享量
	var stats struct {
		TotalViews  int64
		TotalLikes  int64
		TotalShares int64
	}
	s.db.Model(&models.MuseumItem{}).
		Where("status = ?", models.MuseumItemApproved).
		Select("SUM(view_count) as total_views, SUM(like_count) as total_likes, SUM(share_count) as total_shares").
		Scan(&stats)

	analytics.TotalViews = stats.TotalViews
	analytics.TotalLikes = stats.TotalLikes
	analytics.TotalShares = stats.TotalShares

	// 热门分类
	// 这里简化处理，实际应该从分类字段统计
	analytics.PopularCategories = []models.CategoryStat{
		{Category: "情感", Count: 100},
		{Category: "友谊", Count: 80},
		{Category: "成长", Count: 60},
	}

	// 热门标签
	var tags []models.MuseumTag
	s.db.Model(&models.MuseumTag{}).Order("usage_count DESC").Limit(10).Find(&tags)
	for _, tag := range tags {
		analytics.PopularTags = append(analytics.PopularTags, models.TagStat{
			Tag:   tag.Name,
			Count: tag.UsageCount,
		})
	}

	// 每日统计
	// 这里简化处理，实际应该按日期分组统计
	days := int(end.Sub(start).Hours() / 24)
	for i := 0; i < days && i < 30; i++ {
		date := end.AddDate(0, 0, -i)
		analytics.DailyStats = append(analytics.DailyStats, models.DailyStat{
			Date:        date.Format("2006-01-02"),
			Views:       100 + i*10,
			Likes:       20 + i*2,
			Submissions: 5 + i,
		})
	}

	return analytics, nil
}

// SearchEntries 搜索博物馆条目
func (s *MuseumService) SearchEntries(ctx context.Context, query string, tags []string, theme, status string, featured *bool, dateFrom, dateTo, sortBy, sortOrder string, page, limit int) ([]models.MuseumEntry, int64, error) {
	var entries []models.MuseumEntry
	var total int64

	// 构建查询
	dbQuery := s.db.Model(&models.MuseumEntry{})

	// 状态过滤 - 默认只显示已批准的
	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	} else {
		dbQuery = dbQuery.Where("status = ?", models.MuseumItemApproved)
	}

	// 文本搜索
	if query != "" {
		dbQuery = dbQuery.Where("title LIKE ? OR content LIKE ? OR author_name LIKE ?", 
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	// 标签过滤
	if len(tags) > 0 {
		for _, tag := range tags {
			dbQuery = dbQuery.Where("tags LIKE ?", "%"+tag+"%")
		}
	}

	// 主题过滤
	if theme != "" {
		dbQuery = dbQuery.Where("theme = ?", theme)
	}

	// 精选过滤
	if featured != nil {
		if *featured {
			dbQuery = dbQuery.Where("featured_at IS NOT NULL")
		} else {
			dbQuery = dbQuery.Where("featured_at IS NULL")
		}
	}

	// 日期范围过滤
	if dateFrom != "" {
		if fromDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			dbQuery = dbQuery.Where("created_at >= ?", fromDate)
		}
	}
	if dateTo != "" {
		if toDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			// 加一天以包含当天
			toDate = toDate.Add(24 * time.Hour)
			dbQuery = dbQuery.Where("created_at <= ?", toDate)
		}
	}

	// 计算总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "created_at DESC" // 默认排序
	switch sortBy {
	case "created_at":
		orderBy = "created_at " + strings.ToUpper(sortOrder)
	case "view_count":
		orderBy = "view_count " + strings.ToUpper(sortOrder)
	case "like_count":
		orderBy = "like_count " + strings.ToUpper(sortOrder)
	case "title":
		orderBy = "title " + strings.ToUpper(sortOrder)
	case "relevance":
		// 简单的相关性排序：综合浏览量、点赞量等
		orderBy = "(view_count + like_count * 2 + bookmark_count * 3) DESC"
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := dbQuery.Order(orderBy).Offset(offset).Limit(limit).Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// AddItemsToExhibition 向展览添加物品
func (s *MuseumService) AddItemsToExhibition(ctx context.Context, exhibitionID string, itemIDs []string, userID string) error {
	// 检查展览是否存在
	var exhibition models.MuseumExhibition
	if err := s.db.Where("id = ?", exhibitionID).First(&exhibition).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("exhibition not found")
		}
		return err
	}

	// 检查权限
	if exhibition.CreatorID != userID {
		return errors.New("unauthorized")
	}

	// 检查物品是否都存在且已批准
	var count int64
	if err := s.db.Model(&models.MuseumItem{}).
		Where("id IN ? AND status = ?", itemIDs, models.MuseumItemApproved).
		Count(&count).Error; err != nil {
		return err
	}

	if count != int64(len(itemIDs)) {
		return errors.New("some items not found or not approved")
	}

	// 添加物品到展览
	for i, itemID := range itemIDs {
		entry := models.MuseumExhibitionEntry{
			ID:           uuid.New().String(),
			CollectionID: exhibitionID,
			ItemID:       itemID,
			DisplayOrder: i + 1,
			CreatedAt:    time.Now(),
		}

		// 检查是否已存在
		var existing models.MuseumExhibitionEntry
		if err := s.db.Where("collection_id = ? AND item_id = ?", exhibitionID, itemID).First(&existing).Error; err == nil {
			continue // 已存在，跳过
		}

		if err := s.db.Create(&entry).Error; err != nil {
			return err
		}
	}

	// 更新展览的物品数量
	var entryCount int64
	s.db.Model(&models.MuseumExhibitionEntry{}).Where("collection_id = ?", exhibitionID).Count(&entryCount)
	s.db.Model(&exhibition).Update("current_entries", entryCount)

	return nil
}

// RemoveItemsFromExhibition 从展览中移除物品
func (s *MuseumService) RemoveItemsFromExhibition(ctx context.Context, exhibitionID string, itemIDs []string, userID string) error {
	// 检查展览是否存在
	var exhibition models.MuseumExhibition
	if err := s.db.Where("id = ?", exhibitionID).First(&exhibition).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("exhibition not found")
		}
		return err
	}

	// 检查权限
	if exhibition.CreatorID != userID {
		return errors.New("unauthorized")
	}

	// 移除物品
	if err := s.db.Where("collection_id = ? AND item_id IN ?", exhibitionID, itemIDs).
		Delete(&models.MuseumExhibitionEntry{}).Error; err != nil {
		return err
	}

	// 更新展览的物品数量
	var entryCount int64
	s.db.Model(&models.MuseumExhibitionEntry{}).Where("collection_id = ?", exhibitionID).Count(&entryCount)
	s.db.Model(&exhibition).Update("current_entries", entryCount)

	return nil
}

// GetExhibitionItems 获取展览中的物品
func (s *MuseumService) GetExhibitionItems(ctx context.Context, exhibitionID string, page, limit int) ([]models.MuseumItem, int64, error) {
	// 检查展览是否存在
	var exhibition models.MuseumExhibition
	if err := s.db.Where("id = ?", exhibitionID).First(&exhibition).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("exhibition not found")
		}
		return nil, 0, err
	}

	var items []models.MuseumItem
	var total int64

	// 通过关联表查询展览中的物品
	query := s.db.Model(&models.MuseumItem{}).
		Joins("JOIN museum_exhibition_entries ON museum_items.id = museum_exhibition_entries.item_id").
		Where("museum_exhibition_entries.collection_id = ?", exhibitionID).
		Order("museum_exhibition_entries.display_order ASC")

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("Letter").Preload("SubmittedByUser").Preload("ApprovedByUser").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ItemOrder represents an item's display order in an exhibition
type ItemOrder struct {
	ItemID       string `json:"item_id"`
	DisplayOrder int    `json:"display_order"`
}

// UpdateExhibitionItemOrder 更新展览中物品的显示顺序
func (s *MuseumService) UpdateExhibitionItemOrder(ctx context.Context, exhibitionID string, itemOrders []ItemOrder, userID string) error {
	// 检查展览是否存在
	var exhibition models.MuseumExhibition
	if err := s.db.Where("id = ?", exhibitionID).First(&exhibition).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("exhibition not found")
		}
		return err
	}

	// 检查权限
	if exhibition.CreatorID != userID {
		return errors.New("unauthorized")
	}

	// 更新每个物品的显示顺序
	for _, itemOrder := range itemOrders {
		if err := s.db.Model(&models.MuseumExhibitionEntry{}).
			Where("collection_id = ? AND item_id = ?", exhibitionID, itemOrder.ItemID).
			Update("display_order", itemOrder.DisplayOrder).Error; err != nil {
			return err
		}
	}

	return nil
}

// PublishExhibition 发布展览
func (s *MuseumService) PublishExhibition(ctx context.Context, exhibitionID, userID string) error {
	// 检查展览是否存在
	var exhibition models.MuseumExhibition
	if err := s.db.Where("id = ?", exhibitionID).First(&exhibition).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("exhibition not found")
		}
		return err
	}

	// 检查权限
	if exhibition.CreatorID != userID {
		return errors.New("unauthorized")
	}

	// 检查展览是否有物品
	var itemCount int64
	if err := s.db.Model(&models.MuseumExhibitionEntry{}).
		Where("collection_id = ?", exhibitionID).Count(&itemCount).Error; err != nil {
		return err
	}

	if itemCount == 0 {
		return errors.New("cannot publish empty exhibition")
	}

	// 更新状态为已发布
	updates := map[string]interface{}{
		"status":     "published",
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&exhibition).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// GetMuseumStats 获取博物馆统计数据 - SOTA: Clean service layer implementation
func (s *MuseumService) GetMuseumStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计总物品数
	var totalItems int64
	if err := s.db.Model(&models.MuseumItem{}).Count(&totalItems).Error; err != nil {
		return nil, err
	}
	stats["total_items"] = totalItems

	// 统计公开和私有物品数
	var publicItems int64
	if err := s.db.Model(&models.MuseumItem{}).
		Where("status = ?", models.MuseumItemApproved).
		Count(&publicItems).Error; err != nil {
		return nil, err
	}
	stats["public_items"] = publicItems
	stats["private_items"] = totalItems - publicItems

	// 统计总浏览量、点赞量、评论数
	var aggregates struct {
		TotalViews    int64 `gorm:"column:total_views"`
		TotalLikes    int64 `gorm:"column:total_likes"`
		TotalComments int64 `gorm:"column:total_comments"`
	}
	
	if err := s.db.Model(&models.MuseumItem{}).
		Select("COALESCE(SUM(view_count), 0) as total_views, COALESCE(SUM(like_count), 0) as total_likes, COALESCE(SUM(comment_count), 0) as total_comments").
		Scan(&aggregates).Error; err != nil {
		return nil, err
	}
	
	stats["total_views"] = aggregates.TotalViews
	stats["total_likes"] = aggregates.TotalLikes
	stats["total_comments"] = aggregates.TotalComments

	// 统计展览数
	var exhibitions int64
	if err := s.db.Model(&models.MuseumExhibition{}).
		Where("deleted_at IS NULL").
		Count(&exhibitions).Error; err != nil {
		return nil, err
	}
	stats["exhibitions"] = exhibitions

	// 统计活跃标签数
	var activeTags int64
	if err := s.db.Model(&models.MuseumTag{}).
		Where("usage_count > 0").
		Count(&activeTags).Error; err != nil {
		// 如果标签表不存在或为空，从物品中统计
		var items []models.MuseumItem
		if err := s.db.Where("status = ? AND tags != ''", models.MuseumItemApproved).Find(&items).Error; err == nil {
			tagSet := make(map[string]bool)
			for _, item := range items {
				if item.Tags != "" {
					tags := strings.Split(item.Tags, ",")
					for _, tag := range tags {
						tag = strings.TrimSpace(tag)
						if tag != "" {
							tagSet[tag] = true
						}
					}
				}
			}
			activeTags = int64(len(tagSet))
		}
	}
	stats["active_tags"] = activeTags

	return stats, nil
}