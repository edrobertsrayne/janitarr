# CLI Interface: Interactive Terminal Forms

## Context

Janitarr is primarily managed via command-line interface. Many operations require multiple inputs (server name, URL, API key, etc.) that are tedious to provide as individual flags. Interactive forms provide a better user experience for data entry, with inline validation, field navigation, and visual feedback.

This specification defines interactive CLI forms using the [charmbracelet/huh](https://github.com/charmbracelet/huh) library for operations that benefit from guided input.

## Technology Stack

- **Form library**: [charmbracelet/huh](https://github.com/charmbracelet/huh) for interactive terminal forms
- **Console output**: [charmbracelet/log](https://github.com/charmbracelet/log) for logging (see logging.md)

## Requirements

### Story: Interactive Server Addition

- **As a** user
- **I want to** add a server through an interactive form
- **So that** I can enter all required fields with validation feedback

#### Acceptance Criteria

- [ ] `janitarr server add` launches interactive form when no flags provided
- [ ] Form collects: server name, server type, URL, API key
- [ ] Server type selection: Radarr or Sonarr (radio/select field)
- [ ] Server name field validates: required, no duplicates, alphanumeric with dashes
- [ ] URL field validates: required, valid URL format, https preferred (warning for http)
- [ ] API key field: required, masked input (shows `****` while typing)
- [ ] Form shows validation errors inline as user types
- [ ] Tab/Enter navigates between fields
- [ ] Escape cancels form without saving
- [ ] On submit, form tests connection before saving
- [ ] Connection test shows spinner/progress indicator
- [ ] Success: server saved, confirmation message displayed
- [ ] Failure: error displayed, user can retry or cancel

#### Form Layout

```
┌─────────────────────────────────────────────────────────┐
│  Add Server                                             │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Server Type                                            │
│  ● Radarr   ○ Sonarr                                    │
│                                                         │
│  Server Name                                            │
│  ┌─────────────────────────────────────────────────┐    │
│  │ radarr-main                                     │    │
│  └─────────────────────────────────────────────────┘    │
│  Name must be unique, alphanumeric with dashes          │
│                                                         │
│  URL                                                    │
│  ┌─────────────────────────────────────────────────┐    │
│  │ http://192.168.1.100:7878                       │    │
│  └─────────────────────────────────────────────────┘    │
│  ⚠ Using HTTP - HTTPS recommended for security          │
│                                                         │
│  API Key                                                │
│  ┌─────────────────────────────────────────────────┐    │
│  │ ********************************                │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  [ Test & Save ]    [ Cancel ]                          │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Story: Interactive Server Editing

- **As a** user
- **I want to** edit an existing server through an interactive form
- **So that** I can update fields while seeing current values

#### Acceptance Criteria

- [ ] `janitarr server edit <name>` launches form pre-populated with existing values
- [ ] Server type field disabled (cannot change type after creation)
- [ ] Server name can be changed (validates uniqueness excluding current server)
- [ ] URL and API key editable with same validation as add form
- [ ] API key shown as masked (user must clear and re-enter to change)
- [ ] "Keep existing API key" checkbox option
- [ ] Cancel preserves original values (no changes made)
- [ ] Submit tests connection before saving changes

### Story: Interactive Configuration

- **As a** user
- **I want to** configure automation settings through an interactive form
- **So that** I can see all options and their current values in one place

#### Acceptance Criteria

- [ ] `janitarr config` launches interactive configuration form
- [ ] Form displays settings in logical groups:
  - **Automation**: enabled (yes/no), interval (hours), dry-run mode
  - **Search Limits**: missing movies, missing episodes, cutoff movies, cutoff episodes
- [ ] Toggle fields for boolean options (enabled, dry-run)
- [ ] Number fields with validation (interval: 1-168, limits: 0-100)
- [ ] Current values pre-populated from database
- [ ] Changes saved on submit, cancelled on escape
- [ ] Summary of changes displayed before final confirmation

#### Form Layout

```
┌─────────────────────────────────────────────────────────┐
│  Configuration                                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Automation                                             │
│  ─────────────────────────────────────────────────────  │
│  Enabled          [●] Yes  [ ] No                       │
│  Interval         [ 6    ] hours (1-168)                │
│  Dry Run          [ ] Yes  [●] No                       │
│                                                         │
│  Search Limits (per cycle)                              │
│  ─────────────────────────────────────────────────────  │
│  Missing Movies   [ 10   ]                              │
│  Missing Episodes [ 20   ]                              │
│  Cutoff Movies    [ 5    ]                              │
│  Cutoff Episodes  [ 10   ]                              │
│                                                         │
│  [ Save ]    [ Cancel ]                                 │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Story: Flag Override for Non-Interactive Use

- **As a** user running Janitarr in scripts or CI
- **I want to** bypass interactive forms using flags
- **So that** I can automate server management

#### Acceptance Criteria

- [ ] All interactive forms have equivalent flag-based invocation
- [ ] `janitarr server add --name X --type radarr --url Y --api-key Z` skips form
- [ ] `janitarr server edit <name> --url Y` updates only specified fields
- [ ] `janitarr config --interval 12 --dry-run=false` updates only specified settings
- [ ] Mixed mode: partial flags provided prompts only for missing required fields
- [ ] `--non-interactive` flag forces flag-only mode (errors if required fields missing)
- [ ] Piped input (stdin not a TTY) automatically uses non-interactive mode

### Story: Server List with Interactive Selection

- **As a** user
- **I want to** select a server from a list for operations
- **So that** I don't have to remember exact server names

#### Acceptance Criteria

- [ ] `janitarr server edit` (no name) shows interactive server selector
- [ ] `janitarr server delete` (no name) shows interactive server selector
- [ ] `janitarr server test` (no name) shows interactive server selector
- [ ] Selector shows server name, type, and enabled status
- [ ] Arrow keys navigate, Enter selects, Escape cancels
- [ ] Selected server proceeds to appropriate action (edit form, delete confirmation, etc.)

#### Selector Layout

```
┌─────────────────────────────────────────────────────────┐
│  Select Server                                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  > radarr-main      Radarr    Enabled                   │
│    radarr-4k        Radarr    Enabled                   │
│    sonarr-main      Sonarr    Enabled                   │
│    sonarr-anime     Sonarr    Disabled                  │
│                                                         │
│  ↑/↓ Navigate   Enter Select   Esc Cancel               │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Story: Confirmation Dialogs

- **As a** user
- **I want to** confirm destructive actions
- **So that** I don't accidentally delete data

#### Acceptance Criteria

- [ ] `janitarr server delete <name>` shows confirmation prompt
- [ ] Confirmation shows server name and type being deleted
- [ ] User must type server name to confirm (prevents accidental deletion)
- [ ] `--force` flag bypasses confirmation (for scripting)
- [ ] `janitarr logs clear` shows confirmation with log count
- [ ] All destructive actions have clear cancel option

#### Confirmation Layout

```
┌─────────────────────────────────────────────────────────┐
│  Delete Server                                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Are you sure you want to delete this server?           │
│                                                         │
│  Name: radarr-4k                                        │
│  Type: Radarr                                           │
│                                                         │
│  Type the server name to confirm:                       │
│  ┌─────────────────────────────────────────────────┐    │
│  │                                                 │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  [ Delete ]    [ Cancel ]                               │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Edge Cases & Constraints

### Terminal Compatibility

- Forms require TTY (interactive terminal)
- Gracefully degrade when stdin is not a TTY (use flags or error)
- Support standard terminal sizes (minimum 80x24)
- Handle terminal resize during form display
- Work in common terminals: iTerm2, Terminal.app, GNOME Terminal, Windows Terminal, tmux

### Accessibility

- Support keyboard-only navigation (no mouse required)
- Clear focus indicators on active field
- Error messages associated with relevant fields
- Color not the only indicator of state (use symbols too)

### Input Validation

- Validate on field blur and on submit
- Show validation errors inline below field
- Don't allow submit until all required fields valid
- URL validation allows http and https schemes
- API key validation: non-empty, 32 hexadecimal characters (matching Radarr/Sonarr API key format)
- Server name: alphanumeric, dashes, underscores; 1-50 chars

### Error Recovery

- Network errors during connection test show retry option
- Form state preserved on validation errors
- Escape always available to cancel without saving
- Ctrl+C exits cleanly without partial saves

### Security

- API key input masked (show `*` characters)
- API keys never displayed after entry (show "configured" or masked)
- Clipboard paste supported for API key field
- No API keys in command history (interactive mode advantage)

### Performance

- Forms render instantly (< 100ms)
- Connection tests show progress indicator
- Timeout for connection tests: 10 seconds
- Cancel available during long operations

## Implementation Notes

### charmbracelet/huh Form Structure

```go
import "github.com/charmbracelet/huh"

func serverAddForm() (*Server, error) {
    var (
        serverType string
        name       string
        url        string
        apiKey     string
    )

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Server Type").
                Options(
                    huh.NewOption("Radarr", "radarr"),
                    huh.NewOption("Sonarr", "sonarr"),
                ).
                Value(&serverType),

            huh.NewInput().
                Title("Server Name").
                Description("Unique name for this server").
                Validate(validateServerName).
                Value(&name),

            huh.NewInput().
                Title("URL").
                Description("Server URL (e.g., http://localhost:7878)").
                Validate(validateURL).
                Value(&url),

            huh.NewInput().
                Title("API Key").
                Description("Found in Settings > General").
                EchoMode(huh.EchoModePassword).
                Validate(validateAPIKey).
                Value(&apiKey),
        ),
    )

    err := form.Run()
    if err != nil {
        return nil, err
    }

    return &Server{
        Name:   name,
        Type:   serverType,
        URL:    url,
        APIKey: apiKey,
    }, nil
}
```

### Command Integration

```go
// src/cli/server.go
var serverAddCmd = &cobra.Command{
    Use:   "add",
    Short: "Add a new server",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Check if all required flags provided
        if hasAllRequiredFlags(cmd) {
            return addServerFromFlags(cmd)
        }

        // Check if interactive mode available
        if !isInteractive() {
            return fmt.Errorf("missing required flags; use --help for usage")
        }

        // Run interactive form
        return addServerInteractive()
    },
}

func isInteractive() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

### File Structure

```
src/
├── cli/
│   ├── server.go        # Server commands (add, edit, delete, list, test)
│   ├── config.go        # Config command
│   └── forms/
│       ├── server.go    # Server add/edit forms
│       ├── config.go    # Configuration form
│       ├── confirm.go   # Confirmation dialogs
│       └── select.go    # Server selector
```

## Success Metrics

1. **Usability**: Users can add servers without reading documentation
2. **Error Prevention**: Validation catches errors before submission
3. **Efficiency**: Fewer keystrokes than equivalent flag-based commands
4. **Compatibility**: Forms work in 95% of common terminal environments
5. **Scriptability**: Flag-based fallback enables automation

## Future Enhancements (Post-v1)

1. **Setup Wizard**: First-run wizard that guides through initial configuration
2. **Import/Export**: Interactive import of servers from backup file
3. **Bulk Operations**: Multi-select for batch server operations
4. **Themes**: User-selectable color schemes for forms
5. **bubbletea TUI**: Full terminal UI for advanced server management (if needed)
