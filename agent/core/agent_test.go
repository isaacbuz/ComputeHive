package core

import (
	"context"
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	config := &Config{
		ControlPlaneURL:    "http://localhost:8000",
		HeartbeatInterval:  30 * time.Second,
		JobPollingInterval: 10 * time.Second,
		MetricsInterval:    60 * time.Second,
		MaxConcurrentJobs:  5,
		WorkDir:            "/tmp/computehive-test",
		EnableGPU:          false,
		LogLevel:           "info",
	}

	agent, err := NewAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent == nil {
		t.Fatal("Agent is nil")
	}

	if agent.config != config {
		t.Error("Agent config does not match")
	}

	if agent.status != AgentStatusInitializing {
		t.Errorf("Expected status %s, got %s", AgentStatusInitializing, agent.status)
	}
}

func TestAgentMetrics(t *testing.T) {
	metrics := NewAgentMetrics()

	// Test increment functions
	metrics.IncrementJobsStarted()
	if metrics.JobsStarted != 1 {
		t.Errorf("Expected JobsStarted to be 1, got %d", metrics.JobsStarted)
	}

	metrics.IncrementJobsCompleted()
	if metrics.JobsCompleted != 1 {
		t.Errorf("Expected JobsCompleted to be 1, got %d", metrics.JobsCompleted)
	}

	metrics.IncrementJobsFailed()
	if metrics.JobsFailed != 1 {
		t.Errorf("Expected JobsFailed to be 1, got %d", metrics.JobsFailed)
	}

	metrics.IncrementHeartbeatFailures()
	if metrics.HeartbeatFailures != 1 {
		t.Errorf("Expected HeartbeatFailures to be 1, got %d", metrics.HeartbeatFailures)
	}

	// Test snapshot
	snapshot := metrics.GetSnapshot()
	if snapshot.JobsStarted != metrics.JobsStarted {
		t.Error("Snapshot JobsStarted does not match")
	}
	if snapshot.JobsCompleted != metrics.JobsCompleted {
		t.Error("Snapshot JobsCompleted does not match")
	}
	if snapshot.JobsFailed != metrics.JobsFailed {
		t.Error("Snapshot JobsFailed does not match")
	}
	if snapshot.HeartbeatFailures != metrics.HeartbeatFailures {
		t.Error("Snapshot HeartbeatFailures does not match")
	}
}

func TestResourceMonitor(t *testing.T) {
	rm := NewResourceMonitor()
	if rm == nil {
		t.Fatal("ResourceMonitor is nil")
	}

	// Test getting resources
	resources := rm.GetResources()
	if resources == nil {
		t.Fatal("Resources is nil")
	}

	// Start monitoring in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go rm.Start(ctx)

	// Give it time to collect initial data
	time.Sleep(100 * time.Millisecond)

	// Get resources again
	resources = rm.GetResources()
	
	// Basic validation - CPU should have cores
	if resources.CPU.Cores <= 0 {
		t.Error("CPU cores should be greater than 0")
	}

	// Memory should have total
	if resources.Memory.Total <= 0 {
		t.Error("Memory total should be greater than 0")
	}
}

func TestJobExecutor(t *testing.T) {
	config := &Config{
		WorkDir:           "/tmp/computehive-test",
		MaxConcurrentJobs: 5,
	}

	executor := NewJobExecutor(config)
	if executor == nil {
		t.Fatal("JobExecutor is nil")
	}

	// Test active job tracking
	if executor.GetActiveJobCount() != 0 {
		t.Error("Expected 0 active jobs initially")
	}

	jobs := executor.GetActiveJobs()
	if len(jobs) != 0 {
		t.Error("Expected empty active jobs list initially")
	}
}

func TestPlatformInfo(t *testing.T) {
	platform := GetPlatformInfo()
	
	if platform.OS == "" {
		t.Error("Platform OS should not be empty")
	}

	if platform.Arch == "" {
		t.Error("Platform Arch should not be empty")
	}

	if platform.Hostname == "" {
		t.Error("Platform Hostname should not be empty")
	}
}

func TestGenerateAgentID(t *testing.T) {
	id1 := GenerateAgentID()
	id2 := GenerateAgentID()

	if id1 == "" {
		t.Error("Generated ID should not be empty")
	}

	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}
} 