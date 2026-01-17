# Janitarr Troubleshooting Guide

Common issues and solutions for Janitarr users.

## Table of Contents

- [Server Connection Issues](#server-connection-issues)
- [Search Issues](#search-issues)
- [Scheduler Issues](#scheduler-issues)
- [Web Interface Issues](#web-interface-issues)
- [Performance Issues](#performance-issues)
- [Database Issues](#database-issues)
- [Getting Help](#getting-help)

---

## Server Connection Issues

### "Connection failed" when testing server

**Symptoms**:
- Server test fails with connection error
- Dashboard shows server as offline (red status)
- Logs show "Failed to connect to server"

**Possible Causes**:

1. **Incorrect URL**
   - Missing protocol (`http://` or `https://`)
   - Wrong port number
   - Typo in hostname or IP address

   **Solution**:
   ```bash
   # Verify URL format
   janitarr server list

   # Common formats:
   # Local: http://localhost:7878
   # Network: http://192.168.1.100:7878
   # Remote: https://radarr.example.com

   # Edit and fix URL
   janitarr server edit "Server Name"
   ```

2. **Invalid API Key**
   - Wrong key copied from server
   - Extra spaces in key
   - Key regenerated on server side

   **Solution**:
   - Get fresh API key from server Settings → General → Security
   - Copy carefully without extra spaces
   - Update in Janitarr:
     ```bash
     janitarr server edit "Server Name"
     ```

3. **Network Issues**
   - Firewall blocking connection
   - Server not running
   - Network timeout

   **Solution**:
   ```bash
   # Test server accessibility manually
   curl -H "X-Api-Key: YOUR_KEY" http://server-url/api/v3/system/status

   # Check if server is running
   # Check firewall rules
   # Try accessing from same machine as Radarr/Sonarr
   ```

4. **SSL/TLS Certificate Issues** (HTTPS URLs)
   - Self-signed certificate
   - Expired certificate
   - Certificate validation failure

   **Solution**:
   - Use `http://` instead if on local network
   - Fix certificate on server side
   - Ensure system trusts certificate authority

### "Timeout" errors

**Symptoms**:
- Server test times out after 10-15 seconds
- Logs show "Request timed out"

**Possible Causes**:

1. **Server is slow to respond**
   - Server under heavy load
   - Large database query
   - Network latency

   **Solution**:
   - Wait for server to finish current operations
   - Check server performance (CPU, memory, disk I/O)
   - Retry test later

2. **Network latency**
   - Remote server with high ping
   - VPN introducing delay
   - Network congestion

   **Solution**:
   ```bash
   # Check latency
   ping server-hostname

   # If high latency:
   # - Use local network address if available
   # - Disable VPN if not needed
   # - Move Janitarr closer to server (network-wise)
   ```

### "Unauthorized" or "401" errors

**Symptoms**:
- Server test returns "Unauthorized"
- HTTP 401 status code in logs

**Cause**: API key is invalid or missing

**Solution**:
1. Get correct API key from server Settings → General → Security
2. Update in Janitarr:
   ```bash
   janitarr server edit "Server Name"
   ```
3. Ensure no extra spaces or characters in key
4. Test connection after updating

### Server appears online but detection finds nothing

**Symptoms**:
- Server test succeeds
- `janitarr scan` shows 0 missing items
- You know content is missing

**Possible Causes**:

1. **Content not monitored**
   - Movies/episodes marked as "Unmonitored" in Radarr/Sonarr
   - Janitarr only searches for monitored content

   **Solution**:
   - Check Radarr/Sonarr: ensure content is monitored
   - Toggle monitoring in Radarr/Sonarr UI
   - Wait for next scan

2. **Already downloaded**
   - Content appears missing to you but server has it
   - Wrong quality profile confusion

   **Solution**:
   - Check Radarr/Sonarr "Missing" page
   - Verify Janitarr is querying correct endpoint
   - Compare counts: Radarr UI vs Janitarr scan

3. **Server not enabled in Janitarr**
   - Server disabled in configuration

   **Solution**:
   ```bash
   # Check server status
   janitarr server list

   # Enable if disabled
   janitarr server edit "Server Name"
   # Set enabled: true
   ```

---

## Search Issues

### Searches not triggering

**Symptoms**:
- `janitarr run` completes but no searches triggered
- Logs show "0 searches triggered"
- Missing content count is high but searches = 0

**Possible Causes**:

1. **Search limits set to 0**
   - All limits disabled

   **Solution**:
   ```bash
   # Check current limits
   janitarr config show

   # Set appropriate limits
   janitarr config set limits.missing.movies 10
   janitarr config set limits.missing.episodes 10
   janitarr config set limits.cutoff.movies 5
   janitarr config set limits.cutoff.episodes 5
   ```

2. **No missing content detected**
   - All content already downloaded
   - Nothing below quality cutoff

   **Solution**:
   ```bash
   # Verify detection is working
   janitarr scan

   # Should show counts > 0 if content is missing
   ```

3. **Servers disabled**
   - All servers have `enabled: false`

   **Solution**:
   ```bash
   janitarr server list
   # Enable servers that should be active
   ```

### Searches fail with errors

**Symptoms**:
- Logs show "Search failed for [item]"
- Error messages in logs
- Searches triggered but fail to execute

**Possible Causes**:

1. **Indexer issues on Radarr/Sonarr**
   - Indexers disabled or down
   - API limits exceeded on indexers
   - Invalid indexer configuration

   **Solution**:
   - Check Radarr/Sonarr System → Tasks → RSS Sync
   - Verify indexers are enabled in Settings → Indexers
   - Test indexers manually in Radarr/Sonarr
   - Review indexer logs for rate limiting

2. **API command rejected**
   - Server rejected search command
   - Invalid movie/episode ID
   - Permissions issue

   **Solution**:
   ```bash
   # Check logs for specific error
   janitarr logs --limit 50

   # Test server connection
   janitarr server test "Server Name"

   # Try manual search in Radarr/Sonarr UI
   ```

3. **Network issues during search**
   - Timeout during API call
   - Connection dropped

   **Solution**:
   - Check network stability
   - Retry automation cycle
   - Monitor logs for patterns

### Too many searches triggered

**Symptoms**:
- Hundreds of searches triggered per cycle
- Indexer sends rate limit warnings
- Logs flooded with search entries

**Possible Causes**:

1. **Limits too high**
   - Search limits configured too aggressively

   **Solution**:
   ```bash
   # Reduce limits immediately
   janitarr config set limits.missing.movies 5
   janitarr config set limits.missing.episodes 5
   janitarr config set limits.cutoff.movies 3
   janitarr config set limits.cutoff.episodes 3

   # Increase interval
   janitarr config set schedule.interval 8
   ```

2. **Multiple Janitarr instances running**
   - Accidentally running multiple schedulers
   - Each triggering searches independently

   **Solution**:
   ```bash
   # Stop all instances
   janitarr stop

   # Verify only one process
   ps aux | grep janitarr

   # Start single instance
   janitarr start
   ```

### Searches not distributed fairly across servers

**Symptoms**:
- One server gets all searches
- Other servers ignored
- Logs show imbalanced distribution

**Possible Causes**:

1. **One server has significantly more missing content**
   - Round-robin still applies but ratio seems off

   **Solution**:
   - This is expected behavior
   - Example: Server A has 100 missing, Server B has 5 missing, limit is 10
   - Result: 5 from A, 5 from B (fair distribution)
   - If you want proportional: increase limits

2. **Only one server has content of that type**
   - Radarr limits only apply to Radarr servers
   - Sonarr limits only apply to Sonarr servers

   **Solution**:
   - Verify you have multiple servers of same type
   - Check server types: `janitarr server list`

---

## Scheduler Issues

### Services not running

**Symptoms**:
- `janitarr status` shows "Scheduler not running"
- Cannot access web interface
- No automated cycles executing
- Next run time not displayed

**Possible Causes**:

1. **Services not started**
   - Never started or was stopped

   **Solution**:
   ```bash
   janitarr start  # Starts both scheduler and web server
   ```

2. **Automation disabled in config**
   - `schedule.enabled` set to false
   - Web server still runs, but scheduler won't execute cycles

   **Solution**:
   ```bash
   # Enable automation
   janitarr config set schedule.enabled true

   # Restart services
   janitarr stop
   janitarr start
   ```

3. **Process died unexpectedly**
   - Crash or system restart
   - No persistence across reboots

   **Solution**:
   - Check system logs for crashes
   - Restart services: `janitarr start`
   - Consider using systemd or supervisor for persistence (see Development Guide)

### Scheduler runs too frequently or infrequently

**Symptoms**:
- Automation cycles run at wrong intervals
- Not respecting configured schedule

**Possible Causes**:

1. **Incorrect interval configuration**
   - Wrong value set

   **Solution**:
   ```bash
   # Check current interval
   janitarr config show

   # Set correct interval (in hours)
   janitarr config set schedule.interval 6

   # Restart scheduler to apply
   janitarr stop
   janitarr start
   ```

2. **Multiple instances running**
   - Each instance runs its own schedule

   **Solution**:
   ```bash
   # Stop all instances
   janitarr stop

   # Verify
   ps aux | grep janitarr

   # Start single instance
   janitarr start
   ```

### Next run time is incorrect

**Symptoms**:
- `janitarr status` shows wrong next run time
- Schedule seems off

**Possible Causes**:

1. **System clock incorrect**
   - Server time zone wrong
   - Clock drift

   **Solution**:
   ```bash
   # Check system time
   date

   # Sync time if needed (Linux)
   sudo timedatectl set-ntp true

   # Restart scheduler
   janitarr stop
   janitarr start
   ```

2. **Scheduler state corrupted**
   - Rare edge case

   **Solution**:
   ```bash
   # Restart scheduler
   janitarr stop
   janitarr start

   # Verify next run time is correct
   janitarr status
   ```

---

## Web Interface Issues

### Cannot access web interface

**Symptoms**:
- Browser shows "Connection refused"
- Cannot reach http://localhost:3434

**Possible Causes**:

1. **Server not running**
   - Janitarr services not started

   **Solution**:
   ```bash
   # Start services (production mode)
   janitarr start

   # Or in development mode
   janitarr dev
   ```

2. **Wrong port**
   - Server running on different port (default changed from 3000 to 3434)
   - Port 3434 occupied by another process

   **Solution**:
   ```bash
   # Check what's on port 3434
   lsof -i :3434

   # Or use different port
   janitarr start --port 8080
   ```

3. **Firewall blocking**
   - Local firewall blocking connection

   **Solution**:
   - Check firewall rules
   - Temporarily disable firewall to test
   - Add exception for port 3434

4. **Accessing from remote machine**
   - Server bound to localhost only (default)

   **Solution**:
   ```bash
   # Bind to all interfaces for remote access
   janitarr start --host 0.0.0.0 --port 3434

   # Then access from remote machine at http://server-ip:3434
   ```

### Web interface loads but API requests fail

**Symptoms**:
- UI loads but shows errors
- "Failed to fetch" errors in browser console
- Empty data in Dashboard/Servers/Logs

**Possible Causes**:

1. **Backend not running**
   - Frontend loaded but backend server down

   **Solution**:
   ```bash
   # Ensure backend is running
   janitarr start
   ```

2. **CORS issues** (development mode)
   - Browser blocking cross-origin requests

   **Solution**:
   ```bash
   # Use dev mode which includes Vite proxy
   janitarr dev  # Terminal 1
   cd ui && bun run dev  # Terminal 2

   # Access at http://localhost:3434 (not http://localhost:5173)
   ```
   - Check browser console for CORS errors
   - Verify `vite.config.ts` proxy configuration

3. **API endpoint mismatch**
   - Frontend expecting different endpoint

   **Solution**:
   - Check browser Network tab for failed requests
   - Verify API base URL in frontend code
   - Ensure backend routes match frontend expectations

### WebSocket not connecting in Logs view

**Symptoms**:
- Logs view shows "Disconnected" status
- No real-time updates
- Orange/red connection indicator

**Possible Causes**:

1. **Backend not running**
   - WebSocket server requires backend services running

   **Solution**:
   ```bash
   janitarr start
   ```

2. **WebSocket upgrade failed**
   - Proxy not forwarding WebSocket correctly
   - NGINX or reverse proxy misconfiguration

   **Solution**:
   - Check proxy WebSocket configuration
   - Verify `Upgrade` and `Connection` headers forwarded
   - Test direct connection (bypass proxy)

3. **Network issues**
   - Firewall blocking WebSocket
   - Connection timeout

   **Solution**:
   - Check browser console for WebSocket errors
   - Try refreshing page
   - Verify no proxy/firewall blocking WebSocket

### UI not responsive on mobile

**Symptoms**:
- Elements overlapping on mobile
- Horizontal scrolling required
- Touch targets too small

**Possible Causes**:

1. **Old browser**
   - Browser doesn't support modern CSS

   **Solution**:
   - Update browser to latest version
   - Try Chrome, Firefox, or Safari

2. **Zoom level incorrect**
   - Browser zoom set to >100%

   **Solution**:
   - Reset browser zoom to 100%
   - Check viewport meta tag in HTML

3. **Bug in responsive layout**
   - Please report!

   **Solution**:
   - Report issue with screenshot and device info
   - Workaround: use desktop mode temporarily

### Theme not persisting

**Symptoms**:
- Theme resets to light mode on reload
- System theme not detected

**Possible Causes**:

1. **localStorage disabled**
   - Browser privacy settings blocking localStorage

   **Solution**:
   - Enable localStorage in browser settings
   - Check browser privacy mode (incognito/private)

2. **Browser cache cleared**
   - Theme preference stored in localStorage

   **Solution**:
   - Set theme again
   - Preference will persist going forward

---

## Unified Service Startup Issues

### Development mode not proxying to Vite

**Symptoms**:
- `janitarr dev` starts but UI shows errors
- Frontend changes not reflected in browser
- 404 errors for UI assets

**Possible Causes**:

1. **Vite dev server not running**
   - `dev` command only starts backend, Vite must be started separately

   **Solution**:
   ```bash
   # Terminal 1: Start backend with proxy
   janitarr dev

   # Terminal 2: Start Vite dev server
   cd ui && bun run dev
   ```

2. **Wrong port for Vite**
   - Backend expects Vite on port 5173

   **Solution**:
   - Ensure Vite running on default port 5173
   - Check `ui/vite.config.ts` for port configuration

3. **Accessing wrong URL**
   - Should access backend URL, not Vite URL

   **Solution**:
   - Access `http://localhost:3434` (backend with proxy)
   - NOT `http://localhost:5173` (Vite directly)

### Port already in use

**Symptoms**:
- Error: "Address already in use"
- `janitarr start` or `janitarr dev` fails to start

**Possible Causes**:

1. **Port 3434 occupied**
   - Another process using default port
   - Previous Janitarr instance still running

   **Solution**:
   ```bash
   # Check what's using port 3434
   lsof -i :3434

   # Kill previous instance
   janitarr stop

   # Or use different port
   janitarr start --port 8080
   ```

### Graceful shutdown timeout

**Symptoms**:
- Ctrl+C doesn't stop services immediately
- "Waiting for cycle to complete" message appears
- Force exit after 10 seconds

**Possible Causes**:

1. **Active automation cycle**
   - Scheduler completing current cycle before stopping (up to 10 seconds)

   **Solution**:
   - Wait for cycle to complete (automatic, max 10 seconds)
   - Or press Ctrl+C again for immediate force shutdown

2. **Slow server responses**
   - Detection phase taking long time

   **Solution**:
   - Normal behavior, timeout will force exit after 10 seconds
   - Check server performance if cycles consistently slow

### Health check endpoint returning degraded

**Symptoms**:
- `GET /api/health` returns status "degraded"
- HTTP 200 response but scheduler shows disabled

**Possible Causes**:

1. **Scheduler disabled in configuration**
   - Web server running but scheduler not enabled
   - This is expected behavior, not an error

   **Solution**:
   ```bash
   # Enable scheduler
   janitarr config set schedule.enabled true

   # Restart services
   janitarr stop
   janitarr start
   ```

### Metrics endpoint not updating

**Symptoms**:
- `GET /metrics` returns same values
- Counters not incrementing
- Gauges not reflecting current state

**Possible Causes**:

1. **No activity occurring**
   - Metrics only update when events happen

   **Solution**:
   - Trigger automation cycle: `janitarr run`
   - Check if scheduler is running: `janitarr status`
   - Verify servers configured: `janitarr server list`

2. **Accessing stale metrics**
   - Browser caching response

   **Solution**:
   - Hard refresh browser (Ctrl+Shift+R)
   - Use curl for testing: `curl http://localhost:3434/metrics`

---

## Performance Issues

### Slow detection phase

**Symptoms**:
- `janitarr scan` takes minutes to complete
- Automation cycles take very long
- Dashboard slow to load

**Possible Causes**:

1. **Many servers configured**
   - Each server queried sequentially
   - Network latency adds up

   **Solution**:
   - This is normal for many servers
   - Queries run in parallel but still take time
   - Consider: Do you need all servers?

2. **Slow server responses**
   - Radarr/Sonarr slow to respond
   - Large database queries on server side

   **Solution**:
   - Optimize Radarr/Sonarr databases
   - Upgrade server hardware
   - Reduce library size (archive old content)

3. **Network issues**
   - High latency to servers
   - Packet loss

   **Solution**:
   ```bash
   # Check latency
   ping server-hostname

   # Use local network addresses
   # Avoid VPN if possible
   ```

### High memory usage

**Symptoms**:
- Janitarr process using excessive memory
- System swap being used

**Possible Causes**:

1. **Large database**
   - Many log entries
   - Database grown large

   **Solution**:
   ```bash
   # Clear old logs
   janitarr logs --clear

   # Database will shrink automatically
   ```

2. **Memory leak** (rare)
   - Long-running process accumulates memory

   **Solution**:
   - Restart scheduler:
     ```bash
     janitarr stop
     janitarr start
     ```
   - Report issue if recurring

### High CPU usage

**Symptoms**:
- Janitarr process using significant CPU
- System slow during automation

**Possible Causes**:

1. **Database operations**
   - During detection/logging phase
   - Temporary spikes are normal

   **Solution**:
   - Wait for cycle to complete
   - CPU usage should drop after

2. **Too many searches**
   - Triggering hundreds of searches

   **Solution**:
   ```bash
   # Reduce search limits
   janitarr config set limits.missing.movies 5
   janitarr config set limits.missing.episodes 5
   ```

---

## Database Issues

### Database locked

**Symptoms**:
- Error: "database is locked"
- Operations fail intermittently

**Possible Causes**:

1. **Multiple processes accessing database**
   - Two Janitarr instances running
   - Manual database access during automation

   **Solution**:
   ```bash
   # Stop all instances
   janitarr stop

   # Verify no other processes
   lsof data/janitarr.db

   # Start single instance
   janitarr start
   ```

2. **Database on network filesystem**
   - SQLite doesn't handle network FS well
   - File locking issues

   **Solution**:
   - Move database to local filesystem
   - Update `JANITARR_DB_PATH`

### Database corrupted

**Symptoms**:
- Error: "database disk image is malformed"
- Unable to read configuration
- Crashes on startup

**Possible Causes**:
- System crash during write
- Disk full during operation
- Hardware failure

**Solution**:

1. **Try to recover**:
   ```bash
   # Backup corrupted database
   cp data/janitarr.db data/janitarr.db.corrupted

   # Attempt SQLite recovery
   sqlite3 data/janitarr.db ".recover" | sqlite3 data/janitarr-recovered.db
   mv data/janitarr-recovered.db data/janitarr.db
   ```

2. **Start fresh** (if recovery fails):
   ```bash
   # Backup old database
   mv data/janitarr.db data/janitarr.db.old

   # Restart Janitarr (creates new database)
   bun run start

   # Reconfigure servers and settings
   ```

3. **Prevent future corruption**:
   - Ensure disk not full
   - Use reliable storage (SSD)
   - Graceful shutdown (not kill -9)

### Cannot find database

**Symptoms**:
- Error: "unable to open database file"
- Fresh database created unexpectedly

**Possible Causes**:

1. **Wrong working directory**
   - Running Janitarr from different directory
   - Relative path resolves differently

   **Solution**:
   ```bash
   # Use absolute path
   export JANITARR_DB_PATH=/full/path/to/janitarr.db

   # Or always run from project root
   cd /path/to/janitarr
   janitarr status
   ```

2. **Permissions issue**
   - Cannot read/write database file

   **Solution**:
   ```bash
   # Check permissions
   ls -l data/janitarr.db

   # Fix permissions
   chmod 644 data/janitarr.db
   ```

---

## Getting Help

### Before Asking for Help

1. **Check this guide**: Most common issues covered above

2. **Review logs**:
   ```bash
   janitarr logs --limit 100
   ```

3. **Test components**:
   ```bash
   # Test servers
   janitarr server test "Server Name"

   # Test detection
   janitarr scan

   # Test dry-run
   janitarr run --dry-run
   ```

4. **Check configuration**:
   ```bash
   janitarr config show
   janitarr server list
   ```

### Gathering Information

When reporting issues, include:

1. **Janitarr version**:
   ```bash
   git log -1 --oneline
   ```

2. **System information**:
   - Operating system and version
   - Bun version: `bun --version`

3. **Configuration** (sanitized):
   ```bash
   janitarr config show
   janitarr server list  # Remove API keys from output
   ```

4. **Relevant logs**:
   ```bash
   janitarr logs --limit 50
   ```

5. **Steps to reproduce**:
   - What you did
   - What you expected
   - What actually happened

### Where to Get Help

1. **GitHub Issues**: [github.com/yourusername/janitarr/issues](https://github.com/yourusername/janitarr/issues)
   - Search existing issues first
   - Create new issue with information above

2. **Documentation**:
   - [User Guide](user-guide.md)
   - [API Reference](api-reference.md)
   - [Development Guide](development.md)

### Debugging Mode

Enable verbose logging for debugging:

```bash
export JANITARR_LOG_LEVEL=debug
janitarr run
```

This provides detailed output for troubleshooting.

---

## Common Error Messages

### "Server not found"
**Cause**: Server name/ID doesn't exist
**Solution**: Check server list, use correct name

### "Invalid configuration key"
**Cause**: Typo in config key
**Solution**: Use correct key (see config show for valid keys)

### "Invalid interval"
**Cause**: Interval <1 hour
**Solution**: Set interval ≥1: `janitarr config set schedule.interval 6`

### "Database error"
**Cause**: Database corruption or locking
**Solution**: See [Database Issues](#database-issues)

### "Request timeout"
**Cause**: Server slow to respond
**Solution**: Check server performance, network latency

### "Failed to parse response"
**Cause**: Unexpected API response format
**Solution**: Verify server is Radarr/Sonarr v3+, check logs

### "Search limit must be non-negative"
**Cause**: Negative limit value
**Solution**: Use 0 or positive integer

---

Still stuck? Open an issue on GitHub with detailed information!
