package handlers

import (
	"net/http"
	"strconv"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// OPCodeHandler OP Code处理器
type OPCodeHandler struct {
	opcodeService  *services.OPCodeService
	courierService *services.CourierService
}

// NewOPCodeHandler 创建OP Code处理器
func NewOPCodeHandler(opcodeService *services.OPCodeService, courierService *services.CourierService) *OPCodeHandler {
	return &OPCodeHandler{
		opcodeService:  opcodeService,
		courierService: courierService,
	}
}

// ApplyOPCode 申请OP Code
func (h *OPCodeHandler) ApplyOPCode(c *gin.Context) {
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

	var req models.OPCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 创建申请
	application, err := h.opcodeService.ApplyForOPCode(user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "申请失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "申请提交成功",
		"data":    application,
	})
}

// ValidateOPCode 验证OP Code格式和有效性
func (h *OPCodeHandler) ValidateOPCode(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "OP Code不能为空",
		})
		return
	}

	// 验证格式
	if err := models.ValidateOPCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "OP Code格式不正确",
			"error":   err.Error(),
		})
		return
	}

	// 验证是否存在
	isValid, err := h.opcodeService.ValidateOPCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "验证失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "验证成功",
		"data": gin.H{
			"code":     code,
			"is_valid": isValid,
		},
	})
}

// SearchOPCodes 搜索OP Code
func (h *OPCodeHandler) SearchOPCodes(c *gin.Context) {
	var req models.OPCodeSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "查询参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	codes, total, err := h.opcodeService.SearchOPCodes(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "搜索失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "搜索成功",
		"data": gin.H{
			"codes": codes,
			"pagination": gin.H{
				"page":       req.Page,
				"page_size":  req.PageSize,
				"total":      total,
				"total_page": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
			},
		},
	})
}

// GetOPCodeStats 获取OP Code统计信息
func (h *OPCodeHandler) GetOPCodeStats(c *gin.Context) {
	schoolCode := c.Param("school_code")
	if schoolCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "学校代码不能为空",
		})
		return
	}

	stats, err := h.opcodeService.GetOPCodeStats(schoolCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "获取统计失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取统计成功",
		"data":    stats,
	})
}

// AdminReviewApplication 管理员审核申请
func (h *OPCodeHandler) AdminReviewApplication(c *gin.Context) {
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

	// 检查管理员权限
	if user.Role != models.RolePlatformAdmin && user.Role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "权限不足",
		})
		return
	}

	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "申请ID不能为空",
		})
		return
	}

	var req models.OPCodeAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	req.ApplicationID = applicationID

	// 分配OP Code
	if err := h.opcodeService.AssignOPCode(user.ID, req.ApplicationID, req.PointCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "审核失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "审核成功",
	})
}

// GetOPCode 根据编码获取OP Code信息
func (h *OPCodeHandler) GetOPCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "OP Code不能为空",
		})
		return
	}

	// 检查用户权限
	userInterface, exists := c.Get("user")
	includePrivate := false
	if exists {
		user := userInterface.(*models.User)
		// 管理员或信使可以查看私有信息
		includePrivate = user.Role == models.RolePlatformAdmin || 
						user.Role == models.RoleSuperAdmin ||
						user.Role == models.RoleCourierLevel1 ||
						user.Role == models.RoleCourierLevel2 ||
						user.Role == models.RoleCourierLevel3 ||
						user.Role == models.RoleCourierLevel4
	}

	opCode, err := h.opcodeService.GetOPCodeByCode(code, includePrivate)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "OP Code不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    opCode,
	})
}

// SearchAreas 搜索片区
func (h *OPCodeHandler) SearchAreas(c *gin.Context) {
	// 获取学校代码参数
	schoolCode := c.Query("schoolCode")
	if schoolCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "学校代码不能为空",
			"data":    nil,
		})
		return
	}

	result, err := h.opcodeService.SearchAreas(schoolCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "查询成功",
		"data":    result,
	})
}

// SearchBuildings 搜索楼栋
func (h *OPCodeHandler) SearchBuildings(c *gin.Context) {
	schoolCode := c.Query("schoolCode")
	areaCode := c.Query("areaCode")
	
	if schoolCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "学校代码不能为空",
			"data":    nil,
		})
		return
	}

	result, err := h.opcodeService.SearchBuildings(schoolCode, areaCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "查询成功",
		"data":    result,
	})
}

// SearchPoints 搜索投递点
func (h *OPCodeHandler) SearchPoints(c *gin.Context) {
	schoolCode := c.Query("schoolCode")
	areaCode := c.Query("areaCode")
	
	if schoolCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "学校代码不能为空",
			"data":    nil,
		})
		return
	}

	result, err := h.opcodeService.SearchPoints(schoolCode, areaCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "查询成功",
		"data":    result,
	})
}

// SearchSchools 搜索学校
func (h *OPCodeHandler) SearchSchools(c *gin.Context) {
	name := c.Query("name")
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	result, err := h.opcodeService.SearchSchools(name, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "搜索成功",
		"data":    result,
	})
}