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

var serverEditCmd = &cobra.Command{
	Use:   "edit <id-or-name>",
	Short: "Edit an existing media server",
	Args:  cobra.ExactArgs(1),
	RunE:  runServerEdit,
}

func init() {
	serverCmd.AddCommand(serverAddCmd)
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverEditCmd)

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
		fmt.Println(errorMsg("Invalid server type. Must be 'radarr' or 'sonarr'."))
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

func runServerEdit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)
	idOrName := args[0]

	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	serverManager := services.NewServerManagerFunc(db)

	existingServer, err := serverManager.GetServer(ctx, idOrName)
	if err != nil {
		return fmt.Errorf("failed to find server '%s': %w", idOrName, err)
	}
	if existingServer == nil {
		return fmt.Errorf(errorMsg("Server '%s' not found."), idOrName)
	}

	fmt.Println(header(fmt.Sprintf("Edit Server: %s", existingServer.Name)))
	fmt.Println("--------------------\n")
	fmt.Println(info("Leave blank to keep current value."))

	// Name
	fmt.Printf(info("Enter new name (current: %s): "), existingServer.Name)
	newNameInput, _ := reader.ReadString('\n')
	newName := strings.TrimSpace(newNameInput)
	if newName == "" {
		newName = existingServer.Name
	}

	// URL
	fmt.Printf(info("Enter new URL (current: %s): "), existingServer.URL)
	newURLInput, _ := reader.ReadString('\n')
	newURL := strings.TrimSpace(newURLInput)
	if newURL == "" {
		newURL = existingServer.URL
	}

	// API Key
	fmt.Printf(info("Enter new API Key (current: %s...): "), existingServer.APIKey[0:4]) // Only show first 4 chars for security
	newAPIKeyInput, _ := reader.ReadString('\n')
	newAPIKey := strings.TrimSpace(newAPIKeyInput)
	if newAPIKey == "" {
		newAPIKey = existingServer.APIKey
	}

	updates := services.ServerUpdate{}
	if newName != existingServer.Name {
		updates.Name = &newName
	}
	if newURL != existingServer.URL {
		updates.URL = &newURL
	}
	if newAPIKey != existingServer.APIKey {
		updates.APIKey = &newAPIKey
	}

	if updates.Name == nil && updates.URL == nil && updates.APIKey == nil {
		fmt.Println(info("No changes detected. Skipping update."))
		return nil
	}

	hideCursor()
	showProgress("Testing connection and updating server")

	err = serverManager.UpdateServer(ctx, existingServer.ID, updates)

	clearLine()
	showCursor()

	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	fmt.Println(success(fmt.Sprintf("Server '%s' updated successfully!", newName)))
	return nil
}