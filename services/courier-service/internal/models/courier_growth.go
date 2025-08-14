package models

import (
	"time"
)

// CourierGrowthPath 成长路径配置模型
type CourierGrowthPath struct {
	ID           uint         `json:"id" gorm:"primaryKey"`
	FromLevel    CourierLevel `json:"from_level" gorm:"not null"`
	ToLevel      CourierLevel `json:"to_level" gorm:"not null"`
	Name         string       `json:"name" gorm:"not null"`
	Description  string       `json:"description"`
	Requirements string       `json:"requirements" gorm:"type:json"` // JSON格式的升级要求
	IsActive     bool         `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// CourierIncentive 激励奖励模型
type CourierIncentive struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	Type        IncentiveType `json:"type" gorm:"not null"`
	Name        string        `json:"name" gorm:"not null"`
	Description string        `json:"description"`
	Value       float64       `json:"value"`                       // 奖励金额或积分
	Conditions  string        `json:"conditions" gorm:"type:json"` // JSON格式的获得条件
	ValidFrom   time.Time     `json:"valid_from"`
	ValidTo     *time.Time    `json:"valid_to,omitempty"`
	IsActive    bool          `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// CourierStatistics 信使任务统计模型
type CourierStatistics struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	CourierID         string    `json:"courier_id" gorm:"not null;uniqueIndex:idx_courier_date"`
	Date              time.Time `json:"date" gorm:"not null;uniqueIndex:idx_courier_date"`
	TasksAccepted     int       `json:"tasks_accepted" gorm:"default:0"`
	TasksCompleted    int       `json:"tasks_completed" gorm:"default:0"`
	TasksFailed       int       `json:"tasks_failed" gorm:"default:0"`
	TotalDeliveryTime int       `json:"total_delivery_time" gorm:"default:0"` // 总投递时间(分钟)
	CompletionRate    float64   `json:"completion_rate" gorm:"default:0"`
	AverageRating     float64   `json:"average_rating" gorm:"default:0"`
	DistanceTraveled  float64   `json:"distance_traveled" gorm:"default:0"` // 总里程(km)
	EarningsAmount    float64   `json:"earnings_amount" gorm:"default:0"`   // 收入金额
	PointsEarned      int       `json:"points_earned" gorm:"default:0"`     // 获得积分
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CourierBadge 徽章系统模型
type CourierBadge struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	Code        string      `json:"code" gorm:"not null;unique"`
	Name        string      `json:"name" gorm:"not null"`
	Description string      `json:"description"`
	IconURL     string      `json:"icon_url"`
	Conditions  string      `json:"conditions" gorm:"type:json"` // JSON格式的获得条件
	Rarity      BadgeRarity `json:"rarity" gorm:"default:common"`
	Points      int         `json:"points" gorm:"default:0"` // 徽章价值积分
	IsActive    bool        `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// CourierPoints 积分系统模型
type CourierPoints struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CourierID string    `json:"courier_id" gorm:"not null;index"`
	Total     int       `json:"total" gorm:"default:0"`     // 总积分
	Available int       `json:"available" gorm:"default:0"` // 可用积分
	Used      int       `json:"used" gorm:"default:0"`      // 已使用积分
	Earned    int       `json:"earned" gorm:"default:0"`    // 累计获得积分
	UpdatedAt time.Time `json:"updated_at"`
}

// CourierPointsTransaction 积分交易记录
type CourierPointsTransaction struct {
	ID          uint                  `json:"id" gorm:"primaryKey"`
	CourierID   string                `json:"courier_id" gorm:"not null;index"`
	Type        PointsTransactionType `json:"type" gorm:"not null"`
	Amount      int                   `json:"amount" gorm:"not null"`
	Description string                `json:"description"`
	Reference   string                `json:"reference"` // 关联的任务或活动ID
	CreatedAt   time.Time             `json:"created_at"`
}

// CourierBadgeEarned 信使获得的徽章记录
type CourierBadgeEarned struct {
	ID        uint         `json:"id" gorm:"primaryKey"`
	CourierID string       `json:"courier_id" gorm:"not null;index"`
	BadgeID   uint         `json:"badge_id" gorm:"not null"`
	Badge     CourierBadge `json:"badge" gorm:"foreignKey:BadgeID"`
	EarnedAt  time.Time    `json:"earned_at"`
	Reason    string       `json:"reason"`    // 获得原因
	Reference string       `json:"reference"` // 关联的任务或活动
	CreatedAt time.Time    `json:"created_at"`
}

// IncentiveType 激励类型
type IncentiveType string

const (
	IncentiveTypeSubsidy    IncentiveType = "subsidy"    // 投递补贴
	IncentiveTypePoints     IncentiveType = "points"     // 积分奖励
	IncentiveTypeCommission IncentiveType = "commission" // 返佣
	IncentiveTypeBadge      IncentiveType = "badge"      // 徽章奖励
	IncentiveTypeBonus      IncentiveType = "bonus"      // 特殊奖金
)

// BadgeRarity 徽章稀有度
type BadgeRarity string

const (
	BadgeRarityCommon    BadgeRarity = "common"    // 普通
	BadgeRarityUncommon  BadgeRarity = "uncommon"  // 稀有
	BadgeRarityRare      BadgeRarity = "rare"      // 珍稀
	BadgeRarityEpic      BadgeRarity = "epic"      // 史诗
	BadgeRarityLegendary BadgeRarity = "legendary" // 传说
)

// PointsTransactionType 积分交易类型
type PointsTransactionType string

const (
	PointsEarn   PointsTransactionType = "earn"   // 获得积分
	PointsSpend  PointsTransactionType = "spend"  // 消费积分
	PointsRefund PointsTransactionType = "refund" // 退还积分
	PointsExpire PointsTransactionType = "expire" // 积分过期
)

// GrowthRequirement 成长要求结构
type GrowthRequirement struct {
	Type        string      `json:"type"`        // 要求类型
	Description string      `json:"description"` // 描述
	Target      interface{} `json:"target"`      // 目标值
	Current     interface{} `json:"current"`     // 当前值
	Completed   bool        `json:"completed"`   // 是否完成
}

// CourierGrowthProgress 信使成长进度
type CourierGrowthProgress struct {
	CourierID       string              `json:"courier_id"`
	CurrentLevel    CourierLevel        `json:"current_level"`
	NextLevel       *CourierLevel       `json:"next_level,omitempty"`
	CanUpgrade      bool                `json:"can_upgrade"`
	Requirements    []GrowthRequirement `json:"requirements"`
	CompletionRate  float64             `json:"completion_rate"`
	TotalPoints     int                 `json:"total_points"`
	AvailablePoints int                 `json:"available_points"`
	BadgesEarned    int                 `json:"badges_earned"`
	LastUpdated     time.Time           `json:"last_updated"`
}

// CourierRankingInfo 信使排行榜信息
type CourierRankingInfo struct {
	CourierID      string  `json:"courier_id"`
	Name           string  `json:"name"`
	Level          int     `json:"level"`
	Points         int     `json:"points"`
	TasksCompleted int     `json:"tasks_completed"`
	CompletionRate float64 `json:"completion_rate"`
	Rank           int     `json:"rank"`
	ZoneType       string  `json:"zone_type"`
	ZoneName       string  `json:"zone_name"`
}

// DefaultGrowthRequirements 默认成长要求配置
var DefaultGrowthRequirements = map[CourierLevel][]GrowthRequirement{
	LevelTwo: {
		{
			Type:        "delivery_count",
			Description: "累计投递10封信件",
			Target:      10,
			Completed:   false,
		},
		{
			Type:        "consecutive_days",
			Description: "连续7天有投递记录",
			Target:      7,
			Completed:   false,
		},
	},
	LevelThree: {
		{
			Type:        "manage_couriers",
			Description: "管理至少3位一级信使",
			Target:      3,
			Completed:   false,
		},
		{
			Type:        "completion_rate",
			Description: "月任务完成率超过80%",
			Target:      80.0,
			Completed:   false,
		},
		{
			Type:        "experience_months",
			Description: "具备组织经验",
			Target:      true,
			Completed:   false,
		},
	},
	LevelFour: {
		{
			Type:        "service_duration",
			Description: "持续服务3个月以上",
			Target:      3,
			Completed:   false,
		},
		{
			Type:        "school_recommendation",
			Description: "获得校级推荐",
			Target:      true,
			Completed:   false,
		},
		{
			Type:        "platform_approval",
			Description: "通过平台备案审核",
			Target:      true,
			Completed:   false,
		},
	},
}

// DefaultBadges 默认徽章配置
var DefaultBadges = []CourierBadge{
	{
		Code:        "beautiful_courier",
		Name:        "最美信使",
		Description: "连续一个月保持5星评分",
		Rarity:      BadgeRarityRare,
		Points:      100,
		Conditions:  `{"type":"rating","value":5.0,"duration":"30d"}`,
	},
	{
		Code:        "monthly_pioneer",
		Name:        "当月先锋",
		Description: "当月投递量排名前10%",
		Rarity:      BadgeRarityUncommon,
		Points:      50,
		Conditions:  `{"type":"ranking","value":0.1,"period":"monthly"}`,
	},
	{
		Code:        "speed_master",
		Name:        "速度之王",
		Description: "平均投递时间低于1小时",
		Rarity:      BadgeRarityCommon,
		Points:      20,
		Conditions:  `{"type":"delivery_time","value":60,"unit":"minutes"}`,
	},
	{
		Code:        "distance_champion",
		Name:        "里程冠军",
		Description: "单月投递里程超过500公里",
		Rarity:      BadgeRarityEpic,
		Points:      200,
		Conditions:  `{"type":"distance","value":500,"period":"monthly"}`,
	},
	{
		Code:        "loyalty_veteran",
		Name:        "忠诚老兵",
		Description: "连续服务6个月以上",
		Rarity:      BadgeRarityLegendary,
		Points:      500,
		Conditions:  `{"type":"service_duration","value":6,"unit":"months"}`,
	},
}

// DefaultIncentives 默认激励配置
var DefaultIncentives = []CourierIncentive{
	{
		Type:        IncentiveTypeSubsidy,
		Name:        "基础投递补贴",
		Description: "每完成一单投递获得补贴",
		Value:       2.0,
		Conditions:  `{"type":"per_delivery","base_amount":2.0}`,
	},
	{
		Type:        IncentiveTypePoints,
		Name:        "任务完成积分",
		Description: "每完成一个任务获得积分",
		Value:       10.0,
		Conditions:  `{"type":"per_task","points":10}`,
	},
	{
		Type:        IncentiveTypeCommission,
		Name:        "月度返佣",
		Description: "根据月完成量给予返佣",
		Value:       0.05, // 5%返佣
		Conditions:  `{"type":"monthly_commission","rate":0.05,"min_tasks":50}`,
	},
	{
		Type:        IncentiveTypeBonus,
		Name:        "新手奖励",
		Description: "新信使首月额外奖励",
		Value:       50.0,
		Conditions:  `{"type":"first_month_bonus","amount":50.0}`,
	},
}

// GetIncentiveTypeName 获取激励类型名称
func (t IncentiveType) GetName() string {
	switch t {
	case IncentiveTypeSubsidy:
		return "投递补贴"
	case IncentiveTypePoints:
		return "积分奖励"
	case IncentiveTypeCommission:
		return "返佣奖励"
	case IncentiveTypeBadge:
		return "徽章奖励"
	case IncentiveTypeBonus:
		return "特殊奖金"
	default:
		return "未知激励"
	}
}

// GetBadgeRarityName 获取徽章稀有度名称
func (r BadgeRarity) GetName() string {
	switch r {
	case BadgeRarityCommon:
		return "普通"
	case BadgeRarityUncommon:
		return "稀有"
	case BadgeRarityRare:
		return "珍稀"
	case BadgeRarityEpic:
		return "史诗"
	case BadgeRarityLegendary:
		return "传说"
	default:
		return "未知稀有度"
	}
}
