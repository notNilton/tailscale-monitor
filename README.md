# Tailscale Network Monitor

Sistema minimalista e descentralizado de monitoramento para redes Tailscale. Cada dispositivo expõe suas próprias métricas via HTTP (apenas na interface Tailscale) e mantém histórico local em SQLite.

## 🎯 Características

- **Descentralizado**: Sem servidor central, cada nó é autônomo
- **Seguro**: Escuta apenas na interface Tailscale (100.x.y.z)
- **Leve**: Binário único em Go, baixo consumo de recursos
- **Persistente**: Histórico local em SQLite
- **Simples**: CLI intuitivo para consultar qualquer nó

## 📊 Métricas Coletadas

- **CPU**: Uso percentual, load average, número de cores
- **Memória**: Total, usado, disponível, swap
- **Disco**: Uso por partição, I/O
- **Rede**: Bytes enviados/recebidos, interfaces ativas
- **Sistema**: Uptime, hostname, OS, versão do kernel

## 🚀 Instalação

### Opção 1: Docker (Recomendado)

```bash
# Clone o repositório
git clone https://github.com/nilbyte-studios/network-infra.git
cd network-infra

# Crie o arquivo de configuração
cp config.example.yaml config.yaml

# Inicie com Docker Compose
docker-compose up -d
```

### Opção 2: Binário (Linux)

```bash
# Clone o repositório
git clone https://github.com/nilbyte-studios/network-infra.git
cd network-infra

# Execute o script de instalação
sudo ./scripts/install.sh

# Inicie o serviço
sudo systemctl start tailscale-monitor
sudo systemctl enable tailscale-monitor
```

### Opção 3: Build Manual

```bash
# Clone o repositório
git clone https://github.com/nilbyte-studios/network-infra.git
cd network-infra

# Instale dependências
go mod download

# Build
go build -o bin/agent ./cmd/agent
go build -o bin/client ./cmd/client

# Execute
./bin/agent
```

## 📖 Uso

### Cliente CLI

Após a instalação, use o comando `tsmon` (ou `./bin/client` se build manual):

```bash
# Listar todos os peers Tailscale
tsmon list

# Consultar status de um peer específico
tsmon status hostname-do-peer

# Consultar todos os peers online
tsmon status --all

# Output em JSON
tsmon status hostname-do-peer --json
```

### API HTTP

Cada nó expõe os seguintes endpoints (apenas acessíveis via Tailscale):

```bash
# Status atual (métricas em tempo real)
curl http://100.x.y.z:8080/status

# Health check
curl http://100.x.y.z:8080/health

# Histórico (últimas 24 horas por padrão)
curl http://100.x.y.z:8080/metrics/history

# Histórico personalizado (últimas 48 horas)
curl http://100.x.y.z:8080/metrics/history?hours=48
```

## ⚙️ Configuração

Edite `config.yaml`:

```yaml
server:
  port: 8080                # Porta do servidor HTTP
  tailscale_only: true      # Escutar apenas no IP Tailscale

storage:
  path: /data/metrics.db    # Caminho do banco SQLite
  retention_days: 30        # Retenção de dados (dias)

metrics:
  collection_interval: 30s  # Intervalo de coleta
```

## 🔒 Segurança

- **Binding Tailscale**: O servidor escuta apenas no IP Tailscale (100.x.y.z), garantindo que apenas dispositivos na VPN possam acessar
- **Validação de IP**: Todos os endpoints validam que a requisição vem da rede Tailscale
- **Sem autenticação externa**: A segurança é garantida pela própria rede Tailscale

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────┐
│         Dispositivo A (100.x.y.1)       │
│  ┌────────────┐  ┌──────────────────┐   │
│  │  Collector │→ │  HTTP Server     │   │
│  └────────────┘  │  :8080           │   │
│                  └──────────────────┘   │
│                  ┌──────────────────┐   │
│                  │  SQLite DB       │   │
│                  └──────────────────┘   │
└─────────────────────────────────────────┘
                     ↑
                     │ HTTP GET /status
                     │
┌─────────────────────────────────────────┐
│         Admin (100.x.y.3)               │
│  ┌────────────────────────────────────┐ │
│  │  Cliente CLI (tsmon)               │ │
│  │  - Descobre peers via tailscale   │ │
│  │  - Consulta diretamente cada nó   │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

## 📁 Estrutura do Projeto

```
.
├── cmd/
│   ├── agent/          # Agente de monitoramento
│   └── client/         # Cliente CLI
├── internal/
│   ├── config/         # Configuração
│   ├── metrics/        # Coleta de métricas
│   ├── server/         # Servidor HTTP
│   ├── storage/        # Persistência SQLite
│   └── tailscale/      # Integração Tailscale
├── scripts/
│   └── install.sh      # Script de instalação
├── Dockerfile
├── docker-compose.yml
└── config.example.yaml
```

## 🛠️ Desenvolvimento

```bash
# Instalar dependências
go mod download

# Executar testes
go test ./internal/...

# Build
go build -o bin/agent ./cmd/agent
go build -o bin/client ./cmd/client

# Executar localmente
./bin/agent
```

## 📝 Exemplos

### Monitorar todos os dispositivos

```bash
# Lista todos os peers
tsmon list

# Output:
# HOSTNAME        IP              STATUS
# --------        --              ------
# server-prod     100.64.1.10     online
# laptop-dev      100.64.1.20     online
# raspberry-pi    100.64.1.30     offline
```

### Consultar status específico

```bash
tsmon status server-prod

# Output:
# Hostname: server-prod
# Timestamp: 2026-01-18T21:45:00-04:00
# Uptime: 15d 3h 42m
#
# CPU Usage: 23.45%
# CPU Cores: 4
# Load Average: 1.20, 1.15, 1.10
#
# Memory: 3.2 GiB / 8.0 GiB (40.00%)
# Swap: 0 B / 2.0 GiB
#
# Disk Usage:
#   /: 45.2 GiB / 100.0 GiB (45.20%)
#   /data: 120.5 GiB / 500.0 GiB (24.10%)
#
# Network: ↓ 1.2 TiB  ↑ 450.3 GiB
```

### Consultar histórico via API

```bash
curl http://100.64.1.10:8080/metrics/history?hours=1 | jq
```

## 🤝 Contribuindo

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou pull requests.

## 📄 Licença

MIT License - veja LICENSE para detalhes.

## 🔗 Links Úteis

- [Tailscale](https://tailscale.com/)
- [Documentação Go](https://go.dev/doc/)
- [gopsutil](https://github.com/shirou/gopsutil)
