package utils

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketEvent WebSocket事件结构
type WebSocketEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    string      `json:"user_id,omitempty"`
}

// WebSocketClient WebSocket客户端
type WebSocketClient struct {
	ID     string
	UserID string
	Conn   *websocket.Conn
	Send   chan WebSocketEvent
}

// WebSocketManager WebSocket管理器
type WebSocketManager struct {
	clients    map[string]*WebSocketClient
	broadcast  chan WebSocketEvent
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mutex      sync.RWMutex
}

// NewWebSocketManager 创建WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	manager := &WebSocketManager{
		clients:    make(map[string]*WebSocketClient),
		broadcast:  make(chan WebSocketEvent),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}

	go manager.run()
	return manager
}

// run 运行WebSocket管理器
func (manager *WebSocketManager) run() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client.ID] = client
			manager.mutex.Unlock()
			log.Printf("WebSocket client %s connected", client.ID)

		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client.ID]; ok {
				delete(manager.clients, client.ID)
				close(client.Send)
			}
			manager.mutex.Unlock()
			log.Printf("WebSocket client %s disconnected", client.ID)

		case event := <-manager.broadcast:
			manager.mutex.RLock()
			for _, client := range manager.clients {
				select {
				case client.Send <- event:
				default:
					close(client.Send)
					delete(manager.clients, client.ID)
				}
			}
			manager.mutex.RUnlock()
		}
	}
}

// RegisterClient 注册客户端
func (manager *WebSocketManager) RegisterClient(client *WebSocketClient) {
	manager.register <- client
}

// UnregisterClient 注销客户端
func (manager *WebSocketManager) UnregisterClient(client *WebSocketClient) {
	manager.unregister <- client
}

// BroadcastToUser 向特定用户广播消息
func (manager *WebSocketManager) BroadcastToUser(userID string, event WebSocketEvent) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	for _, client := range manager.clients {
		if client.UserID == userID {
			select {
			case client.Send <- event:
			default:
				close(client.Send)
				delete(manager.clients, client.ID)
			}
		}
	}
}

// BroadcastToAll 向所有用户广播消息
func (manager *WebSocketManager) BroadcastToAll(event WebSocketEvent) {
	manager.broadcast <- event
}

// BroadcastToAdmins 向管理员广播消息
func (manager *WebSocketManager) BroadcastToAdmins(event WebSocketEvent) {
	// 这里可以根据实际需求实现管理员筛选逻辑
	manager.broadcast <- event
}

// SendTaskUpdate 发送任务更新通知
func (manager *WebSocketManager) SendTaskUpdate(taskID, status, courierID string) {
	event := WebSocketEvent{
		Type: "COURIER_TASK_UPDATE",
		Data: map[string]interface{}{
			"task_id":    taskID,
			"status":     status,
			"courier_id": courierID,
		},
		Timestamp: time.Now(),
	}

	// 通知信使
	if courierID != "" {
		manager.BroadcastToUser(courierID, event)
	}

	// 通知管理员
	manager.BroadcastToAdmins(event)
}

// SendTaskAssignment 发送任务分配通知
func (manager *WebSocketManager) SendTaskAssignment(taskID, courierID string, taskData interface{}) {
	event := WebSocketEvent{
		Type: "NEW_TASK_ASSIGNMENT",
		Data: map[string]interface{}{
			"task_id":    taskID,
			"courier_id": courierID,
			"task_data":  taskData,
		},
		Timestamp: time.Now(),
	}

	manager.BroadcastToUser(courierID, event)
}

// GenerateTaskID 生成任务ID
func GenerateTaskID() string {
	return "T" + time.Now().Format("20060102150405") + RandomString(6)
}

// RandomString 生成随机字符串
func RandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
