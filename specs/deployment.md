# Deployment Specification

Janitarr supports two distribution methods: Docker images and NixOS modules.

## Docker Image

### Registry & Tags

| Property | Value                             |
| -------- | --------------------------------- |
| Registry | GitHub Container Registry (GHCR)  |
| Image    | `ghcr.io/edrobertsrayne/janitarr` |
| Archs    | `linux/amd64`, `linux/arm64`      |

Docker images are only built and pushed when a git tag is created. This ensures `latest` always points to a tested release rather than bleeding-edge development.

**Tag strategy** (for a release tagged `v1.2.3`):

| Docker Tag | Description                                    |
| ---------- | ---------------------------------------------- |
| `v1.2.3`   | Exact version, immutable                       |
| `v1.2`     | Rolling tag, points to latest `v1.2.x` patch   |
| `v1`       | Rolling tag, points to latest `v1.x.x` release |
| `latest`   | Rolling tag, points to most recent release     |

Rolling tags allow users to receive bug fixes automatically (`v1.2`) while avoiding unexpected feature changes (`v1`), or to pin to exact versions for reproducibility (`v1.2.3`).

### Base Image

Alpine Linux (`alpine:latest`). Provides a small footprint with shell access for debugging.

### Configuration

All configuration via environment variables:

| Variable             | Required | Default | Description                          |
| -------------------- | -------- | ------- | ------------------------------------ |
| `PUID`               | No       | `1000`  | User ID to run as                    |
| `PGID`               | No       | `1000`  | Group ID to run as                   |
| `JANITARR_PORT`      | No       | `3434`  | Port for web UI                      |
| `JANITARR_LOG_LEVEL` | No       | `info`  | Log level (debug, info, warn, error) |

### Data Persistence

| Container Path | Purpose                              |
| -------------- | ------------------------------------ |
| `/data`        | SQLite database and application data |

The container creates `/data/janitarr.db` on first run. Users must mount a volume to persist data across container restarts.

### User Management

The container runs as a non-root user. On startup, an entrypoint script:

1. Creates a user `janitarr` with UID from `PUID` (default: 1000)
2. Creates a group `janitarr` with GID from `PGID` (default: 1000)
3. Adjusts ownership of `/data` to match
4. Drops privileges and runs the application as `janitarr`

This pattern matches Linuxserver.io conventions and allows users to match host filesystem permissions.

### Health Check

The Dockerfile includes a health check against the application's health endpoint:

```dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:${JANITARR_PORT:-3434}/health || exit 1
```

### Example Usage

```bash
docker run -d \
  --name janitarr \
  -p 3434:3434 \
  -e PUID=1000 \
  -e PGID=1000 \
  -v /path/to/data:/data \
  ghcr.io/edrobertsrayne/janitarr:latest
```

Docker Compose:

```yaml
services:
  janitarr:
    image: ghcr.io/edrobertsrayne/janitarr:latest
    container_name: janitarr
    environment:
      - PUID=1000
      - PGID=1000
    volumes:
      - ./data:/data
    ports:
      - "3434:3434"
    restart: unless-stopped
```

### Dockerfile Structure

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o janitarr .

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache su-exec shadow

COPY --from=builder /build/janitarr /usr/local/bin/janitarr
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh

ENV PUID=1000 \
    PGID=1000 \
    JANITARR_PORT=3434 \
    JANITARR_DB_PATH=/data/janitarr.db

EXPOSE 3434

VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:${JANITARR_PORT:-3434}/health || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["janitarr", "start", "--host", "0.0.0.0"]
```

### Entrypoint Script

```bash
#!/bin/sh
set -e

PUID=${PUID:-1000}
PGID=${PGID:-1000}

# Create group if it doesn't exist
if ! getent group janitarr > /dev/null 2>&1; then
    addgroup -g "${PGID}" janitarr
fi

# Create user if it doesn't exist
if ! getent passwd janitarr > /dev/null 2>&1; then
    adduser -D -u "${PUID}" -G janitarr -h /data -s /sbin/nologin janitarr
fi

# Ensure correct ownership
chown -R janitarr:janitarr /data

# Drop privileges and execute
exec su-exec janitarr:janitarr "$@"
```

---

## NixOS Module

### Flake Outputs

The `flake.nix` provides:

| Output                       | Description                      |
| ---------------------------- | -------------------------------- |
| `packages.<system>.janitarr` | Standalone package               |
| `packages.<system>.default`  | Alias to janitarr package        |
| `nixosModules.janitarr`      | NixOS module for systemd service |
| `nixosModules.default`       | Alias to janitarr module         |

Supported systems: `x86_64-linux`, `aarch64-linux`.

### Module Options

```nix
services.janitarr = {
  enable = mkEnableOption "Janitarr media server automation";

  port = mkOption {
    type = types.port;
    default = 3434;
    description = "Port for the web interface.";
  };

  openFirewall = mkOption {
    type = types.bool;
    default = false;
    description = "Whether to open the firewall for the web interface.";
  };

  dataDir = mkOption {
    type = types.path;
    default = "/var/lib/janitarr";
    description = "Directory for database and application data.";
  };

  logLevel = mkOption {
    type = types.enum [ "debug" "info" "warn" "error" ];
    default = "info";
    description = "Log verbosity level.";
  };

  user = mkOption {
    type = types.str;
    default = "janitarr";
    description = "User account under which janitarr runs.";
  };

  group = mkOption {
    type = types.str;
    default = "janitarr";
    description = "Group under which janitarr runs.";
  };

  package = mkOption {
    type = types.package;
    default = pkgs.janitarr;
    description = "The janitarr package to use.";
  };
};
```

### Systemd Service

The module creates a systemd service with security hardening:

```nix
systemd.services.janitarr = {
  description = "Janitarr media server automation";
  wantedBy = [ "multi-user.target" ];
  after = [ "network.target" ];

  serviceConfig = {
    Type = "simple";
    User = cfg.user;
    Group = cfg.group;
    ExecStart = "${cfg.package}/bin/janitarr start --host 0.0.0.0 --port ${toString cfg.port}";
    Restart = "on-failure";
    RestartSec = 5;

    # State management
    StateDirectory = "janitarr";
    StateDirectoryMode = "0750";
    WorkingDirectory = cfg.dataDir;

    # Security hardening
    NoNewPrivileges = true;
    PrivateTmp = true;
    ProtectSystem = "strict";
    ProtectHome = true;
    ProtectKernelTunables = true;
    ProtectKernelModules = true;
    ProtectControlGroups = true;
    RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX" ];
    RestrictNamespaces = true;
    RestrictRealtime = true;
    RestrictSUIDSGID = true;
    MemoryDenyWriteExecute = true;
    LockPersonality = true;
    SystemCallFilter = [ "@system-service" "~@privileged" "~@resources" ];
    SystemCallArchitectures = "native";
    CapabilityBoundingSet = "";
    AmbientCapabilities = "";
  };

  environment = {
    JANITARR_DB_PATH = "${cfg.dataDir}/janitarr.db";
    JANITARR_LOG_LEVEL = cfg.logLevel;
  };
};
```

### User and Group Creation

When using the default user/group:

```nix
users.users.janitarr = mkIf (cfg.user == "janitarr") {
  isSystemUser = true;
  group = cfg.group;
  home = cfg.dataDir;
  description = "Janitarr service user";
};

users.groups.janitarr = mkIf (cfg.group == "janitarr") { };
```

### Firewall Integration

```nix
networking.firewall.allowedTCPPorts = mkIf cfg.openFirewall [ cfg.port ];
```

### Version Pinning

The flake builds from source at whatever commit it's pinned to. For stable deployments, pin to a release tag:

```nix
# Pin to a specific release
janitarr.url = "github:edrobertsrayne/janitarr?ref=v1.2.3";

# Or follow the latest release (update with nix flake update)
janitarr.url = "github:edrobertsrayne/janitarr";
```

Use `nix flake update janitarr` to update to the latest version.

### Example Usage

In a NixOS configuration:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    janitarr.url = "github:edrobertsrayne/janitarr?ref=v1.0.0";
  };

  outputs = { self, nixpkgs, janitarr }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        janitarr.nixosModules.janitarr
        {
          services.janitarr = {
            enable = true;
            port = 3434;
            openFirewall = true;
            logLevel = "info";
          };
        }
      ];
    };
  };
}
```

Standalone package usage:

```nix
environment.systemPackages = [ janitarr.packages.x86_64-linux.janitarr ];
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

The repository includes GitHub Actions workflows for automated builds and validation.

#### Trigger Conditions

| Event        | Condition     | Action                           |
| ------------ | ------------- | -------------------------------- |
| Push         | `main` branch | Build only (validation, no push) |
| Pull Request | Any branch    | Build only (validation, no push) |
| Push         | Tag `v*`      | Build and push with version tags |

Docker images are only published when a version tag is pushed. Builds on `main` and PRs validate the Dockerfile works but do not push to the registry.

#### Docker Build Workflow

`.github/workflows/docker.yml`:

```yaml
name: Docker

on:
  push:
    branches: [main]
    tags: ["v*"]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up QEMU
        uses: docker/setup-qemu-action@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to GHCR
        if: startsWith(github.ref, 'refs/tags/v')
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            # For v1.2.3: creates v1.2.3, v1.2, v1, latest
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
            type=raw,value=latest,enable=${{ startsWith(github.ref, 'refs/tags/v') }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: ${{ startsWith(github.ref, 'refs/tags/v') }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

#### Flake Check Workflow

`.github/workflows/nix.yml`:

```yaml
name: Nix

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install Nix
        uses: DeterminateSystems/nix-installer-action@v9

      - name: Setup Nix cache
        uses: DeterminateSystems/magic-nix-cache-action@v2

      - name: Check flake
        run: nix flake check

      - name: Build package
        run: nix build .#janitarr
```

---

## File Structure

```
/
├── Dockerfile
├── docker-entrypoint.sh
├── flake.nix
├── flake.lock
├── nix/
│   ├── package.nix          # Package derivation
│   └── module.nix           # NixOS module
└── .github/
    └── workflows/
        ├── docker.yml        # Docker build/push
        └── nix.yml           # Flake checks
```

---

## Implementation Checklist

- [ ] Create `Dockerfile` with multi-stage build
- [ ] Create `docker-entrypoint.sh` with PUID/PGID handling
- [ ] Add `/health` endpoint to application (if not present)
- [ ] Create `flake.nix` with package and module outputs
- [ ] Create `nix/package.nix` derivation
- [ ] Create `nix/module.nix` with all options
- [ ] Create `.github/workflows/docker.yml`
- [ ] Create `.github/workflows/nix.yml`
- [ ] Test Docker image locally on amd64
- [ ] Test NixOS module in a VM
- [ ] Configure GHCR repository visibility
