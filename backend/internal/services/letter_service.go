package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/websocket"
	"openpenpal-backend/pkg/utils"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

type LetterService struct {
	db              *gorm.DB
	config          *config.Config
	courierTaskSvc  *CourierTaskService
	notificationSvc *NotificationService
	creditSvc       *CreditService
	aiSvc           *AIService
	wsService       *websocket.WebSocketService
	opcodeService   *OPCodeService // OP Code验证服务
}

func NewLetterService(db *gorm.DB, config *config.Config) *LetterService {
	return &LetterService{
		db:     db,
		config: config,
	}
}

// SetCourierTaskService 设置信使任务服务（避免循环依赖）
func (s *LetterService) SetCourierTaskService(courierTaskSvc *CourierTaskService) {
	s.courierTaskSvc = courierTaskSvc
}

// SetNotificationService 设置通知服务（避免循环依赖）
func (s *LetterService) SetNotificationService(notificationSvc *NotificationService) {
	s.notificationSvc = notificationSvc
}

// SetCreditService 设置积分服务（避免循环依赖）
func (s *LetterService) SetCreditService(creditSvc *CreditService) {
	s.creditSvc = creditSvc
}

// SetWebSocketService 设置WebSocket服务（避免循环依赖）
func (s *LetterService) SetWebSocketService(wsService *websocket.WebSocketService) {
	s.wsService = wsService
}

// SetAIService 设置AI服务（避免循环依赖）
func (s *LetterService) SetAIService(aiSvc *AIService) {
	s.aiSvc = aiSvc
}

// SetOPCodeService 设置OP Code服务（避免循环依赖）
func (s *LetterService) SetOPCodeService(opcodeService *OPCodeService) {
	s.opcodeService = opcodeService
}

// CreateDraft 创建草稿
func (s *LetterService) CreateDraft(userID string, req *models.CreateLetterRequest) (*models.Letter, error) {
	letter := &models.Letter{
		ID:         uuid.New().String(),
		UserID:     userID,
		AuthorID:   userID, // 设置作者ID
		Title:      req.Title,
		Content:    req.Content,
		Style:      req.Style,
		Status:     models.StatusDraft,
		Visibility: models.VisibilityPrivate, // 设置默认可见性
		ReplyTo:    req.ReplyTo,
	}

	if err := s.db.Create(letter).Error; err != nil {
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}

	// 奖励创建信件积分
	if s.creditSvc != nil {
		go func() {
			if err := s.creditSvc.RewardLetterCreated(userID, letter.ID); err != nil {
				fmt.Printf("Failed to reward letter created: %v\n", err)
			}
		}()
	}

	return letter, nil
}

// GenerateCode 生成信件编号和二维码
func (s *LetterService) GenerateCode(letterID string) (*models.LetterCode, error) {
	// 检查信件是否存在
	var letter models.Letter
	if err := s.db.First(&letter, "id = ?", letterID).Error; err != nil {
		return nil, fmt.Errorf("letter not found: %w", err)
	}

	// 检查是否已经生成过编号
	var existingCode models.LetterCode
	if err := s.db.First(&existingCode, "letter_id = ?", letterID).Error; err == nil {
		return &existingCode, nil
	}

	// 生成唯一编号
	code := utils.GenerateLetterCode()

	// 生成包含OP Code信息的结构化QR码数据
	qrData := s.generateEnhancedQRData(letter.ID, code, letter.RecipientOPCode, letter.SenderOPCode)
	
	// 生成二维码（使用增强数据）
	qrCodeFileName := fmt.Sprintf("%s.png", code)
	qrCodePath := filepath.Join(s.config.QRCodeStorePath, qrCodeFileName)
	qrCodeURL := fmt.Sprintf("%s/uploads/qrcodes/%s", s.config.BaseURL, qrCodeFileName)

	// 确保目录存在
	if err := utils.EnsureDir(s.config.QRCodeStorePath); err != nil {
		return nil, fmt.Errorf("failed to create qr code directory: %w", err)
	}

	// 生成二维码文件（使用结构化数据）
	if err := qrcode.WriteFile(qrData, qrcode.Medium, 256, qrCodePath); err != nil {
		return nil, fmt.Errorf("failed to generate qr code: %w", err)
	}

	// 保存到数据库
	letterCode := &models.LetterCode{
		ID:         uuid.New().String(),
		LetterID:   letterID,
		Code:       code,
		QRCodeURL:  qrCodeURL,
		QRCodePath: qrCodePath,
	}

	tx := s.db.Begin()

	// 创建编号记录
	if err := tx.Create(letterCode).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to save letter code: %w", err)
	}

	// 更新信件状态
	if err := tx.Model(&letter).Update("status", models.StatusGenerated).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update letter status: %w", err)
	}

	// 记录状态变更
	statusLog := &models.StatusLog{
		ID:        uuid.New().String(),
		LetterID:  letterID,
		Status:    models.StatusGenerated,
		UpdatedBy: letter.UserID,
		Note:      "编号生成成功",
	}
	if err := tx.Create(statusLog).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create status log: %w", err)
	}

	tx.Commit()

	// 创建信使任务（异步，不影响主流程）
	go func() {
		if s.courierTaskSvc != nil {
			// 获取用户信息以确定取件地址
			var user models.User
			if err := s.db.First(&user, "id = ?", letter.UserID).Error; err == nil {
				// TODO: 创建信使任务
				// taskReq := &models.CreateCourierTaskRequest{
				// 	LetterID:       letterID,
				// 	PickupLocation: "用户宿舍", // TODO: 从用户信息获取实际地址
				// 	PickupCode:     "BJDX",     // TODO: 从用户信息获取实际编码
				// 	DeliveryHint:   "请投递到宿舍楼下信箱",
				// 	SenderContact:  user.Email,
				// }
				//
				// if _, err := s.courierTaskSvc.CreateTask(taskReq); err != nil {
				// 	// 记录错误日志，但不影响主流程
				// 	fmt.Printf("Failed to create courier task: %v\n", err)
				// }
			}
		}

		// 发送通知
		if s.notificationSvc != nil {
			s.notificationSvc.NotifyUser(letter.UserID, "letter_status_update", map[string]interface{}{
				"letter_id": letterID,
				"code":      code,
				"status":    "generated",
				"message":   "信件编号生成成功，请尽快打印贴纸并投递",
			})
		}

		// 奖励生成编号积分
		if s.creditSvc != nil {
			if err := s.creditSvc.RewardLetterGenerated(letter.UserID, letterID); err != nil {
				fmt.Printf("Failed to reward letter generated: %v\n", err)
			}
		}
	}()

	return letterCode, nil
}

// GetLetterByCode 通过编号获取信件
func (s *LetterService) GetLetterByCode(code string) (*models.LetterResponse, error) {
	var letterCode models.LetterCode
	if err := s.db.Preload("Letter").First(&letterCode, "code = ?", code).Error; err != nil {
		return nil, fmt.Errorf("letter not found: %w", err)
	}

	response := &models.LetterResponse{
		Letter:    &letterCode.Letter,
		QRCodeURL: letterCode.QRCodeURL,
		ReadURL:   fmt.Sprintf("%s/read/%s", s.config.FrontendURL, code),
	}

	return response, nil
}

// UpdateStatus 更新信件状态
func (s *LetterService) UpdateStatus(code string, req *models.UpdateLetterStatusRequest, updatedBy string) error {
	// 查找信件
	var letterCode models.LetterCode
	if err := s.db.Preload("Letter").First(&letterCode, "code = ?", code).Error; err != nil {
		return fmt.Errorf("letter not found: %w", err)
	}

	tx := s.db.Begin()

	// 更新信件状态
	if err := tx.Model(&letterCode.Letter).Update("status", req.Status).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update letter status: %w", err)
	}

	// 记录状态变更
	statusLog := &models.StatusLog{
		ID:        uuid.New().String(),
		LetterID:  letterCode.LetterID,
		Status:    req.Status,
		UpdatedBy: updatedBy,
		Location:  req.Location,
		Note:      req.Note,
	}
	if err := tx.Create(statusLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create status log: %w", err)
	}

	tx.Commit()

	// Send status update notifications
	if s.notificationSvc != nil {
		notificationData := map[string]interface{}{
			"letter_id": letterCode.LetterID,
			"code":      code,
			"status":    req.Status,
			"location":  req.Location,
		}

		// Send appropriate notification based on status
		var notificationType string
		switch req.Status {
		case models.StatusCollected:
			notificationType = "letter_collected"
		case models.StatusInTransit:
			notificationType = "letter_in_transit"
		case models.StatusDelivered:
			notificationType = "letter_delivered"
			// Reward sender for successful delivery
			if s.creditSvc != nil {
				go func() {
					if err := s.creditSvc.RewardLetterDelivered(letterCode.Letter.UserID, letterCode.LetterID); err != nil {
						fmt.Printf("Failed to reward letter delivered: %v\n", err)
					}
				}()
			}
			// Also notify and reward recipient when delivered
			if letterCode.Letter.ReplyTo != "" {
				s.notificationSvc.NotifyUser(letterCode.Letter.ReplyTo, "letter_received", map[string]interface{}{
					"letter_id":   letterCode.LetterID,
					"letter_code": code,
					"sender_id":   letterCode.Letter.UserID,
				})
				// Reward recipient for receiving letter
				if s.creditSvc != nil {
					go func() {
						if err := s.creditSvc.RewardReceiveLetter(letterCode.Letter.ReplyTo, letterCode.LetterID); err != nil {
							fmt.Printf("Failed to reward receive letter: %v\n", err)
						}
					}()
				}
			}
		}

		// Notify the sender about status update
		if notificationType != "" {
			if err := s.notificationSvc.NotifyUser(letterCode.Letter.UserID, notificationType, notificationData); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to send notification: %v\n", err)
			}
		}
	}

	// Send real-time WebSocket updates
	if s.wsService != nil {
		statusUpdateData := &websocket.LetterStatusUpdateData{
			LetterID:    letterCode.LetterID,
			Code:        code,
			Status:      string(req.Status),
			Location:    req.Location,
			CourierID:   updatedBy,
			CourierName: "", // Could be populated if we fetch courier info
			UpdatedAt:   time.Now(),
			Message:     req.Note,
		}

		// Broadcast to letter-specific room for anyone tracking this letter
		s.wsService.GetHub().BroadcastToRoom(websocket.GetLetterRoom(letterCode.LetterID),
			websocket.NewMessage(websocket.EventLetterStatusUpdate, map[string]interface{}{
				"letter_id":    statusUpdateData.LetterID,
				"code":         statusUpdateData.Code,
				"status":       statusUpdateData.Status,
				"location":     statusUpdateData.Location,
				"courier_id":   statusUpdateData.CourierID,
				"courier_name": statusUpdateData.CourierName,
				"updated_at":   statusUpdateData.UpdatedAt,
				"message":      statusUpdateData.Message,
			}))

		// Also broadcast to the sender's personal room
		s.wsService.GetHub().BroadcastToRoom(websocket.GetUserRoom(letterCode.Letter.UserID),
			websocket.NewMessage(websocket.EventLetterStatusUpdate, map[string]interface{}{
				"letter_id":  statusUpdateData.LetterID,
				"code":       statusUpdateData.Code,
				"status":     statusUpdateData.Status,
				"location":   statusUpdateData.Location,
				"updated_at": statusUpdateData.UpdatedAt,
				"message":    statusUpdateData.Message,
			}))
	}

	return nil
}

// GetUserLetters 获取用户信件列表
func (s *LetterService) GetUserLetters(userID string, params *models.LetterListParams) ([]models.Letter, int64, error) {
	query := s.db.Model(&models.Letter{}).Where("user_id = ?", userID)

	// 添加过滤条件
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Style != "" {
		query = query.Where("style = ?", params.Style)
	}
	if params.Search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count letters: %w", err)
	}

	// 添加排序和分页
	offset := (params.Page - 1) * params.Limit
	orderBy := fmt.Sprintf("%s %s", params.SortBy, params.SortOrder)

	var letters []models.Letter
	if err := query.Order(orderBy).Offset(offset).Limit(params.Limit).
		Preload("Code").Preload("StatusLogs").Find(&letters).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get letters: %w", err)
	}

	return letters, total, nil
}

// GetUserStats 获取用户统计
func (s *LetterService) GetUserStats(userID string) (*models.LetterStats, error) {
	stats := &models.LetterStats{}

	// 发送的信件总数
	if err := s.db.Model(&models.Letter{}).Where("user_id = ? AND status != ?",
		userID, models.StatusDraft).Count(&stats.TotalSent).Error; err != nil {
		return nil, fmt.Errorf("failed to count sent letters: %w", err)
	}

	// 收到的信件总数
	if err := s.db.Model(&models.Letter{}).Where("reply_to = ?",
		userID).Count(&stats.TotalReceived).Error; err != nil {
		return nil, fmt.Errorf("failed to count received letters: %w", err)
	}

	// 在途中的信件
	if err := s.db.Model(&models.Letter{}).Where("user_id = ? AND status IN ?",
		userID, []models.LetterStatus{models.StatusGenerated, models.StatusCollected, models.StatusInTransit}).
		Count(&stats.InTransit).Error; err != nil {
		return nil, fmt.Errorf("failed to count in transit letters: %w", err)
	}

	// 已送达的信件
	if err := s.db.Model(&models.Letter{}).Where("user_id = ? AND status IN ?",
		userID, []models.LetterStatus{models.StatusDelivered, models.StatusRead}).
		Count(&stats.Delivered).Error; err != nil {
		return nil, fmt.Errorf("failed to count delivered letters: %w", err)
	}

	// 草稿数量
	if err := s.db.Model(&models.Letter{}).Where("user_id = ? AND status = ?",
		userID, models.StatusDraft).Count(&stats.Drafts).Error; err != nil {
		return nil, fmt.Errorf("failed to count drafts: %w", err)
	}

	return stats, nil
}

// GetPublicLetters 获取广场公开信件
func (s *LetterService) GetPublicLetters(params *models.LetterListParams) ([]models.Letter, int64, error) {
	query := s.db.Model(&models.Letter{}).
		Where("status IN ?", []models.LetterStatus{models.StatusGenerated, models.StatusDelivered, models.StatusRead}).
		Where("title IS NOT NULL AND title != ''")

	// 添加分类过滤
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Style != "" {
		query = query.Where("style = ?", params.Style)
	}
	if params.Search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count public letters: %w", err)
	}

	// 添加排序和分页
	offset := (params.Page - 1) * params.Limit
	orderBy := fmt.Sprintf("%s %s", params.SortBy, params.SortOrder)

	var letters []models.Letter
	if err := query.Order(orderBy).Offset(offset).Limit(params.Limit).
		Preload("User").Find(&letters).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get public letters: %w", err)
	}

	return letters, total, nil
}

// MarkAsRead 标记信件为已读
func (s *LetterService) MarkAsRead(code string, userID string) error {
	// 查找信件
	var letterCode models.LetterCode
	if err := s.db.Preload("Letter").First(&letterCode, "code = ?", code).Error; err != nil {
		return fmt.Errorf("letter not found: %w", err)
	}

	// 只有状态为已送达的信件才能标记为已读
	if letterCode.Letter.Status != models.StatusDelivered {
		return fmt.Errorf("letter is not delivered yet")
	}

	tx := s.db.Begin()

	// 更新状态为已读
	if err := tx.Model(&letterCode.Letter).Update("status", models.StatusRead).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to mark as read: %w", err)
	}

	// 记录状态变更
	statusLog := &models.StatusLog{
		ID:        uuid.New().String(),
		LetterID:  letterCode.LetterID,
		Status:    models.StatusRead,
		UpdatedBy: userID,
		Note:      "收件人已查看",
	}
	if err := tx.Create(statusLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create status log: %w", err)
	}

	tx.Commit()

	// Send notification to sender that their letter has been read
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyUser(letterCode.Letter.UserID, "letter_read", map[string]interface{}{
			"letter_id": letterCode.LetterID,
			"code":      code,
			"read_by":   userID,
			"message":   "您的信件已被收件人阅读",
		})
	}

	// Reward sender for letter being read
	if s.creditSvc != nil {
		go func() {
			if err := s.creditSvc.RewardLetterRead(letterCode.Letter.UserID, letterCode.LetterID); err != nil {
				fmt.Printf("Failed to reward letter read: %v\n", err)
			}
		}()
	}

	return nil
}

// GetWritingInspiration 获取AI写作灵感
func (s *LetterService) GetWritingInspiration(userID, theme, style string, tags []string, count int) (*models.AIInspirationResponse, error) {
	if s.aiSvc == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	req := &models.AIInspirationRequest{
		Theme: theme,
		Style: style,
		Tags:  tags,
		Count: count,
	}

	ctx := context.Background()
	return s.aiSvc.GetInspiration(ctx, req)
}

// GetAIReplyAssistance 获取AI回信助手建议
func (s *LetterService) GetAIReplyAssistance(userID, letterID string, persona models.AIPersona) (*models.Letter, error) {
	if s.aiSvc == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	req := &models.AIReplyRequest{
		LetterID:   letterID,
		Persona:    persona,
		DelayHours: 0, // 立即生成
	}

	ctx := context.Background()
	return s.aiSvc.GenerateReply(ctx, req)
}

// MatchPenPalForLetter 为信件匹配笔友
func (s *LetterService) MatchPenPalForLetter(userID, letterID string) (*models.AIMatchResponse, error) {
	if s.aiSvc == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// 检查信件是否属于用户
	var letter models.Letter
	if err := s.db.Where("id = ? AND user_id = ?", letterID, userID).First(&letter).Error; err != nil {
		return nil, fmt.Errorf("letter not found: %w", err)
	}

	req := &models.AIMatchRequest{
		LetterID: letterID,
	}

	ctx := context.Background()
	return s.aiSvc.MatchPenPal(ctx, req)
}

// AutoCurateLetterForMuseum 自动策展信件到博物馆
func (s *LetterService) AutoCurateLetterForMuseum(letterID, exhibitionID string) error {
	if s.aiSvc == nil {
		return fmt.Errorf("AI service not available")
	}

	req := &models.AICurateRequest{
		LetterIDs:    []string{letterID},
		ExhibitionID: exhibitionID,
		AutoApprove:  false, // 需要人工审核
	}

	ctx := context.Background()
	return s.aiSvc.CurateLetters(ctx, req)
}

// ========================= Reply/Thread System =========================

// GetReplyInfoByCode 通过扫码获取回信信息
func (s *LetterService) GetReplyInfoByCode(code string) (*models.LetterResponse, error) {
	// 查找原始信件
	var letterCode models.LetterCode
	if err := s.db.Preload("Letter").Where("code = ?", code).First(&letterCode).Error; err != nil {
		return nil, fmt.Errorf("信件不存在或二维码无效: %w", err)
	}

	// 检查信件是否可以回信
	if letterCode.Letter.Status != models.StatusDelivered && letterCode.Letter.Status != models.StatusRead {
		return nil, fmt.Errorf("只有已送达或已读的信件才能回信")
	}

	// 构建回信信息响应
	response := &models.LetterResponse{
		Letter:    &letterCode.Letter,
		QRCodeURL: letterCode.QRCodeURL,
		ReadURL:   fmt.Sprintf("/letters/read/%s", code),
	}

	return response, nil
}

// CreateReply 创建回信
func (s *LetterService) CreateReply(userID string, req *models.CreateReplyRequest) (*models.LetterReplyResponse, error) {
	// 查找原始信件
	var letterCode models.LetterCode
	if err := s.db.Preload("Letter").Where("code = ?", req.OriginalLetterCode).First(&letterCode).Error; err != nil {
		return nil, fmt.Errorf("原始信件不存在: %w", err)
	}

	originalLetterID := letterCode.LetterID

	// 检查是否存在线程，如果不存在则创建
	var thread models.LetterThread
	err := s.db.Where("original_letter = ?", originalLetterID).First(&thread).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新线程
			thread = models.LetterThread{
				ID:             uuid.New().String(),
				OriginalLetter: originalLetterID,
				Participants:   fmt.Sprintf("[\"%s\", \"%s\"]", letterCode.Letter.UserID, userID),
				ThreadTitle:    fmt.Sprintf("回复: %s", letterCode.Letter.Title),
				LastReplyAt:    time.Now(),
				ReplyCount:     0,
				IsActive:       true,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err := s.db.Create(&thread).Error; err != nil {
				return nil, fmt.Errorf("创建线程失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("查询线程失败: %w", err)
		}
	}

	// 生成回信编号和二维码
	replyCode := fmt.Sprintf("REPLY-%d", time.Now().UnixNano())
	qrCodePath, err := s.generateQRCode(replyCode)
	if err != nil {
		return nil, fmt.Errorf("生成二维码失败: %w", err)
	}

	// 创建回信记录
	reply := models.LetterReply{
		ID:            uuid.New().String(),
		ThreadID:      thread.ID,
		ReplyToLetter: originalLetterID,
		AuthorID:      userID,
		Content:       req.Content,
		Style:         req.Style,
		Status:        models.StatusGenerated,
		IsPublic:      req.IsPublic,
		DeliveryCode:  replyCode,
		QRCodePath:    qrCodePath,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 开始事务
	tx := s.db.Begin()
	if err := tx.Create(&reply).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建回信失败: %w", err)
	}

	// 更新线程信息
	if err := tx.Model(&thread).Updates(map[string]interface{}{
		"last_reply_at": time.Now(),
		"reply_count":   gorm.Expr("reply_count + 1"),
		"updated_at":    time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新线程失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("保存回信失败: %w", err)
	}

	// 发送通知给原始信件作者
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyUser(letterCode.Letter.UserID, "letter_reply_received", map[string]interface{}{
			"reply_id":       reply.ID,
			"author_id":      userID,
			"original_title": letterCode.Letter.Title,
		})
	}

	// 积分奖励
	if s.creditSvc != nil {
		s.creditSvc.RewardReply(userID, reply.ID)
	}

	// 创建信使任务
	if s.courierTaskSvc != nil {
		go func() {
			s.courierTaskSvc.CreateDeliveryTask(reply.ID, reply.DeliveryCode)
		}()
	}

	return &models.LetterReplyResponse{
		ID:            reply.ID,
		ThreadID:      reply.ThreadID,
		ReplyToLetter: reply.ReplyToLetter,
		AuthorID:      reply.AuthorID,
		Content:       reply.Content,
		Style:         reply.Style,
		Status:        reply.Status,
		IsPublic:      reply.IsPublic,
		DeliveryCode:  reply.DeliveryCode,
		QRCodeURL:     fmt.Sprintf("/uploads/qr/%s", filepath.Base(reply.QRCodePath)),
		CreatedAt:     reply.CreatedAt,
		UpdatedAt:     reply.UpdatedAt,
	}, nil
}

// GetUserThreads 获取用户的对话线程列表
func (s *LetterService) GetUserThreads(userID string, page, limit int) ([]models.ThreadResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// 查找用户参与的线程
	var threads []models.LetterThread
	err := s.db.Where("participants LIKE ?", "%\""+userID+"\"%").
		Order("last_reply_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads).Error
	if err != nil {
		return nil, fmt.Errorf("查询线程失败: %w", err)
	}

	var responses []models.ThreadResponse
	for _, thread := range threads {
		// 获取原始信件信息
		var letter models.Letter
		if err := s.db.First(&letter, thread.OriginalLetter).Error; err != nil {
			continue // 跳过无效的线程
		}

		// 获取线程中的回信列表
		var replies []models.LetterReply
		s.db.Where("thread_id = ?", thread.ID).
			Order("created_at ASC").
			Find(&replies)

		// 转换回信格式
		var replyResponses []models.LetterReplyResponse
		for _, reply := range replies {
			replyResponses = append(replyResponses, models.LetterReplyResponse{
				ID:            reply.ID,
				ThreadID:      reply.ThreadID,
				ReplyToLetter: reply.ReplyToLetter,
				AuthorID:      reply.AuthorID,
				Content:       reply.Content,
				Style:         reply.Style,
				Status:        reply.Status,
				IsPublic:      reply.IsPublic,
				DeliveryCode:  reply.DeliveryCode,
				QRCodeURL:     fmt.Sprintf("/uploads/qr/%s", filepath.Base(reply.QRCodePath)),
				ReadAt:        reply.ReadAt,
				CreatedAt:     reply.CreatedAt,
				UpdatedAt:     reply.UpdatedAt,
			})
		}

		// 解析参与者列表
		var participants []string
		if err := json.Unmarshal([]byte(thread.Participants), &participants); err != nil {
			participants = []string{} // 解析失败时使用空列表
		}

		responses = append(responses, models.ThreadResponse{
			ID: thread.ID,
			OriginalLetter: models.LetterResponse{
				Letter:  &letter,
				ReadURL: fmt.Sprintf("/letters/read/%s", thread.OriginalLetter),
			},
			Participants: participants,
			ThreadTitle:  thread.ThreadTitle,
			LastReplyAt:  thread.LastReplyAt,
			ReplyCount:   thread.ReplyCount,
			IsActive:     thread.IsActive,
			Replies:      replyResponses,
			CreatedAt:    thread.CreatedAt,
			UpdatedAt:    thread.UpdatedAt,
		})
	}

	return responses, nil
}

// GetThreadByID 获取指定线程的详细信息
func (s *LetterService) GetThreadByID(userID, threadID string) (*models.ThreadResponse, error) {
	// 获取线程信息
	var thread models.LetterThread
	if err := s.db.First(&thread, threadID).Error; err != nil {
		return nil, fmt.Errorf("线程不存在: %w", err)
	}

	// 检查用户是否有权限访问此线程
	var participants []string
	if err := json.Unmarshal([]byte(thread.Participants), &participants); err != nil {
		return nil, fmt.Errorf("线程数据错误")
	}

	hasPermission := false
	for _, participant := range participants {
		if participant == userID {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, fmt.Errorf("无权限访问此线程")
	}

	// 获取原始信件信息
	var letter models.Letter
	if err := s.db.First(&letter, thread.OriginalLetter).Error; err != nil {
		return nil, fmt.Errorf("原始信件不存在: %w", err)
	}

	// 获取所有回信
	var replies []models.LetterReply
	if err := s.db.Where("thread_id = ?", thread.ID).
		Order("created_at ASC").
		Find(&replies).Error; err != nil {
		return nil, fmt.Errorf("获取回信失败: %w", err)
	}

	// 转换回信格式
	var replyResponses []models.LetterReplyResponse
	for _, reply := range replies {
		replyResponses = append(replyResponses, models.LetterReplyResponse{
			ID:            reply.ID,
			ThreadID:      reply.ThreadID,
			ReplyToLetter: reply.ReplyToLetter,
			AuthorID:      reply.AuthorID,
			Content:       reply.Content,
			Style:         reply.Style,
			Status:        reply.Status,
			IsPublic:      reply.IsPublic,
			DeliveryCode:  reply.DeliveryCode,
			QRCodeURL:     fmt.Sprintf("/uploads/qr/%s", filepath.Base(reply.QRCodePath)),
			ReadAt:        reply.ReadAt,
			CreatedAt:     reply.CreatedAt,
			UpdatedAt:     reply.UpdatedAt,
		})
	}

	return &models.ThreadResponse{
		ID: thread.ID,
		OriginalLetter: models.LetterResponse{
			Letter:  &letter,
			ReadURL: fmt.Sprintf("/letters/read/%s", thread.OriginalLetter),
		},
		Participants: participants,
		ThreadTitle:  thread.ThreadTitle,
		LastReplyAt:  thread.LastReplyAt,
		ReplyCount:   thread.ReplyCount,
		IsActive:     thread.IsActive,
		Replies:      replyResponses,
		CreatedAt:    thread.CreatedAt,
		UpdatedAt:    thread.UpdatedAt,
	}, nil
}

// generateDeliveryCode 生成投递编码
func (s *LetterService) generateDeliveryCode() string {
	return fmt.Sprintf("LETTER-%d", time.Now().UnixNano())
}

// generateQRCode 生成二维码
func (s *LetterService) generateQRCode(code string) (string, error) {
	// 确保上传目录存在
	qrDir := filepath.Join("uploads", "qr")
	if err := utils.EnsureDir(qrDir); err != nil {
		return "", fmt.Errorf("创建二维码目录失败: %w", err)
	}

	// 生成二维码文件名
	qrFileName := fmt.Sprintf("%s.png", code)
	qrFilePath := filepath.Join(qrDir, qrFileName)

	// 生成读取链接
	readURL := fmt.Sprintf("%s/letters/read/%s", s.config.FrontendURL, code)

	// 生成二维码
	err := qrcode.WriteFile(readURL, qrcode.Medium, 256, qrFilePath)
	if err != nil {
		return "", fmt.Errorf("生成二维码失败: %w", err)
	}

	return qrFilePath, nil
}

// generateEnhancedQRData 生成包含OP Code信息的增强QR码数据
func (s *LetterService) generateEnhancedQRData(letterID, code, recipientOPCode, senderOPCode string) string {
	// 创建结构化QR码数据
	qrData := map[string]interface{}{
		"type":             "openpenpal_letter",
		"version":          "1.0",
		"letter_id":        letterID,
		"code":             code,
		"read_url":         fmt.Sprintf("%s/read/%s", s.config.FrontendURL, code),
		"recipient_opcode": recipientOPCode,
		"sender_opcode":    senderOPCode,
		"scan_timestamp":   time.Now().Unix(),
		"app_info": map[string]string{
			"name":    "OpenPenPal",
			"version": "1.0",
		},
	}

	// 将结构化数据序列化为JSON
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		// 如果JSON序列化失败，回退到简单URL
		return fmt.Sprintf("%s/read/%s", s.config.FrontendURL, code)
	}

	return string(jsonData)
}

// QRCodeData QR码数据结构
type QRCodeData struct {
	Type             string            `json:"type"`
	Version          string            `json:"version"`
	LetterID         string            `json:"letter_id"`
	Code             string            `json:"code"`
	ReadURL          string            `json:"read_url"`
	RecipientOPCode  string            `json:"recipient_opcode,omitempty"`
	SenderOPCode     string            `json:"sender_opcode,omitempty"`
	ScanTimestamp    int64             `json:"scan_timestamp"`
	CourierInfo      *CourierQRInfo    `json:"courier_info,omitempty"`
	AppInfo          map[string]string `json:"app_info"`
}

// CourierQRInfo 信使相关的QR码信息
type CourierQRInfo struct {
	RequiredPermission string `json:"required_permission,omitempty"` // 需要的OP Code权限前缀
	TaskID             string `json:"task_id,omitempty"`             // 关联的任务ID
	DeliveryInst       string `json:"delivery_instructions,omitempty"` // 配送说明
}

// GetUserDrafts 获取用户草稿
func (s *LetterService) GetUserDrafts(ctx context.Context, userID string, page, limit int, sortBy, sortOrder string) ([]models.Letter, int64, error) {
	var letters []models.Letter
	var total int64

	query := s.db.Model(&models.Letter{}).Where("author_id = ? AND status = ?", userID, "draft")

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	if sortBy != "created_at" && sortBy != "updated_at" && sortBy != "title" {
		orderClause = "updated_at desc"
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Order(orderClause).Offset(offset).Limit(limit).Find(&letters).Error; err != nil {
		return nil, 0, err
	}

	return letters, total, nil
}

// PublishLetter 发布信件
func (s *LetterService) PublishLetter(ctx context.Context, letterID, userID string, scheduledAt *time.Time, visibility string) (*models.Letter, error) {
	var letter models.Letter
	
	// 查找信件
	if err := s.db.Where("id = ? AND author_id = ?", letterID, userID).First(&letter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("letter not found")
		}
		return nil, err
	}

	// 检查权限
	if letter.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// 更新状态
	updates := map[string]interface{}{
		"status": "published",
		"sent_at": time.Now(),
		"updated_at": time.Now(),
	}

	if visibility != "" {
		updates["visibility"] = visibility
	}

	if scheduledAt != nil && scheduledAt.After(time.Now()) {
		updates["status"] = "scheduled"
		updates["scheduled_at"] = scheduledAt
	}

	if err := s.db.Model(&letter).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 如果立即发布，增加积分
	if letter.Status == "published" && s.creditSvc != nil {
		s.creditSvc.AddPoints(userID, 10, "letter_published", fmt.Sprintf("发布信件《%s》", letter.Title))
	}

	return &letter, nil
}

// LikeLetter 点赞信件
func (s *LetterService) LikeLetter(ctx context.Context, letterID, userID string) error {
	var letter models.Letter
	
	// 查找信件
	if err := s.db.Where("id = ?", letterID).First(&letter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("letter not found")
		}
		return err
	}

	// 更新点赞计数
	if err := s.db.Model(&letter).Update("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
		return err
	}

	// 记录点赞历史
	if userID != "" {
		like := &models.LetterLike{
			ID:        fmt.Sprintf("like_%s_%s", userID, time.Now().Format("20060102150405")),
			LetterID:  letterID,
			UserID:    userID,
			CreatedAt: time.Now(),
		}
		s.db.Create(&like) // 忽略错误，避免重复点赞
	}

	// 发送通知
	if s.notificationSvc != nil && letter.UserID != userID && userID != "" {
		s.notificationSvc.NotifyUser(letter.UserID, "letter_liked", map[string]interface{}{
			"letter_id": letterID,
			"title": letter.Title,
		})
	}

	return nil
}

// ShareLetter 分享信件
func (s *LetterService) ShareLetter(ctx context.Context, letterID, userID, platform string) (map[string]interface{}, error) {
	var letter models.Letter
	
	// 查找信件
	if err := s.db.Where("id = ?", letterID).First(&letter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("letter not found")
		}
		return nil, err
	}

	// 检查信件状态是否允许分享
	if letter.Status != "published" {
		return nil, errors.New("only published letters can be shared")
	}

	// 记录分享（share_count通过LetterShare表统计）

	// 生成分享链接
	shareURL := fmt.Sprintf("https://openpenpal.com/letters/read/%s", letterID)
	shareData := map[string]interface{}{
		"url":      shareURL,
		"title":    letter.Title,
		"platform": platform,
		"code":     letter.Code,
	}

	// 记录分享历史
	if userID != "" {
		share := &models.LetterShare{
			ID:        fmt.Sprintf("share_%s_%s", userID, time.Now().Format("20060102150405")),
			LetterID:  letterID,
			UserID:    userID,
			Platform:  platform,
			CreatedAt: time.Now(),
		}
		s.db.Create(&share)
	}

	return shareData, nil
}

// GetLetterTemplates 获取信件模板
func (s *LetterService) GetLetterTemplates(ctx context.Context, category string, page, limit int) ([]models.LetterTemplate, int64, error) {
	var templates []models.LetterTemplate
	var total int64

	query := s.db.Model(&models.LetterTemplate{}).Where("is_active = ?", true)
	
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Order("usage_count DESC, rating DESC").Offset(offset).Limit(limit).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// GetTemplateByID 根据ID获取模板
func (s *LetterService) GetTemplateByID(ctx context.Context, templateID string) (*models.LetterTemplate, error) {
	var template models.LetterTemplate
	if err := s.db.Where("id = ?", templateID).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}

	// 增加使用次数
	s.db.Model(&template).Update("usage_count", gorm.Expr("usage_count + ?", 1))

	return &template, nil
}

// SearchLetters 搜索信件
func (s *LetterService) SearchLetters(ctx context.Context, userID, query string, tags []string, dateFrom, dateTo, visibility string, page, limit int) ([]models.Letter, int64, error) {
	var letters []models.Letter
	var total int64

	// 构建查询
	dbQuery := s.db.Model(&models.Letter{})

	// 可见性过滤
	if visibility != "" {
		dbQuery = dbQuery.Where("visibility = ?", visibility)
	} else {
		// 默认只搜索公开和学校内的信件
		if userID != "" {
			dbQuery = dbQuery.Where("visibility IN (?) OR author_id = ?", []string{"public", "school"}, userID)
		} else {
			dbQuery = dbQuery.Where("visibility = ?", "public")
		}
	}

	// 关键词搜索
	if query != "" {
		searchPattern := "%" + query + "%"
		dbQuery = dbQuery.Where("(title LIKE ? OR content LIKE ?)", searchPattern, searchPattern)
	}

	// 标签过滤
	if len(tags) > 0 {
		// 这里简化处理，实际应该使用JSON查询
		for _, tag := range tags {
			dbQuery = dbQuery.Where("tags LIKE ?", "%"+tag+"%")
		}
	}

	// 日期范围过滤
	if dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			dbQuery = dbQuery.Where("created_at >= ?", t)
		}
	}
	if dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			dbQuery = dbQuery.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	// 只搜索已发布的信件
	dbQuery = dbQuery.Where("status = ?", "published")

	// 计算总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := dbQuery.Order("created_at DESC").Offset(offset).Limit(limit).Find(&letters).Error; err != nil {
		return nil, 0, err
	}

	return letters, total, nil
}

// GetPopularLetters 获取热门信件
func (s *LetterService) GetPopularLetters(ctx context.Context, timeRange string, page, limit int) ([]models.Letter, int64, error) {
	var letters []models.Letter
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

	query := s.db.Model(&models.Letter{}).
		Where("status = ? AND visibility IN (?)", "published", []string{"public", "school"})

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}

	// 按照热度排序（浏览量 + 点赞量*2 + 分享量*3）
	query = query.Select("*, (view_count + like_count*2 + share_count*3) as score").
		Order("score DESC")

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&letters).Error; err != nil {
		return nil, 0, err
	}

	return letters, total, nil
}

// GetRecommendedLetters 获取推荐信件
func (s *LetterService) GetRecommendedLetters(ctx context.Context, userID string, page, limit int) ([]models.Letter, int64, error) {
	var letters []models.Letter
	var total int64

	// 基础查询：公开的已发布信件
	query := s.db.Model(&models.Letter{}).
		Where("status = ? AND visibility IN (?)", "published", []string{"public", "school"})

	// 如果有用户ID，排除用户自己的信件
	if userID != "" {
		query = query.Where("author_id != ?", userID)
		
		// TODO: 基于用户兴趣推荐
		// 这里可以根据用户的阅读历史、点赞记录等推荐相似内容
	}

	// 推荐最近一周的优质内容
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	query = query.Where("created_at >= ?", oneWeekAgo).
		Where("(like_count > ? OR view_count > ?)", 5, 20)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，随机排序增加多样性
	offset := (page - 1) * limit
	if err := query.Order("RANDOM()").Offset(offset).Limit(limit).Find(&letters).Error; err != nil {
		return nil, 0, err
	}

	return letters, total, nil
}

// BatchOperate 批量操作
func (s *LetterService) BatchOperate(ctx context.Context, userID string, letterIDs []string, operation string, data map[string]interface{}) (map[string]interface{}, error) {
	results := map[string]interface{}{
		"success": 0,
		"failed":  0,
		"errors":  []string{},
	}

	for _, letterID := range letterIDs {
		var err error
		
		switch operation {
		case "delete":
			err = s.deleteLetter(ctx, letterID, userID)
		case "archive":
			err = s.archiveLetter(ctx, letterID, userID)
		case "publish":
			_, err = s.PublishLetter(ctx, letterID, userID, nil, "")
		default:
			err = errors.New("unsupported operation")
		}

		if err != nil {
			results["failed"] = results["failed"].(int) + 1
			errors := results["errors"].([]string)
			results["errors"] = append(errors, fmt.Sprintf("%s: %s", letterID, err.Error()))
		} else {
			results["success"] = results["success"].(int) + 1
		}
	}

	return results, nil
}

// deleteLetter 删除信件（内部方法）
func (s *LetterService) deleteLetter(ctx context.Context, letterID, userID string) error {
	result := s.db.Where("id = ? AND author_id = ?", letterID, userID).Delete(&models.Letter{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("letter not found or unauthorized")
	}
	return nil
}

// archiveLetter 归档信件（内部方法）
func (s *LetterService) archiveLetter(ctx context.Context, letterID, userID string) error {
	result := s.db.Model(&models.Letter{}).
		Where("id = ? AND author_id = ?", letterID, userID).
		Update("status", "archived")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("letter not found or unauthorized")
	}
	return nil
}

// ExportLetters 导出信件
func (s *LetterService) ExportLetters(ctx context.Context, userID string, letterIDs []string, format string, includeAttachments bool) (map[string]interface{}, error) {
	var letters []models.Letter

	// 查询要导出的信件
	query := s.db.Where("author_id = ?", userID)
	if len(letterIDs) > 0 {
		query = query.Where("id IN ?", letterIDs)
	}
	
	if err := query.Find(&letters).Error; err != nil {
		return nil, err
	}

	// 根据格式生成导出数据
	exportData := map[string]interface{}{
		"format":      format,
		"letter_count": len(letters),
		"export_time": time.Now().Format(time.RFC3339),
	}

	switch format {
	case "json":
		exportData["content"] = letters
	case "pdf":
		// TODO: 生成PDF文件
		exportData["file_url"] = fmt.Sprintf("/exports/%s_%s.pdf", userID, time.Now().Format("20060102150405"))
	case "txt":
		// 生成纯文本内容
		var textContent strings.Builder
		for _, letter := range letters {
			textContent.WriteString(fmt.Sprintf("标题：%s\n", letter.Title))
			textContent.WriteString(fmt.Sprintf("日期：%s\n", letter.CreatedAt.Format("2006-01-02")))
			textContent.WriteString(fmt.Sprintf("内容：\n%s\n", letter.Content))
			textContent.WriteString("\n---\n\n")
		}
		exportData["content"] = textContent.String()
	}

	return exportData, nil
}

// AutoSaveDraft 自动保存草稿
func (s *LetterService) AutoSaveDraft(ctx context.Context, letter *models.Letter) (*models.Letter, error) {
	letter.Status = "draft"
	letter.UpdatedAt = time.Now()

	// 如果ID存在，更新；否则创建
	if letter.ID != "" {
		// 检查是否是用户的草稿
		var existing models.Letter
		if err := s.db.Where("id = ? AND user_id = ? AND status = ?", letter.ID, letter.UserID, "draft").First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("draft not found or unauthorized")
			}
			return nil, err
		}

		// 更新草稿
		if err := s.db.Model(&existing).Updates(letter).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	} else {
		// 创建新草稿
		letter.ID = fmt.Sprintf("letter_%s", time.Now().Format("20060102150405"))
		letter.CreatedAt = time.Now()
		
		if err := s.db.Create(letter).Error; err != nil {
			return nil, err
		}
		return letter, nil
	}
}

// GetWritingSuggestions 获取写作建议
func (s *LetterService) GetWritingSuggestions(ctx context.Context, content, style string, tags []string, mood, purpose string) (map[string]interface{}, error) {
	suggestions := map[string]interface{}{
		"general_tips": []string{},
		"style_tips":   []string{},
		"improvements": []string{},
		"examples":     []string{},
	}

	// 基本建议
	generalTips := []string{
		"开头可以用温暖的问候语",
		"结尾加上祝福会让信件更有温度",
		"适当分段让信件更易读",
	}

	// 根据风格给出建议
	styleTips := map[string][]string{
		"formal": {
			"使用敬语和正式称呼",
			"保持措辞严谨",
			"避免使用网络用语",
		},
		"casual": {
			"可以使用轻松的语气",
			"适当加入幽默元素",
			"像和朋友聊天一样自然",
		},
		"emotional": {
			"真诚表达内心感受",
			"使用具体的例子和细节",
			"不要害怕展示脆弱",
		},
	}

	// 根据心情给出建议
	moodTips := map[string][]string{
		"happy": {
			"分享让你快乐的具体事情",
			"传递积极正能量",
		},
		"sad": {
			"诚实表达你的感受",
			"寻求理解和支持是很正常的",
		},
		"grateful": {
			"具体说明你感激的原因",
			"表达感谢会让收信人感到被重视",
		},
	}

	suggestions["general_tips"] = generalTips
	
	if tips, ok := styleTips[style]; ok {
		suggestions["style_tips"] = tips
	}
	
	if tips, ok := moodTips[mood]; ok {
		moodSuggestions := suggestions["style_tips"].([]string)
		suggestions["style_tips"] = append(moodSuggestions, tips...)
	}

	// 如果集成了AI服务，可以提供更智能的建议
	if s.aiSvc != nil {
		// TODO: 调用AI服务获取个性化建议
	}

	return suggestions, nil
}

// ========================= FSD条码系统增强方法 =========================

// BindBarcodeToEnvelope 绑定条码到信封 - FSD 6.2规格
func (s *LetterService) BindBarcodeToEnvelope(req *models.BindBarcodeRequest, operatorID string) (*models.EnvelopeWithBarcodeResponse, error) {
	// 验证OP Code格式
	if s.opcodeService != nil {
		if valid, err := s.opcodeService.ValidateOPCode(req.RecipientCode); !valid {
			return nil, fmt.Errorf("invalid OP Code: %w", err)
		}
	}

	// 查找信件
	var letter models.Letter
	if err := s.db.Where("id = ?", req.LetterID).First(&letter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("letter not found")
		}
		return nil, fmt.Errorf("failed to find letter: %w", err)
	}

	// 查找或创建LetterCode
	var letterCode models.LetterCode
	err := s.db.Where("letter_id = ?", req.LetterID).First(&letterCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果不存在LetterCode，先生成
			generatedCode, genErr := s.GenerateCode(req.LetterID)
			if genErr != nil {
				return nil, fmt.Errorf("failed to generate letter code: %w", genErr)
			}
			letterCode = *generatedCode
		} else {
			return nil, fmt.Errorf("failed to find letter code: %w", err)
		}
	}

	// 检查条码是否可以绑定
	if !letterCode.CanBeBound() {
		return nil, fmt.Errorf("barcode cannot be bound, current status: %s", letterCode.GetStatusDisplayName())
	}

	// 查找或创建信封（如果提供了EnvelopeID）
	var envelope *models.Envelope
	if req.EnvelopeID != "" {
		envelope = &models.Envelope{}
		if err := s.db.Where("id = ?", req.EnvelopeID).First(envelope).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("envelope not found")
			}
			return nil, fmt.Errorf("failed to find envelope: %w", err)
		}
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新LetterCode状态和关联信息
	now := time.Now()
	updates := map[string]interface{}{
		"status":         models.BarcodeStatusBound,
		"recipient_code": req.RecipientCode,
		"bound_at":       &now,
		"last_scanned_by": operatorID,
		"last_scanned_at": &now,
		"scan_count":     gorm.Expr("scan_count + 1"),
		"updated_at":     now,
	}

	if req.EnvelopeID != "" {
		updates["envelope_id"] = req.EnvelopeID
	}

	if err := tx.Model(&letterCode).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update letter code: %w", err)
	}

	// 更新信件的OP Code信息
	if err := tx.Model(&letter).Updates(map[string]interface{}{
		"recipient_op_code": req.RecipientCode,
		"updated_at":        now,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update letter OP code: %w", err)
	}

	// 如果有信封，更新信封的OP Code信息
	if envelope != nil {
		if err := tx.Model(envelope).Updates(map[string]interface{}{
			"recipient_op_code": req.RecipientCode,
			"barcode_id":        letterCode.Code,
			"letter_id":         req.LetterID,
			"updated_at":        now,
		}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update envelope: %w", err)
		}
	}

	// 记录状态变更日志
	statusLog := &models.StatusLog{
		ID:        uuid.New().String(),
		LetterID:  req.LetterID,
		Status:    models.StatusGenerated, // 保持信件状态
		UpdatedBy: operatorID,
		Note:      fmt.Sprintf("条码已绑定到OP Code: %s", req.RecipientCode),
		CreatedAt: now,
	}
	if err := tx.Create(statusLog).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create status log: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 构建响应
	response := &models.EnvelopeWithBarcodeResponse{
		EnvelopeID:      req.EnvelopeID,
		BarcodeID:       letterCode.ID,
		BarcodeCode:     letterCode.Code,
		RecipientOPCode: req.RecipientCode,
		Status:          string(models.BarcodeStatusBound),
		QRURL:           letterCode.QRCodeURL,
	}

	if envelope != nil {
		response.DesignID = envelope.DesignID
	}

	// 发送通知
	if s.notificationSvc != nil {
		go func() {
			s.notificationSvc.NotifyUser(letter.UserID, "barcode_bound", map[string]interface{}{
				"letter_id":        req.LetterID,
				"barcode_code":     letterCode.Code,
				"recipient_opcode": req.RecipientCode,
				"message":          "条码已成功绑定",
			})
		}()
	}

	return response, nil
}

// UpdateBarcodeStatus 更新条码物流状态 - FSD 6.3规格
func (s *LetterService) UpdateBarcodeStatus(barcodeCode string, req *models.UpdateBarcodeStatusRequest) error {
	// 查找LetterCode
	var letterCode models.LetterCode
	if err := s.db.Preload("Letter").Where("code = ?", barcodeCode).First(&letterCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("barcode not found")
		}
		return fmt.Errorf("failed to find barcode: %w", err)
	}

	// 将请求状态映射到BarcodeStatus
	var newStatus models.BarcodeStatus
	switch req.Status {
	case "picked":
		newStatus = models.BarcodeStatusBound // 已取件
	case "in_transit":
		newStatus = models.BarcodeStatusInTransit
	case "delivered":
		newStatus = models.BarcodeStatusDelivered
	case "failed":
		newStatus = models.BarcodeStatusCancelled
	default:
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	// 检查状态转换是否有效
	if !letterCode.IsValidTransition(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", 
			letterCode.GetStatusDisplayName(), string(newStatus))
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新LetterCode状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":          newStatus,
		"last_scanned_by": req.OperatorID,
		"last_scanned_at": &now,
		"scan_count":      gorm.Expr("scan_count + 1"),
		"updated_at":      now,
	}

	// 如果是已送达状态，设置送达时间
	if newStatus == models.BarcodeStatusDelivered {
		updates["delivered_at"] = &now
	}

	if err := tx.Model(&letterCode).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update barcode status: %w", err)
	}

	// 同时更新信件状态
	var letterStatus models.LetterStatus
	switch newStatus {
	case models.BarcodeStatusBound:
		letterStatus = models.StatusCollected
	case models.BarcodeStatusInTransit:
		letterStatus = models.StatusInTransit
	case models.BarcodeStatusDelivered:
		letterStatus = models.StatusDelivered
	case models.BarcodeStatusCancelled:
		letterStatus = models.StatusDraft // 回退到草稿状态
	default:
		letterStatus = letterCode.Letter.Status // 保持原状态
	}

	if err := tx.Model(&letterCode.Letter).Update("status", letterStatus).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update letter status: %w", err)
	}

	// 记录状态变更日志
	statusLog := &models.StatusLog{
		ID:        uuid.New().String(),
		LetterID:  letterCode.LetterID,
		Status:    letterStatus,
		UpdatedBy: req.OperatorID,
		Location:  req.Location,
		Note:      fmt.Sprintf("条码状态更新: %s → %s. %s", letterCode.Status, newStatus, req.Notes),
		CreatedAt: now,
	}
	if err := tx.Create(statusLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create status log: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 发送实时通知
	if s.notificationSvc != nil {
		go func() {
			message := s.getStatusUpdateMessage(string(newStatus))
			s.notificationSvc.NotifyUser(letterCode.Letter.UserID, "barcode_status_updated", map[string]interface{}{
				"letter_id":     letterCode.LetterID,
				"barcode_code":  barcodeCode,
				"status":        string(newStatus),
				"location":      req.Location,
				"operator_id":   req.OperatorID,
				"op_code":       req.OPCode,
				"message":       message,
				"notes":         req.Notes,
			})
		}()
	}

	// 发送WebSocket实时更新
	if s.wsService != nil {
		go func() {
			s.wsService.GetHub().BroadcastToRoom(websocket.GetLetterRoom(letterCode.LetterID),
				websocket.NewMessage("barcode_status_update", map[string]interface{}{
					"letter_id":    letterCode.LetterID,
					"barcode_code": barcodeCode,
					"status":       string(newStatus),
					"location":     req.Location,
					"operator_id":  req.OperatorID,
					"op_code":      req.OPCode,
					"updated_at":   now,
				}))
		}()
	}

	return nil
}

// GetBarcodeStatus 获取条码状态信息 - FSD查询接口
func (s *LetterService) GetBarcodeStatus(barcodeCode string) (*models.LetterCode, error) {
	var letterCode models.LetterCode
	if err := s.db.Preload("Letter").Preload("Envelope").Where("code = ?", barcodeCode).First(&letterCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("barcode not found")
		}
		return nil, fmt.Errorf("failed to find barcode: %w", err)
	}

	return &letterCode, nil
}

// ValidateBarcodeOperation 验证条码操作权限 - FSD权限校验
func (s *LetterService) ValidateBarcodeOperation(barcodeCode, operatorID, requiredOPCode string) error {
	// 查找条码
	var letterCode models.LetterCode
	if err := s.db.Where("code = ?", barcodeCode).First(&letterCode).Error; err != nil {
		return fmt.Errorf("barcode not found: %w", err)
	}

	// 验证操作员权限（如果提供了OP Code服务）
	if s.opcodeService != nil && requiredOPCode != "" {
		if hasPermission, err := s.opcodeService.CheckPermission(operatorID, requiredOPCode); !hasPermission {
			return fmt.Errorf("insufficient permissions: %w", err)
		}
	}

	// 验证条码是否处于可操作状态
	if !letterCode.IsActive() {
		return fmt.Errorf("barcode is not active, current status: %s", letterCode.GetStatusDisplayName())
	}

	return nil
}

// getStatusUpdateMessage 获取状态更新消息
func (s *LetterService) getStatusUpdateMessage(status string) string {
	messages := map[string]string{
		"bound":      "条码已绑定成功",
		"in_transit": "信件正在投递中",
		"delivered":  "信件已成功送达",
		"cancelled":  "投递已取消",
		"expired":    "条码已过期",
	}

	if message, exists := messages[status]; exists {
		return message
	}
	return "状态已更新"
}

