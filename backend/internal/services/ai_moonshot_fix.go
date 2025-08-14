package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"openpenpal-backend/internal/models"
	"time"
)

// callMoonshotFixed is a fixed version of the Moonshot API call
func (s *AIService) callMoonshotFixed(ctx context.Context, config *models.AIConfig, prompt string) (string, error) {
	log.Printf("🌙 [Moonshot] Starting fixed API call...")
	log.Printf("🌙 [Moonshot] API Endpoint: %s", config.APIEndpoint)
	log.Printf("🌙 [Moonshot] Model: %s", config.Model)

	// Ensure we have the correct endpoint
	if config.APIEndpoint == "" {
		config.APIEndpoint = "https://api.moonshot.cn/v1/chat/completions"
	}

	// Validate API key
	if config.APIKey == "" {
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
		"stream":      false, // Explicitly disable streaming
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("❌ [Moonshot] Failed to marshal request body: %v", err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("🌙 [Moonshot] Request body size: %d bytes", len(jsonData))
	log.Printf("🌙 [Moonshot] Request body: %s", string(jsonData))

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ [Moonshot] Failed to create request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Accept", "application/json")

	// 安全日志：不记录任何API密钥信息
	if len(config.APIKey) > 0 {
		log.Printf("🔑 [Moonshot] API Key configured")
	} else {
		log.Printf("⚠️ [Moonshot] No API Key configured")
	}

	// Send request
	log.Printf("🚀 [Moonshot] Sending request to %s", config.APIEndpoint)
	startTime := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ [Moonshot] Request failed: %v", err)
		return "", fmt.Errorf("moonshot API request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	log.Printf("⏱️ [Moonshot] Request took %v", duration)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [Moonshot] Failed to read response body: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("📥 [Moonshot] Response status: %d", resp.StatusCode)
	log.Printf("📥 [Moonshot] Response headers: %v", resp.Header)
	log.Printf("📥 [Moonshot] Response body size: %d bytes", len(body))

	// Log first 500 chars of response for debugging
	if len(body) > 0 {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		log.Printf("📥 [Moonshot] Response preview: %s", preview)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [Moonshot] API error response: %s", string(body))

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
		log.Printf("❌ [Moonshot] Failed to parse response JSON: %v", err)
		log.Printf("❌ [Moonshot] Raw response: %s", string(body))
		return "", fmt.Errorf("failed to parse moonshot response: %w", err)
	}

	// Validate response structure
	if len(result.Choices) == 0 {
		log.Printf("❌ [Moonshot] No choices in response")
		return "", fmt.Errorf("no choices in moonshot response")
	}

	content := result.Choices[0].Message.Content
	if content == "" {
		log.Printf("❌ [Moonshot] Empty content in response")
		return "", fmt.Errorf("empty content in moonshot response")
	}

	log.Printf("✅ [Moonshot] Successfully received response")
	log.Printf("✅ [Moonshot] Response ID: %s", result.ID)
	log.Printf("✅ [Moonshot] Model used: %s", result.Model)
	log.Printf("✅ [Moonshot] Tokens used: prompt=%d, completion=%d, total=%d",
		result.Usage.PromptTokens, result.Usage.CompletionTokens, result.Usage.TotalTokens)
	log.Printf("✅ [Moonshot] Content length: %d characters", len(content))

	return content, nil
}
