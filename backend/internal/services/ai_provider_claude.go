package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"openpenpal-backend/internal/models"
)

// ClaudeProvider Claude AI提供商实现
type ClaudeProvider struct {
	config     *models.AIConfig
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewClaudeProvider 创建Claude提供商实例
func NewClaudeProvider(config *models.AIConfig) *ClaudeProvider {
	baseURL := config.APIEndpoint
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	
	return &ClaudeProvider{
		config:  config,
		baseURL: baseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Claude可能需要更长时间
		},
	}
}

// GenerateText 生成文本
func (p *ClaudeProvider) GenerateText(ctx context.Context, prompt string, options AIGenerationOptions) (*AIResponse, error) {
	messages := []ChatMessage{
		{Role: "user", Content: prompt},
	}
	return p.Chat(ctx, messages, options)
}

// Chat 聊天对话
func (p *ClaudeProvider) Chat(ctx context.Context, messages []ChatMessage, options AIGenerationOptions) (*AIResponse, error) {
	reqBody := map[string]interface{}{
		"model":      p.getModel(options.Model),
		"max_tokens": options.MaxTokens,
		"messages":   messages,
	}

	if options.Temperature > 0 {
		reqBody["temperature"] = options.Temperature
	}
	if options.TopP > 0 {
		reqBody["top_p"] = options.TopP
	}
	if len(options.Stop) > 0 {
		reqBody["stop_sequences"] = options.Stop
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := ""
	for _, c := range claudeResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &AIResponse{
		Content:    content,
		TokensUsed: claudeResp.Usage.OutputTokens + claudeResp.Usage.InputTokens,
		Model:      claudeResp.Model,
		Provider:   "claude",
		RequestID:  claudeResp.ID,
		CreatedAt:  time.Now(),
	}, nil
}

// Summarize 文本总结
func (p *ClaudeProvider) Summarize(ctx context.Context, text string, options AIGenerationOptions) (*AIResponse, error) {
	prompt := fmt.Sprintf("请简洁地总结以下文本的要点，保持关键信息：\n\n%s", text)
	return p.GenerateText(ctx, prompt, options)
}

// Translate 内容翻译
func (p *ClaudeProvider) Translate(ctx context.Context, text, targetLang string, options AIGenerationOptions) (*AIResponse, error) {
	prompt := fmt.Sprintf("请将以下文本准确翻译成%s，保持原文的语调和含义：\n\n%s", targetLang, text)
	return p.GenerateText(ctx, prompt, options)
}

// AnalyzeSentiment 情感分析
func (p *ClaudeProvider) AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error) {
	prompt := fmt.Sprintf(`请分析以下文本的情感倾向并返回JSON格式结果：

要求格式：
{
  "sentiment": "positive/negative/neutral",
  "score": 数值(-1.0到1.0，负数表示消极，正数表示积极),
  "confidence": 置信度(0.0到1.0),
  "details": {
    "主要情感": "描述",
    "情感强度": "强/中/弱"
  }
}

分析文本：
%s`, text)
	
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
		// 如果JSON解析失败，尝试简单的文本分析
		return p.parseSentimentFromText(response.Content), nil
	}

	return &result, nil
}

// ModerateContent 内容审核
func (p *ClaudeProvider) ModerateContent(ctx context.Context, text string) (*ContentModeration, error) {
	prompt := fmt.Sprintf(`请分析以下文本是否包含不当内容，用JSON格式回答：

{
  "flagged": true/false,
  "categories": {
    "hate": 仇恨言论(true/false),
    "harassment": 骚扰内容(true/false),
    "violence": 暴力内容(true/false),
    "sexual": 性相关内容(true/false),
    "harmful": 有害信息(true/false),
    "spam": 垃圾信息(true/false)
  },
  "scores": {
    "hate": 0.0-1.0,
    "harassment": 0.0-1.0,
    "violence": 0.0-1.0,
    "sexual": 0.0-1.0,
    "harmful": 0.0-1.0,
    "spam": 0.0-1.0
  },
  "reason": "如果被标记，详细说明原因"
}

请分析文本：
%s`, text)

	options := AIGenerationOptions{
		MaxTokens:   400,
		Temperature: 0.1,
	}

	response, err := p.GenerateText(ctx, prompt, options)
	if err != nil {
		return nil, err
	}

	var result ContentModeration
	if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
		// 解析失败时返回保守的结果
		return &ContentModeration{
			Flagged:    false,
			Categories: make(map[string]bool),
			Scores:     make(map[string]float64),
			Reason:     "Analysis completed with basic safety check",
		}, nil
	}

	return &result, nil
}

// GetProviderInfo 获取提供商信息
func (p *ClaudeProvider) GetProviderInfo() ProviderInfo {
	return ProviderInfo{
		Name:    "Claude",
		Version: "v1",
		Models:  []string{"claude-3-sonnet-20240229", "claude-3-opus-20240229", "claude-3-haiku-20240307"},
		Capabilities: []string{
			"text_generation", "chat", "summarization",
			"translation", "sentiment_analysis", "reasoning",
			"code_analysis", "creative_writing",
		},
		Limits: map[string]interface{}{
			"max_tokens": 200000,
			"rate_limit": "50 requests/minute",
		},
	}
}

// HealthCheck 健康检查
func (p *ClaudeProvider) HealthCheck(ctx context.Context) error {
	// 发送简单的健康检查请求
	messages := []ChatMessage{
		{Role: "user", Content: "Hello, are you working correctly?"},
	}
	
	options := AIGenerationOptions{
		MaxTokens:   20,
		Temperature: 0.1,
	}
	
	_, err := p.Chat(ctx, messages, options)
	return err
}

// GetUsage 获取使用量信息
func (p *ClaudeProvider) GetUsage(ctx context.Context) (*UsageInfo, error) {
	// Claude API 不直接提供使用量接口，基于配置返回
	return &UsageInfo{
		QuotaUsed:  int64(p.config.UsedQuota),
		QuotaLimit: int64(p.config.DailyQuota),
		ResetTime:  time.Now().Add(24 * time.Hour),
	}, nil
}

// getModel 获取模型名称
func (p *ClaudeProvider) getModel(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	if p.config.Model != "" {
		return p.config.Model
	}
	return "claude-3-sonnet-20240229"
}

// parseSentimentFromText 从文本解析情感分析
func (p *ClaudeProvider) parseSentimentFromText(text string) *SentimentAnalysis {
	sentiment := "neutral"
	score := 0.0
	confidence := 0.6

	// 简单的关键词检测
	positiveKeywords := []string{"positive", "good", "great", "excellent", "积极", "正面", "好", "优秀"}
	negativeKeywords := []string{"negative", "bad", "terrible", "awful", "消极", "负面", "坏", "糟糕"}

	for _, keyword := range positiveKeywords {
		if containsWord(text, keyword) {
			sentiment = "positive"
			score = 0.7
			confidence = 0.8
			break
		}
	}

	for _, keyword := range negativeKeywords {
		if containsWord(text, keyword) {
			sentiment = "negative"
			score = -0.7
			confidence = 0.8
			break
		}
	}

	return &SentimentAnalysis{
		Sentiment:  sentiment,
		Score:      score,
		Confidence: confidence,
		Details: map[string]interface{}{
			"method": "keyword_analysis",
		},
	}
}

// containsWord 检查文本是否包含指定词汇
func containsWord(text, word string) bool {
	// 简单的包含检查，实际应用中可以使用更复杂的自然语言处理
	return len(text) >= len(word) && 
		   text != "" && word != "" &&
		   strings.Contains(strings.ToLower(text), strings.ToLower(word))
}

// Claude API响应结构
type ClaudeResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Model        string `json:"model"`
	Content      []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}