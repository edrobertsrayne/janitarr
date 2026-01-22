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

## Completed Phases (Recent)

### Most Recent: Phase 23 - Enable Skipped Database Tests ✓

**Completed:** 2026-01-22 | **Commit:** `956e156`

Enabled three previously skipped database tests (`TestLogsPurge`, `TestServerStats`, `TestSystemStats`) that validate critical log retention and statistics functionality. Tests were failing due to missing timestamps/IDs in test data, now fixed.

### Phase 22 - E2E Test Suite Improvements ✓

**Completed:** 2026-01-22 | **Final:** `65552a8`

Comprehensive E2E test suite overhaul. Added 9 new test files covering all critical workflows. 61 tests passing, validates complete user journey from server addition to automation.

### Phase 21 - ISSUES.md Fixes ✓

**Completed:** 2026-01-21 | **9 commits**

Fixed all 10 reported issues including dashboard URL population, DaisyUI modal integration, port configuration, and UI enhancements.

---

**For complete implementation history:** See [IMPLEMENTATION_HISTORY.md](./IMPLEMENTATION_HISTORY.md) for detailed summaries of Phases 17-23 and earlier phases.

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
