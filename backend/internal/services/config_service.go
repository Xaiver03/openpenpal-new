package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"openpenpal-backend/internal/models"

	"gorm.io/gorm"
)

// ConfigService 配置服务 - 管理AI相关的动态配置
type ConfigService struct {
	db              *gorm.DB
	cache           map[string]interface{}
	templateCache   map[string][]models.AIContentTemplate
	mutex           sync.RWMutex
	lastRefresh     time.Time
	refreshInterval time.Duration
}

// AIConfigData AI配置数据结构
type AIConfigData struct {
	ID          string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ConfigType  string          `json:"config_type" gorm:"type:varchar(50);not null;index"`
	ConfigKey   string          `json:"config_key" gorm:"type:varchar(100);not null"`
	ConfigValue json.RawMessage `json:"config_value" gorm:"type:jsonb;not null"`
	Category    string          `json:"category" gorm:"type:varchar(50);index"`
	IsActive    bool            `json:"is_active" gorm:"default:true;index"`
	Priority    int             `json:"priority" gorm:"default:0"`
	Version     int             `json:"version" gorm:"default:1"`
	CreatedBy   string          `json:"created_by" gorm:"type:varchar(36)"`
	CreatedAt   time.Time       `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// AIContentTemplateData AI内容模板数据结构
type AIContentTemplateData struct {
	ID           string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TemplateType string          `json:"template_type" gorm:"type:varchar(50);not null;index"`
	Category     string          `json:"category" gorm:"type:varchar(50);index"`
	Title        string          `json:"title" gorm:"type:varchar(200);not null"`
	Content      string          `json:"content" gorm:"type:text;not null"`
	Tags         []string        `json:"tags" gorm:"type:text[]"`
	Metadata     json.RawMessage `json:"metadata" gorm:"type:jsonb"`
	UsageCount   int             `json:"usage_count" gorm:"default:0"`
	Rating       float64         `json:"rating" gorm:"type:decimal(3,2);default:0"`
	QualityScore int             `json:"quality_score" gorm:"default:0"`
	IsActive     bool            `json:"is_active" gorm:"default:true;index"`
	CreatedBy    string          `json:"created_by" gorm:"type:varchar(36)"`
	CreatedAt    time.Time       `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// AIConfigHistoryData AI配置历史数据结构
type AIConfigHistoryData struct {
	ID           string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ConfigID     string          `json:"config_id" gorm:"type:varchar(36);not null"`
	OldValue     json.RawMessage `json:"old_value" gorm:"type:jsonb"`
	NewValue     json.RawMessage `json:"new_value" gorm:"type:jsonb"`
	ChangeReason string          `json:"change_reason" gorm:"type:text"`
	ChangedBy    string          `json:"changed_by" gorm:"type:varchar(36)"`
	ChangedAt    time.Time       `json:"changed_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// PersonaConfig 人设配置结构
type PersonaConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Prompt      string                 `json:"prompt"`
	Style       string                 `json:"style"`
	Personality map[string]interface{} `json:"personality"`
	Constraints map[string]interface{} `json:"constraints"`
}

// SystemPromptConfig 系统提示词配置结构
type SystemPromptConfig struct {
	Prompt        string   `json:"prompt"`
	Temperature   float64  `json:"temperature"`
	MaxTokens     int      `json:"max_tokens"`
	ContextWindow int      `json:"context_window"`
	Guidelines    []string `json:"guidelines"`
}

// NewConfigService 创建配置服务实例
func NewConfigService(db *gorm.DB) *ConfigService {
	service := &ConfigService{
		db:              db,
		cache:           make(map[string]interface{}),
		templateCache:   make(map[string][]models.AIContentTemplate),
		refreshInterval: 5 * time.Minute, // 5分钟缓存刷新
	}

	// 启动配置监控
	go service.startConfigWatcher()

	// 初始化缓存
	if err := service.RefreshCache(); err != nil {
		log.Printf("⚠️ [ConfigService] 初始化缓存失败: %v", err)
	}

	log.Println("✅ [ConfigService] 配置服务初始化完成")
	return service
}

// GetConfig 获取配置
func (s *ConfigService) GetConfig(configType, key string) (*AIConfigData, error) {
	cacheKey := fmt.Sprintf("%s:%s", configType, key)

	// 尝试从缓存获取
	s.mutex.RLock()
	if cached, exists := s.cache[cacheKey]; exists {
		s.mutex.RUnlock()
		if config, ok := cached.(AIConfigData); ok {
			return &config, nil
		}
	}
	s.mutex.RUnlock()

	// 从数据库获取
	var config AIConfigData
	err := s.db.Table("ai_configs").
		Where("config_type = ? AND config_key = ? AND is_active = ?", configType, key, true).
		Order("priority DESC").
		First(&config).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("配置不存在: %s:%s", configType, key)
		}
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}

	// 更新缓存
	s.mutex.Lock()
	s.cache[cacheKey] = config
	s.mutex.Unlock()

	return &config, nil
}

// SetConfig 设置配置
func (s *ConfigService) SetConfig(configType, key string, value interface{}, changedBy string) error {
	// 序列化配置值
	configValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化配置值失败: %w", err)
	}

	// 获取现有配置（用于历史记录）
	var oldConfig AIConfigData
	existingConfig := s.db.Table("ai_configs").
		Where("config_type = ? AND config_key = ?", configType, key).
		First(&oldConfig)

	// 开启事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if existingConfig.Error == nil {
		// 更新现有配置
		err = tx.Table("ai_configs").
			Where("config_type = ? AND config_key = ?", configType, key).
			Updates(map[string]interface{}{
				"config_value": configValue,
				"updated_at":   time.Now(),
				"version":      gorm.Expr("version + 1"),
			}).Error

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("更新配置失败: %w", err)
		}

		// 记录配置历史
		if err := s.recordConfigHistory(tx, oldConfig.ID, oldConfig.ConfigValue, configValue, "配置更新", changedBy); err != nil {
			tx.Rollback()
			return fmt.Errorf("记录配置历史失败: %w", err)
		}
	} else {
		// 创建新配置
		newConfig := AIConfigData{
			ID:          generateConfigUUID(),
			ConfigType:  configType,
			ConfigKey:   key,
			ConfigValue: configValue,
			IsActive:    true,
			Priority:    0,
			Version:     1,
			CreatedBy:   changedBy,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := tx.Table("ai_configs").Create(&newConfig).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("创建配置失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 更新缓存
	cacheKey := fmt.Sprintf("%s:%s", configType, key)
	s.mutex.Lock()
	delete(s.cache, cacheKey) // 删除缓存，下次访问时重新加载
	s.mutex.Unlock()

	log.Printf("✅ [ConfigService] 配置已更新: %s:%s", configType, key)
	return nil
}

// GetTemplates 获取内容模板
func (s *ConfigService) GetTemplates(templateType string) ([]models.AIContentTemplate, error) {
	// 尝试从缓存获取
	s.mutex.RLock()
	if cached, exists := s.templateCache[templateType]; exists {
		s.mutex.RUnlock()
		return cached, nil
	}
	s.mutex.RUnlock()

	// 从数据库获取
	var templates []AIContentTemplateData
	err := s.db.Table("ai_content_templates").
		Where("template_type = ? AND is_active = ?", templateType, true).
		Order("priority DESC, rating DESC, usage_count DESC").
		Find(&templates).Error

	if err != nil {
		return nil, fmt.Errorf("获取模板失败: %w", err)
	}

	// 转换为业务模型
	result := make([]models.AIContentTemplate, len(templates))
	for i, t := range templates {
		result[i] = models.AIContentTemplate{
			ID:           t.ID,
			TemplateType: t.TemplateType,
			Category:     t.Category,
			Title:        t.Title,
			Content:      t.Content,
			Tags:         t.Tags,
			Metadata:     parseMetadata(t.Metadata),
			UsageCount:   t.UsageCount,
			Rating:       t.Rating,
			QualityScore: t.QualityScore,
			IsActive:     t.IsActive,
		}
	}

	// 更新缓存
	s.mutex.Lock()
	s.templateCache[templateType] = result
	s.mutex.Unlock()

	return result, nil
}

// GetPersonaConfig 获取人设配置
func (s *ConfigService) GetPersonaConfig(persona string) (*PersonaConfig, error) {
	config, err := s.GetConfig("persona", persona)
	if err != nil {
		return nil, err
	}

	var personaConfig PersonaConfig
	if err := json.Unmarshal(config.ConfigValue, &personaConfig); err != nil {
		return nil, fmt.Errorf("解析人设配置失败: %w", err)
	}

	return &personaConfig, nil
}

// GetSystemPrompt 获取系统提示词
func (s *ConfigService) GetSystemPrompt(promptType string) (*SystemPromptConfig, error) {
	config, err := s.GetConfig("system_prompt", promptType)
	if err != nil {
		// 返回默认提示词
		return s.getDefaultSystemPrompt(promptType), nil
	}

	var promptConfig SystemPromptConfig
	if err := json.Unmarshal(config.ConfigValue, &promptConfig); err != nil {
		return s.getDefaultSystemPrompt(promptType), nil
	}

	return &promptConfig, nil
}

// RefreshCache 刷新缓存
func (s *ConfigService) RefreshCache() error {
	log.Println("🔄 [ConfigService] 刷新配置缓存...")

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 清空缓存
	s.cache = make(map[string]interface{})
	s.templateCache = make(map[string][]models.AIContentTemplate)

	// 预加载常用配置
	commonConfigs := []struct {
		ConfigType string
		ConfigKey  string
	}{
		{"persona", "poet"},
		{"persona", "friend"},
		{"persona", "mentor"},
		{"system_prompt", "default"},
		{"system_prompt", "inspiration"},
	}

	for _, cfg := range commonConfigs {
		var config AIConfigData
		err := s.db.Table("ai_configs").
			Where("config_type = ? AND config_key = ? AND is_active = ?",
				cfg.ConfigType, cfg.ConfigKey, true).
			First(&config).Error

		if err == nil {
			cacheKey := fmt.Sprintf("%s:%s", cfg.ConfigType, cfg.ConfigKey)
			s.cache[cacheKey] = config
		}
	}

	// 预加载常用模板
	templateTypes := []string{"inspiration", "persona", "system_prompt"}
	for _, templateType := range templateTypes {
		templates, err := s.getTemplatesFromDB(templateType)
		if err == nil {
			s.templateCache[templateType] = templates
		}
	}

	s.lastRefresh = time.Now()
	log.Printf("✅ [ConfigService] 缓存刷新完成，加载了 %d 个配置项", len(s.cache))
	return nil
}

// UpdateTemplateUsage 更新模板使用统计
func (s *ConfigService) UpdateTemplateUsage(templateID string) error {
	return s.db.Table("ai_content_templates").
		Where("id = ?", templateID).
		Update("usage_count", gorm.Expr("usage_count + 1")).Error
}

// 启动配置监控
func (s *ConfigService) startConfigWatcher() {
	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		if time.Since(s.lastRefresh) >= s.refreshInterval {
			if err := s.RefreshCache(); err != nil {
				log.Printf("❌ [ConfigService] 定时刷新缓存失败: %v", err)
			}
		}
	}
}

// 记录配置历史
func (s *ConfigService) recordConfigHistory(tx *gorm.DB, configID string, oldValue, newValue json.RawMessage, reason, changedBy string) error {
	history := AIConfigHistoryData{
		ID:           generateConfigUUID(),
		ConfigID:     configID,
		OldValue:     oldValue,
		NewValue:     newValue,
		ChangeReason: reason,
		ChangedBy:    changedBy,
		ChangedAt:    time.Now(),
	}

	return tx.Table("ai_config_history").Create(&history).Error
}

// 从数据库获取模板
func (s *ConfigService) getTemplatesFromDB(templateType string) ([]models.AIContentTemplate, error) {
	var templates []AIContentTemplateData
	err := s.db.Table("ai_content_templates").
		Where("template_type = ? AND is_active = ?", templateType, true).
		Order("priority DESC, rating DESC").
		Find(&templates).Error

	if err != nil {
		return nil, err
	}

	result := make([]models.AIContentTemplate, len(templates))
	for i, t := range templates {
		result[i] = models.AIContentTemplate{
			ID:           t.ID,
			TemplateType: t.TemplateType,
			Category:     t.Category,
			Title:        t.Title,
			Content:      t.Content,
			Tags:         t.Tags,
			Metadata:     parseMetadata(t.Metadata),
			UsageCount:   t.UsageCount,
			Rating:       t.Rating,
			QualityScore: t.QualityScore,
			IsActive:     t.IsActive,
		}
	}

	return result, nil
}

// 获取默认系统提示词
func (s *ConfigService) getDefaultSystemPrompt(promptType string) *SystemPromptConfig {
	defaultPrompts := map[string]*SystemPromptConfig{
		"default": {
			Prompt:        "你是OpenPenPal的AI助手，在这个温暖的数字书信平台上，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好、富有人文情怀的语气回应。",
			Temperature:   0.9,
			MaxTokens:     1000,
			ContextWindow: 4000,
			Guidelines:    []string{"保持温暖友好的语气", "避免生硬的AI腔调", "重视情感表达和人文关怀"},
		},
		"inspiration": {
			Prompt:        "你是一位富有创造力的写作导师，专门为OpenPenPal用户提供深刻而富有诗意的写作灵感。你的建议应该温暖人心，激发用户的创作热情。",
			Temperature:   0.95,
			MaxTokens:     800,
			ContextWindow: 2000,
			Guidelines:    []string{"激发创作热情", "提供具体的写作建议", "保持诗意和深度"},
		},
		"matching": {
			Prompt:        "你是一位善解人意的笔友媒人，能够理解信件背后的情感需求，为用户匹配最合适的笔友。注重情感共鸣和兴趣契合。",
			Temperature:   0.8,
			MaxTokens:     600,
			ContextWindow: 3000,
			Guidelines:    []string{"关注情感共鸣", "分析兴趣匹配度", "提供匹配理由"},
		},
	}

	if prompt, exists := defaultPrompts[promptType]; exists {
		return prompt
	}
	return defaultPrompts["default"]
}

// 解析元数据
func parseMetadata(raw json.RawMessage) map[string]interface{} {
	var metadata map[string]interface{}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return make(map[string]interface{})
	}
	return metadata
}

// 生成UUID的辅助函数（简化版）
func generateConfigUUID() string {
	// 这里应该使用 uuid.New().String()，但为了简化依赖，使用时间戳
	return fmt.Sprintf("cfg_%d", time.Now().UnixNano())
}
