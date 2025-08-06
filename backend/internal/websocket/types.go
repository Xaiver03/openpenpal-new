package websocket

import (
	"encoding/json"
	"time"
)

// EventType 事件类型枚举
type EventType string

const (
	// 信件相关事件
	EventLetterStatusUpdate EventType = "LETTER_STATUS_UPDATE"
	EventLetterCreated      EventType = "LETTER_CREATED"
	EventLetterRead         EventType = "LETTER_READ"
	EventLetterDelivered    EventType = "LETTER_DELIVERED"

	// 信使相关事件
	EventCourierLocationUpdate EventType = "COURIER_LOCATION_UPDATE"
	EventNewTaskAssignment     EventType = "NEW_TASK_ASSIGNMENT"
	EventTaskStatusUpdate      EventType = "TASK_STATUS_UPDATE"
	EventCourierOnline         EventType = "COURIER_ONLINE"
	EventCourierOffline        EventType = "COURIER_OFFLINE"

	// 用户相关事件
	EventUserOnline   EventType = "USER_ONLINE"
	EventUserOffline  EventType = "USER_OFFLINE"
	EventNotification EventType = "NOTIFICATION"

	// 系统相关事件
	EventSystemMessage EventType = "SYSTEM_MESSAGE"
	EventHeartbeat     EventType = "HEARTBEAT"
	EventError         EventType = "ERROR"
	EventConnected     EventType = "CONNECTED"
	EventDisconnected  EventType = "DISCONNECTED"
)

// Message WebSocket消息结构
type Message struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	Room      string                 `json:"room,omitempty"`
}

// NewMessage 创建新消息
func NewMessage(eventType EventType, data map[string]interface{}) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ToJSON 转换为JSON字符串
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从JSON字符串创建消息
func FromJSON(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}

// LetterStatusUpdateData 信件状态更新数据
type LetterStatusUpdateData struct {
	LetterID    string    `json:"letter_id"`
	Code        string    `json:"code"`
	Status      string    `json:"status"`
	Location    string    `json:"location,omitempty"`
	CourierID   string    `json:"courier_id"`
	CourierName string    `json:"courier_name"`
	UpdatedAt   time.Time `json:"updated_at"`
	Message     string    `json:"message,omitempty"`
}

// CourierLocationData 信使位置数据
type CourierLocationData struct {
	CourierID   string    `json:"courier_id"`
	CourierName string    `json:"courier_name"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Accuracy    float64   `json:"accuracy"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"` // online, busy, offline
}

// TaskAssignmentData 任务分配数据
type TaskAssignmentData struct {
	TaskID           string    `json:"task_id"`
	CourierID        string    `json:"courier_id"`
	LetterID         string    `json:"letter_id"`
	Priority         string    `json:"priority"`
	Deadline         time.Time `json:"deadline"`
	PickupLocation   string    `json:"pickup_location"`
	DeliveryLocation string    `json:"delivery_location"`
	Reward           float64   `json:"reward"`
	AssignedAt       time.Time `json:"assigned_at"`
}

// NotificationData 通知数据
type NotificationData struct {
	NotificationID string     `json:"notification_id"`
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	Type           string     `json:"type"`     // info, success, warning, error
	Priority       string     `json:"priority"` // low, normal, high
	ActionURL      string     `json:"action_url,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// UserPresenceData 用户在线状态数据
type UserPresenceData struct {
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	Status     string    `json:"status"` // online, away, busy, offline
	LastSeen   time.Time `json:"last_seen"`
	DeviceInfo string    `json:"device_info,omitempty"`
}

// SystemMessageData 系统消息数据
type SystemMessageData struct {
	MessageID string    `json:"message_id"`
	Level     string    `json:"level"` // info, warning, error
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Broadcast bool      `json:"broadcast"`
	CreatedAt time.Time `json:"created_at"`
}

// HeartbeatData 心跳数据
type HeartbeatData struct {
	ServerTime time.Time `json:"server_time"`
	ClientTime time.Time `json:"client_time,omitempty"`
	Latency    int64     `json:"latency,omitempty"` // 延迟毫秒数
}

// ErrorData 错误数据
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Room 房间/频道定义
type Room string

const (
	// 全局房间
	RoomGlobal Room = "global"
	RoomSystem Room = "system"

	// 按学校分组
	RoomSchoolPrefix Room = "school:"

	// 按角色分组
	RoomCouriers Room = "couriers"
	RoomAdmins   Room = "admins"
	RoomUsers    Room = "users"

	// 个人房间
	RoomUserPrefix Room = "user:"

	// 信件追踪房间
	RoomLetterPrefix Room = "letter:"
)

// GetSchoolRoom 获取学校房间名
func GetSchoolRoom(schoolCode string) string {
	return string(RoomSchoolPrefix) + schoolCode
}

// GetUserRoom 获取用户个人房间名
func GetUserRoom(userID string) string {
	return string(RoomUserPrefix) + userID
}

// GetLetterRoom 获取信件追踪房间名
func GetLetterRoom(letterID string) string {
	return string(RoomLetterPrefix) + letterID
}

// MessageFilter 消息过滤器
type MessageFilter struct {
	UserID    string     `json:"user_id,omitempty"`
	EventType EventType  `json:"event_type,omitempty"`
	Room      string     `json:"room,omitempty"`
	Since     *time.Time `json:"since,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	SchoolCode   string    `json:"school_code"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastActivity time.Time `json:"last_activity"`
	Rooms        []string  `json:"rooms"`
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return "msg_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
