package core

import (
	"time"
)

// Job represents a compute job to be executed
type Job struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Priority     int                    `json:"priority"`
	Requirements ResourceRequirements   `json:"requirements"`
	Payload      JobPayload            `json:"payload"`
	Timeout      time.Duration         `json:"timeout"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// JobPayload contains the actual work to be done
type JobPayload struct {
	Runtime     string            `json:"runtime"` // docker, wasm, native
	Image       string            `json:"image,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	InputData   []byte            `json:"input_data,omitempty"`
	OutputPath  string            `json:"output_path,omitempty"`
}

// ResourceRequirements specifies required resources for a job
type ResourceRequirements struct {
	CPUCores     int     `json:"cpu_cores"`
	MemoryMB     int     `json:"memory_mb"`
	DiskMB       int     `json:"disk_mb"`
	GPUCount     int     `json:"gpu_count,omitempty"`
	GPUType      string  `json:"gpu_type,omitempty"`
	FPGACount    int     `json:"fpga_count,omitempty"`
	TPUCount     int     `json:"tpu_count,omitempty"`
	NetworkMbps  int     `json:"network_mbps,omitempty"`
}

// JobResult contains the results of job execution
type JobResult struct {
	JobID        string    `json:"job_id"`
	Status       string    `json:"status"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	OutputData   []byte    `json:"output_data,omitempty"`
	OutputHash   string    `json:"output_hash"`
	ResourceUsed ResourceUsage `json:"resource_used"`
	ExitCode     int       `json:"exit_code"`
	Logs         string    `json:"logs,omitempty"`
}

// SystemInfo contains information about the system
type SystemInfo struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
	CPUModel     string `json:"cpu_model"`
	CPUCores     int    `json:"cpu_cores"`
	MemoryTotal  uint64 `json:"memory_total"`
	DiskTotal    uint64 `json:"disk_total"`
}

// ResourceInfo contains detailed resource information
type ResourceInfo struct {
	CPUCores        int        `json:"cpu_cores"`
	MemoryTotal     uint64     `json:"memory_total"`
	MemoryAvailable uint64     `json:"memory_available"`
	DiskTotal       uint64     `json:"disk_total"`
	DiskAvailable   uint64     `json:"disk_available"`
	GPUs            []GPUInfo  `json:"gpus,omitempty"`
	FPGAs           []FPGAInfo `json:"fpgas,omitempty"`
	TPUs            []TPUInfo  `json:"tpus,omitempty"`
}

// ResourceUsage represents current resource usage
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryUsedMB  uint64  `json:"memory_used_mb"`
	DiskPercent   float64 `json:"disk_percent"`
	DiskUsedMB    uint64  `json:"disk_used_mb"`
	NetworkInMbps float64 `json:"network_in_mbps"`
	NetworkOutMbps float64 `json:"network_out_mbps"`
}

// GPUInfo contains GPU information
type GPUInfo struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Vendor      string `json:"vendor"`
	MemoryMB    int    `json:"memory_mb"`
	ComputeCap  string `json:"compute_capability"`
	Temperature float64 `json:"temperature"`
	Utilization float64 `json:"utilization"`
}

// FPGAInfo contains FPGA information
type FPGAInfo struct {
	Index    int    `json:"index"`
	Model    string `json:"model"`
	Vendor   string `json:"vendor"`
	Version  string `json:"version"`
}

// TPUInfo contains TPU information
type TPUInfo struct {
	Index   int    `json:"index"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

// Request types for control plane communication

// RegisterRequest is sent when an agent registers
type RegisterRequest struct {
	AgentID      string       `json:"agent_id"`
	SystemInfo   *SystemInfo  `json:"system_info"`
	Resources    *ResourceInfo `json:"resources"`
	Capabilities []string     `json:"capabilities"`
	Version      string       `json:"version"`
}

// DeregisterRequest is sent when an agent deregisters
type DeregisterRequest struct {
	AgentID string `json:"agent_id"`
	Reason  string `json:"reason"`
}

// HeartbeatRequest is sent periodically to maintain connection
type HeartbeatRequest struct {
	AgentID       string        `json:"agent_id"`
	Status        AgentStatus   `json:"status"`
	ResourceUsage *ResourceUsage `json:"resource_usage"`
	Timestamp     time.Time     `json:"timestamp"`
}

// JobPollRequest is sent to request new jobs
type JobPollRequest struct {
	AgentID string `json:"agent_id"`
	MaxJobs int    `json:"max_jobs"`
}

// JobResultRequest is sent to report job results
type JobResultRequest struct {
	AgentID   string     `json:"agent_id"`
	JobID     string     `json:"job_id"`
	Status    string     `json:"status"`
	Result    *JobResult `json:"result,omitempty"`
	Error     string     `json:"error,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// Response types

// RegisterResponse is received after registration
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

// HeartbeatResponse is received after heartbeat
type HeartbeatResponse struct {
	Success  bool     `json:"success"`
	Commands []string `json:"commands,omitempty"`
}

// JobPollResponse contains jobs to execute
type JobPollResponse struct {
	Jobs []*Job `json:"jobs"`
}

// Error types

// AgentError represents an error in agent operations
type AgentError struct {
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Time    time.Time `json:"time"`
}

func (e *AgentError) Error() string {
	return e.Message
} 