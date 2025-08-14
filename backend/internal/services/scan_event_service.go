package services

import (
	"errors"
	"fmt"
	"time"

	"openpenpal-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScanEventService 扫描事件服务 - PRD要求的完整扫描历史管理
type ScanEventService struct {
	db *gorm.DB
}

// NewScanEventService 创建扫描事件服务
func NewScanEventService(db *gorm.DB) *ScanEventService {
	return &ScanEventService{db: db}
}

// CreateScanEvent 创建扫描事件
func (s *ScanEventService) CreateScanEvent(req *models.ScanEventCreateRequest, scannedBy, userAgent, ipAddress string) (*models.ScanEvent, error) {
	// 验证扫描类型
	if !models.ValidateScanType(req.ScanType) {
		return nil, errors.New("无效的扫描类型")
	}

	// 验证条码是否存在
	var letterCode models.LetterCode
	if err := s.db.Where("id = ?", req.BarcodeID).First(&letterCode).Error; err != nil {
		return nil, fmt.Errorf("条码不存在: %w", err)
	}

	// 创建扫描事件
	scanEvent := &models.ScanEvent{
		ID:           uuid.New().String(),
		BarcodeID:    req.BarcodeID,
		LetterCodeID: letterCode.ID,
		ScannedBy:    scannedBy,
		ScanType:     req.ScanType,
		Location:     req.Location,
		OPCode:       req.OPCode,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		OldStatus:    req.OldStatus,
		NewStatus:    req.NewStatus,
		DeviceInfo:   req.DeviceInfo,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		Note:         req.Note,
		Metadata:     req.Metadata,
		Timestamp:    time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(scanEvent).Error; err != nil {
		return nil, fmt.Errorf("创建扫描事件失败: %w", err)
	}

	return scanEvent, nil
}

// GetScanHistory 获取扫描历史
func (s *ScanEventService) GetScanHistory(query *models.ScanEventQuery) ([]models.ScanEvent, int64, error) {
	var events []models.ScanEvent
	var total int64

	// 构建查询
	db := s.db.Model(&models.ScanEvent{})

	// 添加查询条件
	if query.BarcodeID != "" {
		db = db.Where("barcode_id = ?", query.BarcodeID)
	}
	if query.LetterCodeID != "" {
		db = db.Where("letter_code_id = ?", query.LetterCodeID)
	}
	if query.ScannedBy != "" {
		db = db.Where("scanned_by = ?", query.ScannedBy)
	}
	if query.ScanType != "" {
		db = db.Where("scan_type = ?", query.ScanType)
	}
	if query.OPCode != "" {
		db = db.Where("op_code = ?", query.OPCode)
	}
	if query.OldStatus != "" {
		db = db.Where("old_status = ?", query.OldStatus)
	}
	if query.NewStatus != "" {
		db = db.Where("new_status = ?", query.NewStatus)
	}
	if query.StartTime != nil {
		db = db.Where("timestamp >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("timestamp <= ?", query.EndTime)
	}

	// 计算总数
	db.Count(&total)

	// 分页和排序
	orderBy := query.OrderBy
	if orderBy == "" {
		orderBy = "timestamp"
	}
	
	if query.OrderDesc {
		orderBy += " DESC"
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Preload("Scanner").Preload("LetterCode").
		Order(orderBy).
		Offset(offset).
		Limit(query.PageSize).
		Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("查询扫描历史失败: %w", err)
	}

	return events, total, nil
}

// GetScanEventByID 根据ID获取扫描事件
func (s *ScanEventService) GetScanEventByID(id string) (*models.ScanEvent, error) {
	var event models.ScanEvent
	if err := s.db.Preload("Scanner").Preload("LetterCode").
		Where("id = ?", id).First(&event).Error; err != nil {
		return nil, fmt.Errorf("扫描事件不存在: %w", err)
	}
	return &event, nil
}

// GetBarcodeTimeline 获取条码时间线
func (s *ScanEventService) GetBarcodeTimeline(barcodeID string) ([]models.ScanEvent, error) {
	var events []models.ScanEvent
	if err := s.db.Preload("Scanner").
		Where("barcode_id = ?", barcodeID).
		Order("timestamp ASC").
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("获取条码时间线失败: %w", err)
	}
	return events, nil
}

// GetScanEventSummary 获取扫描事件统计摘要
func (s *ScanEventService) GetScanEventSummary(query *models.ScanEventQuery) (*models.ScanEventSummary, error) {
	summary := &models.ScanEventSummary{
		ByType:   make(map[models.ScanEventType]int64),
		ByStatus: make(map[models.BarcodeStatus]int64),
		ByHour:   make(map[int]int64),
	}

	// 构建基础查询
	baseQuery := s.db.Model(&models.ScanEvent{})
	if query.StartTime != nil {
		baseQuery = baseQuery.Where("timestamp >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		baseQuery = baseQuery.Where("timestamp <= ?", query.EndTime)
	}
	if query.OPCode != "" {
		baseQuery = baseQuery.Where("op_code = ?", query.OPCode)
	}

	// 总扫描次数
	baseQuery.Count(&summary.TotalScans)

	// 唯一用户数
	baseQuery.Distinct("scanned_by").Count(&summary.UniqueUsers)

	// 唯一位置数
	baseQuery.Where("op_code != ''").Distinct("op_code").Count(&summary.UniqueLocations)

	// 按类型统计
	var typeStats []struct {
		ScanType models.ScanEventType `json:"scan_type"`
		Count    int64                `json:"count"`
	}
	baseQuery.Select("scan_type, COUNT(*) as count").
		Group("scan_type").Scan(&typeStats)
	for _, stat := range typeStats {
		summary.ByType[stat.ScanType] = stat.Count
	}

	// 按状态统计
	var statusStats []struct {
		NewStatus models.BarcodeStatus `json:"new_status"`
		Count     int64                `json:"count"`
	}
	baseQuery.Select("new_status, COUNT(*) as count").
		Group("new_status").Scan(&statusStats)
	for _, stat := range statusStats {
		summary.ByStatus[stat.NewStatus] = stat.Count
	}

	// 按小时统计
	var hourStats []struct {
		Hour  int   `json:"hour"`
		Count int64 `json:"count"`
	}
	baseQuery.Select("EXTRACT(hour FROM timestamp) as hour, COUNT(*) as count").
		Group("hour").Scan(&hourStats)
	for _, stat := range hourStats {
		summary.ByHour[stat.Hour] = stat.Count
	}

	// 最近事件
	var recentEvents []models.ScanEvent
	baseQuery.Preload("Scanner").
		Order("timestamp DESC").
		Limit(10).
		Find(&recentEvents)
	summary.RecentEvents = recentEvents

	return summary, nil
}

// RecordBarcodeStatusChange 记录条码状态变更（便捷方法）
func (s *ScanEventService) RecordBarcodeStatusChange(
	barcodeID, scannedBy string,
	scanType models.ScanEventType,
	oldStatus, newStatus models.BarcodeStatus,
	location, opCode, note string,
	userAgent, ipAddress string,
) error {
	req := &models.ScanEventCreateRequest{
		BarcodeID: barcodeID,
		ScanType:  scanType,
		Location:  location,
		OPCode:    opCode,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Note:      note,
	}

	_, err := s.CreateScanEvent(req, scannedBy, userAgent, ipAddress)
	return err
}

// GetUserScanActivity 获取用户扫描活动
func (s *ScanEventService) GetUserScanActivity(userID string, days int) ([]models.ScanEvent, error) {
	var events []models.ScanEvent
	startTime := time.Now().AddDate(0, 0, -days)
	
	if err := s.db.Preload("LetterCode").
		Where("scanned_by = ? AND timestamp >= ?", userID, startTime).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("获取用户扫描活动失败: %w", err)
	}
	
	return events, nil
}

// GetLocationScanStats 获取位置扫描统计
func (s *ScanEventService) GetLocationScanStats(opCode string, days int) (map[string]interface{}, error) {
	startTime := time.Now().AddDate(0, 0, -days)
	
	var total int64
	var uniqueUsers int64
	var recentEvents []models.ScanEvent
	
	baseQuery := s.db.Model(&models.ScanEvent{}).
		Where("op_code = ? AND timestamp >= ?", opCode, startTime)
	
	// 总扫描次数
	baseQuery.Count(&total)
	
	// 唯一用户数
	baseQuery.Distinct("scanned_by").Count(&uniqueUsers)
	
	// 最近事件
	baseQuery.Preload("Scanner").
		Order("timestamp DESC").
		Limit(5).
		Find(&recentEvents)
	
	// 按天统计
	var dailyStats []struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}
	baseQuery.Select("DATE(timestamp) as date, COUNT(*) as count").
		Group("date").
		Order("date DESC").
		Scan(&dailyStats)
	
	return map[string]interface{}{
		"op_code":       opCode,
		"total_scans":   total,
		"unique_users":  uniqueUsers,
		"daily_stats":   dailyStats,
		"recent_events": recentEvents,
	}, nil
}

// CleanupOldScanEvents 清理旧的扫描事件（超过指定天数）
func (s *ScanEventService) CleanupOldScanEvents(days int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	
	result := s.db.Where("timestamp < ?", cutoffTime).Delete(&models.ScanEvent{})
	if result.Error != nil {
		return 0, fmt.Errorf("清理旧扫描事件失败: %w", result.Error)
	}
	
	return result.RowsAffected, nil
}

// GetDB 获取数据库连接（用于其他服务访问）
func (s *ScanEventService) GetDB() *gorm.DB {
	return s.db
}