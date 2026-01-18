package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/services"
)

// MockServerManager is a mock implementation of services.ServerManagerInterface for testing.
type MockServerManager struct {
	mock.Mock
}

func (m *MockServerManager) AddServer(ctx context.Context, name, url, apiKey, serverType string) (*services.ServerInfo, error) {
	args := m.Called(ctx, name, url, apiKey, serverType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ServerInfo), args.Error(1)
}

func (m *MockServerManager) UpdateServer(ctx context.Context, id string, updates services.ServerUpdate) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockServerManager) RemoveServer(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockServerManager) TestConnection(ctx context.Context, idOrURL string) (*services.ConnectionResult, error) {
	args := m.Called(ctx, idOrURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ConnectionResult), args.Error(1)
}

func (m *MockServerManager) ListServers() ([]services.ServerInfo, error) {
	args := m.Called()
	return args.Get(0).([]services.ServerInfo), args.Error(1)
}

func (m *MockServerManager) GetServer(ctx context.Context, idOrName string) (*services.ServerInfo, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ServerInfo), args.Error(1)
}

func TestServerHandlers(t *testing.T) {
	assert := assert.New(t)

	mockServerManager := new(MockServerManager)
	handlers := NewServerHandlers(mockServerManager, nil) // DB is not directly used by handlers for this set of tests

	router := chi.NewRouter()
	router.Get("/servers", handlers.ListServers)
	router.Post("/servers", handlers.CreateServer)
	router.Post("/servers/test", handlers.TestNewServerConnection)
	router.Route("/servers/{id}", func(r chi.Router) {
		r.Get("/", handlers.GetServer)
		r.Put("/", handlers.UpdateServer)
		r.Delete("/", handlers.DeleteServer)
		r.Post("/test", handlers.TestServerConnection)
	})

	t.Run("ListServers - success with data", func(t *testing.T) {
		expectedServers := []services.ServerInfo{
			{ID: uuid.NewString(), Name: "Radarr1", Type: "radarr", URL: "http://radarr.com"},
			{ID: uuid.NewString(), Name: "Sonarr1", Type: "sonarr", URL: "http://sonarr.com"},
		}
		mockServerManager.On("ListServers").Return(expectedServers, nil).Once()

		req, _ := http.NewRequest("GET", "/servers", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualServers []services.ServerInfo
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualServers) // Convert resp.Data to JSON string then unmarshal
		assert.Equal(expectedServers, actualServers)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("ListServers - no data", func(t *testing.T) {
		mockServerManager.On("ListServers").Return([]services.ServerInfo{}, nil).Once()

		req, _ := http.NewRequest("GET", "/servers", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualServers []services.ServerInfo
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualServers)
		assert.Empty(actualServers)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("ListServers - error", func(t *testing.T) {
		mockServerManager.On("ListServers").Return(([]services.ServerInfo)(nil), errors.New("db error")).Once()

		req, _ := http.NewRequest("GET", "/servers", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusInternalServerError, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Contains(resp.Error, "Failed to retrieve servers")
		mockServerManager.AssertExpectations(t)
	})

	t.Run("GetServer - success", func(t *testing.T) {
		serverID := uuid.NewString()
		expectedServer := &services.ServerInfo{ID: serverID, Name: "Radarr1", Type: "radarr", URL: "http://radarr.com"}
		mockServerManager.On("GetServer", mock.Anything, serverID).Return(expectedServer, nil).Once()

		req, _ := http.NewRequest("GET", "/servers/"+serverID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualServer services.ServerInfo
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualServer)
		assert.Equal(*expectedServer, actualServer)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("GetServer - not found", func(t *testing.T) {
		serverID := uuid.NewString()
		mockServerManager.On("GetServer", mock.Anything, serverID).Return((*services.ServerInfo)(nil), services.ErrServerNotFound).Once()

		req, _ := http.NewRequest("GET", "/servers/"+serverID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusNotFound, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Server not found", resp.Error)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("CreateServer - success", func(t *testing.T) {
		payload := map[string]string{"name": "NewServer", "url": "http://new.com", "apiKey": "abc", "type": "radarr"}
		body, _ := json.Marshal(payload)
		expectedServer := &services.ServerInfo{ID: uuid.NewString(), Name: "NewServer", Type: "radarr", URL: "http://new.com"}
		mockServerManager.On("AddServer", mock.Anything, "NewServer", "http://new.com", "abc", "radarr").Return(expectedServer, nil).Once()

		req, _ := http.NewRequest("POST", "/servers", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualServer services.ServerInfo
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualServer)
		assert.Equal(*expectedServer, actualServer)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("CreateServer - invalid payload", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/servers", strings.NewReader(`{invalid json`))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusBadRequest, rr.Code)
	})

	t.Run("CreateServer - validation error", func(t *testing.T) {
		payload := map[string]string{"name": "", "url": "http://new.com", "apiKey": "abc", "type": "radarr"}
		body, _ := json.Marshal(payload)
		mockServerManager.On("AddServer", mock.Anything, "", "http://new.com", "abc", "radarr").Return((*services.ServerInfo)(nil), services.ErrServerValidation).Once()

		req, _ := http.NewRequest("POST", "/servers", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusBadRequest, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(services.ErrServerValidation.Error(), resp.Error)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("UpdateServer - success", func(t *testing.T) {
		serverID := uuid.NewString()
		nameUpdate := "UpdatedName"
		payload := services.ServerUpdate{Name: &nameUpdate}
		body, _ := json.Marshal(payload)
		mockServerManager.On("UpdateServer", mock.Anything, serverID, payload).Return(nil).Once()

		req, _ := http.NewRequest("PUT", "/servers/"+serverID, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Server updated successfully", resp.Message)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("UpdateServer - not found", func(t *testing.T) {
		serverID := uuid.NewString()
		nameUpdate := "UpdatedName"
		payload := services.ServerUpdate{Name: &nameUpdate}
		body, _ := json.Marshal(payload)
		mockServerManager.On("UpdateServer", mock.Anything, serverID, payload).Return(services.ErrServerNotFound).Once()

		req, _ := http.NewRequest("PUT", "/servers/"+serverID, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusNotFound, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Server not found", resp.Error)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("DeleteServer - success", func(t *testing.T) {
		serverID := uuid.NewString()
		mockServerManager.On("RemoveServer", serverID).Return(nil).Once()

		req, _ := http.NewRequest("DELETE", "/servers/"+serverID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Server removed successfully", resp.Message)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("DeleteServer - not found", func(t *testing.T) {
		serverID := uuid.NewString()
		mockServerManager.On("RemoveServer", serverID).Return(services.ErrServerNotFound).Once()

		req, _ := http.NewRequest("DELETE", "/servers/"+serverID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusNotFound, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Server not found", resp.Error)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("TestServerConnection - success", func(t *testing.T) {
		serverID := uuid.NewString()
		expectedResult := &services.ConnectionResult{Success: true, AppName: "Radarr", Version: "4.0"}
		mockServerManager.On("TestConnection", mock.Anything, serverID).Return(expectedResult, nil).Once()

		req, _ := http.NewRequest("POST", "/servers/"+serverID+"/test", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualResult services.ConnectionResult
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualResult)
		assert.Equal(*expectedResult, actualResult)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("TestServerConnection - failure", func(t *testing.T) {
		serverID := uuid.NewString()
		expectedResult := &services.ConnectionResult{Success: false, Error: "connection timed out"}
		mockServerManager.On("TestConnection", mock.Anything, serverID).Return(expectedResult, nil).Once()

		req, _ := http.NewRequest("POST", "/servers/"+serverID+"/test", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code) // HTTP status is OK, but result indicates failure
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualResult services.ConnectionResult
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualResult)
		assert.Equal(*expectedResult, actualResult)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("TestNewServerConnection - success", func(t *testing.T) {
		payload := map[string]string{"url": "http://new.com", "apiKey": "abc", "type": "radarr"}
		body, _ := json.Marshal(payload)
		expectedResult := &services.ConnectionResult{Success: true, AppName: "Radarr", Version: "4.0"}
		mockServerManager.On("TestConnection", mock.Anything, mock.AnythingOfType("string")).Return(expectedResult, nil).Once() // Will be called with empty ID

		req, _ := http.NewRequest("POST", "/servers/test", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualResult services.ConnectionResult
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualResult)
		assert.Equal(*expectedResult, actualResult)
		mockServerManager.AssertExpectations(t)
	})
}
