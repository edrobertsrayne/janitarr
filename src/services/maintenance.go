package services

import (
	"context"
	"fmt"

	"github.com/edrobertsrayne/janitarr/src/database"
)

// LogCleanupLogger interface for logging cleanup operations
type LogCleanupLogger interface {
	Info(msg string, keyvals ...any)
	Error(msg string, keyvals ...any)
}

// RunLogCleanup deletes log entries older than the configured retention period.
// Returns the number of logs deleted and any error encountered.
func RunLogCleanup(ctx context.Context, db *database.DB, logger LogCleanupLogger) (int, error) {
	// Get retention configuration
	config := db.GetAppConfig()
	retentionDays := config.Logs.RetentionDays

	// Validate retention days (enforce minimum of 7 days as a safety measure)
	if retentionDays < 7 {
		retentionDays = 7
	}

	logger.Info("Starting log cleanup", "retention_days", retentionDays)

	// Purge old logs
	deletedCount, err := db.PurgeOldLogs(ctx, retentionDays)
	if err != nil {
		logger.Error("Failed to purge old logs", "error", err)
		return 0, fmt.Errorf("purging old logs: %w", err)
	}

	if deletedCount > 0 {
		logger.Info("Log cleanup completed", "deleted_count", deletedCount, "retention_days", retentionDays)
	} else {
		logger.Info("Log cleanup completed", "deleted_count", 0, "retention_days", retentionDays)
	}

	return deletedCount, nil
}
