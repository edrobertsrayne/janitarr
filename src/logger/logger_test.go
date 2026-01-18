package logger

import (
	"testing"
	"time"

	"github.com/user/janitarr/src/database"
)

func testLoggerDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.New(":memory:", t.TempDir()+"/.key")
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestLogCycleStart_Persists(t *testing.T) {
	db := testLoggerDB(t)
	logger := NewLogger(db)

	logger.LogCycleStart(true)

	logs, _, err := db.GetLogs(1, 0, "", "")
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}
	if logs[0].Type != database.LogEntryType(LogTypeCycleStart) {
		t.Errorf("expected log type %s, got %s", LogTypeCycleStart, logs[0].Type)
	}
}

func TestLogCycleEnd_Persists(t *testing.T) {
	db := testLoggerDB(t)
	logger := NewLogger(db)

	logger.LogCycleEnd(10, 2, false)

	logs, _, err := db.GetLogs(1, 0, "", "")
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}
	if logs[0].Type != database.LogEntryType(LogTypeCycleEnd) {
		t.Errorf("expected log type %s, got %s", LogTypeCycleEnd, logs[0].Type)
	}
}

func TestLogSearches_Persists(t *testing.T) {
	db := testLoggerDB(t)
	logger := NewLogger(db)

	logger.LogSearches("radarr", "radarr", "missing", 5, true)

	logs, _, err := db.GetLogs(1, 0, "", "")
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}
	if logs[0].Type != database.LogEntryType(LogTypeSearch) {
		t.Errorf("expected log type %s, got %s", LogTypeSearch, logs[0].Type)
	}
}

func TestLogError_Persists(t *testing.T) {
	db := testLoggerDB(t)
	logger := NewLogger(db)

	logger.LogServerError("radarr", "radarr", "connection failed")

	logs, _, err := db.GetLogs(1, 0, "", "")
	if err != nil {
		t.Fatalf("failed to get logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}
	if logs[0].Type != database.LogEntryType(LogTypeError) {
		t.Errorf("expected log type %s, got %s", LogTypeError, logs[0].Type)
	}
}

func TestBroadcast_SendsToSubscribers(t *testing.T) {
	db := testLoggerDB(t)
	logger := NewLogger(db)

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
	db := testLoggerDB(t)
	logger := NewLogger(db)

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
