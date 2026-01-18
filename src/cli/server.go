package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/services"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage Radarr/Sonarr server configurations",
}

var serverAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new media server",
	RunE:  runServerAdd,
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured media servers",
	RunE:  runServerList,
}

func init() {
	serverCmd.AddCommand(serverAddCmd)
	serverCmd.AddCommand(serverListCmd)

	serverListCmd.Flags().Bool("json", false, "Output list as JSON")
}

func runServerAdd(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not implemented")
}

func runServerList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := database.New(dbPath, "./data/.janitarr.key") // Assuming keyPath is managed globally or passed
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	serverManager := services.NewServerManager(db)
	servers, err := serverManager.ListServers()
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	outputJSON, _ := cmd.Flags().GetBool("json")

	if outputJSON {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(servers)
	}

	fmt.Println(formatServerTable(servers))
	return nil
}