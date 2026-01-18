package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show scheduler and server status",
	Long:  `Display the current status of the scheduler, configured servers, and last cycle summary.`,
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().Bool("json", false, "Output as JSON")
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}
