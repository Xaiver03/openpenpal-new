package websocket

// WebSocketAdapter - SOTA: Adapter Pattern to bridge WebSocketService with CourierService interface
type WebSocketAdapter struct {
	service *WebSocketService
}

// NewWebSocketAdapter creates a new adapter
func NewWebSocketAdapter(service *WebSocketService) *WebSocketAdapter {
	return &WebSocketAdapter{service: service}
}

// BroadcastToUser implements the WebSocketNotifier interface
func (a *WebSocketAdapter) BroadcastToUser(userID string, message interface{}) error {
	// Convert interface{} to *Message format expected by Hub
	var messageData map[string]interface{}

	// Type assertion to handle different message formats
	switch v := message.(type) {
	case map[string]interface{}:
		messageData = v
	default:
		// Wrap in generic structure
		messageData = map[string]interface{}{
			"data": v,
		}
	}

	wsMessage := NewMessage("courier_notification", messageData)
	a.service.GetHub().BroadcastToUser(userID, wsMessage)
	return nil
}
