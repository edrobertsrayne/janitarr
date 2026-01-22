# Janitarr: Implementation Plan

## Overview

This document tracks implementation tasks for Janitarr, an automation tool for Radarr and Sonarr media servers written in Go.

## Agent Instructions

This document is designed for AI coding agents. Each task:

- Has a checkbox `[ ]` that should be marked `[x]` when complete
- Includes specific file paths and exact code changes
- Has clear completion criteria
- Maps directly to issues in `ISSUES.md` by line number

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

**Active Phase:** None - All phases complete
**Previous Phase:** Phase 23 - Enable Skipped Database Tests (Complete ✓)

---

## Issue Mapping

Issues are numbered 1-10 based on their line order in `ISSUES.md`:

| Issue # | Description                               | Status |
| ------- | ----------------------------------------- | ------ |
| 1       | Add Server button doesn't open modal      | Fixed  |
| 2       | Web logs don't match CLI logs             | Fixed  |
| 3       | Run Now button missing icon               | Fixed  |
| 4       | Theme chooser still on Settings page      | Fixed  |
| 5       | Dashboard URL field empty                 | Fixed  |
| 6       | Test connection shows "connection failed" | Fixed  |
| 7       | Edit server button does nothing           | Fixed  |
| 8       | Delete uses browser modal, not DaisyUI    | Fixed  |
| 9       | Dev mode should use different port        | Fixed  |
| 10      | Port availability checking needed         | Fixed  |

**Note:** Phase 19 (commit `804e332`) added `hx-on::after-swap` to the Edit button. Issue #7 was fully resolved in commit `58a1a6f` by adding a setTimeout delay to ensure the modal is ready in the DOM.

---

## Completed Phases (Archive)

### Phase 23: Enable Skipped Database Tests ✓

**Completed:** 2026-01-22

**Commit:** `956e156` - test: enable skipped database tests for PurgeOldLogs, GetServerStats, and GetSystemStats

**Summary:** Enabled three previously skipped database tests that were marked as not implemented but had their implementations available. The tests were failing due to missing timestamps and IDs in test LogEntry instances. Fixed by adding explicit Timestamp values to prevent zero-time defaults and explicit ID values to prevent primary key collisions. All tests now verify correct behavior for log purging, server-specific statistics, and system-wide statistics.

**Tests Updated:**

- `TestLogsPurge` - Tests PurgeOldLogs method with proper retention period handling
- `TestServerStats` - Tests GetServerStats method aggregating searches and errors per server
- `TestSystemStats` - Tests GetSystemStats method with multi-server statistics

**Why This Matters:** These tests validate critical functionality for log retention and statistics reporting that powers the dashboard. Previously, this functionality was untested, creating risk for production deployments. The statistics queries use SQL aggregation (`SUM(count)` for searches, `COUNT(*)` for errors) that needed verification with realistic test data.

---

### Phase 22: E2E Test Suite Improvements ✓

**Completed:** 2026-01-22

**Commits:**

- `65552a8` - test(e2e): complete Phase 22 test suite improvements
- `3fae069` - test(e2e): add full integration flow tests
- `c3aea9d` - test(e2e): add settings persistence tests
- `7d0cd31` - test(e2e): add error handling tests
- `3c68c0b` - test(e2e): add dashboard integration tests
- `33fc45b` - test(e2e): add connection test functionality tests
- `72a5ffc` - test(e2e): add delete server workflow tests
- `9e3e9be` - test(e2e): add edit server workflow tests
- `4ae0b73` - fix(test): rewrite add-server tests with correct selectors
- `42d3ff0` - fix(test): use ID selectors for server form fields

**Summary:** Comprehensive improvement of the Playwright E2E test suite. Fixed existing test selectors to use ID-based selectors compatible with DaisyUI's label structure. Added 9 new test files covering critical user workflows: edit server, delete server (with DaisyUI modal), connection testing, dashboard integration, error handling, settings persistence, and full integration flows. Final test suite: 61 tests passing, 9 conditional tests (skip when prerequisites not met). All critical user journeys now validated through automated E2E tests.

**New Test Files:**

- `tests/e2e/edit-server.spec.ts` - Edit workflow tests
- `tests/e2e/delete-server.spec.ts` - Delete workflow tests
- `tests/e2e/test-connection.spec.ts` - Connection testing tests
- `tests/e2e/dashboard-integration.spec.ts` - Dashboard stats tests
- `tests/e2e/error-handling.spec.ts` - Error handling tests
- `tests/e2e/settings-persistence.spec.ts` - Settings save/load tests
- `tests/e2e/full-flow.spec.ts` - End-to-end integration tests

---

### Phase 21: ISSUES.md Fixes ✓

**Completed:** 2026-01-21

**Commits:**

- `7cd1c5b` - feat(cli): add port availability checking with clear error messages
- `58a1a6f` - fix(ui): display log count as number instead of ASCII
- `e12f4f3` - feat(ui): replace browser confirm with DaisyUI modal for delete
- `c0c78af` - fix(ui): ensure modal opens after DOM swap in Edit button
- `2c8acfd` - fix(ui): add modal trigger to Add Server buttons
- `f3f9bb6` - feat(ui): add play icon to Run Now button
- `22ce9f5` - fix(cli): use port 3435 for dev mode to avoid conflicts
- `d12deec` - fix(ui): remove deprecated theme chooser from settings
- `5e9cbcc` - fix(web): populate server URL in dashboard table

**Summary:** Fixed all 10 issues from ISSUES.md covering UI bugs and enhancements. Key improvements: Dashboard URL field population (Issue #5), theme chooser removal from Settings (Issue #4), dev mode default port changed to 3435 (Issue #9), Run Now button with play icon (Issue #3), Add Server modal trigger (Issue #1), Edit Server button with setTimeout fix (Issue #7), DaisyUI delete confirmation modal replacing browser confirm (Issue #8), log count display fix (Issue #2), and port availability checking with clear error messages (Issue #10). All manual testing checklist items verified.

**Files Modified:**

- `src/web/handlers/pages/dashboard.go` - Dashboard URL population
- `src/templates/pages/settings.templ` - Theme chooser removal
- `src/cli/dev.go` - Dev mode port and availability checking
- `src/templates/pages/dashboard.templ` - Run Now icon
- `src/templates/pages/servers.templ` - Add Server modal trigger
- `src/templates/components/server_card.templ` - Edit/Delete/Test buttons
- `src/templates/components/log_entry.templ` - Log count display
- `src/web/port.go` (new) - Port availability checking
- `src/web/port_test.go` (new) - Port checking tests
- `src/cli/start.go` - Port availability checking

---

### Phase 20: Build-Time Version Information ✓

**Completed:** 2026-01-21

**Commit:** `9675487` - feat: implement build-time version information from git

**Summary:** Implemented dynamic version information from git instead of hardcoded version strings. Created a new `version` package with build-time variables (`Version`, `Commit`, `BuildDate`) that are set via ldflags during compilation. Updated the Makefile to inject version information from `git describe`, commit hash, and build timestamp. Updated both the CLI (`--version` flag) and web server (Prometheus metrics `janitarr_info`) to use the version package.

---

### Phase 19: Web Interface Bug Fixes ✓

**Completed:** 2026-01-21

**Commits:**

- `bc8a873` - fix(web): correct server ID interpolation in Test button
- `804e332` - fix(web): open modal dialog when Edit button is clicked
- `88f6a34` - fix(web): populate log metadata for web/terminal parity

**Summary:** Fixed three critical web interface bugs: (1) Web logs now display the same metadata detail as terminal logs by populating the `Metadata` field in all logger methods and rendering it in the UI, (2) Edit button now properly opens the modal dialog using htmx's `hx-on::after-swap` event, (3) Test button now correctly interpolates server IDs using htmx attributes instead of Alpine.js string concatenation.

---

### Phase 18: Enable Tests for GetEnabledServers and SetServerEnabled ✓

**Completed:** 2026-01-21

**Commit:** `0e39409` - test: enable tests for GetEnabledServers and SetServerEnabled

**Summary:** Enabled previously disabled database tests for `GetEnabledServers` and `SetServerEnabled` methods. Tests now verify correct behavior for server enabling/disabling functionality and proper filtering of enabled servers.

---

### Phase 17: DaisyUI Version Compatibility Fix ✓

**Completed:** 2026-01-21

**Commit:** `dd18216` - fix(ui): downgrade DaisyUI to v4 for Tailwind CSS 3 compatibility

**Summary:** Fixed DaisyUI compatibility issues by downgrading from v5 (which requires Tailwind CSS v4) to v4 (compatible with Tailwind CSS v3). Updated package.json and tailwind.config.js to use the correct configuration method for DaisyUI v4.

---

### Phase 16 and Earlier

**Note:** Phases 1-16 implemented the core application functionality including CLI commands, database layer, web server, API endpoints, UI templates, server management, automation scheduling, and initial testing infrastructure.

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
make build                  # Generate templates + build binary
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
