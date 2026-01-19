package cli

import (
	"bufio" // Added for bufio.NewScanner
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/janitarr/src/cli/forms"
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
	Use:   "edit [id-or-name]",
	Short: "Edit an existing media server",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runServerEdit,
}

var serverRemoveCmd = &cobra.Command{
	Use:   "remove [id-or-name]",
	Short: "Remove a media server",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runServerRemove,
}

func init() {
	serverCmd.AddCommand(serverAddCmd)
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverEditCmd)
	serverCmd.AddCommand(serverRemoveCmd)

	// Server add flags
	serverAddCmd.Flags().String("name", "", "Server name")
	serverAddCmd.Flags().String("type", "", "Server type (radarr/sonarr)")
	serverAddCmd.Flags().String("url", "", "Server URL")
	serverAddCmd.Flags().String("api-key", "", "Server API key")

	// Server edit flags
	serverEditCmd.Flags().String("name", "", "New server name")
	serverEditCmd.Flags().String("url", "", "New server URL")
	serverEditCmd.Flags().String("api-key", "", "New server API key")

	serverListCmd.Flags().Bool("json", false, "Output list as JSON")
	serverRemoveCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runServerAdd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	var name, url, apiKey, serverType string

	// Check if flags are provided
	flagName, _ := cmd.Flags().GetString("name")
	flagType, _ := cmd.Flags().GetString("type")
	flagURL, _ := cmd.Flags().GetString("url")
	flagAPIKey, _ := cmd.Flags().GetString("api-key")

	hasAllFlags := flagName != "" && flagType != "" && flagURL != "" && flagAPIKey != ""

	// Use interactive form if no flags and terminal is interactive
	if !hasAllFlags && forms.ShouldUseInteractiveMode(nonInteractive) {
		fmt.Println(header("Add New Server"))
		fmt.Println()

		result, err := forms.ServerAddForm(ctx, db)
		if err != nil {
			return fmt.Errorf("form cancelled or failed: %w", err)
		}

		name = result.Name
		serverType = result.Type
		url = result.URL
		apiKey = result.APIKey
	} else if hasAllFlags {
		// Use flags
		name = flagName
		serverType = strings.ToLower(flagType)
		url = flagURL
		apiKey = flagAPIKey

		// Validate inputs
		if err := forms.ValidateServerName(name); err != nil {
			return fmt.Errorf("invalid server name: %w", err)
		}
		if err := forms.ValidateServerType(serverType); err != nil {
			return fmt.Errorf("invalid server type: %w", err)
		}
		if err := forms.ValidateURL(url); err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		if err := forms.ValidateAPIKey(apiKey); err != nil {
			return fmt.Errorf("invalid API key: %w", err)
		}
	} else {
		return fmt.Errorf("missing required flags: --name, --type, --url, --api-key (or run without flags for interactive mode)")
	}

	serverManager := services.NewServerManagerFunc(db)

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

	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	var existingServer *database.Server

	// If no argument provided, show server selector in interactive mode
	if len(args) == 0 {
		if !forms.ShouldUseInteractiveMode(nonInteractive) {
			return fmt.Errorf("server name or ID required in non-interactive mode")
		}

		// Get all servers for selector
		serverManager := services.NewServerManagerFunc(db)
		servers, err := serverManager.ListServers()
		if err != nil {
			return fmt.Errorf("failed to list servers: %w", err)
		}

		if len(servers) == 0 {
			return fmt.Errorf("no servers configured")
		}

		// Convert to ServerInfo format
		serverInfos := make([]forms.ServerInfo, len(servers))
		for i, srv := range servers {
			serverInfos[i] = forms.ServerInfo{
				ID:      srv.ID,
				Name:    srv.Name,
				Type:    string(srv.Type),
				Enabled: srv.Enabled,
			}
		}

		// Show selector
		selected, err := forms.ServerSelector(serverInfos)
		if err != nil {
			return fmt.Errorf("server selection cancelled or failed: %w", err)
		}

		// Get full server object with API key
		existingServer, err = db.GetServer(selected.ID)
		if err != nil || existingServer == nil {
			return fmt.Errorf("failed to load server: %w", err)
		}
	} else {
		// Argument provided - look up server
		idOrName := args[0]

		// First try by ID, then by name to get full Server object (includes APIKey)
		existingServer, err = db.GetServer(idOrName)
		if err != nil || existingServer == nil {
			existingServer, err = db.GetServerByName(idOrName)
			if err != nil {
				return fmt.Errorf("failed to find server '%s': %w", idOrName, err)
			}
			if existingServer == nil {
				return fmt.Errorf("server '%s' not found", idOrName)
			}
		}
	}

	serverManager := services.NewServerManagerFunc(db)

	// Check if flags are provided
	flagName, _ := cmd.Flags().GetString("name")
	flagURL, _ := cmd.Flags().GetString("url")
	flagAPIKey, _ := cmd.Flags().GetString("api-key")

	hasFlags := flagName != "" || flagURL != "" || flagAPIKey != ""

	var result *forms.ServerFormResult

	// If interactive and no flags provided, use the form
	if forms.ShouldUseInteractiveMode(nonInteractive) && !hasFlags {
		result, err = forms.ServerEditForm(ctx, db, existingServer)
		if err != nil {
			return fmt.Errorf("form cancelled or failed: %w", err)
		}
	} else {
		// Use flags or fall back to old behavior
		result = &forms.ServerFormResult{
			Name:       flagName,
			URL:        flagURL,
			APIKey:     flagAPIKey,
			KeepAPIKey: flagAPIKey == "",
		}

		// If no flags provided and not interactive, error
		if !hasFlags {
			return fmt.Errorf("no flags provided and not in interactive mode")
		}
	}

	// Build updates from result
	updates := services.ServerUpdate{}

	if result.Name != "" && result.Name != existingServer.Name {
		updates.Name = &result.Name
	}

	if result.URL != "" && result.URL != existingServer.URL {
		updates.URL = &result.URL
	}

	if !result.KeepAPIKey && result.APIKey != "" {
		updates.APIKey = &result.APIKey
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

	finalName := existingServer.Name
	if updates.Name != nil {
		finalName = *updates.Name
	}

	fmt.Println(success(fmt.Sprintf("Server '%s' updated successfully!", finalName)))
	return nil
}

func runServerRemove(cmd *cobra.Command, args []string) error {
	db, err := database.New(dbPath, "./data/.janitarr.key")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	serverManager := services.NewServerManagerFunc(db)

	var serverToRemove *database.Server

	// If no argument provided, show server selector in interactive mode
	if len(args) == 0 {
		if !forms.ShouldUseInteractiveMode(nonInteractive) {
			return fmt.Errorf("server name or ID required in non-interactive mode")
		}

		// Get all servers for selector
		servers, err := serverManager.ListServers()
		if err != nil {
			return fmt.Errorf("failed to list servers: %w", err)
		}

		if len(servers) == 0 {
			return fmt.Errorf("no servers configured")
		}

		// Convert to ServerInfo format
		serverInfos := make([]forms.ServerInfo, len(servers))
		for i, srv := range servers {
			serverInfos[i] = forms.ServerInfo{
				ID:      srv.ID,
				Name:    srv.Name,
				Type:    string(srv.Type),
				Enabled: srv.Enabled,
			}
		}

		// Show selector
		selected, err := forms.ServerSelector(serverInfos)
		if err != nil {
			return fmt.Errorf("server selection cancelled or failed: %w", err)
		}

		// Get full server object from database
		serverToRemove, err = db.GetServer(selected.ID)
		if err != nil || serverToRemove == nil {
			return fmt.Errorf("failed to load server: %w", err)
		}
	} else {
		// Argument provided - look up server
		idOrName := args[0]

		// Try by ID first
		serverToRemove, err = db.GetServer(idOrName)
		if err != nil || serverToRemove == nil {
			// Try by name
			serverToRemove, err = db.GetServerByName(idOrName)
			if err != nil {
				return fmt.Errorf("failed to find server '%s': %w", idOrName, err)
			}
			if serverToRemove == nil {
				return fmt.Errorf("server '%s' not found", idOrName)
			}
		}
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force {
		// Use interactive confirmation if available
		if forms.ShouldUseInteractiveMode(nonInteractive) {
			confirmed, err := forms.ConfirmDelete("Server", serverToRemove.Name)
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if !confirmed {
				fmt.Println(info("Server removal cancelled."))
				return nil
			}
		} else {
			// Fallback to basic confirmation if not interactive
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf(warning("Are you sure you want to remove server '%s' (%s)? (y/N): "), serverToRemove.Name, serverToRemove.Type)
			confirmation, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(confirmation)) != "y" {
				fmt.Println(info("Server removal cancelled."))
				return nil
			}
		}
	}

	hideCursor()
	showProgress(fmt.Sprintf("Removing server '%s'", serverToRemove.Name))

	err = serverManager.RemoveServer(serverToRemove.ID)

	clearLine()
	showCursor()

	if err != nil {
		return fmt.Errorf("failed to remove server: %w", err)
	}

	fmt.Println(success(fmt.Sprintf("Server '%s' removed successfully!", serverToRemove.Name)))
	return nil
}
