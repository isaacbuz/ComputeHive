package core

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// JobExecutor handles job execution
type JobExecutor struct {
	config      *Config
	activeJobs  map[string]*ActiveJob
	mu          sync.RWMutex
	workDir     string
	dockerAvailable bool
}

// ActiveJob represents a currently running job
type ActiveJob struct {
	Job       *Job
	Context   context.Context
	Cancel    context.CancelFunc
	StartTime time.Time
	Process   *os.Process
}

// NewJobExecutor creates a new job executor
func NewJobExecutor(config *Config) *JobExecutor {
	executor := &JobExecutor{
		config:     config,
		activeJobs: make(map[string]*ActiveJob),
		workDir:    config.WorkDir,
	}
	
	// Check Docker availability
	executor.dockerAvailable = executor.checkDockerAvailable()
	
	// Create work directory if it doesn't exist
	if err := os.MkdirAll(executor.workDir, 0755); err != nil {
		log.Printf("Warning: failed to create work directory: %v", err)
	}
	
	return executor
}

// Execute runs a job
func (je *JobExecutor) Execute(ctx context.Context, job *Job) (*JobResult, error) {
	// Create job context
	jobCtx, cancel := context.WithTimeout(ctx, job.Timeout)
	defer cancel()
	
	// Create job directory
	jobDir := filepath.Join(je.workDir, job.ID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create job directory: %w", err)
	}
	defer os.RemoveAll(jobDir) // Clean up after job
	
	// Register active job
	activeJob := &ActiveJob{
		Job:       job,
		Context:   jobCtx,
		Cancel:    cancel,
		StartTime: time.Now(),
	}
	
	je.mu.Lock()
	je.activeJobs[job.ID] = activeJob
	je.mu.Unlock()
	
	defer func() {
		je.mu.Lock()
		delete(je.activeJobs, job.ID)
		je.mu.Unlock()
	}()
	
	// Execute based on job type
	var result *JobResult
	var err error
	
	switch job.Type {
	case JobTypeDocker:
		result, err = je.executeDockerJob(jobCtx, job, jobDir)
	case JobTypeBinary:
		result, err = je.executeBinaryJob(jobCtx, job, jobDir)
	case JobTypeScript:
		result, err = je.executeScriptJob(jobCtx, job, jobDir)
	case JobTypeWASM:
		result, err = je.executeWASMJob(jobCtx, job, jobDir)
	default:
		err = fmt.Errorf("unsupported job type: %s", job.Type)
	}
	
	if err != nil {
		return &JobResult{
			JobID:      job.ID,
			AgentID:    GenerateAgentID(),
			Status:     JobStatusFailed,
			Error:      err.Error(),
			StartedAt:  activeJob.StartTime,
			FinishedAt: time.Now(),
		}, nil
	}
	
	return result, nil
}

// executeDockerJob runs a Docker-based job
func (je *JobExecutor) executeDockerJob(ctx context.Context, job *Job, workDir string) (*JobResult, error) {
	if !je.dockerAvailable {
		return nil, fmt.Errorf("Docker is not available on this system")
	}
	
	// Build Docker command
	args := []string{"run", "--rm"}
	
	// Add resource limits
	if job.Requirements.CPUCores > 0 {
		args = append(args, fmt.Sprintf("--cpus=%d", job.Requirements.CPUCores))
	}
	if job.Requirements.MemoryMB > 0 {
		args = append(args, fmt.Sprintf("--memory=%dm", job.Requirements.MemoryMB))
	}
	
	// Add work directory as volume
	args = append(args, "-v", fmt.Sprintf("%s:/work", workDir))
	args = append(args, "-w", "/work")
	
	// Add environment variables
	for _, env := range job.Payload.Env {
		args = append(args, "-e", env)
	}
	
	// Add image and command
	args = append(args, job.Payload.Image)
	args = append(args, job.Payload.Command...)
	
	// Execute Docker command
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	
	result := &JobResult{
		JobID:      job.ID,
		AgentID:    GenerateAgentID(),
		Status:     JobStatusCompleted,
		Output:     string(output),
		ExitCode:   0,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}
	
	if err != nil {
		result.Status = JobStatusFailed
		result.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}
	
	return result, nil
}

// executeBinaryJob runs a binary executable job
func (je *JobExecutor) executeBinaryJob(ctx context.Context, job *Job, workDir string) (*JobResult, error) {
	// Download binary if URL is provided
	binaryPath := job.Payload.BinaryURL
	if isURL(binaryPath) {
		downloadedPath := filepath.Join(workDir, "executable")
		if err := downloadFile(ctx, binaryPath, downloadedPath); err != nil {
			return nil, fmt.Errorf("failed to download binary: %w", err)
		}
		binaryPath = downloadedPath
		
		// Make executable
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to make binary executable: %w", err)
		}
	}
	
	// Execute binary
	cmd := exec.CommandContext(ctx, binaryPath, job.Payload.Args...)
	cmd.Dir = workDir
	
	// Set environment variables
	cmd.Env = append(os.Environ(), job.Payload.Env...)
	
	// Capture output
	output, err := cmd.CombinedOutput()
	
	result := &JobResult{
		JobID:      job.ID,
		AgentID:    GenerateAgentID(),
		Status:     JobStatusCompleted,
		Output:     string(output),
		ExitCode:   0,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}
	
	if err != nil {
		result.Status = JobStatusFailed
		result.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}
	
	return result, nil
}

// executeScriptJob runs a script-based job
func (je *JobExecutor) executeScriptJob(ctx context.Context, job *Job, workDir string) (*JobResult, error) {
	// Determine interpreter based on language
	var interpreter string
	var args []string
	
	switch job.Payload.Language {
	case "python":
		interpreter = "python3"
	case "javascript", "js":
		interpreter = "node"
	case "bash", "sh":
		interpreter = "bash"
	case "ruby":
		interpreter = "ruby"
	case "perl":
		interpreter = "perl"
	default:
		return nil, fmt.Errorf("unsupported script language: %s", job.Payload.Language)
	}
	
	// Write script to file
	scriptPath := filepath.Join(workDir, "script")
	if err := os.WriteFile(scriptPath, []byte(job.Payload.Script), 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}
	
	// Execute script
	args = append(args, scriptPath)
	cmd := exec.CommandContext(ctx, interpreter, args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), job.Payload.Env...)
	
	// Handle input data if provided
	if job.Payload.InputData != "" {
		cmd.Stdin = nil // Could pipe input data here if needed
	}
	
	output, err := cmd.CombinedOutput()
	
	result := &JobResult{
		JobID:      job.ID,
		AgentID:    GenerateAgentID(),
		Status:     JobStatusCompleted,
		Output:     string(output),
		ExitCode:   0,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}
	
	if err != nil {
		result.Status = JobStatusFailed
		result.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}
	
	// Collect output artifacts if specified
	if job.Payload.OutputPath != "" {
		outputPath := filepath.Join(workDir, job.Payload.OutputPath)
		if info, err := os.Stat(outputPath); err == nil {
			artifact := JobArtifact{
				Name: filepath.Base(outputPath),
				Path: outputPath,
				Size: info.Size(),
			}
			result.Artifacts = append(result.Artifacts, artifact)
		}
	}
	
	return result, nil
}

// executeWASMJob runs a WebAssembly job
func (je *JobExecutor) executeWASMJob(ctx context.Context, job *Job, workDir string) (*JobResult, error) {
	// This would require a WASM runtime like wasmtime or wasmer
	return nil, fmt.Errorf("WASM execution not yet implemented")
}

// GetActiveJobs returns the list of active job IDs
func (je *JobExecutor) GetActiveJobs() []string {
	je.mu.RLock()
	defer je.mu.RUnlock()
	
	jobs := make([]string, 0, len(je.activeJobs))
	for id := range je.activeJobs {
		jobs = append(jobs, id)
	}
	return jobs
}

// GetActiveJobCount returns the number of active jobs
func (je *JobExecutor) GetActiveJobCount() int {
	je.mu.RLock()
	defer je.mu.RUnlock()
	return len(je.activeJobs)
}

// CancelJob cancels a running job
func (je *JobExecutor) CancelJob(jobID string) error {
	je.mu.RLock()
	activeJob, exists := je.activeJobs[jobID]
	je.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}
	
	activeJob.Cancel()
	return nil
}

// WaitForCompletion waits for all active jobs to complete
func (je *JobExecutor) WaitForCompletion(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if je.GetActiveJobCount() == 0 {
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for jobs to complete")
		}
	}
}

// checkDockerAvailable checks if Docker is available
func (je *JobExecutor) checkDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

// isURL checks if a string is a URL
func isURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")
}

// downloadFile downloads a file from a URL
func downloadFile(ctx context.Context, url, dest string) error {
	// This is a simplified implementation
	// In production, use proper HTTP client with timeouts and retries
	cmd := exec.CommandContext(ctx, "curl", "-L", "-o", dest, url)
	return cmd.Run()
} 