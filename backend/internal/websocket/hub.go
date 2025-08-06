package websocket

import (
	"log"
	"sync"
	"time"
)

// RoomOperation 房间操作
type RoomOperation struct {
	Client *Client
	Room   string
	Action string // join, leave
}

// Hub WebSocket连接管理中心
type Hub struct {
	// 已注册的客户端
	clients map[*Client]bool

	// 房间到客户端的映射
	rooms map[string]map[*Client]bool

	// 广播消息通道
	broadcast chan *Message

	// 注册客户端通道
	register chan *Client

	// 注销客户端通道
	unregister chan *Client

	// 房间加入通道
	joinRoom chan *RoomOperation

	// 房间离开通道
	leaveRoom chan *RoomOperation

	// 定向消息通道
	directMessage chan *DirectMessage

	// 房间广播通道
	roomBroadcast chan *RoomMessage

	// 统计信息
	stats *HubStats

	// 互斥锁
	mutex sync.RWMutex

	// 是否运行中
	running bool

	// 消息历史（可选的内存缓存）
	messageHistory []StoredMessage
	maxHistorySize int
}

// DirectMessage 定向消息
type DirectMessage struct {
	TargetUserID string
	Message      *Message
}

// RoomMessage 房间消息
type RoomMessage struct {
	Room          string
	Message       *Message
	ExcludeClient *Client // 可选：排除特定客户端
}

// HubStats Hub统计信息
type HubStats struct {
	TotalConnections    int64            `json:"total_connections"`
	ActiveConnections   int              `json:"active_connections"`
	TotalMessages       int64            `json:"total_messages"`
	MessagesByType      map[string]int64 `json:"messages_by_type"`
	RoomStats           map[string]int   `json:"room_stats"`
	StartTime           time.Time        `json:"start_time"`
	LastActivity        time.Time        `json:"last_activity"`
	ConnectionsByRole   map[string]int   `json:"connections_by_role"`
	ConnectionsBySchool map[string]int   `json:"connections_by_school"`
}

// StoredMessage 存储的消息（用于历史记录）
type StoredMessage struct {
	Message   *Message  `json:"message"`
	Room      string    `json:"room,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		broadcast:     make(chan *Message, 1000),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		joinRoom:      make(chan *RoomOperation),
		leaveRoom:     make(chan *RoomOperation),
		directMessage: make(chan *DirectMessage, 500),
		roomBroadcast: make(chan *RoomMessage, 500),
		stats: &HubStats{
			MessagesByType:      make(map[string]int64),
			RoomStats:           make(map[string]int),
			ConnectionsByRole:   make(map[string]int),
			ConnectionsBySchool: make(map[string]int),
			StartTime:           time.Now(),
			LastActivity:        time.Now(),
		},
		messageHistory: make([]StoredMessage, 0),
		maxHistorySize: 1000, // 保留最近1000条消息
		running:        false,
	}
}

// Run 启动Hub
func (h *Hub) Run() {
	h.running = true
	log.Println("WebSocket Hub started")

	// 启动定时清理任务
	go h.cleanupTask()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case directMsg := <-h.directMessage:
			h.sendDirectMessage(directMsg)

		case roomMsg := <-h.roomBroadcast:
			h.sendRoomMessage(roomMsg)

		case roomOp := <-h.joinRoom:
			h.handleJoinRoom(roomOp)

		case roomOp := <-h.leaveRoom:
			h.handleLeaveRoom(roomOp)
		}
	}
}

// registerClient 注册客户端
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true
	h.stats.TotalConnections++
	h.stats.ActiveConnections++
	h.stats.LastActivity = time.Now()

	// 更新角色统计
	h.stats.ConnectionsByRole[client.user.Role.String()]++

	// 更新学校统计
	h.stats.ConnectionsBySchool[client.user.SchoolCode]++

	log.Printf("Client registered: %s (%s)", client.user.Username, client.connectionInfo.ID)

	// 发送连接成功消息
	welcomeMessage := NewMessage(EventConnected, map[string]interface{}{
		"connection_id":   client.connectionInfo.ID,
		"server_time":     time.Now(),
		"welcome_message": "Welcome to OpenPenPal Real-time System",
		"user_info": map[string]interface{}{
			"user_id":     client.user.ID,
			"username":    client.user.Username,
			"role":        client.user.Role,
			"school_code": client.user.SchoolCode,
		},
	})

	client.SendMessage(welcomeMessage)

	// 通知其他客户端用户上线
	h.broadcastUserPresence(client, EventUserOnline)
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		h.stats.ActiveConnections--
		h.stats.LastActivity = time.Now()

		// 更新角色统计
		h.stats.ConnectionsByRole[client.user.Role.String()]--
		if h.stats.ConnectionsByRole[client.user.Role.String()] <= 0 {
			delete(h.stats.ConnectionsByRole, client.user.Role.String())
		}

		// 更新学校统计
		h.stats.ConnectionsBySchool[client.user.SchoolCode]--
		if h.stats.ConnectionsBySchool[client.user.SchoolCode] <= 0 {
			delete(h.stats.ConnectionsBySchool, client.user.SchoolCode)
		}

		// 从所有房间移除
		for room := range client.rooms {
			h.removeClientFromRoom(client, room)
		}

		close(client.send)
		log.Printf("Client unregistered: %s (%s)", client.user.Username, client.connectionInfo.ID)

		// 通知其他客户端用户离线
		h.broadcastUserPresence(client, EventUserOffline)
	}
}

// broadcastMessage 广播消息到所有客户端
func (h *Hub) broadcastMessage(message *Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	h.updateMessageStats(message)
	h.storeMessage(message, "")

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			h.cleanupClient(client)
		}
	}
}

// sendDirectMessage 发送定向消息
func (h *Hub) sendDirectMessage(directMsg *DirectMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	h.updateMessageStats(directMsg.Message)

	for client := range h.clients {
		if client.user.ID == directMsg.TargetUserID {
			select {
			case client.send <- directMsg.Message:
			default:
				h.cleanupClient(client)
			}
			break
		}
	}
}

// sendRoomMessage 发送房间消息
func (h *Hub) sendRoomMessage(roomMsg *RoomMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	h.updateMessageStats(roomMsg.Message)
	h.storeMessage(roomMsg.Message, roomMsg.Room)

	if roomClients, exists := h.rooms[roomMsg.Room]; exists {
		for client := range roomClients {
			// 排除指定客户端
			if roomMsg.ExcludeClient != nil && client == roomMsg.ExcludeClient {
				continue
			}

			select {
			case client.send <- roomMsg.Message:
			default:
				h.cleanupClient(client)
			}
		}
	}
}

// handleJoinRoom 处理加入房间
func (h *Hub) handleJoinRoom(roomOp *RoomOperation) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.rooms[roomOp.Room] == nil {
		h.rooms[roomOp.Room] = make(map[*Client]bool)
	}

	h.rooms[roomOp.Room][roomOp.Client] = true
	h.stats.RoomStats[roomOp.Room]++

	log.Printf("Client %s joined room %s", roomOp.Client.user.Username, roomOp.Room)
}

// handleLeaveRoom 处理离开房间
func (h *Hub) handleLeaveRoom(roomOp *RoomOperation) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.removeClientFromRoom(roomOp.Client, roomOp.Room)
	log.Printf("Client %s left room %s", roomOp.Client.user.Username, roomOp.Room)
}

// removeClientFromRoom 从房间移除客户端
func (h *Hub) removeClientFromRoom(client *Client, room string) {
	if roomClients, exists := h.rooms[room]; exists {
		delete(roomClients, client)
		h.stats.RoomStats[room]--

		if h.stats.RoomStats[room] <= 0 {
			delete(h.stats.RoomStats, room)
			delete(h.rooms, room)
		}
	}
}

// broadcastUserPresence 广播用户在线状态
func (h *Hub) broadcastUserPresence(client *Client, eventType EventType) {
	presenceData := &UserPresenceData{
		UserID:   client.user.ID,
		Username: client.user.Username,
		Status:   "online",
		LastSeen: time.Now(),
	}

	if eventType == EventUserOffline {
		presenceData.Status = "offline"
	}

	message := NewMessage(eventType, map[string]interface{}{
		"user": presenceData,
	})

	// 广播到相关房间
	for room := range client.rooms {
		h.roomBroadcast <- &RoomMessage{
			Room:          room,
			Message:       message,
			ExcludeClient: client,
		}
	}
}

// cleanupClient 清理客户端
func (h *Hub) cleanupClient(client *Client) {
	close(client.send)
	delete(h.clients, client)
	h.stats.ActiveConnections--

	// 从所有房间移除
	for room := range client.rooms {
		h.removeClientFromRoom(client, room)
	}
}

// updateMessageStats 更新消息统计
func (h *Hub) updateMessageStats(message *Message) {
	h.stats.TotalMessages++
	h.stats.MessagesByType[string(message.Type)]++
	h.stats.LastActivity = time.Now()
}

// storeMessage 存储消息到历史记录
func (h *Hub) storeMessage(message *Message, room string) {
	storedMsg := StoredMessage{
		Message:   message,
		Room:      room,
		Timestamp: time.Now(),
	}

	h.messageHistory = append(h.messageHistory, storedMsg)

	// 保持历史记录大小限制
	if len(h.messageHistory) > h.maxHistorySize {
		h.messageHistory = h.messageHistory[1:]
	}
}

// cleanupTask 定时清理任务
func (h *Hub) cleanupTask() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !h.running {
			break
		}

		h.mutex.Lock()
		now := time.Now()

		// 清理非活跃连接（超过5分钟无活动）
		for client := range h.clients {
			if now.Sub(client.lastActivity) > 5*time.Minute {
				log.Printf("Cleaning up inactive client: %s", client.user.Username)
				h.cleanupClient(client)
			}
		}

		h.mutex.Unlock()
	}
}

// BroadcastToAll 广播消息到所有客户端
func (h *Hub) BroadcastToAll(message *Message) {
	h.broadcast <- message
}

// BroadcastToUser 广播消息到特定用户
func (h *Hub) BroadcastToUser(userID string, message *Message) {
	h.directMessage <- &DirectMessage{
		TargetUserID: userID,
		Message:      message,
	}
}

// broadcastToRoom 广播消息到特定房间
func (h *Hub) broadcastToRoom(room string, message *Message) {
	h.roomBroadcast <- &RoomMessage{
		Room:    room,
		Message: message,
	}
}

// BroadcastToRoom 公开方法：广播消息到特定房间
func (h *Hub) BroadcastToRoom(room string, message *Message) {
	h.broadcastToRoom(room, message)
}

// GetStats 获取Hub统计信息
func (h *Hub) GetStats() *HubStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.stats
}

// GetConnectedUsers 获取已连接用户列表
func (h *Hub) GetConnectedUsers() []*ConnectionInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	connections := make([]*ConnectionInfo, 0, len(h.clients))
	for client := range h.clients {
		connections = append(connections, client.GetConnectionInfo())
	}

	return connections
}

// GetRoomUsers 获取房间内用户列表
func (h *Hub) GetRoomUsers(room string) []*ConnectionInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var connections []*ConnectionInfo
	if roomClients, exists := h.rooms[room]; exists {
		connections = make([]*ConnectionInfo, 0, len(roomClients))
		for client := range roomClients {
			connections = append(connections, client.GetConnectionInfo())
		}
	}

	return connections
}

// GetMessageHistory 获取消息历史
func (h *Hub) GetMessageHistory(filter *MessageFilter) []StoredMessage {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if filter == nil {
		return h.messageHistory
	}

	var filtered []StoredMessage
	for _, msg := range h.messageHistory {
		if h.matchesFilter(&msg, filter) {
			filtered = append(filtered, msg)
		}
	}

	return filtered
}

// matchesFilter 检查消息是否匹配过滤条件
func (h *Hub) matchesFilter(msg *StoredMessage, filter *MessageFilter) bool {
	if filter.UserID != "" && msg.Message.UserID != filter.UserID {
		return false
	}

	if filter.EventType != "" && msg.Message.Type != filter.EventType {
		return false
	}

	if filter.Room != "" && msg.Room != filter.Room {
		return false
	}

	if filter.Since != nil && msg.Timestamp.Before(*filter.Since) {
		return false
	}

	return true
}

// Stop 停止Hub
func (h *Hub) Stop() {
	h.running = false
	log.Println("WebSocket Hub stopped")
}
