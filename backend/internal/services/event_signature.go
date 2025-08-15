package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	
	"openpenpal-backend/internal/models"
)

var (
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrExpiredSignature   = errors.New("signature expired")
	ErrInvalidPayload     = errors.New("invalid payload")
	ErrMissingSignature   = errors.New("missing signature")
	ErrReplayedEvent      = errors.New("event already processed")
)

// EventSignatureService handles event signature verification for secure task triggering
type EventSignatureService struct {
	secretKey          string
	signatureHeader    string
	maxClockSkew       time.Duration
	replayProtection   bool
	processedEvents    map[string]time.Time // In production, use Redis
}

// NewEventSignatureService creates a new event signature service
func NewEventSignatureService(secretKey string) *EventSignatureService {
	if secretKey == "" {
		// Generate a default secret key (in production, this should come from config)
		secretKey = generateSecretKey()
		log.Printf("[EventSignature] WARNING: Using generated secret key. Configure a permanent key for production!")
	}
	
	return &EventSignatureService{
		secretKey:          secretKey,
		signatureHeader:    "X-OpenPenPal-Signature",
		maxClockSkew:       5 * time.Minute,
		replayProtection:   true,
		processedEvents:    make(map[string]time.Time),
	}
}

// SignedEvent represents an event with signature verification
type SignedEvent struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	Timestamp int64                  `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	Signature string                 `json:"signature,omitempty"`
}

// GenerateSignature generates a signature for an event
func (ess *EventSignatureService) GenerateSignature(event *SignedEvent) (string, error) {
	// Ensure event has required fields
	if event.EventID == "" {
		event.EventID = generateID()
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}
	
	// Create canonical representation of the event
	canonical, err := ess.canonicalizeEvent(event)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalize event: %w", err)
	}
	
	// Generate HMAC signature
	h := hmac.New(sha256.New, []byte(ess.secretKey))
	h.Write([]byte(canonical))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// Include timestamp in signature to prevent replay attacks
	timestampedSignature := fmt.Sprintf("v1=%s,t=%d", signature, event.Timestamp)
	
	return timestampedSignature, nil
}

// VerifySignature verifies an event's signature
func (ess *EventSignatureService) VerifySignature(event *SignedEvent, providedSignature string) error {
	if providedSignature == "" {
		return ErrMissingSignature
	}
	
	// Parse the signature
	parts := strings.Split(providedSignature, ",")
	if len(parts) != 2 {
		return ErrInvalidSignature
	}
	
	var signature string
	var timestamp int64
	
	for _, part := range parts {
		if strings.HasPrefix(part, "v1=") {
			signature = strings.TrimPrefix(part, "v1=")
		} else if strings.HasPrefix(part, "t=") {
			fmt.Sscanf(part, "t=%d", &timestamp)
		}
	}
	
	if signature == "" || timestamp == 0 {
		return ErrInvalidSignature
	}
	
	// Check timestamp to prevent replay attacks
	now := time.Now().Unix()
	if abs(now-timestamp) > int64(ess.maxClockSkew.Seconds()) {
		return ErrExpiredSignature
	}
	
	// Verify the signature
	expectedSignature, err := ess.GenerateSignature(event)
	if err != nil {
		return err
	}
	
	expectedParts := strings.Split(expectedSignature, ",")
	expectedSig := strings.TrimPrefix(expectedParts[0], "v1=")
	
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return ErrInvalidSignature
	}
	
	// Check for replay attacks
	if ess.replayProtection {
		if err := ess.checkReplay(event.EventID); err != nil {
			return err
		}
	}
	
	return nil
}

// VerifyWebhook verifies a webhook request
func (ess *EventSignatureService) VerifyWebhook(body []byte, signature string) (*SignedEvent, error) {
	// Parse the event
	var event SignedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}
	
	// Verify the signature
	if err := ess.VerifySignature(&event, signature); err != nil {
		return nil, err
	}
	
	return &event, nil
}

// SignWebhookPayload signs a webhook payload for outgoing requests
func (ess *EventSignatureService) SignWebhookPayload(eventType string, payload map[string]interface{}) ([]byte, string, error) {
	event := &SignedEvent{
		EventID:   generateID(),
		EventType: eventType,
		Timestamp: time.Now().Unix(),
		Payload:   payload,
	}
	
	// Generate signature
	signature, err := ess.GenerateSignature(event)
	if err != nil {
		return nil, "", err
	}
	
	// Marshal the event
	body, err := json.Marshal(event)
	if err != nil {
		return nil, "", err
	}
	
	return body, signature, nil
}

// canonicalizeEvent creates a canonical string representation of an event
func (ess *EventSignatureService) canonicalizeEvent(event *SignedEvent) (string, error) {
	// Create a consistent representation
	canonical := map[string]interface{}{
		"event_id":   event.EventID,
		"event_type": event.EventType,
		"timestamp":  event.Timestamp,
		"payload":    event.Payload,
	}
	
	// Marshal to JSON with sorted keys
	data, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

// checkReplay checks if an event has already been processed
func (ess *EventSignatureService) checkReplay(eventID string) error {
	// In production, this should use Redis with TTL
	if _, exists := ess.processedEvents[eventID]; exists {
		return ErrReplayedEvent
	}
	
	// Mark event as processed
	ess.processedEvents[eventID] = time.Now()
	
	// Clean up old events (in production, use Redis TTL)
	ess.cleanupOldEvents()
	
	return nil
}

// cleanupOldEvents removes old processed events from memory
func (ess *EventSignatureService) cleanupOldEvents() {
	cutoff := time.Now().Add(-24 * time.Hour)
	for eventID, timestamp := range ess.processedEvents {
		if timestamp.Before(cutoff) {
			delete(ess.processedEvents, eventID)
		}
	}
}

// Middleware for HTTP handlers

// VerifySignatureMiddleware creates a Gin middleware for signature verification
func (ess *EventSignatureService) VerifySignatureMiddleware() func(c interface{}) {
	return func(c interface{}) {
		// This would be implemented based on your HTTP framework
		// For Gin framework example:
		/*
		ctx := c.(*gin.Context)
		
		// Get signature from header
		signature := ctx.GetHeader(ess.signatureHeader)
		if signature == "" {
			ctx.JSON(401, gin.H{"error": "missing signature"})
			ctx.Abort()
			return
		}
		
		// Read body
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "failed to read body"})
			ctx.Abort()
			return
		}
		
		// Restore body for downstream handlers
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		
		// Verify signature
		event, err := ess.VerifyWebhook(body, signature)
		if err != nil {
			ctx.JSON(401, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}
		
		// Store verified event in context
		ctx.Set("verified_event", event)
		ctx.Next()
		*/
	}
}

// Helper functions

func generateSecretKey() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i + 65) // Simple generation for demo
	}
	return base64.StdEncoding.EncodeToString(b)
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// EventTriggerService handles secure event-based task triggering
type EventTriggerService struct {
	schedulerService *EnhancedSchedulerService
	signatureService *EventSignatureService
}

// NewEventTriggerService creates a new event trigger service
func NewEventTriggerService(
	schedulerService *EnhancedSchedulerService,
	signatureService *EventSignatureService,
) *EventTriggerService {
	return &EventTriggerService{
		schedulerService: schedulerService,
		signatureService: signatureService,
	}
}

// TriggerTaskByEvent triggers a task execution based on a verified event
func (ets *EventTriggerService) TriggerTaskByEvent(event *SignedEvent) error {
	// Map event types to task types
	taskType, ok := eventTypeToTaskType[event.EventType]
	if !ok {
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
	
	// Find the task to execute
	var task models.ScheduledTask
	err := ets.schedulerService.db.
		Where("task_type = ? AND is_active = ?", taskType, true).
		First(&task).Error
	
	if err != nil {
		return fmt.Errorf("task not found for event type %s: %w", event.EventType, err)
	}
	
	// Execute the task with event context
	go ets.schedulerService.executeTaskWithLock(&task)
	
	log.Printf("[EventTrigger] Triggered task %s for event %s", task.Name, event.EventID)
	return nil
}

// Event type to task type mapping
var eventTypeToTaskType = map[string]models.TaskType{
	"letter.scheduled":         models.TaskType("future_letter_unlock"),
	"courier.timeout":          models.TaskType("courier_timeout_check"),
	"envelope.contest.closing": models.TaskType("envelope_contest_close"),
	"ai.reply.scheduled":       models.TaskType("ai_penpal_reply"),
	"maintenance.required":     models.TaskTypeSystemMaintenance,
}