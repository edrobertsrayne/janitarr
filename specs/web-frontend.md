# Web Frontend Specification

## Overview

Add a modern, responsive web interface to janitarr that enables users to manage settings, configure servers, and monitor logs through a browser. The web UI will follow Material Design 3 Expressive principles and provide an intuitive alternative to the CLI interface.

## Goals

1. **Accessibility**: Provide a user-friendly web interface for all janitarr features
2. **Real-time Monitoring**: Enable live log streaming and server status updates
3. **Mobile Support**: Ensure full functionality on mobile and tablet devices
4. **Zero Configuration**: Work out-of-the-box with sensible defaults
5. **Performance**: Fast, responsive UI with minimal resource overhead

## Architecture

### Frontend Stack

- **Framework**: React 18+ with TypeScript
- **Build Tool**: Vite 5+ (fast HMR, optimized builds)
- **UI Library**: MUI (Material-UI v6) - Material Design 3 implementation
- **Routing**: React Router v6
- **State Management**: React Context API + hooks (lightweight, no external deps)
- **HTTP Client**: Native Fetch API with type-safe wrappers
- **WebSocket**: Native WebSocket API for real-time log streaming
- **Theme**: Material Design 3 Expressive with dark/light mode support

### Backend Stack

- **HTTP Server**: Bun's native HTTP server (`Bun.serve()`)
- **API Style**: RESTful JSON API
- **WebSocket**: Native WebSocket support via Bun.serve upgrade
- **Integration**: Embedded server running as part of janitarr process
- **Port**: Configurable (default: 3000)
- **Static Files**: Serve built frontend from `dist/` or `public/` directory

### Integration Model

```
┌─────────────────────────────────────────────────┐
│              Janitarr Process                    │
│                                                  │
│  ┌──────────────┐         ┌──────────────────┐ │
│  │   CLI Layer  │         │   Web Server     │ │
│  │  (existing)  │         │   (new)          │ │
│  └──────────────┘         └──────────────────┘ │
│         │                          │            │
│         │                          │            │
│         ├──────────┬───────────────┤            │
│                    │                            │
│           ┌────────▼────────┐                   │
│           │  Service Layer  │                   │
│           │  - ServerManager │                  │
│           │  - DatabaseMgr   │                  │
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

4. **Virtualization**
   - Use `react-window` or similar for efficient rendering of large log lists
   - Load initial batch (last 100 logs), lazy-load more on scroll up
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
    "types": ["search", "error"],  // optional
    "servers": ["server-uuid"],     // optional
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

### Phase 1: Backend API Foundation

**Tasks**:
1. Create `src/web/` directory structure:
   ```
   src/web/
   ├── server.ts          # Main HTTP server setup
   ├── routes/
   │   ├── config.ts      # Config endpoints
   │   ├── servers.ts     # Server management endpoints
   │   ├── logs.ts        # Log endpoints
   │   ├── automation.ts  # Automation control endpoints
   │   └── stats.ts       # Statistics endpoints
   ├── middleware/
   │   ├── error-handler.ts
   │   ├── logger.ts
   │   └── cors.ts
   ├── websocket.ts       # WebSocket log streaming
   └── types.ts           # API request/response types
   ```

2. Implement REST API endpoints using Bun.serve:
   ```typescript
   export function createWebServer(db: DatabaseManager, automation: AutomationService) {
     return Bun.serve({
       port: 3000,
       fetch(req, server) {
         // Route handling logic
       },
       websocket: {
         // WebSocket handlers for log streaming
       },
     });
   }
   ```

3. Add WebSocket log streaming:
   - Create broadcast mechanism for new logs
   - Implement subscription filtering
   - Handle connection lifecycle

4. Add CLI command:
   ```bash
   janitarr serve [options]
     --port, -p <number>   Port to listen on (default: 3000)
     --host <string>       Host to bind to (default: localhost)
     --open, -o            Open browser automatically
   ```

5. Update `DatabaseManager` with query methods needed for API:
   - `getServerStats(serverId)`
   - `getLogsPaginated(filters, limit, offset)`
   - `searchLogs(query)`
   - `getSystemStats()`

**Acceptance Criteria**:
- All REST endpoints functional and returning correct data
- WebSocket streaming logs in real-time
- API integration tests passing
- OpenAPI/Swagger documentation generated

---

### Phase 2: Frontend Project Setup

**Tasks**:
1. Initialize React + Vite project in `ui/` directory:
   ```bash
   cd ui/
   bun create vite . --template react-ts
   bun add @mui/material @emotion/react @emotion/styled
   bun add react-router-dom
   bun add @mui/icons-material
   ```

2. Configure Vite for development:
   ```typescript
   // vite.config.ts
   export default {
     server: {
       proxy: {
         '/api': 'http://localhost:3000',
         '/ws': { target: 'ws://localhost:3000', ws: true }
       }
     },
     build: {
       outDir: '../dist/public',
       emptyOutDir: true
     }
   }
   ```

3. Set up project structure:
   ```
   ui/
   ├── src/
   │   ├── App.tsx
   │   ├── main.tsx
   │   ├── components/
   │   │   ├── layout/
   │   │   │   ├── AppBar.tsx
   │   │   │   ├── NavDrawer.tsx
   │   │   │   └── Layout.tsx
   │   │   ├── servers/
   │   │   ├── logs/
   │   │   ├── settings/
   │   │   ├── dashboard/
   │   │   └── common/      # Shared components
   │   ├── hooks/
   │   │   ├── useApi.ts
   │   │   ├── useWebSocket.ts
   │   │   └── useTheme.ts
   │   ├── services/
   │   │   ├── api.ts       # REST API client
   │   │   └── websocket.ts # WebSocket client
   │   ├── types/
   │   │   └── index.ts     # Shared types with backend
   │   ├── contexts/
   │   │   ├── ThemeContext.tsx
   │   │   └── ConfigContext.tsx
   │   └── theme.ts         # MUI theme configuration
   ├── index.html
   ├── package.json
   └── vite.config.ts
   ```

4. Configure MUI theme with Material Design 3:
   ```typescript
   import { createTheme } from '@mui/material/styles';

   export const lightTheme = createTheme({
     palette: {
       mode: 'light',
       primary: { main: '#6750A4' },
       secondary: { main: '#E91E63' },
       // ... Material Design 3 colors
     },
     shape: { borderRadius: 12 }, // MD3 rounded corners
     // ... typography, components customization
   });
   ```

5. Implement routing structure:
   ```typescript
   <BrowserRouter>
     <Routes>
       <Route path="/" element={<Layout />}>
         <Route index element={<Dashboard />} />
         <Route path="servers" element={<Servers />} />
         <Route path="logs" element={<Logs />} />
         <Route path="settings" element={<Settings />} />
       </Route>
     </Routes>
   </BrowserRouter>
   ```

**Acceptance Criteria**:
- Vite dev server running with HMR
- MUI components rendering correctly
- Routing functional
- Theme switching working (light/dark)
- TypeScript compilation with no errors

---

### Phase 3: Core Components Implementation

**Tasks**:

1. **Layout Components**:
   - `AppBar`: Top bar with title, theme toggle, status indicators
   - `NavDrawer`: Responsive navigation drawer with menu items
   - `Layout`: Main layout wrapper with drawer + content area

2. **Dashboard View**:
   - Status cards with live data from `/api/stats/summary`
   - Server status list with real-time updates
   - Recent activity timeline (last 10 logs)
   - Quick action buttons (FAB or prominent)

3. **Servers View**:
   - Server list/card grid with toggle view
   - Add server dialog with form validation
   - Edit server dialog
   - Delete confirmation dialog
   - Test connection button with loading state and feedback
   - Enable/disable toggle with instant update

4. **Logs View**:
   - Log list with virtualization (use `react-window`)
   - Search/filter toolbar
   - WebSocket integration for real-time logs
   - Export functionality (JSON/CSV download)
   - Auto-scroll behavior

5. **Settings View**:
   - Form sections for config groups
   - Auto-save or save button
   - Validation and error handling
   - Reset to defaults functionality

**Shared/Common Components**:
- `LoadingSpinner`: Consistent loading indicator
- `ErrorBoundary`: Catch and display React errors
- `ConfirmDialog`: Reusable confirmation dialog
- `Snackbar`: Toast notifications for success/error
- `StatusBadge`: Color-coded status indicator
- `ServerIcon`: Radarr/Sonarr logo icons

**Acceptance Criteria**:
- All views render and match design spec
- Navigation between views working
- Forms submit correctly to API
- Error states handled gracefully
- Loading states displayed during async operations

---

### Phase 4: Real-time Features & Polish

**Tasks**:

1. **WebSocket Integration**:
   - Create `useWebSocket` hook for log streaming
   - Auto-reconnect logic with exponential backoff
   - Connection status indicator in UI
   - Subscribe/unsubscribe based on active view
   - Filter logs based on user-selected filters

2. **API Integration**:
   - Create `useApi` hook for consistent API calls
   - Error handling and retry logic
   - Loading states
   - Optimistic updates for instant feedback

3. **Real-time Updates**:
   - Dashboard stats refresh on new logs
   - Server status updates
   - Notification snackbars for important events
   - "New logs available" indicator when scrolled away

4. **Performance Optimizations**:
   - Lazy load routes with React.lazy
   - Virtualize long lists (logs, servers if many)
   - Debounce search inputs
   - Memoize expensive computations
   - Code splitting for smaller bundles

5. **Mobile Responsiveness**:
   - Test on mobile viewports
   - Adjust layouts for small screens
   - Touch-friendly buttons and interactions
   - Responsive tables (card view on mobile)

6. **Accessibility**:
   - ARIA labels for icons and actions
   - Keyboard navigation support
   - Focus management in dialogs
   - Color contrast validation
   - Screen reader testing

**Acceptance Criteria**:
- Logs stream in real-time without lag
- UI responsive on mobile devices
- No console errors or warnings
- Lighthouse score: 90+ (Performance, Accessibility)
- Works in Chrome, Firefox, Safari, Edge

---

### Phase 5: Testing & Documentation

**Tasks**:

1. **Backend Tests**:
   - Unit tests for API routes
   - Integration tests for WebSocket
   - Test error handling
   - Test concurrent requests

2. **Frontend Tests**:
   - Component tests with React Testing Library
   - Integration tests for user flows
   - Test WebSocket reconnection
   - Test form validation

3. **End-to-End Tests**:
   - Use Playwright or Cypress
   - Test critical user journeys:
     - Add and configure server
     - View logs and filter
     - Change settings and save
     - Trigger manual automation

4. **Documentation**:
   - Update main README with web UI instructions
   - Create `docs/web-ui.md` with:
     - Screenshots of each view
     - Feature overview
     - Configuration options
     - Troubleshooting guide
   - API documentation (OpenAPI/Swagger)
   - Developer guide for contributing to UI

5. **Deployment**:
   - Build script: `bun run build` in ui/ directory
   - Copy built files to `dist/public/`
   - Update `src/web/server.ts` to serve static files
   - Add production build to CI/CD pipeline

**Acceptance Criteria**:
- Test coverage >80% for critical paths
- All E2E tests passing
- Documentation complete and accurate
- Production build working
- No security vulnerabilities in dependencies

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
7. **Mobile Apps**: Native iOS/Android apps (React Native or PWA)
8. **Plugins**: Extensibility system for custom integrations
9. **Internationalization**: Multi-language support (i18n)
10. **Configuration Backup**: Scheduled backups and version history

---

## Dependencies

### New Backend Dependencies

```json
{
  "dependencies": {
    // No additional dependencies needed!
    // Bun provides everything: HTTP server, WebSocket, static file serving
  }
}
```

### Frontend Dependencies

```json
{
  "dependencies": {
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "react-router-dom": "^6.22.0",
    "@mui/material": "^6.1.0",
    "@mui/icons-material": "^6.1.0",
    "@emotion/react": "^11.13.0",
    "@emotion/styled": "^11.13.0",
    "react-window": "^1.8.10"
  },
  "devDependencies": {
    "@types/react": "^18.3.1",
    "@types/react-dom": "^18.3.0",
    "@types/react-window": "^1.8.8",
    "@vitejs/plugin-react": "^4.3.0",
    "vite": "^5.4.0",
    "typescript": "^5.6.0",
    "@testing-library/react": "^16.0.0",
    "@testing-library/jest-dom": "^6.5.0",
    "vitest": "^2.0.0"
  }
}
```

---

## Technical Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| WebSocket connection instability | Users miss real-time logs | Implement robust reconnection logic, fallback to polling |
| Large log datasets causing performance issues | UI becomes sluggish | Use virtualization, pagination, and efficient filtering |
| Browser compatibility issues | Users can't access UI | Test on all major browsers, use polyfills if needed |
| Build size too large | Slow initial load | Code splitting, tree shaking, lazy loading routes |
| API breaking changes affecting UI | Frontend breaks on backend updates | Version API, maintain backwards compatibility |
| Memory leaks from WebSocket subscriptions | Browser tab becomes unresponsive | Proper cleanup in useEffect hooks, connection monitoring |

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
