package core

import (
	"context"
	"fmt"
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
	
	// Register with control plane
	if err := a.register(); err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}
	
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
	return nil
}

// register registers the agent with the control plane
func (a *Agent) register() error {
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
	return nil
}

// deregister removes the agent from the control plane
func (a *Agent) deregister() error {
	return a.client.Deregister(a.ctx, a.id)
}

// heartbeatLoop sends periodic heartbeats to the control plane
func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(a.config.HeartbeatInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				a.metrics.IncrementHeartbeatFailures()
			}
		case <-a.ctx.Done():
			return
		}
	}
}

// sendHeartbeat sends a heartbeat to the control plane
func (a *Agent) sendHeartbeat() error {
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
}

// jobPollingLoop polls for new jobs from the control plane
func (a *Agent) jobPollingLoop() {
	ticker := time.NewTicker(a.config.JobPollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := a.pollJobs(); err != nil {
				log.Printf("Failed to poll jobs: %v", err)
			}
		case <-a.ctx.Done():
			return
		}
	}
}

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
			}
		case <-a.ctx.Done():
			return
		}
	}
}

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
func (a *Agent) getStatus() AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
} 