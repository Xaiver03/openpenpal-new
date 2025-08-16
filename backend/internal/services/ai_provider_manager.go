package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"openpenpal-backend/internal/models"
	"gorm.io/gorm"
)

// AIProviderManager 管理多个AI提供商
type AIProviderManager struct {
	providers       map[string]AIProviderInterface
	configs         map[string]*models.AIConfig
	db              *gorm.DB
	defaultProvider string
	failoverChain   []string
	mutex           sync.RWMutex
	healthCheck     *HealthChecker
}

// NewAIProviderManager 创建AI提供商管理器
func NewAIProviderManager(db *gorm.DB) *AIProviderManager {
	manager := &AIProviderManager{
		providers:     make(map[string]AIProviderInterface),
		configs:       make(map[string]*models.AIConfig),
		db:            db,
		failoverChain: []string{"moonshot", "openai", "claude", "local"},
		healthCheck:   NewHealthChecker(),
	}
	
	// 加载配置并初始化提供商
	if err := manager.LoadConfigurations(); err != nil {
		log.Printf("Failed to load AI configurations: %v", err)
	}
	
	if err := manager.InitializeProviders(); err != nil {
		log.Printf("Failed to initialize AI providers: %v", err)
	}
	
	// 启动健康检查
	go manager.StartHealthMonitoring()
	
	log.Println("✅ AI Provider Manager initialized successfully")
	return manager
}

// LoadConfigurations 加载AI配置
func (m *AIProviderManager) LoadConfigurations() error {
	var configs []models.AIConfig
	if err := m.db.Where("provider IS NOT NULL AND provider != ''").Find(&configs).Error; err != nil {
		return fmt.Errorf("failed to load AI configs: %w", err)
	}
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for _, config := range configs {
		m.configs[string(config.Provider)] = &config
		log.Printf("Loaded config for provider: %s", config.Provider)
		
		// 设置默认提供商
		if config.IsActive && m.defaultProvider == "" {
			m.defaultProvider = string(config.Provider)
		}
	}
	
	return nil
}

// InitializeProviders 初始化所有AI提供商
func (m *AIProviderManager) InitializeProviders() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for providerName, config := range m.configs {
		if !config.IsActive {
			continue
		}
		
		provider, err := m.createProvider(providerName, config)
		if err != nil {
			log.Printf("Failed to create provider %s: %v", providerName, err)
			continue
		}
		
		m.providers[providerName] = provider
		log.Printf("✅ Initialized AI provider: %s", providerName)
	}
	
	return nil
}

// createProvider 根据配置创建具体的提供商实例
func (m *AIProviderManager) createProvider(name string, config *models.AIConfig) (AIProviderInterface, error) {
	switch name {
	case "openai":
		return NewOpenAIProvider(config), nil
	case "claude":
		return NewClaudeProvider(config), nil
	case "moonshot":
		return NewMoonshotProvider(config), nil
	case "gemini":
		return NewGeminiProvider(config), nil
	case "siliconflow":
		return NewSiliconFlowProvider(config), nil
	case "local":
		return NewLocalProvider(config), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}

// GetProvider 获取指定的AI提供商
func (m *AIProviderManager) GetProvider(name string) (AIProviderInterface, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found or not active", name)
	}
	
	return provider, nil
}

// GetAvailableProvider 获取可用的AI提供商（支持故障转移）
func (m *AIProviderManager) GetAvailableProvider(preferredProvider string) (AIProviderInterface, string, error) {
	// 首先尝试指定的提供商
	if preferredProvider != "" {
		if provider, err := m.GetProvider(preferredProvider); err == nil {
			if m.healthCheck.IsHealthy(preferredProvider) {
				return provider, preferredProvider, nil
			}
		}
	}
	
	// 尝试默认提供商
	if m.defaultProvider != "" {
		if provider, err := m.GetProvider(m.defaultProvider); err == nil {
			if m.healthCheck.IsHealthy(m.defaultProvider) {
				return provider, m.defaultProvider, nil
			}
		}
	}
	
	// 故障转移链
	for _, providerName := range m.failoverChain {
		if provider, err := m.GetProvider(providerName); err == nil {
			if m.healthCheck.IsHealthy(providerName) {
				log.Printf("Failover to provider: %s", providerName)
				return provider, providerName, nil
			}
		}
	}
	
	return nil, "", fmt.Errorf("no available AI provider found")
}

// GenerateText 统一的文本生成接口（支持故障转移）
func (m *AIProviderManager) GenerateText(ctx context.Context, prompt string, options AIGenerationOptions, preferredProvider string) (*AIResponse, error) {
	provider, usedProvider, err := m.GetAvailableProvider(preferredProvider)
	if err != nil {
		return nil, err
	}
	
	response, err := provider.GenerateText(ctx, prompt, options)
	if err != nil {
		// 标记提供商为不健康并重试
		m.healthCheck.MarkUnhealthy(usedProvider)
		
		// 尝试故障转移
		provider, newProvider, retryErr := m.GetAvailableProvider("")
		if retryErr != nil {
			return nil, fmt.Errorf("all providers failed: %w", err)
		}
		usedProvider = newProvider
		
		response, err = provider.GenerateText(ctx, prompt, options)
		if err != nil {
			return nil, err
		}
	}
	
	// 更新使用统计
	go m.updateUsageStats(usedProvider, response.TokensUsed)
	
	response.Provider = usedProvider
	return response, nil
}

// Chat 统一的聊天接口
func (m *AIProviderManager) Chat(ctx context.Context, messages []ChatMessage, options AIGenerationOptions, preferredProvider string) (*AIResponse, error) {
	provider, usedProvider, err := m.GetAvailableProvider(preferredProvider)
	if err != nil {
		return nil, err
	}
	
	response, err := provider.Chat(ctx, messages, options)
	if err != nil {
		// 故障转移逻辑
		m.healthCheck.MarkUnhealthy(usedProvider)
		provider, newProvider, retryErr := m.GetAvailableProvider("")
		if retryErr != nil {
			return nil, fmt.Errorf("chat failed on all providers: %w", err)
		}
		usedProvider = newProvider
		
		response, err = provider.Chat(ctx, messages, options)
		if err != nil {
			return nil, err
		}
	}
	
	go m.updateUsageStats(usedProvider, response.TokensUsed)
	response.Provider = usedProvider
	return response, nil
}

// GetProviderStats 获取所有提供商状态
func (m *AIProviderManager) GetProviderStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	
	for name, provider := range m.providers {
		info := provider.GetProviderInfo()
		usage, _ := provider.GetUsage(context.Background())
		
		stats[name] = map[string]interface{}{
			"info":      info,
			"usage":     usage,
			"healthy":   m.healthCheck.IsHealthy(name),
			"is_active": m.configs[name].IsActive,
		}
	}
	
	stats["manager"] = map[string]interface{}{
		"default_provider": m.defaultProvider,
		"failover_chain":   m.failoverChain,
		"total_providers":  len(m.providers),
	}
	
	return stats
}

// updateUsageStats 更新使用统计
func (m *AIProviderManager) updateUsageStats(provider string, tokensUsed int) {
	config, exists := m.configs[provider]
	if !exists {
		return
	}
	
	// 更新数据库中的使用量
	m.db.Model(config).UpdateColumn("used_quota", gorm.Expr("used_quota + ?", tokensUsed))
	
	// 检查配额限制
	if config.UsedQuota >= config.DailyQuota {
		log.Printf("Provider %s has exceeded daily quota", provider)
		m.healthCheck.MarkUnhealthy(provider)
	}
}

// StartHealthMonitoring 启动健康监控
func (m *AIProviderManager) StartHealthMonitoring() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (m *AIProviderManager) performHealthCheck() {
	m.mutex.RLock()
	providers := make(map[string]AIProviderInterface)
	for name, provider := range m.providers {
		providers[name] = provider
	}
	m.mutex.RUnlock()
	
	for name, provider := range providers {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := provider.HealthCheck(ctx); err != nil {
			log.Printf("Health check failed for provider %s: %v", name, err)
			m.healthCheck.MarkUnhealthy(name)
		} else {
			m.healthCheck.MarkHealthy(name)
		}
		cancel()
	}
}

// ReloadConfigurations 重新加载配置
func (m *AIProviderManager) ReloadConfigurations() error {
	log.Println("Reloading AI provider configurations...")
	
	if err := m.LoadConfigurations(); err != nil {
		return err
	}
	
	if err := m.InitializeProviders(); err != nil {
		return err
	}
	
	log.Println("✅ AI provider configurations reloaded successfully")
	return nil
}

// HealthChecker 健康检查器
type HealthChecker struct {
	healthStatus map[string]bool
	mutex        sync.RWMutex
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		healthStatus: make(map[string]bool),
	}
}

// IsHealthy 检查提供商是否健康
func (h *HealthChecker) IsHealthy(provider string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	healthy, exists := h.healthStatus[provider]
	return !exists || healthy // 默认认为是健康的
}

// MarkHealthy 标记提供商为健康
func (h *HealthChecker) MarkHealthy(provider string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.healthStatus[provider] = true
}

// MarkUnhealthy 标记提供商为不健康
func (h *HealthChecker) MarkUnhealthy(provider string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.healthStatus[provider] = false
}