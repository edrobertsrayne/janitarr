## Build & Run

### Setup

**With Nix:**
```bash
nix develop
```

**Without Nix:**
- Install [Bun](https://bun.sh/) manually
- Install dependencies: `bun install`

### Running the Project

- **Development mode** (with auto-reload): `bun run dev`
- **Production mode**: `bun run start`

## Validation

Run these after implementing to get immediate feedback:

- Tests: `bun test`
- Typecheck: `bunx tsc --noEmit`
- Lint: `bunx eslint .`

Test API details can be found in @.env

## Operational Notes

Succinct learnings about how to RUN the project:

### Running Commands

All CLI commands use the pattern: `bun run src/index.ts <command>`

For development with auto-reload: `bun run dev`

### First-Time Setup

1. Run `bun install` to install dependencies
2. Database is auto-created at `./data/janitarr.db` on first run
3. Add servers: `bun run src/index.ts server add`
4. Configure limits: `bun run src/index.ts config set limits.missing 10`

### Common Workflows

**Testing a server connection:**
```bash
bun run src/index.ts server add  # Add server with validation
bun run src/index.ts server test <name>  # Test existing server
```

**Manual automation run:**
```bash
bun run src/index.ts scan  # Preview what will be searched
bun run src/index.ts run   # Execute searches
bun run src/index.ts logs  # Review results
```

**Scheduler workflow:**
```bash
bun run src/index.ts start  # Start daemon
bun run src/index.ts status # Check next run time
bun run src/index.ts stop   # Stop daemon
```

### Database Location

Default: `./data/janitarr.db`
Override: Set `JANITARR_DB_PATH` environment variable

The `data/` directory is gitignored and created automatically.

### Test Environment

Test API credentials are in `.env` (development only, not for production).
Integration tests connect to real Radarr/Sonarr instances for validation.

### Codebase Patterns

**Architecture: Layered Service Model**

```
CLI Layer (commands.ts)
    ↓
Service Layer (server-manager, detector, search-trigger, automation)
    ↓
Library Layer (api-client, logger, scheduler)
    ↓
Storage Layer (database.ts)
```

**Result Pattern:**
Services return typed result objects: `{ success: boolean, data?: T, error?: string }`
This pattern enables graceful error handling without exceptions.

**Singleton Pattern:**
- Database: `getDatabase()` returns shared instance
- Scheduler: Single global scheduler with `start()`/`stop()`/`isRunning()`

**Dependency Injection:**
Services accept dependencies explicitly (e.g., `detectAll()` queries servers via `getDatabase()`)

**Type Safety:**
- All types defined in `src/types.ts`
- Strict TypeScript with no implicit `any`
- Database row types separate from domain types (conversion in `database.ts`)

**Error Handling:**
- Validation errors return early with descriptive messages
- API failures logged but don't crash the application
- Partial results returned when some operations fail (e.g., one server fails, others continue)

**Testing Strategy:**
- Unit tests for pure logic (validation, formatting, utilities)
- Integration tests for API client (real server connections)
- Mock-free where possible (use in-memory SQLite for tests)

**Naming Conventions:**
- Services: Verb-based functions (`addServer`, `detectAll`, `triggerSearches`)
- Types: Noun-based interfaces (`ServerConfig`, `DetectionResult`, `LogEntry`)
- Files: Lowercase kebab-case (`server-manager.ts`, `api-client.ts`)

**Database Access:**
- Raw SQL queries via Bun's SQLite driver
- Type-safe result mapping from row types to domain types
- Automatic schema initialization on first run
