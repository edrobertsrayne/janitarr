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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/services"
)

// MockAutomationService is a mock implementation of the Automation service.
type MockAutomationService struct {
	mock.Mock
}

func (m *MockAutomationService) RunCycle(ctx context.Context, isManual, dryRun bool) (*services.CycleResult, error) {
	args := m.Called(ctx, isManual, dryRun)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CycleResult), args.Error(1)
}

// MockSchedulerService is a mock implementation of the Scheduler service.
type MockSchedulerService struct {
	mock.Mock
}

func (m *MockSchedulerService) IsCycleActive() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSchedulerService) GetStatus() services.SchedulerStatus {
	args := m.Called()
	return args.Get(0).(services.SchedulerStatus)
}

// MockLoggerService is a mock implementation of the Logger service.
type MockLoggerService struct {
	mock.Mock
}

func (m *MockLoggerService) LogServerError(serverName, serverType, reason string) *logger.LogEntry {
	args := m.Called(serverName, serverType, reason)
	if args.Get(0) == nil {
		return nil // Or return a mock logger.LogEntry if needed
	}
	return args.Get(0).(*logger.LogEntry)
}

// MockDBAutomation provides GetAppConfig
type MockDBAutomation struct {
	mock.Mock
}

func (m *MockDBAutomation) GetAppConfig() database.AppConfig {
	args := m.Called()
	return args.Get(0).(database.AppConfig)
}

func TestAutomationHandlers(t *testing.T) {
	assert := assert.New(t)

	mockAutomation := new(MockAutomationService)
	mockScheduler := new(MockSchedulerService)
	mockLogger := new(MockLoggerService)
	mockDB := new(MockDBAutomation) // This mock is for database.GetAppConfigFunc in services.Automation

	handlers := NewAutomationHandlers(mockDB, mockAutomation, mockScheduler, mockLogger)

	router := chi.NewRouter()
	router.Post("/automation/trigger", handlers.TriggerAutomationCycle)
	router.Get("/automation/status", handlers.GetSchedulerStatus)

	// Override global functions for services.NewAutomation to use our mocks
	originalNewAutomation := services.NewAutomation
	defer func() { services.NewAutomation = originalNewAutomation }()
	originalNewDetector := services.NewDetector
	defer func() { services.NewDetector = originalNewDetector }()
	originalNewSearchTrigger := services.NewSearchTrigger
	defer func() { services.NewSearchTrigger = originalNewSearchTrigger }()
	originalNewLogger := logger.NewLogger
	defer func() { logger.NewLogger = originalNewLogger }()

	// For AutomationHandlers to construct the internal services.Automation instance:
	// services.NewAutomation = func(db services.AutomationDB, detector services.AutomationDetector, trigger services.AutomationSearchTrigger, logger services.AutomationLogger) *services.Automation {
	// 	return mockAutomation // This would create a cycle in mocking if used directly
	// }

	t.Run("TriggerAutomationCycle - success (dry run)", func(t *testing.T) {
		mockScheduler.On("IsCycleActive").Return(false).Once()
		// Mock the logger.LogServerError call within the goroutine, needs context of when it's called
		// As the goroutine is separate, we primarily test the HTTP response for now.
		// A more robust test would use channels or wait groups to ensure goroutine completion.

		payload := map[string]bool{"dryRun": true}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/automation/trigger", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusAccepted, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Dry-run automation cycle started in background.", resp.Message)
		mockScheduler.AssertExpectations(t)
		mockAutomation.AssertExpectations(t) // Should not be called directly by handler, but by goroutine
	})

	t.Run("TriggerAutomationCycle - success (live run)", func(t *testing.T) {
		mockScheduler.On("IsCycleActive").Return(false).Once()

		payload := map[string]bool{"dryRun": false}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/automation/trigger", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusAccepted, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Automation cycle started in background.", resp.Message)
		mockScheduler.AssertExpectations(t)
		mockAutomation.AssertExpectations(t)
	})

	t.Run("TriggerAutomationCycle - cycle already active", func(t *testing.T) {
		mockScheduler.On("IsCycleActive").Return(true).Once()

		payload := map[string]bool{"dryRun": false}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/automation/trigger", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusConflict, rr.Code)
		var resp ErrorResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal("Automation cycle already active", resp.Error)
		mockScheduler.AssertExpectations(t)
	})

	t.Run("GetSchedulerStatus - success", func(t *testing.T) {
		now := time.Now()
		nextRun := now.Add(4 * time.Hour)
		expectedStatus := services.SchedulerStatus{
			IsRunning:     true,
			IsCycleActive: false,
			NextRun:       &nextRun,
			LastRun:       nil,
			IntervalHours: 4,
		}
		mockScheduler.On("GetStatus").Return(expectedStatus).Once()

		req, _ := http.NewRequest("GET", "/automation/status", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualStatus services.SchedulerStatus
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualStatus)
		assert.Equal(expectedStatus.IsRunning, actualStatus.IsRunning)
		assert.Equal(expectedStatus.IntervalHours, actualStatus.IntervalHours)
		mockScheduler.AssertExpectations(t)
	})

	t.Run("GetSchedulerStatus - not running", func(t *testing.T) {
		expectedStatus := services.SchedulerStatus{
			IsRunning:     false,
			IsCycleActive: false,
			NextRun:       nil,
			LastRun:       nil,
			IntervalHours: 6,
		}
		mockScheduler.On("GetStatus").Return(expectedStatus).Once()

		req, _ := http.NewRequest("GET", "/automation/status", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		var resp SuccessResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		var actualStatus services.SchedulerStatus
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &actualStatus)
		assert.Equal(expectedStatus.IsRunning, actualStatus.IsRunning)
		mockScheduler.AssertExpectations(t)
	})
}
