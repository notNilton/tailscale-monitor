package tailscale

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
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
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Online   bool   `json:"online"`
}

// APIClient cliente para API do Tailscale
type APIClient struct {
	apiKey  string
	tailnet string
	client  *http.Client
}

// NewAPIClient cria um novo cliente da API Tailscale
func NewAPIClient(apiKey, tailnet string) *APIClient {
	return &APIClient{
		apiKey:  apiKey,
		tailnet: tailnet,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// tailscaleDevice representa um dispositivo da API do Tailscale
type tailscaleDevice struct {
	Name      string   `json:"name"`
	Hostname  string   `json:"hostname"`
	Addresses []string `json:"addresses"`
	Online    bool     `json:"online"`
}

// tailscaleDevicesResponse resposta da API de devices
type tailscaleDevicesResponse struct {
	Devices []tailscaleDevice `json:"devices"`
}

// GetPeersViaAPI busca peers usando a API do Tailscale
func (c *APIClient) GetPeersViaAPI() ([]PeerInfo, error) {
	if c.apiKey == "" || c.tailnet == "" {
		return nil, fmt.Errorf("tailscale API key and tailnet are required")
	}

	url := fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/devices", c.tailnet)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var devicesResp tailscaleDevicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&devicesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var peers []PeerInfo
	for _, device := range devicesResp.Devices {
		// Pega o primeiro endereço IPv4 Tailscale
		var ip string
		for _, addr := range device.Addresses {
			if strings.HasPrefix(addr, "100.") {
				// Remove CIDR notation se presente
				ip = strings.Split(addr, "/")[0]
				break
			}
		}

		if ip == "" {
			continue
		}

		hostname := device.Hostname
		if hostname == "" {
			hostname = device.Name
		}

		// Skip localhost entries
		if hostname == "localhost" {
			continue
		}

		peers = append(peers, PeerInfo{
			Hostname: hostname,
			IP:       ip,
			Online:   device.Online,
		})
	}

	return peers, nil
}

// GetPeers retorna lista de peers Tailscale (tenta CLI primeiro, depois API)
func GetPeers() ([]PeerInfo, error) {
	// Tenta via CLI primeiro
	peers, err := getPeersViaCLI()
	if err == nil && len(peers) > 0 {
		return peers, nil
	}

	// Se CLI falhar, retorna erro indicando que API é necessária
	return nil, fmt.Errorf("failed to get tailscale status: %w (hint: configure TAILSCALE_API_KEY and TAILSCALE_TAILNET environment variables)", err)
}

// GetPeersWithAPI retorna peers usando API ou CLI baseado na config
func GetPeersWithAPI(apiKey, tailnet string, useCLI bool) ([]PeerInfo, error) {
	if !useCLI && apiKey != "" && tailnet != "" {
		client := NewAPIClient(apiKey, tailnet)
		return client.GetPeersViaAPI()
	}

	// Fallback para CLI
	return getPeersViaCLI()
}

// getPeersViaCLI busca peers via comando tailscale
func getPeersViaCLI() ([]PeerInfo, error) {
	cmd := exec.Command("tailscale", "status")
	output, err := cmd.Output()
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
