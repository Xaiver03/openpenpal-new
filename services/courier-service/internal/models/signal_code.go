package models

import (
	"fmt"
	"time"
)

// SignalCode 六位信号编码表 - OP Code System核心实现
// 格式: AABBCC，其中AA=学校代码(2位)，BB=片区代码(2位)，CC=具体位置(2位)
type SignalCode struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Code string `gorm:"not null;unique;size:6" json:"code"` // 六位编码，如 "PK5F3D"

	// OP Code结构化字段
	SchoolCode string `gorm:"not null;size:2;index" json:"school_code"` // 前2位: 学校代码 (如: PK)
	AreaCode   string `gorm:"not null;size:2;index" json:"area_code"`   // 中2位: 片区代码 (如: 5F)
	PointCode  string `gorm:"not null;size:2" json:"point_code"`        // 后2位: 位置代码 (如: 3D)

	// 类型和位置信息
	CodeType   string  `gorm:"not null;index" json:"code_type"` // 编码类型：dormitory, shop, box, club
	SchoolID   string  `gorm:"index" json:"school_id"`          // 所属学校ID
	AreaID     string  `gorm:"index" json:"area_id"`            // 所属片区ID
	BuildingID *string `json:"building_id,omitempty"`           // 所属楼栋ID（可选）
	ZoneCode   string  `gorm:"index" json:"zone_code"`          // 区域编码（兼容旧系统）

	// 隐私控制
	IsPublic bool `gorm:"default:false" json:"is_public"` // 后两位是否公开显示

	// 使用状态
	IsUsed   bool       `gorm:"default:false;index" json:"is_used"`  // 是否已使用
	IsActive bool       `gorm:"default:true;index" json:"is_active"` // 是否激活
	UsedBy   *string    `json:"used_by,omitempty"`                   // 使用者ID（信使或用户）
	UsedAt   *time.Time `json:"used_at,omitempty"`                   // 使用时间
	LetterID *string    `json:"letter_id,omitempty"`                 // 关联的信件ID

	// 元数据
	Description string    `json:"description"`                // 编码描述
	CreatedBy   string    `gorm:"not null" json:"created_by"` // 创建者ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SignalCodeBatch 信号编码批次表
type SignalCodeBatch struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	BatchNo    string    `gorm:"not null;unique" json:"batch_no"` // 批次号
	SchoolID   string    `gorm:"not null;index" json:"school_id"` // 学校ID
	AreaID     string    `gorm:"index" json:"area_id"`            // 片区ID
	CodeType   string    `gorm:"not null" json:"code_type"`       // 编码类型
	StartCode  string    `gorm:"not null" json:"start_code"`      // 起始编码
	EndCode    string    `gorm:"not null" json:"end_code"`        // 结束编码
	TotalCount int       `gorm:"not null" json:"total_count"`     // 总数量
	UsedCount  int       `gorm:"default:0" json:"used_count"`     // 已使用数量
	Status     string    `gorm:"default:active" json:"status"`    // 批次状态: active, exhausted, suspended
	CreatedBy  string    `gorm:"not null" json:"created_by"`      // 创建者ID
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SignalCodeUsageLog 编码使用日志表
type SignalCodeUsageLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Code       string    `gorm:"not null;index" json:"code"`    // 六位编码
	Action     string    `gorm:"not null" json:"action"`        // 操作类型: assign, use, release, deactivate
	UserID     string    `gorm:"not null;index" json:"user_id"` // 操作用户ID
	UserType   string    `gorm:"not null" json:"user_type"`     // 用户类型: courier, admin, system
	TargetID   *string   `json:"target_id,omitempty"`           // 目标ID（信件ID或其他）
	TargetType *string   `json:"target_type,omitempty"`         // 目标类型: letter, zone, etc.
	OldStatus  string    `json:"old_status"`                    // 原状态
	NewStatus  string    `json:"new_status"`                    // 新状态
	Reason     string    `json:"reason"`                        // 操作原因
	Metadata   string    `gorm:"type:json" json:"metadata"`     // 额外元数据JSON
	IPAddress  string    `json:"ip_address"`                    // 操作IP
	CreatedAt  time.Time `json:"created_at"`
}

// SignalCodeRule 编码生成规则表
type SignalCodeRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	RuleName    string    `gorm:"not null;unique" json:"rule_name"` // 规则名称
	SchoolID    string    `gorm:"not null;index" json:"school_id"`  // 学校ID
	CodeType    string    `gorm:"not null" json:"code_type"`        // 编码类型
	Pattern     string    `gorm:"not null" json:"pattern"`          // 编码模式，如 "##@@##"
	Description string    `json:"description"`                      // 规则描述
	IsActive    bool      `gorm:"default:true" json:"is_active"`    // 是否激活
	Priority    int       `gorm:"default:0" json:"priority"`        // 优先级
	CreatedBy   string    `gorm:"not null" json:"created_by"`       // 创建者ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 编码类型常量 - 符合PRD的OP Code类型
const (
	SignalCodeTypeDormitory = "dormitory" // 宿舍
	SignalCodeTypeShop      = "shop"      // 商店
	SignalCodeTypeBox       = "box"       // 投递箱
	SignalCodeTypeClub      = "club"      // 社团空间

	// 兼容旧系统
	SignalCodeTypeLetter   = "letter"   // 信件编码（废弃）
	SignalCodeTypeZone     = "zone"     // 区域编码（废弃）
	SignalCodeTypeSchool   = "school"   // 学校编码（废弃）
	SignalCodeTypeBuilding = "building" // 楼栋编码（废弃）
	SignalCodeTypeSpecial  = "special"  // 特殊编码（保留）
)

// 编码状态常量
const (
	SignalCodeStatusAvailable = "available" // 可用
	SignalCodeStatusAssigned  = "assigned"  // 已分配
	SignalCodeStatusUsed      = "used"      // 已使用
	SignalCodeStatusExpired   = "expired"   // 已过期
	SignalCodeStatusSuspended = "suspended" // 已暂停
)

// 批次状态常量
const (
	BatchStatusActive    = "active"    // 活跃
	BatchStatusExhausted = "exhausted" // 已用尽
	BatchStatusSuspended = "suspended" // 已暂停
)

// 操作类型常量
const (
	ActionAssign     = "assign"     // 分配
	ActionUse        = "use"        // 使用
	ActionRelease    = "release"    // 释放
	ActionDeactivate = "deactivate" // 停用
	ActionReactivate = "reactivate" // 重新激活
)

// SignalCodeRequest 编码申请请求
type SignalCodeRequest struct {
	SchoolID   string `json:"school_id" binding:"required"`
	AreaID     string `json:"area_id" binding:"required"`
	CodeType   string `json:"code_type" binding:"required,oneof=letter zone school building special"`
	Quantity   int    `json:"quantity" binding:"required,min=1,max=1000"`
	Reason     string `json:"reason" binding:"required"`
	BuildingID string `json:"building_id,omitempty"`
}

// SignalCodeAssignRequest 编码分配请求
type SignalCodeAssignRequest struct {
	Code       string `json:"code" binding:"required,len=6"`
	UserID     string `json:"user_id" binding:"required"`
	TargetID   string `json:"target_id,omitempty"`
	TargetType string `json:"target_type,omitempty"`
	Reason     string `json:"reason"`
}

// SignalCodeBatchRequest 批量编码请求
type SignalCodeBatchRequest struct {
	SchoolID  string `json:"school_id" binding:"required"`
	AreaID    string `json:"area_id" binding:"required"`
	CodeType  string `json:"code_type" binding:"required"`
	StartCode string `json:"start_code" binding:"required,len=6"`
	EndCode   string `json:"end_code" binding:"required,len=6"`
	BatchNo   string `json:"batch_no" binding:"required"`
}

// SignalCodeSearchRequest 编码搜索请求
type SignalCodeSearchRequest struct {
	Code     string `form:"code"`
	SchoolID string `form:"school_id"`
	AreaID   string `form:"area_id"`
	CodeType string `form:"code_type"`
	IsUsed   *bool  `form:"is_used"`
	IsActive *bool  `form:"is_active"`
	UsedBy   string `form:"used_by"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// SignalCodeStats 编码统计信息
type SignalCodeStats struct {
	SchoolID       string         `json:"school_id"`
	SchoolName     string         `json:"school_name"`
	TotalCodes     int64          `json:"total_codes"`
	UsedCodes      int64          `json:"used_codes"`
	AvailableCodes int64          `json:"available_codes"`
	UsageRate      float64        `json:"usage_rate"`
	ByType         map[string]int `json:"by_type"`
	ByArea         map[string]int `json:"by_area"`
}

// GenerateOPCode 生成符合PRD规范的OP Code
func GenerateOPCode(schoolCode, areaCode, pointCode string) string {
	return fmt.Sprintf("%s%s%s", schoolCode, areaCode, pointCode)
}

// ParseOPCode 解析OP Code为三段
func ParseOPCode(code string) (schoolCode, areaCode, pointCode string, err error) {
	if !IsValidSignalCode(code) {
		return "", "", "", fmt.Errorf("invalid OP Code format")
	}
	return code[:2], code[2:4], code[4:6], nil
}

// IsValidCode 验证编码格式
func IsValidSignalCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	// OP Code必须是大写字母或数字
	for _, c := range code {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// FormatOPCode 格式化OP Code显示（支持隐私保护）
func FormatOPCode(code string, hidePrivate bool) string {
	if !IsValidSignalCode(code) {
		return code
	}
	if hidePrivate {
		// 隐藏后两位
		return code[:4] + "**"
	}
	return code
}

// GetStatusName 获取状态名称
func (s *SignalCode) GetStatusName() string {
	if !s.IsActive {
		return "已停用"
	}
	if s.IsUsed {
		return "已使用"
	}
	return "可用"
}

// CanBeUsed 检查编码是否可以使用
func (s *SignalCode) CanBeUsed() bool {
	return s.IsActive && !s.IsUsed
}

// MarkAsUsed 标记为已使用
func (s *SignalCode) MarkAsUsed(userID string, letterID *string) {
	s.IsUsed = true
	s.UsedBy = &userID
	if letterID != nil {
		s.LetterID = letterID
	}
	now := time.Now()
	s.UsedAt = &now
}

// Release 释放编码
func (s *SignalCode) Release() {
	s.IsUsed = false
	s.UsedBy = nil
	s.LetterID = nil
	s.UsedAt = nil
}

// DefaultSignalCodeRules 默认编码规则配置
var DefaultSignalCodeRules = []SignalCodeRule{
	{
		RuleName:    "letter_code_pattern",
		SchoolID:    "default",
		CodeType:    SignalCodeTypeLetter,
		Pattern:     "##@@##", // 数字+字母+数字
		Description: "信件编码模式：两位数字+两位字母+两位数字",
		IsActive:    true,
		Priority:    1,
		CreatedBy:   "system",
	},
	{
		RuleName:    "zone_code_pattern",
		SchoolID:    "default",
		CodeType:    SignalCodeTypeZone,
		Pattern:     "@@####", // 字母+数字
		Description: "区域编码模式：两位字母+四位数字",
		IsActive:    true,
		Priority:    1,
		CreatedBy:   "system",
	},
	{
		RuleName:    "school_code_pattern",
		SchoolID:    "default",
		CodeType:    SignalCodeTypeSchool,
		Pattern:     "###@@@", // 数字+字母
		Description: "学校编码模式：三位数字+三位字母",
		IsActive:    true,
		Priority:    1,
		CreatedBy:   "system",
	},
}
