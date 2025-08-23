package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"openpenpal-backend/internal/models"
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
		Code:          fullCode,
		SchoolCode:    application.SchoolCode,
		AreaCode:      application.AreaCode,
		PointCode:     pointCode,
		PointType:     application.PointType,
		PointName:     application.PointName,
		FullAddress:   application.FullAddress,
		IsPublic:      false, // Default to false for privacy
		IsActive:      true,
		BindingType:   "user",
		BindingID:     &application.UserID,
		BindingStatus: "approved",
		ManagedBy:     reviewerID,
		ApprovedBy:    &reviewerID,
		ApprovedAt:    &now,
		CreatedAt:     now,
		UpdatedAt:     now,
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

// GetAllApplications 获取所有OP Code申请
func (s *OPCodeService) GetAllApplications() ([]*models.OPCodeApplication, error) {
	var applications []*models.OPCodeApplication
	err := s.db.Order("created_at DESC").Find(&applications).Error
	return applications, err
}

// GetApplicationByID 根据ID获取申请
func (s *OPCodeService) GetApplicationByID(id string) (*models.OPCodeApplication, error) {
	var application models.OPCodeApplication
	err := s.db.First(&application, "id = ?", id).Error
	return &application, err
}

// RejectApplication 拒绝申请
func (s *OPCodeService) RejectApplication(applicationID string, reviewerID string, reason string) error {
	now := time.Now()
	return s.db.Model(&models.OPCodeApplication{}).
		Where("id = ?", applicationID).
		Updates(map[string]interface{}{
			"status":      models.OPCodeStatusRejected,
			"reviewer_id": reviewerID,
			"reviewed_at": now,
			"updated_at":  now,
		}).Error
}

// CreateOPCode 创建OP Code
func (s *OPCodeService) CreateOPCode(opCode *models.SignalCode) error {
	opCode.Code = strings.ToUpper(opCode.Code)
	opCode.SchoolCode = opCode.Code[:2]
	opCode.AreaCode = opCode.Code[2:4]
	opCode.PointCode = opCode.Code[4:6]
	opCode.CreatedAt = time.Now()
	opCode.UpdatedAt = time.Now()
	return s.db.Create(opCode).Error
}

// UpdateOPCode 更新OP Code
func (s *OPCodeService) UpdateOPCode(code string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.db.Model(&models.SignalCode{}).
		Where("code = ?", strings.ToUpper(code)).
		Updates(updates).Error
}

// DeleteOPCode 删除OP Code
func (s *OPCodeService) DeleteOPCode(code string) error {
	// 软删除，设置为不活跃
	return s.db.Model(&models.SignalCode{}).
		Where("code = ?", strings.ToUpper(code)).
		Update("is_active", false).Error
}

func (s *OPCodeService) MigrateZoneToOPCode(zone string) (string, error) {
	// 这是一个示例映射函数，实际项目需要根据具体的Zone格式设计映射规则
	// 例如: "BJDX-A-101" -> "BD1A01"

	// 简单的映射逻辑示例
	mappings := map[string]string{
		"BEIJING":    "BJ",
		"BJDX":       "BD",
		"BJDX-A":     "BD1A",
		"BJDX-A-101": "BD1A01",
		// 添加更多映射...
	}

	if opCode, exists := mappings[zone]; exists {
		return opCode, nil
	}

	// 如果没有找到映射，尝试生成一个
	// 这里需要根据实际的Zone命名规则来设计
	return "", fmt.Errorf("无法将Zone '%s' 转换为OP Code", zone)
}

// SearchAreas 搜索片区
func (s *OPCodeService) SearchAreas(schoolCode string) (map[string]interface{}, error) {
	// 参数验证
	if schoolCode == "" {
		return nil, errors.New("学校代码不能为空")
	}
	schoolCode = strings.ToUpper(schoolCode)

	// 验证学校代码是否存在
	var school models.OPCodeSchool
	if err := s.db.Where("school_code = ? AND is_active = ?", schoolCode, true).First(&school).Error; err != nil {
		return nil, errors.New("学校代码不存在")
	}

	var areas []models.OPCodeArea

	// 查询片区数据
	if err := s.db.Where("school_code = ? AND is_active = ?", schoolCode, true).
		Order("area_code").Find(&areas).Error; err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	result := make(map[string]interface{})
	result["school_code"] = schoolCode
	result["school_name"] = school.SchoolName
	result["areas"] = make([]map[string]interface{}, 0, len(areas))

	for _, area := range areas {
		areaData := map[string]interface{}{
			"area_code":   area.AreaCode,
			"area_name":   area.AreaName,
			"description": area.Description,
		}
		result["areas"] = append(result["areas"].([]map[string]interface{}), areaData)
	}

	return result, nil
}

// SearchBuildings 搜索楼栋
func (s *OPCodeService) SearchBuildings(schoolCode, areaCode string) (map[string]interface{}, error) {
	// 参数验证
	if schoolCode == "" {
		return nil, errors.New("学校代码不能为空")
	}

	schoolCode = strings.ToUpper(schoolCode)
	if areaCode != "" {
		areaCode = strings.ToUpper(areaCode)
	}

	// 构建查询条件
	query := s.db.Table("signal_codes").Where("school_code = ? AND is_active = ?", schoolCode, true)
	if areaCode != "" {
		query = query.Where("area_code = ?", areaCode)
	}

	// 查询并按楼栋分组
	var buildings []struct {
		SchoolCode string `json:"school_code"`
		AreaCode   string `json:"area_code"`
		PointCode  string `json:"point_code"`
		PointName  string `json:"point_name"`
		PointType  string `json:"point_type"`
	}

	if err := query.Select("school_code, area_code, point_code, point_name, point_type").
		Group("school_code, area_code, point_code, point_name, point_type").
		Order("area_code, point_code").Scan(&buildings).Error; err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	result := make(map[string]interface{})
	result["school_code"] = schoolCode
	if areaCode != "" {
		result["area_code"] = areaCode
	}
	result["buildings"] = make([]map[string]interface{}, 0, len(buildings))

	for _, building := range buildings {
		buildingData := map[string]interface{}{
			"school_code": building.SchoolCode,
			"area_code":   building.AreaCode,
			"point_code":  building.PointCode,
			"point_name":  building.PointName,
			"point_type":  building.PointType,
		}
		result["buildings"] = append(result["buildings"].([]map[string]interface{}), buildingData)
	}

	return result, nil
}

// SearchPoints 搜索投递点
func (s *OPCodeService) SearchPoints(schoolCode, areaCode string) (map[string]interface{}, error) {
	// 参数验证
	if schoolCode == "" {
		return nil, errors.New("学校代码不能为空")
	}

	schoolCode = strings.ToUpper(schoolCode)
	if areaCode != "" {
		areaCode = strings.ToUpper(areaCode)
	}

	// 使用临时结构体映射数据库字段
	type TempSignalCode struct {
		Code        string `json:"code"`
		SchoolCode  string `json:"school_code"`
		AreaCode    string `json:"area_code"`
		PointCode   string `json:"point_code"`
		Description string `json:"description"` // 映射数据库中的description字段
		CodeType    string `json:"code_type"`
		IsPublic    bool   `json:"is_public"`
	}

	// 构建查询条件
	query := s.db.Table("signal_codes").Where("school_code = ? AND is_active = ?", schoolCode, true)
	if areaCode != "" {
		query = query.Where("area_code = ?", areaCode)
	}

	var points []TempSignalCode

	if err := query.Select("code, school_code, area_code, point_code, description, code_type, is_public").
		Order("area_code, point_code").Scan(&points).Error; err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	result := make(map[string]interface{})
	result["school_code"] = schoolCode
	if areaCode != "" {
		result["area_code"] = areaCode
	}
	result["points"] = make([]map[string]interface{}, 0, len(points))

	for _, point := range points {
		pointData := map[string]interface{}{
			"code":        point.Code,
			"school_code": point.SchoolCode,
			"area_code":   point.AreaCode,
			"point_code":  point.PointCode,
			"point_name":  point.Description, // 使用description作为point_name
			"point_type":  point.CodeType,
			"is_public":   point.IsPublic,
		}
		result["points"] = append(result["points"].([]map[string]interface{}), pointData)
	}

	return result, nil
}

// SearchSchools 搜索学校
func (s *OPCodeService) SearchSchools(name string, page, limit int) (map[string]interface{}, error) {
	var schools []models.OPCodeSchool
	var total int64

	query := s.db.Model(&models.OPCodeSchool{}).Where("is_active = ?", true)

	// 如果提供了名称，进行模糊搜索
	if name != "" {
		query = query.Where("school_name ILIKE ? OR full_name ILIKE ?", "%"+name+"%", "%"+name+"%")
	}

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&schools).Error; err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	result := make(map[string]interface{})
	result["schools"] = schools
	result["pagination"] = map[string]interface{}{
		"page":       page,
		"limit":      limit,
		"total":      total,
		"total_page": (total + int64(limit) - 1) / int64(limit),
	}

	return result, nil
}

// SearchSchoolsByCity 按城市搜索学校的OPCode信息
func (s *OPCodeService) SearchSchoolsByCity(cityName string, page, limit int) (map[string]interface{}, error) {
	var schools []models.OPCodeSchool
	var total int64

	if cityName == "" {
		return nil, errors.New("城市名称不能为空")
	}

	query := s.db.Model(&models.OPCodeSchool{}).Where("is_active = ?", true)
	
	// 城市模糊搜索 - 支持"北京"查找"北京市"等
	query = query.Where("city ILIKE ?", "%"+cityName+"%")

	// 计算总数
	query.Count(&total)

	// 分页查询，按学校名称排序
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("school_name").Find(&schools).Error; err != nil {
		return nil, err
	}

	// 转换为前端需要的格式，包含城市信息
	result := make(map[string]interface{})
	result["schools"] = schools
	result["city"] = cityName
	result["pagination"] = map[string]interface{}{
		"page":       page,
		"limit":      limit,
		"total":      total,
		"total_page": (total + int64(limit) - 1) / int64(limit),
	}

	return result, nil
}

// SearchSchoolsAdvanced 高级学校搜索（支持多条件）
func (s *OPCodeService) SearchSchoolsAdvanced(req *models.AdvancedSchoolSearchRequest) (map[string]interface{}, error) {
	var schools []models.OPCodeSchool
	var total int64

	query := s.db.Model(&models.OPCodeSchool{}).Where("is_active = ?", true)

	// 学校名称搜索
	if req.SchoolName != "" {
		query = query.Where("school_name ILIKE ? OR full_name ILIKE ?", "%"+req.SchoolName+"%", "%"+req.SchoolName+"%")
	}

	// 城市搜索
	if req.City != "" {
		query = query.Where("city ILIKE ?", "%"+req.City+"%")
	}

	// 省份搜索
	if req.Province != "" {
		query = query.Where("province ILIKE ?", "%"+req.Province+"%")
	}

	// 学校代码搜索
	if req.SchoolCode != "" {
		query = query.Where("school_code ILIKE ?", "%"+strings.ToUpper(req.SchoolCode)+"%")
	}

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.Limit
	orderBy := "school_name"
	if req.SortBy != "" {
		orderBy = req.SortBy
	}
	if req.SortOrder == "desc" {
		orderBy += " DESC"
	}
	
	if err := query.Offset(offset).Limit(req.Limit).Order(orderBy).Find(&schools).Error; err != nil {
		return nil, err
	}

	// 转换为前端需要的格式
	result := make(map[string]interface{})
	result["schools"] = schools
	result["search_criteria"] = map[string]interface{}{
		"school_name": req.SchoolName,
		"city":        req.City,
		"province":    req.Province,
		"school_code": req.SchoolCode,
	}
	result["pagination"] = map[string]interface{}{
		"page":       req.Page,
		"limit":      req.Limit,
		"total":      total,
		"total_page": (total + int64(req.Limit) - 1) / int64(req.Limit),
	}

	return result, nil
}

// generateID 生成UUID（简化版）
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetPendingApplications 获取待审核的申请列表
func (s *OPCodeService) GetPendingApplications() ([]*models.OPCodeApplication, error) {
	var applications []*models.OPCodeApplication
	err := s.db.Where("status = ?", models.OPCodeStatusPending).
		Order("created_at DESC").
		Find(&applications).Error
	return applications, err
}

// GetOPCodesByPrefix 根据前缀获取OP Code列表
func (s *OPCodeService) GetOPCodesByPrefix(prefix string) ([]models.SignalCode, error) {
	var codes []models.SignalCode
	query := s.db.Where("code LIKE ?", prefix+"%")
	
	if err := query.Find(&codes).Error; err != nil {
		return nil, err
	}
	
	return codes, nil
}

// GetAllOPCodes 获取所有OP Code列表
func (s *OPCodeService) GetAllOPCodes() ([]*models.SignalCode, error) {
	var codes []*models.SignalCode
	if err := s.db.Order("code").Find(&codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

// GetOPCodeByID 根据ID获取OP Code
func (s *OPCodeService) GetOPCodeByID(id string) (*models.SignalCode, error) {
	var code models.SignalCode
	err := s.db.First(&code, "id = ?", id).Error
	return &code, err
}

// UpdateOPCodeByModel 更新OP Code（接受SignalCode对象）
func (s *OPCodeService) UpdateOPCodeByModel(opCode *models.SignalCode) error {
	opCode.UpdatedAt = time.Now()
	return s.db.Save(opCode).Error
}

// DeleteOPCodeByID 删除OP Code（根据ID）
func (s *OPCodeService) DeleteOPCodeByID(id string) error {
	// 软删除，设置为不活跃
	return s.db.Model(&models.SignalCode{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "inactive",
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error
}
