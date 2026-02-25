# Podman Installation Guide

This guide covers installing and running HelixAgent using Podman, a daemonless container engine.

## Prerequisites

- Podman 4.0+
- Podman Compose
- 8 GB RAM minimum
- 20 GB disk space

## Installation

### Fedora/RHEL/CentOS

```bash
sudo dnf install podman podman-compose
```

### Ubuntu/Debian

```bash
sudo apt update
sudo apt install podman podman-compose
```

### macOS

```bash
brew install podman podman-compose
podman machine init
podman machine start
```

## Quick Start

```bash
# Clone repository
git clone git@github.com:anomaly/helixagent.git
cd helixagent

# Start services with Podman
make podman-run
```

## Podman vs Docker

| Feature | Docker | Podman |
|---------|--------|--------|
| Daemon | Required | None |
| Root | Required | Rootless |
| Socket | `/var/run/docker.sock` | `XDG_RUNTIME_DIR/podman/podman.sock` |
| Compose | `docker compose` | `podman-compose` |

## Rootless Mode

Podman runs rootless by default, providing better security:

```bash
# Configure rootless storage
mkdir -p ~/.config/containers
cat > ~/.config/containers/storage.conf <<EOF
[storage]
driver = "overlay"
runroot = "/run/user/$(id -u)/containers"
graphroot = "$HOME/.local/share/containers/storage"
EOF
```

## Running HelixAgent

### Using Makefile

```bash
# Start all services
make podman-run

# View logs
make podman-logs

# Stop services
make podman-stop
```

### Using Podman Compose Directly

```bash
# Start services
podman-compose up -d

# View logs
podman-compose logs -f helixagent

# Stop services
podman-compose down
```

## Pod Configuration

### pods/docker-compose.yml Equivalent

```yaml
version: "3.8"
services:
  helixagent:
    image: helixagent:latest
    ports:
      - "7061:7061"
    environment:
      - GIN_MODE=release
    volumes:
      - ./configs:/app/configs:ro,Z
```

Note the `:Z` suffix for SELinux labeling.

## SELinux Considerations

On SELinux-enabled systems, use volume labels:

```bash
# :z - Shared label (multiple containers)
# :Z - Private label (single container)

podman run -v ./data:/app/data:Z helixagent:latest
```

## Systemd Integration

Create a systemd service for automatic startup:

```bash
# Generate systemd unit files
podman generate systemd --name helixagent --files --new

# Move to systemd directory
mv container-helixagent.service ~/.config/systemd/user/

# Enable and start
systemctl --user enable --now container-helixagent
```

## Troubleshooting

### Permission Denied

```bash
# Check subuid/subgid
cat /etc/subuid /etc/subgid

# If missing, add mappings
sudo usermod --add-subuids 100000-165535 --add-subgids 100000-165535 $USER
```

### Port Binding Issues

```bash
# For rootless ports < 1024
sudo sysctl net.ipv4.ip_unprivileged_port_start=0

# Or use higher ports
podman run -p 7061:7061 helixagent:latest
```

### Storage Issues

```bash
# Check storage
podman system df

# Prune unused data
podman system prune -a
```

## Migration from Docker

```bash
# Pull images from Docker
podman pull docker.io/helixagent:latest

# Migrate compose files (usually compatible)
podman-compose up -d
```
