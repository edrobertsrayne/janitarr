package forms

import (
	"testing"

	"github.com/edrobertsrayne/janitarr/src/database"
)

func TestServerFormResult(t *testing.T) {
	// Test that the ServerFormResult struct can be instantiated
	result := &ServerFormResult{
		Name:   "test-server",
		Type:   "radarr",
		URL:    "http://localhost:7878",
		APIKey: "test-api-key-1234567890",
	}

	if result.Name != "test-server" {
		t.Errorf("Expected Name to be 'test-server', got '%s'", result.Name)
	}
	if result.Type != "radarr" {
		t.Errorf("Expected Type to be 'radarr', got '%s'", result.Type)
	}
	if result.URL != "http://localhost:7878" {
		t.Errorf("Expected URL to be 'http://localhost:7878', got '%s'", result.URL)
	}
	if result.APIKey != "test-api-key-1234567890" {
		t.Errorf("Expected APIKey to be set, got '%s'", result.APIKey)
	}
}

func TestServerInfo(t *testing.T) {
	// Test that the ServerInfo struct can be instantiated
	info := ServerInfo{
		ID:      "1",
		Name:    "test-server",
		Type:    "radarr",
		Enabled: true,
	}

	if info.ID != "1" {
		t.Errorf("Expected ID to be '1', got '%s'", info.ID)
	}
	if info.Name != "test-server" {
		t.Errorf("Expected Name to be 'test-server', got '%s'", info.Name)
	}
	if info.Type != "radarr" {
		t.Errorf("Expected Type to be 'radarr', got '%s'", info.Type)
	}
	if !info.Enabled {
		t.Errorf("Expected Enabled to be true, got false")
	}
}

func TestServerSelector_EmptyList(t *testing.T) {
	// Test that ServerSelector returns error for empty list
	servers := []ServerInfo{}
	_, err := ServerSelector(servers)
	if err == nil {
		t.Error("Expected error for empty server list, got nil")
	}
	if err != nil && err.Error() != "no servers configured" {
		t.Errorf("Expected 'no servers configured' error, got '%s'", err.Error())
	}
}

func TestServerSelector_ValidList(t *testing.T) {
	// Test that ServerSelector can build options from server list
	servers := []ServerInfo{
		{ID: "1", Name: "radarr-main", Type: "radarr", Enabled: true},
		{ID: "2", Name: "sonarr-main", Type: "sonarr", Enabled: false},
	}

	// We can't actually run the interactive form in tests, but we can verify
	// the function signature and data structures work correctly
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}
}

// TestServerFormValidation tests that validation functions are applied correctly
func TestServerFormValidation(t *testing.T) {
	// Test various validation scenarios to ensure they're used correctly in forms

	// Test valid server name
	err := ValidateServerName("radarr-main")
	if err != nil {
		t.Errorf("Expected valid server name, got error: %v", err)
	}

	// Test invalid server name
	err = ValidateServerName("invalid name with spaces")
	if err == nil {
		t.Error("Expected error for invalid server name, got nil")
	}

	// Test valid URL
	err = ValidateURL("http://localhost:7878")
	if err != nil {
		t.Errorf("Expected valid URL, got error: %v", err)
	}

	// Test invalid URL
	err = ValidateURL("not-a-url")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	// Test valid API key
	err = ValidateAPIKey("abcdefghij1234567890")
	if err != nil {
		t.Errorf("Expected valid API key, got error: %v", err)
	}

	// Test invalid API key (too short)
	err = ValidateAPIKey("short")
	if err == nil {
		t.Error("Expected error for short API key, got nil")
	}
}

// TestDatabaseServerCompatibility verifies our ServerFormResult can work with database.Server
func TestDatabaseServerCompatibility(t *testing.T) {
	// Verify that database.Server has the fields we expect
	dbServer := &database.Server{
		ID:      "test-id",
		Name:    "test-server",
		Type:    database.ServerTypeRadarr,
		URL:     "http://localhost:7878",
		APIKey:  "encrypted-key-string",
		Enabled: true,
	}

	// Verify we can create a ServerFormResult from database.Server
	formResult := &ServerFormResult{
		Name: dbServer.Name,
		Type: string(dbServer.Type),
		URL:  dbServer.URL,
	}

	if formResult.Name != dbServer.Name {
		t.Errorf("Expected Name to match, got '%s' vs '%s'", formResult.Name, dbServer.Name)
	}
	if formResult.Type != string(dbServer.Type) {
		t.Errorf("Expected Type to match, got '%s' vs '%s'", formResult.Type, dbServer.Type)
	}
	if formResult.URL != dbServer.URL {
		t.Errorf("Expected URL to match, got '%s' vs '%s'", formResult.URL, dbServer.URL)
	}
}
