# Specification Audit Report: Janitarr

**Date:** 2026-01-23
**Scope:** Review of all specification files in `specs/` directory

---

## 1. Executive Summary

The Janitarr specification suite is comprehensive and well-organized, covering the core automation functionality, web interface, CLI, and operational concerns. However, the review identified several areas requiring attention:

**Critical Issues:**

- Port number inconsistency (3000 vs 3434) in web-frontend.md
- Search limits terminology mismatch (2 vs 4 limits between specs)
- Log retention range inconsistency (1-365 vs 7-90 days)
- Encryption key derivation contradiction (machine ID vs stored file)

**High Priority:**

- Significant overlap between `logging.md` and `activity-logging.md`
- Dry-run mode duplicated in two specs
- Several vague performance requirements need concrete metrics

**Recommendations:**

1. Merge `logging.md` and `activity-logging.md` into a single spec
2. Resolve the web-frontend.md port and limit inconsistencies
3. Archive `daisyui-migration.md` after implementation (one-time migration)
4. Add concrete performance metrics to replace "reasonable time" language

---

## 2. Overlapping/Inconsistent Requirements

### 2.1 Web Server Port Inconsistency (Critical)

**Files:** `web-frontend.md`, `unified-service-startup.md`

**Issue:** `web-frontend.md` contains conflicting port references:

- Line 26: "Port: Configurable (default: 3434)"
- Line 365: "Base URL: `http://localhost:3000/api`"
- Line 840: "JANITARR_WEB_PORT=3000"

Meanwhile, `unified-service-startup.md` consistently uses port 3434.

**Recommendation:** Update all port references in `web-frontend.md` to 3434 for consistency.

---

### 2.2 Search Limits Terminology Mismatch (Critical)

**Files:** `search-triggering.md`, `cli-interface.md`, `web-frontend.md`

**Issue:**

- `search-triggering.md` (lines 25-30) defines 4 separate limits: missing movies, missing episodes, cutoff movies, cutoff episodes
- `cli-interface.md` (lines 98-99) mirrors this with 4 limits
- `web-frontend.md` (lines 299-312) defines only 2 limits: "Missing Content Limit" and "Quality Cutoff Limit"

**Impact:** Implementation uncertainty—does the web UI combine movies/episodes or expose all 4?

**Recommendation:** Decide on 2 or 4 limits and update all specs consistently. The 4-limit approach provides more control; the web UI should expose all 4.

---

### 2.3 Log Retention Range Inconsistency (High)

**Files:** `logging.md`, `activity-logging.md`, `web-frontend.md`

**Issue:**

- `logging.md` (line 194): "Retention period configurable via settings (7-90 days range)"
- `activity-logging.md` (line 155): "configurable 7-90 days"
- `web-frontend.md` (line 333): "Min: 1, Max: 365"

**Recommendation:** Standardize to 7-90 days across all specs (more reasonable default range) or explicitly decide on 1-365 if broader range is desired.

---

### 2.4 Dry-Run Mode Duplication (Medium)

**Files:** `search-triggering.md`, `automatic-scheduling.md`

**Issue:** Dry-run mode is fully specified in both:

- `search-triggering.md` lines 133-149
- `automatic-scheduling.md` lines 62-83

Both describe the same feature with slightly different wording.

**Recommendation:** Define dry-run in `search-triggering.md` (the action spec) and reference it from `automatic-scheduling.md`. Remove duplicate acceptance criteria.

---

### 2.5 Logging Specs Overlap (Medium)

**Files:** `logging.md`, `activity-logging.md`

**Issue:** `activity-logging.md` acknowledges it's "part of the unified logging system" but still duplicates:

- Log format examples
- Retention requirements
- Web interface viewing requirements

**Recommendation:** Merge into a single `logging.md` with clear sections for: (1) system logging, (2) activity/audit logging, (3) web interface, (4) retention.

---

### 2.6 Encryption Key Source Contradiction (High)

**Files:** `server-configuration.md`, `go-architecture.md`

**Issue:**

- `server-configuration.md` (lines 101-103): "Encryption key is derived from machine ID (using crypto.randomUUID() or similar)"
- `go-architecture.md` (line 625): "Encryption key stored in `data/.janitarr.key`" and "Key is 32 bytes, base64 encoded for storage"

These are contradictory—one suggests derivation, the other suggests stored file.

**Recommendation:** Clarify the implementation approach. The stored key file approach (from go-architecture.md) appears correct based on actual implementation. Update `server-configuration.md` to match.

---

### 2.7 Connection Test Timeout (Low)

**Files:** `server-configuration.md`, `cli-interface.md`

**Issue:**

- `server-configuration.md` (line 78): "suggest 10-15 seconds"
- `cli-interface.md` (line 259): "Timeout for connection tests: 10 seconds"

**Recommendation:** Standardize to 10 seconds across both specs.

---

## 3. Ambiguous/Unclear Requirements

### 3.1 "Reasonable Time" Performance Metrics

**Files:** `missing-content-detection.md`, `quality-cutoff-detection.md`, `automatic-scheduling.md`

**Issue:** Multiple specs use vague language:

- "Detection should complete in reasonable time" (`missing-content-detection.md:79`)
- "Detection should complete in reasonable time" (`quality-cutoff-detection.md:84`)
- "suggest < 5 minutes for typical libraries" (`automatic-scheduling.md:128`)

**Proposed Fix:** Replace with concrete metrics:

- Detection phase: < 30 seconds per server for libraries up to 10,000 items
- Complete automation cycle: < 5 minutes total
- Define "large library" as > 10,000 items with proportionally longer acceptable times

---

### 3.2 Search Distribution Logic

**File:** `search-triggering.md` (lines 93-99)

**Issue:** "distribute fairly based on each server's proportion of total items" is ambiguous.

**Current text:**

> When triggering searches across multiple servers, distribute fairly based on each server's proportion of total items

**Questions:**

- If Server A has 90 missing items and Server B has 10, does a limit of 10 mean Server A gets 9 and Server B gets 1?
- What about rounding? What if proportional calculation yields 0 for a server?

**Proposed Fix:**

> Distribute search limit proportionally: if Server A has 90% of missing items and limit is 10, Server A receives 9 searches and Server B receives 1. Minimum 1 search per server with missing items, even if proportion is < 1/limit. If total exceeds limit after minimum allocation, reduce proportionally.

---

### 3.3 Manual Trigger Queuing Behavior

**File:** `automatic-scheduling.md` (line 59)

**Issue:** "If a cycle is already running, manual triggers are queued" doesn't specify queue depth.

**Questions:**

- What if 5 manual triggers are submitted while a cycle runs?
- Are they all queued, or only 1?

**Proposed Fix:**

> If a cycle is already running, one manual trigger may be queued. Additional manual triggers while one is queued are rejected with an appropriate message. The queued trigger executes immediately after the current cycle completes.

---

### 3.4 WebSocket Reconnection After Max Backoff

**File:** `web-frontend.md` (line 535)

**Issue:** Exponential backoff maxes at 30 seconds, but what happens next?

**Proposed Fix:**

> Auto-reconnect on disconnect with exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s. After reaching 30s, continue attempting every 30 seconds indefinitely until connection succeeds or page is closed.

---

### 3.5 API Key Length Validation

**File:** `cli-interface.md` (line 238)

**Issue:** "API key validation: non-empty, reasonable length (20-100 chars)"

Radarr/Sonarr API keys are 32 hex characters. The range 20-100 seems arbitrary.

**Proposed Fix:**

> API key validation: non-empty, 32 hexadecimal characters (matching Radarr/Sonarr API key format)

Or if flexibility is needed:

> API key validation: non-empty, 20-64 characters

---

### 3.6 Undefined "High Limits"

**File:** `search-triggering.md` (line 130)

**Issue:** "User should be able to set high limits if they want aggressive searching" without defining constraints.

**Proposed Fix:**

> Search limits accept values from 0 to 1000. Values above 100 display a warning about potential indexer strain. No hard maximum enforced—users control their own risk.

---

### 3.7 Rate Limiting Implementation

**File:** `search-triggering.md` (lines 120-123)

**Issue:** "Respect Radarr/Sonarr API rate limits" and "implement brief delays" are undefined.

**Proposed Fix:**

> Between batch search commands, wait 100ms minimum. If a server returns HTTP 429 (rate limited), honor the `Retry-After` header if present, otherwise wait 30 seconds before retrying. Log rate limit events at WARN level.

---

### 3.8 Configuration Precedence

**Files:** `go-architecture.md`, `web-frontend.md`

**Issue:** Multiple config sources mentioned (env vars, CLI flags, database) without clear precedence.

**Proposed Fix:** Add to `go-architecture.md`:

> Configuration precedence (highest to lowest):
>
> 1. CLI flags (`--port`, `--host`, `--db-path`, `--log-level`)
> 2. Environment variables (`JANITARR_*`)
> 3. Database-stored configuration
> 4. Hardcoded defaults

---

## 4. Feature Suggestions

### 4.1 Notifications/Webhooks (High Value)

**Rationale:** Users managing media automation typically want alerts for errors or completed cycles. Discord, Slack, and email notifications are common in the Radarr/Sonarr ecosystem.

**Suggested Scope:**

- Webhook support for cycle completion and errors
- Configurable notification triggers (errors only, all cycles, etc.)
- Discord webhook format support (native to the ecosystem)

---

### 4.2 Per-Server Search Limits (Medium Value)

**Rationale:** Users with multiple servers (e.g., 4K + standard quality Radarr instances) may want different search behaviors per server.

**Suggested Scope:**

- Optional per-server override of global limits
- Default: inherit global limits
- Web UI: expandable "Advanced" section in server edit form

---

### 4.3 Search History Deduplication (Medium Value)

**Rationale:** Specs acknowledge "the same 5 items may be searched repeatedly." This could overwhelm indexers with duplicate searches.

**Suggested Scope:**

- Track last search timestamp per item (movie ID / episode ID)
- Skip items searched within configurable cooldown period (default: 24 hours)
- Option to disable for users who want aggressive re-searching

---

### 4.4 Import/Export Configuration (Medium Value)

**Rationale:** No backup/restore capability mentioned. Important for disaster recovery and migration.

**Suggested Scope:**

- Export: servers (without decrypted API keys), automation settings
- Import: validate and add/update servers with new API keys required
- CLI: `janitarr config export/import`

---

### 4.5 Manual Search for Specific Item (Low Value for v1)

**Rationale:** Current automation is batch-only. Users might want to trigger a specific movie/episode.

**Suggested Scope:** Consider for post-v1 as the web UI matures.

---

### 4.6 Server Priority/Ordering (Low Value)

**Rationale:** When limits constrain searches, users might want certain servers prioritized.

**Suggested Scope:** Consider adding a priority field (1-10) to server configuration, higher priority servers get searches first.

---

## 5. Documentation Improvements

### 5.1 Merge Logging Specifications

**Action:** Merge `logging.md` and `activity-logging.md` into a single comprehensive `logging.md`.

**Structure:**

1. Overview and Goals
2. Technology Stack
3. Log Levels and Configuration
4. Console Logging (charmbracelet/log)
5. Activity Logging (audit trail events)
6. Web Interface Log Viewer
7. WebSocket Streaming
8. Database Storage and Retention
9. Implementation Notes

---

### 5.2 Archive DaisyUI Migration Spec

**Action:** After implementation is complete, move `daisyui-migration.md` to an `archive/` directory or rename to `_completed_daisyui-migration.md`.

**Rationale:** This is a one-time migration guide, not an ongoing specification. Keeping it clutters the active spec directory.

---

### 5.3 Update README.md Tables

**Action:** Add `daisyui-migration.md` to the README tables and add a "Status" column.

```markdown
| Spec                 | Code | Purpose                     | Status    |
| -------------------- | ---- | --------------------------- | --------- |
| web-frontend.md      | ...  | ...                         | Active    |
| daisyui-migration.md | ...  | DaisyUI component migration | Completed |
```

---

### 5.4 Resolve Open Questions

**File:** `web-frontend.md` (lines 959-965)

**Action:** Address each open question:

1. "Multiple instances with shared database" → Out of scope for v1, document as unsupported
2. "Configuration wizard on first launch" → Document as post-v1 enhancement
3. "Get Started tutorial" → Post-v1 enhancement
4. "Custom themes beyond light/dark" → Already addressed in daisyui-migration.md (light/dark only)
5. "Logs in separate table" → Already implemented per schema in logging.md

---

### 5.5 Add Sequence Diagrams

**Action:** Add diagrams for complex flows in `go-architecture.md`:

- Automation cycle flow (scheduler → detector → search trigger → logger)
- Web request flow (HTTP → handler → service → database)
- WebSocket log streaming flow

---

### 5.6 Standardize Acceptance Criteria Format

**Action:** All specs should use checkbox format `- [ ]` for testable acceptance criteria.

Files needing updates: `missing-content-detection.md`, `quality-cutoff-detection.md` (they already use this format, but verify consistency).

---

### 5.7 Add API Error Response Specification

**Action:** Add a section to `web-frontend.md` or create new `api-conventions.md`:

```json
{
  "error": {
    "code": "SERVER_NAME_EXISTS",
    "message": "Server name already exists",
    "field": "name"
  }
}
```

Standard error codes, HTTP status mapping, and field-level validation response format.

---

## 6. Recommended Actions

### Priority 1: Critical Fixes (Before Implementation)

| Action                                  | Files                     | Scope                 |
| --------------------------------------- | ------------------------- | --------------------- |
| Fix port references to 3434             | `web-frontend.md`         | ~5 line changes       |
| Align search limits (4 separate limits) | `web-frontend.md`         | Rewrite Section 2     |
| Fix log retention range (7-90 days)     | `web-frontend.md`         | 1 line change         |
| Clarify encryption key storage          | `server-configuration.md` | Rewrite lines 101-104 |

### Priority 2: Consolidation (Immediate)

| Action                     | Files                               | Scope                          |
| -------------------------- | ----------------------------------- | ------------------------------ |
| Merge logging specs        | `logging.md`, `activity-logging.md` | Medium restructure             |
| Remove dry-run duplication | `automatic-scheduling.md`           | Reference search-triggering.md |

### Priority 3: Clarification (Short-term)

| Action                               | Files                                                                                    | Scope           |
| ------------------------------------ | ---------------------------------------------------------------------------------------- | --------------- |
| Add concrete performance metrics     | `missing-content-detection.md`, `quality-cutoff-detection.md`, `automatic-scheduling.md` | Small additions |
| Define search distribution algorithm | `search-triggering.md`                                                                   | Add 1 paragraph |
| Define manual trigger queue behavior | `automatic-scheduling.md`                                                                | Add 1 sentence  |
| Specify rate limiting behavior       | `search-triggering.md`                                                                   | Add 1 paragraph |
| Fix API key length validation        | `cli-interface.md`                                                                       | 1 line change   |

### Priority 4: Enhancements (Post-v1 Consideration)

| Action                 | New Spec                      | Rationale             |
| ---------------------- | ----------------------------- | --------------------- |
| Notifications/webhooks | `notifications.md`            | High user value       |
| Per-server limits      | Update `search-triggering.md` | Flexibility           |
| Search deduplication   | Update `search-triggering.md` | Reduce indexer strain |
| Config import/export   | `backup-restore.md`           | Disaster recovery     |

### Priority 5: Documentation Hygiene (Ongoing)

| Action                                        | Files                            | Scope             |
| --------------------------------------------- | -------------------------------- | ----------------- |
| Archive daisyui-migration.md after completion | Move to `archive/`               | File move         |
| Update README.md with status column           | `README.md`                      | Table restructure |
| Add sequence diagrams                         | `go-architecture.md`             | Add diagrams      |
| Add API error conventions                     | New section in `web-frontend.md` | Medium addition   |

---

## Appendix: File-by-File Issue Summary

| File                           | Critical         | High                          | Medium             | Low                |
| ------------------------------ | ---------------- | ----------------------------- | ------------------ | ------------------ |
| `web-frontend.md`              | 2 (port, limits) | 1 (retention)                 | 1 (open questions) | 0                  |
| `server-configuration.md`      | 1 (encryption)   | 0                             | 1 (timeout)        | 0                  |
| `search-triggering.md`         | 0                | 2 (distribution, rate limits) | 1 (dry-run dup)    | 1 (high limits)    |
| `automatic-scheduling.md`      | 0                | 1 (queue behavior)            | 1 (dry-run dup)    | 0                  |
| `logging.md`                   | 0                | 0                             | 1 (overlap)        | 0                  |
| `activity-logging.md`          | 0                | 0                             | 1 (overlap)        | 0                  |
| `cli-interface.md`             | 0                | 0                             | 0                  | 1 (API key length) |
| `missing-content-detection.md` | 0                | 0                             | 1 (perf metrics)   | 0                  |
| `quality-cutoff-detection.md`  | 0                | 0                             | 1 (perf metrics)   | 0                  |
| `unified-service-startup.md`   | 0                | 0                             | 0                  | 0                  |
| `go-architecture.md`           | 0                | 0                             | 0                  | 1 (diagrams)       |
| `daisyui-migration.md`         | 0                | 0                             | 1 (should archive) | 0                  |
| `README.md`                    | 0                | 0                             | 0                  | 1 (missing links)  |

---

## Summary

This audit identifies **3 critical issues**, **4 high-priority items**, **9 medium-priority items**, and **4 low-priority items** across 13 specification files. The specifications are generally well-written and comprehensive, with most issues being inconsistencies between files rather than fundamental gaps.

**Total Issues:** 20

- Critical: 3
- High: 4
- Medium: 9
- Low: 4
