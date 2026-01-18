package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/database"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Update a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)

	configShowCmd.Flags().Bool("json", false, "Output configuration as JSON")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	appConfig := database.GetAppConfigFunc(db)

	if outputJSON {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(appConfig)
	}

	fmt.Println(formatConfigTable(&appConfig))
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	appConfig := db.GetAppConfig()

	switch strings.ToLower(key) {
	case "schedule.interval":
		intVal, parseErr := strconv.Atoi(value)
		if parseErr != nil || intVal < 1 {
			return fmt.Errorf(errorMsg("Invalid value for schedule.interval: must be a positive integer"))
		}
		appConfig.Schedule.IntervalHours = intVal
	case "schedule.enabled":
		boolVal, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return fmt.Errorf(errorMsg("Invalid value for schedule.enabled: must be 'true' or 'false'"))
		}
		appConfig.Schedule.Enabled = boolVal
	case "limits.missing.movies":
		intVal, parseErr := strconv.Atoi(value)
		if parseErr != nil || intVal < 0 {
			return fmt.Errorf(errorMsg("Invalid value for limits.missing.movies: must be a non-negative integer"))
		}
		appConfig.SearchLimits.MissingMoviesLimit = intVal
	case "limits.missing.episodes":
		intVal, parseErr := strconv.Atoi(value)
		if parseErr != nil || intVal < 0 {
			return fmt.Errorf(errorMsg("Invalid value for limits.missing.episodes: must be a non-negative integer"))
		}
		appConfig.SearchLimits.MissingEpisodesLimit = intVal
	case "limits.cutoff.movies":
		intVal, parseErr := strconv.Atoi(value)
		if parseErr != nil || intVal < 0 {
			return fmt.Errorf(errorMsg("Invalid value for limits.cutoff.movies: must be a non-negative integer"))
		}
		appConfig.SearchLimits.CutoffMoviesLimit = intVal
	case "limits.cutoff.episodes":
		intVal, parseErr := strconv.Atoi(value)
		if parseErr != nil || intVal < 0 {
			return fmt.Errorf(errorMsg("Invalid value for limits.cutoff.episodes: must be a non-negative integer"))
		}
		appConfig.SearchLimits.CutoffEpisodesLimit = intVal
	default:
		return fmt.Errorf(errorMsg("Unknown configuration key: %s", key))
	}

	if err := database.SetAppConfigFunc(db, appConfig); err != nil {
		return fmt.Errorf("failed to set app config: %w", err)
	}

	fmt.Println(success(fmt.Sprintf("Configuration key '%s' updated to '%s'.", key, value)))
	fmt.Println(formatConfigTable(&appConfig))
	return nil
}