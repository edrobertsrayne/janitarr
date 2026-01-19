package forms

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/user/janitarr/src/database"
)

// ServerFormResult holds the result of a server form
type ServerFormResult struct {
	Name       string
	Type       string
	URL        string
	APIKey     string
	KeepAPIKey bool // For edit form: if true, don't update API key
}

// ServerAddForm displays an interactive form for adding a new server
func ServerAddForm(ctx context.Context, db *database.DB) (*ServerFormResult, error) {
	var result ServerFormResult

	// Validator that checks for duplicate server names
	validateUniqueName := func(name string) error {
		if err := ValidateServerName(name); err != nil {
			return err
		}

		// Check if name already exists
		existing, err := db.GetServerByName(name)
		if err == nil && existing != nil {
			return fmt.Errorf("server '%s' already exists", name)
		}
		return nil
	}

	// Validator that warns about HTTP (but doesn't block)
	validateURLWithWarning := func(urlStr string) error {
		if err := ValidateURL(urlStr); err != nil {
			return err
		}

		if strings.HasPrefix(strings.ToLower(urlStr), "http://") {
			// Return nil to allow, but we could add a note field if needed
			// For now, just validate it's a proper URL
		}
		return nil
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Server Type").
				Description("Select the type of media server").
				Options(
					huh.NewOption("Radarr (Movies)", "radarr"),
					huh.NewOption("Sonarr (TV Shows)", "sonarr"),
				).
				Value(&result.Type),

			huh.NewInput().
				Title("Server Name").
				Description("Unique identifier (e.g., radarr-main)").
				Placeholder("radarr-main").
				Validate(validateUniqueName).
				Value(&result.Name),

			huh.NewInput().
				Title("URL").
				Description("Server URL including port (e.g., http://localhost:7878)").
				Placeholder("http://localhost:7878").
				Validate(validateURLWithWarning).
				Value(&result.URL),

			huh.NewInput().
				Title("API Key").
				Description("Found in Settings > General in your server").
				EchoMode(huh.EchoModePassword).
				Validate(ValidateAPIKey).
				Value(&result.APIKey),
		),
	).WithTheme(huh.ThemeBase())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ServerEditForm displays an interactive form for editing an existing server
func ServerEditForm(ctx context.Context, db *database.DB, current *database.Server) (*ServerFormResult, error) {
	var result ServerFormResult

	// Pre-populate with current values
	result.Type = string(current.Type)
	result.Name = current.Name
	result.URL = current.URL
	result.KeepAPIKey = true // Default to keeping existing key

	var newAPIKey string
	var keepKey string = "yes"

	// Validator that checks for duplicate server names (excluding current server)
	validateUniqueNameEdit := func(name string) error {
		if err := ValidateServerName(name); err != nil {
			return err
		}

		// If name hasn't changed, it's valid
		if name == current.Name {
			return nil
		}

		// Check if new name already exists
		existing, err := db.GetServerByName(name)
		if err == nil && existing != nil {
			return fmt.Errorf("server '%s' already exists", name)
		}
		return nil
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Edit Server").
				Description(fmt.Sprintf("Editing: %s (%s)", current.Name, current.Type)),

			huh.NewInput().
				Title("Server Name").
				Description("Unique identifier").
				Value(&result.Name).
				Validate(validateUniqueNameEdit),

			huh.NewInput().
				Title("URL").
				Description("Server URL including port").
				Value(&result.URL).
				Validate(ValidateURL),

			huh.NewSelect[string]().
				Title("Update API Key?").
				Description("Choose whether to update the API key").
				Options(
					huh.NewOption("Keep existing key", "yes"),
					huh.NewOption("Enter new key", "no"),
				).
				Value(&keepKey),
		),
	).WithTheme(huh.ThemeBase())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// If user chose to update API key, show another form
	if keepKey == "no" {
		keyForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("New API Key").
					Description("Enter the new API key").
					EchoMode(huh.EchoModePassword).
					Validate(ValidateAPIKey).
					Value(&newAPIKey),
			),
		).WithTheme(huh.ThemeBase())

		err := keyForm.Run()
		if err != nil {
			return nil, err
		}
		result.APIKey = newAPIKey
		result.KeepAPIKey = false
	}

	return &result, nil
}

// ServerInfo represents a server for selection purposes
type ServerInfo struct {
	ID      string
	Name    string
	Type    string
	Enabled bool
}

// ServerSelector displays an interactive list for selecting a server
func ServerSelector(servers []ServerInfo) (*ServerInfo, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers configured")
	}

	// Build options with formatted display
	options := make([]huh.Option[string], len(servers))
	serverMap := make(map[string]*ServerInfo)

	for i, srv := range servers {
		status := "Enabled"
		if !srv.Enabled {
			status = "Disabled"
		}

		// Format: "radarr-main (Radarr) - Enabled"
		label := fmt.Sprintf("%s (%s) - %s", srv.Name, srv.Type, status)
		options[i] = huh.NewOption(label, srv.ID)

		// Store copy for lookup
		srvCopy := srv
		serverMap[srv.ID] = &srvCopy
	}

	var selectedID string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Server").
				Description("Choose a server from the list").
				Options(options...).
				Value(&selectedID),
		),
	).WithTheme(huh.ThemeBase())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	selected := serverMap[selectedID]
	if selected == nil {
		return nil, fmt.Errorf("selected server not found")
	}

	return selected, nil
}
