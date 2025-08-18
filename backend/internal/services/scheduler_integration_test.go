package services

import (
	"context"
	"fmt"
	"openpenpal-backend/internal/models"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// TODO: Use PostgreSQL for testing when available
	// "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestSchedulerIntegration tests the complete scheduler system with all enhancements
// TODO: Re-enable when PostgreSQL test database is available
func TestSchedulerIntegration(t *testing.T) {
	t.Skip("Scheduler integration test disabled - PostgreSQL test environment needed")
	
	// TODO: Setup PostgreSQL test database
	// db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	// require.NoError(t, err)
	
	// Migrate schemas
	err = db.AutoMigrate(
		&models.ScheduledTask{},
		&models.TaskExecution{},
		&models.Letter{},
		&models.User{},
		&models.Notification{},
	)
	require.NoError(t, err)

	// Setup Redis client (use mock in tests)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use test DB
	})
	
	// Clear test Redis DB
	redisClient.FlushDB(context.Background())

	// Create services
	letterService := &LetterService{db: db}
	notificationService := &NotificationService{db: db}
	futureLetterService := NewFutureLetterService(db, letterService, notificationService)
	
	// Create enhanced scheduler
	scheduler := NewEnhancedSchedulerService(
		db,
		redisClient,
		futureLetterService,
		letterService,
		nil, // AI service
		notificationService,
		nil, // Envelope service
		nil, // Courier service
	)

	t.Run("TestFutureLetterUnlock", func(t *testing.T) {
		// Create a scheduled letter
		letter := &models.Letter{
			ID:          "test-letter-1",
			UserID:      "test-user-1",
			Title:       "Future Letter",
			Content:     "This is a future letter",
			Status:      "scheduled",
			ScheduledAt: &[]time.Time{time.Now().Add(-1 * time.Minute)}[0], // Scheduled in the past
		}
		err := db.Create(letter).Error
		require.NoError(t, err)

		// Process scheduled letters
		err = futureLetterService.ProcessScheduledLetters(context.Background())
		assert.NoError(t, err)

		// Verify letter was unlocked
		var updatedLetter models.Letter
		err = db.First(&updatedLetter, "id = ?", letter.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "published", updatedLetter.Status)
		assert.NotNil(t, updatedLetter.PublishedAt)
	})

	t.Run("TestDistributedLocking", func(t *testing.T) {
		lockManager := NewDistributedLockManager(redisClient, "test:")
		
		// Test basic lock acquisition
		lock1 := lockManager.NewLock("test-resource", 5*time.Second)
		err := lock1.Acquire(context.Background())
		assert.NoError(t, err)
		
		// Try to acquire same lock from another instance
		lock2 := lockManager.NewLock("test-resource", 5*time.Second)
		err = lock2.Acquire(context.Background())
		assert.Error(t, err)
		assert.Equal(t, ErrLockNotAcquired, err)
		
		// Release first lock
		err = lock1.Release(context.Background())
		assert.NoError(t, err)
		
		// Now second lock should succeed
		err = lock2.Acquire(context.Background())
		assert.NoError(t, err)
		
		// Clean up
		lock2.Release(context.Background())
	})

	t.Run("TestEventSignatureVerification", func(t *testing.T) {
		signatureService := NewEventSignatureService("test-secret-key")
		
		// Create an event
		event := &SignedEvent{
			EventID:   "test-event-1",
			EventType: "letter.scheduled",
			Timestamp: time.Now().Unix(),
			Payload: map[string]interface{}{
				"letter_id": "test-letter-1",
				"user_id":   "test-user-1",
			},
		}
		
		// Generate signature
		signature, err := signatureService.GenerateSignature(event)
		assert.NoError(t, err)
		assert.NotEmpty(t, signature)
		
		// Verify signature
		err = signatureService.VerifySignature(event, signature)
		assert.NoError(t, err)
		
		// Test invalid signature
		err = signatureService.VerifySignature(event, "invalid-signature")
		assert.Error(t, err)
		
		// Test replay protection
		err = signatureService.VerifySignature(event, signature)
		assert.Error(t, err)
		assert.Equal(t, ErrReplayedEvent, err)
	})

	t.Run("TestSchedulerTaskRegistration", func(t *testing.T) {
		// Register FSD tasks
		err := scheduler.RegisterFSDTasks()
		assert.NoError(t, err)
		
		// Verify tasks were created
		var taskCount int64
		err = db.Model(&models.ScheduledTask{}).Count(&taskCount).Error
		require.NoError(t, err)
		assert.GreaterOrEqual(t, taskCount, int64(5)) // At least 5 FSD tasks
		
		// Verify specific task exists
		var futureLetterTask models.ScheduledTask
		err = db.Where("task_type = ?", "future_letter_unlock").First(&futureLetterTask).Error
		assert.NoError(t, err)
		assert.Equal(t, "*/10 * * * *", futureLetterTask.CronExpression)
	})

	t.Run("TestConcurrentTaskExecution", func(t *testing.T) {
		// Create a task
		task := &models.ScheduledTask{
			ID:             "concurrent-test-task",
			Name:           "Concurrent Test",
			TaskType:       models.TaskTypeSystemMaintenance,
			Status:         models.TaskStatusPending,
			CronExpression: "* * * * *",
			IsActive:       true,
			TimeoutSecs:    30,
		}
		err := db.Create(task).Error
		require.NoError(t, err)

		// Simulate concurrent execution
		done := make(chan bool, 2)
		errors := make(chan error, 2)
		
		for i := 0; i < 2; i++ {
			go func(instance int) {
				// Each instance tries to execute the same task
				scheduler.executeTaskWithLock(task)
				done <- true
			}(i)
		}
		
		// Wait for both to complete
		for i := 0; i < 2; i++ {
			select {
			case <-done:
				// Success
			case err := <-errors:
				t.Errorf("Concurrent execution error: %v", err)
			case <-time.After(5 * time.Second):
				t.Error("Timeout waiting for concurrent execution")
			}
		}
		
		// Verify only one execution succeeded
		var executions []models.TaskExecution
		err = db.Where("task_id = ?", task.ID).Find(&executions).Error
		require.NoError(t, err)
		
		// Due to distributed locking, only one should succeed
		successCount := 0
		for _, exec := range executions {
			if exec.Status == models.TaskStatusCompleted {
				successCount++
			}
		}
		assert.LessOrEqual(t, successCount, 1, "Only one execution should succeed due to locking")
	})
}

// TestPerformanceMetrics tests the performance of the scheduler system
func TestPerformanceMetrics(t *testing.T) {
	// This test would measure:
	// 1. Task execution latency
	// 2. Lock acquisition time
	// 3. Signature verification speed
	// 4. Database query performance
	
	t.Skip("Performance tests should be run separately")
}

// BenchmarkFutureLetterProcessing benchmarks the future letter processing
func BenchmarkFutureLetterProcessing(b *testing.B) {
	// Setup
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Letter{})
	
	letterService := &LetterService{db: db}
	notificationService := &NotificationService{db: db}
	futureLetterService := NewFutureLetterService(db, letterService, notificationService)
	
	// Create test letters
	for i := 0; i < 1000; i++ {
		letter := &models.Letter{
			ID:          fmt.Sprintf("bench-letter-%d", i),
			UserID:      fmt.Sprintf("user-%d", i%100),
			Title:       "Benchmark Letter",
			Status:      "scheduled",
			ScheduledAt: &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
		}
		db.Create(letter)
	}
	
	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		futureLetterService.ProcessScheduledLetters(context.Background())
	}
}