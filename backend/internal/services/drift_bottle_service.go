package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"openpenpal-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DriftBottleService 漂流瓶服务
type DriftBottleService struct {
	db                *gorm.DB
	letterService     *LetterService
	notificationSvc   *NotificationService
	transactionHelper *TransactionHelper
}

// NewDriftBottleService 创建漂流瓶服务实例
func NewDriftBottleService(db *gorm.DB, letterService *LetterService, notificationSvc *NotificationService) *DriftBottleService {
	return &DriftBottleService{
		db:                db,
		letterService:     letterService,
		notificationSvc:   notificationSvc,
		transactionHelper: NewTransactionHelper(db),
	}
}

// DriftBottleCreateRequest 创建漂流瓶请求
type DriftBottleCreateRequest struct {
	LetterID string `json:"letter_id" binding:"required"`
	Theme    string `json:"theme"`
	Region   string `json:"region"`
	Days     int    `json:"days"` // 漂流天数，默认7天
}

// DriftBottleResponse 漂流瓶响应
type DriftBottleResponse struct {
	ID          string                     `json:"id"`
	Letter      *models.Letter             `json:"letter"`
	Status      models.DriftBottleStatus   `json:"status"`
	Theme       string                     `json:"theme"`
	Region      string                     `json:"region"`
	ExpiresAt   time.Time                  `json:"expires_at"`
	CollectedAt *time.Time                 `json:"collected_at,omitempty"`
	Collector   *models.User               `json:"collector,omitempty"`
	CreatedAt   time.Time                  `json:"created_at"`
}

// CreateDriftBottle 创建漂流瓶
func (s *DriftBottleService) CreateDriftBottle(ctx context.Context, userID string, req *DriftBottleCreateRequest) (*DriftBottleResponse, error) {
	// 验证信件归属
	letter, err := s.letterService.GetLetterByID(req.LetterID, userID)
	if err != nil {
		return nil, fmt.Errorf("letter not found: %w", err)
	}
	
	if letter.UserID != userID {
		return nil, errors.New("unauthorized: letter does not belong to user")
	}
	
	// 检查信件是否已经是漂流瓶
	var existingBottle models.DriftBottle
	if err := s.db.Where("letter_id = ?", req.LetterID).First(&existingBottle).Error; err == nil {
		return nil, errors.New("letter is already a drift bottle")
	}
	
	// 设置默认值
	if req.Days <= 0 {
		req.Days = 7
	}
	
	// 创建漂流瓶
	bottle := &models.DriftBottle{
		ID:        uuid.New().String(),
		LetterID:  req.LetterID,
		SenderID:  userID,
		Status:    models.DriftBottleStatusFloating,
		Theme:     req.Theme,
		Region:    req.Region,
		ExpiresAt: time.Now().AddDate(0, 0, req.Days),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	err = s.transactionHelper.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 创建漂流瓶
		if err := tx.Create(bottle).Error; err != nil {
			return fmt.Errorf("failed to create drift bottle: %w", err)
		}
		
		// 更新信件状态为漂流中
		if err := tx.Model(&models.Letter{}).Where("id = ?", req.LetterID).
			Update("visibility", models.VisibilityDrift).Error; err != nil {
			return fmt.Errorf("failed to update letter visibility: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// 发送通知
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyUser(userID, "drift_bottle_created", map[string]interface{}{
			"bottle_id": bottle.ID,
			"letter_title": letter.Title,
			"expires_at": bottle.ExpiresAt,
		})
	}
	
	return s.buildDriftBottleResponse(bottle, letter, nil), nil
}

// CollectDriftBottle 捞取漂流瓶
func (s *DriftBottleService) CollectDriftBottle(ctx context.Context, userID string) (*DriftBottleResponse, error) {
	var bottle models.DriftBottle
	
	// 使用事务确保原子性
	err := s.transactionHelper.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 随机获取一个未被捞取的漂流瓶（排除自己的）
		subQuery := tx.Model(&models.DriftBottle{}).
			Where("status = ?", models.DriftBottleStatusFloating).
			Where("expires_at > ?", time.Now()).
			Where("sender_id != ?", userID).
			Select("id")
		
		// 随机排序并选择一个
		if err := tx.Where("id IN (?)", subQuery).
			Order("RANDOM()").
			Limit(1).
			First(&bottle).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("no drift bottles available")
			}
			return fmt.Errorf("failed to find drift bottle: %w", err)
		}
		
		// 更新漂流瓶状态
		now := time.Now()
		bottle.Status = models.DriftBottleStatusCollected
		bottle.CollectorID = userID
		bottle.CollectedAt = &now
		bottle.UpdatedAt = now
		
		if err := tx.Save(&bottle).Error; err != nil {
			return fmt.Errorf("failed to update drift bottle: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// 加载关联数据
	if err := s.db.Preload("Letter").Preload("Sender").First(&bottle, bottle.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load bottle details: %w", err)
	}
	
	// 发送通知给发送者
	if s.notificationSvc != nil && bottle.SenderID != "" {
		s.notificationSvc.NotifyUser(bottle.SenderID, "drift_bottle_collected", map[string]interface{}{
			"bottle_id": bottle.ID,
			"letter_title": bottle.Letter.Title,
			"collected_by": userID,
		})
	}
	
	collector := &models.User{ID: userID}
	s.db.First(collector, userID)
	
	return s.buildDriftBottleResponse(&bottle, bottle.Letter, collector), nil
}

// GetMyDriftBottles 获取我的漂流瓶列表
func (s *DriftBottleService) GetMyDriftBottles(ctx context.Context, userID string, page, limit int) ([]DriftBottleResponse, int64, error) {
	var bottles []models.DriftBottle
	var total int64
	
	query := s.db.Model(&models.DriftBottle{}).Where("sender_id = ?", userID)
	
	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count drift bottles: %w", err)
	}
	
	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("Letter").
		Preload("Collector").
		Order("created_at DESC").
		Find(&bottles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query drift bottles: %w", err)
	}
	
	// 构建响应
	responses := make([]DriftBottleResponse, len(bottles))
	for i, bottle := range bottles {
		responses[i] = *s.buildDriftBottleResponse(&bottle, bottle.Letter, bottle.Collector)
	}
	
	return responses, total, nil
}

// GetCollectedBottles 获取我捞取的漂流瓶
func (s *DriftBottleService) GetCollectedBottles(ctx context.Context, userID string, page, limit int) ([]DriftBottleResponse, int64, error) {
	var bottles []models.DriftBottle
	var total int64
	
	query := s.db.Model(&models.DriftBottle{}).Where("collector_id = ?", userID)
	
	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count collected bottles: %w", err)
	}
	
	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("Letter").
		Preload("Sender").
		Order("collected_at DESC").
		Find(&bottles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query collected bottles: %w", err)
	}
	
	// 构建响应
	responses := make([]DriftBottleResponse, len(bottles))
	for i, bottle := range bottles {
		responses[i] = *s.buildDriftBottleResponse(&bottle, bottle.Letter, bottle.Sender)
	}
	
	return responses, total, nil
}

// GetFloatingBottles 获取漂流中的瓶子（用于展示）
func (s *DriftBottleService) GetFloatingBottles(ctx context.Context, region string, limit int) ([]DriftBottleResponse, error) {
	var bottles []models.DriftBottle
	
	query := s.db.Model(&models.DriftBottle{}).
		Where("status = ?", models.DriftBottleStatusFloating).
		Where("expires_at > ?", time.Now())
	
	if region != "" {
		query = query.Where("region = ?", region)
	}
	
	// 随机获取一批漂流瓶
	if err := query.Order("RANDOM()").
		Limit(limit).
		Preload("Letter").
		Find(&bottles).Error; err != nil {
		return nil, fmt.Errorf("failed to query floating bottles: %w", err)
	}
	
	// 构建响应（不包含发送者信息，保护隐私）
	responses := make([]DriftBottleResponse, len(bottles))
	for i, bottle := range bottles {
		responses[i] = *s.buildDriftBottleResponse(&bottle, bottle.Letter, nil)
		// 隐藏详细内容，只显示预览
		if responses[i].Letter != nil {
			responses[i].Letter.Content = s.getContentPreview(responses[i].Letter.Content, 50)
		}
	}
	
	return responses, nil
}

// ExpireOldBottles 过期旧的漂流瓶（定时任务）
func (s *DriftBottleService) ExpireOldBottles(ctx context.Context) error {
	result := s.db.Model(&models.DriftBottle{}).
		Where("status = ?", models.DriftBottleStatusFloating).
		Where("expires_at <= ?", time.Now()).
		Update("status", models.DriftBottleStatusExpired)
	
	if result.Error != nil {
		return fmt.Errorf("failed to expire old bottles: %w", result.Error)
	}
	
	return nil
}

// buildDriftBottleResponse 构建漂流瓶响应
func (s *DriftBottleService) buildDriftBottleResponse(bottle *models.DriftBottle, letter *models.Letter, user *models.User) *DriftBottleResponse {
	resp := &DriftBottleResponse{
		ID:        bottle.ID,
		Status:    bottle.Status,
		Theme:     bottle.Theme,
		Region:    bottle.Region,
		ExpiresAt: bottle.ExpiresAt,
		CreatedAt: bottle.CreatedAt,
	}
	
	if letter != nil {
		resp.Letter = letter
	}
	
	if bottle.CollectedAt != nil {
		resp.CollectedAt = bottle.CollectedAt
	}
	
	if user != nil {
		resp.Collector = user
	}
	
	return resp
}

// getContentPreview 获取内容预览
func (s *DriftBottleService) getContentPreview(content string, maxLength int) string {
	runes := []rune(content)
	if len(runes) <= maxLength {
		return content
	}
	return string(runes[:maxLength]) + "..."
}

// GetDB 获取数据库实例（用于统计查询）
func (s *DriftBottleService) GetDB() *gorm.DB {
	return s.db
}