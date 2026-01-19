package forms

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/user/janitarr/src/database"
)

// ConfigForm displays an interactive form for editing application configuration
func ConfigForm(current database.AppConfig) (*database.AppConfig, error) {
	var result database.AppConfig = current

	// Convert booleans to strings for select fields
	var enabled string = "yes"
	if !current.Schedule.Enabled {
		enabled = "no"
	}

	// Convert integers to strings for input fields
	intervalStr := strconv.Itoa(current.Schedule.IntervalHours)
	missingMoviesStr := strconv.Itoa(current.SearchLimits.MissingMoviesLimit)
	missingEpisodesStr := strconv.Itoa(current.SearchLimits.MissingEpisodesLimit)
	cutoffMoviesStr := strconv.Itoa(current.SearchLimits.CutoffMoviesLimit)
	cutoffEpisodesStr := strconv.Itoa(current.SearchLimits.CutoffEpisodesLimit)
	retentionDaysStr := strconv.Itoa(current.Logs.RetentionDays)

	// Validator for interval (1-168 hours = 1 week)
	validateInterval := func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if val < 1 || val > 168 {
			return fmt.Errorf("must be between 1 and 168 hours")
		}
		return nil
	}

	// Validator for search limits (0-100)
	validateLimit := func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if val < 0 || val > 100 {
			return fmt.Errorf("must be between 0 and 100")
		}
		return nil
	}

	// Validator for retention days (7-90)
	validateRetention := func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if val < 7 || val > 90 {
			return fmt.Errorf("must be between 7 and 90 days")
		}
		return nil
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Configuration").
				Description("Configure automation settings"),

			huh.NewSelect[string]().
				Title("Automation Enabled").
				Description("Enable or disable automatic scheduling").
				Options(
					huh.NewOption("Yes", "yes"),
					huh.NewOption("No", "no"),
				).
				Value(&enabled),

			huh.NewInput().
				Title("Schedule Interval (hours)").
				Description("How often to run automation (1-168 hours)").
				Value(&intervalStr).
				Validate(validateInterval),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Search Limits").
				Description("Maximum items to search per automation cycle"),

			huh.NewInput().
				Title("Missing Movies Limit").
				Description("Maximum missing movies to search (0-100, 0=disabled)").
				Value(&missingMoviesStr).
				Validate(validateLimit),

			huh.NewInput().
				Title("Missing Episodes Limit").
				Description("Maximum missing episodes to search (0-100, 0=disabled)").
				Value(&missingEpisodesStr).
				Validate(validateLimit),

			huh.NewInput().
				Title("Cutoff Movies Limit").
				Description("Maximum movies needing quality upgrade (0-100, 0=disabled)").
				Value(&cutoffMoviesStr).
				Validate(validateLimit),

			huh.NewInput().
				Title("Cutoff Episodes Limit").
				Description("Maximum episodes needing quality upgrade (0-100, 0=disabled)").
				Value(&cutoffEpisodesStr).
				Validate(validateLimit),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Log Retention").
				Description("Configure log cleanup settings"),

			huh.NewInput().
				Title("Retention Days").
				Description("Days to keep logs before cleanup (7-90 days)").
				Value(&retentionDaysStr).
				Validate(validateRetention),
		),
	).WithTheme(huh.ThemeBase())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Parse results back to AppConfig
	result.Schedule.Enabled = (enabled == "yes")

	interval, _ := strconv.Atoi(intervalStr)
	result.Schedule.IntervalHours = interval

	missingMovies, _ := strconv.Atoi(missingMoviesStr)
	result.SearchLimits.MissingMoviesLimit = missingMovies

	missingEpisodes, _ := strconv.Atoi(missingEpisodesStr)
	result.SearchLimits.MissingEpisodesLimit = missingEpisodes

	cutoffMovies, _ := strconv.Atoi(cutoffMoviesStr)
	result.SearchLimits.CutoffMoviesLimit = cutoffMovies

	cutoffEpisodes, _ := strconv.Atoi(cutoffEpisodesStr)
	result.SearchLimits.CutoffEpisodesLimit = cutoffEpisodes

	retentionDays, _ := strconv.Atoi(retentionDaysStr)
	result.Logs.RetentionDays = retentionDays

	return &result, nil
}
