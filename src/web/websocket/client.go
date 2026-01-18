package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/user/janitarr/src/logger"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client represents a WebSocket client connection.
type Client struct {
	hub       *LogHub
	conn      *websocket.Conn
	send      chan []byte
	filtersMu sync.RWMutex
	filters   *WebSocketFilters
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close
			}
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle client messages
		switch msg.Type {
		case "subscribe":
			if msg.Filters != nil {
				c.filtersMu.Lock()
				c.filters = msg.Filters
				c.filtersMu.Unlock()
			}
		case "unsubscribe":
			c.filtersMu.Lock()
			c.filters = nil
			c.filtersMu.Unlock()
		case "ping":
			// Send pong
			response := ServerMessage{
				Type: "pong",
			}
			if data, err := json.Marshal(response); err == nil {
				select {
				case c.send <- data:
				default:
				}
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
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
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
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

// shouldSend determines if an entry should be sent to the client based on filters.
func (c *Client) shouldSend(entry *logger.LogEntry) bool {
	c.filtersMu.RLock()
	defer c.filtersMu.RUnlock()

	if c.filters == nil {
		return true
	}

	// Filter by type
	if len(c.filters.Types) > 0 {
		typeMatch := false
		for _, t := range c.filters.Types {
			if string(entry.Type) == t {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			return false
		}
	}

	// Filter by server
	if len(c.filters.Servers) > 0 {
		serverMatch := false
		for _, s := range c.filters.Servers {
			if entry.ServerName == s {
				serverMatch = true
				break
			}
		}
		if !serverMatch {
			return false
		}
	}

	return true
}
