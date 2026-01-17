# Janitarr Implementation Plan

**Last Updated:** 2026-01-17
**Status:** ⚠️ FEATURE PENDING - Unified Service Startup Not Implemented
**Overall Completion:** 95% (Core features complete, new spec pending)

---

## Executive Summary

Janitarr is a production-ready automation tool for managing Radarr/Sonarr media servers with both CLI and web interfaces. All original core functionality has been implemented and tested. A new specification (`unified-service-startup.md`) has been added requiring significant changes to service startup commands, health checks, and metrics.

**Current State:**
- ✅ **CLI Application**: 100% complete (original spec)
- ✅ **Backend Services**: 100% complete with comprehensive test coverage
- ✅ **Web Backend API**: 100% complete with REST + WebSocket
- ✅ **Web Frontend**: 100% core functionality implemented
- ✅ **Testing**: 181 tests passing (142 backend, 36 frontend, 3 E2E)
- ⚠️ **Unified Service Startup**: NOT IMPLEMENTED (new spec)
- ✅ **Health Check Endpoint**: COMPLETE with comprehensive status reporting
- ⚠️ **Prometheus Metrics**: NOT IMPLEMENTED

---

## Priority 1: Unified Service Startup (MUST DO) ⚠️ NEW

### Task 1.1: Update `start` Command for Unified Startup
**Impact:** HIGH - Breaking change, required for deployment simplicity
**Spec:** `specs/unified-service-startup.md` lines 17-33
**Status:** ❌ NOT STARTED

**Current Behavior:**
- `janitarr start` - Only starts scheduler daemon
- `janitarr serve` - Only starts web server
- Users must run two commands/processes separately

**Required Changes:**
- [ ] Modify `start` command to launch both scheduler AND web server together
- [ ] Accept `--port <number>` flag (default: 3434)
- [ ] Accept `--host <string>` flag (default: localhost)
- [ ] Validate port number (1-65535)
- [ ] If scheduler disabled in config, only start web server with warning
- [ ] Display startup confirmation for both services with URLs
- [ ] Handle graceful shutdown on SIGINT for both services

**Files to Modify:**
- `src/cli/commands.ts` - Update `start` command implementation

**Acceptance Criteria:**
- [ ] `janitarr start` launches scheduler + web server in single process
- [ ] `janitarr start --port 8080 --host 0.0.0.0` works correctly
- [ ] Scheduler-disabled config shows warning but web server still starts
- [ ] Ctrl+C gracefully stops both services

---

### Task 1.2: Add `dev` Command for Development Mode
**Impact:** HIGH - Critical for developer experience
**Spec:** `specs/unified-service-startup.md` lines 34-50
**Status:** ❌ NOT STARTED

**Requirements:**
- [ ] New `janitarr dev` command launches both services in development mode
- [ ] Proxy non-API requests to Vite dev server at `http://localhost:5173`
- [ ] Enable verbose logging (DEBUG level)
- [ ] Log all HTTP requests with method, path, status, response time
- [ ] Include stack traces in API error responses
- [ ] Accept same `--port` and `--host` flags as production mode
- [ ] Clear console indication that development mode is active

**Files to Modify:**
- `src/cli/commands.ts` - Add new `dev` command
- `src/web/server.ts` - Add `isDev` parameter and proxy support

**Acceptance Criteria:**
- [ ] `janitarr dev` starts both services with verbose logging
- [ ] Non-API requests proxied to Vite (port 5173)
- [ ] API errors include full stack traces
- [ ] HTTP request logging shows method, path, status, duration

---

### Task 1.3: Remove `serve` Command
**Impact:** MEDIUM - Breaking change for existing users
**Spec:** `specs/unified-service-startup.md` lines 161-173
**Status:** ❌ NOT STARTED

**Requirements:**
- [ ] Completely remove `janitarr serve` command from CLI
- [ ] Users running `serve` receive "unknown command" error
- [ ] Update documentation to show only `start` and `dev` commands

**Files to Modify:**
- `src/cli/commands.ts` - Remove `serve` command (lines 601-636)
- `docs/user-guide.md` - Update command documentation
- `README.md` - Update quick start examples

**Acceptance Criteria:**
- [ ] `janitarr serve` returns "unknown command" error
- [ ] All documentation references `start` instead of `serve`

---

### Task 1.4: Enhanced Health Check Endpoint ✅ COMPLETE
**Impact:** HIGH - Required for deployments and monitoring
**Spec:** `specs/unified-service-startup.md` lines 51-84
**Status:** ✅ COMPLETE (2026-01-17)

**Implementation:**
- Created `src/web/routes/health.ts` with comprehensive health check handler
- Updated `src/web/server.ts` to use new handler
- Added `HttpStatus.SERVICE_UNAVAILABLE` (503) to types
- Created comprehensive test suite in `tests/web/routes/health.test.ts`

**Completed Response Format:**
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

**Requirements:**
- [x] Return comprehensive status for all services
- [x] Overall status `ok` when all services healthy
- [x] Overall status `degraded` when scheduler disabled
- [x] Overall status `error` when critical component failing
- [x] HTTP 200 for `ok` and `degraded`, HTTP 503 for `error`
- [x] Response time < 100ms (lightweight check)
- [x] Database health verified by simple query

**Files Created/Modified:**
- `src/web/routes/health.ts` - New dedicated health check handler
- `src/web/server.ts` - Updated route to use new handler
- `src/web/types.ts` - Added SERVICE_UNAVAILABLE status code
- `tests/web/routes/health.test.ts` - Comprehensive test suite (6 tests, all passing)

**Acceptance Criteria:**
- [x] `/api/health` returns full service status
- [x] Database connectivity verified
- [x] Scheduler status includes nextRun time
- [x] Response within 100ms
- [x] All tests passing (142 unit tests total)

---

### Task 1.5: Prometheus Metrics Endpoint
**Impact:** HIGH - Required for production monitoring
**Spec:** `specs/unified-service-startup.md` lines 86-127
**Status:** ❌ NOT STARTED

**Required Metrics:**
- **Application info:**
  - `janitarr_info{version}` - Application version (gauge)
  - `janitarr_uptime_seconds` - Time since process start (counter)
- **Scheduler metrics:**
  - `janitarr_scheduler_enabled` - Whether scheduler enabled (gauge)
  - `janitarr_scheduler_running` - Whether scheduler running (gauge)
  - `janitarr_scheduler_cycle_active` - Whether cycle active (gauge)
  - `janitarr_scheduler_cycles_total` - Total cycles executed (counter)
  - `janitarr_scheduler_cycles_failed_total` - Failed cycles (counter)
  - `janitarr_scheduler_next_run_timestamp` - Next run unix timestamp (gauge)
- **Search metrics:**
  - `janitarr_searches_triggered_total{server_type,category}` - Searches by type (counter)
  - `janitarr_searches_failed_total{server_type,category}` - Failed searches (counter)
- **Server metrics:**
  - `janitarr_servers_configured{type}` - Configured servers by type (gauge)
  - `janitarr_servers_enabled{type}` - Enabled servers by type (gauge)
- **Database metrics:**
  - `janitarr_database_connected` - Database status (gauge)
  - `janitarr_logs_total` - Total log entries (gauge)
- **HTTP metrics:**
  - `janitarr_http_requests_total{method,path,status}` - Request counter
  - `janitarr_http_request_duration_seconds{method,path}` - Request duration histogram

**Requirements:**
- [ ] `GET /metrics` returns Prometheus text format
- [ ] Content-Type: `text/plain; version=0.0.4; charset=utf-8`
- [ ] HELP and TYPE annotations for each metric
- [ ] All metrics prefixed with `janitarr_`
- [ ] Labels follow snake_case convention
- [ ] Response time < 200ms
- [ ] Counters never decrease (monotonic)
- [ ] Invalid/missing data reported as NaN or omitted

**Files to Create:**
- `src/lib/metrics.ts` - Metrics collection and formatting utilities
- `src/web/routes/metrics.ts` - Metrics endpoint handler

**Files to Modify:**
- `src/web/server.ts` - Add metrics route, add HTTP request middleware
- `src/services/automation.ts` - Track cycle counts
- `src/services/search-trigger.ts` - Track search counts

**Acceptance Criteria:**
- [ ] `/metrics` returns valid Prometheus format
- [ ] All specified metrics exposed
- [ ] HTTP request metrics collected via middleware
- [ ] Response within 200ms

---

### Task 1.6: Graceful Shutdown
**Impact:** MEDIUM - Required for clean deployments
**Spec:** `specs/unified-service-startup.md` lines 129-144
**Status:** ⚠️ PARTIAL (basic SIGINT handling exists)

**Current Implementation:**
- Basic SIGINT handler in `start` command stops scheduler
- Web server has `stopWebServer` function but not integrated

**Requirements:**
- [ ] SIGINT triggers graceful shutdown sequence
- [ ] Scheduler waits for active cycle to complete (with timeout)
- [ ] Web server completes in-flight requests
- [ ] WebSocket connections closed with proper close frames
- [ ] Console output confirms each service stopped
- [ ] Process exits with code 0 after clean shutdown
- [ ] Maximum 10-second timeout before force exit

**Files to Modify:**
- `src/cli/commands.ts` - Enhance shutdown handling in `start` command
- `src/web/server.ts` - Add graceful shutdown support
- `src/web/websocket.ts` - Add graceful close for all connections

**Acceptance Criteria:**
- [ ] Ctrl+C during cycle waits for completion (up to 10s)
- [ ] In-flight HTTP requests complete
- [ ] WebSocket clients receive close frames
- [ ] Exit code 0 on clean shutdown

---

## Priority 2: Testing for New Features (MUST DO)

### Task 2.1: Unit Tests for Unified Startup
**Impact:** HIGH - Critical for maintainability
**Status:** ❌ NOT STARTED

**Test Cases:**
- [ ] `start` command with default options
- [ ] `start` command with custom port/host
- [ ] `start` command with invalid port (error handling)
- [ ] `start` command with scheduler disabled (warning + web only)
- [ ] `dev` command proxy behavior
- [ ] `dev` command verbose logging

**Files to Create:**
- `tests/cli/commands.test.ts` - CLI command unit tests

---

### Task 2.2: Unit Tests for Health Check ✅ COMPLETE
**Impact:** MEDIUM - Required for reliable health monitoring
**Status:** ✅ COMPLETE (2026-01-17)

**Test Cases:**
- [x] Health returns `degraded` when scheduler disabled
- [x] Health returns `error` when scheduler enabled but not running
- [x] Response includes all required fields
- [x] Timestamp is valid ISO 8601 format
- [x] Returns JSON content type
- [x] Database status is ok when accessible

**Files Created:**
- `tests/web/routes/health.test.ts` - Health endpoint tests (6 tests, all passing)

---

### Task 2.3: Unit Tests for Metrics
**Impact:** MEDIUM - Required for reliable metrics
**Status:** ❌ NOT STARTED

**Test Cases:**
- [ ] Metrics formatting follows Prometheus spec
- [ ] All required metrics present
- [ ] Counter increment behavior
- [ ] Gauge update behavior
- [ ] Label formatting correct

**Files to Create:**
- `tests/lib/metrics.test.ts` - Metrics utility tests
- `tests/web/routes/metrics.test.ts` - Metrics endpoint tests

---

## Priority 3: Documentation Updates (SHOULD DO)

### Task 3.1: Update User Documentation
**Impact:** MEDIUM - Required for user adoption
**Status:** ❌ NOT STARTED

**Updates Required:**
- [ ] `docs/user-guide.md` - Update startup commands section
- [ ] `docs/troubleshooting.md` - Add unified startup issues
- [ ] `README.md` - Update quick start with new commands
- [ ] Remove all references to `serve` command

---

### Task 3.2: Update API Documentation
**Impact:** MEDIUM - Required for API consumers
**Status:** ❌ NOT STARTED

**Updates Required:**
- [ ] `docs/api-reference.md` - Document enhanced `/api/health` response
- [ ] `docs/api-reference.md` - Document new `/metrics` endpoint

---

## Completed Features ✅ (Original Spec)

### 1. Server Configuration (100% Complete)
**Spec:** `specs/server-configuration.md`
- ✅ Add/edit/remove Radarr and Sonarr servers
- ✅ URL normalization and validation
- ✅ Connection testing with 10-15 second timeout
- ✅ API key encryption at rest (AES-256-GCM)

### 2. Missing Content Detection (100% Complete)
**Spec:** `specs/missing-content-detection.md`
- ✅ Query Radarr for missing monitored movies
- ✅ Query Sonarr for missing monitored episodes
- ✅ Aggregate results across all servers

### 3. Quality Cutoff Detection (100% Complete)
**Spec:** `specs/quality-cutoff-detection.md`
- ✅ Query Radarr/Sonarr for items below quality cutoff
- ✅ Aggregate results across all servers

### 4. Search Triggering with Granular Limits (100% Complete)
**Spec:** `specs/search-triggering.md`
- ✅ 4 independent search limits
- ✅ Fair round-robin distribution across servers
- ✅ Dry-run mode for previewing searches

### 5. Automatic Scheduling (100% Complete)
**Spec:** `specs/automatic-scheduling.md`
- ✅ Configurable interval (minimum 1 hour)
- ✅ Background daemon with persistent schedule
- ✅ Manual trigger without affecting schedule

### 6. Activity Logging (100% Complete)
**Spec:** `specs/activity-logging.md`
- ✅ Individual search entries with timestamps
- ✅ 30-day automatic log retention
- ✅ Real-time WebSocket streaming

### 7. Web Backend API (100% Complete)
**Spec:** `specs/web-frontend.md` Phase 2.1
- ✅ All REST API endpoints
- ✅ WebSocket server for log streaming

### 8. Web Frontend (100% Complete)
**Spec:** `specs/web-frontend.md` Phases 2.2-2.3
- ✅ Dashboard, Servers, Logs, Settings views
- ✅ Mobile responsive design
- ✅ WCAG 2.1 Level AA accessibility

---

## Test Suite Summary

**Current Status:** 181 tests passing (142 unit, 36 frontend, 3 E2E)

**Test Commands:**
```bash
bun run test          # Backend unit tests (142 tests)
bun run test:ui       # Frontend tests (36 tests)
bun run test:e2e      # E2E tests (3 tests)
bun run test:all      # All unit + frontend tests
```

---

## Task Priority Summary

| Priority | Task | Status | Impact |
|----------|------|--------|--------|
| P1 | 1.1 Update `start` command | ❌ Not Started | HIGH |
| P1 | 1.2 Add `dev` command | ❌ Not Started | HIGH |
| P1 | 1.3 Remove `serve` command | ❌ Not Started | MEDIUM |
| P1 | 1.4 Enhanced health check | ✅ Complete | HIGH |
| P1 | 1.5 Prometheus metrics | ❌ Not Started | HIGH |
| P1 | 1.6 Graceful shutdown | ⚠️ Partial | MEDIUM |
| P2 | 2.1 Tests for unified startup | ❌ Not Started | HIGH |
| P2 | 2.2 Tests for health check | ✅ Complete | MEDIUM |
| P2 | 2.3 Tests for metrics | ❌ Not Started | MEDIUM |
| P3 | 3.1 Update user docs | ❌ Not Started | MEDIUM |
| P3 | 3.2 Update API docs | ❌ Not Started | MEDIUM |

---

## Recommended Implementation Order

1. **Task 1.4: Enhanced Health Check** - Foundation for monitoring
2. **Task 1.5: Prometheus Metrics** - Foundation for observability
3. **Task 1.1: Update `start` command** - Core unified startup
4. **Task 1.2: Add `dev` command** - Developer experience
5. **Task 1.6: Graceful shutdown** - Production reliability
6. **Task 1.3: Remove `serve` command** - Cleanup
7. **Task 2.x: Testing** - Validation
8. **Task 3.x: Documentation** - User communication

---

## Overall Assessment

**Status: ⚠️ FEATURE PENDING**

All original specifications are complete and working. A new specification (`unified-service-startup.md`) has been added that requires:
- Combining scheduler and web server into single process
- New `dev` command for development mode
- Removal of `serve` command (breaking change)
- Enhanced health check endpoint
- Prometheus metrics endpoint
- Improved graceful shutdown

**Estimated Effort:** 6-8 tasks across Priority 1-3
**Breaking Changes:** Yes (`start` behavior changes, `serve` removed)

---

**Last Reviewed:** 2026-01-17
**Next Action:** Implement Task 1.4 (Enhanced Health Check) as foundation
