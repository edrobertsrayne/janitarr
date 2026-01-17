# Janitarr User Guide

Comprehensive guide to using Janitarr for automating your Radarr and Sonarr media server management.

## Table of Contents

- [Introduction](#introduction)
- [Getting Started](#getting-started)
- [Web Interface](#web-interface)
- [Command-Line Interface](#command-line-interface)
- [Configuration](#configuration)
- [Workflows](#workflows)
- [Best Practices](#best-practices)

---

## Introduction

### What is Janitarr?

Janitarr is an automation tool designed to keep your media library up-to-date by:

- **Detecting missing content**: Finds monitored movies and TV episodes that haven't been downloaded yet
- **Finding quality upgrades**: Identifies content that's below your quality cutoff settings
- **Triggering searches automatically**: Searches for content on a configurable schedule
- **Managing multiple servers**: Supports multiple Radarr and Sonarr instances simultaneously

### Why Use Janitarr?

Without Janitarr, you would need to:
- Manually check each server for missing content
- Remember to search for quality upgrades
- Risk overwhelming your indexers with too many searches

With Janitarr:
- Automated detection runs on your schedule
- Search limits prevent indexer bans
- Fair distribution across multiple servers
- Comprehensive logging for audit trails

---

## Getting Started

### Prerequisites

Before using Janitarr, you need:

1. **Bun runtime** (v1.0 or later) - [Install from bun.sh](https://bun.sh/)
2. **At least one media server**:
   - Radarr v3 or later, OR
   - Sonarr v3 or later
3. **API access** to your media server(s)

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd janitarr
   ```

2. Install dependencies:
   ```bash
   bun install
   ```

3. Build the web interface (optional, if using web UI):
   ```bash
   cd ui
   bun install
   bun run build
   cd ..
   ```

### First-Time Setup

#### Option 1: Using the Web Interface

1. Start the Janitarr server:
   ```bash
   bun run start
   ```

2. Open your browser to `http://localhost:3434`

3. Navigate to the **Servers** page and click **Add Server**

4. Fill in your server details:
   - **Name**: A friendly name (e.g., "Main Radarr")
   - **Type**: Choose Radarr or Sonarr
   - **URL**: Your server's URL (e.g., `http://192.168.1.100:7878`)
   - **API Key**: Found in your server's Settings → General → Security
   - **Enabled**: Check to enable the server

5. Click **Test Connection** to verify, then **Add**

6. Go to **Settings** to configure:
   - **Automation Schedule**: How often to run (default: 6 hours)
   - **Search Limits**: How many searches per cycle (default: 10 missing movies, 10 missing episodes, 5 movie upgrades, 5 episode upgrades)

#### Option 2: Using the Command-Line Interface

1. Add a server interactively:
   ```bash
   bun run src/index.ts server add
   ```

   Follow the prompts to enter your server details.

2. Configure search limits:
   ```bash
   bun run src/index.ts config set limits.missing.movies 10
   bun run src/index.ts config set limits.missing.episodes 10
   bun run src/index.ts config set limits.cutoff.movies 5
   bun run src/index.ts config set limits.cutoff.episodes 5
   ```

3. Test your server connection:
   ```bash
   bun run src/index.ts server test "Main Radarr"
   ```

4. Run your first automation cycle:
   ```bash
   bun run src/index.ts run
   ```

---

## Web Interface

The web interface provides a modern, user-friendly way to manage Janitarr.

### Accessing the Web Interface

1. Start the server (production mode):
   ```bash
   bun run start
   ```

2. Open your browser to: `http://localhost:3434`

**Custom port:**
```bash
bun run src/index.ts start --port 8080
```

**Remote access:**
```bash
bun run src/index.ts start --host 0.0.0.0 --port 3434
```

**Development mode** (with hot-reloading):
```bash
# Terminal 1: Start backend with Vite proxy
bun run src/index.ts dev

# Terminal 2: Start Vite dev server
cd ui && bun run dev
```
Access at `http://localhost:3434` (backend proxies to Vite).

The web interface is fully responsive and works on desktop, tablet, and mobile devices.

### Dashboard

The Dashboard provides an at-a-glance view of your automation system.

**Status Cards** (top row):
- **Missing Movies**: Count of movies waiting to be downloaded
- **Missing Episodes**: Count of TV episodes waiting to be downloaded
- **Cutoff Upgrades**: Count of items below quality cutoff
- **Total Searches**: Lifetime count of triggered searches

**Server Status Table**:
- Lists all configured servers
- Shows connection status (green = connected, red = error)
- Displays server type (Radarr/Sonarr)
- Quick actions: Test connection, Edit, Disable/Enable

**Recent Activity Timeline**:
- Shows the last 10 automation events
- Color-coded by event type
- Displays timestamps and details

**Quick Actions**:
- **Scan Now**: Preview what would be searched (dry-run mode)
- **Run Automation**: Trigger a full automation cycle immediately
- **Refresh**: Update the dashboard with latest data

### Servers Page

Manage all your Radarr and Sonarr servers.

**View Modes**:
- **List View**: Compact table with all server details
- **Card View**: Visual cards with server information

**Adding a Server**:
1. Click **Add Server**
2. Fill in the form:
   - **Name**: Unique identifier (e.g., "4K Radarr", "Anime Sonarr")
   - **Type**: Radarr or Sonarr
   - **URL**: Full URL including protocol (`http://` or `https://`)
   - **API Key**: From your server's settings
   - **Enabled**: Whether to include in automation cycles
3. Click **Test Connection** to verify
4. Click **Add** to save

**Testing a Server**:
- Click the test icon next to any server
- Verifies URL accessibility and API key validity
- Shows success or error message

**Editing a Server**:
- Click the edit icon next to any server
- Modify any field except the type
- Test connection before saving

**Deleting a Server**:
- Click the delete icon next to any server
- Confirm the deletion
- Server configuration and logs are permanently removed

**Disabling a Server**:
- Toggle the enabled status
- Disabled servers are skipped during automation cycles
- Useful for temporarily excluding a server

### Logs Page

View and analyze all automation activity.

**Features**:
- **Real-time Streaming**: New logs appear automatically via WebSocket
- **Search**: Filter logs by text content
- **Type Filter**: Show only specific event types (automation, search, error, etc.)
- **Server Filter**: Show logs for a specific server
- **Export**: Download logs as JSON or CSV

**Log Entry Details**:
- **Timestamp**: When the event occurred
- **Type**: Event category (automation, search, server-test, etc.)
- **Server**: Which server was involved (if applicable)
- **Details**: Description of what happened

**Managing Logs**:
- **Refresh**: Manually fetch latest logs
- **Clear All**: Delete all logs (with confirmation)
- Logs automatically expire after 30 days

**Connection Status**:
- Green chip: Connected to WebSocket, receiving real-time updates
- Orange chip: Disconnected, using polling fallback
- Click "Refresh" if not receiving updates

### Settings Page

Configure all automation behavior.

**Automation Schedule Section**:
- **Enable Automation**: Master switch for scheduled automation
- **Interval (hours)**: Time between automation cycles (minimum: 1 hour)
- **Next Run**: Displays when the next cycle will execute

**Search Limits Section**:

Four independent limits control how many searches are triggered per cycle:

- **Missing Movies** (`limits.missing.movies`): Max Radarr missing movie searches
- **Missing Episodes** (`limits.missing.episodes`): Max Sonarr missing episode searches
- **Movie Upgrades** (`limits.cutoff.movies`): Max Radarr quality upgrade searches
- **Episode Upgrades** (`limits.cutoff.episodes`): Max Sonarr quality upgrade searches

Each limit is applied independently. For example, if you set:
- Missing Movies: 10
- Missing Episodes: 20
- Movie Upgrades: 5
- Episode Upgrades: 10

A single cycle could trigger up to 45 total searches (10+20+5+10).

**Why Separate Limits?**

- **Prevent indexer bans**: Most indexers limit requests per day
- **Prioritize content types**: Give more quota to TV shows if desired
- **Balance missing vs upgrades**: Focus on new content over quality improvements
- **Distribute fairly**: Searches are distributed round-robin across servers

**Advanced Section**:
- **Database Path**: Location of SQLite database (read-only display)
- **Log Retention**: Days to keep logs (30 days, not configurable)

**Saving Changes**:
- Click **Save Changes** to apply configuration
- Click **Reset** to discard changes and reload current settings
- Success/error messages appear at top of page

---

## Command-Line Interface

The CLI provides full control over Janitarr from your terminal.

### Server Management

#### Add a Server

Interactive mode (recommended):
```bash
janitarr server add
```

You'll be prompted for:
- Server name
- Type (Radarr or Sonarr)
- URL
- API key

The CLI validates your input and tests the connection before saving.

#### List Servers

```bash
janitarr server list
```

Shows all configured servers with their type, URL, and status.

#### Test Server Connection

```bash
janitarr server test <name>
```

Example:
```bash
janitarr server test "Main Radarr"
```

Verifies:
- URL is accessible
- API key is valid
- Server responds within timeout

#### Edit Server

```bash
janitarr server edit <name>
```

Interactive editing of existing server configuration.

#### Remove Server

```bash
janitarr server remove <name>
```

Permanently deletes server configuration. Requires confirmation.

### Detection & Status

#### View System Status

```bash
janitarr status
```

Displays:
- Scheduler state (running/stopped)
- Current configuration
- Next scheduled run time
- Database location

#### Scan for Content

```bash
janitarr scan
```

Performs detection across all servers without triggering searches:
- Shows counts of missing movies, missing episodes, and quality upgrades
- Displays which servers were checked
- Lists sample items that would be searched

Options:
- `--json`: Output in JSON format for scripting

**Use Cases**:
- Preview automation results before running
- Check if servers have any work to do
- Verify detection is working correctly

### Automation

#### Run Automation Cycle

```bash
janitarr run
```

Executes a full automation cycle immediately:
1. Detects missing content and quality upgrades on all servers
2. Applies configured search limits
3. Triggers searches
4. Logs all activity

Options:
- `--dry-run`: Preview mode - shows what would be searched without actually searching

Example dry-run:
```bash
janitarr run --dry-run
```

**When to Use Manual Runs**:
- Testing your configuration
- Running automation on-demand (outside schedule)
- After adding new content to your library

#### Start Services

```bash
janitarr start
```

Starts both scheduler and web server in a single process:
- Runs automation on configured interval
- Serves web interface at `http://localhost:3434`
- Runs first cycle immediately on startup
- Graceful shutdown on Ctrl+C

**Options:**
```bash
janitarr start --port 8080              # Custom port
janitarr start --host 0.0.0.0           # Bind to all interfaces
janitarr start --port 3000 --host 0.0.0.0  # Both options
```

#### Start in Development Mode

```bash
janitarr dev
```

Starts services in development mode:
- Verbose logging for all HTTP requests
- Proxies non-API requests to Vite dev server (`http://localhost:5173`)
- Detailed stack traces in API errors
- Automation cycle logging with timestamps
- Same `--port` and `--host` options as production mode

**Development workflow:**
```bash
# Terminal 1: Start backend services
bun run src/index.ts dev

# Terminal 2: Start Vite dev server
cd ui && bun run dev
```

#### Stop Services

```bash
janitarr stop
```

Stops the running services:
- Current cycle completes before stopping (with timeout)
- WebSocket connections closed gracefully
- No more automated cycles until restarted
- Manual runs still work

#### Check Service Status

```bash
janitarr status
```

Shows whether services are running and when next cycle will execute.

### Configuration

#### Display Configuration

```bash
janitarr config show
```

Shows all current settings:
- Schedule interval and enabled state
- All four search limits
- Database path

Options:
- `--json`: Output in JSON format

#### Set Configuration Values

```bash
janitarr config set <key> <value>
```

**Schedule Settings**:
```bash
janitarr config set schedule.interval 6      # hours between cycles
janitarr config set schedule.enabled true    # enable/disable scheduler
```

**Search Limits**:
```bash
janitarr config set limits.missing.movies 15    # missing Radarr searches
janitarr config set limits.missing.episodes 20  # missing Sonarr searches
janitarr config set limits.cutoff.movies 5      # Radarr upgrade searches
janitarr config set limits.cutoff.episodes 10   # Sonarr upgrade searches
```

**Configuration Keys**:
- `schedule.interval` - Hours between cycles (min: 1, default: 6)
- `schedule.enabled` - Enable/disable automation (true/false)
- `limits.missing.movies` - Max Radarr missing searches per cycle
- `limits.missing.episodes` - Max Sonarr missing searches per cycle
- `limits.cutoff.movies` - Max Radarr upgrade searches per cycle
- `limits.cutoff.episodes` - Max Sonarr upgrade searches per cycle

### Activity Logs

#### View Logs

```bash
janitarr logs
```

Shows recent activity (default: 50 entries).

Options:
- `--all`: Display all logs (paginated)
- `--limit N`: Show only N most recent entries
- `--json`: Output in JSON format for scripting

**Log Types**:
- `automation`: Cycle start/end with summary
- `search`: Individual search triggered
- `server-test`: Connection test result
- `error`: Failures and errors

#### Clear Logs

```bash
janitarr logs --clear
```

Permanently deletes all logs. Requires confirmation.

**Note**: Logs older than 30 days are automatically purged.

---

## Configuration

### Configuration Storage

All configuration is stored in a SQLite database at `./data/janitarr.db`.

**Override the database location**:
```bash
export JANITARR_DB_PATH=/path/to/custom.db
```

### Schedule Configuration

**Interval**: Time between automation cycles
- Minimum: 1 hour
- Default: 6 hours
- Recommended: 4-8 hours (balances freshness with indexer limits)

**Enabled**: Master switch for automation
- `true`: Scheduler runs on interval
- `false`: Only manual runs work

### Search Limits

Janitarr uses **four independent limits** to control search volume:

| Limit | Applies To | Default |
|-------|-----------|---------|
| `limits.missing.movies` | Radarr missing movies | 10 |
| `limits.missing.episodes` | Sonarr missing episodes | 10 |
| `limits.cutoff.movies` | Radarr quality upgrades | 5 |
| `limits.cutoff.episodes` | Sonarr quality upgrades | 5 |

**How Limits Work**:

1. Detection runs on all servers
2. Results are categorized by type (missing movies, missing episodes, etc.)
3. Each category is limited independently
4. Within each category, searches are distributed fairly across servers using round-robin

**Example**:

You have 2 Radarr servers and 1 Sonarr server:
- Radarr 1: 50 missing movies, 10 upgrades available
- Radarr 2: 30 missing movies, 15 upgrades available
- Sonarr 1: 100 missing episodes, 20 upgrades available

With default limits (10/10/5/5):
- **Missing movies** (limit: 10): 5 from Radarr 1, 5 from Radarr 2 (round-robin)
- **Missing episodes** (limit: 10): All 10 from Sonarr 1
- **Movie upgrades** (limit: 5): 3 from Radarr 1, 2 from Radarr 2 (round-robin)
- **Episode upgrades** (limit: 5): All 5 from Sonarr 1

Total: 30 searches triggered across 3 servers.

**Choosing Limits**:

Consider your indexer's daily limits. Common indexer limits:
- Free tier: 100-500 API hits/day
- Paid tier: 1,000-10,000 API hits/day

Each search = 1 API hit to Radarr/Sonarr + multiple hits to indexers.

**Recommended limits** based on indexer tier:
- **Free tier** (100 hits/day): 5/5/3/3 per cycle, 6-hour interval = ~64 searches/day
- **Mid tier** (500 hits/day): 10/10/5/5 per cycle, 6-hour interval = ~120 searches/day
- **High tier** (1000+ hits/day): 20/20/10/10 per cycle, 4-hour interval = ~360 searches/day

### Server Configuration

**Required fields**:
- **Name**: Unique identifier (any string)
- **Type**: `radarr` or `sonarr` (cannot be changed after creation)
- **URL**: Full URL including protocol (normalized automatically)
- **API Key**: Found in server Settings → General → Security

**Optional fields**:
- **Enabled**: Whether to include in automation (default: true)

**Security**:
- API keys are encrypted at rest using AES-256-GCM
- Keys are only decrypted when making API calls
- CLI displays keys masked (e.g., `r4nd0m...xyz`)

**URL Normalization**:
Janitarr automatically normalizes URLs:
- Adds `http://` if protocol missing
- Removes trailing slashes
- Validates hostname format

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `JANITARR_DB_PATH` | SQLite database location | `./data/janitarr.db` |
| `JANITARR_LOG_LEVEL` | Logging verbosity | `info` |

---

## Workflows

### Daily Operations

**Typical daily workflow** (fully automated):

1. Scheduler wakes up every 6 hours
2. Automation cycle runs:
   - Queries all servers for missing/upgrade content
   - Applies search limits
   - Triggers searches fairly across servers
   - Logs all activity
3. You review logs occasionally to monitor progress

**No manual intervention needed!**

### Adding a New Server

**Workflow**:

1. Add server via Web UI or CLI
2. Test connection to verify configuration
3. Run a manual dry-run to preview behavior:
   ```bash
   janitarr run --dry-run
   ```
4. If dry-run looks good, run actual cycle:
   ```bash
   janitarr run
   ```
5. Monitor logs for first few cycles

### Adjusting Search Limits

**Scenario**: Your indexer sent a warning about approaching API limits.

**Workflow**:

1. Check your current configuration:
   ```bash
   janitarr config show
   ```

2. Review recent search volume:
   ```bash
   janitarr logs | grep "Triggered searches"
   ```

3. Reduce limits:
   ```bash
   janitarr config set limits.missing.movies 5
   janitarr config set limits.missing.episodes 5
   janitarr config set limits.cutoff.movies 3
   janitarr config set limits.cutoff.episodes 3
   ```

4. Consider increasing interval:
   ```bash
   janitarr config set schedule.interval 8
   ```

### Temporary Disable

**Scenario**: You're doing maintenance and want to pause automation.

**Option 1: Disable specific server**:
- Web UI: Go to Servers, toggle the server's enabled status
- CLI: Edit server and set enabled=false

**Option 2: Stop scheduler**:
```bash
janitarr stop
```

**Option 3: Disable automation entirely**:
```bash
janitarr config set schedule.enabled false
```

**To re-enable**:
```bash
janitarr start
janitarr config set schedule.enabled true
```

### Investigating Issues

**Scenario**: Searches aren't triggering as expected.

**Troubleshooting workflow**:

1. Check scheduler status:
   ```bash
   janitarr status
   ```

2. Run a scan to see what would be detected:
   ```bash
   janitarr scan
   ```

3. Review recent logs for errors:
   ```bash
   janitarr logs --limit 100
   ```

4. Test each server connection:
   ```bash
   janitarr server test "Server Name"
   ```

5. Run a dry-run to preview:
   ```bash
   janitarr run --dry-run
   ```

6. Check configuration:
   ```bash
   janitarr config show
   ```

See [Troubleshooting Guide](troubleshooting.md) for detailed issue resolution.

---

## Best Practices

### Search Limit Strategy

**Start conservative, increase gradually**:
1. Begin with default limits (10/10/5/5)
2. Monitor indexer usage for a week
3. If under limits, increase by 5 per category
4. If approaching limits, decrease by 5 per category

**Prioritize missing content over upgrades**:
- Missing content: higher limits (10-20)
- Quality upgrades: lower limits (5-10)

**Consider server balance**:
- More Radarr servers? Allocate more to movie limits
- More Sonarr servers? Allocate more to episode limits

### Schedule Interval

**Recommended intervals**:
- **4 hours**: High-volume libraries, generous indexer limits
- **6 hours** (default): Balanced approach, works for most users
- **8-12 hours**: Conservative, limited indexer API quotas

**Avoid**:
- Less than 2 hours (excessive API usage)
- More than 24 hours (content stays missing too long)

### Server Organization

**Naming conventions**:
- Include purpose: "4K Movies", "Anime Shows", "Kids Content"
- Include location if multiple instances: "Radarr Local", "Radarr VPS"
- Be descriptive: "Main Radarr" is better than "Radarr1"

**Disable vs Delete**:
- **Disable** when temporarily excluding (e.g., server maintenance)
- **Delete** when permanently removing (e.g., decommissioned server)

### Log Management

**Regular review**:
- Check logs weekly for patterns
- Look for repeated errors
- Verify searches are distributed fairly

**Retention**:
- 30-day automatic retention is usually sufficient
- Export logs before clearing if you need historical data
- Use `--json` format for analysis with external tools

### Performance Tips

**Database maintenance**:
- Database is automatically optimized (no manual VACUUM needed)
- Keep database on fast storage (SSD preferred)
- If database grows large (>100MB), consider clearing old logs

**Network**:
- Ensure low-latency connection to servers
- Use local network addresses when possible
- Avoid VPN if servers are on same network

**Resource usage**:
- Janitarr is lightweight (minimal CPU/memory)
- Most time spent waiting for API responses
- No significant I/O except logging

### Security

**API Key Protection**:
- Never share your Janitarr database (contains encrypted keys)
- Don't commit `.env` files or database to version control
- API keys are encrypted at rest but decrypted in memory

**Network Security**:
- Use HTTPS URLs for remote servers when possible
- Consider firewall rules to restrict Janitarr to local network
- Web UI has no authentication (run on trusted network only)

**Access Control**:
- Janitarr has full control over your media servers
- Only give access to trusted users
- Monitor logs for unexpected activity

### Monitoring

**Health checks**:
- Review Dashboard daily (quick glance)
- Check logs weekly (audit trail)
- Test server connections monthly (catch configuration drift)

**Success metrics**:
- Missing content count decreasing over time
- Quality upgrades completing gradually
- No repeated search failures in logs

**Warning signs**:
- Server connection errors
- Search failures (may indicate indexer issues)
- Large missing/upgrade counts (may need higher limits)

---

## Next Steps

- [Troubleshooting Guide](troubleshooting.md) - Solve common issues
- [API Reference](api-reference.md) - Integrate with Janitarr's API
- [Development Guide](development.md) - Contribute to Janitarr

Need help? Check the [GitHub Issues](https://github.com/yourusername/janitarr/issues) or open a new issue.
