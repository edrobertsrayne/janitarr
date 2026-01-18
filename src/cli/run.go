package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute automation cycle manually",
	Long:  `Run the full automation cycle: detect missing/cutoff content and trigger searches.`,
	RunE:  runAutomation,
}

func init() {
	runCmd.Flags().BoolP("dry-run", "d", false, "Preview without triggering searches")
	runCmd.Flags().Bool("json", false, "Output as JSON")
}

func runAutomation(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}
