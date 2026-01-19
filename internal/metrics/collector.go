package metrics

import (
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Collector coleta métricas do sistema
type Collector struct {
	hostname string
}

// NewCollector cria um novo coletor de métricas
func NewCollector() (*Collector, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	return &Collector{
		hostname: hostname,
	}, nil
}

// Collect coleta todas as métricas do sistema
func (c *Collector) Collect() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
		Hostname:  c.hostname,
	}

	var err error

	// CPU
	metrics.CPU, err = c.collectCPU()
	if err != nil {
		return nil, fmt.Errorf("failed to collect CPU metrics: %w", err)
	}

	// Memory
	metrics.Memory, err = c.collectMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to collect memory metrics: %w", err)
	}

	// Disk
	metrics.Disk, err = c.collectDisk()
	if err != nil {
		return nil, fmt.Errorf("failed to collect disk metrics: %w", err)
	}

	// Network
	metrics.Network, err = c.collectNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to collect network metrics: %w", err)
	}

	// System
	metrics.System, err = c.collectSystem()
	if err != nil {
		return nil, fmt.Errorf("failed to collect system metrics: %w", err)
	}

	return metrics, nil
}

func (c *Collector) collectCPU() (CPUMetrics, error) {
	cpuMetrics := CPUMetrics{}

	// CPU usage
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return cpuMetrics, err
	}
	if len(percentages) > 0 {
		cpuMetrics.UsagePercent = percentages[0]
	}

	// Load average
	loadAvg, err := load.Avg()
	if err == nil {
		cpuMetrics.LoadAverage = []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15}
	}

	// CPU cores
	cores, err := cpu.Counts(true)
	if err == nil {
		cpuMetrics.Cores = cores
	}

	return cpuMetrics, nil
}

func (c *Collector) collectMemory() (MemoryMetrics, error) {
	memMetrics := MemoryMetrics{}

	// Virtual memory
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return memMetrics, err
	}

	memMetrics.TotalBytes = vmem.Total
	memMetrics.UsedBytes = vmem.Used
	memMetrics.AvailableBytes = vmem.Available
	memMetrics.UsagePercent = vmem.UsedPercent

	// Swap
	swap, err := mem.SwapMemory()
	if err == nil {
		memMetrics.SwapTotal = swap.Total
		memMetrics.SwapUsed = swap.Used
	}

	return memMetrics, nil
}

func (c *Collector) collectDisk() ([]DiskMetrics, error) {
	var diskMetrics []DiskMetrics

	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue // Skip partitions we can't read
		}

		diskMetrics = append(diskMetrics, DiskMetrics{
			Device:       partition.Device,
			MountPoint:   partition.Mountpoint,
			TotalBytes:   usage.Total,
			UsedBytes:    usage.Used,
			FreeBytes:    usage.Free,
			UsagePercent: usage.UsedPercent,
		})
	}

	return diskMetrics, nil
}

func (c *Collector) collectNetwork() (NetworkMetrics, error) {
	netMetrics := NetworkMetrics{}

	// Network I/O counters
	ioCounters, err := net.IOCounters(false)
	if err == nil && len(ioCounters) > 0 {
		netMetrics.BytesSent = ioCounters[0].BytesSent
		netMetrics.BytesRecv = ioCounters[0].BytesRecv
		netMetrics.PacketsSent = ioCounters[0].PacketsSent
		netMetrics.PacketsRecv = ioCounters[0].PacketsRecv
	}

	// Network interfaces
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			netIface := NetworkInterface{
				Name:      iface.Name,
				Addresses: []string{},
			}

			for _, addr := range iface.Addrs {
				netIface.Addresses = append(netIface.Addresses, addr.Addr)
			}

			// Check if interface is up
			if len(iface.Flags) > 0 {
				netIface.IsUp = iface.Flags[0] == "up"
			}

			netMetrics.Interfaces = append(netMetrics.Interfaces, netIface)
		}
	}

	return netMetrics, nil
}

func (c *Collector) collectSystem() (SystemInfo, error) {
	sysInfo := SystemInfo{}

	info, err := host.Info()
	if err != nil {
		return sysInfo, err
	}

	sysInfo.OS = info.OS
	sysInfo.Platform = info.Platform
	sysInfo.PlatformVersion = info.PlatformVersion
	sysInfo.KernelVersion = info.KernelVersion
	sysInfo.UptimeSeconds = info.Uptime

	return sysInfo, nil
}
