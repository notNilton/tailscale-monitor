# tailscale-monitor

Decentralized monitoring system for Tailscale networks. Each peer runs a lightweight agent exposing system metrics via HTTP and persisting local historical data in SQLite.

## Architecture

```
cmd/
  agent/              Agent daemon entrypoint
  client/             CLI client utility
internal/
  config/             Configuration loading
  metrics/            System metrics collection (CPU, Memory, Disk, Network)
  server/             HTTP API server and web dashboard handlers
  storage/            SQLite persistence engine
  tailscale/          Tailscale API and network interface validation
web/                  Frontend dashboard assets
scripts/              Installation and systemd scripts
```

### Components

- `cmd/agent`: Daemon that collects system metrics, exposes HTTP endpoints on the Tailscale interface, and stores time-series metrics in SQLite.
- `cmd/client`: CLI utility (`tsmon`) to discover and query metrics across tailnet peers.
- `web`: Built-in web dashboard serving real-time peer metrics and system health.

## Development

### Prerequisites

- Go 1.22+
- Tailscale installed and running on the host machine
- Docker / Podman (optional)

### Running Services

Build agent and CLI client binaries:

```bash
go build -o bin/agent ./cmd/agent
go build -o bin/client ./cmd/client
```

Run agent daemon locally:

```bash
./bin/agent
```

Run using Docker Compose:

```bash
docker compose up -d
```

### Systemd Installation

```bash
sudo ./scripts/install.sh
```

### Environment Variables

```env
TAILSCALE_API_KEY=tskey-api-xxxxx
TAILSCALE_TAILNET=example.ts.net
```

### Service Endpoints

| Service | Type | Port | Endpoint |
|---------|------|------|----------|
| Web Dashboard | Frontend | `8080` | http://100.x.y.z:8080/static/ |
| Status API | Backend | `8080` | http://100.x.y.z:8080/status |
| Peer List API | Backend | `8080` | http://100.x.y.z:8080/api/peers |
| Health Check | Backend | `8080` | http://100.x.y.z:8080/health |


## Documentation

- [📋 Roadmap & TODOs](docs/TODO.md) - Planned features and project roadmap
- [📐 Architecture](docs/ARCHITECTURE.md) - System architecture and components
