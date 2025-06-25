//go:build windows
// +build windows

package core

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	
	"github.com/google/uuid"
	"golang.org/x/sys/windows"
)

// GenerateAgentID generates a unique agent ID
func GenerateAgentID() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s-%s", hostname, uuid.New().String()[:8])
}

// GetPlatformInfo returns platform-specific information
func GetPlatformInfo() Platform {
	platform := Platform{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: getHostname(),
	}
	
	// Get Windows version
	platform.Version = getWindowsVersion()
	
	// Check container runtime
	platform.ContainerRuntime = detectContainerRuntime()
	
	return platform
}

// GetPlatformCapabilities returns platform-specific capabilities
func GetPlatformCapabilities() []string {
	caps := []string{}
	
	// Check for Docker Desktop
	if _, err := exec.LookPath("docker"); err == nil {
		caps = append(caps, "docker")
	}
	
	// Check for WSL2
	if isWSL2Available() {
		caps = append(caps, "wsl2")
	}
	
	// Check for Hyper-V
	if isHyperVAvailable() {
		caps = append(caps, "hyperv")
	}
	
	return caps
}

// detectGPUs detects available GPUs on Windows
func detectGPUs() []GPUInfo {
	var gpus []GPUInfo
	
	// Try NVIDIA GPUs
	if nvidiaGPUs := detectNVIDIAGPUs(); len(nvidiaGPUs) > 0 {
		gpus = append(gpus, nvidiaGPUs...)
	}
	
	// Try AMD GPUs (would need Windows-specific implementation)
	// For now, we'll use WMI to detect GPUs
	if wmiGPUs := detectGPUsViaWMI(); len(wmiGPUs) > 0 {
		gpus = append(gpus, wmiGPUs...)
	}
	
	return gpus
}

// detectNVIDIAGPUs detects NVIDIA GPUs using nvidia-smi on Windows
func detectNVIDIAGPUs() []GPUInfo {
	var gpus []GPUInfo
	
	// nvidia-smi is typically in C:\Program Files\NVIDIA Corporation\NVSMI\
	nvidiaSMIPath := `C:\Program Files\NVIDIA Corporation\NVSMI\nvidia-smi.exe`
	
	// Check if nvidia-smi exists
	if _, err := os.Stat(nvidiaSMIPath); err != nil {
		// Try in PATH
		nvidiaSMIPath = "nvidia-smi"
	}
	
	output, err := exec.Command(nvidiaSMIPath, "--query-gpu=index,name,memory.total,utilization.gpu,temperature.gpu,power.draw", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return gpus
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ", ")
		if len(parts) >= 6 {
			gpu := GPUInfo{
				ID:     parts[0],
				Model:  parts[1],
				Vendor: "NVIDIA",
			}
			
			// Parse memory (in MB)
			if _, err := fmt.Sscanf(parts[2], "%d", &gpu.MemoryMB); err != nil {
				gpu.MemoryMB = 0
			}
			
			// Parse usage
			if _, err := fmt.Sscanf(parts[3], "%f", &gpu.Usage); err != nil {
				gpu.Usage = 0
			}
			
			// Parse temperature
			if _, err := fmt.Sscanf(parts[4], "%f", &gpu.Temperature); err != nil {
				gpu.Temperature = 0
			}
			
			// Parse power
			if _, err := fmt.Sscanf(parts[5], "%f", &gpu.PowerWatts); err != nil {
				gpu.PowerWatts = 0
			}
			
			gpus = append(gpus, gpu)
		}
	}
	
	return gpus
}

// detectGPUsViaWMI uses WMI to detect GPUs on Windows
func detectGPUsViaWMI() []GPUInfo {
	var gpus []GPUInfo
	
	// Use WMIC to query video controllers
	output, err := exec.Command("wmic", "path", "win32_VideoController", "get", "Name,AdapterRAM", "/format:csv").Output()
	if err != nil {
		return gpus
	}
	
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		// Skip header and empty lines
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}
		
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			memoryBytes := 0
			fmt.Sscanf(parts[1], "%d", &memoryBytes)
			
			gpu := GPUInfo{
				ID:       fmt.Sprintf("%d", i-2),
				Model:    strings.TrimSpace(parts[2]),
				MemoryMB: memoryBytes / (1024 * 1024),
			}
			
			// Determine vendor from model name
			modelLower := strings.ToLower(gpu.Model)
			if strings.Contains(modelLower, "nvidia") {
				gpu.Vendor = "NVIDIA"
			} else if strings.Contains(modelLower, "amd") || strings.Contains(modelLower, "radeon") {
				gpu.Vendor = "AMD"
			} else if strings.Contains(modelLower, "intel") {
				gpu.Vendor = "Intel"
			}
			
			gpus = append(gpus, gpu)
		}
	}
	
	return gpus
}

// getWindowsVersion returns the Windows version
func getWindowsVersion() string {
	output, err := exec.Command("cmd", "/c", "ver").Output()
	if err != nil {
		return "Windows"
	}
	
	version := strings.TrimSpace(string(output))
	// Extract version number from output like "Microsoft Windows [Version 10.0.19043.1234]"
	if strings.Contains(version, "[Version") {
		start := strings.Index(version, "[Version") + 9
		end := strings.Index(version[start:], "]")
		if end > 0 {
			return "Windows " + version[start:start+end]
		}
	}
	
	return version
}

// isWSL2Available checks if WSL2 is available
func isWSL2Available() bool {
	output, err := exec.Command("wsl", "--list", "--verbose").Output()
	return err == nil && strings.Contains(string(output), "2")
}

// isHyperVAvailable checks if Hyper-V is available
func isHyperVAvailable() bool {
	// Check if Hyper-V service exists
	output, err := exec.Command("sc", "query", "vmms").Output()
	return err == nil && strings.Contains(string(output), "RUNNING")
}

// detectContainerRuntime detects the available container runtime on Windows
func detectContainerRuntime() string {
	// Check for Docker Desktop
	if _, err := exec.LookPath("docker"); err == nil {
		// Verify Docker is actually running
		if err := exec.Command("docker", "version").Run(); err == nil {
			return "docker"
		}
	}
	
	// Check for containerd
	if _, err := exec.LookPath("containerd"); err == nil {
		return "containerd"
	}
	
	return ""
}

// getHostname returns the system hostname
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
} 