package handlers

import (
	"net/http"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// GetDashboardStats 获取管理后台统计数据
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取统计数据失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
		"message": "获取统计数据成功",
	})
}

// GetRecentActivities 获取最近活动
func (h *AdminHandler) GetRecentActivities(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	activities, err := h.adminService.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取活动记录失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
		"message": "获取活动记录成功",
	})
}

// InjectSeedData 注入种子数据
func (h *AdminHandler) InjectSeedData(c *gin.Context) {
	// 检查用户权限（应该只有admin才能执行）
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以注入种子数据",
		})
		return
	}

	err := h.adminService.InjectSeedData()
	if err != nil {
		if err.Error() == "seed data already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "种子数据已存在，无需重复注入",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "注入种子数据失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "种子数据注入成功",
	})
}

// GetUserManagement 获取用户管理数据
func (h *AdminHandler) GetUserManagement(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	response, err := h.adminService.GetUserManagement(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户管理数据失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "获取用户管理数据成功",
	})
}

// GetSystemSettings 获取系统设置
func (h *AdminHandler) GetSystemSettings(c *gin.Context) {
	settings, err := h.adminService.GetSystemSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取系统设置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"message": "获取系统设置成功",
	})
}

// GetAnalyticsData 获取分析数据
func (h *AdminHandler) GetAnalyticsData(c *gin.Context) {
	// 构建模拟的分析数据
	analyticsData := &models.AnalyticsData{
		UserGrowth: &models.ChartData{
			Labels: []string{"1月", "2月", "3月", "4月", "5月", "6月"},
			Datasets: []models.Dataset{
				{
					Label:           "用户增长",
					Data:            []float64{10, 25, 45, 78, 120, 156},
					BackgroundColor: "rgba(54, 162, 235, 0.6)",
					BorderColor:     "rgba(54, 162, 235, 1)",
				},
			},
		},
		LetterTrends: &models.ChartData{
			Labels: []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"},
			Datasets: []models.Dataset{
				{
					Label:           "信件投递",
					Data:            []float64{12, 19, 3, 5, 15, 8, 10},
					BackgroundColor: "rgba(255, 99, 132, 0.6)",
					BorderColor:     "rgba(255, 99, 132, 1)",
				},
			},
		},
		CourierStats: &models.ChartData{
			Labels: []string{"一级信使", "二级信使", "三级信使", "四级信使"},
			Datasets: []models.Dataset{
				{
					Label:           "信使分布",
					Data:            []float64{45, 25, 15, 8},
					BackgroundColor: "rgba(75, 192, 192, 0.6)",
					BorderColor:     "rgba(75, 192, 192, 1)",
				},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analyticsData,
		"message": "获取分析数据成功",
	})
}
