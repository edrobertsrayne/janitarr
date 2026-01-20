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

| Spec                                       | Code   | Purpose                                     |
| ------------------------------------------ | ------ | ------------------------------------------- |
| [go-architecture.md](./go-architecture.md) | `src/` | Go patterns, conventions, project structure |

## Server Configuration

| Spec                                                 | Code                                                          | Purpose                                            |
| ---------------------------------------------------- | ------------------------------------------------------------- | -------------------------------------------------- |
| [server-configuration.md](./server-configuration.md) | `src/services/server_manager.go`<br>`src/database/servers.go` | Radarr/Sonarr connections, credentials, validation |

## Content Detection

| Spec                                                           | Code                                                                     | Purpose                                    |
| -------------------------------------------------------------- | ------------------------------------------------------------------------ | ------------------------------------------ |
| [missing-content-detection.md](./missing-content-detection.md) | `src/services/detector.go`<br>`src/api/radarr.go`<br>`src/api/sonarr.go` | Identify missing monitored episodes/movies |
| [quality-cutoff-detection.md](./quality-cutoff-detection.md)   | `src/services/detector.go`<br>`src/api/radarr.go`<br>`src/api/sonarr.go` | Identify media below quality cutoff        |

## Search & Automation

| Spec                                                       | Code                                                                                      | Purpose                                                   |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| [search-triggering.md](./search-triggering.md)             | `src/services/search_trigger.go`<br>`src/api/client.go`                                   | Trigger searches with limits, dry-run mode                |
| [automatic-scheduling.md](./automatic-scheduling.md)       | `src/services/scheduler.go`<br>`src/services/automation.go`                               | Scheduled detection/search cycles, manual triggers        |
| [unified-service-startup.md](./unified-service-startup.md) | `src/cli/start.go`<br>`src/cli/dev.go`<br>`src/web/server.go`<br>`src/metrics/metrics.go` | Unified daemon startup, health checks, Prometheus metrics |

## Logging & Monitoring

| Spec                                         | Code                                                                            | Purpose                                           |
| -------------------------------------------- | ------------------------------------------------------------------------------- | ------------------------------------------------- |
| [logging.md](./logging.md)                   | `src/logger/logger.go`<br>`src/database/logs.go`<br>`src/web/websocket/logs.go` | Unified logging: console, web streaming, database |
| [activity-logging.md](./activity-logging.md) | `src/logger/logger.go`<br>`src/database/logs.go`                                | Audit trail for searches, cycles, failures        |

## Web Frontend

| Spec                                           | Code                                          | Purpose                                              |
| ---------------------------------------------- | --------------------------------------------- | ---------------------------------------------------- |
| [web-frontend.md](./web-frontend.md)           | `src/templates/`<br>`src/web/handlers/pages/` | templ + htmx + Alpine.js UI, WebSocket log streaming |
| [daisyui-migration.md](./daisyui-migration.md) | `src/templates/`<br>`tailwind.config.cjs`     | DaisyUI components, 32-theme switcher                |

## CLI Interface

| Spec                                   | Code                           | Purpose                                        |
| -------------------------------------- | ------------------------------ | ---------------------------------------------- |
| [cli-interface.md](./cli-interface.md) | `src/cli/`<br>`src/cli/forms/` | Interactive terminal forms (charmbracelet/huh) |
