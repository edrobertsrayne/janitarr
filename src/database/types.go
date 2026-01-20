package database

import "time"

// ServerType represents the type of media server
type ServerType string

const (
	ServerTypeRadarr ServerType = "radarr"
	ServerTypeSonarr ServerType = "sonarr"
)

// LogEntryType represents the type of activity log entry
type LogEntryType string

const (
	LogTypeCycleStart LogEntryType = "cycle_start"
	LogTypeCycleEnd   LogEntryType = "cycle_end"
	LogTypeSearch     LogEntryType = "search"
	LogTypeError      LogEntryType = "error"
)

// SearchCategory represents the category of search (missing or cutoff)
type SearchCategory string

const (
	SearchCategoryMissing SearchCategory = "missing"
	SearchCategoryCutoff  SearchCategory = "cutoff"
)

// Server represents a configured media server
type Server struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	URL       string     `json:"url"`
	APIKey    string     `json:"apiKey"`
	Type      ServerType `json:"type"`
	Enabled   bool       `json:"enabled"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// LogEntry represents an activity log entry
type LogEntry struct {
	ID         string         `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	Type       LogEntryType   `json:"type"`
	ServerName string         `json:"serverName,omitempty"`
	ServerType ServerType     `json:"serverType,omitempty"`
	Category   SearchCategory `json:"category,omitempty"`
	Count      int            `json:"count,omitempty"`
	Message    string         `json:"message"`
	IsManual   bool           `json:"isManual"`
}

// ScheduleConfig represents scheduler configuration
type ScheduleConfig struct {
	IntervalHours int  `json:"intervalHours"`
	Enabled       bool `json:"enabled"`
}

// SearchLimits represents search limit configuration
type SearchLimits struct {
	MissingMoviesLimit   int `json:"missingMoviesLimit"`
	MissingEpisodesLimit int `json:"missingEpisodesLimit"`
	CutoffMoviesLimit    int `json:"cutoffMoviesLimit"`
	CutoffEpisodesLimit  int `json:"cutoffEpisodesLimit"`
}

// LogsConfig represents logging configuration
type LogsConfig struct {
	RetentionDays int `json:"retentionDays"`
}

// AppConfig represents the full application configuration
type AppConfig struct {
	Schedule     ScheduleConfig `json:"schedule"`
	SearchLimits SearchLimits   `json:"searchLimits"`
	Logs         LogsConfig     `json:"logs"`
}

// DefaultAppConfig returns the default application configuration
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Schedule: ScheduleConfig{
			IntervalHours: 6,
			Enabled:       true,
		},
		SearchLimits: SearchLimits{
			MissingMoviesLimit:   10,
			MissingEpisodesLimit: 10,
			CutoffMoviesLimit:    5,
			CutoffEpisodesLimit:  5,
		},
		Logs: LogsConfig{
			RetentionDays: 30,
		},
	}
}

// LogFilters represents filters for log queries
type LogFilters struct {
	Type      LogEntryType
	Server    string
	StartDate string
	EndDate   string
	Search    string
}

// ServerStats represents statistics for a single server
type ServerStats struct {
	TotalSearches int    `json:"totalSearches"`
	ErrorCount    int    `json:"errorCount"`
	LastCheckTime string `json:"lastCheckTime,omitempty"`
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalServers    int    `json:"totalServers"`
	LastCycleTime   string `json:"lastCycleTime,omitempty"`
	SearchesLast24h int    `json:"searchesLast24h"`
	ErrorsLast24h   int    `json:"errorsLast24h"`
}

// ServerCounts represents counts of servers by type
type ServerCounts struct {
	Configured int `json:"configured"` // Total servers configured
	Enabled    int `json:"enabled"`    // Servers that are enabled
}
