## Setup

Enter the development environment:
```bash
nix develop
```

This provides all necessary dependencies including Bun runtime.

Install project dependencies:
```bash
nix develop -c bun install
```

## Build & Run

**Development mode** (with auto-reload):
```bash
nix develop -c bun run dev
```

**Production mode**:
```bash
nix develop -c bun run start
```

**CLI commands** use the pattern:
```bash
nix develop -c bun run src/index.ts <command>
```

## Validation

Run after making changes:

```bash
nix develop -c bun test              # Run test suite
nix develop -c bunx tsc --noEmit     # Type checking
nix develop -c bunx eslint .         # Linting
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
nix develop -c bun run src/index.ts server add                     # Add servers with validation
nix develop -c bun run src/index.ts config set limits.missing 10   # Configure limits
```

**Testing server connections:**
```bash
nix develop -c bun run src/index.ts server test <name>
```

**Manual automation run:**
```bash
nix develop -c bun run src/index.ts scan   # Preview what will be searched
nix develop -c bun run src/index.ts run    # Execute searches
nix develop -c bun run src/index.ts logs   # Review results
```

**Scheduler operations:**
```bash
nix develop -c bun run src/index.ts start   # Start daemon
nix develop -c bun run src/index.ts status  # Check next run time
nix develop -c bun run src/index.ts stop    # Stop daemon
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
