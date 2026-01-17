# Unified Service Startup: Running Scheduler and Web Server Together

## Context

Janitarr consists of two primary runtime components:
1. **Scheduler daemon** - Executes automated detection and search cycles on a configured interval
2. **Web server** - Provides REST API and web interface for managing configuration and monitoring activity

Currently, these services must be started separately using different commands (`start` for scheduler, `serve` for web server). This creates friction for users and complicates deployment. The services should start together as a unified application.

Following the Next.js convention, the system should provide two startup modes:
- **Development mode** (`dev` command) - Optimized for local development with verbose logging and frontend dev server proxy
- **Production mode** (`start` command) - Optimized for deployed environments with built assets and minimal logging

## Requirements

### Story: Start Production Service

- **As a** user deploying Janitarr
- **I want to** start both the scheduler and web server with a single command
- **So that** I can run the complete application without managing multiple processes

#### Acceptance Criteria

- [ ] `janitarr start` command launches both scheduler daemon and web server
- [ ] Web server serves built frontend assets from `dist/public/`
- [ ] Both services run in the same process
- [ ] Command accepts `--port <number>` flag to configure web server port (default: 3434)
- [ ] Command accepts `--host <string>` flag to configure web server bind address (default: localhost)
- [ ] If scheduler is disabled in config, only web server starts (with warning message)
- [ ] Process displays startup confirmation for both services with URLs and configuration
- [ ] Graceful shutdown on SIGINT (Ctrl+C) stops both services cleanly

### Story: Development Mode with Frontend Proxy

- **As a** developer working on Janitarr
- **I want to** run the backend services while proxying frontend requests to Vite dev server
- **So that** I can work on the UI with hot module reloading

#### Acceptance Criteria

- [ ] `janitarr dev` command launches both scheduler daemon and web server in development mode
- [ ] Web server proxies non-API requests to Vite dev server at `http://localhost:5173`
- [ ] API requests (`/api/*`, `/ws/*`) are handled by Janitarr backend
- [ ] Verbose logging enabled: all HTTP requests, scheduler events, automation cycles logged to console
- [ ] API error responses include detailed stack traces and debugging information
- [ ] Command accepts same `--port` and `--host` flags as production mode
- [ ] Clear indication in console output that development mode is active
- [ ] Frontend developer can run `bun run dev` in `ui/` directory separately to start Vite

### Story: Health Check Endpoint

- **As a** deployment system or monitoring tool
- **I want to** query the application health status
- **So that** I can verify both services are running correctly

#### Acceptance Criteria

- [ ] `GET /api/health` endpoint returns service status
- [ ] Response includes both scheduler and web server status
- [ ] Response format:
  ```json
  {
    "status": "ok" | "degraded" | "error",
    "timestamp": "2026-01-17T10:30:00Z",
    "services": {
      "webServer": { "status": "ok" },
      "scheduler": {
        "status": "ok" | "disabled" | "error",
        "isRunning": true,
        "isCycleActive": false,
        "nextRun": "2026-01-17T11:30:00Z"
      }
    },
    "database": { "status": "ok" | "error" }
  }
  ```
- [ ] Overall status is `ok` when all services healthy
- [ ] Overall status is `degraded` when scheduler disabled but web server running
- [ ] Overall status is `error` when any critical component failing
- [ ] Endpoint returns HTTP 200 for `ok` and `degraded`
- [ ] Endpoint returns HTTP 503 for `error`
- [ ] Endpoint accessible in both dev and production modes
- [ ] Response time < 100ms (lightweight check, no expensive operations)

### Story: Prometheus Metrics Endpoint

- **As a** monitoring system administrator
- **I want to** scrape Prometheus-compatible metrics from Janitarr
- **So that** I can monitor performance, track trends, and set up alerts

#### Acceptance Criteria

- [ ] `GET /metrics` endpoint returns Prometheus text format metrics
- [ ] Content-Type header is `text/plain; version=0.0.4; charset=utf-8`
- [ ] Endpoint accessible in both dev and production modes
- [ ] Response time < 200ms (efficient metric collection)
- [ ] Metrics exposed:
  - **Application info**:
    - `janitarr_info{version}` - Application version (gauge, always 1)
    - `janitarr_uptime_seconds` - Time since process start (counter)
  - **Scheduler metrics**:
    - `janitarr_scheduler_enabled` - Whether scheduler is enabled (gauge, 0 or 1)
    - `janitarr_scheduler_running` - Whether scheduler is running (gauge, 0 or 1)
    - `janitarr_scheduler_cycle_active` - Whether automation cycle is active (gauge, 0 or 1)
    - `janitarr_scheduler_cycles_total` - Total automation cycles executed (counter)
    - `janitarr_scheduler_cycles_failed_total` - Total failed cycles (counter)
    - `janitarr_scheduler_next_run_timestamp` - Unix timestamp of next scheduled run (gauge)
  - **Search metrics**:
    - `janitarr_searches_triggered_total{server_type,category}` - Total searches triggered by type (counter)
      - Labels: `server_type` (radarr/sonarr), `category` (missing/cutoff)
    - `janitarr_searches_failed_total{server_type,category}` - Total failed searches (counter)
  - **Server metrics**:
    - `janitarr_servers_configured{type}` - Number of configured servers by type (gauge)
    - `janitarr_servers_enabled{type}` - Number of enabled servers by type (gauge)
  - **Database metrics**:
    - `janitarr_database_connected` - Database connection status (gauge, 0 or 1)
    - `janitarr_logs_total` - Total log entries in database (gauge)
  - **HTTP metrics**:
    - `janitarr_http_requests_total{method,path,status}` - Total HTTP requests (counter)
    - `janitarr_http_request_duration_seconds{method,path}` - HTTP request duration histogram (histogram)
- [ ] Counters never decrease (monotonic)
- [ ] Gauges reflect current state
- [ ] Labels follow Prometheus naming conventions (snake_case, lowercase)
- [ ] All metric names prefixed with `janitarr_`
- [ ] HELP and TYPE annotations included for each metric
- [ ] Invalid/missing data reported as NaN or omitted (not as 0)

### Story: Graceful Shutdown of Both Services

- **As a** user running Janitarr
- **I want to** stop the application cleanly with Ctrl+C
- **So that** ongoing operations complete and resources are released properly

#### Acceptance Criteria

- [ ] SIGINT signal (Ctrl+C) triggers graceful shutdown sequence
- [ ] Scheduler stops accepting new cycles and waits for active cycle to complete (if any)
- [ ] Web server stops accepting new connections but completes in-flight requests
- [ ] WebSocket connections are closed gracefully with proper close frames
- [ ] Console output confirms each service has stopped successfully
- [ ] Process exits with code 0 after clean shutdown
- [ ] Maximum shutdown timeout of 10 seconds before force exit

### Story: Command-Line Configuration

- **As a** user
- **I want to** configure web server port and host via command-line flags
- **So that** I can adapt to different network environments without changing configuration

#### Acceptance Criteria

- [ ] Both `dev` and `start` commands accept `--port` / `-p` flag
- [ ] Both `dev` and `start` commands accept `--host` / `-h` flag
- [ ] Port must be valid integer between 1 and 65535
- [ ] Invalid port number displays error and exits without starting services
- [ ] Host defaults to `localhost` (security: prevents external access by default)
- [ ] Port defaults to `3434` (avoids common conflicts with port 3000)
- [ ] Configuration flags override any database-stored settings

### Story: Migration from Separate Commands

- **As a** existing Janitarr user
- **I want to** the `serve` command to be removed
- **So that** the interface is simplified and there's one clear way to start services

#### Acceptance Criteria

- [ ] `janitarr serve` command is completely removed from CLI
- [ ] Users running old `serve` command receive "unknown command" error
- [ ] Documentation updated to show `start` and `dev` commands only
- [ ] Migration path clear: `serve` → `start`, `serve --port 8080` → `start --port 8080`

## Edge Cases & Constraints

### Service Interdependencies

- Scheduler can run independently without web server, but web server depends on database and services that scheduler uses
- If scheduler is disabled in config (`schedule.enabled = false`), only web server starts
- Web API includes endpoints to manually trigger automation cycles even when scheduler is disabled

### Port Conflicts

- If specified port is already in use, display clear error message with port number
- Suggest checking for existing Janitarr instances or other services
- Exit gracefully without starting scheduler if web server fails to bind

### Development Mode Requirements

- Development mode assumes Vite dev server is running separately on port 5173
- If Vite dev server is not accessible, proxy requests fail with 502 Bad Gateway
- Frontend developers must run `cd ui && bun run dev` in separate terminal
- Backend-only developers can use production mode (`start`) and access built UI

### Logging Behavior

**Production Mode (`start`):**
- Log level: INFO
- Output: Startup messages, scheduler events, automation cycle summaries, errors
- HTTP request logging: Disabled (only errors logged)
- API errors: Generic messages without stack traces

**Development Mode (`dev`):**
- Log level: DEBUG
- Output: All production logs plus HTTP requests, WebSocket messages, detailed timing
- HTTP request logging: Enabled (method, path, status code, response time)
- API errors: Full stack traces and request details
- Scheduler events: Verbose output showing detection counts, search decisions

### Resource Management

- Web server and scheduler share single database connection pool
- WebSocket broadcasts from logger must be thread-safe (if Bun uses worker threads)
- Memory: Both services in single process may use more RAM than separate processes
- Suggested minimum: 128MB RAM for typical deployments

### Configuration Persistence

- Command-line flags (`--port`, `--host`) are runtime-only and not saved to database
- Scheduler configuration (`intervalHours`, `enabled`) persists in database
- Search limits configuration persists in database
- Users can modify database config while services running (changes apply on next scheduler cycle)

### Error Recovery

- If scheduler encounters fatal error during startup, web server continues running
- If web server encounters fatal error during startup, scheduler stops and process exits
- Partial startup states are avoided: both services start or neither starts (web server failure case)
- Runtime errors in scheduler (e.g., API failures) are logged but don't crash process
- Runtime errors in web server (e.g., malformed requests) return HTTP errors but don't crash process

### Health Check Implementation

- Health check must be lightweight and not trigger actual service operations
- Database health verified by simple query (e.g., `SELECT 1`)
- Scheduler status read from in-memory state (no expensive checks)
- Health check should not be blocked by ongoing automation cycles
- Useful for Docker HEALTHCHECK, Kubernetes liveness/readiness probes, monitoring systems

### Prometheus Metrics Implementation

- Metrics collected and stored in-memory (no persistent storage)
- Counters incremented atomically to prevent race conditions
- Metrics reset on application restart (expected Prometheus behavior)
- HTTP request metrics collected via middleware (minimal overhead)
- Expensive metrics (e.g., log counts) may query database but cached briefly
- Example output format:
  ```
  # HELP janitarr_info Application version information
  # TYPE janitarr_info gauge
  janitarr_info{version="0.1.0"} 1

  # HELP janitarr_uptime_seconds Time since process start
  # TYPE janitarr_uptime_seconds counter
  janitarr_uptime_seconds 3600

  # HELP janitarr_scheduler_running Whether scheduler is running
  # TYPE janitarr_scheduler_running gauge
  janitarr_scheduler_running 1

  # HELP janitarr_searches_triggered_total Total searches triggered
  # TYPE janitarr_searches_triggered_total counter
  janitarr_searches_triggered_total{server_type="radarr",category="missing"} 42
  janitarr_searches_triggered_total{server_type="sonarr",category="cutoff"} 15

  # HELP janitarr_http_requests_total Total HTTP requests
  # TYPE janitarr_http_requests_total counter
  janitarr_http_requests_total{method="GET",path="/api/servers",status="200"} 123
  ```
- Prometheus scrape configuration recommendation: 15-30 second intervals
- Metrics library: Implement custom formatting (no external dependencies needed)

### Backwards Compatibility

- Existing `janitarr start` behavior changes from scheduler-only to unified startup
- Users currently running `janitarr start` in background will now also run web server
- Breaking change acceptable: v0.x semantic versioning, document in release notes
- Docker deployments need single port exposure instead of potential multiple ports

### Security Considerations

- Default host `localhost` prevents external network access (intentional for home use)
- Users must explicitly set `--host 0.0.0.0` to allow external connections
- Development mode should only be used in trusted environments (exposes stack traces)
- No authentication in v1 - rely on network-level access control

## Implementation Notes

### Command Structure

```bash
# Production mode - both scheduler and web server
janitarr start [--port 3434] [--host localhost]

# Development mode - both services with verbose logging and Vite proxy
janitarr dev [--port 3434] [--host localhost]

# Removed (no longer available)
janitarr serve  # ❌ Command removed
```

### API Endpoints

The existing API at `/api/health` needs to be enhanced to include scheduler status. Current implementation returns basic status only.

**Updated response schema:**
```typescript
interface HealthResponse {
  status: 'ok' | 'degraded' | 'error';
  timestamp: string; // ISO 8601
  services: {
    webServer: { status: 'ok' };
    scheduler: {
      status: 'ok' | 'disabled' | 'error';
      isRunning: boolean;
      isCycleActive: boolean;
      nextRun: string | null; // ISO 8601
    };
  };
  database: { status: 'ok' | 'error' };
}
```

### Web Server Modifications Required

The `createWebServer` function in `src/web/server.ts` needs:

1. New `isDev` parameter to enable development mode
2. Proxy middleware for non-API requests when `isDev === true`
3. Conditional logging based on mode
4. Detailed error responses in dev mode
5. Return server instance for shutdown capability
6. Enhanced `/api/health` endpoint with scheduler and database status

### Scheduler Integration

The scheduler already supports:
- `startScheduler()` - Begin scheduled cycles
- `stopScheduler()` - Stop accepting new cycles
- `isSchedulerRunning()` - Check if active
- `registerCycleCallback()` - Register automation cycle function
- `getStatus()` - Get current scheduler state
- `getTimeUntilNextRun()` - Get time until next cycle

No changes needed to scheduler itself, only integration in startup command and health check.

### File Structure

```
src/
├── cli/
│   └── commands.ts        # Modified: update start, add dev, remove serve
├── web/
│   ├── server.ts          # Modified: add isDev support, proxy, enhanced health, metrics
│   └── routes/
│       ├── health.ts      # New: dedicated health check route handler
│       └── metrics.ts     # New: Prometheus metrics endpoint handler
└── lib/
    ├── scheduler.ts       # No changes needed
    └── metrics.ts         # New: metrics collection and formatting utilities
```

## Success Metrics

1. **Developer Experience**: Single command to start full development environment
2. **Deployment Simplicity**: One process to manage in production
3. **Migration Path**: Clear upgrade path from separate commands
4. **Reliability**: Both services shutdown cleanly without data loss
5. **Performance**: No measurable overhead from running in single process vs separate
6. **Observability**: Health endpoint responds in < 100ms and accurately reports status
7. **Monitoring**: Prometheus metrics endpoint provides comprehensive operational visibility

## Future Enhancements (Post-v1)

1. **Process Manager Integration**: systemd/pm2 service files for production deployments
2. **Configuration File**: `janitarr.config.js` for setting defaults (port, host, etc.)
3. **Graceful Reload**: SIGHUP signal to reload configuration without downtime
4. **Cluster Mode**: Multiple worker processes for high-traffic deployments (Bun cluster support)
5. **Auto-restart**: Watch for critical failures and auto-restart services
6. **Advanced Metrics**: Additional custom metrics, histogram buckets configuration, metric retention
