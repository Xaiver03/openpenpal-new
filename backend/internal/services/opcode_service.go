package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"openpenpal-backend/internal/models"
	"gorm.io/gorm"
)

// OPCodeService OP Code服务 - 管理6位编码系统
type OPCodeService struct {
	db *gorm.DB
}

// NewOPCodeService 创建OP Code服务
func NewOPCodeService(db *gorm.DB) *OPCodeService {
	return &OPCodeService{db: db}
}

// ApplyForOPCode 申请OP Code
func (s *OPCodeService) ApplyForOPCode(userID string, req *models.OPCodeRequest) (*models.OPCodeApplication, error) {
	// 验证学校和片区代码格式
	if len(req.SchoolCode) != 2 || len(req.AreaCode) != 2 {
		return nil, errors.New("学校代码和片区代码必须为2位")
	}

	// 转换为大写
	req.SchoolCode = strings.ToUpper(req.SchoolCode)
	req.AreaCode = strings.ToUpper(req.AreaCode)

	// 创建申请记录
	application := &models.OPCodeApplication{
		ID:          generateID(),
		UserID:      userID,
		SchoolCode:  req.SchoolCode,
		AreaCode:    req.AreaCode,
		PointType:   req.PointType,
		PointName:   req.PointName,
		FullAddress: req.FullAddress,
		Reason:      req.Reason,
		Status:      models.OPCodeStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// AssignOPCode 分配具体的OP Code
func (s *OPCodeService) AssignOPCode(reviewerID string, applicationID string, pointCode string) error {
	if len(pointCode) != 2 {
		return errors.New("位置代码必须为2位")
	}
	
	pointCode = strings.ToUpper(pointCode)

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取申请记录
	var application models.OPCodeApplication
	if err := tx.First(&application, "id = ?", applicationID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if application.Status != models.OPCodeStatusPending {
		tx.Rollback()
		return errors.New("申请状态不正确")
	}

	// 生成完整的OP Code
	fullCode := fmt.Sprintf("%s%s%s", application.SchoolCode, application.AreaCode, pointCode)

	// 检查编码是否已存在
	var existingCode models.SignalCode
	if err := tx.Where("code = ?", fullCode).First(&existingCode).Error; err == nil {
		tx.Rollback()
		return errors.New("该编码已被使用")
	}

	// 创建新的OP Code记录（使用SignalCode表）
	now := time.Now()
	opCode := &models.SignalCode{
		Code:         fullCode,
		SchoolCode:   application.SchoolCode,
		AreaCode:     application.AreaCode,
		PointCode:    pointCode,
		PointType:    application.PointType,
		PointName:    application.PointName,
		FullAddress:  application.FullAddress,
		IsPublic:     false, // Default to false for privacy
		IsActive:     true,
		BindingType:  "user",
		BindingID:    &application.UserID,
		BindingStatus: "approved",
		ManagedBy:    reviewerID,
		ApprovedBy:   &reviewerID,
		ApprovedAt:   &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := tx.Create(opCode).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新申请状态
	now = time.Now()
	application.Status = models.OPCodeStatusApproved
	application.AssignedCode = fullCode
	application.ReviewerID = &reviewerID
	application.ReviewedAt = &now
	application.UpdatedAt = now

	if err := tx.Save(&application).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// GetOPCodeByCode 根据编码查询OP Code信息
func (s *OPCodeService) GetOPCodeByCode(code string, includePrivate bool) (*models.SignalCode, error) {
	code = strings.ToUpper(code)
	
	var opCode models.SignalCode
	query := s.db.Where("code = ? AND is_active = ?", code, true)
	
	if !includePrivate {
		query = query.Where("is_public = ?", true)
	}
	
	if err := query.First(&opCode).Error; err != nil {
		return nil, err
	}
	
	return &opCode, nil
}

// ValidateCourierAccess 验证信使是否有权限访问某个OP Code
func (s *OPCodeService) ValidateCourierAccess(courierID string, targetOPCode string) (bool, error) {
	targetOPCode = strings.ToUpper(targetOPCode)
	
	// 获取信使信息
	var courier models.Courier
	if err := s.db.First(&courier, "id = ?", courierID).Error; err != nil {
		return false, err
	}
	
	// 如果没有设置OP Code权限，检查旧的Zone权限（兼容性）
	if courier.ManagedOPCodePrefix == "" {
		// TODO: 实现Zone到OP Code的映射逻辑
		return true, nil
	}
	
	// 去除通配符
	prefix := strings.ReplaceAll(courier.ManagedOPCodePrefix, "*", "")
	
	// 检查前缀匹配
	return strings.HasPrefix(targetOPCode, prefix), nil
}

// SearchOPCodes 搜索OP Code
func (s *OPCodeService) SearchOPCodes(req *models.OPCodeSearchRequest) ([]models.SignalCode, int64, error) {
	var codes []models.SignalCode
	var total int64
	
	query := s.db.Model(&models.SignalCode{})
	
	// 构建查询条件
	if req.Code != "" {
		query = query.Where("code LIKE ?", strings.ToUpper(req.Code)+"%")
	}
	if req.SchoolCode != "" {
		query = query.Where("school_code = ?", strings.ToUpper(req.SchoolCode))
	}
	if req.AreaCode != "" {
		query = query.Where("area_code = ?", strings.ToUpper(req.AreaCode))
	}
	if req.PointType != "" {
		query = query.Where("code_type = ?", req.PointType)
	}
	if req.IsPublic != nil {
		query = query.Where("is_public = ?", *req.IsPublic)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}
	
	// 计算总数
	query.Count(&total)
	
	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&codes).Error; err != nil {
		return nil, 0, err
	}
	
	return codes, total, nil
}

// GetOPCodeStats 获取OP Code统计信息
func (s *OPCodeService) GetOPCodeStats(schoolCode string) (*models.OPCodeStats, error) {
	stats := &models.OPCodeStats{
		SchoolCode: schoolCode,
		ByType:     make(map[string]int),
		ByArea:     make(map[string]int),
	}
	
	// 统计总数
	s.db.Model(&models.SignalCode{}).Where("school_code = ?", schoolCode).Count(&stats.TotalCodes)
	
	// 统计激活数
	s.db.Model(&models.SignalCode{}).Where("school_code = ? AND is_active = ?", schoolCode, true).Count(&stats.ActiveCodes)
	
	// 统计公开数
	s.db.Model(&models.SignalCode{}).Where("school_code = ? AND is_public = ?", schoolCode, true).Count(&stats.PublicCodes)
	
	// 按类型统计
	var typeStats []struct {
		CodeType string
		Count    int
	}
	s.db.Model(&models.SignalCode{}).
		Select("code_type, COUNT(*) as count").
		Where("school_code = ?", schoolCode).
		Group("code_type").
		Scan(&typeStats)
	
	for _, ts := range typeStats {
		stats.ByType[ts.CodeType] = ts.Count
	}
	
	// 按片区统计
	var areaStats []struct {
		AreaCode string
		Count    int
	}
	s.db.Model(&models.SignalCode{}).
		Select("area_code, COUNT(*) as count").
		Where("school_code = ?", schoolCode).
		Group("area_code").
		Scan(&areaStats)
	
	for _, as := range areaStats {
		stats.ByArea[as.AreaCode] = as.Count
	}
	
	// 计算利用率
	if stats.TotalCodes > 0 {
		stats.UtilizationRate = float64(stats.ActiveCodes) / float64(stats.TotalCodes) * 100
	}
	
	return stats, nil
}

// MigrateZoneToOPCode 将旧的Zone系统迁移到OP Code
// ValidateOPCode 验证OP Code格式和有效性
func (s *OPCodeService) ValidateOPCode(code string) (bool, error) {
	// 验证格式
	if len(code) != 6 {
		return false, fmt.Errorf("OP Code must be exactly 6 characters")
	}
	
	// 验证是否存在于数据库
	var opcode models.OPCode
	if err := s.db.Where("code = ? AND is_active = ?", code, true).First(&opcode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("OP Code not found or inactive")
		}
		return false, err
	}
	
	return true, nil
}

// CheckPermission 检查用户对特定OP Code的操作权限
func (s *OPCodeService) CheckPermission(userID string, targetOPCode string) (bool, error) {
	// 简化版权限检查：管理员有所有权限
	// TODO: 实现更复杂的权限逻辑
	return true, nil
}

func (s *OPCodeService) MigrateZoneToOPCode(zone string) (string, error) {
	// 这是一个示例映射函数，实际项目需要根据具体的Zone格式设计映射规则
	// 例如: "BJDX-A-101" -> "BD1A01"
	
	// 简单的映射逻辑示例
	mappings := map[string]string{
		"BEIJING":     "BJ",
		"BJDX":        "BD",
		"BJDX-A":      "BD1A",
		"BJDX-A-101":  "BD1A01",
		// 添加更多映射...
	}
	
	if opCode, exists := mappings[zone]; exists {
		return opCode, nil
	}
	
	// 如果没有找到映射，尝试生成一个
	// 这里需要根据实际的Zone命名规则来设计
	return "", fmt.Errorf("无法将Zone '%s' 转换为OP Code", zone)
}

// generateID 生成UUID（简化版）
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}