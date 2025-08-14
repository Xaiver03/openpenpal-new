package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"openpenpal-backend/internal/config"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// HardcodedInspiration 硬编码的灵感结构
type HardcodedInspiration struct {
	Theme  string
	Prompt string
	Style  string
	Tags   []string
}

// HardcodedPersona 硬编码的人设结构
type HardcodedPersona struct {
	Name        string
	Description string
	Prompt      string
	Style       string
}

func main() {
	log.Println("🚀 开始AI配置数据迁移...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}
	
	// 连接数据库
	db, err := connectDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 创建表结构
	if err := createTables(db); err != nil {
		log.Fatalf("❌ 创建表失败: %v", err)
	}

	// 迁移硬编码数据
	if err := migrateHardcodedData(db); err != nil {
		log.Fatalf("❌ 数据迁移失败: %v", err)
	}

	log.Println("✅ AI配置数据迁移完成!")
}

func connectDatabase(cfg *config.Config) (*gorm.DB, error) {
	// 构建数据库连接字符串
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DatabaseName, cfg.DBSSLMode)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	log.Println("✅ 数据库连接成功")
	return db, nil
}

func createTables(db *gorm.DB) error {
	log.Println("📦 创建AI配置相关表...")

	// 删除旧的AI配置表（如果存在）
	if err := db.Exec("DROP TABLE IF EXISTS ai_configs CASCADE").Error; err != nil {
		log.Printf("⚠️ 删除旧表失败: %v", err)
	}
	if err := db.Exec("DROP TABLE IF EXISTS ai_content_templates CASCADE").Error; err != nil {
		log.Printf("⚠️ 删除旧表失败: %v", err)
	}
	if err := db.Exec("DROP TABLE IF EXISTS ai_config_history CASCADE").Error; err != nil {
		log.Printf("⚠️ 删除旧表失败: %v", err)
	}

	// AI配置表
	type AIConfig struct {
		ID          string          `gorm:"primaryKey;type:varchar(36)"`
		ConfigType  string          `gorm:"type:varchar(50);not null;index"`
		ConfigKey   string          `gorm:"type:varchar(100);not null"`
		ConfigValue json.RawMessage `gorm:"type:jsonb;not null"`
		Category    string          `gorm:"type:varchar(50);index"`
		IsActive    bool            `gorm:"default:true;index"`
		Priority    int             `gorm:"default:0"`
		Version     int             `gorm:"default:1"`
		CreatedBy   string          `gorm:"type:varchar(36)"`
		CreatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
		UpdatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	}

	// AI内容模板表
	type AIContentTemplate struct {
		ID           string          `gorm:"primaryKey;type:varchar(36)"`
		TemplateType string          `gorm:"type:varchar(50);not null;index"`
		Category     string          `gorm:"type:varchar(50);index"`
		Title        string          `gorm:"type:varchar(200);not null"`
		Content      string          `gorm:"type:text;not null"`
		Tags         pq.StringArray  `gorm:"type:text[]"`
		Metadata     json.RawMessage `gorm:"type:jsonb"`
		UsageCount   int             `gorm:"default:0"`
		Rating       float64         `gorm:"type:decimal(3,2);default:0"`
		QualityScore int             `gorm:"default:0"`
		IsActive     bool            `gorm:"default:true;index"`
		CreatedBy    string          `gorm:"type:varchar(36)"`
		CreatedAt    time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
		UpdatedAt    time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	}

	// AI配置历史表
	type AIConfigHistory struct {
		ID           string          `gorm:"primaryKey;type:varchar(36)"`
		ConfigID     string          `gorm:"type:varchar(36);not null"`
		OldValue     json.RawMessage `gorm:"type:jsonb"`
		NewValue     json.RawMessage `gorm:"type:jsonb"`
		ChangeReason string          `gorm:"type:text"`
		ChangedBy    string          `gorm:"type:varchar(36)"`
		ChangedAt    time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	}

	// 创建表
	tables := []interface{}{
		&AIConfig{},
		&AIContentTemplate{},
		&AIConfigHistory{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			return fmt.Errorf("创建表失败: %w", err)
		}
	}

	// 创建唯一约束
	if err := db.Exec("ALTER TABLE ai_configs ADD CONSTRAINT IF NOT EXISTS uk_ai_configs_type_key UNIQUE (config_type, config_key)").Error; err != nil {
		log.Printf("⚠️ 创建唯一约束失败 (可能已存在): %v", err)
	}

	log.Println("✅ 表创建完成")
	return nil
}

func migrateHardcodedData(db *gorm.DB) error {
	log.Println("🔄 开始迁移硬编码数据...")

	// 检查是否已经迁移过
	var count int64
	db.Table("ai_content_templates").Where("metadata @> ?", `{"source": "hardcoded_migration"}`).Count(&count)
	if count > 0 {
		log.Println("⚠️ 检测到已存在迁移数据，跳过重复迁移")
		return nil
	}

	// 1. 迁移灵感内容池
	if err := migrateInspirations(db); err != nil {
		return fmt.Errorf("迁移灵感数据失败: %w", err)
	}

	// 2. 迁移AI人设配置
	if err := migratePersonas(db); err != nil {
		return fmt.Errorf("迁移人设数据失败: %w", err)
	}

	// 3. 迁移系统提示词配置
	if err := migrateSystemPrompts(db); err != nil {
		return fmt.Errorf("迁移系统提示词失败: %w", err)
	}

	log.Println("✅ 硬编码数据迁移完成")
	return nil
}

func migrateInspirations(db *gorm.DB) error {
	log.Println("📝 迁移灵感内容池...")

	// 从原始代码提取的灵感内容池
	inspirations := getHardcodedInspirations()

	for i, insp := range inspirations {
		metadata, _ := json.Marshal(map[string]interface{}{
			"source":         "hardcoded_migration",
			"original_style": insp.Style,
			"migration_date": time.Now().Format("2006-01-02"),
			"order":          i + 1,
		})

		template := struct {
			ID           string          `gorm:"primaryKey;type:varchar(36)"`
			TemplateType string          `gorm:"type:varchar(50);not null;index"`
			Category     string          `gorm:"type:varchar(50);index"`
			Title        string          `gorm:"type:varchar(200);not null"`
			Content      string          `gorm:"type:text;not null"`
			Tags         pq.StringArray  `gorm:"type:text[]"`
			Metadata     json.RawMessage `gorm:"type:jsonb"`
			UsageCount   int             `gorm:"default:0"`
			Rating       float64         `gorm:"type:decimal(3,2);default:0"`
			QualityScore int             `gorm:"default:0"`
			IsActive     bool            `gorm:"default:true;index"`
			CreatedBy    string          `gorm:"type:varchar(36)"`
			CreatedAt    time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
			UpdatedAt    time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
		}{
			ID:           uuid.New().String(),
			TemplateType: "inspiration",
			Category:     insp.Theme,
			Title:        extractTitle(insp.Prompt),
			Content:      insp.Prompt,
			Tags:         pq.StringArray(insp.Tags),
			Metadata:     metadata,
			UsageCount:   0,
			Rating:       0.0,
			QualityScore: 80, // 初始质量评分
			IsActive:     true,
			CreatedBy:    "system",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := db.Table("ai_content_templates").Create(&template).Error; err != nil {
			return fmt.Errorf("创建灵感模板失败: %w", err)
		}
	}

	log.Printf("✅ 成功迁移 %d 条灵感内容", len(inspirations))
	return nil
}

func migratePersonas(db *gorm.DB) error {
	log.Println("👤 迁移AI人设配置...")

	personas := getHardcodedPersonas()

	for persona, config := range personas {
		configValue, _ := json.Marshal(map[string]interface{}{
			"name":        config.Name,
			"description": config.Description,
			"prompt":      config.Prompt,
			"style":       config.Style,
			"personality": map[string]interface{}{
				"warmth":     0.8,
				"creativity": 0.9,
				"formality":  0.3,
			},
			"constraints": map[string]interface{}{
				"max_length":    1000,
				"tone":          "friendly",
				"avoid_topics":  []string{"politics", "religion"},
			},
		})

		aiConfig := struct {
			ID          string          `gorm:"primaryKey;type:varchar(36)"`
			ConfigType  string          `gorm:"type:varchar(50);not null;index"`
			ConfigKey   string          `gorm:"type:varchar(100);not null"`
			ConfigValue json.RawMessage `gorm:"type:jsonb;not null"`
			Category    string          `gorm:"type:varchar(50);index"`
			IsActive    bool            `gorm:"default:true;index"`
			Priority    int             `gorm:"default:0"`
			Version     int             `gorm:"default:1"`
			CreatedBy   string          `gorm:"type:varchar(36)"`
			CreatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
			UpdatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
		}{
			ID:          uuid.New().String(),
			ConfigType:  "persona",
			ConfigKey:   persona,
			ConfigValue: configValue,
			Category:    "character",
			IsActive:    true,
			Priority:    getPriority(persona),
			Version:     1,
			CreatedBy:   "system",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Table("ai_configs").Create(&aiConfig).Error; err != nil {
			return fmt.Errorf("创建人设配置失败: %w", err)
		}
	}

	log.Printf("✅ 成功迁移 %d 个AI人设", len(personas))
	return nil
}

func migrateSystemPrompts(db *gorm.DB) error {
	log.Println("💬 迁移系统提示词配置...")

	systemPrompts := map[string]string{
		"default": "你是OpenPenPal的AI助手，在这个温暖的数字书信平台上，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好、富有人文情怀的语气回应。",
		"inspiration": "你是一位富有创造力的写作导师，专门为OpenPenPal用户提供深刻而富有诗意的写作灵感。你的建议应该温暖人心，激发用户的创作热情。",
		"matching": "你是一位善解人意的笔友媒人，能够理解信件背后的情感需求，为用户匹配最合适的笔友。注重情感共鸣和兴趣契合。",
		"reply": "你是一位温暖的回信助手，帮助用户写出真诚、感人的回信。保持人文关怀，避免生硬的模板化表达。",
	}

	for key, prompt := range systemPrompts {
		configValue, _ := json.Marshal(map[string]interface{}{
			"prompt": prompt,
			"temperature": 0.9,
			"max_tokens": 1000,
			"context_window": 4000,
			"guidelines": []string{
				"保持温暖友好的语气",
				"避免生硬的AI腔调",
				"重视情感表达和人文关怀",
			},
		})

		config := struct {
			ID          string          `gorm:"primaryKey;type:varchar(36)"`
			ConfigType  string          `gorm:"type:varchar(50);not null;index"`
			ConfigKey   string          `gorm:"type:varchar(100);not null"`
			ConfigValue json.RawMessage `gorm:"type:jsonb;not null"`
			Category    string          `gorm:"type:varchar(50);index"`
			IsActive    bool            `gorm:"default:true;index"`
			Priority    int             `gorm:"default:0"`
			Version     int             `gorm:"default:1"`
			CreatedBy   string          `gorm:"type:varchar(36)"`
			CreatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
			UpdatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
		}{
			ID:          uuid.New().String(),
			ConfigType:  "system_prompt",
			ConfigKey:   key,
			ConfigValue: configValue,
			Category:    "prompts",
			IsActive:    true,
			Priority:    100,
			Version:     1,
			CreatedBy:   "system",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Table("ai_configs").Create(&config).Error; err != nil {
			return fmt.Errorf("创建系统提示词配置失败: %w", err)
		}
	}

	log.Printf("✅ 成功迁移 %d 个系统提示词", len(systemPrompts))
	return nil
}

// 提取标题（取前30个字符作为标题）
func extractTitle(prompt string) string {
	title := strings.TrimSpace(prompt)
	if len(title) > 30 {
		// 找到第一个句号或逗号
		for i, r := range title[:30] {
			if r == '。' || r == '，' || r == '：' {
				return title[:i+1]
			}
		}
		return title[:30] + "..."
	}
	return title
}

// 获取人设优先级
func getPriority(persona string) int {
	priorities := map[string]int{
		"friend":      100,
		"mentor":      90,
		"poet":        80,
		"philosopher": 70,
		"artist":      60,
		"scientist":   50,
		"traveler":    40,
		"historian":   30,
	}
	if p, ok := priorities[persona]; ok {
		return p
	}
	return 50
}

// 从原始代码提取的硬编码灵感数据（先迁移安全的数据）
func getHardcodedInspirations() []HardcodedInspiration {
	return []HardcodedInspiration{
		{
			Theme:  "日常感悟",
			Prompt: "写一写今天早晨醒来时的第一个念头，以及它如何影响了你一天的心情。可以是对新一天的期待，也可以是昨夜梦境的延续。",
			Style:  "温暖治愈",
			Tags:   []string{"日常", "感悟", "晨光"},
		},
		{
			Theme:  "情感表达",
			Prompt: "有什么话你一直想对某个人说，但还没有找到合适的机会？写一封信给ta，不必考虑是否真的会寄出。",
			Style:  "深度情感",
			Tags:   []string{"情感", "表达", "未说出口"},
		},
		{
			Theme:  "时光回忆",
			Prompt: "回想一个让你印象深刻的雨天，那天发生了什么？雨声、气味、当时的心情，都可以成为你笔下的诗意。",
			Style:  "怀旧诗意",
			Tags:   []string{"回忆", "雨天", "诗意"},
		},
		{
			Theme:  "书信情怀",
			Prompt: "想象你是一位古代的书信家，用现代的心境写一封充满古典韵味的信件。可以给月亮、给风、给一朵花。",
			Style:  "古典优雅",
			Tags:   []string{"古典", "书信", "想象"},
		},
		{
			Theme:  "友情珍惜",
			Prompt: "友情中最珍贵的是什么？写一写你和某位朋友之间的温暖时光，或者你想对朋友说的感谢话语。",
			Style:  "温暖感恩",
			Tags:   []string{"友情", "珍惜", "感恩"},
		},
	}
}

// 从原始代码提取的硬编码人设数据
func getHardcodedPersonas() map[string]HardcodedPersona {
	return map[string]HardcodedPersona{
		"poet": {
			Name:        "诗人",
			Description: "一位富有诗意和浪漫情怀的灵魂，善于发现生活中的美好与深意",
			Prompt:      "我是一位诗人，用富有诗意和浪漫的语言与你交流。我善于发现生活中的美好，用温柔的文字表达深刻的情感。",
			Style:       "诗意浪漫",
		},
		"philosopher": {
			Name:        "哲学家",
			Description: "睿智深刻的思考者，喜欢探讨人生的意义和存在的本质",
			Prompt:      "我是一位哲学家，喜欢深入思考人生的意义。我会用理性而温暖的方式与你探讨存在的本质和生活的智慧。",
			Style:       "深刻理性",
		},
		"artist": {
			Name:        "艺术家",
			Description: "充满创造力的艺术创作者，用独特的视角看待世界",
			Prompt:      "我是一位艺术家，对世界有着独特而敏感的观察。我会用创造性的思维和你分享美的发现，激发你的想象力。",
			Style:       "创意感性",
		},
		"scientist": {
			Name:        "科学家",
			Description: "充满好奇心的科学探索者，用理性和逻辑解释世界的奥秘",
			Prompt:      "我是一位科学家，对世界充满好奇和求知欲。我会用科学的思维和你探讨自然的奥秘，分享知识的魅力。",
			Style:       "理性好奇",
		},
		"traveler": {
			Name:        "旅行者",
			Description: "热爱探索的自由灵魂，走过许多地方，见过许多风景",
			Prompt:      "我是一位旅行者，走过很多地方，见过不同的风景和文化。我愿意与你分享旅途中的故事和感悟。",
			Style:       "自由洒脱",
		},
		"historian": {
			Name:        "历史学家",
			Description: "博学的历史研究者，深谙古今变迁，善于从历史中汲取智慧",
			Prompt:      "我是一位历史学家，对古今变迁有着深刻的理解。我会用历史的智慧与你对话，分享时间长河中的故事。",
			Style:       "博学厚重",
		},
		"mentor": {
			Name:        "人生导师",
			Description: "温和睿智的人生指导者，乐于分享人生智慧和经验",
			Prompt:      "我是你的人生导师，拥有丰富的人生阅历。我会用温和而智慧的方式为你答疑解惑，分享人生的智慧。",
			Style:       "温和智慧",
		},
		"friend": {
			Name:        "知心朋友",
			Description: "温暖贴心的好朋友，总是愿意倾听和陪伴",
			Prompt:      "我是你的知心朋友，总是愿意倾听你的心声。我会用最真诚和温暖的态度与你分享生活中的点点滴滴。",
			Style:       "温暖亲切",
		},
	}
}