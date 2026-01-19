# Janitarr: Implementation Plan

## Overview

This document tracks implementation tasks for Janitarr, an automation tool for Radarr and Sonarr media servers written in Go.

## Agent Instructions

This document is designed for AI coding agents. Each task:

- Has a checkbox `[ ]` that should be marked `[x]` when complete
- Includes specific file paths and commands to execute
- Has clear completion criteria
- References specification documents in `specs/`

**Before starting each phase:**

1. Read the relevant specification documents
2. Write tests first (TDD approach)
3. Run `go test ./...` after each implementation
4. Commit working code before moving to the next task

**Environment:** Development tools are provided by devenv. Run `direnv allow` to load.

## Technology Stack

| Component       | Technology          | Purpose                      |
| --------------- | ------------------- | ---------------------------- |
| Language        | Go 1.22+            | Main application             |
| Web Framework   | Chi (go-chi/chi/v5) | HTTP routing                 |
| Database        | modernc.org/sqlite  | SQLite (pure Go, no CGO)     |
| CLI             | Cobra (spf13/cobra) | Command-line interface       |
| CLI Forms       | charmbracelet/huh   | Interactive terminal forms   |
| Console Logging | charmbracelet/log   | Colorized structured logging |
| Templates       | templ (a-h/templ)   | Type-safe HTML templates     |
| Interactivity   | htmx + Alpine.js    | Dynamic UI without React     |
| CSS             | Tailwind CSS        | Utility-first styling        |

---

## Gap Analysis Summary

### Implemented (Phases 0-9)

The following core functionality is complete:

- **Foundation**: Database, crypto, CLI skeleton
- **Core Services**: API clients, server manager, detector, search trigger
- **Scheduler & Automation**: Scheduler, automation orchestrator, basic activity logger
- **CLI Commands**: Server/config/automation/log commands (flag-based only)
- **Web Server & API**: HTTP server, middleware, API handlers, WebSocket log streaming
- **Frontend**: Templates, components, pages with HTMX partial updates
- **Integration**: Start/dev commands, graceful shutdown

### Gaps Identified

| Spec                  | Gap                                                         | Priority |
| --------------------- | ----------------------------------------------------------- | -------- |
| `logging.md`          | charmbracelet/log not integrated (uses fmt.Printf)          | High     |
| `logging.md`          | No log levels (debug/info/warn/error)                       | High     |
| `logging.md`          | No --log-level CLI flag                                     | High     |
| `logging.md`          | Search logs missing title/year/quality metadata             | Medium   |
| `logging.md`          | Web filters missing: date range, operation type             | Medium   |
| `logging.md`          | Log retention not implemented (constant defined but unused) | Medium   |
| `activity-logging.md` | No `operation` or `metadata` columns in logs table          | Medium   |
| `cli-interface.md`    | charmbracelet/huh not integrated                            | Medium   |
| `cli-interface.md`    | No interactive forms (uses basic bufio prompts)             | Medium   |
| `cli-interface.md`    | No server selector for edit/delete/test                     | Medium   |
| `cli-interface.md`    | No --non-interactive flag                                   | Low      |
| `services/types.go`   | TriggerResult missing Title, Year, QualityProfile fields    | Medium   |

---

## Phase 10: Enhanced Logging System

**Reference:** `specs/logging.md`, `specs/activity-logging.md`
**Verification:** `go test ./src/logger/... && go test ./src/services/...`

### 10.1 Add charmbracelet/log Dependency

- [x] Add dependency: `go get github.com/charmbracelet/log`
- [x] Verify import works in a test file

### 10.2 Create Log Level Support

**Reference:** `specs/logging.md` (Configure Log Levels section)

- [x] Create `src/logger/level.go`:
  - [x] Define `Level` type with constants: `LevelDebug`, `LevelInfo`, `LevelWarn`, `LevelError`
  - [x] Add `ParseLevel(s string) (Level, error)` function
  - [x] Add `String()` method for Level type

### 10.3 Create Console Logger

**Reference:** `specs/logging.md` (Console Logging with charmbracelet/log section)

- [x] Create `src/logger/console.go`:
  - [x] `type ConsoleLogger struct` wrapping `*log.Logger` from charmbracelet/log
  - [x] `NewConsoleLogger(level Level, isDev bool) *ConsoleLogger`
  - [x] Configure output: stdout for dev, stderr for production
  - [x] Configure timestamp format: `15:04:05`
  - [x] Configure colors: debug=gray, info=blue, warn=yellow, error=red
  - [x] Methods: `Debug()`, `Info()`, `Warn()`, `Error()` with structured key-value args

### 10.4 Update Logger to Use Console Logger

**Reference:** `specs/logging.md` (Unified Logging Backend section)

Current state: Logger at `src/logger/logger.go:11-15` only has `storer`, `mu`, `subscribers`.

- [x] Update `src/logger/logger.go`:
  - [x] Add `console *ConsoleLogger` field to Logger struct
  - [x] Add `level Level` field to Logger struct
  - [x] Update constructor: `NewLogger(storer LogStorer, level Level, isDev bool) *Logger`
  - [x] Update all log methods to:
    1. Check level before logging (delegated to console logger)
    2. Write to console via charmbracelet/log
    3. Write to database (existing behavior)
    4. Broadcast to WebSocket subscribers (existing behavior)

- [x] Update all call sites of `NewLogger()`:
  - [x] `src/cli/start.go:58` - pass level and isDev
  - [x] `src/cli/dev.go` - pass debug level
  - [x] Test files using mock logger

### 10.5 Add Log Level CLI Flags

**Reference:** `specs/logging.md` (Configure Log Levels section)

- [ ] Update `src/cli/root.go`:
  - [ ] Add `--log-level` persistent flag (default: "info")
  - [ ] Parse `JANITARR_LOG_LEVEL` environment variable
  - [ ] CLI flag takes precedence over env var
  - [ ] Validate log level, exit with error if invalid

- [ ] Update `src/cli/start.go`:
  - [ ] Pass log level from flag to logger initialization
  - [ ] Default to "info" in production mode

- [ ] Update `src/cli/dev.go`:
  - [ ] Default to "debug" in dev mode
  - [ ] Allow override via `--log-level` flag

### 10.6 Add Detection Summary Logging

**Reference:** `specs/logging.md` (Log Automation Cycle Summary section)

- [ ] Add to `src/logger/logger.go`:
  - [ ] `LogDetectionComplete(serverName, serverType string, missing, cutoffUnmet int)`
  - [ ] Console format: `INFO Detection complete server=X missing=Y cutoff_unmet=Z`

- [ ] Update `src/services/automation.go`:
  - [ ] After detection, call `logger.LogDetectionComplete()` for each server result

### 10.7 Add Detailed Search Logging with Metadata

**Reference:** `specs/logging.md` (Log Individual Search Triggers), `specs/activity-logging.md`

Current state: `TriggerResult` at `src/services/types.go:63-72` only has ServerID, ServerName, ServerType, Category, ItemIDs, Success, Error.

- [ ] Update `src/services/types.go`:
  - [ ] Add `Title string` field to TriggerResult
  - [ ] Add `Year int` field to TriggerResult (for movies)
  - [ ] Add `SeriesTitle string` field (for episodes)
  - [ ] Add `SeasonNumber int` field
  - [ ] Add `EpisodeNumber int` field
  - [ ] Add `QualityProfile string` field

- [ ] Update `src/api/types.go` (if needed):
  - [ ] Ensure `Movie` has `Year`, `QualityProfileId` fields
  - [ ] Ensure `Episode` has quality profile access

- [ ] Update `src/services/search_trigger.go`:
  - [ ] When building TriggerResult, populate metadata fields from MediaItem

- [ ] Update `src/logger/logger.go`:
  - [ ] Add `LogMovieSearch(server, title string, year int, quality, category string)`
  - [ ] Add `LogEpisodeSearch(server, series, episodeTitle string, season, episode int, quality, category string)`
  - [ ] Console format per spec:
    - Movies: `INFO Search triggered title="X" year=Y quality="Z" server=S category=C`
    - Episodes: `INFO Search triggered series="X" episode="S01E02" title="Y" quality="Z" server=S category=C`

### 10.8 Add Development Mode Verbose Logging

**Reference:** `specs/logging.md` (Development Mode Verbose Logging section)

Current state: `src/web/middleware/logging.go:17` has placeholder comment "will integrate with actual logger later".

- [ ] Update `src/web/middleware/logging.go`:
  - [ ] Accept logger in constructor
  - [ ] Log HTTP requests at debug level: `DEBUG HTTP request method=GET path=/api/servers status=200 duration=12ms`

- [ ] Update `src/api/client.go`:
  - [ ] Add optional `logger` field
  - [ ] Add `WithLogger(l *Logger) *Client` method
  - [ ] Log API requests in debug mode (without API keys)

- [ ] Update `src/services/scheduler.go`:
  - [ ] Add logger field
  - [ ] Log scheduler events: `DEBUG Scheduler sleeping until=X`

- [ ] Update `src/web/websocket/hub.go`:
  - [ ] Log WebSocket connections at debug level

### 10.9 Update Database Schema for Enhanced Logs

**Reference:** `specs/logging.md` (Web Log Entry Format section)

Current schema at `src/database/migrations/001_initial_schema.sql:17-27` lacks `operation` and `metadata` columns.

- [ ] Create `src/database/migrations/002_enhanced_logs.sql`:

  ```sql
  ALTER TABLE logs ADD COLUMN operation TEXT;
  ALTER TABLE logs ADD COLUMN metadata TEXT;
  CREATE INDEX IF NOT EXISTS idx_logs_operation ON logs(operation);
  ```

- [ ] Update `src/database/database.go`:
  - [ ] Add migration 002 to embedded migrations

- [ ] Update `src/database/logs.go`:
  - [ ] Add `operation` and `metadata` to insert/select queries
  - [ ] Add filter method: `GetLogsByOperation(operation string)`

- [ ] Update `src/logger/types.go`:
  - [ ] Add `Operation string` field to LogEntry
  - [ ] Add `Metadata map[string]interface{}` field to LogEntry

### 10.10 Web Log Viewer Filter Enhancements

**Reference:** `specs/logging.md` (Filter Logs section)

Current state: `src/web/handlers/api/logs.go` only supports `type` and `server` filters.

- [ ] Update `src/web/handlers/api/logs.go`:
  - [ ] Add `level` query param filter
  - [ ] Add `operation` query param filter
  - [ ] Add `from` datetime query param filter
  - [ ] Add `to` datetime query param filter
  - [ ] Return total count in response for pagination

- [ ] Update `src/database/logs.go`:
  - [ ] Add date range filtering to GetLogs
  - [ ] Add operation filtering

- [ ] Update `src/templates/pages/logs.templ`:
  - [ ] Add date range pickers (from/to)
  - [ ] Add operation type dropdown
  - [ ] Add "Clear filters" button
  - [ ] Sync filter state to URL query params

### 10.11 Dashboard Log Summary Widget

**Reference:** `specs/logging.md` (Dashboard Log Summary section)

Current state: Dashboard at `src/templates/pages/dashboard.templ:116-159` shows recent activity but no error count badge.

- [ ] Update `src/templates/pages/dashboard.templ`:
  - [ ] Add 24-hour error count badge to recent activity section
  - [ ] Add "View all logs" link

- [ ] Update `src/web/handlers/pages/dashboard.go`:
  - [ ] Add 24-hour error count to dashboard data

- [ ] Update `src/database/logs.go`:
  - [ ] Add `GetErrorCount(since time.Time) (int, error)` method

### 10.12 Implement Log Retention

**Reference:** `specs/logging.md` (Log Retention and Cleanup section)

Current state: `LogRetentionDays = 30` constant exists at `src/database/database.go:19` but is unused.

- [ ] Update `src/database/logs.go`:
  - [ ] Add `PurgeOldLogs(retentionDays int) (int, error)` method
  - [ ] Add `GetLogCount() (int, error)` method

- [ ] Update `src/database/config.go`:
  - [ ] Add `logs.retention_days` config key (default: 30, range: 7-90)

- [ ] Create `src/services/maintenance.go`:
  - [ ] `RunLogCleanup(db *database.DB)` function
  - [ ] Delete logs older than configured retention

- [ ] Update `src/services/scheduler.go`:
  - [ ] Run log cleanup daily (at midnight or on first cycle of day)

- [ ] Update `src/templates/pages/settings.templ`:
  - [ ] Add log retention setting dropdown (7, 14, 30, 60, 90 days)
  - [ ] Display current log count

- [ ] Update `src/web/handlers/api/config.go`:
  - [ ] Handle `logs.retention_days` in PATCH endpoint

### 10.13 Write Tests

- [ ] Create `src/logger/level_test.go`:
  - [ ] Test ParseLevel with valid/invalid inputs
  - [ ] Test Level.String()

- [ ] Create `src/logger/console_test.go`:
  - [ ] Test log level filtering

- [ ] Update `src/logger/logger_test.go`:
  - [ ] Test new constructor signature
  - [ ] Test level filtering

### 10.14 Verification

- [ ] Run unit tests: `go test ./src/logger/...`
- [ ] Run integration tests: `go test ./src/services/...`
- [ ] Manual testing:
  - [ ] `./janitarr dev` shows debug logs with colors
  - [ ] `./janitarr start` shows info logs only
  - [ ] `./janitarr start --log-level debug` shows debug logs
  - [ ] Web UI logs page shows all filters
  - [ ] Dashboard shows error count badge

---

## Phase 11: Interactive CLI Forms

**Reference:** `specs/cli-interface.md`
**Verification:** `go test ./src/cli/... && go build ./src`

### 11.1 Add charmbracelet/huh Dependency

- [ ] Add dependency: `go get github.com/charmbracelet/huh`
- [ ] Verify import works

### 11.2 Create Forms Package Structure

- [ ] Create `src/cli/forms/` directory
- [ ] Create `src/cli/forms/helpers.go`:
  - [ ] `IsInteractive() bool` - check if stdin is a TTY using `golang.org/x/term`
  - [ ] Common validation functions:
    - [ ] `ValidateServerName(s string) error`
    - [ ] `ValidateURL(s string) error`
    - [ ] `ValidateAPIKey(s string) error`
    - [ ] `ValidateServerType(s string) error`

### 11.3 Server Add Form

**Reference:** `specs/cli-interface.md` (Interactive Server Addition section)

Current state: `src/cli/server.go:57-124` uses `bufio.NewReader` with manual prompts.

- [ ] Create `src/cli/forms/server.go`:
  - [ ] `ServerAddForm() (*ServerFormResult, error)`:
    - [ ] Select field for server type (Radarr/Sonarr)
    - [ ] Input field for name with validation
    - [ ] Input field for URL with validation
    - [ ] Input field for API key with `EchoMode(huh.EchoModePassword)`
  - [ ] Return nil on Escape/cancel

- [ ] Update `src/cli/server.go` (`runServerAdd`):
  - [ ] Check `forms.IsInteractive()`
  - [ ] If interactive and no flags provided, call `forms.ServerAddForm()`
  - [ ] If non-interactive or all flags provided, use existing flag-based logic
  - [ ] Show spinner during connection test

### 11.4 Server Edit Form

**Reference:** `specs/cli-interface.md` (Interactive Server Editing section)

Current state: `src/cli/server.go:151-227` uses basic prompts.

- [ ] Add to `src/cli/forms/server.go`:
  - [ ] `ServerEditForm(current *ServerFormResult) (*ServerFormResult, error)`:
    - [ ] Pre-populate fields with current values
    - [ ] Server type displayed but disabled
    - [ ] "Keep existing API key" option
    - [ ] Return only changed fields

- [ ] Update `src/cli/server.go` (`runServerEdit`):
  - [ ] If interactive and only server name provided, show edit form
  - [ ] Pre-populate form with existing server values

### 11.5 Server Selector

**Reference:** `specs/cli-interface.md` (Server List with Interactive Selection section)

- [ ] Add to `src/cli/forms/server.go`:
  - [ ] `ServerSelector(servers []ServerInfo) (*ServerInfo, error)`:
    - [ ] Use `huh.NewSelect()` with server list
    - [ ] Display: name, type, enabled status
    - [ ] Return selected server or nil on cancel

- [ ] Update `src/cli/server.go`:
  - [ ] `server edit` (no name arg): show selector, then edit form
  - [ ] `server remove` (no name arg): show selector, then confirmation
  - [ ] `server test` (no name arg): show selector, then test

### 11.6 Configuration Form

**Reference:** `specs/cli-interface.md` (Interactive Configuration section)

Current state: `src/cli/config.go` only has flag-based `config show` and `config set`.

- [ ] Create `src/cli/forms/config.go`:
  - [ ] `ConfigForm(current *AppConfig) (*AppConfig, error)`:
    - [ ] Group: Automation (enabled toggle, interval number, dry-run toggle)
    - [ ] Group: Search Limits (4 number inputs)
    - [ ] Pre-populate with current values

- [ ] Update `src/cli/config.go`:
  - [ ] Add `config` command (no subcommand) that launches form when interactive
  - [ ] Keep `config show` and `config set` for non-interactive use

### 11.7 Confirmation Dialogs

**Reference:** `specs/cli-interface.md` (Confirmation Dialogs section)

Current state: `src/cli/server.go:252-257` uses basic Y/N prompt.

- [ ] Create `src/cli/forms/confirm.go`:
  - [ ] `ConfirmDelete(itemType, itemName string) (bool, error)`:
    - [ ] Show item details
    - [ ] Require typing item name to confirm
  - [ ] `ConfirmAction(message string) (bool, error)`:
    - [ ] Simple yes/no confirmation

- [ ] Update `src/cli/server.go` (`runServerRemove`):
  - [ ] If interactive and no `--force`, show `ConfirmDelete`

- [ ] Update `src/cli/logs.go`:
  - [ ] If interactive and clearing logs, show `ConfirmAction` with log count

### 11.8 Non-Interactive Mode Flag

**Reference:** `specs/cli-interface.md` (Flag Override section)

- [ ] Update `src/cli/root.go`:
  - [ ] Add `--non-interactive` global flag
  - [ ] When set, skip all interactive forms and require flags

- [ ] Update all form-using commands:
  - [ ] Check `--non-interactive` flag
  - [ ] Error with usage if required flags missing

### 11.9 Write Tests

- [ ] Create `src/cli/forms/helpers_test.go`:
  - [ ] Test validation functions

- [ ] Create `src/cli/forms/server_test.go`:
  - [ ] Test form field configurations (mock form execution)

### 11.10 Verification

- [ ] Run tests: `go test ./src/cli/...`
- [ ] Manual testing:
  - [ ] `./janitarr server add` - interactive form works
  - [ ] `./janitarr server add --name X --type radarr --url Y --api-key Z` - flags work
  - [ ] `./janitarr server edit` - shows selector then form
  - [ ] `./janitarr server remove` - shows selector then confirmation
  - [ ] `./janitarr config` - shows configuration form
  - [ ] `./janitarr server add --non-interactive` - errors if flags missing
  - [ ] Piped input detects non-TTY correctly

---

## Completed Phases Summary

The following phases have been completed in prior work:

- **Phase 0:** Setup - Directory structure, Go module, tooling
- **Phase 1:** Foundation - Crypto module, database module, CLI skeleton
- **Phase 2:** Core Services - API client, server manager, detector, search trigger
- **Phase 3:** Scheduler & Automation - Scheduler, automation orchestrator, activity logger
- **Phase 4:** CLI Commands - Formatters, server/config/automation/log commands
- **Phase 5:** Web Server & API - HTTP server, middleware, API handlers, WebSocket
- **Phase 6:** Frontend with templ - Templates, components, pages, static assets
- **Phase 7:** Integration & Polish - Start/dev commands, graceful shutdown, E2E tests
- **Phase 8:** Bug Fixes - Server connection test fix
- **Phase 9:** Test Suite Cleanup - Refactored tests, removed obsolete files

---

## Verification Checklist

### Phase 10: Enhanced Logging

- [ ] charmbracelet/log integrated for console output
- [ ] Log levels work: debug, info, warn, error
- [ ] `--log-level` flag and `JANITARR_LOG_LEVEL` env var work
- [ ] Dev mode shows debug logs to stdout with colors
- [ ] Production mode shows info logs to stderr
- [ ] Detection summaries logged per server
- [ ] Search triggers include title/year/quality metadata
- [ ] Database has operation and metadata columns
- [ ] Web UI log viewer has date range and operation filters
- [ ] Dashboard shows 24-hour error count
- [ ] Log retention is configurable and runs automatically

### Phase 11: Interactive CLI Forms

- [ ] charmbracelet/huh integrated
- [ ] `server add` has interactive form with masked API key input
- [ ] `server edit` has form with pre-populated values
- [ ] `server remove` has confirmation requiring name match
- [ ] `config` has interactive form for all settings
- [ ] Server selector works for edit/remove/test
- [ ] `--non-interactive` flag works
- [ ] Non-TTY input detected and handled

---

## Notes

### Dependencies to Add

```bash
go get github.com/charmbracelet/log
go get github.com/charmbracelet/huh
go get golang.org/x/term  # for IsTerminal check
```

### Breaking Changes

- Logger constructor signature: `NewLogger(storer, level, isDev)` instead of `NewLogger(storer)`
- Database schema migration 002 adds columns
- New config key: `logs.retention_days`

### Performance Considerations

- Console logging should not block main operations
- Database log writes could be batched/async for high-volume scenarios
- WebSocket broadcasts use non-blocking channel sends
