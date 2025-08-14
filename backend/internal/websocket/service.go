package websocket

import (
	"log"
	"sync"
)

// WebSocketService WebSocket服务
type WebSocketService struct {
	hub     *Hub
	handler *WebSocketHandler
	mu      sync.RWMutex
	started bool
}

// NewWebSocketService 创建WebSocket服务
func NewWebSocketService() *WebSocketService {
	hub := NewHub()
	handler := NewWebSocketHandler(hub)

	return &WebSocketService{
		hub:     hub,
		handler: handler,
		started: false,
	}
}

// Start 启动WebSocket服务
func (s *WebSocketService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		go s.hub.Run()
		s.started = true
		log.Println("WebSocket service started")
	}
}

// GetHub 获取Hub实例
func (s *WebSocketService) GetHub() *Hub {
	return s.hub
}

// GetHandler 获取Handler实例
func (s *WebSocketService) GetHandler() *WebSocketHandler {
	return s.handler
}

// IsStarted 检查服务是否已启动
func (s *WebSocketService) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

// Stop 停止服务
func (s *WebSocketService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		s.hub.Stop()
		s.started = false
		log.Println("WebSocket service stopped")
	}
}

// BroadcastLetterStatusUpdate 广播信件状态更新
func (s *WebSocketService) BroadcastLetterStatusUpdate(data *LetterStatusUpdateData) {
	message := NewMessage(EventLetterStatusUpdate, map[string]interface{}{
		"letter_id":    data.LetterID,
		"code":         data.Code,
		"status":       data.Status,
		"location":     data.Location,
		"courier_id":   data.CourierID,
		"courier_name": data.CourierName,
		"updated_at":   data.UpdatedAt,
		"message":      data.Message,
	})

	// 广播到信件追踪房间
	s.hub.BroadcastToRoom(GetLetterRoom(data.LetterID), message)

	// 广播到全局房间
	s.hub.BroadcastToRoom(string(RoomGlobal), message)
}

// BroadcastCourierLocationUpdate 广播信使位置更新
func (s *WebSocketService) BroadcastCourierLocationUpdate(data *CourierLocationData) {
	message := NewMessage(EventCourierLocationUpdate, map[string]interface{}{
		"courier_id":   data.CourierID,
		"courier_name": data.CourierName,
		"latitude":     data.Latitude,
		"longitude":    data.Longitude,
		"accuracy":     data.Accuracy,
		"timestamp":    data.Timestamp,
		"status":       data.Status,
	})

	// 广播到信使房间
	s.hub.BroadcastToRoom(string(RoomCouriers), message)
}

// BroadcastTaskAssignment 广播任务分配
func (s *WebSocketService) BroadcastTaskAssignment(data *TaskAssignmentData) {
	message := NewMessage(EventNewTaskAssignment, map[string]interface{}{
		"task_id":           data.TaskID,
		"courier_id":        data.CourierID,
		"letter_id":         data.LetterID,
		"priority":          data.Priority,
		"deadline":          data.Deadline,
		"pickup_location":   data.PickupLocation,
		"delivery_location": data.DeliveryLocation,
		"reward":            data.Reward,
		"assigned_at":       data.AssignedAt,
	})

	// 发送给特定信使
	s.hub.BroadcastToUser(data.CourierID, message)

	// 广播到信使房间
	s.hub.BroadcastToRoom(string(RoomCouriers), message)
}

// BroadcastNotification 广播通知
func (s *WebSocketService) BroadcastNotification(userID string, data *NotificationData) {
	message := NewMessage(EventNotification, map[string]interface{}{
		"notification_id": data.NotificationID,
		"title":           data.Title,
		"content":         data.Content,
		"type":            data.Type,
		"priority":        data.Priority,
		"action_url":      data.ActionURL,
		"created_at":      data.CreatedAt,
		"expires_at":      data.ExpiresAt,
	})

	if userID != "" {
		// 发送给特定用户
		s.hub.BroadcastToUser(userID, message)
	} else {
		// 广播到所有用户
		s.hub.BroadcastToAll(message)
	}
}

// BroadcastSystemMessage 广播系统消息
func (s *WebSocketService) BroadcastSystemMessage(data *SystemMessageData) {
	message := NewMessage(EventSystemMessage, map[string]interface{}{
		"message_id": data.MessageID,
		"level":      data.Level,
		"title":      data.Title,
		"content":    data.Content,
		"created_at": data.CreatedAt,
	})

	if data.Broadcast {
		s.hub.BroadcastToAll(message)
	} else {
		s.hub.BroadcastToRoom(string(RoomSystem), message)
	}
}
