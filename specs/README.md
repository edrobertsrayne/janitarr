# Janitarr Specifications

This directory contains the design specifications for Janitarr, an automation tool for managing Radarr and Sonarr media servers. These specifications define the requirements, acceptance criteria, and implementation constraints for each feature area.

## About These Specifications

Each specification document follows a consistent structure:

- **Context**: Background and purpose of the feature
- **Requirements**: User stories with acceptance criteria
- **Edge Cases & Constraints**: Technical considerations and limitations

The specifications are implementation-agnostic and focus on _what_ the system should do rather than _how_ it should be built.

## Technology Stack

Janitarr is implemented in Go with the following technologies:

- **Language**: Go 1.22+
- **Web Framework**: Chi (go-chi/chi/v5)
- **Database**: modernc.org/sqlite (pure Go, no CGO)
- **CLI**: Cobra (spf13/cobra)
- **Templates**: templ (a-h/templ)
- **Frontend**: htmx + Alpine.js + Tailwind CSS

See [go-architecture.md](./go-architecture.md) for Go-specific patterns and conventions.

## Specifications by Category

### Configuration

| Spec                                                 | Code                                                          | Purpose                                                                    |
| ---------------------------------------------------- | ------------------------------------------------------------- | -------------------------------------------------------------------------- |
| [server-configuration.md](./server-configuration.md) | `src/services/server_manager.go`<br>`src/database/servers.go` | Managing Radarr and Sonarr server connections, credentials, and validation |

### Detection

| Spec                                                           | Code                                                                     | Purpose                                                                            |
| -------------------------------------------------------------- | ------------------------------------------------------------------------ | ---------------------------------------------------------------------------------- |
| [missing-content-detection.md](./missing-content-detection.md) | `src/services/detector.go`<br>`src/api/radarr.go`<br>`src/api/sonarr.go` | Identifying monitored episodes and movies that are missing from media libraries    |
| [quality-cutoff-detection.md](./quality-cutoff-detection.md)   | `src/services/detector.go`<br>`src/api/radarr.go`<br>`src/api/sonarr.go` | Identifying media that exists but hasn't met the configured quality profile cutoff |

### Actions

| Spec                                           | Code                                                    | Purpose                                                                                                                                  |
| ---------------------------------------------- | ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| [search-triggering.md](./search-triggering.md) | `src/services/search_trigger.go`<br>`src/api/client.go` | Initiating content searches in Radarr/Sonarr with user-defined limits per category<br>**Includes:** Dry-run mode for previewing searches |

### Automation

| Spec                                                       | Code                                                                                      | Purpose                                                                                                                                                              |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [automatic-scheduling.md](./automatic-scheduling.md)       | `src/services/scheduler.go`<br>`src/services/automation.go`                               | Configuring and executing detection and search operations on a scheduled interval<br>**Includes:** Manual triggers, dry-run preview mode                             |
| [unified-service-startup.md](./unified-service-startup.md) | `src/cli/start.go`<br>`src/cli/dev.go`<br>`src/web/server.go`<br>`src/metrics/metrics.go` | Running scheduler daemon and web server together with unified startup commands<br>**Includes:** Development mode, production mode, health checks, Prometheus metrics |

### Monitoring

| Spec                                         | Code                                             | Purpose                                                                                         |
| -------------------------------------------- | ------------------------------------------------ | ----------------------------------------------------------------------------------------------- |
| [activity-logging.md](./activity-logging.md) | `src/logger/logger.go`<br>`src/database/logs.go` | Recording all triggered searches, automation cycles, and failures for audit and troubleshooting |

### User Interface

| Spec                                 | Code                                          | Purpose                                                                                                                                                              |
| ------------------------------------ | --------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [web-frontend.md](./web-frontend.md) | `src/templates/`<br>`src/web/handlers/pages/` | Modern web interface using templ templates, htmx for dynamic updates, and Alpine.js for interactivity<br>**Includes:** Tailwind CSS styling, WebSocket log streaming |

### Architecture

| Spec                                       | Purpose                                                       |
| ------------------------------------------ | ------------------------------------------------------------- |
| [go-architecture.md](./go-architecture.md) | Go-specific patterns, conventions, and implementation details |

## Implementation Flow

The specifications are organized to reflect the logical flow of the automation system:

1. **Configuration** → User configures server connections
2. **Detection** → System identifies missing content and quality upgrade opportunities
3. **Actions** → System triggers searches based on configured limits
4. **Automation** → System executes the detection and action cycle on schedule
5. **Monitoring** → System logs all operations for visibility and troubleshooting

## Reading Guide

If you're new to the project, read the specifications in this recommended order:

1. Start with [server-configuration.md](./server-configuration.md) - the foundation
2. Read the detection specs: [missing-content-detection.md](./missing-content-detection.md) and [quality-cutoff-detection.md](./quality-cutoff-detection.md)
3. Understand actions via [search-triggering.md](./search-triggering.md)
4. Learn about automation in [automatic-scheduling.md](./automatic-scheduling.md)
5. See [unified-service-startup.md](./unified-service-startup.md) for running the complete application
6. Review [activity-logging.md](./activity-logging.md) for visibility requirements
7. (Optional) See [web-frontend.md](./web-frontend.md) for web UI specifications

## Contributing

When adding new features:

1. Create a specification document in this directory first
2. Follow the existing document structure (Context → Requirements → Edge Cases)
3. Update this README to include your new specification in the appropriate category
4. Link to the relevant implementation code locations
