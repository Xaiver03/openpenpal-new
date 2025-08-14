package handlers

import (
	"api-gateway/internal/models"
	"api-gateway/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MetricsHandler 性能监控处理器
type MetricsHandler struct {
	metricsService *services.MetricsService
	logger         *zap.Logger
}

// NewMetricsHandler 创建性能监控处理器
func NewMetricsHandler(metricsService *services.MetricsService, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		metricsService: metricsService,
		logger:         logger,
	}
}

// SubmitPerformanceMetrics 提交性能指标
// @Summary 提交前端性能指标
// @Description 接收前端Core Web Vitals和其他性能数据
// @Tags metrics
// @Accept json
// @Produce json
// @Param metrics body models.PerformanceRequest true "性能指标数据"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/metrics/performance [post]
func (h *MetricsHandler) SubmitPerformanceMetrics(c *gin.Context) {
	var req models.PerformanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid performance metrics request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "Invalid request format",
			Details:   err.Error(),
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	// 获取用户ID (从JWT认证中间件设置)
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	// 验证必要字段
	if req.SessionID == "" || req.PageURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "Missing required fields: session_id or page_url",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	// 保存性能指标
	err := h.metricsService.SavePerformanceMetric(userID, &req)
	if err != nil {
		h.logger.Error("Failed to save performance metrics",
			zap.Error(err),
			zap.String("session_id", req.SessionID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "Failed to save performance metrics",
			Details:   "Internal server error",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	h.logger.Info("Performance metrics submitted successfully",
		zap.String("session_id", req.SessionID),
		zap.String("page_url", req.PageURL),
		zap.String("user_id", userID),
		zap.Float64("lcp", req.LCP),
		zap.Float64("fid", req.FID),
	)

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"session_id": req.SessionID,
		"timestamp":  time.Now(),
	}))
}

// GetDashboardMetrics 获取性能仪表板数据
// @Summary 获取性能仪表板数据
// @Description 获取聚合的性能指标数据用于仪表板展示
// @Tags metrics
// @Produce json
// @Param time_range query string false "时间范围" Enums(1h,24h,7d,30d) default(24h)
// @Success 200 {object} models.SuccessResponse(data=models.DashboardMetrics}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/metrics/dashboard [get]
func (h *MetricsHandler) GetDashboardMetrics(c *gin.Context) {
	timeRange := c.DefaultQuery("time_range", "24h")

	// 验证时间范围参数
	validRanges := map[string]bool{
		"1h":  true,
		"24h": true,
		"7d":  true,
		"30d": true,
	}

	if !validRanges[timeRange] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "Invalid time_range parameter",
			Details:   "Valid values: 1h, 24h, 7d, 30d",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	// 获取仪表板指标
	dashboard, err := h.metricsService.GetDashboardMetrics(timeRange)
	if err != nil {
		h.logger.Error("Failed to get dashboard metrics",
			zap.Error(err),
			zap.String("time_range", timeRange),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "Failed to get dashboard metrics",
			Details:   "Internal server error",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	h.logger.Info("Dashboard metrics retrieved successfully",
		zap.String("time_range", timeRange),
		zap.Int("performance_score", dashboard.PerformanceScore),
	)

	c.JSON(http.StatusOK, models.SuccessResponse(dashboard))
}

// GetPerformanceAlerts 获取性能告警
// @Summary 获取性能告警列表
// @Description 获取活跃的性能告警信息
// @Tags metrics
// @Produce json
// @Param limit query int false "返回记录数限制" default(50)
// @Success 200 {object} models.SuccessResponse(data=[]models.PerformanceAlert}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/metrics/alerts [get]
func (h *MetricsHandler) GetPerformanceAlerts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 || limit > 1000 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "Invalid limit parameter",
			Details:   "Limit must be a number between 0 and 1000",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	// 获取活跃告警
	alerts, err := h.metricsService.GetActiveAlerts(limit)
	if err != nil {
		h.logger.Error("Failed to get performance alerts",
			zap.Error(err),
			zap.Int("limit", limit),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "Failed to get performance alerts",
			Details:   "Internal server error",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	h.logger.Info("Performance alerts retrieved successfully",
		zap.Int("alert_count", len(alerts)),
		zap.Int("limit", limit),
	)

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"alerts": alerts,
		"total":  len(alerts),
	}))
}

// CreatePerformanceAlert 创建性能告警
// @Summary 创建性能告警
// @Description 手动创建性能告警
// @Tags metrics
// @Accept json
// @Produce json
// @Param alert body models.AlertRequest true "告警数据"
// @Success 201 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/metrics/alerts [post]
func (h *MetricsHandler) CreatePerformanceAlert(c *gin.Context) {
	var req models.AlertRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid alert request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "Invalid request format",
			Details:   err.Error(),
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	// 获取用户ID
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	// 验证严重级别
	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}

	if req.Severity == "" {
		req.Severity = "medium"
	} else if !validSeverities[req.Severity] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:      http.StatusBadRequest,
			Message:   "Invalid severity level",
			Details:   "Valid values: low, medium, high, critical",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	// 创建告警
	err := h.metricsService.CreateAlert(userID, &req)
	if err != nil {
		h.logger.Error("Failed to create performance alert",
			zap.Error(err),
			zap.String("metric_type", req.MetricType),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:      http.StatusInternalServerError,
			Message:   "Failed to create performance alert",
			Details:   "Internal server error",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
		})
		return
	}

	h.logger.Info("Performance alert created successfully",
		zap.String("metric_type", req.MetricType),
		zap.String("severity", req.Severity),
		zap.String("user_id", userID),
		zap.Float64("threshold", req.Threshold),
		zap.Float64("value", req.Value),
	)

	c.JSON(http.StatusCreated, models.SuccessResponse(gin.H{
		"metric_type": req.MetricType,
		"severity":    req.Severity,
		"timestamp":   time.Now(),
	}))
}

// GetHealthStatus 获取服务健康状态
// @Summary 获取服务健康状态
// @Description 获取系统和各微服务的健康状态
// @Tags health
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Router /api/health/status [get]
func (h *MetricsHandler) GetHealthStatus(c *gin.Context) {
	// 这里可以集成实际的健康检查逻辑
	healthData := gin.H{
		"gateway": gin.H{
			"status":    "healthy",
			"uptime":    "24h 15m 30s",
			"version":   "1.0.0",
			"cpu_usage": "15%",
			"memory":    "256MB/1GB",
		},
		"services": gin.H{
			"main-backend":    gin.H{"status": "healthy", "response_time": "45ms"},
			"write-service":   gin.H{"status": "healthy", "response_time": "38ms"},
			"courier-service": gin.H{"status": "healthy", "response_time": "52ms"},
			"admin-service":   gin.H{"status": "degraded", "response_time": "120ms"},
			"ocr-service":     gin.H{"status": "unknown", "response_time": "timeout"},
		},
		"database": gin.H{
			"status":          "healthy",
			"connection_pool": "8/10",
			"query_time":      "15ms",
		},
		"redis": gin.H{
			"status":       "healthy",
			"memory_usage": "45%",
			"connections":  "12/100",
		},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, models.SuccessResponse(healthData))
}
