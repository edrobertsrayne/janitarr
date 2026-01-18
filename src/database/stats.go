package database

import (
	"database/sql"
	"time"
)

// GetSystemStats retrieves system-wide statistics
func (db *DB) GetSystemStats() SystemStats {
	stats := SystemStats{}

	// Get server count
	_ = db.conn.QueryRow("SELECT COUNT(*) FROM servers WHERE enabled = 1").Scan(&stats.TotalServers)

	// Get searches in last 24 hours
	oneDayAgo := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	_ = db.conn.QueryRow("SELECT COALESCE(SUM(count), 0) FROM logs WHERE type = 'search' AND timestamp >= ?", oneDayAgo).Scan(&stats.SearchesLast24h)

	// Get errors in last 24 hours
	_ = db.conn.QueryRow("SELECT COUNT(*) FROM logs WHERE type = 'error' AND timestamp >= ?", oneDayAgo).Scan(&stats.ErrorsLast24h)

	// Get last cycle time
	var lastCycleStr sql.NullString
	_ = db.conn.QueryRow("SELECT MAX(timestamp) FROM logs WHERE type = 'cycle_end'").Scan(&lastCycleStr)
	if lastCycleStr.Valid {
		stats.LastCycleTime = lastCycleStr.String
	}

	return stats
}

// GetServerStats retrieves statistics for a specific server
func (db *DB) GetServerStats(serverID string) ServerStats {
	stats := ServerStats{}

	// Get server name first
	var serverName string
	err := db.conn.QueryRow("SELECT name FROM servers WHERE id = ?", serverID).Scan(&serverName)
	if err != nil {
		// Server not found, return empty stats
		return stats
	}

	// Get total searches for this server
	_ = db.conn.QueryRow("SELECT COALESCE(SUM(count), 0) FROM logs WHERE type = 'search' AND server_name = ?", serverName).Scan(&stats.TotalSearches)

	// Get total errors for this server
	_ = db.conn.QueryRow("SELECT COUNT(*) FROM logs WHERE type = 'error' AND server_name = ?", serverName).Scan(&stats.ErrorCount)

	// Get last activity time
	var lastActivityStr sql.NullString
	_ = db.conn.QueryRow("SELECT MAX(timestamp) FROM logs WHERE server_name = ?", serverName).Scan(&lastActivityStr)
	if lastActivityStr.Valid {
		stats.LastCheckTime = lastActivityStr.String
	}

	return stats
}
