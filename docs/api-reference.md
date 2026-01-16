# Janitarr API Reference

Complete reference for the Janitarr REST API and WebSocket protocol.

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [REST API Endpoints](#rest-api-endpoints)
  - [Configuration](#configuration)
  - [Servers](#servers)
  - [Logs](#logs)
  - [Automation](#automation)
  - [Statistics](#statistics)
  - [Health](#health)
- [WebSocket Protocol](#websocket-protocol)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)

---

## Overview

**Base URL**: `http://localhost:3000/api`

**Protocol**: HTTP/1.1

**Data Format**: JSON

**CORS**: Enabled for all origins (development mode)

**Content Type**: `application/json` for request/response bodies

---

## Authentication

**Current Status**: No authentication required

The API is designed for local network use and does not currently implement authentication. It is recommended to:
- Run behind a reverse proxy with authentication
- Use firewall rules to restrict access
- Not expose directly to the internet

---

## REST API Endpoints

### Configuration

Manage application configuration including schedule and search limits.

#### Get Configuration

Retrieve current configuration.

**Endpoint**: `GET /api/config`

**Response**: `200 OK`

```json
{
  "schedule": {
    "enabled": true,
    "interval": 6
  },
  "limits": {
    "missing": {
      "movies": 10,
      "episodes": 10
    },
    "cutoff": {
      "movies": 5,
      "episodes": 5
    }
  }
}
```

**Response Fields**:
- `schedule.enabled` (boolean): Whether automation is enabled
- `schedule.interval` (number): Hours between automation cycles (min: 1)
- `limits.missing.movies` (number): Max Radarr missing movie searches per cycle
- `limits.missing.episodes` (number): Max Sonarr missing episode searches per cycle
- `limits.cutoff.movies` (number): Max Radarr quality upgrade searches per cycle
- `limits.cutoff.episodes` (number): Max Sonarr quality upgrade searches per cycle

---

#### Update Configuration

Update configuration values.

**Endpoint**: `PATCH /api/config`

**Request Body**:

```json
{
  "schedule": {
    "enabled": true,
    "interval": 8
  },
  "limits": {
    "missing": {
      "movies": 15,
      "episodes": 20
    },
    "cutoff": {
      "movies": 5,
      "episodes": 10
    }
  }
}
```

**Partial updates supported** - only include fields you want to change:

```json
{
  "schedule": {
    "interval": 4
  }
}
```

**Response**: `200 OK`

```json
{
  "schedule": {
    "enabled": true,
    "interval": 4
  },
  "limits": {
    "missing": {
      "movies": 10,
      "episodes": 10
    },
    "cutoff": {
      "movies": 5,
      "episodes": 5
    }
  }
}
```

**Validation**:
- `schedule.interval` must be ≥ 1
- All limit values must be ≥ 0
- `schedule.enabled` must be boolean

**Errors**:
- `400 Bad Request`: Invalid configuration values

---

#### Reset Configuration

Reset configuration to default values.

**Endpoint**: `PUT /api/config/reset`

**Response**: `200 OK`

```json
{
  "schedule": {
    "enabled": true,
    "interval": 6
  },
  "limits": {
    "missing": {
      "movies": 10,
      "episodes": 10
    },
    "cutoff": {
      "movies": 5,
      "episodes": 5
    }
  }
}
```

---

### Servers

Manage Radarr and Sonarr server configurations.

#### List Servers

Retrieve all configured servers.

**Endpoint**: `GET /api/servers`

**Query Parameters**:
- `type` (optional): Filter by server type (`radarr` or `sonarr`)

**Examples**:
```
GET /api/servers
GET /api/servers?type=radarr
GET /api/servers?type=sonarr
```

**Response**: `200 OK`

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Main Radarr",
    "type": "radarr",
    "url": "http://192.168.1.100:7878",
    "apiKey": "r4nd0m...xyz",
    "enabled": true,
    "createdAt": "2024-01-15T10:30:00.000Z",
    "updatedAt": "2024-01-15T10:30:00.000Z"
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "Main Sonarr",
    "type": "sonarr",
    "url": "http://192.168.1.100:8989",
    "apiKey": "s0m3th...abc",
    "enabled": true,
    "createdAt": "2024-01-15T10:35:00.000Z",
    "updatedAt": "2024-01-15T10:35:00.000Z"
  }
]
```

**Response Fields**:
- `id` (string): Unique server identifier (UUID)
- `name` (string): Server display name
- `type` (string): Server type (`radarr` or `sonarr`)
- `url` (string): Server base URL
- `apiKey` (string): Masked API key (first 6 + last 3 characters)
- `enabled` (boolean): Whether server is active
- `createdAt` (string): ISO 8601 timestamp
- `updatedAt` (string): ISO 8601 timestamp

---

#### Get Server

Retrieve a single server by ID.

**Endpoint**: `GET /api/servers/:id`

**Path Parameters**:
- `id` (string): Server UUID

**Example**:
```
GET /api/servers/550e8400-e29b-41d4-a716-446655440000
```

**Response**: `200 OK`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Main Radarr",
  "type": "radarr",
  "url": "http://192.168.1.100:7878",
  "apiKey": "r4nd0m...xyz",
  "enabled": true,
  "createdAt": "2024-01-15T10:30:00.000Z",
  "updatedAt": "2024-01-15T10:30:00.000Z"
}
```

**Errors**:
- `404 Not Found`: Server ID does not exist

---

#### Create Server

Add a new server.

**Endpoint**: `POST /api/servers`

**Request Body**:

```json
{
  "name": "4K Radarr",
  "type": "radarr",
  "url": "http://192.168.1.101:7878",
  "apiKey": "your-api-key-here",
  "enabled": true
}
```

**Request Fields**:
- `name` (string, required): Unique server name
- `type` (string, required): `radarr` or `sonarr`
- `url` (string, required): Server base URL
- `apiKey` (string, required): API key from server settings
- `enabled` (boolean, optional): Default `true`

**Response**: `201 Created`

```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "name": "4K Radarr",
  "type": "radarr",
  "url": "http://192.168.1.101:7878",
  "apiKey": "your-a...ere",
  "enabled": true,
  "createdAt": "2024-01-15T11:00:00.000Z",
  "updatedAt": "2024-01-15T11:00:00.000Z"
}
```

**Validation**:
- `name` must be unique
- `type` must be `radarr` or `sonarr`
- `url` must be valid URL format
- `apiKey` cannot be empty

**Errors**:
- `400 Bad Request`: Validation failed
- `409 Conflict`: Server name already exists

---

#### Update Server

Update existing server configuration.

**Endpoint**: `PUT /api/servers/:id`

**Path Parameters**:
- `id` (string): Server UUID

**Request Body**:

```json
{
  "name": "4K Radarr",
  "url": "http://192.168.1.101:7878",
  "apiKey": "new-api-key",
  "enabled": false
}
```

**Request Fields** (all optional except name):
- `name` (string): Server name
- `url` (string): Server URL
- `apiKey` (string): API key
- `enabled` (boolean): Active status

**Note**: `type` cannot be changed after creation

**Response**: `200 OK`

```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "name": "4K Radarr",
  "type": "radarr",
  "url": "http://192.168.1.101:7878",
  "apiKey": "new-ap...key",
  "enabled": false,
  "createdAt": "2024-01-15T11:00:00.000Z",
  "updatedAt": "2024-01-15T11:15:00.000Z"
}
```

**Errors**:
- `400 Bad Request`: Validation failed
- `404 Not Found`: Server ID does not exist
- `409 Conflict`: New name conflicts with existing server

---

#### Delete Server

Remove a server configuration.

**Endpoint**: `DELETE /api/servers/:id`

**Path Parameters**:
- `id` (string): Server UUID

**Example**:
```
DELETE /api/servers/770e8400-e29b-41d4-a716-446655440002
```

**Response**: `204 No Content`

**Errors**:
- `404 Not Found`: Server ID does not exist

---

#### Test Server Connection

Verify server connectivity and API key.

**Endpoint**: `POST /api/servers/:id/test`

**Path Parameters**:
- `id` (string): Server UUID

**Example**:
```
POST /api/servers/550e8400-e29b-41d4-a716-446655440000/test
```

**Response**: `200 OK`

```json
{
  "success": true,
  "serverName": "Main Radarr",
  "serverVersion": "4.3.2.6857",
  "message": "Connection successful"
}
```

**Response Fields**:
- `success` (boolean): Connection test result
- `serverName` (string): Server name from API response
- `serverVersion` (string, optional): Server version if successful
- `message` (string): Status message

**Error Response**: `200 OK` (with `success: false`)

```json
{
  "success": false,
  "message": "Connection failed: Network timeout"
}
```

---

### Logs

Retrieve and manage activity logs.

#### Get Logs

Retrieve activity logs with optional filtering.

**Endpoint**: `GET /api/logs`

**Query Parameters**:
- `type` (optional): Filter by log type (`automation`, `search`, `server-test`, `error`)
- `serverId` (optional): Filter by server UUID
- `search` (optional): Text search in log details
- `limit` (optional): Max results to return (default: 100)
- `offset` (optional): Pagination offset (default: 0)

**Examples**:
```
GET /api/logs
GET /api/logs?type=search
GET /api/logs?serverId=550e8400-e29b-41d4-a716-446655440000
GET /api/logs?search=Movie+Title
GET /api/logs?limit=50&offset=100
```

**Response**: `200 OK`

```json
{
  "logs": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440010",
      "timestamp": "2024-01-15T12:00:00.000Z",
      "type": "automation",
      "serverId": null,
      "details": "Automation cycle started (manual)",
      "metadata": {
        "trigger": "manual",
        "cycleId": "cycle-123"
      }
    },
    {
      "id": "990e8400-e29b-41d4-a716-446655440011",
      "timestamp": "2024-01-15T12:00:15.000Z",
      "type": "search",
      "serverId": "550e8400-e29b-41d4-a716-446655440000",
      "details": "Triggered search for: The Matrix (1999)",
      "metadata": {
        "category": "missing",
        "itemId": 123,
        "title": "The Matrix",
        "year": 1999
      }
    }
  ],
  "total": 2,
  "limit": 100,
  "offset": 0
}
```

**Response Fields**:
- `logs` (array): Log entries
  - `id` (string): Unique log entry UUID
  - `timestamp` (string): ISO 8601 timestamp
  - `type` (string): Log type
  - `serverId` (string | null): Associated server UUID
  - `details` (string): Human-readable description
  - `metadata` (object): Additional structured data
- `total` (number): Total log entries matching filters
- `limit` (number): Applied limit
- `offset` (number): Applied offset

**Log Types**:
- `automation`: Automation cycle events
- `search`: Search trigger events
- `server-test`: Connection test events
- `error`: Error events

---

#### Delete Logs

Clear all activity logs.

**Endpoint**: `DELETE /api/logs`

**Response**: `204 No Content`

**Note**: This action is permanent and cannot be undone.

---

#### Export Logs

Export logs in JSON or CSV format.

**Endpoint**: `GET /api/logs/export`

**Query Parameters**:
- `format` (required): Export format (`json` or `csv`)
- `type` (optional): Filter by log type
- `serverId` (optional): Filter by server UUID
- `search` (optional): Text search filter

**Examples**:
```
GET /api/logs/export?format=json
GET /api/logs/export?format=csv
GET /api/logs/export?format=json&type=search
```

**Response** (format=json): `200 OK`

```json
[
  {
    "id": "880e8400-e29b-41d4-a716-446655440010",
    "timestamp": "2024-01-15T12:00:00.000Z",
    "type": "automation",
    "serverId": null,
    "details": "Automation cycle started (manual)",
    "metadata": "{\"trigger\":\"manual\"}"
  }
]
```

**Response** (format=csv): `200 OK`

```csv
id,timestamp,type,serverId,details,metadata
880e8400-e29b-41d4-a716-446655440010,2024-01-15T12:00:00.000Z,automation,,Automation cycle started (manual),"{""trigger"":""manual""}"
```

**Content-Type**:
- JSON: `application/json`
- CSV: `text/csv`

**Content-Disposition**: `attachment; filename="janitarr-logs-YYYYMMDD-HHMMSS.{json|csv}"`

**Errors**:
- `400 Bad Request`: Invalid format parameter

---

### Automation

Trigger and monitor automation cycles.

#### Trigger Automation

Manually trigger an automation cycle.

**Endpoint**: `POST /api/automation/trigger`

**Request Body** (optional):

```json
{
  "dryRun": false
}
```

**Request Fields**:
- `dryRun` (boolean, optional): If true, preview only (default: false)

**Response**: `202 Accepted`

```json
{
  "message": "Automation cycle triggered",
  "dryRun": false
}
```

**Response Fields**:
- `message` (string): Status message
- `dryRun` (boolean): Whether this is a dry-run

**Note**: Automation runs asynchronously. Check logs for results.

**Errors**:
- `409 Conflict`: Automation cycle already running

---

#### Get Automation Status

Retrieve automation scheduler status.

**Endpoint**: `GET /api/automation/status`

**Response**: `200 OK`

```json
{
  "enabled": true,
  "running": false,
  "nextRunTime": "2024-01-15T18:00:00.000Z",
  "lastRunTime": "2024-01-15T12:00:00.000Z",
  "interval": 6
}
```

**Response Fields**:
- `enabled` (boolean): Whether automation is enabled
- `running` (boolean): Whether a cycle is currently running
- `nextRunTime` (string | null): ISO 8601 timestamp of next cycle
- `lastRunTime` (string | null): ISO 8601 timestamp of last cycle
- `interval` (number): Hours between cycles

---

### Statistics

Retrieve dashboard statistics.

#### Get Stats Summary

Get aggregated statistics across all servers.

**Endpoint**: `GET /api/stats/summary`

**Response**: `200 OK`

```json
{
  "missingMovies": 150,
  "missingEpisodes": 320,
  "cutoffUpgrades": 45,
  "totalSearches": 1250,
  "lastAutomationRun": "2024-01-15T12:00:00.000Z"
}
```

**Response Fields**:
- `missingMovies` (number): Total missing movies across all Radarr servers
- `missingEpisodes` (number): Total missing episodes across all Sonarr servers
- `cutoffUpgrades` (number): Total items below quality cutoff
- `totalSearches` (number): Lifetime search count from logs
- `lastAutomationRun` (string | null): ISO 8601 timestamp of last cycle

---

#### Get Server Stats

Get statistics for a specific server.

**Endpoint**: `GET /api/stats/servers/:id`

**Path Parameters**:
- `id` (string): Server UUID

**Example**:
```
GET /api/stats/servers/550e8400-e29b-41d4-a716-446655440000
```

**Response**: `200 OK`

```json
{
  "serverId": "550e8400-e29b-41d4-a716-446655440000",
  "serverName": "Main Radarr",
  "missingCount": 75,
  "cutoffCount": 20,
  "searchCount": 450,
  "lastSearch": "2024-01-15T12:00:15.000Z"
}
```

**Response Fields**:
- `serverId` (string): Server UUID
- `serverName` (string): Server display name
- `missingCount` (number): Missing items for this server
- `cutoffCount` (number): Items below cutoff for this server
- `searchCount` (number): Total searches triggered for this server
- `lastSearch` (string | null): ISO 8601 timestamp of last search

**Errors**:
- `404 Not Found`: Server ID does not exist

---

### Health

Health check endpoint for monitoring.

#### Health Check

Check if the API is running.

**Endpoint**: `GET /api/health`

**Response**: `200 OK`

```json
{
  "status": "ok",
  "timestamp": "2024-01-15T12:30:00.000Z"
}
```

**Response Fields**:
- `status` (string): Always `"ok"` if server is running
- `timestamp` (string): Current server time in ISO 8601 format

**Use Cases**:
- Load balancer health checks
- Monitoring systems
- Uptime verification

---

## WebSocket Protocol

Real-time log streaming via WebSocket.

### Connection

**Endpoint**: `ws://localhost:3000/ws/logs`

**Protocol**: WebSocket (RFC 6455)

### Connection Lifecycle

1. **Connect**: Open WebSocket connection to `/ws/logs`
2. **Subscribe**: Receive all new log entries in real-time
3. **Filter** (optional): Send filter message to receive subset of logs
4. **Disconnect**: Close connection when done

### Message Format

All messages are JSON-encoded.

### Server → Client Messages

#### Log Entry

Sent whenever a new log entry is created.

```json
{
  "type": "log",
  "data": {
    "id": "880e8400-e29b-41d4-a716-446655440010",
    "timestamp": "2024-01-15T12:00:00.000Z",
    "type": "automation",
    "serverId": null,
    "details": "Automation cycle started (manual)",
    "metadata": {
      "trigger": "manual"
    }
  }
}
```

#### Welcome Message

Sent immediately after connection established.

```json
{
  "type": "welcome",
  "message": "Connected to Janitarr logs stream"
}
```

#### Error Message

Sent if an error occurs.

```json
{
  "type": "error",
  "message": "Filter parse error: Invalid JSON"
}
```

### Client → Server Messages

#### Set Filter

Apply server-side filtering to log stream.

```json
{
  "type": "filter",
  "filter": {
    "type": "search",
    "serverId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Filter Fields** (all optional):
- `type` (string): Log type filter
- `serverId` (string): Server UUID filter
- `search` (string): Text search filter

**Clear Filter**:

```json
{
  "type": "filter",
  "filter": null
}
```

#### Ping

Keep connection alive (optional).

```json
{
  "type": "ping"
}
```

**Response**: Pong message

```json
{
  "type": "pong"
}
```

### Example JavaScript Client

```javascript
const ws = new WebSocket('ws://localhost:3000/ws/logs');

ws.onopen = () => {
  console.log('Connected to logs stream');

  // Apply filter (optional)
  ws.send(JSON.stringify({
    type: 'filter',
    filter: { type: 'search' }
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  if (message.type === 'log') {
    console.log('New log:', message.data);
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from logs stream');
};
```

### Reconnection Strategy

**Recommended**: Implement exponential backoff for reconnection.

```javascript
let reconnectDelay = 1000;
const maxDelay = 30000;

function connect() {
  const ws = new WebSocket('ws://localhost:3000/ws/logs');

  ws.onopen = () => {
    reconnectDelay = 1000; // Reset on successful connection
  };

  ws.onclose = () => {
    setTimeout(() => {
      reconnectDelay = Math.min(reconnectDelay * 2, maxDelay);
      connect();
    }, reconnectDelay);
  };
}

connect();
```

---

## Error Handling

### HTTP Status Codes

- `200 OK`: Request succeeded
- `201 Created`: Resource created successfully
- `202 Accepted`: Request accepted for processing
- `204 No Content`: Request succeeded, no response body
- `400 Bad Request`: Invalid request data
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (duplicate, concurrent operation)
- `500 Internal Server Error`: Server error

### Error Response Format

All errors return JSON:

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common Error Scenarios

**Invalid JSON**:
```json
{
  "error": "Invalid JSON in request body"
}
```

**Validation Error**:
```json
{
  "error": "Validation failed: name is required"
}
```

**Not Found**:
```json
{
  "error": "Server not found"
}
```

**Conflict**:
```json
{
  "error": "Server name already exists"
}
```

---

## Rate Limiting

**Current Status**: No rate limiting implemented

For production deployments, implement rate limiting at the reverse proxy level (NGINX, Caddy, etc.) to prevent abuse.

**Recommended limits**:
- GET requests: 100 requests/minute
- POST/PUT/DELETE requests: 20 requests/minute
- WebSocket connections: 5 connections/minute

---

## API Examples

### cURL Examples

**Get configuration**:
```bash
curl http://localhost:3000/api/config
```

**Update configuration**:
```bash
curl -X PATCH http://localhost:3000/api/config \
  -H "Content-Type: application/json" \
  -d '{"schedule":{"interval":4}}'
```

**Add server**:
```bash
curl -X POST http://localhost:3000/api/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Radarr",
    "type": "radarr",
    "url": "http://localhost:7878",
    "apiKey": "your-api-key",
    "enabled": true
  }'
```

**Get logs**:
```bash
curl http://localhost:3000/api/logs?type=search&limit=10
```

**Export logs to CSV**:
```bash
curl http://localhost:3000/api/logs/export?format=csv > logs.csv
```

**Trigger automation**:
```bash
curl -X POST http://localhost:3000/api/automation/trigger \
  -H "Content-Type: application/json" \
  -d '{"dryRun":false}'
```

### JavaScript/TypeScript Examples

**Fetch configuration**:
```typescript
const config = await fetch('http://localhost:3000/api/config')
  .then(res => res.json());
console.log(config);
```

**Create server**:
```typescript
const server = await fetch('http://localhost:3000/api/servers', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    name: 'Test Radarr',
    type: 'radarr',
    url: 'http://localhost:7878',
    apiKey: 'your-api-key',
    enabled: true,
  }),
}).then(res => res.json());
console.log('Created server:', server);
```

**WebSocket connection**:
```typescript
const ws = new WebSocket('ws://localhost:3000/ws/logs');

ws.addEventListener('message', (event) => {
  const message = JSON.parse(event.data);
  if (message.type === 'log') {
    console.log('New log:', message.data.details);
  }
});
```

---

## Versioning

**Current Version**: 1.0

The API is currently unversioned. Breaking changes will be avoided when possible. When breaking changes are necessary, a versioning scheme will be introduced (e.g., `/api/v2/...`).

---

## Support

For issues, questions, or contributions:
- [User Guide](user-guide.md)
- [Troubleshooting Guide](troubleshooting.md)
- [GitHub Issues](https://github.com/yourusername/janitarr/issues)
