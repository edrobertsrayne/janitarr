package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/edrobertsrayne/janitarr/src/cli/forms"
	"github.com/edrobertsrayne/janitarr/src/database"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify configuration",
	Long:  "View and modify configuration. Run without subcommands to launch interactive form.",
	RunE:  runConfigInteractive,
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

	appConfig := db.GetAppConfig()

	outputJSON, _ := cmd.Flags().GetBool("json")
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
			return fmt.Errorf("invalid value for schedule.interval: must be a positive integer")
		}
		appConfig.Schedule.IntervalHours = intVal
	case "schedule.enabled":
		boolVal, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return fmt.Errorf("invalid value for schedule.enabled: must be 'true' or 'false'")
		}
		appConfig.Schedule.Enabled = boolVal
	case "limits.missing.movies":
		intVal, parseErr := strconv.Atoi(value)
		if parseErr != nil || intVal < 0 {
			return fmt.Errorf("invalid value for limits.missing.movies: must be a non-negative integer")
		}
		appConfig.SearchLimits.MissingMoviesLimit = intVal
	case "limits.missing.episodes":
		intVal, parseErr := strconv.Atoi(value)
		if parseErr != nil || intVal < 0 {
			return fmt.Errorf("invalid value for limits.missing.episodes: must be a non-negative integer")
		}
		appConfig.SearchLimits.MissingEpisodesLimit = intVal
	case "limits.cutoff.movies":
		intVal, parseErr := strconv.Atoi(value)
		if parseErr != nil || intVal < 0 {
			return fmt.Errorf("invalid value for limits.cutoff.movies: must be a non-negative integer")
		}
		appConfig.SearchLimits.CutoffMoviesLimit = intVal
	case "limits.cutoff.episodes":
		intVal, parseErr := strconv.Atoi(value)
		if parseErr != nil || intVal < 0 {
			return fmt.Errorf("invalid value for limits.cutoff.episodes: must be a non-negative integer")
		}
		appConfig.SearchLimits.CutoffEpisodesLimit = intVal
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if err := db.SetAppConfig(appConfig); err != nil {
		return fmt.Errorf("failed to set app config: %w", err)
	}

	fmt.Println(success(fmt.Sprintf("Configuration key '%s' updated to '%s'.", key, value)))
	fmt.Println(formatConfigTable(&appConfig))
	return nil
}

func runConfigInteractive(cmd *cobra.Command, args []string) error {
	// If not interactive, show help and available subcommands
	if !forms.ShouldUseInteractiveMode(nonInteractive) {
		fmt.Println(info("Not in interactive mode. Use 'config show' or 'config set' subcommands."))
		return cmd.Help()
	}

	// Open database
	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get current configuration
	currentConfig := db.GetAppConfig()

	// Show interactive form
	fmt.Println(header("Interactive Configuration"))
	fmt.Println()

	updatedConfig, err := forms.ConfigForm(currentConfig)
	if err != nil {
		// User cancelled or error occurred
		return nil
	}

	// Save updated configuration
	if err := db.SetAppConfig(*updatedConfig); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Show success message and updated configuration
	fmt.Println()
	fmt.Println(success("Configuration saved successfully!"))
	fmt.Println()
	fmt.Println(formatConfigTable(updatedConfig))

	return nil
}
