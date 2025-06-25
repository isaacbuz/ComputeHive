package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

// Resource represents a compute resource
type Resource struct {
	ID                    string                 `json:"id"`
	AgentID               string                 `json:"agent_id"`
	Type                  string                 `json:"type"` // cpu, gpu, storage, network
	Status                string                 `json:"status"` // available, allocated, maintenance
	TotalCapacity         map[string]interface{} `json:"total_capacity"`
	AllocatedCapacity     map[string]interface{} `json:"allocated_capacity"`
	AvailableCapacity     map[string]interface{} `json:"available_capacity"`
	Metadata              map[string]string      `json:"metadata"`
	LastUpdated           time.Time              `json:"last_updated"`
}

// ResourceAllocation represents an allocation of resources
type ResourceAllocation struct {
	ID              string                 `json:"id"`
	ResourceID      string                 `json:"resource_id"`
	JobID           string                 `json:"job_id"`
	UserID          string                 `json:"user_id"`
	AllocatedAmount map[string]interface{} `json:"allocated_amount"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         *time.Time             `json:"end_time,omitempty"`
	Status          string                 `json:"status"` // active, completed, cancelled
}

// ResourceService manages compute resources
type ResourceService struct {
	resources      map[string]*Resource
	allocations    map[string]*ResourceAllocation
	mu             sync.RWMutex
	nats           *nats.Conn
	
	// Metrics
	totalResources     *prometheus.GaugeVec
	allocatedResources *prometheus.GaugeVec
	allocationDuration *prometheus.HistogramVec
}

// NewResourceService creates a new resource service
func NewResourceService() (*ResourceService, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	
	s := &ResourceService{
		resources:   make(map[string]*Resource),
		allocations: make(map[string]*ResourceAllocation),
		nats:        nc,
		
		totalResources: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "resource_service_total_resources",
				Help: "Total resources by type",
			},
			[]string{"type", "metric"},
		),
		allocatedResources: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "resource_service_allocated_resources",
				Help: "Allocated resources by type",
			},
			[]string{"type", "metric"},
		),
		allocationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "resource_service_allocation_duration_seconds",
				Help:    "Resource allocation duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"type"},
		),
	}
	
	prometheus.MustRegister(
		s.totalResources,
		s.allocatedResources,
		s.allocationDuration,
	)
	
	// Subscribe to events
	s.subscribeToEvents()
	
	// Start background workers
	go s.resourceMonitor()
	go s.allocationCleanup()
	
	return s, nil
}

// RegisterResource registers a new resource
func (s *ResourceService) RegisterResource(w http.ResponseWriter, r *http.Request) {
	var resource Resource
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	resource.ID = generateID()
	resource.Status = "available"
	resource.LastUpdated = time.Now()
	
	// Calculate available capacity
	resource.AvailableCapacity = make(map[string]interface{})
	for k, v := range resource.TotalCapacity {
		if allocated, ok := resource.AllocatedCapacity[k]; ok {
			if vFloat, ok := v.(float64); ok {
				if allocFloat, ok := allocated.(float64); ok {
					resource.AvailableCapacity[k] = vFloat - allocFloat
				}
			}
		} else {
			resource.AvailableCapacity[k] = v
		}
	}
	
	s.mu.Lock()
	s.resources[resource.ID] = &resource
	s.mu.Unlock()
	
	// Update metrics
	s.updateResourceMetrics()
	
	// Publish resource registered event
	s.publishResourceEvent("resource.registered", &resource)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

// GetResources returns all resources or filtered by query params
func (s *ResourceService) GetResources(w http.ResponseWriter, r *http.Request) {
	resourceType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	agentID := r.URL.Query().Get("agent_id")
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	resources := make([]*Resource, 0)
	for _, resource := range s.resources {
		// Apply filters
		if resourceType != "" && resource.Type != resourceType {
			continue
		}
		if status != "" && resource.Status != status {
			continue
		}
		if agentID != "" && resource.AgentID != agentID {
			continue
		}
		
		resources = append(resources, resource)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

// AllocateResource allocates resources for a job
func (s *ResourceService) AllocateResource(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ResourceID string                 `json:"resource_id"`
		JobID      string                 `json:"job_id"`
		UserID     string                 `json:"user_id"`
		Amount     map[string]interface{} `json:"amount"`
		Duration   int                    `json:"duration"` // in seconds
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check if resource exists and is available
	resource, exists := s.resources[req.ResourceID]
	if !exists {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}
	
	// Check if sufficient capacity is available
	for k, v := range req.Amount {
		available, ok := resource.AvailableCapacity[k]
		if !ok {
			http.Error(w, fmt.Sprintf("Resource metric %s not found", k), http.StatusBadRequest)
			return
		}
		
		if vFloat, ok := v.(float64); ok {
			if availFloat, ok := available.(float64); ok {
				if vFloat > availFloat {
					http.Error(w, fmt.Sprintf("Insufficient %s capacity", k), http.StatusConflict)
					return
				}
			}
		}
	}
	
	// Create allocation
	allocation := &ResourceAllocation{
		ID:              generateID(),
		ResourceID:      req.ResourceID,
		JobID:           req.JobID,
		UserID:          req.UserID,
		AllocatedAmount: req.Amount,
		StartTime:       time.Now(),
		Status:          "active",
	}
	
	if req.Duration > 0 {
		endTime := time.Now().Add(time.Duration(req.Duration) * time.Second)
		allocation.EndTime = &endTime
	}
	
	// Update resource capacity
	for k, v := range req.Amount {
		if vFloat, ok := v.(float64); ok {
			if currentAlloc, exists := resource.AllocatedCapacity[k]; exists {
				if allocFloat, ok := currentAlloc.(float64); ok {
					resource.AllocatedCapacity[k] = allocFloat + vFloat
				}
			} else {
				resource.AllocatedCapacity[k] = vFloat
			}
			
			// Update available capacity
			if total, ok := resource.TotalCapacity[k].(float64); ok {
				if allocated, ok := resource.AllocatedCapacity[k].(float64); ok {
					resource.AvailableCapacity[k] = total - allocated
				}
			}
		}
	}
	
	resource.LastUpdated = time.Now()
	s.allocations[allocation.ID] = allocation
	
	// Update metrics
	s.updateResourceMetrics()
	
	// Publish allocation event
	s.publishAllocationEvent("allocation.created", allocation)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allocation)
}

// ReleaseResource releases an allocation
func (s *ResourceService) ReleaseResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allocationID := vars["id"]
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	allocation, exists := s.allocations[allocationID]
	if !exists {
		http.Error(w, "Allocation not found", http.StatusNotFound)
		return
	}
	
	if allocation.Status != "active" {
		http.Error(w, "Allocation is not active", http.StatusBadRequest)
		return
	}
	
	// Get resource
	resource, exists := s.resources[allocation.ResourceID]
	if !exists {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}
	
	// Release allocated capacity
	for k, v := range allocation.AllocatedAmount {
		if vFloat, ok := v.(float64); ok {
			if currentAlloc, exists := resource.AllocatedCapacity[k]; exists {
				if allocFloat, ok := currentAlloc.(float64); ok {
					resource.AllocatedCapacity[k] = allocFloat - vFloat
					
					// Update available capacity
					if total, ok := resource.TotalCapacity[k].(float64); ok {
						resource.AvailableCapacity[k] = total - resource.AllocatedCapacity[k].(float64)
					}
				}
			}
		}
	}
	
	// Update allocation status
	allocation.Status = "completed"
	now := time.Now()
	allocation.EndTime = &now
	resource.LastUpdated = now
	
	// Update metrics
	s.updateResourceMetrics()
	
	// Record allocation duration
	duration := now.Sub(allocation.StartTime).Seconds()
	s.allocationDuration.WithLabelValues(resource.Type).Observe(duration)
	
	// Publish release event
	s.publishAllocationEvent("allocation.released", allocation)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(allocation)
}

// GetAllocations returns allocations
func (s *ResourceService) GetAllocations(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	userID := r.URL.Query().Get("user_id")
	status := r.URL.Query().Get("status")
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	allocations := make([]*ResourceAllocation, 0)
	for _, allocation := range s.allocations {
		// Apply filters
		if jobID != "" && allocation.JobID != jobID {
			continue
		}
		if userID != "" && allocation.UserID != userID {
			continue
		}
		if status != "" && allocation.Status != status {
			continue
		}
		
		allocations = append(allocations, allocation)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allocations)
}

// Background workers

func (s *ResourceService) resourceMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		s.mu.RLock()
		for _, resource := range s.resources {
			// Check resource health
			if time.Since(resource.LastUpdated) > 5*time.Minute {
				log.Printf("Resource %s hasn't been updated in 5 minutes", resource.ID)
			}
		}
		s.mu.RUnlock()
	}
}

func (s *ResourceService) allocationCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		
		for id, allocation := range s.allocations {
			// Clean up expired allocations
			if allocation.EndTime != nil && allocation.EndTime.Before(now) && allocation.Status == "active" {
				// Release the allocation
				if resource, exists := s.resources[allocation.ResourceID]; exists {
					for k, v := range allocation.AllocatedAmount {
						if vFloat, ok := v.(float64); ok {
							if currentAlloc, exists := resource.AllocatedCapacity[k]; exists {
								if allocFloat, ok := currentAlloc.(float64); ok {
									resource.AllocatedCapacity[k] = allocFloat - vFloat
									
									// Update available capacity
									if total, ok := resource.TotalCapacity[k].(float64); ok {
										resource.AvailableCapacity[k] = total - resource.AllocatedCapacity[k].(float64)
									}
								}
							}
						}
					}
					resource.LastUpdated = now
				}
				
				allocation.Status = "completed"
				log.Printf("Auto-released expired allocation %s", id)
			}
		}
		
		s.mu.Unlock()
		
		// Update metrics after cleanup
		s.updateResourceMetrics()
	}
}

// Event handling

func (s *ResourceService) subscribeToEvents() {
	// Subscribe to agent heartbeats for resource updates
	s.nats.Subscribe("agent.heartbeat", func(msg *nats.Msg) {
		var heartbeat map[string]interface{}
		if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
			return
		}
		
		// Update resource information based on heartbeat
		if agentID, ok := heartbeat["agent_id"].(string); ok {
			s.updateAgentResources(agentID, heartbeat)
		}
	})
	
	// Subscribe to job events
	s.nats.Subscribe("job.completed", func(msg *nats.Msg) {
		var job map[string]interface{}
		if err := json.Unmarshal(msg.Data, &job); err != nil {
			return
		}
		
		// Release resources allocated to the job
		if jobID, ok := job["id"].(string); ok {
			s.releaseJobResources(jobID)
		}
	})
}

func (s *ResourceService) updateAgentResources(agentID string, heartbeat map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Find resources for this agent
	for _, resource := range s.resources {
		if resource.AgentID == agentID {
			// Update resource metrics from heartbeat
			if metrics, ok := heartbeat["metrics"].(map[string]interface{}); ok {
				// Update capacity information
				resource.LastUpdated = time.Now()
			}
		}
	}
}

func (s *ResourceService) releaseJobResources(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for _, allocation := range s.allocations {
		if allocation.JobID == jobID && allocation.Status == "active" {
			// Release the allocation
			if resource, exists := s.resources[allocation.ResourceID]; exists {
				for k, v := range allocation.AllocatedAmount {
					if vFloat, ok := v.(float64); ok {
						if currentAlloc, exists := resource.AllocatedCapacity[k]; exists {
							if allocFloat, ok := currentAlloc.(float64); ok {
								resource.AllocatedCapacity[k] = allocFloat - vFloat
								
								// Update available capacity
								if total, ok := resource.TotalCapacity[k].(float64); ok {
									resource.AvailableCapacity[k] = total - resource.AllocatedCapacity[k].(float64)
								}
							}
						}
					}
				}
				resource.LastUpdated = time.Now()
			}
			
			allocation.Status = "completed"
			now := time.Now()
			allocation.EndTime = &now
			
			log.Printf("Released resources for completed job %s", jobID)
		}
	}
	
	s.updateResourceMetrics()
}

// Metrics update

func (s *ResourceService) updateResourceMetrics() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Aggregate metrics by resource type
	typeMetrics := make(map[string]map[string]float64)
	
	for _, resource := range s.resources {
		if _, exists := typeMetrics[resource.Type]; !exists {
			typeMetrics[resource.Type] = make(map[string]float64)
		}
		
		// Sum up metrics
		for k, v := range resource.TotalCapacity {
			if vFloat, ok := v.(float64); ok {
				typeMetrics[resource.Type]["total_"+k] += vFloat
			}
		}
		
		for k, v := range resource.AllocatedCapacity {
			if vFloat, ok := v.(float64); ok {
				typeMetrics[resource.Type]["allocated_"+k] += vFloat
			}
		}
	}
	
	// Update Prometheus metrics
	for resourceType, metrics := range typeMetrics {
		for metric, value := range metrics {
			if metric[:6] == "total_" {
				s.totalResources.WithLabelValues(resourceType, metric[6:]).Set(value)
			} else if metric[:10] == "allocated_" {
				s.allocatedResources.WithLabelValues(resourceType, metric[10:]).Set(value)
			}
		}
	}
}

// Event publishing

func (s *ResourceService) publishResourceEvent(event string, resource *Resource) {
	data, _ := json.Marshal(resource)
	s.nats.Publish(event, data)
}

func (s *ResourceService) publishAllocationEvent(event string, allocation *ResourceAllocation) {
	data, _ := json.Marshal(allocation)
	s.nats.Publish(event, data)
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	resourceService, err := NewResourceService()
	if err != nil {
		log.Fatalf("Failed to create resource service: %v", err)
	}
	
	router := mux.NewRouter()
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler())
	
	// Resource endpoints
	router.HandleFunc("/api/v1/resources", resourceService.RegisterResource).Methods("POST")
	router.HandleFunc("/api/v1/resources", resourceService.GetResources).Methods("GET")
	router.HandleFunc("/api/v1/allocations", resourceService.AllocateResource).Methods("POST")
	router.HandleFunc("/api/v1/allocations/{id}/release", resourceService.ReleaseResource).Methods("POST")
	router.HandleFunc("/api/v1/allocations", resourceService.GetAllocations).Methods("GET")
	
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
		port = "8006"
	}
	
	log.Printf("Resource service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
