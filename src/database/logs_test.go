package database

import (
	"context"
	"testing"
	"time"

	"github.com/edrobertsrayne/janitarr/src/logger"
)

func TestGetLogs_SearchFilter(t *testing.T) {
	// Create in-memory database
	db := testDB(t)

	ctx := context.Background()

	// Add test log entries
	testLogs := []logger.LogEntry{
		{
			ID:        "1",
			Timestamp: time.Now(),
			Type:      logger.LogTypeSearch,
			Message:   "Searching for Breaking Bad S01E01",
		},
		{
			ID:        "2",
			Timestamp: time.Now(),
			Type:      logger.LogTypeError,
			Message:   "Connection failed to Radarr",
		},
		{
			ID:        "3",
			Timestamp: time.Now(),
			Type:      logger.LogTypeDetection,
			Message:   "Found 15 missing items",
		},
		{
			ID:        "4",
			Timestamp: time.Now(),
			Type:      logger.LogTypeSearch,
			Message:   "Searching for The Matrix",
		},
	}

	for _, entry := range testLogs {
		if err := db.AddLog(entry); err != nil {
			t.Fatalf("Failed to add log: %v", err)
		}
	}

	tests := []struct {
		name          string
		searchTerm    string
		expectedCount int
		expectedIDs   []string
		description   string
	}{
		{
			name:          "Search for 'Breaking Bad'",
			searchTerm:    "Breaking Bad",
			expectedCount: 1,
			expectedIDs:   []string{"1"},
			description:   "Should find log with 'Breaking Bad' in message",
		},
		{
			name:          "Search for 'Searching'",
			searchTerm:    "Searching",
			expectedCount: 2,
			expectedIDs:   []string{"1", "4"},
			description:   "Should find all logs containing 'Searching'",
		},
		{
			name:          "Search for 'Radarr'",
			searchTerm:    "Radarr",
			expectedCount: 1,
			expectedIDs:   []string{"2"},
			description:   "Should find log with 'Radarr' in message",
		},
		{
			name:          "Search for 'missing'",
			searchTerm:    "missing",
			expectedCount: 1,
			expectedIDs:   []string{"3"},
			description:   "Should find log with 'missing' (case-insensitive)",
		},
		{
			name:          "Search for non-existent term",
			searchTerm:    "NonExistentTerm",
			expectedCount: 0,
			expectedIDs:   []string{},
			description:   "Should return no results for non-existent term",
		},
		{
			name:          "Empty search string",
			searchTerm:    "",
			expectedCount: 4,
			expectedIDs:   []string{"1", "2", "3", "4"},
			description:   "Empty search should return all logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := logger.LogFilters{}
			if tt.searchTerm != "" {
				filters.Search = &tt.searchTerm
			}

			logs, err := db.GetLogs(ctx, 100, 0, filters)
			if err != nil {
				t.Fatalf("GetLogs failed: %v", err)
			}

			if len(logs) != tt.expectedCount {
				t.Errorf("Expected %d logs, got %d. Description: %s", tt.expectedCount, len(logs), tt.description)
			}

			// Verify the correct log IDs are returned
			gotIDs := make(map[string]bool)
			for _, log := range logs {
				gotIDs[log.ID] = true
			}

			for _, expectedID := range tt.expectedIDs {
				if !gotIDs[expectedID] {
					t.Errorf("Expected log ID %s not found in results", expectedID)
				}
			}
		})
	}
}

func TestGetLogs_SearchWithOtherFilters(t *testing.T) {
	// Create in-memory database
	db := testDB(t)

	ctx := context.Background()

	// Add test log entries with different types and servers
	testLogs := []logger.LogEntry{
		{
			ID:         "1",
			Timestamp:  time.Now(),
			Type:       logger.LogTypeSearch,
			ServerName: "Radarr-Main",
			Message:    "Searching for Breaking Bad",
		},
		{
			ID:         "2",
			Timestamp:  time.Now(),
			Type:       logger.LogTypeError,
			ServerName: "Radarr-Main",
			Message:    "Connection failed",
		},
		{
			ID:         "3",
			Timestamp:  time.Now(),
			Type:       logger.LogTypeSearch,
			ServerName: "Sonarr-Main",
			Message:    "Searching for Breaking Bad",
		},
		{
			ID:         "4",
			Timestamp:  time.Now(),
			Type:       logger.LogTypeDetection,
			ServerName: "Sonarr-Main",
			Message:    "Found 10 missing items",
		},
	}

	for _, entry := range testLogs {
		if err := db.AddLog(entry); err != nil {
			t.Fatalf("Failed to add log: %v", err)
		}
	}

	// Test combining search with type filter
	searchTerm := "Breaking Bad"
	logType := string(logger.LogTypeSearch)
	filters := logger.LogFilters{
		Search: &searchTerm,
		Type:   &logType,
	}

	logs, err := db.GetLogs(ctx, 100, 0, filters)
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs (search type + 'Breaking Bad'), got %d", len(logs))
	}

	// Test combining search with server filter
	serverName := "Radarr-Main"
	filters = logger.LogFilters{
		Search: &searchTerm,
		Server: &serverName,
	}

	logs, err = db.GetLogs(ctx, 100, 0, filters)
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log (Radarr-Main + 'Breaking Bad'), got %d", len(logs))
	}

	if logs[0].ID != "1" {
		t.Errorf("Expected log ID '1', got '%s'", logs[0].ID)
	}
}

func TestGetLogs_SearchCaseInsensitivity(t *testing.T) {
	// Create in-memory database
	db := testDB(t)

	ctx := context.Background()

	// Add test log entry
	testLog := logger.LogEntry{
		ID:        "1",
		Timestamp: time.Now(),
		Type:      logger.LogTypeSearch,
		Message:   "Searching for The Matrix",
	}

	if err := db.AddLog(testLog); err != nil {
		t.Fatalf("Failed to add log: %v", err)
	}

	// Test various case variations
	searchTerms := []string{"matrix", "MATRIX", "MaTrIx", "the matrix", "THE MATRIX"}

	for _, term := range searchTerms {
		t.Run("Search: "+term, func(t *testing.T) {
			filters := logger.LogFilters{
				Search: &term,
			}

			logs, err := db.GetLogs(ctx, 100, 0, filters)
			if err != nil {
				t.Fatalf("GetLogs failed: %v", err)
			}

			if len(logs) != 1 {
				t.Errorf("Expected 1 log for search term '%s', got %d", term, len(logs))
			}
		})
	}
}
