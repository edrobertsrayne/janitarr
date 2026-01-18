package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan servers for missing and cutoff content (detection only)",
	Long:  `Scan all configured servers to detect missing and cutoff unmet content without triggering searches.`,
	RunE:  runScan,
}

func init() {
	scanCmd.Flags().Bool("json", false, "Output as JSON")
}

func runScan(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}
