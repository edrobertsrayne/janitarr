# Janitarr

Automation tool for managing Radarr and Sonarr media servers. Automatically detects missing content and quality upgrades, then triggers searches on a configurable schedule.

## Features

- **Multi-server support**: Manage multiple Radarr and Sonarr instances from a single tool
- **Smart detection**: Automatically finds missing episodes/movies and content below quality cutoffs
- **Flexible scheduling**: Run automation on a custom interval or manually trigger searches
- **Granular search limits**: Four independent limits for movies/episodes, missing/upgrades
- **Activity logging**: Track all automation activity with detailed logs
- **Web interface**: Modern, responsive web UI with real-time updates
- **CLI interface**: Simple, intuitive command-line interface for all operations
- **Dry-run mode**: Preview automation cycles before executing searches
- **Single binary**: No runtime dependencies, compiled to native code
- **Secure**: API keys encrypted at rest with AES-256-GCM

## Quick Start

### Prerequisites

- Go 1.22+ (for building from source)
- At least one Radarr or Sonarr instance with API access

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd janitarr

# Build
make build

# Run in production mode
./janitarr start

# Run in development mode (verbose logging)
./janitarr dev
```

### Basic Usage

#### Using the Web Interface

1. **Start the server:**

```bash
./janitarr start
```

2. **Open your browser to:** `http://localhost:3434`

3. **Add servers** via the web UI, configure settings, and monitor activity

#### Using the CLI

1. **Add a media server:**

```bash
./janitarr server add
```

2. **Configure search limits:**

```bash
./janitarr config set limits.missing.movies 10
./janitarr config set limits.missing.episodes 10
./janitarr config set limits.cutoff.movies 5
./janitarr config set limits.cutoff.episodes 5
```

3. **Run a manual cycle:**

```bash
./janitarr run
```

4. **Start both scheduler and web server:**

```bash
./janitarr start
```

For development with hot-reloading:

```bash
make dev  # Runs Air for auto-rebuild on file changes
```

## Web Interface

Janitarr includes a modern, responsive web interface for easy management.

### Features

- **Dashboard**: Real-time overview of missing content, quality upgrades, and server status
- **Server Management**: Add, edit, test, and manage Radarr/Sonarr servers
- **Activity Logs**: View and filter automation logs with real-time streaming
- **Settings**: Configure automation schedule and search limits
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Dark/Light Mode**: Manual dark mode toggle with localStorage persistence

### Accessing the Web UI

1. Start the Janitarr server:

   ```bash
   ./janitarr start
   ```

2. Open your browser to: `http://localhost:3434`

3. Navigate between views using the sidebar menu

**Development mode** (with hot-reloading):

```bash
make dev  # Runs Air for auto-rebuild
```

Access the app at `http://localhost:3434`.

## CLI Commands

| Command                             | Description                           |
| ----------------------------------- | ------------------------------------- |
| `janitarr start`                    | Start scheduler and web server        |
| `janitarr dev`                      | Development mode with verbose logging |
| `janitarr server add`               | Add a new server                      |
| `janitarr server list`              | List all servers                      |
| `janitarr server test <id\|name>`   | Test connection to a server           |
| `janitarr server edit <id\|name>`   | Edit server configuration             |
| `janitarr server remove <id\|name>` | Remove a server                       |
| `janitarr run`                      | Run automation cycle manually         |
| `janitarr run --dry-run`            | Preview what would be searched        |
| `janitarr scan`                     | Scan for missing/cutoff content       |
| `janitarr status`                   | Show scheduler status                 |
| `janitarr logs`                     | View activity logs                    |
| `janitarr logs -n 50`               | Show last 50 log entries              |
| `janitarr logs --clear`             | Clear all logs                        |
| `janitarr config show`              | Show configuration                    |
| `janitarr config set <key> <value>` | Update configuration                  |

### Server Management

```bash
# Add a new server (interactive)
./janitarr server add

# List all configured servers
./janitarr server list

# List servers as JSON
./janitarr server list --json

# Test connection to a server
./janitarr server test <name|id>

# Edit server configuration
./janitarr server edit <name|id>

# Remove a server
./janitarr server remove <name|id>

# Remove server without confirmation
./janitarr server remove <name|id> --force
```

### Detection & Status

```bash
# Show current status (scheduler state, config, next run time)
./janitarr status

# Show status as JSON
./janitarr status --json

# Scan for missing/cutoff content without triggering searches
./janitarr scan

# Show detailed scan results
./janitarr scan --json
```

### Automation

```bash
# Execute a full automation cycle immediately
./janitarr run

# Preview what would be searched (dry-run)
./janitarr run --dry-run

# Run and output as JSON
./janitarr run --json

# Start scheduler + web server (production mode)
./janitarr start

# Start in development mode (with verbose logging)
./janitarr dev
```

**Port configuration:**

```bash
# Start on custom port
./janitarr start --port 8080

# Bind to all interfaces (for remote access)
./janitarr start --host 0.0.0.0 --port 3434
```

### Configuration

```bash
# Display current configuration
./janitarr config show

# Display configuration as JSON
./janitarr config show --json

# Set configuration values
./janitarr config set schedule.interval 6              # hours between cycles
./janitarr config set schedule.enabled true            # enable/disable scheduler
./janitarr config set limits.missing.movies 10         # max missing movie searches
./janitarr config set limits.missing.episodes 10       # max missing episode searches
./janitarr config set limits.cutoff.movies 5           # max cutoff movie searches
./janitarr config set limits.cutoff.episodes 5         # max cutoff episode searches
```

### Activity Logs

```bash
# Display recent logs (default: 20 entries)
./janitarr logs

# Display specific number of logs
./janitarr logs -n 50

# Display all logs
./janitarr logs --all

# Display logs as JSON
./janitarr logs --json

# Clear all logs (with confirmation)
./janitarr logs --clear
```

## Configuration

### Environment Variables

| Variable             | Purpose                  | Default              |
| -------------------- | ------------------------ | -------------------- |
| `JANITARR_DB_PATH`   | SQLite database location | `./data/janitarr.db` |
| `JANITARR_LOG_LEVEL` | Logging verbosity        | `info`               |

### Default Settings

- **Schedule interval**: 6 hours
- **Schedule enabled**: Yes
- **Missing movies limit**: 10 items per cycle
- **Missing episodes limit**: 10 items per cycle
- **Movie upgrades limit**: 5 items per cycle
- **Episode upgrades limit**: 5 items per cycle

### Data Storage

All configuration, server credentials, and logs are stored in a SQLite database at `./data/janitarr.db` (configurable via `JANITARR_DB_PATH`).

## How It Works

1. **Detection Phase**: Janitarr queries each configured server for:
   - Missing monitored content (movies/episodes marked as wanted but not downloaded)
   - Content below quality cutoffs (downloaded but below the desired quality level)

2. **Search Triggering**: Based on configured limits, Janitarr:
   - Distributes searches fairly across all servers
   - Triggers searches using Radarr/Sonarr's command API
   - Logs all triggered searches for audit purposes

3. **Scheduling**: The automation cycle runs:
   - On a configurable interval (default: every 6 hours)
   - Immediately on startup (first run)
   - On demand via manual trigger

## Development

### Environment Setup

The project uses [devenv](https://devenv.sh) with direnv for automatic environment loading:

```bash
direnv allow              # Authorize the development environment
```

This provides Go, templ, Air, Tailwind CSS, and Playwright. The environment loads automatically when entering the project directory.

### Building

```bash
# Generate templates and build
make build

# Run the application
./janitarr --help

# Development with hot reload
make dev                  # Runs Air for auto-rebuild on file changes

# Generate templates only
templ generate

# Build Tailwind CSS
npx tailwindcss -i ./static/css/input.css -o ./static/css/app.css
```

### Running Tests

```bash
# Run all Go tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./src/crypto/...
go test ./src/database/...
go test ./src/api/...
go test ./src/services/...

# Run E2E tests (requires running server)
bunx playwright test --headless
bunx playwright show-report  # View test report
```

**Before running E2E tests**, start the server:

```bash
./janitarr start                      # Terminal 1
bunx playwright test --headless       # Terminal 2
```

### Project Structure

```
janitarr/
├── src/                    # All Go source code
│   ├── main.go             # Entry point
│   ├── cli/                # Cobra CLI commands
│   ├── api/                # Radarr/Sonarr API clients
│   ├── database/           # SQLite operations
│   ├── services/           # Business logic
│   ├── web/                # HTTP server and handlers
│   ├── templates/          # templ HTML templates
│   ├── logger/             # Activity logging
│   └── metrics/            # Prometheus metrics
├── static/                 # CSS and JS assets
├── migrations/             # SQL migration files
├── tests/                  # E2E tests
├── src-ts/                 # Original TypeScript (reference)
├── ui-ts/                  # Original React UI (reference)
└── specs/                  # Requirements documentation
```

## Architecture

### Key Design Principles

- **Single-instance**: Designed to run as a single process (no distributed state)
- **Fail gracefully**: Continue processing other servers if one fails
- **Audit trail**: Log all automation actions for transparency
- **API-first**: All operations use official Radarr/Sonarr APIs

### API Integration

Janitarr uses the following Radarr/Sonarr v3 API endpoints:

- `GET /api/v3/system/status` - Connection validation
- `GET /api/v3/wanted/missing` - Missing content detection
- `GET /api/v3/wanted/cutoff` - Quality cutoff detection
- `POST /api/v3/command` - Search triggering

Authentication is handled via the `X-Api-Key` header.

## Troubleshooting

### Server connection fails

- Verify the URL is correct (include http:// or https://)
- Ensure the API key is valid
- Check network connectivity to the server
- Verify the server is running and accessible

### Searches not triggering

- Check that search limits are not set to 0 (disabled)
- Verify servers have detected missing/cutoff content (`janitarr scan`)
- Review logs for errors (`janitarr logs`)
- Ensure scheduler is enabled (`janitarr config show`)

### Services not running

- Start services with `janitarr start` (runs scheduler + web server)
- Check scheduler status with `janitarr status`
- Verify `schedule.enabled` is set to `true` in configuration
- Use `janitarr dev` for development mode with verbose logging

## Technology Stack

| Component     | Technology          | Purpose                  |
| ------------- | ------------------- | ------------------------ |
| Language      | Go 1.22+            | Main application         |
| Web Framework | Chi (go-chi/chi/v5) | HTTP routing             |
| Database      | modernc.org/sqlite  | SQLite (pure Go, no CGO) |
| CLI           | Cobra (spf13/cobra) | Command-line interface   |
| Templates     | templ (a-h/templ)   | Type-safe HTML templates |
| Interactivity | htmx + Alpine.js    | Dynamic UI               |
| CSS           | Tailwind CSS        | Utility-first styling    |
| Hot Reload    | Air                 | Development workflow     |

## Security Notes

- API keys are encrypted at rest using AES-256-GCM
- Encryption key stored in `data/.janitarr.key`
- Default host binding is `localhost` (prevents external access)
- No authentication in v1 - relies on network-level access control

## Migration from TypeScript Version

This is a complete rewrite from TypeScript to Go. Key changes:

- **Fresh database** - users must re-add servers
- **New encryption key** - not compatible with old encrypted data
- **Server-rendered HTML** - React UI replaced with templ + htmx + Alpine.js
- **Single binary** - no Node.js/Bun runtime required
- **Better performance** - lower memory footprint, faster startup

## Documentation

- **[CLAUDE.md](CLAUDE.md)** - AI assistant guidelines and development workflow
- **[specs/](specs/)** - Feature specifications and requirements
- **[IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)** - Migration roadmap and task tracking

## Support

- **Issues**: Report bugs or request features on GitHub Issues
- **Specifications**: See detailed specs in the `specs/` directory

## License

See LICENSE file for details.
