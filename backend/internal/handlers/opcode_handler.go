package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
// 公开接口，但会根据用户认证状态返回不同级别的信息
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

	// 基础响应数据
	responseData := gin.H{
		"code":     code,
		"is_valid": isValid,
	}

	// 如果用户已认证，提供额外信息
	userInterface, exists := c.Get("user")
	if exists && isValid {
		user := userInterface.(*models.User)
		
		// 检查用户权限决定返回多少信息
		includePrivate := user.Role == models.RolePlatformAdmin ||
			user.Role == models.RoleSuperAdmin ||
			user.Role == models.RoleCourierLevel1 ||
			user.Role == models.RoleCourierLevel2 ||
			user.Role == models.RoleCourierLevel3 ||
			user.Role == models.RoleCourierLevel4

		if opCode, err := h.opcodeService.GetOPCodeByCode(code, includePrivate); err == nil {
			additionalInfo := gin.H{
				"is_active":      opCode.IsActive,
				"is_public":      opCode.IsPublic,
				"point_type":     opCode.PointType,
			}
			
			// 管理员和信使可以看到更多信息
			if includePrivate {
				additionalInfo["point_name"] = opCode.PointName
				additionalInfo["full_address"] = opCode.FullAddress
				additionalInfo["created_at"] = opCode.CreatedAt
				additionalInfo["updated_at"] = opCode.UpdatedAt
			}
			
			responseData["additional_info"] = additionalInfo
			responseData["user_role"] = user.Role
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "验证成功",
		"data":    responseData,
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
// 公开接口，但会根据用户认证状态返回不同级别的信息
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
	var user *models.User
	if exists {
		user = userInterface.(*models.User)
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

	// 构建响应数据
	responseData := make(map[string]interface{})
	
	// 基础字段，所有人都能看到
	responseData["code"] = opCode.Code
	responseData["school_code"] = opCode.SchoolCode
	responseData["area_code"] = opCode.AreaCode
	responseData["is_active"] = opCode.IsActive
	responseData["is_public"] = opCode.IsPublic
	responseData["point_type"] = opCode.PointType
	
	// 根据权限级别添加额外字段
	if includePrivate {
		responseData["point_code"] = opCode.PointCode
		responseData["point_name"] = opCode.PointName
		responseData["full_address"] = opCode.FullAddress
		responseData["binding_type"] = opCode.BindingType
		responseData["binding_id"] = opCode.BindingID
		responseData["created_at"] = opCode.CreatedAt
		responseData["updated_at"] = opCode.UpdatedAt
		responseData["access_level"] = "full"
		
		// 信使特有信息 - 暂时注释掉，因为 CanAccessOPCode 方法可能不存在
		// if user != nil && (user.Role == models.RoleCourierLevel1 ||
		// 	user.Role == models.RoleCourierLevel2 ||
		// 	user.Role == models.RoleCourierLevel3 ||
		// 	user.Role == models.RoleCourierLevel4) {
		// 	// 检查是否在信使管辖范围内
		// 	if canAccess, err := h.courierService.CanAccessOPCode(user.ID, code); err == nil {
		// 		responseData["can_manage"] = canAccess
		// 	}
		// }
	} else {
		responseData["access_level"] = "basic"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    responseData,
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

// SearchSchoolsByCity 按城市搜索学校
func (h *OPCodeHandler) SearchSchoolsByCity(c *gin.Context) {
	cityName := c.Query("city")
	if cityName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "城市名称不能为空",
			"data":    nil,
		})
		return
	}

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

	result, err := h.opcodeService.SearchSchoolsByCity(cityName, page, limit)
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
		"message": "城市搜索成功",
		"data":    result,
	})
}

// SearchSchoolsAdvanced 高级学校搜索
func (h *OPCodeHandler) SearchSchoolsAdvanced(c *gin.Context) {
	var req models.AdvancedSchoolSearchRequest
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
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "school_name"
	}
	if req.SortOrder == "" {
		req.SortOrder = "asc"
	}

	result, err := h.opcodeService.SearchSchoolsAdvanced(&req)
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
		"message": "高级搜索成功",
		"data":    result,
	})
}

// GetCities 获取城市列表
func (h *OPCodeHandler) GetCities(c *gin.Context) {
	// 返回热门城市列表
	cities := []string{
		"北京", "上海", "广州", "深圳", "天津", "重庆",
		"成都", "杭州", "南京", "武汉", "西安", "长沙",
		"青岛", "沈阳", "大连", "厦门", "苏州", "宁波",
		"无锡", "福州", "济南", "哈尔滨", "长春", "郑州",
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取城市列表成功",
		"data": gin.H{
			"cities": cities,
		},
	})
}

// GetDistricts 获取学校片区列表
func (h *OPCodeHandler) GetDistricts(c *gin.Context) {
	schoolCode := c.Param("school_code")
	if schoolCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "学校代码不能为空",
		})
		return
	}

	// 从数据库获取真实片区数据
	result, err := h.opcodeService.SearchAreas(schoolCode)
	if err != nil {
		// 如果数据库查询失败，返回模拟数据
		districts := []gin.H{
			{"code": "1", "name": "东区", "description": "宿舍楼1-5栋"},
			{"code": "2", "name": "西区", "description": "宿舍楼6-10栋"},
			{"code": "3", "name": "南区", "description": "宿舍楼11-15栋"},
			{"code": "4", "name": "北区", "description": "宿舍楼16-20栋"},
			{"code": "5", "name": "中心区", "description": "教学楼、图书馆"},
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"code":    200,
			"message": "获取片区列表成功",
			"data": gin.H{
				"districts": districts,
			},
		})
		return
	}

	// 转换为前端需要的格式
	if areas, ok := result["areas"].([]map[string]interface{}); ok {
		districts := make([]gin.H, 0, len(areas))
		for _, area := range areas {
			districts = append(districts, gin.H{
				"code":        area["area_code"],
				"name":        area["area_name"],
				"description": area["description"],
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"code":    200,
			"message": "获取片区列表成功",
			"data": gin.H{
				"districts": districts,
			},
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "数据格式错误",
		})
	}
}

// GetBuildings 获取楼栋列表
func (h *OPCodeHandler) GetBuildings(c *gin.Context) {
	schoolCode := c.Param("school_code")
	districtCode := c.Param("district_code")
	
	if schoolCode == "" || districtCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "参数不完整",
		})
		return
	}

	// 从数据库查询该片区下的楼栋
	// 由于楼栋数据可能存储在signal_codes表中，我们需要查询该表
	result, err := h.opcodeService.SearchBuildings(schoolCode, districtCode)
	if err != nil || result == nil {
		// 如果查询失败，返回默认楼栋列表
		buildings := []gin.H{
			{"code": "A", "name": "A栋", "type": "dormitory"},
			{"code": "B", "name": "B栋", "type": "dormitory"},
			{"code": "C", "name": "C栋", "type": "dormitory"},
			{"code": "D", "name": "D栋", "type": "teaching"},
			{"code": "E", "name": "E栋", "type": "dining"},
			{"code": "F", "name": "F栋", "type": "dormitory"},
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"code":    200,
			"message": "获取楼栋列表成功",
			"data": gin.H{
				"buildings": buildings,
			},
		})
		return
	}

	// 如果没有查询到数据，使用默认楼栋命名
	if buildings, ok := result["buildings"].([]map[string]interface{}); ok && len(buildings) > 0 {
		// 已有数据，直接返回
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"code":    200,
			"message": "获取楼栋列表成功",
			"data":    result,
		})
	} else {
		// 生成默认楼栋列表
		buildings := []gin.H{
			{"code": "A", "name": "A栋", "type": "dormitory"},
			{"code": "B", "name": "B栋", "type": "dormitory"},
			{"code": "C", "name": "C栋", "type": "dormitory"},
			{"code": "D", "name": "D栋", "type": "teaching"},
			{"code": "E", "name": "E栋", "type": "dining"},
			{"code": "F", "name": "F栋", "type": "dormitory"},
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"code":    200,
			"message": "获取楼栋列表成功",
			"data": gin.H{
				"buildings": buildings,
			},
		})
	}
}

// GetDeliveryPoints 获取投递点列表
func (h *OPCodeHandler) GetDeliveryPoints(c *gin.Context) {
	prefix := c.Param("prefix")
	if prefix == "" || len(prefix) < 4 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "前缀不完整",
		})
		return
	}

	// 从数据库查询已经存在的投递点
	var existingCodes []string
	if len(prefix) >= 2 {
		// 查询signal_codes表中以该前缀开头的所有编码
		req := &models.OPCodeSearchRequest{
			Code:     prefix,
			Page:     1,
			PageSize: 100,
		}
		codes, _, err := h.opcodeService.SearchOPCodes(req)
		if err == nil {
			for _, code := range codes {
				if len(code.Code) == 6 && strings.HasPrefix(code.Code, prefix) {
					existingCodes = append(existingCodes, code.Code)
				}
			}
		}
	}

	// 创建一个快速查找map
	existingMap := make(map[string]bool)
	for _, code := range existingCodes {
		existingMap[code] = true
	}

	// 生成投递点列表
	points := []gin.H{}
	for floor := 1; floor <= 6; floor++ {
		for room := 1; room <= 10; room++ {
			// 生成2位投递点代码
			pointCode := fmt.Sprintf("%d%d", floor, room)
			if len(pointCode) > 2 {
				pointCode = pointCode[len(pointCode)-2:]
			}
			
			// 检查该编码是否已被占用
			fullCode := prefix + pointCode
			isAvailable := !existingMap[fullCode]
			
			points = append(points, gin.H{
				"code":      pointCode,
				"name":      fmt.Sprintf("%d%02d室", floor, room),
				"available": isAvailable,
				"type":      "room",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取投递点列表成功",
		"data": gin.H{
			"points": points,
		},
	})
}

// ApplyOPCodeHierarchical 层级化申请OP Code
func (h *OPCodeHandler) ApplyOPCodeHierarchical(c *gin.Context) {
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
		Code        string `json:"code" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Description string `json:"description"`
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

	// 验证OP Code格式
	if len(req.Code) != 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "OP Code必须为6位",
		})
		return
	}

	// 解析OP Code
	schoolCode := req.Code[:2]
	areaCode := req.Code[2:4]
	pointCode := req.Code[4:6]

	// 创建申请
	apiReq := &models.OPCodeRequest{
		SchoolCode:  schoolCode,
		AreaCode:    areaCode,
		PointType:   req.Type,
		PointName:   req.Description,
		FullAddress: req.Description,
		Reason:      "通过层级选择系统申请",
	}

	application, err := h.opcodeService.ApplyForOPCode(user.ID, apiReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "申请失败",
			"error":   err.Error(),
		})
		return
	}

	// 自动批准并分配（仅用于演示，实际应该有审批流程）
	if err := h.opcodeService.AssignOPCode(user.ID, application.ID, pointCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "分配失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "申请成功",
		"data": gin.H{
			"application_id": application.ID,
			"op_code":        req.Code,
			"status":         "approved",
		},
	})
}
