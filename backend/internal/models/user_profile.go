package models

import (
	"time"
	"gorm.io/gorm"
)

// UserProfileExtended 扩展用户档案信息
type UserProfileExtended struct {
	UserID       string    `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	Bio          string    `json:"bio" gorm:"type:text"`
	School       string    `json:"school" gorm:"type:varchar(100)"`
	OPCode       string    `json:"op_code" gorm:"type:varchar(6);index"`
	WritingLevel int       `json:"writing_level" gorm:"default:1;check:writing_level >= 0 AND writing_level <= 5"`
	CourierLevel int       `json:"courier_level" gorm:"default:0;check:courier_level >= 0 AND courier_level <= 4"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联
	User         User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Stats        UserStatsData `json:"stats,omitempty" gorm:"foreignKey:UserID"`
	Privacy      UserPrivacy   `json:"privacy,omitempty" gorm:"foreignKey:UserID"`
	Achievements []UserAchievement `json:"achievements,omitempty" gorm:"foreignKey:UserID"`
}

// UserStatsData 用户统计数据
type UserStatsData struct {
	UserID              string    `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	LettersSent         int       `json:"letters_sent" gorm:"default:0"`
	LettersReceived     int       `json:"letters_received" gorm:"default:0"`
	MuseumContributions int       `json:"museum_contributions" gorm:"default:0"`
	TotalPoints         int       `json:"total_points" gorm:"default:0"`
	WritingPoints       int       `json:"writing_points" gorm:"default:0"`
	CourierPoints       int       `json:"courier_points" gorm:"default:0"`
	CurrentStreak       int       `json:"current_streak" gorm:"default:0"`
	MaxStreak           int       `json:"max_streak" gorm:"default:0"`
	LastActiveDate      time.Time `json:"last_active_date"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// UserPrivacy 用户隐私设置
type UserPrivacy struct {
	UserID         string             `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	ShowEmail      bool               `json:"show_email" gorm:"default:false"`
	ShowOPCode     bool               `json:"show_op_code" gorm:"default:true"`
	ShowStats      bool               `json:"show_stats" gorm:"default:true"`
	OPCodePrivacy  OPCodePrivacyLevel `json:"op_code_privacy" gorm:"type:varchar(20);default:'partial'"`
	ProfileVisible bool               `json:"profile_visible" gorm:"default:true"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// OPCodePrivacyLevel OP Code 隐私级别
type OPCodePrivacyLevel string

const (
	OPCodePrivacyFull    OPCodePrivacyLevel = "full"    // 显示完整编码
	OPCodePrivacyPartial OPCodePrivacyLevel = "partial" // 显示部分编码 (如 PK5F**)
	OPCodePrivacyHidden  OPCodePrivacyLevel = "hidden"  // 完全隐藏
)

// UserAchievement 用户成就
type UserAchievement struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"type:varchar(36);index"`
	Code        string    `json:"code" gorm:"type:varchar(50);uniqueIndex:idx_user_achievement"`
	Name        string    `json:"name" gorm:"type:varchar(100)"`
	Description string    `json:"description" gorm:"type:text"`
	Icon        string    `json:"icon" gorm:"type:varchar(50)"`
	Category    string    `json:"category" gorm:"type:varchar(50)"`
	UnlockedAt  time.Time `json:"unlocked_at"`
}

// AchievementDefinition 成就定义
type AchievementDefinition struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Category    string `json:"category"`
}

// 预定义成就
var Achievements = []AchievementDefinition{
	// 写信相关
	{Code: "first_letter", Name: "初次来信", Description: "发送第一封信", Icon: "✉️", Category: "writing"},
	{Code: "active_writer", Name: "活跃写手", Description: "发送10封信", Icon: "✍️", Category: "writing"},
	{Code: "prolific_writer", Name: "多产作家", Description: "发送50封信", Icon: "📚", Category: "writing"},
	{Code: "letter_master", Name: "信件大师", Description: "发送100封信", Icon: "🏆", Category: "writing"},
	
	// 博物馆相关
	{Code: "museum_contributor", Name: "博物馆贡献者", Description: "贡献第一封信到博物馆", Icon: "🏛️", Category: "museum"},
	{Code: "museum_curator", Name: "博物馆策展人", Description: "贡献10封信到博物馆", Icon: "🎨", Category: "museum"},
	
	// 社交相关
	{Code: "popular_writer", Name: "人气写手", Description: "收到100个赞", Icon: "⭐", Category: "social"},
	{Code: "social_butterfly", Name: "社交达人", Description: "与20个不同的人通信", Icon: "🦋", Category: "social"},
	
	// 信使相关
	{Code: "rookie_courier", Name: "新手信使", Description: "成为一级信使", Icon: "🎒", Category: "courier"},
	{Code: "area_coordinator", Name: "片区协调员", Description: "成为二级信使", Icon: "📍", Category: "courier"},
	{Code: "school_leader", Name: "校区负责人", Description: "成为三级信使", Icon: "🏫", Category: "courier"},
	{Code: "city_coordinator", Name: "城市总代", Description: "成为四级信使", Icon: "🌆", Category: "courier"},
	
	// 系统相关
	{Code: "early_bird", Name: "早期用户", Description: "平台前1000名用户", Icon: "🐦", Category: "system"},
	{Code: "beta_tester", Name: "测试先锋", Description: "参与测试阶段", Icon: "🧪", Category: "system"},
	{Code: "bug_reporter", Name: "问题猎手", Description: "报告有效bug", Icon: "🐛", Category: "system"},
	
	// 活跃度相关
	{Code: "week_streak", Name: "周连续", Description: "连续活跃7天", Icon: "🔥", Category: "activity"},
	{Code: "month_streak", Name: "月连续", Description: "连续活跃30天", Icon: "💫", Category: "activity"},
	{Code: "year_member", Name: "年度会员", Description: "注册满一年", Icon: "🎂", Category: "activity"},
}

// GetWritingLevelName 获取写信等级名称
func GetWritingLevelName(level int) string {
	levelNames := map[int]string{
		1: "新手写手",
		2: "熟练写手",
		3: "优秀写手",
		4: "资深写手",
		5: "大师写手",
	}
	if name, exists := levelNames[level]; exists {
		return name
	}
	return "未知等级"
}

// GetCourierLevelName 获取信使等级名称
func GetCourierLevelName(level int) string {
	levelNames := map[int]string{
		0: "非信使",
		1: "楼栋信使",
		2: "片区信使",
		3: "校级信使",
		4: "城市总代",
	}
	if name, exists := levelNames[level]; exists {
		return name
	}
	return "未知等级"
}

// TableName 指定表名
func (UserProfileExtended) TableName() string {
	return "user_profiles_extended"
}

func (UserStatsData) TableName() string {
	return "user_stats"
}

func (UserPrivacy) TableName() string {
	return "user_privacy_settings"
}

func (UserAchievement) TableName() string {
	return "user_achievements"
}

// GetFormattedOPCode 获取格式化的OP Code（根据隐私设置）
func (u *UserProfileExtended) GetFormattedOPCode(privacy *UserPrivacy) string {
	if privacy == nil || !privacy.ShowOPCode || u.OPCode == "" {
		return ""
	}
	
	switch privacy.OPCodePrivacy {
	case OPCodePrivacyFull:
		return u.OPCode
	case OPCodePrivacyPartial:
		if len(u.OPCode) >= 4 {
			return u.OPCode[:4] + "**"
		}
		return u.OPCode
	case OPCodePrivacyHidden:
		return ""
	default:
		return u.OPCode[:4] + "**"
	}
}

// CanLevelUp 检查是否可以升级
func (s *UserStatsData) CanLevelUp(currentLevel int, isWriting bool) bool {
	if isWriting {
		// 写信等级升级条件
		levelRequirements := map[int]int{
			1: 100,   // 升到2级需要100积分
			2: 300,   // 升到3级需要300积分
			3: 600,   // 升到4级需要600积分
			4: 1000,  // 升到5级需要1000积分
		}
		if req, exists := levelRequirements[currentLevel]; exists {
			return s.WritingPoints >= req
		}
	} else {
		// 信使等级升级条件
		levelRequirements := map[int]int{
			0: 50,    // 成为1级信使需要50积分
			1: 200,   // 升到2级需要200积分
			2: 500,   // 升到3级需要500积分
			3: 1000,  // 升到4级需要1000积分
		}
		if req, exists := levelRequirements[currentLevel]; exists {
			return s.CourierPoints >= req
		}
	}
	return false
}

// BeforeCreate GORM钩子
func (u *UserProfileExtended) BeforeCreate(tx *gorm.DB) error {
	// 设置默认值
	if u.WritingLevel == 0 {
		u.WritingLevel = 1
	}
	return nil
}

// BeforeCreate GORM钩子
func (s *UserStatsData) BeforeCreate(tx *gorm.DB) error {
	s.LastActiveDate = time.Now()
	return nil
}

// UpdateStreak 更新连续天数
func (s *UserStatsData) UpdateStreak() {
	now := time.Now()
	lastActive := s.LastActiveDate
	
	// 计算天数差
	daysSince := int(now.Sub(lastActive).Hours() / 24)
	
	if daysSince == 0 {
		// 同一天，不更新
		return
	} else if daysSince == 1 {
		// 连续一天
		s.CurrentStreak++
		if s.CurrentStreak > s.MaxStreak {
			s.MaxStreak = s.CurrentStreak
		}
	} else {
		// 中断了，重置
		s.CurrentStreak = 1
	}
	
	s.LastActiveDate = now
}