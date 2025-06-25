package core

import (
	"context"
	"runtime"
	"sync"
	"time"
	
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
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
} 