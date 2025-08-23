package services

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"openpenpal-backend/internal/models"
)

// LocalProvider 本地开发环境AI提供商实现
type LocalProvider struct {
	config *models.AIConfig
}

// NewLocalProvider 创建本地提供商实例
func NewLocalProvider(config *models.AIConfig) *LocalProvider {
	return &LocalProvider{
		config: config,
	}
}

// GenerateText 生成文本（模拟）
func (p *LocalProvider) GenerateText(ctx context.Context, prompt string, options AIGenerationOptions) (*AIResponse, error) {
	// 模拟API延迟
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	
	content := p.generateMockContent(prompt)
	tokensUsed := len(strings.Fields(content)) + len(strings.Fields(prompt))
	
	return &AIResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Model:      "local-mock-model",
		Provider:   "local",
		RequestID:  fmt.Sprintf("local-%d", time.Now().UnixNano()),
		Metadata: map[string]interface{}{
			"simulated": true,
			"prompt_length": len(prompt),
		},
		CreatedAt: time.Now(),
	}, nil
}

// Chat 聊天对话（模拟）
func (p *LocalProvider) Chat(ctx context.Context, messages []ChatMessage, options AIGenerationOptions) (*AIResponse, error) {
	// 模拟延迟
	time.Sleep(time.Duration(rand.Intn(800)) * time.Millisecond)
	
	lastMessage := ""
	if len(messages) > 0 {
		lastMessage = messages[len(messages)-1].Content
	}
	
	content := p.generateMockResponse(lastMessage)
	tokensUsed := p.calculateTokens(messages) + len(strings.Fields(content))
	
	return &AIResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Model:      "local-chat-model",
		Provider:   "local",
		RequestID:  fmt.Sprintf("local-chat-%d", time.Now().UnixNano()),
		Metadata: map[string]interface{}{
			"simulated": true,
			"message_count": len(messages),
		},
		CreatedAt: time.Now(),
	}, nil
}

// Summarize 文本总结（模拟）
func (p *LocalProvider) Summarize(ctx context.Context, text string, options AIGenerationOptions) (*AIResponse, error) {
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	
	// 简单的总结模拟
	sentences := strings.Split(text, "。")
	summary := ""
	
	if len(sentences) > 3 {
		summary = fmt.Sprintf("本文主要讨论了相关主题，包含%d个要点。核心内容涉及文本中提到的关键概念和观点。", len(sentences))
	} else {
		summary = "这是一段简短的文本，主要表达了作者的观点和想法。"
	}
	
	return &AIResponse{
		Content:    summary,
		TokensUsed: len(strings.Fields(text)) + len(strings.Fields(summary)),
		Model:      "local-summary-model",
		Provider:   "local",
		RequestID:  fmt.Sprintf("local-summary-%d", time.Now().UnixNano()),
		CreatedAt:  time.Now(),
	}, nil
}

// Translate 内容翻译（模拟）
func (p *LocalProvider) Translate(ctx context.Context, text, targetLang string, options AIGenerationOptions) (*AIResponse, error) {
	time.Sleep(time.Duration(rand.Intn(400)) * time.Millisecond)
	
	var translated string
	
	switch targetLang {
	case "English", "英语", "english":
		translated = fmt.Sprintf("[Translated to English] %s", text)
	case "中文", "Chinese", "chinese":
		translated = fmt.Sprintf("【翻译为中文】%s", text)
	case "日语", "Japanese", "japanese":
		translated = fmt.Sprintf("【日本語翻訳】%s", text)
	default:
		translated = fmt.Sprintf("[Translated to %s] %s", targetLang, text)
	}
	
	return &AIResponse{
		Content:    translated,
		TokensUsed: len(strings.Fields(text)) + len(strings.Fields(translated)),
		Model:      "local-translate-model",
		Provider:   "local",
		RequestID:  fmt.Sprintf("local-translate-%d", time.Now().UnixNano()),
		CreatedAt:  time.Now(),
	}, nil
}

// AnalyzeSentiment 情感分析（模拟）
func (p *LocalProvider) AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error) {
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	
	// 简单的情感分析模拟
	positiveWords := []string{"好", "棒", "优秀", "喜欢", "爱", "开心", "happy", "good", "great", "love", "like"}
	negativeWords := []string{"坏", "差", "讨厌", "恨", "难过", "sad", "bad", "hate", "terrible", "awful"}
	
	positiveCount := 0
	negativeCount := 0
	
	lowerText := strings.ToLower(text)
	
	for _, word := range positiveWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			positiveCount++
		}
	}
	
	for _, word := range negativeWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			negativeCount++
		}
	}
	
	sentiment := "neutral"
	score := 0.0
	confidence := 0.7
	
	if positiveCount > negativeCount {
		sentiment = "positive"
		score = 0.6 + (float64(positiveCount-negativeCount) * 0.1)
		if score > 1.0 {
			score = 1.0
		}
	} else if negativeCount > positiveCount {
		sentiment = "negative"
		score = -0.6 - (float64(negativeCount-positiveCount) * 0.1)
		if score < -1.0 {
			score = -1.0
		}
	}
	
	return &SentimentAnalysis{
		Sentiment:  sentiment,
		Score:      score,
		Confidence: confidence,
		Details: map[string]interface{}{
			"positive_words": positiveCount,
			"negative_words": negativeCount,
			"method":         "keyword_analysis",
			"simulated":      true,
		},
	}, nil
}

// ModerateContent 内容审核（模拟）
func (p *LocalProvider) ModerateContent(ctx context.Context, text string) (*ContentModeration, error) {
	time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)
	
	// 简单的内容审核模拟
	flaggedWords := []string{"spam", "advertisement", "垃圾", "广告", "暴力", "violence", "hate", "仇恨"}
	
	flagged := false
	categories := make(map[string]bool)
	scores := make(map[string]float64)
	reason := ""
	
	lowerText := strings.ToLower(text)
	
	for _, word := range flaggedWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			flagged = true
			
			switch word {
			case "spam", "advertisement", "垃圾", "广告":
				categories["spam"] = true
				scores["spam"] = 0.8
				reason = "Detected potential spam content"
			case "暴力", "violence":
				categories["violence"] = true
				scores["violence"] = 0.7
				reason = "Detected potential violent content"
			case "hate", "仇恨":
				categories["hate"] = true
				scores["hate"] = 0.6
				reason = "Detected potential hate speech"
			}
		}
	}
	
	// 设置默认分数
	for _, category := range []string{"hate", "harassment", "violence", "sexual", "spam"} {
		if _, exists := scores[category]; !exists {
			scores[category] = 0.1
		}
		if _, exists := categories[category]; !exists {
			categories[category] = false
		}
	}
	
	return &ContentModeration{
		Flagged:    flagged,
		Categories: categories,
		Scores:     scores,
		Reason:     reason,
	}, nil
}

// GetProviderInfo 获取提供商信息
func (p *LocalProvider) GetProviderInfo() ProviderInfo {
	return ProviderInfo{
		Name:    "Local Development Provider",
		Version: "1.0.0",
		Models:  []string{"local-mock-model", "local-chat-model", "local-summary-model"},
		Capabilities: []string{
			"text_generation", "chat", "summarization",
			"translation", "sentiment_analysis", "moderation",
			"development_mode", "offline_capable",
		},
		Limits: map[string]interface{}{
			"max_tokens":      "unlimited",
			"rate_limit":      "unlimited",
			"cost":           "free",
			"response_time":   "100-800ms",
		},
	}
}

// HealthCheck 健康检查
func (p *LocalProvider) HealthCheck(ctx context.Context) error {
	// 本地提供商总是健康的
	return nil
}

// GetUsage 获取使用量信息
func (p *LocalProvider) GetUsage(ctx context.Context) (*UsageInfo, error) {
	return &UsageInfo{
		TotalTokens:   int64(rand.Intn(10000)),
		TotalRequests: int64(rand.Intn(1000)),
		QuotaUsed:     int64(p.config.UsedQuota),
		QuotaLimit:    int64(p.config.DailyQuota),
		ResetTime:     time.Now().Add(24 * time.Hour),
	}, nil
}

// generateMockContent 生成模拟内容
func (p *LocalProvider) generateMockContent(prompt string) string {
	templates := []string{
		"这是对您问题的模拟回答。在实际部署中，这里会调用真实的AI服务。您的问题：%s",
		"基于您的输入，我提供以下建议：%s。这是本地开发环境的模拟响应。",
		"针对您提到的内容，我认为可以从以下几个方面考虑：%s。（注：这是开发环境模拟）",
		"您好！关于您的询问：%s，我建议您可以尝试以下方法。这是本地AI模拟服务的回复。",
	}
	
	template := templates[rand.Intn(len(templates))]
	
	// 截取prompt的前50个字符避免响应过长
	shortPrompt := prompt
	if len(prompt) > 50 {
		shortPrompt = prompt[:50] + "..."
	}
	
	return fmt.Sprintf(template, shortPrompt)
}

// generateMockResponse 生成模拟对话响应
func (p *LocalProvider) generateMockResponse(message string) string {
	responses := []string{
		"我理解您的观点，这确实是一个值得深入思考的问题。",
		"根据您提供的信息，我建议您可以从以下角度考虑这个问题。",
		"这是一个很有意思的话题，让我们一起探讨一下。",
		"感谢您的分享，我觉得您的想法很有启发性。",
		"关于这个问题，我认为可以从多个维度来分析。",
		"您提到的这个点很重要，确实值得我们进一步讨论。",
	}
	
	response := responses[rand.Intn(len(responses))]
	
	// 如果消息包含特定关键词，给出更具体的回应
	if strings.Contains(message, "写信") || strings.Contains(message, "letter") {
		response += " 关于信件写作，我建议您可以从表达真诚情感开始，用心传达您想要分享的内容。"
	} else if strings.Contains(message, "朋友") || strings.Contains(message, "friend") {
		response += " 友谊是珍贵的，真诚和理解是维系友谊的重要因素。"
	}
	
	response += "（本回复来自本地开发环境AI模拟服务）"
	
	return response
}

// calculateTokens 计算消息的近似token数
func (p *LocalProvider) calculateTokens(messages []ChatMessage) int {
	total := 0
	for _, msg := range messages {
		total += len(strings.Fields(msg.Content))
	}
	return total
}

// Additional providers can be added here...

// NewGeminiProvider 创建Gemini提供商实例  
func NewGeminiProvider(config *models.AIConfig) *LocalProvider {
	// Gemini可以实现Google API，这里简化为本地提供商
	return NewLocalProvider(config)
}