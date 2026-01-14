# Server Configuration: Managing Radarr and Sonarr Connections

## Context

Users need to configure one or more Radarr and Sonarr server instances before
the system can detect missing content or trigger searches. Each server requires
a URL and API key. The system must validate connectivity before accepting a
configuration to prevent automation failures.

This is the foundational topic - without valid server configurations, no other
automation can occur.

## Requirements

### Story: Add New Media Server

- **As a** user
- **I want to** add Radarr and Sonarr server details (URL and API key)
- **So that** the system can connect to my media servers for automation

#### Acceptance Criteria

- [ ] User can input server URL and API key through the interface
- [ ] User can specify whether the server is Radarr or Sonarr type
- [ ] System validates URL format before attempting connection
- [ ] System tests API connectivity using provided credentials
- [ ] Server is only saved if connection test passes
- [ ] User receives clear feedback if connection test fails (invalid URL, wrong
      API key, server unreachable)
- [ ] Successfully added servers are immediately available for automation

### Story: View Configured Servers

- **As a** user
- **I want to** see all my configured Radarr and Sonarr servers
- **So that** I know which servers are being managed by the automation

#### Acceptance Criteria

- [ ] User can view a list of all configured servers
- [ ] Each server entry shows: server type (Radarr/Sonarr), URL, and a masked
      version of the API key
- [ ] List distinguishes between Radarr and Sonarr servers visually

### Story: Edit Existing Server

- **As a** user
- **I want to** modify the URL or API key of a configured server
- **So that** I can update credentials or server locations without recreating
  the configuration

#### Acceptance Criteria

- [ ] User can select an existing server to edit
- [ ] User can modify URL and/or API key
- [ ] System re-validates connectivity with new details before saving changes
- [ ] Changes are only applied if new connection test passes
- [ ] User receives feedback if validation fails

### Story: Remove Server

- **As a** user
- **I want to** delete a configured server
- **So that** I can stop managing servers I no longer use

#### Acceptance Criteria

- [ ] User can select a server and remove it from configuration
- [ ] Removal is immediate and the server is no longer checked during automation
      runs
- [ ] User receives confirmation before deletion occurs

## Edge Cases & Constraints

### Connection Validation

- Test connections with a timeout (suggest 10-15 seconds) to avoid indefinite
  waiting
- API key validation should use a minimal API call that confirms authentication
  without retrieving large datasets
- Handle common Radarr/Sonarr API response codes: 200 (success), 401
  (unauthorized), 404 (not found)

### Data Integrity

- Prevent duplicate server entries (same URL and type)
- Handle trailing slashes in URLs consistently
- Validate that URL uses http:// or https:// protocol

### Security

- Never display full API keys in the interface (show only first/last few
  characters or fully mask)
- Store API keys securely (encryption at rest if possible)
- Do not log API keys in activity logs

### User Experience

- Connection tests should provide specific failure reasons (network error, wrong
  credentials, invalid API endpoint)
- If a server is temporarily unreachable during automated runs, log the failure
  but don't disable the server configuration
- Support both local network URLs (192.168.x.x) and remote URLs (domain names)
