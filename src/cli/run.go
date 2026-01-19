package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/services"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute automation cycle manually",
	RunE:  runAutomation,
}

func init() {
	runCmd.Flags().BoolP("dry-run", "d", false, "Preview without triggering searches")
	runCmd.Flags().Bool("json", false, "Output as JSON")
}

func runAutomation(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	outputJSON, _ := cmd.Flags().GetBool("json")

	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Initialize services
	detector := services.NewDetector(db)
	appLogger := logger.NewLogger(db, logger.LevelInfo, false)
	trigger := services.NewSearchTrigger(db, appLogger)

	automation := services.NewAutomation(db, detector, trigger, appLogger)

	if dryRun {
		hideCursor()
		showProgress("Running automation cycle (DRY RUN - no searches will be triggered)")
	} else {
		hideCursor()
		showProgress("Running automation cycle")
	}

	cycleResult, err := automation.RunCycle(ctx, true, dryRun) // isManual = true for CLI run

	clearLine()
	showCursor()

	if err != nil && !outputJSON {
		fmt.Println(errorMsg(fmt.Sprintf("Automation cycle completed with errors: %s", err.Error())))
		if len(cycleResult.Errors) > 0 {
			fmt.Println(errorMsg("Details:"))
			for _, e := range cycleResult.Errors {
				fmt.Println(errorMsg(fmt.Sprintf("  - %s", e)))
			}
		}
	}

	if outputJSON {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(cycleResult)
	}

	fmt.Println(services.FormatCycleResult(cycleResult))
	return nil
}
