package models_test

import (
	"courier-service/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TaskTestSuite struct {
	suite.Suite
}

func TestTaskTestSuite(t *testing.T) {
	suite.Run(t, new(TaskTestSuite))
}

func (suite *TaskTestSuite) TestTaskCanTransitionTo() {
	tests := []struct {
		name         string
		currentStatus string
		targetStatus string
		expected     bool
	}{
		{"Available to Accepted", models.TaskStatusAvailable, models.TaskStatusAccepted, true},
		{"Available to Canceled", models.TaskStatusAvailable, models.TaskStatusCanceled, true},
		{"Available to Collected", models.TaskStatusAvailable, models.TaskStatusCollected, false},
		{"Accepted to Collected", models.TaskStatusAccepted, models.TaskStatusCollected, true},
		{"Accepted to Canceled", models.TaskStatusAccepted, models.TaskStatusCanceled, true},
		{"Collected to InTransit", models.TaskStatusCollected, models.TaskStatusInTransit, true},
		{"Collected to Delivered", models.TaskStatusCollected, models.TaskStatusDelivered, false},
		{"InTransit to Delivered", models.TaskStatusInTransit, models.TaskStatusDelivered, true},
		{"InTransit to Failed", models.TaskStatusInTransit, models.TaskStatusFailed, true},
		{"Delivered to any", models.TaskStatusDelivered, models.TaskStatusFailed, false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			task := &models.Task{Status: tt.currentStatus}
			result := task.CanTransitionTo(tt.targetStatus)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *TaskTestSuite) TestTaskIsAssigned() {
	tests := []struct {
		name      string
		courierID *string
		expected  bool
	}{
		{"Assigned courier", stringPtr("courier123"), true},
		{"Nil courier", nil, false},
		{"Empty courier", stringPtr(""), false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			task := &models.Task{CourierID: tt.courierID}
			result := task.IsAssigned()
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *TaskTestSuite) TestTaskIsCompleted() {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"Delivered task", models.TaskStatusDelivered, true},
		{"Available task", models.TaskStatusAvailable, false},
		{"Failed task", models.TaskStatusFailed, false},
		{"InTransit task", models.TaskStatusInTransit, false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			task := &models.Task{Status: tt.status}
			result := task.IsCompleted()
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *TaskTestSuite) TestTaskIsActive() {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"Available task", models.TaskStatusAvailable, true},
		{"Accepted task", models.TaskStatusAccepted, true},
		{"Collected task", models.TaskStatusCollected, true},
		{"InTransit task", models.TaskStatusInTransit, true},
		{"Delivered task", models.TaskStatusDelivered, false},
		{"Failed task", models.TaskStatusFailed, false},
		{"Canceled task", models.TaskStatusCanceled, false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			task := &models.Task{Status: tt.status}
			result := task.IsActive()
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *TaskTestSuite) TestTaskOPCodeIntegration() {
	suite.Run("Task with valid OP codes", func() {
		task := &models.Task{
			TaskID:         "task123",
			LetterID:       "letter123",
			PickupOPCode:   "PK5F3D",
			DeliveryOPCode: "PK5G2A",
			CurrentOPCode:  "PK5F3D",
			Status:         models.TaskStatusAvailable,
		}

		assert.Equal(suite.T(), "PK5F3D", task.PickupOPCode)
		assert.Equal(suite.T(), "PK5G2A", task.DeliveryOPCode)
		assert.Equal(suite.T(), "PK5F3D", task.CurrentOPCode)
		assert.True(suite.T(), task.IsActive())
	})

	suite.Run("Task OP code format validation", func() {
		// Test 6-character OP code format
		task := &models.Task{
			PickupOPCode:   "PK5F3D", // Valid: 6 chars
			DeliveryOPCode: "QH2A1B", // Valid: 6 chars
		}

		assert.Len(suite.T(), task.PickupOPCode, 6)
		assert.Len(suite.T(), task.DeliveryOPCode, 6)
	})
}

func (suite *TaskTestSuite) TestTaskLifecycle() {
	suite.Run("Complete task lifecycle", func() {
		now := time.Now()
		task := &models.Task{
			TaskID:           "task123",
			LetterID:         "letter123",
			Status:           models.TaskStatusAvailable,
			Priority:         models.TaskPriorityNormal,
			PickupLocation:   "北京大学",
			DeliveryLocation: "清华大学",
			CreatedAt:        now,
		}

		// Initially available and active
		assert.True(suite.T(), task.IsActive())
		assert.False(suite.T(), task.IsAssigned())
		assert.False(suite.T(), task.IsCompleted())

		// Can be accepted
		assert.True(suite.T(), task.CanTransitionTo(models.TaskStatusAccepted))

		// Accept the task
		courierID := "courier123"
		task.CourierID = &courierID
		task.Status = models.TaskStatusAccepted
		acceptedAt := now.Add(1 * time.Hour)
		task.AcceptedAt = &acceptedAt

		assert.True(suite.T(), task.IsAssigned())
		assert.True(suite.T(), task.IsActive())
		assert.False(suite.T(), task.IsCompleted())

		// Can be collected
		assert.True(suite.T(), task.CanTransitionTo(models.TaskStatusCollected))

		// Collect the task
		task.Status = models.TaskStatusCollected

		// Can go in transit
		assert.True(suite.T(), task.CanTransitionTo(models.TaskStatusInTransit))

		// In transit
		task.Status = models.TaskStatusInTransit

		// Can be delivered
		assert.True(suite.T(), task.CanTransitionTo(models.TaskStatusDelivered))

		// Deliver the task
		task.Status = models.TaskStatusDelivered
		completedAt := now.Add(3 * time.Hour)
		task.CompletedAt = &completedAt

		assert.True(suite.T(), task.IsCompleted())
		assert.False(suite.T(), task.IsActive())
	})
}

func (suite *TaskTestSuite) TestTaskPriorityConstants() {
	assert.Equal(suite.T(), "normal", models.TaskPriorityNormal)
	assert.Equal(suite.T(), "urgent", models.TaskPriorityUrgent)
	assert.Equal(suite.T(), "express", models.TaskPriorityExpress)
}

func (suite *TaskTestSuite) TestTaskStatusConstants() {
	assert.Equal(suite.T(), "available", models.TaskStatusAvailable)
	assert.Equal(suite.T(), "accepted", models.TaskStatusAccepted)
	assert.Equal(suite.T(), "collected", models.TaskStatusCollected)
	assert.Equal(suite.T(), "in_transit", models.TaskStatusInTransit)
	assert.Equal(suite.T(), "delivered", models.TaskStatusDelivered)
	assert.Equal(suite.T(), "failed", models.TaskStatusFailed)
	assert.Equal(suite.T(), "canceled", models.TaskStatusCanceled)
}

func (suite *TaskTestSuite) TestValidTaskTransitions() {
	// Test the ValidTaskTransitions map
	availableTransitions := models.ValidTaskTransitions[models.TaskStatusAvailable]
	assert.Contains(suite.T(), availableTransitions, models.TaskStatusAccepted)
	assert.Contains(suite.T(), availableTransitions, models.TaskStatusCanceled)

	acceptedTransitions := models.ValidTaskTransitions[models.TaskStatusAccepted]
	assert.Contains(suite.T(), acceptedTransitions, models.TaskStatusCollected)
	assert.Contains(suite.T(), acceptedTransitions, models.TaskStatusCanceled)

	collectedTransitions := models.ValidTaskTransitions[models.TaskStatusCollected]
	assert.Contains(suite.T(), collectedTransitions, models.TaskStatusInTransit)

	inTransitTransitions := models.ValidTaskTransitions[models.TaskStatusInTransit]
	assert.Contains(suite.T(), inTransitTransitions, models.TaskStatusDelivered)
	assert.Contains(suite.T(), inTransitTransitions, models.TaskStatusFailed)

	// Terminal states should not have transitions
	_, exists := models.ValidTaskTransitions[models.TaskStatusDelivered]
	assert.False(suite.T(), exists)
	_, exists = models.ValidTaskTransitions[models.TaskStatusFailed]
	assert.False(suite.T(), exists)
	_, exists = models.ValidTaskTransitions[models.TaskStatusCanceled]
	assert.False(suite.T(), exists)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}