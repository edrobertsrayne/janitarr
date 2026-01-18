package api

import (
	"context"
	"fmt"
	"time"
)

// RadarrClient is an API client for Radarr servers.
type RadarrClient struct {
	*Client
}

// NewRadarrClient creates a new Radarr API client with default timeout.
func NewRadarrClient(url, apiKey string) *RadarrClient {
	return &RadarrClient{Client: NewClient(url, apiKey)}
}

// NewRadarrClientWithTimeout creates a new Radarr API client with a custom timeout.
func NewRadarrClientWithTimeout(url, apiKey string, timeout time.Duration) *RadarrClient {
	return &RadarrClient{Client: NewClientWithTimeout(url, apiKey, timeout)}
}

// TestConnection tests the connection to the Radarr server.
func (c *RadarrClient) TestConnection(ctx context.Context) (*SystemStatus, error) {
	var result SystemStatus
	if err := c.Get(ctx, "/system/status", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMissing returns a paginated list of missing movies.
func (c *RadarrClient) GetMissing(ctx context.Context, page, pageSize int) (*PagedResponse[Movie], error) {
	var result PagedResponse[Movie]
	endpoint := fmt.Sprintf("/wanted/missing?page=%d&pageSize=%d&sortKey=id&sortDirection=ascending", page, pageSize)
	if err := c.Get(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCutoffUnmet returns a paginated list of movies not meeting quality cutoff.
func (c *RadarrClient) GetCutoffUnmet(ctx context.Context, page, pageSize int) (*PagedResponse[Movie], error) {
	var result PagedResponse[Movie]
	endpoint := fmt.Sprintf("/wanted/cutoff?page=%d&pageSize=%d&sortKey=id&sortDirection=ascending", page, pageSize)
	if err := c.Get(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TriggerSearch triggers a search for the specified movie IDs.
func (c *RadarrClient) TriggerSearch(ctx context.Context, movieIDs []int) error {
	body := map[string]any{
		"name":     "MoviesSearch",
		"movieIds": movieIDs,
	}
	var result CommandResponse
	return c.Post(ctx, "/command", body, &result)
}

// GetAllMissing retrieves all missing movies across all pages.
func (c *RadarrClient) GetAllMissing(ctx context.Context) ([]MediaItem, error) {
	return c.getAllItems(ctx, c.GetMissing)
}

// GetAllCutoffUnmet retrieves all cutoff unmet movies across all pages.
func (c *RadarrClient) GetAllCutoffUnmet(ctx context.Context) ([]MediaItem, error) {
	return c.getAllItems(ctx, c.GetCutoffUnmet)
}

// getAllItems is a helper to paginate through all items.
func (c *RadarrClient) getAllItems(ctx context.Context, fetcher func(context.Context, int, int) (*PagedResponse[Movie], error)) ([]MediaItem, error) {
	var items []MediaItem
	page := 1
	pageSize := 100

	for {
		result, err := fetcher(ctx, page, pageSize)
		if err != nil {
			return nil, err
		}

		for _, movie := range result.Records {
			items = append(items, MediaItem{
				ID:    movie.ID,
				Title: movie.Title,
				Type:  "movie",
			})
		}

		if len(items) >= result.TotalRecords {
			break
		}
		page++
	}

	return items, nil
}
