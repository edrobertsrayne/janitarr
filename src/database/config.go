package database

import (
	"database/sql"
	"strconv"
)

// GetAppConfigFunc is a variable that holds the function to retrieve the full application configuration.
// It can be overridden in tests to inject mock implementations.
var GetAppConfigFunc = func(db *DB) AppConfig {
	config := DefaultAppConfig()

	// Schedule settings
	if val := db.GetConfig("schedule.intervalHours"); val != nil {
		if i, err := strconv.Atoi(*val); err == nil {
			config.Schedule.IntervalHours = i
		}
	}

	if val := db.GetConfig("schedule.enabled"); val != nil {
		config.Schedule.Enabled = *val == "true"
	}

	// Search limits
	if val := db.GetConfig("limits.missing.movies"); val != nil {
		if i, err := strconv.Atoi(*val); err == nil {
			config.SearchLimits.MissingMoviesLimit = i
		}
	}

	if val := db.GetConfig("limits.missing.episodes"); val != nil {
		if i, err := strconv.Atoi(*val); err == nil {
			config.SearchLimits.MissingEpisodesLimit = i
		}
	}

	if val := db.GetConfig("limits.cutoff.movies"); val != nil {
		if i, err := strconv.Atoi(*val); err == nil {
			config.SearchLimits.CutoffMoviesLimit = i
		}
	}

	if val := db.GetConfig("limits.cutoff.episodes"); val != nil {
		if i, err := strconv.Atoi(*val); err == nil {
			config.SearchLimits.CutoffEpisodesLimit = i
		}
	}

	return config
}

// SetAppConfigFunc is a variable that holds the function to update application configuration.
// It can be overridden in tests to inject mock implementations.
var SetAppConfigFunc = func(db *DB, update AppConfig) error {
	// Instead of taking AppConfigUpdate, it takes a full AppConfig to simplify usage
	// and ensure all values are always written.
	if err := db.SetConfig("schedule.intervalHours", strconv.Itoa(update.Schedule.IntervalHours)); err != nil {
		return err
	}
	if err := db.SetConfig("schedule.enabled", strconv.FormatBool(update.Schedule.Enabled)); err != nil {
		return err
	}
	if err := db.SetConfig("limits.missing.movies", strconv.Itoa(update.SearchLimits.MissingMoviesLimit)); err != nil {
		return err
	}
	if err := db.SetConfig("limits.missing.episodes", strconv.Itoa(update.SearchLimits.MissingEpisodesLimit)); err != nil {
		return err
	}
	if err := db.SetConfig("limits.cutoff.movies", strconv.Itoa(update.SearchLimits.CutoffMoviesLimit)); err != nil {
		return err
	}
	if err := db.SetConfig("limits.cutoff.episodes", strconv.Itoa(update.SearchLimits.CutoffEpisodesLimit)); err != nil {
		return err
	}
	return nil
}

// GetConfig retrieves a single configuration value by key
func (db *DB) GetConfig(key string) *string {
	var value string
	err := db.conn.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	return &value
}

// SetConfig sets a configuration value
func (db *DB) SetConfig(key, value string) error {
	_, err := db.conn.Exec(`
		INSERT INTO config (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

// GetAppConfig retrieves the full application configuration.
// This calls the globally exposed GetAppConfigFunc.
func (db *DB) GetAppConfig() AppConfig {
	return GetAppConfigFunc(db)
}

// SetAppConfig updates application configuration.
// This calls the globally exposed SetAppConfigFunc.
func (db *DB) SetAppConfig(config AppConfig) error {
	return SetAppConfigFunc(db, config)
}

// ScheduleConfigUpdate represents optional schedule config updates
type ScheduleConfigUpdate struct {
	IntervalHours *int
	Enabled       *bool
}

// SearchLimitsUpdate represents optional search limits updates
type SearchLimitsUpdate struct {
	MissingMoviesLimit   *int
	MissingEpisodesLimit *int
	CutoffMoviesLimit    *int
	CutoffEpisodesLimit  *int
}

// AppConfigUpdate represents optional app config updates
type AppConfigUpdate struct {
	Schedule     *ScheduleConfigUpdate
	SearchLimits *SearchLimitsUpdate
}