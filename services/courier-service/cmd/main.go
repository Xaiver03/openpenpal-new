package main

import (
	"courier-service/internal/config"
	"courier-service/internal/handlers"
	"courier-service/internal/logging"
	"courier-service/internal/middleware"
	"courier-service/internal/monitoring"
	"courier-service/internal/resilience"
	"courier-service/internal/services"
	"courier-service/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化结构化日志系统
	var logLevel logging.LogLevel
	switch cfg.Environment {
	case "development":
		logLevel = logging.LevelDebug
	case "testing":
		logLevel = logging.LevelInfo
	case "production":
		logLevel = logging.LevelWarn
	default:
		logLevel = logging.LevelInfo
	}
	
	logging.InitDefaultLogger("courier-service", logLevel)
	logger := logging.GetDefaultLogger()
	
	logger.Info("Starting courier service", 
		"version", "1.0.0",
		"environment", cfg.Environment,
		"port", cfg.Port,
	)

	// 初始化监控系统
	monitoring.InitGlobalMetrics(logger)
	monitoring.InitGlobalAlertManager(monitoring.GetGlobalRegistry(), logger)

	// 初始化数据库
	db, err := config.InitDatabase(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	logger.Info("Database connected successfully")

	// 初始化Redis
	redisClient := utils.InitRedis(cfg.RedisURL)
	logger.Info("Redis connected successfully")

	// 初始化WebSocket管理器
	wsManager := utils.NewWebSocketManager()

	// 初始化服务层
	courierService := services.NewCourierService(db, redisClient, wsManager)
	taskService := services.NewTaskService(db, redisClient, wsManager)
	locationService := services.NewLocationService()
	assignmentService := services.NewAssignmentService(db, locationService, wsManager)
	queueService := services.NewQueueService(redisClient, db, wsManager, assignmentService)
	levelService := services.NewCourierLevelService(db, redisClient, wsManager)
	growthService := services.NewCourierGrowthService(db, redisClient, wsManager)
	postalService := services.NewPostalManagementService(db, redisClient, wsManager)
	hierarchyService := services.NewHierarchyService(db, wsManager)
	leaderboardService := services.NewLeaderboardService(db, wsManager)
	hierarchicalAssignmentService := services.NewHierarchicalAssignmentService(db, assignmentService, hierarchyService, wsManager)
	signalCodeService := services.NewSignalCodeService(db)

	// 启动队列消费者
	go queueService.ConsumeTaskQueues()
	go queueService.ConsumeAssignmentQueue()
	go queueService.ConsumeNotificationQueue()
	go queueService.ProcessRetryQueue()

	// 初始化路由
	router := gin.New() // 使用gin.New()而不是gin.Default()来完全控制中间件

	// 应用核心中间件（按顺序）
	router.Use(middleware.RequestID()) // 请求ID必须在最前面
	
	// 根据环境选择恢复中间件
	if cfg.Environment == "development" {
		router.Use(middleware.DebugRecovery())
		router.Use(middleware.DebugErrorHandler())
	} else {
		router.Use(middleware.DefaultRecovery())
		router.Use(middleware.DefaultErrorHandler())
	}
	
	// 监控中间件
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		
		// 记录响应时间
		monitoring.RecordResponseTime(duration)
		
		// 如果有错误，记录错误指标
		if len(c.Errors) > 0 {
			monitoring.RecordError(c.Errors.Last().Err, map[string]string{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
			})
		}
	})
	
	// 其他中间件
	router.Use(middleware.CORS())
	router.Use(func(c *gin.Context) {
		// 自定义日志中间件，使用结构化日志
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		
		c.Next()
		
		duration := time.Since(start)
		requestID := middleware.GetRequestID(c)
		
		logger.Info("HTTP Request",
			"method", method,
			"path", path,
			"query", raw,
			"status", c.Writer.Status(),
			"duration", duration,
			"request_id", requestID,
			"user_agent", c.Request.UserAgent(),
			"remote_addr", c.ClientIP(),
		)
	})

	// API路由组
	api := router.Group("/api/courier")
	api.Use(middleware.JWTAuth(cfg.JWTSecret))

	// 注册路由
	handlers.RegisterCourierRoutes(api, courierService)
	handlers.RegisterTaskRoutes(api, taskService, queueService)
	handlers.RegisterScanRoutes(api, taskService, locationService)
	handlers.RegisterCourierLevelRoutes(api, courierService, levelService)
	handlers.RegisterCourierGrowthRoutes(api, growthService)
	handlers.RegisterPostalManagementRoutes(api, postalService)
	handlers.RegisterHierarchyRoutes(api, hierarchyService)
	handlers.RegisterLeaderboardRoutes(api, leaderboardService)
	handlers.RegisterHierarchicalAssignmentRoutes(api, hierarchicalAssignmentService)
	
	// 注册信号编码路由 (不需要JWT认证，因为有些接口是公开的)
	signalCodeHandler := handlers.NewSignalCodeHandler(signalCodeService)
	handlers.RegisterSignalCodeRoutes(router, signalCodeHandler)

	// 健康检查和监控端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "courier-service",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
		})
	})

	// 指标端点
	router.GET("/metrics", func(c *gin.Context) {
		registry := monitoring.GetGlobalRegistry()
		metrics := registry.GetAllMetrics()
		
		result := make(map[string]interface{})
		for name, metric := range metrics {
			result[name] = gin.H{
				"type":      metric.Type(),
				"value":     metric.Value(),
				"labels":    metric.Labels(),
				"timestamp": metric.Timestamp(),
			}
		}
		
		c.JSON(200, result)
	})

	// 告警状态端点
	router.GET("/alerts", func(c *gin.Context) {
		alertManager := monitoring.GetGlobalAlertManager()
		alerts := alertManager.GetActiveAlerts()
		c.JSON(200, alerts)
	})

	// 熔断器状态端点
	router.GET("/circuit-breakers", func(c *gin.Context) {
		manager := resilience.GetGlobalManager()
		stats := manager.GetAllStats()
		c.JSON(200, stats)
	})

	logger.Info("Courier service starting", "port", cfg.Port)
	
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatal("Failed to start server", "error", err)
	}
}