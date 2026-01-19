# API Documentation

HTTP API exposed by the monitoring agent.

## Endpoints

All endpoints are only available via Tailscale network (100.x.y.z).

### GET /static/

Serves the web dashboard interface.

**Access:** Open in browser `http://100.x.y.z:8080/static/`

### GET /api/peers

Returns list of Tailscale peers with online/offline status.

**Response:**
```json
[
  {
    "hostname": "sleipnir",
    "ip": "100.117.120.13",
    "online": true
  },
  {
    "hostname": "niflheim",
    "ip": "100.124.253.7",
    "online": true
  },
  {
    "hostname": "hermod",
    "ip": "100.98.38.44",
    "online": false
  }
]
```

### GET /status

Returns current system metrics.

**Response:**
```json
{
  "timestamp": "2026-01-19T02:04:19Z",
  "hostname": "sleipnir",
  "cpu": {
    "usage_percent": 17.72,
    "load_average": [0.41, 0.84, 1.03],
    "cores": 8
  },
  "memory": {
    "total_bytes": 36567764992,
    "used_bytes": 10795655168,
    "available_bytes": 25958637568,
    "usage_percent": 29.52,
    "swap_total_bytes": 8589930496,
    "swap_used_bytes": 0
  },
  "disk": [
    {
      "device": "/dev/nvme1n1p2",
      "mount_point": "/",
      "total_bytes": 485923799040,
      "used_bytes": 103049240576,
      "free_bytes": 358115794944,
      "usage_percent": 22.35
    }
  ],
  "network": {
    "bytes_sent": 1832973757,
    "bytes_recv": 2056563938,
    "packets_sent": 403355,
    "packets_recv": 689509,
    "interfaces": [
      {
        "name": "tailscale0",
        "addresses": ["100.117.120.13/32"],
        "is_up": true
      }
    ]
  },
  "system": {
    "os": "linux",
    "platform": "alpine",
    "platform_version": "3.23.2",
    "kernel_version": "6.14.0-37-generic",
    "uptime_seconds": 5705
  }
}
```

### GET /health

Returns service health status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-01-19T02:04:19Z"
}
```

### GET /metrics/history

Returns metrics history.

**Query Parameters:**
- `hours` (optional): Number of hours of history (default: 24)

**Example:**
```bash
curl http://100.117.120.13:8080/metrics/history?hours=6
```

**Response:**
```json
{
  "hours": 6,
  "count": 720,
  "metrics": [
    {
      "timestamp": "2026-01-19T02:04:19Z",
      "hostname": "sleipnir",
      "cpu": { ... },
      "memory": { ... },
      "disk": [ ... ],
      "network": { ... },
      "system": { ... }
    }
  ]
}
```

## Security

### Network Restriction

All endpoints validate that requests come from the Tailscale network (100.64.0.0/10).

Requests from outside the Tailscale network receive:
```
HTTP/1.1 403 Forbidden
Forbidden: Only Tailscale network allowed
```

### CORS Headers

All endpoints include CORS headers to allow cross-origin requests within the Tailscale network:
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

## Error Responses

### 403 Forbidden
Request from outside Tailscale network.

### 500 Internal Server Error
Error collecting metrics or querying database.

**Example:**
```json
{
  "error": "Failed to collect metrics: ..."
}
```

## Usage Examples

### Get Current Metrics
```bash
curl http://100.117.120.13:8080/status | jq '.cpu.usage_percent'
```

### Get Device List
```bash
curl http://100.117.120.13:8080/api/peers | jq '.[] | select(.online==true)'
```

### Get 24h History
```bash
curl http://100.117.120.13:8080/metrics/history?hours=24 | jq '.count'
```

### Health Check
```bash
curl http://100.117.120.13:8080/health
```

## Integration

### Prometheus

You can create a Prometheus exporter using the `/status` endpoint:

```yaml
scrape_configs:
  - job_name: 'tailscale-monitor'
    static_configs:
      - targets: ['100.117.120.13:8080']
    metrics_path: '/status'
```

### Grafana

Create a JSON datasource pointing to the `/metrics/history` endpoint for historical data visualization.

### Custom Scripts

```python
import requests

def get_cpu_usage(ip):
    response = requests.get(f'http://{ip}:8080/status')
    data = response.json()
    return data['cpu']['usage_percent']

# Get CPU usage from all online devices
peers = requests.get('http://100.117.120.13:8080/api/peers').json()
for peer in peers:
    if peer['online']:
        cpu = get_cpu_usage(peer['ip'])
        print(f"{peer['hostname']}: {cpu}%")
```

## Rate Limiting

Currently, there is no rate limiting. The agent can handle multiple concurrent requests.

## Versioning

The API follows semantic versioning. Current version: `v1.0.0`

Breaking changes will be announced in advance and will increment the major version.
