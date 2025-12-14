# Home-Run

A self-hosted service monitoring dashboard.

## Quick Start

### Prerequisites

- Docker and Docker Compose
- (For local dev) Go 1.24+, Node.js 20+

### Docker Setup

1. Create your configuration file:

```bash
cp backend/config.yml backend/config.yml
```

2. Edit `backend/config.yml` with your settings (see Configuration section below).

3. Run with Docker Compose:

```bash
docker-compose up -d
```

The app will be available at `http://localhost:8085`.

## Configuration

Create `backend/config.yml` with the following structure:

```yaml
server:
  port: 8085
  session_secret: "change-this-to-a-random-32-char-string!"
  cors_allow_origin: "http://localhost:8085"

auth:
  username: admin
  password: your-secure-password
  api_token: "your-federation-api-token-here"

services:
  - name: My Service
    url: http://localhost
    port: 8080
    backend: docker
    container_name: my-container
```

### Configuration Options

| Field | Description |
|-------|-------------|
| `server.port` | Port the server listens on |
| `server.session_secret` | Secret for session encryption (min 32 chars) |
| `auth.username` | Login username |
| `auth.password` | Login password |
| `auth.api_token` | Token for federation between hosts |
| `services[].name` | Display name for the service |
| `services[].url` | Base URL of the service |
| `services[].port` | Port number |
| `services[].backend` | Backend type (`docker` or `uptime_kuma`) |
| `services[].container_name` | Docker container name (required for `docker` backend) |
| `services[].kuma_monitor_id` | Uptime Kuma monitor ID (required for `uptime_kuma` backend) |
| `services[].configs` | List of config file paths on host to display in UI |

### Service Examples

#### Docker Backend

Monitor services running in Docker containers:

```yaml
services:
  - name: Nginx Proxy
    url: http://localhost
    port: 80
    backend: docker
    container_name: nginx-proxy

  - name: Postgres Database
    url: http://localhost
    port: 5432
    backend: docker
    container_name: postgres-db
    configs:
      - /opt/postgres/postgresql.conf
      - /opt/postgres/pg_hba.conf
```

#### Uptime Kuma Backend

Monitor services via Uptime Kuma (requires `uptime_kuma` config):

```yaml
uptime_kuma:
  url: http://localhost:3001
  api_key: "uk_xxxxxxxxxxxxxxxxxxxx"

services:
  - name: Public Website
    url: https://example.com
    port: 443
    backend: uptime_kuma
    kuma_monitor_id: 1

  - name: API Gateway
    url: https://api.example.com
    port: 443
    backend: uptime_kuma
    kuma_monitor_id: 2

  - name: Internal Service
    url: http://192.168.1.50
    port: 8080
    backend: uptime_kuma
    kuma_monitor_id: 5
```

#### Mixed Backends

Combine both backends in a single config:

```yaml
uptime_kuma:
  url: http://localhost:3001
  api_key: "uk_xxxxxxxxxxxxxxxxxxxx"

services:
  # Local Docker services
  - name: Redis Cache
    url: http://localhost
    port: 6379
    backend: docker
    container_name: redis

  - name: App Server
    url: http://localhost
    port: 3000
    backend: docker
    container_name: app-server
    configs:
      - /opt/app/config.json
      - /opt/app/.env

  # External services via Uptime Kuma
  - name: CDN Endpoint
    url: https://cdn.example.com
    port: 443
    backend: uptime_kuma
    kuma_monitor_id: 10
```

### Service Config Files

Use the `configs` field to specify paths to config files on the host system. These will be viewable in the dashboard UI.

When running Home-Run in Docker, you must mount these paths in `docker-compose.yml`:

```yaml
# docker-compose.yml
services:
  home-run:
    build: .
    container_name: home-run
    ports:
      - "8085:8085"
    volumes:
      - ./backend/config.yml:/app/config.yml:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
      # Mount config files you want to view in the UI
      - /opt/traefik:/opt/traefik:ro
      - /opt/homeassistant:/opt/homeassistant:ro
    restart: unless-stopped
```

Then reference them in your `config.yml`:

```yaml
services:
  - name: Traefik
    url: http://localhost
    port: 80
    backend: docker
    container_name: traefik
    configs:
      - /opt/traefik/traefik.yml
      - /opt/traefik/dynamic.yml

  - name: Home Assistant
    url: http://localhost
    port: 8123
    backend: docker
    container_name: homeassistant
    configs:
      - /opt/homeassistant/configuration.yaml
      - /opt/homeassistant/automations.yaml
```

### Uptime Kuma Integration

To use the `uptime_kuma` backend, configure the connection:

```yaml
uptime_kuma:
  url: http://localhost:3001
  api_key: "uk_xxxxxxxxxxxxxxxxxxxx"  # From Uptime Kuma Settings > API Keys
```

You can find the monitor ID in Uptime Kuma by clicking on a monitor - the ID is in the URL (e.g., `/dashboard/1` means `kuma_monitor_id: 1`).

### Optional: Remote Host Federation

```yaml
remote_hosts:
  - name: Server 2
    endpoint: http://192.168.1.100:8080/api
    token: "their-api-token"
```

## Local Development

```bash
# Install dependencies
npm install
cd backend && go mod download

# Run backend
make run-backend

# Run frontend (separate terminal)
npm run dev
```

## Make Targets

```
make help          # Show all available targets
make build         # Build backend and frontend
make test          # Run tests
make lint          # Run all linters
make clean         # Clean build artifacts
```
