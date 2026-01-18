package services

import "time"

// ServerInfo represents a server for display (without API key).
type ServerInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Type      string    `json:"type"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ServerUpdate represents optional fields for updating a server.
type ServerUpdate struct {
	Name   *string `json:"name,omitempty"`
	URL    *string `json:"url,omitempty"`
	APIKey *string `json:"apiKey,omitempty"`
}

// ConnectionResult represents the result of testing a server connection.
type ConnectionResult struct {
	Success bool   `json:"success"`
	Version string `json:"version,omitempty"`
	AppName string `json:"appName,omitempty"`
	Error   string `json:"error,omitempty"`
}

// DetectionResult represents detection results for a single server.
type DetectionResult struct {
	ServerID   string `json:"serverId"`
	ServerName string `json:"serverName"`
	ServerType string `json:"serverType"`
	Missing    []int  `json:"missing"`
	Cutoff     []int  `json:"cutoff"`
	Error      string `json:"error,omitempty"`
}

// DetectionResults represents aggregated detection results.
type DetectionResults struct {
	Results      []DetectionResult `json:"results"`
	TotalMissing int               `json:"totalMissing"`
	TotalCutoff  int               `json:"totalCutoff"`
	SuccessCount int               `json:"successCount"`
	FailureCount int               `json:"failureCount"`
}

// TriggerResult represents the result of triggering searches for one category on one server.
type TriggerResult struct {
	ServerID   string `json:"serverId"`
	ServerName string `json:"serverName"`
	ServerType string `json:"serverType"`
	Category   string `json:"category"` // "missing" or "cutoff"
	ItemIDs    []int  `json:"itemIDs"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}

// TriggerResults represents aggregated trigger results.
type TriggerResults struct {
	Results          []TriggerResult `json:"results"`
	MissingTriggered int             `json:"missingTriggered"`
	CutoffTriggered  int             `json:"cutoffTriggered"`
	SuccessCount     int             `json:"successCount"`
	FailureCount     int             `json:"failureCount"`
}

// SchedulerStatus represents the current state of the scheduler.
type SchedulerStatus struct {
	IsRunning     bool       `json:"isRunning"`
	IsCycleActive bool       `json:"isCycleActive"`
	NextRun       *time.Time `json:"nextRun,omitempty"`
	LastRun       *time.Time `json:"lastRun,omitempty"`
	IntervalHours int        `json:"intervalHours"`
}

// CycleResult represents the result of an automation cycle.
type CycleResult struct {
	Success          bool             `json:"success"`
	DetectionResults DetectionResults `json:"detectionResults"`
	SearchResults    TriggerResults   `json:"searchResults"`
	TotalSearches    int              `json:"totalSearches"`
	TotalFailures    int              `json:"totalFailures"`
	Errors           []string         `json:"errors"`
	Duration         time.Duration    `json:"duration"`
}

// ServerManagerInterface defines the interface for the ServerManager service.
type ServerManagerInterface interface {
	AddServer(ctx context.Context, name, url, apiKey, serverType string) (*ServerInfo, error)
	UpdateServer(ctx context.Context, id string, updates ServerUpdate) error
	RemoveServer(id string) error
	TestConnection(ctx context.Context, id string) (*ConnectionResult, error)
	ListServers() ([]ServerInfo, error)
	GetServer(ctx context.Context, idOrName string) (*ServerInfo, error)
}

// StringPtr is a helper function to return a pointer to a string.
func StringPtr(s string) *string {
	return &s
}