package cli

import "github.com/spf13/cobra"

var (
	dbPath  string
	version = "0.1.0"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "janitarr",
		Short:   "Automation tool for Radarr and Sonarr",
		Long:    `Janitarr automates content discovery and search triggering for media servers.`,
		Version: version,
	}
	cmd.PersistentFlags().StringVar(&dbPath, "db-path", "./data/janitarr.db", "Database path")
	return cmd
}

func Execute() error {
	return NewRootCmd().Execute()
}
