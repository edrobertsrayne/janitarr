# Missing Content Detection: Identifying Gaps in Media Libraries

## Context

The system must identify episodes (in Sonarr) and movies (in Radarr) that are
marked as monitored but are not present in the library. This is one of two
detection mechanisms (the other being quality cutoff detection) that feeds into
the search triggering system.

Missing content represents gaps in the library that need to be filled. The
detection runs against all configured servers during each automation cycle.

## Requirements

### Story: Detect Missing Episodes

- **As a** user
- **I want to** the system to identify monitored TV episodes that are missing
  from my library
- **So that** searches can be triggered to acquire them

#### Acceptance Criteria

- [ ] System queries each configured Sonarr server for monitored episodes marked
      as missing
- [ ] System counts total missing episodes across all Sonarr servers
- [ ] Missing count is available to the search triggering system
- [ ] Detection runs against all configured Sonarr servers in each automation
      cycle

### Story: Detect Missing Movies

- **As a** user
- **I want to** the system to identify monitored movies that are missing from my
  library
- **So that** searches can be triggered to acquire them

#### Acceptance Criteria

- [ ] System queries each configured Radarr server for monitored movies marked
      as missing
- [ ] System counts total missing movies across all Radarr servers
- [ ] Missing count is available to the search triggering system
- [ ] Detection runs against all configured Radarr servers in each automation
      cycle

### Story: Handle Detection Failures

- **As a** user
- **I want to** the system to continue operating even if one server is
  temporarily unreachable
- **So that** automation doesn't completely fail due to a single server issue

#### Acceptance Criteria

- [ ] If a server is unreachable during detection, system logs the failure and
      continues checking other servers
- [ ] Detection results reflect only successful server queries
- [ ] Failed server queries are logged with timestamp and reason

## Edge Cases & Constraints

### API Interaction

- Use Radarr/Sonarr API endpoints that filter for missing content server-side
  (don't retrieve all items and filter client-side)
- Handle API pagination if a server has a very large number of missing items
- Respect API rate limits (if Radarr/Sonarr have any)

### Monitoring Status

- Only count items that are marked as "monitored" in Radarr/Sonarr
- Ignore items that users have explicitly unmonitored
- Respect series/season monitoring settings in Sonarr (don't count episodes from
  unmonitored seasons)

### Performance

- Detection target: < 30 seconds per server for libraries up to 10,000 items
- Large libraries (> 10,000 items): Allow proportionally longer times (roughly
  linear scaling)
- Consider caching server responses briefly if detection runs very frequently
  (optional optimization)

### Data Accuracy

- Missing count should reflect the state at the time of detection
- If an item is found between detection and search triggering, Radarr/Sonarr
  will handle the duplicate gracefully (they won't re-download)

### Known Limitations

- The system does not track which specific episodes/movies are missing, only the
  total count
- Missing items from failed server queries are not retried until the next
  scheduled automation cycle
