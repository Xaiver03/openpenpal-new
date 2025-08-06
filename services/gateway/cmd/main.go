package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/database"
	"api-gateway/internal/discovery"
	"api-gateway/internal/handlers"
	"api-gateway/internal/monitor"
	"api-gateway/internal/proxy"
	"api-gateway/internal/router"
	"api-gateway/internal/services"
	"log"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志
	logger := monitor.InitLogger(cfg.LogLevel)
	defer logger.Sync()

	// 初始化监控
	monitor.InitMetrics()

	// 初始化数据库
	db, err := database.InitDB(cfg.DatabaseURL, logger)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		if err := database.CloseDB(db); err != nil {
			logger.Error("Failed to close database", zap.Error(err))
		}
	}()

	// 初始化性能监控服务
	metricsService := services.NewMetricsService(db, logger)
	metricsHandler := handlers.NewMetricsHandler(metricsService, logger)

	// 初始化服务发现
	serviceDiscovery := discovery.NewServiceDiscovery(cfg)
	serviceDiscovery.SetLogger(logger)
	
	// 启动服务健康检查
	go serviceDiscovery.StartHealthCheck()

	// 初始化代理管理器
	proxyManager := proxy.NewProxyManager(serviceDiscovery, logger)

	// 初始化路由管理器
	routerManager := router.NewRouterManager(cfg, proxyManager, logger)
	routerManager.SetMetricsHandler(metricsHandler)
	
	// 设置所有路由
	ginRouter := routerManager.SetupRoutes()

	// 启动服务器
	logger.Info("Starting API Gateway",
		zap.String("port", cfg.Port),
		zap.String("environment", cfg.Environment),
		zap.Int("proxy_timeout", cfg.ProxyTimeout),
		zap.Bool("metrics_enabled", cfg.MetricsEnabled),
	)

	if err := ginRouter.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}