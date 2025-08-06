package handlers

import (
	"net/http"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler 数据分析处理器
type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

// NewAnalyticsHandler 创建数据分析处理器实例
func NewAnalyticsHandler(analyticsService *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetDashboard 获取仪表板数据
// @Summary 获取仪表板数据
// @Description 获取系统概览和趋势数据
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} models.DashboardData "仪表板数据"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/analytics/dashboard [get]
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	data, err := h.analyticsService.GetDashboardData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get dashboard data",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetMetrics 获取分析指标
// @Summary 获取分析指标
// @Description 根据查询条件获取分析指标数据
// @Tags analytics
// @Accept json
// @Produce json
// @Param metricType query string false "指标类型"
// @Param granularity query string false "数据粒度"
// @Param startDate query string false "开始日期"
// @Param endDate query string false "结束日期"
// @Param dimension query string false "维度"
// @Param limit query int false "限制数量" default(100)
// @Success 200 {object} []models.AnalyticsMetric "指标数据列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/analytics/metrics [get]
func (h *AnalyticsHandler) GetMetrics(c *gin.Context) {
	query := &models.AnalyticsQuery{}

	// 解析查询参数
	if metricType := c.Query("metricType"); metricType != "" {
		query.MetricType = models.AnalyticsMetricType(metricType)
	}
	if granularity := c.Query("granularity"); granularity != "" {
		query.Granularity = models.AnalyticsGranularity(granularity)
	}
	if startDate := c.Query("startDate"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			query.StartDate = parsed
		}
	}
	if endDate := c.Query("endDate"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			query.EndDate = parsed
		}
	}
	query.Dimension = c.Query("dimension")

	if limit := c.Query("limit"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			query.Limit = parsed
		}
	} else {
		query.Limit = 100
	}

	metrics, err := h.analyticsService.GetMetrics(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"count":   len(metrics),
		"query":   query,
	})
}

// GetMetricSummary 获取指标摘要
// @Summary 获取指标摘要
// @Description 获取指定指标的统计摘要
// @Tags analytics
// @Accept json
// @Produce json
// @Param metricType query string true "指标类型"
// @Param metricName query string true "指标名称"
// @Param startDate query string true "开始日期"
// @Param endDate query string true "结束日期"
// @Success 200 {object} models.MetricSummary "指标摘要"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/analytics/metrics/summary [get]
func (h *AnalyticsHandler) GetMetricSummary(c *gin.Context) {
	metricType := c.Query("metricType")
	metricName := c.Query("metricName")
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	if metricType == "" || metricName == "" || startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "metricType, metricName, startDate, and endDate are required",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid startDate format",
			"details": err.Error(),
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid endDate format",
			"details": err.Error(),
		})
		return
	}

	summary, err := h.analyticsService.GetMetricSummary(
		models.AnalyticsMetricType(metricType),
		metricName,
		startDate,
		endDate,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get metric summary",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// RecordMetric 记录分析指标
// @Summary 记录分析指标
// @Description 记录新的分析指标数据点
// @Tags analytics
// @Accept json
// @Produce json
// @Param metric body map[string]interface{} true "指标数据"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/analytics/metrics [post]
func (h *AnalyticsHandler) RecordMetric(c *gin.Context) {
	var req struct {
		MetricType  string                 `json:"metricType" binding:"required"`
		MetricName  string                 `json:"metricName" binding:"required"`
		Value       float64                `json:"value" binding:"required"`
		Unit        string                 `json:"unit"`
		Dimension   string                 `json:"dimension"`
		Granularity string                 `json:"granularity" binding:"required"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	err := h.analyticsService.RecordMetric(
		models.AnalyticsMetricType(req.MetricType),
		req.MetricName,
		req.Value,
		req.Unit,
		req.Dimension,
		models.AnalyticsGranularity(req.Granularity),
		req.Metadata,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to record metric",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Metric recorded successfully",
	})
}

// GenerateReport 生成分析报告
// @Summary 生成分析报告
// @Description 生成指定类型的分析报告
// @Tags analytics
// @Accept json
// @Produce json
// @Param report body models.GenerateReportRequest true "报告生成请求"
// @Success 200 {object} models.AnalyticsReport "生成的报告"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/analytics/reports [post]
func (h *AnalyticsHandler) GenerateReport(c *gin.Context) {
	var req models.GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	report, err := h.analyticsService.GenerateReport(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetUserAnalytics 获取用户分析数据
// @Summary 获取用户分析数据
// @Description 获取指定时间范围内的用户分析数据
// @Tags analytics
// @Accept json
// @Produce json
// @Param startDate query string false "开始日期" default("7 days ago")
// @Param endDate query string false "结束日期" default("today")
// @Success 200 {object} []models.UserAnalytics "用户分析数据"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/analytics/users [get]
func (h *AnalyticsHandler) GetUserAnalytics(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 默认查询最近7天的数据
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	if startDateStr := c.Query("startDate"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("endDate"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	// 更新用户分析数据（如果需要）
	for d := startDate; d.Before(endDate.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
		h.analyticsService.UpdateUserAnalytics(userID, d)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User analytics updated",
		"period": gin.H{
			"startDate": startDate.Format("2006-01-02"),
			"endDate":   endDate.Format("2006-01-02"),
		},
	})
}

// GetSystemAnalytics 获取系统分析数据
// @Summary 获取系统分析数据
// @Description 获取系统级别的分析数据（仅管理员）
// @Tags analytics
// @Accept json
// @Produce json
// @Param startDate query string false "开始日期"
// @Param endDate query string false "结束日期"
// @Success 200 {object} []models.SystemAnalytics "系统分析数据"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/analytics/system [get]
func (h *AnalyticsHandler) GetSystemAnalytics(c *gin.Context) {
	// 检查管理员权限
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Admin access required",
		})
		return
	}

	// 默认查询最近30天的数据
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("startDate"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("endDate"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	// 更新系统分析数据
	for d := startDate; d.Before(endDate.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
		h.analyticsService.UpdateSystemAnalytics(d)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "System analytics updated",
		"period": gin.H{
			"startDate": startDate.Format("2006-01-02"),
			"endDate":   endDate.Format("2006-01-02"),
		},
	})
}

// RecordPerformance 记录性能指标
// @Summary 记录性能指标
// @Description 记录API请求的性能指标
// @Tags analytics
// @Accept json
// @Produce json
// @Param performance body map[string]interface{} true "性能数据"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/analytics/performance [post]
func (h *AnalyticsHandler) RecordPerformance(c *gin.Context) {
	var req struct {
		Endpoint     string  `json:"endpoint" binding:"required"`
		Method       string  `json:"method" binding:"required"`
		ResponseTime float64 `json:"responseTime" binding:"required"`
		StatusCode   int     `json:"statusCode" binding:"required"`
		UserAgent    string  `json:"userAgent"`
		IPAddress    string  `json:"ipAddress"`
		UserID       *string `json:"userId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	err := h.analyticsService.RecordPerformanceMetric(
		req.Endpoint,
		req.Method,
		req.ResponseTime,
		req.StatusCode,
		req.UserAgent,
		req.IPAddress,
		req.UserID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to record performance metric",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Performance metric recorded successfully",
	})
}

// GetReports 获取分析报告列表
// @Summary 获取分析报告列表
// @Description 获取用户的分析报告列表
// @Tags analytics
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param reportType query string false "报告类型"
// @Success 200 {object} map[string]interface{} "报告列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/analytics/reports [get]
func (h *AnalyticsHandler) GetReports(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	reportType := c.Query("reportType")

	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// offset calculation would be: (page - 1) * pageSize

	// 这里应该从数据库查询报告列表，暂时返回示例数据
	c.JSON(http.StatusOK, gin.H{
		"reports": []map[string]interface{}{
			{
				"id":          "report-1",
				"title":       "User Engagement Report",
				"reportType":  "user",
				"status":      "completed",
				"generatedBy": userID,
				"createdAt":   time.Now().AddDate(0, 0, -1),
			},
		},
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    1,
		},
		"filters": gin.H{
			"reportType": reportType,
		},
	})
}
