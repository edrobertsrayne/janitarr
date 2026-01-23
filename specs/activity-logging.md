# Activity Logging: Recording Triggered Search Operations

## Context

The system must maintain a log of all searches triggered by the automation,
providing visibility into what actions the system has taken. This log serves as
an audit trail and troubleshooting tool, allowing the user to verify that
automation is working as expected.

Activity logging is part of the unified logging system (see [logging.md](./logging.md)).
All activity events are written to:

- **Database**: Persistent storage for history and audit trail
- **Console**: Real-time output via charmbracelet/log
- **Web interface**: Real-time streaming via WebSocket

This specification focuses on _what_ activity events are logged. For details on
_how_ logging works (levels, filtering, retention, web UI), see [logging.md](./logging.md).

## Requirements

### Story: Log Triggered Searches

- **As a** user
- **I want** the system to record every search it triggers
- **So that** I can verify automation is working and see what content is being
  searched for

#### Acceptance Criteria

- [ ] Every triggered search is logged with: timestamp, server name, server type
      (Radarr/Sonarr), search category (missing or cutoff-not-met), content title,
      and Radarr/Sonarr item ID
- [ ] Movie searches include: title, year, quality profile
- [ ] Episode searches include: series title, season/episode number, episode title, quality profile
- [ ] Each search creates a separate log entry (one entry per movie/episode
      searched, not grouped)
- [ ] Log entries are created immediately when searches are triggered
- [ ] Log persists across application restarts (stored in database)
- [ ] Content titles and IDs allow cross-referencing with Radarr/Sonarr UIs

#### Log Format Examples

Movie search:

```
INFO Search triggered title="The Matrix" year=1999 quality="HD-1080p" server=radarr-main category=missing id=12345
```

Episode search:

```
INFO Search triggered series="Breaking Bad" episode="S01E01" title="Pilot" quality="HD-1080p" server=sonarr-main category=cutoff_unmet id=67890
```

### Story: Log Automation Cycle Events

- **As a** user
- **I want** the system to record when automation cycles start and complete
- **So that** I can see the system is running on schedule

#### Acceptance Criteria

- [ ] Each automation cycle start is logged with timestamp and trigger type
- [ ] Detection results logged per server: count of missing items and cutoff-unmet items
- [ ] Each automation cycle completion is logged with timestamp, duration, and summary
      (total searches triggered, failures)
- [ ] Manual triggers are clearly marked as manual vs scheduled

#### Log Format Examples

Cycle start:

```
INFO Automation cycle started trigger=scheduled
```

Detection results (per server):

```
INFO Detection complete server=radarr-main missing=5 cutoff_unmet=12
INFO Detection complete server=sonarr-main missing=23 cutoff_unmet=8
```

Cycle end:

```
INFO Automation cycle completed duration=45s searches_triggered=15 failures=0
```

### Story: Log Failures

- **As a** user
- **I want** the system to record when operations fail
- **So that** I can troubleshoot issues and understand why automation isn't
  working as expected

#### Acceptance Criteria

- [ ] Server connection failures during detection are logged with timestamp,
      server name, and failure reason
- [ ] Failed search triggers are logged with timestamp, server name, content title, and failure
      reason
- [ ] Rate limiting events logged with retry information
- [ ] Failed operations logged at `error` level for visibility
- [ ] Failed operations are visually distinguishable from successful operations
      in all outputs (console color, web UI styling)

#### Log Format Examples

Connection failure:

```
ERROR Server connection failed server=radarr-main error="connection refused"
```

Search failure:

```
ERROR Search failed title="The Matrix" server=radarr-main error="API timeout"
```

Rate limiting:

```
WARN Rate limited server=radarr-main retry_after=30s
```

### Story: View Activity Log

- **As a** user
- **I want to** see recent activity in a clear, readable format
- **So that** I can quickly understand what the system has been doing

#### Acceptance Criteria

Activity log viewing is handled by the unified logging system. See [logging.md](./logging.md) for:

- Web interface log viewer requirements (reverse chronological order, pagination, filtering)
- Console output formatting and color coding
- Real-time streaming via WebSocket
- Dashboard log summary widget

### Story: Clear Old Logs

- **As a** user
- **I want** the system to automatically manage log size
- **So that** logs don't grow indefinitely and consume excessive storage

#### Acceptance Criteria

Log retention and cleanup is handled by the unified logging system. See [logging.md](./logging.md) for:

- Retention period configuration (30 days default, 7-90 day range)
- Automatic daily cleanup of old logs
- Manual "Clear all logs" option
- Log count display in settings

## Edge Cases & Constraints

For general logging constraints (performance, storage, sensitive data handling, time zones,
concurrency, error handling), see [logging.md](./logging.md#edge-cases--constraints).

### Activity-Specific Constraints

**Log Detail Level:**

- Log meaningful events: cycle start/end, detection results, individual searches triggered
  (with titles and IDs), and failures
- Individual log entries per search provide granular audit trail
- Content titles and IDs allow cross-referencing with Radarr/Sonarr UIs

**Known Limitations:**

- The system logs what searches were triggered, not whether searches found
  content (that's Radarr/Sonarr's responsibility)
- Individual log entries mean higher log volumes compared to grouped entries

## Related Specifications

- [logging.md](./logging.md): Unified logging system (console, web, database)
- [search-triggering.md](./search-triggering.md): Search operation details
- [automatic-scheduling.md](./automatic-scheduling.md): Automation cycle scheduling
