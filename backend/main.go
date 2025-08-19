package main

import (
	"fmt"
	"time"

	"openpenpal-backend/internal/adapters"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/handlers"
	"openpenpal-backend/internal/logger"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/routes"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化智能日志系统
	log := logger.GetLogger()
	
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: %v", err)
	}

	// 初始化数据库
	// 使用直接的数据库连接方法，绕过 shared 包的配置问题
	db, err := config.SetupDatabaseDirect(cfg)
	if err != nil {
		log.Fatal("Failed to setup database: %v", err)
	}

	// 在开发环境下初始化测试数据
	if cfg.Environment == "development" {
		// 重新启用数据种子功能
		if err := config.SeedData(db); err != nil {
			log.Warn("Failed to seed data: %v", err)
		} else {
			log.Info("Test data seeded successfully")
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
	aiManager := services.NewAIProviderManager(db)
	notificationService := services.NewNotificationService(db, cfg)
	analyticsService := services.NewAnalyticsService(db)
	schedulerService := services.NewSchedulerService(db)
	creditService := services.NewCreditService(db)
	
	// 初始化Redis连接（用于限制系统）
	redisClient, err := config.SetupRedis(cfg)
	if err != nil {
		log.Warn("Failed to setup Redis: %v", err)
		redisClient = nil // 继续运行，但限制功能不可用
	}
	
	// 初始化积分限制服务
	var creditLimiterService *services.CreditLimiterService
	if redisClient != nil {
		creditLimiterService = services.NewCreditLimiterService(db, redisClient)
	}
	
	creditTaskService := services.NewCreditTaskService(db, creditService, creditLimiterService) // 新增：模块化积分任务服务
	courierTaskService := services.NewCourierTaskService(db)
	adminService := services.NewAdminService(db, cfg)
	systemSettingsService := services.NewSystemSettingsService(db, cfg)
	storageService := services.NewStorageService(db, cfg)
	moderationService := services.NewModerationService(db, cfg, aiService)
	shopService := services.NewShopService(db)
	creditShopService := services.NewCreditShopService(db, creditService, creditLimiterService) // Phase 2: 积分商城服务
	creditActivityService := services.NewCreditActivityService(db, creditService, creditLimiterService) // Phase 3: 积分活动服务
	creditActivityScheduler := services.NewCreditActivityScheduler(db, creditActivityService) // Phase 3.3: 活动调度器
	commentService := services.NewCommentService(db, cfg)
	followService := services.NewFollowService(db) // 关注系统服务
	privacyService := services.NewPrivacyService(db) // 隐私设置服务
	opcodeService := services.NewOPCodeService(db)       // OP Code服务 - 重新启用
	scanEventService := services.NewScanEventService(db) // 扫描事件服务 - PRD要求
	cloudLetterService := services.NewCloudLetterService(db, cfg) // 云中锦书服务 - 自定义现实角色
	contentSecurityService := services.NewContentSecurityService(db, cfg, aiService) // 内容安全服务 - XSS防护和敏感词管理
	tagService := services.NewTagService(db) // 标签服务 - 内容发现与分类
	auditService := services.NewAuditService(db) // 审计服务 - 企业级安全审计

	// Phase 4.1: 初始化积分过期服务
	creditExpirationService := services.NewCreditExpirationService(db, creditService, notificationService)
	creditService.SetCreditExpirationService(creditExpirationService) // 设置双向依赖

	// Phase 4.2: 初始化积分转赠服务
	creditTransferService := services.NewCreditTransferService(db, creditService, notificationService, creditLimiterService)

	// 延迟队列服务 - 使用修复版本防止无限循环
	log.Info("Initializing fixed delay queue service with circuit breaker...")
	
	delayQueueService, err := services.NewDelayQueueService(db, cfg)
	if err != nil {
		log.Warn("Failed to initialize fixed delay queue service: %v", err)
	} else {
		// 在后台启动延迟队列工作进程
		go delayQueueService.StartWorker()
		log.Info("Fixed delay queue worker started successfully")
	}

	// 初始化WebSocket服务
	wsService := websocket.NewWebSocketService()
	wsService.Start()

	// 创建WebSocket适配器用于服务间通信
	wsAdapter := websocket.NewWebSocketAdapter(wsService)

	// Configure service dependencies with credit system integration
	letterService.SetCreditService(creditService)
	letterService.SetCreditTaskService(creditTaskService) // 新增：积分任务服务依赖
	letterService.SetCourierTaskService(courierTaskService)
	letterService.SetNotificationService(notificationService)
	letterService.SetWebSocketService(wsService)
	letterService.SetAIService(aiService)
	letterService.SetOPCodeService(opcodeService) // PRD要求：集成OP Code验证
	aiService.SetCreditTaskService(creditTaskService) // 新增：AI服务积分任务依赖
	letterService.SetUserService(userService)     // 添加用户服务依赖
	envelopeService.SetCreditService(creditService)
	envelopeService.SetUserService(userService) // FSD增强：OP Code区域验证
	museumService.SetCreditService(creditService)
	museumService.SetCreditTaskService(creditTaskService) // 新增：博物馆服务积分任务依赖
	museumService.SetNotificationService(notificationService)
	museumService.SetAIService(aiService)
	courierTaskService.SetNotificationService(notificationService)
	courierService.SetWebSocketService(wsAdapter) // SOTA: Dependency Injection for real-time notifications
	notificationService.SetWebSocketService(wsService)
	commentService.SetLetterService(letterService)
	commentService.SetCreditService(creditService)
	commentService.SetModerationService(moderationService)
	commentService.SetContentSecurityService(contentSecurityService)
	// 配置云中锦书服务依赖
	cloudLetterService.SetLetterService(letterService)
	cloudLetterService.SetAIService(aiService)
	cloudLetterService.SetCourierService(courierService)
	cloudLetterService.SetNotificationService(notificationService)
	// 配置标签服务依赖
	tagService.SetAIService(aiService)

	// 启动任务调度服务
	if err := schedulerService.Start(); err != nil {
		log.Warn("Failed to start scheduler service: %v", err)
	} else {
		log.Info("Scheduler service started successfully")
	}
	
	// 启动积分活动调度器
	if err := creditActivityScheduler.Start(); err != nil {
		log.Warn("Failed to start credit activity scheduler: %v", err)
	} else {
		log.Info("Credit activity scheduler started successfully")
	}

	// 注册默认调度任务
	log.Info("Registering default scheduler tasks...")
	futureLetterService := services.NewFutureLetterService(db, letterService, notificationService)
	schedulerTasks := services.NewSchedulerTasks(
		futureLetterService,
		letterService,
		aiService,
		notificationService,
		envelopeService,
		courierService,
	)
	
	if err := schedulerTasks.RegisterDefaultTasks(schedulerService); err != nil {
		log.Warn("Failed to register default scheduler tasks: %v", err)
	} else {
		log.Info("Default scheduler tasks registered successfully")
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
	aiHandler := handlers.NewAIHandler(aiService, configService, aiManager)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	schedulerHandler := handlers.NewSchedulerHandler(schedulerService)
	creditHandler := handlers.NewCreditHandler(creditService)
	creditTaskHandler := handlers.NewCreditTaskHandler(creditTaskService, creditService) // 新增：积分任务处理器
	
	// 初始化积分限制处理器（如果限制服务可用）
	var creditLimitHandler *handlers.CreditLimitHandler
	var creditLimitAdminHandler *handlers.CreditLimitAdminHandler
	if creditLimiterService != nil {
		creditLimitHandler = handlers.NewCreditLimitHandler(creditLimiterService)
		creditLimitAdminHandler = handlers.NewCreditLimitAdminHandler(creditLimiterService) // Phase 1.4: 管理界面处理器
	}
	adminHandler := handlers.NewAdminHandler(adminService)
	systemSettingsHandler := handlers.NewSystemSettingsHandler(systemSettingsService, cfg)
	storageHandler := handlers.NewStorageHandler(storageService)
	moderationHandler := handlers.NewModerationHandler(moderationService)
	opcodeHandler := handlers.NewOPCodeHandler(opcodeService, courierService)
	courierOPCodeHandler := handlers.NewCourierOPCodeHandler(opcodeService, courierService)      // 信使OP Code管理处理器
	barcodeHandler := handlers.NewBarcodeHandler(letterService, opcodeService, scanEventService) // PRD条码系统处理器
	scanEventHandler := handlers.NewScanEventHandler(scanEventService)                           // 扫描事件处理器
	batchHandler := handlers.NewBatchHandler(letterService, courierService, opcodeService)       // 批量管理处理器
	cloudLetterHandler := handlers.NewCloudLetterHandler(cloudLetterService)                     // 云中锦书处理器
	shopHandler := handlers.NewShopHandler(shopService, userService)
	creditShopHandler := handlers.NewCreditShopHandler(creditShopService, creditService) // Phase 2: 积分商城处理器
	creditActivityHandler := handlers.NewCreditActivityHandler(creditActivityService, creditService) // Phase 3: 积分活动处理器
	creditActivitySchedulerHandler := handlers.NewCreditActivitySchedulerHandler(creditActivityScheduler) // Phase 3.3: 活动调度器处理器
	creditExpirationHandler := handlers.NewCreditExpirationHandler(creditExpirationService, creditService) // Phase 4.1: 积分过期处理器
	creditTransferHandler := handlers.NewCreditTransferHandler(creditTransferService, creditService) // Phase 4.2: 积分转赠处理器
	commentHandler := handlers.NewCommentHandler(commentService)
	followHandler := handlers.NewFollowHandler(followService) // 关注系统处理器
	privacyHandler := handlers.NewPrivacyHandler(privacyService) // 隐私设置处理器
	userProfileHandler := handlers.NewUserProfileHandler(db) // 用户档案处理器
	sensitiveWordHandler := handlers.NewSensitiveWordHandler(contentSecurityService) // 敏感词管理处理器
	// 初始化完整性服务
	integrityService := services.NewIntegrityService(db, cfg)
	
	validationHandler := handlers.NewValidationHandler(integrityService, auditService) // 安全验证处理器
	tagHandler := handlers.NewTagHandler(tagService) // 标签处理器 - 内容发现与分类
	auditHandler := handlers.NewAuditHandler(auditService) // 审计处理器 - 企业级安全审计

	// QR/条码扫描服务 - SOTA集成：复用现有依赖
	log.Info("Initializing QR/barcode scanning functionality...")
	qrScanService := services.NewQRScanService(db, letterService, courierService, wsAdapter)
	qrScanService.SetCreditTaskService(creditTaskService) // 积分任务服务依赖
	log.Info("QR scan service initialized successfully - using existing barcode handler for scanning")
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
	router.Use(middleware.RequestIDMiddleware()) // 请求追踪
	router.Use(middleware.LoggerMiddleware())    // 日志记录
	router.Use(middleware.RecoveryMiddleware())  // 错误恢复
	metricsMiddleware := middleware.NewMetricsMiddleware("openpenpal-backend")
	router.Use(metricsMiddleware.RequestMetrics())   // 性能监控 - 重新启用

	// 2. 安全中间件
	router.Use(middleware.SecurityHeadersMiddleware())                                  // 安全头
	router.Use(middleware.CORSMiddleware())                                             // CORS
	router.Use(middleware.InputValidation())                                           // 输入验证
	router.Use(middleware.ContentLengthValidation(10 * 1024 * 1024))                  // 内容长度限制 (10MB)
	router.Use(middleware.RequestSizeLimitMiddleware(middleware.DefaultMaxRequestSize)) // 请求大小限制
	router.Use(middleware.SecurityMonitoringMiddleware())                              // 安全监控
	router.Use(middleware.ThreatDetectionMiddleware())                                 // 威胁检测

	// 3. 频率限制（IP级别作为默认）
	router.Use(middleware.RateLimitMiddleware())

	// 4. API转换中间件 - SOTA实现
	router.Use(middleware.RequestTransformMiddleware())  // 请求转换 (camelCase -> snake_case)
	router.Use(middleware.ResponseTransformMiddleware()) // 响应转换 (snake_case -> camelCase)

	// CSP 违规报告端点
	router.POST("/csp-report", middleware.CSPViolationHandler())

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
	
	// 设置AI路由 - 多提供商AI系统
	routes.SetupAIRoutes(router, aiHandler, cfg, db)
	routes.SetupAIWebSocketRoutes(router, aiHandler, cfg, db)

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
			auth.GET("/csrf", authHandler.GetCSRFToken) // 获取CSRF令牌

			// 状态改变操作需要CSRF保护
			csrfProtected := auth.Group("/")
			csrfProtected.Use(middleware.CSRFMiddleware())
			{
				csrfProtected.POST("/register", authHandler.Register) // 用户注册
				csrfProtected.POST("/login", authHandler.Login)       // 用户登录
			}

			// 用户信息端点（需要认证）
			authGroup := auth.Group("/")
			authGroup.Use(middleware.AuthMiddleware(cfg, db))
			{
				authGroup.GET("/me", authHandler.GetCurrentUser)             // 获取当前用户信息
				authGroup.POST("/logout", authHandler.Logout)                // 登出
				authGroup.POST("/refresh", authHandler.RefreshToken)         // 刷新令牌
				authGroup.GET("/check-expiry", authHandler.CheckTokenExpiry) // 检查令牌过期
			}
		}

		// 公开的信件读取
		letters := public.Group("/letters")
		{
			letters.GET("/read/:code", letterHandler.GetLetterByCode)
			letters.POST("/read/:code/mark-read", letterHandler.MarkAsRead)
			letters.GET("/public", letterHandler.GetPublicLetters)           // 新增：广场信件
			letters.GET("/popular", letterHandler.GetPopularLetters)         // 新增：热门信件
			letters.GET("/recommended", letterHandler.GetRecommendedLetters) // 新增：推荐信件
			letters.GET("/templates", letterHandler.GetLetterTemplates)      // 新增：信件模板（公开）
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
			museum.GET("/popular", museumHandler.GetPopularMuseumEntries)          // 获取热门条目
			museum.GET("/exhibitions/:id", museumHandler.GetMuseumExhibitionByID)  // 获取展览详情
			museum.GET("/exhibitions/:id/items", museumHandler.GetExhibitionItems) // 获取展览中的物品
			museum.GET("/tags", museumHandler.GetMuseumTags)                       // 获取标签列表
			museum.GET("/stats", museumHandler.GetMuseumStats)                     // 获取博物馆统计
		}

		// 公开的AI相关（无需认证）- 这些已移至ai_routes.go，需要认证
		// ai := public.Group("/ai")
		// {
		//	ai.POST("/match", aiHandler.MatchPenPal) // 已移至ai_routes.go，需要认证
		//	ai.POST("/reply", aiHandler.GenerateReply)
		//	ai.POST("/reply-advice", aiHandler.GenerateReplyAdvice) // 新增：生成回信角度建议
		//	ai.POST("/inspiration", aiHandler.GetInspiration)
		//	ai.POST("/curate", aiHandler.CurateLetters)
		//	ai.GET("/personas", aiHandler.GetPersonas)
		//	ai.GET("/stats", aiHandler.GetAIStats)
		//	ai.GET("/daily-inspiration", aiHandler.GetDailyInspiration)
		// }

		// 公开的商店信息（无需认证）
		shop := public.Group("/shop")
		{
			shop.GET("/products", shopHandler.GetProducts)                   // 获取商品列表（公开）
			shop.GET("/products/:id", shopHandler.GetProduct)                // 获取商品详情（公开）
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
			letters.POST("/search", letterHandler.SearchLetters) // 搜索信件
			// Popular and recommended letters routes are already in public section

			// 批量操作和导出
			letters.POST("/batch", letterHandler.BatchOperateLetters) // 批量操作
			letters.POST("/export", letterHandler.ExportLetters)      // 导出信件

			// 写作辅助
			letters.POST("/auto-save", letterHandler.AutoSaveDraft)                   // 自动保存草稿
			letters.POST("/writing-suggestions", letterHandler.GetWritingSuggestions) // 获取写作建议
		}

		// 云中锦书相关（自定义现实角色写信）
		cloudLetters := protected.Group("/cloud-letters")
		{
			// 人物角色管理
			cloudLetters.POST("/personas", cloudLetterHandler.CreatePersona)          // 创建自定义人物角色
			cloudLetters.GET("/personas", cloudLetterHandler.GetPersonas)             // 获取用户的人物角色列表
			cloudLetters.PUT("/personas/:persona_id", cloudLetterHandler.UpdatePersona) // 更新人物角色
			cloudLetters.GET("/persona-types", cloudLetterHandler.GetPersonaTypes)    // 获取支持的人物关系类型
			
			// 云信件管理
			cloudLetters.POST("/", cloudLetterHandler.CreateCloudLetter)              // 创建云信件
			cloudLetters.GET("/", cloudLetterHandler.GetCloudLetters)                 // 获取用户的云信件列表
			cloudLetters.GET("/:letter_id", cloudLetterHandler.GetCloudLetter)        // 获取云信件详情
			cloudLetters.GET("/status-options", cloudLetterHandler.GetLetterStatusOptions) // 获取信件状态选项
			
			// L3/L4信使审核功能
			cloudLetters.GET("/pending-reviews", cloudLetterHandler.GetPendingReviews)  // 获取待审核的云信件（L3/L4专用）
			cloudLetters.POST("/:letter_id/review", cloudLetterHandler.ReviewCloudLetter) // 审核云信件（L3/L4专用）
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
				growth.POST("/apply", courierGrowthHandler.SubmitUpgradeRequest)                    // 提交晋升申请
				growth.GET("/applications", courierGrowthHandler.GetUpgradeRequests)                // 获取申请列表
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

			// 批量管理API (L3/L4信使专用)
			batch := courier.Group("/batch")
			{
				batch.POST("/generate", batchHandler.GenerateBatch)           // 批量生成条码
				batch.GET("", batchHandler.GetBatches)                       // 获取批次列表
				batch.GET("/:id", batchHandler.GetBatchDetails)              // 获取批次详情
				batch.GET("/stats", batchHandler.GetBatchStats)              // 获取统计信息
				batch.PATCH("/:id/status", batchHandler.UpdateBatchStatus)   // 更新批次状态
			}
			
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
			
			// OP Code 管理API (信使专用)
			opcodeManage := courier.Group("/opcode")
			{
				opcodeManage.GET("/applications", courierOPCodeHandler.GetApplications)                      // 获取申请列表
				opcodeManage.POST("/applications/:application_id/review", courierOPCodeHandler.ReviewApplication) // 审核申请
				opcodeManage.POST("/create", courierOPCodeHandler.CreateOPCode)                             // 创建OP Code
				opcodeManage.GET("/managed", courierOPCodeHandler.GetManagedOPCodes)                        // 获取管理的OP Code列表
				opcodeManage.PUT("/:id", courierOPCodeHandler.UpdateOPCode)                                // 更新OP Code
				opcodeManage.DELETE("/:id", courierOPCodeHandler.DeleteOPCode)                             // 删除OP Code
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
			museum.POST("/submit", museumHandler.SubmitLetterToMuseum)                      // 新增：提交信件到博物馆
			museum.POST("/entries/:id/interact", museumHandler.InteractWithEntry)           // 记录互动（浏览、点赞等）
			museum.POST("/entries/:id/react", museumHandler.ReactToEntry)                   // 添加反应
			museum.DELETE("/entries/:id/withdraw", museumHandler.WithdrawMuseumEntry)       // 撤回条目
			museum.GET("/my-submissions", museumHandler.GetMySubmissions)                   // 获取我的提交记录
			museum.POST("/search", museumHandler.SearchMuseumEntries)                       // 搜索博物馆条目
			
			// 批量操作
			museum.POST("/items/batch", museumHandler.BatchOperateMuseumItems)              // 批量操作博物馆条目
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
			
			// 积分任务系统 - 模块化积分奖励
			credits.GET("/me/tasks", creditTaskHandler.GetUserTasks)                        // 获取用户积分任务
			credits.GET("/me/summary", creditTaskHandler.GetUserCreditSummary)              // 获取用户积分摘要
			credits.GET("/task-rules", creditTaskHandler.GetCreditTaskRules)                // 获取积分任务规则
			
			// 快速触发积分奖励 - FSD规格实现
			credits.POST("/trigger/letter/:letter_id", creditTaskHandler.TriggerLetterCreatedReward)    // 触发写信奖励
			credits.POST("/trigger/like", creditTaskHandler.TriggerPublicLetterLikeReward)              // 触发点赞奖励
			credits.POST("/trigger/ai-interaction", creditTaskHandler.TriggerAIInteractionReward)       // 触发AI互动奖励
			credits.POST("/trigger/courier/:task_id", creditTaskHandler.TriggerCourierDeliveryReward)   // 触发信使送达奖励
			
			// 积分限制相关端点（如果服务可用）
			if creditLimitHandler != nil {
				credits.GET("/limits/:action_type", creditLimitHandler.GetUserLimitStatus) // 获取用户限制状态
				credits.GET("/risk-info", creditLimitHandler.GetUserRiskInfo)               // 获取用户风险信息
			}

			// Phase 4.1: 积分过期相关端点
			credits.GET("/expiring", creditExpirationHandler.GetUserExpiringCredits)           // 获取即将过期积分
			credits.GET("/expiration-history", creditExpirationHandler.GetUserExpirationHistory) // 获取过期历史

			// Phase 4.2: 积分转赠相关端点
			credits.POST("/transfer", creditTransferHandler.CreateTransfer)                      // 创建积分转赠
			credits.GET("/transfers", creditTransferHandler.GetUserTransfers)                   // 获取用户转赠记录
			credits.GET("/transfers/stats", creditTransferHandler.GetTransferStats)             // 获取转赠统计
			credits.GET("/transfers/pending", creditTransferHandler.GetPendingTransfers)        // 获取待处理转赠
			credits.GET("/transfers/:id", creditTransferHandler.GetTransfer)                    // 获取转赠详情
			credits.POST("/transfers/:id/process", creditTransferHandler.ProcessTransfer)       // 处理积分转赠
			credits.DELETE("/transfers/:id", creditTransferHandler.CancelTransfer)             // 取消积分转赠
			credits.POST("/transfer/batch", creditTransferHandler.BatchTransfer)               // 批量转赠
			credits.POST("/transfer/validate", creditTransferHandler.ValidateTransfer)         // 验证转赠可行性
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

		// 标签系统 - 内容发现与分类
		tags := protected.Group("/tags")
		{
			// 标签基础操作
			tags.POST("", tagHandler.CreateTag)                     // 创建标签
			tags.GET("/:id", tagHandler.GetTag)                     // 获取标签详情
			tags.PUT("/:id", tagHandler.UpdateTag)                  // 更新标签
			tags.DELETE("/:id", tagHandler.DeleteTag)               // 删除标签
			
			// 标签搜索和发现
			tags.GET("/search", tagHandler.SearchTags)              // 搜索标签
			tags.GET("/popular", tagHandler.GetPopularTags)         // 获取热门标签
			tags.GET("/trending", tagHandler.GetTrendingTags)       // 获取趋势标签
			tags.POST("/suggest", tagHandler.SuggestTags)           // 标签建议
			
			// 内容标记
			tags.POST("/content", tagHandler.TagContent)            // 标记内容
			tags.DELETE("/content", tagHandler.UntagContent)        // 取消标记
			tags.GET("/content/:content_type/:content_id", tagHandler.GetContentTags) // 获取内容标签
			
			// 标签关注
			tags.POST("/:id/follow", tagHandler.FollowTag)          // 关注标签
			tags.DELETE("/:id/follow", tagHandler.UnfollowTag)      // 取消关注标签
			tags.GET("/followed", tagHandler.GetFollowedTags)       // 获取关注的标签
			
			// 标签分类
			tags.GET("/categories", tagHandler.GetTagCategories)    // 获取标签分类
			tags.POST("/categories", tagHandler.CreateTagCategory)  // 创建标签分类
			tags.GET("/categories/:id", tagHandler.GetTagCategory)  // 获取分类详情
			
			// 标签统计和分析
			tags.GET("/stats", tagHandler.GetTagStats)              // 获取标签统计
			tags.GET("/:id/trend", tagHandler.GetTagTrend)          // 获取标签趋势
			
			// 批量操作
			tags.POST("/batch", tagHandler.BatchOperateTags)        // 批量操作标签
		}

		// 关注系统
		follow := protected.Group("/follow")
		{
			// 关注操作
			follow.POST("/users", followHandler.FollowUser)                // 关注用户
			follow.DELETE("/users/:user_id", followHandler.UnfollowUser)   // 取消关注用户
			follow.POST("/users/batch", followHandler.FollowMultipleUsers) // 批量关注用户

			// 关注列表
			follow.GET("/followers", followHandler.GetFollowers)                // 获取当前用户粉丝列表
			follow.GET("/following", followHandler.GetFollowing)                // 获取当前用户关注列表
			follow.GET("/users/:user_id/followers", followHandler.GetFollowers) // 获取指定用户粉丝列表
			follow.GET("/users/:user_id/following", followHandler.GetFollowing) // 获取指定用户关注列表

			// 关注状态
			follow.GET("/users/:user_id/status", followHandler.GetFollowStatus) // 获取关注状态
			follow.GET("/users/:user_id/stats", followHandler.GetUserStats)     // 获取用户统计

			// 用户搜索和发现
			follow.GET("/users/search", followHandler.SearchUsers)                // 搜索用户
			follow.GET("/suggestions", followHandler.GetUserSuggestions)          // 获取用户推荐
			follow.POST("/suggestions/refresh", followHandler.RefreshSuggestions) // 刷新推荐

			// 粉丝管理
			follow.DELETE("/followers/:user_id", followHandler.RemoveFollower) // 移除粉丝
		}

		// 隐私设置系统
		privacy := protected.Group("/privacy")
		{
			// 隐私设置管理
			privacy.GET("/settings", privacyHandler.GetPrivacySettings)        // 获取隐私设置
			privacy.PUT("/settings", privacyHandler.UpdatePrivacySettings)     // 更新隐私设置
			privacy.POST("/settings/reset", privacyHandler.ResetPrivacySettings) // 重置隐私设置

			// 隐私权限检查
			privacy.GET("/check/:user_id", privacyHandler.CheckPrivacy)        // 检查隐私权限
			privacy.POST("/check/:user_id/batch", privacyHandler.BatchCheckPrivacy) // 批量检查隐私权限

			// 用户屏蔽管理
			privacy.POST("/block", privacyHandler.BlockUser)                   // 屏蔽用户
			privacy.DELETE("/block/:user_id", privacyHandler.UnblockUser)      // 取消屏蔽用户
			privacy.GET("/blocked", privacyHandler.GetBlockedUsers)            // 获取屏蔽用户列表

			// 用户静音管理
			privacy.POST("/mute", privacyHandler.MuteUser)                     // 静音用户
			privacy.DELETE("/mute/:user_id", privacyHandler.UnmuteUser)        // 取消静音用户
			privacy.GET("/muted", privacyHandler.GetMutedUsers)                // 获取静音用户列表

			// 关键词过滤管理
			privacy.POST("/keywords/block", privacyHandler.AddBlockedKeyword)  // 添加屏蔽关键词
			privacy.DELETE("/keywords/block/:keyword", privacyHandler.RemoveBlockedKeyword) // 移除屏蔽关键词
			privacy.GET("/keywords/blocked", privacyHandler.GetBlockedKeywords) // 获取屏蔽关键词列表
		}

		// 商店系统（需要认证的部分）
		shopAuth := protected.Group("/shop")
		{
			// 商品评价（需要认证）
			shopAuth.POST("/products/:id/reviews", shopHandler.CreateProductReview) // 创建商品评价

			// 购物车相关
			shopAuth.GET("/cart", shopHandler.GetCart)                     // 获取购物车
			shopAuth.POST("/cart/items", shopHandler.AddToCart)            // 添加商品到购物车
			shopAuth.PUT("/cart/items/:id", shopHandler.UpdateCartItem)    // 更新购物车项目
			shopAuth.DELETE("/cart/items/:id", shopHandler.RemoveFromCart) // 从购物车移除商品
			shopAuth.DELETE("/cart", shopHandler.ClearCart)                // 清空购物车

			// 订单相关
			shopAuth.POST("/orders", shopHandler.CreateOrder)      // 创建订单
			shopAuth.GET("/orders", shopHandler.GetOrders)         // 获取订单列表
			shopAuth.GET("/orders/:id", shopHandler.GetOrder)      // 获取订单详情
			shopAuth.POST("/orders/:id/pay", shopHandler.PayOrder) // 支付订单

			// 收藏相关
			shopAuth.GET("/favorites", shopHandler.GetFavorites)               // 获取收藏列表
			shopAuth.POST("/favorites", shopHandler.AddToFavorites)            // 添加收藏
			shopAuth.DELETE("/favorites/:id", shopHandler.RemoveFromFavorites) // 取消收藏
		}

		// Phase 2: 积分商城系统
		creditShop := v1.Group("/credit-shop")
		{
			// 公开商品信息
			creditShop.GET("/products", creditShopHandler.GetCreditShopProducts)     // 获取积分商品列表
			creditShop.GET("/products/:id", creditShopHandler.GetCreditShopProduct)  // 获取积分商品详情
			creditShop.GET("/categories", creditShopHandler.GetCreditShopCategories) // 获取商品分类
			creditShop.GET("/categories/:id", creditShopHandler.GetCreditShopCategory) // 获取分类详情
		}

		// 积分商城认证功能
		creditShopAuth := protected.Group("/credit-shop")
		{
			// 用户功能
			creditShopAuth.GET("/balance", creditShopHandler.GetUserCreditBalance)      // 获取积分余额
			creditShopAuth.POST("/validate", creditShopHandler.ValidatePurchase)        // 验证购买能力
			
			// 购物车管理
			creditShopAuth.GET("/cart", creditShopHandler.GetCreditCart)                // 获取积分购物车
			creditShopAuth.POST("/cart/items", creditShopHandler.AddToCreditCart)       // 添加商品到积分购物车
			creditShopAuth.PUT("/cart/items/:id", creditShopHandler.UpdateCreditCartItem) // 更新积分购物车项目
			creditShopAuth.DELETE("/cart/items/:id", creditShopHandler.RemoveFromCreditCart) // 从积分购物车移除商品
			creditShopAuth.DELETE("/cart", creditShopHandler.ClearCreditCart)           // 清空积分购物车
			
			// Phase 2.3: 兑换订单管理
			creditShopAuth.POST("/redemptions", creditShopHandler.CreateCreditRedemption)        // 创建兑换订单
			creditShopAuth.POST("/redemptions/from-cart", creditShopHandler.CreateCreditRedemptionFromCart) // 从购物车创建兑换订单
			creditShopAuth.GET("/redemptions", creditShopHandler.GetCreditRedemptions)           // 获取用户兑换订单列表
			creditShopAuth.GET("/redemptions/:id", creditShopHandler.GetCreditRedemption)        // 获取兑换订单详情
			creditShopAuth.DELETE("/redemptions/:id", creditShopHandler.CancelCreditRedemption)  // 取消兑换订单
		}

		// Phase 3: 积分活动系统
		creditActivity := v1.Group("/credit-activities")
		{
			// 公开活动信息
			creditActivity.GET("/active", creditActivityHandler.GetActiveActivities)           // 获取进行中的活动
			creditActivity.GET("", creditActivityHandler.GetActivities)                       // 获取活动列表
			creditActivity.GET("/:id", creditActivityHandler.GetActivity)                     // 获取活动详情
			creditActivity.GET("/templates", creditActivityHandler.GetActivityTemplates)      // 获取活动模板
		}

		// 积分活动认证功能
		creditActivityAuth := protected.Group("/credit-activities")
		{
			// 用户参与功能
			creditActivityAuth.POST("/:id/participate", creditActivityHandler.ParticipateActivity)           // 参与活动
			creditActivityAuth.GET("/my-participations", creditActivityHandler.GetUserParticipations)        // 获取用户参与记录
			creditActivityAuth.POST("/trigger", creditActivityHandler.TriggerActivity)                       // 触发活动事件
		}

		// OP Code系统 - OpenPenPal核心地理编码系统 - 重新启用
		opcode := protected.Group("/opcode")
		{
			// 用户功能
			opcode.POST("/apply", opcodeHandler.ApplyOPCodeHierarchical)    // 层级化申请OP Code - 新增
			opcode.GET("/validate", opcodeHandler.ValidateOPCode)           // 验证格式
			opcode.GET("/search", opcodeHandler.SearchOPCodes)              // 搜索OP Code
			opcode.GET("/search/schools", opcodeHandler.SearchSchools)      // 搜索学校
			opcode.GET("/search/schools/by-city", opcodeHandler.SearchSchoolsByCity)      // 按城市搜索学校 - 新增
			opcode.GET("/search/schools/advanced", opcodeHandler.SearchSchoolsAdvanced)   // 高级学校搜索 - 新增
			opcode.GET("/search/areas", opcodeHandler.SearchAreas)          // 搜索片区
			opcode.GET("/search/buildings", opcodeHandler.SearchBuildings)  // 搜索楼栋
			opcode.GET("/search/points", opcodeHandler.SearchPoints)        // 搜索投递点
			opcode.GET("/stats/:school_code", opcodeHandler.GetOPCodeStats) // 获取统计
			
			// 层级选择系统新增路由
			opcode.GET("/cities", opcodeHandler.GetCities)                                      // 获取城市列表
			opcode.GET("/districts/:school_code", opcodeHandler.GetDistricts)                   // 获取学校片区列表
			opcode.GET("/buildings/:school_code/:district_code", opcodeHandler.GetBuildings)    // 获取楼栋列表
			opcode.GET("/delivery-points/:prefix", opcodeHandler.GetDeliveryPoints)             // 获取投递点列表（含可用性）
			
			opcode.GET("/:code", opcodeHandler.GetOPCode)                   // 获取OP Code信息

			// 管理功能（需要额外权限验证）
			opcodeAdmin := opcode.Group("/admin")
			{
				opcodeAdmin.POST("/applications/:application_id/review", opcodeHandler.AdminReviewApplication) // 审核申请
			}
		}

		// 条码系统 - PRD规格实现
		barcodes := protected.Group("/barcodes")
		{
			barcodes.POST("", barcodeHandler.CreateBarcode)                         // 创建条码
			barcodes.PATCH("/:id/bind", barcodeHandler.BindBarcode)                 // 绑定条码
			barcodes.PATCH("/:id/status", barcodeHandler.UpdateBarcodeStatus)       // 更新状态
			barcodes.GET("/:id/status", barcodeHandler.GetBarcodeStatus)            // 获取状态
			barcodes.POST("/:id/validate", barcodeHandler.ValidateBarcodeOperation) // 验证操作权限
			barcodes.POST("/verify", barcodeHandler.VerifyBarcodeAuthenticity)      // 验证条码真实性
		}

		// 扫描事件系统 - PRD要求的完整扫描历史
		scanEvents := protected.Group("/scan-events")
		{
			scanEvents.GET("", scanEventHandler.GetScanHistory)                                  // 获取扫描历史
			scanEvents.GET("/:id", scanEventHandler.GetScanEventByID)                            // 获取扫描事件详情
			scanEvents.GET("/barcode/:barcode_id/timeline", scanEventHandler.GetBarcodeTimeline) // 获取条码时间线
			scanEvents.GET("/summary", scanEventHandler.GetScanEventSummary)                     // 获取统计摘要
			scanEvents.GET("/user/activity", scanEventHandler.GetUserScanActivity)               // 获取用户扫描活动
			scanEvents.GET("/location/:op_code/stats", scanEventHandler.GetLocationScanStats)    // 获取位置统计

			// 管理员功能
			scanEvents.POST("", scanEventHandler.CreateScanEvent)              // 手动创建扫描事件
			scanEvents.POST("/cleanup", scanEventHandler.CleanupOldScanEvents) // 清理旧事件
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
		admin.GET("/dashboard", adminHandler.GetDashboardStats)                      // 前端期望的主要仪表盘接口
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
			adminUsers.GET("/", adminHandler.GetUsers)                                    // 获取用户列表 (前端期望的格式)
			adminUsers.GET("/management", adminHandler.GetUserManagement)                // 保留原有的管理接口
			adminUsers.GET("/:id", userHandler.AdminGetUser)
			adminUsers.PUT("/:id", adminHandler.UpdateUser)
			adminUsers.PUT("/:id/status", adminHandler.UpdateUserStatus)                 // 新增：更新用户状态
			adminUsers.POST("/:id/reset-password", adminHandler.ResetUserPassword)       // 新增：重置用户密码
			adminUsers.DELETE("/:id", userHandler.AdminDeactivateUser)
			adminUsers.POST("/:id/reactivate", userHandler.AdminReactivateUser)
		}

		// 角色和任命管理
		admin.GET("/roles", adminHandler.GetAppointableRoles)                     // 获取可任命角色列表
		adminAppointments := admin.Group("/appointments")
		{
			adminAppointments.GET("/", adminHandler.GetAppointmentRecords)       // 获取任命记录
			adminAppointments.POST("/", adminHandler.AppointUser)                // 任命用户角色
			adminAppointments.PUT("/:id", adminHandler.ReviewAppointment)        // 审批任命申请
		}

		// 信件管理
		adminLetters := admin.Group("/letters")
		{
			adminLetters.GET("/", adminHandler.GetLetters)                            // 获取信件列表
			adminLetters.POST("/:id/moderate", adminHandler.ModerateLetter)          // 审核信件
		}

		// 信使管理
		adminCouriers := admin.Group("/couriers")
		{
			adminCouriers.GET("/", adminHandler.GetCouriers)                         // 获取信使列表
		}

		// 信使申请管理（保留原有功能）
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
			adminMuseum.POST("/entries/:id/moderate", museumHandler.ModerateMuseumEntry)             // 审核条目
			adminMuseum.GET("/entries/pending", museumHandler.GetPendingMuseumEntries)               // 获取待审核条目
			adminMuseum.POST("/exhibitions", museumHandler.CreateMuseumExhibition)                   // 创建展览
			adminMuseum.PUT("/exhibitions/:id", museumHandler.UpdateMuseumExhibition)                // 更新展览
			adminMuseum.DELETE("/exhibitions/:id", museumHandler.DeleteMuseumExhibition)             // 删除展览
			adminMuseum.POST("/exhibitions/:id/items", museumHandler.AddItemsToExhibition)           // 向展览添加物品
			adminMuseum.DELETE("/exhibitions/:id/items", museumHandler.RemoveItemsFromExhibition)    // 从展览移除物品
			adminMuseum.PUT("/exhibitions/:id/items/order", museumHandler.UpdateExhibitionItemOrder) // 更新物品显示顺序
			adminMuseum.POST("/exhibitions/:id/publish", museumHandler.PublishExhibition)            // 发布展览
			adminMuseum.POST("/refresh-stats", museumHandler.RefreshMuseumStats)                     // 刷新统计数据
			adminMuseum.GET("/analytics", museumHandler.GetMuseumAnalytics)                          // 获取分析数据
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

			// 敏感词管理（通过审核模块访问）
			adminModeration.GET("/sensitive-words", func(c *gin.Context) {
				// 转发到专门的敏感词管理处理器（需要四级信使或平台管理员权限）
				c.Request.URL.Path = "/api/v1/admin/sensitive-words"
				router.HandleContext(c)
			})

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
			
			// 积分任务管理 - 模块化积分系统
			adminCredits.POST("/tasks", creditTaskHandler.CreateTask)                      // 创建积分任务
			adminCredits.POST("/tasks/:task_id/execute", creditTaskHandler.ExecuteTask)    // 手动执行任务
			adminCredits.GET("/tasks/statistics", creditTaskHandler.GetTaskStatistics)     // 获取任务统计
			adminCredits.POST("/tasks/batch", creditTaskHandler.CreateBatchTasks)          // 批量创建任务
			
			// 积分限制和风控管理（如果服务可用）
			if creditLimitHandler != nil {
				adminCredits.GET("/limit-rules", creditLimitHandler.GetLimitRules)               // 获取限制规则
				adminCredits.POST("/limit-rules", creditLimitHandler.CreateLimitRule)            // 创建限制规则
				adminCredits.PUT("/limit-rules/:id", creditLimitHandler.UpdateLimitRule)         // 更新限制规则
				adminCredits.DELETE("/limit-rules/:id", creditLimitHandler.DeleteLimitRule)      // 删除限制规则
				adminCredits.GET("/risk-users", creditLimitHandler.GetRiskUsers)                 // 获取风险用户
				adminCredits.POST("/users/block", creditLimitHandler.BlockUser)                  // 封禁用户
				adminCredits.DELETE("/users/:user_id/block", creditLimitHandler.UnblockUser)     // 解封用户
				adminCredits.GET("/users/:user_id/actions", creditLimitHandler.GetUserActions)   // 获取用户行为记录
			}
			
			// Phase 1.4: 增强管理界面（如果服务可用）
			if creditLimitAdminHandler != nil {
				// 批量操作
				adminCredits.POST("/limit-rules/batch", creditLimitAdminHandler.BatchCreateRules)     // 批量创建规则
				adminCredits.PUT("/limit-rules/batch", creditLimitAdminHandler.BatchUpdateRules)      // 批量更新规则
				
				// 导入导出
				adminCredits.GET("/limit-rules/export", creditLimitAdminHandler.ExportRules)          // 导出规则配置
				adminCredits.POST("/limit-rules/import", creditLimitAdminHandler.ImportRules)         // 导入规则配置
				
				// 统计报表
				adminCredits.GET("/dashboard/stats", creditLimitAdminHandler.GetDashboardStats)       // 仪表板统计
				adminCredits.GET("/reports/usage", creditLimitAdminHandler.GetLimitUsageReport)       // 限制使用报告
				adminCredits.GET("/reports/fraud", creditLimitAdminHandler.GetFraudDetectionReport)   // 防作弊检测报告
				
				// 实时监控
				adminCredits.GET("/monitoring/alerts", creditLimitAdminHandler.GetRealTimeAlerts)     // 实时告警
				adminCredits.GET("/monitoring/health", creditLimitAdminHandler.GetSystemHealth)       // 系统健康状态
				
				// 高级搜索
				adminCredits.POST("/search/advanced", creditLimitAdminHandler.AdvancedSearch)          // 高级搜索
			}

			// Phase 4.1: 积分过期管理
			adminCredits.GET("/expiration/rules", creditExpirationHandler.GetExpirationRules)                      // 获取过期规则
			adminCredits.POST("/expiration/rules", creditExpirationHandler.CreateExpirationRule)                   // 创建过期规则
			adminCredits.PUT("/expiration/rules/:id", creditExpirationHandler.UpdateExpirationRule)                // 更新过期规则
			adminCredits.DELETE("/expiration/rules/:id", creditExpirationHandler.DeleteExpirationRule)             // 删除过期规则
			adminCredits.POST("/expiration/process", creditExpirationHandler.ProcessExpiredCredits)                 // 手动处理过期积分
			adminCredits.POST("/expiration/warnings", creditExpirationHandler.SendExpirationWarnings)              // 发送过期警告
			adminCredits.GET("/expiration/statistics", creditExpirationHandler.GetExpirationStatistics)            // 获取过期统计
			adminCredits.GET("/expiration/batches", creditExpirationHandler.GetExpirationBatches)                  // 获取过期批次
			adminCredits.GET("/expiration/logs", creditExpirationHandler.GetExpirationLogs)                        // 获取过期日志
			adminCredits.GET("/expiration/notifications", creditExpirationHandler.GetExpirationNotifications)      // 获取过期通知记录

			// Phase 4.2: 积分转赠管理
			adminCredits.GET("/transfers/all", creditTransferHandler.GetAllTransfers)                           // 获取所有转赠记录
			adminCredits.GET("/transfers/statistics", creditTransferHandler.GetTransferStatistics)             // 获取转赠统计
			adminCredits.POST("/transfers/process-expired", creditTransferHandler.ProcessExpiredTransfers)      // 处理过期转赠
			adminCredits.DELETE("/transfers/:id/cancel", creditTransferHandler.AdminCancelTransfer)            // 管理员取消转赠
		}

		// AI管理 - 已移至ai_routes.go，避免重复注册
		// adminAI := admin.Group("/ai")
		// {
		//	adminAI.GET("/config", aiHandler.GetAIConfig)               // 获取AI配置
		//	adminAI.PUT("/config", aiHandler.UpdateAIConfig)            // 更新AI配置
		//	adminAI.GET("/templates", aiHandler.GetContentTemplates)    // 获取内容模板
		//	adminAI.POST("/templates", aiHandler.CreateContentTemplate) // 创建内容模板
		//	adminAI.GET("/monitoring", aiHandler.GetAIMonitoring)       // 获取AI监控数据
		//	adminAI.GET("/analytics", aiHandler.GetAIAnalytics)         // 获取AI分析数据
		//	adminAI.GET("/logs", aiHandler.GetAILogs)                   // 获取AI操作日志
		//	adminAI.POST("/test-provider", aiHandler.TestAIProvider)    // 测试AI提供商连接
		// }

		// 商店管理
		adminShop := admin.Group("/shop")
		{
			adminShop.POST("/products", shopHandler.CreateProduct)             // 创建商品
			adminShop.PUT("/products/:id", shopHandler.UpdateProduct)          // 更新商品
			adminShop.DELETE("/products/:id", shopHandler.DeleteProduct)       // 删除商品
			adminShop.PUT("/orders/:id/status", shopHandler.UpdateOrderStatus) // 更新订单状态
			adminShop.GET("/stats", shopHandler.GetShopStatistics)             // 获取商店统计
		}

		// Phase 2: 积分商城管理
		adminCreditShop := admin.Group("/credit-shop")
		{
			// 商品管理
			adminCreditShop.POST("/products", creditShopHandler.CreateCreditShopProduct)   // 创建积分商品
			adminCreditShop.PUT("/products/:id", creditShopHandler.UpdateCreditShopProduct) // 更新积分商品
			adminCreditShop.DELETE("/products/:id", creditShopHandler.DeleteCreditShopProduct) // 删除积分商品
			
			// 分类管理
			adminCreditShop.POST("/categories", creditShopHandler.CreateCreditShopCategory)   // 创建商品分类
			adminCreditShop.PUT("/categories/:id", creditShopHandler.UpdateCreditShopCategory) // 更新商品分类
			adminCreditShop.DELETE("/categories/:id", creditShopHandler.DeleteCreditShopCategory) // 删除商品分类
			
			// 配置管理
			adminCreditShop.GET("/config", creditShopHandler.GetCreditShopConfig)    // 获取系统配置
			adminCreditShop.POST("/config", creditShopHandler.UpdateCreditShopConfig) // 更新系统配置
			
			// 统计数据
			adminCreditShop.GET("/stats", creditShopHandler.GetCreditShopStatistics) // 获取积分商城统计
		}

		// Phase 3: 积分活动管理
		adminCreditActivity := admin.Group("/credit-activities")
		{
			// 活动管理
			adminCreditActivity.POST("", creditActivityHandler.CreateActivity)                           // 创建活动
			adminCreditActivity.PUT("/:id", creditActivityHandler.UpdateActivity)                       // 更新活动
			adminCreditActivity.DELETE("/:id", creditActivityHandler.DeleteActivity)                    // 删除活动
			
			// 活动状态管理
			adminCreditActivity.POST("/:id/start", creditActivityHandler.StartActivity)                 // 启动活动
			adminCreditActivity.POST("/:id/pause", creditActivityHandler.PauseActivity)                 // 暂停活动
			adminCreditActivity.POST("/:id/resume", creditActivityHandler.ResumeActivity)               // 恢复活动
			adminCreditActivity.POST("/:id/complete", creditActivityHandler.CompleteActivity)           // 结束活动
			
			// 活动统计
			adminCreditActivity.GET("/:id/statistics", creditActivityHandler.GetActivityStatistics)     // 获取活动统计
			adminCreditActivity.GET("/statistics", creditActivityHandler.GetAllActivitiesStatistics)    // 获取所有活动统计
			
			// 活动模板管理
			adminCreditActivity.POST("/templates", creditActivityHandler.CreateActivityTemplate)        // 创建活动模板
			adminCreditActivity.POST("/templates/:id/create", creditActivityHandler.CreateActivityFromTemplate) // 从模板创建活动
			
			// 定时处理
			adminCreditActivity.POST("/process-scheduled", creditActivityHandler.ProcessScheduledActivities) // 处理定时活动
		}

		// Phase 3.3: 积分活动调度管理
		adminCreditActivityScheduler := admin.Group("/credit-activities/scheduler")
		{
			// 调度器控制
			adminCreditActivityScheduler.POST("/start", creditActivitySchedulerHandler.StartScheduler)   // 启动调度器
			adminCreditActivityScheduler.POST("/stop", creditActivitySchedulerHandler.StopScheduler)     // 停止调度器
			adminCreditActivityScheduler.GET("/status", creditActivitySchedulerHandler.GetSchedulerStatus) // 获取调度器状态
			
			// 任务调度管理
			adminCreditActivityScheduler.POST("/schedule", creditActivitySchedulerHandler.ScheduleActivity)        // 安排活动执行
			adminCreditActivityScheduler.GET("/tasks", creditActivitySchedulerHandler.GetScheduledTasks)          // 获取调度任务列表
			adminCreditActivityScheduler.DELETE("/tasks/:id", creditActivitySchedulerHandler.CancelScheduledTask) // 取消调度任务
			
			// 批量操作
			adminCreditActivityScheduler.POST("/schedule/recurring", creditActivitySchedulerHandler.ScheduleRecurringActivities)     // 安排重复活动
			adminCreditActivityScheduler.POST("/schedule/immediate", creditActivitySchedulerHandler.ProcessImmediateExecution)       // 立即执行活动
			
			// 调度统计
			adminCreditActivityScheduler.GET("/statistics", creditActivitySchedulerHandler.GetSchedulingStatistics) // 获取调度统计
		}

		// 安全监控管理（需要平台管理员或超级管理员权限）
		adminSecurity := admin.Group("/security")
		adminSecurity.Use(func(c *gin.Context) {
			userInterface, exists := c.Get("user")
			if !exists {
				c.JSON(401, gin.H{"error": "用户未认证"})
				c.Abort()
				return
			}
			user := userInterface.(*models.User)
			if user.Role == models.RolePlatformAdmin || user.Role == models.RoleSuperAdmin {
				c.Next()
				return
			}
			c.JSON(403, gin.H{"error": "需要平台管理员或超级管理员权限"})
			c.Abort()
		})
		{
			securityHandler := handlers.NewSecurityHandler()
			adminSecurity.GET("/events", securityHandler.GetSecurityEvents)       // 获取安全事件
			adminSecurity.GET("/stats", securityHandler.GetSecurityStats)         // 获取安全统计
			adminSecurity.GET("/dashboard", securityHandler.GetSecurityDashboard) // 安全仪表板
			adminSecurity.POST("/events", securityHandler.RecordCustomSecurityEvent) // 记录自定义安全事件
			
			// 安全验证相关端点
			adminSecurity.POST("/validate", validationHandler.RunSecurityValidation) // 运行完整安全验证
			adminSecurity.GET("/validate/summary", validationHandler.GetValidationSummary) // 获取验证摘要
			adminSecurity.GET("/validate/results", validationHandler.GetValidationResults) // 获取详细结果
			adminSecurity.GET("/validate/export", validationHandler.ExportValidationReport) // 导出验证报告
			adminSecurity.POST("/validate/:component", validationHandler.ValidateSpecificComponent) // 验证特定组件
			adminSecurity.POST("/validate/continuous", validationHandler.RunContinuousValidation) // 持续验证
			adminSecurity.GET("/validate/health", validationHandler.GetValidationHealth) // 验证系统健康
		}

		// 审计日志管理（需要管理员权限）
		adminAudit := admin.Group("/audit")
		{
			adminAudit.GET("/logs", auditHandler.GetAuditLogs)      // 获取审计日志
			adminAudit.GET("/stats", auditHandler.GetAuditStats)    // 获取审计统计
			adminAudit.GET("/export", auditHandler.ExportAuditLogs) // 导出审计日志
		}

		// 敏感词管理（独立路由组，需要四级信使或平台管理员权限）
		adminSensitiveWords := admin.Group("/sensitive-words")
		adminSensitiveWords.Use(func(c *gin.Context) {
			userInterface, exists := c.Get("user")
			if !exists {
				c.JSON(401, gin.H{"error": "用户未认证"})
				c.Abort()
				return
			}
			user := userInterface.(*models.User)
			// 只允许四级信使和平台管理员访问
			if user.Role == models.RoleCourierLevel4 || 
			   user.Role == models.RolePlatformAdmin || 
			   user.Role == models.RoleSuperAdmin {
				c.Next()
				return
			}
			c.JSON(403, gin.H{"error": "需要四级信使或平台管理员权限", "user_role": string(user.Role)})
			c.Abort()
		})
		{
			adminSensitiveWords.GET("", sensitiveWordHandler.ListSensitiveWords)         // 获取敏感词列表
			adminSensitiveWords.POST("", sensitiveWordHandler.CreateSensitiveWord)       // 创建敏感词
			adminSensitiveWords.PUT("/:id", sensitiveWordHandler.UpdateSensitiveWord)    // 更新敏感词
			adminSensitiveWords.DELETE("/:id", sensitiveWordHandler.DeleteSensitiveWord) // 删除敏感词
			adminSensitiveWords.POST("/batch-import", sensitiveWordHandler.BatchImportSensitiveWords) // 批量导入
			adminSensitiveWords.GET("/export", sensitiveWordHandler.ExportSensitiveWords)            // 导出敏感词
			adminSensitiveWords.POST("/refresh", sensitiveWordHandler.RefreshSensitiveWords)         // 刷新敏感词库
			adminSensitiveWords.GET("/stats", sensitiveWordHandler.GetSensitiveWordStats)            // 获取统计信息
		}

		// ==================== SOTA 管理API适配路由 ====================
		// 兼容Java前端期待的API格式和路径

		// 用户管理 - 适配Java前端
		admin.GET("/users", adminAdapter.GetUsersCompat)
		// admin.GET("/users/:id", adminAdapter.GetUserCompat) // 与adminUsers组路由冲突，已注释
		// admin.PUT("/users/:id", adminAdapter.UpdateUserCompat) // 与adminUsers组路由冲突，已注释
		admin.POST("/users/:id/unlock", adminAdapter.UnlockUserCompat)
		// Removed duplicate route - already defined in adminUsers group
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
	log.Info("Starting server on %s", addr)
	log.Info("Environment: %s", cfg.Environment)
	log.Info("Frontend URL: %s", cfg.FrontendURL)

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server: %v", err)
	}
}
