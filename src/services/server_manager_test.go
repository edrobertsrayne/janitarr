package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/janitarr/src/api"
	"github.com/user/janitarr/src/database"
)

// testDB creates a test database with an in-memory SQLite database.
func testDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.New(":memory:", t.TempDir()+"/.key")
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// mockRadarrServer creates a mock Radarr server for testing.
func mockRadarrServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/system/status" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"appName": "Radarr", "version": "4.7.5.7809"}`))
			return
		}
		http.NotFound(w, r)
	}))
}

// mockSonarrServer creates a mock Sonarr server for testing.
func mockSonarrServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/system/status" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"appName": "Sonarr", "version": "3.0.10.1567"}`))
			return
		}
		http.NotFound(w, r)
	}))
}

// mockFailingServer creates a mock server that always returns an error.
func mockFailingServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
}

// mockUnauthorizedServer creates a mock server that returns 401.
func mockUnauthorizedServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
}

func TestAddServer_Success(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	info, err := mgr.AddServer(context.Background(), "Test Radarr", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Name != "Test Radarr" {
		t.Errorf("expected name 'Test Radarr', got '%s'", info.Name)
	}
	if info.Type != "radarr" {
		t.Errorf("expected type 'radarr', got '%s'", info.Type)
	}
	if info.URL != api.NormalizeURL(server.URL) {
		t.Errorf("expected URL '%s', got '%s'", api.NormalizeURL(server.URL), info.URL)
	}
	if !info.Enabled {
		t.Error("expected server to be enabled")
	}
	if info.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestAddServer_Sonarr(t *testing.T) {
	db := testDB(t)
	server := mockSonarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	info, err := mgr.AddServer(context.Background(), "Test Sonarr", server.URL, "test-api-key", "sonarr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Name != "Test Sonarr" {
		t.Errorf("expected name 'Test Sonarr', got '%s'", info.Name)
	}
	if info.Type != "sonarr" {
		t.Errorf("expected type 'sonarr', got '%s'", info.Type)
	}
}

func TestAddServer_DuplicateName(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	// Add first server
	_, err := mgr.AddServer(context.Background(), "Test Server", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding first server: %v", err)
	}

	// Try to add server with same name
	_, err = mgr.AddServer(context.Background(), "Test Server", server.URL+"/other", "test-api-key", "radarr")
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
}

func TestAddServer_DuplicateURLType(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	// Add first server
	_, err := mgr.AddServer(context.Background(), "First Radarr", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding first server: %v", err)
	}

	// Try to add server with same URL and type but different name
	_, err = mgr.AddServer(context.Background(), "Second Radarr", server.URL, "test-api-key", "radarr")
	if err == nil {
		t.Fatal("expected error for duplicate URL+type, got nil")
	}
}

func TestAddServer_SameURLDifferentType(t *testing.T) {
	db := testDB(t)
	radarrServer := mockRadarrServer()
	sonarrServer := mockSonarrServer()
	defer radarrServer.Close()
	defer sonarrServer.Close()

	mgr := NewServerManager(db)

	// Add Radarr server
	_, err := mgr.AddServer(context.Background(), "Test Radarr", radarrServer.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding radarr: %v", err)
	}

	// Add Sonarr server with different URL should succeed
	_, err = mgr.AddServer(context.Background(), "Test Sonarr", sonarrServer.URL, "test-api-key", "sonarr")
	if err != nil {
		t.Fatalf("unexpected error adding sonarr with different URL: %v", err)
	}
}

func TestAddServer_ConnectionFailed(t *testing.T) {
	db := testDB(t)
	server := mockFailingServer()
	defer server.Close()

	mgr := NewServerManager(db)

	_, err := mgr.AddServer(context.Background(), "Bad Server", server.URL, "test-api-key", "radarr")
	if err == nil {
		t.Fatal("expected error for failed connection, got nil")
	}
}

func TestAddServer_WrongServerType(t *testing.T) {
	db := testDB(t)
	// Mock Radarr server but try to add as Sonarr
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	_, err := mgr.AddServer(context.Background(), "Wrong Type", server.URL, "test-api-key", "sonarr")
	if err == nil {
		t.Fatal("expected error for wrong server type, got nil")
	}
}

func TestUpdateServer_Success(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	// Add a server first
	info, err := mgr.AddServer(context.Background(), "Original Name", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Update the server's name
	newName := "Updated Name"
	err = mgr.UpdateServer(context.Background(), info.ID, ServerUpdate{Name: &newName})
	if err != nil {
		t.Fatalf("unexpected error updating server: %v", err)
	}

	// Verify the update
	updated, err := mgr.GetServer(info.ID)
	if err != nil {
		t.Fatalf("unexpected error getting server: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", updated.Name)
	}
}

func TestUpdateServer_NotFound(t *testing.T) {
	db := testDB(t)
	mgr := NewServerManager(db)

	newName := "Updated Name"
	err := mgr.UpdateServer(context.Background(), "nonexistent-id", ServerUpdate{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found server, got nil")
	}
}

func TestRemoveServer_Success(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	// Add a server
	info, err := mgr.AddServer(context.Background(), "To Delete", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Remove the server
	err = mgr.RemoveServer(info.ID)
	if err != nil {
		t.Fatalf("unexpected error removing server: %v", err)
	}

	// Verify it's gone
	_, err = mgr.GetServer(info.ID)
	if err == nil {
		t.Fatal("expected error getting removed server, got nil")
	}
}

func TestRemoveServer_NotFound(t *testing.T) {
	db := testDB(t)
	mgr := NewServerManager(db)

	err := mgr.RemoveServer("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for not found server, got nil")
	}
}

func TestTestConnection_Success(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	// Add a server first
	info, err := mgr.AddServer(context.Background(), "Test Server", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Test the connection
	result, err := mgr.TestConnection(context.Background(), info.ID)
	if err != nil {
		t.Fatalf("unexpected error testing connection: %v", err)
	}

	if !result.Success {
		t.Error("expected successful connection")
	}
	if result.Version != "4.7.5.7809" {
		t.Errorf("expected version '4.7.5.7809', got '%s'", result.Version)
	}
	if result.AppName != "Radarr" {
		t.Errorf("expected appName 'Radarr', got '%s'", result.AppName)
	}
}

func TestTestConnection_Unauthorized(t *testing.T) {
	db := testDB(t)
	mockServer := mockUnauthorizedServer()
	defer mockServer.Close()

	mgr := NewServerManager(db)

	// We need to bypass the normal AddServer which would fail connection test
	// Add server directly to database
	_, err := db.AddServer("Test Server", mockServer.URL, "bad-key", database.ServerTypeRadarr)
	if err != nil {
		t.Fatalf("unexpected error adding server to db: %v", err)
	}

	servers, _ := mgr.ListServers()
	if len(servers) == 0 {
		t.Fatal("expected at least one server")
	}

	result, err := mgr.TestConnection(context.Background(), servers[0].ID)
	if err != nil {
		t.Fatalf("unexpected error testing connection: %v", err)
	}

	if result.Success {
		t.Error("expected failed connection")
	}
	if result.Error == "" {
		t.Error("expected error message in result")
	}
}

func TestGetServer_ByID(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	// Add a server
	info, err := mgr.AddServer(context.Background(), "Find By ID", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Get by ID
	found, err := mgr.GetServer(info.ID)
	if err != nil {
		t.Fatalf("unexpected error getting server by ID: %v", err)
	}
	if found.ID != info.ID {
		t.Errorf("expected ID '%s', got '%s'", info.ID, found.ID)
	}
	if found.Name != "Find By ID" {
		t.Errorf("expected name 'Find By ID', got '%s'", found.Name)
	}
}

func TestGetServer_ByName(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	// Add a server
	info, err := mgr.AddServer(context.Background(), "Find By Name", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Get by name
	found, err := mgr.GetServer("Find By Name")
	if err != nil {
		t.Fatalf("unexpected error getting server by name: %v", err)
	}
	if found.ID != info.ID {
		t.Errorf("expected ID '%s', got '%s'", info.ID, found.ID)
	}
}

func TestGetServer_ByNameCaseInsensitive(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	// Add a server
	info, err := mgr.AddServer(context.Background(), "Case Test", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Get by name with different case
	found, err := mgr.GetServer("case test")
	if err != nil {
		t.Fatalf("unexpected error getting server by lowercase name: %v", err)
	}
	if found.ID != info.ID {
		t.Errorf("expected ID '%s', got '%s'", info.ID, found.ID)
	}
}

func TestGetServer_NotFound(t *testing.T) {
	db := testDB(t)
	mgr := NewServerManager(db)

	_, err := mgr.GetServer("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found server, got nil")
	}
}

func TestListServers(t *testing.T) {
	db := testDB(t)
	radarrServer := mockRadarrServer()
	sonarrServer := mockSonarrServer()
	defer radarrServer.Close()
	defer sonarrServer.Close()

	mgr := NewServerManager(db)

	// Initially empty
	servers, err := mgr.ListServers()
	if err != nil {
		t.Fatalf("unexpected error listing servers: %v", err)
	}
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}

	// Add servers
	_, err = mgr.AddServer(context.Background(), "Radarr 1", radarrServer.URL, "key1", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	_, err = mgr.AddServer(context.Background(), "Sonarr 1", sonarrServer.URL, "key2", "sonarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// List should return both
	servers, err = mgr.ListServers()
	if err != nil {
		t.Fatalf("unexpected error listing servers: %v", err)
	}
	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}

func TestUpdateServer_URLRequiresConnectionTest(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	failingServer := mockFailingServer()
	defer server.Close()
	defer failingServer.Close()

	mgr := NewServerManager(db)

	// Add a server
	info, err := mgr.AddServer(context.Background(), "Test Server", server.URL, "test-api-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Try to update URL to a failing server
	badURL := failingServer.URL
	err = mgr.UpdateServer(context.Background(), info.ID, ServerUpdate{URL: &badURL})
	if err == nil {
		t.Fatal("expected error for failed connection on URL update, got nil")
	}
}

func TestUpdateServer_APIKeyRequiresConnectionTest(t *testing.T) {
	db := testDB(t)
	mockServer := mockUnauthorizedServer()
	realServer := mockRadarrServer()
	defer mockServer.Close()
	defer realServer.Close()

	mgr := NewServerManager(db)

	// Add a server with working connection
	info, err := mgr.AddServer(context.Background(), "Test Server", realServer.URL, "good-key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Update to use the unauthorized server's URL to simulate bad key
	badURL := mockServer.URL
	err = mgr.UpdateServer(context.Background(), info.ID, ServerUpdate{URL: &badURL})
	if err == nil {
		t.Fatal("expected error for unauthorized connection on update, got nil")
	}
}

func TestGetEnabledServers(t *testing.T) {
	db := testDB(t)
	radarrServer := mockRadarrServer()
	defer radarrServer.Close()

	mgr := NewServerManager(db)

	// Add a server
	info, err := mgr.AddServer(context.Background(), "Test Server", radarrServer.URL, "key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Should be enabled by default
	enabled, err := mgr.GetEnabledServers()
	if err != nil {
		t.Fatalf("unexpected error getting enabled servers: %v", err)
	}
	if len(enabled) != 1 {
		t.Errorf("expected 1 enabled server, got %d", len(enabled))
	}

	// Disable the server
	isEnabled := false
	err = mgr.SetServerEnabled(info.ID, isEnabled)
	if err != nil {
		t.Fatalf("unexpected error disabling server: %v", err)
	}

	// Should now have no enabled servers
	enabled, err = mgr.GetEnabledServers()
	if err != nil {
		t.Fatalf("unexpected error getting enabled servers: %v", err)
	}
	if len(enabled) != 0 {
		t.Errorf("expected 0 enabled servers, got %d", len(enabled))
	}
}

func TestSetServerEnabled(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	// Add a server (enabled by default)
	info, err := mgr.AddServer(context.Background(), "Test Server", server.URL, "key", "radarr")
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	// Verify enabled
	serverInfo, _ := mgr.GetServer(info.ID)
	if !serverInfo.Enabled {
		t.Error("expected server to be enabled by default")
	}

	// Disable
	err = mgr.SetServerEnabled(info.ID, false)
	if err != nil {
		t.Fatalf("unexpected error disabling server: %v", err)
	}

	// Verify disabled
	serverInfo, _ = mgr.GetServer(info.ID)
	if serverInfo.Enabled {
		t.Error("expected server to be disabled")
	}

	// Re-enable
	err = mgr.SetServerEnabled(info.ID, true)
	if err != nil {
		t.Fatalf("unexpected error enabling server: %v", err)
	}

	// Verify enabled again
	serverInfo, _ = mgr.GetServer(info.ID)
	if !serverInfo.Enabled {
		t.Error("expected server to be enabled")
	}
}

func TestEmptyNameValidation(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	_, err := mgr.AddServer(context.Background(), "", server.URL, "test-api-key", "radarr")
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestEmptyURLValidation(t *testing.T) {
	db := testDB(t)
	mgr := NewServerManager(db)

	_, err := mgr.AddServer(context.Background(), "Test Server", "", "test-api-key", "radarr")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestEmptyAPIKeyValidation(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	_, err := mgr.AddServer(context.Background(), "Test Server", server.URL, "", "radarr")
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

func TestInvalidServerType(t *testing.T) {
	db := testDB(t)
	server := mockRadarrServer()
	defer server.Close()

	mgr := NewServerManager(db)

	_, err := mgr.AddServer(context.Background(), "Test Server", server.URL, "test-api-key", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid server type, got nil")
	}
}
