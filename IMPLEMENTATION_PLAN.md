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

- [x] Update `src/cli/root.go`:
  - [x] Add `--log-level` persistent flag (default: "info")
  - [x] Parse `JANITARR_LOG_LEVEL` environment variable
  - [x] CLI flag takes precedence over env var
  - [x] Validate log level, exit with error if invalid

- [x] Update `src/cli/start.go`:
  - [x] Pass log level from flag to logger initialization
  - [x] Default to "info" in production mode

- [x] Update `src/cli/dev.go`:
  - [x] Default to "debug" in dev mode
  - [x] Allow override via `--log-level` flag

### 10.6 Add Detection Summary Logging

**Reference:** `specs/logging.md` (Log Automation Cycle Summary section)

- [x] Add to `src/logger/logger.go`:
  - [x] `LogDetectionComplete(serverName, serverType string, missing, cutoffUnmet int)`
  - [x] Console format: `INFO Detection complete server=X missing=Y cutoff_unmet=Z`

- [x] Update `src/services/automation.go`:
  - [x] After detection, call `logger.LogDetectionComplete()` for each server result

### 10.7 Add Detailed Search Logging with Metadata

**Reference:** `specs/logging.md` (Log Individual Search Triggers), `specs/activity-logging.md`

- [x] Update `src/services/types.go`:
  - [x] Add `Title string` field to TriggerResult
  - [x] Add `Year int` field to TriggerResult (for movies)
  - [x] Add `SeriesTitle string` field (for episodes)
  - [x] Add `SeasonNumber int` field
  - [x] Add `EpisodeNumber int` field
  - [x] Add `QualityProfile string` field
  - [x] Add `MissingItems` and `CutoffItems` maps to DetectionResult for metadata

- [x] Update `src/api/types.go`:
  - [x] Added `Year` and `QualityProfile` fields to `Movie`
  - [x] Added `QualityProfile` to `Series` (used by episodes)
  - [x] Added `EpisodeTitle` field to `MediaItem` for raw episode title

- [x] Update `src/services/search_trigger.go`:
  - [x] Added logger interface and field to SearchTrigger
  - [x] Updated to log each item individually with metadata before triggering
  - [x] Populated metadata from DetectionResult's MediaItem maps

- [x] Update `src/logger/logger.go`:
  - [x] Added `LogMovieSearch(server, title string, year int, quality, category string)`
  - [x] Added `LogEpisodeSearch(server, series, episodeTitle string, season, episode int, quality, category string)`
  - [x] Console format per spec:
    - Movies: `INFO Search triggered title="X" year=Y quality="Z" server=S category=C`
    - Episodes: `INFO Search triggered series="X" episode="S01E02" title="Y" quality="Z" server=S category=C`

### 10.8 Add Development Mode Verbose Logging

**Reference:** `specs/logging.md` (Development Mode Verbose Logging section)

- [x] Update `src/web/middleware/logging.go`:
  - [x] Accept logger in constructor
  - [x] Log HTTP requests at debug level: `DEBUG HTTP request method=GET path=/api/servers status=200 duration=12ms`

- [x] Update `src/api/client.go`:
  - [x] Add optional `logger` field via `DebugLogger` interface
  - [x] Add `WithLogger(l DebugLogger, serverName string) *Client` method
  - [x] Log API requests in debug mode (without API keys)

- [x] Update `src/services/scheduler.go`:
  - [x] Add logger field via `DebugLogger` interface
  - [x] Add `WithLogger(l DebugLogger) *Scheduler` method
  - [x] Log scheduler events: `DEBUG Scheduler sleeping until=X`, `DEBUG Scheduler woke up reason=timer`

- [x] Update `src/web/websocket/hub.go`:
  - [x] Log WebSocket connections at debug level (connect/disconnect with client counts)

- [x] Update initialization code:
  - [x] `src/web/server.go` - Pass logger to RequestLogger middleware
  - [x] `src/cli/start.go` - Attach logger to scheduler
  - [x] `src/cli/dev.go` - Attach logger to scheduler

- [x] Add `Debug()` method to `src/logger/logger.go` for console-only debug logging

### 10.9 Update Database Schema for Enhanced Logs

**Reference:** `specs/logging.md` (Web Log Entry Format section)

Current schema at `src/database/migrations/001_initial_schema.sql:17-27` lacks `operation` and `metadata` columns.

- [x] Create `src/database/migrations/002_enhanced_logs.sql`:

  ```sql
  ALTER TABLE logs ADD COLUMN operation TEXT;
  ALTER TABLE logs ADD COLUMN metadata TEXT;
  CREATE INDEX IF NOT EXISTS idx_logs_operation ON logs(operation);
  ```

- [x] Update `src/database/database.go`:
  - [x] Add migration 002 to embedded migrations
  - [x] Implement proper migration tracking with `schema_migrations` table

- [x] Update `src/database/logs.go`:
  - [x] Add `operation` and `metadata` to insert/select queries
  - [x] Add filter method: `GetLogsByOperation(operation string)`
  - [x] Add JSON serialization/deserialization for metadata

- [x] Update `src/logger/types.go`:
  - [x] Add `Operation string` field to LogEntry
  - [x] Add `Metadata map[string]interface{}` field to LogEntry

### 10.10 Web Log Viewer Filter Enhancements

**Reference:** `specs/logging.md` (Filter Logs section)

Current state: `src/web/handlers/api/logs.go` only supports `type` and `server` filters.

- [x] Update `src/web/handlers/api/logs.go`:
  - [x] Add `level` query param filter (filters by type)
  - [x] Add `operation` query param filter
  - [x] Add `from` datetime query param filter
  - [x] Add `to` datetime query param filter
  - [ ] Return total count in response for pagination (deferred)

- [x] Update `src/database/logs.go`:
  - [x] Add date range filtering to GetLogs
  - [x] Add operation filtering
  - [x] Updated LogFilters struct in logger/types.go with Type, Server, Operation, FromDate, ToDate

- [x] Update `src/templates/pages/logs.templ`:
  - [x] Add date range pickers (from/to)
  - [x] Add operation type dropdown
  - [x] Add "Clear filters" button
  - [ ] Sync filter state to URL query params (not implemented - filters work via HTMX)

**Implementation Notes:**

- Created `logger.LogFilters` struct to consolidate filter parameters
- Updated all GetLogs call sites to use new LogFilters struct
- Updated mockDB in logger tests to match new interface
- Filters are applied via HTMX and work correctly with partial updates
- Date inputs use HTML5 datetime-local type for easy date/time selection

### 10.11 Dashboard Log Summary Widget

**Reference:** `specs/logging.md` (Dashboard Log Summary section)

Current state: Dashboard at `src/templates/pages/dashboard.templ:116-159` shows recent activity but no error count badge.

- [x] Update `src/templates/pages/dashboard.templ`:
  - [x] Add 24-hour error count badge to recent activity section (already implemented via error stats card)
  - [x] Add "View all logs" link (already implemented at line 155)

- [x] Update `src/web/handlers/pages/dashboard.go`:
  - [x] Add 24-hour error count to dashboard data (lines 56-62, 100-106)

- [x] Update `src/database/logs.go`:
  - [x] Add `GetErrorCount(since time.Time) (int, error)` method (already implemented at lines 247-256)

**Implementation Notes:**

- The dashboard already had an error count badge in the stats cards section
- The "View all logs" link was already present in the Recent Activity section
- Updated the handler to use `GetErrorCount()` with a 24-hour time window instead of counting errors from recent logs
- Applied the same change to both `HandleDashboard` and `HandleStatsPartial` for consistency

### 10.12 Implement Log Retention

**Reference:** `specs/logging.md` (Log Retention and Cleanup section)

- [x] Update `src/database/logs.go`:
  - [x] Add `PurgeOldLogs(ctx, retentionDays int) (int, error)` method (lines 269-290)
  - [x] Add `GetLogCount(ctx) (int, error)` method (lines 258-267)

- [x] Update `src/database/types.go`:
  - [x] Add `LogsConfig` struct with `RetentionDays` field (lines 70-73)
  - [x] Add `Logs` field to `AppConfig` struct (line 79)
  - [x] Add default retention to `DefaultAppConfig()` (lines 95-97)

- [x] Update `src/database/config.go`:
  - [x] Add `logs.retention_days` config key with 7-90 day range validation (lines 49-57)
  - [x] Add logs config handling in SetAppConfigFunc (lines 85-87)

- [x] Create `src/services/maintenance.go`:
  - [x] `RunLogCleanup(ctx, db, logger)` function (lines 17-44)
  - [x] Delete logs older than configured retention with safety minimum of 7 days

- [x] Update `src/services/scheduler.go`:
  - [x] Add `lastCleanupDate` field to track daily cleanup (line 30)
  - [x] Add `runDailyCleanup()` method (lines 196-213)
  - [x] Integrate cleanup into run loop (line 182)
  - [x] Updated DebugLogger interface to include Info and Error methods (lines 13-17)

- [x] Update `src/templates/pages/settings.templ`:
  - [x] Updated signature to accept logCount parameter (line 9)

- [x] Update `src/templates/components/forms/config_form.templ`:
  - [x] Updated signature to accept logCount parameter (line 6)
  - [x] Add log retention setting dropdown with 7, 14, 30, 60, 90 days options (lines 118-144)
  - [x] Display current log count (lines 140-142)

- [x] Update `src/web/handlers/pages/settings.go`:
  - [x] Get log count and pass to template (lines 13-20)

- [x] Update `src/web/handlers/api/config.go`:
  - [x] Add PostConfig handler for form submission (lines 109-170)
  - [x] Handle `logs.retention_days` in POST endpoint (lines 157-162)

- [x] Update `src/web/server.go`:
  - [x] Add POST route for /api/config (line 116)

- [x] Update `src/logger/logger.go`:
  - [x] Add Info() method for console logging (lines 241-244)
  - [x] Add Error() method for console logging (lines 246-249)

**Implementation Notes:**

- Log cleanup runs once per day after each automation cycle
- Cleanup runs in background goroutine to avoid blocking automation
- Retention period validated to 7-90 day range for safety
- Current log count displayed in settings UI
- All tests pass, build successful

### 10.13 Write Tests

- [x] Create `src/logger/level_test.go`:
  - [x] Test ParseLevel with valid/invalid inputs
  - [x] Test Level.String()

- [x] Create `src/logger/console_test.go`:
  - [x] Test log level filtering

- [x] Update `src/logger/logger_test.go`:
  - [x] Test new constructor signature
  - [x] Test level filtering

**Implementation Notes:**

- All tests implemented and passing
- level_test.go: Tests ParseLevel with valid/invalid inputs, mixed case, whitespace
- console_test.go: Tests log level filtering, SetLevel method, toCharmLevel conversion
- logger_test.go: Tests updated to use new constructor signature with level and isDev params
- All logger tests pass: `go test ./src/logger/...`
- All services tests pass: `go test ./src/services/...`
- Full test suite passes: `go test ./...`

### 10.14 Verification

- [x] Run unit tests: `go test ./src/logger/...`
- [x] Run integration tests: `go test ./src/services/...`
- [x] Build binary: `make build`
- [ ] Manual testing:
  - [ ] `./janitarr dev` shows debug logs with colors
  - [ ] `./janitarr start` shows info logs only
  - [ ] `./janitarr start --log-level debug` shows debug logs
  - [ ] Web UI logs page shows all filters
  - [ ] Dashboard shows error count badge

**Automated Verification Complete:**

- All unit tests pass (logger, database, API, services, crypto, websocket)
- All integration tests pass
- Binary builds successfully with templ and tailwind
- No compilation errors or warnings

---

## Phase 11: Interactive CLI Forms

**Reference:** `specs/cli-interface.md`
**Verification:** `go test ./src/cli/... && go build ./src`

### 11.1 Add charmbracelet/huh Dependency

- [x] Add dependency: `go get github.com/charmbracelet/huh`
- [x] Verify import works

### 11.2 Create Forms Package Structure

- [x] Create `src/cli/forms/` directory
- [x] Create `src/cli/forms/helpers.go`:
  - [x] `IsInteractive() bool` - check if stdin is a TTY using `golang.org/x/term`
  - [x] Common validation functions:
    - [x] `ValidateServerName(s string) error`
    - [x] `ValidateURL(s string) error`
    - [x] `ValidateAPIKey(s string) error`
    - [x] `ValidateServerType(s string) error`

### 11.3 Server Add Form

**Reference:** `specs/cli-interface.md` (Interactive Server Addition section)

Current state: `src/cli/server.go:57-124` uses `bufio.NewReader` with manual prompts.

- [x] Create `src/cli/forms/server.go`:
  - [x] `ServerAddForm() (*ServerFormResult, error)`:
    - [x] Select field for server type (Radarr/Sonarr)
    - [x] Input field for name with validation
    - [x] Input field for URL with validation
    - [x] Input field for API key with `EchoMode(huh.EchoModePassword)`
  - [x] Return nil on Escape/cancel

- [x] Update `src/cli/server.go` (`runServerAdd`):
  - [x] Check `forms.IsInteractive()`
  - [x] If interactive and no flags provided, call `forms.ServerAddForm()`
  - [x] If non-interactive or all flags provided, use existing flag-based logic
  - [x] Show spinner during connection test
  - [x] Add flags: --name, --type, --url, --api-key

### 11.4 Server Edit Form

**Reference:** `specs/cli-interface.md` (Interactive Server Editing section)

Current state: `src/cli/server.go:151-227` uses basic prompts.

- [x] Add to `src/cli/forms/server.go`:
  - [x] `ServerEditForm(current *ServerFormResult) (*ServerFormResult, error)`:
    - [x] Pre-populate fields with current values
    - [x] Server type displayed but disabled
    - [x] "Keep existing API key" option
    - [x] Return only changed fields

- [x] Update `src/cli/server.go` (`runServerEdit`):
  - [x] If interactive and only server name provided, show edit form
  - [x] Pre-populate form with existing server values

**Implementation Notes:**

- `ServerEditForm` already existed in `src/cli/forms/server.go` (lines 94-182)
- Updated `runServerEdit` to detect interactive mode and use the form when no flags are provided
- Added fallback to flag-based editing when `--name`, `--url`, or `--api-key` flags are used
- Form provides "Keep existing key" / "Enter new key" option for API key updates
- Uses `db.GetServer()` / `db.GetServerByName()` to get full Server object with APIKey
- All tests pass, build successful

### 11.5 Server Selector

**Reference:** `specs/cli-interface.md` (Server List with Interactive Selection section)

- [x] Add to `src/cli/forms/server.go`:
  - [x] `ServerSelector(servers []ServerInfo) (*ServerInfo, error)`:
    - [x] Use `huh.NewSelect()` with server list
    - [x] Display: name, type, enabled status
    - [x] Return selected server or nil on cancel

- [x] Update `src/cli/server.go`:
  - [x] `server edit` (no name arg): show selector, then edit form
  - [x] `server remove` (no name arg): show selector, then confirmation
  - [x] `server test` command does not exist (skipped)

**Implementation Notes:**

- `ServerSelector` already existed in `src/cli/forms/server.go` (lines 192-239)
- Updated `serverEditCmd` to use `cobra.MaximumNArgs(1)` instead of `cobra.ExactArgs(1)`
- Updated `runServerEdit` to detect when no argument is provided and show selector in interactive mode
- Updated `serverRemoveCmd` to use `cobra.MaximumNArgs(1)` instead of `cobra.ExactArgs(1)`
- Updated `runServerRemove` to detect when no argument is provided and show selector in interactive mode
- When no argument provided in non-interactive mode, commands error appropriately
- Used `db.GetServer()` and `db.GetServerByName()` to retrieve full server objects with API keys
- All tests pass, build successful

### 11.6 Configuration Form

**Reference:** `specs/cli-interface.md` (Interactive Configuration section)

Current state: `src/cli/config.go` only has flag-based `config show` and `config set`.

- [x] Create `src/cli/forms/config.go`:
  - [x] `ConfigForm(current AppConfig) (*AppConfig, error)`:
    - [x] Group: Automation (enabled toggle, interval number)
    - [x] Group: Search Limits (4 number inputs)
    - [x] Group: Log Retention (retention days)
    - [x] Pre-populate with current values
    - [x] Validation for interval (1-168 hours)
    - [x] Validation for search limits (0-100)
    - [x] Validation for retention days (7-90)

- [x] Update `src/cli/config.go`:
  - [x] Add `config` command (no subcommand) that launches form when interactive
  - [x] Keep `config show` and `config set` for non-interactive use
  - [x] Import forms package and add runConfigInteractive function
  - [x] Check IsInteractive() before showing form

**Implementation Notes:**

- Created ConfigForm with three groups: Automation, Search Limits, and Log Retention
- Form validates all inputs with appropriate ranges per spec
- runConfigInteractive checks if terminal is interactive, shows help if not
- Pre-populates all fields with current database values
- Successfully saves updated configuration to database
- All tests pass, build successful

### 11.7 Confirmation Dialogs

**Reference:** `specs/cli-interface.md` (Confirmation Dialogs section)

Current state: `src/cli/server.go:252-257` uses basic Y/N prompt.

- [x] Create `src/cli/forms/confirm.go`:
  - [x] `ConfirmDelete(itemType, itemName string) (bool, error)`:
    - [x] Show item details
    - [x] Require typing item name to confirm
  - [x] `ConfirmAction(message string) (bool, error)`:
    - [x] Simple yes/no confirmation
  - [x] `ConfirmActionWithDetails(title, details string) (bool, error)`:
    - [x] Confirmation with additional context

- [x] Update `src/cli/server.go` (`runServerRemove`):
  - [x] If interactive and no `--force`, show `ConfirmDelete`
  - [x] Fallback to basic Y/N prompt if not interactive

- [ ] Update `src/cli/logs.go`:
  - [ ] If interactive and clearing logs, show `ConfirmAction` with log count

### 11.8 Non-Interactive Mode Flag

**Reference:** `specs/cli-interface.md` (Flag Override section)

- [x] Update `src/cli/root.go`:
  - [x] Add `--non-interactive` global flag
  - [x] When set, skip all interactive forms and require flags

- [x] Update all form-using commands:
  - [x] Check `--non-interactive` flag
  - [x] Error with usage if required flags missing

**Implementation Notes:**

- Added `nonInteractive` bool variable to `src/cli/root.go` (line 14)
- Added `--non-interactive` persistent flag to root command (line 43)
- Created `ShouldUseInteractiveMode(nonInteractiveFlag bool)` helper in `src/cli/forms/helpers.go` (lines 18-27)
- Updated `runServerAdd` to use `ShouldUseInteractiveMode(nonInteractive)` instead of `IsInteractive()` (line 89)
- Updated `runServerEdit` to check flag in two places: selector logic (line 183) and form logic (line 249)
- Updated `runServerRemove` to check flag in two places: selector logic (line 323) and confirmation logic (line 380)
- Updated `runConfigInteractive` to check flag before showing form (line 125)
- All tests pass: `go test ./...`
- Binary builds successfully: `make build`

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
