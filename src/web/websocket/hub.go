package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/user/janitarr/src/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now. In production, you may want to restrict this.
		return true
	},
}

// LogHub maintains the set of active clients and broadcasts log messages to them.
type LogHub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan *logger.LogEntry
	register   chan *Client
	unregister chan *Client
	logger     *logger.Logger
}

// NewLogHub creates a new LogHub and subscribes to the logger.
func NewLogHub(log *logger.Logger) *LogHub {
	hub := &LogHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *logger.LogEntry, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     log,
	}

	// Subscribe to logger
	logChan := log.Subscribe()
	go func() {
		for entry := range logChan {
			hub.broadcast <- &entry
		}
	}()

	return hub
}

// Run starts the hub's main loop.
func (h *LogHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			// Send connected message
			msg := ServerMessage{
				Type:    "connected",
				Message: "WebSocket connection established",
			}
			if data, err := json.Marshal(msg); err == nil {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case entry := <-h.broadcast:
			h.Broadcast(entry)
		}
	}
}

// Broadcast sends a log entry to all matching clients.
func (h *LogHub) Broadcast(entry *logger.LogEntry) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := ServerMessage{
		Type: "log",
		Data: entry,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for client := range h.clients {
		if client.shouldSend(entry) {
			select {
			case client.send <- data:
			default:
				// Client's send buffer is full, close it
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// ServeWS handles websocket requests from the peer.
func (h *LogHub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Log error
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// Close shuts down the hub and disconnects all clients.
func (h *LogHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
}
