/**
 * 共享响应包集成 - 将内部响应格式迁移到共享包
 * 提供向后兼容的API，同时使用SOTA共享响应实现
 */

package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"shared/pkg/response"
)

// MigrateToSharedResponse 启用共享响应包的标志
var MigrateToSharedResponse = true

// SharedSuccessResponse 使用共享包的成功响应
func SharedSuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	if !MigrateToSharedResponse {
		// 回退到原始实现
		SuccessResponse(c, statusCode, message, data)
		return
	}

	switch statusCode {
	case http.StatusCreated:
		response.Created(c, data, message)
	case http.StatusNoContent:
		response.NoContent(c)
	default:
		if message != "" {
			response.SuccessWithMessage(c, data, message)
		} else {
			response.Success(c, data)
		}
	}
}

// SharedErrorResponse 使用共享包的错误响应
func SharedErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	if !MigrateToSharedResponse {
		// 回退到原始实现
		ErrorResponse(c, statusCode, message, err)
		return
	}

	errorMessage := message
	if err != nil {
		errorMessage = message + ": " + err.Error()
	}

	switch statusCode {
	case http.StatusBadRequest:
		response.BadRequest(c, errorMessage)
	case http.StatusUnauthorized:
		response.Unauthorized(c, errorMessage)
	case http.StatusForbidden:
		response.Forbidden(c, errorMessage)
	case http.StatusNotFound:
		response.NotFound(c, errorMessage)
	case http.StatusConflict:
		response.Conflict(c, errorMessage)
	case http.StatusUnprocessableEntity:
		response.UnprocessableEntity(c, errorMessage)
	case http.StatusTooManyRequests:
		response.TooManyRequests(c, errorMessage)
	case http.StatusInternalServerError:
		response.InternalServerError(c, errorMessage)
	case http.StatusBadGateway:
		response.BadGateway(c, errorMessage)
	case http.StatusServiceUnavailable:
		response.ServiceUnavailable(c, errorMessage)
	case http.StatusGatewayTimeout:
		response.GatewayTimeout(c, errorMessage)
	default:
		response.Error(c, statusCode, errorMessage)
	}
}

// SharedBadRequestResponse 使用共享包的400错误响应
func SharedBadRequestResponse(c *gin.Context, message string, err error) {
	SharedErrorResponse(c, http.StatusBadRequest, message, err)
}

// SharedUnauthorizedResponse 使用共享包的401错误响应  
func SharedUnauthorizedResponse(c *gin.Context, message string) {
	SharedErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// SharedForbiddenResponse 使用共享包的403错误响应
func SharedForbiddenResponse(c *gin.Context, message string) {
	SharedErrorResponse(c, http.StatusForbidden, message, nil)
}

// SharedNotFoundResponse 使用共享包的404错误响应
func SharedNotFoundResponse(c *gin.Context, message string) {
	SharedErrorResponse(c, http.StatusNotFound, message, nil)
}

// SharedConflictResponse 使用共享包的409错误响应
func SharedConflictResponse(c *gin.Context, message string) {
	SharedErrorResponse(c, http.StatusConflict, message, nil)
}

// SharedInternalServerErrorResponse 使用共享包的500错误响应
func SharedInternalServerErrorResponse(c *gin.Context, message string, err error) {
	SharedErrorResponse(c, http.StatusInternalServerError, message, err)
}

// ================================
// 向后兼容的别名函数 - 渐进式迁移
// ================================

// EnableSharedResponsePackage 启用共享响应包
func EnableSharedResponsePackage() {
	MigrateToSharedResponse = true
}

// DisableSharedResponsePackage 禁用共享响应包（回退到原始实现）
func DisableSharedResponsePackage() {
	MigrateToSharedResponse = false
}

// 业务特定响应
func SharedPermissionDeniedResponse(c *gin.Context, permission string) {
	if MigrateToSharedResponse {
		response.PermissionDenied(c, permission)
	} else {
		ErrorResponse(c, http.StatusForbidden, "权限不足: "+permission, nil)
	}
}

func SharedResourceNotFoundResponse(c *gin.Context, resource string) {
	if MigrateToSharedResponse {
		response.ResourceNotFound(c, resource)
	} else {
		ErrorResponse(c, http.StatusNotFound, resource+"未找到", nil)
	}
}

func SharedDataConflictResponse(c *gin.Context, message string) {
	if MigrateToSharedResponse {
		response.DataConflict(c, message)
	} else {
		ErrorResponse(c, http.StatusConflict, "数据冲突: "+message, nil)
	}
}

func SharedRateLimitExceededResponse(c *gin.Context) {
	if MigrateToSharedResponse {
		response.RateLimitExceeded(c)
	} else {
		ErrorResponse(c, http.StatusTooManyRequests, "请求频率过高", nil)
	}
}

func SharedMaintenanceModeResponse(c *gin.Context) {
	if MigrateToSharedResponse {
		response.MaintenanceMode(c)
	} else {
		ErrorResponse(c, http.StatusServiceUnavailable, "系统维护中", nil)
	}
}

// ================================
// 渐进式迁移助手
// ================================

// BeginResponseMigration 开始响应包迁移的助手函数
func BeginResponseMigration() {
	EnableSharedResponsePackage()
}

// GetMigrationStatus 获取迁移状态
func GetMigrationStatus() bool {
	return MigrateToSharedResponse
}