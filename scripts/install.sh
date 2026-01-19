#!/bin/bash
set -e

# Script de instalação do Tailscale Network Monitor Agent

echo "=== Tailscale Network Monitor - Instalação ==="
echo

# Verifica se está rodando como root
if [ "$EUID" -ne 0 ]; then 
    echo "Por favor, execute como root (use sudo)"
    exit 1
fi

# Verifica se Tailscale está instalado
if ! command -v tailscale &> /dev/null; then
    echo "Erro: Tailscale não está instalado"
    echo "Instale o Tailscale primeiro: https://tailscale.com/download"
    exit 1
fi

# Verifica se Go está instalado
if ! command -v go &> /dev/null; then
    echo "Erro: Go não está instalado"
    echo "Instale Go 1.21 ou superior: https://go.dev/dl/"
    exit 1
fi

# Diretórios
INSTALL_DIR="/opt/tailscale-monitor"
DATA_DIR="/var/lib/tailscale-monitor"
CONFIG_FILE="/etc/tailscale-monitor/config.yaml"

echo "Criando diretórios..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$DATA_DIR"
mkdir -p "$(dirname $CONFIG_FILE)"

# Build dos binários
echo "Compilando binários..."
go build -o "$INSTALL_DIR/agent" ./cmd/agent
go build -o "$INSTALL_DIR/client" ./cmd/client

chmod +x "$INSTALL_DIR/agent"
chmod +x "$INSTALL_DIR/client"

# Cria link simbólico para o cliente
ln -sf "$INSTALL_DIR/client" /usr/local/bin/tsmon

# Cria configuração se não existir
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Criando configuração padrão..."
    cat > "$CONFIG_FILE" <<EOF
server:
  port: 8080
  tailscale_only: true

storage:
  path: $DATA_DIR/metrics.db
  retention_days: 30

metrics:
  collection_interval: 30s
EOF
fi

# Cria serviço systemd
echo "Criando serviço systemd..."
cat > /etc/systemd/system/tailscale-monitor.service <<EOF
[Unit]
Description=Tailscale Network Monitor Agent
After=network.target tailscaled.service
Requires=tailscaled.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/agent
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Recarrega systemd
systemctl daemon-reload

echo
echo "=== Instalação concluída! ==="
echo
echo "Para iniciar o serviço:"
echo "  sudo systemctl start tailscale-monitor"
echo
echo "Para habilitar na inicialização:"
echo "  sudo systemctl enable tailscale-monitor"
echo
echo "Para verificar status:"
echo "  sudo systemctl status tailscale-monitor"
echo
echo "Para usar o cliente:"
echo "  tsmon list                    # Lista peers"
echo "  tsmon status <hostname>       # Status de um peer"
echo "  tsmon status --all            # Status de todos"
echo
echo "Configuração: $CONFIG_FILE"
echo "Dados: $DATA_DIR"
echo
