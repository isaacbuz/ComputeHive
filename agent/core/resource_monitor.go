package core

import (
<<<<<<< HEAD
	"fmt"
	"sync"
	"time"

=======
	"context"
	"runtime"
	"sync"
	"time"
	
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
<<<<<<< HEAD
	"go.uber.org/zap"
)

// ResourceMonitor monitors system resource usage
type ResourceMonitor struct {
	logger         *zap.Logger
	mu             sync.RWMutex
	lastCPUTimes   []cpu.TimesStat
	lastNetIOStats []net.IOCountersStat
	lastCheckTime  time.Time
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(logger *zap.Logger) *ResourceMonitor {
	return &ResourceMonitor{
		logger:        logger,
		lastCheckTime: time.Now(),
	}
}

// GetCurrentUsage returns current resource usage
func (rm *ResourceMonitor) GetCurrentUsage() (*ResourceUsage, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	usage := &ResourceUsage{}

	// CPU usage
	cpuPercent, err := rm.getCPUUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	usage.CPUPercent = cpuPercent

	// Memory usage
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}
	usage.MemoryPercent = memStats.UsedPercent
	usage.MemoryUsedMB = memStats.Used / (1024 * 1024)

	// Disk usage
	diskStats, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk stats: %w", err)
	}
	usage.DiskPercent = diskStats.UsedPercent
	usage.DiskUsedMB = diskStats.Used / (1024 * 1024)

	// Network usage
	netIn, netOut, err := rm.getNetworkUsage()
	if err != nil {
		rm.logger.Warn("Failed to get network usage", zap.Error(err))
		// Don't fail the whole operation if network stats fail
	} else {
		usage.NetworkInMbps = netIn
		usage.NetworkOutMbps = netOut
	}

	return usage, nil
}

// getCPUUsage calculates CPU usage percentage
func (rm *ResourceMonitor) getCPUUsage() (float64, error) {
	// Get current CPU times
	currentTimes, err := cpu.Times(false)
	if err != nil {
		return 0, err
	}

	// If this is the first check, just store the times
	if rm.lastCPUTimes == nil {
		rm.lastCPUTimes = currentTimes
		// Use instant CPU percent for first reading
		percentages, err := cpu.Percent(100*time.Millisecond, false)
		if err != nil {
			return 0, err
		}
		if len(percentages) > 0 {
			return percentages[0], nil
		}
		return 0, nil
	}

	// Calculate CPU usage based on time differences
	totalDelta := float64(0)
	idleDelta := float64(0)

	for i, current := range currentTimes {
		if i < len(rm.lastCPUTimes) {
			last := rm.lastCPUTimes[i]
			
			total := (current.User - last.User) +
				(current.System - last.System) +
				(current.Idle - last.Idle) +
				(current.Nice - last.Nice) +
				(current.Iowait - last.Iowait) +
				(current.Irq - last.Irq) +
				(current.Softirq - last.Softirq) +
				(current.Steal - last.Steal)
			
			totalDelta += total
			idleDelta += (current.Idle - last.Idle)
		}
	}

	rm.lastCPUTimes = currentTimes

	if totalDelta == 0 {
		return 0, nil
	}

	// Calculate percentage of non-idle time
	usage := 100.0 * (1.0 - idleDelta/totalDelta)
	
	// Clamp to valid range
	if usage < 0 {
		usage = 0
	} else if usage > 100 {
		usage = 100
	}

	return usage, nil
}

// getNetworkUsage calculates network usage in Mbps
func (rm *ResourceMonitor) getNetworkUsage() (inMbps, outMbps float64, err error) {
	// Get current network stats
	currentStats, err := net.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}

	if len(currentStats) == 0 {
		return 0, 0, fmt.Errorf("no network interfaces found")
	}

	// If this is the first check, just store the stats
	if rm.lastNetIOStats == nil {
		rm.lastNetIOStats = currentStats
		return 0, 0, nil
	}

	// Calculate time delta
	now := time.Now()
	timeDelta := now.Sub(rm.lastCheckTime).Seconds()
	rm.lastCheckTime = now

	if timeDelta == 0 {
		return 0, 0, nil
	}

	// Calculate bytes per second
	var bytesSentDelta, bytesRecvDelta uint64
	
	for i, current := range currentStats {
		if i < len(rm.lastNetIOStats) {
			last := rm.lastNetIOStats[i]
			
			// Check for counter wrap-around
			if current.BytesSent >= last.BytesSent {
				bytesSentDelta += current.BytesSent - last.BytesSent
			}
			if current.BytesRecv >= last.BytesRecv {
				bytesRecvDelta += current.BytesRecv - last.BytesRecv
			}
		}
	}

	rm.lastNetIOStats = currentStats

	// Convert to Mbps (megabits per second)
	inMbps = float64(bytesRecvDelta) * 8 / (timeDelta * 1000000)
	outMbps = float64(bytesSentDelta) * 8 / (timeDelta * 1000000)

	return inMbps, outMbps, nil
}

// GetDetailedMetrics returns detailed system metrics
func (rm *ResourceMonitor) GetDetailedMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// CPU metrics
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		metrics["cpu_model"] = cpuInfo[0].ModelName
		metrics["cpu_cores"] = len(cpuInfo)
		metrics["cpu_mhz"] = cpuInfo[0].Mhz
	}

	// Per-CPU usage
	perCPU, err := cpu.Percent(100*time.Millisecond, true)
	if err == nil {
		metrics["cpu_per_core"] = perCPU
	}

	// Memory metrics
	memStats, err := mem.VirtualMemory()
	if err == nil {
		metrics["memory_total_mb"] = memStats.Total / (1024 * 1024)
		metrics["memory_available_mb"] = memStats.Available / (1024 * 1024)
		metrics["memory_used_mb"] = memStats.Used / (1024 * 1024)
		metrics["memory_cached_mb"] = memStats.Cached / (1024 * 1024)
		metrics["memory_buffers_mb"] = memStats.Buffers / (1024 * 1024)
	}

	// Swap metrics
	swapStats, err := mem.SwapMemory()
	if err == nil {
		metrics["swap_total_mb"] = swapStats.Total / (1024 * 1024)
		metrics["swap_used_mb"] = swapStats.Used / (1024 * 1024)
		metrics["swap_percent"] = swapStats.UsedPercent
	}

	// Disk metrics for all partitions
	partitions, err := disk.Partitions(false)
	if err == nil {
		diskMetrics := make([]map[string]interface{}, 0)
		for _, partition := range partitions {
			usage, err := disk.Usage(partition.Mountpoint)
			if err == nil {
				diskMetrics = append(diskMetrics, map[string]interface{}{
					"mountpoint":   partition.Mountpoint,
					"device":       partition.Device,
					"fstype":       partition.Fstype,
					"total_mb":     usage.Total / (1024 * 1024),
					"used_mb":      usage.Used / (1024 * 1024),
					"free_mb":      usage.Free / (1024 * 1024),
					"used_percent": usage.UsedPercent,
				})
			}
		}
		metrics["disks"] = diskMetrics
	}

	// Network interface metrics
	interfaces, err := net.Interfaces()
	if err == nil {
		netMetrics := make([]map[string]interface{}, 0)
		for _, iface := range interfaces {
			// Skip loopback interfaces
			if iface.Name == "lo" || iface.Name == "lo0" {
				continue
			}
			
			netMetrics = append(netMetrics, map[string]interface{}{
				"name":          iface.Name,
				"mtu":           iface.MTU,
				"hardwareaddr":  iface.HardwareAddr,
				"flags":         iface.Flags,
			})
		}
		metrics["network_interfaces"] = netMetrics
	}

	// Load average (Unix-like systems)
	loadAvg, err := getLoadAverage()
	if err == nil {
		metrics["load_average"] = loadAvg
	}

	return metrics, nil
}

// StartContinuousMonitoring starts continuous resource monitoring
func (rm *ResourceMonitor) StartContinuousMonitoring(interval time.Duration, callback func(*ResourceUsage)) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			usage, err := rm.GetCurrentUsage()
			if err != nil {
				rm.logger.Error("Failed to get resource usage", zap.Error(err))
				continue
			}
			
			if callback != nil {
				callback(usage)
			}
		}
	}()
=======
)

// ResourceMonitor monitors system resources
type ResourceMonitor struct {
	resources *Resources
	mu        sync.RWMutex
	interval  time.Duration
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor() *ResourceMonitor {
	return &ResourceMonitor{
		resources: &Resources{},
		interval:  5 * time.Second,
	}
}

// Start begins monitoring resources
func (rm *ResourceMonitor) Start(ctx context.Context) {
	// Initial resource scan
	rm.updateResources()
	
	ticker := time.NewTicker(rm.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			rm.updateResources()
		case <-ctx.Done():
			return
		}
	}
}

// GetResources returns the current resource snapshot
func (rm *ResourceMonitor) GetResources() *Resources {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Return a copy to prevent race conditions
	return &Resources{
		CPU:     rm.resources.CPU,
		Memory:  rm.resources.Memory,
		GPUs:    append([]GPUInfo{}, rm.resources.GPUs...),
		Storage: rm.resources.Storage,
		Network: rm.resources.Network,
	}
}

// updateResources updates the resource information
func (rm *ResourceMonitor) updateResources() {
	resources := &Resources{}
	
	// Update CPU info
	resources.CPU = rm.getCPUInfo()
	
	// Update memory info
	resources.Memory = rm.getMemoryInfo()
	
	// Update storage info
	resources.Storage = rm.getStorageInfo()
	
	// Update network info
	resources.Network = rm.getNetworkInfo()
	
	// Update GPU info (platform-specific)
	resources.GPUs = rm.getGPUInfo()
	
	rm.mu.Lock()
	rm.resources = resources
	rm.mu.Unlock()
}

// getCPUInfo retrieves CPU information
func (rm *ResourceMonitor) getCPUInfo() CPUInfo {
	info := CPUInfo{
		Cores:   runtime.NumCPU(),
		Threads: runtime.GOMAXPROCS(0),
	}
	
	// Get CPU info
	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		info.Model = cpuInfo[0].ModelName
		info.FrequencyHz = int64(cpuInfo[0].Mhz * 1000000)
	}
	
	// Get CPU usage
	if usage, err := cpu.Percent(time.Second, false); err == nil && len(usage) > 0 {
		info.Usage = usage[0]
	}
	
	return info
}

// getMemoryInfo retrieves memory information
func (rm *ResourceMonitor) getMemoryInfo() MemoryInfo {
	info := MemoryInfo{}
	
	if vmStat, err := mem.VirtualMemory(); err == nil {
		info.Total = int64(vmStat.Total)
		info.Available = int64(vmStat.Available)
		info.Used = int64(vmStat.Used)
		info.Usage = vmStat.UsedPercent
	}
	
	return info
}

// getStorageInfo retrieves storage information
func (rm *ResourceMonitor) getStorageInfo() StorageInfo {
	info := StorageInfo{}
	
	if usage, err := disk.Usage("/"); err == nil {
		info.Total = int64(usage.Total)
		info.Available = int64(usage.Free)
		info.Used = int64(usage.Used)
		info.Usage = usage.UsedPercent
	}
	
	return info
}

// getNetworkInfo retrieves network information
func (rm *ResourceMonitor) getNetworkInfo() NetworkInfo {
	info := NetworkInfo{
		Interfaces: []NetworkInterface{},
		Bandwidth:  1000, // Default 1Gbps
	}
	
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			if len(iface.Addrs) > 0 {
				ni := NetworkInterface{
					Name: iface.Name,
					Type: "ethernet",
				}
				
				// Get first non-loopback IP
				for _, addr := range iface.Addrs {
					ip := addr.Addr
					if ip != "" && !isLoopback(ip) {
						ni.IP = ip
						break
					}
				}
				
				if ni.IP != "" {
					info.Interfaces = append(info.Interfaces, ni)
				}
			}
		}
	}
	
	return info
}

// getGPUInfo retrieves GPU information (stub - implemented in platform-specific files)
func (rm *ResourceMonitor) getGPUInfo() []GPUInfo {
	// This is overridden in platform-specific implementations
	return detectGPUs()
}

// isLoopback checks if an IP address is a loopback address
func isLoopback(ip string) bool {
	return len(ip) >= 3 && ip[:3] == "127"
}

// MonitorJob monitors resources for a specific job
func (rm *ResourceMonitor) MonitorJob(ctx context.Context, jobID string) *JobMetrics {
	metrics := &JobMetrics{}
	startTime := time.Now()
	
	// Monitor resources during job execution
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	var maxMemory int64
	var totalCPUTime time.Duration
	
	for {
		select {
		case <-ticker.C:
			// Update memory peak
			if mem := rm.resources.Memory.Used; mem > maxMemory {
				maxMemory = mem
			}
			
			// Estimate CPU time (simplified)
			totalCPUTime += time.Second * time.Duration(rm.resources.CPU.Usage/100)
			
		case <-ctx.Done():
			metrics.CPUTime = totalCPUTime
			metrics.MemoryPeakMB = maxMemory / (1024 * 1024)
			return metrics
		}
	}
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
} 