package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"openpenpal-backend/internal/models"
)

// MoonshotProvider Moonshot AI提供商实现
type MoonshotProvider struct {
	config     *models.AIConfig
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewMoonshotProvider 创建Moonshot提供商实例
func NewMoonshotProvider(config *models.AIConfig) *MoonshotProvider {
	return &MoonshotProvider{
		config:  config,
		baseURL: config.APIEndpoint,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateText 生成文本
func (p *MoonshotProvider) GenerateText(ctx context.Context, prompt string, options AIGenerationOptions) (*AIResponse, error) {
	messages := []ChatMessage{
		{Role: "user", Content: prompt},
	}
	return p.Chat(ctx, messages, options)
}

// Chat 聊天对话
func (p *MoonshotProvider) Chat(ctx context.Context, messages []ChatMessage, options AIGenerationOptions) (*AIResponse, error) {
	reqBody := map[string]interface{}{
		"model":       p.getModel(options.Model),
		"messages":    messages,
		"max_tokens":  options.MaxTokens,
		"temperature": options.Temperature,
	}

	if options.TopP > 0 {
		reqBody["top_p"] = options.TopP
	}
	if len(options.Stop) > 0 {
		reqBody["stop"] = options.Stop
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var moonshotResp MoonshotResponse
	if err := json.Unmarshal(body, &moonshotResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(moonshotResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &AIResponse{
		Content:    moonshotResp.Choices[0].Message.Content,
		TokensUsed: moonshotResp.Usage.TotalTokens,
		Model:      moonshotResp.Model,
		Provider:   "moonshot",
		RequestID:  moonshotResp.ID,
		CreatedAt:  time.Now(),
	}, nil
}

// Summarize 文本总结
func (p *MoonshotProvider) Summarize(ctx context.Context, text string, options AIGenerationOptions) (*AIResponse, error) {
	prompt := fmt.Sprintf("请简洁地总结以下文本的核心内容：\n\n%s\n\n总结：", text)
	return p.GenerateText(ctx, prompt, options)
}

// Translate 内容翻译
func (p *MoonshotProvider) Translate(ctx context.Context, text, targetLang string, options AIGenerationOptions) (*AIResponse, error) {
	prompt := fmt.Sprintf("请将以下内容翻译成%s，保持原文的语调和风格：\n\n%s", targetLang, text)
	return p.GenerateText(ctx, prompt, options)
}

// AnalyzeSentiment 情感分析
func (p *MoonshotProvider) AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error) {
	prompt := fmt.Sprintf(`请分析以下文本的情感倾向，用JSON格式回答：
{
  "sentiment": "positive/negative/neutral",
  "score": 数值范围-1.0到1.0(负数表示消极，正数表示积极),
  "confidence": 置信度0.0到1.0
}

文本：%s`, text)
	
	options := AIGenerationOptions{
		MaxTokens:   150,
		Temperature: 0.1,
	}
	
	response, err := p.GenerateText(ctx, prompt, options)
	if err != nil {
		return nil, err
	}

	var result SentimentAnalysis
	if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
		// 如果JSON解析失败，使用文本解析
		return p.parseSentimentFromText(response.Content), nil
	}

	return &result, nil
}

// ModerateContent 内容审核
func (p *MoonshotProvider) ModerateContent(ctx context.Context, text string) (*ContentModeration, error) {
	prompt := fmt.Sprintf(`请分析以下文本是否包含不当内容，用JSON格式回答：
{
  "flagged": true/false,
  "categories": {
    "hate": true/false,
    "harassment": true/false,
    "violence": true/false,
    "sexual": true/false,
    "spam": true/false
  },
  "scores": {
    "hate": 0.0-1.0,
    "harassment": 0.0-1.0,
    "violence": 0.0-1.0,
    "sexual": 0.0-1.0,
    "spam": 0.0-1.0
  },
  "reason": "如果被标记，说明原因"
}

文本：%s`, text)

	options := AIGenerationOptions{
		MaxTokens:   300,
		Temperature: 0.1,
	}

	response, err := p.GenerateText(ctx, prompt, options)
	if err != nil {
		return nil, err
	}

	var result ContentModeration
	if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
		// 如果解析失败，返回安全的默认值
		return &ContentModeration{
			Flagged:    false,
			Categories: make(map[string]bool),
			Scores:     make(map[string]float64),
		}, nil
	}

	return &result, nil
}

// GetProviderInfo 获取提供商信息
func (p *MoonshotProvider) GetProviderInfo() ProviderInfo {
	return ProviderInfo{
		Name:    "Moonshot AI",
		Version: "v1",
		Models:  []string{"moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k"},
		Capabilities: []string{
			"text_generation", "chat", "summarization", 
			"translation", "sentiment_analysis", "chinese_optimized",
		},
		Limits: map[string]interface{}{
			"max_tokens": 8192,
			"rate_limit": "1000 requests/minute",
		},
	}
}

// HealthCheck 健康检查
func (p *MoonshotProvider) HealthCheck(ctx context.Context) error {
	// 发送一个简单的请求来检查API状态
	messages := []ChatMessage{
		{Role: "user", Content: "Hello"},
	}
	
	options := AIGenerationOptions{
		MaxTokens:   10,
		Temperature: 0.1,
	}
	
	_, err := p.Chat(ctx, messages, options)
	return err
}

// GetUsage 获取使用量信息
func (p *MoonshotProvider) GetUsage(ctx context.Context) (*UsageInfo, error) {
	return &UsageInfo{
		QuotaUsed:  int64(p.config.UsedQuota),
		QuotaLimit: int64(p.config.DailyQuota),
		ResetTime:  time.Now().Add(24 * time.Hour),
	}, nil
}

// getModel 获取模型名称
func (p *MoonshotProvider) getModel(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	if p.config.Model != "" {
		return p.config.Model
	}
	return "moonshot-v1-8k"
}

// parseSentimentFromText 从文本中解析情感分析结果
func (p *MoonshotProvider) parseSentimentFromText(text string) *SentimentAnalysis {
	// 简单的文本解析逻辑
	sentiment := "neutral"
	score := 0.0
	confidence := 0.5

	if contains(text, []string{"positive", "积极", "正面"}) {
		sentiment = "positive"
		score = 0.7
		confidence = 0.8
	} else if contains(text, []string{"negative", "消极", "负面"}) {
		sentiment = "negative"
		score = -0.7
		confidence = 0.8
	}

	return &SentimentAnalysis{
		Sentiment:  sentiment,
		Score:      score,
		Confidence: confidence,
	}
}

// contains 检查文本是否包含任何指定的关键词
func contains(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(text) > 0 && len(keyword) > 0 {
			// 简单的包含检查
			for i := 0; i <= len(text)-len(keyword); i++ {
				if text[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// Moonshot API响应结构
type MoonshotResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}