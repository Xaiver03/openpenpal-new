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

// StorageHandler 存储处理器
type StorageHandler struct {
	storageService *services.StorageService
}

// NewStorageHandler 创建存储处理器实例
func NewStorageHandler(storageService *services.StorageService) *StorageHandler {
	return &StorageHandler{
		storageService: storageService,
	}
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到存储系统
// @Tags storage
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上传的文件"
// @Param category formData string true "文件分类"
// @Param related_type formData string false "关联类型"
// @Param related_id formData string false "关联ID"
// @Param is_public formData bool false "是否公开" default(false)
// @Param expires_in formData int false "过期时间（秒）"
// @Success 200 {object} models.UploadResponse "上传成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 413 {object} map[string]interface{} "文件过大"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/storage/upload [post]
func (h *StorageHandler) UploadFile(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "获取上传文件失败",
			"details": err.Error(),
		})
		return
	}

	// 解析请求参数
	req := &models.UploadRequest{}
	req.Category = models.FileCategory(c.PostForm("category"))
	req.RelatedType = c.PostForm("related_type")
	req.RelatedID = c.PostForm("related_id")

	if isPublicStr := c.PostForm("is_public"); isPublicStr != "" {
		req.IsPublic, _ = strconv.ParseBool(isPublicStr)
	}

	if expiresInStr := c.PostForm("expires_in"); expiresInStr != "" {
		req.ExpiresIn, _ = strconv.Atoi(expiresInStr)
	}

	// 验证必需参数
	if req.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件分类不能为空",
		})
		return
	}

	// 上传文件
	response, err := h.storageService.UploadFile(file, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "文件上传失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetFile 获取文件信息
// @Summary 获取文件信息
// @Description 根据文件ID获取文件详细信息
// @Tags storage
// @Accept json
// @Produce json
// @Param file_id path string true "文件ID"
// @Success 200 {object} models.StorageFile "文件信息"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/storage/files/{file_id} [get]
func (h *StorageHandler) GetFile(c *gin.Context) {
	fileID := c.Param("file_id")

	file, err := h.storageService.GetFile(fileID)
	if err != nil {
		if err.Error() == "文件不存在" || err.Error() == "文件已过期" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取文件信息失败",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, file)
}

// GetFiles 获取文件列表
// @Summary 获取文件列表
// @Description 根据查询条件获取文件列表
// @Tags storage
// @Accept json
// @Produce json
// @Param category query string false "文件分类"
// @Param provider query string false "存储提供商"
// @Param status query string false "文件状态"
// @Param related_type query string false "关联类型"
// @Param related_id query string false "关联ID"
// @Param uploaded_by query string false "上传者ID"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段"
// @Param sort_order query string false "排序方向"
// @Success 200 {object} map[string]interface{} "文件列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/storage/files [get]
func (h *StorageHandler) GetFiles(c *gin.Context) {
	query := &models.FileQuery{
		Page:     1,
		PageSize: 20,
	}

	// 解析查询参数
	if category := c.Query("category"); category != "" {
		query.Category = models.FileCategory(category)
	}
	if provider := c.Query("provider"); provider != "" {
		query.Provider = models.StorageProvider(provider)
	}
	if status := c.Query("status"); status != "" {
		query.Status = models.FileStatus(status)
	}
	query.RelatedType = c.Query("related_type")
	query.RelatedID = c.Query("related_id")
	query.UploadedBy = c.Query("uploaded_by")

	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			query.StartDate = &parsed
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			query.EndDate = &parsed
		}
	}

	if page := c.Query("page"); page != "" {
		if parsed, err := strconv.Atoi(page); err == nil && parsed > 0 {
			query.Page = parsed
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if parsed, err := strconv.Atoi(pageSize); err == nil && parsed > 0 && parsed <= 100 {
			query.PageSize = parsed
		}
	}

	query.SortBy = c.Query("sort_by")
	query.SortOrder = c.Query("sort_order")

	files, total, err := h.storageService.GetFiles(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取文件列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"pagination": gin.H{
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total":       total,
			"total_pages": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
		},
	})
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除指定的文件
// @Tags storage
// @Accept json
// @Produce json
// @Param file_id path string true "文件ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/storage/files/{file_id} [delete]
func (h *StorageHandler) DeleteFile(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	fileID := c.Param("file_id")

	err := h.storageService.DeleteFile(fileID, userID)
	if err != nil {
		if err.Error() == "文件不存在" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "删除文件失败",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "文件删除成功",
	})
}

// GetStorageStats 获取存储统计信息
// @Summary 获取存储统计信息
// @Description 获取存储系统的统计数据
// @Tags storage
// @Accept json
// @Produce json
// @Success 200 {object} models.StorageStats "存储统计信息"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/storage/stats [get]
func (h *StorageHandler) GetStorageStats(c *gin.Context) {
	stats, err := h.storageService.GetStorageStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取存储统计失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// DownloadFile 下载文件
// @Summary 下载文件
// @Description 下载指定的文件
// @Tags storage
// @Accept json
// @Produce application/octet-stream
// @Param file_id path string true "文件ID"
// @Success 200 "文件内容"
// @Failure 404 {object} map[string]interface{} "文件不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/storage/files/{file_id}/download [get]
func (h *StorageHandler) DownloadFile(c *gin.Context) {
	fileID := c.Param("file_id")

	// 获取文件信息
	file, err := h.storageService.GetFile(fileID)
	if err != nil {
		if err.Error() == "文件不存在" || err.Error() == "文件已过期" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "获取文件信息失败",
				"details": err.Error(),
			})
		}
		return
	}

	// 如果是公开文件，直接重定向到公共URL
	if file.IsPublic && file.PublicURL != "" {
		c.Redirect(http.StatusFound, file.PublicURL)
		return
	}

	// 对于私有文件，需要通过服务下载（这里简化处理）
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":    "私有文件下载功能暂未实现",
		"file_url": file.PublicURL,
	})
}
