package models

import (
	"time"
)

// AIProvider AI服务提供商
type AIProvider string

const (
	ProviderOpenAI     AIProvider = "openai"
	ProviderClaude     AIProvider = "claude"
	ProviderSiliconFlow AIProvider = "siliconflow"
	ProviderMoonshot   AIProvider = "moonshot"
	ProviderGemini     AIProvider = "gemini"
	ProviderLocal      AIProvider = "local"      // 本地生成（开发环境）
	ProviderDefault    AIProvider = "moonshot"
)

// AIPersona AI笔友人设
type AIPersona string

const (
	PersonaPoet        AIPersona = "poet"        // 诗人
	PersonaPhilosopher AIPersona = "philosopher" // 哲学家
	PersonaArtist      AIPersona = "artist"      // 艺术家
	PersonaScientist   AIPersona = "scientist"   // 科学家
	PersonaTraveler    AIPersona = "traveler"    // 旅行者
	PersonaHistorian   AIPersona = "historian"   // 历史学家
	PersonaMentor      AIPersona = "mentor"      // 人生导师
	PersonaFriend      AIPersona = "friend"      // 知心朋友
)

// AITaskType AI任务类型
type AITaskType string

const (
	TaskTypeMatch       AITaskType = "match"       // 笔友匹配
	TaskTypeReply       AITaskType = "reply"       // 生成回信
	TaskTypeInspiration AITaskType = "inspiration" // 写作灵感
	TaskTypeCurate      AITaskType = "curate"      // 内容策展
	TaskTypeModerate    AITaskType = "moderate"    // 内容审核
)

// AIContentTemplate AI内容模板
type AIContentTemplate struct {
	ID           string                 `json:"id"`
	TemplateType string                 `json:"template_type"`
	Category     string                 `json:"category"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	UsageCount   int                    `json:"usage_count"`
	Rating       float64                `json:"rating"`
	QualityScore int                    `json:"quality_score"`
	IsActive     bool                   `json:"is_active"`
}

// AIPersonaInfo AI人设信息
type AIPersonaInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Style       string `json:"style"`
	Available   bool   `json:"available"`
}

// AIMatch AI匹配记录
type AIMatch struct {
	ID            string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID      string     `json:"letter_id" gorm:"type:varchar(36);not null;index"`
	MatchedUserID string     `json:"matched_user_id" gorm:"type:varchar(36);index"`
	MatchScore    float64    `json:"match_score" gorm:"default:0"`
	MatchReason   string     `json:"match_reason" gorm:"type:text"`
	Status        string     `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending/accepted/rejected
	Provider      AIProvider `json:"provider" gorm:"type:varchar(20)"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// AIReply AI回信记录
type AIReply struct {
	ID               string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	OriginalLetterID string     `json:"original_letter_id" gorm:"type:varchar(36);not null;index"`
	ReplyLetterID    string     `json:"reply_letter_id" gorm:"type:varchar(36);index"`
	Persona          AIPersona  `json:"persona" gorm:"type:varchar(20);not null"`
	Provider         AIProvider `json:"provider" gorm:"type:varchar(20)"`
	DelayHours       int        `json:"delay_hours" gorm:"default:24"` // 延迟回信时间
	ScheduledAt      time.Time  `json:"scheduled_at"`
	SentAt           *time.Time `json:"sent_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

// AIInspiration AI写作灵感
type AIInspiration struct {
	ID         string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID     string     `json:"user_id" gorm:"type:varchar(36);index"`
	Theme      string     `json:"theme" gorm:"type:varchar(100)"`
	Prompt     string     `json:"prompt" gorm:"type:text;not null"`
	Style      string     `json:"style" gorm:"type:varchar(50)"`
	Tags       string     `json:"tags" gorm:"type:text"`
	UsageCount int        `json:"usage_count" gorm:"default:0"`
	Provider   AIProvider `json:"provider" gorm:"type:varchar(20)"`
	CreatedAt  time.Time  `json:"created_at"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
}

// AICuration AI策展记录
type AICuration struct {
	ID           string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID     string     `json:"letter_id" gorm:"type:varchar(36);not null;index"`
	ExhibitionID *string    `json:"exhibition_id" gorm:"type:varchar(36);index"`
	Category     string     `json:"category" gorm:"type:varchar(50)"`
	Tags         string     `json:"tags" gorm:"type:text"`
	Summary      string     `json:"summary" gorm:"type:text"`
	Highlights   string     `json:"highlights" gorm:"type:text"`
	Score        float64    `json:"score" gorm:"default:0"` // 策展质量分
	Provider     AIProvider `json:"provider" gorm:"type:varchar(20)"`
	CreatedAt    time.Time  `json:"created_at"`
	ApprovedAt   *time.Time `json:"approved_at"`
}

// AIReplyAdvice AI回信角度建议
type AIReplyAdvice struct {
	ID               string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID         string     `json:"letter_id" gorm:"type:varchar(36);not null;index"`
	UserID           string     `json:"user_id" gorm:"type:varchar(36);not null;index"`
	PersonaType      string     `json:"persona_type" gorm:"type:varchar(50)"` // custom, predefined
	PersonaName      string     `json:"persona_name" gorm:"type:varchar(100)"`
	PersonaDesc      string     `json:"persona_desc" gorm:"type:text"`
	Perspectives     string     `json:"perspectives" gorm:"type:text"` // JSON array of perspectives
	EmotionalTone    string     `json:"emotional_tone" gorm:"type:varchar(50)"`
	SuggestedTopics  string     `json:"suggested_topics" gorm:"type:text"`
	WritingStyle     string     `json:"writing_style" gorm:"type:varchar(50)"`
	KeyPoints        string     `json:"key_points" gorm:"type:text"`
	DeliveryDelay    int        `json:"delivery_delay" gorm:"default:0"` // 延迟天数
	ScheduledFor     *time.Time `json:"scheduled_for"`
	Provider         AIProvider `json:"provider" gorm:"type:varchar(20)"`
	CreatedAt        time.Time  `json:"created_at"`
	UsedAt           *time.Time `json:"used_at"`
}

// AIConfig AI配置
type AIConfig struct {
	ID           string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Provider     AIProvider `json:"provider" gorm:"type:varchar(20);uniqueIndex"`
	APIKey       string     `json:"-" gorm:"type:varchar(255)"` // 加密存储
	APIEndpoint  string     `json:"api_endpoint" gorm:"type:varchar(500)"`
	Model        string     `json:"model" gorm:"type:varchar(100)"`
	Temperature  float64    `json:"temperature" gorm:"default:0.7"`
	MaxTokens    int        `json:"max_tokens" gorm:"default:1000"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	Priority     int        `json:"priority" gorm:"default:0"` // 优先级，用于负载均衡
	DailyQuota   int        `json:"daily_quota" gorm:"default:10000"`
	UsedQuota    int        `json:"used_quota" gorm:"default:0"`
	QuotaResetAt time.Time  `json:"quota_reset_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// AIUsageLog AI使用日志
type AIUsageLog struct {
	ID           string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID       string     `json:"user_id" gorm:"type:varchar(36);index"`
	TaskType     AITaskType `json:"task_type" gorm:"type:varchar(20);not null"`
	TaskID       string     `json:"task_id" gorm:"type:varchar(36);index"` // 关联的任务ID
	Provider     AIProvider `json:"provider" gorm:"type:varchar(20)"`
	Model        string     `json:"model" gorm:"type:varchar(100)"`
	InputTokens  int        `json:"input_tokens" gorm:"default:0"`
	OutputTokens int        `json:"output_tokens" gorm:"default:0"`
	TotalTokens  int        `json:"total_tokens" gorm:"default:0"`
	ResponseTime int        `json:"response_time" gorm:"default:0"` // 毫秒
	Status       string     `json:"status" gorm:"type:varchar(20)"` // success/failed/timeout
	ErrorMessage string     `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Request/Response DTOs

// DelayConfig 精确延迟配置
type DelayConfig struct {
	Type            string     `json:"type"`                      // "preset", "relative", "absolute"
	PresetOption    string     `json:"preset_option,omitempty"`   // "1hour", "tomorrow", "nextweek", "weekend"
	RelativeDays    int        `json:"relative_days,omitempty"`   // 相对天数
	RelativeHours   int        `json:"relative_hours,omitempty"`  // 相对小时
	RelativeMinutes int        `json:"relative_minutes,omitempty"` // 相对分钟
	AbsoluteTime    *time.Time `json:"absolute_time,omitempty"`   // 绝对时间
	Timezone        string     `json:"timezone,omitempty"`        // 时区，默认用户本地时区
	UserDescription string     `json:"user_description,omitempty"` // 用户自定义描述
}

// AIMatchRequest AI匹配请求
type AIMatchRequest struct {
	LetterID     string       `json:"letter_id" binding:"required"`
	MaxMatches   int          `json:"max_matches"` // 最大匹配数，默认3
	DelayOption  string       `json:"delay_option,omitempty"`  // 兼容旧版：quick/normal/slow
	DelayConfig  *DelayConfig `json:"delay_config,omitempty"`  // 新版精确配置
}

// AIReplyRequest AI回信请求
type AIReplyRequest struct {
	LetterID       string    `json:"letter_id" binding:"required"`
	Persona        AIPersona `json:"persona" binding:"required"`
	PersonaType    string    `json:"persona_type,omitempty"` // "custom" or "predefined"
	PersonaDesc    string    `json:"persona_desc,omitempty"` // 自定义人设描述
	DelayHours     int       `json:"delay_hours"` // 延迟小时数，默认24
	DeliveryDays   int       `json:"delivery_days,omitempty"` // 延迟天数
	OriginalLetter *Letter   `json:"original_letter,omitempty"` // 原始信件信息
}

// AIReplyAdviceRequest AI回信建议请求
type AIReplyAdviceRequest struct {
	LetterID     string `json:"letter_id" binding:"required"`
	PersonaType  string `json:"persona_type" binding:"required"` // "custom", "predefined", "deceased", "distant_friend", "unspoken_love" 
	PersonaName  string `json:"persona_name" binding:"required"`
	PersonaDesc  string `json:"persona_desc,omitempty"` // 详细人设描述
	Relationship string `json:"relationship,omitempty"` // 关系描述：如"已故的奶奶"、"多年未见的好友"、"暗恋的人"
	DeliveryDays int    `json:"delivery_days,omitempty"` // 延迟天数，0表示立即
}

// AIInspirationRequest AI灵感请求
type AIInspirationRequest struct {
	Theme string   `json:"theme"`
	Style string   `json:"style"`
	Tags  []string `json:"tags"`
	Count int      `json:"count"` // 生成数量，默认1
}

// AICurateRequest AI策展请求
type AICurateRequest struct {
	LetterIDs    []string `json:"letter_ids" binding:"required"`
	ExhibitionID string   `json:"exhibition_id"`
	AutoApprove  bool     `json:"auto_approve"`
}

// AIMatchResponse AI匹配响应
type AIMatchResponse struct {
	Status   string                 `json:"status,omitempty"`
	Message  string                 `json:"message,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Matches  []struct {
		UserID     string   `json:"user_id"`
		Username   string   `json:"username"`
		Score      float64  `json:"score"`
		Reason     string   `json:"reason"`
		CommonTags []string `json:"common_tags"`
	} `json:"matches"`
}

// AIInspirationResponse AI灵感响应
type AIInspirationResponse struct {
	Inspirations []struct {
		ID     string   `json:"id"`
		Theme  string   `json:"theme"`
		Prompt string   `json:"prompt"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
	} `json:"inspirations"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DelayQueueRecord 延迟队列记录
type DelayQueueRecord struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TaskType     string    `json:"task_type" gorm:"type:varchar(50);not null"`
	Payload      string    `json:"payload" gorm:"type:text;not null"`
	DelayedUntil time.Time `json:"delayed_until" gorm:"not null;index"`
	Status       string    `json:"status" gorm:"type:varchar(20);default:'pending';index"` // pending/processing/completed/failed
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// AIUsageStats AI使用统计
type AIUsageStats struct {
	UserID          string    `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	RequestCount    int       `json:"request_count" gorm:"default:0"`
	LastRequestAt   time.Time `json:"last_request_at"`
	DailyLimit      int       `json:"daily_limit" gorm:"default:10"`
	MonthlyLimit    int       `json:"monthly_limit" gorm:"default:100"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
