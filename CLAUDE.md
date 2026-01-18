# Janitarr - AI Agent Instructions

Janitarr is an automation tool for Radarr and Sonarr media servers, written in Go. It automates content discovery and search triggering with configurable schedules and limits.

## Development Environment

The project uses [devenv](https://devenv.sh) with direnv for automatic environment loading.

**First-time setup:**

```bash
direnv allow                          # Authorize the development environment
```

This provides Go, templ, Tailwind CSS, and Playwright. The environment loads automatically when entering the project directory.

## Build Commands

```bash
# Generate templ templates and build
make build

# Run the application
./janitarr --help
./janitarr start                      # Production mode
./janitarr dev                        # Development mode with verbose logging

# Generate templates only
templ generate

# Build Tailwind CSS
npx tailwindcss -i ./static/css/input.css -o ./static/css/app.css
```

## Test Commands

Run these after making changes:

```bash
# Run all Go tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./src/crypto/...
go test ./src/database/...
go test ./src/api/...
go test ./src/services/...

# Run E2E tests (requires running server)
bunx playwright test --headless
bunx playwright show-report           # View test report
```

**Before running E2E tests**, start the server:

```bash
./janitarr start                      # Terminal 1
bunx playwright test --headless       # Terminal 2
```

## Code Style

### Go Conventions

- **Package naming**: lowercase, single word when possible (`crypto`, `database`, `api`)
- **File naming**: lowercase with underscores (`server_manager.go`, `api_client.go`)
- **Exports**: Only export what needs to be public; prefer unexported by default
- **Error handling**: Return errors, wrap with context using `fmt.Errorf("context: %w", err)`
- **Testing**: Table-driven tests preferred, use `testify` assertions sparingly

### Error Pattern

```go
func (s *ServerManager) AddServer(name, url string) (*Server, error) {
    if name == "" {
        return nil, fmt.Errorf("server name is required")
    }

    server, err := s.db.AddServer(name, url)
    if err != nil {
        return nil, fmt.Errorf("adding server: %w", err)
    }

    return server, nil
}
```

### Project Structure

```
src/                    # All Go source code
├── main.go             # Entry point
├── cli/                # Cobra CLI commands
├── api/                # Radarr/Sonarr API clients
├── database/           # SQLite operations
├── services/           # Business logic
├── web/                # HTTP server and handlers
├── templates/          # templ HTML templates
├── logger/             # Activity logging
└── metrics/            # Prometheus metrics
static/                 # CSS and JS assets
migrations/             # SQL migration files
tests/                  # E2E tests
```

## Testing Strategy

- **Unit tests**: In `*_test.go` files alongside implementation
- **Table-driven tests**: Preferred for functions with multiple cases
- **Mock HTTP**: Use `httptest.Server` for API client tests
- **In-memory SQLite**: Use `:memory:` for database tests
- **E2E tests**: Playwright in `tests/ui/` directory

### Test Helpers

```go
// Create test database
func testDB(t *testing.T) *database.DB {
    t.Helper()
    db, err := database.New(":memory:", t.TempDir()+"/key")
    if err != nil {
        t.Fatalf("creating test db: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
```

## Database

**Location:** `./data/janitarr.db` (auto-created on first run)
**Override:** Set `JANITARR_DB_PATH` environment variable
**Driver:** modernc.org/sqlite (pure Go, no CGO)

The `data/` directory is gitignored.

## Common Workflows

**Add a new CLI command:**

1. Create `src/cli/<command>.go`
2. Add command to `src/cli/root.go`
3. Test with `go run ./src <command> --help`

**Add a new API endpoint:**

1. Create handler in `src/web/handlers/api/<handler>.go`
2. Add tests in `src/web/handlers/api/<handler>_test.go`
3. Register route in `src/web/server.go`

**Add a new page:**

1. Create template in `src/templates/pages/<page>.templ`
2. Run `templ generate`
3. Create handler in `src/web/handlers/pages/<page>.go`
4. Register route in `src/web/server.go`

**Modify database schema:**

1. Create new migration in `migrations/<number>_<name>.sql`
2. Update `src/database/database.go` migration logic
3. Update affected Go structs

## Integration Testing

Test API credentials are in `.env` (development only).
Integration tests connect to real Radarr/Sonarr instances specified in `.env`:

```bash
RADARR_URL=http://localhost:7878
RADARR_API_KEY=your-api-key
SONARR_URL=http://localhost:8989
SONARR_API_KEY=your-api-key
```

## Security Notes

- API keys are encrypted at rest using AES-256-GCM
- Encryption key stored in `data/.janitarr.key`
- Default host binding is `localhost` (prevents external access)
- No authentication in v1 - relies on network-level access control
- Never log decrypted API keys

## AI Assistant Guidelines

**Use Context7 MCP** for:

- Go standard library documentation
- Chi router patterns
- templ template syntax
- htmx attributes and patterns
- Cobra CLI patterns

**When implementing features:**

1. Write tests first (TDD approach)
2. Run `go test ./...` after changes
3. Run `templ generate` after modifying `.templ` files
4. Check for race conditions with `go test -race ./...`

**Reference files:**

- `src-ts/` - Original TypeScript implementation (reference only)
- `ui-ts/` - Original React UI (reference only)
- `specs/` - Feature specifications
- `MIGRATION_PLAN.md` - Detailed migration tasks
