# Activity Logging: Recording Triggered Search Operations

## Context

The system must maintain a log of all searches triggered by the automation,
providing visibility into what actions the system has taken. This log serves as
an audit trail and troubleshooting tool, allowing the user to verify that
automation is working as expected.

This is the only mechanism for the user to see what the automation has done.

## Requirements

### Story: Log Triggered Searches

- **As a** user
- **I want to** the system to record every search it triggers
- **So that** I can verify automation is working and see what content is being
  searched for

#### Acceptance Criteria

- [ ] Every triggered search is logged with: timestamp, server name, server type
      (Radarr/Sonarr), search category (missing or cutoff-not-met), content title,
      and Radarr/Sonarr item ID
- [ ] Each search creates a separate log entry (one entry per movie/episode
      searched, not grouped)
- [ ] Log entries are created immediately when searches are triggered
- [ ] Log persists across application restarts
- [ ] Content titles and IDs allow cross-referencing with Radarr/Sonarr UIs

### Story: Log Automation Cycle Events

- **As a** user
- **I want to** the system to record when automation cycles start and complete
- **So that** I can see the system is running on schedule

#### Acceptance Criteria

- [ ] Each automation cycle start is logged with timestamp
- [ ] Each automation cycle completion is logged with timestamp and summary
      (total searches triggered)
- [ ] Manual triggers are clearly marked as manual vs scheduled

### Story: Log Failures

- **As a** user
- **I want to** the system to record when operations fail
- **So that** I can troubleshoot issues and understand why automation isn't
  working as expected

#### Acceptance Criteria

- [ ] Server connection failures during detection are logged with timestamp,
      server name, and failure reason
- [ ] Failed search triggers are logged with timestamp, server name, and failure
      reason
- [ ] Failed operations are visually distinguishable from successful operations
      in the log

### Story: View Activity Log

- **As a** user
- **I want to** see recent activity in a clear, readable format
- **So that** I can quickly understand what the system has been doing

#### Acceptance Criteria

- [ ] Activity log is displayed in reverse chronological order (newest first)
- [ ] Log entries show date and time in readable format
- [ ] User can see at least the most recent 100 log entries
- [ ] Log interface clearly distinguishes between: cycle events, successful
      searches, and failures

### Story: Clear Old Logs

- **As a** user
- **I want to** the system to automatically manage log size
- **So that** logs don't grow indefinitely and consume excessive storage

#### Acceptance Criteria

- [ ] System retains logs for at least 30 days
- [ ] Logs older than 30 days are automatically purged
- [ ] User can manually clear all logs if desired
- [ ] Log clearing is confirmed before executing

## Edge Cases & Constraints

### Log Storage

- Log entries should be lightweight (text-based, minimal data per entry)
- Consider a maximum log size (e.g., 10MB) with automatic rotation if needed
- Don't log sensitive information (no API keys, no full URLs if they contain
  credentials)

### Performance

- Logging should not significantly impact automation performance
- Log display should be fast even with thousands of entries
- Consider pagination or virtual scrolling if displaying very large logs

### Log Detail Level

- Log should be detailed enough to troubleshoot issues and provide full
  visibility into what content is being searched
- Log meaningful events: cycle start/end, individual searches triggered
  (with titles and IDs), and failures
- Individual log entries per search provide granular audit trail
  - Example: "Triggered search for Breaking Bad S01E01 [ID:12345]"
  - More verbose but enables users to see exactly what's being searched
- Use efficient storage and display techniques (virtualization, pagination) to
  handle potentially large log volumes

### Time Representation

- All timestamps should use consistent time zone (user's local time or UTC,
  clearly indicated)
- Timestamps should include date and time, not just time

### Data Integrity

- Log should survive application crashes when possible
- Write log entries immediately, don't buffer them in memory

### User Experience

- Failed operations should be easy to spot (color coding, icons, or clear
  labels)
- Recent activity should be visible on the main interface (don't require digging
  through menus)
- Consider a summary view: "Last cycle: 12 searches triggered, 2 failures" with
  option to expand details

### Known Limitations

- The system logs what searches were triggered, not whether searches found
  content (that's Radarr/Sonarr's responsibility)
- Individual log entries mean higher log volumes compared to grouped entries
- Very old logs are purged automatically to prevent unbounded growth
- Recommend using react-window or similar virtualization for displaying large
  log lists efficiently
