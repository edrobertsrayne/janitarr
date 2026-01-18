package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
)

// MockDetector for testing the Automation service
type MockDetector struct {
	mock.Mock
}

func (m *MockDetector) DetectAll(ctx context.Context) (*DetectionResults, error) {
	args := m.Called(ctx)
	return args.Get(0).(*DetectionResults), args.Error(1)
}

// MockSearchTrigger for testing the Automation service
type MockSearchTrigger struct {
	mock.Mock
}

func (m *MockSearchTrigger) TriggerSearches(ctx context.Context, detectionResults *DetectionResults, limits database.SearchLimits, dryRun bool) (*TriggerResults, error) {
	args := m.Called(ctx, detectionResults, limits, dryRun)
	return args.Get(0).(*TriggerResults), args.Error(1)
}

// MockLogger for testing the Automation service
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) LogCycleStart(isManual bool) *logger.LogEntry {
	args := m.Called(isManual)
	return args.Get(0).(*logger.LogEntry)
}

func (m *MockLogger) LogCycleEnd(totalSearches, failures int, isManual bool) *logger.LogEntry {
	args := m.Called(totalSearches, failures, isManual)
	return args.Get(0).(*logger.LogEntry)
}

func (m *MockLogger) LogSearches(serverName, serverType, category string, count int, isManual bool) *logger.LogEntry {
	args := m.Called(serverName, serverType, category, count, isManual)
	return args.Get(0).(*logger.LogEntry)
}

func (m *MockLogger) LogServerError(serverName, serverType, reason string) *logger.LogEntry {
	args := m.Called(serverName, serverType, reason)
	return args.Get(0).(*logger.LogEntry)
}

func (m *MockLogger) LogSearchError(serverName, serverType, category, reason string) *logger.LogEntry {
	args := m.Called(serverName, serverType, category, reason)
	return args.Get(0).(*logger.LogEntry)
}

// MockDB for testing the Automation service
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetAppConfig() (*database.AppConfig, error) {
	args := m.Called()
	return args.Get(0).(*database.AppConfig), args.Error(1)
}

func (m *MockDB) AddLogEntry(entry *database.LogEntry) error {
	args := m.Called(entry)
	return args.Error(0)
}

func (m *MockDB) GetAllServers() ([]database.Server, error) {
	args := m.Called()
	return args.Get(0).([]database.Server), args.Error(1)
}

func (m *MockDB) GetServer(idOrName string) (*database.Server, error) {
	args := m.Called(idOrName)
	return args.Get(0).(*database.Server), args.Error(1)
}

func (m *MockDB) GetServersByType(serverType database.ServerType) ([]database.Server, error) {
	args := m.Called(serverType)
	return args.Get(0).([]database.Server), args.Error(1)
}

func (m *MockDB) AddServer(name, url, apiKey string, serverType database.ServerType) (*database.Server, error) {
	args := m.Called(name, url, apiKey, serverType)
	return args.Get(0).(*database.Server), args.Error(1)
}

func (m *MockDB) UpdateServer(serverID string, name, url, apiKey string, serverType database.ServerType, enabled bool) error {
	args := m.Called(serverID, name, url, apiKey, serverType, enabled)
	return args.Error(0)
}

func (m *MockDB) DeleteServer(serverID string) error {
	args := m.Called(serverID)
	return args.Error(0)
}

func (m *MockDB) SetConfig(key, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockDB) GetLogs(limit, offset int) ([]database.LogEntry, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]database.LogEntry), args.Error(1)
}

func (m *MockDB) PurgeOldLogs(days int) error {
	args := m.Called(days)
	return args.Error(0)
}

// Helper to create default AppConfig
func defaultAppConfig() *database.AppConfig {
	return &database.AppConfig{
		Schedule: database.ScheduleConfig{
			IntervalHours: 6,
			Enabled:       true,
		},
		SearchLimits: database.SearchLimits{
			MissingMoviesLimit:   10,
			MissingEpisodesLimit: 10,
			CutoffMoviesLimit:    5,
			CutoffEpisodesLimit:  5,
		},
	}
}

// TestRunCycle_Success verifies a successful automation cycle.
func TestRunCycle_Success(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	mockDB := new(MockDB)
	mockDetector := new(MockDetector)
	mockSearchTrigger := new(MockSearchTrigger)
	mockLogger := new(MockLogger)

	// Expected configuration
	appConfig := defaultAppConfig()

	// Mock DB calls
	mockDB.On("GetAppConfig").Return(appConfig, nil).Once()

	// Mock Logger calls (Note: AddLogEntry is mocked directly as a function of the DB mock for the logger)
	mockLogger.On("LogCycleStart", true).Return(&logger.LogEntry{Type: logger.LogTypeCycleStart}).Once()
	mockLogger.On("LogSearches", "Server1", "radarr", "missing", 2, true).Return(&logger.LogEntry{Type: logger.LogTypeSearch}).Once()
	mockLogger.On("LogSearches", "Server1", "radarr", "cutoff", 1, true).Return(&logger.LogEntry{Type: logger.LogTypeSearch}).Once()
	mockLogger.On("LogCycleEnd", 3, 0, true).Return(&logger.LogEntry{Type: logger.LogTypeCycleEnd}).Once()
	mockLogger.On("LogServerError", mock.Anything, mock.Anything, mock.Anything).Return(&logger.LogEntry{}).Maybe()
	mockLogger.On("LogSearchError", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&logger.LogEntry{}).Maybe()

	// Mock Detector calls
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   "server1",
				ServerName: "Server1",
				ServerType: "radarr",
				Missing:    []int{101, 102},
				Cutoff:     []int{201},
			},
		},
		TotalMissing: 2,
		TotalCutoff:  1,
		SuccessCount: 1,
		FailureCount: 0,
	}
	mockDetector.On("DetectAll", ctx).Return(detectionResults, nil).Once()

	// Mock SearchTrigger calls
	triggerResults := &TriggerResults{
		Results: []TriggerResult{
			{
				ServerID:   "server1",
				ServerName: "Server1",
				ServerType: "radarr",
				Category:   "missing",
				ItemIDs:    []int{101, 102},
				Success:    true,
			},
			{
				ServerID:   "server1",
				ServerName: "Server1",
				ServerType: "radarr",
				Category:   "cutoff",
				ItemIDs:    []int{201},
				Success:    true,
			},
		},
		MissingTriggered: 2,
		CutoffTriggered:  1,
		SuccessCount:     2,
		FailureCount:     0,
	}
	mockSearchTrigger.On("TriggerSearches", ctx, detectionResults, appConfig.SearchLimits, false).Return(triggerResults, nil).Once()

	automation := NewAutomation(mockDB, mockDetector, mockSearchTrigger, mockLogger)
	result, err := automation.RunCycle(ctx, true, false)

	assert.NoError(err)
	assert.True(result.Success)
	assert.Equal(3, result.TotalSearches)
	assert.Equal(0, result.TotalFailures)
	assert.Len(result.Errors, 0)
	assert.WithinDuration(time.Now(), time.Now().Add(-result.Duration), 1*time.Second) // Check duration is set
	assert.Equal(detectionResults.TotalMissing, result.DetectionResults.TotalMissing)  // Check relevant fields
	assert.Equal(detectionResults.TotalCutoff, result.DetectionResults.TotalCutoff)
	assert.Equal(triggerResults.MissingTriggered, result.SearchResults.MissingTriggered)
	assert.Equal(triggerResults.CutoffTriggered, result.SearchResults.CutoffTriggered)

	mockDB.AssertExpectations(t)
	mockDetector.AssertExpectations(t)
	mockSearchTrigger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestRunCycle_DetectionFailure verifies the behavior when detection fails.
func TestRunCycle_DetectionFailure(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	mockDB := new(MockDB)
	mockDetector := new(MockDetector)
	mockSearchTrigger := new(MockSearchTrigger)
	mockLogger := new(MockLogger)

	appConfig := defaultAppConfig()
	mockDB.On("GetAppConfig").Return(appConfig, nil).Once()

	// Mock Logger calls
	mockLogger.On("LogCycleStart", false).Return(&logger.LogEntry{Type: logger.LogTypeCycleStart}).Once()
	mockLogger.On("LogServerError", "ServerWithErr", "radarr", "detection error: failed to detect").Return(&logger.LogEntry{}).Once() // For internal detection error logging
	mockLogger.On("LogCycleEnd", 0, 1, false).Return(&logger.LogEntry{Type: logger.LogTypeCycleEnd}).Once()                           // 1 failure from detection
	mockLogger.On("LogSearches", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&logger.LogEntry{}).Maybe()
	mockLogger.On("LogSearchError", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&logger.LogEntry{}).Maybe()

	// Mock Detector call to return an error, and also a partial result with a server error
	detectionErr := errors.New("failed to detect")
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   "serverWithErr",
				ServerName: "ServerWithErr",
				ServerType: "radarr",
				Error:      detectionErr.Error(),
			},
		},
		TotalMissing: 0,
		TotalCutoff:  0,
		SuccessCount: 0,
		FailureCount: 1,
	}
	mockDetector.On("DetectAll", ctx).Return(detectionResults, detectionErr).Once()

	// SearchTrigger should still be called, but with the (empty) detection results
	triggerResults := &TriggerResults{Results: []TriggerResult{}} // Empty as no successful detections
	mockSearchTrigger.On("TriggerSearches", ctx, detectionResults, appConfig.SearchLimits, false).Return(triggerResults, nil).Once()

	automation := NewAutomation(mockDB, mockDetector, mockSearchTrigger, mockLogger)
	result, err := automation.RunCycle(ctx, false, false)

	assert.Error(err) // Should return an error because detection failed
	assert.Contains(err.Error(), "automation cycle completed with 2 errors")
	assert.False(result.Success)
	assert.Equal(0, result.TotalSearches)
	assert.Equal(1, result.TotalFailures) // 1 failure from detection
	assert.Len(result.Errors, 2)          // 1 from detection error, 1 from server error
	assert.Contains(result.Errors[0], "detection failed")
	assert.Contains(result.Errors[1], "server ServerWithErr detection failed")
	assert.Equal(detectionResults.TotalFailureCount(), result.DetectionResults.TotalFailureCount())

	mockDB.AssertExpectations(t)
	mockDetector.AssertExpectations(t)
	mockSearchTrigger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestRunCycle_TriggerFailure verifies the behavior when search triggering fails.
func TestRunCycle_TriggerFailure(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	mockDB := new(MockDB)
	mockDetector := new(MockDetector)
	mockSearchTrigger := new(MockSearchTrigger)
	mockLogger := new(MockLogger)

	appConfig := defaultAppConfig()
	mockDB.On("GetAppConfig").Return(appConfig, nil).Once()

	// Mock Logger calls
	mockLogger.On("LogCycleStart", true).Return(&logger.LogEntry{Type: logger.LogTypeCycleStart}).Once()
	mockLogger.On("LogSearchError", "Server1", "radarr", "missing", "failed to trigger").Return(&logger.LogEntry{}).Once()
	mockLogger.On("LogCycleEnd", 0, 1, true).Return(&logger.LogEntry{Type: logger.LogTypeCycleEnd}).Once() // 1 failure from trigger
	mockLogger.On("LogServerError", mock.Anything, mock.Anything, mock.Anything).Return(&logger.LogEntry{}).Maybe()
	mockLogger.On("LogSearches", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&logger.LogEntry{}).Maybe()

	// Mock Detector call (successful detection)
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   "server1",
				ServerName: "Server1",
				ServerType: "radarr",
				Missing:    []int{101, 102},
			},
		},
		TotalMissing: 2,
		TotalCutoff:  0,
		SuccessCount: 1,
		FailureCount: 0,
	}
	mockDetector.On("DetectAll", ctx).Return(detectionResults, nil).Once()

	// Mock SearchTrigger call to return an error, and also a partial result with a server error
	triggerErr := errors.New("failed to trigger")
	triggerResults := &TriggerResults{
		Results: []TriggerResult{
			{
				ServerID:   "server1",
				ServerName: "Server1",
				ServerType: "radarr",
				Category:   "missing",
				ItemIDs:    []int{101, 102},
				Success:    false,
				Error:      triggerErr.Error(),
			},
		},
		MissingTriggered: 0, // No actual successful triggers
		CutoffTriggered:  0,
		SuccessCount:     0,
		FailureCount:     1,
	}
	mockSearchTrigger.On("TriggerSearches", ctx, detectionResults, appConfig.SearchLimits, false).Return(triggerResults, triggerErr).Once()

	automation := NewAutomation(mockDB, mockDetector, mockSearchTrigger, mockLogger)
	result, err := automation.RunCycle(ctx, true, false)

	assert.Error(err) // Should return an error because triggering failed
	assert.Contains(err.Error(), "automation cycle completed with 2 errors")
	assert.False(result.Success)
	assert.Equal(0, result.TotalSearches)
	assert.Equal(1, result.TotalFailures) // 1 failure from trigger
	assert.Len(result.Errors, 2)
	assert.Contains(result.Errors[0], "triggering searches failed")
	assert.Contains(result.Errors[1], "server Server1 search failed for missing (radarr): failed to trigger")
	assert.Equal(triggerResults.TotalFailureCount(), result.SearchResults.TotalFailureCount())

	mockDB.AssertExpectations(t)
	mockDetector.AssertExpectations(t)
	mockSearchTrigger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestRunCycle_DryRun verifies that no API calls or logs are made in dry-run mode.
func TestRunCycle_DryRun(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	mockDB := new(MockDB)
	mockDetector := new(MockDetector)
	mockSearchTrigger := new(MockSearchTrigger)
	mockLogger := new(MockLogger)

	appConfig := defaultAppConfig()
	mockDB.On("GetAppConfig").Return(appConfig, nil).Once()

	// Mock Logger calls - only LogCycleStart and LogCycleEnd should be called, but with dryRun=true
	mockLogger.On("LogCycleStart", true).Return(&logger.LogEntry{Type: logger.LogTypeCycleStart}).Once()
	mockLogger.On("LogCycleEnd", 3, 0, true).Return(&logger.LogEntry{Type: logger.LogTypeCycleEnd}).Once()

	// Mock Detector calls (successful detection)
	detectionResults := &DetectionResults{
		Results: []DetectionResult{
			{
				ServerID:   "server1",
				ServerName: "Server1",
				ServerType: "radarr",
				Missing:    []int{101, 102},
				Cutoff:     []int{201},
			},
		},
		TotalMissing: 2,
		TotalCutoff:  1,
		SuccessCount: 1,
		FailureCount: 0,
	}
	mockDetector.On("DetectAll", ctx).Return(detectionResults, nil).Once()

	// Mock SearchTrigger calls - dryRun should be true
	triggerResults := &TriggerResults{
		Results: []TriggerResult{
			{
				ServerID:   "server1",
				ServerName: "Server1",
				ServerType: "radarr",
				Category:   "missing",
				ItemIDs:    []int{101, 102},
				Success:    true,
			},
			{
				ServerID:   "server1",
				ServerName: "Server1",
				ServerType: "radarr",
				Category:   "cutoff",
				ItemIDs:    []int{201},
				Success:    true,
			},
		},
		MissingTriggered: 2,
		CutoffTriggered:  1,
		SuccessCount:     2,
		FailureCount:     0,
	}
	mockSearchTrigger.On("TriggerSearches", ctx, detectionResults, appConfig.SearchLimits, true).Return(triggerResults, nil).Once()

	automation := NewAutomation(mockDB, mockDetector, mockSearchTrigger, mockLogger)
	result, err := automation.RunCycle(ctx, true, true) // Dry-run is true

	assert.NoError(err)
	assert.True(result.Success)
	assert.Equal(3, result.TotalSearches)
	assert.Equal(0, result.TotalFailures)
	assert.Len(result.Errors, 0)

	mockDB.AssertExpectations(t)
	mockDetector.AssertExpectations(t)
	mockSearchTrigger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestRunCycle_ManualScheduledLogging verifies isManual flag is passed correctly to logger.
func TestRunCycle_ManualScheduledLogging(t *testing.T) {
	tests := []struct {
		name     string
		isManual bool
	}{
		{"Manual Run", true},
		{"Scheduled Run", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()

			mockDB := new(MockDB)
			mockDetector := new(MockDetector)
			mockSearchTrigger := new(MockSearchTrigger)
			mockLogger := new(MockLogger)

			appConfig := defaultAppConfig()
			mockDB.On("GetAppConfig").Return(appConfig, nil).Once()

			// Mock Logger calls, checking isManual flag
			mockLogger.On("LogCycleStart", tt.isManual).Return(&logger.LogEntry{Type: logger.LogTypeCycleStart}).Once()
			mockLogger.On("LogSearches", "Server1", "radarr", "missing", 2, tt.isManual).Return(&logger.LogEntry{Type: logger.LogTypeSearch}).Once()
			mockLogger.On("LogCycleEnd", 2, 0, tt.isManual).Return(&logger.LogEntry{Type: logger.LogTypeCycleEnd}).Once()
			mockLogger.On("LogServerError", mock.Anything, mock.Anything, mock.Anything).Return(&logger.LogEntry{}).Maybe()
			mockLogger.On("LogSearchError", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&logger.LogEntry{}).Maybe()

			// Mock Detector calls (successful detection)
			detectionResults := &DetectionResults{
				Results: []DetectionResult{
					{
						ServerID:   "server1",
						ServerName: "Server1",
						ServerType: "radarr",
						Missing:    []int{101, 102},
					},
				},
				TotalMissing: 2,
				SuccessCount: 1,
			}
			mockDetector.On("DetectAll", ctx).Return(detectionResults, nil).Once()

			// Mock SearchTrigger calls
			triggerResults := &TriggerResults{
				Results: []TriggerResult{
					{
						ServerID:   "server1",
						ServerName: "Server1",
						ServerType: "radarr",
						Category:   "missing",
						ItemIDs:    []int{101, 102},
						Success:    true,
					},
				},
				MissingTriggered: 2,
				SuccessCount:     1,
			}
			mockSearchTrigger.On("TriggerSearches", ctx, detectionResults, appConfig.SearchLimits, false).Return(triggerResults, nil).Once()

			automation := NewAutomation(mockDB, mockDetector, mockSearchTrigger, mockLogger)
			_, err := automation.RunCycle(ctx, tt.isManual, false) // isManual based on test case

			assert.NoError(err)

			mockDB.AssertExpectations(t)
			mockDetector.AssertExpectations(t)
			mockSearchTrigger.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

// TestRunCycle_EmptyResults verifies behavior with no detection results.
func TestRunCycle_EmptyResults(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	mockDB := new(MockDB)
	mockDetector := new(MockDetector)
	mockSearchTrigger := new(MockSearchTrigger)
	mockLogger := new(MockLogger)

	appConfig := defaultAppConfig()
	mockDB.On("GetAppConfig").Return(appConfig, nil).Once()

	// Mock Logger calls - only cycle start/end
	mockLogger.On("LogCycleStart", false).Return(&logger.LogEntry{Type: logger.LogTypeCycleStart}).Once()
	mockLogger.On("LogCycleEnd", 0, 0, false).Return(&logger.LogEntry{Type: logger.LogTypeCycleEnd}).Once()

	// Mock Detector call - returns empty results
	detectionResults := &DetectionResults{Results: []DetectionResult{}}
	mockDetector.On("DetectAll", ctx).Return(detectionResults, nil).Once()

	// Mock SearchTrigger call - expects empty detection results
	triggerResults := &TriggerResults{Results: []TriggerResult{}}
	mockSearchTrigger.On("TriggerSearches", ctx, detectionResults, appConfig.SearchLimits, false).Return(triggerResults, nil).Once()

	automation := NewAutomation(mockDB, mockDetector, mockSearchTrigger, mockLogger)
	result, err := automation.RunCycle(ctx, false, false)

	assert.NoError(err)
	assert.True(result.Success)
	assert.Equal(0, result.TotalSearches)
	assert.Equal(0, result.TotalFailures)
	assert.Len(result.Errors, 0)
	assert.Equal(0, len(result.DetectionResults.Results))
	assert.Equal(0, len(result.SearchResults.Results))

	mockDB.AssertExpectations(t)
	mockDetector.AssertExpectations(t)
	mockSearchTrigger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func (dr *DetectionResults) TotalFailureCount() int {
	return dr.FailureCount
}

func (tr *TriggerResults) TotalFailureCount() int {
	return tr.FailureCount
}
