package core

import (
	"time"
)

const (
	// Version is the agent version
	Version = "1.0.0"
)

// Config represents agent configuration
type Config struct {
	ControlPlaneURL    string        `json:"control_plane_url"`
	Token              string        `json:"token"`
	HeartbeatInterval  time.Duration `json:"heartbeat_interval"`
	JobPollingInterval time.Duration `json:"job_polling_interval"`
	MetricsInterval    time.Duration `json:"metrics_interval"`
	MaxConcurrentJobs  int           `json:"max_concurrent_jobs"`
	WorkDir            string        `json:"work_dir"`
	EnableGPU          bool          `json:"enable_gpu"`
	EnableTrustedExec  bool          `json:"enable_trusted_exec"`
	LogLevel           string        `json:"log_level"`
}

// AgentStatus represents the agent's current status
type AgentStatus string

const (
	AgentStatusInitializing AgentStatus = "initializing"
	AgentStatusActive       AgentStatus = "active"
	AgentStatusBusy         AgentStatus = "busy"
	AgentStatusShuttingDown AgentStatus = "shutting_down"
	AgentStatusStopped      AgentStatus = "stopped"
	AgentStatusError        AgentStatus = "error"
)

// Job represents a compute job
type Job struct {
	ID           string            `json:"id"`
	Type         JobType           `json:"type"`
	Requirements ResourceRequirements `json:"requirements"`
	Payload      JobPayload        `json:"payload"`
	Priority     int               `json:"priority"`
	Timeout      time.Duration     `json:"timeout"`
	CreatedAt    time.Time         `json:"created_at"`
	MaxRetries   int               `json:"max_retries"`
}

// JobType represents the type of job
type JobType string

const (
	JobTypeDocker     JobType = "docker"
	JobTypeKubernetes JobType = "kubernetes"
	JobTypeBinary     JobType = "binary"
	JobTypeWASM       JobType = "wasm"
	JobTypeScript     JobType = "script"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// JobPayload contains job-specific execution details
type JobPayload struct {
	// Docker job fields
	Image   string   `json:"image,omitempty"`
	Command []string `json:"command,omitempty"`
	Env     []string `json:"env,omitempty"`
	
	// Binary job fields
	BinaryURL string   `json:"binary_url,omitempty"`
	Args      []string `json:"args,omitempty"`
	
	// Script job fields
	Script   string `json:"script,omitempty"`
	Language string `json:"language,omitempty"`
	
	// Input/output
	InputData  string `json:"input_data,omitempty"`
	OutputPath string `json:"output_path,omitempty"`
}

// ResourceRequirements specifies job resource needs
type ResourceRequirements struct {
	CPUCores     int      `json:"cpu_cores"`
	MemoryMB     int      `json:"memory_mb"`
	GPUCount     int      `json:"gpu_count"`
	GPUType      string   `json:"gpu_type,omitempty"`
	StorageMB    int      `json:"storage_mb"`
	NetworkMbps  int      `json:"network_mbps"`
	TrustedExec  bool     `json:"trusted_exec"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// Resources represents available system resources
type Resources struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	GPUs    []GPUInfo   `json:"gpus,omitempty"`
	Storage StorageInfo `json:"storage"`
	Network NetworkInfo `json:"network"`
}

// CPUInfo contains CPU information
type CPUInfo struct {
	Model       string  `json:"model"`
	Cores       int     `json:"cores"`
	Threads     int     `json:"threads"`
	FrequencyHz int64   `json:"frequency_hz"`
	Usage       float64 `json:"usage"`
}

// MemoryInfo contains memory information
type MemoryInfo struct {
	Total     int64   `json:"total"`
	Available int64   `json:"available"`
	Used      int64   `json:"used"`
	Usage     float64 `json:"usage"`
}

// GPUInfo contains GPU information
type GPUInfo struct {
	ID         string  `json:"id"`
	Model      string  `json:"model"`
	Vendor     string  `json:"vendor"`
	MemoryMB   int     `json:"memory_mb"`
	Usage      float64 `json:"usage"`
	Temperature float64 `json:"temperature"`
	PowerWatts float64 `json:"power_watts"`
}

// StorageInfo contains storage information
type StorageInfo struct {
	Total     int64   `json:"total"`
	Available int64   `json:"available"`
	Used      int64   `json:"used"`
	Usage     float64 `json:"usage"`
}

// NetworkInfo contains network information
type NetworkInfo struct {
	Interfaces []NetworkInterface `json:"interfaces"`
	Bandwidth  int                `json:"bandwidth_mbps"`
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Type string `json:"type"`
}

// JobResult represents the result of a job execution
type JobResult struct {
	JobID      string         `json:"job_id"`
	AgentID    string         `json:"agent_id"`
	Status     JobStatus      `json:"status"`
	Output     string         `json:"output,omitempty"`
	Error      string         `json:"error,omitempty"`
	ExitCode   int            `json:"exit_code"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt time.Time      `json:"finished_at"`
	Metrics    *JobMetrics    `json:"metrics,omitempty"`
	Artifacts  []JobArtifact  `json:"artifacts,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

// JobMetrics contains job execution metrics
type JobMetrics struct {
	CPUTime      time.Duration `json:"cpu_time"`
	MemoryPeakMB int64         `json:"memory_peak_mb"`
	NetworkInMB  int64         `json:"network_in_mb"`
	NetworkOutMB int64         `json:"network_out_mb"`
	DiskReadMB   int64         `json:"disk_read_mb"`
	DiskWriteMB  int64         `json:"disk_write_mb"`
}

// JobArtifact represents an output artifact from a job
type JobArtifact struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Checksum  string `json:"checksum"`
	MimeType  string `json:"mime_type"`
}

// RegisterRequest is sent to register an agent
type RegisterRequest struct {
	AgentID      string     `json:"agent_id"`
	Version      string     `json:"version"`
	Platform     Platform   `json:"platform"`
	Resources    *Resources `json:"resources"`
	Capabilities []string   `json:"capabilities"`
}

// RegisterResponse is received after registration
type RegisterResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Platform contains platform information
type Platform struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	Version      string `json:"version"`
	Hostname     string `json:"hostname"`
	ContainerRuntime string `json:"container_runtime,omitempty"`
}

// Heartbeat is sent periodically to the control plane
type Heartbeat struct {
	AgentID    string           `json:"agent_id"`
	Timestamp  time.Time        `json:"timestamp"`
	Status     AgentStatus      `json:"status"`
	Resources  *Resources       `json:"resources"`
	ActiveJobs []string         `json:"active_jobs"`
	Metrics    *AgentMetrics    `json:"metrics"`
}

// AgentMetrics contains agent performance metrics
type AgentMetrics struct {
	JobsStarted        int64     `json:"jobs_started"`
	JobsCompleted      int64     `json:"jobs_completed"`
	JobsFailed         int64     `json:"jobs_failed"`
	HeartbeatFailures  int64     `json:"heartbeat_failures"`
	UptimeSeconds      int64     `json:"uptime_seconds"`
	LastReportTime     time.Time `json:"last_report_time"`
}

// MetricsReport contains detailed metrics for reporting
type MetricsReport struct {
	AgentID   string         `json:"agent_id"`
	Timestamp time.Time      `json:"timestamp"`
	Metrics   *AgentMetrics  `json:"metrics"`
	Resources *Resources     `json:"resources"`
}

// NewAgentMetrics creates a new AgentMetrics instance
func NewAgentMetrics() *AgentMetrics {
	return &AgentMetrics{
		LastReportTime: time.Now(),
	}
}

// IncrementJobsStarted increments the jobs started counter
func (m *AgentMetrics) IncrementJobsStarted() {
	m.JobsStarted++
}

// IncrementJobsCompleted increments the jobs completed counter
func (m *AgentMetrics) IncrementJobsCompleted() {
	m.JobsCompleted++
}

// IncrementJobsFailed increments the jobs failed counter
func (m *AgentMetrics) IncrementJobsFailed() {
	m.JobsFailed++
}

// IncrementHeartbeatFailures increments the heartbeat failures counter
func (m *AgentMetrics) IncrementHeartbeatFailures() {
	m.HeartbeatFailures++
}

// GetSnapshot returns a copy of the metrics
func (m *AgentMetrics) GetSnapshot() *AgentMetrics {
	return &AgentMetrics{
		JobsStarted:       m.JobsStarted,
		JobsCompleted:     m.JobsCompleted,
		JobsFailed:        m.JobsFailed,
		HeartbeatFailures: m.HeartbeatFailures,
		UptimeSeconds:     int64(time.Since(m.LastReportTime).Seconds()),
		LastReportTime:    time.Now(),
	}
} 