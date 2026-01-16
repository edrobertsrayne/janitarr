# Janitarr Development Guide

Guide for developers contributing to Janitarr.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Architecture](#project-architecture)
- [Backend Development](#backend-development)
- [Frontend Development](#frontend-development)
- [Testing](#testing)
- [Code Standards](#code-standards)
- [Deployment](#deployment)
- [Contributing](#contributing)

---

## Development Setup

### Prerequisites

- **Bun** v1.0+ - [Install from bun.sh](https://bun.sh/)
- **Git** for version control
- **Text editor** - VS Code recommended
- **Development media server** - Optional but recommended for testing

### Quick Start

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd janitarr
   ```

2. **Allow devenv** (if using direnv):
   ```bash
   direnv allow
   ```

   This loads the development environment with all necessary tools.

3. **Install backend dependencies**:
   ```bash
   bun install
   ```

4. **Install frontend dependencies**:
   ```bash
   cd ui
   bun install
   cd ..
   ```

5. **Run tests to verify setup**:
   ```bash
   bun test
   ```

### Development Environment

The project uses [devenv](https://devenv.sh) with direnv for automatic environment loading.

**First-time setup**:
```bash
direnv allow
```

This provides:
- Bun runtime
- Node.js (for tooling)
- Playwright (for UI tests)
- Git and build tools

The environment loads automatically when entering the project directory.

### Environment Variables

Create a `.env` file for development:

```bash
# Database location (optional, defaults to ./data/janitarr.db)
JANITARR_DB_PATH=./data/dev.db

# Logging level (debug, info, warn, error)
JANITARR_LOG_LEVEL=debug

# Server port (optional, defaults to 3000)
PORT=3000
```

**Never commit `.env` files!** They're gitignored by default.

### Test Media Servers

For development, set up test Radarr/Sonarr instances:

1. **Using Docker** (recommended):
   ```bash
   # Radarr
   docker run -d \
     --name radarr-test \
     -p 7878:7878 \
     -v radarr-config:/config \
     linuxserver/radarr:latest

   # Sonarr
   docker run -d \
     --name sonarr-test \
     -p 8989:8989 \
     -v sonarr-config:/config \
     linuxserver/sonarr:latest
   ```

2. **Configure test servers**:
   - Access Radarr at `http://localhost:7878`
   - Access Sonarr at `http://localhost:8989`
   - Get API keys from Settings → General → Security
   - Add to Janitarr for testing

---

## Project Architecture

### Directory Structure

```
janitarr/
├── src/                    # Backend source code
│   ├── lib/                # Shared utilities
│   │   ├── api-client.ts   # Radarr/Sonarr API client
│   │   ├── crypto.ts       # Encryption utilities
│   │   ├── logger.ts       # Activity logging
│   │   └── scheduler.ts    # Background scheduling
│   ├── services/           # Business logic
│   │   ├── server-manager.ts  # Server CRUD
│   │   ├── detector.ts        # Content detection
│   │   ├── search-trigger.ts  # Search execution
│   │   └── automation.ts      # Orchestration
│   ├── storage/            # Data layer
│   │   └── database.ts     # SQLite interface
│   ├── web/                # Web backend
│   │   ├── routes.ts       # REST API routes
│   │   ├── websocket.ts    # WebSocket server
│   │   └── server.ts       # HTTP server
│   ├── cli/                # CLI interface
│   │   ├── commands.ts     # Command definitions
│   │   └── formatters.ts   # Output formatting
│   ├── types.ts            # TypeScript types
│   └── index.ts            # CLI entry point
├── ui/                     # Frontend source code
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── views/          # Page components
│   │   ├── services/       # API client
│   │   ├── contexts/       # React contexts
│   │   └── test/           # Test utilities
│   ├── public/             # Static assets
│   └── vite.config.ts      # Build configuration
├── tests/                  # Backend tests
│   ├── lib/                # Utility tests
│   ├── services/           # Service tests
│   ├── storage/            # Database tests
│   └── integration/        # Integration tests
├── specs/                  # Requirements docs
├── docs/                   # Documentation
├── data/                   # Database (gitignored)
└── dist/                   # Build output (gitignored)
```

### Backend Architecture

**Layered architecture**:

1. **CLI Layer** (`src/cli/`):
   - User interface
   - Command parsing
   - Output formatting

2. **Service Layer** (`src/services/`):
   - Business logic
   - Orchestration
   - Validation

3. **Library Layer** (`src/lib/`):
   - Reusable utilities
   - API clients
   - Scheduling

4. **Storage Layer** (`src/storage/`):
   - Data persistence
   - Database operations
   - Migrations

5. **Web Layer** (`src/web/`):
   - REST API
   - WebSocket server
   - HTTP server

**Key design patterns**:

- **Result pattern**: Services return `{ success, data?, error? }`
- **Dependency injection**: Services receive dependencies as parameters
- **Separation of concerns**: Each layer has clear responsibility
- **Fail gracefully**: Partial failures don't crash the system

### Frontend Architecture

**React SPA with Material-UI**:

- **React 19** - UI framework
- **React Router v7** - Client-side routing
- **Material-UI v7** - Component library
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server

**Component hierarchy**:

```
App
├── Layout (AppBar + Navigation)
│   ├── Dashboard
│   │   ├── StatusCard
│   │   ├── ServerStatusTable
│   │   └── RecentActivity
│   ├── Servers
│   │   ├── ServerList/ServerCards
│   │   └── AddServerDialog
│   ├── Logs
│   │   ├── LogList
│   │   └── LogFilters
│   └── Settings
│       └── SettingsForm
└── Common Components
    ├── LoadingSpinner
    ├── StatusBadge
    └── ConfirmDialog
```

**State management**:

- **React Context** for theme
- **Component state** for UI state
- **API calls** for server state
- **WebSocket** for real-time updates

### Database Schema

**SQLite database with 3 tables**:

1. **config**:
   - Key-value configuration storage
   - Holds limits, schedule settings

2. **servers**:
   - Server configurations
   - Encrypted API keys
   - Type, URL, enabled status

3. **logs**:
   - Activity logging
   - Automation events
   - Search history

**Schema migrations**:
- Handled in `src/storage/database.ts`
- Automatic on startup
- Version tracked in config table

---

## Backend Development

### Running the Backend

**Development mode** (auto-reload):
```bash
bun run dev
```

**Production mode**:
```bash
bun run start
```

**CLI commands** (without starting server):
```bash
bun run src/index.ts <command>
```

### Adding a New CLI Command

1. **Define command** in `src/cli/commands.ts`:
   ```typescript
   program
     .command('my-command')
     .description('Description of command')
     .option('-f, --flag', 'Optional flag')
     .action(async (options) => {
       // Implementation
     });
   ```

2. **Add service logic** in appropriate service file

3. **Add tests** in `tests/cli/` or `tests/services/`

4. **Update documentation** in `docs/user-guide.md`

### Adding a New API Endpoint

1. **Define route** in `src/web/routes.ts`:
   ```typescript
   export async function handleMyEndpoint(req: Request): Promise<Response> {
     // Validate request
     // Call service
     // Return JSON response
   }
   ```

2. **Add to router** in `src/web/server.ts`:
   ```typescript
   if (path === '/api/my-endpoint' && method === 'GET') {
     return handleMyEndpoint(req);
   }
   ```

3. **Add frontend client** in `ui/src/services/api.ts`:
   ```typescript
   export async function myApiCall(): Promise<MyData> {
     const response = await fetch('/api/my-endpoint');
     if (!response.ok) throw new Error('Failed to fetch');
     return response.json();
   }
   ```

4. **Add tests** for both backend and frontend

5. **Update documentation** in `docs/api-reference.md`

### Working with the Database

**Using the database**:
```typescript
import { getDatabase } from './storage/database';

const db = getDatabase();
const result = db.query('SELECT * FROM servers').all();
```

**Best practices**:
- Use prepared statements for SQL queries
- Handle errors gracefully
- Close database connections properly
- Use transactions for multi-step operations

**Schema changes**:
1. Update `initializeDatabase()` in `src/storage/database.ts`
2. Add migration logic if needed
3. Test with existing database
4. Document changes

### Encryption and Security

**API Key Encryption**:

API keys are encrypted at rest using AES-256-GCM:

```typescript
import { encrypt, decrypt } from './lib/crypto';

const encrypted = encrypt(apiKey);  // Store this
const decrypted = decrypt(encrypted);  // Use this for API calls
```

**Security considerations**:
- Keys encrypted at rest in database
- Keys decrypted only when needed
- Encryption key derived from system entropy
- No key material in logs or error messages

### Logging

**Activity logging**:
```typescript
import { ActivityLogger } from './lib/logger';

const logger = new ActivityLogger();

logger.logAutomationStart(trigger);
logger.logSearchTrigger(serverId, category, item, success, error);
logger.logAutomationComplete(summary);
```

**Console logging** (development):
```typescript
console.log('[DEBUG]', 'Detailed information');
console.info('[INFO]', 'General information');
console.warn('[WARN]', 'Warning message');
console.error('[ERROR]', 'Error message');
```

Use `JANITARR_LOG_LEVEL` environment variable to control verbosity.

---

## Frontend Development

### Running the Frontend

**Development mode** (with hot reload):
```bash
cd ui
bun run dev
```

Access at `http://localhost:5173`

**Production build**:
```bash
cd ui
bun run build
```

Builds to `../dist/public/` for serving by backend.

### Component Development

**Creating a new component**:

1. Create file in appropriate directory:
   - `ui/src/components/common/` - Reusable components
   - `ui/src/components/layout/` - Layout components
   - `ui/src/views/` - Page components

2. Use TypeScript and Material-UI:
   ```typescript
   import { Box, Button } from '@mui/material';

   interface MyComponentProps {
     title: string;
     onAction: () => void;
   }

   export function MyComponent({ title, onAction }: MyComponentProps) {
     return (
       <Box>
         <Button onClick={onAction}>{title}</Button>
       </Box>
     );
   }
   ```

3. **Write tests** in `ui/src/components/common/*.test.tsx`

4. **Export** from component directory if needed

### Styling

**Material-UI theming**:

Theme is configured in `ui/src/contexts/ThemeContext.tsx`:
- Light mode
- Dark mode
- System preference

**Custom styles**:

Use Material-UI's `sx` prop:
```typescript
<Box sx={{ padding: 2, backgroundColor: 'primary.main' }}>
  Content
</Box>
```

Or `styled` components:
```typescript
import { styled } from '@mui/material/styles';

const StyledBox = styled(Box)(({ theme }) => ({
  padding: theme.spacing(2),
  backgroundColor: theme.palette.primary.main,
}));
```

### API Integration

**API client** is in `ui/src/services/api.ts`:

```typescript
// GET request
export async function getServers(): Promise<ServerConfig[]> {
  const response = await fetch('/api/servers');
  if (!response.ok) throw new Error('Failed to fetch servers');
  return response.json();
}

// POST request
export async function createServer(server: ServerInput): Promise<ServerConfig> {
  const response = await fetch('/api/servers', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(server),
  });
  if (!response.ok) throw new Error('Failed to create server');
  return response.json();
}
```

**Using in components**:
```typescript
import { getServers } from '../services/api';

function MyComponent() {
  const [servers, setServers] = useState<ServerConfig[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getServers()
      .then(setServers)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingSpinner />;
  return <div>{/* Render servers */}</div>;
}
```

### WebSocket

**WebSocket client** is in `ui/src/services/websocket.ts`:

```typescript
import { WebSocketClient } from '../services/websocket';

function LogsView() {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const ws = useRef<WebSocketClient>();

  useEffect(() => {
    ws.current = new WebSocketClient('ws://localhost:3000/ws/logs');

    ws.current.onMessage((log: LogEntry) => {
      setLogs(prev => [log, ...prev]);
    });

    ws.current.connect();

    return () => ws.current?.disconnect();
  }, []);

  // ...
}
```

### Responsive Design

**Material-UI breakpoints**:
- `xs`: 0px+ (mobile)
- `sm`: 600px+ (small tablet)
- `md`: 960px+ (tablet/laptop)
- `lg`: 1280px+ (desktop)
- `xl`: 1920px+ (large desktop)

**Responsive props**:
```typescript
<Box
  sx={{
    width: { xs: '100%', md: '50%' },
    display: { xs: 'block', md: 'flex' },
  }}
>
```

**Responsive Grid**:
```typescript
<Grid container spacing={2}>
  <Grid item xs={12} sm={6} md={4}>
    {/* Full width on mobile, half on tablet, third on desktop */}
  </Grid>
</Grid>
```

### Accessibility

**ARIA labels**:
```typescript
<IconButton aria-label="Delete server">
  <DeleteIcon />
</IconButton>
```

**Keyboard navigation**:
- Use semantic HTML (`<button>`, `<a>`, etc.)
- Material-UI handles focus management
- Test with keyboard only (Tab, Enter, Escape)

**Screen readers**:
- Use `aria-live` for dynamic content
- Provide `aria-label` for icon-only buttons
- Use proper heading hierarchy (`h1`, `h2`, etc.)

---

## Testing

### Backend Tests

**Run all tests**:
```bash
bun test
```

**Run specific test file**:
```bash
bun test tests/services/detector.test.ts
```

**Run with coverage** (when supported):
```bash
bun test --coverage
```

**Writing tests**:

Use the built-in `test` function:

```typescript
import { describe, test, expect } from 'bun:test';

describe('My Service', () => {
  test('should do something', () => {
    const result = myFunction();
    expect(result).toBe(expected);
  });

  test('should handle errors', () => {
    expect(() => myFunction()).toThrow('Error message');
  });
});
```

**Testing async code**:
```typescript
test('should fetch data', async () => {
  const data = await fetchData();
  expect(data).toBeDefined();
});
```

**Testing database operations**:

Use in-memory database for tests:

```typescript
import { Database } from 'bun:sqlite';

test('should store data', () => {
  const db = new Database(':memory:');
  // Initialize schema
  // Test operations
  db.close();
});
```

### Frontend Tests

**Run frontend tests**:
```bash
cd ui
bun test
```

**Run with UI**:
```bash
cd ui
bun test --ui
```

**Run specific test**:
```bash
cd ui
bun test src/components/common/StatusBadge.test.tsx
```

**Writing component tests**:

Use Vitest + React Testing Library:

```typescript
import { render, screen } from '../test/utils';
import { MyComponent } from './MyComponent';

describe('MyComponent', () => {
  it('renders title', () => {
    render(<MyComponent title="Hello" />);
    expect(screen.getByText('Hello')).toBeInTheDocument();
  });

  it('calls handler on click', async () => {
    const handler = vi.fn();
    render(<MyComponent onAction={handler} />);

    await userEvent.click(screen.getByRole('button'));
    expect(handler).toHaveBeenCalled();
  });
});
```

**Testing API calls**:

Mock fetch:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getServers } from './api';

beforeEach(() => {
  global.fetch = vi.fn();
});

it('fetches servers', async () => {
  const mockServers = [{ id: '1', name: 'Test' }];
  (global.fetch as any).mockResolvedValue({
    ok: true,
    json: async () => mockServers,
  });

  const servers = await getServers();
  expect(servers).toEqual(mockServers);
  expect(fetch).toHaveBeenCalledWith('/api/servers');
});
```

### Integration Tests

**Backend integration tests** test real API calls to Radarr/Sonarr.

**Setup**:

Create `.env` with test credentials:
```bash
TEST_RADARR_URL=http://localhost:7878
TEST_RADARR_API_KEY=your-key-here
TEST_SONARR_URL=http://localhost:8989
TEST_SONARR_API_KEY=your-key-here
```

**Run integration tests**:
```bash
bun test tests/integration/
```

**Writing integration tests**:
```typescript
import { describe, test, expect } from 'bun:test';
import { ApiClient } from '../src/lib/api-client';

describe('Radarr Integration', () => {
  test('should connect to real server', async () => {
    const client = new ApiClient(
      process.env.TEST_RADARR_URL!,
      process.env.TEST_RADARR_API_KEY!
    );

    const status = await client.testConnection();
    expect(status.success).toBe(true);
  });
});
```

### E2E Tests

**Playwright tests** for end-to-end UI testing.

**Run E2E tests**:
```bash
bunx playwright test
```

**Run with UI**:
```bash
bunx playwright test --ui
```

**Before running E2E tests**, start both servers:
```bash
# Terminal 1
cd ui && bun run dev

# Terminal 2
bun run start
```

**Writing E2E tests**:
```typescript
import { test, expect } from '@playwright/test';

test('should add server', async ({ page }) => {
  await page.goto('http://localhost:5173');

  await page.click('text=Servers');
  await page.click('text=Add Server');

  await page.fill('input[name="name"]', 'Test Server');
  await page.fill('input[name="url"]', 'http://localhost:7878');
  await page.fill('input[name="apiKey"]', 'test-key');

  await page.click('text=Add');

  await expect(page.locator('text=Test Server')).toBeVisible();
});
```

---

## Code Standards

### TypeScript

**Strict mode enabled**:
- No implicit `any`
- Null checks enforced
- All types defined in `src/types.ts`

**Naming conventions**:
- **Files**: `kebab-case.ts`
- **Types**: `PascalCase`
- **Interfaces**: `PascalCase`
- **Functions**: `camelCase`
- **Constants**: `UPPER_SNAKE_CASE`

**Type definitions**:
```typescript
// Domain types
export interface ServerConfig {
  id: string;
  name: string;
  type: 'radarr' | 'sonarr';
  // ...
}

// Database row types (separate from domain types)
interface ServerRow {
  id: string;
  name: string;
  type: string;
  api_key_encrypted: string;
  // ...
}
```

### Code Organization

**File structure**:
- One class/service per file
- Related functions in same file
- Separate concerns (API, business logic, storage)

**Function size**:
- Keep functions small (<50 lines)
- Extract complex logic to separate functions
- Single responsibility principle

**Error handling**:
```typescript
// Services return Result pattern
return { success: true, data: result };
return { success: false, error: 'Error message' };

// Throw errors for unexpected conditions
throw new Error('Unexpected condition');
```

### Code Style

**Formatting**:
- 2 spaces for indentation
- Single quotes for strings
- Semicolons optional
- Trailing commas in multi-line

**Use ESLint**:
```bash
bunx eslint .
```

**Use TypeScript compiler**:
```bash
bunx tsc --noEmit
```

### Documentation

**Code comments**:
- Use JSDoc for public APIs
- Explain "why" not "what"
- Document non-obvious behavior

```typescript
/**
 * Triggers searches for detected content across all servers.
 *
 * Applies configured limits per category and distributes searches
 * fairly across servers using round-robin.
 *
 * @param detectionResults - Results from content detection
 * @param limits - Search limits per category
 * @param dryRun - If true, skips actual search triggering
 * @returns Summary of triggered searches
 */
export function triggerSearches(
  detectionResults: DetectionResults,
  limits: SearchLimits,
  dryRun = false
): SearchSummary {
  // Implementation
}
```

---

## Deployment

### Production Build

1. **Build frontend**:
   ```bash
   cd ui
   bun install
   bun run build
   cd ..
   ```

   Output: `dist/public/`

2. **Install backend dependencies**:
   ```bash
   bun install --production
   ```

3. **Start server**:
   ```bash
   bun run start
   ```

### Running as a Service

**Using systemd** (Linux):

Create `/etc/systemd/system/janitarr.service`:

```ini
[Unit]
Description=Janitarr Automation Service
After=network.target

[Service]
Type=simple
User=janitarr
WorkingDirectory=/opt/janitarr
ExecStart=/usr/local/bin/bun run start
Restart=on-failure
RestartSec=10
Environment=JANITARR_DB_PATH=/var/lib/janitarr/janitarr.db
Environment=JANITARR_LOG_LEVEL=info

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable janitarr
sudo systemctl start janitarr
sudo systemctl status janitarr
```

**Using Docker**:

Create `Dockerfile`:
```dockerfile
FROM oven/bun:1

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --production

COPY src ./src
COPY dist ./dist

EXPOSE 3000

CMD ["bun", "run", "start"]
```

Build and run:
```bash
docker build -t janitarr .
docker run -d \
  -p 3000:3000 \
  -v janitarr-data:/app/data \
  --name janitarr \
  janitarr
```

### Reverse Proxy

**NGINX configuration**:

```nginx
server {
    listen 80;
    server_name janitarr.example.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    location /ws {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }
}
```

### Environment Configuration

**Production environment variables**:

```bash
# Database location
JANITARR_DB_PATH=/var/lib/janitarr/janitarr.db

# Logging
JANITARR_LOG_LEVEL=info

# Server port
PORT=3000
```

### Backup and Restore

**Backup database**:
```bash
cp /path/to/janitarr.db /backup/location/janitarr-$(date +%Y%m%d).db
```

**Automated backups** (cron):
```cron
# Daily backup at 2 AM
0 2 * * * cp /var/lib/janitarr/janitarr.db /backup/janitarr-$(date +\%Y\%m\%d).db
```

**Restore database**:
```bash
# Stop Janitarr
sudo systemctl stop janitarr

# Restore backup
cp /backup/janitarr-20240115.db /var/lib/janitarr/janitarr.db

# Start Janitarr
sudo systemctl start janitarr
```

---

## Contributing

### Getting Started

1. **Fork the repository**
2. **Clone your fork**:
   ```bash
   git clone https://github.com/yourusername/janitarr
   cd janitarr
   ```
3. **Create a branch**:
   ```bash
   git checkout -b feature/my-feature
   ```

### Development Workflow

1. **Make changes**
   - Follow code standards
   - Write tests
   - Update documentation

2. **Run tests**:
   ```bash
   bun test
   bunx tsc --noEmit
   bunx eslint .
   ```

3. **Commit changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

   Use conventional commit format:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation
   - `refactor:` - Code refactoring
   - `test:` - Test changes
   - `chore:` - Build/tooling changes

4. **Push to your fork**:
   ```bash
   git push origin feature/my-feature
   ```

5. **Create pull request**
   - Describe changes
   - Reference issues
   - Include screenshots if UI changes

### Pull Request Guidelines

**Before submitting**:
- [ ] Tests pass (`bun test`)
- [ ] TypeScript compiles (`bunx tsc --noEmit`)
- [ ] ESLint passes (`bunx eslint .`)
- [ ] Documentation updated
- [ ] Commit messages follow convention

**PR description should include**:
- What: Brief description of changes
- Why: Reason for changes
- How: Implementation approach
- Testing: How you tested changes
- Screenshots: If UI changes

### Code Review Process

1. **Maintainer reviews PR**
2. **Address feedback**:
   - Make requested changes
   - Push to same branch
   - PR updates automatically
3. **Approval and merge**:
   - Maintainer merges when ready
   - Branch deleted after merge

### Reporting Issues

**Bug reports should include**:
- Janitarr version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Relevant logs

**Feature requests should include**:
- Use case description
- Proposed solution
- Alternatives considered
- Additional context

### Community Guidelines

- Be respectful and inclusive
- Help others learn and grow
- Provide constructive feedback
- Follow code of conduct

---

## Additional Resources

- [User Guide](user-guide.md) - For end users
- [API Reference](api-reference.md) - REST API documentation
- [Troubleshooting Guide](troubleshooting.md) - Common issues

Need help? Open an issue on GitHub!
