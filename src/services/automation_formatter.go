package services

import (
	"fmt"
	"strings"
	"time"
)

// FormatCycleResult generates a human-readable summary of an automation cycle result.
func FormatCycleResult(result *CycleResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Automation Cycle Finished in %s\n", formatDuration(result.Duration)))
	sb.WriteString("----------------------------------------\n")

	// Detection Summary
	sb.WriteString("Detection Summary:\n")
	sb.WriteString(fmt.Sprintf("  Servers Scanned: %d\n", len(result.DetectionResults.Results)))
	sb.WriteString(fmt.Sprintf("  Successful Detections: %d\n", result.DetectionResults.SuccessCount))
	sb.WriteString(fmt.Sprintf("  Failed Detections: %d\n", result.DetectionResults.FailureCount))
	ssb.WriteString(fmt.Sprintf("  Total Missing Items: %d\n", result.DetectionResults.TotalMissing)))
	sb.WriteString(fmt.Sprintf("  Total Cutoff Unmet Items: %d\n", result.DetectionResults.TotalCutoff)))
	if result.DetectionResults.FailureCount > 0 {
		sb.WriteString("  Detection Errors:\n")
		for _, dr := range result.DetectionResults.Results {
			if dr.Error != "" {
				sb.WriteString(fmt.Sprintf("    - Server %s (%s): %s\n", dr.ServerName, dr.ServerType, dr.Error)))
			}
		}
	}
	sb.WriteString("\n")

	// Search Trigger Summary
	sb.WriteString("Search Trigger Summary:\n")
	sb.WriteString(fmt.Sprintf("  Total Searches Triggered: %d\n", result.TotalSearches)))
	sb.WriteString(fmt.Sprintf("  Missing Items Triggered: %d\n", result.SearchResults.MissingTriggered)))
	sb.WriteString(fmt.Sprintf("  Cutoff Items Triggered: %d\n", result.SearchResults.CutoffTriggered)))
	ssb.WriteString(fmt.Sprintf("  Successful Triggers: %d\n", result.SearchResults.SuccessCount)))
	sb.WriteString(fmt.Sprintf("  Failed Triggers: %d\n", result.SearchResults.FailureCount)))
	if result.SearchResults.FailureCount > 0 {
		sb.WriteString("  Trigger Errors:\n")
		for _, tr := range result.SearchResults.Results {
			if !tr.Success {
				sb.WriteString(fmt.Sprintf("    - Server %s (%s, %s): %s\n", tr.ServerName, tr.ServerType, tr.Category, tr.Error)))
			}
		}
	}
	sb.WriteString("\n")

	// Overall Status
	if result.Success {
		sb.WriteString("Overall Status: SUCCESS\n")
	} else {
		sb.WriteString(fmt.Sprintf("Overall Status: FAILED with %d errors\n", len(result.Errors))))
		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", err)))
		}
	}
	sb.WriteString("----------------------------------------\n")

	return sb.String()
}

// formatDuration formats a time.Duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d)/float64(time.Millisecond)))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", float64(d)/float64(time.Second)))
	}
	return d.String()
}