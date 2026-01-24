package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edrobertsrayne/janitarr/src/database"
	"github.com/edrobertsrayne/janitarr/src/logger"
	"github.com/edrobertsrayne/janitarr/src/services"
	"github.com/edrobertsrayne/janitarr/src/web/handlers/api"
)

// testDB creates a new in-memory database for testing.
func testDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.New(":memory:", t.TempDir()+"/key")
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestHealthRouteAlias(t *testing.T) {
	db := testDB(t)

	// Create a logger for testing
	log := logger.NewLogger(db, logger.LevelInfo, false)

	// Create a scheduler with the required arguments
	// NewScheduler(db *database.DB, intervalHours int, cycleFunc func(ctx context.Context, isManual bool) error)
	scheduler := services.NewScheduler(db, 6, func(ctx context.Context, isManual bool) error {
		return nil
	})

	// Create a Server instance and setup routes
	server := NewServer(ServerConfig{
		Port:      3434,
		Host:      "localhost",
		DB:        db,
		Logger:    log,
		Scheduler: scheduler,
		IsDev:     false,
	})
	server.setupRoutes()

	// Test /health endpoint (the route we're adding)
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	// Accept 200 (OK) or 503 (degraded/error) - both are valid health check responses
	if rr.Code != http.StatusOK && rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 200 or 503 for /health, got %d", rr.Code)
	}

	var resp api.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal /health response: %v\nBody: %s", err, rr.Body.String())
	}

	if resp.Status == "" {
		t.Errorf("expected status field in response, got empty string")
	}

	// Verify /api/health still works
	req2 := httptest.NewRequest("GET", "/api/health", nil)
	rr2 := httptest.NewRecorder()

	server.router.ServeHTTP(rr2, req2)

	// Accept 200 (OK) or 503 (degraded/error) - both are valid health check responses
	if rr2.Code != http.StatusOK && rr2.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 200 or 503 for /api/health, got %d", rr2.Code)
	}

	var resp2 api.HealthResponse
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("failed to unmarshal /api/health response: %v\nBody: %s", err, rr2.Body.String())
	}

	// Verify both endpoints return the same response
	if resp.Status != resp2.Status {
		t.Errorf("expected both endpoints to return same status, got %s and %s", resp.Status, resp2.Status)
	}
}
