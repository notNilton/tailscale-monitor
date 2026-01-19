package metrics

import "time"

// SystemMetrics representa todas as métricas coletadas do sistema
type SystemMetrics struct {
	Timestamp time.Time      `json:"timestamp"`
	Hostname  string         `json:"hostname"`
	CPU       CPUMetrics     `json:"cpu"`
	Memory    MemoryMetrics  `json:"memory"`
	Disk      []DiskMetrics  `json:"disk"`
	Network   NetworkMetrics `json:"network"`
	System    SystemInfo     `json:"system"`
}

// CPUMetrics representa métricas de CPU
type CPUMetrics struct {
	UsagePercent float64   `json:"usage_percent"`
	LoadAverage  []float64 `json:"load_average"` // 1, 5, 15 minutos
	Cores        int       `json:"cores"`
}

// MemoryMetrics representa métricas de memória
type MemoryMetrics struct {
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
	SwapTotal      uint64  `json:"swap_total_bytes"`
	SwapUsed       uint64  `json:"swap_used_bytes"`
}

// DiskMetrics representa métricas de disco por partição
type DiskMetrics struct {
	Device       string  `json:"device"`
	MountPoint   string  `json:"mount_point"`
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	FreeBytes    uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

// NetworkMetrics representa métricas de rede
type NetworkMetrics struct {
	BytesSent   uint64             `json:"bytes_sent"`
	BytesRecv   uint64             `json:"bytes_recv"`
	PacketsSent uint64             `json:"packets_sent"`
	PacketsRecv uint64             `json:"packets_recv"`
	Interfaces  []NetworkInterface `json:"interfaces"`
}

// NetworkInterface representa uma interface de rede
type NetworkInterface struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	IsUp      bool     `json:"is_up"`
}

// SystemInfo representa informações do sistema
type SystemInfo struct {
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	UptimeSeconds   uint64 `json:"uptime_seconds"`
}
