package services

import (
	"context"
	"errors"
	"fmt"
	"time"

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
// Note: This factory doesn't have access to logger, so API request logging
// is attached separately in triggerForServer if needed.
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
	serverID       string
	serverName     string
	serverType     string
	serverURL      string
	apiKey         string
	missing        []int
	cutoff         []int
	missingItems   map[int]api.MediaItem // Metadata for missing items
	cutoffItems    map[int]api.MediaItem // Metadata for cutoff items
	rateLimitCount int                   // Consecutive 429 errors
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

// allocateItems distributes items across servers using proportional allocation, respecting limits.
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

	// Distribute missing items with proportional allocation
	totalMissingLimit := limits.MissingMoviesLimit + limits.MissingEpisodesLimit
	if totalMissingLimit > 0 {
		s.distributeProportional(detectionResults, allocations, "missing", totalMissingLimit)
	}

	// Distribute cutoff items with proportional allocation
	totalCutoffLimit := limits.CutoffMoviesLimit + limits.CutoffEpisodesLimit
	if totalCutoffLimit > 0 {
		s.distributeProportional(detectionResults, allocations, "cutoff", totalCutoffLimit)
	}

	// Convert map to slice
	result := make([]serverItemAllocation, 0, len(allocations))
	for _, alloc := range allocations {
		result = append(result, *alloc)
	}

	return result
}

// distributeProportional distributes items across servers using largest remainder method.
// Each server receives items proportional to its item count, with a minimum of 1 per server.
func (s *SearchTrigger) distributeProportional(detectionResults *DetectionResults, allocations map[string]*serverItemAllocation, category string, limit int) {
	// Build server item map
	type serverInfo struct {
		serverID string
		items    []int
	}

	var servers []serverInfo
	totalItems := 0

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

		if len(items) > 0 {
			servers = append(servers, serverInfo{
				serverID: result.ServerID,
				items:    items,
			})
			totalItems += len(items)
		}
	}

	if len(servers) == 0 || totalItems == 0 || limit == 0 {
		return
	}

	// Calculate allocations using largest remainder method
	type allocation struct {
		serverID  string
		quota     float64
		floor     int
		remainder float64
		allocated int
	}

	allocatedCounts := make([]allocation, len(servers))
	totalFloor := 0

	for i, srv := range servers {
		// Calculate proportional quota
		quota := float64(limit) * float64(len(srv.items)) / float64(totalItems)
		floor := int(quota)

		// Ensure minimum of 1 per server (unless limit is very small)
		if floor == 0 && limit >= len(servers) {
			floor = 1
		}

		allocatedCounts[i] = allocation{
			serverID:  srv.serverID,
			quota:     quota,
			floor:     floor,
			remainder: quota - float64(floor),
		}
		totalFloor += floor
	}

	// Distribute remainders to servers with largest fractional parts
	remainingSlots := limit - totalFloor
	if remainingSlots > 0 {
		// Sort by remainder (descending)
		for i := 0; i < len(allocatedCounts)-1; i++ {
			for j := i + 1; j < len(allocatedCounts); j++ {
				if allocatedCounts[j].remainder > allocatedCounts[i].remainder {
					allocatedCounts[i], allocatedCounts[j] = allocatedCounts[j], allocatedCounts[i]
				}
			}
		}

		// Give remainder slots to servers with largest fractions
		for i := 0; i < remainingSlots && i < len(allocatedCounts); i++ {
			allocatedCounts[i].floor++
		}
	}

	// Assign actual items to each server based on calculated allocation
	for _, srv := range servers {
		// Find this server's allocation
		var targetCount int
		for _, alloc := range allocatedCounts {
			if alloc.serverID == srv.serverID {
				targetCount = alloc.floor
				break
			}
		}

		// Don't allocate more items than the server has
		if targetCount > len(srv.items) {
			targetCount = len(srv.items)
		}

		// Take the first N items from this server
		itemsToAllocate := srv.items[:targetCount]

		// Add to allocations
		if category == "missing" {
			allocations[srv.serverID].missing = append(allocations[srv.serverID].missing, itemsToAllocate...)
		} else {
			allocations[srv.serverID].cutoff = append(allocations[srv.serverID].cutoff, itemsToAllocate...)
		}
	}
}

// executeAllocations executes the trigger allocations.
func (s *SearchTrigger) executeAllocations(ctx context.Context, allocations []serverItemAllocation, dryRun bool) (*TriggerResults, error) {
	results := &TriggerResults{
		Results: make([]TriggerResult, 0),
	}

	// Track rate limits across allocations (use map for persistence)
	rateLimits := make(map[string]int)
	for i := range allocations {
		rateLimits[allocations[i].serverID] = allocations[i].rateLimitCount
	}

	isFirstBatch := true
	for i := range allocations {
		alloc := &allocations[i]

		// Skip servers that have hit rate limit threshold (3 strikes)
		if rateLimits[alloc.serverID] >= 3 {
			continue
		}

		// Add 100ms delay between batches (but not before first batch)
		if !isFirstBatch && !dryRun {
			time.Sleep(100 * time.Millisecond)
		}
		isFirstBatch = false

		// Handle missing items
		if len(alloc.missing) > 0 {
			result := s.triggerForServer(ctx, *alloc, "missing", alloc.missing, dryRun)
			results.Results = append(results.Results, result)

			if result.Success {
				results.SuccessCount++
				results.MissingTriggered += len(result.ItemIDs)
				// Reset rate limit counter on success
				rateLimits[alloc.serverID] = 0
			} else {
				results.FailureCount++
				// Check if it's a rate limit error
				if result.Error != "" && (result.Error == "rate_limit" || isRateLimitError(result.Error)) {
					rateLimits[alloc.serverID]++
				}
			}
		}

		// Handle cutoff items (only if not rate limited)
		if len(alloc.cutoff) > 0 && rateLimits[alloc.serverID] < 3 {
			if !isFirstBatch && !dryRun {
				time.Sleep(100 * time.Millisecond)
			}
			result := s.triggerForServer(ctx, *alloc, "cutoff", alloc.cutoff, dryRun)
			results.Results = append(results.Results, result)

			if result.Success {
				results.SuccessCount++
				results.CutoffTriggered += len(result.ItemIDs)
				// Reset rate limit counter on success
				rateLimits[alloc.serverID] = 0
			} else {
				results.FailureCount++
				// Check if it's a rate limit error
				if result.Error != "" && (result.Error == "rate_limit" || isRateLimitError(result.Error)) {
					rateLimits[alloc.serverID]++
				}
			}
		}
	}

	return results, nil
}

// isRateLimitError checks if an error message indicates a rate limit error.
func isRateLimitError(errMsg string) bool {
	return errMsg != "" && (errMsg == "rate_limit" || contains(errMsg, "rate limited") || contains(errMsg, "retry after"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
		// Check if it's a rate limit error
		var rateLimitErr *api.RateLimitError
		if errors.As(err, &rateLimitErr) {
			result.Error = "rate_limit"
		} else {
			result.Error = err.Error()
		}
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
