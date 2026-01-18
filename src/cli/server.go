package cli

import (
	"bufio" // Added for bufio.NewScanner
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(header("Add New Server"))
	fmt.Println("--------------------")

	// Name
	fmt.Print(info("Enter server name: "))
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf(errorMsg("Server name cannot be empty"))
	}

	// Type
	serverType := ""
	for {
		fmt.Print(info("Enter server type (radarr/sonarr): "))
		typeInput, _ := reader.ReadString('\n')
		typeInput = strings.ToLower(strings.TrimSpace(typeInput))
		if typeInput == "radarr" || typeInput == "sonarr" {
			serverType = typeInput
			break
		}
		fmt.Println(errorMsg("Invalid server type. Must be 'radarr' or 'sonarr'"))
	}

	// URL
	fmt.Print(info("Enter server URL (e.g., http://localhost:7878): "))
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf(errorMsg("Server URL cannot be empty"))
	}

	// API Key
	fmt.Print(info("Enter API Key: "))
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf(errorMsg("API Key cannot be empty"))
	}

	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	serverManager := services.NewServerManagerFunc(db) // Use NewServerManagerFunc

	hideCursor()
	showProgress("Testing connection")
	
	// Test connection and add server
	addedServer, err := serverManager.AddServer(ctx, name, url, apiKey, serverType)
	
	clearLine()
	showCursor()

	if err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	fmt.Println(success(fmt.Sprintf("Server '%s' (%s) added successfully!", addedServer.Name, addedServer.Type)))
	return nil
}

func runServerList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := database.New(dbPath, "./data/.janitarr.key") // Assuming keyPath is managed globally or passed
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	serverManager := services.NewServerManagerFunc(db) // Use NewServerManagerFunc
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
