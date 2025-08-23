package models

import (
	"time"
	"gorm.io/gorm"
)

// CourierPromotionType 信使晋升类型
type CourierPromotionType string

const (
	CourierPromotionNormal   CourierPromotionType = "normal"   // 正常晋升
	CourierPromotionEarly    CourierPromotionType = "early"    // 提前晋升
	CourierPromotionSpecial  CourierPromotionType = "special"  // 特殊晋升
	CourierPromotionDemotion CourierPromotionType = "demotion" // 降级
)

// CourierPromotionStatus 晋升状态
type CourierPromotionStatus string

const (
	CourierPromotionStatusPending   CourierPromotionStatus = "pending"   // 待审核
	CourierPromotionStatusApproved  CourierPromotionStatus = "approved"  // 已通过
	CourierPromotionStatusRejected  CourierPromotionStatus = "rejected"  // 已拒绝
	CourierPromotionStatusCancelled CourierPromotionStatus = "cancelled" // 已取消
)

// CourierPromotion 信使晋升记录
type CourierPromotion struct {
	ID          string                 `json:"id" gorm:"primaryKey;type:varchar(36)"`
	CourierID   string                 `json:"courier_id" gorm:"type:varchar(36);not null;index"`
	FromLevel   int                    `json:"from_level" gorm:"not null"`                    // 晋升前等级
	ToLevel     int                    `json:"to_level" gorm:"not null"`                      // 晋升后等级
	Type        CourierPromotionType   `json:"type" gorm:"type:varchar(20);not null"`         // 晋升类型
	Status      CourierPromotionStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"` // 晋升状态
	
	// 晋升条件和评估
	RequiredTasks    int     `json:"required_tasks" gorm:"not null"`           // 需要完成的任务数
	CompletedTasks   int     `json:"completed_tasks" gorm:"not null"`          // 已完成的任务数
	RequiredRating   float64 `json:"required_rating" gorm:"not null"`          // 需要的评分
	CurrentRating    float64 `json:"current_rating" gorm:"not null"`           // 当前评分
	RequiredDays     int     `json:"required_days" gorm:"not null"`            // 需要的服务天数
	CurrentDays      int     `json:"current_days" gorm:"not null"`             // 当前服务天数
	RequiredReviews  int     `json:"required_reviews" gorm:"not null"`         // 需要的好评数
	CurrentReviews   int     `json:"current_reviews" gorm:"not null"`          // 当前好评数
	
	// 额外信息
	Reason        string     `json:"reason" gorm:"type:text"`                   // 晋升理由/备注
	ReviewerID    string     `json:"reviewer_id" gorm:"type:varchar(36);index"` // 审核员ID
	ReviewComment string     `json:"review_comment" gorm:"type:text"`           // 审核意见
	ReviewedAt    *time.Time `json:"reviewed_at"`                               // 审核时间
	
	// 奖励信息
	CreditReward  int    `json:"credit_reward" gorm:"default:0"`     // 积分奖励
	BadgeReward   string `json:"badge_reward" gorm:"type:varchar(100)"` // 徽章奖励
	TitleReward   string `json:"title_reward" gorm:"type:varchar(100)"` // 称号奖励
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Courier  *Courier `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
	Reviewer *User    `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID"`
}

func (CourierPromotion) TableName() string {
	return "courier_promotions"
}

// PromotionEligibility 晋升资格检查结果
type PromotionEligibility struct {
	CourierID      string                 `json:"courier_id"`
	CurrentLevel   int                    `json:"current_level"`
	TargetLevel    int                    `json:"target_level"`
	Eligible       bool                   `json:"eligible"`
	Requirements   PromotionRequirements  `json:"requirements"`
	CurrentStatus  PromotionCurrentStatus `json:"current_status"`
	MissingItems   []string               `json:"missing_items"`
	EstimatedDays  int                    `json:"estimated_days"` // 估计还需要多少天可以晋升
}

// PromotionRequirements 晋升要求
type PromotionRequirements struct {
	MinTasks   int     `json:"min_tasks"`
	MinRating  float64 `json:"min_rating"`
	MinDays    int     `json:"min_days"`
	MinReviews int     `json:"min_reviews"`
}

// PromotionCurrentStatus 当前状态
type PromotionCurrentStatus struct {
	CompletedTasks int     `json:"completed_tasks"`
	CurrentRating  float64 `json:"current_rating"`
	ServiceDays    int     `json:"service_days"`
	PositiveReviews int    `json:"positive_reviews"`
}

// GetPromotionRequirements 获取等级晋升要求
func GetPromotionRequirements(fromLevel, toLevel int) PromotionRequirements {
	// 基础要求矩阵
	baseRequirements := map[int]PromotionRequirements{
		1: {MinTasks: 10, MinRating: 4.0, MinDays: 7, MinReviews: 5},   // L0 -> L1
		2: {MinTasks: 30, MinRating: 4.2, MinDays: 15, MinReviews: 15}, // L1 -> L2
		3: {MinTasks: 100, MinRating: 4.5, MinDays: 30, MinReviews: 40}, // L2 -> L3
		4: {MinTasks: 300, MinRating: 4.7, MinDays: 90, MinReviews: 100}, // L3 -> L4
	}
	
	if req, exists := baseRequirements[toLevel]; exists {
		return req
	}
	
	// 默认要求
	return PromotionRequirements{
		MinTasks:   10,
		MinRating:  4.0,
		MinDays:    7,
		MinReviews: 5,
	}
}

// CalculatePromotionScore 计算晋升得分
func (p *CourierPromotion) CalculatePromotionScore() float64 {
	var score float64 = 0
	
	// 任务完成度权重 30%
	if p.RequiredTasks > 0 {
		taskScore := float64(p.CompletedTasks) / float64(p.RequiredTasks)
		if taskScore > 1.0 {
			taskScore = 1.0
		}
		score += taskScore * 0.3
	}
	
	// 评分权重 40%
	if p.RequiredRating > 0 {
		ratingScore := p.CurrentRating / p.RequiredRating
		if ratingScore > 1.0 {
			ratingScore = 1.0
		}
		score += ratingScore * 0.4
	}
	
	// 服务时间权重 20%
	if p.RequiredDays > 0 {
		daysScore := float64(p.CurrentDays) / float64(p.RequiredDays)
		if daysScore > 1.0 {
			daysScore = 1.0
		}
		score += daysScore * 0.2
	}
	
	// 好评数权重 10%
	if p.RequiredReviews > 0 {
		reviewScore := float64(p.CurrentReviews) / float64(p.RequiredReviews)
		if reviewScore > 1.0 {
			reviewScore = 1.0
		}
		score += reviewScore * 0.1
	}
	
	return score
}

// IsEligible 检查是否符合晋升条件
func (p *CourierPromotion) IsEligible() bool {
	return p.CompletedTasks >= p.RequiredTasks &&
		   p.CurrentRating >= p.RequiredRating &&
		   p.CurrentDays >= p.RequiredDays &&
		   p.CurrentReviews >= p.RequiredReviews
}