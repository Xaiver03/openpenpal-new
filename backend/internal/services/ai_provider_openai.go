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

// OpenAIProvider OpenAI API提供商实现
type OpenAIProvider struct {
	config     *models.AIConfig
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewOpenAIProvider 创建OpenAI提供商实例
func NewOpenAIProvider(config *models.AIConfig) *OpenAIProvider {
	return &OpenAIProvider{
		config:  config,
		baseURL: config.APIEndpoint,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateText 生成文本
func (p *OpenAIProvider) GenerateText(ctx context.Context, prompt string, options AIGenerationOptions) (*AIResponse, error) {
	messages := []ChatMessage{
		{Role: "user", Content: prompt},
	}
	return p.Chat(ctx, messages, options)
}

// Chat 聊天对话
func (p *OpenAIProvider) Chat(ctx context.Context, messages []ChatMessage, options AIGenerationOptions) (*AIResponse, error) {
	reqBody := map[string]interface{}{
		"model":    p.getModel(options.Model),
		"messages": messages,
		"max_tokens": options.MaxTokens,
		"temperature": options.Temperature,
	}

	if options.TopP > 0 {
		reqBody["top_p"] = options.TopP
	}
	if options.FrequencyPenalty != 0 {
		reqBody["frequency_penalty"] = options.FrequencyPenalty
	}
	if options.PresencePenalty != 0 {
		reqBody["presence_penalty"] = options.PresencePenalty
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

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &AIResponse{
		Content:    openaiResp.Choices[0].Message.Content,
		TokensUsed: openaiResp.Usage.TotalTokens,
		Model:      openaiResp.Model,
		Provider:   "openai",
		RequestID:  openaiResp.ID,
		CreatedAt:  time.Now(),
	}, nil
}

// Summarize 文本总结
func (p *OpenAIProvider) Summarize(ctx context.Context, text string, options AIGenerationOptions) (*AIResponse, error) {
	prompt := fmt.Sprintf("请总结以下文本的主要内容：\n\n%s", text)
	return p.GenerateText(ctx, prompt, options)
}

// Translate 内容翻译
func (p *OpenAIProvider) Translate(ctx context.Context, text, targetLang string, options AIGenerationOptions) (*AIResponse, error) {
	prompt := fmt.Sprintf("请将以下文本翻译成%s：\n\n%s", targetLang, text)
	return p.GenerateText(ctx, prompt, options)
}

// AnalyzeSentiment 情感分析
func (p *OpenAIProvider) AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error) {
	prompt := fmt.Sprintf("请分析以下文本的情感倾向，返回JSON格式：{\"sentiment\": \"positive/negative/neutral\", \"score\": 0.0-1.0, \"confidence\": 0.0-1.0}：\n\n%s", text)
	
	options := AIGenerationOptions{
		MaxTokens:   200,
		Temperature: 0.1,
	}
	
	response, err := p.GenerateText(ctx, prompt, options)
	if err != nil {
		return nil, err
	}

	var result SentimentAnalysis
	if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
		// 如果解析失败，返回默认值
		return &SentimentAnalysis{
			Sentiment:  "neutral",
			Score:      0.0,
			Confidence: 0.5,
		}, nil
	}

	return &result, nil
}

// ModerateContent 内容审核
func (p *OpenAIProvider) ModerateContent(ctx context.Context, text string) (*ContentModeration, error) {
	reqBody := map[string]interface{}{
		"input": text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/moderations", bytes.NewBuffer(jsonData))
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

	var moderationResp OpenAIModerationResponse
	if err := json.Unmarshal(body, &moderationResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(moderationResp.Results) == 0 {
		return nil, fmt.Errorf("no results in moderation response")
	}

	result := moderationResp.Results[0]
	return &ContentModeration{
		Flagged:    result.Flagged,
		Categories: result.Categories,
		Scores:     result.CategoryScores,
	}, nil
}

// GetProviderInfo 获取提供商信息
func (p *OpenAIProvider) GetProviderInfo() ProviderInfo {
	return ProviderInfo{
		Name:    "OpenAI",
		Version: "v1",
		Models:  []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"},
		Capabilities: []string{
			"text_generation", "chat", "summarization", 
			"translation", "sentiment_analysis", "moderation",
		},
		Limits: map[string]interface{}{
			"max_tokens": 4096,
			"rate_limit": "3500 requests/minute",
		},
	}
}

// HealthCheck 健康检查
func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetUsage 获取使用量信息
func (p *OpenAIProvider) GetUsage(ctx context.Context) (*UsageInfo, error) {
	// OpenAI API doesn't provide direct usage endpoint
	// Return information based on config
	return &UsageInfo{
		QuotaUsed:  int64(p.config.UsedQuota),
		QuotaLimit: int64(p.config.DailyQuota),
		ResetTime:  time.Now().Add(24 * time.Hour),
	}, nil
}

// getModel 获取模型名称
func (p *OpenAIProvider) getModel(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	if p.config.Model != "" {
		return p.config.Model
	}
	return "gpt-3.5-turbo"
}

// OpenAI API响应结构
type OpenAIResponse struct {
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

type OpenAIModerationResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Results []struct {
		Flagged        bool               `json:"flagged"`
		Categories     map[string]bool    `json:"categories"`
		CategoryScores map[string]float64 `json:"category_scores"`
	} `json:"results"`
}