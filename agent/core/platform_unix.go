//go:build !windows
// +build !windows

package core

import (
<<<<<<< HEAD
	"github.com/shirou/gopsutil/v3/load"
)

// getLoadAverage returns system load averages for Unix-like systems
func getLoadAverage() (map[string]float64, error) {
	avg, err := load.Avg()
	if err != nil {
		return nil, err
	}
	
	return map[string]float64{
		"1min":  avg.Load1,
		"5min":  avg.Load5,
		"15min": avg.Load15,
	}, nil
=======
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	
	"github.com/google/uuid"
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
	
	// Get OS version
	if runtime.GOOS == "darwin" {
		if version, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			platform.Version = strings.TrimSpace(string(version))
		}
	} else if runtime.GOOS == "linux" {
		if version, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(version), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					platform.Version = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
					break
				}
			}
		}
	}
	
	// Check container runtime
	platform.ContainerRuntime = detectContainerRuntime()
	
	return platform
}

// GetPlatformCapabilities returns platform-specific capabilities
func GetPlatformCapabilities() []string {
	caps := []string{}
	
	// Check for specific capabilities
	if _, err := exec.LookPath("docker"); err == nil {
		caps = append(caps, "docker")
	}
	
	if _, err := exec.LookPath("podman"); err == nil {
		caps = append(caps, "podman")
	}
	
	if _, err := exec.LookPath("singularity"); err == nil {
		caps = append(caps, "singularity")
	}
	
	// Check for hardware features
	if runtime.GOOS == "linux" {
		// Check for KVM
		if _, err := os.Stat("/dev/kvm"); err == nil {
			caps = append(caps, "kvm")
		}
		
		// Check for SGX
		if _, err := os.Stat("/dev/sgx"); err == nil {
			caps = append(caps, "sgx")
		}
	}
	
	return caps
}

// detectGPUs detects available GPUs on the system
func detectGPUs() []GPUInfo {
	var gpus []GPUInfo
	
	// Try NVIDIA GPUs
	if nvidiaGPUs := detectNVIDIAGPUs(); len(nvidiaGPUs) > 0 {
		gpus = append(gpus, nvidiaGPUs...)
	}
	
	// Try AMD GPUs
	if amdGPUs := detectAMDGPUs(); len(amdGPUs) > 0 {
		gpus = append(gpus, amdGPUs...)
	}
	
	// Try Intel GPUs (for integrated graphics)
	if intelGPUs := detectIntelGPUs(); len(intelGPUs) > 0 {
		gpus = append(gpus, intelGPUs...)
	}
	
	return gpus
}

// detectNVIDIAGPUs detects NVIDIA GPUs using nvidia-smi
func detectNVIDIAGPUs() []GPUInfo {
	var gpus []GPUInfo
	
	// Check if nvidia-smi is available
	output, err := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.total,utilization.gpu,temperature.gpu,power.draw", "--format=csv,noheader,nounits").Output()
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

// detectAMDGPUs detects AMD GPUs using rocm-smi
func detectAMDGPUs() []GPUInfo {
	var gpus []GPUInfo
	
	// Check if rocm-smi is available
	output, err := exec.Command("rocm-smi", "--showid", "--showproductname", "--showmeminfo", "vram", "--showuse", "--showtemp", "--showpower").Output()
	if err != nil {
		return gpus
	}
	
	// Parse rocm-smi output (simplified)
	// In a real implementation, this would need proper parsing
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if strings.Contains(line, "GPU[") {
			gpu := GPUInfo{
				ID:     fmt.Sprintf("%d", i),
				Vendor: "AMD",
				Model:  "AMD GPU", // Would need to parse actual model
			}
			gpus = append(gpus, gpu)
		}
	}
	
	return gpus
}

// detectIntelGPUs detects Intel integrated GPUs
func detectIntelGPUs() []GPUInfo {
	var gpus []GPUInfo
	
	// On Linux, check for Intel GPU in /sys
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/sys/class/drm/card0"); err == nil {
			// Check if it's Intel
			vendor, _ := os.ReadFile("/sys/class/drm/card0/device/vendor")
			if strings.TrimSpace(string(vendor)) == "0x8086" { // Intel vendor ID
				gpu := GPUInfo{
					ID:     "0",
					Vendor: "Intel",
					Model:  "Intel Integrated Graphics",
				}
				gpus = append(gpus, gpu)
			}
		}
	}
	
	return gpus
}

// detectContainerRuntime detects the available container runtime
func detectContainerRuntime() string {
	if _, err := exec.LookPath("docker"); err == nil {
		return "docker"
	}
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman"
	}
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
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
} 