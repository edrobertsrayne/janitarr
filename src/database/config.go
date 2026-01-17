package database

import (
	"database/sql"
	"strconv"
)

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

// GetAppConfig retrieves the full application configuration
func (db *DB) GetAppConfig() AppConfig {
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

// SetAppConfig updates application configuration
func (db *DB) SetAppConfig(update AppConfigUpdate) error {
	if update.Schedule != nil {
		if update.Schedule.IntervalHours != nil {
			if err := db.SetConfig("schedule.intervalHours", strconv.Itoa(*update.Schedule.IntervalHours)); err != nil {
				return err
			}
		}
		if update.Schedule.Enabled != nil {
			val := "false"
			if *update.Schedule.Enabled {
				val = "true"
			}
			if err := db.SetConfig("schedule.enabled", val); err != nil {
				return err
			}
		}
	}

	if update.SearchLimits != nil {
		if update.SearchLimits.MissingMoviesLimit != nil {
			if err := db.SetConfig("limits.missing.movies", strconv.Itoa(*update.SearchLimits.MissingMoviesLimit)); err != nil {
				return err
			}
		}
		if update.SearchLimits.MissingEpisodesLimit != nil {
			if err := db.SetConfig("limits.missing.episodes", strconv.Itoa(*update.SearchLimits.MissingEpisodesLimit)); err != nil {
				return err
			}
		}
		if update.SearchLimits.CutoffMoviesLimit != nil {
			if err := db.SetConfig("limits.cutoff.movies", strconv.Itoa(*update.SearchLimits.CutoffMoviesLimit)); err != nil {
				return err
			}
		}
		if update.SearchLimits.CutoffEpisodesLimit != nil {
			if err := db.SetConfig("limits.cutoff.episodes", strconv.Itoa(*update.SearchLimits.CutoffEpisodesLimit)); err != nil {
				return err
			}
		}
	}

	return nil
}
