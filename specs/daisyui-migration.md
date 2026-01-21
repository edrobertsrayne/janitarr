# DaisyUI Migration Specification

## Overview

This specification describes the migration of Janitarr's web frontend from raw Tailwind CSS utility classes to DaisyUI, a Tailwind CSS component library. The migration includes implementing a simple light/dark mode toggle in the navigation sidebar with custom theme definitions for future customization.

## Goals

1. **Consistency**: Replace ad-hoc Tailwind styling with semantic DaisyUI component classes
2. **Theming**: Enable users to toggle between light and dark mode via a navigation sidebar switch
3. **Maintainability**: Reduce verbose inline classes with DaisyUI's semantic components
4. **Customizability**: Define custom "light" and "dark" themes (clones of DaisyUI defaults) for future color adjustments
5. **User Experience**: Provide a polished, cohesive look with simple light/dark preference

## Non-Goals

- Multiple theme selection (only light/dark toggle)
- Server-side theme persistence (localStorage is sufficient)
- Gradual/incremental migration (this is a big-bang replacement)

## Technical Approach

### Version Compatibility

> **CRITICAL WARNING:** DaisyUI v5 is NOT compatible with Tailwind CSS v3!
>
> If you install the wrong version, DaisyUI classes will have NO styling. The UI will render as plain unstyled HTML. This is a **silent failure** - no errors are shown during build or at runtime.

**IMPORTANT:** DaisyUI version must be compatible with the installed Tailwind CSS version:

| DaisyUI Version | Tailwind CSS Version | Configuration Method                    |
| --------------- | -------------------- | --------------------------------------- |
| v4.x            | v3.x                 | `require("daisyui")` in tailwind.config |
| v5.x            | v4.x                 | `@plugin "daisyui"` in CSS file         |

This project uses **Tailwind CSS v3**, so **DaisyUI v4.x** must be used.

### DaisyUI Installation

Add DaisyUI v4 as a Tailwind plugin:

```bash
npm install -D daisyui@^4.12.24
# or with bun:
bun add -D daisyui@^4.12.24
```

**DO NOT use `daisyui@latest`** - this installs v5 which requires Tailwind CSS v4.

### Verification After Installation

Always verify the correct version is installed and CSS is compiled properly:

```bash
# Check DaisyUI version (must be 4.x.x, NOT 5.x.x)
cat node_modules/daisyui/package.json | grep '"version"'

# Rebuild CSS
make generate

# Verify DaisyUI classes are in compiled CSS
grep -c "btn" static/css/app.css       # Should be > 100
grep -c "drawer" static/css/app.css    # Should be > 10
grep -c "data-theme" static/css/app.css # Should be > 0
grep -c "base-100" static/css/app.css   # Should be > 0
```

If any grep returns 0, DaisyUI is not working. Check the version.

Update `tailwind.config.cjs` with custom theme definitions (DaisyUI v4 format):

```javascript
module.exports = {
  content: ["./src/templates/**/*.templ", "./src/templates/**/*_templ.go"],
  theme: { extend: {} },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        // Custom light theme (clone of DaisyUI light for future customization)
        light: {
          primary: "#570df8",
          secondary: "#f000b8",
          accent: "#37cdbe",
          neutral: "#3d4451",
          "base-100": "#ffffff",
          "base-200": "#f9fafb",
          "base-300": "#d1d5db",
          "base-content": "#1f2937",
          info: "#3abff8",
          success: "#36d399",
          warning: "#fbbd23",
          error: "#f87272",
        },
      },
      {
        // Custom dark theme (clone of DaisyUI dark for future customization)
        dark: {
          primary: "#661ae6",
          secondary: "#d926aa",
          accent: "#1fb2a5",
          neutral: "#2a323c",
          "base-100": "#1d232a",
          "base-200": "#191e24",
          "base-300": "#15191e",
          "base-content": "#a6adba",
          info: "#3abff8",
          success: "#36d399",
          warning: "#fbbd23",
          error: "#f87272",
        },
      },
    ],
    darkTheme: "dark", // Default for @media(prefers-color-scheme: dark)
  },
};
```

**Note:** DaisyUI v4 uses simple hex color values, not oklch() or CSS custom properties. The v5 syntax (oklch colors, `--color-*` variables) will NOT work with Tailwind CSS v3.

### Theme System

#### Available Themes

- **light**: Custom light theme (clone of DaisyUI light)
- **dark**: Custom dark theme (clone of DaisyUI dark)

#### Default Behavior

1. On first visit (no localStorage value): use "dark" theme
2. When user toggles: persist choice in localStorage as "light" or "dark"
3. On subsequent visits: load theme from localStorage

#### Theme Persistence

```javascript
// Theme initialization in base.templ
document.documentElement.setAttribute(
  "data-theme",
  localStorage.getItem("janitarr-theme") || "dark",
);
```

### Component Migration Strategy

Use DaisyUI classes directly for most elements. Create templ helper components only for complex patterns (modals) to reduce boilerplate.

#### Mapping: Current Classes → DaisyUI

| Element          | Current Approach                                                                                       | DaisyUI Replacement             |
| ---------------- | ------------------------------------------------------------------------------------------------------ | ------------------------------- |
| Primary Button   | `px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-500...`                    | `btn btn-primary`               |
| Secondary Button | `px-4 py-2 bg-gray-200 text-gray-800 rounded-lg...`                                                    | `btn btn-ghost`                 |
| Danger Button    | `px-4 py-2 bg-red-600 text-white...`                                                                   | `btn btn-error`                 |
| Text Input       | `mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500...`                      | `input input-bordered w-full`   |
| Select           | Similar verbose classes                                                                                | `select select-bordered w-full` |
| Checkbox         | Custom styling                                                                                         | `checkbox checkbox-primary`     |
| Card             | `bg-white dark:bg-gray-800 rounded-lg shadow p-6`                                                      | `card bg-base-100 shadow-xl`    |
| Badge (Radarr)   | `inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800...` | `badge badge-primary`           |
| Badge (Sonarr)   | `...bg-purple-100 text-purple-800...`                                                                  | `badge badge-secondary`         |
| Badge (Enabled)  | `...bg-green-100 text-green-800...`                                                                    | `badge badge-success`           |
| Badge (Disabled) | `...bg-gray-100 text-gray-800...`                                                                      | `badge badge-ghost`             |
| Badge (Error)    | `...bg-red-100 text-red-800...`                                                                        | `badge badge-error`             |
| Modal            | Custom Alpine.js + Tailwind                                                                            | `modal` + `modal-box`           |
| Toast Success    | `.toast-success` custom CSS                                                                            | `alert alert-success`           |
| Toast Error      | `.toast-error` custom CSS                                                                              | `alert alert-error`             |
| Navigation       | Custom sidebar styling                                                                                 | `menu` + `drawer`               |
| Stats Card       | Custom card styling                                                                                    | `stat` component                |

### File Changes

#### 1. Configuration Files

**`tailwind.config.cjs`** - Add DaisyUI plugin and theme configuration

**`static/css/input.css`** - Remove custom `.toast` classes (replaced by DaisyUI alerts)

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

/* No custom styles needed - DaisyUI provides everything */
```

#### 2. Base Layout

**`src/templates/layouts/base.templ`**

- Remove `darkMode` Alpine.js state
- Remove `:class="{ 'dark': darkMode }"` binding
- Add `data-theme` attribute with localStorage initialization
- Update body classes to use DaisyUI's `bg-base-100` instead of `bg-gray-100 dark:bg-gray-900`

```go
templ Base(title string) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
        <title>{ title } - Janitarr</title>
        <link href="/static/css/app.css" rel="stylesheet"/>
        <script src="/static/js/htmx.min.js"></script>
        <script src="/static/js/alpine.min.js" defer></script>
        <script>
            document.documentElement.setAttribute(
                'data-theme',
                localStorage.getItem('janitarr-theme') || 'dark'
            );
        </script>
    </head>
    <body class="bg-base-100 min-h-screen">
        // ... drawer structure wrapping content
    </body>
    </html>
}
```

#### 3. Navigation Component

**`src/templates/components/nav.templ`**

Convert to DaisyUI drawer + menu pattern with light/dark toggle:

```go
templ Nav(currentPath string) {
    <div class="drawer lg:drawer-open">
        <input id="nav-drawer" type="checkbox" class="drawer-toggle"/>
        <div class="drawer-content">
            // Mobile navbar with hamburger
            <div class="navbar bg-base-100 lg:hidden">
                <div class="flex-none">
                    <label for="nav-drawer" class="btn btn-square btn-ghost">
                        <svg>...</svg> // Hamburger icon
                    </label>
                </div>
                <div class="flex-1">
                    <span class="text-xl font-bold">Janitarr</span>
                </div>
            </div>
            // Main content slot
            { children... }
        </div>
        <div class="drawer-side">
            <label for="nav-drawer" class="drawer-overlay"></label>
            <aside class="bg-base-200 w-64 min-h-full flex flex-col">
                <div class="p-4 text-xl font-bold">Janitarr</div>
                <ul class="menu p-4 gap-2 flex-1">
                    @NavItem("/", "Dashboard", currentPath)
                    @NavItem("/servers", "Servers", currentPath)
                    @NavItem("/logs", "Activity Logs", currentPath)
                    @NavItem("/settings", "Settings", currentPath)
                </ul>
                // Light/Dark toggle at bottom of sidebar
                @ThemeToggle()
            </aside>
        </div>
    </div>
}

templ NavItem(href, label, currentPath string) {
    <li>
        <a href={ templ.SafeURL(href) }
           class={ templ.KV("active", href == currentPath) }>
            { label }
        </a>
    </li>
}

templ ThemeToggle() {
    <div class="p-4 border-t border-base-300">
        <label class="flex items-center gap-3 cursor-pointer"
               x-data="{ isDark: localStorage.getItem('janitarr-theme') !== 'light' }"
               x-init="$watch('isDark', val => {
                   const theme = val ? 'dark' : 'light';
                   localStorage.setItem('janitarr-theme', theme);
                   document.documentElement.setAttribute('data-theme', theme);
               })">
            // Sun icon (light mode)
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"/>
            </svg>
            <input type="checkbox" class="toggle toggle-sm"
                   x-model="isDark"/>
            // Moon icon (dark mode)
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"/>
            </svg>
        </label>
    </div>
}
```

- Light/dark toggle placed at bottom of navigation sidebar
- Uses DaisyUI `toggle` component with sun/moon icons
- Alpine.js manages state and persists to localStorage
- Toggle is checked (right position) for dark mode, unchecked (left position) for light mode

#### 4. Server Card Component

**`src/templates/components/server_card.templ`**

```go
templ ServerCard(server Server) {
    <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
            <div class="flex items-center justify-between">
                <h2 class="card-title">{ server.Name }</h2>
                @ServerTypeBadge(server.Type)
            </div>
            <p class="text-base-content/70">{ server.URL }</p>
            <div class="flex items-center gap-2">
                @StatusBadge(server.Enabled, server.LastError)
            </div>
            <div class="card-actions justify-end">
                <button class="btn btn-ghost btn-sm" ...>Test</button>
                <button class="btn btn-ghost btn-sm" ...>Edit</button>
                <button class="btn btn-ghost btn-sm text-error" ...>Delete</button>
            </div>
        </div>
    </div>
}

templ ServerTypeBadge(serverType string) {
    if serverType == "radarr" {
        <span class="badge badge-primary">Radarr</span>
    } else {
        <span class="badge badge-secondary">Sonarr</span>
    }
}

templ StatusBadge(enabled bool, lastError string) {
    if !enabled {
        <span class="badge badge-ghost">Disabled</span>
    } else if lastError != "" {
        <span class="badge badge-error">Error</span>
    } else {
        <span class="badge badge-success">Connected</span>
    }
}
```

#### 5. Stats Card Component

**`src/templates/components/stats_card.templ`**

Convert to DaisyUI stat component:

```go
templ StatsCard(title, value, description string) {
    <div class="stat bg-base-100 rounded-box shadow">
        <div class="stat-title">{ title }</div>
        <div class="stat-value">{ value }</div>
        <div class="stat-desc">{ description }</div>
    </div>
}
```

#### 6. Form Components

**`src/templates/components/forms/server_form.templ`**

```go
templ ServerForm(server *Server) {
    <dialog id="server-modal" class="modal">
        <div class="modal-box">
            <h3 class="font-bold text-lg">
                if server == nil {
                    Add Server
                } else {
                    Edit Server
                }
            </h3>
            <form method="dialog" class="space-y-4 mt-4">
                <div class="form-control w-full">
                    <label class="label">
                        <span class="label-text">Name</span>
                    </label>
                    <input type="text" name="name"
                           class="input input-bordered w-full"
                           required/>
                </div>

                <div class="form-control w-full">
                    <label class="label">
                        <span class="label-text">Type</span>
                    </label>
                    <select name="type" class="select select-bordered w-full">
                        <option value="radarr">Radarr</option>
                        <option value="sonarr">Sonarr</option>
                    </select>
                </div>

                <div class="form-control w-full">
                    <label class="label">
                        <span class="label-text">URL</span>
                    </label>
                    <input type="url" name="url"
                           class="input input-bordered w-full"
                           placeholder="http://localhost:7878"
                           required/>
                </div>

                <div class="form-control w-full">
                    <label class="label">
                        <span class="label-text">API Key</span>
                    </label>
                    <input type="password" name="apiKey"
                           class="input input-bordered w-full"
                           required/>
                </div>

                <div class="form-control">
                    <label class="label cursor-pointer justify-start gap-4">
                        <input type="checkbox" name="enabled"
                               class="checkbox checkbox-primary"
                               checked/>
                        <span class="label-text">Enabled</span>
                    </label>
                </div>

                <div class="modal-action">
                    <button type="button" class="btn btn-ghost"
                            onclick="this.closest('dialog').close()">
                        Cancel
                    </button>
                    <button type="submit" class="btn btn-primary">
                        Save
                    </button>
                </div>
            </form>
        </div>
        <form method="dialog" class="modal-backdrop">
            <button>close</button>
        </form>
    </dialog>
}
```

**`src/templates/components/forms/config_form.templ`**

- Convert all inputs to DaisyUI form controls
- Use `input input-bordered`, `select select-bordered`, `checkbox`
- Group related settings in cards with `card bg-base-100`

#### 7. Log Entry Component

**`src/templates/components/log_entry.templ`**

```go
templ LogEntry(log LogEntry) {
    <div class="card bg-base-100 shadow-sm">
        <div class="card-body p-4">
            <div class="flex items-center gap-2">
                @LogIcon(log.Type)
                @LogTypeBadge(log.Type)
                <span class="text-base-content/70 text-sm">{ log.ServerName }</span>
                <span class="text-base-content/50 text-sm ml-auto">{ log.Timestamp }</span>
            </div>
            <p>{ log.Message }</p>
        </div>
    </div>
}

templ LogTypeBadge(logType string) {
    switch logType {
    case "cycle_start":
        <span class="badge badge-info">Cycle Start</span>
    case "cycle_end":
        <span class="badge badge-success">Cycle End</span>
    case "search":
        <span class="badge badge-primary">Search</span>
    case "error":
        <span class="badge badge-error">Error</span>
    default:
        <span class="badge">{ logType }</span>
    }
}
```

#### 8. Pages

**`src/templates/pages/dashboard.templ`**

- Wrap stats in `<div class="stats stats-vertical lg:stats-horizontal shadow">` or individual stat cards
- Use card components for server list section
- Use DaisyUI timeline or card list for recent activity

**`src/templates/pages/servers.templ`**

- Use `btn btn-primary` for "Add Server" button
- Grid of server cards using converted `ServerCard` component

**`src/templates/pages/logs.templ`**

- Toolbar: use `join` for grouped filter buttons, `select select-bordered` for dropdowns
- Log list: converted `LogEntry` components

**`src/templates/pages/settings.templ`**

- Each settings section in a `card bg-base-100 shadow-xl`
- No theme selector needed (toggle is in navigation sidebar)

### Toast Notifications

Replace custom toast CSS with DaisyUI toast + alert:

```go
templ Toast(message string, toastType string) {
    <div class="toast toast-end">
        <div class={ "alert", templ.KV("alert-success", toastType == "success"),
                              templ.KV("alert-error", toastType == "error") }>
            <span>{ message }</span>
        </div>
    </div>
}
```

## Implementation Plan

### Phase 1: Setup and Configuration

1. Install DaisyUI v4: `npm install -D daisyui@^4.12.24` (**NOT `@latest`** which installs v5)
2. Update `tailwind.config.cjs` with DaisyUI plugin and theme config
3. Remove custom CSS from `static/css/input.css`
4. Run `make generate` to rebuild CSS
5. Verify DaisyUI classes in compiled CSS: `grep -c "btn" static/css/app.css` should return > 100

### Phase 2: Base Layout and Navigation

1. Update `base.templ` with theme initialization script (default to "dark")
2. Convert `nav.templ` to drawer + menu pattern
3. Add light/dark theme toggle to navigation sidebar

### Phase 3: Component Migration

1. Convert `stats_card.templ` to DaisyUI stat
2. Convert `server_card.templ` to DaisyUI card + badges
3. Convert `log_entry.templ` to DaisyUI card + badges
4. Convert `server_form.templ` to DaisyUI modal + form controls
5. Convert `config_form.templ` to DaisyUI form controls

### Phase 4: Page Migration

1. Update `dashboard.templ` with converted components
2. Update `servers.templ` with converted components
3. Update `logs.templ` with converted components
4. Update `settings.templ` with converted components (no theme selector needed)

### Phase 5: Testing and Polish

1. Test light and dark themes render correctly
2. Verify theme persistence across page reloads
3. Test responsive behavior (drawer collapse on mobile)
4. Verify all htmx interactions still work
5. Run existing Playwright E2E tests

## Acceptance Criteria

1. **Theme Toggle**: Users can toggle between light and dark mode via navigation sidebar switch
2. **Theme Persistence**: Selected theme persists across browser sessions via localStorage
3. **Default Theme**: "Dark" theme is applied on first visit before user toggles
4. **Component Consistency**: All UI elements use DaisyUI component classes
5. **No Visual Regression**: All existing functionality works correctly
6. **Responsive Design**: Drawer collapses to hamburger menu on mobile
7. **Build Success**: `make build` completes without errors
8. **Tests Pass**: All existing tests continue to pass
9. **Custom Themes**: Custom "light" and "dark" themes defined in tailwind.config.cjs for future customization

## Dependencies

### NPM (dev only)

```json
{
  "devDependencies": {
    "tailwindcss": "^3.4.0",
    "daisyui": "^4.12.24"
  }
}
```

**CRITICAL:** DaisyUI v4.x is required for Tailwind CSS v3.x compatibility. Do NOT use v5.x.

### Bundle Size Impact

DaisyUI adds ~10-15KB to the compiled CSS (with only 2 custom themes enabled). This is minimal given:

- CSS is cached after first load
- Replaces significant custom styling
- Only 2 themes (light/dark) instead of 35

## Risks and Mitigations

| Risk                                           | Impact             | Mitigation                                                            |
| ---------------------------------------------- | ------------------ | --------------------------------------------------------------------- |
| DaisyUI v5 installed with Tailwind CSS v3      | No styles rendered | Pin DaisyUI to v4.x; verify CSS contains DaisyUI classes after build  |
| DaisyUI class conflicts with existing Tailwind | Broken styling     | Big-bang migration removes old classes entirely                       |
| Theme switch causes layout shift               | Poor UX            | Theme is set synchronously before render via inline script            |
| Breaking htmx/Alpine.js interactions           | Lost functionality | Test thoroughly; DaisyUI is class-based and doesn't interfere with JS |

## Pattern Mismatches and Resolutions

The current UI uses several patterns that differ from DaisyUI's approach. This section documents each mismatch and how to resolve it using DaisyUI components.

### 1. Custom Modal Implementation

**Location:** `server_form.templ:6-148`

**Current:** Uses a `fixed inset-0` overlay with manual positioning and Alpine.js for show/hide:

```html
<div x-show="showModal" class="fixed inset-0 bg-gray-500 bg-opacity-75">
  <div class="bg-white rounded-lg">...</div>
</div>
```

**Resolution:** Convert to DaisyUI's dialog-based modal:

```html
<dialog id="server-modal" class="modal">
  <div class="modal-box">
    <!-- form content -->
  </div>
  <form method="dialog" class="modal-backdrop">
    <button>close</button>
  </form>
</dialog>
```

Open with `document.getElementById('server-modal').showModal()`. The native `<dialog>` element handles backdrop, focus trapping, and escape-to-close automatically.

---

### 2. Loading Spinners

**Locations:** `server_form.templ:133-136`, `config_form.templ:153-156`, `dashboard.templ:45-48`

**Current:** Custom SVG spinner with Tailwind animation:

```html
<svg class="animate-spin h-5 w-5">
  <circle
    class="opacity-25"
    cx="12"
    cy="12"
    r="10"
    stroke="currentColor"
    stroke-width="4"
  />
  <path
    class="opacity-75"
    fill="currentColor"
    d="M4 12a8 8 0 018-8V0C5.373..."
  />
</svg>
```

**Resolution:** Use DaisyUI loading component:

```html
<span class="loading loading-spinner loading-sm"></span>
```

For htmx indicators, wrap in the indicator span:

```html
<span id="run-spinner" class="htmx-indicator">
  <span class="loading loading-spinner loading-sm"></span>
</span>
```

---

### 3. Table Styling

**Location:** `dashboard.templ:80-111`

**Current:** Manual table with verbose utility classes:

```html
<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
  <thead>
    <tr>
      <th
        class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase"
      >
        Name
      </th>
    </tr>
  </thead>
</table>
```

**Resolution:** Use DaisyUI table component:

```html
<table class="table">
  <thead>
    <tr>
      <th>Name</th>
      <th>Type</th>
      <th>URL</th>
      <th>Status</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>{ server.Name }</td>
      ...
    </tr>
  </tbody>
</table>
```

Optional modifiers: `table-zebra` for alternating rows, `table-pin-rows` for sticky headers.

---

### 4. Always-Visible Sidebar

**Location:** `nav.templ`, `base.templ:18-23`

**Current:** Sidebar is always visible with fixed width, no responsive collapse:

```html
<div class="flex h-full">
  <nav class="w-64 bg-gray-800">...</nav>
  <main class="flex-1">...</main>
</div>
```

**Resolution:** Use DaisyUI drawer with `lg:drawer-open` for desktop-permanent, mobile-collapsible:

```html
<div class="drawer lg:drawer-open">
  <input id="nav-drawer" type="checkbox" class="drawer-toggle" />
  <div class="drawer-content">
    <!-- Mobile navbar (hidden on lg:) -->
    <div class="navbar bg-base-100 lg:hidden">
      <label for="nav-drawer" class="btn btn-square btn-ghost">
        <svg><!-- hamburger icon --></svg>
      </label>
      <span class="text-xl font-bold">Janitarr</span>
    </div>
    <main class="p-6">{ children... }</main>
  </div>
  <div class="drawer-side">
    <label for="nav-drawer" class="drawer-overlay"></label>
    <aside class="bg-base-200 w-64 min-h-full">
      <ul class="menu p-4">
        ...
      </ul>
    </aside>
  </div>
</div>
```

---

### 5. Radio Button Group

**Location:** `server_form.templ:40-65`

**Current:** Custom flex container with styled radio inputs:

```html
<div class="flex space-x-4">
  <label class="inline-flex items-center">
    <input
      type="radio"
      class="form-radio text-blue-600"
      name="type"
      value="radarr"
    />
    <span class="ml-2">Radarr</span>
  </label>
</div>
```

**Resolution:** Use DaisyUI radio with form-control:

```html
<div class="form-control">
  <label class="label cursor-pointer justify-start gap-4">
    <input
      type="radio"
      name="type"
      value="radarr"
      class="radio radio-primary"
      checked
    />
    <span class="label-text">Radarr</span>
  </label>
</div>
<div class="form-control">
  <label class="label cursor-pointer justify-start gap-4">
    <input
      type="radio"
      name="type"
      value="sonarr"
      class="radio radio-secondary"
    />
    <span class="label-text">Sonarr</span>
  </label>
</div>
```

---

### 6. Section Headers with Bottom Borders

**Locations:** `dashboard.templ:69-72`, `dashboard.templ:117-119`

**Current:** Cards with internal border dividers:

```html
<div class="bg-white rounded-lg shadow">
  <div class="px-6 py-4 border-b border-gray-200">
    <h2 class="text-xl font-semibold">Servers</h2>
  </div>
  <div class="p-6">...</div>
</div>
```

**Resolution:** Use DaisyUI card with divider component:

```html
<div class="card bg-base-100 shadow-xl">
  <div class="card-body">
    <h2 class="card-title">Servers</h2>
    <div class="divider mt-0"></div>
    <!-- content -->
  </div>
</div>
```

---

### 7. Hardcoded Icon Colors

**Location:** `log_entry.templ:10-26`

**Current:** Explicit color classes that don't adapt to themes:

```html
<svg class="w-5 h-5 text-blue-600 dark:text-blue-400">
  <svg class="w-5 h-5 text-green-600 dark:text-green-400">
    <svg class="w-5 h-5 text-red-600 dark:text-red-400"></svg>
  </svg>
</svg>
```

**Resolution:** Use DaisyUI semantic colors:

```html
<svg class="w-5 h-5 text-info">
  <!-- cycle start -->
  <svg class="w-5 h-5 text-success">
    <!-- cycle end -->
    <svg class="w-5 h-5 text-secondary">
      <!-- search -->
      <svg class="w-5 h-5 text-error"><!-- error --></svg>
    </svg>
  </svg>
</svg>
```

These colors automatically adapt to the selected theme.

---

### 8. Empty State Pattern

**Locations:** `logs.templ:134-140`, `dashboard.templ:74-77`, `dashboard.templ:122-123`

**Current:** Custom centered div with icon and text:

```html
<div class="text-center py-8">
  <svg class="mx-auto h-12 w-12 text-gray-400">...</svg>
  <p class="mt-2 text-gray-500">No servers configured</p>
</div>
```

**Issue:** DaisyUI doesn't have a specific empty state component.

**Resolution:** Keep the pattern but use DaisyUI semantic colors:

```html
<div class="p-12 text-center">
  <svg class="mx-auto h-12 w-12 text-base-content/30">...</svg>
  <h3 class="mt-2 text-lg font-semibold">No servers configured</h3>
  <p class="text-base-content/60">
    <a href="/servers" class="link link-primary">Add a server</a> to get
    started.
  </p>
</div>
```

---

### 9. Filter Toolbar

**Location:** `logs.templ:38-130`

**Current:** Custom grid with 5 columns of dropdowns and date inputs with verbose Tailwind classes.

**Issue:** DaisyUI doesn't have a specific filter bar component.

**Resolution:** Keep the grid layout but convert form elements to DaisyUI:

```html
<div class="card bg-base-100 shadow mb-4">
  <div class="card-body p-4">
    <div class="grid grid-cols-1 md:grid-cols-5 gap-4">
      <select class="select select-bordered select-sm">
        <option>All Types</option>
        ...
      </select>
      <select class="select select-bordered select-sm">
        <option>All Servers</option>
        ...
      </select>
      <input type="date" class="input input-bordered input-sm" />
      <input type="date" class="input input-bordered input-sm" />
      <div class="flex gap-2">
        <button class="btn btn-primary btn-sm">Filter</button>
        <button class="btn btn-ghost btn-sm">Clear</button>
      </div>
    </div>
  </div>
</div>
```

---

### 10. Inline Dynamic Status Messages

**Locations:** `server_card.templ:61-66`, `server_form.templ:117`, `config_form.templ:160-162`

**Current:** Alpine.js-driven text with conditional Tailwind colors:

```html
<div
  x-show="testResult"
  :class="testResult.startsWith('Connected') ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'"
  x-text="testResult"
></div>
```

**Resolution:** Use DaisyUI semantic colors:

```html
<div
  x-show="testResult"
  :class="testResult.startsWith('Connected') ? 'text-success' : 'text-error'"
  x-text="testResult"
></div>
```

---

### 11. Conditional Row Highlighting

**Locations:** `log_entry.templ:6`, `dashboard.templ:127-129`

**Current:** Explicit background colors for error rows:

```html
<div class={ "p-4", templ.KV("bg-red-50 dark:bg-red-900/20", entry.Type == logger.LogTypeError) }>
```

**Resolution:** Use DaisyUI semantic colors with opacity:

```html
<div class={ "p-4 rounded-lg", templ.KV("bg-error/10", entry.Type == logger.LogTypeError) }>
```

---

### 12. Link Styling

**Locations:** `dashboard.templ:76`, `dashboard.templ:156`

**Current:** Manual link styling:

```html
<a href="/servers" class="text-blue-600 hover:underline">Add a server</a>
```

**Resolution:** Use DaisyUI link component:

```html
<a href="/servers" class="link link-primary">Add a server</a>
```

---

## Open Questions

None - all requirements have been clarified.
