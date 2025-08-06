package proxy

import (
	"api-gateway/internal/discovery"
	"api-gateway/internal/models"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProxyManager 代理管理器
type ProxyManager struct {
	serviceDiscovery *discovery.ServiceDiscovery
	logger           *zap.Logger
	proxies          map[string]*httputil.ReverseProxy
}

// NewProxyManager 创建代理管理器
func NewProxyManager(serviceDiscovery *discovery.ServiceDiscovery, logger *zap.Logger) *ProxyManager {
	return &ProxyManager{
		serviceDiscovery: serviceDiscovery,
		logger:           logger,
		proxies:          make(map[string]*httputil.ReverseProxy),
	}
}

// ProxyHandler 创建代理处理器
func (pm *ProxyManager) ProxyHandler(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取目标服务实例
		instance, err := pm.serviceDiscovery.GetHealthyInstance(serviceName)
		if err != nil {
			pm.logger.Error("Failed to get healthy instance", 
				zap.String("service", serviceName),
				zap.Error(err),
			)
			
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Code:      http.StatusServiceUnavailable,
				Message:   "Service temporarily unavailable",
				Timestamp: time.Now(),
				Path:      c.Request.URL.Path,
			})
			return
		}

		// 创建代理
		proxy := pm.getOrCreateProxy(serviceName, instance.Host)

		// 记录请求开始时间
		startTime := time.Now()

		// 设置请求上下文
		pm.setupProxyRequest(c, serviceName)

		// 执行代理
		proxy.ServeHTTP(c.Writer, c.Request)

		// 记录请求完成
		duration := time.Since(startTime)
		pm.logRequest(c, serviceName, instance.Host, duration)
	}
}

// getOrCreateProxy 获取或创建代理
func (pm *ProxyManager) getOrCreateProxy(serviceName, targetHost string) *httputil.ReverseProxy {
	key := fmt.Sprintf("%s-%s", serviceName, targetHost)
	
	if proxy, exists := pm.proxies[key]; exists {
		return proxy
	}

	target, _ := url.Parse(targetHost)
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义Director函数
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		pm.modifyRequest(req, serviceName)
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		pm.handleProxyError(w, r, serviceName, err)
	}

	// 自定义响应修改
	proxy.ModifyResponse = func(resp *http.Response) error {
		return pm.modifyResponse(resp, serviceName)
	}

	pm.proxies[key] = proxy
	return proxy
}

// setupProxyRequest 设置代理请求
func (pm *ProxyManager) setupProxyRequest(c *gin.Context, serviceName string) {
	// 添加跟踪ID
	traceID := c.GetHeader("X-Trace-ID")
	if traceID == "" {
		traceID = generateTraceID()
		c.Header("X-Trace-ID", traceID)
	}

	// 添加网关信息
	c.Request.Header.Set("X-Gateway", "openpenpal-api-gateway")
	c.Request.Header.Set("X-Service-Name", serviceName)
	c.Request.Header.Set("X-Forwarded-For", c.ClientIP())
	c.Request.Header.Set("X-Real-IP", c.ClientIP())

	// 保留原始Host信息
	if c.Request.Header.Get("X-Original-Host") == "" {
		c.Request.Header.Set("X-Original-Host", c.Request.Host)
	}
}

// modifyRequest 修改请求
func (pm *ProxyManager) modifyRequest(req *http.Request, serviceName string) {
	// 重写路径：移除服务前缀
	req.URL.Path = pm.rewritePath(req.URL.Path, serviceName)
	
	// 设置User-Agent
	req.Header.Set("User-Agent", "OpenPenPal-API-Gateway/1.0")
	
	pm.logger.Debug("Proxy request modified",
		zap.String("service", serviceName),
		zap.String("path", req.URL.Path),
		zap.String("method", req.Method),
	)
}

// rewritePath 重写请求路径
func (pm *ProxyManager) rewritePath(originalPath, serviceName string) string {
	// 移除 /api/v1 前缀
	path := strings.TrimPrefix(originalPath, "/api/v1")
	
	// 根据服务名重写路径
	switch serviceName {
	case "main-backend":
		// 保持原有路径结构
		return "/api/v1" + path
	case "write-service":
		// 移除 /letters 前缀
		return "/api" + strings.TrimPrefix(path, "/letters")
	case "courier-service":
		// 保持 /courier 前缀
		return "/api" + path
	case "admin-service":
		// 保持 /admin 前缀
		return "/api" + path
	case "ocr-service":
		// 保持 /ocr 前缀
		return "/api" + path
	default:
		return path
	}
}

// modifyResponse 修改响应
func (pm *ProxyManager) modifyResponse(resp *http.Response, serviceName string) error {
	// 添加服务标识头
	resp.Header.Set("X-Service-Source", serviceName)
	resp.Header.Set("X-Gateway", "openpenpal-api-gateway")
	
	// 统一CORS头
	resp.Header.Set("Access-Control-Allow-Origin", "*")
	resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	resp.Header.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

	return nil
}

// handleProxyError 处理代理错误
func (pm *ProxyManager) handleProxyError(w http.ResponseWriter, r *http.Request, serviceName string, err error) {
	pm.logger.Error("Proxy error", 
		zap.String("service", serviceName),
		zap.String("path", r.URL.Path),
		zap.Error(err),
	)

	// 标记服务实例为不健康
	pm.serviceDiscovery.MarkUnhealthy(serviceName, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	
	errorResponse := models.ErrorResponse{
		Code:      http.StatusBadGateway,
		Message:   "Service temporarily unavailable",
		Details:   err.Error(),
		Timestamp: time.Now(),
		Path:      r.URL.Path,
	}

	// 写入错误响应
	if respBytes, err := errorResponse.ToJSON(); err == nil {
		w.Write(respBytes)
	}
}

// logRequest 记录请求日志
func (pm *ProxyManager) logRequest(c *gin.Context, serviceName, targetHost string, duration time.Duration) {
	pm.logger.Info("Proxy request completed",
		zap.String("service", serviceName),
		zap.String("target", targetHost),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int("status", c.Writer.Status()),
		zap.Duration("duration", duration),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("trace_id", c.GetHeader("X-Trace-ID")),
	)
}

// HealthCheckHandler 健康检查处理器
func (pm *ProxyManager) HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("service")
		
		if serviceName == "" {
			// 检查所有服务
			status := pm.serviceDiscovery.GetAllServicesHealth()
			c.JSON(http.StatusOK, gin.H{
				"status":   "ok",
				"services": status,
			})
			return
		}

		// 检查特定服务
		healthy := pm.serviceDiscovery.IsServiceHealthy(serviceName)
		if healthy {
			c.JSON(http.StatusOK, gin.H{
				"status":  "healthy",
				"service": serviceName,
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"service": serviceName,
			})
		}
	}
}

// generateTraceID 生成跟踪ID
func generateTraceID() string {
	return fmt.Sprintf("gw-%d-%s", time.Now().UnixNano(), randomString(8))
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}