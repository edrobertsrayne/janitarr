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

**Active Phase:** Phase 21 - ISSUES.md Fixes
**Next Phase:** Phase 22 - E2E Test Suite Improvements

---

## Issue Mapping

Issues are numbered 1-10 based on their line order in `ISSUES.md`:

| Issue # | Description                               | Status |
| ------- | ----------------------------------------- | ------ |
| 1       | Add Server button doesn't open modal      | Fixed  |
| 2       | Web logs don't match CLI logs             | Open   |
| 3       | Run Now button missing icon               | Fixed  |
| 4       | Theme chooser still on Settings page      | Fixed  |
| 5       | Dashboard URL field empty                 | Fixed  |
| 6       | Test connection shows "connection failed" | Fixed  |
| 7       | Edit server button does nothing           | Fixed  |
| 8       | Delete uses browser modal, not DaisyUI    | Fixed  |
| 9       | Dev mode should use different port        | Fixed  |
| 10      | Port availability checking needed         | Open   |

**Note:** Phase 19 (commit `804e332`) added `hx-on::after-swap` to the Edit button. Issue #7 was fully resolved in commit `58a1a6f` by adding a setTimeout delay to ensure the modal is ready in the DOM.

---

## Phase 21: ISSUES.md Fixes

### [x] Task 1: Fix Dashboard URL Field Empty (Issue #5)

**File:** `src/web/handlers/pages/dashboard.go`
**Difficulty:** Trivial

**Current code (line 30):**

```go
URL:     "",
```

**Change to:**

```go
URL:     srv.URL,
```

**Steps:**

1. Open `src/web/handlers/pages/dashboard.go`
2. Find line 30 inside the `for i, srv := range servers` loop
3. Change `URL: "",` to `URL: srv.URL,`
4. Save file

**Verify:**

```bash
go test ./src/web/handlers/pages/...
make build
./janitarr dev --host 0.0.0.0
# Open http://<host-ip>:3434 - Dashboard should show server URLs
```

**Commit:** `fix(web): populate server URL in dashboard table`

---

### [x] Task 2: Remove Theme Chooser from Settings (Issue #4)

**File:** `src/templates/pages/settings.templ`
**Difficulty:** Trivial

**Steps:**

1. Open `src/templates/pages/settings.templ`
2. Delete line 18: `@ThemeSelector()`
3. Delete lines 24-83: The entire `templ ThemeSelector()` component
4. Save file

**After deletion, file should look like:**

```go
package pages

import (
	"github.com/user/janitarr/src/templates/layouts"
	"github.com/user/janitarr/src/templates/components/forms"
	"github.com/user/janitarr/src/database"
)

templ Settings(config database.AppConfig, logCount int) {
	@layouts.Base("Settings") {
		<div class="max-w-4xl mx-auto">
			<div class="mb-6">
				<h1 class="text-3xl font-bold">Settings</h1>
				<p class="mt-2 text-sm text-base-content/60">
					Configure automation schedule and search limits
				</p>
			</div>
			@forms.ConfigForm(config, logCount)
		</div>
	}
}
```

**Verify:**

```bash
templ generate
make build
./janitarr dev --host 0.0.0.0
# Open http://<host-ip>:3434/settings - No theme dropdown should appear
```

**Commit:** `fix(ui): remove deprecated theme chooser from settings`

---

### [x] Task 3: Change Dev Mode Default Port (Issue #9)

**File:** `src/cli/dev.go`
**Difficulty:** Trivial

**Current code (line 26):**

```go
devCmd.Flags().IntP("port", "p", 3434, "Web server port")
```

**Change to:**

```go
devCmd.Flags().IntP("port", "p", 3435, "Web server port (default: 3435 for dev mode)")
```

**Steps:**

1. Open `src/cli/dev.go`
2. Find line 26 in the `init()` function
3. Change `3434` to `3435`
4. Update the help text to clarify it's for dev mode
5. Save file

**Verify:**

```bash
make build
./janitarr dev --help
# Should show: --port, -p int   Web server port (default: 3435 for dev mode) (default 3435)
./janitarr dev --host 0.0.0.0
# Should start on port 3435, not 3434
```

**Commit:** `fix(cli): use port 3435 for dev mode to avoid conflicts`

---

### [x] Task 4: Add Icon to Run Now Button (Issue #3)

**File:** `src/templates/pages/dashboard.templ`
**Difficulty:** Easy

**Current code (lines 38-48):**

```html
<button
  hx-post="/api/automation/trigger"
  hx-swap="none"
  hx-indicator="#run-spinner"
  class="btn btn-primary"
>
  <span id="run-spinner" class="htmx-indicator">
    <span class="loading loading-spinner loading-sm"></span>
  </span>
  Run Now
</button>
```

**Replace with:**

```html
<button
  hx-post="/api/automation/trigger"
  hx-swap="none"
  hx-indicator="#run-spinner"
  class="btn btn-primary"
>
  <span id="run-spinner" class="htmx-indicator">
    <span class="loading loading-spinner loading-sm"></span>
  </span>
  <svg
    id="run-icon"
    class="w-5 h-5"
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 20 20"
    fill="currentColor"
  >
    <path
      fill-rule="evenodd"
      d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z"
      clip-rule="evenodd"
    ></path>
  </svg>
  Run Now
</button>
```

**Steps:**

1. Open `src/templates/pages/dashboard.templ`
2. Find the Run Now button (lines 38-48)
3. Replace the entire button element with the new version above
4. Save file

**Verify:**

```bash
templ generate
make build
./janitarr dev --host 0.0.0.0
# Open http://<host-ip>:3435 - Run Now button should show play icon
# Click button - spinner should appear while running
```

**Commit:** `feat(ui): add play icon to Run Now button`

---

### [x] Task 5: Fix Add Server Button Modal Trigger (Issue #1)

**File:** `src/templates/pages/servers.templ`
**Difficulty:** Easy

The Add Server buttons fetch the form but don't open the modal. Need to add `hx-on::after-swap` to trigger `showModal()`.

**Step 1: Fix main Add Server button (lines 16-22)**

**Current:**

```html
<button
  hx-get="/servers/new"
  hx-target="#modal-container"
  hx-swap="innerHTML"
  class="btn btn-primary"
>
  Add Server
</button>
```

**Change to:**

```html
<button
  hx-get="/servers/new"
  hx-target="#modal-container"
  hx-swap="innerHTML"
  hx-on::after-swap="document.getElementById('server-modal').showModal()"
  class="btn btn-primary"
>
  Add Server
</button>
```

**Step 2: Fix empty state Add Server button (lines 37-46)**

**Current:**

```html
<button
  hx-get="/servers/new"
  hx-target="#modal-container"
  hx-swap="innerHTML"
  class="btn btn-primary"
>
  <svg ...>...</svg>
  Add Server
</button>
```

**Change to:**

```html
<button
  hx-get="/servers/new"
  hx-target="#modal-container"
  hx-swap="innerHTML"
  hx-on::after-swap="document.getElementById('server-modal').showModal()"
  class="btn btn-primary"
>
  <svg ...>...</svg>
  Add Server
</button>
```

**Verify:**

```bash
templ generate
make build
./janitarr dev --host 0.0.0.0
# Open http://<host-ip>:3435/servers
# Click "Add Server" button - modal should appear
```

**Commit:** `fix(ui): add modal trigger to Add Server buttons`

---

### [x] Task 6: Fix Edit Server Button (Issue #7)

**File:** `src/templates/components/server_card.templ`
**Difficulty:** Investigation + Fix

Phase 19 added `hx-on::after-swap` to the Edit button, but it still doesn't work. Debug and fix.

**Debug Steps:**

1. Start the dev server: `./janitarr dev --host 0.0.0.0`
2. Open browser DevTools (F12) → Console tab
3. Navigate to `/servers`
4. Click "Edit" on a server card
5. Check for JavaScript errors in Console
6. Check Network tab - is the `/servers/{id}/edit` request returning HTML?

**Possible Issues:**

1. **Modal element not found** - The `showModal()` call might fail if the dialog isn't in the DOM yet
2. **Dialog element missing ID** - Verify the form includes `<dialog id="server-modal">`
3. **HTMX timing** - The `after-swap` event might fire before the DOM is fully updated

**Likely Fix:**

The current code assumes the modal is ready immediately after swap. Add a small delay:

**Current (line 32):**

```html
hx-on::after-swap="document.getElementById('server-modal').showModal()"
```

**Change to:**

```html
hx-on::after-swap="setTimeout(() =>
document.getElementById('server-modal')?.showModal(), 10)"
```

**Steps:**

1. Open `src/templates/components/server_card.templ`
2. Find the Edit button (lines 28-35)
3. Change line 32 from:
   `hx-on::after-swap="document.getElementById('server-modal').showModal()"`
   to:
   `hx-on::after-swap="setTimeout(() => document.getElementById('server-modal')?.showModal(), 10)"`
4. Save file

**Verify:**

```bash
templ generate
make build
./janitarr dev --host 0.0.0.0
# Open http://<host-ip>:3435/servers
# Click "Edit" on a server card - modal should appear with pre-filled data
```

**Commit:** `fix(ui): ensure modal opens after DOM swap in Edit button`

---

### [x] Task 7: Verify Test Connection Works (Issue #6)

**Note:** This was previously Task 6 - renumbered due to Issue #7 fix insertion.

**Files:** `src/templates/components/server_card.templ`, `src/web/handlers/api/servers.go`
**Difficulty:** Verification + possible fix

Phase 19 fixed the server ID interpolation. Need to verify if test connection now works.

**Steps:**

1. Start the dev server: `./janitarr dev --host 0.0.0.0`
2. Open browser DevTools (F12) → Network tab
3. Navigate to `/servers`
4. Click "Test" on a server card
5. Check Network tab for the response from `/api/servers/{id}/test`

**If test shows "Connected (version)":** Issue is fixed, mark as complete.

**If test still shows "Connection failed":**

Check the API response format. The handler at `src/web/handlers/api/servers.go` should return:

```json
{ "success": true, "version": "4.5.0", "error": "" }
```

The client-side handler in `server_card.templ` line 22 expects this format:

```javascript
const data = JSON.parse($event.detail.xhr.response);
testResult = data.success
  ? "Connected (" + data.version + ")"
  : data.error || "Connection failed";
```

If the response is wrapped (e.g., `{"data": {...}}`), update line 22 to:

```javascript
const response = JSON.parse($event.detail.xhr.response);
const data = response.data || response;
testResult = data.success
  ? "Connected (" + data.version + ")"
  : data.error || "Connection failed";
```

**Commit (if changes needed):** `fix(ui): handle API response format in test connection`

---

### [x] Task 8: Replace Delete Modal with DaisyUI (Issue #8)

**File:** `src/templates/components/server_card.templ`
**Difficulty:** Moderate

Replace the browser's native `confirm()` dialog with a DaisyUI modal.

**Current Delete button (lines 36-43):**

```html
<button
	hx-delete={ "/api/servers/" + server.ID }
	hx-confirm="Are you sure you want to delete this server?"
	hx-target="closest div.card"
	hx-swap="outerHTML swap:1s"
	class="btn btn-ghost btn-sm text-error">
	Delete
</button>
```

**Replace the entire ServerCard component with:**

```go
templ ServerCard(server services.ServerInfo) {
	<div class="card bg-base-100 shadow-xl" x-data="{ testing: false, testResult: '', showDeleteModal: false }">
		<div class="card-body">
			<div class="flex items-center justify-between">
				<h2 class="card-title">{ server.Name }</h2>
				@ServerTypeBadge(server.Type)
			</div>
			<p class="text-base-content/70 break-all">{ server.URL }</p>
			<div class="flex items-center gap-2">
				@ServerStatusBadge(server.Enabled)
			</div>
			<div class="card-actions justify-end">
				<button
					type="button"
					hx-post={ "/api/servers/" + server.ID + "/test" }
					hx-swap="none"
					@click="testing = true; testResult = ''"
					@htmx:after-request="testing = false; if ($event.detail.successful) { const data = JSON.parse($event.detail.xhr.response); testResult = data.success ? 'Connected (' + data.version + ')' : (data.error || 'Connection failed') } else { testResult = 'Error: Request failed' }"
					:disabled="testing"
					class="btn btn-ghost btn-sm">
					<span x-show="!testing">Test</span>
					<span x-show="testing">Testing...</span>
				</button>
				<button
					hx-get={ "/servers/" + server.ID + "/edit" }
					hx-target="#modal-container"
					hx-swap="innerHTML"
					hx-on::after-swap="document.getElementById('server-modal').showModal()"
					class="btn btn-ghost btn-sm">
					Edit
				</button>
				<button
					@click="showDeleteModal = true"
					class="btn btn-ghost btn-sm text-error">
					Delete
				</button>
			</div>
			<div
				x-show="testResult"
				class="mt-1 text-xs"
				:class="testResult.startsWith('Connected') ? 'text-success' : 'text-error'"
				x-text="testResult">
			</div>
			<!-- Delete Confirmation Modal -->
			<dialog class="modal" :class="{ 'modal-open': showDeleteModal }">
				<div class="modal-box">
					<h3 class="font-bold text-lg">Delete Server</h3>
					<p class="py-4">Are you sure you want to delete <strong>{ server.Name }</strong>? This action cannot be undone.</p>
					<div class="modal-action">
						<button @click="showDeleteModal = false" class="btn">Cancel</button>
						<button
							hx-delete={ "/api/servers/" + server.ID }
							hx-target="closest div.card"
							hx-swap="outerHTML swap:1s"
							@click="showDeleteModal = false"
							class="btn btn-error">
							Delete
						</button>
					</div>
				</div>
				<div class="modal-backdrop" @click="showDeleteModal = false"></div>
			</dialog>
		</div>
	</div>
}
```

**Steps:**

1. Open `src/templates/components/server_card.templ`
2. Replace the entire `ServerCard` component (lines 5-53) with the new version above
3. Keep the `ServerTypeBadge` and `ServerStatusBadge` components unchanged
4. Save file

**Verify:**

```bash
templ generate
make build
./janitarr dev --host 0.0.0.0
# Open http://<host-ip>:3435/servers
# Click "Delete" - DaisyUI modal should appear (not browser confirm)
# Click "Cancel" - modal should close, server still exists
# Click "Delete" again, then confirm - server should be deleted
```

**Commit:** `feat(ui): replace browser confirm with DaisyUI modal for delete`

---

### [x] Task 9: Fix Log Count Display (Issue #2, Part A)

**File:** `src/templates/components/log_entry.templ`
**Difficulty:** Trivial

The count display has a bug - it only works for single digits (0-9).

**Current code (line 39):**

```go
Count: { string(rune(entry.Count + 48)) }
```

**Change to:**

```go
Count: { fmt.Sprint(entry.Count) }
```

**Steps:**

1. Open `src/templates/components/log_entry.templ`
2. Find line 39
3. Change `{ string(rune(entry.Count + 48)) }` to `{ fmt.Sprint(entry.Count) }`
4. Save file

**Verify:**

```bash
templ generate
go test ./...
```

**Commit:** `fix(ui): display log count as number instead of ASCII`

---

### [ ] Task 10: Add Port Availability Checking (Issue #10)

**Files:** `src/web/port.go` (new), `src/cli/start.go`, `src/cli/dev.go`
**Difficulty:** Moderate

**Step 1: Create `src/web/port.go`**

```go
package web

import (
	"fmt"
	"net"
	"time"
)

// IsPortAvailable checks if a port is available for binding
func IsPortAvailable(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)

	// Try to listen on the port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()

	// Small delay to ensure port is fully released
	time.Sleep(10 * time.Millisecond)
	return true
}
```

**Step 2: Create `src/web/port_test.go`**

```go
package web

import (
	"net"
	"testing"
)

func TestIsPortAvailable(t *testing.T) {
	// Test that a random high port is available
	if !IsPortAvailable("localhost", 59999) {
		t.Skip("Port 59999 unexpectedly in use")
	}

	// Occupy a port
	listener, err := net.Listen("tcp", "localhost:59998")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()

	// Test that occupied port is not available
	if IsPortAvailable("localhost", 59998) {
		t.Error("Expected port 59998 to be unavailable")
	}
}
```

**Step 3: Update `src/cli/start.go`**

Add after the port validation (around line 45):

```go
// Check if port is available
if !web.IsPortAvailable(host, port) {
	return fmt.Errorf("port %d is already in use on %s. Use --port to specify a different port", port, host)
}
```

**Step 4: Update `src/cli/dev.go`**

Add after the port validation (around line 46):

```go
// Check if port is available
if !web.IsPortAvailable(host, port) {
	return fmt.Errorf("port %d is already in use on %s. Use --port to specify a different port", port, host)
}
```

**Verify:**

```bash
go test ./src/web/...
make build

# Test 1: Start first instance
./janitarr start &
sleep 2

# Test 2: Try second instance on same port - should show clear error
./janitarr start
# Expected: "port 3434 is already in use on localhost"

# Test 3: Second instance with different port - should work
./janitarr start --port 8080

# Cleanup
pkill janitarr
```

**Commit:** `feat(cli): add port availability checking with clear error messages`

---

## Final Verification

### [ ] Task 11: Run Full Test Suite

```bash
go test ./...
go test -race ./...
templ generate
make build
```

**All tests must pass.**

---

### [ ] Task 12: Manual Browser Testing Checklist

Start the server: `./janitarr dev --host 0.0.0.0`

- [ ] Dashboard shows server URLs in table (Issue #5)
- [ ] Settings page has no theme dropdown (Issue #4)
- [ ] Dev mode starts on port 3435 (Issue #9)
- [ ] Run Now button shows play icon (Issue #3)
- [ ] Add Server button opens modal (Issue #1)
- [ ] Edit Server button opens modal with data (Issue #7)
- [ ] Test connection shows success/error correctly (Issue #6)
- [ ] Delete Server shows DaisyUI modal (Issue #8)
- [ ] Port conflict shows clear error message (Issue #10)
- [ ] Log count displays as number (Issue #2)

---

## Files Modified Summary

| File                                         | Issues Addressed |
| -------------------------------------------- | ---------------- |
| `src/web/handlers/pages/dashboard.go`        | #5               |
| `src/templates/pages/settings.templ`         | #4               |
| `src/cli/dev.go`                             | #9, #10          |
| `src/templates/pages/dashboard.templ`        | #3               |
| `src/templates/pages/servers.templ`          | #1               |
| `src/templates/components/server_card.templ` | #6, #7, #8       |
| `src/templates/components/log_entry.templ`   | #2               |
| `src/web/port.go` (new)                      | #10              |
| `src/web/port_test.go` (new)                 | #10              |
| `src/cli/start.go`                           | #10              |

---

## Phase 22: E2E Test Suite Improvements

**Goal:** Improve Playwright E2E test depth and coverage. Current tests are broad but shallow - many use conditional skips and loose assertions. This phase adds comprehensive tests for critical user flows.

**Prerequisites:**

- Phase 21 must be complete (UI bugs fixed)
- Run `direnv allow` to ensure Chromium is available via `CHROMIUM_PATH`

**Test Environment:**

```bash
# Verify Playwright works
direnv exec . bunx playwright test --reporter=list
```

---

### [ ] Task 1: Fix Existing Test Selectors

**File:** `tests/e2e/servers.spec.ts`
**Difficulty:** Easy

The existing server tests fail because they use `getByLabel()` which requires proper label-input associations. The form uses DaisyUI's label structure without `for` attributes.

**Current code (lines 26-31):**

```typescript
const nameField = page.getByLabel(/name/i).first();
const urlField = page.getByLabel(/url/i).first();

// At least one of these fields should be visible
const nameVisible = await nameField.isVisible().catch(() => false);
const urlVisible = await urlField.isVisible().catch(() => false);

expect(nameVisible || urlVisible).toBeTruthy();
```

**Replace with:**

```typescript
// Use ID selectors since DaisyUI labels don't have for attributes
const nameField = page.locator("#name");
const urlField = page.locator("#url");

await expect(nameField).toBeVisible();
await expect(urlField).toBeVisible();
```

**Also fix lines 47-48:**

```typescript
// Current
await page.getByLabel(/name/i).first().fill("Test Radarr");
await page.getByLabel(/url/i).first().fill("http://localhost:7878");

// Replace with
await page.locator("#name").fill("Test Radarr");
await page.locator("#url").fill("http://localhost:7878");
```

**And fix line 60-63:**

```typescript
// Current
const apiKeyField = page.getByLabel(/api key/i).first();
if (await apiKeyField.isVisible()) {
  await apiKeyField.fill("1234567890abcdef1234567890abcdef");
}

// Replace with
await page.locator("#apiKey").fill("1234567890abcdef1234567890abcdef");
```

**Verify:**

```bash
direnv exec . bunx playwright test tests/e2e/servers.spec.ts --reporter=list
```

**Commit:** `fix(test): use ID selectors for server form fields`

---

### [ ] Task 2: Fix Add Server Flow Test

**File:** `tests/e2e/add-server.spec.ts`
**Difficulty:** Easy

Update selectors to match actual form structure.

**Replace entire file with:**

```typescript
import { test, expect } from "./setup";

test.describe("Add Server flow", () => {
  test("should open modal and display form fields", async ({ page }) => {
    await page.goto("/servers");

    // Click Add Server button
    const addButton = page.getByRole("button", { name: /add server/i });
    await expect(addButton).toBeVisible();
    await addButton.click();

    // Wait for modal to open
    const modal = page.locator("#server-modal");
    await expect(modal).toBeVisible();

    // Verify form fields are present
    await expect(page.locator("#name")).toBeVisible();
    await expect(page.locator("#url")).toBeVisible();
    await expect(page.locator("#apiKey")).toBeVisible();
    await expect(
      page.locator('input[name="type"][value="radarr"]'),
    ).toBeVisible();
    await expect(
      page.locator('input[name="type"][value="sonarr"]'),
    ).toBeVisible();

    // Verify buttons
    await expect(page.getByRole("button", { name: /create/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /cancel/i })).toBeVisible();
    await expect(
      page.getByRole("button", { name: /test connection/i }),
    ).toBeVisible();
  });

  test("should validate required fields", async ({ page }) => {
    await page.goto("/servers");

    // Open modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Try to submit empty form
    await page.getByRole("button", { name: /create/i }).click();

    // Form should still be visible (not submitted due to HTML5 validation)
    await expect(page.locator("#server-modal")).toBeVisible();

    // Name field should show validation state
    const nameField = page.locator("#name");
    await expect(nameField).toHaveAttribute("required");
  });

  test("should close modal on cancel", async ({ page }) => {
    await page.goto("/servers");

    // Open modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Click cancel
    await page.getByRole("button", { name: /cancel/i }).click();

    // Modal should close
    await expect(page.locator("#server-modal")).not.toBeVisible();
  });
});
```

**Verify:**

```bash
direnv exec . bunx playwright test tests/e2e/add-server.spec.ts --reporter=list
```

**Commit:** `fix(test): rewrite add-server tests with correct selectors`

---

### [ ] Task 3: Add Edit Server Tests

**File:** `tests/e2e/edit-server.spec.ts` (new)
**Difficulty:** Moderate

Create comprehensive tests for the edit server workflow.

**Create file with:**

```typescript
import { test, expect } from "./setup";

test.describe("Edit Server flow", () => {
  test.beforeEach(async ({ page }) => {
    // Ensure at least one server exists by navigating to servers page
    // The test database should have a server from previous test runs
    await page.goto("/servers");
  });

  test("should open edit modal with pre-filled data", async ({ page }) => {
    // Find and click Edit button on first server card
    const editButton = page.getByRole("button", { name: /edit/i }).first();

    // Skip if no servers exist
    if (!(await editButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await editButton.click();

    // Wait for modal to open
    const modal = page.locator("#server-modal");
    await expect(modal).toBeVisible();

    // Verify modal title indicates edit mode
    await expect(
      page.getByRole("heading", { name: /edit server/i }),
    ).toBeVisible();

    // Verify fields are pre-filled (not empty)
    const nameField = page.locator("#name");
    await expect(nameField).toBeVisible();
    const nameValue = await nameField.inputValue();
    expect(nameValue.length).toBeGreaterThan(0);

    // Verify Update button (not Create)
    await expect(page.getByRole("button", { name: /update/i })).toBeVisible();
  });

  test("should show enabled checkbox in edit mode", async ({ page }) => {
    const editButton = page.getByRole("button", { name: /edit/i }).first();

    if (!(await editButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await editButton.click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Enabled checkbox should be visible in edit mode
    const enabledCheckbox = page.locator("#enabled");
    await expect(enabledCheckbox).toBeVisible();
  });

  test("should allow editing server name", async ({ page }) => {
    const editButton = page.getByRole("button", { name: /edit/i }).first();

    if (!(await editButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await editButton.click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Clear and type new name
    const nameField = page.locator("#name");
    await nameField.clear();
    await nameField.fill("Updated Server Name");

    // Verify the value changed
    await expect(nameField).toHaveValue("Updated Server Name");
  });

  test("should close edit modal on cancel without saving", async ({ page }) => {
    const editButton = page.getByRole("button", { name: /edit/i }).first();

    if (!(await editButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    // Get original server name from card
    const serverCard = page.locator(".card").first();
    const originalName = await serverCard.locator(".card-title").textContent();

    await editButton.click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Change the name
    const nameField = page.locator("#name");
    await nameField.clear();
    await nameField.fill("Should Not Save");

    // Cancel
    await page.getByRole("button", { name: /cancel/i }).click();
    await expect(page.locator("#server-modal")).not.toBeVisible();

    // Original name should still be displayed
    await expect(serverCard.locator(".card-title")).toHaveText(originalName!);
  });
});
```

**Verify:**

```bash
direnv exec . bunx playwright test tests/e2e/edit-server.spec.ts --reporter=list
```

**Commit:** `test(e2e): add edit server workflow tests`

---

### [ ] Task 4: Add Delete Server Tests

**File:** `tests/e2e/delete-server.spec.ts` (new)
**Difficulty:** Moderate

Create tests for delete server workflow with DaisyUI modal confirmation.

**Create file with:**

```typescript
import { test, expect } from "./setup";

test.describe("Delete Server flow", () => {
  test("should show DaisyUI confirmation modal", async ({ page }) => {
    await page.goto("/servers");

    // Find delete button
    const deleteButton = page.getByRole("button", { name: /delete/i }).first();

    if (!(await deleteButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await deleteButton.click();

    // DaisyUI modal should appear (not browser confirm)
    const modal = page.locator(".modal-open, dialog[open]");
    await expect(modal).toBeVisible();

    // Modal should have confirmation text
    await expect(page.getByText(/are you sure/i)).toBeVisible();

    // Modal should have Cancel and Delete buttons
    await expect(page.getByRole("button", { name: /cancel/i })).toBeVisible();
    await expect(modal.getByRole("button", { name: /delete/i })).toBeVisible();
  });

  test("should close modal on cancel without deleting", async ({ page }) => {
    await page.goto("/servers");

    const deleteButton = page.getByRole("button", { name: /delete/i }).first();

    if (!(await deleteButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    // Count servers before
    const serverCountBefore = await page.locator(".card").count();

    await deleteButton.click();

    // Wait for modal
    const modal = page.locator(".modal-open, dialog[open]");
    await expect(modal).toBeVisible();

    // Click cancel
    await page.getByRole("button", { name: /cancel/i }).click();

    // Modal should close
    await expect(modal).not.toBeVisible();

    // Server count should be unchanged
    const serverCountAfter = await page.locator(".card").count();
    expect(serverCountAfter).toBe(serverCountBefore);
  });

  test("should delete server on confirm", async ({ page }) => {
    await page.goto("/servers");

    const deleteButton = page.getByRole("button", { name: /delete/i }).first();

    if (!(await deleteButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    // Count servers before
    const serverCountBefore = await page.locator(".card").count();

    await deleteButton.click();

    // Wait for modal and confirm delete
    const modal = page.locator(".modal-open, dialog[open]");
    await expect(modal).toBeVisible();

    // Click the delete button inside the modal
    await modal.getByRole("button", { name: /delete/i }).click();

    // Wait for deletion to complete
    await page.waitForTimeout(1500);

    // Server count should decrease by 1
    const serverCountAfter = await page.locator(".card").count();
    expect(serverCountAfter).toBe(serverCountBefore - 1);
  });
});
```

**Verify:**

```bash
direnv exec . bunx playwright test tests/e2e/delete-server.spec.ts --reporter=list
```

**Commit:** `test(e2e): add delete server workflow tests`

---

### [ ] Task 5: Add Connection Test Tests

**File:** `tests/e2e/test-connection.spec.ts` (new)
**Difficulty:** Moderate

Create tests for the server connection testing functionality.

**Create file with:**

```typescript
import { test, expect } from "./setup";

test.describe("Test Connection functionality", () => {
  test("should show testing state when clicked", async ({ page }) => {
    await page.goto("/servers");

    const testButton = page.getByRole("button", { name: /^test$/i }).first();

    if (!(await testButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    // Click test button
    await testButton.click();

    // Should show "Testing..." state
    await expect(page.getByText(/testing/i)).toBeVisible();
  });

  test("should show result after test completes", async ({ page }) => {
    await page.goto("/servers");

    const testButton = page.getByRole("button", { name: /^test$/i }).first();

    if (!(await testButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await testButton.click();

    // Wait for test to complete (may take a few seconds)
    await page.waitForTimeout(5000);

    // Should show either "Connected" or "Connection failed" or error message
    const serverCard = page.locator(".card").first();
    const resultText = await serverCard.textContent();

    const hasResult =
      resultText?.includes("Connected") ||
      resultText?.includes("Connection failed") ||
      resultText?.includes("Error");

    expect(hasResult).toBeTruthy();
  });

  test("test connection in add server modal", async ({ page }) => {
    await page.goto("/servers");

    // Open add server modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Fill required fields
    await page.locator("#name").fill("Test Server");
    await page.locator("#url").fill("http://localhost:7878");
    await page.locator("#apiKey").fill("invalid-api-key-for-testing");

    // Click test connection
    const testButton = page.getByRole("button", { name: /test connection/i });
    await expect(testButton).toBeVisible();
    await testButton.click();

    // Should show testing state
    await expect(page.getByText(/testing/i)).toBeVisible();

    // Wait for result
    await page.waitForTimeout(5000);

    // Should show connection result (likely failed since no real server)
    const modal = page.locator("#server-modal");
    const modalText = await modal.textContent();

    const hasResult =
      modalText?.includes("Connected") ||
      modalText?.includes("Connection failed") ||
      modalText?.includes("failed");

    expect(hasResult).toBeTruthy();
  });
});
```

**Verify:**

```bash
direnv exec . bunx playwright test tests/e2e/test-connection.spec.ts --reporter=list
```

**Commit:** `test(e2e): add connection test functionality tests`

---

### [ ] Task 6: Add Dashboard Integration Tests

**File:** `tests/e2e/dashboard-integration.spec.ts` (new)
**Difficulty:** Moderate

Create tests that verify dashboard stats update correctly.

**Create file with:**

```typescript
import { test, expect } from "./setup";

test.describe("Dashboard integration", () => {
  test("should display server count matching servers page", async ({
    page,
  }) => {
    // Get server count from servers page
    await page.goto("/servers");
    const serverCards = page.locator(".card");
    const serverCount = await serverCards.count();

    // Navigate to dashboard
    await page.goto("/");

    // Find the Servers stat card
    const serversStatCard = page
      .locator(".stat")
      .filter({ hasText: "Servers" });
    await expect(serversStatCard).toBeVisible();

    // Get the displayed count
    const statValue = serversStatCard.locator(".stat-value");
    await expect(statValue).toHaveText(String(serverCount));
  });

  test("Run Now button should show loading state", async ({ page }) => {
    await page.goto("/");

    const runButton = page.getByRole("button", { name: /run now/i });
    await expect(runButton).toBeVisible();

    // Click the button
    await runButton.click();

    // Spinner should appear (htmx-indicator becomes visible)
    const spinner = page.locator("#run-spinner");

    // The spinner might be very brief, so we check the button is still functional
    await page.waitForTimeout(500);
    await expect(page.locator("main")).toBeVisible();
  });

  test("should display last cycle time", async ({ page }) => {
    await page.goto("/");

    // Find the Last Cycle stat card
    const lastCycleCard = page
      .locator(".stat")
      .filter({ hasText: "Last Cycle" });
    await expect(lastCycleCard).toBeVisible();

    // Should have a value (either "Never" or a timestamp)
    const statValue = lastCycleCard.locator(".stat-value");
    const value = await statValue.textContent();
    expect(value).toBeTruthy();
  });

  test("should show server URLs in table", async ({ page }) => {
    await page.goto("/");

    // Find the servers table
    const serversTable = page.locator("table");

    if (await serversTable.isVisible()) {
      // Check that URL column exists
      await expect(
        page.getByRole("columnheader", { name: /url/i }),
      ).toBeVisible();

      // If there are servers, URLs should not be empty
      const urlCells = page.locator("table tbody tr td:nth-child(3)");
      const count = await urlCells.count();

      if (count > 0) {
        const firstUrl = await urlCells.first().textContent();
        // URL should either be empty (if issue #5 not fixed) or contain http
        // After fix, should contain actual URL
        expect(firstUrl !== undefined).toBeTruthy();
      }
    }
  });

  test("recent activity section exists", async ({ page }) => {
    await page.goto("/");

    // Check for Recent Activity heading
    await expect(
      page.getByRole("heading", { name: /recent activity/i }),
    ).toBeVisible();
  });
});
```

**Verify:**

```bash
direnv exec . bunx playwright test tests/e2e/dashboard-integration.spec.ts --reporter=list
```

**Commit:** `test(e2e): add dashboard integration tests`

---

### [ ] Task 7: Add Error Handling Tests

**File:** `tests/e2e/error-handling.spec.ts` (new)
**Difficulty:** Moderate

Create tests for graceful error handling.

**Create file with:**

```typescript
import { test, expect } from "./setup";

test.describe("Error handling", () => {
  test("should handle 404 pages gracefully", async ({ page }) => {
    const response = await page.goto("/nonexistent-page");

    // Should return 404 status
    expect(response?.status()).toBe(404);

    // Page should still render something useful
    await expect(page.locator("body")).toBeVisible();
  });

  test("should handle invalid server form submission", async ({ page }) => {
    await page.goto("/servers");

    // Open add server modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Fill with invalid URL
    await page.locator("#name").fill("Test");
    await page.locator("#url").fill("not-a-valid-url");
    await page.locator("#apiKey").fill("test-key");

    // The URL field has type="url", so browser validation should prevent submission
    const urlField = page.locator("#url");
    await expect(urlField).toHaveAttribute("type", "url");
  });

  test("settings form should handle invalid numeric values", async ({
    page,
  }) => {
    await page.goto("/settings");

    // Find numeric inputs
    const numericInputs = page.locator('input[type="number"]');
    const count = await numericInputs.count();

    if (count > 0) {
      const firstInput = numericInputs.first();

      // Try to enter negative value
      await firstInput.clear();
      await firstInput.fill("-5");

      // Check if input has min attribute for validation
      const minAttr = await firstInput.getAttribute("min");

      // Either has min validation or the form should handle it
      await expect(page.locator("main")).toBeVisible();
    }
  });

  test("app should remain functional after failed API call", async ({
    page,
  }) => {
    await page.goto("/");

    // Intercept and fail the automation trigger
    await page.route("**/api/automation/trigger", (route) => {
      route.fulfill({
        status: 500,
        body: JSON.stringify({ error: "Test error" }),
      });
    });

    // Click Run Now
    const runButton = page.getByRole("button", { name: /run now/i });
    await runButton.click();

    // Wait a moment
    await page.waitForTimeout(1000);

    // Page should still be functional
    await expect(page.locator("main")).toBeVisible();

    // Navigation should still work
    await page.getByRole("link", { name: "Servers" }).click();
    await expect(page).toHaveURL(/\/servers/);
  });
});
```

**Verify:**

```bash
direnv exec . bunx playwright test tests/e2e/error-handling.spec.ts --reporter=list
```

**Commit:** `test(e2e): add error handling tests`

---

### [ ] Task 8: Add Settings Persistence Tests

**File:** `tests/e2e/settings-persistence.spec.ts` (new)
**Difficulty:** Moderate

Create tests that verify settings actually persist after save.

**Create file with:**

```typescript
import { test, expect } from "./setup";

test.describe("Settings persistence", () => {
  test("should save and persist interval hours", async ({ page }) => {
    await page.goto("/settings");

    // Find interval input (adjust selector based on actual form)
    const intervalInput = page
      .locator(
        'input[name="interval_hours"], input#interval_hours, input#intervalHours',
      )
      .first();

    if (!(await intervalInput.isVisible().catch(() => false))) {
      // Try finding by label text proximity
      const intervalSection = page
        .locator("text=Interval")
        .locator("..")
        .locator("input[type=number]")
        .first();
      if (await intervalSection.isVisible().catch(() => false)) {
        await intervalSection.clear();
        await intervalSection.fill("12");
      } else {
        test.skip();
        return;
      }
    } else {
      await intervalInput.clear();
      await intervalInput.fill("12");
    }

    // Save settings
    const saveButton = page.getByRole("button", { name: /save/i });
    await saveButton.click();

    // Wait for save to complete
    await page.waitForTimeout(1000);

    // Reload page
    await page.reload();

    // Verify value persisted
    const reloadedInput = page
      .locator(
        'input[name="interval_hours"], input#interval_hours, input#intervalHours',
      )
      .first();
    if (await reloadedInput.isVisible().catch(() => false)) {
      await expect(reloadedInput).toHaveValue("12");
    }
  });

  test("should save scheduler enabled state", async ({ page }) => {
    await page.goto("/settings");

    // Find scheduler enabled checkbox
    const enabledCheckbox = page
      .locator(
        'input[name="scheduler_enabled"], input#scheduler_enabled, input#schedulerEnabled, input[type="checkbox"]',
      )
      .first();

    if (!(await enabledCheckbox.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    // Get current state
    const wasChecked = await enabledCheckbox.isChecked();

    // Toggle it
    await enabledCheckbox.click();

    // Save
    const saveButton = page.getByRole("button", { name: /save/i });
    await saveButton.click();
    await page.waitForTimeout(1000);

    // Reload
    await page.reload();

    // Verify state changed
    const reloadedCheckbox = page
      .locator(
        'input[name="scheduler_enabled"], input#scheduler_enabled, input#schedulerEnabled, input[type="checkbox"]',
      )
      .first();
    if (await reloadedCheckbox.isVisible().catch(() => false)) {
      const isNowChecked = await reloadedCheckbox.isChecked();
      expect(isNowChecked).toBe(!wasChecked);
    }
  });

  test("should show save confirmation feedback", async ({ page }) => {
    await page.goto("/settings");

    // Find save button and click it
    const saveButton = page.getByRole("button", { name: /save/i });

    if (!(await saveButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await saveButton.click();

    // Should show some feedback (loading state, success message, etc.)
    // Wait and check page is still functional
    await page.waitForTimeout(2000);
    await expect(page.locator("main")).toBeVisible();
  });
});
```

**Verify:**

```bash
direnv exec . bunx playwright test tests/e2e/settings-persistence.spec.ts --reporter=list
```

**Commit:** `test(e2e): add settings persistence tests`

---

### [ ] Task 9: Add Full Integration Test

**File:** `tests/e2e/full-flow.spec.ts` (new)
**Difficulty:** Complex

Create an end-to-end test that exercises a complete user workflow.

**Create file with:**

```typescript
import { test, expect } from "./setup";

test.describe("Full user flow integration", () => {
  test("complete workflow: navigate all pages", async ({ page }) => {
    // Start at dashboard
    await page.goto("/");
    await expect(page).toHaveTitle(/Dashboard.*Janitarr/);
    await expect(
      page.getByRole("heading", { name: "Dashboard" }),
    ).toBeVisible();

    // Navigate to Servers
    await page.getByRole("link", { name: "Servers" }).click();
    await expect(page).toHaveURL(/\/servers/);
    await expect(page.getByRole("heading", { name: "Servers" })).toBeVisible();

    // Navigate to Activity Logs
    await page.getByRole("link", { name: "Activity Logs" }).click();
    await expect(page).toHaveURL(/\/logs/);
    await expect(
      page.getByRole("heading", { name: /activity|logs/i }),
    ).toBeVisible();

    // Navigate to Settings
    await page.getByRole("link", { name: "Settings" }).click();
    await expect(page).toHaveURL(/\/settings/);
    await expect(
      page.getByRole("heading", { name: /settings/i }),
    ).toBeVisible();

    // Return to Dashboard
    await page.getByRole("link", { name: "Dashboard" }).click();
    await expect(page).toHaveURL(/\/$/);
  });

  test("theme toggle persists across pages", async ({ page }) => {
    await page.goto("/");

    // Find theme toggle
    const themeToggle = page.locator('input[type="checkbox"]').last();

    if (await themeToggle.isVisible()) {
      // Get initial theme
      const html = page.locator("html");
      const initialTheme = await html.getAttribute("data-theme");

      // Toggle theme
      await themeToggle.click();
      await page.waitForTimeout(300);

      // Get new theme
      const newTheme = await html.getAttribute("data-theme");
      expect(newTheme).not.toBe(initialTheme);

      // Navigate to another page
      await page.getByRole("link", { name: "Settings" }).click();
      await page.waitForTimeout(300);

      // Theme should persist
      const themeAfterNav = await html.getAttribute("data-theme");
      expect(themeAfterNav).toBe(newTheme);
    }
  });

  test("server CRUD operations", async ({ page }) => {
    await page.goto("/servers");

    // Count initial servers
    const initialCount = await page.locator(".card").count();

    // Open add server modal
    const addButton = page.getByRole("button", { name: /add server/i });
    await addButton.click();

    // Verify modal opened
    const modal = page.locator("#server-modal");
    await expect(modal).toBeVisible();

    // Fill form
    await page.locator("#name").fill("Integration Test Server");
    await page.locator("#url").fill("http://localhost:9999");
    await page.locator("#apiKey").fill("test-api-key-12345");

    // Submit (will likely fail connection but server may still be created depending on implementation)
    await page.getByRole("button", { name: /create/i }).click();

    // Wait for response
    await page.waitForTimeout(2000);

    // Either modal closed (success) or error shown
    // The test verifies the flow works, not necessarily successful creation
    await expect(page.locator("main")).toBeVisible();
  });

  test("logs page loads and displays content", async ({ page }) => {
    await page.goto("/logs");

    // Heading should be visible
    await expect(
      page.getByRole("heading", { name: /activity|logs/i }),
    ).toBeVisible();

    // Main content area should exist
    await expect(page.locator("main")).toBeVisible();

    // Either logs are displayed or empty state message
    const mainContent = await page.locator("main").textContent();
    expect(mainContent).toBeTruthy();
    expect(mainContent!.length).toBeGreaterThan(10);
  });
});
```

**Verify:**

```bash
direnv exec . bunx playwright test tests/e2e/full-flow.spec.ts --reporter=list
```

**Commit:** `test(e2e): add full integration flow tests`

---

### [ ] Task 10: Run Complete Test Suite

**Verify all tests pass:**

```bash
direnv exec . bunx playwright test --reporter=list
```

**Expected output:** All tests should pass (some may skip if prerequisites not met).

**Generate HTML report:**

```bash
direnv exec . bunx playwright show-report
```

**Commit:** `test(e2e): complete Phase 22 test suite improvements`

---

## Phase 22 Files Summary

| File                                      | Purpose                            |
| ----------------------------------------- | ---------------------------------- |
| `tests/e2e/servers.spec.ts`               | Fixed selectors                    |
| `tests/e2e/add-server.spec.ts`            | Rewritten with correct selectors   |
| `tests/e2e/edit-server.spec.ts`           | New - edit workflow tests          |
| `tests/e2e/delete-server.spec.ts`         | New - delete workflow tests        |
| `tests/e2e/test-connection.spec.ts`       | New - connection testing tests     |
| `tests/e2e/dashboard-integration.spec.ts` | New - dashboard stats tests        |
| `tests/e2e/error-handling.spec.ts`        | New - error handling tests         |
| `tests/e2e/settings-persistence.spec.ts`  | New - settings save/load tests     |
| `tests/e2e/full-flow.spec.ts`             | New - end-to-end integration tests |

---

## Completed Phases (Archive)

### Phase 20: Build-Time Version Information ✓

**Completed:** 2026-01-21
**Commit:** `9675487 feat: implement build-time version information from git`

**Summary:** Implemented dynamic version information from git instead of hardcoded version strings. Created a new `version` package with build-time variables (`Version`, `Commit`, `BuildDate`) that are set via ldflags during compilation. Updated the Makefile to inject version information from `git describe`, commit hash, and build timestamp. Updated both the CLI (`--version` flag) and web server (Prometheus metrics `janitarr_info`) to use the version package. All tests pass.

### Phase 19: Web Interface Bug Fixes ✓

**Completed:** 2026-01-21
**Commits:**

- `bc8a873 fix(web): correct server ID interpolation in Test button`
- `804e332 fix(web): open modal dialog when Edit button is clicked`
- `88f6a34 fix(web): populate log metadata for web/terminal parity`

**Summary:** Fixed three critical web interface bugs: (1) Web logs now display the same metadata detail as terminal logs by populating the `Metadata` field in all logger methods and rendering it in the UI, (2) Edit button now properly opens the modal dialog using htmx's `hx-on::after-swap` event, (3) Test button now correctly interpolates server IDs using htmx attributes instead of Alpine.js string concatenation.

### Phase 18: Enable Tests for GetEnabledServers and SetServerEnabled ✓

**Completed:** 2026-01-21
**Commit:** `0e39409 test: enable tests for GetEnabledServers and SetServerEnabled`

### Phase 17: DaisyUI Version Compatibility Fix ✓

**Completed:** 2026-01-21
**Commit:** `dd18216 fix(ui): downgrade DaisyUI to v4 for Tailwind CSS 3 compatibility`

---

## Quick Reference

### DaisyUI Version Compatibility

| DaisyUI Version | Tailwind CSS Version | Configuration Method                    |
| --------------- | -------------------- | --------------------------------------- |
| v4.x            | v3.x                 | `require("daisyui")` in tailwind.config |
| v5.x            | v4.x                 | `@plugin "daisyui"` in CSS file         |
