# Automation Scheduling: Configuring When Detection and Search Occurs

## Context

The system runs detection and search operations on a schedule defined by the
user. The schedule determines how often the complete automation cycle executes:
detect missing content, detect cutoff-not-met content, and trigger searches
according to configured limits.

This is the timing control for the entire automation system.

## Requirements

### Story: Configure Schedule Frequency

- **As a** user
- **I want to** define how often the automation runs
- **So that** I can balance responsiveness against system load and indexer usage

#### Acceptance Criteria

- [ ] User can set a time interval for how often automation runs (e.g., every 1
      hour, every 6 hours, daily)
- [ ] Minimum interval is 1 hour to prevent excessive API usage
- [ ] Schedule configuration is persisted and survives application restarts
- [ ] Changes to schedule take effect on the next scheduled run (not
      retroactively)

### Story: Execute Automation Cycle

- **As a** user
- **I want to** the system to automatically run detection and search triggering
  on schedule
- **So that** my libraries are maintained without manual intervention

#### Acceptance Criteria

- [ ] System executes complete automation cycle at the configured interval
- [ ] Each cycle includes: detect missing content, detect cutoff-not-met
      content, trigger searches up to limits
- [ ] Automation continues running indefinitely until stopped by user or
      application shutdown
- [ ] Each automation cycle completion is logged with timestamp

### Story: Manual Trigger

- **As a** user
- **I want to** manually trigger an automation cycle on demand
- **So that** I can test configuration or run immediately without waiting for
  the next scheduled time

#### Acceptance Criteria

- [ ] User can manually initiate an automation cycle through the interface
- [ ] Manual trigger executes the same complete cycle as scheduled automation
- [ ] Manual trigger does not affect the regular schedule (next scheduled run
      occurs at the original time)
- [ ] If a cycle is already running, one manual trigger may be queued; additional
      manual triggers while one is queued are rejected with an appropriate message
- [ ] The queued trigger executes immediately after the current cycle completes
- [ ] User receives feedback when manual cycle completes

### Story: Preview Mode (Dry-Run)

- **As a** user
- **I want to** preview what would be searched without actually triggering
  searches
- **So that** I can validate my configuration and limits before running actual
  automation

> **Note:** Dry-run functionality is fully specified in
> [`search-triggering.md`](./search-triggering.md#dry-run-mode). This story
> enables dry-run mode to be triggered as part of the automation cycle via the
> `--dry-run` CLI flag.

#### Acceptance Criteria

- [ ] User can run automation in dry-run/preview mode via CLI flag (e.g.,
      `--dry-run`)
- [ ] Dry-run mode executes the full automation cycle with dry-run behavior as
      defined in `search-triggering.md`

### Story: View Schedule Status

- **As a** user
- **I want to** see when the next automation cycle will run
- **So that** I know the system is operating as configured

#### Acceptance Criteria

- [ ] Interface displays current schedule configuration (interval)
- [ ] Interface shows time until next scheduled run
- [ ] Status updates in real-time or when user refreshes

## Edge Cases & Constraints

### Scheduling Behavior

- If an automation cycle takes longer than the configured interval, the next
  cycle should wait until the current one completes (don't run concurrent
  cycles)
- If the application is stopped and restarted, the next automation cycle waits
  for the next scheduled time (no immediate run on startup). Users can manually
  trigger if immediate execution is desired.

### Error Handling

- If an automation cycle fails completely (e.g., all servers unreachable), log
  the failure and continue with the next scheduled cycle
- Partial failures (some servers work, some don't) are handled within
  detection/triggering specs - automation continues

### Resource Management

- Long-running automation cycles should not block the user interface
- System should remain responsive during detection and search operations

### Time Zones

- All scheduling and logging should use consistent time zone (user's local time
  or UTC, clearly indicated)
- Schedule intervals are duration-based, not time-of-day based (e.g., "every 6
  hours" not "at 6am and 6pm")

### Performance

- Complete automation cycle target: < 5 minutes total for typical libraries (up
  to 10,000 items per server)
- Large libraries (> 10,000 items): Allow proportionally longer cycle times
- If cycle duration approaches interval duration, consider warning user to
  increase interval

### User Control

- User should be able to disable scheduled automation entirely (set to "manual
  only" mode)
- No upper limit on schedule interval (user can set very infrequent schedules if
  desired)
