package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// OPCode 统一的6位编码模型 - OpenPenPal核心地理标识体系
// 格式: XXYYZI，其中XX=学校代码(2位)，YY=片区代码(2位)，ZZ=具体位置(2位)
type OPCode struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Code       string    `json:"code" gorm:"unique;not null;size:6;index"` // 完整6位编码，如: PK5F3D
	SchoolCode string    `json:"school_code" gorm:"not null;size:2;index"` // 前2位: 学校代码
	AreaCode   string    `json:"area_code" gorm:"not null;size:2;index"`   // 中2位: 片区/楼栋代码
	PointCode  string    `json:"point_code" gorm:"not null;size:2"`        // 后2位: 具体位置代码
	
	// 类型和属性
	PointType     string `json:"point_type" gorm:"not null;size:20"`        // 类型: dormitory/shop/box/club
	PointName     string `json:"point_name" gorm:"size:100"`                // 位置名称
	FullAddress   string `json:"full_address" gorm:"size:200"`              // 完整地址描述
	IsPublic      bool   `json:"is_public" gorm:"default:false"`            // 后两位是否公开
	IsActive      bool   `json:"is_active" gorm:"default:true"`             // 是否激活
	
	// 绑定信息
	BindingType   string    `json:"binding_type" gorm:"size:20"`               // 绑定类型: user/shop/public
	BindingID     *string   `json:"binding_id,omitempty"`                      // 绑定对象ID
	BindingStatus string    `json:"binding_status" gorm:"default:'pending'"`   // 绑定状态: pending/approved/rejected
	
	// 管理信息
	ManagedBy     string    `json:"managed_by" gorm:"not null"`                // 管理者ID (二级信使)
	ApprovedBy    *string   `json:"approved_by,omitempty"`                     // 审核者ID
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`                    // 审核时间
	
	// 使用统计
	UsageCount    int       `json:"usage_count" gorm:"default:0"`              // 使用次数
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`                   // 最后使用时间
	
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// SignalCode 为了向后兼容，SignalCode是OPCode的别名
// 这样可以让现有代码无缝使用新的OP Code系统
type SignalCode = OPCode

// 第一个OPCodeApplication定义已删除，使用下面更完整的定义

// 常量定义移至下方，避免重复

// OPCodeRequest定义移至下方，使用更完整的版本

// OPCodeSchool 学校编码映射表
type OPCodeSchool struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	SchoolCode   string    `json:"school_code" gorm:"unique;not null;size:2"` // 2位学校代码
	SchoolName   string    `json:"school_name" gorm:"not null;size:100"`
	FullName     string    `json:"full_name" gorm:"size:200"`
	City         string    `json:"city" gorm:"size:50"`
	Province     string    `json:"province" gorm:"size:50"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	ManagedBy    string    `json:"managed_by"`                                 // 四级信使ID
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OPCodeArea 片区编码映射表
type OPCodeArea struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	SchoolCode   string    `json:"school_code" gorm:"not null;size:2;index"`
	AreaCode     string    `json:"area_code" gorm:"not null;size:2"`
	AreaName     string    `json:"area_name" gorm:"not null;size:100"`
	Description  string    `json:"description" gorm:"size:200"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	ManagedBy    string    `json:"managed_by"`                                 // 三级信使ID
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// 联合唯一索引
	UniqueIndex  string    `gorm:"uniqueIndex:idx_school_area,unique"`
}

// OPCodeApplication OP Code申请记录
type OPCodeApplication struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID        string    `json:"user_id" gorm:"not null;index"`
	RequestedCode string    `json:"requested_code" gorm:"size:6"`              // 申请的完整编码
	SchoolCode    string    `json:"school_code" gorm:"not null;size:2"`
	AreaCode      string    `json:"area_code" gorm:"not null;size:2"`
	PointType     string    `json:"point_type" gorm:"not null;size:20"`
	PointName     string    `json:"point_name" gorm:"size:100"`
	FullAddress   string    `json:"full_address" gorm:"size:200"`
	Reason        string    `json:"reason" gorm:"type:text"`
	Evidence      string    `json:"evidence" gorm:"type:json"`                 // 证明材料JSON
	
	Status        string    `json:"status" gorm:"default:'pending'"`           // pending/approved/rejected
	AssignedCode  string    `json:"assigned_code" gorm:"size:6"`               // 最终分配的编码
	ReviewerID    *string   `json:"reviewer_id,omitempty"`
	ReviewComment string    `json:"review_comment" gorm:"type:text"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// OPCodePermission OP Code权限表
type OPCodePermission struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	CourierID    string    `json:"courier_id" gorm:"not null;index"`
	CourierLevel int       `json:"courier_level" gorm:"not null"`
	CodePrefix   string    `json:"code_prefix" gorm:"not null;size:6;index"`  // 管理的编码前缀
	Permission   string    `json:"permission" gorm:"not null;size:20"`        // view/assign/approve
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// 常量定义
const (
	// 点位类型
	OPCodeTypeDormitory = "dormitory" // 宿舍
	OPCodeTypeShop      = "shop"      // 商店
	OPCodeTypeBox       = "box"       // 投递箱
	OPCodeTypeClub      = "club"      // 社团空间
	
	// 绑定类型
	OPCodeBindingUser   = "user"      // 用户绑定
	OPCodeBindingShop   = "shop"      // 商店绑定
	OPCodeBindingPublic = "public"    // 公共点位
	
	// 状态
	OPCodeStatusPending  = "pending"   // 待审核
	OPCodeStatusApproved = "approved"  // 已批准
	OPCodeStatusRejected = "rejected"  // 已拒绝
	
	// 权限类型
	OPCodePermissionView    = "view"    // 查看权限
	OPCodePermissionAssign  = "assign"  // 分配权限
	OPCodePermissionApprove = "approve" // 审批权限
)

// TableName 指定表名
func (OPCode) TableName() string {
	return "op_codes"
}

func (OPCodeSchool) TableName() string {
	return "op_code_schools"
}

func (OPCodeArea) TableName() string {
	return "op_code_areas"
}

func (OPCodeApplication) TableName() string {
	return "op_code_applications"
}

func (OPCodePermission) TableName() string {
	return "op_code_permissions"
}

// ValidateOPCode 验证OP Code格式
func ValidateOPCode(code string) error {
	if len(code) != 6 {
		return fmt.Errorf("OP Code必须为6位")
	}
	
	// 验证格式：2位大写字母/数字 + 2位字母/数字 + 2位字母/数字
	pattern := `^[A-Z0-9]{2}[A-Z0-9]{2}[A-Z0-9]{2}$`
	matched, _ := regexp.MatchString(pattern, strings.ToUpper(code))
	if !matched {
		return fmt.Errorf("OP Code格式不正确，应为6位大写字母或数字组合")
	}
	
	return nil
}

// ParseOPCode 解析OP Code
func ParseOPCode(code string) (schoolCode, areaCode, pointCode string, err error) {
	if err = ValidateOPCode(code); err != nil {
		return "", "", "", err
	}
	
	code = strings.ToUpper(code)
	return code[:2], code[2:4], code[4:6], nil
}

// FormatOPCode 格式化OP Code显示
func FormatOPCode(code string, hidePrivate bool) string {
	if err := ValidateOPCode(code); err != nil {
		return code
	}
	
	code = strings.ToUpper(code)
	if hidePrivate {
		// 隐藏后两位
		return code[:4] + "**"
	}
	return code
}

// GetOPCodePrefix 获取OP Code前缀（用于权限控制）
func GetOPCodePrefix(code string, level int) string {
	if err := ValidateOPCode(code); err != nil {
		return ""
	}
	
	code = strings.ToUpper(code)
	switch level {
	case 4: // 四级信使 - 管理学校级别
		return code[:2] + "****"
	case 3: // 三级信使 - 管理片区级别
		return code[:4] + "**"
	case 2: // 二级信使 - 管理具体位置
		return code
	default:
		return ""
	}
}

// CanManageOPCode 检查是否有权限管理某个OP Code
func CanManageOPCode(managerPrefix, targetCode string) bool {
	if err := ValidateOPCode(targetCode); err != nil {
		return false
	}
	
	// 去除通配符
	prefix := strings.ReplaceAll(managerPrefix, "*", "")
	
	// 检查前缀匹配
	return strings.HasPrefix(strings.ToUpper(targetCode), prefix)
}

// OPCodeRequest OP Code申请请求
type OPCodeRequest struct {
	SchoolCode  string                 `json:"school_code" binding:"required,len=2"`
	AreaCode    string                 `json:"area_code" binding:"required,len=2"`
	PointType   string                 `json:"point_type" binding:"required,oneof=dormitory shop box club"`
	PointName   string                 `json:"point_name" binding:"required,min=2,max=100"`
	FullAddress string                 `json:"full_address" binding:"required,min=5,max=200"`
	Reason      string                 `json:"reason" binding:"required,min=10,max=500"`
	Evidence    map[string]interface{} `json:"evidence"`
	IsPublic    bool                   `json:"is_public"`
}

// OPCodeAssignRequest OP Code分配请求
type OPCodeAssignRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
	PointCode     string `json:"point_code" binding:"required,len=2"`
	Comment       string `json:"comment"`
}

// OPCodeSearchRequest OP Code搜索请求
type OPCodeSearchRequest struct {
	Code       string `form:"code"`
	SchoolCode string `form:"school_code"`
	AreaCode   string `form:"area_code"`
	PointType  string `form:"point_type"`
	IsPublic   *bool  `form:"is_public"`
	IsActive   *bool  `form:"is_active"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

// OPCodeStats OP Code统计信息
type OPCodeStats struct {
	SchoolCode      string         `json:"school_code"`
	SchoolName      string         `json:"school_name"`
	TotalCodes      int64          `json:"total_codes"`
	ActiveCodes     int64          `json:"active_codes"`
	PublicCodes     int64          `json:"public_codes"`
	ByType          map[string]int `json:"by_type"`
	ByArea          map[string]int `json:"by_area"`
	UtilizationRate float64        `json:"utilization_rate"`
}