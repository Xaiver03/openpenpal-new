// Package services provides enhanced moderation with optimized database queries
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"openpenpal-backend/internal/logger"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/pkg/cache"
)

// EnhancedModerationService provides optimized content moderation with query caching
type EnhancedModerationService struct {
	db           *gorm.DB
	cacheManager *cache.EnhancedCacheManager
	queryCacheService *QueryCacheService
	enableCache  bool
}

// NewEnhancedModerationService creates a new enhanced moderation service
func NewEnhancedModerationService(db *gorm.DB, cacheManager *cache.EnhancedCacheManager) *EnhancedModerationService {
	queryCacheService := NewQueryCacheService(db, cacheManager)
	
	return &EnhancedModerationService{
		db:                db,
		cacheManager:      cacheManager,
		queryCacheService: queryCacheService,
		enableCache:       true,
	}
}

// ModerationResult represents the result of content moderation
type ModerationResult struct {
	Level      models.ModerationLevel `json:"level"`
	Reasons    []string         `json:"reasons"`
	Categories []string         `json:"categories"`
	Score      int              `json:"score"`      // 0-100 risk score
	Blocked    bool             `json:"blocked"`    // Whether content should be blocked
	Warnings   []string         `json:"warnings"`   // Non-blocking warnings
}

// OptimizedCheckContent performs optimized content moderation using database-side filtering
func (s *EnhancedModerationService) OptimizedCheckContent(ctx context.Context, content string) (*ModerationResult, error) {
	result := &ModerationResult{
		Level:      models.LevelLow,
		Reasons:    []string{},
		Categories: []string{},
		Score:      0,
		Blocked:    false,
		Warnings:   []string{},
	}

	if strings.TrimSpace(content) == "" {
		return result, nil
	}

	// Method 1: Use database-side pattern matching (most efficient)
	if err := s.checkContentWithDatabaseQuery(ctx, content, result); err != nil {
		logger.Error("Database-side moderation failed, falling back to cache", err)
		
		// Method 2: Fallback to cached sensitive words
		if err := s.checkContentWithCache(ctx, content, result); err != nil {
			logger.Error("Cache-based moderation failed", err)
			return nil, err
		}
	}

	// Calculate final score and blocking decision
	s.calculateFinalScore(result)

	return result, nil
}

// checkContentWithDatabaseQuery uses database-side pattern matching for optimal performance
func (s *EnhancedModerationService) checkContentWithDatabaseQuery(ctx context.Context, content string, result *ModerationResult) error {
	// Use SQL pattern matching to find sensitive words directly in database
	var matches []struct {
		Word     string           `json:"word"`
		Level    models.RiskLevel `json:"level"`
		Category string           `json:"category"`
		Score    int              `json:"score"`
	}

	contentLower := strings.ToLower(content)
	
	// Use database LIKE queries for pattern matching
	err := s.db.Table("sensitive_words").
		Select("word, level, category, score").
		Where("is_active = ?", true).
		Where("? LIKE CONCAT('%', LOWER(word), '%')", contentLower).
		Find(&matches).Error

	if err != nil {
		return err
	}

	// Process matches
	s.processMatches(matches, result)
	
	logger.Debug("Database-side content moderation completed",
		"contentLength", len(content),
		"matches", len(matches),
		"level", result.Level,
	)

	return nil
}

// checkContentWithCache uses cached sensitive words for moderation
func (s *EnhancedModerationService) checkContentWithCache(ctx context.Context, content string, result *ModerationResult) error {
	// Get cached sensitive words
	words, err := s.queryCacheService.GetSensitiveWords(ctx)
	if err != nil {
		return err
	}

	contentLower := strings.ToLower(content)
	
	// Check each cached word
	for _, word := range words {
		if strings.Contains(contentLower, strings.ToLower(word)) {
			// Get word details from database
			var wordDetail models.SensitiveWord
			if err := s.db.Where("word = ? AND is_active = ?", word, true).First(&wordDetail).Error; err != nil {
				continue
			}

			result.Reasons = append(result.Reasons, fmt.Sprintf("包含敏感词: %s", word))
			
			if wordDetail.Category != "" && !containsString(result.Categories, wordDetail.Category) {
				result.Categories = append(result.Categories, wordDetail.Category)
			}

			// Update risk level
			if wordDetail.Level == models.LevelBlock {
				result.Level = models.LevelBlock
			} else if wordDetail.Level == models.LevelHigh && result.Level != models.LevelBlock {
				result.Level = models.LevelHigh
			} else if wordDetail.Level == models.LevelMedium && result.Level == models.LevelLow {
				result.Level = models.LevelMedium
			}
		}
	}

	logger.Debug("Cache-based content moderation completed",
		"contentLength", len(content),
		"wordsChecked", len(words),
		"level", result.Level,
	)

	return nil
}

// processMatches processes database query matches
func (s *EnhancedModerationService) processMatches(matches []struct {
	Word     string           `json:"word"`
	Level    models.RiskLevel `json:"level"`
	Category string           `json:"category"`
	Score    int              `json:"score"`
}, result *ModerationResult) {
	
	for _, match := range matches {
		result.Reasons = append(result.Reasons, fmt.Sprintf("包含敏感词: %s", match.Word))
		
		if match.Category != "" && !containsString(result.Categories, match.Category) {
			result.Categories = append(result.Categories, match.Category)
		}

		// Update risk level
		if match.Level == models.RiskLevelBlocked {
			result.Level = models.LevelBlock
		} else if match.Level == models.RiskLevelHigh && result.Level != models.LevelBlock {
			result.Level = models.LevelHigh
		} else if match.Level == models.RiskLevelMedium && result.Level == models.LevelLow {
			result.Level = models.LevelMedium
		}

		// Accumulate score
		result.Score += match.Score
	}
}

// calculateFinalScore calculates final moderation score and blocking decision
func (s *EnhancedModerationService) calculateFinalScore(result *ModerationResult) {
	// Ensure score doesn't exceed 100
	if result.Score > 100 {
		result.Score = 100
	}

	// Determine blocking based on level and score
	switch result.Level {
	case models.LevelBlock:
		result.Blocked = true
	case models.LevelHigh:
		result.Blocked = result.Score >= 80
	case models.LevelMedium:
		result.Blocked = result.Score >= 90
		if result.Score >= 60 && result.Score < 90 {
			result.Warnings = append(result.Warnings, "内容可能包含敏感信息，请谨慎处理")
		}
	default:
		result.Blocked = false
		if result.Score >= 30 {
			result.Warnings = append(result.Warnings, "建议人工审核")
		}
	}
}

// BatchCheckContent performs batch content moderation for multiple texts
func (s *EnhancedModerationService) BatchCheckContent(ctx context.Context, contents []string) (map[int]*ModerationResult, error) {
	results := make(map[int]*ModerationResult)

	// Use concurrent processing for batch operations
	type contentItem struct {
		index   int
		content string
	}

	// Process in batches to avoid overwhelming the database
	batchSize := 50
	for i := 0; i < len(contents); i += batchSize {
		end := i + batchSize
		if end > len(contents) {
			end = len(contents)
		}

		batch := contents[i:end]
		for j, content := range batch {
			result, err := s.OptimizedCheckContent(ctx, content)
			if err != nil {
				logger.Error("Failed to moderate content in batch", err, "index", i+j)
				results[i+j] = &ModerationResult{Level: models.LevelLow} // Default safe result
			} else {
				results[i+j] = result
			}
		}
	}

	logger.Info("Batch content moderation completed", "total", len(contents))
	return results, nil
}

// UpdateSensitiveWord updates a sensitive word and invalidates cache
func (s *EnhancedModerationService) UpdateSensitiveWord(ctx context.Context, word *models.SensitiveWord) error {
	if err := s.db.Save(word).Error; err != nil {
		return err
	}

	// Invalidate cache
	if err := s.queryCacheService.InvalidateSensitiveWords(ctx); err != nil {
		logger.Error("Failed to invalidate sensitive words cache", err)
	}

	logger.Info("Sensitive word updated and cache invalidated", "word", word.Word)
	return nil
}

// AddSensitiveWord adds a new sensitive word and invalidates cache
func (s *EnhancedModerationService) AddSensitiveWord(ctx context.Context, word *models.SensitiveWord) error {
	if err := s.db.Create(word).Error; err != nil {
		return err
	}

	// Invalidate cache
	if err := s.queryCacheService.InvalidateSensitiveWords(ctx); err != nil {
		logger.Error("Failed to invalidate sensitive words cache", err)
	}

	logger.Info("Sensitive word added and cache invalidated", "word", word.Word)
	return nil
}

// DeleteSensitiveWord deletes a sensitive word and invalidates cache
func (s *EnhancedModerationService) DeleteSensitiveWord(ctx context.Context, wordID string) error {
	if err := s.db.Delete(&models.SensitiveWord{}, "id = ?", wordID).Error; err != nil {
		return err
	}

	// Invalidate cache
	if err := s.queryCacheService.InvalidateSensitiveWords(ctx); err != nil {
		logger.Error("Failed to invalidate sensitive words cache", err)
	}

	logger.Info("Sensitive word deleted and cache invalidated", "wordID", wordID)
	return nil
}

// GetModerationStats returns moderation statistics with caching
func (s *EnhancedModerationService) GetModerationStats(ctx context.Context) (map[string]interface{}, error) {
	cacheKey := "moderation:stats"
	
	var stats map[string]interface{}
	if s.enableCache {
		if err := s.cacheManager.Get(ctx, cacheKey, &stats); err == nil {
			return stats, nil
		}
	}

	// Load stats from database
	var result struct {
		TotalWords     int64 `json:"total_words"`
		ActiveWords    int64 `json:"active_words"`
		BlockedWords   int64 `json:"blocked_words"`
		HighRiskWords  int64 `json:"high_risk_words"`
		MediumRiskWords int64 `json:"medium_risk_words"`
	}

	if err := s.db.Model(&models.SensitiveWord{}).Count(&result.TotalWords).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.SensitiveWord{}).Where("is_active = ?", true).Count(&result.ActiveWords).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.SensitiveWord{}).Where("is_active = ? AND level = ?", true, models.LevelBlock).Count(&result.BlockedWords).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.SensitiveWord{}).Where("is_active = ? AND level = ?", true, models.LevelHigh).Count(&result.HighRiskWords).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.SensitiveWord{}).Where("is_active = ? AND level = ?", true, models.LevelMedium).Count(&result.MediumRiskWords).Error; err != nil {
		return nil, err
	}

	stats = map[string]interface{}{
		"total_words":       result.TotalWords,
		"active_words":      result.ActiveWords,
		"blocked_words":     result.BlockedWords,
		"high_risk_words":   result.HighRiskWords,
		"medium_risk_words": result.MediumRiskWords,
		"last_updated":      time.Now(),
	}

	// Cache stats for 10 minutes
	if s.enableCache {
		if err := s.cacheManager.Set(ctx, cacheKey, stats, 10*time.Minute); err != nil {
			logger.Error("Failed to cache moderation stats", err)
		}
	}

	return stats, nil
}

// WarmupModerationCache preloads moderation-related cache
func (s *EnhancedModerationService) WarmupModerationCache(ctx context.Context) error {
	logger.Info("Starting moderation cache warmup")

	// Warmup sensitive words
	if _, err := s.queryCacheService.GetSensitiveWords(ctx); err != nil {
		logger.Error("Failed to warmup sensitive words", err)
		return err
	}

	// Warmup moderation stats
	if _, err := s.GetModerationStats(ctx); err != nil {
		logger.Error("Failed to warmup moderation stats", err)
		return err
	}

	logger.Info("Moderation cache warmup completed")
	return nil
}

// SetCacheEnabled enables or disables caching for moderation
func (s *EnhancedModerationService) SetCacheEnabled(enabled bool) {
	s.enableCache = enabled
	s.queryCacheService.SetCacheEnabled(enabled)
	logger.Info("Enhanced moderation cache enabled status changed", "enabled", enabled)
}

// Helper function
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}