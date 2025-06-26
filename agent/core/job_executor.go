package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// JobExecutor handles job execution
type JobExecutor struct {
	logger           *zap.Logger
	maxConcurrent    int
	activeJobs       sync.Map
	semaphore        chan struct{}
	workDir          string
	containerRuntime string
}

// NewJobExecutor creates a new job executor
func NewJobExecutor(maxConcurrent int, logger *zap.Logger) *JobExecutor {
	workDir := filepath.Join(os.TempDir(), "computehive", "jobs")
	os.MkdirAll(workDir, 0755)

	return &JobExecutor{
		logger:           logger,
		maxConcurrent:    maxConcurrent,
		semaphore:        make(chan struct{}, maxConcurrent),
		workDir:          workDir,
		containerRuntime: detectContainerRuntime(),
	}
}

// Execute runs a job and returns the result
func (je *JobExecutor) Execute(ctx context.Context, job *Job) (*JobResult, error) {
	// Acquire semaphore
	select {
	case je.semaphore <- struct{}{}:
		defer func() { <-je.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Track active job
	je.activeJobs.Store(job.ID, job)
	defer je.activeJobs.Delete(job.ID)

	// Create job directory
	jobDir := filepath.Join(je.workDir, job.ID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create job directory: %w", err)
	}
	defer os.RemoveAll(jobDir) // Clean up after execution

	// Start timing
	startTime := time.Now()

	// Execute based on runtime
	var result *JobResult
	var err error

	switch job.Payload.Runtime {
	case "docker":
		result, err = je.executeDockerJob(ctx, job, jobDir)
	case "wasm":
		result, err = je.executeWasmJob(ctx, job, jobDir)
	case "native":
		result, err = je.executeNativeJob(ctx, job, jobDir)
	default:
		err = fmt.Errorf("unsupported runtime: %s", job.Payload.Runtime)
	}

	if err != nil {
		return nil, err
	}

	// Complete result
	result.JobID = job.ID
	result.StartTime = startTime
	result.EndTime = time.Now()

	return result, nil
}

// executeDockerJob executes a job using Docker
func (je *JobExecutor) executeDockerJob(ctx context.Context, job *Job, jobDir string) (*JobResult, error) {
	if je.containerRuntime == "" {
		return nil, fmt.Errorf("no container runtime available")
	}

	// Write input data if provided
	var inputFile string
	if len(job.Payload.InputData) > 0 {
		inputFile = filepath.Join(jobDir, "input")
		if err := os.WriteFile(inputFile, job.Payload.InputData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write input data: %w", err)
		}
	}

	// Prepare output file
	outputFile := filepath.Join(jobDir, "output")

	// Build docker command
	args := []string{
		"run", "--rm",
		"--cpus", fmt.Sprintf("%d", job.Requirements.CPUCores),
		"--memory", fmt.Sprintf("%dm", job.Requirements.MemoryMB),
		"-v", fmt.Sprintf("%s:/workspace", jobDir),
		"--workdir", "/workspace",
	}

	// Add GPU support if requested
	if job.Requirements.GPUCount > 0 && je.hasNvidiaRuntime() {
		args = append(args, "--runtime", "nvidia")
		args = append(args, "--gpus", fmt.Sprintf("%d", job.Requirements.GPUCount))
	}

	// Add environment variables
	for k, v := range job.Payload.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add image and command
	args = append(args, job.Payload.Image)
	args = append(args, job.Payload.Command...)

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, job.Timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, je.containerRuntime, args...)
	
	// Capture output
	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	errorPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Read output
	var outputData []byte
	var logs string
	
	go func() {
		output, _ := io.ReadAll(outputPipe)
		outputData = output
	}()
	
	go func() {
		stderr, _ := io.ReadAll(errorPipe)
		logs = string(stderr)
	}()

	// Wait for completion
	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("container execution failed: %w", err)
		}
	}

	// Read output file if specified
	if job.Payload.OutputPath != "" {
		outputFilePath := filepath.Join(jobDir, job.Payload.OutputPath)
		if data, err := os.ReadFile(outputFilePath); err == nil {
			outputData = data
		}
	}

	// Calculate output hash
	hash := sha256.Sum256(outputData)
	outputHash := hex.EncodeToString(hash[:])

	// Get resource usage from container stats
	resourceUsage := je.getContainerResourceUsage(job.ID)

	return &JobResult{
		Status:       "completed",
		OutputData:   outputData,
		OutputHash:   outputHash,
		ResourceUsed: resourceUsage,
		ExitCode:     exitCode,
		Logs:         logs,
	}, nil
}

// executeWasmJob executes a WebAssembly job
func (je *JobExecutor) executeWasmJob(ctx context.Context, job *Job, jobDir string) (*JobResult, error) {
	// Check for wasmtime or wasmer
	wasmRuntime := ""
	if _, err := exec.LookPath("wasmtime"); err == nil {
		wasmRuntime = "wasmtime"
	} else if _, err := exec.LookPath("wasmer"); err == nil {
		wasmRuntime = "wasmer"
	} else {
		return nil, fmt.Errorf("no WASM runtime available (wasmtime or wasmer required)")
	}
	
	// Write WASM binary if provided
	wasmFile := filepath.Join(jobDir, "module.wasm")
	if len(job.Payload.WasmBinary) > 0 {
		if err := os.WriteFile(wasmFile, job.Payload.WasmBinary, 0644); err != nil {
			return nil, fmt.Errorf("failed to write WASM binary: %w", err)
		}
	} else if job.Payload.WasmPath != "" {
		wasmFile = job.Payload.WasmPath
	} else {
		return nil, fmt.Errorf("no WASM binary or path provided")
	}
	
	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, job.Timeout)
	defer cancel()
	
	var cmd *exec.Cmd
	if wasmRuntime == "wasmtime" {
		args := []string{wasmFile}
		// Add WASI directories
		if job.Payload.WasiPreopens != nil {
			for host, guest := range job.Payload.WasiPreopens {
				args = append(args, "--dir", fmt.Sprintf("%s::%s", guest, host))
			}
		}
		// Add environment variables
		for k, v := range job.Payload.Environment {
			args = append(args, "--env", fmt.Sprintf("%s=%s", k, v))
		}
		// Add command arguments
		args = append(args, "--")
		args = append(args, job.Payload.Command...)
		
		cmd = exec.CommandContext(cmdCtx, wasmRuntime, args...)
	} else { // wasmer
		args := []string{"run", wasmFile}
		// Add WASI directories
		if job.Payload.WasiPreopens != nil {
			for host, guest := range job.Payload.WasiPreopens {
				args = append(args, "--mapdir", fmt.Sprintf("%s:%s", guest, host))
			}
		}
		// Add environment variables
		for k, v := range job.Payload.Environment {
			args = append(args, "--env", fmt.Sprintf("%s=%s", k, v))
		}
		// Add command arguments
		args = append(args, "--")
		args = append(args, job.Payload.Command...)
		
		cmd = exec.CommandContext(cmdCtx, wasmRuntime, args...)
	}
	
	// Set working directory
	cmd.Dir = jobDir
	
	// Write input data if provided
	if len(job.Payload.InputData) > 0 {
		stdinPipe, err := cmd.StdinPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
		}
		go func() {
			defer stdinPipe.Close()
			stdinPipe.Write(job.Payload.InputData)
		}()
	}
	
	// Execute command
	output, err := cmd.CombinedOutput()
	
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("WASM execution failed: %w", err)
		}
	}
	
	// Calculate output hash
	hash := sha256.Sum256(output)
	outputHash := hex.EncodeToString(hash[:])
	
	// Get resource usage (simplified)
	resourceUsage := ResourceUsage{
		CPUPercent:    20.0,
		MemoryPercent: 15.0,
		MemoryUsedMB:  150,
	}
	
	return &JobResult{
		Status:       "completed",
		OutputData:   output,
		OutputHash:   outputHash,
		ResourceUsed: resourceUsage,
		ExitCode:     exitCode,
		Logs:         string(output),
	}, nil
}

// executeNativeJob executes a native binary job
func (je *JobExecutor) executeNativeJob(ctx context.Context, job *Job, jobDir string) (*JobResult, error) {
	if len(job.Payload.Command) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, job.Timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, job.Payload.Command[0], job.Payload.Command[1:]...)
	cmd.Dir = jobDir
	
	// Set environment
	cmd.Env = os.Environ()
	for k, v := range job.Payload.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Write input data
	if len(job.Payload.InputData) > 0 {
		cmd.Stdin = os.NewFile(0, "stdin")
		go func() {
			cmd.Stdin.Write(job.Payload.InputData)
			cmd.Stdin.Close()
		}()
	}

	// Execute command
	output, err := cmd.CombinedOutput()
	
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}

	// Calculate output hash
	hash := sha256.Sum256(output)
	outputHash := hex.EncodeToString(hash[:])

	// Get resource usage (simplified)
	resourceUsage := ResourceUsage{
		CPUPercent:    25.0,
		MemoryPercent: 10.0,
		MemoryUsedMB:  100,
	}

	return &JobResult{
		Status:       "completed",
		OutputData:   output,
		OutputHash:   outputHash,
		ResourceUsed: resourceUsage,
		ExitCode:     exitCode,
		Logs:         string(output),
	}, nil
}

// Shutdown gracefully shuts down the job executor
func (je *JobExecutor) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Wait for active jobs to complete
	done := make(chan struct{})
	go func() {
		defer close(done)
		
		for {
			count := 0
			je.activeJobs.Range(func(_, _ interface{}) bool {
				count++
				return true
			})
			
			if count == 0 {
				return
			}
			
			je.logger.Info("Waiting for active jobs to complete", zap.Int("count", count))
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// Force stop active jobs
		je.activeJobs.Range(func(key, value interface{}) bool {
			jobID := key.(string)
			job := value.(*Job)
			je.logger.Warn("Force stopping job", zap.String("job_id", jobID))
			
			// Force stop based on runtime
			switch job.Payload.Runtime {
			case "docker":
				// Force stop docker container
				cmd := exec.Command(je.containerRuntime, "kill", jobID)
				if err := cmd.Run(); err != nil {
					je.logger.Error("Failed to kill container", 
						zap.String("job_id", jobID),
						zap.Error(err))
				}
				// Remove container
				cmd = exec.Command(je.containerRuntime, "rm", "-f", jobID)
				cmd.Run()
			case "native":
				// For native processes, we would need to track PIDs
				// This is a simplified approach
				je.logger.Info("Native job force stop not fully implemented", 
					zap.String("job_id", jobID))
			case "wasm":
				// WASM processes should terminate with context cancellation
				je.logger.Info("WASM job should terminate with context", 
					zap.String("job_id", jobID))
			}
			
			return true
		})
		return ctx.Err()
	}
}

// GetActiveJobs returns the list of currently active jobs
func (je *JobExecutor) GetActiveJobs() []string {
	var jobs []string
	je.activeJobs.Range(func(key, _ interface{}) bool {
		jobs = append(jobs, key.(string))
		return true
	})
	return jobs
}

// Helper functions

func detectContainerRuntime() string {
	// Check for Docker
	if _, err := exec.LookPath("docker"); err == nil {
		return "docker"
	}
	
	// Check for Podman
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman"
	}
	
	// Check for containerd
	if _, err := exec.LookPath("nerdctl"); err == nil {
		return "nerdctl"
	}
	
	return ""
}

func (je *JobExecutor) hasNvidiaRuntime() bool {
	// Check if nvidia-container-runtime is available
	cmd := exec.Command(je.containerRuntime, "info", "-f", "{{.Runtimes}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return string(output) != "" && (string(output) == "nvidia" || string(output) == "nvidia-container-runtime")
}

// getContainerResourceUsage gets actual resource usage from container
func (je *JobExecutor) getContainerResourceUsage(containerID string) ResourceUsage {
	// Default values
	usage := ResourceUsage{
		CPUPercent:    0.0,
		MemoryPercent: 0.0,
		MemoryUsedMB:  0,
	}
	
	// Try to get container stats
	cmd := exec.Command(je.containerRuntime, "stats", "--no-stream", "--format", "json", containerID)
	output, err := cmd.Output()
	if err != nil {
		// Return default values if stats not available
		return usage
	}
	
	// Parse stats based on runtime
	if je.containerRuntime == "docker" {
		// Docker returns JSON with CPU and memory stats
		var stats struct {
			CPUPerc string `json:"CPUPerc"`
			MemPerc string `json:"MemPerc"`
			MemUsage string `json:"MemUsage"`
		}
		if err := json.Unmarshal(output, &stats); err == nil {
			// Parse CPU percentage
			if cpuStr := strings.TrimSuffix(stats.CPUPerc, "%"); cpuStr != "" {
				if cpu, err := strconv.ParseFloat(cpuStr, 64); err == nil {
					usage.CPUPercent = cpu
				}
			}
			// Parse memory percentage
			if memStr := strings.TrimSuffix(stats.MemPerc, "%"); memStr != "" {
				if mem, err := strconv.ParseFloat(memStr, 64); err == nil {
					usage.MemoryPercent = mem
				}
			}
			// Parse memory usage
			if parts := strings.Split(stats.MemUsage, "/"); len(parts) > 0 {
				memStr := strings.TrimSpace(parts[0])
				// Convert to MB
				if strings.HasSuffix(memStr, "MiB") {
					if mem, err := strconv.ParseFloat(strings.TrimSuffix(memStr, "MiB"), 64); err == nil {
						usage.MemoryUsedMB = uint64(mem)
					}
				} else if strings.HasSuffix(memStr, "GiB") {
					if mem, err := strconv.ParseFloat(strings.TrimSuffix(memStr, "GiB"), 64); err == nil {
						usage.MemoryUsedMB = uint64(mem * 1024)
					}
				}
			}
		}
	}
	
	return usage
} 