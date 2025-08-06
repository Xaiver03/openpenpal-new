package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"openpenpal-backend/internal/models"
)

const (
	// 写入等待时间
	writeWait = 10 * time.Second

	// Pong等待时间
	pongWait = 60 * time.Second

	// Ping发送间隔，必须小于pongWait
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境中应该检查Origin
		return true
	},
}

// Client WebSocket客户端
type Client struct {
	// WebSocket连接
	conn *websocket.Conn

	// Hub引用
	hub *Hub

	// 用户信息
	user *models.User

	// 连接信息
	connectionInfo *ConnectionInfo

	// 发送消息的缓冲通道
	send chan *Message

	// 客户端所在的房间
	rooms map[string]bool

	// 最后活动时间
	lastActivity time.Time

	// 客户端状态
	isActive bool
}

// NewClient 创建新的WebSocket客户端
func NewClient(conn *websocket.Conn, hub *Hub, user *models.User, r *http.Request) *Client {
	connectionInfo := &ConnectionInfo{
		ID:           generateConnectionID(),
		UserID:       user.ID,
		Username:     user.Username,
		Role:         string(user.Role),
		SchoolCode:   user.SchoolCode,
		IPAddress:    getClientIP(r),
		UserAgent:    r.UserAgent(),
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		Rooms:        []string{},
	}

	client := &Client{
		conn:           conn,
		hub:            hub,
		user:           user,
		connectionInfo: connectionInfo,
		send:           make(chan *Message, 256),
		rooms:          make(map[string]bool),
		lastActivity:   time.Now(),
		isActive:       true,
	}

	// 自动加入默认房间
	client.joinDefaultRooms()

	return client
}

// joinDefaultRooms 加入默认房间
func (c *Client) joinDefaultRooms() {
	// 加入全局房间
	c.JoinRoom(string(RoomGlobal))

	// 加入学校房间
	c.JoinRoom(GetSchoolRoom(c.user.SchoolCode))

	// 加入角色房间
	switch c.user.Role {
	case models.RoleCourier, models.RoleSeniorCourier, models.RoleCourierCoordinator:
		c.JoinRoom(string(RoomCouriers))
	case models.RoleSchoolAdmin, models.RolePlatformAdmin, models.RoleSuperAdmin:
		c.JoinRoom(string(RoomAdmins))
	default:
		c.JoinRoom(string(RoomUsers))
	}

	// 加入个人房间
	c.JoinRoom(GetUserRoom(c.user.ID))
}

// ReadPump 读取消息泵
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.updateActivity()
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.updateActivity()
		c.handleMessage(message)
	}
}

// WritePump 写入消息泵
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.writeMessage(message); err != nil {
				log.Printf("WriteMessage error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// writeMessage 写入消息
func (c *Client) writeMessage(message *Message) error {
	data, err := message.ToJSON()
	if err != nil {
		return err
	}

	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(data []byte) {
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		log.Printf("JSON unmarshal error: %v", err)
		c.sendError("INVALID_MESSAGE", "Invalid message format")
		return
	}

	// 设置消息发送者
	message.UserID = c.user.ID

	// 处理不同类型的消息
	switch message.Type {
	case EventHeartbeat:
		c.handleHeartbeat(&message)
	case EventCourierLocationUpdate:
		c.handleLocationUpdate(&message)
	case EventUserOnline:
		c.handleUserOnline(&message)
	default:
		// 转发到Hub处理
		c.hub.broadcast <- &message
	}
}

// handleHeartbeat 处理心跳消息
func (c *Client) handleHeartbeat(message *Message) {
	response := NewMessage(EventHeartbeat, map[string]interface{}{
		"server_time": time.Now(),
		"client_time": message.Data["client_time"],
	})

	select {
	case c.send <- response:
	default:
		close(c.send)
	}
}

// handleLocationUpdate 处理位置更新
func (c *Client) handleLocationUpdate(message *Message) {
	// 验证用户是否有信使权限
	if !c.user.HasRole(models.RoleCourier) {
		c.sendError("PERMISSION_DENIED", "Only couriers can update location")
		return
	}

	// 广播位置更新到相关房间
	c.hub.broadcastToRoom(GetSchoolRoom(c.user.SchoolCode), message)
	c.hub.broadcastToRoom(string(RoomCouriers), message)
}

// handleUserOnline 处理用户上线
func (c *Client) handleUserOnline(message *Message) {
	presenceData := &UserPresenceData{
		UserID:   c.user.ID,
		Username: c.user.Username,
		Status:   "online",
		LastSeen: time.Now(),
	}

	onlineMessage := NewMessage(EventUserOnline, map[string]interface{}{
		"user":      presenceData,
		"room_info": c.getRoomInfo(),
	})

	// 广播到相关房间
	for room := range c.rooms {
		c.hub.broadcastToRoom(room, onlineMessage)
	}
}

// sendError 发送错误消息
func (c *Client) sendError(code, message string) {
	errorMsg := NewMessage(EventError, map[string]interface{}{
		"code":    code,
		"message": message,
	})

	select {
	case c.send <- errorMsg:
	default:
		close(c.send)
	}
}

// JoinRoom 加入房间
func (c *Client) JoinRoom(room string) {
	c.rooms[room] = true
	c.connectionInfo.Rooms = append(c.connectionInfo.Rooms, room)

	// 通知Hub
	c.hub.joinRoom <- &RoomOperation{
		Client: c,
		Room:   room,
		Action: "join",
	}
}

// LeaveRoom 离开房间
func (c *Client) LeaveRoom(room string) {
	delete(c.rooms, room)

	// 从连接信息中移除
	for i, r := range c.connectionInfo.Rooms {
		if r == room {
			c.connectionInfo.Rooms = append(c.connectionInfo.Rooms[:i], c.connectionInfo.Rooms[i+1:]...)
			break
		}
	}

	// 通知Hub
	c.hub.leaveRoom <- &RoomOperation{
		Client: c,
		Room:   room,
		Action: "leave",
	}
}

// SendMessage 发送消息给客户端
func (c *Client) SendMessage(message *Message) {
	if !c.isActive {
		return
	}

	select {
	case c.send <- message:
	default:
		close(c.send)
		c.isActive = false
	}
}

// IsInRoom 检查是否在指定房间
func (c *Client) IsInRoom(room string) bool {
	return c.rooms[room]
}

// GetConnectionInfo 获取连接信息
func (c *Client) GetConnectionInfo() *ConnectionInfo {
	c.connectionInfo.LastActivity = c.lastActivity
	return c.connectionInfo
}

// updateActivity 更新活动时间
func (c *Client) updateActivity() {
	c.lastActivity = time.Now()
	c.connectionInfo.LastActivity = c.lastActivity
}

// getRoomInfo 获取房间信息
func (c *Client) getRoomInfo() map[string]interface{} {
	rooms := make([]string, 0, len(c.rooms))
	for room := range c.rooms {
		rooms = append(rooms, room)
	}

	return map[string]interface{}{
		"rooms":      rooms,
		"room_count": len(rooms),
	}
}

// Close 关闭客户端连接
func (c *Client) Close() {
	c.isActive = false
	close(c.send)
	c.conn.Close()
}

// generateConnectionID 生成连接ID
func generateConnectionID() string {
	return "conn_" + time.Now().Format("20060102150405") + "_" + randomString(12)
}

// getClientIP 获取客户端IP地址
func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}
