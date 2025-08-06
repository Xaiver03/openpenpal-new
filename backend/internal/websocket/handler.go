package websocket

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"openpenpal-backend/internal/models"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	hub *Hub
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(hub *Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocketConnection 处理WebSocket连接
func (h *WebSocketHandler) HandleWebSocketConnection(c *gin.Context) {
	// 获取用户信息（应该在认证中间件中设置）
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户信息格式错误"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket升级失败"})
		return
	}

	// 创建客户端
	client := NewClient(conn, h.hub, user, c.Request)

	// 先启动读写协程
	go client.WritePump()
	go client.ReadPump()

	// 给WritePump一点时间启动
	time.Sleep(10 * time.Millisecond)

	// 然后注册客户端（这会发送欢迎消息）
	h.hub.register <- client

	log.Printf("WebSocket connection established for user: %s", user.Username)
}

// HandleGetConnections 获取连接信息
func (h *WebSocketHandler) HandleGetConnections(c *gin.Context) {
	connections := h.hub.GetConnectedUsers()
	c.JSON(http.StatusOK, gin.H{
		"connections": connections,
		"count":       len(connections),
	})
}

// HandleGetStats 获取Hub统计信息
func (h *WebSocketHandler) HandleGetStats(c *gin.Context) {
	// 确保hub不为空
	if h.hub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebSocket service not initialized",
		})
		return
	}

	stats := h.hub.GetStats()
	if stats == nil {
		// 如果stats为空，返回默认值
		c.JSON(http.StatusOK, gin.H{
			"total_connections":     0,
			"active_connections":    0,
			"total_messages":        0,
			"messages_by_type":      map[string]int64{},
			"room_stats":            map[string]int{},
			"start_time":            time.Now().Format(time.RFC3339),
			"last_activity":         time.Now().Format(time.RFC3339),
			"connections_by_role":   map[string]int{},
			"connections_by_school": map[string]int{},
		})
		return
	}
	
	// 安全地转换map类型，确保不会返回nil
	messagesByType := make(map[string]int64)
	if stats.MessagesByType != nil {
		messagesByType = stats.MessagesByType
	}
	
	roomStats := make(map[string]int)
	if stats.RoomStats != nil {
		roomStats = stats.RoomStats
	}
	
	connectionsByRole := make(map[string]int)
	if stats.ConnectionsByRole != nil {
		connectionsByRole = stats.ConnectionsByRole
	}
	
	connectionsBySchool := make(map[string]int)
	if stats.ConnectionsBySchool != nil {
		connectionsBySchool = stats.ConnectionsBySchool
	}
	
	// 确保时间字段正确序列化
	c.JSON(http.StatusOK, gin.H{
		"total_connections":     stats.TotalConnections,
		"active_connections":    stats.ActiveConnections,
		"total_messages":        stats.TotalMessages,
		"messages_by_type":      messagesByType,
		"room_stats":            roomStats,
		"start_time":            stats.StartTime.Format(time.RFC3339),
		"last_activity":         stats.LastActivity.Format(time.RFC3339),
		"connections_by_role":   connectionsByRole,
		"connections_by_school": connectionsBySchool,
	})
}

// HandleGetRoomUsers 获取房间用户列表
func (h *WebSocketHandler) HandleGetRoomUsers(c *gin.Context) {
	room := c.Param("room")
	if room == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "房间名不能为空"})
		return
	}

	users := h.hub.GetRoomUsers(room)
	c.JSON(http.StatusOK, gin.H{
		"room":  room,
		"users": users,
		"count": len(users),
	})
}

// HandleBroadcastMessage 广播消息
func (h *WebSocketHandler) HandleBroadcastMessage(c *gin.Context) {
	var req struct {
		Type EventType              `json:"type" binding:"required"`
		Data map[string]interface{} `json:"data" binding:"required"`
		Room string                 `json:"room,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取发送者信息
	userInterface, _ := c.Get("user")
	user := userInterface.(*models.User)

	// 创建消息
	message := NewMessage(req.Type, req.Data)
	message.UserID = user.ID

	// 根据是否指定房间进行广播
	if req.Room != "" {
		h.hub.BroadcastToRoom(req.Room, message)
	} else {
		h.hub.BroadcastToAll(message)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message_id": message.ID,
		"timestamp":  message.Timestamp,
	})
}

// HandleSendDirectMessage 发送定向消息
func (h *WebSocketHandler) HandleSendDirectMessage(c *gin.Context) {
	var req struct {
		TargetUserID string                 `json:"target_user_id" binding:"required"`
		Type         EventType              `json:"type" binding:"required"`
		Data         map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取发送者信息
	userInterface, _ := c.Get("user")
	user := userInterface.(*models.User)

	// 创建消息
	message := NewMessage(req.Type, req.Data)
	message.UserID = user.ID

	// 发送定向消息
	h.hub.BroadcastToUser(req.TargetUserID, message)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message_id":  message.ID,
		"target_user": req.TargetUserID,
		"timestamp":   message.Timestamp,
	})
}

// HandleGetMessageHistory 获取消息历史
func (h *WebSocketHandler) HandleGetMessageHistory(c *gin.Context) {
	filter := &MessageFilter{}

	// 解析查询参数
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}

	if eventType := c.Query("event_type"); eventType != "" {
		filter.EventType = EventType(eventType)
	}

	if room := c.Query("room"); room != "" {
		filter.Room = room
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	// 获取消息历史
	messages := h.hub.GetMessageHistory(filter)

	// 应用限制
	if filter.Limit > 0 && len(messages) > filter.Limit {
		messages = messages[len(messages)-filter.Limit:]
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
		"filter":   filter,
	})
}

// Remove duplicate WebSocketService definition from handler.go
// This is now defined in service.go
