package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

// MetricPoint represents a single metric data point
type MetricPoint struct {
	Name       string                 `json:"name"`
	Value      float64                `json:"value"`
	Tags       map[string]string      `json:"tags"`
	Fields     map[string]interface{} `json:"fields"`
	Timestamp  time.Time              `json:"timestamp"`
	AgentID    string                 `json:"agent_id"`
	MetricType string                 `json:"metric_type"`
}

// TelemetryService handles metrics collection
type TelemetryService struct {
	metrics    map[string][]*MetricPoint
	nats       *nats.Conn
	metricsReceived prometheus.Counter
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService() (*TelemetryService, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	
	s := &TelemetryService{
		metrics: make(map[string][]*MetricPoint),
		nats:    nc,
		metricsReceived: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "telemetry_metrics_received_total",
				Help: "Total number of metrics received",
			},
		),
	}
	
	prometheus.MustRegister(s.metricsReceived)
	
	return s, nil
}

// IngestMetrics handles metric ingestion
func (s *TelemetryService) IngestMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics []MetricPoint
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Store metrics
	for _, metric := range metrics {
		if s.metrics[metric.Name] == nil {
			s.metrics[metric.Name] = make([]*MetricPoint, 0)
		}
		s.metrics[metric.Name] = append(s.metrics[metric.Name], &metric)
		
		// Keep only last 1000 points per metric
		if len(s.metrics[metric.Name]) > 1000 {
			s.metrics[metric.Name] = s.metrics[metric.Name][100:]
		}
	}
	
	s.metricsReceived.Add(float64(len(metrics)))
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	telemetryService, err := NewTelemetryService()
	if err != nil {
		log.Fatalf("Failed to create telemetry service: %v", err)
	}
	
	router := mux.NewRouter()
	
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/api/v1/metrics", telemetryService.IngestMetrics).Methods("POST")
	
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://computehive.io"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	
	handler := c.Handler(router)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}
	
	log.Printf("Telemetry service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
