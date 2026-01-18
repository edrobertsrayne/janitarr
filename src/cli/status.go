package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/services"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show scheduler and server status",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().Bool("json", false, "Output status as JSON")
}

func runStatus(cmd *cobra.Command, args []string) error {
	outputJSON, _ := cmd.Flags().GetBool("json")

	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Scheduler Status
	schedulerStatus := services.GetSchedulerStatusFunc(db)

	// Server counts
	servers, err := services.NewServerManager(db).ListServers()
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}
	radarrCount := 0
	sonarrCount := 0
	for _, s := range servers {
		if s.Type == "radarr" {
			radarrCount++
		} else if s.Type == "sonarr" {
			sonarrCount++
		}
	}

	// Last cycle summary (fetch from logs or a dedicated config value if available)
	// For now, we'll use a placeholder or assume it's part of schedulerStatus if possible
	// or fetch from logs directly. As there's no direct "last cycle summary" in DB,
	// let's just indicate if a cycle is active.

	statusInfo := struct {
		Scheduler      services.SchedulerStatus `json:"scheduler"`
		ServerCounts   struct {
			Total   int `json:"total"`
			Radarr  int `json:"radarr"`
			Sonarr  int `json:"sonarr"`
		} `json:"serverCounts"`
		LastCycle struct {
			Active bool `json:"active"`
			LastRun *time.Time `json:"lastRun,omitempty"`
			NextRun *time.Time `json:"nextRun,omitempty"`
		} `json:"lastCycle"`
	}{
		Scheduler: schedulerStatus,
		ServerCounts: struct {
			Total   int `json:"total"`
			Radarr  int `json:"radarr"`
			Sonarr  int `json:"sonarr"`
		}{
			Total:   len(servers),
			Radarr:  radarrCount,
			Sonarr:  sonarrCount,
		},
		LastCycle: struct {
			Active bool `json:"active"`
			LastRun *time.Time `json:"lastRun,omitempty"`
			NextRun *time.Time `json:"nextRun,omitempty"`
		}{
			Active: schedulerStatus.IsCycleActive,
			LastRun: schedulerStatus.LastRun,
			NextRun: schedulerStatus.NextRun,
		},
	}

	if outputJSON {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(statusInfo)
	}

	fmt.Println(header("Janitarr Status:"))
	fmt.Println("--------------------")

	fmt.Println(info("Scheduler Status:"))
	fmt.Printf("  Running: %s\n", formatBool(schedulerStatus.IsRunning))
	fmt.Printf("  Cycle Active: %s\n", formatBool(schedulerStatus.IsCycleActive))
	if schedulerStatus.NextRun != nil {
		fmt.Printf("  Next Run: %s (in %s)\n", schedulerStatus.NextRun.Format(time.RFC822), time.Until(*schedulerStatus.NextRun).Round(time.Second))
	} else {
		fmt.Println("  Next Run: N/A")
	}
	if schedulerStatus.LastRun != nil {
		fmt.Printf("  Last Run: %s (%s ago)\n", schedulerStatus.LastRun.Format(time.RFC822), time.Since(*schedulerStatus.LastRun).Round(time.Second))
	} else {
		fmt.Println("  Last Run: N/A")
	}
	fmt.Printf("  Interval: %d hours\n", schedulerStatus.IntervalHours)
	fmt.Println()

	fmt.Println(info("Server Overview:"))
	fmt.Printf("  Total Configured: %d\n", statusInfo.ServerCounts.Total)
	fmt.Printf("  Radarr Servers: %d\n", statusInfo.ServerCounts.Radarr)
	fmt.Printf("  Sonarr Servers: %d\n", statusInfo.ServerCounts.Sonarr)
	fmt.Println()

	// Placeholder for last cycle summary until actual implementation exists
	// fmt.Println(info("Last Automation Cycle:"))
	// fmt.Printf("  Status: %s\n", "N/A")
	// fmt.Printf("  Searches Triggered: %s\n", "N/A")
	// fmt.Printf("  Errors: %s\n", "N/A")

	return nil
}

func formatBool(b bool) string {
	if b {
		return success("Yes")
	}
	return warning("No")
}