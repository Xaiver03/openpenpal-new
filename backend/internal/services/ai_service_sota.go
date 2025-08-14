package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AIServiceMetrics tracks performance metrics
type AIServiceMetrics struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	FallbackCount   int64
	AvgResponseTime float64
	LastError       string
	LastErrorTime   time.Time
	mu              sync.RWMutex
}

// IncrementRequest increments the total request counter
func (m *AIServiceMetrics) IncrementRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests++
}

// IncrementSuccess increments the success request counter
func (m *AIServiceMetrics) IncrementSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SuccessRequests++
}

// IncrementFallback increments the fallback counter
func (m *AIServiceMetrics) IncrementFallback() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FallbackCount++
}

// RecordResponseTime records the response time
func (m *AIServiceMetrics) RecordResponseTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Simple moving average calculation
	if m.TotalRequests > 0 {
		m.AvgResponseTime = (m.AvgResponseTime*float64(m.TotalRequests-1) + float64(duration.Milliseconds())) / float64(m.TotalRequests)
	} else {
		m.AvgResponseTime = float64(duration.Milliseconds())
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failures     int32
	lastFailTime time.Time
	state        int32 // 0: closed, 1: open, 2: half-open
	mu           sync.Mutex
}

const (
	circuitClosed = iota
	circuitOpen
	circuitHalfOpen
)

// EnhancedAIService is the SOTA implementation with production-ready features
type EnhancedAIService struct {
	*AIService
	metrics        *AIServiceMetrics
	circuitBreaker *CircuitBreaker
	retryConfig    RetryConfig
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

// NewEnhancedAIService creates a new enhanced AI service with SOTA features
func NewEnhancedAIService(db *gorm.DB, config *config.Config) *EnhancedAIService {
	baseService := NewAIService(db, config)
	
	return &EnhancedAIService{
		AIService: baseService,
		metrics: &AIServiceMetrics{},
		circuitBreaker: &CircuitBreaker{
			maxFailures:  5,
			resetTimeout: 30 * time.Second,
		},
		retryConfig: RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     30 * time.Second,
			BackoffFactor:  2.0,
		},
	}
}

// callMoonshotWithRetry implements the enhanced Moonshot API call with retry and circuit breaker
func (s *EnhancedAIService) callMoonshotWithRetry(ctx context.Context, config *models.AIConfig, prompt string) (string, error) {
	// Start metrics
	startTime := time.Now()
	atomic.AddInt64(&s.metrics.TotalRequests, 1)
	
	// Check circuit breaker
	if !s.circuitBreaker.canRequest() {
		atomic.AddInt64(&s.metrics.FailedRequests, 1)
		s.updateLastError("Circuit breaker is open - too many recent failures")
		return "", fmt.Errorf("circuit breaker is open")
	}
	
	var lastErr error
	backoff := s.retryConfig.InitialBackoff
	
	for attempt := 0; attempt <= s.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("🔄 [Moonshot] Retry attempt %d/%d after %v backoff", attempt, s.retryConfig.MaxRetries, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", ctx.Err()
			}
			backoff = s.calculateBackoff(backoff)
		}
		
		result, err := s.callMoonshotOnce(ctx, config, prompt)
		if err == nil {
			// Success
			s.circuitBreaker.recordSuccess()
			atomic.AddInt64(&s.metrics.SuccessRequests, 1)
			s.updateResponseTime(time.Since(startTime))
			log.Printf("✅ [Moonshot] API call successful (attempt %d)", attempt+1)
			return result, nil
		}
		
		lastErr = err
		
		// Analyze error type
		if s.isRetriableError(err) {
			log.Printf("⚠️ [Moonshot] Retriable error on attempt %d: %v", attempt+1, err)
			continue
		} else {
			// Non-retriable error
			log.Printf("❌ [Moonshot] Non-retriable error on attempt %d: %v", attempt+1, err)
			break
		}
	}
	
	// All attempts failed
	s.circuitBreaker.recordFailure()
	atomic.AddInt64(&s.metrics.FailedRequests, 1)
	s.updateLastError(lastErr.Error())
	
	return "", fmt.Errorf("moonshot API call failed after %d attempts: %w", s.retryConfig.MaxRetries+1, lastErr)
}

// callMoonshotOnce performs a single API call with enhanced error handling
func (s *EnhancedAIService) callMoonshotOnce(ctx context.Context, config *models.AIConfig, prompt string) (string, error) {
	log.Printf("🌙 [Moonshot] Starting API call (timeout: 30s)...")
	
	// Create context with timeout
	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Validate configuration
	if config.APIKey == "" {
		return "", errors.New("moonshot API key is empty")
	}
	
	if config.APIEndpoint == "" {
		config.APIEndpoint = "https://api.moonshot.cn/v1/chat/completions"
	}
	
	// Build request body
	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是OpenPenPal的AI助手，在这个温暖的数字书信平台上，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好、富有人文情怀的语气回应。回复时请使用中文。",
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
	
	log.Printf("🌙 [Moonshot] Request size: %d bytes", len(jsonData))
	
	// Create request
	req, err := http.NewRequestWithContext(callCtx, "POST", config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("User-Agent", "OpenPenPal/1.0")
	
	// 安全日志：不记录任何API密钥信息
	if len(config.APIKey) > 0 {
		log.Printf("🔑 [Moonshot] API Key configured")
	} else {
		log.Printf("⚠️ [Moonshot] No API Key configured")
	}
	
	// Send request
	log.Printf("🚀 [Moonshot] Sending request to %s", config.APIEndpoint)
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	log.Printf("📥 [Moonshot] Response status: %d", resp.StatusCode)
	log.Printf("📥 [Moonshot] Response size: %d bytes", len(body))
	
	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return "", fmt.Errorf("moonshot API error (%d): %s", resp.StatusCode, errorResp.Error.Message)
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
			Index        int    `json:"index"`
			Message      struct {
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
		log.Printf("❌ [Moonshot] Failed to parse response: %v", err)
		log.Printf("❌ [Moonshot] Raw response: %s", string(body))
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Validate response
	if len(result.Choices) == 0 {
		return "", errors.New("no choices in response")
	}
	
	content := result.Choices[0].Message.Content
	if content == "" {
		return "", errors.New("empty content in response")
	}
	
	log.Printf("✅ [Moonshot] Successfully received response: %d tokens used", result.Usage.TotalTokens)
	
	// Update usage metrics
	s.logAIUsage("system", models.TaskTypeInspiration, "", config, 
		result.Usage.PromptTokens, result.Usage.CompletionTokens, "success", "")
	
	return content, nil
}

// parseInspirationResponse delegates to the base service
func (s *EnhancedAIService) parseInspirationResponse(aiResponse string) (*models.AIInspirationResponse, error) {
	return s.AIService.parseInspirationResponse(aiResponse)
}

// getFallbackInspiration delegates to the base service
func (s *EnhancedAIService) getFallbackInspiration(req *models.AIInspirationRequest) *models.AIInspirationResponse {
	// Get the handler instance through a type assertion
	// This is a temporary solution - in production, this should be refactored
	inspirationPool := []struct {
		ID     string   `json:"id"`
		Theme  string   `json:"theme"`
		Prompt string   `json:"prompt"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
	}{
		{
			ID:     "1",
			Theme:  "日常生活",
			Prompt: "写一写你今天遇到的一个有趣的人或事，可以是在路上、在学校，或是在任何地方的小小惊喜。",
			Style:  "轻松随意",
			Tags:   []string{"日常", "生活", "观察"},
		},
		{
			ID:     "2",
			Theme:  "情感表达",
			Prompt: "想起一个让你印象深刻的瞬间，可能是开心、感动，或是有些失落的时刻，把这份情感写出来。",
			Style:  "真诚温暖",
			Tags:   []string{"情感", "回忆", "真诚"},
		},
		{
			ID:     "3",
			Theme:  "梦想话题",
			Prompt: "如果你能实现一个小小的愿望，会是什么？不需要很宏大，就是那种想想就会微笑的心愿。",
			Style:  "充满希望",
			Tags:   []string{"梦想", "愿望", "未来"},
		},
	}

	// Filter by theme if specified
	var selectedInspirations []struct {
		ID     string   `json:"id"`
		Theme  string   `json:"theme"`
		Prompt string   `json:"prompt"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
	}
	
	if req.Theme != "" {
		for _, insp := range inspirationPool {
			if insp.Theme == req.Theme {
				selectedInspirations = append(selectedInspirations, insp)
			}
		}
	}
	
	if len(selectedInspirations) == 0 {
		selectedInspirations = inspirationPool[:req.Count]
	}

	return &models.AIInspirationResponse{
		Inspirations: selectedInspirations,
	}
}

// Enhanced GetInspiration with comprehensive error handling
func (s *EnhancedAIService) GetInspiration(ctx context.Context, req *models.AIInspirationRequest) (*models.AIInspirationResponse, error) {
	log.Printf("🎯 [GetInspiration] Starting enhanced inspiration generation...")
	
	// Get AI configuration
	aiConfig, err := s.GetActiveProvider()
	if err != nil {
		log.Printf("❌ [GetInspiration] Failed to get AI provider: %v", err)
		atomic.AddInt64(&s.metrics.FallbackCount, 1)
		return s.getFallbackInspirationWithMetrics(req)
	}
	
	log.Printf("🔧 [GetInspiration] Using provider: %s, Model: %s", aiConfig.Provider, aiConfig.Model)
	
	// Build prompt
	prompt := s.buildInspirationPrompt(req)
	log.Printf("📝 [GetInspiration] Generated prompt: %d characters", len(prompt))
	
	// Call API with retry and circuit breaker
	var aiResponse string
	if aiConfig.Provider == models.ProviderMoonshot {
		aiResponse, err = s.callMoonshotWithRetry(ctx, aiConfig, prompt)
	} else {
		// Fallback to original implementation for other providers
		aiResponse, err = s.callAIAPI(ctx, aiConfig, prompt, models.TaskTypeInspiration)
	}
	
	if err != nil {
		log.Printf("❌ [GetInspiration] AI API call failed: %v", err)
		atomic.AddInt64(&s.metrics.FallbackCount, 1)
		return s.getFallbackInspirationWithMetrics(req)
	}
	
	log.Printf("✅ [GetInspiration] AI API response received: %d characters", len(aiResponse))
	
	// Parse response
	inspirations, err := s.parseInspirationResponse(aiResponse)
	if err != nil {
		log.Printf("❌ [GetInspiration] Failed to parse AI response: %v", err)
		atomic.AddInt64(&s.metrics.FallbackCount, 1)
		return s.getFallbackInspirationWithMetrics(req)
	}
	
	// Save to database
	for i, insp := range inspirations.Inspirations {
		inspiration := &models.AIInspiration{
			ID:        uuid.New().String(),
			Theme:     insp.Theme,
			Prompt:    insp.Prompt,
			Style:     insp.Style,
			Tags:      fmt.Sprintf("[%s]", strings.Join(insp.Tags, ",")),
			Provider:  aiConfig.Provider,
			CreatedAt: time.Now(),
			IsActive:  true,
		}
		if err := s.db.Create(inspiration).Error; err != nil {
			log.Printf("⚠️ [GetInspiration] Failed to save inspiration %d: %v", i, err)
		}
		inspirations.Inspirations[i].ID = inspiration.ID
	}
	
	log.Printf("✅ [GetInspiration] Successfully generated %d inspirations", len(inspirations.Inspirations))
	
	return inspirations, nil
}

// Helper methods

func (s *EnhancedAIService) isRetriableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	
	// Network errors are retriable
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "temporary failure") {
		return true
	}
	
	// Rate limit errors are retriable
	if strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "429") {
		return true
	}
	
	// 5xx errors are retriable
	if strings.Contains(errStr, "status 5") {
		return true
	}
	
	// Auth errors and 4xx errors are not retriable
	if strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "invalid api key") {
		return false
	}
	
	return false
}

func (s *EnhancedAIService) calculateBackoff(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * s.retryConfig.BackoffFactor)
	if next > s.retryConfig.MaxBackoff {
		return s.retryConfig.MaxBackoff
	}
	return next
}

func (s *EnhancedAIService) updateResponseTime(duration time.Duration) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()
	
	// Simple moving average
	if s.metrics.AvgResponseTime == 0 {
		s.metrics.AvgResponseTime = duration.Seconds()
	} else {
		s.metrics.AvgResponseTime = (s.metrics.AvgResponseTime*0.9) + (duration.Seconds()*0.1)
	}
}

func (s *EnhancedAIService) updateLastError(err string) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()
	
	s.metrics.LastError = err
	s.metrics.LastErrorTime = time.Now()
}

func (s *EnhancedAIService) getFallbackInspirationWithMetrics(req *models.AIInspirationRequest) (*models.AIInspirationResponse, error) {
	log.Printf("⚠️ [GetInspiration] Using fallback inspiration due to API issues")
	
	// Use the original fallback method
	response := s.getFallbackInspiration(req)
	
	// Mark response as fallback
	if response != nil && len(response.Inspirations) > 0 {
		for i := range response.Inspirations {
			// Add a marker to indicate this is fallback content
			response.Inspirations[i].ID = fmt.Sprintf("fallback_%s", response.Inspirations[i].ID)
		}
	}
	
	return response, nil
}

// Circuit breaker methods

func (cb *CircuitBreaker) canRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	state := atomic.LoadInt32(&cb.state)
	
	switch state {
	case circuitClosed:
		return true
	case circuitOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			atomic.StoreInt32(&cb.state, circuitHalfOpen)
			log.Printf("🔄 [CircuitBreaker] Transitioning to half-open state")
			return true
		}
		return false
	case circuitHalfOpen:
		return true
	}
	
	return false
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	atomic.StoreInt32(&cb.failures, 0)
	if atomic.LoadInt32(&cb.state) == circuitHalfOpen {
		atomic.StoreInt32(&cb.state, circuitClosed)
		log.Printf("✅ [CircuitBreaker] Circuit closed after successful request")
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	failures := atomic.AddInt32(&cb.failures, 1)
	cb.lastFailTime = time.Now()
	
	if failures >= int32(cb.maxFailures) {
		atomic.StoreInt32(&cb.state, circuitOpen)
		log.Printf("❌ [CircuitBreaker] Circuit opened after %d failures", failures)
	}
}

// GetMetrics returns current service metrics
func (s *EnhancedAIService) GetMetrics() AIServiceMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()
	
	return AIServiceMetrics{
		TotalRequests:   atomic.LoadInt64(&s.metrics.TotalRequests),
		SuccessRequests: atomic.LoadInt64(&s.metrics.SuccessRequests),
		FailedRequests:  atomic.LoadInt64(&s.metrics.FailedRequests),
		FallbackCount:   atomic.LoadInt64(&s.metrics.FallbackCount),
		AvgResponseTime: s.metrics.AvgResponseTime,
		LastError:       s.metrics.LastError,
		LastErrorTime:   s.metrics.LastErrorTime,
	}
}