package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/cli/forms"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
)

// confirmAction prompts the user for y/N confirmation (non-interactive fallback)
var confirmAction = func(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(warning(prompt + " (y/N): "))
	confirmation, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(confirmation)) == "y"
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View activity logs",
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().IntP("limit", "n", 20, "Number of entries to show")
	logsCmd.Flags().Bool("all", false, "Show all entries")
	logsCmd.Flags().Bool("json", false, "Output as JSON")
	logsCmd.Flags().Bool("clear", false, "Clear all logs")
}

func runLogs(cmd *cobra.Command, args []string) error {
	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	outputJSON, _ := cmd.Flags().GetBool("json")
	clearLogs, _ := cmd.Flags().GetBool("clear")
	showAll, _ := cmd.Flags().GetBool("all")
	limit, _ := cmd.Flags().GetInt("limit")

	if clearLogs {
		// Get log count for confirmation message
		logCount, err := db.GetLogCount(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get log count: %w", err)
		}

		// Use interactive confirmation if in TTY and not forced non-interactive
		var confirmed bool
		if forms.ShouldUseInteractiveMode(nonInteractive) {
			details := fmt.Sprintf("This will permanently delete %d log entries.\nThis action cannot be undone.", logCount)
			confirmed, err = forms.ConfirmActionWithDetails("Clear All Logs", details)
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
		} else {
			// Fall back to basic Y/N prompt for non-interactive mode
			confirmed = confirmAction(fmt.Sprintf("Are you sure you want to clear all %d logs? This action cannot be undone.", logCount))
		}

		if !confirmed {
			fmt.Println(info("Log clearing cancelled."))
			return nil
		}

		if err := db.ClearLogs(); err != nil {
			return fmt.Errorf("failed to clear logs: %w", err)
		}
		fmt.Println(success("All logs cleared successfully."))
		return nil
	}

	var logEntries []logger.LogEntry
	if showAll {
		// Implement pagination if needed for very large datasets, for now fetch all
		logEntries, err = db.GetLogs(context.Background(), 0, 0, logger.LogFilters{}) // Limit 0 means all
	} else {
		logEntries, err = db.GetLogs(context.Background(), limit, 0, logger.LogFilters{})
	}

	if err != nil {
		return fmt.Errorf("failed to retrieve logs: %w", err)
	}

	if outputJSON {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(logEntries)
	}

	if len(logEntries) == 0 {
		fmt.Println(info("No log entries found."))
		return nil
	}

	fmt.Println(header("Activity Logs:"))
	fmt.Println(formatLogTable(logEntries))

	return nil
}
