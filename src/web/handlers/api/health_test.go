package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/services"
)

// MockDBHealth is a mock for database operations for health checks
type MockDBHealth struct {
	mock.Mock
}

func (m *MockDBHealth) Ping() error {
	args := m.Called()
	return args.Error(0)
}

// MockSchedulerHealth is a mock for scheduler operations for health checks
type MockSchedulerHealth struct {
	mock.Mock
}

func (m *MockSchedulerHealth) GetStatus() services.SchedulerStatus {
	args := m.Called()
	return args.Get(0).(services.SchedulerStatus)
}

func TestHealthHandlers(t *testing.T) {
	assert := assert.New(t)

	mockDB := new(MockDBHealth)
	mockScheduler := new(MockSchedulerHealth)
	handlers := NewHealthHandlers(nil, nil) // DB and Scheduler are mocked through their methods

	// Override global functions for testing if needed, or directly mock methods on the struct passed to NewHealthHandlers
	// For these tests, we'll mock the Ping method of the DB struct and GetStatus of Scheduler
	originalDBPingFunc := database.PingFunc
	defer func() { database.PingFunc = originalDBPingFunc }()
	originalSchedulerGetStatusFunc := services.GetSchedulerStatusFunc
	defer func() { services.GetSchedulerStatusFunc = originalSchedulerGetStatusFunc }()

	database.PingFunc = func(db *database.DB) error {
		return mockDB.Ping()
	}
	services.GetSchedulerStatusFunc = func(scheduler *services.Scheduler) services.SchedulerStatus {
		return mockScheduler.GetStatus()
	}

	router := chi.NewRouter()
	router.Get("/health", handlers.GetHealth)

	t.Run("GetHealth - All OK", func(t *testing.T) {
		mockDB.On("Ping").Return(nil).Once()
		mockScheduler.On("GetStatus").Return(services.SchedulerStatus{IsRunning: true}).Once()

		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp HealthResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("ok", resp.Status)
		assert.Equal("ok", resp.Database["status"])
		assert.Equal(true, resp.Services["scheduler"].(map[string]interface{})["running"])
		mockDB.AssertExpectations(t)
		mockScheduler.AssertExpectations(t)
	})

	t.Run("GetHealth - Database Error", func(t *testing.T) {
		mockDB.On("Ping").Return(sql.ErrConnDone).Once()
		mockScheduler.On("GetStatus").Return(services.SchedulerStatus{IsRunning: true}).Once()

		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusServiceUnavailable, rr.Code)
		var resp HealthResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("error", resp.Status)
		assert.Equal("error", resp.Database["status"])
		assert.Contains(resp.Database["message"], "Database connection failed")
		mockDB.AssertExpectations(t)
		mockScheduler.AssertExpectations(t)
	})

	t.Run("GetHealth - Scheduler Not Running", func(t *testing.T) {
		mockDB.On("Ping").Return(nil).Once()
		mockScheduler.On("GetStatus").Return(services.SchedulerStatus{IsRunning: false}).Once()

		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusServiceUnavailable, rr.Code)
		var resp HealthResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("degraded", resp.Status)
		assert.Equal("ok", resp.Database["status"])
		assert.Equal(false, resp.Services["scheduler"].(map[string]interface{})["running"])
		assert.Contains(resp.Services["scheduler"].(map[string]interface{})["message"], "Scheduler is not running")
		mockDB.AssertExpectations(t)
		mockScheduler.AssertExpectations(t)
	})

	t.Run("GetHealth - Both Degraded", func(t *testing.T) {
		mockDB.On("Ping").Return(sql.ErrConnDone).Once()
		mockScheduler.On("GetStatus").Return(services.SchedulerStatus{IsRunning: false}).Once()

		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusServiceUnavailable, rr.Code)
		var resp HealthResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("error", resp.Status) // Database error takes precedence for overall status
		assert.Equal("error", resp.Database["status"])
		assert.Equal(false, resp.Services["scheduler"].(map[string]interface{})["running"])
		mockDB.AssertExpectations(t)
		mockScheduler.AssertExpectations(t)
	})
}
