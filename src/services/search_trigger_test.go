package services

import (
	"context"
	"errors"
	"testing"

	"github.com/user/janitarr/src/api"
	"github.com/user/janitarr/src/database"
)

// mockTriggerAPIClient is a mock implementation of SearchTriggerAPIClient for testing.
type mockTriggerAPIClient struct {
	serverType   string
	triggerErr   error
	triggerCalls [][]int
}

func (m *mockTriggerAPIClient) TestConnection(ctx context.Context) (*api.SystemStatus, error) {
	return &api.SystemStatus{AppName: m.serverType, Version: "1.0"}, nil
}

func (m *mockTriggerAPIClient) GetAllMissing(ctx context.Context) ([]api.MediaItem, error) {
	return []api.MediaItem{}, nil
}

func (m *mockTriggerAPIClient) GetAllCutoffUnmet(ctx context.Context) ([]api.MediaItem, error) {
	return []api.MediaItem{}, nil
}

func (m *mockTriggerAPIClient) TriggerSearch(ctx context.Context, ids []int) error {
	m.triggerCalls = append(m.triggerCalls, ids)
	return m.triggerErr
}

func (m *mockTriggerAPIClient) getTriggerCalls() [][]int {
	return m.triggerCalls
}

// testTriggerDB creates a test database for SearchTrigger tests.
func testTriggerDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.New(":memory:", t.TempDir()+"/.key")
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestTriggerSearches_RespectsLimits(t *testing.T) {
	db := testTriggerDB(t)

	// Add a radarr server
	server1, err := db.AddServer("radarr1", "http://localhost:7878", "api1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	// Create mock client
	mockClient := &mockTriggerAPIClient{serverType: "radarr"}

	// Create SearchTrigger with mock factory
	trigger := NewSearchTriggerWithFactory(db, func(url, apiKey, serverType string) SearchTriggerAPIClient {
		return mockClient
	})

	// Create detection results with more items than the limit
	// We need 15 missing items, but the default limit is 10
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   server1.ID,
				ServerName: "radarr1",
				ServerType: "radarr",
				Missing:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
				Cutoff:     []int{},
			},
		},
		TotalMissing: 15,
		TotalCutoff:  0,
		SuccessCount: 1,
		FailureCount: 0,
	}

	// Set limits: 5 missing movies, 0 cutoff
	limits := database.SearchLimits{MissingMoviesLimit: 5, CutoffMoviesLimit: 0}

	// Trigger searches
	ctx := context.Background()
	results, err := trigger.TriggerSearches(ctx, detectionResults, limits, false)
	if err != nil {
		t.Fatalf("TriggerSearches failed: %v", err)
	}

	// Check that we only triggered up to the limit
	if results.MissingTriggered > 5 {
		t.Errorf("expected at most 5 missing triggered, got %d", results.MissingTriggered)
	}

	// Check the trigger calls
	calls := mockClient.getTriggerCalls()
	totalItems := 0
	for _, call := range calls {
		totalItems += len(call)
	}
	if totalItems > 5 {
		t.Errorf("expected at most 5 total items triggered, got %d", totalItems)
	}
}

func TestTriggerSearches_RoundRobin(t *testing.T) {
	db := testTriggerDB(t)

	// Add two radarr servers
	server1, err := db.AddServer("radarr1", "http://localhost:7878", "api1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server 1: %v", err)
	}
	server2, err := db.AddServer("radarr2", "http://localhost:7879", "api2", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server 2: %v", err)
	}

	// Create mock clients (track separately)
	mockClients := make(map[string]*mockTriggerAPIClient)
	mockClients["http://localhost:7878"] = &mockTriggerAPIClient{serverType: "radarr"}
	mockClients["http://localhost:7879"] = &mockTriggerAPIClient{serverType: "radarr"}

	// Create SearchTrigger with mock factory
	trigger := NewSearchTriggerWithFactory(db, func(url, apiKey, serverType string) SearchTriggerAPIClient {
		return mockClients[url]
	})

	// Create detection results with items from both servers
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   server1.ID,
				ServerName: "radarr1",
				ServerType: "radarr",
				Missing:    []int{1, 2, 3, 4, 5}, // Server 1 has 5 items
				Cutoff:     []int{},
			},
			{
				ServerID:   server2.ID,
				ServerName: "radarr2",
				ServerType: "radarr",
				Missing:    []int{10, 11, 12, 13, 14}, // Server 2 has 5 items
				Cutoff:     []int{},
			},
		},
		TotalMissing: 10,
		TotalCutoff:  0,
		SuccessCount: 2,
		FailureCount: 0,
	}

	// Set limits: 6 missing (should distribute 3 to each server if round-robin)
	limits := database.SearchLimits{MissingMoviesLimit: 6, CutoffMoviesLimit: 0}

	// Trigger searches
	ctx := context.Background()
	results, err := trigger.TriggerSearches(ctx, detectionResults, limits, false)
	if err != nil {
		t.Fatalf("TriggerSearches failed: %v", err)
	}

	// Check that we triggered from both servers
	if results.MissingTriggered != 6 {
		t.Errorf("expected 6 missing triggered, got %d", results.MissingTriggered)
	}

	// Both servers should have received trigger calls
	calls1 := mockClients["http://localhost:7878"].getTriggerCalls()
	calls2 := mockClients["http://localhost:7879"].getTriggerCalls()

	if len(calls1) == 0 && len(calls2) == 0 {
		t.Errorf("expected at least one server to receive trigger calls")
	}

	// Count total items per server
	totalItems1 := 0
	for _, call := range calls1 {
		totalItems1 += len(call)
	}
	totalItems2 := 0
	for _, call := range calls2 {
		totalItems2 += len(call)
	}

	// With round-robin, the distribution should be roughly equal
	if totalItems1+totalItems2 != 6 {
		t.Errorf("expected 6 total items, got %d", totalItems1+totalItems2)
	}
}

func TestTriggerSearches_DryRun(t *testing.T) {
	db := testTriggerDB(t)

	// Add a server
	server1, err := db.AddServer("radarr1", "http://localhost:7878", "api1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	// Create mock client
	mockClient := &mockTriggerAPIClient{serverType: "radarr"}

	// Create SearchTrigger with mock factory
	trigger := NewSearchTriggerWithFactory(db, func(url, apiKey, serverType string) SearchTriggerAPIClient {
		return mockClient
	})

	// Create detection results
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   server1.ID,
				ServerName: "radarr1",
				ServerType: "radarr",
				Missing:    []int{1, 2, 3},
				Cutoff:     []int{10, 11},
			},
		},
		TotalMissing: 3,
		TotalCutoff:  2,
		SuccessCount: 1,
		FailureCount: 0,
	}

	// Set limits
	limits := database.SearchLimits{MissingMoviesLimit: 10, CutoffMoviesLimit: 10}

	// Trigger searches with dry-run
	ctx := context.Background()
	results, err := trigger.TriggerSearches(ctx, detectionResults, limits, true)
	if err != nil {
		t.Fatalf("TriggerSearches failed: %v", err)
	}

	// Check that results are reported
	if results.MissingTriggered != 3 {
		t.Errorf("expected 3 missing triggered (dry-run), got %d", results.MissingTriggered)
	}
	if results.CutoffTriggered != 2 {
		t.Errorf("expected 2 cutoff triggered (dry-run), got %d", results.CutoffTriggered)
	}

	// Check that NO API calls were made
	calls := mockClient.getTriggerCalls()
	if len(calls) > 0 {
		t.Errorf("expected no API calls in dry-run, got %d", len(calls))
	}
}

func TestTriggerSearches_PartialFailure(t *testing.T) {
	db := testTriggerDB(t)

	// Add two servers
	server1, err := db.AddServer("radarr1", "http://localhost:7878", "api1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server 1: %v", err)
	}
	server2, err := db.AddServer("radarr2", "http://localhost:7879", "api2", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server 2: %v", err)
	}

	// Create mock clients - one succeeds, one fails
	mockClient1 := &mockTriggerAPIClient{serverType: "radarr"}
	mockClient2 := &mockTriggerAPIClient{serverType: "radarr", triggerErr: errors.New("API error")}

	mockClients := map[string]*mockTriggerAPIClient{
		"http://localhost:7878": mockClient1,
		"http://localhost:7879": mockClient2,
	}

	// Create SearchTrigger with mock factory
	trigger := NewSearchTriggerWithFactory(db, func(url, apiKey, serverType string) SearchTriggerAPIClient {
		return mockClients[url]
	})

	// Create detection results with items from both servers
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   server1.ID,
				ServerName: "radarr1",
				ServerType: "radarr",
				Missing:    []int{1, 2, 3},
				Cutoff:     []int{},
			},
			{
				ServerID:   server2.ID,
				ServerName: "radarr2",
				ServerType: "radarr",
				Missing:    []int{10, 11, 12},
				Cutoff:     []int{},
			},
		},
		TotalMissing: 6,
		TotalCutoff:  0,
		SuccessCount: 2,
		FailureCount: 0,
	}

	// Set limits
	limits := database.SearchLimits{MissingMoviesLimit: 10, CutoffMoviesLimit: 0}

	// Trigger searches
	ctx := context.Background()
	results, err := trigger.TriggerSearches(ctx, detectionResults, limits, false)
	if err != nil {
		t.Fatalf("TriggerSearches failed: %v", err)
	}

	// Should have partial success
	if results.SuccessCount == 0 {
		t.Error("expected at least one success")
	}
	if results.FailureCount == 0 {
		t.Error("expected at least one failure")
	}

	// Check that we recorded the error
	hasError := false
	for _, r := range results.Results {
		if r.Error != "" {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected at least one result with an error")
	}
}

func TestTriggerSearches_NoResults(t *testing.T) {
	db := testTriggerDB(t)

	// Create mock client
	mockClient := &mockTriggerAPIClient{serverType: "radarr"}

	// Create SearchTrigger with mock factory
	trigger := NewSearchTriggerWithFactory(db, func(url, apiKey, serverType string) SearchTriggerAPIClient {
		return mockClient
	})

	// Create empty detection results
	detectionResults := &DetectionResults{
		Results:      []DetectionResult{},
		TotalMissing: 0,
		TotalCutoff:  0,
		SuccessCount: 0,
		FailureCount: 0,
	}

	// Set limits
	limits := database.SearchLimits{MissingMoviesLimit: 10, CutoffMoviesLimit: 10}

	// Trigger searches
	ctx := context.Background()
	results, err := trigger.TriggerSearches(ctx, detectionResults, limits, false)
	if err != nil {
		t.Fatalf("TriggerSearches failed: %v", err)
	}

	// Should handle gracefully with no results
	if results.MissingTriggered != 0 {
		t.Errorf("expected 0 missing triggered, got %d", results.MissingTriggered)
	}
	if results.CutoffTriggered != 0 {
		t.Errorf("expected 0 cutoff triggered, got %d", results.CutoffTriggered)
	}

	// No API calls should have been made
	calls := mockClient.getTriggerCalls()
	if len(calls) > 0 {
		t.Errorf("expected no API calls for empty results, got %d", len(calls))
	}
}

func TestTriggerSearches_ZeroLimit(t *testing.T) {
	db := testTriggerDB(t)

	// Add a server
	server1, err := db.AddServer("radarr1", "http://localhost:7878", "api1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	// Create mock client
	mockClient := &mockTriggerAPIClient{serverType: "radarr"}

	// Create SearchTrigger with mock factory
	trigger := NewSearchTriggerWithFactory(db, func(url, apiKey, serverType string) SearchTriggerAPIClient {
		return mockClient
	})

	// Create detection results
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   server1.ID,
				ServerName: "radarr1",
				ServerType: "radarr",
				Missing:    []int{1, 2, 3},
				Cutoff:     []int{10, 11},
			},
		},
		TotalMissing: 3,
		TotalCutoff:  2,
		SuccessCount: 1,
		FailureCount: 0,
	}

	// Set limits: 0 for both categories
	limits := database.SearchLimits{MissingMoviesLimit: 0, CutoffMoviesLimit: 0}

	// Trigger searches
	ctx := context.Background()
	results, err := trigger.TriggerSearches(ctx, detectionResults, limits, false)
	if err != nil {
		t.Fatalf("TriggerSearches failed: %v", err)
	}

	// Should not trigger any searches with zero limits
	if results.MissingTriggered != 0 {
		t.Errorf("expected 0 missing triggered with zero limit, got %d", results.MissingTriggered)
	}
	if results.CutoffTriggered != 0 {
		t.Errorf("expected 0 cutoff triggered with zero limit, got %d", results.CutoffTriggered)
	}

	// No API calls should have been made
	calls := mockClient.getTriggerCalls()
	if len(calls) > 0 {
		t.Errorf("expected no API calls with zero limits, got %d", len(calls))
	}
}

func TestTriggerSearches_MixedServerTypes(t *testing.T) {
	db := testTriggerDB(t)

	// Add both radarr and sonarr servers
	radarr, err := db.AddServer("radarr1", "http://localhost:7878", "api1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding radarr server: %v", err)
	}
	sonarr, err := db.AddServer("sonarr1", "http://localhost:8989", "api2", database.ServerTypeSonarr)
	if err != nil {
		t.Fatalf("adding sonarr server: %v", err)
	}

	// Create mock clients
	radarrClient := &mockTriggerAPIClient{serverType: "radarr"}
	sonarrClient := &mockTriggerAPIClient{serverType: "sonarr"}

	mockClients := map[string]*mockTriggerAPIClient{
		"http://localhost:7878": radarrClient,
		"http://localhost:8989": sonarrClient,
	}

	// Create SearchTrigger with mock factory
	trigger := NewSearchTriggerWithFactory(db, func(url, apiKey, serverType string) SearchTriggerAPIClient {
		return mockClients[url]
	})

	// Create detection results with items from both servers
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   radarr.ID,
				ServerName: "radarr1",
				ServerType: "radarr",
				Missing:    []int{1, 2, 3}, // Movies
				Cutoff:     []int{10},
			},
			{
				ServerID:   sonarr.ID,
				ServerName: "sonarr1",
				ServerType: "sonarr",
				Missing:    []int{100, 101, 102, 103}, // Episodes
				Cutoff:     []int{200, 201},
			},
		},
		TotalMissing: 7,
		TotalCutoff:  3,
		SuccessCount: 2,
		FailureCount: 0,
	}

	// Set limits
	limits := database.SearchLimits{MissingMoviesLimit: 10, CutoffMoviesLimit: 10, MissingEpisodesLimit: 10, CutoffEpisodesLimit: 10}

	// Trigger searches
	ctx := context.Background()
	results, err := trigger.TriggerSearches(ctx, detectionResults, limits, false)
	if err != nil {
		t.Fatalf("TriggerSearches failed: %v", err)
	}

	// Check totals
	if results.MissingTriggered != 7 {
		t.Errorf("expected 7 missing triggered, got %d", results.MissingTriggered)
	}
	if results.CutoffTriggered != 3 {
		t.Errorf("expected 3 cutoff triggered, got %d", results.CutoffTriggered)
	}

	// Both clients should have received calls
	radarrCalls := radarrClient.getTriggerCalls()
	sonarrCalls := sonarrClient.getTriggerCalls()

	if len(radarrCalls) == 0 {
		t.Error("expected radarr client to receive trigger calls")
	}
	if len(sonarrCalls) == 0 {
		t.Error("expected sonarr client to receive trigger calls")
	}
}

func TestTriggerSearches_SkipsFailedDetectionServers(t *testing.T) {
	db := testTriggerDB(t)

	// Add a server
	server1, err := db.AddServer("radarr1", "http://localhost:7878", "api1", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("adding server: %v", err)
	}

	// Create mock client
	mockClient := &mockTriggerAPIClient{serverType: "radarr"}

	// Create SearchTrigger with mock factory
	trigger := NewSearchTriggerWithFactory(db, func(url, apiKey, serverType string) SearchTriggerAPIClient {
		return mockClient
	})

	// Create detection results with a failed server (has Error)
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   server1.ID,
				ServerName: "radarr1",
				ServerType: "radarr",
				Missing:    []int{1, 2, 3},
				Cutoff:     []int{10, 11},
				Error:      "detection failed", // This server had an error
			},
		},
		TotalMissing: 3,
		TotalCutoff:  2,
		SuccessCount: 0,
		FailureCount: 1,
	}

	// Set limits
	limits := database.SearchLimits{MissingMoviesLimit: 10, CutoffMoviesLimit: 10}

	// Trigger searches
	ctx := context.Background()
	results, err := trigger.TriggerSearches(ctx, detectionResults, limits, false)
	if err != nil {
		t.Fatalf("TriggerSearches failed: %v", err)
	}

	// Should not trigger searches for servers with detection errors
	if results.MissingTriggered != 0 {
		t.Errorf("expected 0 missing triggered for failed detection, got %d", results.MissingTriggered)
	}
	if results.CutoffTriggered != 0 {
		t.Errorf("expected 0 cutoff triggered for failed detection, got %d", results.CutoffTriggered)
	}

	// No API calls should have been made
	calls := mockClient.getTriggerCalls()
	if len(calls) > 0 {
		t.Errorf("expected no API calls for failed detection, got %d", len(calls))
	}
}
