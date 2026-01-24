# Janitarr: Implementation Plan

## Overview

This document tracks implementation tasks for Janitarr, an automation tool for Radarr and Sonarr media servers written in Go.

## Agent Instructions

This document is designed for AI coding agents. Each task:

- Has a checkbox `[ ]` that should be marked `[x]` when complete
- Includes specific file paths and exact code changes
- Has clear completion criteria
- Follows established code patterns from the codebase

**Workflow for each task:**

1. Read the task completely before starting
2. Make the specified code changes
3. Run the verification commands
4. Commit with the specified message
5. Mark the checkbox `[x]`

**Environment:** Run `direnv allow` to load development tools.

## Technology Stack

| Component     | Technology          | Purpose                    |
| ------------- | ------------------- | -------------------------- |
| Language      | Go 1.22+            | Main application           |
| Web Framework | Chi (go-chi/chi/v5) | HTTP routing               |
| Database      | modernc.org/sqlite  | SQLite (pure Go, no CGO)   |
| CLI           | Cobra (spf13/cobra) | Command-line interface     |
| Templates     | templ (a-h/templ)   | Type-safe HTML templates   |
| Interactivity | htmx + Alpine.js    | Dynamic UI without React   |
| CSS           | Tailwind CSS v3     | Utility-first styling      |
| UI Components | DaisyUI v4          | Semantic component classes |

---

## Current Status

**Active Phase:** Phase 28 - Deployment Infrastructure
**Previous Phase:** Phase 27 - Spec-Code Alignment (Complete)
**Test Status:** Go unit tests passing, E2E tests 88% pass rate (63/72 passing, 9 intentionally skipped)

### Gap Analysis Summary (2026-01-24)

All critical spec-code alignment issues from Phase 27 have been resolved. The remaining work is deployment infrastructure plus one minor validation fix.

| Gap                                         | Spec File            | Severity | Status      |
| ------------------------------------------- | -------------------- | -------- | ----------- |
| Search limit upper bound validation missing | search-triggering.md | Low      | Not Started |
| /health route alias missing (Docker health) | deployment.md        | Low      | Not Started |
| Dockerfile not created                      | deployment.md        | High     | Not Started |
| docker-entrypoint.sh not created            | deployment.md        | High     | Not Started |
| flake.nix with package/module not created   | deployment.md        | High     | Not Started |
| nix/package.nix not created                 | deployment.md        | Medium   | Not Started |
| nix/module.nix not created                  | deployment.md        | Medium   | Not Started |
| .github/workflows/docker.yml not created    | deployment.md        | Medium   | Not Started |
| .github/workflows/nix.yml not created       | deployment.md        | Low      | Not Started |

**Note on /health:** The `/api/health` endpoint exists at `src/web/server.go:125`. Docker health checks expect `/health` at root level. Task 2 adds a route alias.

**Note on search limits:** HTML form enforces 0-1000 range via `min`/`max` attributes, but server-side validation in `src/web/handlers/api/config.go:134-152` only checks `>= 0`. API bypass possible.

### Implementation Completeness

Features from `/specs/` that ARE fully implemented:

- CLI Interface with interactive forms (cli-interface.md)
- Logging system with web viewer (logging.md)
- Web frontend with templ + htmx + Alpine.js (web-frontend.md)
- DaisyUI v4 integration (archived: daisyui-migration.md)
- Unified service startup (unified-service-startup.md)
- Server configuration management (server-configuration.md)
- Activity logging (activity-logging.md)
- Missing content & quality cutoff detection (missing-content-detection.md, quality-cutoff-detection.md)
- Automatic scheduling with manual triggers (automatic-scheduling.md)
- Search triggering with 4 limits, rate limiting, proportional distribution (search-triggering.md)
- Dry-run mode (search-triggering.md)
- Cycle duration monitoring (automatic-scheduling.md)
- 100ms inter-batch delay (search-triggering.md)
- 429 rate limit handling with Retry-After (search-triggering.md)
- 3-strike server skip on rate limits (search-triggering.md)
- High limit warnings (>100) (search-triggering.md)

Features NOT yet implemented:

- Docker deployment (deployment.md)
- NixOS module (deployment.md)
- CI/CD workflows (deployment.md)
- Search limit upper bound validation (search-triggering.md) - minor fix

---

## Phase 28: Deployment Infrastructure

**Status:** Not Started
**Priority:** High (required for release)
**Spec:** `specs/deployment.md`

### Task Summary

| #   | Task                         | Dependency | Effort | Priority |
| --- | ---------------------------- | ---------- | ------ | -------- |
| 1   | Add search limit upper bound | None       | Low    | Low      |
| 2   | Add /health route alias      | None       | Low    | Low      |
| 3   | Create Dockerfile            | None       | Medium | High     |
| 4   | Create docker-entrypoint.sh  | Task 3     | Low    | High     |
| 5   | Create flake.nix             | None       | Medium | High     |
| 6   | Create nix/package.nix       | Task 5     | Low    | Medium   |
| 7   | Create nix/module.nix        | Task 6     | Medium | Medium   |
| 8   | Create docker.yml workflow   | Task 3     | Low    | Medium   |
| 9   | Create nix.yml workflow      | Task 5     | Low    | Low      |

---

### Task 1: Add Search Limit Upper Bound Validation

Add server-side validation to enforce the 0-1000 range specified in search-triggering.md.

**Files to Modify:**

- `src/web/handlers/api/config.go`

**Implementation:**

In `PostConfig()`, update the search limit validation (lines 134, 140, 146, 152) to include upper bound:

```go
// Line 134: Change
if i, err := strconv.Atoi(val); err == nil && i >= 0 {
// To
if i, err := strconv.Atoi(val); err == nil && i >= 0 && i <= 1000 {

// Apply same change to lines 140, 146, 152
```

**Tests:**

- **Unit:** Add test case in `src/web/handlers/api/config_test.go` for limits > 1000 being rejected (silently ignored, keeping old value)
- **E2E:** N/A (form already prevents via HTML attributes)

**Verification:**

```bash
go test ./src/web/handlers/api/... -v -run TestConfig
```

---

### Task 2: Add /health Route Alias

Add `/health` route that mirrors `/api/health` for Docker health checks.

**Files to Modify:**

- `src/web/server.go`

**Implementation:**

Add route outside the `/api` group. Around line 121, before the `r.Route("/api", ...)` block:

```go
// Health check alias for Docker health checks
r.Get("/health", apiHandlers.Health.GetHealth)
```

**Tests:**

- **Unit:** Add test in `src/web/server_test.go` verifying `/health` returns 200 with JSON body containing `"status"`
- **E2E:** `curl http://localhost:3434/health` returns JSON with status

**Verification:**

```bash
go test ./src/web/... -v
# Manual verification after starting server:
curl -s http://localhost:3434/health | jq .status
# Should output: "ok"
```

---

### Task 3: Create Dockerfile

Create multi-stage Dockerfile with Alpine base per specs/deployment.md.

**Files to Create:**

- `Dockerfile`

**Implementation:**

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /build

# Download dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and generate templates
COPY . .
RUN templ generate
RUN go build -ldflags="-s -w" -o janitarr .

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache su-exec shadow wget

COPY --from=builder /build/janitarr /usr/local/bin/janitarr
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh

ENV PUID=1000 \
    PGID=1000 \
    JANITARR_PORT=3434 \
    JANITARR_DB_PATH=/data/janitarr.db

EXPOSE 3434

VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:${JANITARR_PORT:-3434}/health || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["janitarr", "start", "--host", "0.0.0.0"]
```

**Reference:** See specs/deployment.md lines 100-137 for exact Dockerfile structure.

**Tests:**

- **Unit:** N/A (Dockerfile not testable with Go)
- **E2E:** Manual `docker build` verification

**Verification:**

```bash
docker build -t janitarr:test .
docker run --rm janitarr:test janitarr --version
```

---

### Task 4: Create docker-entrypoint.sh

Create entrypoint script for PUID/PGID user management per specs/deployment.md.

**Files to Create:**

- `docker-entrypoint.sh`

**Implementation:**

```bash
#!/bin/sh
set -e

PUID=${PUID:-1000}
PGID=${PGID:-1000}

# Create group if it doesn't exist
if ! getent group janitarr > /dev/null 2>&1; then
    addgroup -g "${PGID}" janitarr
fi

# Create user if it doesn't exist
if ! getent passwd janitarr > /dev/null 2>&1; then
    adduser -D -u "${PUID}" -G janitarr -h /data -s /sbin/nologin janitarr
fi

# Ensure correct ownership
chown -R janitarr:janitarr /data

# Drop privileges and execute
exec su-exec janitarr:janitarr "$@"
```

**Reference:** See specs/deployment.md lines 139-163 for exact script.

**Tests:**

- **Unit:** N/A (shell script)
- **E2E:** Verify container runs as expected user

**Verification:**

```bash
docker run --rm -e PUID=1001 -e PGID=1001 janitarr:test id
# Should show uid=1001 gid=1001
```

---

### Task 5: Create flake.nix

Create Nix flake with package and module outputs per specs/deployment.md.

**Files to Create:**

- `flake.nix`

**Implementation:**

```nix
{
  description = "Janitarr - Automation tool for Radarr and Sonarr";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachSystem [ "x86_64-linux" "aarch64-linux" ] (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        janitarr = pkgs.callPackage ./nix/package.nix { };
      in
      {
        packages = {
          janitarr = janitarr;
          default = janitarr;
        };
      }
    ) // {
      nixosModules = {
        janitarr = import ./nix/module.nix;
        default = self.nixosModules.janitarr;
      };
    };
}
```

**Tests:**

- **Unit:** N/A
- **E2E:** `nix flake check` passes

**Verification:**

```bash
nix flake check
nix build .#janitarr --dry-run
```

---

### Task 6: Create nix/package.nix

Create Nix derivation for the janitarr package.

**Files to Create:**

- `nix/package.nix`

**Implementation:**

```nix
{ lib
, buildGoModule
, templ
, fetchFromGitHub
}:

buildGoModule rec {
  pname = "janitarr";
  version = "0.1.0";

  src = ./..;

  vendorHash = lib.fakeHash; # Update after first build attempt

  nativeBuildInputs = [ templ ];

  preBuild = ''
    templ generate
  '';

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "Automation tool for Radarr and Sonarr media servers";
    homepage = "https://github.com/edrobertsrayne/janitarr";
    license = licenses.mit;
    maintainers = [ ];
    mainProgram = "janitarr";
  };
}
```

**Note:** The `vendorHash` will need to be updated after the first build attempt. Nix will report the correct hash.

**Tests:**

- **Unit:** N/A
- **E2E:** `nix build` produces working binary

**Verification:**

```bash
nix build .#janitarr
./result/bin/janitarr --version
```

---

### Task 7: Create nix/module.nix

Create NixOS module with systemd service and security hardening per specs/deployment.md.

**Files to Create:**

- `nix/module.nix`

**Implementation:**

```nix
{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.services.janitarr;
in
{
  options.services.janitarr = {
    enable = mkEnableOption "Janitarr media server automation";

    port = mkOption {
      type = types.port;
      default = 3434;
      description = "Port for the web interface.";
    };

    openFirewall = mkOption {
      type = types.bool;
      default = false;
      description = "Whether to open the firewall for the web interface.";
    };

    dataDir = mkOption {
      type = types.path;
      default = "/var/lib/janitarr";
      description = "Directory for database and application data.";
    };

    logLevel = mkOption {
      type = types.enum [ "debug" "info" "warn" "error" ];
      default = "info";
      description = "Log verbosity level.";
    };

    user = mkOption {
      type = types.str;
      default = "janitarr";
      description = "User account under which janitarr runs.";
    };

    group = mkOption {
      type = types.str;
      default = "janitarr";
      description = "Group under which janitarr runs.";
    };

    package = mkOption {
      type = types.package;
      default = pkgs.janitarr or (pkgs.callPackage ./package.nix { });
      description = "The janitarr package to use.";
    };
  };

  config = mkIf cfg.enable {
    users.users.janitarr = mkIf (cfg.user == "janitarr") {
      isSystemUser = true;
      group = cfg.group;
      home = cfg.dataDir;
      description = "Janitarr service user";
    };

    users.groups.janitarr = mkIf (cfg.group == "janitarr") { };

    networking.firewall.allowedTCPPorts = mkIf cfg.openFirewall [ cfg.port ];

    systemd.services.janitarr = {
      description = "Janitarr media server automation";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        Type = "simple";
        User = cfg.user;
        Group = cfg.group;
        ExecStart = "${cfg.package}/bin/janitarr start --host 0.0.0.0 --port ${toString cfg.port}";
        Restart = "on-failure";
        RestartSec = 5;

        # State management
        StateDirectory = "janitarr";
        StateDirectoryMode = "0750";
        WorkingDirectory = cfg.dataDir;

        # Security hardening
        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX" ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        MemoryDenyWriteExecute = true;
        LockPersonality = true;
        SystemCallFilter = [ "@system-service" "~@privileged" "~@resources" ];
        SystemCallArchitectures = "native";
        CapabilityBoundingSet = "";
        AmbientCapabilities = "";
      };

      environment = {
        JANITARR_DB_PATH = "${cfg.dataDir}/janitarr.db";
        JANITARR_LOG_LEVEL = cfg.logLevel;
      };
    };
  };
}
```

**Reference:** See specs/deployment.md lines 231-279 for exact service config.

**Tests:**

- **Unit:** N/A
- **E2E:** `nix flake check` type-checks the module

**Verification:**

```bash
nix flake check
```

---

### Task 8: Create docker.yml Workflow

Create GitHub Actions workflow for Docker builds per specs/deployment.md.

**Files to Create:**

- `.github/workflows/docker.yml`

**Implementation:**

```yaml
name: Docker

on:
  push:
    branches: [main]
    tags: ["v*"]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up QEMU
        uses: docker/setup-qemu-action@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to GHCR
        if: startsWith(github.ref, 'refs/tags/v')
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            # For v1.2.3: creates v1.2.3, v1.2, v1, latest
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
            type=raw,value=latest,enable=${{ startsWith(github.ref, 'refs/tags/v') }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: ${{ startsWith(github.ref, 'refs/tags/v') }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

**Reference:** See specs/deployment.md lines 375-436 for exact workflow.

**Tests:**

- **Unit:** N/A
- **E2E:** Workflow runs successfully on push

**Verification:**

```bash
# After pushing, check GitHub Actions tab for successful build
```

---

### Task 9: Create nix.yml Workflow

Create GitHub Actions workflow for Nix flake checks per specs/deployment.md.

**Files to Create:**

- `.github/workflows/nix.yml`

**Implementation:**

```yaml
name: Nix

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install Nix
        uses: DeterminateSystems/nix-installer-action@v9

      - name: Setup Nix cache
        uses: DeterminateSystems/magic-nix-cache-action@v2

      - name: Check flake
        run: nix flake check

      - name: Build package
        run: nix build .#janitarr
```

**Reference:** See specs/deployment.md lines 438-469 for exact workflow.

**Tests:**

- **Unit:** N/A
- **E2E:** Workflow runs successfully on push

**Verification:**

```bash
# After pushing, check GitHub Actions tab for successful build
```

---

### Completion Checklist

- [x] Task 1: Add search limit upper bound validation
- [ ] Task 2: Add /health route alias
- [ ] Task 3: Create Dockerfile
- [ ] Task 4: Create docker-entrypoint.sh
- [ ] Task 5: Create flake.nix
- [ ] Task 6: Create nix/package.nix
- [ ] Task 7: Create nix/module.nix
- [ ] Task 8: Create docker.yml workflow
- [ ] Task 9: Create nix.yml workflow

**Final Verification:**

```bash
# Go tests
go test ./...

# Docker
docker build -t janitarr:test .
docker run --rm -d --name janitarr-test -p 3434:3434 -v $(pwd)/data:/data janitarr:test
curl -s http://localhost:3434/health
docker stop janitarr-test

# Nix
nix flake check
nix build .#janitarr
./result/bin/janitarr --version
```

---

## Spec Revisions

This section documents changes made to specification files during the planning process.

### 2026-01-24: Planning Review

- Verified all audit report issues from 2026-01-23 have been resolved in specs
- Confirmed port consistency (3434 throughout web-frontend.md)
- Confirmed 4 separate search limits defined in web-frontend.md (lines 315-337)
- Confirmed log retention range standardized to 7-90 days (line 357)
- Verified README.md status column updated for deployment.md
- No new spec changes required

### 2026-01-24: README Status Update

- Updated `specs/README.md`: Changed deployment.md status from "Planned" to "Not Started" to accurately reflect current state (no deployment files exist in repository)

### 2026-01-23: Specification Audit Complete

All 21 issues identified in the spec audit have been resolved. See `specs/AUDIT_REPORT.md` for detailed changes:

- **Critical (4 resolved):** Port consistency (3434), search limits (4 separate), log retention (7-90 days), encryption key storage
- **High (4 resolved):** Logging consolidation, dry-run deduplication, queue behavior, performance metrics
- **Medium (9 resolved):** Distribution algorithm, rate limiting, WebSocket backoff, API key validation, search limit constraints, configuration precedence, API error responses
- **Low (4 resolved):** Connection timeout, README status column, archive migration spec, sequence diagrams

---

## Completed Phases (Recent)

### Phase 27 - Spec-Code Alignment

**Completed:** 2026-01-23

Aligned implementation with specs from 2026-01-23 audit:

- Proportional search distribution (replaced round-robin)
- Rate limiting (100ms delay, 429 handling, 3-strike skip)
- Search limit validation (0-1000 range) - HTML form only, server-side partial
- High limit warnings (>100)
- Cycle duration monitoring

### Phase 26 - Modal Z-Index Fix

**Completed:** 2026-01-23 | **Commit:** `f1206a2`

Fixed modal z-index issue by moving modal-container outside `<main>` element. Improved E2E test pass rate from 86% to 88% (63/72 passing).

### Phase 25 - E2E Test Encryption Key Fix

**Completed:** 2026-01-23 | **Commit:** `5adb9f6`

Fixed E2E test encryption-related failures by preserving encryption key file across test runs. Improved test pass rate from 66% to 86%.

### Phase 24 - UI Bug Fixes & E2E Tests

**Completed:** 2026-01-23 | **Commit:** `1b8e643`

Fixed Alpine.js scoping issues, added favicon and navigation icons, improved UI contrast and visual separation.

---

**For complete implementation history:** See [IMPLEMENTATION_HISTORY.md](./IMPLEMENTATION_HISTORY.md)

---

## Quick Reference

### DaisyUI Version Compatibility

| DaisyUI Version | Tailwind CSS Version | Configuration Method                    |
| --------------- | -------------------- | --------------------------------------- |
| v4.x            | v3.x                 | `require("daisyui")` in tailwind.config |
| v5.x            | v4.x                 | `@plugin "daisyui"` in CSS file         |

### Development Commands

```bash
# Environment setup
direnv allow                # Load Go, templ, Tailwind, Playwright

# Build and run
just build                  # Generate templates + build binary
./janitarr start            # Production mode
./janitarr dev              # Development mode (verbose logging)

# Testing
go test ./...               # All tests
go test -race ./...         # Race detection
templ generate              # After .templ changes

# E2E testing
direnv exec . bunx playwright test --reporter=list

# Port configuration
./janitarr start            # Default port: 3434
./janitarr dev              # Default port: 3435
./janitarr dev --host 0.0.0.0  # Required for Playwright testing
```

### Database

- **Location:** `./data/janitarr.db` (override: `JANITARR_DB_PATH`)
- **Driver:** modernc.org/sqlite (pure Go, no CGO)
- **Testing:** Use `:memory:` for tests

### Code Patterns

- **Errors:** Wrap with context: `fmt.Errorf("context: %w", err)`
- **Tests:** Table-driven, use `httptest.Server` for API mocks
- **Exports:** Prefer unexported by default
- **API Keys:** Encrypted at rest (AES-256-GCM), never log decrypted
