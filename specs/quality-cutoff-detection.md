# Quality Cutoff Detection: Identifying Upgradeable Media

## Context

The system must identify episodes and movies that exist in the library but have
not yet met the quality profile cutoff defined in Radarr/Sonarr. This is the
second detection mechanism (alongside missing content detection) that feeds into
the search triggering system.

Quality cutoff represents opportunities to upgrade existing content to better
quality. The detection runs against all configured servers during each
automation cycle.

## Requirements

### Story: Detect Episodes Below Quality Cutoff

- **As a** user
- **I want to** the system to identify TV episodes that haven't met the
  configured quality cutoff
- **So that** searches can be triggered to upgrade them to preferred quality

#### Acceptance Criteria

- [ ] System queries each configured Sonarr server for monitored episodes that
      exist but are below quality cutoff
- [ ] System counts total upgradeable episodes across all Sonarr servers
- [ ] Cutoff-not-met count is available to the search triggering system
- [ ] Detection runs against all configured Sonarr servers in each automation
      cycle

### Story: Detect Movies Below Quality Cutoff

- **As a** user
- **I want to** the system to identify movies that haven't met the configured
  quality cutoff
- **So that** searches can be triggered to upgrade them to preferred quality

#### Acceptance Criteria

- [ ] System queries each configured Radarr server for monitored movies that
      exist but are below quality cutoff
- [ ] System counts total upgradeable movies across all Radarr servers
- [ ] Cutoff-not-met count is available to the search triggering system
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

- Use Radarr/Sonarr API endpoints that filter for cutoff-not-met content
  server-side
- Handle API pagination if a server has many items below cutoff
- Quality cutoff is defined per item in Radarr/Sonarr's quality profiles - the
  system reads this, doesn't define it

### Quality Profile Behavior

- Respect the quality cutoff configured in each server's quality profiles
- Only count items where current quality is below the cutoff threshold
- Ignore items that have already met or exceeded their quality cutoff

### Monitoring Status

- Only count items that are marked as "monitored" in Radarr/Sonarr
- Respect series/season monitoring settings in Sonarr

### Performance

- Detection target: < 30 seconds per server for libraries up to 10,000 items
- Large libraries (> 10,000 items): Allow proportionally longer times (roughly
  linear scaling)
- Cutoff detection may be slower than missing detection as it requires quality
  comparison

### Data Accuracy

- Cutoff-not-met count should reflect the state at the time of detection
- If an item is upgraded between detection and search triggering, Radarr/Sonarr
  will handle gracefully

### Known Limitations

- The system does not track which specific episodes/movies need upgrades, only
  the total count
- Items from failed server queries are not retried until the next scheduled
  automation cycle
- System relies on Radarr/Sonarr's quality profile configuration - it does not
  define or validate quality standards
