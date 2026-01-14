# Janitarr Implementation Plan

Automation tool for managing Radarr and Sonarr media servers. Detects missing content and quality upgrades, then triggers searches on a configurable schedule.

## Project Status: In Progress

**Last Updated:** 2026-01-14

### Technology Decisions (Confirmed)
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Runtime | Bun | Fast, TypeScript-native, built-in SQLite |
| Storage | SQLite via `bun:sqlite` | ACID compliance, no external deps, built into Bun |
| Interface | CLI first | Core functionality first, web UI can be added in Phase 7 |
| Deployment | Docker-ready | Environment variables for config |
| Instance model | Single-instance | Simplifies state management and scheduling |

### Validation Commands
```bash
bun test          # Run tests
bunx tsc --noEmit # Type check
bunx eslint .     # Lint
```

### Test Environment
Test API credentials available in `.env` (Radarr at thor:7878, Sonarr at thor:8989)

### Progress Overview
| Phase | Status | Completion |
|-------|--------|------------|
| Phase 1: Project Foundation | Complete | 100% |
| Phase 2: Server Configuration | Complete | 100% |
| Phase 3: Content Detection | Not Started | 0% |
| Phase 4: Search Triggering | Not Started | 0% |
| Phase 5: Activity Logging | Not Started | 0% |
| Phase 6: Automatic Scheduling | Not Started | 0% |
| Phase 7: User Interface | Not Started | 0% |

### Immediate Next Steps
Execute Phase 3 in order:
1. **Create `src/services/detector.ts`** - Missing and cutoff content detection
2. **Test detection against real Radarr/Sonarr servers**

---

## Architecture Overview

```
janitarr/
├── src/
│   ├── lib/                    # Shared utilities (project standard library)
│   │   ├── api-client.ts       # Radarr/Sonarr API client
│   │   ├── config.ts           # Configuration management
│   │   ├── logger.ts           # Activity logging system
│   │   └── scheduler.ts        # Scheduling utilities
│   ├── services/
│   │   ├── server-manager.ts   # Server CRUD operations
│   │   ├── detector.ts         # Missing/cutoff detection
│   │   ├── search-trigger.ts   # Search triggering logic
│   │   └── automation.ts       # Automation cycle orchestration
│   ├── storage/
│   │   └── database.ts         # Persistent storage (SQLite/JSON)
│   ├── ui/                     # Web interface (if applicable)
│   └── index.ts                # Application entry point
├── specs/                      # Requirements documentation
└── tests/                      # Test files
```

### Dependency Graph

```
Server Configuration ─────────────────────────────┐
        │                                         │
        ▼                                         ▼
Missing Content Detection    Quality Cutoff Detection
        │                           │
        └───────────┬───────────────┘
                    ▼
            Search Triggering
                    │
                    ▼
           Automatic Scheduling ◄── Activity Logging (cross-cutting)
```

---

## Phase 1: Project Foundation

### 1.1 Project Setup
- [x] Create `package.json` with Bun runtime configuration:
  ```json
  {
    "name": "janitarr",
    "version": "0.1.0",
    "type": "module",
    "scripts": {
      "start": "bun run src/index.ts",
      "dev": "bun --watch run src/index.ts",
      "test": "bun test",
      "typecheck": "bunx tsc --noEmit",
      "lint": "bunx eslint ."
    },
    "devDependencies": {
      "@types/bun": "latest",
      "typescript": "^5.0.0",
      "eslint": "^8.0.0",
      "@typescript-eslint/eslint-plugin": "^6.0.0",
      "@typescript-eslint/parser": "^6.0.0"
    }
  }
  ```
- [x] Create `tsconfig.json` for TypeScript compilation:
  ```json
  {
    "compilerOptions": {
      "target": "ESNext",
      "module": "ESNext",
      "moduleResolution": "bundler",
      "strict": true,
      "skipLibCheck": true,
      "noEmit": true,
      "esModuleInterop": true,
      "allowSyntheticDefaultImports": true,
      "resolveJsonModule": true,
      "isolatedModules": true,
      "outDir": "./dist",
      "rootDir": "./src",
      "types": ["bun-types"]
    },
    "include": ["src/**/*", "tests/**/*"],
    "exclude": ["node_modules", "dist"]
  }
  ```
- [x] Create `.eslintrc.json` for linting rules
- [x] Create initial directory structure:
  ```
  mkdir -p src/lib src/services src/storage tests
  ```
- [x] Run `bun install` to install dependencies

### 1.2 Core Types & Schemas (`src/types.ts`)
- [x] Define `ServerType` enum (`radarr` | `sonarr`)
- [x] Define `ServerConfig` interface:
  ```typescript
  interface ServerConfig {
    id: string;           // UUID
    name: string;         // User-friendly display name
    url: string;          // Base URL (validated, normalized)
    apiKey: string;       // API key (stored encrypted)
    type: ServerType;
    createdAt: Date;
    updatedAt: Date;
  }
  ```
- [x] Define `DetectionResult` interface:
  ```typescript
  interface DetectionResult {
    serverId: string;
    serverName: string;
    serverType: ServerType;
    missingCount: number;
    cutoffCount: number;
    missingItems: MediaItem[];  // For search triggering
    cutoffItems: MediaItem[];   // For search triggering
    error?: string;             // If detection failed
  }
  ```
- [x] Define `MediaItem` interface (for search targeting):
  ```typescript
  interface MediaItem {
    id: number;           // Radarr/Sonarr internal ID
    title: string;        // For logging
    type: 'movie' | 'episode';
  }
  ```
- [x] Define `LogEntry` interface:
  ```typescript
  interface LogEntry {
    id: string;
    timestamp: Date;
    type: 'cycle_start' | 'cycle_end' | 'search' | 'error';
    serverName?: string;
    serverType?: ServerType;
    category?: 'missing' | 'cutoff';
    count?: number;
    message: string;
    isManual?: boolean;   // Manual vs scheduled trigger
  }
  ```
- [x] Define `ScheduleConfig` interface:
  ```typescript
  interface ScheduleConfig {
    intervalHours: number;  // Minimum 1
    enabled: boolean;       // false = manual only mode
  }
  ```
- [x] Define `SearchLimits` interface:
  ```typescript
  interface SearchLimits {
    missingLimit: number;   // 0 = disabled
    cutoffLimit: number;    // 0 = disabled
  }
  ```
- [x] Define `AppConfig` interface (aggregates all config):
  ```typescript
  interface AppConfig {
    schedule: ScheduleConfig;
    searchLimits: SearchLimits;
  }
  ```

---

## Phase 2: Server Configuration (specs/server-configuration.md)

**Dependency:** Phase 1 complete

### 2.1 API Client Library (`src/lib/api-client.ts`)
- [x] Implement base HTTP client with timeout handling (10-15 second timeout)
- [x] Implement URL normalization (trailing slashes, protocol validation)
- [x] Create `RadarrClient` class with authentication header injection
- [x] Create `SonarrClient` class with authentication header injection
- [x] Implement connection validation using minimal API call (system/status endpoint)
- [x] Handle API response codes: 200 (success), 401 (unauthorized), 404 (not found)
- [x] Return specific error messages for different failure modes

### 2.2 Server Storage (`src/storage/database.ts`)
- [x] Implement persistent storage for server configurations
- [x] Store API keys securely (encryption at rest if possible)
- [x] Prevent duplicate server entries (same URL + type)
- [x] Support CRUD operations for server records

### 2.3 Server Manager Service (`src/services/server-manager.ts`)

**Story: Add New Media Server**
- [x] Accept server URL and API key input
- [x] Accept server type (Radarr/Sonarr)
- [x] Validate URL format (http:// or https:// protocol required)
- [x] Test API connectivity before saving
- [x] Only save if connection test passes
- [x] Return clear error messages on failure

**Story: View Configured Servers**
- [x] Return list of all configured servers
- [x] Include: server type, URL, masked API key (first/last chars only)
- [x] Distinguish between Radarr and Sonarr servers

**Story: Edit Existing Server**
- [x] Allow modification of URL and/or API key
- [x] Re-validate connectivity before saving changes
- [x] Only apply changes if new connection test passes

**Story: Remove Server**
- [x] Remove server from configuration
- [x] Require confirmation before deletion
- [x] Server immediately excluded from automation

---

## Phase 3: Content Detection

**Dependency:** Phase 2 complete (Server Configuration)

### 3.1 Missing Content Detection (`src/services/detector.ts`) - specs/missing-content-detection.md

**Story: Detect Missing Episodes (Sonarr)**
- [ ] Query each Sonarr server for monitored episodes marked as missing
- [ ] Use API endpoint that filters server-side (not client-side filtering)
- [ ] Handle API pagination for large libraries
- [ ] Count total missing episodes across all Sonarr servers
- [ ] Only count monitored items
- [ ] Respect series/season monitoring settings

**Story: Detect Missing Movies (Radarr)**
- [ ] Query each Radarr server for monitored movies marked as missing
- [ ] Use API endpoint that filters server-side
- [ ] Handle API pagination for large libraries
- [ ] Count total missing movies across all Radarr servers
- [ ] Only count monitored items

**Story: Handle Detection Failures**
- [ ] Continue checking other servers if one fails
- [ ] Log failures with timestamp, server name, and reason
- [ ] Return partial results from successful queries

### 3.2 Quality Cutoff Detection (`src/services/detector.ts`) - specs/quality-cutoff-detection.md

**Story: Detect Episodes Below Quality Cutoff (Sonarr)**
- [ ] Query each Sonarr server for episodes below quality cutoff
- [ ] Use API endpoint that filters for cutoff-not-met server-side
- [ ] Handle API pagination
- [ ] Count total upgradeable episodes across all Sonarr servers
- [ ] Only count monitored items
- [ ] Respect quality profile cutoff settings

**Story: Detect Movies Below Quality Cutoff (Radarr)**
- [ ] Query each Radarr server for movies below quality cutoff
- [ ] Use API endpoint that filters for cutoff-not-met server-side
- [ ] Handle API pagination
- [ ] Count total upgradeable movies across all Radarr servers
- [ ] Only count monitored items

**Story: Handle Detection Failures**
- [ ] Continue checking other servers if one fails
- [ ] Log failures with timestamp, server name, and reason
- [ ] Return partial results from successful queries

---

## Phase 4: Search Triggering (specs/search-triggering.md)

**Dependency:** Phase 3 complete (Content Detection)

### 4.1 Search Limits Configuration
- [ ] Allow user to set numeric limit for missing content searches (0 or greater)
- [ ] Allow user to set separate numeric limit for cutoff searches (0 or greater)
- [ ] Persist limits across application restarts
- [ ] Setting limit to 0 disables that category

### 4.2 Search Trigger Service (`src/services/search-trigger.ts`)

**Story: Trigger Missing Content Searches**
- [ ] Trigger searches up to user-configured missing limit
- [ ] Use Radarr/Sonarr CommandController API (MoviesSearch, EpisodeSearch)
- [ ] Distribute searches fairly across all configured servers
- [ ] Log each triggered search with timestamp, server, item type
- [ ] If fewer items than limit, search all available

**Story: Trigger Quality Upgrade Searches**
- [ ] Trigger searches up to user-configured cutoff limit
- [ ] Use Radarr/Sonarr CommandController API
- [ ] Distribute searches fairly across all configured servers
- [ ] Log each triggered search with timestamp, server, item type
- [ ] If fewer items than limit, search all available

**Story: Handle Search Failures**
- [ ] Log failures with timestamp, server name, and reason
- [ ] Failed searches do not count against limit
- [ ] Continue triggering remaining searches after failure
- [ ] Implement brief delays between commands if needed for rate limiting

---

## Phase 5: Activity Logging (specs/activity-logging.md)

**Dependency:** Can be implemented alongside Phase 3-4, integrated throughout

### 5.1 Log Storage (`src/lib/logger.ts`)
- [ ] Implement persistent log storage (survives restarts)
- [ ] Write entries immediately (no buffering)
- [ ] Retain logs for at least 30 days
- [ ] Auto-purge logs older than 30 days
- [ ] Consider max log size (10MB) with rotation
- [ ] Never log API keys or credentials

### 5.2 Log Entry Types
- [ ] Log triggered searches: timestamp, server name, server type, category (missing/cutoff), count
- [ ] Log automation cycle start: timestamp
- [ ] Log automation cycle completion: timestamp, summary (total searches)
- [ ] Log server connection failures: timestamp, server name, reason
- [ ] Log search trigger failures: timestamp, server name, reason
- [ ] Mark manual triggers vs scheduled triggers
- [ ] Group related operations where sensible (e.g., "Triggered 5 missing searches on Server1" vs 5 entries)

### 5.3 Log Display
- [ ] Display logs in reverse chronological order (newest first)
- [ ] Show date and time in readable format
- [ ] Show at least 100 most recent entries
- [ ] Visually distinguish: cycle events, successful searches, failures
- [ ] Consider pagination for large logs

### 5.4 Log Management
- [ ] Allow user to manually clear all logs
- [ ] Require confirmation before clearing
- [ ] Display summary view: "Last cycle: N searches triggered, M failures"

---

## Phase 6: Automatic Scheduling (specs/automatic-scheduling.md)

**Dependency:** Phases 3-5 complete (Detection, Triggering, Logging)

### 6.1 Scheduler (`src/lib/scheduler.ts`)

**Story: Configure Schedule Frequency**
- [ ] Allow user to set time interval (e.g., 1 hour, 6 hours, daily)
- [ ] Enforce minimum interval of 1 hour
- [ ] Allow "manual only" mode (disable scheduled automation)
- [ ] Persist schedule configuration across restarts
- [ ] Changes take effect on next scheduled run

**Story: Execute Automation Cycle**
- [ ] Execute complete cycle: detect missing, detect cutoff, trigger searches
- [ ] Prevent concurrent cycles (if cycle takes longer than interval, wait)
- [ ] Continue running indefinitely until stopped
- [ ] Log cycle start and completion

**Story: Manual Trigger**
- [ ] Allow user to manually initiate cycle through interface
- [ ] Execute same cycle as scheduled automation
- [ ] Manual trigger does not affect regular schedule
- [ ] Provide feedback when cycle completes

**Story: View Schedule Status**
- [ ] Display current schedule configuration
- [ ] Show time until next scheduled run
- [ ] Update status in real-time or on refresh

### 6.2 Automation Orchestrator (`src/services/automation.ts`)
- [ ] Coordinate detection and search triggering
- [ ] Handle partial failures gracefully
- [ ] Run first cycle immediately on application startup
- [ ] Resume regular schedule after restart
- [ ] Keep UI responsive during automation (non-blocking)

---

## Phase 7: User Interface (CLI)

**Dependency:** Phases 1-6 complete (all backend services)

### 7.1 CLI Command Structure

The CLI is the primary interface. Commands follow the pattern: `janitarr <command> [subcommand] [options]`

**Server Management:**
- [ ] `janitarr server add` - Interactive prompt for URL, API key, type, name
- [ ] `janitarr server list` - Display all servers with masked API keys
- [ ] `janitarr server edit <id|name>` - Modify server configuration
- [ ] `janitarr server remove <id|name>` - Delete server (with confirmation)
- [ ] `janitarr server test <id|name>` - Test connection to specific server

**Detection & Status:**
- [ ] `janitarr status` - Show next run time, last run summary, server count
- [ ] `janitarr scan` - Run detection only (no searches), display counts

**Automation:**
- [ ] `janitarr run` - Execute full automation cycle immediately (manual trigger)
- [ ] `janitarr start` - Start daemon with scheduled automation
- [ ] `janitarr stop` - Stop running daemon gracefully

**Configuration:**
- [ ] `janitarr config show` - Display current configuration
- [ ] `janitarr config set <key> <value>` - Update config values
  - Keys: `schedule.interval`, `schedule.enabled`, `limits.missing`, `limits.cutoff`

**Activity Logs:**
- [ ] `janitarr logs` - Display recent activity (default: 50 entries)
- [ ] `janitarr logs --all` - Display all logs with pagination
- [ ] `janitarr logs --clear` - Clear all logs (with confirmation)

### 7.2 CLI Output Formatting
- [ ] Use colored output for success/failure indicators
- [ ] Table formatting for server lists and status
- [ ] Progress indicators for long-running operations
- [ ] JSON output option (`--json`) for scripting integration

---

## Configuration Strategy

### Environment Variables
| Variable | Purpose | Default |
|----------|---------|---------|
| `JANITARR_DB_PATH` | SQLite database file location | `./data/janitarr.db` |
| `JANITARR_LOG_LEVEL` | Logging verbosity (debug/info/warn/error) | `info` |

**Note:** Server credentials are stored in the SQLite database, NOT in environment variables. The `.env` file in the repository is for development/testing only.

### Database Schema (SQLite)

**servers table:**
```sql
CREATE TABLE servers (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  url TEXT NOT NULL,
  api_key TEXT NOT NULL,  -- Consider encryption
  type TEXT NOT NULL CHECK(type IN ('radarr', 'sonarr')),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(url, type)
);
```

**config table:**
```sql
CREATE TABLE config (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
-- Keys: schedule.interval, schedule.enabled, limits.missing, limits.cutoff
```

**logs table:**
```sql
CREATE TABLE logs (
  id TEXT PRIMARY KEY,
  timestamp TEXT NOT NULL,
  type TEXT NOT NULL,
  server_name TEXT,
  server_type TEXT,
  category TEXT,
  count INTEGER,
  message TEXT NOT NULL,
  is_manual INTEGER DEFAULT 0
);
CREATE INDEX idx_logs_timestamp ON logs(timestamp DESC);
```

---

## API Research Notes

### Radarr API Endpoints (v3)
- `GET /api/v3/system/status` - Connection validation
- `GET /api/v3/movie` with `?monitored=true` - Get movies
- `GET /api/v3/wanted/missing` - Missing movies (paginated)
- `GET /api/v3/wanted/cutoff` - Cutoff unmet movies (paginated)
- `POST /api/v3/command` with body `{"name": "MoviesSearch", "movieIds": [...]}` - Trigger search

### Sonarr API Endpoints (v3)
- `GET /api/v3/system/status` - Connection validation
- `GET /api/v3/series` - Get series
- `GET /api/v3/wanted/missing` - Missing episodes (paginated)
- `GET /api/v3/wanted/cutoff` - Cutoff unmet episodes (paginated)
- `POST /api/v3/command` with body `{"name": "EpisodeSearch", "episodeIds": [...]}` - Trigger search

### Authentication
Both APIs use `X-Api-Key` header for authentication.

---

## Implementation Order Summary

1. **Phase 1:** Project setup, types, schemas
2. **Phase 2:** Server configuration (API client, storage, CRUD)
3. **Phase 3:** Content detection (missing + cutoff)
4. **Phase 4:** Search triggering with limits
5. **Phase 5:** Activity logging (integrate throughout 3-4)
6. **Phase 6:** Automatic scheduling
7. **Phase 7:** User interface

---

## Technology Decisions (Resolved)

| Question | Decision | Notes |
|----------|----------|-------|
| UI Technology | CLI first, Web UI in Phase 7 | Core services designed as importable modules |
| Storage Backend | SQLite via `bun:sqlite` | ACID compliance, built into Bun, no external deps |
| Deployment | Docker-ready | Environment variables for DB path, ports |
| Multi-instance | Single-instance only | Simplifies scheduling and state management |

All technology decisions have been finalized. Implementation can proceed.

---

## Gap Analysis (2026-01-14, Updated)

### Current State
| Category | Status |
|----------|--------|
| Source code | Phase 2 complete (types, API client, database, server manager) |
| Test code | 52 tests passing (unit + integration) |
| Build config | Complete (`package.json`, `tsconfig.json`, `.eslintrc.json`) |
| Specifications | Complete (6 spec files) |
| Implementation plan | Complete - all specs mapped to phases |
| Test environment | Ready (`.env` with API credentials) |

### Project Structure

**Existing files:**
```
janitarr/
├── src/
│   ├── lib/
│   │   └── api-client.ts       # ✅ Phase 2.1 - Radarr/Sonarr API client
│   ├── services/
│   │   └── server-manager.ts   # ✅ Phase 2.3 - Server CRUD operations
│   ├── storage/
│   │   └── database.ts         # ✅ Phase 2.2 - SQLite persistence
│   ├── types.ts                # ✅ Phase 1.2 - Core type definitions
│   └── index.ts                # ✅ Phase 1.1 - Entry point stub
├── tests/
│   ├── lib/
│   │   └── api-client.test.ts  # ✅ URL normalization/validation tests
│   ├── services/
│   │   └── server-manager.test.ts  # ✅ Server manager tests
│   ├── storage/
│   │   └── database.test.ts    # ✅ Database operations tests
│   └── integration/
│       └── api-client.integration.test.ts  # ✅ Live API tests
├── specs/                      # ✅ Complete - 6 specification files
├── package.json                # ✅ Phase 1.1 - Bun runtime config
├── tsconfig.json               # ✅ Phase 1.1 - TypeScript config
├── .eslintrc.json              # ✅ Phase 1.1 - Linting config
├── IMPLEMENTATION_PLAN.md      # ✅ This file (comprehensive)
├── AGENTS.md                   # ✅ Build instructions
├── .gitignore                  # ✅ Configured
└── .env                        # ✅ Test API credentials (dev only)
```

**Files to create:**
```
janitarr/
├── package.json                # Phase 1.1
├── tsconfig.json               # Phase 1.1
├── data/                       # Created at runtime for SQLite DB
├── src/
│   ├── index.ts                # Phase 1.1 - Entry point & CLI router
│   ├── types.ts                # Phase 1.2 - Core type definitions
│   ├── lib/
│   │   ├── api-client.ts       # Phase 2.1 - HTTP client for Radarr/Sonarr
│   │   ├── config.ts           # Phase 2 - App configuration loading
│   │   ├── logger.ts           # Phase 5.1 - Activity logging
│   │   └── scheduler.ts        # Phase 6.1 - Scheduling utilities
│   ├── services/
│   │   ├── server-manager.ts   # Phase 2.3 - Server CRUD
│   │   ├── detector.ts         # Phase 3 - Missing/cutoff detection
│   │   ├── search-trigger.ts   # Phase 4.2 - Search triggering
│   │   └── automation.ts       # Phase 6.2 - Cycle orchestration
│   ├── storage/
│   │   └── database.ts         # Phase 2.2 - SQLite persistence
│   └── cli/
│       ├── commands.ts         # Phase 7 - CLI command definitions
│       └── formatters.ts       # Phase 7 - Output formatting
└── tests/                      # Mirror src/ structure
    ├── lib/
    ├── services/
    └── storage/
```

### Specification → Phase Mapping
| Spec File | Implementation Phase | Coverage |
|-----------|---------------------|----------|
| `server-configuration.md` | Phase 2 | ✅ Full |
| `missing-content-detection.md` | Phase 3.1 | ✅ Full |
| `quality-cutoff-detection.md` | Phase 3.2 | ✅ Full |
| `search-triggering.md` | Phase 4 | ✅ Full |
| `activity-logging.md` | Phase 5 | ✅ Full |
| `automatic-scheduling.md` | Phase 6 | ✅ Full |

### Test File Structure
Tests should be placed alongside source files or in a parallel `tests/` directory:
```
tests/
├── lib/
│   └── api-client.test.ts      # API client unit tests
├── services/
│   ├── server-manager.test.ts  # CRUD operation tests
│   ├── detector.test.ts        # Detection logic tests
│   └── search-trigger.test.ts  # Search triggering tests
├── storage/
│   └── database.test.ts        # SQLite operations tests
└── integration/
    └── automation.test.ts      # End-to-end cycle tests
```

### Ready to Begin
Phase 2 complete. Begin with Phase 3 (Content Detection).
