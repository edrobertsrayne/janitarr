// Package api provides clients for interacting with Radarr and Sonarr APIs.
package api

// SystemStatus represents the system status response from Radarr/Sonarr.
type SystemStatus struct {
	AppName      string `json:"appName"`
	Version      string `json:"version"`
	InstanceName string `json:"instanceName,omitempty"`
}

// Movie represents a movie item from Radarr's wanted/missing or cutoff unmet endpoints.
type Movie struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	HasFile   bool   `json:"hasFile"`
	Monitored bool   `json:"monitored"`
}

// Series represents series info nested in Sonarr episode responses.
type Series struct {
	Title string `json:"title"`
}

// Episode represents an episode item from Sonarr's wanted/missing or cutoff unmet endpoints.
type Episode struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	HasFile       bool    `json:"hasFile"`
	Monitored     bool    `json:"monitored"`
	SeriesTitle   string  `json:"seriesTitle,omitempty"`
	Series        *Series `json:"series,omitempty"`
	SeasonNumber  int     `json:"seasonNumber"`
	EpisodeNumber int     `json:"episodeNumber"`
}

// PagedResponse wraps paginated API responses.
type PagedResponse[T any] struct {
	Page         int `json:"page"`
	PageSize     int `json:"pageSize"`
	TotalRecords int `json:"totalRecords"`
	Records      []T `json:"records"`
}

// CommandResponse represents the response from command endpoints.
type CommandResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// MediaItem is a simplified representation of a media item for search operations.
type MediaItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"` // "movie" or "episode"
}
