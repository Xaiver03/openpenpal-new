package services

import (
	"courier-service/internal/models"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SignalCodeService 六位信号编码服务
type SignalCodeService struct {
	DB *gorm.DB
}

// NewSignalCodeService 创建信号编码服务
func NewSignalCodeService(db *gorm.DB) *SignalCodeService {
	return &SignalCodeService{DB: db}
}

// GenerateCodeBatch 生成编码批次
func (s *SignalCodeService) GenerateCodeBatch(req *models.SignalCodeBatchRequest) (*models.SignalCodeBatch, error) {
	// 验证起始和结束编码
	if !models.IsValidSignalCode(req.StartCode) || !models.IsValidSignalCode(req.EndCode) {
		return nil, errors.New("invalid start or end code format")
	}

	startNum, err := strconv.Atoi(req.StartCode)
	if err != nil {
		return nil, errors.New("invalid start code number")
	}

	endNum, err := strconv.Atoi(req.EndCode)
	if err != nil {
		return nil, errors.New("invalid end code number")
	}

	if startNum >= endNum {
		return nil, errors.New("start code must be less than end code")
	}

	totalCount := endNum - startNum + 1

	// 创建批次记录
	batch := &models.SignalCodeBatch{
		BatchNo:    req.BatchNo,
		SchoolID:   req.SchoolID,
		AreaID:     req.AreaID,
		CodeType:   req.CodeType,
		StartCode:  req.StartCode,
		EndCode:    req.EndCode,
		TotalCount: totalCount,
		UsedCount:  0,
		Status:     models.BatchStatusActive,
		CreatedBy:  "system", // 应该从context获取
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.DB.Create(batch).Error; err != nil {
		return nil, err
	}

	// 批量生成编码
	var codes []models.SignalCode
	for i := startNum; i <= endNum; i++ {
		code := models.SignalCode{
			Code:        fmt.Sprintf("%06d", i),
			CodeType:    req.CodeType,
			SchoolID:    req.SchoolID,
			AreaID:      req.AreaID,
			ZoneCode:    fmt.Sprintf("%s_%s", req.SchoolID, req.AreaID),
			IsUsed:      false,
			IsActive:    true,
			Description: fmt.Sprintf("批次 %s 生成的 %s 类型编码", req.BatchNo, req.CodeType),
			CreatedBy:   "system",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		codes = append(codes, code)
	}

	// 批量插入编码
	if err := s.DB.CreateInBatches(codes, 100).Error; err != nil {
		return nil, err
	}

	return batch, nil
}

// RequestSignalCode 申请信号编码
func (s *SignalCodeService) RequestSignalCode(req *models.SignalCodeRequest) (*models.SignalCode, error) {
	// 查找可用的编码
	var code models.SignalCode
	query := s.DB.Where("school_id = ? AND area_id = ? AND code_type = ? AND is_used = ? AND is_active = ?",
		req.SchoolID, req.AreaID, req.CodeType, false, true)

	if req.BuildingID != "" {
		query = query.Where("building_id = ?", req.BuildingID)
	}

	if err := query.First(&code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no available signal code found")
		}
		return nil, err
	}

	return &code, nil
}

// AssignSignalCode 分配信号编码
func (s *SignalCodeService) AssignSignalCode(req *models.SignalCodeAssignRequest) error {
	// 开始事务
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找编码
	var code models.SignalCode
	if err := tx.Where("code = ?", req.Code).First(&code).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 检查编码是否可用
	if !code.CanBeUsed() {
		tx.Rollback()
		return errors.New("signal code is not available")
	}

	// 更新编码状态
	oldStatus := code.GetStatusName()
	code.MarkAsUsed(req.UserID, &req.TargetID)

	if err := tx.Save(&code).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 记录使用日志
	log := models.SignalCodeUsageLog{
		Code:       req.Code,
		Action:     models.ActionAssign,
		UserID:     req.UserID,
		UserType:   "courier", // 应该从context获取
		TargetID:   &req.TargetID,
		TargetType: &req.TargetType,
		OldStatus:  oldStatus,
		NewStatus:  code.GetStatusName(),
		Reason:     req.Reason,
		IPAddress:  "127.0.0.1", // 应该从context获取
		CreatedAt:  time.Now(),
	}

	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ReleaseSignalCode 释放信号编码
func (s *SignalCodeService) ReleaseSignalCode(code string, userID string, reason string) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var signalCode models.SignalCode
	if err := tx.Where("code = ?", code).First(&signalCode).Error; err != nil {
		tx.Rollback()
		return err
	}

	oldStatus := signalCode.GetStatusName()
	signalCode.Release()

	if err := tx.Save(&signalCode).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 记录使用日志
	log := models.SignalCodeUsageLog{
		Code:      code,
		Action:    models.ActionRelease,
		UserID:    userID,
		UserType:  "courier",
		OldStatus: oldStatus,
		NewStatus: signalCode.GetStatusName(),
		Reason:    reason,
		IPAddress: "127.0.0.1",
		CreatedAt: time.Now(),
	}

	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// SearchSignalCodes 搜索信号编码
func (s *SignalCodeService) SearchSignalCodes(req *models.SignalCodeSearchRequest) ([]models.SignalCode, int64, error) {
	var codes []models.SignalCode
	var total int64

	query := s.DB.Model(&models.SignalCode{})

	// 构建查询条件
	if req.Code != "" {
		query = query.Where("code LIKE ?", "%"+req.Code+"%")
	}
	if req.SchoolID != "" {
		query = query.Where("school_id = ?", req.SchoolID)
	}
	if req.AreaID != "" {
		query = query.Where("area_id = ?", req.AreaID)
	}
	if req.CodeType != "" {
		query = query.Where("code_type = ?", req.CodeType)
	}
	if req.IsUsed != nil {
		query = query.Where("is_used = ?", *req.IsUsed)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}
	if req.UsedBy != "" {
		query = query.Where("used_by = ?", req.UsedBy)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&codes).Error; err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

// GetSignalCodeStats 获取信号编码统计
func (s *SignalCodeService) GetSignalCodeStats(schoolID string) (*models.SignalCodeStats, error) {
	var stats models.SignalCodeStats

	// 获取学校基本信息
	stats.SchoolID = schoolID

	// 统计总编码数
	if err := s.DB.Model(&models.SignalCode{}).Where("school_id = ?", schoolID).Count(&stats.TotalCodes).Error; err != nil {
		return nil, err
	}

	// 统计已使用编码数
	if err := s.DB.Model(&models.SignalCode{}).Where("school_id = ? AND is_used = ?", schoolID, true).Count(&stats.UsedCodes).Error; err != nil {
		return nil, err
	}

	// 计算可用编码数
	stats.AvailableCodes = stats.TotalCodes - stats.UsedCodes

	// 计算使用率
	if stats.TotalCodes > 0 {
		stats.UsageRate = float64(stats.UsedCodes) / float64(stats.TotalCodes) * 100
	}

	// 按类型统计
	var typeStats []struct {
		CodeType string `json:"code_type"`
		Count    int    `json:"count"`
	}
	if err := s.DB.Model(&models.SignalCode{}).Select("code_type, count(*) as count").Where("school_id = ?", schoolID).Group("code_type").Scan(&typeStats).Error; err != nil {
		return nil, err
	}

	stats.ByType = make(map[string]int)
	for _, stat := range typeStats {
		stats.ByType[stat.CodeType] = stat.Count
	}

	// 按区域统计
	var areaStats []struct {
		AreaID string `json:"area_id"`
		Count  int    `json:"count"`
	}
	if err := s.DB.Model(&models.SignalCode{}).Select("area_id, count(*) as count").Where("school_id = ?", schoolID).Group("area_id").Scan(&areaStats).Error; err != nil {
		return nil, err
	}

	stats.ByArea = make(map[string]int)
	for _, stat := range areaStats {
		stats.ByArea[stat.AreaID] = stat.Count
	}

	return &stats, nil
}

// GetUsageLogs 获取编码使用日志
func (s *SignalCodeService) GetUsageLogs(code string, limit int) ([]models.SignalCodeUsageLog, error) {
	var logs []models.SignalCodeUsageLog
	query := s.DB.Where("code = ?", code).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// CreateMockData 创建测试数据
func (s *SignalCodeService) CreateMockData() error {
	// 创建默认规则
	for _, rule := range models.DefaultSignalCodeRules {
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()
		if err := s.DB.Create(&rule).Error; err != nil {
			// 忽略重复插入错误
			if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "UNIQUE constraint") {
				return err
			}
		}
	}

	// 创建测试学校的编码批次
	schools := []struct {
		ID   string
		Name string
	}{
		{"school_001", "北京大学"},
		{"school_002", "清华大学"},
		{"school_003", "复旦大学"},
	}

	areas := []struct {
		ID   string
		Name string
	}{
		{"area_01", "主校区"},
		{"area_02", "分校区"},
	}

	for _, school := range schools {
		for _, area := range areas {
			// 创建信件编码批次
			batchReq := &models.SignalCodeBatchRequest{
				SchoolID:  school.ID,
				AreaID:    area.ID,
				CodeType:  models.SignalCodeTypeLetter,
				StartCode: fmt.Sprintf("%s%s%s", school.ID[7:], area.ID[5:], "001"),
				EndCode:   fmt.Sprintf("%s%s%s", school.ID[7:], area.ID[5:], "100"),
				BatchNo:   fmt.Sprintf("BATCH_%s_%s_%d", school.ID, area.ID, time.Now().Unix()),
			}

			// 生成一些随机的6位编码
			startNum := rand.Intn(100000) + 100000
			endNum := startNum + 99

			batchReq.StartCode = fmt.Sprintf("%06d", startNum)
			batchReq.EndCode = fmt.Sprintf("%06d", endNum)

			if _, err := s.GenerateCodeBatch(batchReq); err != nil {
				// 如果批次已存在，跳过
				if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "UNIQUE constraint") {
					return err
				}
			}
		}
	}

	return nil
}
