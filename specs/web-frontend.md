# Web Frontend Specification

## Overview

Janitarr provides a modern, responsive web interface for managing settings, configuring servers, and monitoring logs. The UI uses server-rendered HTML with templ templates, enhanced with htmx for dynamic updates and Alpine.js for client-side interactivity.

## Goals

1. **Accessibility**: Provide a user-friendly web interface for all janitarr features
2. **Real-time Monitoring**: Enable live log streaming and server status updates
3. **Mobile Support**: Ensure full functionality on mobile and tablet devices
4. **Zero Configuration**: Work out-of-the-box with sensible defaults
5. **Performance**: Fast, responsive UI with minimal resource overhead
6. **Simplicity**: No JavaScript build step, single binary deployment

## Architecture

### Technology Stack

- **Templates**: templ (a-h/templ) - Type-safe Go HTML templates
- **Dynamic Updates**: htmx - HTML over the wire
- **Interactivity**: Alpine.js - Lightweight reactive framework
- **Styling**: Tailwind CSS - Utility-first CSS
- **HTTP Server**: Chi router (go-chi/chi/v5)
- **WebSocket**: gorilla/websocket for real-time log streaming
- **Port**: Configurable (default: 3434)

### Why This Stack?

| Choice       | Rationale                                                     |
| ------------ | ------------------------------------------------------------- |
| templ        | Type-safe templates that compile to Go, excellent IDE support |
| htmx         | Progressive enhancement, no client-side state management      |
| Alpine.js    | Minimal JS for interactions (modals, toggles, dark mode)      |
| Tailwind CSS | Utility-first, works great with templ, easy dark mode         |
| No React     | Simpler deployment, smaller bundle, faster initial load       |

### Integration Model

```
┌─────────────────────────────────────────────────┐
│              Janitarr Process                    │
│                                                  │
│  ┌──────────────┐         ┌──────────────────┐ │
│  │   CLI Layer  │         │   Web Server     │ │
│  │   (Cobra)    │         │   (Chi + templ)  │ │
│  └──────────────┘         └──────────────────┘ │
│         │                          │            │
│         │                          │            │
│         ├──────────┬───────────────┤            │
│                    │                            │
│           ┌────────▼────────┐                   │
│           │  Service Layer  │                   │
│           │  - ServerManager │                  │
│           │  - Detector      │                  │
│           │  - Automation    │                  │
│           │  - Logger        │                  │
│           └─────────────────┘                   │
│                    │                            │
│           ┌────────▼────────┐                   │
│           │  SQLite Database │                  │
│           └─────────────────┘                   │
└─────────────────────────────────────────────────┘
```

## User Interface Design

### Layout Structure

Material Design 3 navigation with responsive drawer:

```
┌─────────────────────────────────────────────────┐
│  App Bar (Top)                           [☰] [◐]│
├─────────────────────────────────────────────────┤
│     │                                            │
│  N  │         Main Content Area                 │
│  a  │                                            │
│  v  │         (Dashboard / Servers / Settings   │
│     │          / Logs view)                      │
│  D  │                                            │
│  r  │                                            │
│  a  │                                            │
│  w  │                                            │
│  e  │                                            │
│  r  │                                            │
│     │                                            │
└─────┴────────────────────────────────────────────┘
```

**Mobile/Tablet**: Navigation drawer collapses to hamburger menu, content fills screen

### Navigation Structure

1. **Dashboard** (default view)
   - Icon: `HomeIcon`
   - Route: `/`

2. **Servers**
   - Icon: `StorageIcon`
   - Route: `/servers`

3. **Logs**
   - Icon: `HistoryIcon`
   - Route: `/logs`

4. **Settings**
   - Icon: `SettingsIcon`
   - Route: `/settings`

### Color Scheme (Material Design 3)

**Light Theme:**

- Primary: Purple/Blue (`#6750A4`)
- Secondary: Pink (`#E91E63`)
- Surface: White/Off-white
- Background: Light grey (`#FAFAFA`)

**Dark Theme:**

- Primary: Light purple (`#D0BCFF`)
- Secondary: Light pink
- Surface: Dark grey (`#1C1B1F`)
- Background: Darker grey (`#121212`)

## Feature Specifications

### 1. Dashboard View

**Purpose**: Overview of janitarr status and recent activity

**Components:**

1. **Status Cards** (4-card grid, responsive to 2x2 or 1-column on mobile)
   - **Total Servers**: Count of configured servers with active/inactive breakdown
   - **Last Automation Cycle**: Time since last cycle, next scheduled time
   - **Recent Searches**: Count of searches in last 24 hours
   - **Error Count**: Errors in last 24 hours with severity indicator

2. **Server Status List**
   - Table/List of all servers with:
     - Server name
     - Type (Radarr/Sonarr) with icon
     - Status indicator: ✓ Connected (green) / ✗ Error (red) / ◌ Disabled (grey)
     - Last check timestamp
     - Quick actions: Test, Edit, Disable/Enable
   - Click row to navigate to server detail

3. **Recent Activity Timeline**
   - Last 10 log entries in timeline format
   - Color-coded by type (cycle, search, error)
   - Expandable entries for full details
   - "View All Logs" button to navigate to Logs view

4. **Quick Actions** (Floating Action Button or prominent buttons)
   - "Run Automation Now" - Trigger manual cycle
   - "Add Server" - Open server creation dialog

**Real-time Updates:**

- WebSocket connection updates status cards when logs arrive
- Auto-refresh server status every 60 seconds (configurable)

---

### 2. Servers View

**Purpose**: Manage Radarr/Sonarr server configurations

**Layout:**

1. **Server List/Grid** (switchable view)
   - **List View**: Detailed table with all columns
   - **Card View**: Material cards with key info (better for mobile)

2. **Server Card/Row Information**:
   - Name (editable in-place or via dialog)
   - Type (Radarr/Sonarr) with logo
   - URL
   - Status: Connected / Disconnected / Error / Disabled
   - Statistics (if enabled):
     - Total searches triggered
     - Last successful check
     - Success rate percentage
   - Actions:
     - **Test Connection**: Validate API key and connectivity
     - **Edit**: Open edit dialog
     - **Enable/Disable**: Toggle server active state
     - **Delete**: Remove server (with confirmation)

3. **Add Server Button** (Primary FAB or prominent button)
   - Opens dialog/modal with form:
     - Name (text input, required)
     - Type (radio buttons: Radarr / Sonarr, required)
     - URL (text input with validation, required)
     - API Key (password input, required)
     - Enabled (checkbox, default: true)
   - "Test Connection" button in dialog before saving
   - Validation feedback inline
   - Save creates server via POST /api/servers

4. **Edit Server Dialog**
   - Same form as Add, pre-populated
   - Update via PUT /api/servers/:id

5. **Server Statistics Panel** (optional, expandable section)
   - Chart showing searches over time
   - Breakdown by missing vs. cutoff searches
   - Error history

**Interactions:**

- Drag-and-drop to reorder servers (future enhancement)
- Bulk actions: Enable/Disable multiple, Delete multiple (with multi-select)
- Search/filter servers by name or type
- Sort by name, type, status, or last activity

---

### 3. Logs View

**Purpose**: Monitor and search janitarr activity logs

**Layout:**

1. **Toolbar** (sticky at top)
   - **Search Input**: Full-text search across log messages
   - **Type Filter**: Multi-select dropdown (Cycle Start, Cycle End, Search, Error, All)
   - **Server Filter**: Dropdown to filter by specific server or "All Servers"
   - **Date Range Picker**: Filter logs by date/time range
   - **Export Button**: Download filtered logs as JSON or CSV
   - **Clear Logs Button**: Clear all logs with confirmation dialog
   - **Refresh Button**: Manually refresh if WebSocket disconnected

2. **Log Stream** (auto-scrolling list with virtualization for performance)
   - Each log entry displayed as a Material Card or List Item:
     ```
     ┌─────────────────────────────────────────────┐
     │ [Icon] [Type Badge]  Server Name            │
     │        Timestamp (relative: "2 mins ago")   │
     │        Message text                         │
     │        [Details Chip: "5 items searched"]   │
     └─────────────────────────────────────────────┘
     ```
   - **Icon & Color Coding**:
     - Cycle Start: `PlayArrowIcon` (Blue)
     - Cycle End: `CheckCircleIcon` (Green)
     - Search: `SearchIcon` (Cyan)
     - Error: `ErrorIcon` (Red)
   - **Badges**: "Manual" badge for manual triggers
   - **Expandable Details**: Click to expand full JSON or additional metadata
   - **Action Menu**: Export single entry, Copy to clipboard

3. **Real-time Streaming**
   - WebSocket connection indicator in toolbar (Connected / Reconnecting / Disconnected)
   - New logs auto-append to bottom with smooth animation
   - Auto-scroll to bottom when new log arrives (unless user scrolled up)
   - "New logs available" snackbar if user scrolled away
   - Smooth fade-in animation for new entries

4. **Pagination**
   - Server-side pagination with htmx for efficient rendering
   - Load initial batch (last 100 logs), load more on scroll or button click
   - Maintain performance with thousands of entries

5. **Export Functionality**
   - **JSON Export**: All filtered logs as JSON array
   - **CSV Export**: Flattened format with columns: Timestamp, Type, Server, Category, Count, Message
   - Download triggers browser download dialog

---

### 4. Settings View

**Purpose**: Configure janitarr application settings

**Layout: Grouped Settings Sections**

#### Section 1: Automation Schedule

Material Card with form inputs:

1. **Enable Automation**
   - Switch toggle (on/off)
   - Description: "Automatically run detection and search cycles on schedule"

2. **Interval Hours**
   - Number input with stepper (+/- buttons)
   - Min: 1, Max: 168 (7 days)
   - Suffix: "hours"
   - Description: "Time between automatic cycles"

3. **Next Scheduled Run**
   - Read-only display showing next execution time
   - Updates in real-time

#### Section 2: Search Limits

Material Card with form inputs:

1. **Missing Content Limit**
   - Number input with stepper
   - Min: 0, Max: 1000
   - Suffix: "items per cycle"
   - Description: "Maximum missing items to search per automation cycle"

2. **Quality Cutoff Limit**
   - Number input with stepper
   - Min: 0, Max: 1000
   - Suffix: "items per cycle"
   - Description: "Maximum quality upgrade items to search per cycle"

#### Section 3: Web Interface Settings

Material Card with form inputs:

1. **Server Port**
   - Number input
   - Min: 1024, Max: 65535
   - Description: "Port for web interface (requires restart)"

2. **Theme Preference**
   - Radio buttons or Segmented button group:
     - Light
     - Dark
     - System (auto-detect)
   - Persisted in localStorage, not in database

3. **Log Retention Days**
   - Number input
   - Min: 1, Max: 365
   - Description: "Days to keep logs before auto-deletion"

#### Section 4: Advanced

Material Expansion Panel (collapsed by default):

1. **Database Path**
   - Read-only text input showing current path
   - "Open in Explorer" button (if applicable)

2. **API Base URL**
   - Read-only display for API endpoint (for external integrations)
   - Copy button

3. **Reset to Defaults**
   - Button to restore default settings (with confirmation)

**Interactions:**

- Auto-save on blur (save individual field changes via PATCH /api/config)
- Or "Save Changes" button at bottom (save all changes at once)
- Success snackbar on save: "Settings saved successfully"
- Error snackbar on failure with retry button
- Validation feedback inline (red text for errors)
- Reset confirmation dialog with warning about data loss

---

## API Specifications

### REST API Endpoints

Base URL: `http://localhost:3000/api`

#### Configuration

```
GET    /api/config
  Response: { schedule: { intervalHours, enabled }, searchLimits: { missingLimit, cutoffLimit } }

PATCH  /api/config
  Body: Partial<AppConfig>
  Response: Updated AppConfig

PUT    /api/config/reset
  Response: Default AppConfig
```

#### Servers

```
GET    /api/servers
  Response: Server[]
  Query params: ?type=radarr|sonarr (optional filter)

GET    /api/servers/:id
  Response: Server

POST   /api/servers
  Body: { name, type, url, apiKey, enabled? }
  Response: Created Server (with generated id)

PUT    /api/servers/:id
  Body: { name?, url?, apiKey?, enabled? }
  Response: Updated Server

DELETE /api/servers/:id
  Response: 204 No Content

POST   /api/servers/:id/test
  Response: { success: boolean, message: string, status?: SystemStatus }
  Tests connection and returns server status if successful
```

#### Logs

```
GET    /api/logs
  Response: Log[]
  Query params:
    - limit: number (default: 100, max: 1000)
    - offset: number (default: 0)
    - type: string (filter by log type)
    - server: string (filter by server name)
    - startDate: ISO timestamp
    - endDate: ISO timestamp
    - search: string (full-text search in message)

DELETE /api/logs
  Response: { deletedCount: number }
  Clears all logs

GET    /api/logs/export
  Response: Log[] as JSON or CSV (based on Accept header or ?format=csv)
  Same query params as GET /api/logs
```

#### Automation

```
POST   /api/automation/trigger
  Body: { type?: "full" | "missing" | "cutoff" } (optional, default: "full")
  Response: { jobId: string, message: string }
  Triggers manual automation cycle, returns immediately

GET    /api/automation/status
  Response: {
    running: boolean,
    nextScheduledRun: ISO timestamp | null,
    lastRunTime: ISO timestamp | null,
    lastRunResults?: { searchesTriggered, failedServers }
  }
```

#### Statistics (for dashboard)

```
GET    /api/stats/summary
  Response: {
    totalServers: number,
    activeServers: number,
    lastCycleTime: ISO timestamp | null,
    nextScheduledTime: ISO timestamp | null,
    searchesLast24h: number,
    errorsLast24h: number
  }

GET    /api/stats/servers/:id
  Response: {
    totalSearches: number,
    successRate: number,
    lastCheckTime: ISO timestamp,
    errorCount: number
  }
```

### WebSocket API

**Endpoint**: `ws://localhost:3000/ws/logs`

**Protocol**: JSON messages

**Client → Server Messages**:

```json
{
  "type": "subscribe",
  "filters": {
    "types": ["search", "error"], // optional
    "servers": ["server-uuid"] // optional
  }
}
```

```json
{
  "type": "unsubscribe"
}
```

```json
{
  "type": "ping"
}
```

**Server → Client Messages**:

```json
{
  "type": "log",
  "data": {
    "id": "uuid",
    "timestamp": "2026-01-15T10:30:00Z",
    "type": "search",
    "serverName": "My Radarr",
    "serverType": "radarr",
    "category": "missing",
    "count": 5,
    "message": "Triggered search for 5 missing movies",
    "isManual": false
  }
}
```

```json
{
  "type": "connected",
  "message": "WebSocket connection established"
}
```

```json
{
  "type": "pong"
}
```

**Connection Management**:

- Auto-reconnect on disconnect (exponential backoff: 1s, 2s, 4s, 8s, max 30s)
- Ping/pong every 30s to keep connection alive
- Close connection on page unload

---

## Implementation Plan

### Phase 1: templ Setup and Base Layout

**Tasks**:

1. Install templ CLI:

   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   ```

2. Create template directory structure:

   ```
   src/templates/
   ├── layouts/
   │   └── base.templ        # HTML5 document, nav, content slot
   ├── components/
   │   ├── nav.templ         # Navigation sidebar
   │   ├── server_card.templ # Server display card
   │   ├── log_entry.templ   # Single log entry
   │   ├── stats_card.templ  # Dashboard stat card
   │   └── forms/
   │       ├── server_form.templ
   │       └── config_form.templ
   └── pages/
       ├── dashboard.templ
       ├── servers.templ
       ├── logs.templ
       └── settings.templ
   ```

3. Create base layout with dark mode support:

   ```go
   // templates/layouts/base.templ
   package layouts

   templ Base(title string) {
       <!DOCTYPE html>
       <html lang="en" x-data="{ darkMode: localStorage.getItem('darkMode') === 'true' }" :class="{ 'dark': darkMode }">
       <head>
           <meta charset="UTF-8"/>
           <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
           <title>{ title } - Janitarr</title>
           <link href="/static/css/app.css" rel="stylesheet"/>
           <script src="/static/js/htmx.min.js"></script>
           <script src="/static/js/alpine.min.js" defer></script>
       </head>
       <body class="bg-gray-100 dark:bg-gray-900">
           { children... }
       </body>
       </html>
   }
   ```

4. Set up Tailwind CSS:

   ```bash
   npm init -y
   npm install -D tailwindcss
   npx tailwindcss init
   ```

   Configure `tailwind.config.js` to scan templ files:

   ```javascript
   module.exports = {
     content: ["./src/templates/**/*.templ"],
     darkMode: "class",
     theme: { extend: {} },
     plugins: [],
   };
   ```

5. Download htmx and Alpine.js:
   ```bash
   mkdir -p static/js
   curl -o static/js/htmx.min.js https://unpkg.com/htmx.org@1.9/dist/htmx.min.js
   curl -o static/js/alpine.min.js https://unpkg.com/alpinejs@3/dist/cdn.min.js
   ```

**Acceptance Criteria**:

- `templ generate` compiles templates without errors
- Base layout renders with navigation
- Dark mode toggle works with Alpine.js
- Tailwind CSS styles applied correctly

---

### Phase 2: Page Handlers and Routing

**Tasks**:

1. Create page handlers in `src/web/handlers/pages/`:

   ```go
   // dashboard.go
   func Dashboard(db *database.DB, scheduler *services.Scheduler) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           stats := db.GetStats()
           status := scheduler.GetStatus()
           pages.Dashboard(stats, status).Render(r.Context(), w)
       }
   }
   ```

2. Register routes in Chi router:

   ```go
   r.Get("/", handlers.Dashboard(db, scheduler))
   r.Get("/servers", handlers.ServersPage(db))
   r.Get("/logs", handlers.LogsPage(db))
   r.Get("/settings", handlers.SettingsPage(db))
   ```

3. Implement each page template:
   - Dashboard: Stats cards, server list, recent activity
   - Servers: Server grid, add/edit forms
   - Logs: Log list, filters, export
   - Settings: Configuration forms

**Acceptance Criteria**:

- All pages render correctly
- Navigation between pages works
- Data displays correctly from database

---

### Phase 3: htmx Dynamic Updates

**Tasks**:

1. Add htmx attributes for dynamic updates:

   ```go
   // Server card with test button
   templ ServerCard(server database.Server) {
       <div class="bg-white dark:bg-gray-800 rounded-lg p-4">
           <h3>{ server.Name }</h3>
           <button
               hx-post={ "/api/servers/" + server.ID + "/test" }
               hx-target="#test-result"
               hx-indicator="#spinner"
               class="btn"
           >
               Test Connection
           </button>
           <span id="test-result"></span>
       </div>
   }
   ```

2. Create partial templates for htmx responses:

   ```go
   // Partial for stats refresh
   templ StatsPartial(stats Stats) {
       <div class="grid grid-cols-4 gap-4">
           @StatsCard("Servers", stats.ServerCount)
           @StatsCard("Last Cycle", stats.LastCycle)
           @StatsCard("Searches", stats.SearchCount)
           @StatsCard("Errors", stats.ErrorCount)
       </div>
   }
   ```

3. Add polling for real-time updates:

   ```html
   <div hx-get="/partials/stats" hx-trigger="every 60s" hx-swap="innerHTML">
     <!-- Stats content -->
   </div>
   ```

4. Implement form submissions:
   - Add server form
   - Edit server form
   - Settings form
   - All use htmx for seamless updates

**Acceptance Criteria**:

- Forms submit without page reload
- Test connection shows result inline
- Stats auto-refresh every 60 seconds
- Delete confirmation works with Alpine.js

---

### Phase 4: WebSocket Log Streaming

**Tasks**:

1. Implement WebSocket handler:

   ```go
   // src/web/websocket/logs.go
   func LogsHandler(logger *logger.Logger) http.HandlerFunc {
       upgrader := websocket.Upgrader{
           CheckOrigin: func(r *http.Request) bool { return true },
       }

       return func(w http.ResponseWriter, r *http.Request) {
           conn, err := upgrader.Upgrade(w, r, nil)
           if err != nil {
               return
           }
           defer conn.Close()

           ch := logger.Subscribe()
           defer logger.Unsubscribe(ch)

           for entry := range ch {
               if err := conn.WriteJSON(entry); err != nil {
                   break
               }
           }
       }
   }
   ```

2. Add JavaScript for WebSocket connection:

   ```javascript
   // In logs.templ
   <script>
   function logViewer() {
       return {
           ws: null,
           connect() {
               this.ws = new WebSocket(`ws://${window.location.host}/ws/logs`);
               this.ws.onmessage = (e) => {
                   const log = JSON.parse(e.data);
                   htmx.ajax('GET', '/partials/log-entry?id=' + log.id, {
                       target: '#log-list',
                       swap: 'afterbegin'
                   });
               };
               this.ws.onclose = () => setTimeout(() => this.connect(), 1000);
           }
       }
   }
   </script>
   ```

3. Create log entry partial for htmx insertion

**Acceptance Criteria**:

- Logs appear in real-time
- WebSocket reconnects on disconnect
- New logs insert at top of list
- Filter updates work correctly

---

### Phase 5: Testing & Polish

**Tasks**:

1. **Unit Tests**:
   - Handler tests with httptest
   - Template rendering tests
   - WebSocket connection tests

2. **E2E Tests with Playwright**:
   - Dashboard loads correctly
   - Add/edit/delete server flow
   - Settings save and persist
   - Log filtering works
   - Dark mode toggle

3. **Mobile Responsiveness**:
   - Test on small screens
   - Collapsible navigation
   - Touch-friendly buttons

4. **Accessibility**:
   - ARIA labels
   - Keyboard navigation
   - Color contrast
   - Focus management

**Acceptance Criteria**:

- All tests pass
- Responsive on mobile
- Accessible with screen readers
- No console errors

---

## Configuration

### Environment Variables

```bash
# Web server configuration
JANITARR_WEB_PORT=3000              # Port for web interface (default: 3000)
JANITARR_WEB_HOST=localhost         # Host to bind to (default: localhost)
JANITARR_WEB_ENABLED=true           # Enable web server (default: true)

# Existing variables
JANITARR_DB_PATH=./data/janitarr.db # Database path
LOG_RETENTION_DAYS=30               # Log retention period
```

### Config File (Optional Future Enhancement)

```json
{
  "web": {
    "port": 3000,
    "host": "localhost",
    "enabled": true,
    "openBrowser": false,
    "corsOrigins": ["http://localhost:5173"]
  }
}
```

---

## Security Considerations

### Current Scope (No Authentication)

- **Localhost Only**: Default host binding to `localhost` prevents external access
- **Internal Network**: Suitable for home networks with trusted users
- **No API Keys**: No authentication required for initial version
- **Warning**: Display warning in UI if accessed from non-localhost origin
- **HTTP Only**: Web interface uses HTTP protocol only

### Future Authentication (Out of Scope for v1)

- Basic username/password with bcrypt hashing
- Session management with secure cookies
- CSRF protection
- Rate limiting on API endpoints
- API key generation for programmatic access

---

## Success Metrics

1. **Usability**: Users can complete common tasks (add server, view logs, change settings) in <30 seconds
2. **Performance**: Page load time <2 seconds, log streaming latency <100ms
3. **Reliability**: WebSocket reconnection success rate >99%
4. **Mobile**: Fully functional on devices with screen width ≥320px
5. **Accessibility**: WCAG 2.1 Level AA compliance

---

## Future Enhancements (Post-v1)

1. **Multi-user Support**: User accounts with role-based permissions
2. **Notifications**: Browser push notifications for critical events
3. **Advanced Charts**: Historical data visualization with Chart.js or Recharts
4. **Server Groups**: Organize servers into logical groups
5. **Bulk Operations**: Multi-select and bulk actions on servers/logs
6. **API Webhooks**: Outbound webhooks for integration with other tools
7. **Mobile Apps**: Progressive Web App (PWA) support
8. **Plugins**: Extensibility system for custom integrations
9. **Internationalization**: Multi-language support (i18n)
10. **Configuration Backup**: Scheduled backups and version history

---

## Dependencies

### Go Dependencies

```go
// go.mod
require (
    github.com/go-chi/chi/v5 v5.0.12
    github.com/gorilla/websocket v1.5.1
    modernc.org/sqlite v1.29.1
    github.com/spf13/cobra v1.8.0
    github.com/a-h/templ v0.2.543
)
```

### Frontend Dependencies (npm - for Tailwind build only)

```json
{
  "devDependencies": {
    "tailwindcss": "^3.4.0"
  }
}
```

### Static Assets (CDN downloads)

- htmx v1.9+ (30KB minified)
- Alpine.js v3+ (15KB minified)

No JavaScript build step required. Static files served directly from `static/` directory.

---

## Technical Risks & Mitigations

| Risk                                          | Impact                             | Mitigation                                               |
| --------------------------------------------- | ---------------------------------- | -------------------------------------------------------- |
| WebSocket connection instability              | Users miss real-time logs          | Implement robust reconnection logic, fallback to polling |
| Large log datasets causing performance issues | UI becomes sluggish                | Use virtualization, pagination, and efficient filtering  |
| Browser compatibility issues                  | Users can't access UI              | Test on all major browsers, use polyfills if needed      |
| Build size too large                          | Slow initial load                  | Code splitting, tree shaking, lazy loading routes        |
| API breaking changes affecting UI             | Frontend breaks on backend updates | Version API, maintain backwards compatibility            |
| Memory leaks from WebSocket subscriptions     | Browser tab becomes unresponsive   | Proper cleanup in useEffect hooks, connection monitoring |

---

## Open Questions

1. Should we support multiple instances of janitarr with a shared database?
2. Do we need a configuration wizard on first launch?
3. Should we include a "Get Started" tutorial/onboarding flow?
4. Do we want to support custom themes beyond light/dark?
5. Should logs be stored in a separate table optimized for querying?

---

## Conclusion

This specification provides a comprehensive roadmap for implementing a modern, Material Design 3 web interface for janitarr. The embedded server architecture ensures zero-configuration deployment while maintaining the simplicity of the existing CLI tool. By following the phased implementation plan, we can deliver a polished, fully-featured web UI that enhances the janitarr user experience while maintaining the performance and reliability of the existing system.
