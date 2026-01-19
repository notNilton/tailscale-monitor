package tailscale

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

// GetTailscaleIP retorna o IP Tailscale local (100.x.y.z)
func GetTailscaleIP() (string, error) {
	// Tenta obter via comando tailscale
	cmd := exec.Command("tailscale", "ip", "-4")
	output, err := cmd.Output()
	if err == nil {
		ip := strings.TrimSpace(string(output))
		if ip != "" {
			return ip, nil
		}
	}

	// Fallback: procura interface com IP 100.x.y.z
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}

			// Verifica se é um IP Tailscale (100.64.0.0/10)
			if ip[0] == 100 && (ip[1]&0xC0) == 64 {
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("tailscale IP not found")
}

// PeerInfo representa informações de um peer Tailscale
type PeerInfo struct {
	Hostname string
	IP       string
	Online   bool
}

// GetPeers retorna lista de peers Tailscale
func GetPeers() ([]PeerInfo, error) {
	cmd := exec.Command("tailscale", "status", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tailscale status: %w", err)
	}

	// Parse simples do JSON (poderia usar encoding/json para mais robustez)
	// Por simplicidade, vamos usar o formato texto
	cmd = exec.Command("tailscale", "status")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tailscale status: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var peers []PeerInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Formato: IP hostname status
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		ip := fields[0]
		hostname := fields[1]
		online := !strings.Contains(line, "offline")

		// Ignora linha de cabeçalho
		if ip == "#" || !strings.HasPrefix(ip, "100.") {
			continue
		}

		peers = append(peers, PeerInfo{
			Hostname: hostname,
			IP:       ip,
			Online:   online,
		})
	}

	return peers, nil
}

// IsTailscaleIP verifica se um IP é da rede Tailscale
func IsTailscaleIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	ip = ip.To4()
	if ip == nil {
		return false
	}

	// Verifica se é 100.64.0.0/10
	return ip[0] == 100 && (ip[1]&0xC0) == 64
}
