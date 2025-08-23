package models

import (
	"time"
	"gorm.io/gorm"
)

// CourierStatsPeriod 统计周期
type CourierStatsPeriod string

const (
	CourierStatsPeriodDaily   CourierStatsPeriod = "daily"   // 日统计
	CourierStatsPeriodWeekly  CourierStatsPeriod = "weekly"  // 周统计
	CourierStatsPeriodMonthly CourierStatsPeriod = "monthly" // 月统计
	CourierStatsPeriodAnnual  CourierStatsPeriod = "annual"  // 年统计
)

// CourierStatsType 统计类型
type CourierStatsType string

const (
	CourierStatsTypePerformance CourierStatsType = "performance" // 绩效统计
	CourierStatsTypeEarnings    CourierStatsType = "earnings"    // 收益统计
	CourierStatsTypeActivity    CourierStatsType = "activity"    // 活动统计
	CourierStatsTypeRating      CourierStatsType = "rating"      // 评分统计
)

// CourierStats 信使统计数据
type CourierStats struct {
	ID         string             `json:"id" gorm:"primaryKey;type:varchar(36)"`
	CourierID  string             `json:"courier_id" gorm:"type:varchar(36);not null;index"`
	Period     CourierStatsPeriod `json:"period" gorm:"type:varchar(20);not null;index"`
	StatsType  CourierStatsType   `json:"stats_type" gorm:"type:varchar(20);not null;index"`
	StatsDate  time.Time          `json:"stats_date" gorm:"not null;index"` // 统计日期
	
	// 任务相关统计
	TasksAssigned   int `json:"tasks_assigned" gorm:"default:0"`   // 分配的任务数
	TasksCompleted  int `json:"tasks_completed" gorm:"default:0"`  // 完成的任务数
	TasksCancelled  int `json:"tasks_cancelled" gorm:"default:0"`  // 取消的任务数
	TasksFailed     int `json:"tasks_failed" gorm:"default:0"`     // 失败的任务数
	TasksInProgress int `json:"tasks_in_progress" gorm:"default:0"` // 进行中的任务数
	
	// 绩效相关统计
	SuccessRate        float64 `json:"success_rate" gorm:"default:0"`        // 成功率
	AverageRating      float64 `json:"average_rating" gorm:"default:0"`      // 平均评分
	CompletionTime     int     `json:"completion_time" gorm:"default:0"`     // 平均完成时间(分钟)
	OnTimeDeliveryRate float64 `json:"on_time_delivery_rate" gorm:"default:0"` // 准时送达率
	
	// 收益相关统计
	TotalEarnings    float64 `json:"total_earnings" gorm:"default:0"`    // 总收益
	BaseEarnings     float64 `json:"base_earnings" gorm:"default:0"`     // 基础收益
	BonusEarnings    float64 `json:"bonus_earnings" gorm:"default:0"`    // 奖励收益
	PenaltyDeduction float64 `json:"penalty_deduction" gorm:"default:0"` // 罚款扣除
	
	// 活动相关统计
	ActiveHours      float64 `json:"active_hours" gorm:"default:0"`      // 活跃时间(小时)
	ActiveDays       int     `json:"active_days" gorm:"default:0"`       // 活跃天数
	TotalDistance    float64 `json:"total_distance" gorm:"default:0"`    // 总配送距离(公里)
	AverageDistance  float64 `json:"average_distance" gorm:"default:0"`  // 平均配送距离(公里)
	
	// 客户满意度统计
	FiveStarReviews  int     `json:"five_star_reviews" gorm:"default:0"`  // 5星评价数
	FourStarReviews  int     `json:"four_star_reviews" gorm:"default:0"`  // 4星评价数
	ThreeStarReviews int     `json:"three_star_reviews" gorm:"default:0"` // 3星评价数
	TwoStarReviews   int     `json:"two_star_reviews" gorm:"default:0"`   // 2星评价数
	OneStarReviews   int     `json:"one_star_reviews" gorm:"default:0"`   // 1星评价数
	TotalReviews     int     `json:"total_reviews" gorm:"default:0"`      // 总评价数
	ComplaintCount   int     `json:"complaint_count" gorm:"default:0"`    // 投诉数量
	
	// 排名相关
	LevelRank      int `json:"level_rank" gorm:"default:0"`      // 同级别排名
	OverallRank    int `json:"overall_rank" gorm:"default:0"`    // 总体排名
	RegionRank     int `json:"region_rank" gorm:"default:0"`     // 区域排名
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Courier *Courier `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
}

func (CourierStats) TableName() string {
	return "courier_stats"
}

// CourierPerformanceSummary 信使绩效汇总
type CourierPerformanceSummary struct {
	CourierID      string    `json:"courier_id"`
	Period         string    `json:"period"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	
	// 基础数据
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	SuccessRate    float64 `json:"success_rate"`
	AverageRating  float64 `json:"average_rating"`
	
	// 时间相关
	AverageCompletionTime int     `json:"average_completion_time"` // 分钟
	OnTimeDeliveryRate    float64 `json:"on_time_delivery_rate"`
	TotalActiveHours      float64 `json:"total_active_hours"`
	
	// 收益相关
	TotalEarnings  float64 `json:"total_earnings"`
	DailyAverage   float64 `json:"daily_average"`
	
	// 客户满意度
	CustomerSatisfaction float64 `json:"customer_satisfaction"`
	TotalReviews        int     `json:"total_reviews"`
	ComplaintRate       float64 `json:"complaint_rate"`
	
	// 排名信息
	LevelRanking   int `json:"level_ranking"`
	OverallRanking int `json:"overall_ranking"`
	
	// 改进建议
	Strengths      []string `json:"strengths"`
	Improvements   []string `json:"improvements"`
	NextLevelGoals []string `json:"next_level_goals"`
}

// CourierTrendData 信使趋势数据
type CourierTrendData struct {
	Date           time.Time `json:"date"`
	TasksCompleted int       `json:"tasks_completed"`
	Earnings       float64   `json:"earnings"`
	Rating         float64   `json:"rating"`
	ActiveHours    float64   `json:"active_hours"`
}

// CalculateSuccessRate 计算成功率
func (s *CourierStats) CalculateSuccessRate() float64 {
	total := s.TasksCompleted + s.TasksCancelled + s.TasksFailed
	if total == 0 {
		return 0
	}
	return float64(s.TasksCompleted) / float64(total) * 100
}

// CalculateCustomerSatisfaction 计算客户满意度
func (s *CourierStats) CalculateCustomerSatisfaction() float64 {
	if s.TotalReviews == 0 {
		return 0
	}
	
	totalScore := float64(s.FiveStarReviews*5 + s.FourStarReviews*4 + s.ThreeStarReviews*3 + s.TwoStarReviews*2 + s.OneStarReviews*1)
	return totalScore / float64(s.TotalReviews)
}

// CalculateComplaintRate 计算投诉率
func (s *CourierStats) CalculateComplaintRate() float64 {
	if s.TasksCompleted == 0 {
		return 0
	}
	return float64(s.ComplaintCount) / float64(s.TasksCompleted) * 100
}

// GetPerformanceGrade 获取绩效等级
func (s *CourierStats) GetPerformanceGrade() string {
	score := s.CalculatePerformanceScore()
	
	switch {
	case score >= 90:
		return "A+"
	case score >= 85:
		return "A"
	case score >= 80:
		return "B+"
	case score >= 75:
		return "B"
	case score >= 70:
		return "C+"
	case score >= 65:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// CalculatePerformanceScore 计算综合绩效分数
func (s *CourierStats) CalculatePerformanceScore() float64 {
	var score float64 = 0
	
	// 成功率 30%
	score += s.CalculateSuccessRate() * 0.3
	
	// 客户满意度 25%
	satisfaction := s.CalculateCustomerSatisfaction()
	score += (satisfaction / 5.0 * 100) * 0.25
	
	// 准时送达率 20%
	score += s.OnTimeDeliveryRate * 0.2
	
	// 投诉率影响 15% (投诉率越低越好)
	complaintImpact := (100 - s.CalculateComplaintRate()) * 0.15
	if complaintImpact < 0 {
		complaintImpact = 0
	}
	score += complaintImpact
	
	// 活跃度 10%
	activityScore := 0.0
	if s.ActiveDays > 0 {
		// 假设一个周期最多30天
		activityScore = float64(s.ActiveDays) / 30.0 * 100
		if activityScore > 100 {
			activityScore = 100
		}
	}
	score += activityScore * 0.1
	
	return score
}

// GenerateImprovementSuggestions 生成改进建议
func (s *CourierStats) GenerateImprovementSuggestions() ([]string, []string) {
	strengths := []string{}
	improvements := []string{}
	
	successRate := s.CalculateSuccessRate()
	satisfaction := s.CalculateCustomerSatisfaction()
	complaintRate := s.CalculateComplaintRate()
	
	// 分析优势
	if successRate >= 95 {
		strengths = append(strengths, "任务完成率出色")
	}
	if satisfaction >= 4.5 {
		strengths = append(strengths, "客户满意度很高")
	}
	if s.OnTimeDeliveryRate >= 90 {
		strengths = append(strengths, "准时送达率优秀")
	}
	if complaintRate <= 2 {
		strengths = append(strengths, "投诉率很低")
	}
	
	// 分析改进点
	if successRate < 85 {
		improvements = append(improvements, "提高任务完成率")
	}
	if satisfaction < 4.0 {
		improvements = append(improvements, "改善服务质量，提高客户满意度")
	}
	if s.OnTimeDeliveryRate < 80 {
		improvements = append(improvements, "提高时间管理能力，减少延迟")
	}
	if complaintRate > 5 {
		improvements = append(improvements, "加强服务规范，减少客户投诉")
	}
	if s.CompletionTime > 120 { // 超过2小时
		improvements = append(improvements, "优化配送路线，提高效率")
	}
	
	return strengths, improvements
}

// IsPromotionReady 检查是否符合晋升条件
func (s *CourierStats) IsPromotionReady(targetLevel int) bool {
	score := s.CalculatePerformanceScore()
	
	// 根据目标等级设定不同的标准
	requiredScore := map[int]float64{
		2: 70.0, // L1->L2 需要70分
		3: 80.0, // L2->L3 需要80分
		4: 90.0, // L3->L4 需要90分
	}
	
	if required, exists := requiredScore[targetLevel]; exists {
		return score >= required && s.TasksCompleted >= 10 && s.TotalReviews >= 5
	}
	
	return false
}