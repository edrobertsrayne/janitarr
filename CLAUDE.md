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
