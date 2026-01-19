package services

import (
	"context"
	"fmt"
	"time"

	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
)

// AutomationDetector defines the interface for content detection.
type AutomationDetector interface {
	DetectAll(ctx context.Context) (*DetectionResults, error)
}

// AutomationSearchTrigger defines the interface for triggering searches.
type AutomationSearchTrigger interface {
	TriggerSearches(ctx context.Context, detectionResults *DetectionResults, limits database.SearchLimits, dryRun bool) (*TriggerResults, error)
}

// AutomationLogger defines the interface for logging automation events.
type AutomationLogger interface {
	LogCycleStart(isManual bool) *logger.LogEntry
	LogCycleEnd(totalSearches, failures int, isManual bool) *logger.LogEntry
	LogDetectionComplete(serverName, serverType string, missing, cutoffUnmet int) *logger.LogEntry
	LogSearches(serverName, serverType, category string, count int, isManual bool) *logger.LogEntry
	LogServerError(serverName, serverType, reason string) *logger.LogEntry
	LogSearchError(serverName, serverType, category, reason string) *logger.LogEntry
}

// AutomationDB defines the interface for database operations needed by Automation.
// Note: AddLogEntry removed as logger handles that.
type AutomationDB interface {
	GetAppConfig() database.AppConfig
}

// Automation orchestrates the detection, search triggering, and logging process.
type Automation struct {
	db       AutomationDB
	detector AutomationDetector
	trigger  AutomationSearchTrigger
	logger   AutomationLogger
}

// NewAutomation creates a new Automation service.
// The db parameter should now represent the database operations needed by Automation,
// and the logger parameter should be a logger.Logger instance.
func NewAutomation(db AutomationDB, detector AutomationDetector, trigger AutomationSearchTrigger, appLogger AutomationLogger) *Automation {
	return &Automation{
		db:       db,
		detector: detector,
		trigger:  trigger,
		logger:   appLogger,
	}
}

// RunCycle executes a full automation cycle: detect, trigger searches, and log results.
func (a *Automation) RunCycle(ctx context.Context, isManual, dryRun bool) (*CycleResult, error) {
	startTime := time.Now()
	a.logger.LogCycleStart(isManual)

	cycleResult := &CycleResult{
		Success: true,
		Errors:  []string{},
	}

	// 1. Get application configuration for search limits
	config := a.db.GetAppConfig()

	// 2. Detect missing and cutoff content
	detectionResults, err := a.detector.DetectAll(ctx)
	if err != nil {
		cycleResult.Success = false
		cycleResult.Errors = append(cycleResult.Errors, fmt.Sprintf("detection failed: %v", err))
		// Continue with partial detection results if any were returned
	}
	cycleResult.DetectionResults = *detectionResults

	// Log detection completion and errors for each server
	for _, res := range detectionResults.Results {
		if res.Error != "" {
			if !dryRun { // Added condition for dryRun
				a.logger.LogServerError(res.ServerName, res.ServerType, fmt.Sprintf("detection error: %s", res.Error))
			}
			cycleResult.Success = false
			cycleResult.TotalFailures++
			cycleResult.Errors = append(cycleResult.Errors, fmt.Sprintf("server %s detection failed: %s", res.ServerName, res.Error))
		} else {
			// Log successful detection completion
			a.logger.LogDetectionComplete(res.ServerName, res.ServerType, len(res.Missing), len(res.Cutoff))
		}
	}
	// 3. Trigger searches
	triggerResults, err := a.trigger.TriggerSearches(ctx, detectionResults, config.SearchLimits, dryRun)
	if err != nil {
		cycleResult.Success = false
		cycleResult.Errors = append(cycleResult.Errors, fmt.Sprintf("triggering searches failed: %v", err))
		// Continue with partial trigger results if any were returned
	}
	cycleResult.SearchResults = *triggerResults

	cycleResult.TotalSearches = triggerResults.MissingTriggered + triggerResults.CutoffTriggered
	cycleResult.TotalFailures += triggerResults.FailureCount

	// 4. Log triggered searches
	if !dryRun {
		for _, result := range triggerResults.Results {
			if result.Success {
				if len(result.ItemIDs) > 0 {
					a.logger.LogSearches(result.ServerName, result.ServerType, result.Category, len(result.ItemIDs), isManual)
				}
			} else {
				a.logger.LogSearchError(result.ServerName, result.ServerType, result.Category, result.Error)
				cycleResult.Errors = append(cycleResult.Errors, fmt.Sprintf("server %s search failed for %s (%s): %s", result.ServerName, result.Category, result.ServerType, result.Error))
			}
		}
	}

	cycleResult.Duration = time.Since(startTime)
	a.logger.LogCycleEnd(cycleResult.TotalSearches, cycleResult.TotalFailures, isManual)

	if len(cycleResult.Errors) > 0 {
		return cycleResult, fmt.Errorf("automation cycle completed with %d errors", len(cycleResult.Errors))
	}

	return cycleResult, nil
}
