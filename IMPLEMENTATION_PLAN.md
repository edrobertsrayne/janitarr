# Janitarr: Implementation Plan

## Overview

This document tracks implementation tasks for Janitarr, an automation tool for Radarr and Sonarr media servers written in Go.

## Agent Instructions

This document is designed for AI coding agents. Each task:

- Has a checkbox `[ ]` that should be marked `[x]` when complete
- Includes specific file paths and commands to execute
- Has clear completion criteria
- References specification documents in `specs/`

**Before starting each phase:**

1. Read the relevant specification documents
2. Write tests first (TDD approach)
3. Run `go test ./...` after each implementation
4. Commit working code before moving to the next task

**Environment:** Development tools are provided by devenv. Run `direnv allow` to load.

## Technology Stack

| Component       | Technology          | Purpose                      |
| --------------- | ------------------- | ---------------------------- |
| Language        | Go 1.22+            | Main application             |
| Web Framework   | Chi (go-chi/chi/v5) | HTTP routing                 |
| Database        | modernc.org/sqlite  | SQLite (pure Go, no CGO)     |
| CLI             | Cobra (spf13/cobra) | Command-line interface       |
| CLI Forms       | charmbracelet/huh   | Interactive terminal forms   |
| Console Logging | charmbracelet/log   | Colorized structured logging |
| Templates       | templ (a-h/templ)   | Type-safe HTML templates     |
| Interactivity   | htmx + Alpine.js    | Dynamic UI without React     |
| CSS             | Tailwind CSS v3     | Utility-first styling        |
| UI Components   | DaisyUI v4          | Semantic component classes   |

---

## Current Status

**Active Phase:** All current phases complete. Ready for next phase.

---

## Completed Phases (Archive)

### Phase 20: Build-Time Version Information ✓

**Completed:** 2026-01-21
**Commit:** TBD

**Summary:** Implemented dynamic version information from git instead of hardcoded version strings. Created a new `version` package with build-time variables (`Version`, `Commit`, `BuildDate`) that are set via ldflags during compilation. Updated the Makefile to inject version information from `git describe`, commit hash, and build timestamp. Updated both the CLI (`--version` flag) and web server (Prometheus metrics `janitarr_info`) to use the version package. All tests pass.

**Files Changed:**

- `src/version/version.go` - New package with build-time variables
- `src/version/version_test.go` - Tests for version package
- `Makefile` - Added ldflags to inject version info
- `src/cli/root.go` - Use version package instead of hardcoded version
- `src/web/server.go` - Use version package for Prometheus metrics

### Phase 19: Web Interface Bug Fixes ✓

**Reference:** `specs/logging.md`, `specs/web-frontend.md`
**Completed:** 2026-01-21
**Commits:**

- `bc8a873 fix(web): correct server ID interpolation in Test button`
- `804e332 fix(web): open modal dialog when Edit button is clicked`
- `88f6a34 fix(web): populate log metadata for web/terminal parity`

**Summary:** Fixed three critical web interface bugs: (1) Web logs now display the same metadata detail as terminal logs by populating the `Metadata` field in all logger methods and rendering it in the UI, (2) Edit button now properly opens the modal dialog using htmx's `hx-on::after-swap` event, (3) Test button now correctly interpolates server IDs using htmx attributes instead of Alpine.js string concatenation. All automated tests pass; manual browser verification recommended for visual confirmation.

### Phase 18: Enable Tests for GetEnabledServers and SetServerEnabled ✓

**Completed:** 2026-01-21
**Commit:** `0e39409 test: enable tests for GetEnabledServers and SetServerEnabled`

**Summary:** Activated previously commented-out tests for `GetEnabledServers` and `SetServerEnabled` methods. Added these methods to the `ServerManagerInterface` and updated the mock implementation in web handlers. All tests pass successfully, improving test coverage for server management functionality.

### Phase 17: DaisyUI Version Compatibility Fix ✓

**Reference:** `specs/daisyui-migration.md`
**Completed:** 2026-01-21
**Commit:** `dd18216 fix(ui): downgrade DaisyUI to v4 for Tailwind CSS 3 compatibility`

**Summary:** Fixed UI styling by downgrading DaisyUI from v5.5.14 to v4.12.24 to maintain compatibility with Tailwind CSS v3. All automated verification passed; manual browser testing recommended on next UI review.

---

## Quick Reference

### DaisyUI Version Compatibility

| DaisyUI Version | Tailwind CSS Version | Configuration Method                    |
| --------------- | -------------------- | --------------------------------------- |
| v4.x            | v3.x                 | `require("daisyui")` in tailwind.config |
| v5.x            | v4.x                 | `@plugin "daisyui"` in CSS file         |
