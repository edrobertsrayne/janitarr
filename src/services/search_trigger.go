package services

import (
	"context"
	"fmt"

	"github.com/user/janitarr/src/api"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
)

// SearchTriggerAPIClient is the interface for API clients used by the SearchTrigger.
type SearchTriggerAPIClient interface {
	TestConnection(ctx context.Context) (*api.SystemStatus, error)
	GetAllMissing(ctx context.Context) ([]api.MediaItem, error)
	GetAllCutoffUnmet(ctx context.Context) ([]api.MediaItem, error)
	TriggerSearch(ctx context.Context, ids []int) error
}

// SearchTriggerAPIClientFactory creates API clients for search triggering.
type SearchTriggerAPIClientFactory func(url, apiKey, serverType string) SearchTriggerAPIClient

// defaultSearchTriggerAPIClientFactory creates real API clients.
func defaultSearchTriggerAPIClientFactory(url, apiKey, serverType string) SearchTriggerAPIClient {
	if serverType == "sonarr" {
		return api.NewSonarrClient(url, apiKey)
	}
	return api.NewRadarrClient(url, apiKey)
}

// SearchTriggerLogger is the interface for logging search operations.
type SearchTriggerLogger interface {
	LogMovieSearch(serverName, serverType, title string, year int, qualityProfile, category string) *logger.LogEntry
	LogEpisodeSearch(serverName, serverType, seriesTitle, episodeTitle string, season, episode int, qualityProfile, category string) *logger.LogEntry
}

// SearchTrigger triggers searches for missing and cutoff content.
type SearchTrigger struct {
	db         *database.DB
	apiFactory SearchTriggerAPIClientFactory
	logger     SearchTriggerLogger
}

// NewSearchTrigger creates a new SearchTrigger with the given database.
func NewSearchTrigger(db *database.DB, logger SearchTriggerLogger) *SearchTrigger {
	return &SearchTrigger{
		db:         db,
		apiFactory: defaultSearchTriggerAPIClientFactory,
		logger:     logger,
	}
}

// NewSearchTriggerWithFactory creates a new SearchTrigger with a custom API factory.
// Useful for testing.
func NewSearchTriggerWithFactory(db *database.DB, factory SearchTriggerAPIClientFactory, logger SearchTriggerLogger) *SearchTrigger {
	return &SearchTrigger{
		db:         db,
		apiFactory: factory,
		logger:     logger,
	}
}

// serverItemAllocation tracks items to be triggered for a server.
type serverItemAllocation struct {
	serverID     string
	serverName   string
	serverType   string
	serverURL    string
	apiKey       string
	missing      []int
	cutoff       []int
	missingItems map[int]api.MediaItem // Metadata for missing items
	cutoffItems  map[int]api.MediaItem // Metadata for cutoff items
}

// TriggerSearches triggers searches based on detection results and limits.
// If dryRun is true, it returns what would be searched without making API calls.
func (s *SearchTrigger) TriggerSearches(ctx context.Context, detectionResults *DetectionResults, limits database.SearchLimits, dryRun bool) (*TriggerResults, error) {
	// Get all servers for URL/API key lookup
	servers, err := s.db.GetAllServers()
	if err != nil {
		return nil, fmt.Errorf("getting servers: %w", err)
	}

	// Create a map of server ID to server for quick lookup
	serverMap := make(map[string]*database.Server)
	for i := range servers {
		serverMap[servers[i].ID] = &servers[i]
	}

	// Allocate items to servers respecting limits and using round-robin distribution
	allocations := s.allocateItems(detectionResults, serverMap, limits)

	// Execute triggers (or simulate in dry-run mode)
	return s.executeAllocations(ctx, allocations, dryRun)
}

// allocateItems distributes items across servers using round-robin, respecting limits.
func (s *SearchTrigger) allocateItems(detectionResults *DetectionResults, serverMap map[string]*database.Server, limits database.SearchLimits) []serverItemAllocation {
	// Initialize allocations for each server with successful detection
	allocations := make(map[string]*serverItemAllocation)

	for _, result := range detectionResults.Results {
		// Skip servers with detection errors
		if result.Error != "" {
			continue
		}

		server, ok := serverMap[result.ServerID]
		if !ok {
			continue // Server not found, skip
		}

		allocations[result.ServerID] = &serverItemAllocation{
			serverID:     result.ServerID,
			serverName:   result.ServerName,
			serverType:   result.ServerType,
			serverURL:    server.URL,
			apiKey:       server.APIKey,
			missing:      []int{},
			cutoff:       []int{},
			missingItems: result.MissingItems,
			cutoffItems:  result.CutoffItems,
		}
	}

	// Distribute missing items with round-robin
	totalMissingLimit := limits.MissingMoviesLimit + limits.MissingEpisodesLimit
	if totalMissingLimit > 0 {
		s.distributeRoundRobin(detectionResults, allocations, "missing", totalMissingLimit)
	}

	// Distribute cutoff items with round-robin
	totalCutoffLimit := limits.CutoffMoviesLimit + limits.CutoffEpisodesLimit
	if totalCutoffLimit > 0 {
		s.distributeRoundRobin(detectionResults, allocations, "cutoff", totalCutoffLimit)
	}

	// Convert map to slice
	result := make([]serverItemAllocation, 0, len(allocations))
	for _, alloc := range allocations {
		result = append(result, *alloc)
	}

	return result
}

// distributeRoundRobin distributes items across servers in round-robin fashion.
func (s *SearchTrigger) distributeRoundRobin(detectionResults *DetectionResults, allocations map[string]*serverItemAllocation, category string, limit int) {
	// Collect all items with their server IDs
	type itemWithServer struct {
		serverID string
		itemID   int
	}

	var allItems []itemWithServer
	for _, result := range detectionResults.Results {
		// Skip servers with errors or not in allocations
		if result.Error != "" {
			continue
		}
		if _, ok := allocations[result.ServerID]; !ok {
			continue
		}

		var items []int
		if category == "missing" {
			items = result.Missing
		} else {
			items = result.Cutoff
		}

		for _, itemID := range items {
			allItems = append(allItems, itemWithServer{
				serverID: result.ServerID,
				itemID:   itemID,
			})
		}
	}

	// Get list of server IDs for round-robin
	serverIDs := make([]string, 0, len(allocations))
	for serverID := range allocations {
		serverIDs = append(serverIDs, serverID)
	}

	if len(serverIDs) == 0 || len(allItems) == 0 {
		return
	}

	// Round-robin distribution
	distributed := 0
	serverIndex := 0

	for distributed < limit && len(allItems) > 0 {
		serverID := serverIDs[serverIndex%len(serverIDs)]

		// Find next available item for this server
		found := false
		for i, item := range allItems {
			if item.serverID == serverID {
				// Add to allocation
				if category == "missing" {
					allocations[serverID].missing = append(allocations[serverID].missing, item.itemID)
				} else {
					allocations[serverID].cutoff = append(allocations[serverID].cutoff, item.itemID)
				}

				// Remove from available items
				allItems = append(allItems[:i], allItems[i+1:]...)
				distributed++
				found = true
				break
			}
		}

		// If no item found for this server, remove it from rotation
		if !found {
			// Find and remove this server from serverIDs
			for i, id := range serverIDs {
				if id == serverID {
					serverIDs = append(serverIDs[:i], serverIDs[i+1:]...)
					break
				}
			}
			// Don't increment serverIndex since we removed a server
			if len(serverIDs) == 0 {
				break
			}
			continue
		}

		serverIndex++
	}
} // Correct closing brace for distributeRoundRobin

// executeAllocations executes the trigger allocations.
func (s *SearchTrigger) executeAllocations(ctx context.Context, allocations []serverItemAllocation, dryRun bool) (*TriggerResults, error) {
	results := &TriggerResults{
		Results: make([]TriggerResult, 0),
	}

	for _, alloc := range allocations {
		// Handle missing items
		if len(alloc.missing) > 0 {
			result := s.triggerForServer(ctx, alloc, "missing", alloc.missing, dryRun)
			results.Results = append(results.Results, result)

			if result.Success {
				results.SuccessCount++
				results.MissingTriggered += len(result.ItemIDs)
			} else {
				results.FailureCount++
			}
		}

		// Handle cutoff items
		if len(alloc.cutoff) > 0 {
			result := s.triggerForServer(ctx, alloc, "cutoff", alloc.cutoff, dryRun)
			results.Results = append(results.Results, result)

			if result.Success {
				results.SuccessCount++
				results.CutoffTriggered += len(result.ItemIDs)
			} else {
				results.FailureCount++
			}
		}
	}

	return results, nil
}

// triggerForServer triggers a search for items on a specific server.
func (s *SearchTrigger) triggerForServer(ctx context.Context, alloc serverItemAllocation, category string, itemIDs []int, dryRun bool) TriggerResult {
	result := TriggerResult{
		ServerID:   alloc.serverID,
		ServerName: alloc.serverName,
		ServerType: alloc.serverType,
		Category:   category,
		ItemIDs:    itemIDs,
		Success:    true,
	}

	// Get the item metadata map based on category
	var itemMetadata map[int]api.MediaItem
	if category == "missing" {
		itemMetadata = alloc.missingItems
	} else {
		itemMetadata = alloc.cutoffItems
	}

	// Log each item individually before triggering the search
	if s.logger != nil && !dryRun {
		for _, itemID := range itemIDs {
			item, ok := itemMetadata[itemID]
			if !ok {
				continue // Skip if metadata not available
			}

			if item.Type == "movie" {
				s.logger.LogMovieSearch(alloc.serverName, alloc.serverType, item.Title, item.Year, item.QualityProfile, category)
			} else if item.Type == "episode" {
				s.logger.LogEpisodeSearch(alloc.serverName, alloc.serverType, item.SeriesTitle, item.EpisodeTitle, item.SeasonNumber, item.EpisodeNumber, item.QualityProfile, category)
			}
		}
	}

	// In dry-run mode, don't make actual API calls
	if dryRun {
		return result
	}

	// Create API client and trigger search
	client := s.apiFactory(alloc.serverURL, alloc.apiKey, alloc.serverType)
	if err := client.TriggerSearch(ctx, itemIDs); err != nil {
		result.Success = false
		result.Error = err.Error()
	}

	return result
}

// DistributeSearches is a utility function that distributes a slice of item IDs
// across a given number of servers using round-robin.
func DistributeSearches(items []int, limit int, serverCount int) [][]int {
	if serverCount <= 0 || limit <= 0 || len(items) == 0 {
		return nil
	}

	// Initialize result slices
	result := make([][]int, serverCount)
	for i := range result {
		result[i] = []int{}
	}

	// Distribute items round-robin
	distributed := 0
	for i, itemID := range items {
		if distributed >= limit {
			break
		}
		serverIdx := i % serverCount
		result[serverIdx] = append(result[serverIdx], itemID)
		distributed++
	}

	return result
}
