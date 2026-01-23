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

**Active Phase:** Phase 24 - UI Bug Fixes & E2E Tests
**Previous Phase:** Phase 23 - Enable Skipped Database Tests (Complete ✓)
**Test Status:** All tests passing (Go unit tests + E2E tests)

### Implementation Completeness

All specifications from `/specs/` have been implemented:

- ✅ CLI Interface with interactive forms (cli-interface.md)
- ✅ Logging system with web viewer (logging.md)
- ✅ Web frontend with templ + htmx + Alpine.js (web-frontend.md)
- ✅ DaisyUI v4 integration (daisyui-migration.md)
- ✅ Unified service startup (unified-service-startup.md)
- ✅ Server configuration management (server-configuration.md)
- ✅ Activity logging (activity-logging.md)
- ✅ Missing content & quality cutoff detection (missing-content-detection.md, quality-cutoff-detection.md)
- ✅ Search triggering (search-triggering.md)
- ✅ Automatic scheduling (automatic-scheduling.md)

### Key Features

- **CLI**: Interactive forms for server/config management, server selector for edit/delete
- **Web UI**: Dashboard, servers, logs, settings pages with real-time updates
- **Logging**: WebSocket streaming, full-text search, export (JSON/CSV), filtering
- **Automation**: Configurable scheduling, search limits, dry-run mode
- **Testing**: Comprehensive unit tests + 61 E2E tests covering all workflows

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

## Phase 24: UI Bug Fixes & E2E Tests

**Status:** Not Started
**Priority:** Critical (modal bugs block core functionality)

This phase addresses critical UI bugs discovered during Playwright testing, adds missing UI polish, and implements E2E tests for server management workflows.

### Background

UI analysis revealed that the "Add Server" modal has Alpine.js scoping issues that prevent the Save button from displaying text and cause Cancel/Escape to not close the modal. The root cause is that `x-data="{ loading: false }"` is defined on the `<form>` element, but the Save button is in a sibling `<div class="modal-action">` outside the form's scope.

### Task 1: Fix Alpine.js x-data Scoping in Server Form Modal

**File:** `src/templates/components/forms/server_form.templ`

**Problem:** The `loading` variable is defined in `x-data` on the `<form>` element (line 23), but the Save button (lines 188-204) is outside the form in `<div class="modal-action">`. This causes:

- `x-show="!loading"` to fail silently (button text hidden)
- `x-bind:disabled="loading"` to not work
- Alpine.js console errors: `loading is not defined`

**Solution:** Move `x-data` from the `<form>` to the `<div class="modal-box">` wrapper so both the form and modal-action buttons share the same scope.

**Changes:**

- [x] **Line 7:** Add `x-data` to modal-box div:

  ```html
  <!-- BEFORE -->
  <div class="modal-box">
    <!-- AFTER -->
    <div class="modal-box" x-data="{ loading: false }"></div>
  </div>
  ```

- [x] **Lines 15-25:** Remove `x-data` and move htmx event handlers to form (keep only htmx attributes):

  ```html
  <!-- BEFORE -->
  <form
    id="server-form"
    ...
    x-data="{ loading: false }"
    @htmx:before-request="loading = true"
    @htmx:after-request="loading = false; if (event.detail.successful) { ... }"
    class="space-y-4 mt-4"
  >
    <!-- AFTER -->
    <form
      id="server-form"
      ...
      @htmx:before-request="loading = true"
      @htmx:after-request="loading = false; if (event.detail.successful) { document.getElementById('server-modal')?.close(); window.location.reload(); } else { try { const resp = JSON.parse(event.detail.xhr.responseText); alert(resp.error || 'Failed to save server'); } catch(e) { alert('Failed to save server'); } }"
      class="space-y-4 mt-4"
    ></form>
  </form>
  ```

**Verification:**

```bash
templ generate
make build
./janitarr dev --host 0.0.0.0
# Test: Click "Add Server", verify Save button shows "Create" text
# Test: Fill form and submit, verify loading state appears
```

---

### Task 2: Fix Modal Cancel Button

**File:** `src/templates/components/forms/server_form.templ`

**Problem:** The Cancel button (lines 182-186) uses `onclick` but the dialog may not close properly.

**Solution:** Use Alpine.js `@click` for consistency and add the close method to the modal-box x-data.

**Changes:**

- [x] **Line 7:** Extend x-data to include a close helper (already added in Task 1, update to):

  ```html
  <div
    class="modal-box"
    x-data="{ loading: false, closeModal() { document.getElementById('server-modal').close() } }"
  ></div>
  ```

- [x] **Lines 182-186:** Update Cancel button to use Alpine.js:

  ```html
  <!-- BEFORE -->
  <button
    type="button"
    onclick="document.getElementById('server-modal').close()"
    class="btn btn-ghost"
  >
    Cancel
  </button>

  <!-- AFTER -->
  <button type="button" @click="closeModal()" class="btn btn-ghost">
    Cancel
  </button>
  ```

**Verification:**

```bash
templ generate
make build
# Test: Open Add Server modal, click Cancel, verify modal closes
# Test: Open modal, press Escape, verify modal closes (native dialog behavior)
```

---

### Task 3: Add Favicon

**Problem:** Every page load logs a 404 error for `/favicon.ico`.

**Solution:** Add a simple favicon to static assets.

**Changes:**

- [x] Create `static/favicon.ico` - A simple 32x32 or 16x16 icon. Can use an SVG favicon for simplicity:

- [x] **File:** `src/templates/layouts/base.templ` - Add favicon link in `<head>`:

  ```html
  <!-- Add after <title> tag, around line 10 -->
  <link rel="icon" type="image/svg+xml" href="/static/favicon.svg" />
  ```

- [x] **Create file:** `static/favicon.svg`:
  ```svg
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
    <rect width="32" height="32" rx="6" fill="#661ae6"/>
    <text x="16" y="23" text-anchor="middle" font-family="system-ui" font-weight="bold" font-size="18" fill="white">J</text>
  </svg>
  ```

**Verification:**

```bash
make build
./janitarr dev
# Check browser console - no 404 for favicon
# Check browser tab shows favicon
```

---

### Task 4: Add Navigation Icons

**File:** `src/templates/components/nav.templ`

**Problem:** Navigation items are text-only. The spec (web-frontend.md) mentions icons should be present.

**Solution:** Add Heroicons SVG icons next to each navigation item.

**Changes:**

- [x] Update each `NavItem` call in the `<ul class="menu">` to include an icon. Modify the nav items (around lines 45-48):

  ```html
  <!-- Dashboard icon (HomeIcon) -->
  <li>
      <a href="/" class={ templ.KV("active", currentPath == "/") }>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z"/>
          </svg>
          Dashboard
      </a>
  </li>

  <!-- Servers icon (ServerIcon) -->
  <li>
      <a href="/servers" class={ templ.KV("active", currentPath == "/servers") }>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M2 5a2 2 0 012-2h12a2 2 0 012 2v2a2 2 0 01-2 2H4a2 2 0 01-2-2V5zm14 1a1 1 0 11-2 0 1 1 0 012 0zM2 13a2 2 0 012-2h12a2 2 0 012 2v2a2 2 0 01-2 2H4a2 2 0 01-2-2v-2zm14 1a1 1 0 11-2 0 1 1 0 012 0z" clip-rule="evenodd"/>
          </svg>
          Servers
      </a>
  </li>

  <!-- Activity Logs icon (ClockIcon) -->
  <li>
      <a href="/logs" class={ templ.KV("active", currentPath == "/logs") }>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd"/>
          </svg>
          Activity Logs
      </a>
  </li>

  <!-- Settings icon (CogIcon) -->
  <li>
      <a href="/settings" class={ templ.KV("active", currentPath == "/settings") }>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clip-rule="evenodd"/>
          </svg>
          Settings
      </a>
  </li>
  ```

**Verification:**

```bash
templ generate
make build
./janitarr dev
# Verify icons appear next to each navigation item
```

---

### Task 5: Improve Empty State Icons

**Files:** `src/templates/pages/servers.templ`, `src/templates/pages/dashboard.templ`

**Problem:** Arrow icon (`→`) used for empty states is semantically unclear.

**Solution:** Replace with contextual server/plus icons.

**Changes:**

- [x] **File:** `src/templates/pages/servers.templ` (lines 32-34) - Replace arrow with server + plus icon:

  ```html
  <!-- BEFORE -->
  <svg
    class="mx-auto h-12 w-12 text-base-content/30"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path
      stroke-linecap="round"
      stroke-linejoin="round"
      stroke-width="2"
      d="M5 12h14M12 5l7 7-7 7"
    ></path>
  </svg>

  <!-- AFTER - Server stack icon -->
  <svg
    class="mx-auto h-12 w-12 text-base-content/30"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path
      stroke-linecap="round"
      stroke-linejoin="round"
      stroke-width="2"
      d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"
    />
  </svg>
  ```

- [x] **File:** `src/templates/pages/dashboard.templ` - Find the "No servers configured" empty state (around line 74) and apply the same icon change.

**Verification:**

```bash
templ generate
make build
# Visual check: Empty states show server icon instead of arrow
```

---

### Task 6: Improve Dashboard Stats Card Separation

**File:** `src/templates/components/stats_card.templ`

**Problem:** The 4 stat cards appear as one continuous bar without visual separation.

**Solution:** Add shadow and ensure proper card styling to each stat.

**Changes:**

- [x] Updated StatsCard component to use `shadow-lg` instead of `shadow` for better visual separation

  ```html
  <!-- Ensure each stat div has shadow and rounded corners -->
  <div class="stat bg-base-100 rounded-box shadow-lg"></div>
  ```

  If the stats are using a shared `<div class="stats">` wrapper, consider switching to individual cards:

  ```html
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
    <div class="stat bg-base-100 rounded-box shadow-lg">
      <div class="stat-title">Servers</div>
      <div class="stat-value">{ stats.ServerCount }</div>
    </div>
    <!-- Repeat for each stat -->
  </div>
  ```

**Verification:**

```bash
templ generate
make build
# Visual check: Each stat card has visible shadow and spacing
```

---

### Task 7: Improve Light Theme Active Nav Contrast

**File:** `tailwind.config.cjs` or `src/templates/components/nav.templ`

**Problem:** Active navigation item background is barely visible in light theme.

**Solution:** Add explicit active state styling or use DaisyUI's menu-active classes.

**Changes:**

- [x] In `nav.templ`, added explicit active state styling with better contrast for light theme:

  ```html
  <!-- Active class now includes background color, primary text color, and font weight -->
  <a href="/" class={ "flex items-center gap-2", templ.KV("active bg-primary/10 text-primary font-semibold", currentPath == "/") }>
  ```

  Applied to all four navigation items (Dashboard, Servers, Activity Logs, Settings). The `bg-primary/10` adds a light background tint, `text-primary` colors the text with the primary theme color (purple), and `font-semibold` makes the text bolder for better visibility.

**Verification:**

```bash
templ generate
make build
# Toggle to light theme, verify active nav item is clearly visible
```

---

### Task 8: Add E2E Tests for Server Modal

**File:** `tests/e2e/add-server.spec.ts`

**Purpose:** Test that the server add/edit modal works correctly after the Alpine.js fixes.

**Status:** ✅ Complete - Most tests already existed; added missing "Escape key closes modal" test

**Changes:**

- [x] Tests already existed in add-server.spec.ts, edit-server.spec.ts, and test-connection.spec.ts
- [x] Added missing "Escape key closes modal" test to add-server.spec.ts (line 77-92)

**Test Coverage:**

- ✅ Add Server button opens modal - `tests/e2e/add-server.spec.ts:4`
- ✅ Save button displays text - `tests/e2e/add-server.spec.ts:28` (uses workaround selector)
- ✅ Cancel button closes modal - `tests/e2e/add-server.spec.ts:55`
- ✅ Escape key closes modal - `tests/e2e/add-server.spec.ts:77` (newly added)
- ✅ Test Connection shows feedback - `tests/e2e/test-connection.spec.ts:48`

---

### Task 9: Add E2E Tests for Theme Toggle

**File:** `tests/e2e/full-flow.spec.ts`

**Status:** ✅ Complete - Theme toggle test already existed; added missing reload persistence test

**Changes:**

- [x] Theme toggle test already existed in full-flow.spec.ts:38
- [x] Added missing "theme persists across page reload" test to full-flow.spec.ts (line 67-94)

**Test Coverage:**

- ✅ Theme toggle switches between light and dark - `tests/e2e/full-flow.spec.ts:38`
- ✅ Theme persists across page navigation - `tests/e2e/full-flow.spec.ts:57`
- ✅ Theme persists across page reload - `tests/e2e/full-flow.spec.ts:67` (newly added)

---

### Task 10: Update Specifications

**Files:** `specs/web-frontend.md`, `specs/daisyui-migration.md`

**Changes:**

- [ ] **File:** `specs/daisyui-migration.md` - Add warning about Alpine.js x-data scoping in modals section (Pattern #1):

  Add after line ~635 (after the modal resolution example):

  ````markdown
  **Important:** When using Alpine.js with DaisyUI modals, ensure `x-data` is placed on a parent element that encompasses ALL interactive elements. If `x-data` is on the `<form>` but buttons are in a sibling `<div class="modal-action">`, those buttons won't have access to the reactive state.

  ```html
  <!-- WRONG: buttons outside x-data scope -->
  <div class="modal-box">
    <form x-data="{ loading: false }">...</form>
    <div class="modal-action">
      <button x-show="!loading">Save</button>
      <!-- loading is undefined! -->
    </div>
  </div>

  <!-- CORRECT: x-data on parent -->
  <div class="modal-box" x-data="{ loading: false }">
    <form>...</form>
    <div class="modal-action">
      <button x-show="!loading">Save</button>
      <!-- loading is accessible -->
    </div>
  </div>
  ```
  ````

  ```

  ```

- [ ] **File:** `specs/web-frontend.md` - Add note about htmx + showModal() timing in the Modal section:

  Add to the "Interactions" section around line 203:

  ````markdown
  **Modal Opening with htmx:** When using htmx to load modal content dynamically, use `hx-on::after-swap` to call `showModal()` after the content is inserted:

  ```html
  <button
    hx-get="/servers/new"
    hx-target="#modal-container"
    hx-swap="innerHTML"
    hx-on::after-swap="document.getElementById('server-modal').showModal()"
  >
    Add Server
  </button>
  ```
  ````

  ```

  ```

**Verification:**

```bash
# Review the spec changes for accuracy
cat specs/daisyui-migration.md | grep -A 20 "Alpine.js x-data scoping"
```

---

### Completion Checklist

- [x] Task 1: Fix Alpine.js x-data scoping
- [x] Task 2: Fix modal Cancel button
- [x] Task 3: Add favicon
- [x] Task 4: Add navigation icons
- [x] Task 5: Improve empty state icons
- [x] Task 6: Improve stats card separation
- [x] Task 7: Improve light theme active nav contrast
- [x] Task 8: Add server modal E2E tests
- [x] Task 9: Add theme toggle E2E tests
- [ ] Task 10: Update specifications

**Final Verification:**

```bash
# Run all tests
go test ./...
templ generate
make build
direnv exec . bunx playwright test --reporter=list

# Manual testing
./janitarr dev --host 0.0.0.0
# 1. Add Server modal opens, shows "Create" button, Cancel works
# 2. Favicon appears in browser tab
# 3. Navigation icons visible
# 4. Theme toggle works and persists
# 5. No console errors
```

**Commit Message:**

```
fix: resolve server modal Alpine.js scoping and UI improvements

- Fix x-data scope so Save button displays text
- Fix Cancel button to properly close modal
- Add favicon.svg
- Add navigation icons (Home, Server, Clock, Cog)
- Improve empty state icons
- Improve stats card visual separation
- Add E2E tests for server modal and theme toggle
- Update specs with Alpine.js scoping guidance
```

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
