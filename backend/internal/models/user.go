package models

import (
	"gorm.io/gorm"
	"time"
)

// UserRole 用户角色枚举
type UserRole string

const (
	// 基础角色
	RoleUser UserRole = "user" // 普通用户

	// 四级信使体系 - 符合PRD设计
	RoleCourierLevel1 UserRole = "courier_level1" // 一级信使（基础投递信使）
	RoleCourierLevel2 UserRole = "courier_level2" // 二级信使（片区协调员）
	RoleCourierLevel3 UserRole = "courier_level3" // 三级信使（校区负责人）
	RoleCourierLevel4 UserRole = "courier_level4" // 四级信使（城市负责人）

	// 管理角色
	RolePlatformAdmin UserRole = "platform_admin" // 平台管理员
	RoleSuperAdmin    UserRole = "super_admin"    // 超级管理员
)

// String 返回角色字符串
func (r UserRole) String() string {
	return string(r)
}

// RoleHierarchy 角色层级（数字越大权限越高）
var RoleHierarchy = map[UserRole]int{
	RoleUser:          1,
	RoleCourierLevel1: 2,
	RoleCourierLevel2: 3,
	RoleCourierLevel3: 4,
	RoleCourierLevel4: 5,
	RolePlatformAdmin: 6,
	RoleSuperAdmin:    7,
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

// RolePermissions 角色权限映射 - 简化版本符合PRD
var RolePermissions = map[UserRole][]Permission{
	RoleUser: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
	},

	// 一级信使：基础投递
	RoleCourierLevel1: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
	},

	// 二级信使：片区协调员 - 可以分发任务给一级信使
	RoleCourierLevel2: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionAssignTasks, // 分发任务给一级信使
	},

	// 三级信使：校区负责人 - 可以任命二级信使，查看报告
	RoleCourierLevel3: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionAssignTasks,
		PermissionManageCouriers, // 任命二级信使
		PermissionViewReports,    // 查看报告
	},

	// 四级信使：城市负责人 - 拥有城市级权限
	RoleCourierLevel4: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionAssignTasks,
		PermissionManageCouriers,
		PermissionViewReports,
		PermissionManageSchool,  // 开通新学校
		PermissionViewAnalytics, // 查看城市级数据
	},

	// 平台管理员
	RolePlatformAdmin: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionManageUsers,
		PermissionViewReports,
		PermissionViewAnalytics,
		PermissionManageSystem,
	},

	// 超级管理员 - 拥有所有权限
	RoleSuperAdmin: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionManageCouriers,
		PermissionAssignTasks,
		PermissionViewReports,
		PermissionManageUsers,
		PermissionManageSchool,
		PermissionViewAnalytics,
		PermissionManageSystem,
		PermissionManagePlatform,
		PermissionManageAdmins,
		PermissionSystemConfig,
	},
}

// User 用户模型
type User struct {
	ID           string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Username     string         `json:"username" gorm:"type:varchar(50);uniqueIndex:unique_username;not null"`
	Email        string         `json:"email" gorm:"type:varchar(100);uniqueIndex:unique_email"`
	PasswordHash string         `json:"-" gorm:"type:varchar(255);not null"`
	Nickname     string         `json:"nickname" gorm:"type:varchar(50)"`
	Avatar       string         `json:"avatar" gorm:"type:varchar(500)"`
	Role         UserRole       `json:"role" gorm:"type:varchar(20);not null;default:'user'"`
	SchoolCode   string         `json:"school_code" gorm:"type:varchar(20);index"`
	OPCode       string         `json:"op_code" gorm:"type:varchar(6);index"` // OP Code地址
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
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"` // 刷新令牌，与前端类型保持一致
	User         *User     `json:"user"`
	ExpiresAt    time.Time `json:"expires_at"`
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

// AdminUpdateUserRequest 管理员更新用户请求
type AdminUpdateUserRequest struct {
	Nickname   string `json:"nickname"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	SchoolCode string `json:"school_code"`
	IsActive   bool   `json:"is_active"`
}

// UserStats 用户统计
type UserStats struct {
	UserID         string    `json:"user_id"`
	LettersSent     int64 `json:"letters_sent"`
	LettersReceived int64 `json:"letters_received"`
	DraftsCount     int64 `json:"drafts_count"`
	DeliveredCount  int64 `json:"delivered_count"`
	FollowingCount int64 `json:"following_count"`
	FollowersCount int64 `json:"followers_count"`
	MutualCount    int64 `json:"mutual_count"`
	LastActive     time.Time `json:"last_active"`
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
		RoleUser:          "普通用户",
		RoleCourierLevel1: "一级信使（基础投递）",
		RoleCourierLevel2: "二级信使（片区协调员）",
		RoleCourierLevel3: "三级信使（校区负责人）",
		RoleCourierLevel4: "四级信使（城市负责人）",
		RolePlatformAdmin: "平台管理员",
		RoleSuperAdmin:    "超级管理员",
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
