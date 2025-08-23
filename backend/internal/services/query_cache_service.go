// Package services provides query caching functionality to optimize database performance
package services

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"openpenpal-backend/internal/logger"
	"openpenpal-backend/pkg/cache"
)

// QueryCacheService provides caching for database queries to reduce N+1 problems and improve performance
type QueryCacheService struct {
	db           *gorm.DB
	cacheManager *cache.EnhancedCacheManager
	defaultTTL   time.Duration
	enableCache  bool
}

// NewQueryCacheService creates a new query cache service
func NewQueryCacheService(db *gorm.DB, cacheManager *cache.EnhancedCacheManager) *QueryCacheService {
	return &QueryCacheService{
		db:           db,
		cacheManager: cacheManager,
		defaultTTL:   5 * time.Minute, // Default 5 minute cache
		enableCache:  true,
	}
}

// UserCacheKey generates cache key for user data
func (s *QueryCacheService) UserCacheKey(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

// UsersBatchCacheKey generates cache key for batch user lookup
func (s *QueryCacheService) UsersBatchCacheKey(userIDs []string) string {
	return fmt.Sprintf("users:batch:%v", userIDs)
}

// SensitiveWordsCacheKey generates cache key for sensitive words
func (s *QueryCacheService) SensitiveWordsCacheKey() string {
	return "sensitive_words:active"
}

// ConfigCacheKey generates cache key for configuration
func (s *QueryCacheService) ConfigCacheKey(configType, key string) string {
	return fmt.Sprintf("config:%s:%s", configType, key)
}

// SchoolCacheKey generates cache key for school information
func (s *QueryCacheService) SchoolCacheKey(schoolCode string) string {
	return fmt.Sprintf("school:%s", schoolCode)
}

// PermissionCacheKey generates cache key for user permissions
func (s *QueryCacheService) PermissionCacheKey(userID string) string {
	return fmt.Sprintf("permissions:%s", userID)
}

// CacheUser stores user information in cache
func (s *QueryCacheService) CacheUser(ctx context.Context, user interface{}) error {
	if !s.enableCache {
		return nil
	}

	// Extract user ID for key generation (assumes user has ID field)
	userMap, ok := user.(map[string]interface{})
	if !ok {
		return fmt.Errorf("user must be a map[string]interface{}")
	}

	userID, ok := userMap["id"].(string)
	if !ok {
		return fmt.Errorf("user ID not found or not string")
	}

	key := s.UserCacheKey(userID)
	return s.cacheManager.Set(ctx, key, user, s.defaultTTL)
}

// GetUser retrieves user from cache or database
func (s *QueryCacheService) GetUser(ctx context.Context, userID string, result interface{}) error {
	if !s.enableCache {
		return s.db.First(result, "id = ?", userID).Error
	}

	key := s.UserCacheKey(userID)
	
	// Try cache first
	if err := s.cacheManager.Get(ctx, key, result); err == nil {
		logger.Debug("Cache hit for user", "userID", userID)
		return nil
	}

	// Cache miss - load from database
	if err := s.db.First(result, "id = ?", userID).Error; err != nil {
		return err
	}

	// Cache the result
	if err := s.cacheManager.Set(ctx, key, result, s.defaultTTL); err != nil {
		logger.Error("Failed to cache user", err, "userID", userID)
	}

	logger.Debug("Cache miss for user - loaded from database", "userID", userID)
	return nil
}

// GetSensitiveWords retrieves active sensitive words with caching
func (s *QueryCacheService) GetSensitiveWords(ctx context.Context) ([]string, error) {
	if !s.enableCache {
		return s.loadSensitiveWordsFromDB()
	}

	key := s.SensitiveWordsCacheKey()
	
	// Try cache first
	var words []string
	if err := s.cacheManager.Get(ctx, key, &words); err == nil {
		logger.Debug("Cache hit for sensitive words", "count", len(words))
		return words, nil
	}

	// Cache miss - load from database
	words, err := s.loadSensitiveWordsFromDB()
	if err != nil {
		return nil, err
	}

	// Cache the result with shorter TTL for sensitive data
	cacheTTL := 30 * time.Minute
	if err := s.cacheManager.Set(ctx, key, words, cacheTTL); err != nil {
		logger.Error("Failed to cache sensitive words", err)
	}

	logger.Debug("Cache miss for sensitive words - loaded from database", "count", len(words))
	return words, nil
}

// loadSensitiveWordsFromDB loads sensitive words from database
func (s *QueryCacheService) loadSensitiveWordsFromDB() ([]string, error) {
	var words []struct {
		Word string `json:"word"`
	}
	
	err := s.db.Table("sensitive_words").
		Select("word").
		Where("is_active = ?", true).
		Find(&words).Error
	
	if err != nil {
		return nil, err
	}

	result := make([]string, len(words))
	for i, w := range words {
		result[i] = w.Word
	}
	
	return result, nil
}

// GetSchoolInfo retrieves school information with caching
func (s *QueryCacheService) GetSchoolInfo(ctx context.Context, schoolCode string, result interface{}) error {
	if !s.enableCache {
		return s.db.First(result, "code = ?", schoolCode).Error
	}

	key := s.SchoolCacheKey(schoolCode)
	
	// Try cache first
	if err := s.cacheManager.Get(ctx, key, result); err == nil {
		logger.Debug("Cache hit for school", "schoolCode", schoolCode)
		return nil
	}

	// Cache miss - load from database
	if err := s.db.First(result, "code = ?", schoolCode).Error; err != nil {
		return err
	}

	// Cache with longer TTL as school info changes infrequently
	cacheTTL := 2 * time.Hour
	if err := s.cacheManager.Set(ctx, key, result, cacheTTL); err != nil {
		logger.Error("Failed to cache school info", err, "schoolCode", schoolCode)
	}

	logger.Debug("Cache miss for school - loaded from database", "schoolCode", schoolCode)
	return nil
}

// GetUserPermissions retrieves user permissions with caching
func (s *QueryCacheService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	if !s.enableCache {
		return s.loadUserPermissionsFromDB(userID)
	}

	key := s.PermissionCacheKey(userID)
	
	// Try cache first
	var permissions []string
	if err := s.cacheManager.Get(ctx, key, &permissions); err == nil {
		logger.Debug("Cache hit for user permissions", "userID", userID)
		return permissions, nil
	}

	// Cache miss - load from database
	permissions, err := s.loadUserPermissionsFromDB(userID)
	if err != nil {
		return nil, err
	}

	// Cache with moderate TTL
	cacheTTL := 15 * time.Minute
	if err := s.cacheManager.Set(ctx, key, permissions, cacheTTL); err != nil {
		logger.Error("Failed to cache user permissions", err, "userID", userID)
	}

	logger.Debug("Cache miss for user permissions - loaded from database", "userID", userID)
	return permissions, nil
}

// loadUserPermissionsFromDB loads user permissions from database
func (s *QueryCacheService) loadUserPermissionsFromDB(userID string) ([]string, error) {
	var user struct {
		Role string `json:"role"`
	}
	
	if err := s.db.Table("users").Select("role").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	// Map roles to permissions (simplified example)
	rolePermissions := map[string][]string{
		"super_admin":         {"*"},
		"platform_admin":      {"admin.*", "user.*", "letter.*", "courier.*"},
		"school_admin":        {"school.*", "user.read", "letter.*", "courier.*"},
		"courier_coordinator": {"courier.*", "letter.read", "user.read"},
		"senior_courier":      {"courier.read", "letter.read", "task.*"},
		"courier":             {"courier.read", "task.own"},
		"user":                {"letter.own", "profile.own"},
	}

	return rolePermissions[user.Role], nil
}

// BatchGetUsers efficiently retrieves multiple users with caching
func (s *QueryCacheService) BatchGetUsers(ctx context.Context, userIDs []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var missingIDs []string

	if s.enableCache {
		// Check cache for each user
		for _, userID := range userIDs {
			key := s.UserCacheKey(userID)
			var user interface{}
			if err := s.cacheManager.Get(ctx, key, &user); err == nil {
				result[userID] = user
			} else {
				missingIDs = append(missingIDs, userID)
			}
		}
	} else {
		missingIDs = userIDs
	}

	// Load missing users from database
	if len(missingIDs) > 0 {
		var users []map[string]interface{}
		if err := s.db.Table("users").Where("id IN ?", missingIDs).Find(&users).Error; err != nil {
			return nil, err
		}

		// Add to result and cache
		for _, user := range users {
			userID := user["id"].(string)
			result[userID] = user

			if s.enableCache {
				key := s.UserCacheKey(userID)
				if err := s.cacheManager.Set(ctx, key, user, s.defaultTTL); err != nil {
					logger.Error("Failed to cache user in batch operation", err, "userID", userID)
				}
			}
		}
	}

	logger.Debug("Batch user lookup completed", 
		"total", len(userIDs),
		"cached", len(userIDs)-len(missingIDs),
		"loaded", len(missingIDs),
	)

	return result, nil
}

// InvalidateUser removes user from cache
func (s *QueryCacheService) InvalidateUser(ctx context.Context, userID string) error {
	if !s.enableCache {
		return nil
	}

	keys := []string{
		s.UserCacheKey(userID),
		s.PermissionCacheKey(userID),
	}

	return s.cacheManager.Delete(ctx, keys...)
}

// InvalidateSensitiveWords removes sensitive words from cache
func (s *QueryCacheService) InvalidateSensitiveWords(ctx context.Context) error {
	if !s.enableCache {
		return nil
	}

	return s.cacheManager.Delete(ctx, s.SensitiveWordsCacheKey())
}

// InvalidateSchool removes school info from cache
func (s *QueryCacheService) InvalidateSchool(ctx context.Context, schoolCode string) error {
	if !s.enableCache {
		return nil
	}

	return s.cacheManager.Delete(ctx, s.SchoolCacheKey(schoolCode))
}

// WarmupCommonData preloads frequently accessed data
func (s *QueryCacheService) WarmupCommonData(ctx context.Context) error {
	if !s.enableCache {
		return nil
	}

	logger.Info("Starting cache warmup for common data")

	// Warmup sensitive words
	if _, err := s.GetSensitiveWords(ctx); err != nil {
		logger.Error("Failed to warmup sensitive words", err)
	}

	// Warmup common schools
	commonSchools := []string{"PK", "QH", "BD", "BU", "RU"} // Common school codes
	for _, schoolCode := range commonSchools {
		var school interface{}
		if err := s.GetSchoolInfo(ctx, schoolCode, &school); err != nil {
			logger.Error("Failed to warmup school info", err, "schoolCode", schoolCode)
		}
	}

	logger.Info("Cache warmup completed")
	return nil
}

// GetCacheStats returns cache statistics
func (s *QueryCacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	return s.cacheManager.GetStats(ctx)
}

// SetCacheEnabled enables or disables caching
func (s *QueryCacheService) SetCacheEnabled(enabled bool) {
	s.enableCache = enabled
	logger.Info("Query cache enabled status changed", "enabled", enabled)
}

// SetDefaultTTL sets the default TTL for cache entries
func (s *QueryCacheService) SetDefaultTTL(ttl time.Duration) {
	s.defaultTTL = ttl
	logger.Info("Default cache TTL changed", "ttl", ttl)
}