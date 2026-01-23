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

- When triggering searches across multiple servers, distribute the search limit
  proportionally based on each server's share of total items in that category.
  - Example: If Server A has 90 missing items and Server B has 10 (total 100),
    and the limit is 10, Server A receives 9 searches and Server B receives 1.
  - Minimum allocation: Each server with items in the category receives at
    least 1 search, even if its proportion would yield less than 1.
  - If minimum allocations exceed the limit, reduce each server's allocation
    proportionally while maintaining at least 1 per server.
  - Rounding: Use floor division for proportional allocation, then distribute
    any remaining searches to servers with the largest fractional remainders.
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

- Between batch search commands, wait 100ms minimum to avoid overwhelming
  servers
- If a server returns HTTP 429 (rate limited):
  - Honor the `Retry-After` header if present
  - If no `Retry-After` header, wait 30 seconds before retrying
  - Log rate limit events at WARN level
- After 3 consecutive rate limit responses from the same server, skip remaining
  searches for that server in the current cycle and log at ERROR level

### User Control

- Limits prevent runaway search behavior (accidentally searching thousands of
  items)
- Search limits accept values from 0 to 1000
- Values above 100 display a warning about potential indexer strain (in both
  web UI and CLI)
- No hard maximum enforced beyond validation—users control their own risk

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
