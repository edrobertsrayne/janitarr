package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
)

// MockDBStats is a mock database for stats-related operations.
type MockDBStats struct {
	mock.Mock
}

func (m *MockDBStats) GetSystemStats() database.SystemStats {
	args := m.Called()
	return args.Get(0).(database.SystemStats)
}

func (m *MockDBStats) GetServerStats(serverID string) database.ServerStats {
	args := m.Called(serverID)
	return args.Get(0).(database.ServerStats)
}

func TestStatsHandlers(t *testing.T) {
	assert := assert.New(t)

	mockDB := new(MockDBStats)
	handlers := NewStatsHandlers(nil) // DB will be mocked via GetSystemStats/GetServerStatsFunc

	// Override global functions for testing if needed, though for direct method calls, mocking the DB struct directly is fine
	// originalGetSystemStatsFunc := database.GetSystemStatsFunc
	// defer func() { database.GetSystemStatsFunc = originalGetSystemStatsFunc }()
	// database.GetSystemStatsFunc = func(db *database.DB) database.SystemStats {
	// 	return mockDB.GetSystemStats()
	// }

	router := chi.NewRouter()
	router.Get("/stats/summary", handlers.GetSummaryStats)
	router.Get("/stats/servers/{id}", handlers.GetServerStats)

	t.Run("GetSummaryStats - success", func(t *testing.T) {
		expectedStats := database.SystemStats{
			TotalServers:    5,
			LastCycleTime:   "2023-10-27T10:00:00Z",
			SearchesLast24h: 10,
			ErrorsLast24h:   2,
		}
		mockDB.On("GetSystemStats").Return(expectedStats).Once()

		req, _ := http.NewRequest("GET", "/stats/summary", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualStats database.SystemStats
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualStats)
		assert.Equal(expectedStats, actualStats)
		mockDB.AssertExpectations(t)
	})

	t.Run("GetServerStats - success", func(t *testing.T) {
		serverID := uuid.NewString()
		expectedStats := database.ServerStats{
			ServerName:    "Radarr",
			TotalSearches: 5,
			ErrorCount:    1,
			LastCheckTime: "2023-10-27T10:05:00Z",
		}
		mockDB.On("GetServerStats", serverID).Return(expectedStats).Once()

		req, _ := http.NewRequest("GET", "/stats/servers/"+serverID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualStats database.ServerStats
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualStats)
		assert.Equal(expectedStats, actualStats)
		mockDB.AssertExpectations(t)
	})

	t.Run("GetServerStats - not found", func(t *testing.T) {
		serverID := uuid.NewString()
		mockDB.On("GetServerStats", serverID).Return(database.ServerStats{}).Once() // Empty stats indicates not found

		req, _ := http.NewRequest("GET", "/stats/servers/"+serverID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusNotFound, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Server not found", resp.Error)
		mockDB.AssertExpectations(t)
	})

	t.Run("GetServerStats - missing server ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/stats/servers/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req) // This should hit a 404 from chi router if ID is empty or not matched
		// For chi, if a route parameter is missing in the URL path definition itself, it might not even reach the handler.
		// Testing this case specifically might require a different router setup or just rely on chi's default behavior.
		// For now, testing the direct handler call.

		// Manually call handler to check its internal logic if serverID is empty
		reqNoID, _ := http.NewRequest("GET", "/stats/servers/", nil)
		rrNoID := httptest.NewRecorder()
		handlers.GetServerStats(rrNoID, reqNoID) // Direct call without chi context

		assert.Equal(http.StatusBadRequest, rrNoID.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rrNoID.Body.Bytes(), &resp)
		assert.Equal("Server ID is required", resp.Error)
	})
}
