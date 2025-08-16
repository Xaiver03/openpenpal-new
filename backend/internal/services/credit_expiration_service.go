package services

import (
	"fmt"
	"log"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreditExpirationService 积分过期服务
type CreditExpirationService struct {
	db                    *gorm.DB
	creditService         *CreditService
	notificationService   *NotificationService
}

// NewCreditExpirationService 创建积分过期服务实例
func NewCreditExpirationService(db *gorm.DB, creditService *CreditService, notificationService *NotificationService) *CreditExpirationService {
	return &CreditExpirationService{
		db:                  db,
		creditService:       creditService,
		notificationService: notificationService,
	}
}

// SetCreditService 设置积分服务依赖
func (s *CreditExpirationService) SetCreditService(creditService *CreditService) {
	s.creditService = creditService
}

// SetNotificationService 设置通知服务依赖
func (s *CreditExpirationService) SetNotificationService(notificationService *NotificationService) {
	s.notificationService = notificationService
}

// AddExpirationToTransaction 为积分交易添加过期时间
func (s *CreditExpirationService) AddExpirationToTransaction(transaction *models.CreditTransaction, creditType string) error {
	// 获取对应的过期规则
	var rule models.CreditExpirationRule
	err := s.db.Where("credit_type = ? AND is_active = ?", creditType, true).
		Order("priority DESC").
		First(&rule).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有找到特定规则，使用默认规则
			err = s.db.Where("credit_type = ? AND is_active = ?", "default", true).
				First(&rule).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// 如果连默认规则都没有，不设置过期时间
					log.Printf("No expiration rule found for credit type: %s", creditType)
					return nil
				}
				return fmt.Errorf("failed to get default expiration rule: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get expiration rule: %w", err)
		}
	}

	// 计算过期时间
	expirationTime := transaction.CreatedAt.AddDate(0, 0, rule.ExpirationDays)
	transaction.ExpiresAt = &expirationTime

	// 更新数据库中的交易记录
	return s.db.Model(transaction).Updates(map[string]interface{}{
		"expires_at": expirationTime,
	}).Error
}

// ProcessExpiredCredits 处理过期积分（定时任务）
func (s *CreditExpirationService) ProcessExpiredCredits() error {
	now := time.Now()
	
	// 创建过期批次记录
	batch := &models.CreditExpirationBatch{
		ID:        uuid.New().String(),
		BatchDate: now,
		Status:    "processing",
		StartedAt: &now,
		ProcessedBy: "system",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.db.Create(batch).Error; err != nil {
		return fmt.Errorf("failed to create expiration batch: %w", err)
	}

	log.Printf("Starting credit expiration batch processing: %s", batch.ID)

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找所有过期的积分交易
	var expiredTransactions []models.CreditTransaction
	err := tx.Where("expires_at IS NOT NULL AND expires_at <= ? AND is_expired = ?", now, false).
		Find(&expiredTransactions).Error
	
	if err != nil {
		tx.Rollback()
		s.updateBatchStatus(batch, "failed", fmt.Sprintf("Failed to find expired transactions: %v", err))
		return fmt.Errorf("failed to find expired transactions: %w", err)
	}

	if len(expiredTransactions) == 0 {
		tx.Commit()
		s.updateBatchStatus(batch, "completed", "No expired credits found")
		log.Printf("No expired credits found in batch: %s", batch.ID)
		return nil
	}

	log.Printf("Found %d expired credit transactions", len(expiredTransactions))

	// 统计信息
	totalExpiredCredits := 0
	affectedUsers := make(map[string]bool)
	
	// 处理每个过期交易
	for _, transaction := range expiredTransactions {
		// 标记为过期
		expiredAt := now
		err = tx.Model(&transaction).Updates(map[string]interface{}{
			"is_expired":  true,
			"expired_at":  expiredAt,
		}).Error

		if err != nil {
			tx.Rollback()
			s.updateBatchStatus(batch, "failed", fmt.Sprintf("Failed to update transaction %s: %v", transaction.ID, err))
			return fmt.Errorf("failed to update expired transaction: %w", err)
		}

		// 从用户可用积分中扣除
		err = s.deductExpiredCreditsFromUser(tx, transaction.UserID, transaction.Amount)
		if err != nil {
			tx.Rollback()
			s.updateBatchStatus(batch, "failed", fmt.Sprintf("Failed to deduct credits from user %s: %v", transaction.UserID, err))
			return fmt.Errorf("failed to deduct expired credits: %w", err)
		}

		// 记录过期日志
		expirationLog := &models.CreditExpirationLog{
			ID:               uuid.New().String(),
			BatchID:          batch.ID,
			UserID:           transaction.UserID,
			TransactionID:    transaction.ID,
			ExpiredCredits:   transaction.Amount,
			OriginalAmount:   transaction.Amount,
			ExpirationReason: "Reached expiration date",
			CreatedAt:        now,
		}

		if err = tx.Create(expirationLog).Error; err != nil {
			tx.Rollback()
			s.updateBatchStatus(batch, "failed", fmt.Sprintf("Failed to create expiration log: %v", err))
			return fmt.Errorf("failed to create expiration log: %w", err)
		}

		totalExpiredCredits += transaction.Amount
		affectedUsers[transaction.UserID] = true
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		s.updateBatchStatus(batch, "failed", fmt.Sprintf("Failed to commit transaction: %v", err))
		return fmt.Errorf("failed to commit expiration transaction: %w", err)
	}

	// 更新批次统计
	completedAt := time.Now()
	batch.TotalCredits = totalExpiredCredits
	batch.TotalUsers = len(affectedUsers)
	batch.TotalTransactions = len(expiredTransactions)
	batch.Status = "completed"
	batch.CompletedAt = &completedAt

	if err = s.db.Save(batch).Error; err != nil {
		log.Printf("Failed to update batch statistics: %v", err)
	}

	log.Printf("Credit expiration batch completed: %s, expired %d credits for %d users", 
		batch.ID, totalExpiredCredits, len(affectedUsers))

	// 异步发送通知
	go s.sendExpirationNotifications(affectedUsers)

	return nil
}

// deductExpiredCreditsFromUser 从用户账户中扣除过期积分
func (s *CreditExpirationService) deductExpiredCreditsFromUser(tx *gorm.DB, userID string, amount int) error {
	var userCredit models.UserCredit
	err := tx.Where("user_id = ?", userID).First(&userCredit).Error
	if err != nil {
		return fmt.Errorf("failed to get user credit: %w", err)
	}

	// 确保可用积分不为负数
	newAvailable := userCredit.Available - amount
	if newAvailable < 0 {
		log.Printf("Warning: User %s available credits would go negative (%d - %d = %d), setting to 0",
			userID, userCredit.Available, amount, newAvailable)
		newAvailable = 0
	}

	// 更新用户积分
	return tx.Model(&userCredit).Updates(map[string]interface{}{
		"available": newAvailable,
	}).Error
}

// SendExpirationWarnings 发送积分即将过期的警告
func (s *CreditExpirationService) SendExpirationWarnings() error {
	log.Println("Starting to send expiration warnings...")

	// 获取所有启用的过期规则
	var rules []models.CreditExpirationRule
	if err := s.db.Where("is_active = ?", true).Find(&rules).Error; err != nil {
		return fmt.Errorf("failed to get expiration rules: %w", err)
	}

	now := time.Now()
	for _, rule := range rules {
		// 计算警告时间范围
		warningStartDate := now.AddDate(0, 0, rule.NotifyDays)
		warningEndDate := now.AddDate(0, 0, rule.NotifyDays+1)

		// 查找即将过期的积分
		var transactions []models.CreditTransaction
		err := s.db.Where("expires_at >= ? AND expires_at < ? AND is_expired = ? AND type = ?",
			warningStartDate, warningEndDate, false, "earn").
			Find(&transactions).Error

		if err != nil {
			log.Printf("Failed to find transactions for warning (rule %s): %v", rule.ID, err)
			continue
		}

		// 按用户分组发送通知
		userCredits := make(map[string]int)
		userDates := make(map[string]time.Time)

		for _, transaction := range transactions {
			userCredits[transaction.UserID] += transaction.Amount
			if userDates[transaction.UserID].IsZero() || transaction.ExpiresAt.Before(userDates[transaction.UserID]) {
				userDates[transaction.UserID] = *transaction.ExpiresAt
			}
		}

		// 发送警告通知
		for userID, credits := range userCredits {
			expirationDate := userDates[userID]
			
			// 检查是否已经发送过通知
			var existingNotification models.CreditExpirationNotification
			err = s.db.Where("user_id = ? AND notification_type = ? AND expiration_date = ?",
				userID, "warning", expirationDate).First(&existingNotification).Error

			if err == nil {
				// 已经发送过通知，跳过
				continue
			}

			// 创建通知记录
			notification := &models.CreditExpirationNotification{
				ID:               uuid.New().String(),
				UserID:           userID,
				NotificationType: "warning",
				CreditsToExpire:  credits,
				ExpirationDate:   expirationDate,
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			if err = s.db.Create(notification).Error; err != nil {
				log.Printf("Failed to create expiration notification: %v", err)
				continue
			}

			// 发送通知（如果通知服务可用）
			s.sendWarningNotification(userID, credits, expirationDate, notification.ID)
		}
	}

	log.Println("Finished sending expiration warnings")
	return nil
}

// sendWarningNotification 发送积分即将过期的警告通知
func (s *CreditExpirationService) sendWarningNotification(userID string, credits int, expirationDate time.Time, notificationID string) {
	if s.notificationService == nil {
		log.Printf("Notification service not available, skipping warning for user %s", userID)
		return
	}

	daysUntilExpiration := int(time.Until(expirationDate).Hours() / 24)
	
	title := "积分即将过期提醒"
	message := fmt.Sprintf("您有 %d 积分将在 %d 天后（%s）过期，请及时使用。", 
		credits, daysUntilExpiration, expirationDate.Format("2006-01-02"))

	// 异步发送通知
	go func() {
		req := &models.SendNotificationRequest{
			UserIDs: []string{userID},
			Type:    models.NotificationSystem,
			Channel: models.ChannelEmail,
			Title:   title,
			Content: message,
			Data: map[string]interface{}{
				"category": "credit_expiration_warning",
			},
		}
		err := s.notificationService.SendEmailNotification(req)
		
		// 更新通知状态
		updateData := map[string]interface{}{
			"notification_sent": err == nil,
			"notification_time": time.Now(),
		}
		
		if err != nil {
			updateData["notification_error"] = err.Error()
			log.Printf("Failed to send expiration warning to user %s: %v", userID, err)
		}

		s.db.Model(&models.CreditExpirationNotification{}).
			Where("id = ?", notificationID).
			Updates(updateData)
	}()
}

// sendExpirationNotifications 发送积分已过期的通知
func (s *CreditExpirationService) sendExpirationNotifications(affectedUsers map[string]bool) {
	if s.notificationService == nil {
		log.Println("Notification service not available, skipping expiration notifications")
		return
	}

	for userID := range affectedUsers {
		// 计算用户过期的积分总数
		var totalExpired int
		err := s.db.Model(&models.CreditExpirationLog{}).
			Where("user_id = ? AND created_at >= ?", userID, time.Now().AddDate(0, 0, -1)).
			Select("COALESCE(SUM(expired_credits), 0)").
			Scan(&totalExpired).Error

		if err != nil {
			log.Printf("Failed to calculate expired credits for user %s: %v", userID, err)
			continue
		}

		if totalExpired == 0 {
			continue
		}

		title := "积分过期通知"
		message := fmt.Sprintf("您有 %d 积分已过期，请注意及时使用积分以免过期。", totalExpired)

		// 创建通知记录
		notification := &models.CreditExpirationNotification{
			ID:               uuid.New().String(),
			UserID:           userID,
			NotificationType: "expired",
			CreditsToExpire:  totalExpired,
			ExpirationDate:   time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err = s.db.Create(notification).Error; err != nil {
			log.Printf("Failed to create expired notification record: %v", err)
			continue
		}

		// 发送通知
		go func(uid, nid string) {
			req := &models.SendNotificationRequest{
				UserIDs: []string{uid},
				Type:    models.NotificationSystem,
				Channel: models.ChannelEmail,
				Title:   title,
				Content: message,
				Data: map[string]interface{}{
					"category": "credit_expired",
				},
			}
			err := s.notificationService.SendEmailNotification(req)
			
			updateData := map[string]interface{}{
				"notification_sent": err == nil,
				"notification_time": time.Now(),
			}
			
			if err != nil {
				updateData["notification_error"] = err.Error()
				log.Printf("Failed to send expiration notification to user %s: %v", uid, err)
			}

			s.db.Model(&models.CreditExpirationNotification{}).
				Where("id = ?", nid).
				Updates(updateData)
		}(userID, notification.ID)
	}
}

// updateBatchStatus 更新批次状态
func (s *CreditExpirationService) updateBatchStatus(batch *models.CreditExpirationBatch, status, errorMessage string) {
	updateData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == "completed" {
		completedAt := time.Now()
		updateData["completed_at"] = completedAt
	}

	if errorMessage != "" {
		updateData["error_message"] = errorMessage
	}

	if err := s.db.Model(batch).Updates(updateData).Error; err != nil {
		log.Printf("Failed to update batch status: %v", err)
	}
}

// GetExpirationRules 获取过期规则列表
func (s *CreditExpirationService) GetExpirationRules() ([]models.CreditExpirationRule, error) {
	var rules []models.CreditExpirationRule
	err := s.db.Order("priority DESC, created_at DESC").Find(&rules).Error
	return rules, err
}

// CreateExpirationRule 创建过期规则
func (s *CreditExpirationService) CreateExpirationRule(rule *models.CreditExpirationRule) error {
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	
	return s.db.Create(rule).Error
}

// UpdateExpirationRule 更新过期规则
func (s *CreditExpirationService) UpdateExpirationRule(ruleID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.db.Model(&models.CreditExpirationRule{}).Where("id = ?", ruleID).Updates(updates).Error
}

// DeleteExpirationRule 删除过期规则
func (s *CreditExpirationService) DeleteExpirationRule(ruleID string) error {
	return s.db.Delete(&models.CreditExpirationRule{}, "id = ?", ruleID).Error
}

// GetUserExpiringCredits 获取用户即将过期的积分
func (s *CreditExpirationService) GetUserExpiringCredits(userID string, days int) ([]models.CreditTransaction, int, error) {
	futureDate := time.Now().AddDate(0, 0, days)
	
	var transactions []models.CreditTransaction
	err := s.db.Where("user_id = ? AND expires_at IS NOT NULL AND expires_at <= ? AND is_expired = ?",
		userID, futureDate, false).
		Order("expires_at ASC").
		Find(&transactions).Error
	
	if err != nil {
		return nil, 0, err
	}

	// 计算总积分
	totalCredits := 0
	for _, transaction := range transactions {
		totalCredits += transaction.Amount
	}

	return transactions, totalCredits, nil
}

// GetExpirationStatistics 获取过期统计信息
func (s *CreditExpirationService) GetExpirationStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 今日过期积分
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)
	
	var todayExpired int
	err := s.db.Model(&models.CreditExpirationLog{}).
		Where("created_at >= ? AND created_at < ?", today, tomorrow).
		Select("COALESCE(SUM(expired_credits), 0)").
		Scan(&todayExpired).Error
	if err != nil {
		return nil, err
	}
	stats["today_expired"] = todayExpired

	// 本周过期积分
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	var weekExpired int
	err = s.db.Model(&models.CreditExpirationLog{}).
		Where("created_at >= ?", weekStart).
		Select("COALESCE(SUM(expired_credits), 0)").
		Scan(&weekExpired).Error
	if err != nil {
		return nil, err
	}
	stats["week_expired"] = weekExpired

	// 即将过期的积分（7天内）
	futureDate := time.Now().AddDate(0, 0, 7)
	var soonToExpire int
	err = s.db.Model(&models.CreditTransaction{}).
		Where("expires_at IS NOT NULL AND expires_at <= ? AND is_expired = ?", futureDate, false).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&soonToExpire).Error
	if err != nil {
		return nil, err
	}
	stats["soon_to_expire"] = soonToExpire

	// 活跃的过期规则数量
	var activeRules int64
	err = s.db.Model(&models.CreditExpirationRule{}).Where("is_active = ?", true).Count(&activeRules).Error
	if err != nil {
		return nil, err
	}
	stats["active_rules"] = activeRules

	return stats, nil
}