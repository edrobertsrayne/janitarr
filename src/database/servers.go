package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ServerUpdate represents optional fields for updating a server
type ServerUpdate struct {
	Name    *string
	URL     *string
	APIKey  *string
	Enabled *bool
}

// AddServer adds a new server to the database
func (db *DB) AddServer(name, url, apiKey string, serverType ServerType) (*Server, error) {
	if name == "" {
		return nil, fmt.Errorf("server name is required")
	}
	if url == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Encrypt API key
	encryptedKey, err := db.encryptAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("encrypting API key: %w", err)
	}

	now := time.Now().UTC()
	server := &Server{
		ID:        uuid.New().String(),
		Name:      name,
		URL:       url,
		APIKey:    apiKey, // Return unencrypted key to caller
		Type:      serverType,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = db.conn.Exec(`
		INSERT INTO servers (id, name, url, api_key, type, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, server.ID, server.Name, server.URL, encryptedKey, server.Type, 1, now.Format(time.RFC3339), now.Format(time.RFC3339))

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, fmt.Errorf("server with name '%s' already exists", name)
		}
		return nil, fmt.Errorf("inserting server: %w", err)
	}

	return server, nil
}

// GetServer retrieves a server by ID
func (db *DB) GetServer(id string) (*Server, error) {
	row := db.conn.QueryRow(`
		SELECT id, name, url, api_key, type, enabled, created_at, updated_at
		FROM servers WHERE id = ?
	`, id)

	return db.scanServer(row)
}

// GetServerByName retrieves a server by name (case-insensitive)
func (db *DB) GetServerByName(name string) (*Server, error) {
	row := db.conn.QueryRow(`
		SELECT id, name, url, api_key, type, enabled, created_at, updated_at
		FROM servers WHERE LOWER(name) = LOWER(?)
	`, name)

	return db.scanServer(row)
}

// GetAllServers retrieves all servers
func (db *DB) GetAllServers() ([]Server, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, url, api_key, type, enabled, created_at, updated_at
		FROM servers ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("querying servers: %w", err)
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		server, err := db.scanServerRow(rows)
		if err != nil {
			return nil, err
		}
		servers = append(servers, *server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating servers: %w", err)
	}

	return servers, nil
}

// GetServersByType retrieves all servers of a specific type
func (db *DB) GetServersByType(serverType ServerType) ([]Server, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, url, api_key, type, enabled, created_at, updated_at
		FROM servers WHERE type = ? ORDER BY name
	`, serverType)
	if err != nil {
		return nil, fmt.Errorf("querying servers: %w", err)
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		server, err := db.scanServerRow(rows)
		if err != nil {
			return nil, err
		}
		servers = append(servers, *server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating servers: %w", err)
	}

	return servers, nil
}

// ListServers returns a list of servers without decrypted API keys (for display)
func (db *DB) ListServers() ([]ServerInfo, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, type, enabled FROM servers ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("querying servers: %w", err)
	}
	defer rows.Close()

	var servers []ServerInfo
	for rows.Next() {
		var info ServerInfo
		var enabled int
		if err := rows.Scan(&info.ID, &info.Name, &info.Type, &enabled); err != nil {
			return nil, fmt.Errorf("scanning server: %w", err)
		}
		info.Enabled = enabled == 1
		servers = append(servers, info)
	}

	return servers, nil
}

// ServerInfo represents basic server info without API key
type ServerInfo struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Type    ServerType `json:"type"`
	Enabled bool       `json:"enabled"`
}

// ServerExists checks if a server with the same URL and type exists
func (db *DB) ServerExists(url string, serverType ServerType, excludeID string) bool {
	var query string
	var args []any

	if excludeID != "" {
		query = "SELECT 1 FROM servers WHERE url = ? AND type = ? AND id != ?"
		args = []any{url, serverType, excludeID}
	} else {
		query = "SELECT 1 FROM servers WHERE url = ? AND type = ?"
		args = []any{url, serverType}
	}

	var exists int
	err := db.conn.QueryRow(query, args...).Scan(&exists)
	return err == nil
}

// UpdateServer updates a server's fields
func (db *DB) UpdateServer(id string, updates *ServerUpdate) error {
	if updates == nil {
		return nil
	}

	// Build update query dynamically
	var setClauses []string
	var args []any

	if updates.Name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *updates.Name)
	}

	if updates.URL != nil {
		setClauses = append(setClauses, "url = ?")
		args = append(args, *updates.URL)
	}

	if updates.APIKey != nil {
		encryptedKey, err := db.encryptAPIKey(*updates.APIKey)
		if err != nil {
			return fmt.Errorf("encrypting API key: %w", err)
		}
		setClauses = append(setClauses, "api_key = ?")
		args = append(args, encryptedKey)
	}

	if updates.Enabled != nil {
		enabled := 0
		if *updates.Enabled {
			enabled = 1
		}
		setClauses = append(setClauses, "enabled = ?")
		args = append(args, enabled)
	}

	if len(setClauses) == 0 {
		return nil // Nothing to update
	}

	// Add updated_at
	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now().UTC().Format(time.RFC3339))

	// Add WHERE clause
	args = append(args, id)

	query := fmt.Sprintf("UPDATE servers SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("updating server: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("server not found: %s", id)
	}

	return nil
}

// DeleteServer removes a server from the database
func (db *DB) DeleteServer(id string) (bool, error) {
	result, err := db.conn.Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("deleting server: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("checking affected rows: %w", err)
	}

	return rows > 0, nil
}

// scanServer scans a single server row
func (db *DB) scanServer(row *sql.Row) (*Server, error) {
	var server Server
	var encryptedKey string
	var enabled int
	var createdAt, updatedAt string

	err := row.Scan(&server.ID, &server.Name, &server.URL, &encryptedKey, &server.Type, &enabled, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning server: %w", err)
	}

	// Decrypt API key
	apiKey, err := db.decryptAPIKey(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("decrypting API key: %w", err)
	}
	server.APIKey = apiKey
	server.Enabled = enabled == 1

	// Parse timestamps
	server.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	server.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &server, nil
}

// scanServerRow scans a server from rows iterator
func (db *DB) scanServerRow(rows *sql.Rows) (*Server, error) {
	var server Server
	var encryptedKey string
	var enabled int
	var createdAt, updatedAt string

	err := rows.Scan(&server.ID, &server.Name, &server.URL, &encryptedKey, &server.Type, &enabled, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("scanning server: %w", err)
	}

	// Decrypt API key
	apiKey, err := db.decryptAPIKey(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("decrypting API key: %w", err)
	}
	server.APIKey = apiKey
	server.Enabled = enabled == 1

	// Parse timestamps
	server.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	server.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &server, nil
}
