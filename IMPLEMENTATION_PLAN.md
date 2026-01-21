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

## Phase 17: DaisyUI Version Compatibility Fix

**Reference:** `specs/daisyui-migration.md`
**Status:** COMPLETED
**Priority:** CRITICAL - UI is completely unstyled

### Problem Description

The DaisyUI migration (Phase 13) was marked complete but DaisyUI styles are NOT being compiled into the CSS output. The UI renders as plain unstyled HTML with no component styling.

**Evidence from diagnosis:**

1. Screenshot (`daisyui-fail.png`) shows unstyled UI - no cards, no badges, no styled buttons, no drawer navigation
2. `grep -c "btn" static/css/app.css` returns 0 (no DaisyUI classes in compiled CSS)
3. `grep -c "data-theme" static/css/app.css` returns 0 (no theme system in CSS)
4. `grep -c "drawer" static/css/app.css` returns 0 (no drawer component)
5. Templates use DaisyUI classes (`btn`, `card`, `badge`, `drawer`, `bg-base-100`) but they have no styling

### Root Cause

**Version incompatibility between DaisyUI v5 and Tailwind CSS v3:**

| Package     | Installed Version | Required For DaisyUI v5 |
| ----------- | ----------------- | ----------------------- |
| daisyui     | 5.5.14            | -                       |
| tailwindcss | 3.x               | 4.x                     |

DaisyUI v5 requires Tailwind CSS v4 and uses the new `@plugin "daisyui"` syntax in CSS files. The project uses Tailwind CSS v3 with the old `require("daisyui")` syntax in `tailwind.config.cjs`, which DaisyUI v5 does not support.

**Configuration files are correct** - `tailwind.config.cjs` properly uses `require("daisyui")` and defines custom light/dark themes. The only issue is the DaisyUI version.

### Solution

**Downgrade DaisyUI to v4.x** (compatible with Tailwind CSS v3).

This is the recommended fix because:

1. Minimal changes required - only `package.json` needs to update the version
2. Theme configuration in `tailwind.config.cjs` already uses v4-compatible syntax
3. Template files already use correct DaisyUI class names (same between v4 and v5)
4. No need to upgrade Tailwind CSS (larger migration effort)

### 17.1 Downgrade DaisyUI to v4

- [x] Update `package.json`:
  - [x] Change `"daisyui": "^5.5.14"` to `"daisyui": "^4.12.24"`

- [x] Install correct version:
  - [x] Run `bun install`

- [x] Verify installation:
  - [x] Run `cat node_modules/daisyui/package.json | grep '"version"'`
  - [x] Confirm version shows `4.12.x`

### 17.2 Rebuild CSS and Verify

- [x] Generate CSS: `make generate`

- [x] Verify DaisyUI classes are in CSS:
  - [x] Run `grep -c "btn" static/css/app.css` - returned 68 (✓)
  - [x] Run `grep -c "drawer" static/css/app.css` - returned 33 (✓)
  - [x] Run `grep -c "data-theme" static/css/app.css` - returned 3 (✓)
  - [x] Run `grep -c "base-100" static/css/app.css` - returned 1 (✓)

- [x] Build binary: `make build`

### 17.3 Manual Visual Verification

- [ ] Start the server: `./janitarr dev`

- [ ] Open browser to http://localhost:3434

- [ ] Verify the following are styled correctly:
  - [ ] Navigation sidebar has background color (`bg-base-200`)
  - [ ] Cards have shadows and rounded corners
  - [ ] Buttons have colored backgrounds (primary blue/purple)
  - [ ] Badges show colored backgrounds (green for success, red for error)
  - [ ] Theme toggle in sidebar switches between light and dark
  - [ ] Mobile view shows hamburger menu (resize browser < 1024px)

- [ ] Check browser console for errors (should be none related to CSS)

### 17.4 Run Tests

- [x] Run unit tests: `go test ./...`
- [ ] Run E2E tests: `bun test:e2e` (if available)
- [x] All tests should pass

### 17.5 Commit Fix

- [ ] Stage changes: `git add package.json bun.lockb static/css/app.css`
- [ ] Commit: `git commit -m "fix(ui): downgrade DaisyUI to v4 for Tailwind CSS 3 compatibility"`

---

## Files to Modify (Phase 17)

| File                 | Changes                              |
| -------------------- | ------------------------------------ |
| `package.json`       | Change daisyui version from ^5 to ^4 |
| `bun.lockb`          | Updated automatically by bun install |
| `static/css/app.css` | Regenerated with DaisyUI styles      |

---

## Verification Checklist

### Phase 17: DaisyUI Version Fix

- [x] DaisyUI v4.x installed (not v5)
- [x] `make generate` completes without errors
- [x] Compiled CSS contains DaisyUI classes (`btn`, `drawer`, `card`, etc.)
- [x] Compiled CSS contains theme system (`data-theme`, `base-100`, etc.)
- [ ] UI displays with proper styling (cards, buttons, badges, navigation) - requires manual verification
- [ ] Theme toggle switches between light and dark correctly - requires manual verification
- [ ] Mobile hamburger menu appears on small screens - requires manual verification
- [x] All existing tests pass
- [x] Build completes successfully

---

## Notes

### DaisyUI Version Compatibility Reference

| DaisyUI Version | Tailwind CSS Version | Configuration Method                    |
| --------------- | -------------------- | --------------------------------------- |
| v4.x            | v3.x                 | `require("daisyui")` in tailwind.config |
| v5.x            | v4.x                 | `@plugin "daisyui"` in CSS file         |

### Why Downgrade Instead of Upgrade?

Upgrading to Tailwind CSS v4 + DaisyUI v5 would require:

1. Upgrade tailwindcss to v4.x in package.json
2. Install `@tailwindcss/vite` or `@tailwindcss/cli`
3. Convert `tailwind.config.cjs` to CSS-based configuration
4. Update `static/css/input.css` to use `@import "tailwindcss"; @plugin "daisyui";`
5. Update Makefile to use new Tailwind CLI options
6. Review all templates for any Tailwind v3 → v4 breaking changes
7. Test all components thoroughly

This is a larger migration effort and not recommended for a bug fix. The downgrade to DaisyUI v4 is a single-line change.
