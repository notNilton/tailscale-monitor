package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/nilbyte-studios/network-infra/internal/metrics"
	"github.com/nilbyte-studios/network-infra/internal/tailscale"
)

const (
	defaultPort    = 8080
	defaultTimeout = 5 * time.Second
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		cmdList()
	case "status":
		cmdStatus()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Tailscale Network Monitor Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client list                    List all Tailscale peers")
	fmt.Println("  client status <hostname>       Get status of a specific peer")
	fmt.Println("  client status --all            Get status of all peers")
	fmt.Println("  client status --json           Output in JSON format")
	fmt.Println()
}

func cmdList() {
	peers, err := tailscale.GetPeers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting peers: %v\n", err)
		os.Exit(1)
	}

	if len(peers) == 0 {
		fmt.Println("No peers found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "HOSTNAME\tIP\tSTATUS")
	fmt.Fprintln(w, "--------\t--\t------")

	for _, peer := range peers {
		status := "online"
		if !peer.Online {
			status = "offline"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", peer.Hostname, peer.IP, status)
	}

	w.Flush()
}

func cmdStatus() {
	var (
		all        bool
		jsonOutput bool
		port       int
	)

	fs := flag.NewFlagSet("status", flag.ExitOnError)
	fs.BoolVar(&all, "all", false, "Get status of all peers")
	fs.BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	fs.IntVar(&port, "port", defaultPort, "Port to connect to")
	fs.Parse(os.Args[2:])

	if all {
		cmdStatusAll(port, jsonOutput)
		return
	}

	if fs.NArg() < 1 {
		fmt.Println("Error: hostname required")
		fmt.Println("Usage: client status <hostname> [--json] [--port PORT]")
		os.Exit(1)
	}

	hostname := fs.Arg(0)
	cmdStatusSingle(hostname, port, jsonOutput)
}

func cmdStatusSingle(hostname string, port int, jsonOutput bool) {
	// Encontra o peer pelo hostname
	peers, err := tailscale.GetPeers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting peers: %v\n", err)
		os.Exit(1)
	}

	var targetIP string
	for _, peer := range peers {
		if peer.Hostname == hostname {
			targetIP = peer.IP
			break
		}
	}

	if targetIP == "" {
		fmt.Fprintf(os.Stderr, "Peer not found: %s\n", hostname)
		os.Exit(1)
	}

	// Busca status
	m, err := fetchMetrics(targetIP, port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching metrics from %s: %v\n", hostname, err)
		os.Exit(1)
	}

	if jsonOutput {
		printJSON(m)
	} else {
		printMetrics(m)
	}
}

func cmdStatusAll(port int, jsonOutput bool) {
	peers, err := tailscale.GetPeers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting peers: %v\n", err)
		os.Exit(1)
	}

	if len(peers) == 0 {
		fmt.Println("No peers found")
		return
	}

	results := make(map[string]*metrics.SystemMetrics)

	for _, peer := range peers {
		if !peer.Online {
			continue
		}

		m, err := fetchMetrics(peer.IP, port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch from %s: %v\n", peer.Hostname, err)
			continue
		}

		results[peer.Hostname] = m
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
	} else {
		for hostname, m := range results {
			fmt.Printf("\n=== %s ===\n", hostname)
			printMetrics(m)
		}
	}
}

func fetchMetrics(ip string, port int) (*metrics.SystemMetrics, error) {
	url := fmt.Sprintf("http://%s:%d/status", ip, port)

	client := &http.Client{
		Timeout: defaultTimeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var m metrics.SystemMetrics
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func printJSON(m *metrics.SystemMetrics) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func printMetrics(m *metrics.SystemMetrics) {
	fmt.Printf("Hostname: %s\n", m.Hostname)
	fmt.Printf("Timestamp: %s\n", m.Timestamp.Format(time.RFC3339))
	fmt.Printf("Uptime: %s\n", formatDuration(time.Duration(m.System.UptimeSeconds)*time.Second))
	fmt.Println()

	// CPU
	fmt.Printf("CPU Usage: %.2f%%\n", m.CPU.UsagePercent)
	fmt.Printf("CPU Cores: %d\n", m.CPU.Cores)
	if len(m.CPU.LoadAverage) >= 3 {
		fmt.Printf("Load Average: %.2f, %.2f, %.2f\n", m.CPU.LoadAverage[0], m.CPU.LoadAverage[1], m.CPU.LoadAverage[2])
	}
	fmt.Println()

	// Memory
	fmt.Printf("Memory: %s / %s (%.2f%%)\n",
		formatBytes(m.Memory.UsedBytes),
		formatBytes(m.Memory.TotalBytes),
		m.Memory.UsagePercent)
	if m.Memory.SwapTotal > 0 {
		fmt.Printf("Swap: %s / %s\n",
			formatBytes(m.Memory.SwapUsed),
			formatBytes(m.Memory.SwapTotal))
	}
	fmt.Println()

	// Disk
	if len(m.Disk) > 0 {
		fmt.Println("Disk Usage:")
		for _, disk := range m.Disk {
			fmt.Printf("  %s: %s / %s (%.2f%%)\n",
				disk.MountPoint,
				formatBytes(disk.UsedBytes),
				formatBytes(disk.TotalBytes),
				disk.UsagePercent)
		}
		fmt.Println()
	}

	// Network
	fmt.Printf("Network: ↓ %s  ↑ %s\n",
		formatBytes(m.Network.BytesRecv),
		formatBytes(m.Network.BytesSent))
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
