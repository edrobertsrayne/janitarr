# Janitarr

Keeps a Sonarr and a Radarr instance tidy over their web API. One Node service, configured by environment variables, with a web UI for the dashboard and for tweaking anything you did not pin in the environment.

## What it does

| Job         | Default interval | Behaviour                                                                                                                                                                                                                                                                                                  |
| ----------- | ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Search**  | 60 min           | Triggers searches for monitored items that are missing and/or have not met their quality cutoff. Capped per cycle, with a per-item cooldown so indexers are not hammered.                                                                                                                                  |
| **Stalled** | 5 min            | Removes downloads that made no meaningful progress for a timeout (tiny trickles below a minimum-progress threshold don't count), queued items stuck at size 0, or that the \*arr itself flags as stalled. Optionally removes from the download client, blocklists the release, and searches again.         |
| **Failed**  | 5 min            | Clears failed downloads/imports from the queue, blocklists the release, removes it from the client and searches again. Each failure is a **strike** against the episode/movie; at the limit the item is no longer searched and is (optionally) unmonitored. Strikes decay after a period without failures. |

Details worth knowing:

- Stalled means: status `downloading`/`warning`, and either `sizeleft` unchanged for `STALLED_TIMEOUT_MINUTES` (tracked by Janitarr) or a stalled warning from the \*arr. Neither applies during the grace period after a download first appears. Paused/queued downloads are never touched.
- Janitarr deletes queue items with `skipRedownload=true` and triggers the re-search itself, so items that have hit the strike limit are not searched again.
- Stalled removals do not add strikes; failed downloads/imports do.
- Items that are given up on stay skipped until you press **Clear** on the dashboard's Strikes panel (do that after re-monitoring them).
- A multi-episode download (season pack) is removed once, and every episode in it gets a strike.
- **Dry run** logs what would happen and changes nothing.

## Run

```sh
npm ci
npm run build   # without a .git directory (e.g. a Nix sandbox): GIT_SHA=<rev> npm run build
node build
```

UI at `http://localhost:3000` (`HOST`/`PORT` set the listen address). State (settings saved from the UI, strikes, cooldowns, activity log) lives in `JANITARR_DATA_DIR` (default `./data`).

There is no authentication; run it behind a reverse proxy or on a private network. When proxied, set `ORIGIN` to the public URL (SvelteKit's CSRF protection compares it with the request's `Origin` header), or `PROTOCOL_HEADER`/`HOST_HEADER` as described in the [adapter-node docs](https://svelte.dev/docs/kit/adapter-node).

## Configuration

Resolution per setting: **environment variable → value saved from the UI → default.** A setting that is set in the environment is locked (read-only) in the UI and shows its variable name.

| Variable                                | Default                         |                                                                                                  |
| --------------------------------------- | ------------------------------- | ------------------------------------------------------------------------------------------------ |
| `SONARR_URL`, `SONARR_API_KEY`          | –                               | Connection. The instance is skipped until both are set.                                          |
| `RADARR_URL`, `RADARR_API_KEY`          | –                               | Connection.                                                                                      |
| `JANITARR_DRY_RUN`                      | `false`                         | Log only.                                                                                        |
| `JANITARR_SEARCH_ENABLED`               | `true`                          |                                                                                                  |
| `JANITARR_SEARCH_INTERVAL_MINUTES`      | `60`                            |                                                                                                  |
| `JANITARR_SEARCH_MISSING`               | `true`                          | Search items with no file.                                                                       |
| `JANITARR_SEARCH_CUTOFF`                | `true`                          | Search items below the quality cutoff.                                                           |
| `JANITARR_SEARCH_MAX_PER_CYCLE`         | `10`                            | Per instance and per list (missing, cutoff).                                                     |
| `JANITARR_SEARCH_COOLDOWN_HOURS`        | `72`                            | Minimum time before searching the same item again.                                               |
| `JANITARR_STALLED_ENABLED`              | `true`                          |                                                                                                  |
| `JANITARR_STALLED_INTERVAL_MINUTES`     | `5`                             |                                                                                                  |
| `JANITARR_STALLED_TIMEOUT_MINUTES`      | `30`                            | No-progress timeout. Also catches queued items stuck at size 0 (no metadata).                    |
| `JANITARR_STALLED_MIN_PROGRESS_PERCENT` | `1`                             | Progress restarts the timeout only after this % of the size (at most 50 MB); `0` = any progress. |
| `JANITARR_STALLED_GRACE_MINUTES`        | `10`                            | Minimum age before a download can be called stalled.                                             |
| `JANITARR_STALLED_REMOVE_FROM_CLIENT`   | `true`                          |                                                                                                  |
| `JANITARR_STALLED_BLOCKLIST`            | `true`                          |                                                                                                  |
| `JANITARR_STALLED_RESEARCH`             | `true`                          |                                                                                                  |
| `JANITARR_FAILED_ENABLED`               | `true`                          |                                                                                                  |
| `JANITARR_FAILED_INTERVAL_MINUTES`      | `5`                             |                                                                                                  |
| `JANITARR_FAILED_IMPORT_BLOCKED`        | `true`                          | Treat "import blocked" as failed.                                                                |
| `JANITARR_FAILED_STRIKES`               | `3`                             | Failures before giving up on an item.                                                            |
| `JANITARR_FAILED_STRIKE_DECAY_DAYS`     | `30`                            | `0` disables decay.                                                                              |
| `JANITARR_FAILED_UNMONITOR`             | `true`                          | Unmonitor at the strike limit.                                                                   |
| `JANITARR_DATA_DIR`                     | `./data`                        |                                                                                                  |

Booleans accept `true/false/1/0/yes/no/on/off`. An invalid value is ignored, falls back to the UI/default value, and is written to the activity log.

## Development

```sh
npm install
npm run dev        # http://localhost:5173, data in ./data
npm run check && npm run lint
npx vitest run --project server
```

Requires Node 24 (uses the built-in `node:sqlite`). Targets the Sonarr/Radarr v3 API (`/api/v3`).
