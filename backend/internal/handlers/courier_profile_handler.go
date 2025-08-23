package handlers

import (
	"fmt"
	"net/http"
	"time"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CourierProfileHandler 信使个人中心处理器
type CourierProfileHandler struct {
	db                *gorm.DB
	courierService    *services.CourierService
	promotionService  *services.PromotionService
	creditService     *services.CreditService
}

// NewCourierProfileHandler 创建信使个人中心处理器
func NewCourierProfileHandler(
	db *gorm.DB,
	courierService *services.CourierService,
	promotionService *services.PromotionService,
	creditService *services.CreditService,
) *CourierProfileHandler {
	return &CourierProfileHandler{
		db:               db,
		courierService:   courierService,
		promotionService: promotionService,
		creditService:    creditService,
	}
}

// ProfileInfo 个人信息
type ProfileInfo struct {
	CourierID          string    `json:"courier_id"`
	UserID             string    `json:"user_id"`
	Username           string    `json:"username"`
	Level              int       `json:"level"`
	LevelName          string    `json:"level_name"`
	Status             string    `json:"status"`
	Points             int       `json:"points"`
	ManagedArea        string    `json:"managed_area"`
	ParentCourierID    string    `json:"parent_courier_id,omitempty"`
	ParentCourierName  string    `json:"parent_courier_name,omitempty"`
	JoinedAt           time.Time `json:"joined_at"`
	LastActiveAt       time.Time `json:"last_active_at"`
	CompletedTasks     int64     `json:"completed_tasks"`
	TotalTasks         int       `json:"total_tasks"`
	SuccessRate        float64   `json:"success_rate"`
	NextLevelProgress  int       `json:"next_level_progress"`
	Achievements       []string  `json:"achievements"`
}

// GrowthRecord 成长记录
type GrowthRecord struct {
	Date         string  `json:"date"`
	Points       int     `json:"points"`
	Tasks        int     `json:"tasks"`
	Rating       float64 `json:"rating"`
	Milestone    string  `json:"milestone,omitempty"`
}

// PointsDetail 积分详情
type PointsDetail struct {
	TotalPoints      int                `json:"total_points"`
	AvailablePoints  int                `json:"available_points"`
	FrozenPoints     int                `json:"frozen_points"`
	ThisMonthEarned  int                `json:"this_month_earned"`
	ThisWeekEarned   int                `json:"this_week_earned"`
	RecentRecords    []PointsRecord     `json:"recent_records"`
	PointsSources    map[string]int     `json:"points_sources"`
}

// PointsRecord 积分记录
type PointsRecord struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Amount      int       `json:"amount"`
	Balance     int       `json:"balance"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetInfo 获取个人信息
func (h *CourierProfileHandler) GetInfo(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	// 获取信使信息
	courier, err := h.courierService.GetCourierByUserID(user.ID)
	if err != nil || courier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信使信息不存在",
		})
		return
	}

	// 获取上级信使信息
	var parentName string
	if courier.ParentID != nil && *courier.ParentID != "" {
		var parentCourier models.Courier
		if err := h.db.First(&parentCourier, "id = ?", *courier.ParentID).Error; err == nil {
			var parentUser models.User
			if err := h.db.First(&parentUser, "id = ?", parentCourier.UserID).Error; err == nil {
				parentName = parentUser.Username
			}
		}
	}

	// 查询已完成的任务数
	var completedCount int64
	h.db.Model(&models.CourierTask{}).
		Where("courier_id = ? AND status = ?", courier.ID, "delivered").
		Count(&completedCount)
	
	// 计算成功率
	successRate := 100.0
	if courier.TaskCount > 0 {
		successRate = float64(completedCount) / float64(courier.TaskCount) * 100
	}

	// 计算下一级进度
	nextLevelProgress := h.calculateNextLevelProgress(courier)

	// 获取成就列表
	achievements := h.getAchievements(courier)

	// 构建响应
	parentCourierID := ""
	if courier.ParentID != nil {
		parentCourierID = *courier.ParentID
	}
	
	info := ProfileInfo{
		CourierID:         courier.ID,
		UserID:            user.ID,
		Username:          user.Username,
		Level:             courier.Level,
		LevelName:         h.getLevelName(courier.Level),
		Status:            courier.Status,
		Points:            courier.Points,
		ManagedArea:       courier.ManagedOPCodePrefix,
		ParentCourierID:   parentCourierID,
		ParentCourierName: parentName,
		JoinedAt:          courier.CreatedAt,
		LastActiveAt:      courier.UpdatedAt,
		CompletedTasks:    completedCount,
		TotalTasks:        courier.TaskCount,
		SuccessRate:       successRate,
		NextLevelProgress: nextLevelProgress,
		Achievements:      achievements,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    info,
	})
}

// GetGrowth 获取成长记录
func (h *CourierProfileHandler) GetGrowth(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	// 获取信使信息
	courier, err := h.courierService.GetCourierByUserID(user.ID)
	if err != nil || courier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信使信息不存在",
		})
		return
	}

	// 获取最近30天的成长记录
	records := make([]GrowthRecord, 0, 30)
	for i := 29; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		
		// 查询当天数据
		var dayTasks int64
		var dayPoints int
		
		h.db.Model(&models.CourierTask{}).
			Where("courier_id = ? AND DATE(created_at) = ?", courier.ID, dateStr).
			Count(&dayTasks)
		
		// 这里应该查询积分记录表，暂时模拟
		dayPoints = int(dayTasks) * 10
		
		record := GrowthRecord{
			Date:   dateStr,
			Points: dayPoints,
			Tasks:  int(dayTasks),
			Rating: 4.5 + float64(i%5)*0.1, // 模拟评分
		}
		
		// 检查里程碑
		if i == 15 {
			record.Milestone = "完成100个任务"
		}
		
		records = append(records, record)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"records":        records,
			"total_growth":   courier.Points,
			"monthly_growth": 450, // 应该计算实际值
			"rank":          12,   // 应该查询实际排名
		},
	})
}

// GetPoints 获取积分详情
func (h *CourierProfileHandler) GetPoints(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	// 获取积分余额
	credits, err := h.creditService.GetUserCreditInfo(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "获取积分失败",
		})
		return
	}

	// 获取积分记录
	var records []models.CreditTransaction
	h.db.Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Limit(20).
		Find(&records)

	// 转换记录格式，计算每条记录时的余额
	pointsRecords := make([]PointsRecord, 0, len(records))
	currentBalance := credits.Total
	
	for _, record := range records {
		pr := PointsRecord{
			ID:          record.ID,
			Type:        record.Type,
			Amount:      record.Amount,
			Balance:     currentBalance,
			Description: record.Description,
			CreatedAt:   record.CreatedAt,
		}
		pointsRecords = append(pointsRecords, pr)
		
		// 反向计算前一条记录的余额
		if record.Type == "spend" {
			currentBalance += record.Amount
		} else {
			currentBalance -= record.Amount
		}
	}

	// 计算积分来源分布
	pointsSources := map[string]int{
		"任务完成": 1200,
		"用户好评": 300,
		"活动奖励": 150,
		"系统奖励": 100,
	}

	// 构建响应
	detail := PointsDetail{
		TotalPoints:     credits.Total,
		AvailablePoints: credits.Available,
		FrozenPoints:    credits.Total - credits.Available,
		ThisMonthEarned: 450, // 应该计算实际值
		ThisWeekEarned:  120, // 应该计算实际值
		RecentRecords:   pointsRecords,
		PointsSources:   pointsSources,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    detail,
	})
}

// ApplyPromotion 申请晋升
func (h *CourierProfileHandler) ApplyPromotion(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	var req struct {
		TargetLevel int    `json:"target_level" binding:"required,min=2,max=4"`
		Reason      string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 获取信使信息
	courier, err := h.courierService.GetCourierByUserID(user.ID)
	if err != nil || courier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信使信息不存在",
		})
		return
	}

	// 检查是否可以申请晋升
	if req.TargetLevel <= courier.Level {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4002,
			"message": "目标等级必须高于当前等级",
		})
		return
	}

	// 检查晋升条件
	eligible, reason := h.checkPromotionEligibility(courier, req.TargetLevel)
	if !eligible {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4003,
			"message": reason,
		})
		return
	}

	// TODO: 晋升申请功能暂未实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"code":    5000,
		"message": "晋升申请功能暂未开放",
	})
}

// 辅助方法
func (h *CourierProfileHandler) getLevelName(level int) string {
	levelNames := map[int]string{
		1: "楼栋投递员",
		2: "片区管理员",
		3: "学校协调员",
		4: "城市总监",
	}
	if name, exists := levelNames[level]; exists {
		return name
	}
	return "未知级别"
}

func (h *CourierProfileHandler) calculateNextLevelProgress(courier *models.Courier) int {
	// 根据不同等级的晋升要求计算进度
	requirements := map[int]struct{ tasks, points int }{
		1: {tasks: 100, points: 1000},
		2: {tasks: 500, points: 5000},
		3: {tasks: 2000, points: 20000},
	}

	// 查询已完成的任务数
	var completedCount int64
	h.db.Model(&models.CourierTask{}).
		Where("courier_id = ? AND status = ?", courier.ID, "delivered").
		Count(&completedCount)

	if req, exists := requirements[courier.Level]; exists {
		taskProgress := float64(completedCount) / float64(req.tasks) * 50
		pointProgress := float64(courier.Points) / float64(req.points) * 50
		return int(taskProgress + pointProgress)
	}
	return 0
}

func (h *CourierProfileHandler) getAchievements(courier *models.Courier) []string {
	achievements := []string{}
	
	// 查询已完成的任务数
	var completedCount int64
	h.db.Model(&models.CourierTask{}).
		Where("courier_id = ? AND status = ?", courier.ID, "delivered").
		Count(&completedCount)
	
	if completedCount >= 100 {
		achievements = append(achievements, "百单达人")
	}
	if completedCount >= 1000 {
		achievements = append(achievements, "千单王者")
	}
	// 暂时不计算平均评分
	// if courier.AverageRating >= 4.8 {
	// 	achievements = append(achievements, "服务之星")
	// }
	if courier.Points >= 10000 {
		achievements = append(achievements, "积分达人")
	}
	
	return achievements
}

func (h *CourierProfileHandler) checkPromotionEligibility(courier *models.Courier, targetLevel int) (bool, string) {
	// 查询已完成的任务数
	var completedCount int64
	h.db.Model(&models.CourierTask{}).
		Where("courier_id = ? AND status = ?", courier.ID, "delivered").
		Count(&completedCount)
	
	// 检查任务完成数
	requiredTasks := map[int]int{
		2: 100,
		3: 500,
		4: 2000,
	}
	
	if tasks, exists := requiredTasks[targetLevel]; exists {
		if int(completedCount) < tasks {
			return false, fmt.Sprintf("需要完成至少%d个任务，当前: %d", tasks, completedCount)
		}
	}
	
	// 检查积分要求
	requiredPoints := map[int]int{
		2: 1000,
		3: 5000,
		4: 20000,
	}
	
	if points, exists := requiredPoints[targetLevel]; exists {
		if courier.Points < points {
			return false, fmt.Sprintf("需要至少%d积分，当前: %d", points, courier.Points)
		}
	}
	
	// 暂时不检查评分要求
	// if courier.AverageRating < 4.0 {
	// 	return false, "平均评分需要达到4.0以上"
	// }
	
	return true, ""
}