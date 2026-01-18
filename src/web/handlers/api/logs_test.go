package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
)

// MockDBLogStore is a mock implementation of the database operations for logs.
type MockDBLogStore struct {
	mock.Mock
}

func (m *MockDBLogStore) GetLogs(ctx context.Context, limit, offset int, logTypeFilter, serverNameFilter *string) ([]logger.LogEntry, error) {
	args := m.Called(ctx, limit, offset, logTypeFilter, serverNameFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]logger.LogEntry), args.Error(1)
}

func (m *MockDBLogStore) ClearLogs() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDBLogStore) AddLog(entry logger.LogEntry) error {
	args := m.Called(entry)
	return args.Error(0)
}

func TestLogHandlers(t *testing.T) {
	assert := assert.New(t)

	mockDB := new(MockDBLogStore)
	handlers := NewLogHandlers(nil) // DB will be mocked via GetLogs/ClearLogsFunc

	// Override global functions for testing
	originalGetLogsFunc := database.GetLogsFunc
	defer func() { database.GetLogsFunc = originalGetLogsFunc }()
	originalClearLogsFunc := database.ClearLogsFunc
	defer func() { database.ClearLogsFunc = originalClearLogsFunc }()
	originalAddLogFunc := database.AddLogFunc // Not directly used by handlers, but good practice if it were
	defer func() { database.AddLogFunc = originalAddLogFunc }()

	database.GetLogsFunc = func(db *database.DB, ctx context.Context, limit, offset int, logTypeFilter, serverNameFilter *string) ([]logger.LogEntry, error) {
		return mockDB.GetLogs(ctx, limit, offset, logTypeFilter, serverNameFilter)
	}
	database.ClearLogsFunc = func(db *database.DB) error {
		return mockDB.ClearLogs()
	}

	router := chi.NewRouter()
	router.Get("/logs", handlers.ListLogs)
	router.Delete("/logs", handlers.ClearLogs)
	router.Get("/logs/export", handlers.ExportLogs)

	t.Run("ListLogs - success with data", func(t *testing.T) {
		entries := []logger.LogEntry{
			{ID: uuid.NewString(), Timestamp: time.Now(), Type: logger.LogTypeCycleStart, Message: "Cycle started"},
			{ID: uuid.NewString(), Timestamp: time.Now().Add(-time.Hour), Type: logger.LogTypeSearch, ServerName: "Radarr", Message: "Searched item"},
		}
		mockDB.On("GetLogs", mock.Anything, 20, 0, (*string)(nil), (*string)(nil)).Return(entries, nil).Once()

		req, _ := http.NewRequest("GET", "/logs", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var actualLogs []logger.LogEntry
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		// Need to re-marshal and unmarshal because resp.Data is interface{}
		dataBytes, _ := json.Marshal(resp.Data)
		err := json.Unmarshal(dataBytes, &actualLogs)
		assert.NoError(err)
		assert.Len(actualLogs, 2)
		assert.Equal(entries[0].Message, actualLogs[0].Message)
		mockDB.AssertExpectations(t)
	})

	t.Run("ListLogs - no data", func(t *testing.T) {
		mockDB.On("GetLogs", mock.Anything, 20, 0, (*string)(nil), (*string)(nil)).Return([]logger.LogEntry{}, nil).Once()

		req, _ := http.NewRequest("GET", "/logs", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var actualLogs []logger.LogEntry
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		dataBytes, _ := json.Marshal(resp.Data)
		err := json.Unmarshal(dataBytes, &actualLogs)
		assert.NoError(err)
		assert.Empty(actualLogs)
		mockDB.AssertExpectations(t)
	})

	t.Run("ListLogs - with limit and offset", func(t *testing.T) {
		entries := []logger.LogEntry{
			{ID: uuid.NewString(), Timestamp: time.Now(), Type: logger.LogTypeCycleStart, Message: "Cycle started"},
		}
		mockDB.On("GetLogs", mock.Anything, 1, 10, (*string)(nil), (*string)(nil)).Return(entries, nil).Once()

		req, _ := http.NewRequest("GET", "/logs?limit=1&offset=10", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("ListLogs - with type and server filter", func(t *testing.T) {
		logType := "search"
		serverName := "Radarr"
		entries := []logger.LogEntry{
			{ID: uuid.NewString(), Timestamp: time.Now(), Type: logger.LogTypeSearch, ServerName: "Radarr", Message: "Searched item"},
		}
		mockDB.On("GetLogs", mock.Anything, 20, 0, &logType, &serverName).Return(entries, nil).Once()

		req, _ := http.NewRequest("GET", "/logs?type=search&server=Radarr", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("ListLogs - error from DB", func(t *testing.T) {
		mockDB.On("GetLogs", mock.Anything, 20, 0, (*string)(nil), (*string)(nil)).Return(([]logger.LogEntry)(nil), errors.New("db error")).Once()

		req, _ := http.NewRequest("GET", "/logs", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusInternalServerError, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Contains(resp.Error, "Failed to retrieve logs")
		mockDB.AssertExpectations(t)
	})

	t.Run("ClearLogs - success", func(t *testing.T) {
		mockDB.On("ClearLogs").Return(nil).Once()

		req, _ := http.NewRequest("DELETE", "/logs", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("All logs cleared successfully", resp.Message)
		mockDB.AssertExpectations(t)
	})

	t.Run("ClearLogs - error from DB", func(t *testing.T) {
		mockDB.On("ClearLogs").Return(errors.New("db error")).Once()

		req, _ := http.NewRequest("DELETE", "/logs", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusInternalServerError, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Contains(resp.Error, "Failed to clear logs")
		mockDB.AssertExpectations(t)
	})

	t.Run("ExportLogs - JSON success", func(t *testing.T) {
		entries := []logger.LogEntry{
			{ID: uuid.NewString(), Timestamp: time.Now(), Type: logger.LogTypeCycleStart, Message: "Cycle started"},
		}
		mockDB.On("GetLogs", mock.Anything, 0, 0, (*string)(nil), (*string)(nil)).Return(entries, nil).Once()

		req, _ := http.NewRequest("GET", "/logs/export?format=json", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		assert.Equal("application/json", rr.Header().Get("Content-Type"))
		assert.Contains(rr.Header().Get("Content-Disposition"), "janitarr_logs.json")
		var actualLogs []logger.LogEntry
		err := json.Unmarshal(rr.Body.Bytes(), &actualLogs)
		assert.NoError(err)
		assert.Len(actualLogs, 1)
		mockDB.AssertExpectations(t)
	})

	t.Run("ExportLogs - CSV success", func(t *testing.T) {
		entries := []logger.LogEntry{
			{ID: "1", Timestamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), Type: logger.LogTypeSearch, ServerName: "Radarr", ServerType: "radarr", Category: "movie", Count: 1, Message: "Movie searched", IsManual: true},
		}
		mockDB.On("GetLogs", mock.Anything, 0, 0, (*string)(nil), (*string)(nil)).Return(entries, nil).Once()

		req, _ := http.NewRequest("GET", "/logs/export?format=csv", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		assert.Equal("text/csv", rr.Header().Get("Content-Type"))
		assert.Contains(rr.Header().Get("Content-Disposition"), "janitarr_logs.csv")
		expectedCSV := "ID,Timestamp,Type,ServerName,ServerType,Category,Count,Message,IsManual\n" +
				`1,2023-01-01T10:00:00Z,search,Radarr,radarr,movie,1,"Movie searched",true` + "\n"
		assert.Equal(expectedCSV, rr.Body.String())
		mockDB.AssertExpectations(t)
	})

	t.Run("ExportLogs - invalid format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/logs/export?format=invalid", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusBadRequest, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Contains(resp.Error, "Invalid export format")
	})
}
