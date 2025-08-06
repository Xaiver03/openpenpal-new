package services

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"openpenpal-backend/internal/models"
)

// QRScanService - SOTA pattern: Single Responsibility + Strategy Pattern
type QRScanService struct {
	db             *gorm.DB
	letterService  *LetterService
	courierService *CourierService
	wsService      WebSocketNotifier
}

// Use the WebSocketNotifier interface from courier_service.go to avoid redeclaration

// NewQRScanService - Factory pattern for clean dependency injection
func NewQRScanService(db *gorm.DB, letterService *LetterService, courierService *CourierService, wsService WebSocketNotifier) *QRScanService {
	return &QRScanService{
		db:             db,
		letterService:  letterService,
		courierService: courierService,
		wsService:      wsService,
	}
}

// QRScanRequest - Clean API contract
type QRScanRequest struct {
	Code        string  `json:"code" binding:"required"`
	CourierID   string  `json:"courier_id" binding:"required"`
	Action      string  `json:"action" binding:"required,oneof=pickup delivery"`
	Location    string  `json:"location"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Notes       string  `json:"notes"`
}

// QRScanResponse - Rich response with full context
type QRScanResponse struct {
	Success     bool                `json:"success"`
	Letter      *LetterInfo         `json:"letter,omitempty"`
	Task        *CourierTaskInfo    `json:"task,omitempty"`
	NextAction  string              `json:"next_action,omitempty"`
	Message     string              `json:"message"`
}

// LetterInfo - Optimized data structure
type LetterInfo struct {
	ID              string    `json:"id"`
	Code            string    `json:"code"`
	Status          string    `json:"status"`
	SenderName      string    `json:"sender_name"`
	RecipientOPCode string    `json:"recipient_op_code"`
	SenderOPCode    string    `json:"sender_op_code,omitempty"`
	Priority        string    `json:"priority"`
	CreatedAt       time.Time `json:"created_at"`
	LastUpdate      time.Time `json:"last_update"`
}

// CourierTaskInfo - Task context for courier
type CourierTaskInfo struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Priority     string `json:"priority"`
	PickupCode   string `json:"pickup_op_code"`
	DeliveryCode string `json:"delivery_op_code"`
	Reward       int    `json:"reward"`
}

// ProcessQRScan - Main business logic with elegant error handling
func (s *QRScanService) ProcessQRScan(req *QRScanRequest) (*QRScanResponse, error) {
	// Strategy Pattern: Different handlers for different actions
	switch req.Action {
	case "pickup":
		return s.processPickup(req)
	case "delivery":
		return s.processDelivery(req)
	default:
		return nil, fmt.Errorf("unsupported action: %s", req.Action)
	}
}

// processPickup - Clean, focused logic
func (s *QRScanService) processPickup(req *QRScanRequest) (*QRScanResponse, error) {
	// 1. Validate letter exists and is in correct state
	letter, err := s.findLetterByCode(req.Code)
	if err != nil {
		return &QRScanResponse{
			Success: false,
			Message: "信件不存在或编号错误",
		}, nil
	}

	if letter.Status != models.StatusGenerated {
		return &QRScanResponse{
			Success: false,
			Letter:  s.mapLetterInfo(letter),
			Message: "此信件已被处理，当前状态：" + s.getStatusText(string(letter.Status)),
		}, nil
	}

	// 2. Validate courier permissions
	canHandle, err := s.validateCourierPermissions(req.CourierID, letter.RecipientOPCode)
	if err != nil {
		return nil, err
	}
	if !canHandle {
		return &QRScanResponse{
			Success: false,
			Letter:  s.mapLetterInfo(letter),
			Message: "您没有权限处理此区域的信件",
		}, nil
	}

	// 3. Create or update courier task
	task, err := s.createCourierTask(letter, req)
	if err != nil {
		return nil, err
	}

	// 4. Update letter status  
	letter.Status = models.StatusCollected
	// Note: Letter model doesn't have CollectedAt field, using status tracking

	if err := s.db.Save(letter).Error; err != nil {
		return nil, fmt.Errorf("failed to update letter status: %w", err)
	}

	// 5. Record scan history
	s.recordScanHistory(req, letter, "pickup")

	// 6. Send real-time notification
	s.notifyLetterStatusUpdate(letter, "已被信使收取")

	return &QRScanResponse{
		Success:    true,
		Letter:     s.mapLetterInfo(letter),
		Task:       s.mapTaskInfo(task),
		NextAction: "delivery",
		Message:    "信件收取成功，请前往投递地址",
	}, nil
}

// processDelivery - Delivery logic with validation
func (s *QRScanService) processDelivery(req *QRScanRequest) (*QRScanResponse, error) {
	// 1. Find letter and validate state
	letter, err := s.findLetterByCode(req.Code)
	if err != nil {
		return &QRScanResponse{
			Success: false,
			Message: "信件不存在或编号错误",
		}, nil
	}

	if letter.Status != models.StatusCollected && letter.Status != models.StatusInTransit {
		return &QRScanResponse{
			Success: false,
			Letter:  s.mapLetterInfo(letter),
			Message: "此信件状态异常，请联系管理员",
		}, nil
	}

	// 2. Validate courier (must be the same courier who picked up)
	// Note: Letter model doesn't have CourierID field, using task-based validation
	var task models.CourierTask
	err = s.db.Where("letter_id = ? AND courier_id = ? AND status IN ?", letter.ID, req.CourierID, []string{"accepted", "in_progress"}).First(&task).Error
	if err != nil {
		return &QRScanResponse{
			Success: false,
			Letter:  s.mapLetterInfo(letter),
			Message: "只有收取此信件的信使才能投递",
		}, nil
	}

	// 3. Update letter to delivered
	letter.Status = models.StatusDelivered
	// Note: Letter model doesn't have DeliveredAt/DeliveryLocation fields, using status tracking

	if err := s.db.Save(letter).Error; err != nil {
		return nil, fmt.Errorf("failed to update letter status: %w", err)
	}

	// 4. Update courier task
	s.completeCourierTask(letter.ID, req)

	// 5. Update courier statistics
	s.updateCourierStats(req.CourierID, true)

	// 6. Record scan history
	s.recordScanHistory(req, letter, "delivery")

	// 7. Send notifications
	s.notifyLetterStatusUpdate(letter, "已成功投递")
	s.notifyDeliveryComplete(letter)

	return &QRScanResponse{
		Success:    true,
		Letter:     s.mapLetterInfo(letter),
		NextAction: "complete",
		Message:    "信件投递成功！",
	}, nil
}

// Helper methods - Clean, focused implementations

func (s *QRScanService) findLetterByCode(code string) (*models.Letter, error) {
	var letter models.Letter
	err := s.db.Where("code = ?", code).First(&letter).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("letter not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &letter, nil
}

func (s *QRScanService) validateCourierPermissions(courierID, targetOPCode string) (bool, error) {
	var courier models.Courier
	err := s.db.Where("user_id = ?", courierID).First(&courier).Error
	if err != nil {
		return false, fmt.Errorf("courier not found: %w", err)
	}

	// Check if courier's managed OP code prefix covers the target
	if courier.ManagedOPCodePrefix == "" {
		return false, nil
	}

	// Simple prefix matching - can be enhanced with more sophisticated logic
	return len(targetOPCode) >= len(courier.ManagedOPCodePrefix) && 
		   targetOPCode[:len(courier.ManagedOPCodePrefix)] == courier.ManagedOPCodePrefix, nil
}

func (s *QRScanService) createCourierTask(letter *models.Letter, req *QRScanRequest) (*models.CourierTask, error) {
	// Get letter code
	var letterCode models.LetterCode
	err := s.db.Where("letter_id = ?", letter.ID).First(&letterCode).Error
	if err != nil {
		return nil, fmt.Errorf("letter code not found: %w", err)
	}

	task := &models.CourierTask{
		CourierID:      req.CourierID,
		LetterCode:     letterCode.Code,
		Title:          "Letter Delivery: " + letter.Title,
		SenderName:     getSenderName(letter),
		TargetLocation: letter.RecipientOPCode,
		Status:         "accepted",
		Priority:       "normal", // Default priority since Letter doesn't have Priority field
		PickupOPCode:   letter.SenderOPCode,
		DeliveryOPCode: letter.RecipientOPCode,
		CurrentOPCode:  req.Location,
		Reward:         s.calculateReward(letter),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create courier task: %w", err)
	}

	return task, nil
}

func (s *QRScanService) calculateReward(letter *models.Letter) int {
	// Simple reward calculation - can be enhanced
	baseReward := 10
	// Note: Letter model doesn't have Priority field, using default calculation
	return baseReward
}

func (s *QRScanService) completeCourierTask(letterID string, req *QRScanRequest) {
	s.db.Model(&models.CourierTask{}).
		Where("letter_id = ? AND courier_id = ?", letterID, req.CourierID).
		Updates(map[string]interface{}{
			"status":           "completed",
			"current_op_code":  req.Location,
			"delivery_notes":   req.Notes,
			"completed_at":     time.Now(),
		})
}

func (s *QRScanService) updateCourierStats(courierID string, success bool) {
	// Update courier statistics - elegant increment
	updates := map[string]interface{}{
		"task_count": gorm.Expr("task_count + 1"),
	}
	
	if success {
		updates["points"] = gorm.Expr("points + 10")
	}

	s.db.Model(&models.Courier{}).Where("user_id = ?", courierID).Updates(updates)
}

func (s *QRScanService) recordScanHistory(req *QRScanRequest, letter *models.Letter, action string) {
	scanRecord := &models.ScanRecord{
		CourierID:   req.CourierID,
		LetterCode:  req.Code,
		ScanType:    action,
		Location:    req.Location,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Timestamp:   time.Now(),
		Notes:       req.Notes,
	}
	s.db.Create(scanRecord)
}

// Real-time notification methods - SOTA: Clean interface usage
func (s *QRScanService) notifyLetterStatusUpdate(letter *models.Letter, message string) {
	if s.wsService != nil {
		notification := map[string]interface{}{
			"type":      "letter_status_update",
			"letter_id": letter.ID,
			"status":    letter.Status,
			"message":   message,
			"timestamp": time.Now().Unix(),
		}
		s.wsService.BroadcastToUser(letter.UserID, notification)
	}
}

func (s *QRScanService) notifyDeliveryComplete(letter *models.Letter) {
	if s.wsService != nil {
		notification := map[string]interface{}{
			"type":      "delivery_complete",
			"letter_id": letter.ID,
			"message":   "您的信件已成功投递",
			"timestamp": time.Now().Unix(),
		}
		s.wsService.BroadcastToUser(letter.UserID, notification)
	}
}

// Mapping methods - Clean data transformation
func (s *QRScanService) mapLetterInfo(letter *models.Letter) *LetterInfo {
	// Get letter code for display
	letterCode := "N/A"
	if letter.Code != nil {
		letterCode = letter.Code.Code
	}
	
	// Get sender name
	senderName := getSenderName(letter)
	
	return &LetterInfo{
		ID:              letter.ID,
		Code:            letterCode,
		Status:          string(letter.Status),
		SenderName:      senderName,
		RecipientOPCode: letter.RecipientOPCode,
		SenderOPCode:    letter.SenderOPCode,
		Priority:        "normal", // Default since Letter doesn't have Priority field
		CreatedAt:       letter.CreatedAt,
		LastUpdate:      letter.UpdatedAt,
	}
}

func (s *QRScanService) mapTaskInfo(task *models.CourierTask) *CourierTaskInfo {
	return &CourierTaskInfo{
		ID:           task.ID,
		Status:       task.Status,
		Priority:     task.Priority,
		PickupCode:   task.PickupOPCode,
		DeliveryCode: task.DeliveryOPCode,
		Reward:       task.Reward,
	}
}

func (s *QRScanService) getStatusText(status string) string {
	statusMap := map[string]string{
		string(models.StatusDraft):     "草稿",
		string(models.StatusGenerated): "待收取",
		string(models.StatusCollected): "已收取",
		string(models.StatusInTransit): "运输中",
		string(models.StatusDelivered): "已投递",
		string(models.StatusRead):      "已阅读",
	}
	if text, ok := statusMap[status]; ok {
		return text
	}
	return status
}

// GetLetterInfoByCode - 复用现有查询逻辑，提供给handler层
func (s *QRScanService) GetLetterInfoByCode(code string) (*LetterInfo, error) {
	letter, err := s.findLetterByCode(code)
	if err != nil {
		return nil, err
	}
	return s.mapLetterInfo(letter), nil
}

// GetScanHistory - 分页查询扫描历史，继承现有分页模式
func (s *QRScanService) GetScanHistory(courierID string, page, limit int) ([]models.ScanRecord, int64, error) {
	var records []models.ScanRecord
	var total int64

	// 计算总数
	if err := s.db.Model(&models.ScanRecord{}).Where("courier_id = ?", courierID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，包含关联数据
	offset := (page - 1) * limit
	if err := s.db.Where("courier_id = ?", courierID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// ValidateQRCode - QR码格式验证，支持多种格式
func (s *QRScanService) ValidateQRCode(qrContent string) (bool, map[string]interface{}) {
	// 尝试解析JSON格式的QR码（增强格式）
	var qrData map[string]interface{}
	info := make(map[string]interface{})
	
	if err := json.Unmarshal([]byte(qrContent), &qrData); err == nil {
		// JSON格式QR码
		if qrType, ok := qrData["type"].(string); ok && qrType == "openpenpal_letter" {
			info["format"] = "json"
			info["version"] = qrData["version"]
			if letterCode, exists := qrData["code"]; exists {
				info["letter_code"] = letterCode
			}
			return true, info
		}
	}

	// 简单字符串格式（兼容性）
	if len(qrContent) >= 8 && len(qrContent) <= 20 {
		info["format"] = "simple"
		info["letter_code"] = qrContent
		return true, info
	}

	info["error"] = "不是有效的OpenPenPal信件QR码"
	return false, info
}

// Helper function to get sender name from letter
func getSenderName(letter *models.Letter) string {
	if letter.User != nil {
		if letter.User.Nickname != "" {
			return letter.User.Nickname
		}
		return letter.User.Username
	}
	return "Unknown"
}