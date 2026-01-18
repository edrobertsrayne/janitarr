# Janitarr Web UI

Modern, responsive web interface for managing Janitarr automation.

## Technology Stack

- **React 19** - UI framework
- **React Router v7** - Client-side routing
- **Material-UI v7** - Component library
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Emotion** - CSS-in-JS styling

## Features

- **Dashboard** - Overview of automation activity and system status
- **Server Management** - Add, edit, and test Radarr/Sonarr connections
- **Activity Logs** - View detailed execution logs with filtering
- **Settings** - Configure automation limits and schedules
- **Theme Support** - Light, dark, and system-based themes
- **Responsive Design** - Optimized for desktop, tablet, and mobile

## Development

### Prerequisites

Ensure you're in the Janitarr devenv (run `direnv allow` in project root).

### Install Dependencies

```bash
bun install
```

### Development Server

```bash
bun run dev
```

The UI will be available at http://localhost:5173

**API Proxy:** The Vite dev server proxies `/api` requests to `http://localhost:3000` where the backend API should be running.

### Build for Production

```bash
bun run build
```

Builds to `../dist/public` for serving by the backend API server.

### Type Checking

```bash
bunx tsc --noEmit
```

### Linting

```bash
bunx eslint .
```

## Testing

### Manual Testing

1. Start the UI dev server: `bun run dev`
2. Start the backend API: `cd .. && bun run start`
3. Open http://localhost:5173 in your browser
4. Test all views and functionality

**Test checklist:**

- [ ] Navigation works between all views
- [ ] Theme toggle cycles through light/dark/system
- [ ] Mobile responsive design (resize browser)
- [ ] Forms validate input correctly
- [ ] API requests succeed (check browser console)
- [ ] Error states display properly

## Project Structure

```
ui/
├── src/
│   ├── components/
│   │   ├── common/          # Reusable components
│   │   │   ├── ConfirmDialog.tsx
│   │   │   ├── LoadingSpinner.tsx
│   │   │   └── StatusBadge.tsx
│   │   └── layout/
│   │       └── Layout.tsx   # AppBar + Navigation Drawer
│   ├── contexts/
│   │   └── ThemeContext.tsx # Theme management
│   ├── views/               # Page components
│   │   ├── Dashboard.tsx
│   │   ├── Servers.tsx
│   │   ├── Logs.tsx
│   │   └── Settings.tsx
│   ├── App.tsx              # Route configuration
│   └── main.tsx             # Entry point
├── public/                  # Static assets
├── vite.config.ts          # Vite configuration
└── tsconfig.json           # TypeScript config
```

## Architecture

### Routing

React Router v7 with client-side routing:

- `/` - Dashboard (default)
- `/servers` - Server management
- `/logs` - Activity logs
- `/settings` - Configuration

All routes share a common layout with:

- Responsive navigation drawer (sidebar on desktop, hamburger menu on mobile)
- AppBar with title and theme toggle
- Main content area

### Theme System

Three-mode theme implementation:

- **Light mode** - Standard light theme
- **Dark mode** - Dark theme
- **System mode** - Follows OS preference

Theme state persists in localStorage and is managed by `ThemeContext`.

### Responsive Design

- **Desktop (≥960px):** Permanent sidebar navigation (240px width)
- **Tablet/Mobile (<960px):** Hamburger menu with temporary drawer

Uses Material-UI breakpoints for consistent responsive behavior.

### API Integration

The UI communicates with the backend API at `/api/*` endpoints:

- `/api/servers` - Server CRUD operations
- `/api/logs` - Activity log retrieval
- `/api/config` - Configuration management
- `/api/status` - System status

In development, Vite proxies these to `http://localhost:3000`.

## Known Limitations

No known limitations at this time.

## Troubleshooting

### Dev server won't start

- Check that port 5173 is available
- Ensure dependencies are installed: `bun install`
- Verify you're in the devenv: `direnv allow`

### API requests fail

- Ensure the backend server is running on port 3000
- Check the proxy configuration in `vite.config.ts`
- Verify CORS settings on the backend

### Build fails

- Run type checking: `bunx tsc --noEmit`
- Check for linting errors: `bunx eslint .`
- Ensure all dependencies are installed

### Theme not persisting

- Check browser localStorage is enabled
- Verify ThemeContext is properly wrapping the app
- Clear browser cache and reload
