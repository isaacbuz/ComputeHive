package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

// Job represents a compute job
type Job struct {
	ID               string               `json:"id"`
	UserID           string               `json:"user_id"`
	Type             string               `json:"type"`
	Status           string               `json:"status"`
	Priority         int                  `json:"priority"`
	Requirements     ResourceRequirements `json:"requirements"`
	Payload          json.RawMessage      `json:"payload"`
	AssignedAgentID  string               `json:"assigned_agent_id,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
	ScheduledAt      *time.Time           `json:"scheduled_at,omitempty"`
	StartedAt        *time.Time           `json:"started_at,omitempty"`
	CompletedAt      *time.Time           `json:"completed_at,omitempty"`
	EstimatedCost    float64              `json:"estimated_cost"`
	ActualCost       float64              `json:"actual_cost,omitempty"`
	MaxRetries       int                  `json:"max_retries"`
	RetryCount       int                  `json:"retry_count"`
	Timeout          time.Duration        `json:"timeout"`
	SLARequirements  *SLARequirements     `json:"sla_requirements,omitempty"`
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

// SLARequirements defines service level agreement requirements
type SLARequirements struct {
	MaxLatencyMs     int     `json:"max_latency_ms"`
	MinAvailability  float64 `json:"min_availability"`
	MaxCostPerHour   float64 `json:"max_cost_per_hour"`
	PreferredRegions []string `json:"preferred_regions,omitempty"`
}

// Agent represents a compute agent
type Agent struct {
	ID           string              `json:"id"`
	Status       string              `json:"status"`
	Resources    AgentResources      `json:"resources"`
	Capabilities []string            `json:"capabilities"`
	Location     string              `json:"location"`
	PricePerHour map[string]float64  `json:"price_per_hour"`
	Reputation   float64             `json:"reputation"`
	LastSeen     time.Time           `json:"last_seen"`
	ActiveJobs   []string            `json:"active_jobs"`
}

// AgentResources represents available resources on an agent
type AgentResources struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	GPUs    []GPUInfo   `json:"gpus"`
	Storage StorageInfo `json:"storage"`
	Network NetworkInfo `json:"network"`
}

// Resource info types
type CPUInfo struct {
	Cores     int     `json:"cores"`
	Available int     `json:"available"`
	Usage     float64 `json:"usage"`
}

type MemoryInfo struct {
	TotalMB     int `json:"total_mb"`
	AvailableMB int `json:"available_mb"`
}

type GPUInfo struct {
	ID       string `json:"id"`
	Model    string `json:"model"`
	MemoryMB int    `json:"memory_mb"`
	InUse    bool   `json:"in_use"`
}

type StorageInfo struct {
	TotalMB     int `json:"total_mb"`
	AvailableMB int `json:"available_mb"`
}

type NetworkInfo struct {
	BandwidthMbps int `json:"bandwidth_mbps"`
}

// SchedulerService handles job scheduling and resource allocation
type SchedulerService struct {
	jobs       map[string]*Job
	agents     map[string]*Agent
	jobQueue   []*Job
	mu         sync.RWMutex
	nats       *nats.Conn
	httpClient *http.Client
	
	// Metrics
	jobsScheduled   prometheus.Counter
	jobsCompleted   prometheus.Counter
	jobsFailed      prometheus.Counter
	schedulingTime  prometheus.Histogram
	queueLength     prometheus.Gauge
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService() (*SchedulerService, error) {
	// Connect to NATS for event streaming
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	
	s := &SchedulerService{
		jobs:       make(map[string]*Job),
		agents:     make(map[string]*Agent),
		jobQueue:   make([]*Job, 0),
		nats:       nc,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		
		// Initialize metrics
		jobsScheduled: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "scheduler_jobs_scheduled_total",
			Help: "Total number of jobs scheduled",
		}),
		jobsCompleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "scheduler_jobs_completed_total",
			Help: "Total number of jobs completed",
		}),
		jobsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "scheduler_jobs_failed_total",
			Help: "Total number of jobs failed",
		}),
		schedulingTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "scheduler_scheduling_duration_seconds",
			Help:    "Time taken to schedule a job",
			Buckets: prometheus.DefBuckets,
		}),
		queueLength: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "scheduler_queue_length",
			Help: "Current number of jobs in queue",
		}),
	}
	
	// Register metrics
	prometheus.MustRegister(s.jobsScheduled, s.jobsCompleted, s.jobsFailed, s.schedulingTime, s.queueLength)
	
	// Subscribe to agent events
	s.subscribeToAgentEvents()
	
	return s, nil
}

// SubmitJob handles job submission
func (s *SchedulerService) SubmitJob(w http.ResponseWriter, r *http.Request) {
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Generate job ID
	job.ID = generateID()
	job.Status = "pending"
	job.CreatedAt = time.Now()
	
	// Extract user ID from JWT token
	claims := r.Context().Value("claims").(*Claims)
	job.UserID = claims.UserID
	
	// Validate job requirements
	if err := s.validateJobRequirements(&job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Estimate cost based on requirements and market rates
	job.EstimatedCost = s.estimateJobCost(&job)
	
	// Store job
	s.mu.Lock()
	s.jobs[job.ID] = &job
	s.jobQueue = append(s.jobQueue, &job)
	s.queueLength.Set(float64(len(s.jobQueue)))
	s.mu.Unlock()
	
	// Trigger scheduling
	go s.scheduleJob(&job)
	
	// Publish job created event
	s.publishJobEvent("job.created", &job)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// GetJob retrieves job details
func (s *SchedulerService) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	
	s.mu.RLock()
	job, exists := s.jobs[jobID]
	s.mu.RUnlock()
	
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	// Check authorization
	claims := r.Context().Value("claims").(*Claims)
	if job.UserID != claims.UserID && claims.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// ListJobs lists jobs for a user
func (s *SchedulerService) ListJobs(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	
	var userJobs []*Job
	s.mu.RLock()
	for _, job := range s.jobs {
		if job.UserID == claims.UserID || claims.Role == "admin" {
			userJobs = append(userJobs, job)
		}
	}
	s.mu.RUnlock()
	
	// Sort by creation time
	sort.Slice(userJobs, func(i, j int) bool {
		return userJobs[i].CreatedAt.After(userJobs[j].CreatedAt)
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userJobs)
}

// CancelJob cancels a pending or running job
func (s *SchedulerService) CancelJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]
	
	s.mu.Lock()
	job, exists := s.jobs[jobID]
	if !exists {
		s.mu.Unlock()
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	// Check authorization
	claims := r.Context().Value("claims").(*Claims)
	if job.UserID != claims.UserID && claims.Role != "admin" {
		s.mu.Unlock()
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}
	
	// Update job status
	job.Status = "cancelled"
	now := time.Now()
	job.CompletedAt = &now
	s.mu.Unlock()
	
	// Notify assigned agent if any
	if job.AssignedAgentID != "" {
		s.notifyAgentJobCancelled(job.AssignedAgentID, jobID)
	}
	
	// Publish cancellation event
	s.publishJobEvent("job.cancelled", job)
	
	w.WriteHeader(http.StatusNoContent)
}

// scheduleJob finds the best agent for a job and assigns it
func (s *SchedulerService) scheduleJob(job *Job) {
	timer := prometheus.NewTimer(s.schedulingTime)
	defer timer.ObserveDuration()
	
	// Find suitable agents
	agents := s.findSuitableAgents(job)
	if len(agents) == 0 {
		log.Printf("No suitable agents found for job %s", job.ID)
		s.requeueJob(job)
		return
	}
	
	// Score and rank agents
	scoredAgents := s.scoreAgents(agents, job)
	
	// Try to assign to the best agent
	for _, sa := range scoredAgents {
		if s.assignJobToAgent(job, sa.agent) {
			s.jobsScheduled.Inc()
			return
		}
	}
	
	// If no agent accepted, requeue
	s.requeueJob(job)
}

// findSuitableAgents finds agents that meet job requirements
func (s *SchedulerService) findSuitableAgents(job *Job) []*Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var suitable []*Agent
	for _, agent := range s.agents {
		if s.agentMeetsRequirements(agent, job) {
			suitable = append(suitable, agent)
		}
	}
	
	return suitable
}

// agentMeetsRequirements checks if an agent can handle a job
func (s *SchedulerService) agentMeetsRequirements(agent *Agent, job *Job) bool {
	// Check agent status
	if agent.Status != "active" {
		return false
	}
	
	// Check last seen time (agent should be recently active)
	if time.Since(agent.LastSeen) > 2*time.Minute {
		return false
	}
	
	// Check CPU requirements
	if agent.Resources.CPU.Available < job.Requirements.CPUCores {
		return false
	}
	
	// Check memory requirements
	if agent.Resources.Memory.AvailableMB < job.Requirements.MemoryMB {
		return false
	}
	
	// Check GPU requirements
	if job.Requirements.GPUCount > 0 {
		availableGPUs := 0
		for _, gpu := range agent.Resources.GPUs {
			if !gpu.InUse {
				if job.Requirements.GPUType == "" || gpu.Model == job.Requirements.GPUType {
					availableGPUs++
				}
			}
		}
		if availableGPUs < job.Requirements.GPUCount {
			return false
		}
	}
	
	// Check storage requirements
	if agent.Resources.Storage.AvailableMB < job.Requirements.StorageMB {
		return false
	}
	
	// Check capabilities
	for _, required := range job.Requirements.Capabilities {
		found := false
		for _, capability := range agent.Capabilities {
			if capability == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check SLA requirements
	if job.SLARequirements != nil {
		// Check cost
		agentHourlyRate := s.calculateAgentHourlyRate(agent, job)
		if agentHourlyRate > job.SLARequirements.MaxCostPerHour {
			return false
		}
		
		// Check location preferences
		if len(job.SLARequirements.PreferredRegions) > 0 {
			found := false
			for _, region := range job.SLARequirements.PreferredRegions {
				if agent.Location == region {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	
	return true
}

// scoreAgents scores agents based on various factors
func (s *SchedulerService) scoreAgents(agents []*Agent, job *Job) []scoredAgent {
	scored := make([]scoredAgent, len(agents))
	
	for i, agent := range agents {
		score := 0.0
		
		// Factor 1: Cost (lower is better)
		hourlyRate := s.calculateAgentHourlyRate(agent, job)
		costScore := 1.0 / (1.0 + hourlyRate/100.0) // Normalize cost impact
		score += costScore * 0.3
		
		// Factor 2: Reputation
		score += agent.Reputation * 0.3
		
		// Factor 3: Resource availability (more available is better)
		availabilityScore := float64(agent.Resources.CPU.Available) / float64(agent.Resources.CPU.Cores)
		score += availabilityScore * 0.2
		
		// Factor 4: Current load (fewer active jobs is better)
		loadScore := 1.0 / (1.0 + float64(len(agent.ActiveJobs)))
		score += loadScore * 0.2
		
		scored[i] = scoredAgent{
			agent: agent,
			score: score,
		}
	}
	
	// Sort by score (highest first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	
	return scored
}

type scoredAgent struct {
	agent *Agent
	score float64
}

// calculateAgentHourlyRate calculates the hourly rate for a job on an agent
func (s *SchedulerService) calculateAgentHourlyRate(agent *Agent, job *Job) float64 {
	baseCPURate := agent.PricePerHour["cpu"] * float64(job.Requirements.CPUCores)
	baseMemoryRate := agent.PricePerHour["memory"] * float64(job.Requirements.MemoryMB) / 1024.0
	baseStorageRate := agent.PricePerHour["storage"] * float64(job.Requirements.StorageMB) / 1024.0
	
	totalRate := baseCPURate + baseMemoryRate + baseStorageRate
	
	// Add GPU rate if needed
	if job.Requirements.GPUCount > 0 {
		gpuRate := agent.PricePerHour["gpu"] * float64(job.Requirements.GPUCount)
		totalRate += gpuRate
	}
	
	return totalRate
}

// assignJobToAgent attempts to assign a job to an agent
func (s *SchedulerService) assignJobToAgent(job *Job, agent *Agent) bool {
	// Send assignment request to agent
	assignment := map[string]interface{}{
		"job_id": job.ID,
		"job":    job,
	}
	
	data, _ := json.Marshal(assignment)
	msg, err := s.nats.Request(fmt.Sprintf("agent.%s.assign", agent.ID), data, 5*time.Second)
	if err != nil {
		log.Printf("Failed to assign job %s to agent %s: %v", job.ID, agent.ID, err)
		return false
	}
	
	var response map[string]bool
	if err := json.Unmarshal(msg.Data, &response); err != nil || !response["accepted"] {
		return false
	}
	
	// Update job status
	s.mu.Lock()
	job.Status = "scheduled"
	job.AssignedAgentID = agent.ID
	now := time.Now()
	job.ScheduledAt = &now
	
	// Update agent's active jobs
	agent.ActiveJobs = append(agent.ActiveJobs, job.ID)
	s.mu.Unlock()
	
	// Publish assignment event
	s.publishJobEvent("job.scheduled", job)
	
	return true
}

// requeueJob puts a job back in the queue for retry
func (s *SchedulerService) requeueJob(job *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	job.RetryCount++
	if job.RetryCount > job.MaxRetries {
		job.Status = "failed"
		s.jobsFailed.Inc()
		s.publishJobEvent("job.failed", job)
		return
	}
	
	// Add back to queue with exponential backoff
	go func() {
		backoff := time.Duration(math.Pow(2, float64(job.RetryCount))) * time.Second
		time.Sleep(backoff)
		
		s.mu.Lock()
		s.jobQueue = append(s.jobQueue, job)
		s.queueLength.Set(float64(len(s.jobQueue)))
		s.mu.Unlock()
	}()
}

// Event handling

func (s *SchedulerService) subscribeToAgentEvents() {
	// Subscribe to agent heartbeats
	s.nats.Subscribe("agent.heartbeat", func(msg *nats.Msg) {
		var heartbeat map[string]interface{}
		if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
			return
		}
		
		agentID := heartbeat["agent_id"].(string)
		s.updateAgentStatus(agentID, heartbeat)
	})
	
	// Subscribe to job results
	s.nats.Subscribe("job.result", func(msg *nats.Msg) {
		var result map[string]interface{}
		if err := json.Unmarshal(msg.Data, &result); err != nil {
			return
		}
		
		jobID := result["job_id"].(string)
		s.handleJobResult(jobID, result)
	})
}

func (s *SchedulerService) updateAgentStatus(agentID string, heartbeat map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	agent, exists := s.agents[agentID]
	if !exists {
		// New agent registration
		agent = &Agent{
			ID:           agentID,
			PricePerHour: make(map[string]float64),
			ActiveJobs:   make([]string, 0),
		}
		s.agents[agentID] = agent
	}
	
	// Update agent info from heartbeat
	agent.Status = heartbeat["status"].(string)
	agent.LastSeen = time.Now()
	
	// Update resources if provided
	if resources, ok := heartbeat["resources"].(map[string]interface{}); ok {
		// Parse and update resources (simplified)
		// In production, this would properly unmarshal the resources
	}
}

func (s *SchedulerService) handleJobResult(jobID string, result map[string]interface{}) {
	s.mu.Lock()
	job, exists := s.jobs[jobID]
	if !exists {
		s.mu.Unlock()
		return
	}
	
	// Update job status
	status := result["status"].(string)
	job.Status = status
	now := time.Now()
	
	if status == "completed" {
		job.CompletedAt = &now
		s.jobsCompleted.Inc()
	} else if status == "failed" {
		job.CompletedAt = &now
		s.jobsFailed.Inc()
	}
	
	// Remove from agent's active jobs
	if agent, exists := s.agents[job.AssignedAgentID]; exists {
		newActiveJobs := make([]string, 0)
		for _, activeJobID := range agent.ActiveJobs {
			if activeJobID != jobID {
				newActiveJobs = append(newActiveJobs, activeJobID)
			}
		}
		agent.ActiveJobs = newActiveJobs
	}
	
	s.mu.Unlock()
	
	// Publish completion event
	s.publishJobEvent(fmt.Sprintf("job.%s", status), job)
}

func (s *SchedulerService) publishJobEvent(event string, job *Job) {
	data, _ := json.Marshal(job)
	s.nats.Publish(event, data)
}

func (s *SchedulerService) notifyAgentJobCancelled(agentID, jobID string) {
	notification := map[string]string{
		"job_id": jobID,
		"action": "cancel",
	}
	data, _ := json.Marshal(notification)
	s.nats.Publish(fmt.Sprintf("agent.%s.job.cancel", agentID), data)
}

// validateJobRequirements validates job requirements
func (s *SchedulerService) validateJobRequirements(job *Job) error {
	if job.Requirements.CPUCores <= 0 {
		return fmt.Errorf("CPU cores must be positive")
	}
	if job.Requirements.MemoryMB <= 0 {
		return fmt.Errorf("memory must be positive")
	}
	if job.Timeout <= 0 {
		job.Timeout = 1 * time.Hour // Default timeout
	}
	if job.MaxRetries <= 0 {
		job.MaxRetries = 3 // Default retries
	}
	if job.Priority < 0 || job.Priority > 10 {
		job.Priority = 5 // Default priority
	}
	return nil
}

// estimateJobCost estimates the cost of running a job
func (s *SchedulerService) estimateJobCost(job *Job) float64 {
	// Base estimates (would be more sophisticated in production)
	cpuHourlyRate := 0.05 * float64(job.Requirements.CPUCores)
	memoryHourlyRate := 0.01 * float64(job.Requirements.MemoryMB) / 1024.0
	storageHourlyRate := 0.001 * float64(job.Requirements.StorageMB) / 1024.0
	
	baseRate := cpuHourlyRate + memoryHourlyRate + storageHourlyRate
	
	// Add GPU premium
	if job.Requirements.GPUCount > 0 {
		gpuRate := 0.5 * float64(job.Requirements.GPUCount) // $0.50 per GPU hour
		baseRate += gpuRate
	}
	
	// Estimate job duration (simplified)
	estimatedHours := float64(job.Timeout) / float64(time.Hour)
	
	return baseRate * estimatedHours
}

// Process job queue periodically
func (s *SchedulerService) processQueue() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		s.mu.Lock()
		if len(s.jobQueue) == 0 {
			s.mu.Unlock()
			continue
		}
		
		// Get jobs to process
		jobsToProcess := make([]*Job, len(s.jobQueue))
		copy(jobsToProcess, s.jobQueue)
		s.jobQueue = s.jobQueue[:0]
		s.queueLength.Set(0)
		s.mu.Unlock()
		
		// Schedule each job
		for _, job := range jobsToProcess {
			go s.scheduleJob(job)
		}
	}
}

// JWT Claims type
type Claims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Role     string   `json:"role"`
	Scopes   []string `json:"scopes"`
	jwt.RegisteredClaims
}

// Auth middleware
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple auth check - in production, validate JWT properly
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}
		
		// Mock claims for development
		claims := &Claims{
			UserID: "user-123",
			Role:   "user",
		}
		
		ctx := context.WithValue(r.Context(), "claims", claims)
		next(w, r.WithContext(ctx))
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	// Create scheduler service
	scheduler, err := NewSchedulerService()
	if err != nil {
		log.Fatalf("Failed to create scheduler service: %v", err)
	}
	
	// Start queue processor
	go scheduler.processQueue()
	
	// Setup routes
	router := mux.NewRouter()
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler())
	
	// Job endpoints
	router.HandleFunc("/api/v1/jobs", authMiddleware(scheduler.SubmitJob)).Methods("POST")
	router.HandleFunc("/api/v1/jobs", authMiddleware(scheduler.ListJobs)).Methods("GET")
	router.HandleFunc("/api/v1/jobs/{id}", authMiddleware(scheduler.GetJob)).Methods("GET")
	router.HandleFunc("/api/v1/jobs/{id}/cancel", authMiddleware(scheduler.CancelJob)).Methods("POST")
	
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://computehive.io"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	
	handler := c.Handler(router)
	
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	
	log.Printf("Scheduler service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 