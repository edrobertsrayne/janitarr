# Search Triggering: Initiating Content Searches with User-Defined Limits

## Context

After the system detects missing content and content below quality cutoff, it
must trigger searches in Radarr and Sonarr. The user defines how many searches
to trigger for each category (missing vs cutoff-not-met) to control resource
usage and avoid overwhelming indexers.

Search triggering is the action phase of the automation - this is where the
system actually tells Radarr/Sonarr to search for content.

## Requirements

### Story: Configure Search Limits

- **As a** user
- **I want to** define how many missing items and how many cutoff-not-met items
  to search for per automation run
- **So that** I can control the volume of searches and avoid overwhelming my
  indexers

#### Acceptance Criteria

- [ ] User can set a numeric limit for missing movies searches (0 or greater)
- [ ] User can set a numeric limit for missing episodes searches (0 or greater)
- [ ] User can set a separate numeric limit for cutoff-not-met movies searches
      (0 or greater)
- [ ] User can set a separate numeric limit for cutoff-not-met episodes searches
      (0 or greater)
- [ ] Limits apply globally across all configured servers but separately per
      content type (movies vs episodes)
- [ ] Setting a limit to 0 disables that category of searches
- [ ] Limits are persisted and apply to all future automation runs until changed

### Story: Trigger Missing Content Searches

- **As a** user
- **I want to** the system to trigger searches for missing items up to my
  configured limit
- **So that** gaps in my library are filled automatically

#### Acceptance Criteria

- [ ] System triggers searches for up to N missing items, where N is the
      user-configured limit
- [ ] If total missing items exceeds limit, only the configured number of
      searches are triggered
- [ ] If total missing items is less than limit, only available items are
      searched
- [ ] Searches are distributed across all configured servers (both Radarr and
      Sonarr)
- [ ] Each triggered search is logged with timestamp, server, and item type

### Story: Trigger Quality Upgrade Searches

- **As a** user
- **I want to** the system to trigger searches for cutoff-not-met items up to my
  configured limit
- **So that** my library quality improves automatically over time

#### Acceptance Criteria

- [ ] System triggers searches for up to N cutoff-not-met items, where N is the
      user-configured limit
- [ ] If total cutoff-not-met items exceeds limit, only the configured number of
      searches are triggered
- [ ] If total cutoff-not-met items is less than limit, only available items are
      searched
- [ ] Searches are distributed across all configured servers (both Radarr and
      Sonarr)
- [ ] Each triggered search is logged with timestamp, server, and item type

### Story: Handle Search Failures

- **As a** user
- **I want to** the system to log failures when searches cannot be triggered
- **So that** I can troubleshoot issues without searches failing silently

#### Acceptance Criteria

- [ ] If a search command fails (server unreachable, API error), the failure is
      logged with reason
- [ ] Failed searches do not count against the user-configured limit
- [ ] System continues attempting to trigger remaining searches even if some
      fail

## Edge Cases & Constraints

### Search Distribution

- When triggering searches across multiple servers, distribute fairly based on
  each server's proportion of total items (don't exhaust one server's quota
  before touching others)
- Limits are separate by category AND content type:
  - Missing movies limit is independent from missing episodes limit
  - Cutoff-not-met movies limit is independent from cutoff-not-met episodes limit
  - Example: With missing movies limit of 10 and missing episodes limit of 10,
    the system can trigger up to 20 total searches (10 + 10)

### API Commands

- Use Radarr/Sonarr's "CommandController" API to trigger searches (e.g.,
  "MoviesSearch", "SeriesSearch")
- Use batch commands where possible (send arrays of IDs in a single request for
  efficiency)
- Triggered searches are queued in Radarr/Sonarr - actual search execution is
  their responsibility
- The system only triggers searches; it doesn't track whether searches succeed
  or find content

### Item Selection

- The system does not need to intelligently prioritize which specific items to
  search when limits are exceeded
- Radarr/Sonarr will handle the actual search prioritization based on their own
  internal logic
- Acceptable to search the first N items returned by detection queries

### Rate Limiting

- Respect Radarr/Sonarr API rate limits when triggering searches
- If triggering many searches rapidly, implement brief delays between commands
  if necessary

### User Control

- Limits prevent runaway search behavior (accidentally searching thousands of
  items)
- User should be able to set high limits if they want aggressive searching
- No arbitrary maximum limit (user controls their own risk)

### Dry-Run Mode

- Users can preview what would be searched without actually triggering searches
- Dry-run mode is useful for:
  - Testing configuration changes before applying them
  - Understanding what the automation will do before enabling it
  - Validating search limits are set appropriately
  - Previewing which items would be searched in the next cycle
- Dry-run execution:
  - Performs full detection (queries Radarr/Sonarr for missing and cutoff items)
  - Applies configured limits and distribution logic
  - Logs or displays what _would_ be searched
  - Does NOT trigger actual searches in Radarr/Sonarr
  - Does NOT create log entries for searches (since none occurred)
- Available via CLI: `janitarr run --dry-run` or `janitarr scan` command
- Should clearly indicate in output that this is a preview/dry-run

### Known Behavior

- Setting all limits to 0 effectively disables all search automation
- Searches are triggered "fire and forget" - the system does not wait for
  Radarr/Sonarr to complete searches
- If detection finds 100 missing items but limit is 5, the same 5 items may be
  searched repeatedly in subsequent runs until they're found
- Janitarr does not perform search deduplication - it relies on Radarr/Sonarr to
  handle duplicate search requests intelligently and not re-download existing
  content
