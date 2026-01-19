# Logging: Unified Console and Web Logging System

## Context

Janitarr requires comprehensive logging for operational visibility, debugging, and troubleshooting. The logging system serves multiple audiences:

1. **Operators** need to monitor automation activity and verify the system is working
2. **Developers** need detailed logs to diagnose issues during development
3. **Web users** need real-time visibility into system operations through the web interface

This specification defines a unified logging system that outputs to multiple destinations (console, web interface, database) with consistent formatting and appropriate detail levels for each context.

## Technology Stack

- **Console logging**: [charmbracelet/log](https://github.com/charmbracelet/log) for colorized, structured terminal output
- **Web streaming**: WebSocket for real-time log delivery to browser
- **Persistence**: SQLite database for log history and audit trail

## Requirements

### Story: Configure Log Levels

- **As a** user
- **I want to** control the verbosity of log output
- **So that** I can see appropriate detail for my current needs

#### Acceptance Criteria

- [ ] System supports four log levels: `debug`, `info`, `warn`, `error`
- [ ] Default log level is `info` in production mode (`janitarr start`)
- [ ] Default log level is `debug` in development mode (`janitarr dev`)
- [ ] Log level configurable via `--log-level` CLI flag (e.g., `janitarr start --log-level debug`)
- [ ] Log level configurable via `JANITARR_LOG_LEVEL` environment variable
- [ ] CLI flag takes precedence over environment variable
- [ ] Invalid log level displays error and exits

### Story: Console Logging with charmbracelet/log

- **As a** user running Janitarr from the terminal
- **I want to** see colorized, well-formatted log output
- **So that** I can quickly scan logs and identify important information

#### Acceptance Criteria

- [ ] Logs use charmbracelet/log for terminal output
- [ ] Log output uses structured key-value format for machine parseability
- [ ] Colors distinguish log levels: debug (gray), info (blue), warn (yellow), error (red)
- [ ] Timestamps displayed in local time with readable format (e.g., `15:04:05`)
- [ ] Long values (URLs, paths) are truncated in terminal output with `...`
- [ ] Production mode (`start`) writes logs to stderr
- [ ] Development mode (`dev`) writes logs to stdout

### Story: Log Automation Cycle Summary

- **As a** user
- **I want to** see summary statistics at the start and end of each automation cycle
- **So that** I can quickly understand what the system found and did

#### Acceptance Criteria

- [ ] Cycle start logs summary of detected items per server:
  ```
  INFO Automation cycle started trigger=scheduled
  INFO Detection complete server=radarr-main missing=5 cutoff_unmet=12
  INFO Detection complete server=sonarr-main missing=23 cutoff_unmet=8
  ```
- [ ] Cycle end logs summary of actions taken:
  ```
  INFO Automation cycle completed duration=45s searches_triggered=15 failures=0
  ```
- [ ] Manual triggers clearly identified: `trigger=manual` vs `trigger=scheduled`
- [ ] Failed cycles log error summary with failure count and reasons

### Story: Log Individual Search Triggers

- **As a** user
- **I want to** see detailed information when searches are triggered
- **So that** I can verify the correct content is being searched

#### Acceptance Criteria

- [ ] Movie searches log title, year, quality profile, and server:
  ```
  INFO Search triggered title="The Matrix" year=1999 quality="HD-1080p" server=radarr-main category=missing
  ```
- [ ] Episode searches log series, season/episode, episode title, quality profile, and server:
  ```
  INFO Search triggered series="Breaking Bad" episode="S01E01" title="Pilot" quality="HD-1080p" server=sonarr-main category=cutoff_unmet
  ```
- [ ] Failed searches logged at `error` level with reason:
  ```
  ERROR Search failed title="The Matrix" server=radarr-main error="API timeout"
  ```

### Story: Development Mode Verbose Logging

- **As a** developer
- **I want to** see all internal operations in development mode
- **So that** I can diagnose issues and verify the application is working correctly

#### Acceptance Criteria

- [ ] Development mode logs all events visible in web interface
- [ ] HTTP requests to web server logged with method, path, status, duration:
  ```
  DEBUG HTTP request method=GET path=/api/servers status=200 duration=12ms
  ```
- [ ] HTTP requests to Radarr/Sonarr APIs logged (without API keys):
  ```
  DEBUG API request server=radarr-main endpoint=/api/v3/movie status=200 duration=234ms
  ```
- [ ] Scheduler events logged (tick, sleep, wake):
  ```
  DEBUG Scheduler sleeping until=2026-01-17T12:00:00Z
  DEBUG Scheduler woke up reason=timer
  ```
- [ ] Database operations logged at debug level (query type, table, duration)
- [ ] WebSocket connections logged (connect, disconnect, message count)

### Story: Web Interface Log Viewer Page

- **As a** user
- **I want to** view logs in a dedicated web page
- **So that** I can review system activity without terminal access

#### Acceptance Criteria

- [ ] Dedicated `/logs` page displays log history
- [ ] Logs displayed in reverse chronological order (newest first)
- [ ] Each log entry shows: timestamp, level, message, and key-value metadata
- [ ] Log levels visually distinguished (color coding or icons)
- [ ] Error-level logs prominently highlighted
- [ ] Page loads last 100 entries initially
- [ ] "Load more" button or infinite scroll for older entries
- [ ] Maximum 1000 entries loadable in single session (pagination for older)

### Story: Real-time Log Streaming

- **As a** user viewing the logs page
- **I want to** see new logs appear in real-time
- **So that** I can monitor active operations without refreshing

#### Acceptance Criteria

- [ ] WebSocket connection streams new log entries to browser
- [ ] New entries animate into view at top of log list
- [ ] Visual indicator shows streaming is active (e.g., "Live" badge)
- [ ] Auto-scroll pauses when user scrolls up to read history
- [ ] Auto-scroll resumes when user scrolls back to top
- [ ] Connection automatically reconnects on disconnect (with backoff)
- [ ] Graceful degradation: page still works if WebSocket unavailable (manual refresh)

### Story: Dashboard Log Summary

- **As a** user viewing the dashboard
- **I want to** see recent log activity at a glance
- **So that** I can quickly assess system health without navigating to logs page

#### Acceptance Criteria

- [ ] Dashboard displays compact log summary widget
- [ ] Widget shows last 5-10 log entries
- [ ] Recent errors prominently displayed (last 24 hours error count)
- [ ] "View all logs" link navigates to full `/logs` page
- [ ] Widget updates in real-time via same WebSocket as logs page

### Story: Filter Logs

- **As a** user viewing logs (web interface)
- **I want to** filter logs by various criteria
- **So that** I can find specific information quickly

#### Acceptance Criteria

- [ ] Filter by log level (show debug/info/warn/error, multi-select)
- [ ] Filter by server (dropdown of configured servers, multi-select)
- [ ] Filter by operation type: search, automation_cycle, connection, system
- [ ] Filter by date/time range (from/to datetime pickers)
- [ ] Filters combinable (AND logic)
- [ ] Active filters displayed as removable chips/tags
- [ ] "Clear all filters" button
- [ ] Filter state preserved in URL query parameters (shareable/bookmarkable)
- [ ] Filters apply to both historical logs and real-time stream

### Story: Log Retention and Cleanup

- **As a** user
- **I want to** logs to be automatically managed
- **So that** storage doesn't grow unbounded

#### Acceptance Criteria

- [ ] Logs retained for 30 days by default
- [ ] Retention period configurable via settings (7-90 days range)
- [ ] Automatic daily cleanup of logs older than retention period
- [ ] Manual "Clear all logs" option in web UI (with confirmation dialog)
- [ ] Cleanup runs during low-activity periods (e.g., midnight local time)
- [ ] Log count displayed in settings for awareness

### Story: Unified Logging Backend

- **As a** developer
- **I want to** a single logging interface that outputs to all destinations
- **So that** logging is consistent and maintainable

#### Acceptance Criteria

- [ ] Single `Logger` interface used throughout codebase
- [ ] Logger writes to: console (charmbracelet/log), database (SQLite), WebSocket broadcast
- [ ] Log level filtering applied before output (don't store debug logs in production)
- [ ] Database writes are asynchronous (don't block main operations)
- [ ] WebSocket broadcast is non-blocking (slow clients don't affect logging)
- [ ] Structured log fields preserved across all outputs

## Edge Cases & Constraints

### Performance

- Logging must not significantly impact automation performance
- Database writes should be batched or async to prevent blocking
- WebSocket broadcasts use non-blocking channel sends (drop if buffer full)
- Console output buffered to prevent I/O blocking
- High-volume debug logging in dev mode may impact performance (acceptable trade-off)

### Log Storage

- Database log entries are lightweight (< 1KB per entry typical)
- Estimate: 1000 entries/day = ~1MB/day, 30 days = ~30MB
- Maximum practical limit: 100,000 entries before UI performance degrades
- Consider SQLite indexes on timestamp, level, server for filter performance

### Sensitive Data

- Never log API keys, passwords, or authentication tokens
- Sanitize URLs to remove query parameters that might contain secrets
- Log server names but not full connection URLs
- Mask sensitive configuration values in logs

### Console Output Format

Using charmbracelet/log structured format:

```
15:04:05 INFO  Automation cycle started trigger=scheduled
15:04:06 INFO  Detection complete server=radarr-main missing=5 cutoff_unmet=12
15:04:07 INFO  Search triggered title="The Matrix" year=1999 server=radarr-main
15:04:08 WARN  Rate limited server=radarr-main retry_after=30s
15:04:38 ERROR Search failed title="Inception" error="connection refused"
15:04:45 INFO  Automation cycle completed duration=40s searches=14 failures=1
```

Development mode adds more detail:

```
15:04:05 DEBUG Scheduler woke up reason=timer
15:04:05 INFO  Automation cycle started trigger=scheduled
15:04:05 DEBUG API request server=radarr-main endpoint=/api/v3/movie status=200 duration=234ms
15:04:05 DEBUG HTTP request method=GET path=/api/health status=200 duration=2ms
...
```

### Web Log Entry Format

Database schema for log entries:

```sql
CREATE TABLE logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    level TEXT NOT NULL,           -- debug, info, warn, error
    message TEXT NOT NULL,
    operation TEXT,                -- search, automation_cycle, connection, system
    server_name TEXT,              -- nullable, for server-specific logs
    metadata TEXT,                 -- JSON blob for structured key-value data
    INDEX idx_logs_timestamp (timestamp),
    INDEX idx_logs_level (level),
    INDEX idx_logs_server (server_name)
);
```

### WebSocket Protocol

Log streaming WebSocket at `/ws/logs`:

```json
// Server -> Client: New log entry
{
  "type": "log",
  "data": {
    "id": 12345,
    "timestamp": "2026-01-17T15:04:05Z",
    "level": "info",
    "message": "Search triggered",
    "operation": "search",
    "server_name": "radarr-main",
    "metadata": {
      "title": "The Matrix",
      "year": 1999,
      "quality": "HD-1080p",
      "category": "missing"
    }
  }
}

// Server -> Client: Connection established
{
  "type": "connected",
  "data": {
    "streaming": true
  }
}
```

### Error Handling

- If database write fails, log to console only (don't lose the log)
- If WebSocket broadcast fails, continue (clients can reconnect)
- If console output fails (rare), continue silently
- Logger initialization failure is fatal (application cannot start)

### Time Zones

- All timestamps stored in UTC in database
- Console output displays local time (user's system timezone)
- Web interface displays local time (browser timezone)
- API responses include ISO 8601 timestamps with timezone

### Concurrency

- Logger must be goroutine-safe (multiple goroutines logging simultaneously)
- Use sync.Mutex or channel-based design for thread safety
- WebSocket broadcast hub manages multiple client connections
- Database connection pool handles concurrent writes

## Implementation Notes

### Logger Interface

```go
// src/logger/logger.go
type Logger interface {
    Debug(msg string, keyvals ...interface{})
    Info(msg string, keyvals ...interface{})
    Warn(msg string, keyvals ...interface{})
    Error(msg string, keyvals ...interface{})

    // Convenience methods for common operations
    LogSearch(server, title string, metadata map[string]interface{})
    LogCycleStart(trigger string)
    LogCycleEnd(duration time.Duration, searches, failures int)
    LogDetectionComplete(server string, missing, cutoffUnmet int)
}

type UnifiedLogger struct {
    console    *log.Logger  // charmbracelet/log
    db         *database.DB
    wsHub      *websocket.Hub
    level      Level
    isDev      bool
}
```

### charmbracelet/log Setup

```go
import "github.com/charmbracelet/log"

func NewConsoleLogger(level Level, isDev bool) *log.Logger {
    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportTimestamp: true,
        TimeFormat:      "15:04:05",
        Level:           toCharmLevel(level),
    })

    if isDev {
        logger.SetOutput(os.Stdout)
    }

    return logger
}
```

### Integration Points

1. **Scheduler** (`src/services/scheduler.go`): Log cycle start/end, detection results
2. **Search Trigger** (`src/services/search_trigger.go`): Log individual searches
3. **Server Manager** (`src/services/server_manager.go`): Log connection tests
4. **API Clients** (`src/api/`): Log API requests in debug mode
5. **Web Server** (`src/web/server.go`): Log HTTP requests in dev mode
6. **WebSocket Handler** (`src/web/websocket/`): Broadcast logs to connected clients

## Success Metrics

1. **Visibility**: Users can see exactly what the automation is doing
2. **Debuggability**: Developers can diagnose issues using dev mode logs
3. **Performance**: Logging adds < 5% overhead to automation cycles
4. **Usability**: Web log viewer loads in < 500ms with 1000 entries
5. **Reliability**: No logs lost due to buffer overflow or connection issues

## Future Enhancements (Post-v1)

1. **Log export**: Download logs as CSV or JSON
2. **Log search**: Full-text search within log messages
3. **Custom log levels**: User-defined verbosity profiles
4. **External logging**: Forward logs to syslog, Loki, or other log aggregators
5. **Alerting**: Send notifications on error patterns
