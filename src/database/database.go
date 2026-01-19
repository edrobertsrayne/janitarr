package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/janitarr/src/crypto"
	_ "modernc.org/sqlite"
)

//go:embed migrations/001_initial_schema.sql
var migration001 string

//go:embed migrations/002_enhanced_logs.sql
var migration002 string

const (
	// LogRetentionDays is the number of days to keep log entries
	LogRetentionDays = 30
)

// DB represents the database connection and encryption key
type DB struct {
	conn      *sql.DB
	cryptoKey []byte
}

// New creates a new database connection and runs migrations.
// dbPath can be a file path or ":memory:" for an in-memory database.
// keyPath is the path to the encryption key file.
func New(dbPath, keyPath string) (*DB, error) {
	// Ensure parent directory exists for file-based databases
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("creating database directory: %w", err)
		}
	}

	// Load or create encryption key
	key, err := crypto.LoadOrCreateKey(keyPath)
	if err != nil {
		return nil, fmt.Errorf("loading encryption key: %w", err)
	}

	// Open database connection
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if dbPath != ":memory:" {
		if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
			conn.Close()
			return nil, fmt.Errorf("setting WAL mode: %w", err)
		}
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	db := &DB{
		conn:      conn,
		cryptoKey: key,
	}

	// Run migrations
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	// Set default config values
	if err := db.initializeDefaults(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("initializing defaults: %w", err)
	}

	return db, nil
}

// migrate runs database migrations with proper version tracking
func (db *DB) migrate() error {
	// Create migration tracking table if it doesn't exist
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("creating migration tracking table: %w", err)
	}

	migrations := []string{
		migration001,
		migration002,
	}

	for i, migration := range migrations {
		version := i + 1

		// Check if migration has already been applied
		var count int
		err := db.conn.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("checking migration status for version %d: %w", version, err)
		}

		if count > 0 {
			// Migration already applied, skip it
			continue
		}

		// Apply the migration
		if _, err := db.conn.Exec(migration); err != nil {
			return fmt.Errorf("migration %03d failed: %w", version, err)
		}

		// Record that migration was applied
		_, err = db.conn.Exec(
			"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
			version,
			time.Now().Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("recording migration %03d: %w", version, err)
		}
	}

	return nil
}

// initializeDefaults sets default configuration values if not present
func (db *DB) initializeDefaults() error {
	defaults := map[string]string{
		"schedule.intervalHours":  "6",
		"schedule.enabled":        "true",
		"limits.missing.movies":   "10",
		"limits.missing.episodes": "10",
		"limits.cutoff.movies":    "5",
		"limits.cutoff.episodes":  "5",
	}

	for key, value := range defaults {
		if err := db.setConfigDefault(key, value); err != nil {
			return fmt.Errorf("setting default %s: %w", key, err)
		}
	}

	return nil
}

// setConfigDefault sets a config value only if not already set
func (db *DB) setConfigDefault(key, value string) error {
	var existing string
	err := db.conn.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&existing)
	if err == nil {
		// Key already exists
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}

	// Key doesn't exist, insert default
	_, err = db.conn.Exec("INSERT INTO config (key, value) VALUES (?, ?)", key, value)
	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// TestConnection verifies the database connection is working
func (db *DB) TestConnection() bool {
	var result int
	err := db.conn.QueryRow("SELECT 1").Scan(&result)
	return err == nil && result == 1
}

// Ping verifies the database connection is alive
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// encryptAPIKey encrypts an API key for storage
func (db *DB) encryptAPIKey(apiKey string) (string, error) {
	return crypto.Encrypt(apiKey, db.cryptoKey)
}

// decryptAPIKey decrypts an API key from storage
func (db *DB) decryptAPIKey(encryptedKey string) (string, error) {
	return crypto.Decrypt(encryptedKey, db.cryptoKey)
}
