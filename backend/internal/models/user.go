package models

import (
	"gorm.io/gorm"
	"time"
)

// UserRole 用户角色枚举
type UserRole string

const (
	RoleUser               UserRole = "user"                // 普通用户
	RoleCourier            UserRole = "courier"             // 普通信使
	RoleSeniorCourier      UserRole = "senior_courier"      // 高级信使
	RoleCourierCoordinator UserRole = "courier_coordinator" // 信使协调员
	RoleSchoolAdmin        UserRole = "school_admin"        // 学校管理员
	RolePlatformAdmin      UserRole = "platform_admin"      // 平台管理员
	RoleSuperAdmin         UserRole = "super_admin"         // 超级管理员

	// 分级信使系统 (兼容性)
	RoleCourierLevel1 UserRole = "courier_level1" // 一级信使
	RoleCourierLevel2 UserRole = "courier_level2" // 二级信使
	RoleCourierLevel3 UserRole = "courier_level3" // 三级信使
	RoleCourierLevel4 UserRole = "courier_level4" // 四级信使
)

// String 返回角色字符串
func (r UserRole) String() string {
	return string(r)
}

// RoleHierarchy 角色层级（数字越大权限越高）
var RoleHierarchy = map[UserRole]int{
	RoleUser:               1,
	RoleCourier:            2,
	RoleSeniorCourier:      3,
	RoleCourierCoordinator: 4,
	RoleSchoolAdmin:        5,
	RolePlatformAdmin:      6,
	RoleSuperAdmin:         7,

	// 分级信使系统映射
	RoleCourierLevel1: 2,
	RoleCourierLevel2: 3,
	RoleCourierLevel3: 4,
	RoleCourierLevel4: 5,
}

// Permission 权限类型
type Permission string

const (
	// 用户权限
	PermissionWriteLetter   Permission = "write_letter"
	PermissionReadLetter    Permission = "read_letter"
	PermissionManageProfile Permission = "manage_profile"

	// 信使权限
	PermissionDeliverLetter Permission = "deliver_letter"
	PermissionScanCode      Permission = "scan_code"
	PermissionViewTasks     Permission = "view_tasks"

	// 协调员权限
	PermissionManageCouriers Permission = "manage_couriers"
	PermissionAssignTasks    Permission = "assign_tasks"
	PermissionViewReports    Permission = "view_reports"

	// 管理员权限
	PermissionManageUsers   Permission = "manage_users"
	PermissionManageSchool  Permission = "manage_school"
	PermissionViewAnalytics Permission = "view_analytics"
	PermissionManageSystem  Permission = "manage_system"

	// 超级管理员权限
	PermissionManagePlatform Permission = "manage_platform"
	PermissionManageAdmins   Permission = "manage_admins"
	PermissionSystemConfig   Permission = "system_config"
)

// RolePermissions 角色权限映射
var RolePermissions = map[UserRole][]Permission{
	RoleUser: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
	},
	RoleCourier: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
	},
	RoleSeniorCourier: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionViewReports,
	},
	RoleCourierCoordinator: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionManageCouriers,
		PermissionAssignTasks,
		PermissionViewReports,
	},
	RoleSchoolAdmin: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionManageUsers,
		PermissionManageCouriers,
		PermissionAssignTasks,
		PermissionViewReports,
		PermissionManageSchool,
		PermissionViewAnalytics,
	},
	RolePlatformAdmin: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionManageUsers,
		PermissionManageCouriers,
		PermissionAssignTasks,
		PermissionViewReports,
		PermissionManageSchool,
		PermissionViewAnalytics,
		PermissionManageSystem,
	},
	RoleSuperAdmin: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionManageUsers,
		PermissionManageCouriers,
		PermissionAssignTasks,
		PermissionViewReports,
		PermissionManageSchool,
		PermissionViewAnalytics,
		PermissionManageSystem,
		PermissionManagePlatform,
		PermissionManageAdmins,
		PermissionSystemConfig,
	},

	// 分级信使系统权限映射 (兼容性)
	RoleCourierLevel1: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
	},
	RoleCourierLevel2: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionViewReports,
	},
	RoleCourierLevel3: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionManageCouriers,
		PermissionAssignTasks,
		PermissionViewReports,
	},
	RoleCourierLevel4: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionManageCouriers,
		PermissionAssignTasks,
		PermissionViewReports,
		PermissionManageSchool,
	},
}

// User 用户模型
type User struct {
	ID           string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Username     string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Email        string         `json:"email" gorm:"type:varchar(100);uniqueIndex"`
	PasswordHash string         `json:"-" gorm:"type:varchar(255);not null"`
	Nickname     string         `json:"nickname" gorm:"type:varchar(50)"`
	Avatar       string         `json:"avatar" gorm:"type:varchar(500)"`
	Role         UserRole       `json:"role" gorm:"type:varchar(20);not null;default:'user'"`
	SchoolCode   string         `json:"school_code" gorm:"type:varchar(20);index"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	SentLetters     []Letter `json:"sent_letters,omitempty" gorm:"foreignKey:UserID"`
	AuthoredLetters []Letter `json:"authored_letters,omitempty" gorm:"foreignKey:AuthorID"`
}

// UserProfile 用户档案
type UserProfile struct {
	UserID      string    `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	RealName    string    `json:"real_name" gorm:"type:varchar(50)"`
	Phone       string    `json:"phone" gorm:"type:varchar(20)"`
	Address     string    `json:"address" gorm:"type:text"`
	Bio         string    `json:"bio" gorm:"type:text"`
	Preferences string    `json:"preferences" gorm:"type:json"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=50"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Nickname   string `json:"nickname" binding:"required,min=1,max=50"`
	SchoolCode string `json:"school_code" binding:"required,len=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	User      *User     `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UpdateProfileRequest 更新档案请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	Address  string `json:"address"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UserStats 用户统计
type UserStats struct {
	LettersSent     int64 `json:"letters_sent"`
	LettersReceived int64 `json:"letters_received"`
	DraftsCount     int64 `json:"drafts_count"`
	DeliveredCount  int64 `json:"delivered_count"`
}

// HasPermission 检查用户是否有指定权限
func (u *User) HasPermission(permission Permission) bool {
	permissions, exists := RolePermissions[u.Role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasRole 检查用户是否有指定角色或更高权限
func (u *User) HasRole(role UserRole) bool {
	userLevel, exists := RoleHierarchy[u.Role]
	if !exists {
		return false
	}

	requiredLevel, exists := RoleHierarchy[role]
	if !exists {
		return false
	}

	return userLevel >= requiredLevel
}

// GetRoleDisplayName 获取角色显示名称
func (u *User) GetRoleDisplayName() string {
	roleNames := map[UserRole]string{
		RoleUser:               "普通用户",
		RoleCourier:            "信使",
		RoleSeniorCourier:      "高级信使",
		RoleCourierCoordinator: "信使协调员",
		RoleSchoolAdmin:        "学校管理员",
		RolePlatformAdmin:      "平台管理员",
		RoleSuperAdmin:         "超级管理员",

		// 分级信使系统显示名称
		RoleCourierLevel1: "一级信使",
		RoleCourierLevel2: "二级信使",
		RoleCourierLevel3: "三级信使",
		RoleCourierLevel4: "四级信使",
	}

	if name, exists := roleNames[u.Role]; exists {
		return name
	}
	return "未知角色"
}

// CanManageUser 检查是否可以管理指定用户
func (u *User) CanManageUser(targetUser *User) bool {
	// 不能管理自己
	if u.ID == targetUser.ID {
		return false
	}

	// 检查权限层级
	userLevel := RoleHierarchy[u.Role]
	targetLevel := RoleHierarchy[targetUser.Role]

	return userLevel > targetLevel
}
