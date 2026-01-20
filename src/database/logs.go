package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/user/janitarr/src/logger" // Import logger package for LogEntry
)

// GetLogsFunc is a variable that holds the function to retrieve log entries.
// It can be overridden in tests to inject mock implementations.
var GetLogsFunc = func(db *DB, ctx context.Context, limit, offset int, filters logger.LogFilters) ([]logger.LogEntry, error) {
	query := "SELECT id, timestamp, type, server_name, server_type, category, count, message, is_manual, operation, metadata FROM logs WHERE 1=1"
	var args []any
	var logs []logger.LogEntry

	if filters.Type != nil && *filters.Type != "" {
		query += " AND type = ?"
		args = append(args, *filters.Type)
	}

	if filters.Server != nil && *filters.Server != "" {
		query += " AND server_name = ?"
		args = append(args, *filters.Server)
	}

	if filters.Operation != nil && *filters.Operation != "" {
		query += " AND operation = ?"
		args = append(args, *filters.Operation)
	}

	if filters.FromDate != nil && *filters.FromDate != "" {
		query += " AND timestamp >= ?"
		args = append(args, *filters.FromDate)
	}

	if filters.ToDate != nil && *filters.ToDate != "" {
		query += " AND timestamp <= ?"
		args = append(args, *filters.ToDate)
	}

	if filters.Search != nil && *filters.Search != "" {
		query += " AND message LIKE '%' || ? || '%'"
		args = append(args, *filters.Search)
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying logs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var logEntry logger.LogEntry
		var timestampStr string
		var serverName, serverType, category, operation, metadataJSON sql.NullString
		var count sql.NullInt64
		var isManual bool

		err := rows.Scan(
			&logEntry.ID,
			&timestampStr,
			&logEntry.Type,
			&serverName,
			&serverType,
			&category,
			&count,
			&logEntry.Message,
			&isManual,
			&operation,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning log entry: %w", err)
		}

		logEntry.Timestamp, err = time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("parsing timestamp: %w", err)
		}

		if serverName.Valid {
			logEntry.ServerName = serverName.String
		}
		if serverType.Valid {
			logEntry.ServerType = serverType.String
		}
		if category.Valid {
			logEntry.Category = category.String
		}
		if count.Valid {
			logEntry.Count = int(count.Int64)
		}
		logEntry.IsManual = isManual

		if operation.Valid {
			logEntry.Operation = operation.String
		}

		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &logEntry.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshaling metadata: %w", err)
			}
		}

		logs = append(logs, logEntry)
	}

	return logs, nil
}

// ClearLogsFunc is a variable that holds the function to clear all log entries.
// It can be overridden in tests to inject mock implementations.
var ClearLogsFunc = func(db *DB) error {
	_, err := db.conn.Exec("DELETE FROM logs")
	if err != nil {
		return fmt.Errorf("clearing logs: %w", err)
	}
	return nil
}

// AddLogFunc is a variable that holds the function to add a log entry.
// It can be overridden in tests to inject mock implementations.
var AddLogFunc = func(db *DB, entry logger.LogEntry) error {
	query := `
		INSERT INTO logs (id, timestamp, type, server_name, server_type, category, count, message, is_manual, operation, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Serialize metadata to JSON if present
	var metadataJSON any
	if entry.Metadata != nil && len(entry.Metadata) > 0 {
		jsonBytes, err := json.Marshal(entry.Metadata)
		if err != nil {
			return fmt.Errorf("marshaling metadata: %w", err)
		}
		metadataJSON = string(jsonBytes)
	}

	_, err := db.conn.Exec(query,
		entry.ID,
		entry.Timestamp.Format(time.RFC3339),
		entry.Type,
		nullString(entry.ServerName),
		nullString(entry.ServerType),
		nullString(entry.Category),
		nullInt(entry.Count),
		entry.Message,
		entry.IsManual,
		nullString(entry.Operation),
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("inserting log entry: %w", err)
	}
	return nil
}

// GetLogs retrieves log entries with optional filters.
// This calls the globally exposed GetLogsFunc.
func (db *DB) GetLogs(ctx context.Context, limit, offset int, filters logger.LogFilters) ([]logger.LogEntry, error) {
	return GetLogsFunc(db, ctx, limit, offset, filters)
}

// ClearLogs removes all log entries.
// This calls the globally exposed ClearLogsFunc.
func (db *DB) ClearLogs() error {
	return ClearLogsFunc(db)
}

// AddLog implements logger.LogStorer.AddLog
// This calls the globally exposed AddLogFunc.
func (db *DB) AddLog(entry logger.LogEntry) error {
	return AddLogFunc(db, entry)
}

// GetLogsByOperation retrieves log entries filtered by operation type.
func (db *DB) GetLogsByOperation(ctx context.Context, operation string, limit, offset int) ([]logger.LogEntry, error) {
	query := "SELECT id, timestamp, type, server_name, server_type, category, count, message, is_manual, operation, metadata FROM logs WHERE operation = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	var logs []logger.LogEntry

	rows, err := db.conn.QueryContext(ctx, query, operation, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("querying logs by operation: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var logEntry logger.LogEntry
		var timestampStr string
		var serverName, serverType, category, op, metadataJSON sql.NullString
		var count sql.NullInt64
		var isManual bool

		err := rows.Scan(
			&logEntry.ID,
			&timestampStr,
			&logEntry.Type,
			&serverName,
			&serverType,
			&category,
			&count,
			&logEntry.Message,
			&isManual,
			&op,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning log entry: %w", err)
		}

		logEntry.Timestamp, err = time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("parsing timestamp: %w", err)
		}

		if serverName.Valid {
			logEntry.ServerName = serverName.String
		}
		if serverType.Valid {
			logEntry.ServerType = serverType.String
		}
		if category.Valid {
			logEntry.Category = category.String
		}
		if count.Valid {
			logEntry.Count = int(count.Int64)
		}
		logEntry.IsManual = isManual

		if op.Valid {
			logEntry.Operation = op.String
		}

		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &logEntry.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshaling metadata: %w", err)
			}
		}

		logs = append(logs, logEntry)
	}

	return logs, nil
}

// GetErrorCount returns the count of error logs since the given time.
func (db *DB) GetErrorCount(ctx context.Context, since time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM logs WHERE type = 'error' AND timestamp >= ?"
	var count int
	err := db.conn.QueryRowContext(ctx, query, since.Format(time.RFC3339)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting error logs: %w", err)
	}
	return count, nil
}

// GetLogCount returns the total count of log entries.
func (db *DB) GetLogCount(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM logs"
	var count int
	err := db.conn.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting logs: %w", err)
	}
	return count, nil
}

// PurgeOldLogs deletes log entries older than the specified number of days.
// Returns the number of rows deleted.
func (db *DB) PurgeOldLogs(ctx context.Context, retentionDays int) (int, error) {
	if retentionDays < 1 {
		return 0, fmt.Errorf("retention days must be at least 1")
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays).Format(time.RFC3339)
	query := "DELETE FROM logs WHERE timestamp < ?"

	result, err := db.conn.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("purging old logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("getting rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// nullString converts an empty string to a nil interface
func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// nullInt converts a zero value int to a nil interface
func nullInt(i int) any {
	if i == 0 {
		return nil
	}
	return i
}
