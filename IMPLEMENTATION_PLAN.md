# Janitarr Implementation Plan

**Last Updated:** 2026-01-16
**Status:** Phase 2.1 Complete - Backend API Foundation Implemented
**Latest Update:** 2026-01-16 - Web Backend API with REST endpoints and WebSocket streaming complete

---

## Executive Summary

Janitarr is a well-architected automation tool for managing Radarr/Sonarr media servers. The core CLI functionality is **fully implemented and operational**, including server management, content detection, search triggering, scheduling, logging, and encryption.

**Current State:**
- ✅ Core CLI features complete (server management, detection, automation, logging)
- ✅ Encryption at rest (AES-256-GCM) for API keys
- ✅ Background scheduling with configurable intervals
- ✅ Comprehensive test suite (144 tests passing)
- ✅ Search limits granularity complete (4 separate limits)
- ✅ Web Backend API complete (REST + WebSocket)
- ❌ Web frontend React application not implemented (remaining work)
- ⚠️ Minor CLI command variations from spec (optional)

---

## Gap Analysis Validation Summary

**Methodology:** Systematic code review comparing implementation against all specifications

**Files Reviewed:**
- ✅ All 8 specification documents in `specs/`
- ✅ All 13 TypeScript source files in `src/`
- ✅ All 10 test files in `tests/`
- ✅ Database schema and configuration (`src/storage/database.ts`)
- ✅ CLI command structure (`src/cli/commands.ts`)
- ✅ Core service implementations (detector, search-trigger, automation)

**Verification Results:**
- ✅ No TODO/FIXME/HACK/PLACEHOLDER comments found in source code
- ✅ 186 test cases across 10 test files (all passing/conditional)
- ✅ No skipped/incomplete tests (only conditional integration tests requiring live servers)
- ✅ All core CLI features operational as documented
- ✅ Encryption implementation verified (AES-256-GCM in `src/lib/crypto.ts`)
- ✅ Package dependencies minimal and appropriate for CLI-only implementation

**Gap Confirmation:**
1. ❌ **Web Frontend**: Directories `src/web/` and `ui/` do not exist; no web dependencies in package.json
2. ❌ **Search Limits**: `SearchLimits` interface has 2 properties (lines 67-70 in types.ts), spec requires 4
3. ⚠️ **Dry-run Flag**: `scan` command exists (line 312 in commands.ts), but `run --dry-run` not implemented

---

## Critical Gaps (High Priority)

### 1. ⚠️ Web Frontend - BACKEND COMPLETE, FRONTEND NOT IMPLEMENTED
**Specification:** `specs/web-frontend.md` (comprehensive 997-line spec)
**Current State:** Backend API complete (`src/web/` implemented), Frontend not started (`ui/` directory missing)
**Package Dependencies:** No frontend dependencies yet (no React, Vite, MUI, React Router in package.json)
**Impact:** Backend operational, but web UI missing

**✅ Completed Implementation (Phase 2.1: Backend API Foundation):**
- **Backend API** (`src/web/`) - **COMPLETE** (2026-01-16):
  - ✅ REST API server with Bun's native HTTP server (`src/web/server.ts`)
  - ✅ WebSocket server for real-time log streaming (`src/web/websocket.ts`)
  - ✅ API endpoints for configuration (`src/web/routes/config.ts`)
  - ✅ API endpoints for server management (`src/web/routes/servers.ts`)
  - ✅ API endpoints for logs (`src/web/routes/logs.ts`)
  - ✅ API endpoints for automation control (`src/web/routes/automation.ts`)
  - ✅ API endpoints for statistics (`src/web/routes/stats.ts`)
  - ✅ CORS support for development
  - ✅ Error handling and validation
  - ✅ Integration with existing services (DatabaseManager, automation, logger)
  - ✅ CLI command: `janitarr serve [--port] [--host]`
  - ✅ WebSocket log broadcasting integrated into logger
  - ✅ DatabaseManager extended with query methods (getLogsPaginated, getServerStats, getSystemStats)
  - ✅ All TypeScript compilation passing
  - ✅ All ESLint validation passing
  - ✅ All 144 tests passing

**❌ Remaining Implementation (Phase 2.2-2.5: Frontend Application):**
- **Frontend React Application** (`ui/`) - **NOT STARTED**:
  - React 18+ with TypeScript, Vite 5+ build system
  - Material-UI v6 (Material Design 3 Expressive theme)
  - React Router v6 for navigation
  - Four main views: Dashboard, Servers, Logs, Settings
  - Real-time WebSocket integration
  - Dark/light theme support

**Acceptance Criteria for Remaining Work:**
- Dashboard view with status cards, server list, activity timeline, quick actions
- Servers view with list/card toggle, add/edit dialogs, statistics
- Logs view with search, filters, virtualized scrolling, real-time streaming
- Settings view with automation, limits, web interface, advanced sections
- Frontend integration with completed REST and WebSocket API endpoints
- Mobile responsive (≥320px width)
- WCAG 2.1 Level AA accessibility compliance

**Dependencies:** Backend API complete ✅ (ready to start frontend)
**Estimated Complexity:** Very High (largest remaining feature)

---

### 2. ✅ Search Limits Granularity - COMPLETE
**Specification:** `specs/search-triggering.md` lines 24-29, 94-98
**Status:** Implemented and tested
**Completed:** 2026-01-15

**Implementation:**
```typescript
interface SearchLimits {
  missingMoviesLimit: number;     // Radarr missing only
  missingEpisodesLimit: number;   // Sonarr missing only
  cutoffMoviesLimit: number;      // Radarr cutoff only
  cutoffEpisodesLimit: number;    // Sonarr cutoff only
}
```

**Files Modified:**
- ✅ `src/types.ts` - Updated SearchLimits interface to 4 properties
- ✅ `src/storage/database.ts` - Updated config defaults, added migration logic
- ✅ `src/services/search-trigger.ts` - Updated distributeItems() to filter by MediaItemType
- ✅ `src/cli/commands.ts` - Updated config commands for 4 separate limits
- ✅ `src/cli/formatters.ts` - Updated config display to show 4 limits
- ✅ `tests/services/search-trigger.test.ts` - Updated tests for 4-limit structure
- ✅ `tests/services/automation.test.ts` - Updated tests for 4-limit structure
- ✅ `tests/storage/database.test.ts` - Added migration test, updated config tests

**Acceptance Criteria Met:**
- ✅ User can set 4 independent limits via CLI (limits.missing.movies, limits.missing.episodes, limits.cutoff.movies, limits.cutoff.episodes)
- ✅ Limits apply separately by content type (movies vs episodes)
- ✅ Search distribution filters items by type before applying limits
- ✅ Example verified: missing movies limit 10 + missing episodes limit 10 = up to 20 total searches
- ✅ Backward compatibility: Old config keys (limits.missing, limits.cutoff) automatically migrate to new granular keys
- ✅ All 144 tests passing
- ✅ TypeScript compilation successful
- ✅ ESLint validation passed

---

## Minor Gaps (Low Priority)

### 3. ⚠️ Dry-Run Flag on Run Command
**Specification:** `specs/search-triggering.md` line 147, `specs/automatic-scheduling.md` line 72
**Current State:** Only `scan` command exists at commands.ts:312 (calls `detectAll()` only, no search triggering)
**Gap:** `run --dry-run` flag not implemented (commands.ts:331 accepts only `--json` option)
**Impact:** Very Low - `scan` command already provides equivalent functionality

**Options:**
1. **Add `--dry-run` flag to `run` command** (matches spec exactly)
2. **No action needed** - `scan` is functionally equivalent and may be clearer
3. **Alias `scan` to `run --dry-run`** (backwards compatibility)

**Recommendation:** Low priority - current `scan` command is clearer and achieves the same goal. Only implement if strict spec compliance required or if users specifically request `run --dry-run` flag.

**Files to Modify (if implementing option 1):**
- `src/cli/commands.ts` - Add `--dry-run` option to run command
- Update automation.ts or search-trigger.ts to support dry-run mode (skip actual searches)

**Acceptance Criteria:**
- `janitarr run --dry-run` performs detection only, no searches triggered
- Output clearly indicates dry-run/preview mode
- No log entries created for "searches" in dry-run mode
- Functionally identical to existing `scan` command

**Dependencies:** None
**Estimated Complexity:** Low (simple flag addition)

---

## Completed Features ✅

The following specifications are **fully implemented** with no gaps identified:

### ✅ Server Configuration
**Spec:** `specs/server-configuration.md`
**Implementation:** `src/services/server-manager.ts`, `src/storage/database.ts`
**Status:** Complete including encryption at rest (AES-256-GCM)

**Features:**
- Add/edit/remove Radarr and Sonarr servers
- URL normalization and validation
- Connection testing with 10-15 second timeout
- API key encryption at rest with machine-specific key
- Unique server name enforcement
- Masked API key display
- CLI commands: `server add`, `server edit`, `server remove`, `server test`, `server list`

---

### ✅ Missing Content Detection
**Spec:** `specs/missing-content-detection.md`
**Implementation:** `src/services/detector.ts`, `src/lib/api-client.ts`
**Status:** Complete

**Features:**
- Query Radarr for missing monitored movies
- Query Sonarr for missing monitored episodes
- Aggregate results across all servers
- Graceful handling of single server failures
- Server-side filtering (monitored only)
- API pagination support
- CLI command: `scan`

---

### ✅ Quality Cutoff Detection
**Spec:** `specs/quality-cutoff-detection.md`
**Implementation:** `src/services/detector.ts`, `src/lib/api-client.ts`
**Status:** Complete

**Features:**
- Query Radarr for movies below quality cutoff
- Query Sonarr for episodes below quality cutoff
- Aggregate results across all servers
- Graceful handling of single server failures
- Server-side filtering (monitored only, below cutoff)
- API pagination support
- CLI command: `scan`

---

### ✅ Search Triggering
**Spec:** `specs/search-triggering.md`
**Implementation:** `src/services/search-trigger.ts`, `src/lib/api-client.ts`
**Status:** Complete

**Features Implemented:**
- Trigger searches for missing movies up to limit
- Trigger searches for missing episodes up to limit
- Trigger searches for cutoff movies up to limit
- Trigger searches for cutoff episodes up to limit
- Fair round-robin distribution across servers
- Content-type filtering (movies vs episodes)
- Search failure logging with detailed errors
- Failed searches don't count against limit
- CLI commands: `run`, `config set limits.missing.movies`, `config set limits.missing.episodes`, `config set limits.cutoff.movies`, `config set limits.cutoff.episodes`
- Automatic migration from old 2-limit config to new 4-limit config

---

### ✅ Automatic Scheduling
**Spec:** `specs/automatic-scheduling.md`
**Implementation:** `src/lib/scheduler.ts`, `src/services/automation.ts`
**Status:** Complete

**Features:**
- Configurable interval (minimum 1 hour, default 6 hours)
- Background daemon with persistent schedule across restarts
- Full automation cycle: detect + trigger searches
- Manual trigger without affecting schedule
- Prevents concurrent cycles
- Status display with time until next run
- Cycle start/end logging with summary
- CLI commands: `start`, `stop`, `status`, `run`, `config set schedule.interval`, `config set schedule.enabled`

---

### ✅ Activity Logging
**Spec:** `specs/activity-logging.md`
**Implementation:** `src/lib/logger.ts`, `src/storage/database.ts`
**Status:** Complete

**Features:**
- Individual search entries with timestamp, server, category, item details
- Automation cycle events (start, completion with summary)
- Manual vs scheduled cycle distinction
- Server connection failure logging
- Failed search trigger logging
- Reverse chronological display (newest first)
- 30-day automatic log retention with purge
- Manual clear with confirmation
- Lightweight SQLite storage with UUID primary keys
- CLI command: `logs` (with `--limit`, `--all`, `--json`, `--clear` options)

---

## Code Quality Assessment ✅

**Strengths:**
- ✅ Strict TypeScript with no implicit any
- ✅ Consistent Result<T> pattern for error handling
- ✅ Clear separation of concerns (lib, services, CLI, storage)
- ✅ Proper async/await throughout
- ✅ Comprehensive error handling with descriptive messages
- ✅ Encryption of sensitive data at rest
- ✅ API key masking in display output
- ✅ No TODO/FIXME/placeholder code found
- ✅ Graceful partial failures in automation
- ✅ Kebab-case files, verb-based service functions
- ✅ Type definitions separate from implementation
- ✅ Database row types separate from domain types
- ✅ Singleton pattern for database consistency
- ✅ Test suite exists (10 test files in tests/ directory)

**Observations:**
- No database migration system (schema versioning)
- No webhook/notification integrations
- No cron-style advanced scheduling
- No per-server search limits (only global limits)

---

## Testing Status

**Test Files Found:**
```
tests/lib/api-client.test.ts
tests/lib/crypto.test.ts
tests/lib/logger.test.ts
tests/lib/scheduler.test.ts
tests/services/automation.test.ts
tests/services/detector.test.ts
tests/services/search-trigger.test.ts
tests/services/server-manager.test.ts
tests/storage/database.test.ts
tests/integration/api-client.integration.test.ts
```

**Coverage:** Not assessed (run `bun test` to verify)
**Recommendation:** Verify all tests pass before implementing gaps

---

## Implementation Priority

### Phase 1: Core Completeness (High Priority)
**Goal:** Complete CLI functionality to match all specifications exactly

1. **Search Limits Granularity** (Gap #2)
   - Update SearchLimits interface to 4 separate limits
   - Modify database schema and config handling
   - Update search distribution logic to filter by server type
   - Update CLI commands and display formatters
   - Add database migration for existing installations
   - Update tests to cover 4-limit scenarios

**Success Criteria:**
- All tests pass with new 4-limit structure
- Existing installations migrate seamlessly
- CLI allows independent control of movie and episode limits
- Search distribution correctly applies limits by server type

**Estimated Effort:** 1-2 days

---

### Phase 2: Web Frontend (Very High Priority)
**Goal:** Implement complete web UI as specified in `specs/web-frontend.md`

**Phase 2.1: Backend API Foundation**
- Create `src/web/` directory structure
- Implement REST API server with Bun's HTTP server
- Add all REST endpoints (config, servers, logs, automation, stats)
- Implement WebSocket server for log streaming
- Add middleware (CORS, error handling, request logging)
- Add CLI command: `janitarr serve [--port]`
- Update DatabaseManager with query methods for web API

**Phase 2.2: Frontend Project Setup**
- Initialize React + Vite project in `ui/` directory
- Install dependencies (MUI, React Router, etc.)
- Set up project structure (components, hooks, services, types)
- Configure Material Design 3 theme
- Implement routing structure

**Phase 2.3: Core Views Implementation**
- Layout components (AppBar, NavDrawer, Layout)
- Dashboard view (status cards, server list, activity timeline)
- Servers view (list/card toggle, add/edit dialogs, statistics)
- Logs view (search, filters, virtualization, real-time streaming)
- Settings view (automation, limits, web config, advanced)

**Phase 2.4: Real-time Features & Polish**
- WebSocket integration with auto-reconnect
- API integration with error handling
- Real-time dashboard updates
- Performance optimizations (lazy loading, virtualization, memoization)
- Mobile responsiveness testing (≥320px width)
- Accessibility (ARIA, keyboard navigation, WCAG 2.1 Level AA)

**Phase 2.5: Testing & Documentation**
- Backend API unit tests
- Backend integration tests
- Frontend component tests
- E2E tests (Playwright or Cypress)
- Documentation with screenshots and user guides
- Production build configuration

**Success Criteria:**
- All acceptance criteria in `specs/web-frontend.md` met
- Dashboard displays real-time status and recent activity
- Servers CRUD operations fully functional
- Logs streaming with WebSocket, search, and filters working
- Settings changes persist and take effect
- Mobile responsive and accessible
- All tests passing
- Production build optimized (<2s page load)

**Estimated Effort:** 3-4 weeks

---

### Phase 3: Polish & Enhancements (Optional)
**Goal:** Address minor gaps and potential improvements

1. **Dry-Run Flag** (Gap #3) - Only if desired
   - Add `--dry-run` flag to `run` command
   - Update documentation to clarify `scan` vs `run --dry-run`

2. **Database Migrations**
   - Add schema versioning system
   - Implement migration runner for future updates

3. **Enhanced Testing**
   - Verify test coverage is >80%
   - Add E2E CLI tests
   - Add performance/load tests for API

4. **Documentation**
   - API documentation (OpenAPI/Swagger)
   - Architecture diagrams
   - Deployment guides

**Success Criteria:**
- Optional enhancements based on user feedback
- Production-ready deployment documentation

**Estimated Effort:** 1-2 weeks

---

## Technical Debt & Future Enhancements

**Not Blocking Current Implementation:**
- Multi-user authentication (mentioned in web-frontend.md as future)
- HTTPS/TLS support (localhost-only in v1)
- Webhook notifications for external integrations
- Advanced scheduling (cron expressions)
- Per-server search limits (currently global only)
- Historical statistics and charts
- Database backup/restore functionality
- Plugin/extension system
- Multi-language support (i18n)
- Native mobile apps
- Rate limiting for API endpoints
- Browser push notifications

---

## Verification Checklist

Before marking any gap as complete, verify:

- [ ] All acceptance criteria in relevant spec are met
- [ ] All existing tests pass (`bun test`)
- [ ] New functionality has test coverage
- [ ] TypeScript compilation succeeds (`bunx tsc --noEmit`)
- [ ] Linting passes (`bunx eslint .`)
- [ ] Manual testing confirms expected behavior
- [ ] Documentation updated (if applicable)
- [ ] Database migrations tested (if schema changed)
- [ ] Backward compatibility maintained (if applicable)

---

## Next Actions

### Immediate (Ready to Start)

1. **✅ COMPLETED: Gap #2 - Search Limits Granularity** (2026-01-15)
   - ✅ Updated type definitions in `src/types.ts`
   - ✅ Updated database schema and added migration logic
   - ✅ Modified search distribution algorithm to filter by content type
   - ✅ Updated CLI commands and formatters
   - ✅ All tests passing (144/144)

2. **✅ COMPLETED: Gap #1 Phase 2.1 - Backend API Foundation** (2026-01-16)
   - ✅ Created `src/web/` directory structure
   - ✅ Implemented all REST API endpoints (config, servers, logs, automation, stats)
   - ✅ Added WebSocket log streaming with real-time broadcasting
   - ✅ Integrated WebSocket broadcasting into logger
   - ✅ Extended DatabaseManager with web API query methods
   - ✅ Added `janitarr serve` CLI command
   - ✅ All TypeScript compilation passing
   - ✅ All ESLint validation passing
   - ✅ All 144 tests passing

3. **TODO: Gap #1 Phase 2.2-2.5 - Frontend React Application** (Largest remaining effort)
   - Initialize React + Vite project in `ui/` directory
   - Install and configure MUI, React Router, dependencies
   - Implement Layout components (AppBar, NavDrawer, Layout)
   - Implement Dashboard view with real-time WebSocket updates
   - Implement Servers view with CRUD operations
   - Implement Logs view with streaming and filtering
   - Implement Settings view with config management
   - Add E2E tests and documentation

### Follow-Up
- Address Gap #3 (dry-run flag) only if strict spec compliance required
- Consider technical debt items based on user feedback
- Monitor production usage for performance bottlenecks

---

## Summary

Janitarr has a **solid, production-ready CLI implementation** with excellent code quality and architecture.

**Phase 1 Status:** ✅ **COMPLETE**
- ✅ Search Limits Granularity implemented (2026-01-15)
- ✅ CLI is now 100% spec-compliant (excluding optional dry-run flag)
- ✅ All 144 tests passing
- ✅ Full backward compatibility with automatic migration

**Phase 2.1 Status:** ✅ **COMPLETE** (2026-01-16)
- ✅ Backend API Foundation implemented
- ✅ All REST API endpoints operational (config, servers, logs, automation, stats)
- ✅ WebSocket log streaming with real-time broadcasting
- ✅ DatabaseManager extended with web query methods
- ✅ `janitarr serve` CLI command added
- ✅ All TypeScript, ESLint, and tests passing (144/144)

**Remaining Work:**
1. **Web Frontend React Application** (Phase 2.2-2.5) - Frontend UI implementation required

Once the frontend React application is implemented, Janitarr will be **100% feature-complete** according to all specifications.

**Next Steps:**
1. ✅ **DONE:** Search limits granularity (Phase 1)
2. ✅ **DONE:** Backend API Foundation (Phase 2.1)
3. **TODO:** Implement frontend React application (Phase 2.2-2.5)
4. **OPTIONAL:** Polish with optional enhancements (Phase 3)

---

---

## Additional Code Review Findings

### Architecture & Structure ✅
- **Service Layer**: Clean separation between CLI, services, and storage
- **Lib Directory**: 4 utility modules (api-client, crypto, logger, scheduler) - all complete
- **Type Safety**: Strict TypeScript with no implicit any, all types in `src/types.ts`
- **Error Handling**: Consistent Result<T> pattern throughout services
- **Database**: SQLite with singleton pattern (DatabaseManager in storage/database.ts)

### Implementation Quality Indicators ✅
- **No Placeholders**: Zero TODO/FIXME/HACK comments found in codebase
- **Test Coverage**: 186 test cases across 10 test files
  - Unit tests for all lib/ and services/ modules
  - Integration tests for API client (conditional on live servers)
  - No skipped/incomplete tests (only conditional .skipIf for integration tests)
- **Security**: AES-256-GCM encryption for API keys at rest (crypto.ts)
- **CLI**: Commander.js with consistent error handling and user feedback
- **Concurrency**: Proper use of Promise.all for parallel server operations

### Key Implementation Patterns Verified
1. **Server Operations**: Round-robin distribution in search-trigger.ts:36-92
2. **Graceful Degradation**: Partial failures don't stop automation cycles
3. **Logging**: Comprehensive activity logging with 30-day retention
4. **Config Management**: Key-value store in SQLite config table
5. **API Client**: Factory pattern with type-safe Radarr/Sonarr clients

### Dependencies (package.json)
**Production**: chalk (5.6.2), commander (14.0.2)
**Dev**: Bun types, TypeScript, ESLint with TypeScript support
**Notable**: No web framework dependencies (confirms web frontend gap)

---

**Document Status:** ✅ Complete, Validated, and Ready for Implementation
**Last Reviewed:** 2026-01-15
**Code Review Completed:** 2026-01-15
