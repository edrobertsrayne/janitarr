package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// LogEntryInput represents input for creating a log entry
type LogEntryInput struct {
	Type       LogEntryType
	ServerName string
	ServerType ServerType
	Category   SearchCategory
	Count      int
	Message    string
	IsManual   bool
}

// AddLog adds a new log entry
func (db *DB) AddLog(input LogEntryInput) LogEntry {
	id := uuid.New().String()
	timestamp := time.Now().UTC()

	isManual := 0
	if input.IsManual {
		isManual = 1
	}

	var count *int
	if input.Count > 0 {
		count = &input.Count
	}

	db.conn.Exec(`
		INSERT INTO logs (id, timestamp, type, server_name, server_type, category, count, message, is_manual)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, timestamp.Format(time.RFC3339), input.Type, nullString(input.ServerName), nullString(string(input.ServerType)), nullString(string(input.Category)), count, input.Message, isManual)

	return LogEntry{
		ID:         id,
		Timestamp:  timestamp,
		Type:       input.Type,
		ServerName: input.ServerName,
		ServerType: input.ServerType,
		Category:   input.Category,
		Count:      input.Count,
		Message:    input.Message,
		IsManual:   input.IsManual,
	}
}

// GetLogs retrieves log entries with pagination
func (db *DB) GetLogs(limit, offset int) []LogEntry {
	rows, err := db.conn.Query(`
		SELECT id, timestamp, type, server_name, server_type, category, count, message, is_manual
		FROM logs ORDER BY timestamp DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil
	}
	defer rows.Close()

	return db.scanLogRows(rows)
}

// GetLogsPaginated retrieves log entries with filters and pagination
func (db *DB) GetLogsPaginated(filters LogFilters, limit, offset int) []LogEntry {
	query := "SELECT id, timestamp, type, server_name, server_type, category, count, message, is_manual FROM logs WHERE 1=1"
	var args []any

	if filters.Type != "" {
		query += " AND type = ?"
		args = append(args, filters.Type)
	}

	if filters.Server != "" {
		query += " AND server_name = ?"
		args = append(args, filters.Server)
	}

	if filters.StartDate != "" {
		query += " AND timestamp >= ?"
		args = append(args, filters.StartDate)
	}

	if filters.EndDate != "" {
		query += " AND timestamp <= ?"
		args = append(args, filters.EndDate)
	}

	if filters.Search != "" {
		query += " AND message LIKE ?"
		args = append(args, "%"+filters.Search+"%")
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	return db.scanLogRows(rows)
}

// GetLogCount returns the total number of log entries
func (db *DB) GetLogCount() int {
	var count int
	db.conn.QueryRow("SELECT COUNT(*) FROM logs").Scan(&count)
	return count
}

// ClearLogs removes all log entries
func (db *DB) ClearLogs() int {
	result, err := db.conn.Exec("DELETE FROM logs")
	if err != nil {
		return 0
	}
	rows, _ := result.RowsAffected()
	return int(rows)
}

// PurgeOldLogs removes log entries older than the retention period
func (db *DB) PurgeOldLogs() int {
	cutoff := time.Now().AddDate(0, 0, -LogRetentionDays).Format(time.RFC3339)
	result, err := db.conn.Exec("DELETE FROM logs WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0
	}
	rows, _ := result.RowsAffected()
	return int(rows)
}

// GetServerStats returns statistics for a specific server
func (db *DB) GetServerStats(serverID string) ServerStats {
	// Get server name first
	var serverName string
	err := db.conn.QueryRow("SELECT name FROM servers WHERE id = ?", serverID).Scan(&serverName)
	if err != nil {
		return ServerStats{}
	}

	var stats ServerStats

	// Count total searches for this server
	db.conn.QueryRow(`
		SELECT COUNT(*) FROM logs WHERE server_name = ? AND type = 'search'
	`, serverName).Scan(&stats.TotalSearches)

	// Count errors for this server
	db.conn.QueryRow(`
		SELECT COUNT(*) FROM logs WHERE server_name = ? AND type = 'error'
	`, serverName).Scan(&stats.ErrorCount)

	// Get last check time
	var lastCheck sql.NullString
	db.conn.QueryRow(`
		SELECT timestamp FROM logs WHERE server_name = ? ORDER BY timestamp DESC LIMIT 1
	`, serverName).Scan(&lastCheck)
	if lastCheck.Valid {
		stats.LastCheckTime = lastCheck.String
	}

	return stats
}

// GetSystemStats returns system-wide statistics
func (db *DB) GetSystemStats() SystemStats {
	var stats SystemStats

	// Count total servers
	db.conn.QueryRow("SELECT COUNT(*) FROM servers").Scan(&stats.TotalServers)

	// Get last cycle end time
	var lastCycle sql.NullString
	db.conn.QueryRow(`
		SELECT timestamp FROM logs WHERE type = 'cycle_end' ORDER BY timestamp DESC LIMIT 1
	`).Scan(&lastCycle)
	if lastCycle.Valid {
		stats.LastCycleTime = lastCycle.String
	}

	// Count searches in last 24 hours
	yesterday := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	db.conn.QueryRow(`
		SELECT COUNT(*) FROM logs WHERE type = 'search' AND timestamp >= ?
	`, yesterday).Scan(&stats.SearchesLast24h)

	// Count errors in last 24 hours
	db.conn.QueryRow(`
		SELECT COUNT(*) FROM logs WHERE type = 'error' AND timestamp >= ?
	`, yesterday).Scan(&stats.ErrorsLast24h)

	return stats
}

// scanLogRows scans multiple log rows
func (db *DB) scanLogRows(rows *sql.Rows) []LogEntry {
	var logs []LogEntry

	for rows.Next() {
		var entry LogEntry
		var timestamp string
		var serverName, serverType, category sql.NullString
		var count sql.NullInt64
		var isManual int

		err := rows.Scan(&entry.ID, &timestamp, &entry.Type, &serverName, &serverType, &category, &count, &entry.Message, &isManual)
		if err != nil {
			continue
		}

		entry.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		if serverName.Valid {
			entry.ServerName = serverName.String
		}
		if serverType.Valid {
			entry.ServerType = ServerType(serverType.String)
		}
		if category.Valid {
			entry.Category = SearchCategory(category.String)
		}
		if count.Valid {
			entry.Count = int(count.Int64)
		}
		entry.IsManual = isManual == 1

		logs = append(logs, entry)
	}

	return logs
}

// nullString converts an empty string to a nil interface
func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
