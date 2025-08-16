package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Phase 4.2: 积分转赠服务

// CreditTransferService 积分转赠服务
type CreditTransferService struct {
	db                      *gorm.DB
	creditService           *CreditService
	notificationService     *NotificationService
	creditLimiterService    *CreditLimiterService
}

// NewCreditTransferService 创建积分转赠服务实例
func NewCreditTransferService(db *gorm.DB, creditService *CreditService, notificationService *NotificationService, creditLimiterService *CreditLimiterService) *CreditTransferService {
	return &CreditTransferService{
		db:                      db,
		creditService:           creditService,
		notificationService:     notificationService,
		creditLimiterService:    creditLimiterService,
	}
}

// GetDB 获取数据库实例（用于其他服务访问）
func (s *CreditTransferService) GetDB() *gorm.DB {
	return s.db
}

// CreateTransfer 创建积分转赠
func (s *CreditTransferService) CreateTransfer(fromUserID string, request *models.CreateCreditTransferRequest) (*models.CreditTransfer, error) {
	// 验证转赠参数
	if err := s.validateTransferRequest(fromUserID, request); err != nil {
		return nil, err
	}

	// 获取适用的转赠规则
	rule, err := s.getApplicableRule(fromUserID, request.Amount, request.TransferType)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer rule: %w", err)
	}

	// 检查转赠限制
	if err := s.checkTransferLimits(fromUserID, request.Amount, rule); err != nil {
		return nil, err
	}

	// 检查用户积分余额
	fromUserCredit, err := s.creditService.GetUserCreditInfo(fromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user credit: %w", err)
	}

	// 计算手续费
	fee := rule.CalculateFee(request.Amount)
	totalRequired := request.Amount + fee

	if fromUserCredit.Available < totalRequired {
		return nil, fmt.Errorf("insufficient credits: available %d, required %d (amount %d + fee %d)", 
			fromUserCredit.Available, totalRequired, request.Amount, fee)
	}

	// 验证目标用户存在
	var toUser models.User
	if err := s.db.Where("id = ?", request.ToUserID).First(&toUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("target user not found")
		}
		return nil, fmt.Errorf("failed to verify target user: %w", err)
	}

	// 创建转赠记录
	transfer := &models.CreditTransfer{
		ID:           uuid.New().String(),
		FromUserID:   fromUserID,
		ToUserID:     request.ToUserID,
		Amount:       request.Amount,
		TransferType: request.TransferType,
		Status:       models.TransferStatusPending,
		Message:      strings.TrimSpace(request.Message),
		ExpiresAt:    time.Now().Add(time.Duration(rule.ExpirationHours) * time.Hour),
		Fee:          fee,
		Reference:    request.Reference,
		Metadata:     datatypes.JSON{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建转赠记录
	if err := tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transfer: %w", err)
	}

	// 冻结转出用户的积分（扣除积分+手续费）
	if err := s.creditService.SpendPoints(fromUserID, totalRequired, 
		fmt.Sprintf("积分转赠冻结 - 转给用户%s", request.ToUserID), transfer.ID); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to freeze credits: %w", err)
	}

	// 更新转赠限制记录
	if err := s.updateTransferLimits(tx, fromUserID, request.Amount); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update transfer limits: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transfer transaction: %w", err)
	}

	// 发送通知
	s.sendTransferNotifications(transfer)

	log.Printf("Credit transfer created: %s from %s to %s, amount: %d, fee: %d", 
		transfer.ID, fromUserID, request.ToUserID, request.Amount, fee)

	return transfer, nil
}

// ProcessTransfer 处理积分转赠
func (s *CreditTransferService) ProcessTransfer(transferID string, userID string, request *models.ProcessCreditTransferRequest) error {
	// 获取转赠记录
	var transfer models.CreditTransfer
	if err := s.db.Where("id = ?", transferID).First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("transfer not found")
		}
		return fmt.Errorf("failed to get transfer: %w", err)
	}

	// 验证用户权限（只有转入用户可以处理）
	if transfer.ToUserID != userID {
		return fmt.Errorf("permission denied: only recipient can process this transfer")
	}

	// 检查转赠状态
	if !transfer.CanBeProcessed() {
		return fmt.Errorf("transfer cannot be processed: status=%s, expired=%v", 
			transfer.Status, transfer.IsExpired())
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var newStatus models.CreditTransferStatus
	var description string

	if request.Action == "accept" {
		// 接受转赠 - 给接收用户增加积分
		if err := s.creditService.AddPoints(transfer.ToUserID, transfer.Amount, 
			fmt.Sprintf("收到积分转赠 - 来自用户%s", transfer.FromUserID), transfer.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to add credits to recipient: %w", err)
		}
		newStatus = models.TransferStatusProcessed
		description = "积分转赠已接受"
		log.Printf("Credit transfer accepted: %s, amount: %d to user %s", 
			transferID, transfer.Amount, transfer.ToUserID)
	} else {
		// 拒绝转赠 - 返还积分给转出用户（不包括手续费）
		if err := s.creditService.AddPoints(transfer.FromUserID, transfer.Amount, 
			fmt.Sprintf("积分转赠被拒绝退款 - 转给用户%s", transfer.ToUserID), transfer.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to refund credits to sender: %w", err)
		}
		newStatus = models.TransferStatusRejected
		description = fmt.Sprintf("积分转赠被拒绝: %s", request.Reason)
		log.Printf("Credit transfer rejected: %s, amount: %d refunded to user %s, reason: %s", 
			transferID, transfer.Amount, transfer.FromUserID, request.Reason)
	}

	// 更新转赠状态
	now := time.Now()
	if err := tx.Model(&transfer).Updates(map[string]interface{}{
		"status":       newStatus,
		"processed_at": &now,
		"updated_at":   now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transfer status: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit process transaction: %w", err)
	}

	// 发送处理结果通知
	s.sendProcessNotifications(&transfer, newStatus, description)

	return nil
}

// CancelTransfer 取消积分转赠
func (s *CreditTransferService) CancelTransfer(transferID string, userID string) error {
	// 获取转赠记录
	var transfer models.CreditTransfer
	if err := s.db.Where("id = ?", transferID).First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("transfer not found")
		}
		return fmt.Errorf("failed to get transfer: %w", err)
	}

	// 验证用户权限（只有转出用户可以取消）
	if transfer.FromUserID != userID {
		return fmt.Errorf("permission denied: only sender can cancel this transfer")
	}

	// 检查转赠状态
	if !transfer.CanBeCanceled() {
		return fmt.Errorf("transfer cannot be canceled: status=%s, expired=%v", 
			transfer.Status, transfer.IsExpired())
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 返还积分给转出用户（包括手续费）
	totalAmount := transfer.Amount + transfer.Fee
	if err := s.creditService.AddPoints(transfer.FromUserID, totalAmount, 
		fmt.Sprintf("取消积分转赠退款 - 转给用户%s", transfer.ToUserID), transfer.ID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to refund credits: %w", err)
	}

	// 更新转赠状态
	now := time.Now()
	if err := tx.Model(&transfer).Updates(map[string]interface{}{
		"status":       models.TransferStatusCanceled,
		"processed_at": &now,
		"updated_at":   now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transfer status: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit cancel transaction: %w", err)
	}

	log.Printf("Credit transfer canceled: %s, refunded amount: %d to user %s", 
		transferID, totalAmount, transfer.FromUserID)

	// 发送取消通知
	s.sendCancelNotifications(&transfer)

	return nil
}

// GetUserTransfers 获取用户转赠记录
func (s *CreditTransferService) GetUserTransfers(userID string, page, limit int, status models.CreditTransferStatus) ([]models.CreditTransfer, int64, error) {
	var transfers []models.CreditTransfer
	var total int64

	query := s.db.Model(&models.CreditTransfer{}).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transfers: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * limit
	if err := query.Preload("FromUser").Preload("ToUser").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&transfers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get transfers: %w", err)
	}

	return transfers, total, nil
}

// GetTransferStats 获取用户转赠统计
func (s *CreditTransferService) GetTransferStats(userID string) (*models.CreditTransferStatsResponse, error) {
	stats := &models.CreditTransferStatsResponse{}

	// 基本统计查询
	type basicStats struct {
		TotalTransfers    int64 `json:"total_transfers"`
		TotalAmount       int   `json:"total_amount"`
		TotalFees         int   `json:"total_fees"`
		PendingTransfers  int64 `json:"pending_transfers"`
		ProcessedTransfers int64 `json:"processed_transfers"`
		CanceledTransfers int64  `json:"canceled_transfers"`
		ExpiredTransfers  int64  `json:"expired_transfers"`
	}

	var bs basicStats
	err := s.db.Model(&models.CreditTransfer{}).
		Select(`
			COUNT(*) as total_transfers,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(fee), 0) as total_fees,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) as pending_transfers,
			COALESCE(SUM(CASE WHEN status = 'processed' THEN 1 ELSE 0 END), 0) as processed_transfers,
			COALESCE(SUM(CASE WHEN status = 'canceled' THEN 1 ELSE 0 END), 0) as canceled_transfers,
			COALESCE(SUM(CASE WHEN status = 'expired' THEN 1 ELSE 0 END), 0) as expired_transfers
		`).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Scan(&bs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}

	stats.TotalTransfers = bs.TotalTransfers
	stats.TotalAmount = bs.TotalAmount
	stats.TotalFees = bs.TotalFees
	stats.PendingTransfers = bs.PendingTransfers
	stats.ProcessedTransfers = bs.ProcessedTransfers
	stats.CanceledTransfers = bs.CanceledTransfers
	stats.ExpiredTransfers = bs.ExpiredTransfers

	// 计算平均转赠数量
	if stats.TotalTransfers > 0 {
		stats.AverageAmount = float64(stats.TotalAmount) / float64(stats.TotalTransfers)
	}

	// 获取当日和当月限制信息
	rule, err := s.getApplicableRule(userID, 100, models.TransferTypeDirect) // 使用默认参数获取规则
	if err == nil {
		stats.DailyLimit = rule.DailyLimit
		stats.MonthlyLimit = rule.MonthlyLimit

		// 获取当日和当月已使用数量
		monthStart := time.Now().Format("2006-01") + "-01"

		var limits models.CreditTransferLimit
		s.db.Where("user_id = ? AND date >= ?", userID, monthStart).
			Order("date DESC").First(&limits)

		stats.DailyUsed = limits.DailyUsed
		stats.MonthlyUsed = limits.MonthlyUsed
	}

	return stats, nil
}

// BatchTransfer 批量转赠积分
func (s *CreditTransferService) BatchTransfer(fromUserID string, request *models.BatchTransferRequest) (*models.BatchTransferResponse, error) {
	response := &models.BatchTransferResponse{
		Transfers: make([]models.CreditTransfer, 0),
		Errors:    make([]string, 0),
	}

	// 验证批量转赠参数
	if len(request.ToUserIDs) == 0 {
		return nil, fmt.Errorf("recipient list cannot be empty")
	}

	if len(request.ToUserIDs) > 50 { // 限制批量数量
		return nil, fmt.Errorf("batch transfer limit exceeded: max 50 recipients")
	}

	// 去重用户ID
	uniqueUserIDs := make(map[string]bool)
	validUserIDs := make([]string, 0)
	for _, userID := range request.ToUserIDs {
		if !uniqueUserIDs[userID] && userID != fromUserID {
			uniqueUserIDs[userID] = true
			validUserIDs = append(validUserIDs, userID)
		}
	}

	totalRequired := request.Amount * len(validUserIDs)

	// 检查发送者积分余额
	fromUserCredit, err := s.creditService.GetUserCreditInfo(fromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user credit: %w", err)
	}

	rule, err := s.getApplicableRule(fromUserID, request.Amount, request.TransferType)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer rule: %w", err)
	}

	totalFees := rule.CalculateFee(request.Amount) * len(validUserIDs)
	if fromUserCredit.Available < totalRequired + totalFees {
		return nil, fmt.Errorf("insufficient credits for batch transfer: available %d, required %d", 
			fromUserCredit.Available, totalRequired + totalFees)
	}

	// 逐个处理转赠
	for _, toUserID := range validUserIDs {
		transferRequest := &models.CreateCreditTransferRequest{
			ToUserID:     toUserID,
			Amount:       request.Amount,
			TransferType: request.TransferType,
			Message:      request.Message,
			Reference:    request.Reference,
		}

		transfer, err := s.CreateTransfer(fromUserID, transferRequest)
		if err != nil {
			response.FailureCount++
			response.Errors = append(response.Errors, 
				fmt.Sprintf("Failed to transfer to %s: %v", toUserID, err))
		} else {
			response.SuccessCount++
			response.Transfers = append(response.Transfers, *transfer)
		}
	}

	log.Printf("Batch transfer completed: from %s, success: %d, failure: %d", 
		fromUserID, response.SuccessCount, response.FailureCount)

	return response, nil
}

// ProcessExpiredTransfers 处理过期转赠
func (s *CreditTransferService) ProcessExpiredTransfers() error {
	log.Println("Processing expired credit transfers...")

	// 查找过期的待处理转赠
	var expiredTransfers []models.CreditTransfer
	err := s.db.Where("status = ? AND expires_at < ?", 
		models.TransferStatusPending, time.Now()).
		Find(&expiredTransfers).Error

	if err != nil {
		return fmt.Errorf("failed to find expired transfers: %w", err)
	}

	if len(expiredTransfers) == 0 {
		log.Println("No expired transfers found")
		return nil
	}

	processed := 0
	for _, transfer := range expiredTransfers {
		if err := s.processExpiredTransfer(&transfer); err != nil {
			log.Printf("Failed to process expired transfer %s: %v", transfer.ID, err)
			continue
		}
		processed++
	}

	log.Printf("Processed %d expired transfers", processed)
	return nil
}

// 私有方法

// validateTransferRequest 验证转赠请求
func (s *CreditTransferService) validateTransferRequest(fromUserID string, request *models.CreateCreditTransferRequest) error {
	if request.ToUserID == fromUserID {
		return fmt.Errorf("cannot transfer credits to yourself")
	}

	if request.Amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	if len(request.Message) > 500 {
		return fmt.Errorf("message too long: max 500 characters")
	}

	validTypes := map[models.CreditTransferType]bool{
		models.TransferTypeDirect: true,
		models.TransferTypeGift:   true,
		models.TransferTypeReward: true,
	}

	if !validTypes[request.TransferType] {
		return fmt.Errorf("invalid transfer type: %s", request.TransferType)
	}

	return nil
}

// getApplicableRule 获取适用的转赠规则
func (s *CreditTransferService) getApplicableRule(userID string, amount int, transferType models.CreditTransferType) (*models.CreditTransferRule, error) {
	var rule models.CreditTransferRule
	
	// 首先查找最高优先级的活跃规则
	err := s.db.Where("is_active = ?", true).
		Order("priority DESC").
		First(&rule).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有找到规则，创建默认规则
			return s.createDefaultRule()
		}
		return nil, fmt.Errorf("failed to get transfer rule: %w", err)
	}

	// 验证转赠数量是否在规则范围内
	if amount < rule.MinAmount || amount > rule.MaxAmount {
		return nil, fmt.Errorf("transfer amount %d outside allowed range [%d, %d]", 
			amount, rule.MinAmount, rule.MaxAmount)
	}

	return &rule, nil
}

// createDefaultRule 创建默认转赠规则
func (s *CreditTransferService) createDefaultRule() (*models.CreditTransferRule, error) {
	rule := &models.CreditTransferRule{
		ID:                    uuid.New().String(),
		RuleName:              "默认转赠规则",
		MinAmount:             1,
		MaxAmount:             1000,
		DailyLimit:            500,
		MonthlyLimit:          5000,
		FeeRate:               0.02, // 2%手续费
		MinFee:                1,
		MaxFee:                50,
		ExpirationHours:       72,
		RequireConfirmation:   true,
		AllowSelfTransfer:     false,
		RestrictedUserLevels:  datatypes.JSON{},
		AllowedTransferTypes:  datatypes.JSON{},
		IsActive:              true,
		Priority:              0,
		ApplicableUserRoles:   datatypes.JSON{},
		Description:           "系统默认的积分转赠规则",
		CreatedBy:             "system",
		UpdatedBy:             "system",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.db.Create(rule).Error; err != nil {
		return nil, fmt.Errorf("failed to create default rule: %w", err)
	}

	return rule, nil
}

// checkTransferLimits 检查转赠限制
func (s *CreditTransferService) checkTransferLimits(userID string, amount int, rule *models.CreditTransferRule) error {
	today := time.Now()
	dateStr := today.Format("2006-01-02")

	// 获取或创建当日限制记录
	var limit models.CreditTransferLimit
	err := s.db.Where("user_id = ? AND date = ?", userID, dateStr).First(&limit).Error
	
	if err == gorm.ErrRecordNotFound {
		// 创建新的限制记录
		limit = models.CreditTransferLimit{
			ID:       uuid.New().String(),
			UserID:   userID,
			Date:     today.Truncate(24 * time.Hour),
		}
	} else if err != nil {
		return fmt.Errorf("failed to get transfer limits: %w", err)
	}

	// 检查每日限制
	if limit.DailyUsed + amount > rule.DailyLimit {
		return fmt.Errorf("daily transfer limit exceeded: used %d, limit %d, attempting %d", 
			limit.DailyUsed, rule.DailyLimit, amount)
	}

	// 检查每月限制
	if limit.MonthlyUsed + amount > rule.MonthlyLimit {
		return fmt.Errorf("monthly transfer limit exceeded: used %d, limit %d, attempting %d", 
			limit.MonthlyUsed, rule.MonthlyLimit, amount)
	}

	return nil
}

// updateTransferLimits 更新转赠限制记录
func (s *CreditTransferService) updateTransferLimits(tx *gorm.DB, userID string, amount int) error {
	today := time.Now()
	dateStr := today.Format("2006-01-02")

	var limit models.CreditTransferLimit
	err := tx.Where("user_id = ? AND date = ?", userID, dateStr).First(&limit).Error
	
	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		limit = models.CreditTransferLimit{
			ID:           uuid.New().String(),
			UserID:       userID,
			Date:         today.Truncate(24 * time.Hour),
			DailyUsed:    amount,
			MonthlyUsed:  amount,
			DailyCount:   1,
			MonthlyCount: 1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		return tx.Create(&limit).Error
	} else if err != nil {
		return err
	}

	// 更新现有记录
	return tx.Model(&limit).Updates(map[string]interface{}{
		"daily_used":    limit.DailyUsed + amount,
		"monthly_used":  limit.MonthlyUsed + amount,
		"daily_count":   limit.DailyCount + 1,
		"monthly_count": limit.MonthlyCount + 1,
		"updated_at":    time.Now(),
	}).Error
}

// processExpiredTransfer 处理单个过期转赠
func (s *CreditTransferService) processExpiredTransfer(transfer *models.CreditTransfer) error {
	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 返还积分给转出用户（包括手续费）
	totalAmount := transfer.Amount + transfer.Fee
	if err := s.creditService.AddPoints(transfer.FromUserID, totalAmount, 
		fmt.Sprintf("积分转赠过期退款 - 转给用户%s", transfer.ToUserID), transfer.ID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to refund expired transfer: %w", err)
	}

	// 更新转赠状态
	now := time.Now()
	if err := tx.Model(transfer).Updates(map[string]interface{}{
		"status":       models.TransferStatusExpired,
		"processed_at": &now,
		"updated_at":   now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update expired transfer status: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit expired transfer transaction: %w", err)
	}

	// 发送过期通知
	s.sendExpiredNotifications(transfer)

	return nil
}

// 通知相关方法

// sendTransferNotifications 发送转赠通知
func (s *CreditTransferService) sendTransferNotifications(transfer *models.CreditTransfer) {
	if s.notificationService == nil {
		return
	}

	// 通知转出用户
	go s.notificationService.NotifyUser(transfer.FromUserID, "credit_transfer_sent", map[string]interface{}{
		"transfer_id": transfer.ID,
		"amount":      transfer.Amount,
		"fee":         transfer.Fee,
		"to_user_id":  transfer.ToUserID,
		"message":     transfer.Message,
	})

	// 通知转入用户
	go s.notificationService.NotifyUser(transfer.ToUserID, "credit_transfer_received", map[string]interface{}{
		"transfer_id":  transfer.ID,
		"amount":       transfer.Amount,
		"from_user_id": transfer.FromUserID,
		"message":      transfer.Message,
		"expires_at":   transfer.ExpiresAt,
	})
}

// sendProcessNotifications 发送处理结果通知
func (s *CreditTransferService) sendProcessNotifications(transfer *models.CreditTransfer, status models.CreditTransferStatus, description string) {
	if s.notificationService == nil {
		return
	}

	notificationType := "credit_transfer_processed"
	if status == models.TransferStatusRejected {
		notificationType = "credit_transfer_rejected"
	}

	// 通知转出用户
	go s.notificationService.NotifyUser(transfer.FromUserID, notificationType, map[string]interface{}{
		"transfer_id": transfer.ID,
		"amount":      transfer.Amount,
		"status":      status,
		"description": description,
		"to_user_id":  transfer.ToUserID,
	})
}

// sendCancelNotifications 发送取消通知
func (s *CreditTransferService) sendCancelNotifications(transfer *models.CreditTransfer) {
	if s.notificationService == nil {
		return
	}

	// 通知转入用户转赠被取消
	go s.notificationService.NotifyUser(transfer.ToUserID, "credit_transfer_canceled", map[string]interface{}{
		"transfer_id":  transfer.ID,
		"amount":       transfer.Amount,
		"from_user_id": transfer.FromUserID,
		"message":      transfer.Message,
	})
}

// sendExpiredNotifications 发送过期通知
func (s *CreditTransferService) sendExpiredNotifications(transfer *models.CreditTransfer) {
	if s.notificationService == nil {
		return
	}

	// 通知双方转赠过期
	go s.notificationService.NotifyUser(transfer.FromUserID, "credit_transfer_expired", map[string]interface{}{
		"transfer_id": transfer.ID,
		"amount":      transfer.Amount,
		"to_user_id":  transfer.ToUserID,
		"refunded":    transfer.Amount + transfer.Fee,
	})

	go s.notificationService.NotifyUser(transfer.ToUserID, "credit_transfer_expired", map[string]interface{}{
		"transfer_id":  transfer.ID,
		"amount":       transfer.Amount,
		"from_user_id": transfer.FromUserID,
	})
}