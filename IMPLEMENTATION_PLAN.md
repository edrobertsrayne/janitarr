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

**Active Phase:** Phase 19 - Web Interface Bug Fixes

---

## Phase 19: Web Interface Bug Fixes

**Reference:** `specs/logging.md`, `specs/web-frontend.md`
**Priority:** High - User-facing functionality is broken

### Overview

Two related issues affect the web interface usability:

1. **Web logs lack terminal log detail** - The web interface shows simplified logs (e.g., "Search triggered. Count: 1") while terminal logs show full details (title, year, quality, etc.)
2. **Server management broken** - Edit button doesn't open modal, Test button fails with literal "{ server.ID }" error

### Bug 1: Web Log Metadata Parity

**Root Cause:** The `Metadata` field in `LogEntry` struct exists but is never populated by logging methods. Terminal logs receive full details via `console.Info()` key-value pairs, but the database entry only stores a simplified message.

**Files to modify:**

- `src/logger/logger.go` - Populate `Metadata` field in all logging methods
- `src/templates/components/log_entry.templ` - Render metadata key-value pairs

**Tasks:**

- [x] Update `LogMovieSearch()` to populate `Metadata` with title, year, quality profile
- [x] Update `LogEpisodeSearch()` to populate `Metadata` with series, season, episode, title, quality
- [x] Update `LogDetectionComplete()` to populate `Metadata` with missing count, cutoff_unmet count
- [x] Update `LogCycleStart()` and `LogCycleEnd()` to populate `Metadata` with relevant stats
- [x] Update `log_entry.templ` to render `entry.Metadata` as key-value pairs
- [x] Run `templ generate` and verify template compiles
- [ ] Test that web logs now show same detail as terminal logs (manual browser verification needed)

### Bug 2a: Edit Modal Not Opening

**Root Cause:** DaisyUI `<dialog>` elements require `.showModal()` to be called via JavaScript. When htmx swaps in the form content, the dialog is inserted but never opened.

**Files to modify:**

- `src/templates/components/server_card.templ` - Add `hx-on::after-swap` to call `showModal()`
- OR `src/templates/components/forms/server_form.templ` - Add auto-open script

**Tasks:**

- [x] Add `hx-on::after-swap="document.getElementById('server-modal').showModal()"` to Edit button
- [x] Run `templ generate` and verify template compiles
- [ ] Test that clicking Edit opens the modal dialog (manual browser verification needed)

### Bug 2b: Test Button Returns Literal "{ server.ID }"

**Root Cause:** Template variable interpolation failure. The code uses `'{ server.ID }'` inside a JavaScript string literal, which Templ does not interpolate. The literal text "{ server.ID }" is sent to the API.

**File:** `src/templates/components/server_card.templ` (line 19)

**Current (broken):**

```templ
@click="fetch('/api/servers/' + '{ server.ID }' + '/test', { method: 'POST' })"
```

**Tasks:**

- [ ] Fix template interpolation using one of these approaches:
  - Use `hx-post={ "/api/servers/" + server.ID + "/test" }` instead of @click fetch
  - OR use `data-server-id={ server.ID }` and read from dataset in JavaScript
  - OR build entire onclick attribute value with Templ concatenation
- [ ] Run `templ generate` and verify template compiles
- [ ] Test that Test button now calls correct endpoint with actual server ID

### Bug 2c: Test Without API Key (Cascading Fix)

**Root Cause:** This is a cascading failure from Bug 2b. The backend implementation (`server_manager.go:TestConnection`) correctly retrieves and uses stored API keys - the issue is that the wrong server ID ("{ server.ID }") is being sent.

**Tasks:**

- [ ] After fixing Bug 2b, verify that testing existing servers without entering a new API key works
- [ ] Confirm the stored decrypted API key is used for the connection test

### Completion Criteria

- [ ] Web log entries display same detail level as terminal logs
- [ ] Clicking Edit on a server card opens the edit modal
- [ ] Test Connection on server cards works and returns actual connection status
- [ ] Testing existing servers without new API key uses stored credentials
- [ ] All tests pass: `go test ./...`
- [ ] Templates compile: `templ generate`

---

## Completed Phases (Archive)

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
