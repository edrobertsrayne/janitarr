package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/janitarr/src/api"
	"github.com/user/janitarr/src/database"
)

// APIClient is an interface for testing server connections.
type APIClient interface {
	TestConnection(ctx context.Context) (*api.SystemStatus, error)
}

// APIClientFactory creates API clients for given URL and API key.
type APIClientFactory func(url, apiKey, serverType string) APIClient

// defaultAPIClientFactory creates real API clients.
func defaultAPIClientFactory(url, apiKey, serverType string) APIClient {
	if serverType == "sonarr" {
		return api.NewSonarrClient(url, apiKey)
	}
	return api.NewRadarrClient(url, apiKey)
}

// ServerManager handles CRUD operations for server configurations.
type ServerManager struct {
	db         *database.DB
	apiFactory APIClientFactory
}

// Ensure ServerManager implements ServerManagerInterface
var _ ServerManagerInterface = (*ServerManager)(nil)

// NewServerManager creates a new ServerManager with the given database.
// This function is assigned to NewServerManagerFunc for testability.
func NewServerManager(db *database.DB) ServerManagerInterface {
	return &ServerManager{
		db:         db,
		apiFactory: defaultAPIClientFactory,
	}
}

// NewServerManagerWithFactory creates a new ServerManager with a custom API factory.
// Useful for testing.
func NewServerManagerWithFactory(db *database.DB, factory APIClientFactory) ServerManagerInterface {
	return &ServerManager{
		db:         db,
		apiFactory: factory,
	}
}

// NewServerManagerFunc is a variable that holds the constructor for ServerManager.
// It can be overridden in tests to inject mock implementations.
var NewServerManagerFunc = NewServerManager

// AddServer adds a new server after validating the configuration and testing the connection.
func (m *ServerManager) AddServer(ctx context.Context, name, url, apiKey, serverType string) (*ServerInfo, error) {
	// Validate inputs
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("server name is required")
	}
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Validate server type
	dbType, err := parseServerType(serverType)
	if err != nil {
		return nil, err
	}

	// Normalize URL
	normalizedURL := api.NormalizeURL(url)

	// Check for duplicate URL+type
	if m.db.ServerExists(normalizedURL, dbType, "") {
		return nil, fmt.Errorf("a %s server with this URL already exists", serverType)
	}

	// Check for duplicate name (will be caught by AddServer, but check early)
	existing, _ := m.db.GetServerByName(name)
	if existing != nil {
		return nil, fmt.Errorf("a server named '%s' already exists", name)
	}

	// Test connection
	client := m.apiFactory(normalizedURL, apiKey, serverType)
	status, err := client.TestConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// Validate server type matches
	expectedApp := "Radarr"
	if serverType == "sonarr" {
		expectedApp = "Sonarr"
	}
	if status.AppName != expectedApp {
		return nil, fmt.Errorf("server is %s, but %s was specified", status.AppName, expectedApp)
	}

	// Save to database
	server, err := m.db.AddServer(name, normalizedURL, apiKey, dbType)
	if err != nil {
		return nil, fmt.Errorf("saving server: %w", err)
	}

	return toServerInfo(server), nil
}

// UpdateServer updates a server's fields. If URL or API key change, connection is re-tested.
func (m *ServerManager) UpdateServer(ctx context.Context, id string, updates ServerUpdate) error {
	// Get current server
	server, err := m.db.GetServer(id)
	if server == nil {
		if err != nil {
			return err
		}
		return fmt.Errorf("server not found: %s", id)
	}

	// Determine new values
	newURL := server.URL
	newAPIKey := server.APIKey
	newName := server.Name

	if updates.URL != nil {
		newURL = api.NormalizeURL(*updates.URL)
	}
	if updates.APIKey != nil {
		newAPIKey = *updates.APIKey
	}
	if updates.Name != nil {
		newName = *updates.Name

		// Check for duplicate name
		if newName != server.Name {
			existing, _ := m.db.GetServerByName(newName)
			if existing != nil && existing.ID != id {
				return fmt.Errorf("a server named '%s' already exists", newName)
			}
		}
	}

	// Check for duplicate URL+type if URL changed
	if newURL != server.URL {
		if m.db.ServerExists(newURL, server.Type, id) {
			return fmt.Errorf("a %s server with this URL already exists", server.Type)
		}
	}

	// Test connection if URL or API key changed
	if newURL != server.URL || newAPIKey != server.APIKey {
		client := m.apiFactory(newURL, newAPIKey, string(server.Type))
		_, err := client.TestConnection(ctx)
		if err != nil {
			return fmt.Errorf("connection failed with new settings: %w", err)
		}
	}

	// Build database update
	dbUpdate := &database.ServerUpdate{}
	if updates.Name != nil {
		dbUpdate.Name = updates.Name
	}
	if updates.URL != nil {
		dbUpdate.URL = &newURL
	}
	if updates.APIKey != nil {
		dbUpdate.APIKey = updates.APIKey
	}

	return m.db.UpdateServer(id, dbUpdate)
}

// RemoveServer removes a server by ID.
func (m *ServerManager) RemoveServer(id string) error {
	deleted, err := m.db.DeleteServer(id)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("server not found: %s", id)
	}
	return nil
}

// TestConnection tests the connection to an existing server.
func (m *ServerManager) TestConnection(ctx context.Context, id string) (*ConnectionResult, error) {
	server, err := m.db.GetServer(id)
	if server == nil {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("server not found: %s", id)
	}

	client := m.apiFactory(server.URL, server.APIKey, string(server.Type))
	status, err := client.TestConnection(ctx)
	if err != nil {
		return &ConnectionResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ConnectionResult{
		Success: true,
		Version: status.Version,
		AppName: status.AppName,
	}, nil
}

// TestNewConnection tests a connection to a new server before saving it.
func (m *ServerManager) TestNewConnection(ctx context.Context, url, apiKey, serverType string) (*ConnectionResult, error) {
	// Validate server type
	serverType = strings.ToLower(serverType)
	if serverType != "radarr" && serverType != "sonarr" {
		return nil, fmt.Errorf("invalid server type: %s", serverType)
	}

	// Normalize URL (remove trailing slash, ensure protocol)
	url = api.NormalizeURL(url)

	// Create API client
	client := m.apiFactory(url, apiKey, serverType)

	// Test connection
	status, err := client.TestConnection(ctx)
	if err != nil {
		return &ConnectionResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ConnectionResult{
		Success: true,
		Version: status.Version,
		AppName: status.AppName,
	}, nil
}

// ListServers returns all servers (without API keys).
func (m *ServerManager) ListServers() ([]ServerInfo, error) {
	servers, err := m.db.GetAllServers()
	if err != nil {
		return nil, err
	}

	result := make([]ServerInfo, len(servers))
	for i, s := range servers {
		result[i] = *toServerInfo(&s)
	}
	return result, nil
}

// GetServer retrieves a server by ID or name.
func (m *ServerManager) GetServer(ctx context.Context, idOrName string) (*ServerInfo, error) {
	// Try by ID first
	server, err := m.db.GetServer(idOrName)
	if err != nil {
		return nil, err
	}

	// Then try by name
	if server == nil {
		server, err = m.db.GetServerByName(idOrName)
		if err != nil {
			return nil, err
		}
	}

	if server == nil {
		return nil, fmt.Errorf("server '%s' not found", idOrName)
	}

	return toServerInfo(server), nil
}

// GetServerWithCredentials retrieves a server by ID or name including the API key.
// Use with caution - only when credentials are needed for API calls.
func (m *ServerManager) GetServerWithCredentials(idOrName string) (*database.Server, error) {
	// Try by ID first
	server, err := m.db.GetServer(idOrName)
	if err != nil {
		return nil, err
	}

	// Then try by name
	if server == nil {
		server, err = m.db.GetServerByName(idOrName)
		if err != nil {
			return nil, err
		}
	}

	if server == nil {
		return nil, fmt.Errorf("server '%s' not found", idOrName)
	}

	return server, nil
}

// GetEnabledServers returns all enabled servers with credentials.
func (m *ServerManager) GetEnabledServers() ([]database.Server, error) {
	servers, err := m.db.GetAllServers()
	if err != nil {
		return nil, err
	}

	var enabled []database.Server
	for _, s := range servers {
		if s.Enabled {
			enabled = append(enabled, s)
		}
	}
	return enabled, nil
}

// SetServerEnabled enables or disables a server.
func (m *ServerManager) SetServerEnabled(id string, enabled bool) error {
	update := &database.ServerUpdate{Enabled: &enabled}
	return m.db.UpdateServer(id, update)
}

// GetServersByType returns all servers of a given type.
func (m *ServerManager) GetServersByType(serverType string) ([]ServerInfo, error) {
	dbType, err := parseServerType(serverType)
	if err != nil {
		return nil, err
	}

	servers, err := m.db.GetServersByType(dbType)
	if err != nil {
		return nil, err
	}

	result := make([]ServerInfo, len(servers))
	for i, s := range servers {
		result[i] = *toServerInfo(&s)
	}
	return result, nil
}

// toServerInfo converts a database.Server to a ServerInfo (without API key).
func toServerInfo(s *database.Server) *ServerInfo {
	return &ServerInfo{
		ID:        s.ID,
		Name:      s.Name,
		URL:       s.URL,
		Type:      string(s.Type),
		Enabled:   s.Enabled,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// parseServerType validates and converts a server type string.
func parseServerType(serverType string) (database.ServerType, error) {
	switch strings.ToLower(serverType) {
	case "radarr":
		return database.ServerTypeRadarr, nil
	case "sonarr":
		return database.ServerTypeSonarr, nil
	default:
		return "", fmt.Errorf("invalid server type '%s': must be 'radarr' or 'sonarr'", serverType)
	}
}
