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
