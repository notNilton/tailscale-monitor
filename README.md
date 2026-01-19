# Tailscale Network Monitor

A decentralized monitoring system for Tailscale networks built in Go. Each device exposes its metrics via a lightweight HTTP server and stores historical data locally in SQLite.

## ✨ Features

- 🔒 **Secure**: Runs only on Tailscale interface (100.x.y.z)
- 📊 **Complete Metrics**: CPU, Memory, Disk, Network, System Info
- 💾 **Local Storage**: SQLite with automatic cleanup
- 🌐 **Web Dashboard**: Modern brutalist UI for monitoring
- 🔍 **Peer Discovery**: Automatic detection of Tailscale devices
- 🎯 **Peer-to-Peer**: Direct communication between devices
- 🐳 **Docker Ready**: Easy deployment with Docker Compose
- 📱 **Responsive**: Works on desktop, tablet, and mobile

## 📊 Collected Metrics

- **CPU**: Usage percentage, load average, number of cores
- **Memory**: Total, used, available, swap
- **Disk**: Usage per partition, I/O
- **Network**: Bytes sent/received, active interfaces
- **System**: Uptime, hostname, OS, kernel version

## 🌐 Web Dashboard

Each agent serves a **modern web interface** that allows peer-to-peer navigation between all Tailscale network devices.

### Access the Dashboard

Open your browser and access any device on the network:

```
http://100.x.y.z:8080/static/
```

### Features

- 📋 **Device List**: View all Tailscale peers with online/offline status
- 🔍 **Search**: Filter devices by hostname or IP
- 📊 **Real-time Metrics**: CPU, memory, disk, network, and system
- 🎨 **Modern Design**: Brutalist interface with clean aesthetics
- 📱 **Responsive**: Works perfectly on desktop, tablet, and mobile
- 🔄 **Auto-refresh**: Automatic updates every 30 seconds
- 🌐 **Peer-to-Peer**: Direct requests between devices (no central server)

### How to Use

1. Access the dashboard on any device: `http://100.x.y.z:8080/static/`
2. See the device list in the sidebar
3. Click on a device to view its metrics
4. Metrics are automatically updated

## 🔑 Tailscale API Configuration

For the dashboard to list devices on your network, you need to configure Tailscale API credentials.

### Step 1: Get API Key

1. Access [Tailscale Admin Console](https://login.tailscale.com/admin/settings/keys)
2. Click **Generate API key**
3. Give it a descriptive name (e.g., "Network Monitor")
4. Copy the generated key (starts with `tskey-api-`)

### Step 2: Identify your Tailnet

Your tailnet is your Tailscale network domain. Examples:
- `example.com`
- `example.ts.net`
- `user@example.com`

You can find it at [General Settings](https://login.tailscale.com/admin/settings/general).

### Step 3: Configure Environment Variables

**Docker:**
```bash
# Create a .env file
cp .env.example .env

# Edit and add your credentials
TAILSCALE_API_KEY=tskey-api-xxxxx
TAILSCALE_TAILNET=example.ts.net
```

**Systemd:**
```bash
# Edit the service file
sudo systemctl edit tailscale-monitor

# Add:
[Service]
Environment="TAILSCALE_API_KEY=tskey-api-xxxxx"
Environment="TAILSCALE_TAILNET=example.ts.net"
```

**Manual:**
```bash
export TAILSCALE_API_KEY=tskey-api-xxxxx
export TAILSCALE_TAILNET=example.ts.net
./bin/agent
```

## 🚀 Installation

### Option 1: Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-user/network-infra
cd network-infra

# Configure environment variables
cp .env.example .env
# Edit .env with your Tailscale API credentials

# Start with Docker Compose
docker compose up -d

# View logs
docker compose logs -f
```

### Option 2: Binary

```bash
# Build
go build -o bin/agent ./cmd/agent
go build -o bin/client ./cmd/client

# Run agent
./bin/agent

# Use client
./bin/client list
./bin/client status <hostname>
```

### Option 3: Systemd (Linux)

```bash
# Run installation script
sudo ./scripts/install.sh

# Check status
sudo systemctl status tailscale-monitor

# View logs
sudo journalctl -u tailscale-monitor -f
```

## 📖 Usage

### Web Dashboard

Access the web interface:
```bash
http://100.117.120.13:8080/static/
```

Features:
- View all Tailscale devices
- Click on any device with monitor running (green)
- See real-time metrics
- Auto-refresh every 30 seconds

### CLI Client

```bash
# List all peers
tsmon list

# Query specific peer
tsmon status sleipnir

# Query all online peers
tsmon status --all

# JSON output
tsmon status sleipnir --json
```

### HTTP API

```bash
# Current metrics
curl http://100.117.120.13:8080/status

# Device list
curl http://100.117.120.13:8080/api/peers

# Metrics history
curl http://100.117.120.13:8080/metrics/history?hours=24

# Health check
curl http://100.117.120.13:8080/health
```

## ⚙️ Configuration

Create a `config.yaml` file:

```yaml
server:
  port: 8080
  tailscale_only: true

storage:
  path: /data/metrics.db
  retention_days: 30

metrics:
  collection_interval: 30s

tailscale:
  api_key: ""  # or use TAILSCALE_API_KEY env var
  tailnet: ""  # or use TAILSCALE_TAILNET env var
  use_cli: false  # true to use CLI, false to use API
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│                 Tailscale Network                   │
│                  (100.x.y.z/10)                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    │
│  │ Device 1 │    │ Device 2 │    │ Device 3 │    │
│  │          │    │          │    │          │    │
│  │  Agent   │◄──►│  Agent   │◄──►│  Agent   │    │
│  │  :8080   │    │  :8080   │    │  :8080   │    │
│  │          │    │          │    │          │    │
│  │ SQLite   │    │ SQLite   │    │ SQLite   │    │
│  └──────────┘    └──────────┘    └──────────┘    │
│                                                     │
└─────────────────────────────────────────────────────┘
```

**Key Components:**
- **Agent**: HTTP server + metrics collector + SQLite storage
- **Client**: CLI for querying peers
- **Web Dashboard**: Browser interface for monitoring
- **Tailscale**: Secure network layer

## 🔒 Security

- ✅ Listens **only** on Tailscale interface (100.x.y.z)
- ✅ Validates that requests come from Tailscale network
- ✅ No public internet exposure
- ✅ Encrypted communication via Tailscale (WireGuard)
- ✅ No central server (decentralized)

## 📚 Documentation

- [API Documentation](docs/API.md)
- [Installation Guide](scripts/install.sh)
- [Configuration Examples](config.example.yaml)

## 🛠️ Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o bin/agent ./cmd/agent
go build -o bin/client ./cmd/client

# Run locally
./bin/agent
```

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Support

For issues and questions, please open an issue on GitHub.
