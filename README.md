# Janitarr

Automation tool for managing Radarr and Sonarr media servers. Automatically detects missing content and quality upgrades, then triggers searches on a configurable schedule.

## Features

- **Multi-server support**: Manage multiple Radarr and Sonarr instances from a single tool
- **Smart detection**: Automatically finds missing episodes/movies and content below quality cutoffs
- **Flexible scheduling**: Run automation on a custom interval or manually trigger searches
- **Search limits**: Control how many searches to trigger per cycle to avoid overwhelming your indexers
- **Activity logging**: Track all automation activity with detailed logs
- **CLI interface**: Simple, intuitive command-line interface for all operations

## Quick Start

### Prerequisites

- [Bun](https://bun.sh/) runtime (v1.0 or later)
- At least one Radarr or Sonarr instance with API access

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd janitarr

# Install dependencies
bun install
```

### Basic Usage

1. **Add a media server:**
```bash
bun run src/index.ts server add
```

2. **Configure search limits:**
```bash
bun run src/index.ts config set limits.missing 10
bun run src/index.ts config set limits.cutoff 5
```

3. **Run a manual cycle:**
```bash
bun run src/index.ts run
```

4. **Start the scheduler:**
```bash
bun run src/index.ts start
```

## CLI Commands

### Server Management

```bash
# Add a new server (interactive)
janitarr server add

# List all configured servers
janitarr server list

# Test connection to a server
janitarr server test <name|id>

# Edit server configuration
janitarr server edit <name|id>

# Remove a server
janitarr server remove <name|id>
```

### Detection & Status

```bash
# Show current status (scheduler state, config, next run time)
janitarr status

# Scan for missing/cutoff content without triggering searches
janitarr scan

# Show detailed scan results
janitarr scan --json
```

### Automation

```bash
# Execute a full automation cycle immediately
janitarr run

# Start the scheduler daemon
janitarr start

# Stop the running scheduler
janitarr stop
```

### Configuration

```bash
# Display current configuration
janitarr config show

# Set configuration values
janitarr config set schedule.interval 6    # hours between cycles
janitarr config set schedule.enabled true  # enable/disable scheduler
janitarr config set limits.missing 10      # max missing content searches
janitarr config set limits.cutoff 5        # max quality upgrade searches
```

### Activity Logs

```bash
# Display recent logs (default: 50 entries)
janitarr logs

# Display all logs with pagination
janitarr logs --all

# Display logs as JSON
janitarr logs --json

# Clear all logs (with confirmation)
janitarr logs --clear
```

## Configuration

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `JANITARR_DB_PATH` | SQLite database location | `./data/janitarr.db` |
| `JANITARR_LOG_LEVEL` | Logging verbosity | `info` |

### Default Settings

- **Schedule interval**: 6 hours
- **Schedule enabled**: Yes
- **Missing content limit**: 10 items per cycle
- **Quality cutoff limit**: 5 items per cycle

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

### Running Tests

```bash
# Run all tests
bun test

# Type checking
bunx tsc --noEmit

# Linting
bunx eslint .
```

### Project Structure

```
janitarr/
├── src/
│   ├── lib/              # Shared utilities
│   │   ├── api-client.ts # Radarr/Sonarr API client
│   │   ├── logger.ts     # Activity logging
│   │   └── scheduler.ts  # Scheduling engine
│   ├── services/         # Business logic
│   │   ├── server-manager.ts # Server CRUD operations
│   │   ├── detector.ts       # Content detection
│   │   ├── search-trigger.ts # Search execution
│   │   └── automation.ts     # Cycle orchestration
│   ├── storage/          # Data persistence
│   │   └── database.ts   # SQLite interface
│   ├── cli/              # CLI interface
│   │   ├── commands.ts   # Command definitions
│   │   └── formatters.ts # Output formatting
│   ├── types.ts          # TypeScript types
│   └── index.ts          # Entry point
├── tests/                # Test files
├── specs/                # Requirements documentation
└── data/                 # Database storage (gitignored)
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

### Scheduler not running

- Start the scheduler with `janitarr start`
- Check scheduler status with `janitarr status`
- Verify `schedule.enabled` is set to `true` in configuration

## License

See LICENSE file for details.
