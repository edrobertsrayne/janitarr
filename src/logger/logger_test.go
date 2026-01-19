package logger

import (
	"context"
	"testing"
	"time"
)

// mockDB implements LogStorer for testing
type mockDB struct {
	logs []LogEntry
}

func (m *mockDB) AddLog(entry LogEntry) error {
	m.logs = append(m.logs, entry)
	return nil
}

func (m *mockDB) GetLogs(ctx context.Context, limit, offset int, logTypeFilter, serverNameFilter *string) ([]LogEntry, error) {
	return m.logs, nil
}

func (m *mockDB) ClearLogs() error {
	m.logs = nil
	return nil
}

func TestLogCycleStart_Persists(t *testing.T) {
	db := &mockDB{}
	logger := NewLogger(db, LevelInfo, false)

	logger.LogCycleStart(true)

	if len(db.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(db.logs))
	}
	if db.logs[0].Type != LogTypeCycleStart {
		t.Errorf("expected log type %s, got %s", LogTypeCycleStart, db.logs[0].Type)
	}
}

func TestLogCycleEnd_Persists(t *testing.T) {
	db := &mockDB{}
	logger := NewLogger(db, LevelInfo, false)

	logger.LogCycleEnd(10, 2, false)

	if len(db.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(db.logs))
	}
	if db.logs[0].Type != LogTypeCycleEnd {
		t.Errorf("expected log type %s, got %s", LogTypeCycleEnd, db.logs[0].Type)
	}
}

func TestLogDetectionComplete_Persists(t *testing.T) {
	db := &mockDB{}
	logger := NewLogger(db, LevelInfo, false)

	logger.LogDetectionComplete("radarr-main", "radarr", 5, 12)

	if len(db.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(db.logs))
	}
	if db.logs[0].Type != LogTypeDetection {
		t.Errorf("expected log type %s, got %s", LogTypeDetection, db.logs[0].Type)
	}
	if db.logs[0].ServerName != "radarr-main" {
		t.Errorf("expected server name %s, got %s", "radarr-main", db.logs[0].ServerName)
	}
	if db.logs[0].ServerType != "radarr" {
		t.Errorf("expected server type %s, got %s", "radarr", db.logs[0].ServerType)
	}
	if db.logs[0].Count != 17 {
		t.Errorf("expected count %d (5+12), got %d", 17, db.logs[0].Count)
	}
}

func TestLogSearches_Persists(t *testing.T) {
	db := &mockDB{}
	logger := NewLogger(db, LevelInfo, false)

	logger.LogSearches("radarr", "radarr", "missing", 5, true)

	if len(db.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(db.logs))
	}
	if db.logs[0].Type != LogTypeSearch {
		t.Errorf("expected log type %s, got %s", LogTypeSearch, db.logs[0].Type)
	}
}

func TestLogError_Persists(t *testing.T) {
	db := &mockDB{}
	logger := NewLogger(db, LevelInfo, false)

	logger.LogServerError("radarr", "radarr", "connection failed")

	if len(db.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(db.logs))
	}
	if db.logs[0].Type != LogTypeError {
		t.Errorf("expected log type %s, got %s", LogTypeError, db.logs[0].Type)
	}
}

func TestBroadcast_SendsToSubscribers(t *testing.T) {
	db := &mockDB{}
	logger := NewLogger(db, LevelInfo, false)

	sub := logger.Subscribe()

	go logger.LogCycleStart(true)

	select {
	case entry := <-sub:
		if entry.Type != LogTypeCycleStart {
			t.Errorf("expected log type %s, got %s", LogTypeCycleStart, entry.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for log entry")
	}
}

func TestUnsubscribe_StopsReceiving(t *testing.T) {
	db := &mockDB{}
	logger := NewLogger(db, LevelInfo, false)

	sub := make(chan LogEntry, 1)
	logger.subscribers[sub] = true

	logger.Unsubscribe(sub)

	logger.LogCycleStart(true)

	select {
	case _, ok := <-sub:
		if ok {
			t.Fatal("received log entry after unsubscribe")
		}
	case <-time.After(100 * time.Millisecond):
		// success
	}
}
