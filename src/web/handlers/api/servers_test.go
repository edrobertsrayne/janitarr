package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edrobertsrayne/janitarr/src/database"
	"github.com/edrobertsrayne/janitarr/src/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// mockServerManager implements ServerManagerInterface for testing
type mockServerManager struct {
	servers          map[string]*services.ServerInfo
	addServerFunc    func(ctx context.Context, name, url, apiKey, serverType string) (*services.ServerInfo, error)
	testNewConnFunc  func(ctx context.Context, url, apiKey, serverType string) (*services.ConnectionResult, error)
	testConnFunc     func(ctx context.Context, id string) (*services.ConnectionResult, error)
	updateServerFunc func(ctx context.Context, id string, updates services.ServerUpdate) error
	removeServerFunc func(id string) error
}

func newMockServerManager() *mockServerManager {
	return &mockServerManager{
		servers: make(map[string]*services.ServerInfo),
	}
}

func (m *mockServerManager) AddServer(ctx context.Context, name, url, apiKey, serverType string) (*services.ServerInfo, error) {
	if m.addServerFunc != nil {
		return m.addServerFunc(ctx, name, url, apiKey, serverType)
	}
	server := &services.ServerInfo{
		ID:        uuid.NewString(),
		Name:      name,
		URL:       url,
		Type:      serverType,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.servers[server.ID] = server
	return server, nil
}

func (m *mockServerManager) UpdateServer(ctx context.Context, id string, updates services.ServerUpdate) error {
	if m.updateServerFunc != nil {
		return m.updateServerFunc(ctx, id, updates)
	}
	server, exists := m.servers[id]
	if !exists {
		return errors.New("server not found")
	}
	if updates.URL != nil {
		server.URL = *updates.URL
	}
	server.UpdatedAt = time.Now()
	return nil
}

func (m *mockServerManager) RemoveServer(id string) error {
	if m.removeServerFunc != nil {
		return m.removeServerFunc(id)
	}
	if _, exists := m.servers[id]; !exists {
		return errors.New("server not found")
	}
	delete(m.servers, id)
	return nil
}

func (m *mockServerManager) TestConnection(ctx context.Context, id string) (*services.ConnectionResult, error) {
	if m.testConnFunc != nil {
		return m.testConnFunc(ctx, id)
	}
	return &services.ConnectionResult{
		Success: true,
		Version: "1.0.0",
		AppName: "Test Server",
	}, nil
}

func (m *mockServerManager) TestNewConnection(ctx context.Context, url, apiKey, serverType string) (*services.ConnectionResult, error) {
	if m.testNewConnFunc != nil {
		return m.testNewConnFunc(ctx, url, apiKey, serverType)
	}
	return &services.ConnectionResult{
		Success: true,
		Version: "1.0.0",
		AppName: "Test Server",
	}, nil
}

func (m *mockServerManager) ListServers() ([]services.ServerInfo, error) {
	var list []services.ServerInfo
	for _, s := range m.servers {
		list = append(list, *s)
	}
	return list, nil
}

func (m *mockServerManager) GetServer(ctx context.Context, idOrName string) (*services.ServerInfo, error) {
	if server, exists := m.servers[idOrName]; exists {
		return server, nil
	}
	return nil, errors.New("server not found")
}

func (m *mockServerManager) GetEnabledServers() ([]database.Server, error) {
	var enabled []database.Server
	for _, s := range m.servers {
		if s.Enabled {
			enabled = append(enabled, database.Server{
				ID:      s.ID,
				Name:    s.Name,
				URL:     s.URL,
				Type:    database.ServerType(s.Type),
				Enabled: s.Enabled,
			})
		}
	}
	return enabled, nil
}

func (m *mockServerManager) SetServerEnabled(id string, enabled bool) error {
	server, exists := m.servers[id]
	if !exists {
		return errors.New("server not found")
	}
	server.Enabled = enabled
	return nil
}

func TestListServers_Empty(t *testing.T) {
	db := testDB(t)
	mockMgr := newMockServerManager()
	handlers := NewServerHandlers(mockMgr, db)

	req := httptest.NewRequest("GET", "/api/servers", nil)
	rr := httptest.NewRecorder()

	handlers.ListServers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp SuccessResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)

	// Extract servers list
	dataBytes, _ := json.Marshal(resp.Data)
	var servers []services.ServerInfo
	json.Unmarshal(dataBytes, &servers)

	if len(servers) != 0 {
		t.Errorf("expected empty list, got %d servers", len(servers))
	}
}

func TestListServers_WithData(t *testing.T) {
	db := testDB(t)
	mockMgr := newMockServerManager()
	handlers := NewServerHandlers(mockMgr, db)

	// Add test servers
	mockMgr.AddServer(context.Background(), "Radarr1", "http://radarr.com", "key1", "radarr")
	mockMgr.AddServer(context.Background(), "Sonarr1", "http://sonarr.com", "key2", "sonarr")

	req := httptest.NewRequest("GET", "/api/servers", nil)
	rr := httptest.NewRecorder()

	handlers.ListServers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp SuccessResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)

	dataBytes, _ := json.Marshal(resp.Data)
	var servers []services.ServerInfo
	json.Unmarshal(dataBytes, &servers)

	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}

func TestCreateServer_Success(t *testing.T) {
	db := testDB(t)
	mockMgr := newMockServerManager()
	handlers := NewServerHandlers(mockMgr, db)

	payload := map[string]string{
		"name":   "TestServer",
		"url":    "http://test.com",
		"apiKey": "testkey",
		"type":   "radarr",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/servers", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handlers.CreateServer(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp SuccessResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)

	dataBytes, _ := json.Marshal(resp.Data)
	var server services.ServerInfo
	json.Unmarshal(dataBytes, &server)

	if server.Name != "TestServer" {
		t.Errorf("expected name TestServer, got %s", server.Name)
	}
}

func TestCreateServer_ConnectionFailed(t *testing.T) {
	db := testDB(t)
	mockMgr := newMockServerManager()
	mockMgr.addServerFunc = func(ctx context.Context, name, url, apiKey, serverType string) (*services.ServerInfo, error) {
		return nil, errors.New("connection failed")
	}
	handlers := NewServerHandlers(mockMgr, db)

	payload := map[string]string{
		"name":   "TestServer",
		"url":    "http://test.com",
		"apiKey": "badkey",
		"type":   "radarr",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/servers", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handlers.CreateServer(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestUpdateServer_Success(t *testing.T) {
	db := testDB(t)
	mockMgr := newMockServerManager()
	handlers := NewServerHandlers(mockMgr, db)

	// Create a server first
	server, _ := mockMgr.AddServer(context.Background(), "Original", "http://old.com", "key", "radarr")

	payload := map[string]string{
		"url": "http://new.com",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/api/servers/"+server.ID, bytes.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", server.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handlers.UpdateServer(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteServer_Success(t *testing.T) {
	db := testDB(t)
	mockMgr := newMockServerManager()
	handlers := NewServerHandlers(mockMgr, db)

	// Create a server first
	server, _ := mockMgr.AddServer(context.Background(), "ToDelete", "http://delete.com", "key", "radarr")

	req := httptest.NewRequest("DELETE", "/api/servers/"+server.ID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", server.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handlers.DeleteServer(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify server was deleted
	_, err := mockMgr.GetServer(context.Background(), server.ID)
	if err == nil {
		t.Error("expected server to be deleted")
	}
}

func TestTestNewServerConnection_Success(t *testing.T) {
	db := testDB(t)
	mockMgr := newMockServerManager()
	handlers := NewServerHandlers(mockMgr, db)

	payload := map[string]string{
		"url":    "http://test.com",
		"apiKey": "testkey",
		"type":   "radarr",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/servers/test", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handlers.TestNewServerConnection(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp SuccessResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)

	dataBytes, _ := json.Marshal(resp.Data)
	var result services.ConnectionResult
	json.Unmarshal(dataBytes, &result)

	if !result.Success {
		t.Error("expected successful connection")
	}
}
