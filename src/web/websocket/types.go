package websocket

// ClientMessage represents a message sent by the client.
type ClientMessage struct {
	Type    string            `json:"type"` // subscribe, unsubscribe, ping
	Filters *WebSocketFilters `json:"filters,omitempty"`
}

// ServerMessage represents a message sent by the server.
type ServerMessage struct {
	Type    string      `json:"type"` // connected, log, pong
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// WebSocketFilters represents filter criteria for log messages.
type WebSocketFilters struct {
	Types   []string `json:"types,omitempty"`
	Servers []string `json:"servers,omitempty"`
}
