package handlers

import (
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SignalCodeHandler 信号编码处理器
type SignalCodeHandler struct {
	signalCodeService *services.SignalCodeService
}

// NewSignalCodeHandler 创建信号编码处理器
func NewSignalCodeHandler(signalCodeService *services.SignalCodeService) *SignalCodeHandler {
	return &SignalCodeHandler{
		signalCodeService: signalCodeService,
	}
}

// GenerateCodeBatch 生成编码批次
// @Summary 生成编码批次
// @Description 批量生成信号编码
// @Tags 信号编码
// @Accept json
// @Produce json
// @Param request body models.SignalCodeBatchRequest true "批次生成请求"
// @Success 200 {object} models.SignalCodeBatch
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/signal-codes/batch [post]
func (h *SignalCodeHandler) GenerateCodeBatch(c *gin.Context) {
	var req models.SignalCodeBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batch, err := h.signalCodeService.GenerateCodeBatch(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    batch,
		"message": "编码批次生成成功",
	})
}

// RequestSignalCode 申请信号编码
// @Summary 申请信号编码
// @Description 申请可用的信号编码
// @Tags 信号编码
// @Accept json
// @Produce json
// @Param request body models.SignalCodeRequest true "编码申请请求"
// @Success 200 {object} models.SignalCode
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/signal-codes/request [post]
func (h *SignalCodeHandler) RequestSignalCode(c *gin.Context) {
	var req models.SignalCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code, err := h.signalCodeService.RequestSignalCode(&req)
	if err != nil {
		if err.Error() == "no available signal code found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "暂无可用编码"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    code,
		"message": "编码申请成功",
	})
}

// AssignSignalCode 分配信号编码
// @Summary 分配信号编码
// @Description 将编码分配给用户
// @Tags 信号编码
// @Accept json
// @Produce json
// @Param request body models.SignalCodeAssignRequest true "编码分配请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/signal-codes/assign [post]
func (h *SignalCodeHandler) AssignSignalCode(c *gin.Context) {
	var req models.SignalCodeAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.signalCodeService.AssignSignalCode(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "编码分配成功",
	})
}

// ReleaseSignalCode 释放信号编码
// @Summary 释放信号编码
// @Description 释放已分配的编码
// @Tags 信号编码
// @Accept json
// @Produce json
// @Param code path string true "编码"
// @Param request body map[string]string true "释放请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/signal-codes/{code}/release [post]
func (h *SignalCodeHandler) ReleaseSignalCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "编码不能为空"})
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.signalCodeService.ReleaseSignalCode(code, req.UserID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "编码释放成功",
	})
}

// SearchSignalCodes 搜索信号编码
// @Summary 搜索信号编码
// @Description 根据条件搜索编码
// @Tags 信号编码
// @Accept json
// @Produce json
// @Param code query string false "编码"
// @Param school_id query string false "学校ID"
// @Param area_id query string false "片区ID"
// @Param code_type query string false "编码类型"
// @Param is_used query bool false "是否已使用"
// @Param is_active query bool false "是否激活"
// @Param used_by query string false "使用者ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "页大小" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/signal-codes/search [get]
func (h *SignalCodeHandler) SearchSignalCodes(c *gin.Context) {
	var req models.SignalCodeSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	codes, total, err := h.signalCodeService.SearchSignalCodes(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"codes":      codes,
			"total":      total,
			"page":       req.Page,
			"page_size":  req.PageSize,
			"total_page": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
		},
		"message": "查询成功",
	})
}

// GetSignalCodeStats 获取编码统计
// @Summary 获取编码统计
// @Description 获取学校编码使用统计
// @Tags 信号编码
// @Accept json
// @Produce json
// @Param school_id path string true "学校ID"
// @Success 200 {object} models.SignalCodeStats
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/signal-codes/stats/{school_id} [get]
func (h *SignalCodeHandler) GetSignalCodeStats(c *gin.Context) {
	schoolID := c.Param("school_id")
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "学校ID不能为空"})
		return
	}

	stats, err := h.signalCodeService.GetSignalCodeStats(schoolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
		"message": "统计信息获取成功",
	})
}

// GetUsageLogs 获取编码使用日志
// @Summary 获取编码使用日志
// @Description 获取指定编码的使用历史
// @Tags 信号编码
// @Accept json
// @Produce json
// @Param code path string true "编码"
// @Param limit query int false "限制条数" default(10)
// @Success 200 {array} models.SignalCodeUsageLog
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/signal-codes/{code}/logs [get]
func (h *SignalCodeHandler) GetUsageLogs(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "编码不能为空"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	logs, err := h.signalCodeService.GetUsageLogs(code, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"message": "使用日志获取成功",
	})
}

// CreateMockData 创建测试数据
// @Summary 创建测试数据
// @Description 为开发测试创建模拟数据
// @Tags 信号编码
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/signal-codes/mock-data [post]
func (h *SignalCodeHandler) CreateMockData(c *gin.Context) {
	if err := h.signalCodeService.CreateMockData(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "测试数据创建成功",
	})
}

// RegisterSignalCodeRoutes 注册信号编码路由
func RegisterSignalCodeRoutes(router *gin.Engine, handler *SignalCodeHandler) {
	api := router.Group("/api/signal-codes")
	{
		// 批次管理
		api.POST("/batch", handler.GenerateCodeBatch)

		// 编码申请和分配
		api.POST("/request", handler.RequestSignalCode)
		api.POST("/assign", handler.AssignSignalCode)
		api.POST("/:code/release", handler.ReleaseSignalCode)

		// 查询和统计
		api.GET("/search", handler.SearchSignalCodes)
		api.GET("/stats/:school_id", handler.GetSignalCodeStats)
		api.GET("/:code/logs", handler.GetUsageLogs)

		// 开发测试
		api.POST("/mock-data", handler.CreateMockData)
	}
}
