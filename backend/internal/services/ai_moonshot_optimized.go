package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"openpenpal-backend/internal/logger"
	"openpenpal-backend/internal/models"
	"time"
)

// callMoonshotOptimized 优化版本的Moonshot API调用，减少日志输出
func (s *AIService) callMoonshotOptimized(ctx context.Context, config *models.AIConfig, prompt string) (string, error) {
	log := logger.GetLogger()
	
	// 仅在DEBUG模式下记录详细信息
	log.Debug("🌙 [Moonshot] Starting API call to %s with model %s", config.APIEndpoint, config.Model)

	// Ensure we have the correct endpoint
	if config.APIEndpoint == "" {
		config.APIEndpoint = "https://api.moonshot.cn/v1/chat/completions"
	}

	// Validate API key
	if config.APIKey == "" {
		log.Error("❌ [Moonshot] API key is empty")
		return "", fmt.Errorf("moonshot API key is empty")
	}

	// Build request body - Moonshot API is OpenAI compatible
	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是OpenPenPal的AI助手，在这个温暖的数字书信平台上，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好、富有人文情怀的语气回应。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": config.Temperature,
		"max_tokens":  config.MaxTokens,
		"stream":      false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Error("❌ [Moonshot] Failed to marshal request body: %v", err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 仅在DEBUG模式下记录请求体详情
	log.DebugWithKey("moonshot_request", "🌙 [Moonshot] Request body size: %d bytes", len(jsonData))

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("❌ [Moonshot] Failed to create request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Accept", "application/json")

	// Send request
	startTime := time.Now()
	log.InfoWithKey("moonshot_api", "🚀 [Moonshot] Sending API request")

	resp, err := client.Do(req)
	if err != nil {
		log.Error("❌ [Moonshot] Request failed: %v", err)
		return "", fmt.Errorf("moonshot API request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	
	// 仅在响应时间过长时记录警告
	if duration > 10*time.Second {
		log.Warn("⏱️ [Moonshot] Slow response: %v", duration)
	} else {
		log.DebugWithKey("moonshot_timing", "⏱️ [Moonshot] Request took %v", duration)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("❌ [Moonshot] Failed to read response body: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 仅在DEBUG模式下记录响应详情
	log.DebugWithKey("moonshot_response", "📥 [Moonshot] Response status: %d, size: %d bytes", resp.StatusCode, len(body))

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		log.Error("❌ [Moonshot] API error (status %d): %s", resp.StatusCode, string(body))

		// Try to parse error response
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return "", fmt.Errorf("moonshot API error (status %d): %s", resp.StatusCode, errorResp.Error.Message)
		}

		return "", fmt.Errorf("moonshot API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var result struct {
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

	if err := json.Unmarshal(body, &result); err != nil {
		log.Error("❌ [Moonshot] Failed to parse response JSON: %v", err)
		log.Debug("❌ [Moonshot] Raw response: %s", string(body)) // 仅DEBUG时记录原始响应
		return "", fmt.Errorf("failed to parse moonshot response: %w", err)
	}

	// Validate response structure
	if len(result.Choices) == 0 {
		log.Error("❌ [Moonshot] No choices in response")
		return "", fmt.Errorf("no choices in moonshot response")
	}

	content := result.Choices[0].Message.Content
	if content == "" {
		log.Error("❌ [Moonshot] Empty content in response")
		return "", fmt.Errorf("empty content in moonshot response")
	}

	// 仅记录成功信息和重要统计
	log.InfoWithKey("moonshot_success", "✅ [Moonshot] Success - tokens: %d, chars: %d, duration: %v",
		result.Usage.TotalTokens, len(content), duration)
	
	// 仅在DEBUG模式下记录详细统计
	log.DebugWithKey("moonshot_stats", "✅ [Moonshot] Details - ID: %s, model: %s, prompt=%d, completion=%d",
		result.ID, result.Model, result.Usage.PromptTokens, result.Usage.CompletionTokens)

	return content, nil
}

// callMoonshotSilent 静默版本的Moonshot API调用，仅记录错误
func (s *AIService) callMoonshotSilent(ctx context.Context, config *models.AIConfig, prompt string) (string, error) {
	// 基本验证，无日志
	if config.APIEndpoint == "" {
		config.APIEndpoint = "https://api.moonshot.cn/v1/chat/completions"
	}
	if config.APIKey == "" {
		return "", fmt.Errorf("moonshot API key is empty")
	}

	// 构建请求
	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是OpenPenPal的AI助手，在这个温暖的数字书信平台上，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好、富有人文情怀的语气回应。",
			},
			{
				"role":    "user", 
				"content": prompt,
			},
		},
		"temperature": config.Temperature,
		"max_tokens":  config.MaxTokens,
		"stream":      false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("❌ [Moonshot] Request failed: %v", err)
		return "", fmt.Errorf("moonshot API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("❌ [Moonshot] API error (status %d)", resp.StatusCode)
		return "", fmt.Errorf("moonshot API error (status %d)", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("❌ [Moonshot] Failed to parse response")
		return "", fmt.Errorf("failed to parse moonshot response: %w", err)
	}

	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		logger.Error("❌ [Moonshot] Empty response")
		return "", fmt.Errorf("empty moonshot response")
	}

	return result.Choices[0].Message.Content, nil
}