package services

import (
	"context"
	"errors"
	"testing"

	"github.com/edrobertsrayne/janitarr/src/api"
	"github.com/edrobertsrayne/janitarr/src/database"
)

// mockDetectorClient implements the DetectorAPIClient interface for testing.
type mockDetectorClient struct {
	missing    []api.MediaItem
	cutoff     []api.MediaItem
	missingErr error
	cutoffErr  error
}

func (m *mockDetectorClient) TestConnection(ctx context.Context) (*api.SystemStatus, error) {
	return &api.SystemStatus{AppName: "Radarr", Version: "4.0.0"}, nil
}

func (m *mockDetectorClient) GetAllMissing(ctx context.Context) ([]api.MediaItem, error) {
	if m.missingErr != nil {
		return nil, m.missingErr
	}
	return m.missing, nil
}

func (m *mockDetectorClient) GetAllCutoffUnmet(ctx context.Context) ([]api.MediaItem, error) {
	if m.cutoffErr != nil {
		return nil, m.cutoffErr
	}
	return m.cutoff, nil
}

func (m *mockDetectorClient) TriggerSearch(ctx context.Context, ids []int) error {
	return nil
}

// testDetectorDB creates an in-memory test database.
func testDetectorDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.New(":memory:", t.TempDir()+"/key")
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestDetectAll_MultipleServers(t *testing.T) {
	db := testDetectorDB(t)
	ctx := context.Background()

	// Add two servers
	_, err := db.AddServer("radarr1", "http://localhost:7878", "test-key-1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server 1: %v", err)
	}
	_, err = db.AddServer("sonarr1", "http://localhost:8989", "test-key-2", database.ServerTypeSonarr)
	if err != nil {
		t.Fatalf("adding server 2: %v", err)
	}

	// Create detector with mock factory
	factory := func(url, apiKey, serverType string) DetectorAPIClient {
		if url == "http://localhost:7878" {
			return &mockDetectorClient{
				missing: []api.MediaItem{{ID: 1, Title: "Movie 1"}, {ID: 2, Title: "Movie 2"}},
				cutoff:  []api.MediaItem{{ID: 3, Title: "Movie 3"}},
			}
		}
		return &mockDetectorClient{
			missing: []api.MediaItem{{ID: 10, Title: "Episode 1"}},
			cutoff:  []api.MediaItem{{ID: 20, Title: "Episode 2"}, {ID: 21, Title: "Episode 3"}},
		}
	}

	detector := NewDetectorWithFactory(db, factory)
	results, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}

	// Verify aggregated results
	if results.TotalMissing != 3 {
		t.Errorf("TotalMissing = %d, want 3", results.TotalMissing)
	}
	if results.TotalCutoff != 3 {
		t.Errorf("TotalCutoff = %d, want 3", results.TotalCutoff)
	}
	if results.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", results.SuccessCount)
	}
	if results.FailureCount != 0 {
		t.Errorf("FailureCount = %d, want 0", results.FailureCount)
	}
	if len(results.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2", len(results.Results))
	}
}

func TestDetectAll_PartialFailure(t *testing.T) {
	db := testDetectorDB(t)
	ctx := context.Background()

	// Add two servers
	_, err := db.AddServer("radarr1", "http://localhost:7878", "test-key-1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server 1: %v", err)
	}
	_, err = db.AddServer("sonarr1", "http://localhost:8989", "test-key-2", database.ServerTypeSonarr)
	if err != nil {
		t.Fatalf("adding server 2: %v", err)
	}

	// Create detector with one failing mock
	factory := func(url, apiKey, serverType string) DetectorAPIClient {
		if url == "http://localhost:7878" {
			return &mockDetectorClient{
				missing: []api.MediaItem{{ID: 1, Title: "Movie 1"}},
				cutoff:  []api.MediaItem{},
			}
		}
		// Second server fails
		return &mockDetectorClient{
			missingErr: errors.New("connection timeout"),
		}
	}

	detector := NewDetectorWithFactory(db, factory)
	results, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}

	// Should have partial results - continues despite failures
	if results.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1", results.SuccessCount)
	}
	if results.FailureCount != 1 {
		t.Errorf("FailureCount = %d, want 1", results.FailureCount)
	}
	if results.TotalMissing != 1 {
		t.Errorf("TotalMissing = %d, want 1", results.TotalMissing)
	}

	// Verify error is recorded
	var foundError bool
	for _, r := range results.Results {
		if r.Error != "" {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("expected error in results, got none")
	}
}

func TestDetectAll_SkipsDisabled(t *testing.T) {
	db := testDetectorDB(t)
	ctx := context.Background()

	// Add two servers, one disabled
	srv1, err := db.AddServer("radarr1", "http://localhost:7878", "test-key-1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server 1: %v", err)
	}
	srv2, err := db.AddServer("radarr2", "http://localhost:7879", "test-key-2", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server 2: %v", err)
	}

	// Disable second server
	enabled := false
	if err := db.UpdateServer(srv2.ID, &database.ServerUpdate{Enabled: &enabled}); err != nil {
		t.Fatalf("disabling server 2: %v", err)
	}

	// Create detector - track which URLs are accessed
	accessedURLs := make(map[string]bool)
	factory := func(url, apiKey, serverType string) DetectorAPIClient {
		accessedURLs[url] = true
		return &mockDetectorClient{
			missing: []api.MediaItem{{ID: 1, Title: "Movie 1"}},
			cutoff:  []api.MediaItem{},
		}
	}

	detector := NewDetectorWithFactory(db, factory)
	results, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}

	// Only enabled server should be detected
	if results.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1", results.SuccessCount)
	}
	if len(results.Results) != 1 {
		t.Errorf("len(Results) = %d, want 1", len(results.Results))
	}
	if results.Results[0].ServerID != srv1.ID {
		t.Errorf("expected results from server %s, got %s", srv1.ID, results.Results[0].ServerID)
	}
	if !accessedURLs["http://localhost:7878"] {
		t.Error("expected enabled server to be accessed")
	}
	if accessedURLs["http://localhost:7879"] {
		t.Error("disabled server should not be accessed")
	}
}

func TestDetectMissing_Radarr(t *testing.T) {
	db := testDetectorDB(t)
	ctx := context.Background()

	// Add a radarr server
	_, err := db.AddServer("radarr1", "http://localhost:7878", "test-key-1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	factory := func(url, apiKey, serverType string) DetectorAPIClient {
		return &mockDetectorClient{
			missing: []api.MediaItem{
				{ID: 1, Title: "Movie 1", Type: "movie"},
				{ID: 2, Title: "Movie 2", Type: "movie"},
				{ID: 3, Title: "Movie 3", Type: "movie"},
			},
			cutoff: []api.MediaItem{},
		}
	}

	detector := NewDetectorWithFactory(db, factory)
	results, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}

	if results.TotalMissing != 3 {
		t.Errorf("TotalMissing = %d, want 3", results.TotalMissing)
	}
	if len(results.Results[0].Missing) != 3 {
		t.Errorf("len(Missing) = %d, want 3", len(results.Results[0].Missing))
	}
}

func TestDetectMissing_Sonarr(t *testing.T) {
	db := testDetectorDB(t)
	ctx := context.Background()

	// Add a sonarr server
	_, err := db.AddServer("sonarr1", "http://localhost:8989", "test-key-1", database.ServerTypeSonarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	factory := func(url, apiKey, serverType string) DetectorAPIClient {
		return &mockDetectorClient{
			missing: []api.MediaItem{
				{ID: 10, Title: "Series - S01E01", Type: "episode"},
				{ID: 11, Title: "Series - S01E02", Type: "episode"},
			},
			cutoff: []api.MediaItem{},
		}
	}

	detector := NewDetectorWithFactory(db, factory)
	results, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}

	if results.TotalMissing != 2 {
		t.Errorf("TotalMissing = %d, want 2", results.TotalMissing)
	}
	if len(results.Results[0].Missing) != 2 {
		t.Errorf("len(Missing) = %d, want 2", len(results.Results[0].Missing))
	}
}

func TestDetectCutoff_Radarr(t *testing.T) {
	db := testDetectorDB(t)
	ctx := context.Background()

	// Add a radarr server
	_, err := db.AddServer("radarr1", "http://localhost:7878", "test-key-1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	factory := func(url, apiKey, serverType string) DetectorAPIClient {
		return &mockDetectorClient{
			missing: []api.MediaItem{},
			cutoff: []api.MediaItem{
				{ID: 1, Title: "Movie 1", Type: "movie"},
				{ID: 2, Title: "Movie 2", Type: "movie"},
			},
		}
	}

	detector := NewDetectorWithFactory(db, factory)
	results, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}

	if results.TotalCutoff != 2 {
		t.Errorf("TotalCutoff = %d, want 2", results.TotalCutoff)
	}
	if len(results.Results[0].Cutoff) != 2 {
		t.Errorf("len(Cutoff) = %d, want 2", len(results.Results[0].Cutoff))
	}
}

func TestDetectCutoff_Sonarr(t *testing.T) {
	db := testDetectorDB(t)
	ctx := context.Background()

	// Add a sonarr server
	_, err := db.AddServer("sonarr1", "http://localhost:8989", "test-key-1", database.ServerTypeSonarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	factory := func(url, apiKey, serverType string) DetectorAPIClient {
		return &mockDetectorClient{
			missing: []api.MediaItem{},
			cutoff: []api.MediaItem{
				{ID: 10, Title: "Series - S01E01", Type: "episode"},
				{ID: 11, Title: "Series - S01E02", Type: "episode"},
				{ID: 12, Title: "Series - S01E03", Type: "episode"},
			},
		}
	}

	detector := NewDetectorWithFactory(db, factory)
	results, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}

	if results.TotalCutoff != 3 {
		t.Errorf("TotalCutoff = %d, want 3", results.TotalCutoff)
	}
	if len(results.Results[0].Cutoff) != 3 {
		t.Errorf("len(Cutoff) = %d, want 3", len(results.Results[0].Cutoff))
	}
}

func TestDetectAll_EmptyServers(t *testing.T) {
	db := testDetectorDB(t)
	ctx := context.Background()

	// No servers added

	factory := func(url, apiKey, serverType string) DetectorAPIClient {
		t.Error("factory should not be called with no servers")
		return nil
	}

	detector := NewDetectorWithFactory(db, factory)
	results, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}

	if results.TotalMissing != 0 {
		t.Errorf("TotalMissing = %d, want 0", results.TotalMissing)
	}
	if results.TotalCutoff != 0 {
		t.Errorf("TotalCutoff = %d, want 0", results.TotalCutoff)
	}
	if results.SuccessCount != 0 {
		t.Errorf("SuccessCount = %d, want 0", results.SuccessCount)
	}
	if results.FailureCount != 0 {
		t.Errorf("FailureCount = %d, want 0", results.FailureCount)
	}
	if len(results.Results) != 0 {
		t.Errorf("len(Results) = %d, want 0", len(results.Results))
	}
}

func TestDetectServer_CutoffFailureAfterMissing(t *testing.T) {
	db := testDetectorDB(t)
	ctx := context.Background()

	// Add a server
	_, err := db.AddServer("radarr1", "http://localhost:7878", "test-key-1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	// Missing succeeds but cutoff fails
	factory := func(url, apiKey, serverType string) DetectorAPIClient {
		return &mockDetectorClient{
			missing:   []api.MediaItem{{ID: 1, Title: "Movie 1"}},
			cutoffErr: errors.New("cutoff endpoint unavailable"),
		}
	}

	detector := NewDetectorWithFactory(db, factory)
	results, err := detector.DetectAll(ctx)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}

	// Should record failure even though missing succeeded
	if results.FailureCount != 1 {
		t.Errorf("FailureCount = %d, want 1", results.FailureCount)
	}
	if results.Results[0].Error == "" {
		t.Error("expected error message, got empty string")
	}
}
