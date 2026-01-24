# Version information from git
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
COMMIT := `git rev-parse --short HEAD 2>/dev/null || echo "unknown"`
BUILD_DATE := `date -u '+%Y-%m-%dT%H:%M:%SZ'`

# Build flags
LDFLAGS := "-s -w \
  -X github.com/edrobertsrayne/janitarr/src/version.Version=" + VERSION + " \
  -X github.com/edrobertsrayne/janitarr/src/version.Commit=" + COMMIT + " \
  -X github.com/edrobertsrayne/janitarr/src/version.BuildDate=" + BUILD_DATE

# Generate templ templates and Tailwind CSS
generate:
  templ generate
  ./node_modules/.bin/tailwindcss -i ./static/css/input.css -o ./static/css/app.css

# Build binary with version information
build: generate
  go build -ldflags {{LDFLAGS}} -o janitarr ./src

# Build and run dev server (accessible from all interfaces for Playwright testing)
dev: build
  ./janitarr dev --host 0.0.0.0

# Run both unit tests (Go) and E2E tests (Playwright)
test:
  @echo "Running Go unit tests with race detection..."
  go test -race ./...
  @echo "\nRunning Playwright E2E tests..."
  bunx playwright test

# Build with Nix
nix-build:
  nix build .#app
  @echo "Binary available at: result/bin/janitarr"
