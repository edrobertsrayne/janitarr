package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/edrobertsrayne/janitarr/src/database"
	"github.com/edrobertsrayne/janitarr/src/logger"
	"github.com/gorilla/websocket"
)

func TestHub_ClientConnect(t *testing.T) {
	// Create test database and logger
	db, err := database.New(":memory:", t.TempDir()+"/key")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	log := logger.NewLogger(db, logger.LevelInfo, false)
	hub := NewLogHub(log)
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	// Connect to the WebSocket
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect to websocket: %v", err)
	}
	defer ws.Close()

	// Wait a bit for connection to be registered
	time.Sleep(100 * time.Millisecond)

	// Check if client was added to hub
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 1 {
		t.Errorf("expected 1 client, got %d", clientCount)
	}
}

func TestHub_ClientDisconnect(t *testing.T) {
	// Create test database and logger
	db, err := database.New(":memory:", t.TempDir()+"/key")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	log := logger.NewLogger(db, logger.LevelInfo, false)
	hub := NewLogHub(log)
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	// Connect to the WebSocket
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect to websocket: %v", err)
	}

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	// Close the connection
	ws.Close()

	// Wait for disconnection to be processed
	time.Sleep(100 * time.Millisecond)

	// Check if client was removed from hub
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("expected 0 clients after disconnect, got %d", clientCount)
	}
}

func TestHub_Broadcast(t *testing.T) {
	// Create test database and logger
	db, err := database.New(":memory:", t.TempDir()+"/key")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	log := logger.NewLogger(db, logger.LevelInfo, false)
	hub := NewLogHub(log)
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	// Connect to the WebSocket
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect to websocket: %v", err)
	}
	defer ws.Close()

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	// Read the "connected" message
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read connected message: %v", err)
	}

	// Log an entry
	entry := log.LogCycleStart(false)

	// Wait a bit for message to be broadcast
	time.Sleep(100 * time.Millisecond)

	// Read the broadcasted message
	_, message, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var msg ServerMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if msg.Type != "log" {
		t.Errorf("expected message type 'log', got '%s'", msg.Type)
	}

	if msg.Data == nil {
		t.Error("expected log data in message")
	}

	// Verify the entry ID matches
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map")
	}

	if dataMap["id"] != entry.ID {
		t.Errorf("expected entry ID %s, got %v", entry.ID, dataMap["id"])
	}
}

func TestHub_FilteredBroadcast(t *testing.T) {
	// Create test database and logger
	db, err := database.New(":memory:", t.TempDir()+"/key")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	log := logger.NewLogger(db, logger.LevelInfo, false)
	hub := NewLogHub(log)
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	// Connect to the WebSocket
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect to websocket: %v", err)
	}
	defer ws.Close()

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	// Read the "connected" message
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read connected message: %v", err)
	}

	// Send subscribe message with filter for only "error" types
	filter := ClientMessage{
		Type: "subscribe",
		Filters: &WebSocketFilters{
			Types: []string{"error"},
		},
	}
	filterData, _ := json.Marshal(filter)
	if err := ws.WriteMessage(websocket.TextMessage, filterData); err != nil {
		t.Fatalf("failed to send filter: %v", err)
	}

	// Wait for filter to be processed
	time.Sleep(100 * time.Millisecond)

	// Log a cycle_start entry (should not be received)
	log.LogCycleStart(false)

	// Log an error entry (should be received)
	errorEntry := log.LogServerError("test-server", "radarr", "test error")

	// Wait for messages to be broadcast
	time.Sleep(100 * time.Millisecond)

	// Set read deadline
	ws.SetReadDeadline(time.Now().Add(1 * time.Second))

	// Try to read message - should receive the error entry
	_, message, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var msg ServerMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if msg.Type != "log" {
		t.Errorf("expected message type 'log', got '%s'", msg.Type)
	}

	// Verify it's the error entry
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map")
	}

	if dataMap["id"] != errorEntry.ID {
		t.Errorf("expected error entry ID %s, got %v", errorEntry.ID, dataMap["id"])
	}

	// Try to read another message with a short timeout - should timeout
	ws.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, err = ws.ReadMessage()
	if err == nil {
		t.Error("expected timeout but got a message")
	}
}
