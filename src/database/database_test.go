package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testDB creates a new in-memory test database
func testDB(t *testing.T) *DB {
	t.Helper()
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, ".key")

	db, err := New(":memory:", keyPath)
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("closing test db: %v", err)
		}
	})

	return db
}

// TestNew tests database creation and migration
func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	keyPath := filepath.Join(tmpDir, ".key")

	// First creation
	db, err := New(dbPath, keyPath)
	if err != nil {
		t.Fatalf("creating database: %v", err)
	}

	// Verify key file was created
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("key file was not created")
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}

	if err := db.Close(); err != nil {
		t.Fatalf("closing database: %v", err)
	}

	// Reopen database
	db2, err := New(dbPath, keyPath)
	if err != nil {
		t.Fatalf("reopening database: %v", err)
	}
	defer db2.Close()

	// Verify it works
	if !db2.TestConnection() {
		t.Error("database connection test failed")
	}
}

func TestNew_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "data", "subdir")
	dbPath := filepath.Join(subDir, "test.db")
	keyPath := filepath.Join(tmpDir, ".key")

	db, err := New(dbPath, keyPath)
	if err != nil {
		t.Fatalf("creating database: %v", err)
	}
	defer db.Close()

	// Verify directory was created
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("database directory was not created")
	}
}

func TestNew_InMemory(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, ".key")

	db, err := New(":memory:", keyPath)
	if err != nil {
		t.Fatalf("creating in-memory database: %v", err)
	}
	defer db.Close()

	if !db.TestConnection() {
		t.Error("in-memory database connection test failed")
	}
}

// TestServerCRUD tests server add, get, update, delete operations
func TestServerCRUD(t *testing.T) {
	db := testDB(t)

	// Add server
	server, err := db.AddServer("test-radarr", "http://localhost:7878", "api-key-123", ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	if server.ID == "" {
		t.Error("server ID should not be empty")
	}
	if server.Name != "test-radarr" {
		t.Errorf("expected name 'test-radarr', got '%s'", server.Name)
	}
	if server.URL != "http://localhost:7878" {
		t.Errorf("expected URL 'http://localhost:7878', got '%s'", server.URL)
	}
	if server.Type != ServerTypeRadarr {
		t.Errorf("expected type 'radarr', got '%s'", server.Type)
	}
	if !server.Enabled {
		t.Error("server should be enabled by default")
	}

	// Get server by ID
	fetched, err := db.GetServer(server.ID)
	if err != nil {
		t.Fatalf("getting server by ID: %v", err)
	}
	if fetched == nil {
		t.Fatal("server not found by ID")
	}
	if fetched.Name != server.Name {
		t.Errorf("expected name '%s', got '%s'", server.Name, fetched.Name)
	}
	// API key should be decrypted
	if fetched.APIKey != "api-key-123" {
		t.Errorf("expected API key 'api-key-123', got '%s'", fetched.APIKey)
	}

	// Get server by name
	byName, err := db.GetServerByName("test-radarr")
	if err != nil {
		t.Fatalf("getting server by name: %v", err)
	}
	if byName == nil {
		t.Fatal("server not found by name")
	}
	if byName.ID != server.ID {
		t.Errorf("expected ID '%s', got '%s'", server.ID, byName.ID)
	}

	// Update server
	newURL := "http://localhost:8080"
	err = db.UpdateServer(server.ID, &ServerUpdate{
		URL: &newURL,
	})
	if err != nil {
		t.Fatalf("updating server: %v", err)
	}

	updated, _ := db.GetServer(server.ID)
	if updated.URL != newURL {
		t.Errorf("expected URL '%s', got '%s'", newURL, updated.URL)
	}

	// List servers
	servers, err := db.GetAllServers()
	if err != nil {
		t.Fatalf("listing servers: %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(servers))
	}

	// Delete server
	deleted, err := db.DeleteServer(server.ID)
	if err != nil {
		t.Fatalf("deleting server: %v", err)
	}
	if !deleted {
		t.Error("delete should return true")
	}

	// Verify deleted
	fetched, err = db.GetServer(server.ID)
	if err != nil {
		t.Fatalf("getting deleted server: %v", err)
	}
	if fetched != nil {
		t.Error("server should be deleted")
	}
}

func TestServerDuplicateName(t *testing.T) {
	db := testDB(t)

	_, err := db.AddServer("my-server", "http://localhost:7878", "key1", ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding first server: %v", err)
	}

	// Try to add server with same name
	_, err = db.AddServer("my-server", "http://localhost:8989", "key2", ServerTypeSonarr)
	if err == nil {
		t.Error("expected error for duplicate server name")
	}
}

func TestServerGetByName_CaseInsensitive(t *testing.T) {
	db := testDB(t)

	_, err := db.AddServer("MyServer", "http://localhost:7878", "key1", ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	// Find by lowercase
	server, err := db.GetServerByName("myserver")
	if err != nil {
		t.Fatalf("getting server: %v", err)
	}
	if server == nil {
		t.Error("server should be found case-insensitively")
	}

	// Find by uppercase
	server, err = db.GetServerByName("MYSERVER")
	if err != nil {
		t.Fatalf("getting server: %v", err)
	}
	if server == nil {
		t.Error("server should be found case-insensitively")
	}
}

func TestServersByType(t *testing.T) {
	db := testDB(t)

	_, _ = db.AddServer("radarr1", "http://localhost:7878", "key1", ServerTypeRadarr)
	_, _ = db.AddServer("sonarr1", "http://localhost:8989", "key2", ServerTypeSonarr)
	_, _ = db.AddServer("radarr2", "http://localhost:7879", "key3", ServerTypeRadarr)

	radarrs, err := db.GetServersByType(ServerTypeRadarr)
	if err != nil {
		t.Fatalf("getting servers by type: %v", err)
	}
	if len(radarrs) != 2 {
		t.Errorf("expected 2 radarr servers, got %d", len(radarrs))
	}

	sonarrs, err := db.GetServersByType(ServerTypeSonarr)
	if err != nil {
		t.Fatalf("getting servers by type: %v", err)
	}
	if len(sonarrs) != 1 {
		t.Errorf("expected 1 sonarr server, got %d", len(sonarrs))
	}
}

// TestConfigGetSet tests configuration persistence
func TestConfigGetSet(t *testing.T) {
	db := testDB(t)

	// Get default config
	config := db.GetAppConfig()
	defaults := DefaultAppConfig()

	if config.Schedule.IntervalHours != defaults.Schedule.IntervalHours {
		t.Errorf("expected default interval %d, got %d", defaults.Schedule.IntervalHours, config.Schedule.IntervalHours)
	}
	if config.Schedule.Enabled != defaults.Schedule.Enabled {
		t.Errorf("expected default enabled %v, got %v", defaults.Schedule.Enabled, config.Schedule.Enabled)
	}

	// Set config values
	db.SetConfig("schedule.intervalHours", "12")
	db.SetConfig("schedule.enabled", "false")

	// Verify changes
	config = db.GetAppConfig()
	if config.Schedule.IntervalHours != 12 {
		t.Errorf("expected interval 12, got %d", config.Schedule.IntervalHours)
	}
	if config.Schedule.Enabled != false {
		t.Error("expected enabled false")
	}

	// Get single config value
	val := db.GetConfig("schedule.intervalHours")
	if val == nil || *val != "12" {
		t.Errorf("expected '12', got %v", val)
	}

	// Get non-existent key
	val = db.GetConfig("nonexistent")
	if val != nil {
		t.Error("expected nil for non-existent key")
	}
}

func TestSetAppConfig(t *testing.T) {
	db := testDB(t)

	// Update partial config
	db.SetAppConfig(AppConfigUpdate{
		Schedule: &ScheduleConfigUpdate{
			IntervalHours: intPtr(24),
		},
		SearchLimits: &SearchLimitsUpdate{
			MissingMoviesLimit: intPtr(20),
		},
	})

	config := db.GetAppConfig()
	if config.Schedule.IntervalHours != 24 {
		t.Errorf("expected interval 24, got %d", config.Schedule.IntervalHours)
	}
	if config.SearchLimits.MissingMoviesLimit != 20 {
		t.Errorf("expected missing movies limit 20, got %d", config.SearchLimits.MissingMoviesLimit)
	}
	// Other values should remain default
	if config.Schedule.Enabled != true {
		t.Error("enabled should remain true")
	}
}

// TestLogsInsertRetrieve tests log operations
func TestLogsInsertRetrieve(t *testing.T) {
	db := testDB(t)

	// Add log entries
	entry1 := db.AddLog(LogEntryInput{
		Type:     LogTypeCycleStart,
		Message:  "Starting cycle",
		IsManual: true,
	})

	if entry1.ID == "" {
		t.Error("log entry ID should not be empty")
	}
	if entry1.Timestamp.IsZero() {
		t.Error("log entry timestamp should be set")
	}

	_ = db.AddLog(LogEntryInput{
		Type:       LogTypeSearch,
		ServerName: "test-radarr",
		ServerType: ServerTypeRadarr,
		Category:   SearchCategoryMissing,
		Count:      5,
		Message:    "Triggered 5 searches",
		IsManual:   true,
	})

	entry3 := db.AddLog(LogEntryInput{
		Type:    LogTypeCycleEnd,
		Message: "Cycle complete",
	})

	// Get logs
	logs := db.GetLogs(100, 0)
	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}

	// Verify entry3 exists (cycle_end)
	var foundCycleEnd bool
	for _, log := range logs {
		if log.ID == entry3.ID {
			foundCycleEnd = true
			break
		}
	}
	if !foundCycleEnd {
		t.Error("cycle_end log entry should be in results")
	}

	// Verify log entry details - find the search log
	var searchLog *LogEntry
	for i := range logs {
		if logs[i].ServerName == "test-radarr" {
			searchLog = &logs[i]
			break
		}
	}
	if searchLog == nil {
		t.Fatal("search log entry not found")
	}
	if searchLog.Count != 5 {
		t.Errorf("expected count 5, got %d", searchLog.Count)
	}
}

// TestLogsPagination tests offset and limit
func TestLogsPagination(t *testing.T) {
	db := testDB(t)

	// Add 10 log entries
	for i := 0; i < 10; i++ {
		db.AddLog(LogEntryInput{
			Type:    LogTypeSearch,
			Message: "Search log",
		})
	}

	// Get first page
	page1 := db.GetLogs(5, 0)
	if len(page1) != 5 {
		t.Errorf("expected 5 logs in page 1, got %d", len(page1))
	}

	// Get second page
	page2 := db.GetLogs(5, 5)
	if len(page2) != 5 {
		t.Errorf("expected 5 logs in page 2, got %d", len(page2))
	}

	// Ensure different entries
	if page1[0].ID == page2[0].ID {
		t.Error("pages should have different entries")
	}

	// Get count
	count := db.GetLogCount()
	if count != 10 {
		t.Errorf("expected 10 log count, got %d", count)
	}
}

// TestLogsPurge tests delete old entries
func TestLogsPurge(t *testing.T) {
	db := testDB(t)

	// Add old log entries (manually set timestamp in past)
	oldTime := time.Now().AddDate(0, 0, -31).Format(time.RFC3339)

	// Use raw SQL to insert old entries for testing
	db.conn.Exec(`
		INSERT INTO logs (id, timestamp, type, message, is_manual)
		VALUES ('old-1', ?, 'search', 'Old log 1', 0),
		       ('old-2', ?, 'search', 'Old log 2', 0)
	`, oldTime, oldTime)

	// Add recent log
	db.AddLog(LogEntryInput{
		Type:    LogTypeSearch,
		Message: "Recent log",
	})

	// Verify we have 3 logs
	count := db.GetLogCount()
	if count != 3 {
		t.Errorf("expected 3 logs before purge, got %d", count)
	}

	// Purge old logs (30 days retention)
	purged := db.PurgeOldLogs()
	if purged != 2 {
		t.Errorf("expected 2 purged logs, got %d", purged)
	}

	// Verify only recent log remains
	count = db.GetLogCount()
	if count != 1 {
		t.Errorf("expected 1 log after purge, got %d", count)
	}
}

func TestLogsFilter(t *testing.T) {
	db := testDB(t)

	db.AddLog(LogEntryInput{Type: LogTypeCycleStart, Message: "Start"})
	db.AddLog(LogEntryInput{Type: LogTypeSearch, ServerName: "radarr1", Message: "Search radarr"})
	db.AddLog(LogEntryInput{Type: LogTypeSearch, ServerName: "sonarr1", Message: "Search sonarr"})
	db.AddLog(LogEntryInput{Type: LogTypeError, ServerName: "radarr1", Message: "Error"})

	// Filter by type
	searchLogs := db.GetLogsPaginated(LogFilters{Type: LogTypeSearch}, 100, 0)
	if len(searchLogs) != 2 {
		t.Errorf("expected 2 search logs, got %d", len(searchLogs))
	}

	// Filter by server
	radarrLogs := db.GetLogsPaginated(LogFilters{Server: "radarr1"}, 100, 0)
	if len(radarrLogs) != 2 {
		t.Errorf("expected 2 radarr logs, got %d", len(radarrLogs))
	}

	// Filter by text search
	errorLogs := db.GetLogsPaginated(LogFilters{Search: "Error"}, 100, 0)
	if len(errorLogs) != 1 {
		t.Errorf("expected 1 error log, got %d", len(errorLogs))
	}
}

func TestClearLogs(t *testing.T) {
	db := testDB(t)

	db.AddLog(LogEntryInput{Type: LogTypeSearch, Message: "Log 1"})
	db.AddLog(LogEntryInput{Type: LogTypeSearch, Message: "Log 2"})

	cleared := db.ClearLogs()
	if cleared != 2 {
		t.Errorf("expected 2 cleared, got %d", cleared)
	}

	count := db.GetLogCount()
	if count != 0 {
		t.Errorf("expected 0 logs after clear, got %d", count)
	}
}

func TestServerStats(t *testing.T) {
	db := testDB(t)

	server, _ := db.AddServer("test-server", "http://localhost:7878", "key", ServerTypeRadarr)

	// Add some logs for this server
	db.AddLog(LogEntryInput{Type: LogTypeSearch, ServerName: "test-server", Message: "Search 1"})
	db.AddLog(LogEntryInput{Type: LogTypeSearch, ServerName: "test-server", Message: "Search 2"})
	db.AddLog(LogEntryInput{Type: LogTypeError, ServerName: "test-server", Message: "Error"})
	db.AddLog(LogEntryInput{Type: LogTypeSearch, ServerName: "other-server", Message: "Other"})

	stats := db.GetServerStats(server.ID)
	if stats.TotalSearches != 2 {
		t.Errorf("expected 2 searches, got %d", stats.TotalSearches)
	}
	if stats.ErrorCount != 1 {
		t.Errorf("expected 1 error, got %d", stats.ErrorCount)
	}
	if stats.LastCheckTime == "" {
		t.Error("expected last check time to be set")
	}
}

func TestSystemStats(t *testing.T) {
	db := testDB(t)

	db.AddServer("server1", "http://localhost:7878", "key1", ServerTypeRadarr)
	db.AddServer("server2", "http://localhost:8989", "key2", ServerTypeSonarr)

	db.AddLog(LogEntryInput{Type: LogTypeCycleEnd, Message: "Cycle done"})
	db.AddLog(LogEntryInput{Type: LogTypeSearch, Message: "Search"})
	db.AddLog(LogEntryInput{Type: LogTypeError, Message: "Error"})

	stats := db.GetSystemStats()
	if stats.TotalServers != 2 {
		t.Errorf("expected 2 servers, got %d", stats.TotalServers)
	}
	if stats.SearchesLast24h != 1 {
		t.Errorf("expected 1 search, got %d", stats.SearchesLast24h)
	}
	if stats.ErrorsLast24h != 1 {
		t.Errorf("expected 1 error, got %d", stats.ErrorsLast24h)
	}
	if stats.LastCycleTime == "" {
		t.Error("expected last cycle time to be set")
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}
