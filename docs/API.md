# API Documentation

## Endpoints

Todos os endpoints estão disponíveis apenas via rede Tailscale (100.x.y.z).

### GET /static/

Serve a interface web do dashboard.

**Acesso:** Abra no navegador `http://100.x.y.z:8080/static/`

### GET /api/peers

Retorna lista de peers Tailscale com status online/offline.

**Resposta:**
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

Retorna as métricas atuais do sistema.

**Resposta:**
```json
{
  "timestamp": "2026-01-18T21:45:00-04:00",
  "hostname": "server-prod",
  "cpu": {
    "usage_percent": 23.45,
    "load_average": [1.20, 1.15, 1.10],
    "cores": 4
  },
  "memory": {
    "total_bytes": 8589934592,
    "used_bytes": 3435973836,
    "available_bytes": 5153960756,
    "usage_percent": 40.0,
    "swap_total_bytes": 2147483648,
    "swap_used_bytes": 0
  },
  "disk": [
    {
      "device": "/dev/sda1",
      "mount_point": "/",
      "total_bytes": 107374182400,
      "used_bytes": 48547348480,
      "free_bytes": 58826833920,
      "usage_percent": 45.2
    }
  ],
  "network": {
    "bytes_sent": 483183820800,
    "bytes_recv": 1319413953331,
    "packets_sent": 1234567,
    "packets_recv": 2345678,
    "interfaces": [
      {
        "name": "tailscale0",
        "addresses": ["100.64.1.10"],
        "is_up": true
      }
    ]
  },
  "system": {
    "os": "linux",
    "platform": "ubuntu",
    "platform_version": "22.04",
    "kernel_version": "5.15.0-91-generic",
    "uptime_seconds": 1314120
  }
}
```

### GET /health

Health check simples.

**Resposta:**
```json
{
  "status": "healthy",
  "timestamp": "2026-01-18T21:45:00-04:00"
}
```

### GET /metrics/history

Retorna histórico de métricas.

**Query Parameters:**
- `hours` (opcional): Número de horas de histórico (padrão: 24)

**Exemplo:**
```bash
curl http://100.64.1.10:8080/metrics/history?hours=48
```

**Resposta:**
```json
{
  "hours": 48,
  "count": 5760,
  "metrics": [
    {
      "timestamp": "2026-01-18T21:45:00-04:00",
      "hostname": "server-prod",
      "cpu": { ... },
      "memory": { ... },
      "disk": [ ... ],
      "network": { ... },
      "system": { ... }
    },
    ...
  ]
}
```

## Segurança

Todos os endpoints (exceto `/health`) validam que a requisição vem da rede Tailscale (100.64.0.0/10). Requisições de outros IPs retornam `403 Forbidden`.

## Exemplos com curl

```bash
# Status atual
curl http://100.64.1.10:8080/status | jq

# Health check
curl http://100.64.1.10:8080/health

# Histórico das últimas 12 horas
curl http://100.64.1.10:8080/metrics/history?hours=12 | jq

# Apenas CPU usage do histórico
curl http://100.64.1.10:8080/metrics/history?hours=1 | jq '.metrics[].cpu.usage_percent'
```

## Códigos de Status HTTP

- `200 OK`: Requisição bem-sucedida
- `403 Forbidden`: IP não é da rede Tailscale
- `500 Internal Server Error`: Erro ao coletar métricas
