# Janitarr - AI Agent Instructions

Janitarr is an automation tool for Radarr/Sonarr media servers, written in Go.

## Environment Setup

```bash
direnv allow  # First-time only - loads Go, templ, Tailwind, Playwright
```

## Build & Run

```bash
make build              # Generate templates + build binary
./janitarr start        # Production mode
./janitarr dev          # Development mode (verbose logging)
```

### Running with Playwright UI Testing

When using Claude Code's Playwright MCP to test the web interface, the dev server must be accessible to the Playwright container. Use the `--host` flag to bind to all network interfaces:

```bash
# Get the host's network IP address
HOST_IP=$(ip a | grep -oP '(?<=inet\s)\d+\.\d+\.\d+\.\d+' | grep -v '^127' | head -1)

# Run dev server with correct host binding
./janitarr dev --host 0.0.0.0

# Then navigate to it in Playwright (in Claude Code):
# http://$HOST_IP:3434
```

**Important**: Always use `--host 0.0.0.0` when testing with Playwright, as Docker containers cannot reach `localhost`. The default `--host localhost` only binds to the loopback interface.

## Validation

Run these after implementing to get immediate feedback:

```bash
go test ./...           # All tests
go test -race ./...     # Race detection
templ generate          # After .templ changes
```

## Operational Notes

- **Database**: `./data/janitarr.db` (override: `JANITARR_DB_PATH`)
- **Driver**: modernc.org/sqlite (pure Go, no CGO)
- **API keys**: Encrypted at rest (AES-256-GCM), never log decrypted

## Code Patterns

- **Errors**: Wrap with context: `fmt.Errorf("context: %w", err)`
- **Tests**: Table-driven, use `httptest.Server` for API mocks, `:memory:` for DB
- **Exports**: Prefer unexported by default
