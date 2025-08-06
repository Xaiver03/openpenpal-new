package discovery

import (
	"api-gateway/internal/config"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	Name       string    `json:"name"`
	Host       string    `json:"host"`
	Healthy    bool      `json:"healthy"`
	Weight     int       `json:"weight"`
	LastCheck  time.Time `json:"last_check"`
	ErrorCount int       `json:"error_count"`
}

// ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	config    *config.Config
	logger    *zap.Logger
	services  map[string][]*ServiceInstance
	mutex     sync.RWMutex
	client    *http.Client
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewServiceDiscovery 创建服务发现实例
func NewServiceDiscovery(cfg *config.Config) *ServiceDiscovery {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(cfg.ConnectTimeout) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     time.Duration(cfg.KeepAliveTimeout) * time.Second,
		},
	}

	sd := &ServiceDiscovery{
		config:   cfg,
		logger:   zap.NewNop(), // 将在main中设置
		services: make(map[string][]*ServiceInstance),
		client:   client,
		ctx:      ctx,
		cancel:   cancel,
	}

	// 初始化服务实例
	sd.initializeServices()

	return sd
}

// SetLogger 设置日志器
func (sd *ServiceDiscovery) SetLogger(logger *zap.Logger) {
	sd.logger = logger
}

// initializeServices 初始化服务实例
func (sd *ServiceDiscovery) initializeServices() {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	for serviceName, serviceConfig := range sd.config.Services {
		var instances []*ServiceInstance
		
		for _, host := range serviceConfig.Hosts {
			instance := &ServiceInstance{
				Name:       serviceName,
				Host:       host,
				Healthy:    false, // 初始状态为不健康，等待健康检查
				Weight:     serviceConfig.Weight,
				LastCheck:  time.Now(),
				ErrorCount: 0,
			}
			instances = append(instances, instance)
		}
		
		sd.services[serviceName] = instances
	}
}

// StartHealthCheck 启动健康检查
func (sd *ServiceDiscovery) StartHealthCheck() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	// 立即执行一次健康检查
	sd.performHealthCheck()

	for {
		select {
		case <-ticker.C:
			sd.performHealthCheck()
		case <-sd.ctx.Done():
			return
		}
	}
}

// performHealthCheck 执行健康检查
func (sd *ServiceDiscovery) performHealthCheck() {
	sd.mutex.RLock()
	services := make(map[string][]*ServiceInstance)
	for name, instances := range sd.services {
		services[name] = instances
	}
	sd.mutex.RUnlock()

	// 并发检查所有服务实例
	var wg sync.WaitGroup
	
	for serviceName, instances := range services {
		for _, instance := range instances {
			wg.Add(1)
			
			go func(svcName string, inst *ServiceInstance) {
				defer wg.Done()
				sd.checkInstanceHealth(svcName, inst)
			}(serviceName, instance)
		}
	}
	
	wg.Wait()
}

// checkInstanceHealth 检查单个实例健康状态
func (sd *ServiceDiscovery) checkInstanceHealth(serviceName string, instance *ServiceInstance) {
	serviceConfig := sd.config.GetServiceConfig(serviceName)
	if serviceConfig == nil {
		return
	}

	healthURL := instance.Host + serviceConfig.HealthCheck
	
	ctx, cancel := context.WithTimeout(sd.ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		sd.markInstanceUnhealthy(instance, fmt.Errorf("create request failed: %w", err))
		return
	}

	resp, err := sd.client.Do(req)
	if err != nil {
		sd.markInstanceUnhealthy(instance, fmt.Errorf("health check failed: %w", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		sd.markInstanceHealthy(instance)
	} else {
		sd.markInstanceUnhealthy(instance, fmt.Errorf("unhealthy status: %d", resp.StatusCode))
	}
}

// markInstanceHealthy 标记实例为健康
func (sd *ServiceDiscovery) markInstanceHealthy(instance *ServiceInstance) {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	wasUnhealthy := !instance.Healthy
	instance.Healthy = true
	instance.LastCheck = time.Now()
	instance.ErrorCount = 0

	if wasUnhealthy {
		sd.logger.Info("Service instance recovered",
			zap.String("service", instance.Name),
			zap.String("host", instance.Host),
		)
	}
}

// markInstanceUnhealthy 标记实例为不健康
func (sd *ServiceDiscovery) markInstanceUnhealthy(instance *ServiceInstance, err error) {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	wasHealthy := instance.Healthy
	instance.Healthy = false
	instance.LastCheck = time.Now()
	instance.ErrorCount++

	if wasHealthy {
		sd.logger.Warn("Service instance became unhealthy",
			zap.String("service", instance.Name),
			zap.String("host", instance.Host),
			zap.Error(err),
		)
	}
}

// GetHealthyInstance 获取健康的服务实例
func (sd *ServiceDiscovery) GetHealthyInstance(serviceName string) (*ServiceInstance, error) {
	sd.mutex.RLock()
	instances := sd.services[serviceName]
	sd.mutex.RUnlock()

	if len(instances) == 0 {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// 获取所有健康的实例
	var healthyInstances []*ServiceInstance
	for _, instance := range instances {
		if instance.Healthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("no healthy instances for service %s", serviceName)
	}

	// 使用加权随机选择算法
	return sd.selectInstanceByWeight(healthyInstances), nil
}

// selectInstanceByWeight 基于权重选择实例
func (sd *ServiceDiscovery) selectInstanceByWeight(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 1 {
		return instances[0]
	}

	// 计算总权重
	totalWeight := 0
	for _, instance := range instances {
		totalWeight += instance.Weight
	}

	if totalWeight == 0 {
		// 如果所有权重都为0，随机选择
		return instances[rand.Intn(len(instances))]
	}

	// 加权随机选择
	randWeight := rand.Intn(totalWeight)
	currentWeight := 0

	for _, instance := range instances {
		currentWeight += instance.Weight
		if randWeight < currentWeight {
			return instance
		}
	}

	// 默认返回第一个实例
	return instances[0]
}

// MarkUnhealthy 标记服务实例为不健康
func (sd *ServiceDiscovery) MarkUnhealthy(serviceName, host string) {
	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	instances := sd.services[serviceName]
	for _, instance := range instances {
		if host == "" || instance.Host == host {
			instance.Healthy = false
			instance.ErrorCount++
			instance.LastCheck = time.Now()
		}
	}
}

// IsServiceHealthy 检查服务是否健康
func (sd *ServiceDiscovery) IsServiceHealthy(serviceName string) bool {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	instances := sd.services[serviceName]
	for _, instance := range instances {
		if instance.Healthy {
			return true
		}
	}
	return false
}

// GetAllServicesHealth 获取所有服务的健康状态
func (sd *ServiceDiscovery) GetAllServicesHealth() map[string]interface{} {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	result := make(map[string]interface{})
	
	for serviceName, instances := range sd.services {
		healthyCount := 0
		totalCount := len(instances)
		
		var instancesStatus []map[string]interface{}
		
		for _, instance := range instances {
			if instance.Healthy {
				healthyCount++
			}
			
			instancesStatus = append(instancesStatus, map[string]interface{}{
				"host":        instance.Host,
				"healthy":     instance.Healthy,
				"weight":      instance.Weight,
				"last_check":  instance.LastCheck,
				"error_count": instance.ErrorCount,
			})
		}
		
		result[serviceName] = map[string]interface{}{
			"healthy_instances": healthyCount,
			"total_instances":   totalCount,
			"status":           fmt.Sprintf("%d/%d healthy", healthyCount, totalCount),
			"instances":        instancesStatus,
		}
	}
	
	return result
}

// GetServiceInstances 获取服务的所有实例
func (sd *ServiceDiscovery) GetServiceInstances(serviceName string) []*ServiceInstance {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	instances := sd.services[serviceName]
	result := make([]*ServiceInstance, len(instances))
	copy(result, instances)
	
	return result
}

// Stop 停止服务发现
func (sd *ServiceDiscovery) Stop() {
	if sd.cancel != nil {
		sd.cancel()
	}
}