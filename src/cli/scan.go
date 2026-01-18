package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/services"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan servers for missing and cutoff content (detection only)",
	RunE:  runScan,
}

func init() {
	scanCmd.Flags().Bool("json", false, "Output results as JSON")
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	outputJSON, _ := cmd.Flags().GetBool("json")

	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	detector := services.NewDetector(db)

	hideCursor()
	showProgress("Scanning servers for missing and cutoff content")

	detectionResults, err := detector.DetectAll(ctx)

	clearLine()
	showCursor()

	if err != nil {
		return fmt.Errorf(errorMsg("Error during scan: %w", err))
	}

	if outputJSON {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(detectionResults)
	}

	if len(detectionResults.Results) == 0 {
		fmt.Println(info("No servers configured or enabled for scanning."))
		return nil
	}

	fmt.Println(header("Scan Results:"))
	fmt.Printf("  Successful Scans: %d\n", detectionResults.SuccessCount)
	fmt.Printf("  Failed Scans: %d\n", detectionResults.FailureCount)
	fmt.Printf("  Total Missing Items: %d\n", detectionResults.TotalMissing)
	fmt.Printf("  Total Cutoff Unmet Items: %d\n", detectionResults.TotalCutoff)
	fmt.Println()

	for _, res := range detectionResults.Results {
		if res.Error != "" {
			fmt.Printf(errorMsg("Server %s (%s) Scan Failed: %s\n"), res.ServerName, res.ServerType, res.Error)
		} else {
			fmt.Printf(success("Server %s (%s) Scan Successful:\n"), res.ServerName, res.ServerType)
			fmt.Printf("  Missing Items: %d\n", len(res.Missing))
			fmt.Printf("  Cutoff Unmet Items: %d\n", len(res.Cutoff))
		}
	}

	return nil
}