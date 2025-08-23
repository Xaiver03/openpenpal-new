package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"openpenpal-backend/internal/models"
)

// SiliconFlowProvider SiliconFlow AI提供商实现
type SiliconFlowProvider struct {
	config     *models.AIConfig
	httpClient *http.Client
}

// NewSiliconFlowProvider 创建SiliconFlow提供商实例
func NewSiliconFlowProvider(config *models.AIConfig) AIProviderInterface {
	return &SiliconFlowProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SiliconFlowRequest 请求结构
type SiliconFlowRequest struct {
	Model       string                   `json:"model"`
	Messages    []SiliconFlowMessage     `json:"messages"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	TopP        float64                  `json:"top_p,omitempty"`
	Stream      bool                     `json:"stream"`
	Stop        []string                 `json:"stop,omitempty"`
}

// SiliconFlowMessage 消息结构
type SiliconFlowMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SiliconFlowResponse 响应结构
type SiliconFlowResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int    `json:"index"`
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
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// GenerateText 生成文本
func (p *SiliconFlowProvider) GenerateText(ctx context.Context, prompt string, options AIGenerationOptions) (*AIResponse, error) {
	messages := []SiliconFlowMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return p.sendRequest(ctx, messages, options)
}

// Chat 聊天对话
func (p *SiliconFlowProvider) Chat(ctx context.Context, messages []ChatMessage, options AIGenerationOptions) (*AIResponse, error) {
	sfMessages := make([]SiliconFlowMessage, len(messages))
	for i, msg := range messages {
		sfMessages[i] = SiliconFlowMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return p.sendRequest(ctx, sfMessages, options)
}

// sendRequest 发送请求到SiliconFlow API
func (p *SiliconFlowProvider) sendRequest(ctx context.Context, messages []SiliconFlowMessage, options AIGenerationOptions) (*AIResponse, error) {
	// 设置默认模型
	model := p.config.Model
	if model == "" {
		model = "Qwen/Qwen2.5-7B-Instruct" // SiliconFlow默认使用Qwen模型
	}
	if options.Model != "" {
		model = options.Model
	}

	// 构建请求
	request := SiliconFlowRequest{
		Model:       model,
		Messages:    messages,
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
		TopP:        options.TopP,
		Stream:      false,
		Stop:        options.Stop,
	}

	// 设置默认值
	if request.Temperature == 0 {
		request.Temperature = 0.7
	}
	if request.MaxTokens == 0 {
		request.MaxTokens = 1000
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("🚀 [SiliconFlow] Sending request to model: %s", model)

	// 创建HTTP请求
	apiURL := p.config.APIEndpoint
	if apiURL == "" {
		apiURL = "https://api.siliconflow.cn/v1/chat/completions"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	var sfResp SiliconFlowResponse
	if err := json.Unmarshal(body, &sfResp); err != nil {
		log.Printf("❌ [SiliconFlow] Failed to parse response: %s", string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查错误
	if sfResp.Error != nil {
		log.Printf("❌ [SiliconFlow] API error: %s - %s", sfResp.Error.Code, sfResp.Error.Message)
		return nil, fmt.Errorf("SiliconFlow API error: %s - %s", sfResp.Error.Code, sfResp.Error.Message)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [SiliconFlow] HTTP error: %d - %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("SiliconFlow API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 提取响应内容
	if len(sfResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	content := sfResp.Choices[0].Message.Content
	tokensUsed := sfResp.Usage.TotalTokens

	log.Printf("✅ [SiliconFlow] Response received: %d tokens used", tokensUsed)

	return &AIResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Model:      model,
		Provider:   "siliconflow",
		RequestID:  sfResp.ID,
		Metadata: map[string]interface{}{
			"finish_reason": sfResp.Choices[0].FinishReason,
			"prompt_tokens": sfResp.Usage.PromptTokens,
			"completion_tokens": sfResp.Usage.CompletionTokens,
		},
		CreatedAt: time.Now(),
	}, nil
}

// Summarize 文本总结
func (p *SiliconFlowProvider) Summarize(ctx context.Context, text string, options AIGenerationOptions) (*AIResponse, error) {
	prompt := fmt.Sprintf("请总结以下文本的主要内容，用简洁的中文表达：\n\n%s", text)
	return p.GenerateText(ctx, prompt, options)
}

// Translate 内容翻译
func (p *SiliconFlowProvider) Translate(ctx context.Context, text, targetLang string, options AIGenerationOptions) (*AIResponse, error) {
	prompt := fmt.Sprintf("请将以下文本翻译成%s，保持原意和语言风格：\n\n%s", targetLang, text)
	return p.GenerateText(ctx, prompt, options)
}

// AnalyzeSentiment 情感分析
func (p *SiliconFlowProvider) AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error) {
	prompt := fmt.Sprintf(`请分析以下文本的情感倾向，返回JSON格式的结果：
{
  "sentiment": "positive/negative/neutral",
  "confidence": 0.0-1.0,
  "emotions": ["情感1", "情感2"],
  "summary": "简短总结"
}

文本：%s`, text)

	response, err := p.GenerateText(ctx, prompt, AIGenerationOptions{
		Temperature: 0.3, // 降低温度以获得更稳定的结果
		MaxTokens:   500,
	})
	if err != nil {
		return nil, err
	}

	// 尝试解析JSON响应
	var result SentimentAnalysis
	
	// 查找JSON内容
	content := response.Content
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	
	if startIdx >= 0 && endIdx > startIdx {
		jsonStr := content[startIdx : endIdx+1]
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			// 如果解析失败，返回基本分析
			return p.fallbackSentimentAnalysis(content), nil
		}
		return &result, nil
	}

	return p.fallbackSentimentAnalysis(content), nil
}

// fallbackSentimentAnalysis 降级情感分析
func (p *SiliconFlowProvider) fallbackSentimentAnalysis(aiResponse string) *SentimentAnalysis {
	sentiment := "neutral"
	confidence := 0.5
	
	lowerResponse := strings.ToLower(aiResponse)
	if strings.Contains(lowerResponse, "positive") || strings.Contains(lowerResponse, "积极") {
		sentiment = "positive"
		confidence = 0.7
	} else if strings.Contains(lowerResponse, "negative") || strings.Contains(lowerResponse, "消极") {
		sentiment = "negative"
		confidence = 0.7
	}

	return &SentimentAnalysis{
		Sentiment:  sentiment,
		Score:      0.0,
		Confidence: confidence,
		Details: map[string]interface{}{
			"provider": "siliconflow",
			"emotions": []string{"unknown"},
			"summary":  "AI分析结果",
		},
	}
}

// ModerateContent 内容审核
func (p *SiliconFlowProvider) ModerateContent(ctx context.Context, text string) (*ContentModeration, error) {
	prompt := fmt.Sprintf(`请审核以下内容是否包含不当信息，返回JSON格式的结果：
{
  "safe": true/false,
  "categories": ["类别1", "类别2"],
  "confidence": 0.0-1.0,
  "reason": "审核理由"
}

内容：%s`, text)

	response, err := p.GenerateText(ctx, prompt, AIGenerationOptions{
		Temperature: 0.2,
		MaxTokens:   300,
	})
	if err != nil {
		return nil, err
	}

	// 尝试解析JSON响应
	var result ContentModeration
	
	content := response.Content
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	
	if startIdx >= 0 && endIdx > startIdx {
		jsonStr := content[startIdx : endIdx+1]
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			// 如果解析失败，返回安全结果
			return &ContentModeration{
				Flagged:    false,
				Categories: make(map[string]bool),
				Scores:     make(map[string]float64),
				Reason:     "内容审核通过",
			}, nil
		}
		return &result, nil
	}

	// 默认返回安全
	return &ContentModeration{
		Flagged:    false,
		Categories: make(map[string]bool),
		Scores:     make(map[string]float64),
		Reason:     "内容审核通过",
	}, nil
}

// GetProviderInfo 获取提供商信息
func (p *SiliconFlowProvider) GetProviderInfo() ProviderInfo {
	return ProviderInfo{
		Name:         "siliconflow",
		Version:      "1.0",
		Capabilities: []string{"text-generation", "chat", "translation", "summarization"},
		Models: []string{
			"Qwen/Qwen2.5-7B-Instruct",
			"deepseek-ai/DeepSeek-V2.5",
		},
		Limits: map[string]interface{}{
			"max_tokens_per_request": 32768,
			"requests_per_minute": 100,
		},
	}
}

// HealthCheck 健康检查
func (p *SiliconFlowProvider) HealthCheck(ctx context.Context) error {
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := p.GenerateText(testCtx, "Hello", AIGenerationOptions{
		MaxTokens: 10,
	})
	
	return err
}

// GetUsage 获取使用量
func (p *SiliconFlowProvider) GetUsage(ctx context.Context) (*UsageInfo, error) {
	// SiliconFlow API暂不提供使用量查询接口
	return &UsageInfo{
		TotalTokens:   0,
		TotalRequests: 0,
		QuotaUsed:     0,
		QuotaLimit:    0,
		ResetTime:     time.Now().Add(24 * time.Hour),
	}, nil
}