package router

import (
	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/middleware"
	"api-gateway/internal/monitor"
	"api-gateway/internal/proxy"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Manager 路由管理器
type Manager struct {
	config             *config.Config
	proxyManager       *proxy.Manager
	logger             *zap.Logger
	router             *gin.Engine
	performanceHandler *handlers.MetricsHandler
}

// NewManager 创建路由管理器
func NewManager(cfg *config.Config, proxyManager *proxy.Manager, logger *zap.Logger) *Manager {
	return &Manager{
		config:       cfg,
		proxyManager: proxyManager,
		logger:       logger,
	}
}

// SetMetricsHandler 设置性能监控处理器
func (rm *Manager) SetMetricsHandler(handler *handlers.MetricsHandler) {
	rm.performanceHandler = handler
}

// SetupRoutes 设置所有路由
func (rm *Manager) SetupRoutes() *gin.Engine {
	// 初始化Gin
	if rm.config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	rm.router = router

	// 应用全局中间件
	rm.setupGlobalMiddleware()

	// 设置基础路由
	rm.setupBaseRoutes()

	// 设置API路由
	rm.setupAPIRoutes()

	// 设置管理路由
	rm.setupAdminRoutes()

	return router
}

// setupGlobalMiddleware 设置全局中间件
func (rm *Manager) setupGlobalMiddleware() {
	// 基础中间件
	rm.router.Use(middleware.RequestID())
	rm.router.Use(middleware.Logger(rm.logger))
	rm.router.Use(middleware.Recovery(rm.logger))
	rm.router.Use(middleware.CORS())
	rm.router.Use(middleware.Security())
	rm.router.Use(middleware.Metrics())

	// 超时中间件
	timeout := time.Duration(rm.config.ProxyTimeout) * time.Second
	rm.router.Use(middleware.Timeout(timeout))
}

// setupBaseRoutes 设置基础路由
func (rm *Manager) setupBaseRoutes() {
	// 健康检查
	rm.router.GET("/health", rm.healthCheckHandler())
	rm.router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "pong"})
	})

	// 监控指标
	rm.router.GET("/metrics", gin.WrapH(monitor.MetricsHandler()))

	// 版本信息
	rm.router.GET("/version", rm.versionHandler())

	// 服务信息
	rm.router.GET("/info", rm.infoHandler())
}

// setupAPIRoutes 设置API路由
func (rm *Manager) setupAPIRoutes() {
	// API v1 路由组
	apiV1 := rm.router.Group("/api/v1")

	// 认证相关路由（无需JWT认证）
	rm.setupAuthRoutes(apiV1)

	// 需要认证的路由
	authenticatedAPI := apiV1.Group("")
	authenticatedAPI.Use(middleware.JWTAuth(rm.config.JWTSecret))

	// 用户相关路由
	rm.setupUserRoutes(authenticatedAPI)

	// 信件相关路由 - 移到apiV1组以支持公开路由
	rm.setupLetterRoutes(apiV1)

	// 信使相关路由
	rm.setupCourierRoutes(authenticatedAPI)

	// OCR相关路由
	rm.setupOCRRoutes(authenticatedAPI)

	// Postcode地址编码路由
	rm.setupPostcodeRoutes(authenticatedAPI)

	// 地址搜索路由（兼容性）
	rm.setupAddressRoutes(authenticatedAPI)

	// 商店相关路由
	rm.setupShopRoutes(authenticatedAPI)

	// AI相关路由
	rm.setupAIRoutes(apiV1)

	// 博物馆相关路由
	rm.setupMuseumRoutes(apiV1)

	// 性能监控路由（新增）
	rm.setupMetricsRoutes(apiV1)

	// 健康检查路由（新增）
	rm.setupHealthRoutes(apiV1)

	// WebSocket路由（新增）
	rm.setupWebSocketRoutes(apiV1)
}

// setupAuthRoutes 设置认证路由
func (rm *Manager) setupAuthRoutes(group *gin.RouterGroup) {
	authGroup := group.Group("/auth")
	authGroup.Use(middleware.NewRateLimiter(60)) // 每分钟60次

	// 转发到主后端服务
	authGroup.Any("/*path", rm.proxyManager.ProxyHandler("main-backend"))
}

// setupUserRoutes 设置用户路由
func (rm *Manager) setupUserRoutes(group *gin.RouterGroup) {
	userGroup := group.Group("/users")
	userGroup.Use(middleware.NewRateLimiter(120))

	// 转发到主后端服务
	userGroup.Any("/*path", rm.proxyManager.ProxyHandler("main-backend"))
}

// setupLetterRoutes 设置信件路由
func (rm *Manager) setupLetterRoutes(group *gin.RouterGroup) {
	letterGroup := group.Group("/letters")
	letterGroup.Use(middleware.NewRateLimiter(100))

	// 转发所有信件请求到主后端 - 由主后端处理认证
	letterGroup.Any("/*path", rm.proxyManager.ProxyHandler("main-backend"))
}

// setupCourierRoutes 设置信使路由
func (rm *Manager) setupCourierRoutes(group *gin.RouterGroup) {
	courierGroup := group.Group("/courier")
	courierGroup.Use(middleware.NewRateLimiter(80))

	// 路由到主后端的信使操作
	courierGroup.POST("/apply", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.GET("/status", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.GET("/profile", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.GET("/me", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.POST("/create", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.GET("/subordinates", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.GET("/candidates", rm.proxyManager.ProxyHandler("main-backend"))

	// 公共统计信息
	courierGroup.GET("/stats", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.GET("/stats/*path", rm.proxyManager.ProxyHandler("main-backend"))

	// 任务相关路由 (主后端处理)
	courierGroup.GET("/tasks", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.PUT("/tasks/*path", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.POST("/tasks/*path", rm.proxyManager.ProxyHandler("main-backend"))

	// 扫码相关路由
	courierGroup.POST("/scan/*path", rm.proxyManager.ProxyHandler("main-backend"))
	courierGroup.GET("/scan/*path", rm.proxyManager.ProxyHandler("main-backend"))

	// 成长系统路由
	courierGroup.Any("/growth/*path", rm.proxyManager.ProxyHandler("main-backend"))

	// 管理API
	courierGroup.Any("/management/*path", rm.proxyManager.ProxyHandler("main-backend"))

	// 需要信使权限的路由 (转发到courier-service)
	courierProtected := courierGroup.Group("")
	courierProtected.Use(middleware.CourierAuth())
	{
		// TODO: 添加需要信使认证的特定路由
		// 例如: 信使个人信息更新、任务管理等
	}

	// 管理员路由
	courierAdmin := courierGroup.Group("/admin")
	courierAdmin.Use(middleware.AdminAuth())
	{
		courierAdmin.Any("/*path", rm.proxyManager.ProxyHandler("courier-service"))
	}
}

// setupOCRRoutes 设置OCR路由
func (rm *Manager) setupOCRRoutes(group *gin.RouterGroup) {
	ocrGroup := group.Group("/ocr")
	ocrGroup.Use(middleware.NewRateLimiter(20)) // OCR操作限制更严格

	// 转发到OCR服务
	ocrGroup.Any("/*path", rm.proxyManager.ProxyHandler("ocr-service"))
}

// setupAdminRoutes 设置管理路由
func (rm *Manager) setupAdminRoutes() {
	adminGroup := rm.router.Group("/admin")
	adminGroup.Use(middleware.JWTAuth(rm.config.JWTSecret))
	adminGroup.Use(middleware.AdminAuth())
	adminGroup.Use(middleware.NewRateLimiter(30))

	// 网关管理路由
	adminGroup.GET("/gateway/status", rm.gatewayStatusHandler())
	adminGroup.GET("/gateway/services", rm.servicesStatusHandler())
	adminGroup.GET("/gateway/metrics", rm.metricsHandler())
	adminGroup.POST("/gateway/reload", rm.reloadConfigHandler())

	// 服务健康检查
	adminGroup.GET("/health/:service", rm.proxyManager.HealthCheckHandler())
	adminGroup.GET("/health", rm.proxyManager.HealthCheckHandler())

	// 转发其他管理请求到管理服务
	adminGroup.Any("/service/*path", rm.proxyManager.ProxyHandler("admin-service"))
}

// healthCheckHandler 健康检查处理器
func (rm *Manager) healthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "api-gateway",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// versionHandler 版本信息处理器
func (rm *Manager) versionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":    "OpenPenPal API Gateway",
			"version":    "1.0.0",
			"go_version": "1.21",
			"build_time": time.Now().Format(time.RFC3339),
			"git_commit": "unknown",
		})
	}
}

// infoHandler 服务信息处理器
func (rm *Manager) infoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		info := gin.H{
			"gateway": gin.H{
				"name":        "OpenPenPal API Gateway",
				"version":     "1.0.0",
				"environment": rm.config.Environment,
				"uptime":      time.Since(time.Now()).String(),
			},
			"services": make(map[string]interface{}),
		}

		// 添加已配置的服务信息
		for serviceName, serviceConfig := range rm.config.Services {
			info["services"].(map[string]interface{})[serviceName] = gin.H{
				"hosts":        serviceConfig.Hosts,
				"health_check": serviceConfig.HealthCheck,
				"timeout":      serviceConfig.Timeout,
				"weight":       serviceConfig.Weight,
			}
		}

		c.JSON(http.StatusOK, info)
	}
}

// gatewayStatusHandler 网关状态处理器
func (rm *Manager) gatewayStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现网关状态收集
		c.JSON(http.StatusOK, gin.H{
			"status": "operational",
			"stats": gin.H{
				"total_requests": 0,
				"error_rate":     0.0,
				"avg_latency":    "0ms",
			},
		})
	}
}

// servicesStatusHandler 服务状态处理器
func (rm *Manager) servicesStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 从服务发现获取服务状态
		c.JSON(http.StatusOK, gin.H{
			"services": gin.H{},
		})
	}
}

// setupWebSocketRoutes 设置WebSocket路由
func (rm *Manager) setupWebSocketRoutes(group *gin.RouterGroup) {
	wsGroup := group.Group("/ws")
	// WebSocket连接不需要速率限制

	// 转发到主后端服务的WebSocket端点
	wsGroup.Any("/*path", rm.proxyManager.ProxyHandler("main-backend"))
}

// metricsHandler 指标处理器
func (rm *Manager) metricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		summary := monitor.GetMetricsSummary()
		c.JSON(http.StatusOK, summary)
	}
}

// reloadConfigHandler 重新加载配置处理器
func (rm *Manager) reloadConfigHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现配置重新加载
		rm.logger.Info("Config reload requested",
			zap.String("user", middleware.GetUserID(c)),
		)

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Configuration reloaded",
		})
	}
}

// setupMetricsRoutes 设置性能监控路由
func (rm *Manager) setupMetricsRoutes(group *gin.RouterGroup) {
	if rm.performanceHandler == nil {
		rm.logger.Warn("MetricsHandler not initialized, using placeholder endpoints")
		rm.setupPlaceholderMetricsRoutes(group)
		return
	}

	metricsGroup := group.Group("/metrics")
	metricsGroup.Use(middleware.NewRateLimiter(100)) // 每分钟100次

	// 性能指标上报 (无需认证，前端直接上报)
	metricsGroup.POST("/performance", rm.performanceHandler.SubmitPerformanceMetrics)

	// 需要认证的监控API
	authenticatedMetrics := metricsGroup.Group("")
	authenticatedMetrics.Use(middleware.JWTAuth(rm.config.JWTSecret))

	// 获取仪表板数据
	authenticatedMetrics.GET("/dashboard", rm.performanceHandler.GetDashboardMetrics)

	// 获取告警信息
	authenticatedMetrics.GET("/alerts", rm.performanceHandler.GetPerformanceAlerts)

	// 创建告警
	authenticatedMetrics.POST("/alerts", rm.performanceHandler.CreatePerformanceAlert)
}

// setupPlaceholderMetricsRoutes 设置占位符性能监控路由
func (rm *Manager) setupPlaceholderMetricsRoutes(group *gin.RouterGroup) {
	metricsGroup := group.Group("/metrics")
	metricsGroup.Use(middleware.NewRateLimiter(100))

	metricsGroup.POST("/performance", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Performance metrics endpoint (handler not initialized)",
			"data":    nil,
		})
	})

	authenticatedMetrics := metricsGroup.Group("")
	authenticatedMetrics.Use(middleware.JWTAuth(rm.config.JWTSecret))

	authenticatedMetrics.GET("/dashboard", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Dashboard metrics endpoint (handler not initialized)",
			"data":    nil,
		})
	})

	authenticatedMetrics.GET("/alerts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Alerts endpoint (handler not initialized)",
			"data":    nil,
		})
	})

	authenticatedMetrics.POST("/alerts", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"code":    0,
			"message": "Alert creation endpoint (handler not initialized)",
			"data":    nil,
		})
	})
}

// setupHealthRoutes 设置健康检查路由
func (rm *Manager) setupHealthRoutes(group *gin.RouterGroup) {
	healthGroup := group.Group("/health")
	healthGroup.Use(middleware.NewRateLimiter(200)) // 每分钟200次

	// 服务状态检查 (无需认证)
	if rm.performanceHandler != nil {
		healthGroup.GET("/status", rm.performanceHandler.GetHealthStatus)
	} else {
		healthGroup.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "Health status endpoint (handler not initialized)",
				"data":    nil,
			})
		})
	}

	// 需要认证的健康检查API
	authenticatedHealth := healthGroup.Group("")
	authenticatedHealth.Use(middleware.JWTAuth(rm.config.JWTSecret))

	// 详细健康信息
	authenticatedHealth.GET("/detailed", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Detailed health endpoint (pending implementation)",
			"data":    nil,
		})
	})

	// 健康告警上报
	authenticatedHealth.POST("/alert", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"code":    0,
			"message": "Health alert endpoint (pending implementation)",
			"data":    nil,
		})
	})
}

// setupPostcodeRoutes 设置Postcode地址编码路由
func (rm *Manager) setupPostcodeRoutes(group *gin.RouterGroup) {
	postcodeGroup := group.Group("/postcode")
	postcodeGroup.Use(middleware.NewRateLimiter(120)) // 每分钟120次

	// 地址查询 - 根据完整6位编码查询
	postcodeGroup.GET("/:code", rm.proxyManager.ProxyHandler("write-service"))

	// 地址搜索 - 模糊搜索
	postcodeGroup.GET("/search", rm.proxyManager.ProxyHandler("write-service"))

	// 学校管理
	schoolGroup := postcodeGroup.Group("/schools")
	schoolGroup.GET("", rm.proxyManager.ProxyHandler("write-service"))          // 获取学校列表
	schoolGroup.POST("", rm.proxyManager.ProxyHandler("write-service"))         // 创建学校
	schoolGroup.GET("/:code", rm.proxyManager.ProxyHandler("write-service"))    // 获取学校详情
	schoolGroup.PUT("/:code", rm.proxyManager.ProxyHandler("write-service"))    // 更新学校
	schoolGroup.DELETE("/:code", rm.proxyManager.ProxyHandler("write-service")) // 删除学校

	// 片区管理
	schoolGroup.GET("/:code/areas", rm.proxyManager.ProxyHandler("write-service"))  // 获取学校的片区列表
	schoolGroup.POST("/:code/areas", rm.proxyManager.ProxyHandler("write-service")) // 创建片区

	// 楼栋管理
	schoolGroup.GET("/:code/areas/:area/buildings", rm.proxyManager.ProxyHandler("write-service"))  // 获取片区的楼栋列表
	schoolGroup.POST("/:code/areas/:area/buildings", rm.proxyManager.ProxyHandler("write-service")) // 创建楼栋

	// 房间管理
	schoolGroup.GET("/:code/areas/:area/buildings/:building/rooms", rm.proxyManager.ProxyHandler("write-service"))  // 获取楼栋的房间列表
	schoolGroup.POST("/:code/areas/:area/buildings/:building/rooms", rm.proxyManager.ProxyHandler("write-service")) // 创建房间

	// 权限管理
	permissionGroup := postcodeGroup.Group("/permissions")
	permissionGroup.GET("/:courier_id", rm.proxyManager.ProxyHandler("write-service"))    // 获取信使权限
	permissionGroup.POST("", rm.proxyManager.ProxyHandler("write-service"))               // 分配权限
	permissionGroup.PUT("/:courier_id", rm.proxyManager.ProxyHandler("write-service"))    // 更新权限
	permissionGroup.DELETE("/:courier_id", rm.proxyManager.ProxyHandler("write-service")) // 删除权限

	// 反馈管理
	feedbackGroup := postcodeGroup.Group("/feedback")
	feedbackGroup.GET("", rm.proxyManager.ProxyHandler("write-service"))             // 获取反馈列表
	feedbackGroup.POST("", rm.proxyManager.ProxyHandler("write-service"))            // 提交反馈
	feedbackGroup.GET("/:id", rm.proxyManager.ProxyHandler("write-service"))         // 获取反馈详情
	feedbackGroup.POST("/:id/review", rm.proxyManager.ProxyHandler("write-service")) // 审核反馈

	// 统计分析
	statsGroup := postcodeGroup.Group("/stats")
	statsGroup.GET("", rm.proxyManager.ProxyHandler("write-service"))             // 获取统计数据
	statsGroup.GET("/popular", rm.proxyManager.ProxyHandler("write-service"))     // 热门地址
	statsGroup.GET("/problematic", rm.proxyManager.ProxyHandler("write-service")) // 问题地址
	statsGroup.POST("/usage", rm.proxyManager.ProxyHandler("write-service"))      // 记录使用统计

	// 管理工具
	toolsGroup := postcodeGroup.Group("/tools")
	toolsGroup.POST("/validate", rm.proxyManager.ProxyHandler("write-service")) // 批量验证
	toolsGroup.POST("/import", rm.proxyManager.ProxyHandler("write-service"))   // 批量导入
	toolsGroup.GET("/export", rm.proxyManager.ProxyHandler("write-service"))    // 数据导出
}

// setupAddressRoutes 设置地址相关路由 (兼容性路由)
func (rm *Manager) setupAddressRoutes(group *gin.RouterGroup) {
	addressGroup := group.Group("/address")
	addressGroup.Use(middleware.NewRateLimiter(100)) // 每分钟100次

	// 地址搜索 - 兼容前端现有调用
	addressGroup.GET("/search", rm.proxyManager.ProxyHandler("write-service"))
}

// setupShopRoutes 设置商店相关路由
func (rm *Manager) setupShopRoutes(group *gin.RouterGroup) {
	shopGroup := group.Group("/shop")
	shopGroup.Use(middleware.NewRateLimiter(100)) // 每分钟100次

	// 转发所有商店请求到写信服务
	shopGroup.Any("/*path", rm.proxyManager.ProxyHandler("write-service"))
}

// setupAIRoutes 设置AI相关路由
func (rm *Manager) setupAIRoutes(group *gin.RouterGroup) {
	aiGroup := group.Group("/ai")
	// AI服务可以有更宽松的限制，因为它们通常比较耗时
	aiGroup.Use(middleware.NewRateLimiter(60)) // 每分钟60次

	// 公开的AI路由（无需认证）
	// 转发所有AI请求到主后端服务
	aiGroup.Any("/*path", rm.proxyManager.ProxyHandler("main-backend"))
}

// setupMuseumRoutes 设置博物馆相关路由
func (rm *Manager) setupMuseumRoutes(group *gin.RouterGroup) {
	museumGroup := group.Group("/museum")
	museumGroup.Use(middleware.NewRateLimiter(100)) // 每分钟100次

	// 公开的博物馆路由（无需认证）
	// 转发所有博物馆请求到主后端服务
	museumGroup.Any("/*path", rm.proxyManager.ProxyHandler("main-backend"))
}
