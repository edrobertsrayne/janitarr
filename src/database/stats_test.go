package database

import (
	"path/filepath"
	"testing"
)

func testStatsDB(t *testing.T) *DB {
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

func TestGetServerCounts(t *testing.T) {
	db := testStatsDB(t)

	// Test with no servers
	counts, err := db.GetServerCounts()
	if err != nil {
		t.Fatalf("GetServerCounts failed: %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected 0 server types, got %d", len(counts))
	}

	// Add some test servers
	servers := []struct {
		name    string
		typ     string
		enabled bool
	}{
		{"radarr1", "radarr", true},
		{"radarr2", "radarr", true},
		{"radarr3", "radarr", false},
		{"sonarr1", "sonarr", true},
		{"sonarr2", "sonarr", false},
	}

	for _, srv := range servers {
		_, err := db.AddServer(srv.name, "http://example.com", "test-key", ServerType(srv.typ))
		if err != nil {
			t.Fatalf("failed to add server: %v", err)
		}
		// Update enabled status if needed
		if !srv.enabled {
			// Use raw SQL to disable the server since there's no UpdateServer in the test context
			_, err = db.conn.Exec("UPDATE servers SET enabled = 0 WHERE name = ?", srv.name)
			if err != nil {
				t.Fatalf("failed to disable server: %v", err)
			}
		}
	}

	// Test server counts
	counts, err = db.GetServerCounts()
	if err != nil {
		t.Fatalf("GetServerCounts failed: %v", err)
	}

	// Check radarr counts
	if radarrCounts, ok := counts["radarr"]; !ok {
		t.Error("radarr counts missing")
	} else {
		if radarrCounts.Configured != 3 {
			t.Errorf("expected 3 configured radarr servers, got %d", radarrCounts.Configured)
		}
		if radarrCounts.Enabled != 2 {
			t.Errorf("expected 2 enabled radarr servers, got %d", radarrCounts.Enabled)
		}
	}

	// Check sonarr counts
	if sonarrCounts, ok := counts["sonarr"]; !ok {
		t.Error("sonarr counts missing")
	} else {
		if sonarrCounts.Configured != 2 {
			t.Errorf("expected 2 configured sonarr servers, got %d", sonarrCounts.Configured)
		}
		if sonarrCounts.Enabled != 1 {
			t.Errorf("expected 1 enabled sonarr server, got %d", sonarrCounts.Enabled)
		}
	}
}

func TestGetServerCounts_EmptyDatabase(t *testing.T) {
	db := testStatsDB(t)

	counts, err := db.GetServerCounts()
	if err != nil {
		t.Fatalf("GetServerCounts failed: %v", err)
	}

	if len(counts) != 0 {
		t.Errorf("expected empty map, got %d entries", len(counts))
	}
}
