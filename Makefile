.PHONY: build test generate nix-build

# Version information from git
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build flags
LDFLAGS = -s -w \
	-X github.com/edrobertsrayne/janitarr/src/version.Version=$(VERSION) \
	-X github.com/edrobertsrayne/janitarr/src/version.Commit=$(COMMIT) \
	-X github.com/edrobertsrayne/janitarr/src/version.BuildDate=$(BUILD_DATE)

generate:
	templ generate
	./node_modules/.bin/tailwindcss -i ./static/css/input.css -o ./static/css/app.css

build: generate
	go build -ldflags "$(LDFLAGS)" -o janitarr ./src

test:
	go test -race ./...

nix-build:
	nix build .#app
	@echo "Binary available at: result/bin/janitarr"
