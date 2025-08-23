package handlers

import (
	"net/http"
	"time"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CourierDashboardHandler 信使仪表板处理器
type CourierDashboardHandler struct {
	db              *gorm.DB
	courierService  *services.CourierService
	taskService     *services.CourierTaskService
	statsService    *services.AnalyticsService
}

// NewCourierDashboardHandler 创建信使仪表板处理器
func NewCourierDashboardHandler(
	db *gorm.DB,
	courierService *services.CourierService, 
	taskService *services.CourierTaskService,
	statsService *services.AnalyticsService,
) *CourierDashboardHandler {
	return &CourierDashboardHandler{
		db:             db,
		courierService: courierService,
		taskService:    taskService,
		statsService:   statsService,
	}
}

// DashboardStats 仪表板统计数据
type DashboardStats struct {
	TodayTasks      int64   `json:"today_tasks"`
	CompletedTasks  int64   `json:"completed_tasks"`
	PendingTasks    int64   `json:"pending_tasks"`
	TotalPoints     int     `json:"total_points"`
	CurrentLevel    int     `json:"current_level"`
	TeamMembers     int64   `json:"team_members"`
	ManagedArea     string  `json:"managed_area"`
	MonthlyGrowth   float64 `json:"monthly_growth"`
	WeeklyTrend     []int   `json:"weekly_trend"`
	RecentTasks     []TaskSummary `json:"recent_tasks"`
}

// TaskSummary 任务摘要
type TaskSummary struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	Address      string    `json:"address"`
	Priority     string    `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	EstimatedTime string   `json:"estimated_time"`
}

// GetStats 获取仪表板统计数据
func (h *CourierDashboardHandler) GetStats(c *gin.Context) {
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

	// 获取今日任务统计
	today := time.Now().Format("2006-01-02")
	var todayTasks, completedTasks, pendingTasks int64
	
	// 今日总任务数
	h.db.Model(&models.CourierTask{}).
		Where("courier_id = ? AND DATE(created_at) = ?", courier.ID, today).
		Count(&todayTasks)
	
	// 今日已完成任务
	h.db.Model(&models.CourierTask{}).
		Where("courier_id = ? AND DATE(created_at) = ? AND status = ?", 
			courier.ID, today, "delivered").
		Count(&completedTasks)
	
	// 待处理任务
	h.db.Model(&models.CourierTask{}).
		Where("courier_id = ? AND status IN (?)", 
			courier.ID, []string{"pending", "accepted", "collected", "in_transit"}).
		Count(&pendingTasks)

	// 获取团队成员数（如果是管理级别）
	var teamMembers int64
	if courier.Level > 1 {
		// 获取下属信使数量
		h.db.Model(&models.Courier{}).
			Where("parent_courier_id = ?", courier.ID).
			Count(&teamMembers)
	}

	// 获取最近任务
	var recentTasks []models.CourierTask
	h.db.Model(&models.CourierTask{}).
		Where("courier_id = ?", courier.ID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentTasks)

	// 转换任务摘要
	taskSummaries := make([]TaskSummary, 0, len(recentTasks))
	for _, task := range recentTasks {
		summary := TaskSummary{
			ID:           task.ID,
			Type:         task.Priority, // 使用Priority作为任务类型
			Status:       task.Status,
			Address:      task.TargetLocation, // 使用TargetLocation作为地址
			Priority:     task.Priority,
			CreatedAt:    task.CreatedAt,
			EstimatedTime: "30分钟", // 可以根据实际逻辑计算
		}
		taskSummaries = append(taskSummaries, summary)
	}

	// 计算月度增长率（示例）
	monthlyGrowth := 23.5 // 实际应该从历史数据计算

	// 获取周趋势数据（最近7天的任务数）
	weeklyTrend := make([]int, 7)
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var count int64
		h.db.Model(&models.CourierTask{}).
			Where("courier_id = ? AND DATE(created_at) = ?", courier.ID, date).
			Count(&count)
		weeklyTrend[6-i] = int(count)
	}

	// 构建响应
	stats := DashboardStats{
		TodayTasks:     todayTasks,
		CompletedTasks: completedTasks,
		PendingTasks:   pendingTasks,
		TotalPoints:    courier.Points,
		CurrentLevel:   courier.Level,
		TeamMembers:    teamMembers,
		ManagedArea:    courier.ManagedOPCodePrefix,
		MonthlyGrowth:  monthlyGrowth,
		WeeklyTrend:    weeklyTrend,
		RecentTasks:    taskSummaries,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// GetSummary 获取仪表板摘要信息
func (h *CourierDashboardHandler) GetSummary(c *gin.Context) {
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

	// 构建摘要信息
	summary := gin.H{
		"courier_id":    courier.ID,
		"level":         courier.Level,
		"level_name":    h.getLevelName(courier.Level),
		"status":        courier.Status,
		"managed_area":  courier.ManagedOPCodePrefix,
		"points":        courier.Points,
		"completed_tasks": courier.TaskCount,
		"success_rate":  h.calculateSuccessRate(courier),
		"joined_at":     courier.CreatedAt,
		"permissions":   h.getPermissionSummary(courier.Level),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    summary,
	})
}

// 辅助方法：获取等级名称
func (h *CourierDashboardHandler) getLevelName(level int) string {
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

// 辅助方法：计算成功率
func (h *CourierDashboardHandler) calculateSuccessRate(courier *models.Courier) float64 {
	if courier.TaskCount == 0 {
		return 100.0
	}
	// 查询已完成的任务数
	var completedCount int64
	h.db.Model(&models.CourierTask{}).
		Where("courier_id = ? AND status = ?", courier.ID, "delivered").
		Count(&completedCount)
	
	return float64(completedCount) / float64(courier.TaskCount) * 100
}

// 辅助方法：获取权限摘要
func (h *CourierDashboardHandler) getPermissionSummary(level int) []string {
	permissions := []string{"查看任务", "扫码投递"}
	
	if level >= 2 {
		permissions = append(permissions, "管理团队", "审核OP Code")
	}
	if level >= 3 {
		permissions = append(permissions, "批量生成", "学校管理")
	}
	if level >= 4 {
		permissions = append(permissions, "城市管理", "数据分析", "系统配置")
	}
	
	return permissions
}