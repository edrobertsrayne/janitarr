package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify configuration",
	Long:  `Display or update Janitarr configuration settings.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  `Show all configuration values with their current settings.`,
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Update a configuration value",
	Long: `Update a configuration setting. Valid keys:
  schedule.interval  - Hours between automation cycles (default: 6)
  schedule.enabled   - Whether scheduler is enabled (default: true)
  limits.missing     - Max missing searches per cycle (default: 10)
  limits.cutoff      - Max cutoff searches per cycle (default: 5)`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

func init() {
	configShowCmd.Flags().Bool("json", false, "Output as JSON")

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}
