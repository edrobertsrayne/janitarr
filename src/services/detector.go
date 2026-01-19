package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/user/janitarr/src/api"
	"github.com/user/janitarr/src/database"
)

// DetectorAPIClient is the interface for API clients used by the Detector.
type DetectorAPIClient interface {
	TestConnection(ctx context.Context) (*api.SystemStatus, error)
	GetAllMissing(ctx context.Context) ([]api.MediaItem, error)
	GetAllCutoffUnmet(ctx context.Context) ([]api.MediaItem, error)
	TriggerSearch(ctx context.Context, ids []int) error
}

// DetectorAPIClientFactory creates API clients for detection.
type DetectorAPIClientFactory func(url, apiKey, serverType string) DetectorAPIClient

// defaultDetectorAPIClientFactory creates real API clients.
func defaultDetectorAPIClientFactory(url, apiKey, serverType string) DetectorAPIClient {
	if serverType == "sonarr" {
		return api.NewSonarrClient(url, apiKey)
	}
	return api.NewRadarrClient(url, apiKey)
}

// Detector detects missing content and content below quality cutoff across all servers.
type Detector struct {
	db         *database.DB
	apiFactory DetectorAPIClientFactory
}

// NewDetector creates a new Detector with the given database.
func NewDetector(db *database.DB) *Detector {
	return &Detector{
		db:         db,
		apiFactory: defaultDetectorAPIClientFactory,
	}
}

// NewDetectorWithFactory creates a new Detector with a custom API factory.
// Useful for testing.
func NewDetectorWithFactory(db *database.DB, factory DetectorAPIClientFactory) *Detector {
	return &Detector{
		db:         db,
		apiFactory: factory,
	}
}

// DetectAll runs detection on all enabled servers concurrently.
func (d *Detector) DetectAll(ctx context.Context) (*DetectionResults, error) {
	servers, err := d.db.GetAllServers()
	if err != nil {
		return nil, fmt.Errorf("getting servers: %w", err)
	}

	// Filter to only enabled servers
	var enabledServers []database.Server
	for _, s := range servers {
		if s.Enabled {
			enabledServers = append(enabledServers, s)
		}
	}

	// Return empty results if no servers
	if len(enabledServers) == 0 {
		return &DetectionResults{
			Results:      []DetectionResult{},
			TotalMissing: 0,
			TotalCutoff:  0,
			SuccessCount: 0,
			FailureCount: 0,
		}, nil
	}

	// Run detection on all servers concurrently
	var wg sync.WaitGroup
	resultCh := make(chan DetectionResult, len(enabledServers))

	for _, server := range enabledServers {
		wg.Add(1)
		go func(s database.Server) {
			defer wg.Done()
			result := d.detectServer(ctx, &s)
			resultCh <- result
		}(server)
	}

	// Wait for all detections to complete
	wg.Wait()
	close(resultCh)

	// Aggregate results
	results := &DetectionResults{
		Results: make([]DetectionResult, 0, len(enabledServers)),
	}

	for result := range resultCh {
		results.Results = append(results.Results, result)

		if result.Error != "" {
			results.FailureCount++
		} else {
			results.SuccessCount++
			results.TotalMissing += len(result.Missing)
			results.TotalCutoff += len(result.Cutoff)
		}
	}

	return results, nil
}

// detectServer runs detection on a single server.
func (d *Detector) detectServer(ctx context.Context, server *database.Server) DetectionResult {
	result := DetectionResult{
		ServerID:     server.ID,
		ServerName:   server.Name,
		ServerType:   string(server.Type),
		Missing:      []int{},
		Cutoff:       []int{},
		MissingItems: make(map[int]api.MediaItem),
		CutoffItems:  make(map[int]api.MediaItem),
	}

	client := d.apiFactory(server.URL, server.APIKey, string(server.Type))

	// Get missing items
	missingItems, err := client.GetAllMissing(ctx)
	if err != nil {
		result.Error = fmt.Sprintf("missing detection failed: %v", err)
		return result
	}

	for _, item := range missingItems {
		result.Missing = append(result.Missing, item.ID)
		result.MissingItems[item.ID] = item
	}

	// Get cutoff unmet items
	cutoffItems, err := client.GetAllCutoffUnmet(ctx)
	if err != nil {
		result.Error = fmt.Sprintf("cutoff detection failed: %v", err)
		return result
	}

	for _, item := range cutoffItems {
		result.Cutoff = append(result.Cutoff, item.ID)
		result.CutoffItems[item.ID] = item
	}

	return result
}

// DetectServer runs detection on a single server by ID.
func (d *Detector) DetectServer(ctx context.Context, serverID string) (*DetectionResult, error) {
	server, err := d.db.GetServer(serverID)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	result := d.detectServer(ctx, server)
	return &result, nil
}

// DetectByType runs detection on all enabled servers of a specific type.
func (d *Detector) DetectByType(ctx context.Context, serverType database.ServerType) (*DetectionResults, error) {
	servers, err := d.db.GetServersByType(serverType)
	if err != nil {
		return nil, fmt.Errorf("getting servers by type: %w", err)
	}

	// Filter to only enabled servers
	var enabledServers []database.Server
	for _, s := range servers {
		if s.Enabled {
			enabledServers = append(enabledServers, s)
		}
	}

	// Return empty results if no servers
	if len(enabledServers) == 0 {
		return &DetectionResults{
			Results:      []DetectionResult{},
			TotalMissing: 0,
			TotalCutoff:  0,
			SuccessCount: 0,
			FailureCount: 0,
		}, nil
	}

	// Run detection concurrently
	var wg sync.WaitGroup
	resultCh := make(chan DetectionResult, len(enabledServers))

	for _, server := range enabledServers {
		wg.Add(1)
		go func(s database.Server) {
			defer wg.Done()
			result := d.detectServer(ctx, &s)
			resultCh <- result
		}(server)
	}

	wg.Wait()
	close(resultCh)

	// Aggregate results
	results := &DetectionResults{
		Results: make([]DetectionResult, 0, len(enabledServers)),
	}

	for result := range resultCh {
		results.Results = append(results.Results, result)

		if result.Error != "" {
			results.FailureCount++
		} else {
			results.SuccessCount++
			results.TotalMissing += len(result.Missing)
			results.TotalCutoff += len(result.Cutoff)
		}
	}

	return results, nil
}
