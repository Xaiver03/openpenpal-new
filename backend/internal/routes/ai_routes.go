package routes

import (
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/handlers"
	"openpenpal-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupAIRoutes 设置AI相关路由
func SetupAIRoutes(router *gin.Engine, aiHandler *handlers.AIHandler, cfg *config.Config, db *gorm.DB) {
	// 公开AI API路由组
	aiAPI := router.Group("/api/ai")
	{
		// 文本生成和处理
		aiAPI.POST("/generate", aiHandler.GenerateText)
		aiAPI.POST("/chat", aiHandler.Chat)
		aiAPI.POST("/summarize", aiHandler.Summarize)
		aiAPI.POST("/translate", aiHandler.Translate)
		aiAPI.POST("/translate/batch", aiHandler.BatchTranslate)
		
		// 分析和审核
		aiAPI.POST("/sentiment", aiHandler.AnalyzeSentiment)
		aiAPI.POST("/moderate", aiHandler.ModerateContent)
		
		// 信件写作辅助
		aiAPI.POST("/letter/assist", aiHandler.LetterWritingAssist)
		
		// 系统状态（公开）
		aiAPI.GET("/providers/status", aiHandler.GetProviderStatus)
		
		// 使用统计（需要认证）
		aiAPI.GET("/usage/stats", middleware.AuthMiddleware(cfg, db), aiHandler.GetAIUsageStats)
	}

	// 管理员AI API路由组
	aiAdmin := router.Group("/api/admin/ai")
	aiAdmin.Use(middleware.AuthMiddleware(cfg, db))
	{
		// 提供商管理
		aiAdmin.POST("/providers/reload", aiHandler.ReloadProviders)
		
		// 配置管理
		aiAdmin.GET("/config", aiHandler.GetAIConfig)
		aiAdmin.PUT("/config", aiHandler.UpdateAIConfig)
		
		// 模板管理
		aiAdmin.GET("/templates", aiHandler.GetContentTemplates)
		aiAdmin.POST("/templates", aiHandler.CreateContentTemplate)
		
		// 监控和分析
		aiAdmin.GET("/monitoring", aiHandler.GetAIMonitoring)
		aiAdmin.GET("/analytics", aiHandler.GetAIAnalytics)
		aiAdmin.GET("/logs", aiHandler.GetAILogs)
		
		// 提供商测试
		aiAdmin.POST("/test-provider", aiHandler.TestAIProvider)
	}

	// 兼容现有的v1 AI路由
	v1 := router.Group("/api/v1/ai")
	v1.Use(middleware.AuthMiddleware(cfg, db))
	{
		// 现有AI功能（保持向后兼容）
		v1.POST("/match", aiHandler.MatchPenPal)
		v1.POST("/reply", aiHandler.GenerateReply)
		v1.POST("/reply-advice", aiHandler.GenerateReplyAdvice)
		v1.POST("/inspiration", aiHandler.GetInspiration)
		v1.POST("/curate", aiHandler.CurateLetters)
		
		// 状态和统计
		v1.GET("/stats", aiHandler.GetAIStats)
		v1.GET("/personas", aiHandler.GetPersonas)
		v1.GET("/daily-inspiration", aiHandler.GetDailyInspiration)
	}

	// 管理员v1 AI路由
	v1Admin := router.Group("/api/v1/admin/ai")
	v1Admin.Use(middleware.AuthMiddleware(cfg, db))
	{
		v1Admin.GET("/config", aiHandler.GetAIConfig)
		v1Admin.PUT("/config", aiHandler.UpdateAIConfig)
		v1Admin.GET("/templates", aiHandler.GetContentTemplates)
		v1Admin.POST("/templates", aiHandler.CreateContentTemplate)
		v1Admin.GET("/monitoring", aiHandler.GetAIMonitoring)
		v1Admin.GET("/analytics", aiHandler.GetAIAnalytics)
		v1Admin.GET("/logs", aiHandler.GetAILogs)
		v1Admin.POST("/test-provider", aiHandler.TestAIProvider)
	}
}

// SetupAIWebSocketRoutes 设置AI WebSocket路由
func SetupAIWebSocketRoutes(router *gin.Engine, aiHandler *handlers.AIHandler, cfg *config.Config, db *gorm.DB) {
	// WebSocket路由用于实时AI交互
	ws := router.Group("/ws/ai")
	ws.Use(middleware.AuthMiddleware(cfg, db))
	{
		// 实时聊天
		ws.GET("/chat", func(c *gin.Context) {
			// TODO: 实现WebSocket聊天处理器
			c.JSON(200, gin.H{"message": "WebSocket AI chat endpoint - implementation pending"})
		})
		
		// 实时文本生成
		ws.GET("/generate", func(c *gin.Context) {
			// TODO: 实现WebSocket文本生成处理器
			c.JSON(200, gin.H{"message": "WebSocket AI generation endpoint - implementation pending"})
		})
	}
}