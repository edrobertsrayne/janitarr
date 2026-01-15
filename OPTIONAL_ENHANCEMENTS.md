# Optional Enhancements for Janitarr

**Status:** All core features complete. These enhancements are optional and non-blocking.

## Overview

Janitarr is production-ready with all planned features implemented. This document outlines optional enhancements that could improve the project based on specific deployment needs or user feedback.

---

## Enhancement 1: Docker Support

**Priority:** Low
**Effort:** 2-4 hours
**Impact:** Simplifies deployment in containerized environments

### Current State
- Application runs directly with Bun runtime
- Users must install Bun and manage dependencies manually
- No standardized deployment method

### Proposed Implementation

#### 1.1 Create Dockerfile

```dockerfile
# Dockerfile
FROM oven/bun:1 AS builder

WORKDIR /app

# Copy package files
COPY package.json bun.lock* ./

# Install dependencies
RUN bun install --frozen-lockfile --production

# Copy source code
COPY src ./src
COPY tsconfig.json ./

# Build if needed (optional for Bun)
# RUN bun build src/index.ts --outdir dist --target bun

FROM oven/bun:1-slim

WORKDIR /app

# Copy from builder
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./
COPY --from=builder /app/src ./src
COPY --from=builder /app/tsconfig.json ./

# Create data directory
RUN mkdir -p /app/data

# Set environment variables
ENV JANITARR_DB_PATH=/app/data/janitarr.db
ENV JANITARR_LOG_LEVEL=info

# Volume for persistent data
VOLUME ["/app/data"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD bun run src/index.ts status --json || exit 1

# Run the application
ENTRYPOINT ["bun", "run", "src/index.ts"]
CMD ["start"]
```

#### 1.2 Create docker-compose.yml

```yaml
version: '3.8'

services:
  janitarr:
    build: .
    container_name: janitarr
    restart: unless-stopped

    environment:
      - JANITARR_DB_PATH=/app/data/janitarr.db
      - JANITARR_LOG_LEVEL=info
      - TZ=UTC

    volumes:
      - ./data:/app/data

    # If Radarr/Sonarr are on the same Docker network
    networks:
      - media

    # Health check
    healthcheck:
      test: ["CMD", "bun", "run", "src/index.ts", "status", "--json"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

networks:
  media:
    external: true
```

#### 1.3 Create .dockerignore

```
node_modules
data
.git
.env
*.md
tests
.gitignore
.eslintrc.json
```

#### 1.4 Update README.md

Add Docker deployment section:

```markdown
## Docker Deployment

### Using Docker Compose (Recommended)

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd janitarr
   ```

2. **Configure your servers:**
   ```bash
   docker-compose run --rm janitarr server add
   ```

3. **Start the scheduler:**
   ```bash
   docker-compose up -d
   ```

4. **View logs:**
   ```bash
   docker-compose logs -f
   ```

### Using Docker Directly

```bash
# Build the image
docker build -t janitarr .

# Run interactively
docker run -it --rm -v $(pwd)/data:/app/data janitarr server add

# Run as daemon
docker run -d --name janitarr -v $(pwd)/data:/app/data janitarr start
```
```

### Testing Plan
1. Build Docker image locally
2. Test interactive commands (server add, config set)
3. Test daemon mode (start, stop)
4. Verify data persistence across container restarts
5. Test health check endpoint
6. Verify volume mounts

---

## Enhancement 2: API Key Encryption

**Priority:** Low
**Effort:** 4-6 hours
**Impact:** Enhanced security for multi-user or shared deployments

### Current State
- API keys stored as plain text in SQLite database
- File system permissions provide basic protection
- Keys never logged or exposed in CLI output (masked)
- Suitable for single-user deployments

### Proposed Implementation

#### 2.1 Add Encryption Library

```typescript
// src/lib/crypto.ts
import { createCipheriv, createDecipheriv, randomBytes } from "crypto";

const ALGORITHM = "aes-256-gcm";
const KEY_LENGTH = 32;
const IV_LENGTH = 16;
const TAG_LENGTH = 16;

/**
 * Get or create encryption key
 * Stored in environment or generated once per database
 */
function getEncryptionKey(): Buffer {
  const keyFromEnv = process.env.JANITARR_ENCRYPTION_KEY;

  if (keyFromEnv) {
    return Buffer.from(keyFromEnv, "hex");
  }

  // For single-user deployments, derive key from machine ID
  // This is transparent to the user but provides encryption at rest
  const machineId = getMachineId(); // Platform-specific
  return createHash("sha256").update(machineId).digest();
}

/**
 * Encrypt API key
 */
export function encryptApiKey(plaintext: string): string {
  const key = getEncryptionKey();
  const iv = randomBytes(IV_LENGTH);

  const cipher = createCipheriv(ALGORITHM, key, iv);

  let encrypted = cipher.update(plaintext, "utf8", "hex");
  encrypted += cipher.final("hex");

  const tag = cipher.getAuthTag();

  // Format: iv:tag:ciphertext
  return `${iv.toString("hex")}:${tag.toString("hex")}:${encrypted}`;
}

/**
 * Decrypt API key
 */
export function decryptApiKey(encrypted: string): string {
  const key = getEncryptionKey();
  const [ivHex, tagHex, ciphertext] = encrypted.split(":");

  const iv = Buffer.from(ivHex, "hex");
  const tag = Buffer.from(tagHex, "hex");

  const decipher = createDecipheriv(ALGORITHM, key, iv);
  decipher.setAuthTag(tag);

  let decrypted = decipher.update(ciphertext, "hex", "utf8");
  decrypted += decipher.final("utf8");

  return decrypted;
}
```

#### 2.2 Update Database Layer

```typescript
// src/storage/database.ts
import { encryptApiKey, decryptApiKey } from "../lib/crypto";

// In addServer method:
const encryptedKey = encryptApiKey(apiKey);
stmt.run(id, name, normalizedUrl, encryptedKey, type, now, now);

// In getServer method:
const decryptedKey = decryptApiKey(row.api_key);
return {
  ...server,
  apiKey: decryptedKey
};
```

#### 2.3 Migration Strategy

```typescript
// src/storage/migrations/001_encrypt_keys.ts
export function migrateApiKeys(db: Database): void {
  const servers = db.prepare("SELECT * FROM servers").all();

  for (const server of servers) {
    // Check if already encrypted (contains colons)
    if (!server.api_key.includes(":")) {
      const encrypted = encryptApiKey(server.api_key);
      db.prepare("UPDATE servers SET api_key = ? WHERE id = ?")
        .run(encrypted, server.id);
    }
  }
}
```

#### 2.4 Environment Variable Documentation

Add to README.md:

```markdown
### Security

**API Key Encryption:**

By default, API keys are encrypted at rest using a key derived from your system. For enhanced security in shared environments, set an explicit encryption key:

```bash
export JANITARR_ENCRYPTION_KEY=$(openssl rand -hex 32)
```

Store this key securely - you'll need it to decrypt existing configurations.
```

### Testing Plan
1. Test encryption/decryption round-trip
2. Test with environment variable key
3. Test with machine-derived key
4. Verify migration from plain text
5. Ensure backwards compatibility
6. Test key rotation scenario

### Considerations
- **Backwards Compatibility:** Migration required for existing databases
- **Key Management:** Users must backup encryption key if explicitly set
- **Performance:** Minimal impact (encrypt/decrypt only on read/write)
- **Security Model:** Protects against casual file inspection, not sophisticated attacks

---

## Enhancement 3: Web UI

**Priority:** Low
**Effort:** 40-60 hours (major feature)
**Impact:** Alternative interface for users who prefer browser-based tools

### Current State
- Complete CLI implementation
- All functionality accessible via command line
- JSON output available for scripting

### Proposed Architecture

#### Technology Stack
- **Backend:** Hono (lightweight web framework for Bun)
- **Frontend:** HTMX + Alpine.js (minimal JavaScript)
- **Styling:** Tailwind CSS (utility-first CSS)
- **Template Engine:** EJS or native template literals

#### High-Level Design

```
src/
├── web/
│   ├── server.ts          # Web server setup
│   ├── routes/
│   │   ├── dashboard.ts   # Main dashboard
│   │   ├── servers.ts     # Server management
│   │   ├── config.ts      # Configuration
│   │   └── logs.ts        # Activity logs
│   ├── templates/
│   │   ├── layout.ejs     # Base layout
│   │   ├── dashboard.ejs  # Dashboard view
│   │   ├── servers.ejs    # Server list/edit
│   │   └── logs.ejs       # Log viewer
│   └── public/
│       ├── styles.css     # Compiled Tailwind
│       └── app.js         # Minimal client-side JS
```

#### Key Features
1. **Dashboard:** Status overview, next run time, recent activity
2. **Server Management:** CRUD operations with inline editing
3. **Configuration:** Form-based config updates
4. **Log Viewer:** Real-time log display with filtering
5. **Manual Triggers:** Button to run cycle immediately
6. **Scheduler Control:** Start/stop scheduler from UI

#### Implementation Phases

**Phase 1: Basic Server (8-10 hours)**
- Set up Hono web server
- Create layout and navigation
- Implement dashboard route

**Phase 2: Server Management (8-10 hours)**
- Server list view with HTMX
- Add/edit server forms
- Delete confirmation dialogs
- Connection testing

**Phase 3: Configuration (6-8 hours)**
- Configuration form
- Real-time validation
- Save/cancel actions

**Phase 4: Logs & Activity (10-12 hours)**
- Log table with pagination
- Real-time updates (SSE or polling)
- Filtering by type/date
- Clear logs functionality

**Phase 5: Automation Controls (8-10 hours)**
- Manual cycle trigger
- Scheduler start/stop
- Real-time status updates
- Progress indicators

### Sample Code

```typescript
// src/web/server.ts
import { Hono } from "hono";
import { serveStatic } from "hono/bun";
import { listServers } from "../services/server-manager";
import { getStatus } from "../lib/scheduler";

const app = new Hono();

// Static files
app.use("/static/*", serveStatic({ root: "./" }));

// Dashboard
app.get("/", (c) => {
  const servers = listServers();
  const status = getStatus();

  return c.html(`
    <!DOCTYPE html>
    <html>
      <head>
        <title>Janitarr</title>
        <link href="/static/styles.css" rel="stylesheet">
        <script src="https://unpkg.com/htmx.org@1.9.10"></script>
      </head>
      <body>
        <div class="container">
          <h1>Janitarr Dashboard</h1>

          <div class="stats">
            <div class="stat">
              <span>Servers</span>
              <span>${servers.length}</span>
            </div>
            <div class="stat">
              <span>Status</span>
              <span>${status.isRunning ? "Running" : "Stopped"}</span>
            </div>
          </div>

          <div class="actions">
            <button hx-post="/api/run" hx-swap="outerHTML">
              Run Now
            </button>
          </div>
        </div>
      </body>
    </html>
  `);
});

// Start server
export function startWebServer(port = 3000) {
  console.log(`Web UI available at http://localhost:${port}`);
  Bun.serve({
    port,
    fetch: app.fetch,
  });
}
```

### Testing Plan
1. Unit tests for route handlers
2. Integration tests for API endpoints
3. E2E tests with Playwright (browser automation)
4. Accessibility testing (WCAG compliance)
5. Mobile responsive testing

### Considerations
- **Port Configuration:** Add `JANITARR_WEB_PORT` environment variable
- **Authentication:** Add basic auth for multi-user deployments
- **API Endpoints:** RESTful API for HTMX to consume
- **Real-time Updates:** WebSocket or SSE for live status
- **Security:** CSRF protection, input sanitization

---

## Enhancement 4: Binary Packaging

**Priority:** Low
**Effort:** 1-2 hours
**Impact:** Simplifies distribution (no Bun runtime required)

### Current State
- Users must install Bun runtime
- Application distributed as source code
- Requires `bun run` to execute

### Proposed Implementation

Use Bun's built-in compilation feature:

```bash
# Build standalone binary
bun build --compile src/index.ts --outfile janitarr

# Results in a single executable
./janitarr --help
```

#### Platform-Specific Builds

```json
// package.json
{
  "scripts": {
    "build:linux": "bun build --compile --target bun-linux-x64 src/index.ts --outfile dist/janitarr-linux",
    "build:macos": "bun build --compile --target bun-darwin-x64 src/index.ts --outfile dist/janitarr-macos",
    "build:macos-arm": "bun build --compile --target bun-darwin-arm64 src/index.ts --outfile dist/janitarr-macos-arm64",
    "build:windows": "bun build --compile --target bun-windows-x64 src/index.ts --outfile dist/janitarr.exe",
    "build:all": "bun run build:linux && bun run build:macos && bun run build:macos-arm && bun run build:windows"
  }
}
```

#### GitHub Actions Release Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            target: bun-linux-x64
            artifact: janitarr-linux-x64
          - os: macos-latest
            target: bun-darwin-x64
            artifact: janitarr-macos-x64
          - os: macos-latest
            target: bun-darwin-arm64
            artifact: janitarr-macos-arm64
          - os: windows-latest
            target: bun-windows-x64
            artifact: janitarr-windows-x64.exe

    steps:
      - uses: actions/checkout@v3

      - uses: oven-sh/setup-bun@v1
        with:
          bun-version: latest

      - run: bun install

      - run: bun test

      - name: Build
        run: bun build --compile --target ${{ matrix.target }} src/index.ts --outfile ${{ matrix.artifact }}

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: ${{ matrix.artifact }}
          path: ${{ matrix.artifact }}
```

### Benefits
- No Bun runtime required for end users
- Single executable file
- Faster startup time
- Easier distribution via GitHub Releases

### Considerations
- Binary size (~40-50 MB with Bun runtime embedded)
- Platform-specific builds required
- Database path handling (relative paths)
- Update mechanism (no auto-update by default)

---

## Priority Matrix

| Enhancement | Priority | Effort | User Value | Implementation Order |
|-------------|----------|--------|------------|---------------------|
| Docker Support | Medium | Low | High | 1st (if deploying to servers) |
| Binary Packaging | Low | Very Low | Medium | 2nd (easy win) |
| API Key Encryption | Low | Medium | Low | 3rd (security conscious) |
| Web UI | Low | High | Medium | 4th (significant effort) |

---

## Implementation Guidelines

### Before Starting
1. Create a new branch for each enhancement
2. Update tests to maintain 100% passing rate
3. Update documentation (README, AGENTS.md)
4. Consider backwards compatibility

### Testing Requirements
- All existing tests must continue passing
- New functionality must have test coverage
- Integration tests for external-facing features
- Documentation examples must be validated

### Documentation Updates
Each enhancement requires:
- README.md updates (installation, configuration, usage)
- AGENTS.md updates (new dependencies, build steps)
- CHANGELOG.md entry
- Updated version number in package.json

---

## Conclusion

These enhancements are entirely optional. Janitarr is production-ready without them. Prioritize based on:
- **Deployment environment** (Docker for servers)
- **Distribution method** (Binary for end users)
- **Security requirements** (Encryption for shared systems)
- **User preference** (Web UI for GUI users)

Start with the enhancement that provides the most value for your specific use case.
