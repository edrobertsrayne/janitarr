package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/services"
)

// MockAutomation for testing CLI commands
type MockAutomation struct {
	mock.Mock
}

func (m *MockAutomation) RunCycle(ctx context.Context, isManual, dryRun bool) (*services.CycleResult, error) {
	args := m.Called(ctx, isManual, dryRun)
	return args.Get(0).(*services.CycleResult), args.Error(1)
}

// MockDetector for testing CLI commands related to scanning
type MockDetector struct {
	mock.Mock
}

func (m *MockDetector) DetectAll(ctx context.Context) (*services.DetectionResults, error) {
	args := m.Called(ctx)
	return args.Get(0).(*services.DetectionResults), args.Error(1)
}

// MockServerManager for testing CLI commands related to servers
type MockServerManager struct {
	mock.Mock
}

func (m *MockServerManager) AddServer(ctx context.Context, name, url, apiKey, serverType string) (*services.ServerInfo, error) {
	args := m.Called(ctx, name, url, apiKey, serverType)
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

func (m *MockServerManager) TestConnection(ctx context.Context, id string) (*services.ConnectionResult, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*services.ConnectionResult), args.Error(1)
}

func (m *MockServerManager) ListServers() ([]services.ServerInfo, error) {
	args := m.Called()
	return args.Get(0).([]services.ServerInfo), args.Error(1)
}

func (m *MockServerManager) GetServer(ctx context.Context, idOrName string) (*services.ServerInfo, error) {
	args := m.Called(ctx, idOrName)
	return args.Get(0).(*services.ServerInfo), args.Error(1)
}

// MockScheduler for testing CLI commands related to scheduler status
type MockScheduler struct {
	mock.Mock
}

func (m *MockScheduler) GetStatus() services.SchedulerStatus {
	args := m.Called()
	return args.Get(0).(services.SchedulerStatus)
}

// MockLogger for testing Automation related CLI commands
type MockLoggerCLI struct {
	mock.Mock
}

func (m *MockLoggerCLI) LogCycleStart(isManual bool) *logger.LogEntry {
	args := m.Called(isManual)
	return args.Get(0).(*logger.LogEntry)
}

func (m *MockLoggerCLI) LogCycleEnd(totalSearches, failures int, isManual bool) *logger.LogEntry {
	args := m.Called(totalSearches, failures, isManual)
	return args.Get(0).(*logger.LogEntry)
}

func (m *MockLoggerCLI) LogSearches(serverName, serverType, category string, count int, isManual bool) *logger.LogEntry {
	args := m.Called(serverName, serverType, category, count, isManual)
	return args.Get(0).(*logger.LogEntry)
}

func (m *MockLoggerCLI) LogServerError(serverName, serverType, reason string) *logger.LogEntry {
	args := m.Called(serverName, serverType, reason)
	return args.Get(0).(*logger.LogEntry)
}

func (m *MockLoggerCLI) LogSearchError(serverName, serverType, category, reason string) *logger.LogEntry {
	args := m.Called(serverName, serverType, category, reason)
	return args.Get(0).(*logger.LogEntry)
}

func createTestDBAutomation(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.New(":memory:", t.TempDir()+"/.janitarr.key")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestRunCommand(t *testing.T) {
	assert := assert.New(t)

	// Override NewAutomation for testing
	originalNewAutomation := services.NewAutomation
	defer func() { services.NewAutomation = originalNewAutomation }()

	mockAutomation := new(MockAutomation)
	services.NewAutomation = func(db services.AutomationDB, detector services.AutomationDetector, trigger services.AutomationSearchTrigger, logger services.AutomationLogger) *services.Automation {
		return &services.Automation{
			DB:       db,
			Detector: detector,
			Trigger:  trigger,
			Logger:   logger,
		}
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(runCmd)

	// Temporarily override database.New to return a mock DB
	originalNewDB := database.New
	defer func() { database.New = originalNewDB }()
	database.New = func(dbPath, keyPath string) (*database.DB, error) {
		return createTestDBAutomation(t), nil
	}

	t.Run("run command - success", func(t *testing.T) {
		cycleResult := &services.CycleResult{
			Success:   true,
			TotalSearches: 5,
			TotalFailures: 0,
			Errors:    []string{},
			Duration:  1 * time.Second,
			DetectionResults: services.DetectionResults{
				Results: []services.DetectionResult{
					{ServerName: "Radarr", Missing: []int{1, 2}},
					{ServerName: "Sonarr", Cutoff: []int{3, 4, 5}},
				},
				TotalMissing: 2,
				TotalCutoff:  3,
			},
			SearchResults: services.TriggerResults{
				MissingTriggered: 2,
				CutoffTriggered:  3,
			},
		}
		mockAutomation.On("RunCycle", mock.Anything, true, false).Return(cycleResult, nil).Once()

		output, err := executeCommand(rootCmd, "run")
		assert.NoError(err)
		assert.Contains(output, "Automation Cycle Finished in 1.0s")
		assert.Contains(output, "Overall Status: SUCCESS")
		assert.Contains(output, "Total Searches Triggered: 5")
		mockAutomation.AssertExpectations(t)
	})

	t.Run("run command - dry run", func(t *testing.T) {
		cycleResult := &services.CycleResult{
			Success:   true,
			TotalSearches: 5,
			TotalFailures: 0,
			Errors:    []string{},
			Duration:  1 * time.Second,
			DetectionResults: services.DetectionResults{
				Results: []services.DetectionResult{
					{ServerName: "Radarr", Missing: []int{1, 2}},
					{ServerName: "Sonarr", Cutoff: []int{3, 4, 5}},
				},
				TotalMissing: 2,
				TotalCutoff:  3,
			},
			SearchResults: services.TriggerResults{
				MissingTriggered: 2,
				CutoffTriggered:  3,
			},
		}
		mockAutomation.On("RunCycle", mock.Anything, true, true).Return(cycleResult, nil).Once()

		output, err := executeCommand(rootCmd, "run", "--dry-run")
		assert.NoError(err)
		assert.Contains(output, "Automation Cycle Finished in 1.0s")
		assert.Contains(output, "Overall Status: SUCCESS")
		assert.Contains(output, "Total Searches Triggered: 5")
		mockAutomation.AssertExpectations(t)
	})

	t.Run("run command - failure", func(t *testing.T) {
		cycleResult := &services.CycleResult{
			Success:   false,
			TotalSearches: 2,
			TotalFailures: 1,
			Errors:    []string{"Simulated error"},
			Duration:  1 * time.Second,
			DetectionResults: services.DetectionResults{
				Results: []services.DetectionResult{
					{ServerName: "Radarr", Missing: []int{1, 2}},
				},
				TotalMissing: 2,
				TotalCutoff:  0,
			},
			SearchResults: services.TriggerResults{
				MissingTriggered: 2,
				CutoffTriggered:  0,
				FailureCount:     1,
				Results: []services.TriggerResult{
					{ServerName: "Radarr", Category: "missing", Error: "failed to trigger"},
				},
			},
		}
		mockError := errors.New("automation cycle completed with 1 errors")
		mockAutomation.On("RunCycle", mock.Anything, true, false).Return(cycleResult, mockError).Once()

		output, err := executeCommand(rootCmd, "run")
		assert.Error(err)
		assert.Contains(output, errorMsg("Automation cycle completed with errors: automation cycle completed with 1 errors"))
		assert.Contains(output, errorMsg("Simulated error"))
		assert.Contains(output, "Overall Status: FAILED with 1 errors")
		mockAutomation.AssertExpectations(t)
	})

	t.Run("run command - json output", func(t *testing.T) {
		cycleResult := &services.CycleResult{
			Success:   true,
			TotalSearches: 5,
			TotalFailures: 0,
			Errors:    []string{},
			Duration:  1 * time.Second,
			DetectionResults: services.DetectionResults{
				Results: []services.DetectionResult{
					{ServerName: "Radarr", Missing: []int{1, 2}},
					{ServerName: "Sonarr", Cutoff: []int{3, 4, 5}},
				},
				TotalMissing: 2,
				TotalCutoff:  3,
			},
			SearchResults: services.TriggerResults{
				MissingTriggered: 2,
				CutoffTriggered:  3,
			},
		}
		mockAutomation.On("RunCycle", mock.Anything, true, false).Return(cycleResult, nil).Once()

		output, err := executeCommand(rootCmd, "run", "--json")
		assert.NoError(err)

		var actualResult services.CycleResult
		err = json.Unmarshal([]byte(output), &actualResult)
		assert.NoError(err)
		assert.True(actualResult.Success)
		assert.Equal(cycleResult.TotalSearches, actualResult.TotalSearches)
		assert.Equal(cycleResult.TotalFailures, actualResult.TotalFailures)
		mockAutomation.AssertExpectations(t)
	})
}

func TestScanCommand(t *testing.T) {
	assert := assert.New(t)

	// Override services.NewDetector for testing
	originalNewDetector := services.NewDetector
	defer func() { services.NewDetector = originalNewDetector }()

	mockDetector := new(MockDetector)
	services.NewDetector = func(db *database.DB) services.DetectorInterface {
		return mockDetector
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(scanCmd)

	// Temporarily override database.New to return a mock DB
	originalNewDB := database.New
	defer func() { database.New = originalNewDB }()
	database.New = func(dbPath, keyPath string) (*database.DB, error) {
		return createTestDBAutomation(t), nil
	}

	t.Run("scan command - success", func(t *testing.T) {
		detectionResults := &services.DetectionResults{
			Results: []services.DetectionResult{
				{ServerName: "Radarr", ServerType: "radarr", Missing: []int{101, 102}, Cutoff: []int{201}},
				{ServerName: "Sonarr", ServerType: "sonarr", Missing: []int{301}, Cutoff: []int{401, 402}},
			},
			TotalMissing: 3,
			TotalCutoff:  3,
			SuccessCount: 2,
			FailureCount: 0,
		}
		mockDetector.On("DetectAll", mock.Anything).Return(detectionResults, nil).Once()

		output, err := executeCommand(rootCmd, "scan")
		assert.NoError(err)
		assert.Contains(output, "Scan Results:")
		assert.Contains(output, "Successful Scans: 2")
		assert.Contains(output, "Failed Scans: 0")
		assert.Contains(output, "Total Missing Items: 3")
		assert.Contains(output, "Total Cutoff Unmet Items: 3")
		assert.Contains(output, success("Server Radarr (radarr) Scan Successful:"))
		assert.Contains(output, success("Server Sonarr (sonarr) Scan Successful:"))
		mockDetector.AssertExpectations(t)
	})

	t.Run("scan command - partial failure", func(t *testing.T) {
		detectionResults := &services.DetectionResults{
			Results: []services.DetectionResult{
				{ServerName: "Radarr", ServerType: "radarr", Missing: []int{101}, Cutoff: []int{}, Error: "API error"},
				{ServerName: "Sonarr", ServerType: "sonarr", Missing: []int{301}, Cutoff: []int{401, 402}},
			},
			TotalMissing: 2,
			TotalCutoff:  2,
			SuccessCount: 1,
			FailureCount: 1,
		}
		mockDetector.On("DetectAll", mock.Anything).Return(detectionResults, nil).Once()

		output, err := executeCommand(rootCmd, "scan")
		assert.NoError(err) // Command itself should not error if detection results contain errors
		assert.Contains(output, "Scan Results:")
		assert.Contains(output, "Successful Scans: 1")
		assert.Contains(output, "Failed Scans: 1")
		assert.Contains(output, errorMsg("Server Radarr (radarr) Scan Failed: API error"))
		assert.Contains(output, success("Server Sonarr (sonarr) Scan Successful:"))
		mockDetector.AssertExpectations(t)
	})

	t.Run("scan command - all failure", func(t *testing.T) {
		detectionResults := &services.DetectionResults{
			Results: []services.DetectionResult{
				{ServerName: "Radarr", ServerType: "radarr", Missing: []int{}, Cutoff: []int{}, Error: "Auth failed"},
				{ServerName: "Sonarr", ServerType: "sonarr", Missing: []int{}, Cutoff: []int{}, Error: "Network issue"},
			},
			TotalMissing: 0,
			TotalCutoff:  0,
			SuccessCount: 0,
			FailureCount: 2,
		}
		mockDetector.On("DetectAll", mock.Anything).Return(detectionResults, nil).Once()

		output, err := executeCommand(rootCmd, "scan")
		assert.NoError(err) // Command itself should not error if detection results contain errors
		assert.Contains(output, "Scan Results:")
		assert.Contains(output, "Successful Scans: 0")
		assert.Contains(output, "Failed Scans: 2")
		assert.Contains(output, errorMsg("Server Radarr (radarr) Scan Failed: Auth failed"))
		assert.Contains(output, errorMsg("Server Sonarr (sonarr) Scan Failed: Network issue"))
		mockDetector.AssertExpectations(t)
	})

	t.Run("scan command - detector returns error", func(t *testing.T) {
		mockDetector.On("DetectAll", mock.Anything).Return((*services.DetectionResults)(nil), errors.New("database unavailable")).Once()

		output, err := executeCommand(rootCmd, "scan")
		assert.Error(err)
		assert.Contains(output, errorMsg("Error during scan: database unavailable"))
		mockDetector.AssertExpectations(t)
	})

	t.Run("scan command - json output", func(t *testing.T) {
		detectionResults := &services.DetectionResults{
			Results: []services.DetectionResult{
				{ServerName: "Radarr", ServerType: "radarr", Missing: []int{101}, Cutoff: []int{}},
				{ServerName: "Sonarr", ServerType: "sonarr", Missing: []int{}, Cutoff: []int{401}},
			},
			TotalMissing: 1,
			TotalCutoff:  1,
			SuccessCount: 2,
			FailureCount: 0,
		}
		mockDetector.On("DetectAll", mock.Anything).Return(detectionResults, nil).Once()

		output, err := executeCommand(rootCmd, "scan", "--json")
		assert.NoError(err)

		var actualResults services.DetectionResults
		err = json.Unmarshal([]byte(output), &actualResults)
		assert.NoError(err)
		assert.Equal(detectionResults.TotalMissing, actualResults.TotalMissing)
		assert.Equal(detectionResults.TotalCutoff, actualResults.TotalCutoff)
		assert.Equal(len(detectionResults.Results), len(actualResults.Results))
		mockDetector.AssertExpectations(t)
	})
}

func TestStatusCommand(t *testing.T) {
	assert := assert.New(t)

	// Override services.NewServerManager and services.GetSchedulerStatusFunc
	originalNewServerManager := services.NewServerManager
	defer func() { services.NewServerManager = originalNewServerManager }()
	originalGetSchedulerStatusFunc := services.GetSchedulerStatusFunc
	defer func() { services.GetSchedulerStatusFunc = originalGetSchedulerStatusFunc }()

	mockServerManager := new(MockServerManager)
	services.NewServerManager = func(db *database.DB) services.ServerManagerInterface {
		return mockServerManager
	}

	mockScheduler := new(MockScheduler)
	services.GetSchedulerStatusFunc = func(db *database.DB) services.SchedulerStatus {
		return mockScheduler.GetStatus()
	}

	rootCmd := NewRootCmd()
	rootCmd.AddCommand(statusCmd)

	// Temporarily override database.New to return a mock DB
	originalNewDB := database.New
	defer func() { database.New = originalNewDB }()
	database.New = func(dbPath, keyPath string) (*database.DB, error) {
		return createTestDBAutomation(t), nil
	}

	t.Run("status command - scheduler running, servers configured", func(t *testing.T) {
		now := time.Now()
		nextRun := now.Add(2 * time.Hour)
		lastRun := now.Add(-4 * time.Hour)

		schedulerStatus := services.SchedulerStatus{
			IsRunning:     true,
			IsCycleActive: false,
			NextRun:       &nextRun,
			LastRun:       &lastRun,
			IntervalHours: 6,
		}
		mockScheduler.On("GetStatus").Return(schedulerStatus).Once()

		servers := []services.ServerInfo{
			{ID: uuid.NewString(), Name: "MyRadarr", Type: "radarr"},
			{ID: uuid.NewString(), Name: "MySonarr", Type: "sonarr"},
			{ID: uuid.NewString(), Name: "AnotherRadarr", Type: "radarr"},
		}
		mockServerManager.On("ListServers").Return(servers, nil).Once()

		output, err := executeCommand(rootCmd, "status")
		assert.NoError(err)
		assert.Contains(output, "Janitarr Status:")
		assert.Contains(output, "Scheduler Status:")
		assert.Contains(output, "Running: "+success("Yes"))
		assert.Contains(output, "Cycle Active: "+warning("No"))
		assert.Contains(output, "Interval: 6 hours")
		assert.Contains(output, "Server Overview:")
		assert.Contains(output, "Total Configured: 3")
		assert.Contains(output, "Radarr Servers: 2")
		assert.Contains(output, "Sonarr Servers: 1")
		mockScheduler.AssertExpectations(t)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("status command - scheduler stopped, no servers", func(t *testing.T) {
		schedulerStatus := services.SchedulerStatus{
			IsRunning:     false,
			IsCycleActive: false,
			NextRun:       nil,
			LastRun:       nil,
			IntervalHours: 6,
		}
		mockScheduler.On("GetStatus").Return(schedulerStatus).Once()
		mockServerManager.On("ListServers").Return([]services.ServerInfo{}, nil).Once()

		output, err := executeCommand(rootCmd, "status")
		assert.NoError(err)
		assert.Contains(output, "Running: "+warning("No"))
		assert.Contains(output, "Next Run: N/A")
		assert.Contains(output, "Last Run: N/A")
		assert.Contains(output, "Total Configured: 0")
		assert.Contains(output, "Radarr Servers: 0")
		assert.Contains(output, "Sonarr Servers: 0")
		mockScheduler.AssertExpectations(t)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("status command - json output", func(t *testing.T) {
		now := time.Now()
		nextRun := now.Add(2 * time.Hour)
		lastRun := now.Add(-4 * time.Hour)

		schedulerStatus := services.SchedulerStatus{
			IsRunning:     true,
			IsCycleActive: true,
			NextRun:       &nextRun,
			LastRun:       &lastRun,
			IntervalHours: 12,
		}
		mockScheduler.On("GetStatus").Return(schedulerStatus).Once()

		servers := []services.ServerInfo{
			{ID: uuid.NewString(), Name: "OnlyRadarr", Type: "radarr"},
		}
		mockServerManager.On("ListServers").Return(servers, nil).Once()

		output, err := executeCommand(rootCmd, "status", "--json")
		assert.NoError(err)

		var statusInfo struct {
			Scheduler services.SchedulerStatus `json:"scheduler"`
			ServerCounts struct {
				Total  int `json:"total"`
				Radarr int `json:"radarr"`
				Sonarr int `json:"sonarr"`
			} `json:"serverCounts"`
			LastCycle struct {
				Active  bool      `json:"active"`
				LastRun *time.Time `json:"lastRun,omitempty"`
				NextRun *time.Time `json:"nextRun,omitempty"`
			} `json:"lastCycle"`
		}
		err = json.Unmarshal([]byte(output), &statusInfo)
		assert.NoError(err)
		assert.True(statusInfo.Scheduler.IsRunning)
		assert.True(statusInfo.Scheduler.IsCycleActive)
		assert.Equal(12, statusInfo.Scheduler.IntervalHours)
		assert.Equal(1, statusInfo.ServerCounts.Total)
		assert.Equal(1, statusInfo.ServerCounts.Radarr)
		assert.Equal(0, statusInfo.ServerCounts.Sonarr)
		mockScheduler.AssertExpectations(t)
		mockServerManager.AssertExpectations(t)
	})

	t.Run("status command - error listing servers", func(t *testing.T) {
		schedulerStatus := services.SchedulerStatus{
			IsRunning:     true,
			IsCycleActive: false,
			NextRun:       nil,
			LastRun:       nil,
			IntervalHours: 6,
		}
		mockScheduler.On("GetStatus").Return(schedulerStatus).Once()
		mockServerManager.On("ListServers").Return(([]services.ServerInfo)(nil), errors.New("db error")).Once()

		output, err := executeCommand(rootCmd, "status")
		assert.Error(err)
		assert.Contains(output, errorMsg("failed to list servers: db error"))
		mockScheduler.AssertExpectations(t)
		mockServerManager.AssertExpectations(t)
	})
}
