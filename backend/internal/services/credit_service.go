package services

import (
	"fmt"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreditService 积分系统服务
type CreditService struct {
	db              *gorm.DB
	notificationSvc *NotificationService
}

// NewCreditService 创建积分服务
func NewCreditService(db *gorm.DB) *CreditService {
	return &CreditService{
		db: db,
	}
}

// SetNotificationService 设置通知服务
func (s *CreditService) SetNotificationService(notificationSvc *NotificationService) {
	s.notificationSvc = notificationSvc
}

// 积分奖励规则 - 严格按照FSD规格配置
const (
	// 信件相关积分 - FSD规格匹配
	PointsLetterCreated    = 10 // 成功写信并绑定条码 (FSD: +10)
	PointsLetterGenerated  = 10 // 生成编号 (与创建信件相同)
	PointsLetterDelivered  = 20 // 信件送达 (内部奖励)
	PointsLetterRead       = 15 // 信件被阅读 (内部奖励)
	PointsReceiveLetter    = 5  // 被回信 (FSD: +5) - 修正
	PointsPublicLetterLike = 1  // 公开信被点赞 (FSD: +1) - 新增

	// 写作与挑战相关积分 - FSD新增
	PointsWritingChallenge = 15 // 参与写作挑战并完成投稿 (FSD: +15)
	PointsAIInteraction    = 3  // 使用AI笔友并留下评价 (FSD: +3)

	// 信使相关积分 - FSD规格
	PointsCourierFirstTask = 20 // 成为信使后首次完成任务 (FSD: +20)
	PointsCourierDelivery  = 5  // 信使每成功送达一封信 (FSD: +5)

	// 信封相关积分
	PointsEnvelopePurchase = 2 // 购买信封
	PointsEnvelopeBinding  = 3 // 绑定信封

	// 博物馆相关积分 - FSD规格修正
	PointsMuseumSubmit   = 25  // 提交作品到博物馆
	PointsMuseumApproved = 100 // 审核通过投稿信封被采纳 (FSD: +100) - 修正
	PointsMuseumLiked    = 5   // 作品获得点赞

	// 系统管理相关积分 - FSD新增
	PointsOPCodeApproval    = 10 // 点位申请审核成功 (FSD: +10)
	PointsCommunityBadge    = 50 // 被授予社区贡献徽章 (FSD: +50)
	PointsAdminReward       = 0  // 管理员手动奖励 (可变)
)

// 等级升级所需积分
var PointsLevelUp = []int{0, 100, 300, 600, 1000, 1500} // 每级所需总积分

// GetOrCreateUserCredit 获取或创建用户积分记录
func (s *CreditService) GetOrCreateUserCredit(userID string) (*models.UserCredit, error) {
	var credit models.UserCredit
	if err := s.db.Where("user_id = ?", userID).First(&credit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			credit = models.UserCredit{
				ID:        uuid.New().String(),
				UserID:    userID,
				Total:     0,
				Available: 0,
				Used:      0,
				Earned:    0,
				Level:     1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := s.db.Create(&credit).Error; err != nil {
				return nil, fmt.Errorf("failed to create user credit: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get user credit: %w", err)
		}
	}
	return &credit, nil
}

// AddPoints 增加用户积分
func (s *CreditService) AddPoints(userID string, points int, description, reference string) error {
	if points <= 0 {
		return fmt.Errorf("points must be positive")
	}

	// 获取用户积分记录
	credit, err := s.GetOrCreateUserCredit(userID)
	if err != nil {
		return err
	}

	// 开始事务
	tx := s.db.Begin()

	// 更新积分
	oldLevel := credit.Level
	credit.Total += points
	credit.Available += points
	credit.Earned += points

	// 检查等级升级
	newLevel := s.calculateLevel(credit.Total)
	if newLevel > credit.Level {
		credit.Level = newLevel
	}

	if err := tx.Save(credit).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user credit: %w", err)
	}

	// 创建交易记录
	transaction := models.CreditTransaction{
		ID:          uuid.New().String(),
		UserID:      userID,
		Type:        "earn",
		Amount:      points,
		Description: description,
		Reference:   reference,
		CreatedAt:   time.Now(),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	tx.Commit()

	// 发送通知
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyUser(userID, "points_earned", map[string]interface{}{
			"points":      points,
			"description": description,
			"reference":   reference,
			"total":       credit.Total,
		})

		// 如果升级了，发送升级通知
		if newLevel > oldLevel {
			s.notificationSvc.NotifyUser(userID, "level_up", map[string]interface{}{
				"old_level": oldLevel,
				"new_level": newLevel,
				"points":    credit.Total,
			})
		}
	}

	return nil
}

// SpendPoints 消费积分
func (s *CreditService) SpendPoints(userID string, points int, description, reference string) error {
	if points <= 0 {
		return fmt.Errorf("points must be positive")
	}

	// 获取用户积分记录
	credit, err := s.GetOrCreateUserCredit(userID)
	if err != nil {
		return err
	}

	// 检查积分是否足够
	if credit.Available < points {
		return fmt.Errorf("insufficient credits: available %d, required %d", credit.Available, points)
	}

	// 开始事务
	tx := s.db.Begin()

	// 更新积分
	credit.Available -= points
	credit.Used += points

	if err := tx.Save(credit).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user credit: %w", err)
	}

	// 创建交易记录
	transaction := models.CreditTransaction{
		ID:          uuid.New().String(),
		UserID:      userID,
		Type:        "spend",
		Amount:      -points, // 负数表示消费
		Description: description,
		Reference:   reference,
		CreatedAt:   time.Now(),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	tx.Commit()

	// 发送通知
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyUser(userID, "points_deducted", map[string]interface{}{
			"points":      points,
			"description": description,
			"reference":   reference,
			"remaining":   credit.Available,
		})
	}

	return nil
}

// calculateLevel 根据总积分计算等级
func (s *CreditService) calculateLevel(totalPoints int) int {
	for level := len(PointsLevelUp) - 1; level >= 1; level-- {
		if totalPoints >= PointsLevelUp[level] {
			return level
		}
	}
	return 1
}

// GetUserCreditInfo 获取用户积分信息
func (s *CreditService) GetUserCreditInfo(userID string) (*models.UserCredit, error) {
	return s.GetOrCreateUserCredit(userID)
}

// GetCreditHistory 获取积分历史记录
func (s *CreditService) GetCreditHistory(userID string, limit, offset int) ([]models.CreditTransaction, int64, error) {
	var transactions []models.CreditTransaction
	var total int64

	// 获取总数
	if err := s.db.Model(&models.CreditTransaction{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// 获取分页数据
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	return transactions, total, nil
}

// 信件相关积分奖励方法

// RewardLetterCreated 奖励创建信件
func (s *CreditService) RewardLetterCreated(userID, letterID string) error {
	return s.AddPoints(userID, PointsLetterCreated, "创建信件", letterID)
}

// RewardLetterGenerated 奖励生成信件编号
func (s *CreditService) RewardLetterGenerated(userID, letterID string) error {
	return s.AddPoints(userID, PointsLetterGenerated, "生成信件编号", letterID)
}

// RewardLetterDelivered 奖励信件送达
func (s *CreditService) RewardLetterDelivered(userID, letterID string) error {
	return s.AddPoints(userID, PointsLetterDelivered, "信件送达", letterID)
}

// RewardLetterRead 奖励信件被阅读
func (s *CreditService) RewardLetterRead(userID, letterID string) error {
	return s.AddPoints(userID, PointsLetterRead, "信件被阅读", letterID)
}

// RewardReceiveLetter 奖励收到信件
func (s *CreditService) RewardReceiveLetter(userID, letterID string) error {
	return s.AddPoints(userID, PointsReceiveLetter, "收到信件", letterID)
}

// RewardReply 奖励回信
func (s *CreditService) RewardReply(userID, replyID string) error {
	return s.AddPoints(userID, PointsLetterCreated, "创建回信", replyID) // 使用与创建信件相同的积分
}

// 信封相关积分奖励方法

// RewardEnvelopePurchase 奖励购买信封
func (s *CreditService) RewardEnvelopePurchase(userID, orderID string, quantity int) error {
	points := PointsEnvelopePurchase * quantity
	return s.AddPoints(userID, points, fmt.Sprintf("购买%d个信封", quantity), orderID)
}

// RewardEnvelopeBinding 奖励绑定信封
func (s *CreditService) RewardEnvelopeBinding(userID, letterID string) error {
	return s.AddPoints(userID, PointsEnvelopeBinding, "绑定信封", letterID)
}

// 博物馆相关积分奖励方法

// RewardMuseumSubmit 奖励提交博物馆作品
func (s *CreditService) RewardMuseumSubmit(userID, submissionID string) error {
	return s.AddPoints(userID, PointsMuseumSubmit, "提交博物馆作品", submissionID)
}

// RewardMuseumApproved 奖励博物馆作品通过审核
func (s *CreditService) RewardMuseumApproved(userID, submissionID string) error {
	return s.AddPoints(userID, PointsMuseumApproved, "博物馆作品通过审核", submissionID)
}

// RewardMuseumLiked 奖励博物馆作品获得点赞
func (s *CreditService) RewardMuseumLiked(userID, submissionID string) error {
	return s.AddPoints(userID, PointsMuseumLiked, "博物馆作品获得点赞", submissionID)
}

// ========================= FSD新增积分奖励方法 =========================

// RewardPublicLetterLike 奖励公开信被点赞 - FSD规格
func (s *CreditService) RewardPublicLetterLike(userID, letterID string) error {
	return s.AddPoints(userID, PointsPublicLetterLike, "公开信被点赞", letterID)
}

// RewardWritingChallenge 奖励参与写作挑战 - FSD规格
func (s *CreditService) RewardWritingChallenge(userID, challengeID string) error {
	return s.AddPoints(userID, PointsWritingChallenge, "参与写作挑战并完成投稿", challengeID)
}

// RewardAIInteraction 奖励AI互动评价 - FSD规格
func (s *CreditService) RewardAIInteraction(userID, sessionID string) error {
	return s.AddPoints(userID, PointsAIInteraction, "使用AI笔友并留下评价", sessionID)
}

// RewardCourierFirstTask 奖励信使首次任务完成 - FSD规格
func (s *CreditService) RewardCourierFirstTask(userID, taskID string) error {
	return s.AddPoints(userID, PointsCourierFirstTask, "成为信使后首次完成任务", taskID)
}

// RewardCourierDelivery 奖励信使送达信件 - FSD规格
func (s *CreditService) RewardCourierDelivery(userID, taskID string) error {
	return s.AddPoints(userID, PointsCourierDelivery, "信使成功送达一封信", taskID)
}

// RewardOPCodeApproval 奖励点位申请审核成功 - FSD规格
func (s *CreditService) RewardOPCodeApproval(userID, applicationID string) error {
	return s.AddPoints(userID, PointsOPCodeApproval, "点位申请审核成功", applicationID)
}

// RewardCommunityBadge 奖励社区贡献徽章 - FSD规格
func (s *CreditService) RewardCommunityBadge(userID, badgeID string) error {
	return s.AddPoints(userID, PointsCommunityBadge, "被授予社区贡献徽章", badgeID)
}

// RewardAdminCustom 管理员手动奖励积分 - FSD规格
func (s *CreditService) RewardAdminCustom(userID string, points int, description, reference string) error {
	if points <= 0 {
		return fmt.Errorf("admin reward points must be positive")
	}
	return s.AddPoints(userID, points, fmt.Sprintf("管理员奖励: %s", description), reference)
}

// ========================= 积分限制与验证机制 =========================

// CheckDailyLimit 检查每日积分限制 - FSD风控要求
func (s *CreditService) CheckDailyLimit(userID, actionType string) (bool, error) {
	today := time.Now().Format("2006-01-02")
	
	var dailyLimits = map[string]int{
		"letter_created":    3,  // 每日上限3封
		"receive_letter":    5,  // 每日上限5封
		"public_like":       20, // 每封信上限20赞
		"writing_challenge": 1,  // 每周限一次
		"ai_interaction":    3,  // 每日限3次
	}
	
	limit, exists := dailyLimits[actionType]
	if !exists {
		return true, nil // 无限制的行为类型
	}
	
	var count int64
	err := s.db.Model(&models.CreditTransaction{}).
		Where("user_id = ? AND description LIKE ? AND DATE(created_at) = ?", 
			userID, "%"+actionType+"%", today).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return int(count) < limit, nil
}

// GetDailyStats 获取用户每日积分统计 - 风控监控
func (s *CreditService) GetDailyStats(userID string) (map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	
	var dailyStats struct {
		TotalPoints int64 `json:"total_points"`
		TotalTrans  int64 `json:"total_transactions"`
	}
	
	err := s.db.Model(&models.CreditTransaction{}).
		Select("COALESCE(SUM(amount), 0) as total_points, COUNT(*) as total_trans").
		Where("user_id = ? AND DATE(created_at) = ? AND amount > 0", userID, today).
		Scan(&dailyStats).Error
	
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"date":          today,
		"total_points":  dailyStats.TotalPoints,
		"transactions":  dailyStats.TotalTrans,
		"user_id":       userID,
	}, nil
}

// GetLeaderboard 获取积分排行榜
func (s *CreditService) GetLeaderboard(limit int) ([]models.UserCredit, error) {
	var leaderboard []models.UserCredit
	if err := s.db.Order("total DESC").Limit(limit).Find(&leaderboard).Error; err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}
	return leaderboard, nil
}
