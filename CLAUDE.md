## Setup

The development environment uses [devenv](https://devenv.sh) with direnv for automatic environment loading.

**First-time setup:**
```bash
direnv allow  # Authorize the development environment
```

This provides all necessary dependencies including Bun runtime. The environment loads automatically when entering the project directory.

**Install project dependencies:**
```bash
bun install
```

## Build & Run

**Development mode** (with auto-reload):
```bash
bun run dev
```

**Production mode**:
```bash
bun run start
```

**CLI commands** use the pattern:
```bash
bun run src/index.ts <command>
```

## Validation

Run after making changes:

```bash
bun test              # Run test suite
bunx tsc --noEmit     # Type checking
bunx eslint .         # Linting
```

## Test Environment

Test API credentials are in `.env` (development only, not for production).
Integration tests connect to real Radarr/Sonarr instances specified in `.env`.

## Database

**Location:** `./data/janitarr.db` (auto-created on first run)
**Override:** Set `JANITARR_DB_PATH` environment variable

The `data/` directory is gitignored.

## Common Workflows

**First-time setup:**
```bash
bun run src/index.ts server add                     # Add servers with validation
bun run src/index.ts config set limits.missing 10   # Configure limits
```

**Testing server connections:**
```bash
bun run src/index.ts server test <name>
```

**Manual automation run:**
```bash
bun run src/index.ts scan   # Preview what will be searched
bun run src/index.ts run    # Execute searches
bun run src/index.ts logs   # Review results
```

**Scheduler operations:**
```bash
bun run src/index.ts start   # Start daemon
bun run src/index.ts status  # Check next run time
bun run src/index.ts stop    # Stop daemon
```

## Code Standards

**Type Safety:**
- All types defined in `src/types.ts`
- Strict TypeScript with no implicit `any`
- Database row types separate from domain types

**Result Pattern:**
Services return typed result objects:
```typescript
{ success: boolean, data?: T, error?: string }
```

**Error Handling:**
- Validation errors return early with descriptive messages
- API failures logged but don't crash the application
- Partial results when some operations fail

**Naming Conventions:**
- Services: Verb-based functions (`addServer`, `detectAll`, `triggerSearches`)
- Types: Noun-based interfaces (`ServerConfig`, `DetectionResult`, `LogEntry`)
- Files: Lowercase kebab-case (`server-manager.ts`, `api-client.ts`)

**Testing:**
- Unit tests for pure logic (validation, formatting, utilities)
- Integration tests for API client (real server connections)
- Mock-free where possible (use in-memory SQLite for tests)
