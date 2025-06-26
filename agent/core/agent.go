package core

import (
	"context"
	"fmt"
<<<<<<< HEAD
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
)

// Agent represents the core compute agent
type Agent struct {
	ID              string
	config          *Config
	logger          *zap.Logger
	client          *ControlPlaneClient
	resourceMonitor *ResourceMonitor
	jobExecutor     *JobExecutor
	heartbeatTicker *time.Ticker
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.RWMutex
	status          AgentStatus
}

// AgentStatus represents the current status of the agent
type AgentStatus struct {
	State       string    `json:"state"`
	LastPing    time.Time `json:"last_ping"`
	ActiveJobs  int       `json:"active_jobs"`
	TotalJobs   int64     `json:"total_jobs"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
}

// Config holds agent configuration
type Config struct {
	ControlPlaneURL string        `json:"control_plane_url"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	MaxConcurrentJobs int          `json:"max_concurrent_jobs"`
	ResourceLimits    ResourceLimits `json:"resource_limits"`
	SecurityConfig    SecurityConfig `json:"security_config"`
}

// ResourceLimits defines resource usage limits
type ResourceLimits struct {
	MaxCPUPercent    float64 `json:"max_cpu_percent"`
	MaxMemoryPercent float64 `json:"max_memory_percent"`
	MaxDiskPercent   float64 `json:"max_disk_percent"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableTLS      bool   `json:"enable_tls"`
	CertFile       string `json:"cert_file"`
	KeyFile        string `json:"key_file"`
	CAFile         string `json:"ca_file"`
	EnableAttestation bool   `json:"enable_attestation"`
}

// NewAgent creates a new compute agent
func NewAgent(config *Config, logger *zap.Logger) (*Agent, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	agent := &Agent{
		ID:     uuid.New().String(),
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		status: AgentStatus{
			State:    "initializing",
			LastPing: time.Now(),
		},
	}

	// Initialize components
	client, err := NewControlPlaneClient(config.ControlPlaneURL, config.SecurityConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create control plane client: %w", err)
	}
	agent.client = client

	agent.resourceMonitor = NewResourceMonitor(logger)
	agent.jobExecutor = NewJobExecutor(config.MaxConcurrentJobs, logger)

	return agent, nil
}

// Start begins agent operations
func (a *Agent) Start() error {
	a.logger.Info("Starting ComputeHive agent", zap.String("agent_id", a.ID))

=======
	"log"
	"sync"
	"time"
)

// Agent represents the main compute agent
type Agent struct {
	id              string
	config          *Config
	client          *Client
	resourceMonitor *ResourceMonitor
	jobExecutor     *JobExecutor
	metrics         *AgentMetrics
	status          AgentStatus
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewAgent creates a new compute agent instance
func NewAgent(config *Config) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	client, err := NewClient(config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	
	resourceMonitor := NewResourceMonitor()
	jobExecutor := NewJobExecutor(config)
	
	agent := &Agent{
		id:              GenerateAgentID(),
		config:          config,
		client:          client,
		resourceMonitor: resourceMonitor,
		jobExecutor:     jobExecutor,
		metrics:         NewAgentMetrics(),
		status:          AgentStatusInitializing,
		ctx:             ctx,
		cancel:          cancel,
	}
	
	return agent, nil
}

// Start begins the agent operation
func (a *Agent) Start() error {
	log.Printf("Starting ComputeHive agent %s", a.id)
	
	// Start resource monitoring
	go a.resourceMonitor.Start(a.ctx)
	
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	// Register with control plane
	if err := a.register(); err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}
<<<<<<< HEAD

	// Start resource monitoring
	a.wg.Add(1)
	go a.monitorResources()

	// Start heartbeat
	a.heartbeatTicker = time.NewTicker(a.config.HeartbeatInterval)
	a.wg.Add(1)
	go a.heartbeatLoop()

	// Start job polling
	a.wg.Add(1)
	go a.jobPollingLoop()

	a.setStatus("running")
	a.logger.Info("Agent started successfully", zap.String("agent_id", a.ID))

	return nil
}

// Stop gracefully stops the agent
func (a *Agent) Stop() error {
	a.logger.Info("Stopping agent", zap.String("agent_id", a.ID))
	a.setStatus("stopping")

	// Cancel context to signal goroutines
	a.cancel()

	// Stop heartbeat
	if a.heartbeatTicker != nil {
		a.heartbeatTicker.Stop()
	}

	// Wait for active jobs to complete
	if err := a.jobExecutor.Shutdown(30 * time.Second); err != nil {
		a.logger.Error("Error shutting down job executor", zap.Error(err))
	}

	// Deregister from control plane
	if err := a.deregister(); err != nil {
		a.logger.Error("Error deregistering agent", zap.Error(err))
	}

	// Wait for all goroutines to finish
	a.wg.Wait()

	a.setStatus("stopped")
	a.logger.Info("Agent stopped", zap.String("agent_id", a.ID))

=======
	
	// Update status
	a.setStatus(AgentStatusActive)
	
	// Start main loops
	go a.heartbeatLoop()
	go a.jobPollingLoop()
	go a.metricsReportingLoop()
	
	log.Printf("Agent %s started successfully", a.id)
	return nil
}

// Stop gracefully shuts down the agent
func (a *Agent) Stop() error {
	log.Printf("Stopping agent %s", a.id)
	
	a.setStatus(AgentStatusShuttingDown)
	
	// Cancel context to stop all goroutines
	a.cancel()
	
	// Wait for active jobs to complete
	if err := a.jobExecutor.WaitForCompletion(30 * time.Second); err != nil {
		log.Printf("Warning: some jobs did not complete cleanly: %v", err)
	}
	
	// Deregister from control plane
	if err := a.deregister(); err != nil {
		log.Printf("Warning: failed to deregister agent: %v", err)
	}
	
	a.setStatus(AgentStatusStopped)
	log.Printf("Agent %s stopped", a.id)
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	return nil
}

// register registers the agent with the control plane
func (a *Agent) register() error {
<<<<<<< HEAD
	systemInfo, err := a.getSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to get system info: %w", err)
	}

	resources, err := a.getResourceInfo()
	if err != nil {
		return fmt.Errorf("failed to get resource info: %w", err)
	}

	req := &RegisterRequest{
		AgentID:    a.ID,
		SystemInfo: systemInfo,
		Resources:  resources,
		Capabilities: a.getCapabilities(),
	}

	if err := a.client.Register(a.ctx, req); err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	a.setStatus("registered")
=======
	resources := a.resourceMonitor.GetResources()
	
	req := &RegisterRequest{
		AgentID:      a.id,
		Version:      Version,
		Platform:     GetPlatformInfo(),
		Resources:    resources,
		Capabilities: a.getCapabilities(),
	}
	
	resp, err := a.client.Register(a.ctx, req)
	if err != nil {
		return err
	}
	
	a.config.Token = resp.Token
	log.Printf("Agent registered successfully with ID: %s", a.id)
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	return nil
}

// deregister removes the agent from the control plane
func (a *Agent) deregister() error {
<<<<<<< HEAD
	req := &DeregisterRequest{
		AgentID: a.ID,
		Reason:  "agent_shutdown",
	}

	return a.client.Deregister(context.Background(), req)
=======
	return a.client.Deregister(a.ctx, a.id)
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
}

// heartbeatLoop sends periodic heartbeats to the control plane
func (a *Agent) heartbeatLoop() {
<<<<<<< HEAD
	defer a.wg.Done()

	for {
		select {
		case <-a.heartbeatTicker.C:
			if err := a.sendHeartbeat(); err != nil {
				a.logger.Error("Failed to send heartbeat", zap.Error(err))
=======
	ticker := time.NewTicker(a.config.HeartbeatInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				a.metrics.IncrementHeartbeatFailures()
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
			}
		case <-a.ctx.Done():
			return
		}
	}
}

// sendHeartbeat sends a heartbeat to the control plane
func (a *Agent) sendHeartbeat() error {
<<<<<<< HEAD
	status := a.getStatus()
	resources, err := a.resourceMonitor.GetCurrentUsage()
	if err != nil {
		a.logger.Warn("Failed to get resource usage", zap.Error(err))
	}

	req := &HeartbeatRequest{
		AgentID:       a.ID,
		Status:        status,
		ResourceUsage: resources,
		Timestamp:     time.Now(),
	}

	if err := a.client.Heartbeat(a.ctx, req); err != nil {
		return fmt.Errorf("heartbeat failed: %w", err)
	}

	a.mu.Lock()
	a.status.LastPing = time.Now()
	a.mu.Unlock()

	return nil
=======
	resources := a.resourceMonitor.GetResources()
	activeJobs := a.jobExecutor.GetActiveJobs()
	
	heartbeat := &Heartbeat{
		AgentID:    a.id,
		Timestamp:  time.Now(),
		Status:     a.getStatus(),
		Resources:  resources,
		ActiveJobs: activeJobs,
		Metrics:    a.metrics.GetSnapshot(),
	}
	
	return a.client.SendHeartbeat(a.ctx, heartbeat)
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
}

// jobPollingLoop polls for new jobs from the control plane
func (a *Agent) jobPollingLoop() {
<<<<<<< HEAD
	defer a.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if a.canAcceptJobs() {
				a.pollForJobs()
=======
	ticker := time.NewTicker(a.config.JobPollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := a.pollJobs(); err != nil {
				log.Printf("Failed to poll jobs: %v", err)
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
			}
		case <-a.ctx.Done():
			return
		}
	}
}

<<<<<<< HEAD
// pollForJobs requests new jobs from the control plane
func (a *Agent) pollForJobs() {
	req := &JobPollRequest{
		AgentID: a.ID,
		MaxJobs: a.getAvailableJobSlots(),
	}

	jobs, err := a.client.PollJobs(a.ctx, req)
	if err != nil {
		a.logger.Error("Failed to poll for jobs", zap.Error(err))
		return
	}

	for _, job := range jobs {
		if err := a.executeJob(job); err != nil {
			a.logger.Error("Failed to execute job",
				zap.String("job_id", job.ID),
				zap.Error(err))
		}
	}
}

// executeJob executes a single job
func (a *Agent) executeJob(job *Job) error {
	a.mu.Lock()
	a.status.ActiveJobs++
	a.mu.Unlock()

	// Execute job asynchronously
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		defer func() {
			a.mu.Lock()
			a.status.ActiveJobs--
			a.status.TotalJobs++
			a.mu.Unlock()
		}()

		result, err := a.jobExecutor.Execute(a.ctx, job)
		if err != nil {
			a.logger.Error("Job execution failed",
				zap.String("job_id", job.ID),
				zap.Error(err))
			// Report failure to control plane
			a.reportJobResult(job.ID, nil, err)
			return
		}

		// Report success to control plane
		a.reportJobResult(job.ID, result, nil)
	}()

	return nil
}

// reportJobResult reports job execution results to the control plane
func (a *Agent) reportJobResult(jobID string, result *JobResult, err error) {
	status := "completed"
	errorMsg := ""
	if err != nil {
		status = "failed"
		errorMsg = err.Error()
	}

	req := &JobResultRequest{
		AgentID:  a.ID,
		JobID:    jobID,
		Status:   status,
		Result:   result,
		Error:    errorMsg,
		Timestamp: time.Now(),
	}

	if err := a.client.ReportJobResult(context.Background(), req); err != nil {
		a.logger.Error("Failed to report job result",
			zap.String("job_id", jobID),
			zap.Error(err))
	}
}

// monitorResources continuously monitors system resources
func (a *Agent) monitorResources() {
	defer a.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			usage, err := a.resourceMonitor.GetCurrentUsage()
			if err != nil {
				a.logger.Error("Failed to get resource usage", zap.Error(err))
				continue
			}

			a.mu.Lock()
			a.status.CPUUsage = usage.CPUPercent
			a.status.MemoryUsage = usage.MemoryPercent
			a.status.DiskUsage = usage.DiskPercent
			a.mu.Unlock()

			// Check if resources are within limits
			if !a.checkResourceLimits(usage) {
				a.logger.Warn("Resource limits exceeded",
					zap.Float64("cpu", usage.CPUPercent),
					zap.Float64("memory", usage.MemoryPercent),
					zap.Float64("disk", usage.DiskPercent))
=======
// pollJobs checks for new jobs and executes them
func (a *Agent) pollJobs() error {
	// Only poll if we have available capacity
	if !a.hasCapacity() {
		return nil
	}
	
	jobs, err := a.client.GetJobs(a.ctx, a.id)
	if err != nil {
		return err
	}
	
	for _, job := range jobs {
		if err := a.executeJob(job); err != nil {
			log.Printf("Failed to execute job %s: %v", job.ID, err)
			a.reportJobFailure(job, err)
		}
	}
	
	return nil
}

// executeJob runs a job on the agent
func (a *Agent) executeJob(job *Job) error {
	log.Printf("Executing job %s", job.ID)
	a.metrics.IncrementJobsStarted()
	
	// Validate job requirements
	if err := a.validateJob(job); err != nil {
		return fmt.Errorf("job validation failed: %w", err)
	}
	
	// Execute the job
	result, err := a.jobExecutor.Execute(a.ctx, job)
	if err != nil {
		a.metrics.IncrementJobsFailed()
		return err
	}
	
	// Report result to control plane
	if err := a.client.ReportJobResult(a.ctx, result); err != nil {
		return fmt.Errorf("failed to report job result: %w", err)
	}
	
	a.metrics.IncrementJobsCompleted()
	log.Printf("Job %s completed successfully", job.ID)
	return nil
}

// validateJob checks if the job can be executed on this agent
func (a *Agent) validateJob(job *Job) error {
	resources := a.resourceMonitor.GetResources()
	
	// Check CPU requirements
	if job.Requirements.CPUCores > resources.CPU.Cores {
		return fmt.Errorf("insufficient CPU cores: required %d, available %d", 
			job.Requirements.CPUCores, resources.CPU.Cores)
	}
	
	// Check memory requirements
	if job.Requirements.MemoryMB > resources.Memory.Available {
		return fmt.Errorf("insufficient memory: required %d MB, available %d MB", 
			job.Requirements.MemoryMB, resources.Memory.Available)
	}
	
	// Check GPU requirements if specified
	if job.Requirements.GPUCount > 0 && len(resources.GPUs) < job.Requirements.GPUCount {
		return fmt.Errorf("insufficient GPUs: required %d, available %d", 
			job.Requirements.GPUCount, len(resources.GPUs))
	}
	
	return nil
}

// hasCapacity checks if the agent can accept new jobs
func (a *Agent) hasCapacity() bool {
	activeJobs := a.jobExecutor.GetActiveJobCount()
	return activeJobs < a.config.MaxConcurrentJobs
}

// reportJobFailure notifies the control plane of a job failure
func (a *Agent) reportJobFailure(job *Job, err error) {
	result := &JobResult{
		JobID:     job.ID,
		AgentID:   a.id,
		Status:    JobStatusFailed,
		Error:     err.Error(),
		Timestamp: time.Now(),
	}
	
	if reportErr := a.client.ReportJobResult(a.ctx, result); reportErr != nil {
		log.Printf("Failed to report job failure: %v", reportErr)
	}
}

// metricsReportingLoop periodically reports metrics
func (a *Agent) metricsReportingLoop() {
	ticker := time.NewTicker(a.config.MetricsInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := a.reportMetrics(); err != nil {
				log.Printf("Failed to report metrics: %v", err)
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
			}
		case <-a.ctx.Done():
			return
		}
	}
}

<<<<<<< HEAD
// Helper methods

func (a *Agent) setStatus(status string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.status.State = status
}

=======
// reportMetrics sends metrics to the control plane
func (a *Agent) reportMetrics() error {
	metrics := &MetricsReport{
		AgentID:   a.id,
		Timestamp: time.Now(),
		Metrics:   a.metrics.GetSnapshot(),
		Resources: a.resourceMonitor.GetResources(),
	}
	
	return a.client.ReportMetrics(a.ctx, metrics)
}

// getCapabilities returns the agent's capabilities
func (a *Agent) getCapabilities() []string {
	caps := []string{"docker", "kubernetes"}
	
	resources := a.resourceMonitor.GetResources()
	if len(resources.GPUs) > 0 {
		caps = append(caps, "gpu")
		// Add specific GPU capabilities
		for _, gpu := range resources.GPUs {
			if gpu.Vendor == "NVIDIA" {
				caps = append(caps, "cuda")
			} else if gpu.Vendor == "AMD" {
				caps = append(caps, "rocm")
			}
		}
	}
	
	// Add platform-specific capabilities
	caps = append(caps, GetPlatformCapabilities()...)
	
	return caps
}

// setStatus updates the agent status
func (a *Agent) setStatus(status AgentStatus) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.status = status
}

// getStatus returns the current agent status
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
func (a *Agent) getStatus() AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
<<<<<<< HEAD
}

func (a *Agent) canAcceptJobs() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status.State == "running" && a.status.ActiveJobs < a.config.MaxConcurrentJobs
}

func (a *Agent) getAvailableJobSlots() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config.MaxConcurrentJobs - a.status.ActiveJobs
}

func (a *Agent) checkResourceLimits(usage *ResourceUsage) bool {
	limits := a.config.ResourceLimits
	return usage.CPUPercent <= limits.MaxCPUPercent &&
		usage.MemoryPercent <= limits.MaxMemoryPercent &&
		usage.DiskPercent <= limits.MaxDiskPercent
}

func (a *Agent) getSystemInfo() (*SystemInfo, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		Hostname:     hostInfo.Hostname,
		OS:           hostInfo.OS,
		Platform:     hostInfo.Platform,
		Architecture: hostInfo.KernelArch,
		CPUModel:     cpuInfo[0].ModelName,
		CPUCores:     len(cpuInfo),
		MemoryTotal:  memInfo.Total,
		DiskTotal:    diskInfo.Total,
	}, nil
}

func (a *Agent) getResourceInfo() (*ResourceInfo, error) {
	cpuCount, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &ResourceInfo{
		CPUCores:     cpuCount,
		MemoryTotal:  memInfo.Total,
		MemoryAvailable: memInfo.Available,
		DiskTotal:    diskInfo.Total,
		DiskAvailable: diskInfo.Free,
		GPUs:         a.detectGPUs(),
		FPGAs:        a.detectFPGAs(),
		TPUs:         a.detectTPUs(),
	}, nil
}

func (a *Agent) getCapabilities() []string {
	caps := []string{
		"docker",
		"containerd",
		"linux-cgroups",
	}

	// Add platform-specific capabilities
	if a.hasNvidiaGPU() {
		caps = append(caps, "nvidia-gpu", "cuda")
	}

	if a.hasIntelSGX() {
		caps = append(caps, "intel-sgx")
	}

	if a.hasAMDSEV() {
		caps = append(caps, "amd-sev")
	}

	return caps
}

// Placeholder methods for hardware detection
func (a *Agent) detectGPUs() []GPUInfo {
	var gpus []GPUInfo
	
	// Try to detect NVIDIA GPUs using nvidia-smi
	if output, err := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.total", "--format=csv,noheader").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			parts := strings.Split(line, ", ")
			if len(parts) >= 3 {
				gpus = append(gpus, GPUInfo{
					Index:  parts[0],
					Name:   parts[1],
					Memory: parts[2],
					Type:   "NVIDIA",
				})
			}
		}
	}
	
	// Try to detect AMD GPUs using rocm-smi
	if output, err := exec.Command("rocm-smi", "--showproductname").Output(); err == nil {
		// Parse AMD GPU info
		if strings.Contains(string(output), "GPU") {
			gpus = append(gpus, GPUInfo{
				Type: "AMD",
				Name: "AMD GPU",
			})
		}
	}
	
	return gpus
}

func (a *Agent) detectFPGAs() []FPGAInfo {
	var fpgas []FPGAInfo
	
	// Try to detect Xilinx FPGAs
	if output, err := exec.Command("xbutil", "list").Output(); err == nil {
		if strings.Contains(string(output), "Device") {
			fpgas = append(fpgas, FPGAInfo{
				Type:   "Xilinx",
				Status: "available",
			})
		}
	}
	
	// Try to detect Intel FPGAs
	if output, err := exec.Command("aocl", "diagnose").Output(); err == nil {
		if strings.Contains(string(output), "FPGA") {
			fpgas = append(fpgas, FPGAInfo{
				Type:   "Intel",
				Status: "available",
			})
		}
	}
	
	return fpgas
}

func (a *Agent) detectTPUs() []TPUInfo {
	var tpus []TPUInfo
	
	// Try to detect Google Cloud TPUs
	if output, err := exec.Command("gcloud", "compute", "tpus", "list", "--format=json").Output(); err == nil {
		// Parse TPU info from JSON
		if len(output) > 0 && string(output) != "[]" {
			tpus = append(tpus, TPUInfo{
				Type:   "Google Cloud TPU",
				Status: "available",
			})
		}
	}
	
	// Try to detect Edge TPUs
	if output, err := exec.Command("lsusb").Output(); err == nil {
		if strings.Contains(string(output), "Global Unichip Corp") || strings.Contains(string(output), "Google Inc") {
			tpus = append(tpus, TPUInfo{
				Type:   "Edge TPU",
				Status: "available",
			})
		}
	}
	
	return tpus
}

func (a *Agent) hasNvidiaGPU() bool {
	_, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return false
	}
	
	output, err := exec.Command("nvidia-smi", "-L").Output()
	return err == nil && strings.Contains(string(output), "GPU")
}

func (a *Agent) hasIntelSGX() bool {
	// Check for Intel SGX support
	if output, err := exec.Command("cpuid").Output(); err == nil {
		return strings.Contains(string(output), "SGX")
	}
	
	// Alternative: check /proc/cpuinfo
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		return strings.Contains(string(data), "sgx")
	}
	
	return false
}

func (a *Agent) hasAMDSEV() bool {
	// Check for AMD SEV support
	if data, err := os.ReadFile("/sys/module/kvm_amd/parameters/sev"); err == nil {
		return strings.TrimSpace(string(data)) == "1" || strings.TrimSpace(string(data)) == "Y"
	}
	
	// Alternative: check cpuid
	if output, err := exec.Command("cpuid").Output(); err == nil {
		return strings.Contains(string(output), "SEV")
	}
	
	return false
=======
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
} 