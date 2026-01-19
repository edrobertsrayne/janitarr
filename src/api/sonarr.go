package api

import (
	"context"
	"fmt"
	"time"
)

// SonarrClient is an API client for Sonarr servers.
type SonarrClient struct {
	*Client
}

// NewSonarrClient creates a new Sonarr API client with default timeout.
func NewSonarrClient(url, apiKey string) *SonarrClient {
	return &SonarrClient{Client: NewClient(url, apiKey)}
}

// NewSonarrClientWithTimeout creates a new Sonarr API client with a custom timeout.
func NewSonarrClientWithTimeout(url, apiKey string, timeout time.Duration) *SonarrClient {
	return &SonarrClient{Client: NewClientWithTimeout(url, apiKey, timeout)}
}

// TestConnection tests the connection to the Sonarr server.
func (c *SonarrClient) TestConnection(ctx context.Context) (*SystemStatus, error) {
	var result SystemStatus
	if err := c.Get(ctx, "/system/status", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMissing returns a paginated list of missing episodes.
func (c *SonarrClient) GetMissing(ctx context.Context, page, pageSize int) (*PagedResponse[Episode], error) {
	var result PagedResponse[Episode]
	endpoint := fmt.Sprintf("/wanted/missing?page=%d&pageSize=%d&sortKey=id&sortDirection=ascending", page, pageSize)
	if err := c.Get(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCutoffUnmet returns a paginated list of episodes not meeting quality cutoff.
func (c *SonarrClient) GetCutoffUnmet(ctx context.Context, page, pageSize int) (*PagedResponse[Episode], error) {
	var result PagedResponse[Episode]
	endpoint := fmt.Sprintf("/wanted/cutoff?page=%d&pageSize=%d&sortKey=id&sortDirection=ascending", page, pageSize)
	if err := c.Get(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TriggerSearch triggers a search for the specified episode IDs.
func (c *SonarrClient) TriggerSearch(ctx context.Context, episodeIDs []int) error {
	body := map[string]any{
		"name":       "EpisodeSearch",
		"episodeIds": episodeIDs,
	}
	var result CommandResponse
	return c.Post(ctx, "/command", body, &result)
}

// GetAllMissing retrieves all missing episodes across all pages.
func (c *SonarrClient) GetAllMissing(ctx context.Context) ([]MediaItem, error) {
	return c.getAllItems(ctx, c.GetMissing)
}

// GetAllCutoffUnmet retrieves all cutoff unmet episodes across all pages.
func (c *SonarrClient) GetAllCutoffUnmet(ctx context.Context) ([]MediaItem, error) {
	return c.getAllItems(ctx, c.GetCutoffUnmet)
}

// getAllItems is a helper to paginate through all items.
func (c *SonarrClient) getAllItems(ctx context.Context, fetcher func(context.Context, int, int) (*PagedResponse[Episode], error)) ([]MediaItem, error) {
	var items []MediaItem
	page := 1
	pageSize := 100

	for {
		result, err := fetcher(ctx, page, pageSize)
		if err != nil {
			return nil, err
		}

		for _, episode := range result.Records {
			seriesTitle := episode.SeriesTitle
			if episode.Series != nil && episode.Series.Title != "" {
				seriesTitle = episode.Series.Title
			}
			qualityProfile := ""
			if episode.Series != nil {
				qualityProfile = episode.Series.QualityProfile.Name
			}

			items = append(items, MediaItem{
				ID:             episode.ID,
				Title:          formatEpisodeTitle(episode),
				EpisodeTitle:   episode.Title, // Raw episode title for logging
				Type:           "episode",
				SeriesTitle:    seriesTitle,
				SeasonNumber:   episode.SeasonNumber,
				EpisodeNumber:  episode.EpisodeNumber,
				QualityProfile: qualityProfile,
			})
		}

		if len(items) >= result.TotalRecords {
			break
		}
		page++
	}

	return items, nil
}

// formatEpisodeTitle formats an episode title like "Series - S01E02 - Episode Title".
func formatEpisodeTitle(ep Episode) string {
	seriesTitle := ep.SeriesTitle
	if ep.Series != nil && ep.Series.Title != "" {
		seriesTitle = ep.Series.Title
	}
	if seriesTitle == "" {
		seriesTitle = "Unknown Series"
	}

	return fmt.Sprintf("%s - S%02dE%02d - %s",
		seriesTitle,
		ep.SeasonNumber,
		ep.EpisodeNumber,
		ep.Title,
	)
}
