package handlers

import (
	"net/http"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CourierGrowthHandler 信使成长系统处理器
type CourierGrowthHandler struct {
	courierService   *services.CourierService
	userService      *services.UserService
	promotionService *services.PromotionService
}

// NewCourierGrowthHandler 创建信使成长处理器
func NewCourierGrowthHandler(courierService *services.CourierService, userService *services.UserService, promotionService *services.PromotionService) *CourierGrowthHandler {
	return &CourierGrowthHandler{
		courierService:   courierService,
		userService:      userService,
		promotionService: promotionService,
	}
}

// GetGrowthPath 获取成长路径
func (h *CourierGrowthHandler) GetGrowthPath(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")

	if !exists || userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户未认证",
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	// 获取当前信使信息
	courier, err := h.courierService.GetCourierByUserID(userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "信使信息不存在",
		})
		return
	}

	// 构建成长路径
	growthPath := map[string]interface{}{
		"courier_id":    courier.ID,
		"current_level": courier.Level,
		"current_name":  getLevelName(courier.Level),
		"paths":         []map[string]interface{}{},
	}

	// 添加可能的晋升路径
	for level := courier.Level + 1; level <= 4; level++ {
		path := map[string]interface{}{
			"target_level": level,
			"target_name":  getLevelName(level),
			"requirements": getUpgradeRequirements(level),
			"zone_type":    getZoneType(level),
			"permissions":  getLevelPermissions(level),
		}

		// 如果是下一级，检查是否满足条件
		if level == courier.Level+1 {
			canUpgrade, completionRate := h.checkUpgradeRequirements(courier, level)
			path["can_upgrade"] = canUpgrade
			path["completion_rate"] = completionRate
			path["detailed_requirements"] = getDetailedRequirements(courier, level)
		}

		growthPath["paths"] = append(growthPath["paths"].([]map[string]interface{}), path)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    growthPath,
	})
}

// GetGrowthProgress 获取成长进度
func (h *CourierGrowthHandler) GetGrowthProgress(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户未认证",
		})
		return
	}
	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	courier, err := h.courierService.GetCourierByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "信使信息不存在",
		})
		return
	}

	progress := map[string]interface{}{
		"courier_id":       courier.ID,
		"current_level":    courier.Level,
		"total_points":     courier.Points,
		"available_points": courier.Points,
		"badges_earned":    0, // 简化实现
		"last_updated":     time.Now(),
	}

	// 如果不是最高等级，添加下一级信息
	if courier.Level < 4 {
		nextLevel := courier.Level + 1
		canUpgrade, completionRate := h.checkUpgradeRequirements(courier, nextLevel)

		progress["next_level"] = nextLevel
		progress["can_upgrade"] = canUpgrade
		progress["completion_rate"] = completionRate
		progress["requirements"] = getDetailedRequirements(courier, nextLevel)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    progress,
	})
}

// GetLevelConfig 获取等级配置
func (h *CourierGrowthHandler) GetLevelConfig(c *gin.Context) {
	config := map[string]interface{}{
		"levels": []map[string]interface{}{
			{
				"level":       1,
				"name":        "一级信使",
				"zone_type":   "building",
				"permissions": []string{"scan", "status_change", "handover"},
				"description": "负责本楼栋信件扫码登记、状态变更和向上转交",
			},
			{
				"level":       2,
				"name":        "二级信使",
				"zone_type":   "area",
				"permissions": []string{"scan", "status_change", "handover", "package", "distribute", "receive_handover"},
				"description": "负责片区管理、打包分拣、信封分发和接收一级转交",
			},
			{
				"level":       3,
				"name":        "三级信使",
				"zone_type":   "campus",
				"permissions": []string{"scan", "status_change", "handover", "package", "distribute", "receive_handover", "feedback", "performance"},
				"description": "负责全校权限、用户反馈处理和校级绩效查看",
			},
			{
				"level":       4,
				"name":        "四级信使",
				"zone_type":   "city",
				"permissions": []string{"scan", "status_change", "package", "distribute", "receive_handover", "feedback", "performance"},
				"description": "负责全域权限、多校管理，不可向上转交",
			},
		},
		"permissions": []map[string]interface{}{
			{"code": "scan", "name": "扫码登记", "description": "扫描信件二维码进行状态更新"},
			{"code": "status_change", "name": "状态变更", "description": "更改信件投递状态"},
			{"code": "handover", "name": "向上转交", "description": "将任务转交给上级信使"},
			{"code": "package", "name": "打包分拣", "description": "对信件进行打包和分拣"},
			{"code": "distribute", "name": "信封分发", "description": "分发信件到下级信使"},
			{"code": "receive_handover", "name": "接收转交", "description": "接收下级信使的转交任务"},
			{"code": "feedback", "name": "用户反馈处理", "description": "处理用户投诉和反馈"},
			{"code": "performance", "name": "绩效查看", "description": "查看绩效和统计数据"},
		},
		"zone_types": []map[string]interface{}{
			{"code": "building", "name": "楼栋", "description": "单个楼栋或宿舍楼"},
			{"code": "area", "name": "片区", "description": "多个楼栋组成的片区"},
			{"code": "campus", "name": "校区", "description": "整个校园区域"},
			{"code": "city", "name": "城市", "description": "整个城市范围"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    config,
	})
}

// CheckLevel 检查信使等级
func (h *CourierGrowthHandler) CheckLevel(c *gin.Context) {
	courierID := c.Param("courier_id")
	if courierID == "" {
		userIDInterface, _ := c.Get("user_id")
		courierID, _ = userIDInterface.(string)
	}

	courier, err := h.courierService.GetCourierByUserID(courierID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "信使信息不存在",
		})
		return
	}

	levelInfo := map[string]interface{}{
		"id":          courier.ID,
		"level":       courier.Level,
		"zone_type":   getZoneType(courier.Level),
		"zone_id":     courier.Zone,
		"zone_name":   courier.Zone,
		"permissions": getLevelPermissions(courier.Level),
	}

	// 添加可创建的下级等级
	canCreateLevels := []int{}
	if courier.Level >= 2 {
		for i := 1; i < courier.Level; i++ {
			canCreateLevels = append(canCreateLevels, i)
		}
	}
	levelInfo["can_create_levels"] = canCreateLevels

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    levelInfo,
	})
}

// SubmitUpgradeRequest 提交晋升申请
func (h *CourierGrowthHandler) SubmitUpgradeRequest(c *gin.Context) {
	userIDInterface, _ := c.Get("user_id")
	userID, _ := userIDInterface.(string)

	var req struct {
		RequestLevel int                    `json:"request_level" binding:"required"`
		Reason       string                 `json:"reason" binding:"required"`
		Evidence     map[string]interface{} `json:"evidence"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取当前信使信息
	courier, err := h.courierService.GetCourierByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "信使信息不存在",
		})
		return
	}

	// 检查是否可以申请
	if req.RequestLevel != courier.Level+1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只能申请晋升到下一级",
		})
		return
	}

	// 使用PromotionService提交真实的晋升申请 (使用userID)
	upgradeRequest, err := h.promotionService.SubmitUpgradeRequest(
		userID,
		courier.Level,
		req.RequestLevel,
		req.Reason,
		req.Evidence,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "晋升申请已提交",
		"data":    upgradeRequest,
	})
}

// GetUpgradeRequests 获取晋升申请列表
func (h *CourierGrowthHandler) GetUpgradeRequests(c *gin.Context) {
	status := c.Query("status")
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	requests, total, err := h.promotionService.GetUpgradeRequests(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取申请列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"requests": requests,
			"total":    total,
			"limit":    limit,
			"offset":   offset,
		},
	})
}

// ProcessUpgradeRequest 处理晋升申请
func (h *CourierGrowthHandler) ProcessUpgradeRequest(c *gin.Context) {
	requestID := c.Param("request_id")

	var req struct {
		Action  string `json:"action" binding:"required,oneof=approve reject"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取审核者ID
	userIDInterface, _ := c.Get("user_id")
	reviewerID, _ := userIDInterface.(string)

	// 处理申请
	err := h.promotionService.ProcessUpgradeRequest(requestID, reviewerID, req.Action, req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	actionMap := map[string]string{"approve": "批准", "reject": "驳回"}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "晋升申请处理成功",
		"data": map[string]interface{}{
			"request_id":   requestID,
			"action":       req.Action,
			"processed_by": reviewerID,
			"message":      "晋升申请已" + actionMap[req.Action],
		},
	})
}

// 辅助函数
func getLevelName(level int) string {
	names := map[int]string{
		1: "一级信使",
		2: "二级信使",
		3: "三级信使",
		4: "四级信使",
	}
	return names[level]
}

func getZoneType(level int) string {
	types := map[int]string{
		1: "building",
		2: "area",
		3: "campus",
		4: "city",
	}
	return types[level]
}

func getLevelPermissions(level int) []string {
	permissions := map[int][]string{
		1: {"scan", "status_change", "handover"},
		2: {"scan", "status_change", "handover", "package", "distribute", "receive_handover"},
		3: {"scan", "status_change", "handover", "package", "distribute", "receive_handover", "feedback", "performance"},
		4: {"scan", "status_change", "package", "distribute", "receive_handover", "feedback", "performance"},
	}
	return permissions[level]
}

func getUpgradeRequirements(level int) []map[string]interface{} {
	requirements := map[int][]map[string]interface{}{
		2: {
			{"type": "delivery_count", "name": "累计投递数量", "description": "完成至少50次投递", "target": 50},
			{"type": "consecutive_days", "name": "连续工作天数", "description": "连续工作7天", "target": 7},
			{"type": "completion_rate", "name": "任务完成率", "description": "近30天完成率达到95%", "target": 95},
		},
		3: {
			{"type": "delivery_count", "name": "累计投递数量", "description": "完成至少200次投递", "target": 200},
			{"type": "manage_couriers", "name": "管理信使数量", "description": "管理至少5名一级信使", "target": 5},
			{"type": "service_duration", "name": "服务时长", "description": "担任二级信使满3个月", "target": 3},
		},
		4: {
			{"type": "delivery_count", "name": "累计投递数量", "description": "完成至少500次投递", "target": 500},
			{"type": "school_recommendation", "name": "学校推荐", "description": "获得学校管理部门推荐", "target": 1},
			{"type": "platform_approval", "name": "平台审核", "description": "通过平台高级审核", "target": 1},
		},
	}
	return requirements[level]
}

func getDetailedRequirements(courier *models.Courier, targetLevel int) []map[string]interface{} {
	requirements := getUpgradeRequirements(targetLevel)
	detailed := make([]map[string]interface{}, len(requirements))

	for i, req := range requirements {
		detailed[i] = req
		// 模拟当前进度
		switch req["type"] {
		case "delivery_count":
			detailed[i]["current"] = courier.TaskCount
			detailed[i]["completed"] = courier.TaskCount >= req["target"].(int)
		case "completion_rate":
			rate := 95.5 // 模拟数据
			detailed[i]["current"] = rate
			detailed[i]["completed"] = rate >= float64(req["target"].(int))
		default:
			detailed[i]["current"] = 0
			detailed[i]["completed"] = false
		}
	}

	return detailed
}

func (h *CourierGrowthHandler) checkUpgradeRequirements(courier *models.Courier, targetLevel int) (bool, float64) {
	requirements := getDetailedRequirements(courier, targetLevel)
	completedCount := 0

	for _, req := range requirements {
		if req["completed"].(bool) {
			completedCount++
		}
	}

	completionRate := float64(completedCount) / float64(len(requirements)) * 100
	canUpgrade := completedCount == len(requirements)

	return canUpgrade, completionRate
}

func generateID() string {
	return "req-" + time.Now().Format("20060102150405")
}
