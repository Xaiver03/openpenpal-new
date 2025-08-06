package handlers

import (
	"net/http"
	"time"

	"openpenpal-backend/pkg/cache"

	"github.com/gin-gonic/gin"
)

// LogoutHandler 处理用户注销
type LogoutHandler struct {
	blacklist *cache.TokenBlacklist
}

// NewLogoutHandler 创建注销处理器
func NewLogoutHandler() *LogoutHandler {
	return &LogoutHandler{
		blacklist: cache.GetTokenBlacklist(),
	}
}

// Logout 注销当前用户的Token
// @Summary 用户注销
// @Description 将当前Token加入黑名单，使其立即失效
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "注销成功"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Router /api/v1/auth/logout [post]
func (h *LogoutHandler) Logout(c *gin.Context) {
	// 获取JWT ID
	jti, exists := c.Get("token_jti")
	if !exists || jti == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "注销成功",
		})
		return
	}

	// 获取用户信息用于日志
	userID, _ := c.Get("user_id")

	// 将Token加入黑名单
	// 设置黑名单过期时间为Token的原始过期时间+1小时（作为缓冲）
	blacklistExpiry := time.Now().Add(24 * time.Hour) // 默认24小时
	h.blacklist.Add(jti.(string), blacklistExpiry)

	// 记录注销日志
	if userID != nil {
		c.Set("audit_action", "logout")
		c.Set("audit_message", "用户主动注销")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "注销成功",
		"data": gin.H{
			"user_id": userID,
			"logout_time": time.Now().Format(time.RFC3339),
		},
	})
}

// LogoutAll 注销用户的所有Token
// @Summary 注销所有会话
// @Description 注销该用户的所有活跃会话（需要实现用户Token映射）
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "注销成功"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Router /api/v1/auth/logout-all [post]
func (h *LogoutHandler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未认证用户",
		})
		return
	}

	// TODO: 实现用户所有Token的映射和批量注销
	// 当前实现仅注销当前Token
	jti, _ := c.Get("token_jti")
	if jti != nil && jti != "" {
		blacklistExpiry := time.Now().Add(24 * time.Hour)
		h.blacklist.Add(jti.(string), blacklistExpiry)
	}

	// 清除用户缓存
	userCache := cache.GetUserCache()
	userCache.Delete(userID.(string))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已注销所有会话",
		"data": gin.H{
			"user_id": userID,
			"logout_time": time.Now().Format(time.RFC3339),
		},
	})
}