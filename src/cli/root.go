package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/version"
)

var (
	dbPath         string
	logLevel       string
	nonInteractive bool
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "janitarr",
		Short:   "Automation tool for Radarr and Sonarr",
		Long:    `Janitarr automates content discovery and search triggering for media servers.`,
		Version: version.Short(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate log level if provided
			if cmd.Flags().Changed("log-level") || logLevel != "" {
				if _, err := logger.ParseLevel(logLevel); err != nil {
					return fmt.Errorf("invalid log level: %w", err)
				}
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&dbPath, "db-path", "./data/janitarr.db", "Database path")

	// Get log level from environment variable first, then allow CLI flag to override
	envLogLevel := os.Getenv("JANITARR_LOG_LEVEL")
	defaultLogLevel := "info"
	if envLogLevel != "" {
		defaultLogLevel = envLogLevel
	}
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", defaultLogLevel, "Log level (debug, info, warn, error)")
	cmd.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "Force non-interactive mode (require all flags)")

	// Register commands
	cmd.AddCommand(startCmd)
	cmd.AddCommand(devCmd)
	cmd.AddCommand(serverCmd)
	cmd.AddCommand(configCmd)
	cmd.AddCommand(runCmd)
	cmd.AddCommand(scanCmd)
	cmd.AddCommand(statusCmd)
	cmd.AddCommand(logsCmd)

	return cmd
}

func Execute() error {
	return NewRootCmd().Execute()
}
