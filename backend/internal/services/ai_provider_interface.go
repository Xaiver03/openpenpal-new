package services

import (
	"context"
	"time"
)

// AIProviderInterface 定义所有AI提供商必须实现的接口
type AIProviderInterface interface {
	// 基础文本生成
	GenerateText(ctx context.Context, prompt string, options AIGenerationOptions) (*AIResponse, error)
	
	// 聊天对话
	Chat(ctx context.Context, messages []ChatMessage, options AIGenerationOptions) (*AIResponse, error)
	
	// 文本总结
	Summarize(ctx context.Context, text string, options AIGenerationOptions) (*AIResponse, error)
	
	// 内容翻译
	Translate(ctx context.Context, text, targetLang string, options AIGenerationOptions) (*AIResponse, error)
	
	// 情感分析
	AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error)
	
	// 内容审核
	ModerateContent(ctx context.Context, text string) (*ContentModeration, error)
	
	// 提供商信息
	GetProviderInfo() ProviderInfo
	
	// 健康检查
	HealthCheck(ctx context.Context) error
	
	// 获取使用量
	GetUsage(ctx context.Context) (*UsageInfo, error)
}

// AIGenerationOptions AI生成选项
type AIGenerationOptions struct {
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p,omitempty"`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64 `json:"presence_penalty,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	Model            string  `json:"model,omitempty"`
	Stream           bool    `json:"stream,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// AIResponse AI响应
type AIResponse struct {
	Content      string            `json:"content"`
	TokensUsed   int              `json:"tokens_used"`
	Model        string           `json:"model"`
	Provider     string           `json:"provider"`
	RequestID    string           `json:"request_id"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}

// SentimentAnalysis 情感分析结果
type SentimentAnalysis struct {
	Sentiment string  `json:"sentiment"` // positive, negative, neutral
	Score     float64 `json:"score"`     // -1.0 到 1.0
	Confidence float64 `json:"confidence"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ContentModeration 内容审核结果
type ContentModeration struct {
	Flagged    bool               `json:"flagged"`
	Categories map[string]bool    `json:"categories"`
	Scores     map[string]float64 `json:"scores"`
	Reason     string             `json:"reason,omitempty"`
}

// ProviderInfo 提供商信息
type ProviderInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Models       []string          `json:"models"`
	Capabilities []string          `json:"capabilities"`
	Limits       map[string]interface{} `json:"limits"`
}

// UsageInfo 使用量信息
type UsageInfo struct {
	TotalTokens   int64     `json:"total_tokens"`
	TotalRequests int64     `json:"total_requests"`
	QuotaUsed     int64     `json:"quota_used"`
	QuotaLimit    int64     `json:"quota_limit"`
	ResetTime     time.Time `json:"reset_time"`
}