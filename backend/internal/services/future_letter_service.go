package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"openpenpal-backend/internal/models"
	"sync"
	"time"

	"gorm.io/gorm"
)

// FutureLetterService handles automated future letter unlocking
type FutureLetterService struct {
	db              *gorm.DB
	letterService   *LetterService
	notificationSvc *NotificationService
	mu              sync.RWMutex
	isRunning       bool
}

// NewFutureLetterService creates a new future letter service
func NewFutureLetterService(db *gorm.DB, letterService *LetterService, notificationSvc *NotificationService) *FutureLetterService {
	return &FutureLetterService{
		db:              db,
		letterService:   letterService,
		notificationSvc: notificationSvc,
	}
}

// ProcessScheduledLetters checks and unlocks letters that are due
// This is the main task that should be called by the scheduler every 10 minutes
func (s *FutureLetterService) ProcessScheduledLetters(ctx context.Context) error {
	// Prevent concurrent execution
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return errors.New("future letter processor is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
	}()

	// Start transaction for atomic operation
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Find all scheduled letters that are due for publishing
	var scheduledLetters []models.Letter
	now := time.Now()
	
	err := tx.Where("status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?", 
		"scheduled", now).
		Find(&scheduledLetters).Error
	
	if err != nil {
		return fmt.Errorf("failed to query scheduled letters: %w", err)
	}

	log.Printf("[FutureLetterService] Found %d letters ready to unlock", len(scheduledLetters))

	// Process each letter
	successCount := 0
	failedCount := 0
	
	for _, letter := range scheduledLetters {
		if err := s.unlockLetter(tx, &letter); err != nil {
			log.Printf("[FutureLetterService] Failed to unlock letter %s: %v", letter.ID, err)
			failedCount++
			continue
		}
		successCount++
		
		// Send notification asynchronously
		go s.notifyRecipient(&letter)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[FutureLetterService] Processed %d letters (success: %d, failed: %d)", 
		len(scheduledLetters), successCount, failedCount)

	// Record metrics
	s.recordMetrics(successCount, failedCount)

	return nil
}

// unlockLetter updates the letter status from scheduled to published
func (s *FutureLetterService) unlockLetter(tx *gorm.DB, letter *models.Letter) error {
	updates := map[string]interface{}{
		"status":       "published",
		"published_at": time.Now(),
		"updated_at":   time.Now(),
	}

	result := tx.Model(letter).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update letter status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

// notifyRecipient sends notification to the letter recipient
func (s *FutureLetterService) notifyRecipient(letter *models.Letter) {
	if s.notificationSvc == nil {
		return
	}

	// Determine recipient based on letter type
	var recipientID string
	if letter.RecipientID != "" {
		recipientID = letter.RecipientID
	} else if letter.UserID != "" && letter.Visibility == "self" {
		// Self-addressed future letter
		recipientID = letter.UserID
	} else {
		// Public letter - no specific recipient notification
		return
	}

	// Create notification payload
	notificationData := map[string]interface{}{
		"letter_id":     letter.ID,
		"letter_title":  letter.Title,
		"author_id":     letter.UserID,
		"author_name":   letter.AuthorName,
		"scheduled_at":  letter.ScheduledAt,
		"unlock_time":   time.Now(),
	}

	// Send notification
	err := s.notificationSvc.NotifyUser(recipientID, "future_letter_unlocked", notificationData)
	if err != nil {
		log.Printf("[FutureLetterService] Failed to send notification for letter %s: %v", 
			letter.ID, err)
	}
}

// recordMetrics records processing metrics for monitoring
func (s *FutureLetterService) recordMetrics(success, failed int) {
	// This would integrate with your monitoring system (Prometheus)
	// For now, just log the metrics
	metrics := map[string]interface{}{
		"timestamp":     time.Now(),
		"success_count": success,
		"failed_count":  failed,
		"total_count":   success + failed,
		"success_rate":  float64(success) / float64(success+failed),
	}
	
	log.Printf("[FutureLetterService] Metrics: %+v", metrics)
}

// GetPendingCount returns the count of pending future letters
func (s *FutureLetterService) GetPendingCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&models.Letter{}).
		Where("status = ? AND scheduled_at IS NOT NULL AND scheduled_at > ?", 
			"scheduled", time.Now()).
		Count(&count).Error
	
	return count, err
}

// GetUpcomingLetters returns letters scheduled to be unlocked soon
func (s *FutureLetterService) GetUpcomingLetters(ctx context.Context, hours int) ([]models.Letter, error) {
	var letters []models.Letter
	deadline := time.Now().Add(time.Duration(hours) * time.Hour)
	
	err := s.db.WithContext(ctx).
		Where("status = ? AND scheduled_at IS NOT NULL AND scheduled_at BETWEEN ? AND ?",
			"scheduled", time.Now(), deadline).
		Order("scheduled_at ASC").
		Find(&letters).Error
	
	return letters, err
}

// CancelScheduledLetter cancels a scheduled future letter
func (s *FutureLetterService) CancelScheduledLetter(ctx context.Context, letterID, userID string) error {
	result := s.db.WithContext(ctx).
		Model(&models.Letter{}).
		Where("id = ? AND author_id = ? AND status = ?", letterID, userID, "scheduled").
		Updates(map[string]interface{}{
			"status":       "draft",
			"scheduled_at": nil,
			"updated_at":   time.Now(),
		})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("letter not found or not scheduled")
	}
	
	return nil
}