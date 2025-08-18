package handlers

import (
	"net/http"
	"strconv"
	"time"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// BatchHandler 批量条码管理处理器 - 桥接信使服务API
type BatchHandler struct {
	letterService  *services.LetterService
	courierService *services.CourierService
	opcodeService  *services.OPCodeService
}

// NewBatchHandler 创建批量管理处理器
func NewBatchHandler(letterService *services.LetterService, courierService *services.CourierService, opcodeService *services.OPCodeService) *BatchHandler {
	return &BatchHandler{
		letterService:  letterService,
		courierService: courierService,
		opcodeService:  opcodeService,
	}
}

// BatchGenerateRequest 批量生成请求
type BatchGenerateRequest struct {
	BatchNo      string `json:"batch_no" binding:"required"`
	SchoolCode   string `json:"school_code" binding:"required"`
	AreaCode     string `json:"area_code,omitempty"`
	Quantity     int    `json:"quantity" binding:"required,min=1,max=1000"`
	CodeType     string `json:"code_type" binding:"required"`
	Description  string `json:"description,omitempty"`
	OperatorID   string `json:"operator_id,omitempty"`
}

// BatchRecord 批次记录
type BatchRecord struct {
	ID             string    `json:"id"`
	BatchNo        string    `json:"batch_no"`
	SchoolCode     string    `json:"school_code"`
	AreaCode       string    `json:"area_code,omitempty"`
	Quantity       int       `json:"quantity"`
	GeneratedCount int       `json:"generated_count"`
	UsedCount      int       `json:"used_count"`
	CodeType       string    `json:"code_type"`
	Status         string    `json:"status"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Description    string    `json:"description,omitempty"`
	DownloadURL    string    `json:"download_url,omitempty"`
}

// GenerateBatch 批量生成条码 - POST /courier/batch/generate
func (h *BatchHandler) GenerateBatch(c *gin.Context) {
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

	// 权限检查：只有L3+信使可以批量生成
	if user.Role != models.RoleCourierLevel3 && user.Role != models.RoleCourierLevel4 && 
		user.Role != models.RolePlatformAdmin && user.Role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "需要三级及以上信使权限",
		})
		return
	}

	var req BatchGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// L3信使只能为自己学校生成
	if user.Role == models.RoleCourierLevel3 {
		// TODO: 从用户信息获取学校代码进行验证
		// if req.SchoolCode != user.SchoolCode {
		//     c.JSON(http.StatusForbidden, gin.H{"message": "只能为所属学校生成条码"})
		//     return
		// }
	}

	// 批量生成条码
	batch := &BatchRecord{
		ID:             generateBatchID(),
		BatchNo:        req.BatchNo,
		SchoolCode:     req.SchoolCode,
		AreaCode:       req.AreaCode,
		Quantity:       req.Quantity,
		GeneratedCount: req.Quantity, // 模拟生成成功
		UsedCount:      0,
		CodeType:       req.CodeType,
		Status:         "active",
		CreatedBy:      user.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Description:    req.Description,
	}

	// TODO: 实际的批量生成逻辑
	// 可以调用信使服务的API或直接在数据库中创建条码记录

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "批次生成成功",
		"data": gin.H{
			"batch_no":        batch.BatchNo,
			"generated_count": batch.GeneratedCount,
			"batch_id":        batch.ID,
		},
	})
}

// GetBatches 获取批次列表 - GET /courier/batch
func (h *BatchHandler) GetBatches(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	schoolCode := c.Query("school_code")

	// L3信使只能查看自己学校的批次
	if user.Role == models.RoleCourierLevel3 && schoolCode == "" {
		// schoolCode = user.SchoolCode // TODO: 从用户信息获取
		schoolCode = "BJDX" // 临时默认值
	}

	// 模拟批次数据
	batches := []BatchRecord{
		{
			ID:             "batch_001",
			BatchNo:        "B20250118001",
			SchoolCode:     "BJDX",
			Quantity:       500,
			GeneratedCount: 500,
			UsedCount:      123,
			CodeType:       "normal",
			Status:         "active",
			CreatedBy:      user.ID,
			CreatedAt:      time.Now().Add(-24 * time.Hour),
			UpdatedAt:      time.Now().Add(-1 * time.Hour),
			Description:    "北京大学批次001",
		},
		{
			ID:             "batch_002", 
			BatchNo:        "B20250118002",
			SchoolCode:     "QHDX",
			Quantity:       300,
			GeneratedCount: 300,
			UsedCount:      45,
			CodeType:       "drift",
			Status:         "active",
			CreatedBy:      user.ID,
			CreatedAt:      time.Now().Add(-12 * time.Hour),
			UpdatedAt:      time.Now().Add(-30 * time.Minute),
			Description:    "清华大学漂流信批次",
		},
	}

	// 根据权限过滤
	if schoolCode != "" {
		filtered := make([]BatchRecord, 0)
		for _, batch := range batches {
			if batch.SchoolCode == schoolCode {
				filtered = append(filtered, batch)
			}
		}
		batches = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取批次列表成功",
		"data": gin.H{
			"batches": batches,
			"total":   len(batches),
			"page":    page,
			"limit":   limit,
		},
	})
}

// GetBatchDetails 获取批次详情 - GET /courier/batch/:id
func (h *BatchHandler) GetBatchDetails(c *gin.Context) {
	batchID := c.Param("id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "批次ID不能为空",
		})
		return
	}

	// 模拟批次详情
	batch := BatchRecord{
		ID:             batchID,
		BatchNo:        "B20250118001",
		SchoolCode:     "BJDX", 
		Quantity:       500,
		GeneratedCount: 500,
		UsedCount:      123,
		CodeType:       "normal",
		Status:         "active",
		CreatedBy:      "user_123",
		CreatedAt:      time.Now().Add(-24 * time.Hour),
		UpdatedAt:      time.Now().Add(-1 * time.Hour),
		Description:    "北京大学批次001",
	}

	// 模拟条码列表（前几个）
	codes := []map[string]interface{}{
		{
			"id":         "code_001",
			"code":       "OPP-BJFU-5F3D-01",
			"status":     "unactivated",
			"batch_id":   batchID,
			"created_at": time.Now().Add(-24 * time.Hour),
		},
		{
			"id":         "code_002", 
			"code":       "OPP-BJFU-5F3D-02",
			"status":     "bound",
			"batch_id":   batchID,
			"bound_at":   time.Now().Add(-2 * time.Hour),
			"created_at": time.Now().Add(-24 * time.Hour),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取批次详情成功",
		"data": gin.H{
			"batch": batch,
			"codes": codes,
		},
	})
}

// GetBatchStats 获取批次统计 - GET /courier/batch/stats
func (h *BatchHandler) GetBatchStats(c *gin.Context) {
	_, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未认证",
		})
		return
	}

	schoolCode := c.Query("school_code")
	
	// 模拟统计数据
	stats := gin.H{
		"total_batches": 15,
		"total_codes":   7500,
		"used_codes":    2340,
		"active_batches": 12,
		"expired_batches": 2,
		"usage_rate":     0.312,
		"batches_by_type": gin.H{
			"normal": 12,
			"drift":  3,
		},
		"usage_by_school": []gin.H{
			{
				"school_code": "BJDX",
				"total":       3000,
				"used":        934,
			},
			{
				"school_code": "QHDX", 
				"total":       2500,
				"used":        756,
			},
		},
	}

	if schoolCode != "" {
		// 按学校过滤统计
		stats["school_filter"] = schoolCode
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取统计信息成功", 
		"data":    stats,
	})
}

// UpdateBatchStatus 更新批次状态 - PATCH /courier/batch/:id/status
func (h *BatchHandler) UpdateBatchStatus(c *gin.Context) {
	batchID := c.Param("id")
	
	var req struct {
		Status string `json:"status" binding:"required,oneof=active completed expired suspended"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 模拟更新操作
	batch := BatchRecord{
		ID:        batchID,
		Status:    req.Status,
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "批次状态更新成功",
		"data":    batch,
	})
}

// 辅助函数
func generateBatchID() string {
	return "batch_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}