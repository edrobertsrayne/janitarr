# Janitarr Specifications

Automation tool for managing Radarr and Sonarr media servers. Written in Go.

## Technology Stack

| Component       | Technology                                |
| --------------- | ----------------------------------------- |
| Language        | Go 1.22+                                  |
| Web Framework   | Chi (go-chi/chi/v5)                       |
| Database        | modernc.org/sqlite (pure Go, no CGO)      |
| CLI             | Cobra (spf13/cobra)                       |
| CLI Forms       | charmbracelet/huh                         |
| Console Logging | charmbracelet/log                         |
| Templates       | templ (a-h/templ)                         |
| Frontend        | htmx + Alpine.js + Tailwind CSS + DaisyUI |

## Core Architecture

| Spec                                       | Code   | Purpose                                     | Status |
| ------------------------------------------ | ------ | ------------------------------------------- | ------ |
| [go-architecture.md](./go-architecture.md) | `src/` | Go patterns, conventions, project structure | Active |

## Server Configuration

| Spec                                                 | Code                                                          | Purpose                                            | Status |
| ---------------------------------------------------- | ------------------------------------------------------------- | -------------------------------------------------- | ------ |
| [server-configuration.md](./server-configuration.md) | `src/services/server_manager.go`<br>`src/database/servers.go` | Radarr/Sonarr connections, credentials, validation | Active |

## Content Detection

| Spec                                                           | Code                                                                     | Purpose                                    | Status |
| -------------------------------------------------------------- | ------------------------------------------------------------------------ | ------------------------------------------ | ------ |
| [missing-content-detection.md](./missing-content-detection.md) | `src/services/detector.go`<br>`src/api/radarr.go`<br>`src/api/sonarr.go` | Identify missing monitored episodes/movies | Active |
| [quality-cutoff-detection.md](./quality-cutoff-detection.md)   | `src/services/detector.go`<br>`src/api/radarr.go`<br>`src/api/sonarr.go` | Identify media below quality cutoff        | Active |

## Search & Automation

| Spec                                                       | Code                                                                                      | Purpose                                                   | Status |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------- | --------------------------------------------------------- | ------ |
| [search-triggering.md](./search-triggering.md)             | `src/services/search_trigger.go`<br>`src/api/client.go`                                   | Trigger searches with limits, dry-run mode                | Active |
| [automatic-scheduling.md](./automatic-scheduling.md)       | `src/services/scheduler.go`<br>`src/services/automation.go`                               | Scheduled detection/search cycles, manual triggers        | Active |
| [unified-service-startup.md](./unified-service-startup.md) | `src/cli/start.go`<br>`src/cli/dev.go`<br>`src/web/server.go`<br>`src/metrics/metrics.go` | Unified daemon startup, health checks, Prometheus metrics | Active |

## Logging & Monitoring

| Spec                                         | Code                                                                            | Purpose                                           | Status |
| -------------------------------------------- | ------------------------------------------------------------------------------- | ------------------------------------------------- | ------ |
| [logging.md](./logging.md)                   | `src/logger/logger.go`<br>`src/database/logs.go`<br>`src/web/websocket/logs.go` | Unified logging: console, web streaming, database | Active |
| [activity-logging.md](./activity-logging.md) | `src/logger/logger.go`<br>`src/database/logs.go`                                | Audit trail for searches, cycles, failures        | Active |

## Web Frontend

| Spec                                                   | Code                                          | Purpose                                              | Status   |
| ------------------------------------------------------ | --------------------------------------------- | ---------------------------------------------------- | -------- |
| [web-frontend.md](./web-frontend.md)                   | `src/templates/`<br>`src/web/handlers/pages/` | templ + htmx + Alpine.js UI, WebSocket log streaming | Active   |
| [daisyui-migration.md](./archive/daisyui-migration.md) | `src/templates/`<br>`tailwind.config.cjs`     | DaisyUI component migration guide                    | Archived |

## CLI Interface

| Spec                                   | Code                           | Purpose                                        | Status |
| -------------------------------------- | ------------------------------ | ---------------------------------------------- | ------ |
| [cli-interface.md](./cli-interface.md) | `src/cli/`<br>`src/cli/forms/` | Interactive terminal forms (charmbracelet/huh) | Active |
