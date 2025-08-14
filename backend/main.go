package main

import (
	"fmt"
	"log"
	"time"

	"openpenpal-backend/internal/adapters"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/handlers"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/routes"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化数据库
	// 使用直接的数据库连接方法，绕过 shared 包的配置问题
	db, err := config.SetupDatabaseDirect(cfg)
	if err != nil {
		log.Fatal("Failed to setup database:", err)
	}

	// 在开发环境下初始化测试数据
	if cfg.Environment == "development" {
		// 重新启用数据种子功能
		if err := config.SeedData(db); err != nil {
			log.Printf("Warning: Failed to seed data: %v", err)
		} else {
			log.Printf("Test data seeded successfully")
		}
	}

	// 初始化服务
	userService := services.NewUserService(db, cfg)
	letterService := services.NewLetterService(db, cfg)
	envelopeService := services.NewEnvelopeService(db)
	courierService := services.NewCourierService(db)
	museumService := services.NewMuseumService(db)
	aiService := services.NewAIService(db, cfg)
	configService := services.NewConfigService(db)
	notificationService := services.NewNotificationService(db, cfg)
	analyticsService := services.NewAnalyticsService(db)
	schedulerService := services.NewSchedulerService(db)
	creditService := services.NewCreditService(db)
	courierTaskService := services.NewCourierTaskService(db)
	adminService := services.NewAdminService(db, cfg)
	systemSettingsService := services.NewSystemSettingsService(db, cfg)
	storageService := services.NewStorageService(db, cfg)
	moderationService := services.NewModerationService(db, cfg, aiService)
	shopService := services.NewShopService(db)
	commentService := services.NewCommentService(db, cfg)
	followService := services.NewFollowService(db) // 关注系统服务
	opcodeService := services.NewOPCodeService(db) // OP Code服务 - 重新启用
	scanEventService := services.NewScanEventService(db) // 扫描事件服务 - PRD要求

	// 初始化延迟队列服务
	delayQueueService, err := services.NewDelayQueueService(db, cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize delay queue service: %v", err)
	} else {
		// 在后台启动延迟队列工作进程
		go delayQueueService.StartWorker()
		log.Println("Delay queue worker started")
	}

	// 初始化WebSocket服务
	wsService := websocket.NewWebSocketService()
	wsService.Start()
	
	// 创建WebSocket适配器用于服务间通信
	wsAdapter := websocket.NewWebSocketAdapter(wsService)

	// Configure service dependencies with credit system integration
	letterService.SetCreditService(creditService)
	letterService.SetCourierTaskService(courierTaskService)
	letterService.SetNotificationService(notificationService)
	letterService.SetWebSocketService(wsService)
	letterService.SetAIService(aiService)
	letterService.SetOPCodeService(opcodeService) // PRD要求：集成OP Code验证
	envelopeService.SetCreditService(creditService)
	envelopeService.SetUserService(userService) // FSD增强：OP Code区域验证
	museumService.SetCreditService(creditService)
	museumService.SetNotificationService(notificationService)
	museumService.SetAIService(aiService)
	courierTaskService.SetNotificationService(notificationService)
	courierService.SetWebSocketService(wsAdapter) // SOTA: Dependency Injection for real-time notifications
	notificationService.SetWebSocketService(wsService)
	commentService.SetLetterService(letterService)
	commentService.SetCreditService(creditService)
	commentService.SetModerationService(moderationService)

	// 启动任务调度服务
	if err := schedulerService.Start(); err != nil {
		log.Printf("Warning: Failed to start scheduler service: %v", err)
	} else {
		log.Println("Scheduler service started successfully")
	}

	// 初始化处理器
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, cfg) // 新增：专门的认证处理器
	letterHandler := handlers.NewLetterHandler(letterService, envelopeService)
	envelopeHandler := handlers.NewEnvelopeHandler(envelopeService)
	courierHandler := handlers.NewCourierHandler(courierService)
	promotionService := services.NewPromotionService(db)
	courierGrowthHandler := handlers.NewCourierGrowthHandler(courierService, userService, promotionService)
	museumHandler := handlers.NewMuseumHandler(museumService)
	aiHandler := handlers.NewAIHandler(aiService, configService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	schedulerHandler := handlers.NewSchedulerHandler(schedulerService)
	creditHandler := handlers.NewCreditHandler(creditService)
	adminHandler := handlers.NewAdminHandler(adminService)
	systemSettingsHandler := handlers.NewSystemSettingsHandler(systemSettingsService, cfg)
	storageHandler := handlers.NewStorageHandler(storageService)
	moderationHandler := handlers.NewModerationHandler(moderationService)
	opcodeHandler := handlers.NewOPCodeHandler(opcodeService, courierService)
	barcodeHandler := handlers.NewBarcodeHandler(letterService, opcodeService, scanEventService) // PRD条码系统处理器
	scanEventHandler := handlers.NewScanEventHandler(scanEventService) // 扫描事件处理器
	shopHandler := handlers.NewShopHandler(shopService, userService)
	commentHandler := handlers.NewCommentHandler(commentService)
	followHandler := handlers.NewFollowHandler(followService) // 关注系统处理器
	userProfileHandler := handlers.NewUserProfileHandler(db) // 用户档案处理器
	
	// QR扫描服务和处理器 - SOTA集成：复用现有依赖 - Temporarily disabled
	// qrScanService := services.NewQRScanService(db, letterService, courierService, wsAdapter)
	// qrScanHandler := handlers.NewQRScanHandler(qrScanService, middleware.NewAuthMiddleware(cfg, db)) - Temporarily disabled
	wsHandler := wsService.GetHandler()

	// SOTA管理API适配器 - 兼容Java前端
	adminAdapter := adapters.NewAdminAdapter(
		adminHandler,
		userHandler,
		letterHandler,
		courierHandler,
		museumHandler,
		adminService,
		userService,
		letterService,
		courierService,
		museumService,
	)

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由器
	router := gin.New()

	// 全局中间件 - SOTA级别配置
	// 1. 基础中间件
	router.Use(middleware.RequestIDMiddleware())    // 请求追踪
	router.Use(middleware.LoggerMiddleware())       // 日志记录
	router.Use(middleware.RecoveryMiddleware())     // 错误恢复
	router.Use(middleware.MetricsMiddleware())      // 性能监控
	
	// 2. 安全中间件
	router.Use(middleware.SecurityHeadersMiddleware())                    // 安全头
	router.Use(middleware.CORSMiddleware())                              // CORS
	router.Use(middleware.RequestSizeLimitMiddleware(middleware.DefaultMaxRequestSize)) // 请求大小限制
	
	// 3. 频率限制（IP级别作为默认）
	router.Use(middleware.RateLimitMiddleware())
	
	// 4. API转换中间件 - SOTA实现
	router.Use(middleware.RequestTransformMiddleware())  // 请求转换 (camelCase -> snake_case)
	router.Use(middleware.ResponseTransformMiddleware()) // 响应转换 (snake_case -> camelCase)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		// 检查数据库连接
		var dbStatus string
		sqlDB, err := db.DB()
		if err != nil {
			dbStatus = "unhealthy: " + err.Error()
		} else if err := sqlDB.Ping(); err != nil {
			dbStatus = "unhealthy: " + err.Error()
		} else {
			dbStatus = "healthy"
		}

		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "openpenpal-backend",
			"version":   cfg.AppVersion,
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  dbStatus,
			"websocket": "healthy",
			"env":       cfg.Environment,
		})
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 设置API路由别名 - SOTA解决方案
	routes.SetupAPIAliases(router, userProfileHandler)

	// API版本组
	v1 := router.Group("/api/v1")

	// 公开路由（无需认证）
	public := v1.Group("/")
	{
		// 用户认证 - 使用新的认证处理器
		auth := public.Group("/auth")
		auth.Use(middleware.AuthRateLimitMiddleware())
		{
			// 基础认证端点（无需CSRF）
			auth.GET("/csrf", authHandler.GetCSRFToken)        // 获取CSRF令牌
			
			// 状态改变操作需要CSRF保护
			csrfProtected := auth.Group("/")
			csrfProtected.Use(middleware.CSRFMiddleware())
			{
				csrfProtected.POST("/register", authHandler.Register)       // 用户注册
				csrfProtected.POST("/login", authHandler.Login)             // 用户登录
			}
			
			// 用户信息端点（需要认证）
			authGroup := auth.Group("/")
			authGroup.Use(middleware.AuthMiddleware(cfg, db))
			{
				authGroup.GET("/me", authHandler.GetCurrentUser)         // 获取当前用户信息
				authGroup.POST("/logout", authHandler.Logout)            // 登出
				authGroup.POST("/refresh", authHandler.RefreshToken)     // 刷新令牌
				authGroup.GET("/check-expiry", authHandler.CheckTokenExpiry) // 检查令牌过期
			}
		}

		// 公开的信件读取
		letters := public.Group("/letters")
		{
			letters.GET("/read/:code", letterHandler.GetLetterByCode)
			letters.POST("/read/:code/mark-read", letterHandler.MarkAsRead)
			letters.GET("/public", letterHandler.GetPublicLetters) // 新增：广场信件
			letters.GET("/popular", letterHandler.GetPopularLetters) // 新增：热门信件
			letters.GET("/recommended", letterHandler.GetRecommendedLetters) // 新增：推荐信件
			letters.GET("/templates", letterHandler.GetLetterTemplates) // 新增：信件模板（公开）
		}

		// 公开的信使统计信息
		courier := public.Group("/courier")
		{
			courier.GET("/stats", courierHandler.GetCourierStats)
		}

		// 公开的博物馆信息
		museum := public.Group("/museum")
		{
			museum.GET("/entries", museumHandler.GetMuseumEntries)
			museum.GET("/entries/:id", museumHandler.GetMuseumEntry)
			museum.GET("/exhibitions", museumHandler.GetMuseumExhibitions)
			museum.GET("/popular", museumHandler.GetPopularMuseumEntries)        // 获取热门条目
			museum.GET("/exhibitions/:id", museumHandler.GetMuseumExhibitionByID) // 获取展览详情
			museum.GET("/exhibitions/:id/items", museumHandler.GetExhibitionItems) // 获取展览中的物品
			museum.GET("/tags", museumHandler.GetMuseumTags)                     // 获取标签列表
			museum.GET("/stats", museumHandler.GetMuseumStats)                   // 获取博物馆统计
		}

		// 公开的AI相关（无需认证）
		ai := public.Group("/ai")
		{
			ai.POST("/match", aiHandler.MatchPenPal)
			ai.POST("/reply", aiHandler.GenerateReply)
			ai.POST("/reply-advice", aiHandler.GenerateReplyAdvice) // 新增：生成回信角度建议
			ai.POST("/inspiration", aiHandler.GetInspiration)
			ai.POST("/curate", aiHandler.CurateLetters)
			ai.GET("/personas", aiHandler.GetPersonas)
			ai.GET("/stats", aiHandler.GetAIStats)
			ai.GET("/daily-inspiration", aiHandler.GetDailyInspiration)
		}
		
		// 公开的商店信息（无需认证）
		shop := public.Group("/shop")
		{
			shop.GET("/products", shopHandler.GetProducts)                   // 获取商品列表（公开）
			shop.GET("/products/:id", shopHandler.GetProduct)               // 获取商品详情（公开）
			shop.GET("/products/:id/reviews", shopHandler.GetProductReviews) // 获取商品评价（公开）
		}
		
		// 公开的OP Code查询（仅公开信息） - Temporarily disabled
		/*
		opcode := public.Group("/opcode")
		{
			opcode.GET("/:code", opcodeHandler.GetOPCode)                        // 查询OP Code公开信息
			opcode.GET("/validate", opcodeHandler.ValidateOPCode)                // 验证OP Code格式
		}

		// 公开的QR码验证（无需认证） - Temporarily disabled
		qr := public.Group("/qr")
		{
			qr.GET("/validate", qrScanHandler.ValidateQRCode)                    // 验证QR码格式
		}
		*/

	}

	// 需要认证的路由
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg, db))
	{
		// 注意：/auth相关的路由已经在上面的authHandler中处理了，这里不重复定义
		
		// 用户相关
		users := protected.Group("/users")
		users.Use(middleware.UserRateLimitMiddleware()) // 用户级频率限制
		{
			users.GET("/me", userHandler.GetProfile)
			users.PUT("/me", userHandler.UpdateProfile)
			users.POST("/me/change-password", userHandler.ChangePassword)
			users.GET("/me/stats", userHandler.GetUserStats)
			users.DELETE("/me", userHandler.DeactivateAccount)
			users.POST("/avatar", userHandler.UploadAvatar)
			users.DELETE("/avatar", userHandler.RemoveAvatar)
		}

		// 信件相关
		letters := protected.Group("/letters")
		{
			letters.POST("/", letterHandler.CreateDraft)
			letters.GET("/", letterHandler.GetUserLetters)
			letters.GET("/stats", letterHandler.GetUserStats)
			letters.GET("/:id", letterHandler.GetLetter)
			letters.PUT("/:id", letterHandler.UpdateLetter)
			letters.DELETE("/:id", letterHandler.DeleteLetter)
			letters.POST("/:id/generate-code", letterHandler.GenerateCode)

			// 信封绑定相关
			letters.POST("/:id/bind-envelope", letterHandler.BindEnvelope)
			letters.DELETE("/:id/bind-envelope", letterHandler.UnbindEnvelope)
			letters.GET("/:id/envelope", letterHandler.GetLetterEnvelope)

			// SOTA 回信系统路由 (扫码回信和线索保持) - 已实现
			letters.GET("/scan-reply/:code", letterHandler.GetReplyInfoByCode) // 扫码获取回信信息
			letters.POST("/replies", letterHandler.CreateReply)                // 创建回信
			letters.GET("/threads", letterHandler.GetUserThreads)              // 获取用户线程列表
			letters.GET("/threads/:id", letterHandler.GetThreadByID)           // 获取线程详情
			
			// 草稿管理
			letters.GET("/drafts", letterHandler.GetDrafts)           // 获取草稿列表
			letters.POST("/:id/publish", letterHandler.PublishLetter) // 发布信件
			
			// 互动功能
			letters.POST("/:id/like", letterHandler.LikeLetter)   // 点赞信件
			letters.POST("/:id/share", letterHandler.ShareLetter) // 分享信件
			
			// 模板功能
			// templates route moved to public section
			letters.GET("/templates/:id", letterHandler.GetLetterTemplateByID) // 获取模板详情
			
			// 搜索和发现
			letters.POST("/search", letterHandler.SearchLetters)             // 搜索信件
			// Popular and recommended letters routes are already in public section
			
			// 批量操作和导出
			letters.POST("/batch", letterHandler.BatchOperateLetters) // 批量操作
			letters.POST("/export", letterHandler.ExportLetters)      // 导出信件
			
			// 写作辅助
			letters.POST("/auto-save", letterHandler.AutoSaveDraft)                   // 自动保存草稿
			letters.POST("/writing-suggestions", letterHandler.GetWritingSuggestions) // 获取写作建议
		}

		// 信使相关
		courier := protected.Group("/courier")
		{
			courier.POST("/apply", courierHandler.ApplyCourier)
			courier.GET("/status", courierHandler.GetCourierStatus)
			courier.GET("/profile", courierHandler.GetCourierProfile)
			courier.POST("/letters/:code/status", letterHandler.UpdateStatus)

			// 四级信使管理API
			courier.POST("/create", courierHandler.CreateCourier)           // 创建下级信使
			courier.GET("/subordinates", courierHandler.GetSubordinates)    // 获取下级信使列表
			courier.GET("/me", courierHandler.GetCourierInfo)               // 获取当前信使信息
			courier.GET("/candidates", courierHandler.GetCourierCandidates) // 获取信使候选人列表
			courier.GET("/tasks", courierHandler.GetCourierTasks)           // 获取信使任务列表
			
			// 晋升系统路由 - SOTA完整实现
			growth := courier.Group("/growth")
			{
				growth.GET("/path", courierGrowthHandler.GetGrowthPath)
				growth.GET("/progress", courierGrowthHandler.GetGrowthProgress)
				growth.POST("/apply", courierGrowthHandler.SubmitUpgradeRequest)      // 提交晋升申请
				growth.GET("/applications", courierGrowthHandler.GetUpgradeRequests)  // 获取申请列表
				growth.PUT("/applications/:request_id", courierGrowthHandler.ProcessUpgradeRequest) // 处理申请
			}
			
			// 等级管理路由 - 本地实现
			level := courier.Group("/level")
			{
				level.GET("/config", courierGrowthHandler.GetLevelConfig)
				level.GET("/check", courierGrowthHandler.CheckLevel)
				level.GET("/check/:courier_id", courierGrowthHandler.CheckLevel)
				level.POST("/upgrade", courierGrowthHandler.SubmitUpgradeRequest)
				level.GET("/upgrade-requests", courierGrowthHandler.GetUpgradeRequests)
				level.PUT("/upgrade/:request_id", courierGrowthHandler.ProcessUpgradeRequest)
			}
			
			// QR扫描相关 - SOTA集成，无缝融入现有架构 - Temporarily disabled
			/*
			courier.POST("/scan", qrScanHandler.ProcessQRScan)              // 处理QR码扫描
			courier.GET("/letters/:code", qrScanHandler.GetLetterByCode)    // 通过编码获取信件信息
			courier.GET("/scan-history", qrScanHandler.GetScanHistory)      // 获取扫描历史记录
			*/

			// 管理级别API
			management := courier.Group("/management")
			{
				// 一级信使管理 (楼栋)
				management.GET("/level-1/stats", courierHandler.GetFirstLevelStats)
				management.GET("/level-1/couriers", courierHandler.GetFirstLevelCouriers)

				// 二级信使管理 (片区)
				management.GET("/level-2/stats", courierHandler.GetSecondLevelStats)
				management.GET("/level-2/couriers", courierHandler.GetSecondLevelCouriers)

				// 三级信使管理 (学校)
				management.GET("/level-3/stats", courierHandler.GetThirdLevelStats)
				management.GET("/level-3/couriers", courierHandler.GetThirdLevelCouriers)

				// 四级信使管理 (城市)
				management.GET("/level-4/stats", courierHandler.GetFourthLevelStats)
				management.GET("/level-4/couriers", courierHandler.GetFourthLevelCouriers)
			}
		}

		// 信封相关
		envelopes := protected.Group("/envelopes")
		{
			envelopes.GET("/my", envelopeHandler.GetMyEnvelopes)
			envelopes.GET("/designs", envelopeHandler.GetEnvelopeDesigns)
			envelopes.POST("/orders", envelopeHandler.CreateEnvelopeOrder)
			envelopes.GET("/orders", envelopeHandler.GetEnvelopeOrders)
			envelopes.POST("/orders/:id/pay", envelopeHandler.ProcessEnvelopePayment)
		}

		// 博物馆相关
		museum := protected.Group("/museum")
		{
			museum.POST("/items", museumHandler.CreateMuseumItem)
			museum.POST("/items/:id/ai-description", museumHandler.GenerateItemDescription) // 新增：AI生成描述
			museum.POST("/submit", museumHandler.SubmitLetterToMuseum) // 新增：提交信件到博物馆
			museum.POST("/entries/:id/interact", museumHandler.InteractWithEntry)  // 记录互动（浏览、点赞等）
			museum.POST("/entries/:id/react", museumHandler.ReactToEntry)         // 添加反应
			museum.DELETE("/entries/:id/withdraw", museumHandler.WithdrawMuseumEntry) // 撤回条目
			museum.GET("/my-submissions", museumHandler.GetMySubmissions)         // 获取我的提交记录
			museum.POST("/search", museumHandler.SearchMuseumEntries)              // 搜索博物馆条目
		}

		// 数据分析相关
		analytics := protected.Group("/analytics")
		{
			analytics.GET("/dashboard", analyticsHandler.GetDashboard)
			analytics.GET("/metrics", analyticsHandler.GetMetrics)
			analytics.POST("/metrics", analyticsHandler.RecordMetric)
			analytics.GET("/metrics/summary", analyticsHandler.GetMetricSummary)
			analytics.GET("/users", analyticsHandler.GetUserAnalytics)
			analytics.POST("/reports", analyticsHandler.GenerateReport)
			analytics.GET("/reports", analyticsHandler.GetReports)
			analytics.POST("/performance", analyticsHandler.RecordPerformance)
		}

		// 任务调度相关
		scheduler := protected.Group("/scheduler")
		{
			scheduler.POST("/tasks", schedulerHandler.CreateTask)
			scheduler.GET("/tasks", schedulerHandler.GetTasks)
			scheduler.GET("/tasks/:id", schedulerHandler.GetTask)
			scheduler.PUT("/tasks/:id/status", schedulerHandler.UpdateTaskStatus)
			scheduler.POST("/tasks/:id/enable", schedulerHandler.EnableTask)
			scheduler.POST("/tasks/:id/disable", schedulerHandler.DisableTask)
			scheduler.POST("/tasks/:id/execute", schedulerHandler.ExecuteTaskNow)
			scheduler.DELETE("/tasks/:id", schedulerHandler.DeleteTask)
			scheduler.GET("/tasks/:id/executions", schedulerHandler.GetTaskExecutions)
			scheduler.GET("/stats", schedulerHandler.GetTaskStats)
			scheduler.POST("/tasks/defaults", schedulerHandler.CreateDefaultTasks)
		}

		// 审核相关（普通用户可以触发审核）
		moderation := protected.Group("/moderation")
		{
			moderation.POST("/check", moderationHandler.ModerateContent)
		}

		// 通知相关
		notifications := protected.Group("/notifications")
		{
			notifications.GET("/", notificationHandler.GetUserNotifications)
			notifications.POST("/send", notificationHandler.SendNotification)
			notifications.POST("/:id/read", notificationHandler.MarkNotificationAsRead)
			notifications.POST("/read-all", notificationHandler.MarkAllNotificationsAsRead)
			notifications.GET("/preferences", notificationHandler.GetUserPreferences)
			notifications.PUT("/preferences", notificationHandler.UpdateUserPreferences)
			notifications.POST("/test-email", notificationHandler.TestEmailNotification)
		}

		// 公开的WebSocket统计信息
		wsPublic := public.Group("/ws")
		{
			wsPublic.GET("/stats", wsHandler.HandleGetStats) // 公开的统计信息
		}

		// WebSocket相关 - 使用专用的WebSocket认证中间件
		ws := v1.Group("/ws")
		ws.Use(middleware.WebSocketAuthMiddleware(cfg, db))
		{
			// WebSocket连接端点 - 支持从查询参数获取token
			ws.GET("/connect", wsHandler.HandleWebSocketConnection)
			// 其他WebSocket管理端点
			ws.GET("/connections", wsHandler.HandleGetConnections)
			ws.GET("/rooms/:room/users", wsHandler.HandleGetRoomUsers)
			ws.POST("/broadcast", wsHandler.HandleBroadcastMessage)
			ws.POST("/direct", wsHandler.HandleSendDirectMessage)
			ws.GET("/history", wsHandler.HandleGetMessageHistory)
		}

		// 积分系统相关
		credits := protected.Group("/credits")
		{
			credits.GET("/me", creditHandler.GetUserCredit)            // 获取当前用户积分信息
			credits.GET("/me/history", creditHandler.GetCreditHistory) // 获取积分历史
			credits.GET("/me/level", creditHandler.GetUserLevel)       // 获取等级信息
			credits.GET("/me/stats", creditHandler.GetCreditStats)     // 获取积分统计
			credits.GET("/leaderboard", creditHandler.GetLeaderboard)  // 获取排行榜
			credits.GET("/rules", creditHandler.GetCreditRules)        // 获取积分规则
		}

		// 文件存储相关
		storage := protected.Group("/storage")
		{
			storage.POST("/upload", storageHandler.UploadFile)                   // 上传文件
			storage.GET("/files", storageHandler.GetFiles)                       // 获取文件列表
			storage.GET("/files/:file_id", storageHandler.GetFile)               // 获取文件信息
			storage.GET("/files/:file_id/download", storageHandler.DownloadFile) // 下载文件
			storage.DELETE("/files/:file_id", storageHandler.DeleteFile)         // 删除文件
			storage.GET("/stats", storageHandler.GetStorageStats)                // 获取存储统计
		}

		// 评论系统
		handlers.RegisterCommentRoutes(protected, commentHandler)

		// 关注系统
		follow := protected.Group("/follow")
		{
			// 关注操作
			follow.POST("/users", followHandler.FollowUser)                    // 关注用户
			follow.DELETE("/users/:user_id", followHandler.UnfollowUser)       // 取消关注用户
			follow.POST("/users/batch", followHandler.FollowMultipleUsers)     // 批量关注用户
			
			// 关注列表
			follow.GET("/followers", followHandler.GetFollowers)               // 获取当前用户粉丝列表
			follow.GET("/following", followHandler.GetFollowing)               // 获取当前用户关注列表
			follow.GET("/users/:user_id/followers", followHandler.GetFollowers) // 获取指定用户粉丝列表
			follow.GET("/users/:user_id/following", followHandler.GetFollowing) // 获取指定用户关注列表
			
			// 关注状态
			follow.GET("/users/:user_id/status", followHandler.GetFollowStatus) // 获取关注状态
			
			// 用户搜索和发现
			follow.GET("/users/search", followHandler.SearchUsers)             // 搜索用户
			follow.GET("/suggestions", followHandler.GetUserSuggestions)       // 获取用户推荐
			follow.POST("/suggestions/refresh", followHandler.RefreshSuggestions) // 刷新推荐
			
			// 粉丝管理
			follow.DELETE("/followers/:user_id", followHandler.RemoveFollower) // 移除粉丝
		}

		// 商店系统（需要认证的部分）
		shopAuth := protected.Group("/shop")
		{
			// 商品评价（需要认证）
			shopAuth.POST("/products/:id/reviews", shopHandler.CreateProductReview) // 创建商品评价

			// 购物车相关
			shopAuth.GET("/cart", shopHandler.GetCart)                       // 获取购物车
			shopAuth.POST("/cart/items", shopHandler.AddToCart)             // 添加商品到购物车
			shopAuth.PUT("/cart/items/:id", shopHandler.UpdateCartItem)     // 更新购物车项目
			shopAuth.DELETE("/cart/items/:id", shopHandler.RemoveFromCart)  // 从购物车移除商品
			shopAuth.DELETE("/cart", shopHandler.ClearCart)                 // 清空购物车

			// 订单相关
			shopAuth.POST("/orders", shopHandler.CreateOrder)              // 创建订单
			shopAuth.GET("/orders", shopHandler.GetOrders)                 // 获取订单列表
			shopAuth.GET("/orders/:id", shopHandler.GetOrder)              // 获取订单详情
			shopAuth.POST("/orders/:id/pay", shopHandler.PayOrder)         // 支付订单

			// 收藏相关
			shopAuth.GET("/favorites", shopHandler.GetFavorites)           // 获取收藏列表
			shopAuth.POST("/favorites", shopHandler.AddToFavorites)        // 添加收藏
			shopAuth.DELETE("/favorites/:id", shopHandler.RemoveFromFavorites) // 取消收藏
		}
		
		// OP Code系统 - OpenPenPal核心地理编码系统 - 重新启用
		opcode := protected.Group("/opcode")
		{
			// 用户功能
			opcode.POST("/apply", opcodeHandler.ApplyOPCode)                     // 申请OP Code
			opcode.GET("/validate", opcodeHandler.ValidateOPCode)                // 验证格式
			opcode.GET("/search", opcodeHandler.SearchOPCodes)                   // 搜索OP Code
			opcode.GET("/search/schools", opcodeHandler.SearchSchools)           // 搜索学校
			opcode.GET("/search/areas", opcodeHandler.SearchAreas)               // 搜索片区
			opcode.GET("/search/buildings", opcodeHandler.SearchBuildings)       // 搜索楼栋
			opcode.GET("/search/points", opcodeHandler.SearchPoints)             // 搜索投递点
			opcode.GET("/stats/:school_code", opcodeHandler.GetOPCodeStats)      // 获取统计
			opcode.GET("/:code", opcodeHandler.GetOPCode)                        // 获取OP Code信息
			
			// 管理功能（需要额外权限验证）
			opcodeAdmin := opcode.Group("/admin")
			{
				opcodeAdmin.POST("/applications/:application_id/review", opcodeHandler.AdminReviewApplication) // 审核申请
			}
		}

		// 条码系统 - PRD规格实现
		barcodes := protected.Group("/barcodes")
		{
			barcodes.POST("", barcodeHandler.CreateBarcode)                      // 创建条码
			barcodes.PATCH("/:id/bind", barcodeHandler.BindBarcode)              // 绑定条码
			barcodes.PATCH("/:id/status", barcodeHandler.UpdateBarcodeStatus)    // 更新状态
			barcodes.GET("/:id/status", barcodeHandler.GetBarcodeStatus)         // 获取状态
			barcodes.POST("/:id/validate", barcodeHandler.ValidateBarcodeOperation) // 验证操作权限
		}

		// 扫描事件系统 - PRD要求的完整扫描历史
		scanEvents := protected.Group("/scan-events")
		{
			scanEvents.GET("", scanEventHandler.GetScanHistory)                  // 获取扫描历史
			scanEvents.GET("/:id", scanEventHandler.GetScanEventByID)            // 获取扫描事件详情
			scanEvents.GET("/barcode/:barcode_id/timeline", scanEventHandler.GetBarcodeTimeline) // 获取条码时间线
			scanEvents.GET("/summary", scanEventHandler.GetScanEventSummary)     // 获取统计摘要
			scanEvents.GET("/user/activity", scanEventHandler.GetUserScanActivity) // 获取用户扫描活动
			scanEvents.GET("/location/:op_code/stats", scanEventHandler.GetLocationScanStats) // 获取位置统计
			
			// 管理员功能
			scanEvents.POST("", scanEventHandler.CreateScanEvent)               // 手动创建扫描事件
			scanEvents.POST("/cleanup", scanEventHandler.CleanupOldScanEvents)  // 清理旧事件
		}
	}

	// 管理员路由
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg, db))
	// 支持多种管理员角色
	admin.Use(func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(401, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}
		user := userInterface.(*models.User)
		// 检查是否是任何管理员角色
		if user.Role == "admin" || user.Role == models.RoleCourierLevel3 || 
		   user.Role == models.RolePlatformAdmin || user.Role == models.RoleSuperAdmin {
			c.Next()
			return
		}
		c.JSON(403, gin.H{"error": "需要管理员权限", "user_role": string(user.Role)})
		c.Abort()
	})
	{
		// 管理后台仪表盘
		admin.GET("/dashboard/stats", adminHandler.GetDashboardStats)
		admin.GET("/dashboard/activities", adminHandler.GetRecentActivities)
		admin.GET("/dashboard/analytics", adminHandler.GetAnalyticsData)
		admin.POST("/seed-data", adminHandler.InjectSeedData)
		// 系统设置管理 - 完整的CRUD操作
		admin.GET("/settings", systemSettingsHandler.GetSettings)
		admin.PUT("/settings", systemSettingsHandler.UpdateSettings)
		admin.POST("/settings", systemSettingsHandler.ResetSettings)
		admin.POST("/settings/test-email", systemSettingsHandler.TestEmailConfig)

		// 用户管理
		adminUsers := admin.Group("/users")
		{
			adminUsers.GET("/", adminHandler.GetUserManagement)
			adminUsers.GET("/:id", userHandler.AdminGetUser)
			adminUsers.PUT("/:id", adminHandler.UpdateUser)
			adminUsers.DELETE("/:id", userHandler.AdminDeactivateUser)
			adminUsers.POST("/:id/reactivate", userHandler.AdminReactivateUser)
		}

		// 信使管理
		adminCourier := admin.Group("/courier")
		{
			adminCourier.GET("/applications", courierHandler.GetPendingApplications)
			adminCourier.POST("/:id/approve", courierHandler.ApproveCourierApplication)
			adminCourier.POST("/:id/reject", courierHandler.RejectCourierApplication)
		}

		// 博物馆管理
		adminMuseum := admin.Group("/museum")
		{
			adminMuseum.POST("/items/:id/approve", museumHandler.ApproveMuseumItem)
			adminMuseum.POST("/entries/:id/moderate", museumHandler.ModerateMuseumEntry)    // 审核条目
			adminMuseum.GET("/entries/pending", museumHandler.GetPendingMuseumEntries)      // 获取待审核条目
			adminMuseum.POST("/exhibitions", museumHandler.CreateMuseumExhibition)          // 创建展览
			adminMuseum.PUT("/exhibitions/:id", museumHandler.UpdateMuseumExhibition)       // 更新展览
			adminMuseum.DELETE("/exhibitions/:id", museumHandler.DeleteMuseumExhibition)    // 删除展览
			adminMuseum.POST("/exhibitions/:id/items", museumHandler.AddItemsToExhibition)  // 向展览添加物品
			adminMuseum.DELETE("/exhibitions/:id/items", museumHandler.RemoveItemsFromExhibition) // 从展览移除物品
			adminMuseum.PUT("/exhibitions/:id/items/order", museumHandler.UpdateExhibitionItemOrder) // 更新物品显示顺序
			adminMuseum.POST("/exhibitions/:id/publish", museumHandler.PublishExhibition)   // 发布展览
			adminMuseum.POST("/refresh-stats", museumHandler.RefreshMuseumStats)            // 刷新统计数据
			adminMuseum.GET("/analytics", museumHandler.GetMuseumAnalytics)                 // 获取分析数据
		}

		// 数据分析管理
		adminAnalytics := admin.Group("/analytics")
		{
			adminAnalytics.GET("/system", analyticsHandler.GetSystemAnalytics)
			adminAnalytics.GET("/dashboard", analyticsHandler.GetDashboard)
			adminAnalytics.GET("/reports", analyticsHandler.GetReports)
		}

		// 审核管理
		adminModeration := admin.Group("/moderation")
		{
			adminModeration.POST("/review", moderationHandler.ReviewContent)
			adminModeration.GET("/queue", moderationHandler.GetModerationQueue)
			adminModeration.GET("/stats", moderationHandler.GetModerationStats)

			// 敏感词管理
			adminModeration.GET("/sensitive-words", moderationHandler.GetSensitiveWords)
			adminModeration.POST("/sensitive-words", moderationHandler.AddSensitiveWord)
			adminModeration.PUT("/sensitive-words/:id", moderationHandler.UpdateSensitiveWord)
			adminModeration.DELETE("/sensitive-words/:id", moderationHandler.DeleteSensitiveWord)

			// 审核规则管理
			adminModeration.GET("/rules", moderationHandler.GetModerationRules)
			adminModeration.POST("/rules", moderationHandler.AddModerationRule)
			adminModeration.PUT("/rules/:id", moderationHandler.UpdateModerationRule)
			adminModeration.DELETE("/rules/:id", moderationHandler.DeleteModerationRule)
		}

		// 积分管理
		adminCredits := admin.Group("/credits")
		{
			adminCredits.GET("/users/:user_id", creditHandler.AdminGetUserCredit)    // 获取指定用户积分信息
			adminCredits.POST("/users/add-points", creditHandler.AdminAddPoints)     // 给用户增加积分
			adminCredits.POST("/users/spend-points", creditHandler.AdminSpendPoints) // 给用户扣除积分
			adminCredits.GET("/leaderboard", creditHandler.GetLeaderboard)           // 积分排行榜（管理员视图）
			adminCredits.GET("/rules", creditHandler.GetCreditRules)                 // 积分规则管理
		}

		// AI管理
		adminAI := admin.Group("/ai")
		{
			adminAI.GET("/config", aiHandler.GetAIConfig)           // 获取AI配置
			adminAI.PUT("/config", aiHandler.UpdateAIConfig)        // 更新AI配置
			adminAI.GET("/templates", aiHandler.GetContentTemplates) // 获取内容模板
			adminAI.POST("/templates", aiHandler.CreateContentTemplate) // 创建内容模板
			adminAI.GET("/monitoring", aiHandler.GetAIMonitoring)   // 获取AI监控数据
			adminAI.GET("/analytics", aiHandler.GetAIAnalytics)     // 获取AI分析数据
			adminAI.GET("/logs", aiHandler.GetAILogs)               // 获取AI操作日志
			adminAI.POST("/test-provider", aiHandler.TestAIProvider) // 测试AI提供商连接
		}

		// 商店管理
		adminShop := admin.Group("/shop")
		{
			adminShop.POST("/products", shopHandler.CreateProduct)              // 创建商品
			adminShop.PUT("/products/:id", shopHandler.UpdateProduct)           // 更新商品
			adminShop.DELETE("/products/:id", shopHandler.DeleteProduct)        // 删除商品
			adminShop.PUT("/orders/:id/status", shopHandler.UpdateOrderStatus)  // 更新订单状态
			adminShop.GET("/stats", shopHandler.GetShopStatistics)             // 获取商店统计
		}

		// ==================== SOTA 管理API适配路由 ====================
		// 兼容Java前端期待的API格式和路径
		
		// 用户管理 - 适配Java前端
		admin.GET("/users", adminAdapter.GetUsersCompat)
		admin.GET("/users/:id", adminAdapter.GetUserCompat) 
		admin.PUT("/users/:id", adminAdapter.UpdateUserCompat)
		admin.POST("/users/:id/unlock", adminAdapter.UnlockUserCompat)
		admin.POST("/users/:id/reset-password", adminAdapter.ResetPasswordCompat)
		admin.GET("/users/stats/role", adminAdapter.GetUserStatsCompat)
		
		// 信件管理 - 适配Java前端
		admin.GET("/letters", adminAdapter.GetLettersCompat)
		admin.GET("/letters/:id", adminAdapter.GetLetterCompat)
		admin.PUT("/letters/:id/status", adminAdapter.UpdateLetterStatusCompat)
		admin.GET("/letters/stats/overview", adminAdapter.GetLetterStatsCompat)
		
		// 系统配置 - 适配Java前端
		admin.GET("/system/config", adminAdapter.GetSystemConfigCompat)
		admin.PUT("/system/config/:key", adminAdapter.UpdateSystemConfigCompat)
		admin.GET("/system/info", adminAdapter.GetSystemInfoCompat)
		admin.GET("/system/health", adminAdapter.GetSystemHealthCompat)
	}

	// 静态文件服务（二维码图片等）
	router.Static("/uploads", "./uploads")

	// 启动服务器
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Frontend URL: %s", cfg.FrontendURL)

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
