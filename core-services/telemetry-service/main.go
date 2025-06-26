package main

import (
	"context"
<<<<<<< HEAD
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
=======
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/shopspring/decimal"
)

// MetricPoint represents a single metric data point
type MetricPoint struct {
<<<<<<< HEAD
	Name       string                 `json:"name"`
	Value      float64                `json:"value"`
	Tags       map[string]string      `json:"tags"`
	Fields     map[string]interface{} `json:"fields"`
	Timestamp  time.Time              `json:"timestamp"`
	AgentID    string                 `json:"agent_id"`
	MetricType string                 `json:"metric_type"` // gauge, counter, histogram
}

// LogEntry represents a log entry
type LogEntry struct {
	ID        string                 `json:"id"`
	Level     string                 `json:"level"` // debug, info, warn, error, fatal
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	AgentID   string                 `json:"agent_id,omitempty"`
	JobID     string                 `json:"job_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}

// Alert represents an alert
type Alert struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Severity    string    `json:"severity"` // info, warning, critical
	Status      string    `json:"status"`   // active, resolved, silenced
	Message     string    `json:"message"`
	Query       string    `json:"query"`
	Threshold   float64   `json:"threshold"`
	CurrentValue float64  `json:"current_value"`
	TriggeredAt time.Time `json:"triggered_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	Tags        map[string]string `json:"tags"`
}

// Dashboard represents a monitoring dashboard
type Dashboard struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Panels      []DashboardPanel `json:"panels"`
	CreatedBy   string         `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// DashboardPanel represents a panel in a dashboard
type DashboardPanel struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title"`
	Type     string                 `json:"type"` // graph, gauge, table, heatmap
	Query    string                 `json:"query"`
	Position Position               `json:"position"`
	Options  map[string]interface{} `json:"options"`
}

// Position represents panel position
type Position struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// TelemetryService handles metrics, logs, and monitoring
type TelemetryService struct {
	metrics       map[string][]*MetricPoint
	logs          []*LogEntry
	alerts        map[string]*Alert
	dashboards    map[string]*Dashboard
	mu            sync.RWMutex
	nats          *nats.Conn
	influxClient  influxdb2.Client
	writeAPI      api.WriteAPIBlocking
	queryAPI      api.QueryAPI
	wsConnections map[string]*websocket.Conn
	wsMutex       sync.RWMutex
	
	// Internal metrics
	metricsReceived    prometheus.Counter
	logsReceived       prometheus.Counter
	alertsTriggered    prometheus.Counter
	queryExecutions    *prometheus.CounterVec
	queryDuration      *prometheus.HistogramVec
	activeConnections  prometheus.Gauge
=======
	Name        string                 `json:"name"`
	Value       float64                `json:"value"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
	Timestamp   time.Time              `json:"timestamp"`
	AgentID     string                 `json:"agent_id"`
	MetricType  string                 `json:"metric_type"` // gauge, counter, histogram
	Unit        string                 `json:"unit"`
	Description string                 `json:"description,omitempty"`
}

// Alert represents a monitoring alert
type Alert struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Condition     string                 `json:"condition"`
	Threshold     float64                `json:"threshold"`
	MetricName    string                 `json:"metric_name"`
	Tags          map[string]string      `json:"tags"`
	Severity      string                 `json:"severity"` // critical, warning, info
	State         string                 `json:"state"`    // firing, resolved
	LastTriggered *time.Time             `json:"last_triggered,omitempty"`
	NotifyWebhook string                 `json:"notify_webhook,omitempty"`
	NotifyEmail   []string               `json:"notify_email,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// AggregatedMetric represents aggregated metric data
type AggregatedMetric struct {
	Name       string            `json:"name"`
	AgentID    string            `json:"agent_id,omitempty"`
	Tags       map[string]string `json:"tags"`
	Period     string            `json:"period"` // 1m, 5m, 1h, 1d
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Count      int64             `json:"count"`
	Sum        float64           `json:"sum"`
	Min        float64           `json:"min"`
	Max        float64           `json:"max"`
	Avg        float64           `json:"avg"`
	P50        float64           `json:"p50"`
	P95        float64           `json:"p95"`
	P99        float64           `json:"p99"`
	StdDev     float64           `json:"std_dev"`
}

// TelemetryService handles metrics collection, storage, and querying
type TelemetryService struct {
	db                *sql.DB
	nats              *nats.Conn
	alerts            map[string]*Alert
	alertMu           sync.RWMutex
	wsClients         map[string]*websocket.Conn
	wsClientsMu       sync.RWMutex
	metricBuffer      []*MetricPoint
	bufferMu          sync.Mutex
	
	// Metrics
	metricsReceived   *prometheus.CounterVec
	metricsStored     *prometheus.CounterVec
	alertsTriggered   *prometheus.CounterVec
	queryDuration     *prometheus.HistogramVec
	wsConnections     prometheus.Gauge
	bufferSize        prometheus.Gauge
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService() (*TelemetryService, error) {
	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	
<<<<<<< HEAD
	// Connect to InfluxDB
	influxURL := os.Getenv("INFLUXDB_URL")
	if influxURL == "" {
		influxURL = "http://localhost:8086"
	}
	
	influxToken := os.Getenv("INFLUXDB_TOKEN")
	if influxToken == "" {
		influxToken = "dev-token"
	}
	
	influxOrg := os.Getenv("INFLUXDB_ORG")
	if influxOrg == "" {
		influxOrg = "computehive"
	}
	
	influxBucket := os.Getenv("INFLUXDB_BUCKET")
	if influxBucket == "" {
		influxBucket = "metrics"
	}
	
	influxClient := influxdb2.NewClient(influxURL, influxToken)
	writeAPI := influxClient.WriteAPIBlocking(influxOrg, influxBucket)
	queryAPI := influxClient.QueryAPI(influxOrg)
	
	s := &TelemetryService{
		metrics:       make(map[string][]*MetricPoint),
		logs:          make([]*LogEntry, 0),
		alerts:        make(map[string]*Alert),
		dashboards:    make(map[string]*Dashboard),
		nats:          nc,
		influxClient:  influxClient,
		writeAPI:      writeAPI,
		queryAPI:      queryAPI,
		wsConnections: make(map[string]*websocket.Conn),
		
		// Initialize internal metrics
		metricsReceived: prometheus.NewCounter(
=======
	// Connect to TimescaleDB
	dbURL := os.Getenv("TIMESCALE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/telemetry?sslmode=disable"
	}
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TimescaleDB: %w", err)
	}
	
	// Initialize schema
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}
	
	s := &TelemetryService{
		db:           db,
		nats:         nc,
		alerts:       make(map[string]*Alert),
		wsClients:    make(map[string]*websocket.Conn),
		metricBuffer: make([]*MetricPoint, 0, 10000),
		
		// Initialize metrics
		metricsReceived: prometheus.NewCounterVec(
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
			prometheus.CounterOpts{
				Name: "telemetry_metrics_received_total",
				Help: "Total number of metrics received",
			},
<<<<<<< HEAD
		),
		logsReceived: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "telemetry_logs_received_total",
				Help: "Total number of logs received",
			},
		),
		alertsTriggered: prometheus.NewCounter(
=======
			[]string{"metric_name", "agent_id"},
		),
		metricsStored: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "telemetry_metrics_stored_total",
				Help: "Total number of metrics stored to database",
			},
			[]string{"status"},
		),
		alertsTriggered: prometheus.NewCounterVec(
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
			prometheus.CounterOpts{
				Name: "telemetry_alerts_triggered_total",
				Help: "Total number of alerts triggered",
			},
<<<<<<< HEAD
		),
		queryExecutions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "telemetry_query_executions_total",
				Help: "Total number of queries executed",
			},
			[]string{"type"},
=======
			[]string{"alert_name", "severity"},
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
		),
		queryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "telemetry_query_duration_seconds",
<<<<<<< HEAD
				Help:    "Query execution duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"type"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "telemetry_active_connections",
				Help: "Number of active WebSocket connections",
=======
				Help:    "Query execution time",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"query_type"},
		),
		wsConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "telemetry_websocket_connections",
				Help: "Current number of WebSocket connections",
			},
		),
		bufferSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "telemetry_buffer_size",
				Help: "Current size of the metrics buffer",
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
			},
		),
	}
	
<<<<<<< HEAD
	// Register internal metrics
	prometheus.MustRegister(
		s.metricsReceived,
		s.logsReceived,
		s.alertsTriggered,
		s.queryExecutions,
		s.queryDuration,
		s.activeConnections,
=======
	// Register metrics
	prometheus.MustRegister(
		s.metricsReceived,
		s.metricsStored,
		s.alertsTriggered,
		s.queryDuration,
		s.wsConnections,
		s.bufferSize,
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	)
	
	// Subscribe to events
	s.subscribeToEvents()
	
	// Start background workers
<<<<<<< HEAD
	go s.alertEvaluator()
	go s.metricAggregator()
	go s.logRotator()
	
	// Create default dashboards
	s.createDefaultDashboards()
=======
	go s.metricFlusher()
	go s.alertEvaluator()
	go s.aggregator()
	go s.retentionManager()
	
	// Load alerts from database
	s.loadAlerts()
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	
	return s, nil
}

// HTTP Handlers

<<<<<<< HEAD
// IngestMetrics handles metric ingestion
=======
// IngestMetrics handles metric ingestion from agents
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
func (s *TelemetryService) IngestMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics []MetricPoint
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
<<<<<<< HEAD
	// Store metrics in memory (for quick access)
	s.mu.Lock()
	for _, metric := range metrics {
		if s.metrics[metric.Name] == nil {
			s.metrics[metric.Name] = make([]*MetricPoint, 0)
		}
		s.metrics[metric.Name] = append(s.metrics[metric.Name], &metric)
		
		// Keep only last 1000 points per metric in memory
		if len(s.metrics[metric.Name]) > 1000 {
			s.metrics[metric.Name] = s.metrics[metric.Name][100:]
		}
	}
	s.mu.Unlock()
	
	// Write to InfluxDB
	for _, metric := range metrics {
		point := influxdb2.NewPointWithMeasurement(metric.Name)
		
		// Add tags
		for k, v := range metric.Tags {
			point.AddTag(k, v)
		}
		point.AddTag("agent_id", metric.AgentID)
		
		// Add fields
		for k, v := range metric.Fields {
			point.AddField(k, v)
		}
		point.AddField("value", metric.Value)
		
		point.SetTime(metric.Timestamp)
		
		// Write point
		if err := s.writeAPI.WritePoint(context.Background(), point); err != nil {
			log.Printf("Failed to write metric to InfluxDB: %v", err)
		}
	}
	
	s.metricsReceived.Add(float64(len(metrics)))
	
	// Broadcast metrics to WebSocket connections
	s.broadcastMetrics(metrics)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "count": fmt.Sprintf("%d", len(metrics))})
}

// IngestLogs handles log ingestion
func (s *TelemetryService) IngestLogs(w http.ResponseWriter, r *http.Request) {
	var logs []LogEntry
	if err := json.NewDecoder(r.Body).Decode(&logs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Store logs
	s.mu.Lock()
	s.logs = append(s.logs, logs...)
	
	// Keep only last 10000 logs in memory
	if len(s.logs) > 10000 {
		s.logs = s.logs[1000:]
	}
	s.mu.Unlock()
	
	// Write to InfluxDB (logs bucket)
	for _, log := range logs {
		point := influxdb2.NewPointWithMeasurement("logs")
		
		point.AddTag("level", log.Level)
		point.AddTag("source", log.Source)
		if log.AgentID != "" {
			point.AddTag("agent_id", log.AgentID)
		}
		if log.JobID != "" {
			point.AddTag("job_id", log.JobID)
		}
		if log.UserID != "" {
			point.AddTag("user_id", log.UserID)
		}
		
		// Add custom tags
		for k, v := range log.Tags {
			point.AddTag(k, v)
		}
		
		// Add fields
		point.AddField("message", log.Message)
		for k, v := range log.Fields {
			point.AddField(k, v)
		}
		
		point.SetTime(log.Timestamp)
		
		// Write point
		if err := s.writeAPI.WritePoint(context.Background(), point); err != nil {
			log.Printf("Failed to write log to InfluxDB: %v", err)
		}
	}
	
	s.logsReceived.Add(float64(len(logs)))
	
	// Check for error logs that might trigger alerts
	for _, log := range logs {
		if log.Level == "error" || log.Level == "fatal" {
			s.checkLogAlert(log)
		}
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "count": fmt.Sprintf("%d", len(logs))})
}

// QueryMetrics queries metrics using PromQL-like syntax
func (s *TelemetryService) QueryMetrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter required", http.StatusBadRequest)
		return
	}
	
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	step := r.URL.Query().Get("step")
	
	if start == "" {
		start = "-1h"
	}
	
	timer := prometheus.NewTimer(s.queryDuration.WithLabelValues("metrics"))
	defer timer.ObserveDuration()
	
	// Convert PromQL-like query to InfluxDB query
	influxQuery := s.convertToInfluxQuery(query, start, end, step)
	
	// Execute query
	result, err := s.queryAPI.Query(context.Background(), influxQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Format results
	var data []map[string]interface{}
	for result.Next() {
		values := result.Record().Values()
		data = append(data, values)
	}
	
	if result.Err() != nil {
		http.Error(w, fmt.Sprintf("Query error: %v", result.Err()), http.StatusInternalServerError)
		return
	}
	
	s.queryExecutions.WithLabelValues("metrics").Inc()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

// QueryLogs queries logs
func (s *TelemetryService) QueryLogs(w http.ResponseWriter, r *http.Request) {
	var filter struct {
		Level    string            `json:"level"`
		Source   string            `json:"source"`
		AgentID  string            `json:"agent_id"`
		JobID    string            `json:"job_id"`
		UserID   string            `json:"user_id"`
		Search   string            `json:"search"`
		Start    string            `json:"start"`
		End      string            `json:"end"`
		Limit    int               `json:"limit"`
		Tags     map[string]string `json:"tags"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		// Try query parameters
		filter.Level = r.URL.Query().Get("level")
		filter.Source = r.URL.Query().Get("source")
		filter.Search = r.URL.Query().Get("search")
		filter.Limit = 100
	}
	
	if filter.Limit == 0 {
		filter.Limit = 100
	}
	
	timer := prometheus.NewTimer(s.queryDuration.WithLabelValues("logs"))
	defer timer.ObserveDuration()
	
	// Build query
	query := `from(bucket: "metrics")
		|> range(start: -24h)
		|> filter(fn: (r) => r["_measurement"] == "logs")`
	
	if filter.Level != "" {
		query += fmt.Sprintf(` |> filter(fn: (r) => r["level"] == "%s")`, filter.Level)
	}
	if filter.Source != "" {
		query += fmt.Sprintf(` |> filter(fn: (r) => r["source"] == "%s")`, filter.Source)
	}
	if filter.AgentID != "" {
		query += fmt.Sprintf(` |> filter(fn: (r) => r["agent_id"] == "%s")`, filter.AgentID)
	}
	if filter.JobID != "" {
		query += fmt.Sprintf(` |> filter(fn: (r) => r["job_id"] == "%s")`, filter.JobID)
	}
	
	query += fmt.Sprintf(` |> limit(n: %d) |> sort(columns: ["_time"], desc: true)`, filter.Limit)
	
	// Execute query
	result, err := s.queryAPI.Query(context.Background(), query)
=======
	// Buffer metrics for batch insertion
	s.bufferMu.Lock()
	s.metricBuffer = append(s.metricBuffer, metrics...)
	bufferLen := len(s.metricBuffer)
	s.bufferMu.Unlock()
	
	// Update metrics
	for _, metric := range metrics {
		s.metricsReceived.WithLabelValues(metric.Name, metric.AgentID).Inc()
	}
	s.bufferSize.Set(float64(bufferLen))
	
	// Stream to WebSocket clients
	go s.streamMetrics(metrics)
	
	// Check if buffer should be flushed
	if bufferLen > 5000 {
		go s.flushBuffer()
	}
	
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "accepted",
		"count":  len(metrics),
	})
}

// QueryMetrics queries stored metrics
func (s *TelemetryService) QueryMetrics(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(s.queryDuration.WithLabelValues("metrics"))
	defer timer.ObserveDuration()
	
	// Parse query parameters
	metricName := r.URL.Query().Get("metric")
	agentID := r.URL.Query().Get("agent_id")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	tagsStr := r.URL.Query().Get("tags")
	aggregation := r.URL.Query().Get("aggregation")
	interval := r.URL.Query().Get("interval")
	
	if metricName == "" {
		http.Error(w, "metric parameter is required", http.StatusBadRequest)
		return
	}
	
	// Parse time range
	end := time.Now()
	start := end.Add(-1 * time.Hour) // Default to last hour
	
	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	
	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}
	
	// Parse tags
	tags := make(map[string]string)
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
			http.Error(w, "Invalid tags format", http.StatusBadRequest)
			return
		}
	}
	
	// Query metrics
	var results interface{}
	var err error
	
	if aggregation != "" {
		results, err = s.queryAggregatedMetrics(metricName, agentID, tags, start, end, aggregation, interval)
	} else {
		results, err = s.queryRawMetrics(metricName, agentID, tags, start, end)
	}
	
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	
<<<<<<< HEAD
	// Format results
	var logs []LogEntry
	for result.Next() {
		record := result.Record()
		log := LogEntry{
			Level:     record.ValueByKey("level").(string),
			Message:   record.ValueByKey("message").(string),
			Timestamp: record.Time(),
			Source:    record.ValueByKey("source").(string),
		}
		
		// Optional fields
		if v, ok := record.ValueByKey("agent_id").(string); ok {
			log.AgentID = v
		}
		if v, ok := record.ValueByKey("job_id").(string); ok {
			log.JobID = v
		}
		if v, ok := record.ValueByKey("user_id").(string); ok {
			log.UserID = v
		}
		
		logs = append(logs, log)
	}
	
	s.queryExecutions.WithLabelValues("logs").Inc()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// GetAlerts returns active alerts
func (s *TelemetryService) GetAlerts(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	alerts := make([]*Alert, 0, len(s.alerts))
	for _, alert := range s.alerts {
		if alert.Status == "active" || r.URL.Query().Get("all") == "true" {
			alerts = append(alerts, alert)
		}
	}
	s.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

=======
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GetAgentMetrics returns real-time metrics for a specific agent
func (s *TelemetryService) GetAgentMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	
	timer := prometheus.NewTimer(s.queryDuration.WithLabelValues("agent_metrics"))
	defer timer.ObserveDuration()
	
	// Query latest metrics for the agent
	query := `
		SELECT DISTINCT ON (name) 
			name, value, tags, timestamp, unit, metric_type
		FROM metrics
		WHERE agent_id = $1 
			AND timestamp > NOW() - INTERVAL '5 minutes'
		ORDER BY name, timestamp DESC
	`
	
	rows, err := s.db.Query(query, agentID)
	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	metrics := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name, unit, metricType string
		var value float64
		var tags map[string]string
		var timestamp time.Time
		
		if err := rows.Scan(&name, &value, &tags, &timestamp, &unit, &metricType); err != nil {
			continue
		}
		
		metrics = append(metrics, map[string]interface{}{
			"name":        name,
			"value":       value,
			"tags":        tags,
			"timestamp":   timestamp,
			"unit":        unit,
			"metric_type": metricType,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// Alert Management

>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
// CreateAlert creates a new alert rule
func (s *TelemetryService) CreateAlert(w http.ResponseWriter, r *http.Request) {
	var alert Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	alert.ID = generateID()
<<<<<<< HEAD
	alert.Status = "active"
	
	s.mu.Lock()
	s.alerts[alert.ID] = &alert
	s.mu.Unlock()
=======
	alert.State = "inactive"
	
	// Validate alert
	if alert.Name == "" || alert.MetricName == "" || alert.Condition == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	
	// Store alert
	s.alertMu.Lock()
	s.alerts[alert.ID] = &alert
	s.alertMu.Unlock()
	
	// Save to database
	if err := s.saveAlert(&alert); err != nil {
		http.Error(w, "Failed to save alert", http.StatusInternalServerError)
		return
	}
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

<<<<<<< HEAD
// GetDashboards returns all dashboards
func (s *TelemetryService) GetDashboards(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	dashboards := make([]*Dashboard, 0, len(s.dashboards))
	for _, dashboard := range s.dashboards {
		dashboards = append(dashboards, dashboard)
	}
	s.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboards)
}

// CreateDashboard creates a new dashboard
func (s *TelemetryService) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	var dashboard Dashboard
	if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	dashboard.ID = generateID()
	dashboard.CreatedAt = time.Now()
	dashboard.UpdatedAt = time.Now()
	
	s.mu.Lock()
	s.dashboards[dashboard.ID] = &dashboard
	s.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// WebSocket handler for real-time metrics
func (s *TelemetryService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in dev
		},
	}
	
=======
// GetAlerts returns all alerts
func (s *TelemetryService) GetAlerts(w http.ResponseWriter, r *http.Request) {
	s.alertMu.RLock()
	defer s.alertMu.RUnlock()
	
	alerts := make([]*Alert, 0, len(s.alerts))
	for _, alert := range s.alerts {
		alerts = append(alerts, alert)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// WebSocket Handler

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, implement proper origin checking
	},
}

// StreamMetricsWS handles WebSocket connections for real-time metrics
func (s *TelemetryService) StreamMetricsWS(w http.ResponseWriter, r *http.Request) {
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
<<<<<<< HEAD
	defer conn.Close()
	
	connID := generateID()
	s.wsMutex.Lock()
	s.wsConnections[connID] = conn
	s.activeConnections.Inc()
	s.wsMutex.Unlock()
	
	defer func() {
		s.wsMutex.Lock()
		delete(s.wsConnections, connID)
		s.activeConnections.Dec()
		s.wsMutex.Unlock()
	}()
	
	// Handle incoming messages
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Handle subscription requests
		if msgType, ok := msg["type"].(string); ok && msgType == "subscribe" {
			// Handle metric subscriptions
			log.Printf("WebSocket subscription: %v", msg)
		}
=======
	
	clientID := generateID()
	
	s.wsClientsMu.Lock()
	s.wsClients[clientID] = conn
	s.wsConnections.Set(float64(len(s.wsClients)))
	s.wsClientsMu.Unlock()
	
	defer func() {
		conn.Close()
		s.wsClientsMu.Lock()
		delete(s.wsClients, clientID)
		s.wsConnections.Set(float64(len(s.wsClients)))
		s.wsClientsMu.Unlock()
	}()
	
	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	}
}

// Background Workers

<<<<<<< HEAD
=======
func (s *TelemetryService) metricFlusher() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		s.flushBuffer()
	}
}

func (s *TelemetryService) flushBuffer() {
	s.bufferMu.Lock()
	if len(s.metricBuffer) == 0 {
		s.bufferMu.Unlock()
		return
	}
	
	metrics := s.metricBuffer
	s.metricBuffer = make([]*MetricPoint, 0, 10000)
	s.bufferMu.Unlock()
	
	// Batch insert metrics
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		s.metricsStored.WithLabelValues("error").Add(float64(len(metrics)))
		return
	}
	
	stmt, err := tx.Prepare(`
		INSERT INTO metrics (name, value, tags, fields, timestamp, agent_id, metric_type, unit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to prepare statement: %v", err)
		s.metricsStored.WithLabelValues("error").Add(float64(len(metrics)))
		return
	}
	defer stmt.Close()
	
	for _, metric := range metrics {
		tagsJSON, _ := json.Marshal(metric.Tags)
		fieldsJSON, _ := json.Marshal(metric.Fields)
		
		_, err := stmt.Exec(
			metric.Name,
			metric.Value,
			tagsJSON,
			fieldsJSON,
			metric.Timestamp,
			metric.AgentID,
			metric.MetricType,
			metric.Unit,
		)
		
		if err != nil {
			log.Printf("Failed to insert metric: %v", err)
			s.metricsStored.WithLabelValues("error").Inc()
		}
	}
	
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		s.metricsStored.WithLabelValues("error").Add(float64(len(metrics)))
	} else {
		s.metricsStored.WithLabelValues("success").Add(float64(len(metrics)))
	}
	
	s.bufferSize.Set(0)
}

>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
func (s *TelemetryService) alertEvaluator() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
<<<<<<< HEAD
		s.mu.RLock()
		alerts := make([]*Alert, 0, len(s.alerts))
		for _, alert := range s.alerts {
			if alert.Status == "active" {
				alerts = append(alerts, alert)
			}
		}
		s.mu.RUnlock()
		
		for _, alert := range alerts {
			// Execute alert query
			result, err := s.queryAPI.Query(context.Background(), alert.Query)
			if err != nil {
				log.Printf("Alert query failed: %v", err)
				continue
			}
			
			// Check threshold
			var currentValue float64
			if result.Next() {
				if v, ok := result.Record().Value().(float64); ok {
					currentValue = v
				}
			}
			
			// Update alert state
			s.mu.Lock()
			alert.CurrentValue = currentValue
			
			if currentValue > alert.Threshold && alert.Status == "active" {
				alert.TriggeredAt = time.Now()
				s.alertsTriggered.Inc()
				
				// Send alert notification
				s.sendAlertNotification(alert)
			} else if currentValue <= alert.Threshold && alert.Status == "triggered" {
				now := time.Now()
				alert.ResolvedAt = &now
				alert.Status = "resolved"
			}
			s.mu.Unlock()
=======
		s.evaluateAlerts()
	}
}

func (s *TelemetryService) evaluateAlerts() {
	s.alertMu.RLock()
	alerts := make([]*Alert, 0, len(s.alerts))
	for _, alert := range s.alerts {
		alerts = append(alerts, alert)
	}
	s.alertMu.RUnlock()
	
	for _, alert := range alerts {
		// Query recent metrics
		var value float64
		query := `
			SELECT AVG(value) 
			FROM metrics 
			WHERE name = $1 
				AND timestamp > NOW() - INTERVAL '5 minutes'
		`
		
		err := s.db.QueryRow(query, alert.MetricName).Scan(&value)
		if err != nil {
			continue
		}
		
		// Evaluate condition
		triggered := false
		switch alert.Condition {
		case "gt", ">":
			triggered = value > alert.Threshold
		case "lt", "<":
			triggered = value < alert.Threshold
		case "gte", ">=":
			triggered = value >= alert.Threshold
		case "lte", "<=":
			triggered = value <= alert.Threshold
		case "eq", "==":
			triggered = math.Abs(value-alert.Threshold) < 0.001
		}
		
		// Update alert state
		if triggered && alert.State != "firing" {
			s.triggerAlert(alert, value)
		} else if !triggered && alert.State == "firing" {
			s.resolveAlert(alert)
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
		}
	}
}

<<<<<<< HEAD
func (s *TelemetryService) metricAggregator() {
=======
func (s *TelemetryService) triggerAlert(alert *Alert, value float64) {
	now := time.Now()
	alert.State = "firing"
	alert.LastTriggered = &now
	
	s.alertsTriggered.WithLabelValues(alert.Name, alert.Severity).Inc()
	
	// Send notifications
	notification := map[string]interface{}{
		"alert_id":   alert.ID,
		"alert_name": alert.Name,
		"severity":   alert.Severity,
		"metric":     alert.MetricName,
		"value":      value,
		"threshold":  alert.Threshold,
		"condition":  alert.Condition,
		"timestamp":  now,
		"state":      "firing",
	}
	
	// Publish to NATS
	data, _ := json.Marshal(notification)
	s.nats.Publish("alerts.triggered", data)
	
	// Update in database
	s.updateAlertState(alert)
	
	log.Printf("Alert triggered: %s (value: %f, threshold: %f)", alert.Name, value, alert.Threshold)
}

func (s *TelemetryService) resolveAlert(alert *Alert) {
	alert.State = "resolved"
	
	// Send resolution notification
	notification := map[string]interface{}{
		"alert_id":   alert.ID,
		"alert_name": alert.Name,
		"timestamp":  time.Now(),
		"state":      "resolved",
	}
	
	data, _ := json.Marshal(notification)
	s.nats.Publish("alerts.resolved", data)
	
	// Update in database
	s.updateAlertState(alert)
	
	log.Printf("Alert resolved: %s", alert.Name)
}

func (s *TelemetryService) aggregator() {
	// Run aggregation every minute
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
<<<<<<< HEAD
		// Aggregate metrics for downsampling
		// This would create hourly/daily aggregates
		log.Println("Running metric aggregation...")
	}
}

func (s *TelemetryService) logRotator() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// Clean up old logs from memory
		s.mu.Lock()
		cutoff := time.Now().Add(-24 * time.Hour)
		var newLogs []*LogEntry
		for _, log := range s.logs {
			if log.Timestamp.After(cutoff) {
				newLogs = append(newLogs, log)
			}
		}
		s.logs = newLogs
		s.mu.Unlock()
		
		log.Printf("Log rotation complete, retained %d logs", len(newLogs))
	}
}

// Helper Methods
=======
		s.runAggregations()
	}
}

func (s *TelemetryService) runAggregations() {
	// Aggregate metrics for different time windows
	windows := []struct {
		name     string
		interval string
		retention string
	}{
		{"1m", "1 minute", "1 hour"},
		{"5m", "5 minutes", "6 hours"},
		{"1h", "1 hour", "7 days"},
		{"1d", "1 day", "90 days"},
	}
	
	for _, window := range windows {
		query := fmt.Sprintf(`
			INSERT INTO metrics_aggregated (name, agent_id, tags, period, start_time, end_time,
				count, sum, min, max, avg, p50, p95, p99)
			SELECT 
				name,
				agent_id,
				tags,
				'%s' as period,
				date_trunc('minute', timestamp) as start_time,
				date_trunc('minute', timestamp) + INTERVAL '%s' as end_time,
				COUNT(*) as count,
				SUM(value) as sum,
				MIN(value) as min,
				MAX(value) as max,
				AVG(value) as avg,
				PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY value) as p50,
				PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY value) as p95,
				PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY value) as p99
			FROM metrics
			WHERE timestamp >= NOW() - INTERVAL '%s'
				AND timestamp < date_trunc('minute', NOW())
			GROUP BY name, agent_id, tags, date_trunc('minute', timestamp)
			ON CONFLICT (name, agent_id, tags, period, start_time) DO NOTHING
		`, window.name, window.interval, window.interval)
		
		if _, err := s.db.Exec(query); err != nil {
			log.Printf("Aggregation failed for %s window: %v", window.name, err)
		}
	}
}

func (s *TelemetryService) retentionManager() {
	// Run retention cleanup daily
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		s.cleanupOldData()
	}
}

func (s *TelemetryService) cleanupOldData() {
	// Clean up old raw metrics
	if _, err := s.db.Exec(`
		DELETE FROM metrics 
		WHERE timestamp < NOW() - INTERVAL '7 days'
	`); err != nil {
		log.Printf("Failed to clean up old metrics: %v", err)
	}
	
	// Clean up old aggregated metrics
	retentions := map[string]string{
		"1m": "1 day",
		"5m": "7 days",
		"1h": "30 days",
		"1d": "365 days",
	}
	
	for period, retention := range retentions {
		query := fmt.Sprintf(`
			DELETE FROM metrics_aggregated 
			WHERE period = '%s' AND start_time < NOW() - INTERVAL '%s'
		`, period, retention)
		
		if _, err := s.db.Exec(query); err != nil {
			log.Printf("Failed to clean up %s aggregations: %v", period, err)
		}
	}
}

// Helper functions

func (s *TelemetryService) streamMetrics(metrics []MetricPoint) {
	s.wsClientsMu.RLock()
	defer s.wsClientsMu.RUnlock()
	
	if len(s.wsClients) == 0 {
		return
	}
	
	data, err := json.Marshal(map[string]interface{}{
		"type":    "metrics",
		"metrics": metrics,
	})
	if err != nil {
		return
	}
	
	for clientID, conn := range s.wsClients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Failed to send to WebSocket client %s: %v", clientID, err)
		}
	}
}

func (s *TelemetryService) queryRawMetrics(name, agentID string, tags map[string]string, start, end time.Time) ([]MetricPoint, error) {
	query := `
		SELECT name, value, tags, fields, timestamp, agent_id, metric_type, unit
		FROM metrics
		WHERE name = $1 AND timestamp >= $2 AND timestamp <= $3
	`
	
	args := []interface{}{name, start, end}
	
	if agentID != "" {
		query += " AND agent_id = $4"
		args = append(args, agentID)
	}
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	metrics := make([]MetricPoint, 0)
	for rows.Next() {
		var m MetricPoint
		var tagsJSON, fieldsJSON []byte
		
		err := rows.Scan(&m.Name, &m.Value, &tagsJSON, &fieldsJSON,
			&m.Timestamp, &m.AgentID, &m.MetricType, &m.Unit)
		if err != nil {
			continue
		}
		
		json.Unmarshal(tagsJSON, &m.Tags)
		json.Unmarshal(fieldsJSON, &m.Fields)
		
		// Filter by tags if specified
		if len(tags) > 0 {
			match := true
			for k, v := range tags {
				if m.Tags[k] != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}
		
		metrics = append(metrics, m)
	}
	
	return metrics, nil
}

func (s *TelemetryService) queryAggregatedMetrics(name, agentID string, tags map[string]string,
	start, end time.Time, aggregation, interval string) ([]AggregatedMetric, error) {
	
	// Map interval to period
	periodMap := map[string]string{
		"1m":  "1m",
		"5m":  "5m",
		"15m": "5m",  // Use 5m aggregations
		"1h":  "1h",
		"1d":  "1d",
	}
	
	period, ok := periodMap[interval]
	if !ok {
		period = "5m" // Default
	}
	
	query := `
		SELECT name, agent_id, tags, period, start_time, end_time,
			count, sum, min, max, avg, p50, p95, p99
		FROM metrics_aggregated
		WHERE name = $1 AND period = $2 AND start_time >= $3 AND end_time <= $4
	`
	
	args := []interface{}{name, period, start, end}
	
	if agentID != "" {
		query += " AND agent_id = $5"
		args = append(args, agentID)
	}
	
	query += " ORDER BY start_time"
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	aggregated := make([]AggregatedMetric, 0)
	for rows.Next() {
		var a AggregatedMetric
		var tagsJSON []byte
		
		err := rows.Scan(&a.Name, &a.AgentID, &tagsJSON, &a.Period,
			&a.StartTime, &a.EndTime, &a.Count, &a.Sum,
			&a.Min, &a.Max, &a.Avg, &a.P50, &a.P95, &a.P99)
		if err != nil {
			continue
		}
		
		json.Unmarshal(tagsJSON, &a.Tags)
		
		// Filter by tags if specified
		if len(tags) > 0 {
			match := true
			for k, v := range tags {
				if a.Tags[k] != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}
		
		aggregated = append(aggregated, a)
	}
	
	return aggregated, nil
}
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849

func (s *TelemetryService) subscribeToEvents() {
	// Subscribe to agent metrics
	s.nats.Subscribe("agent.metrics", func(msg *nats.Msg) {
		var metrics []MetricPoint
		if err := json.Unmarshal(msg.Data, &metrics); err != nil {
			return
		}
		
<<<<<<< HEAD
		// Process metrics
		for _, metric := range metrics {
			s.storeMetric(&metric)
		}
	})
	
	// Subscribe to job events for metrics
	s.nats.Subscribe("job.>", func(msg *nats.Msg) {
		// Extract job metrics from events
		var event map[string]interface{}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return
		}
		
		// Create metric from job event
		metric := MetricPoint{
			Name:      "job_events",
			Value:     1,
			Tags:      map[string]string{"event": msg.Subject},
			Timestamp: time.Now(),
		}
		
		if jobID, ok := event["job_id"].(string); ok {
			metric.Tags["job_id"] = jobID
		}
		
		s.storeMetric(&metric)
	})
}

func (s *TelemetryService) storeMetric(metric *MetricPoint) {
	// Store in memory
	s.mu.Lock()
	if s.metrics[metric.Name] == nil {
		s.metrics[metric.Name] = make([]*MetricPoint, 0)
	}
	s.metrics[metric.Name] = append(s.metrics[metric.Name], metric)
	s.mu.Unlock()
	
	// Write to InfluxDB
	point := influxdb2.NewPointWithMeasurement(metric.Name)
	for k, v := range metric.Tags {
		point.AddTag(k, v)
	}
	point.AddField("value", metric.Value)
	point.SetTime(metric.Timestamp)
	
	s.writeAPI.WritePoint(context.Background(), point)
}

func (s *TelemetryService) broadcastMetrics(metrics []MetricPoint) {
	s.wsMutex.RLock()
	defer s.wsMutex.RUnlock()
	
	for _, conn := range s.wsConnections {
		err := conn.WriteJSON(map[string]interface{}{
			"type":    "metrics",
			"metrics": metrics,
		})
		if err != nil {
			log.Printf("Failed to broadcast metrics: %v", err)
		}
	}
}

func (s *TelemetryService) checkLogAlert(log LogEntry) {
	// Check if this log should trigger an alert
	if strings.Contains(strings.ToLower(log.Message), "critical") ||
		strings.Contains(strings.ToLower(log.Message), "fatal") {
		
		alert := &Alert{
			ID:           generateID(),
			Name:         "Critical Log Alert",
			Severity:     "critical",
			Status:       "active",
			Message:      fmt.Sprintf("Critical log from %s: %s", log.Source, log.Message),
			TriggeredAt:  time.Now(),
			Tags:         log.Tags,
		}
		
		s.mu.Lock()
		s.alerts[alert.ID] = alert
		s.mu.Unlock()
		
		s.sendAlertNotification(alert)
	}
}

func (s *TelemetryService) sendAlertNotification(alert *Alert) {
	// Publish alert to NATS
	data, _ := json.Marshal(alert)
	s.nats.Publish("alert.triggered", data)
	
	// In production, this would also:
	// - Send email notifications
	// - Send Slack/Discord notifications
	// - Create PagerDuty incidents
	// - Send webhooks
}

func (s *TelemetryService) convertToInfluxQuery(promQL, start, end, step string) string {
	// Basic PromQL to InfluxDB query conversion
	// This is a simplified version - production would need full parser
	
	query := `from(bucket: "metrics")`
	
	// Time range
	if start != "" {
		query += fmt.Sprintf(` |> range(start: %s`, start)
		if end != "" {
			query += fmt.Sprintf(`, stop: %s`, end)
		}
		query += `)`
	}
	
	// Parse metric name from PromQL
	if strings.Contains(promQL, "{") {
		parts := strings.Split(promQL, "{")
		metricName := strings.TrimSpace(parts[0])
		query += fmt.Sprintf(` |> filter(fn: (r) => r["_measurement"] == "%s")`, metricName)
		
		// Parse labels
		if len(parts) > 1 {
			labelsPart := strings.TrimSuffix(parts[1], "}")
			labels := strings.Split(labelsPart, ",")
			for _, label := range labels {
				kv := strings.Split(label, "=")
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.Trim(strings.TrimSpace(kv[1]), `"`)
					query += fmt.Sprintf(` |> filter(fn: (r) => r["%s"] == "%s")`, key, value)
				}
			}
		}
	} else {
		query += fmt.Sprintf(` |> filter(fn: (r) => r["_measurement"] == "%s")`, promQL)
	}
	
	// Aggregation window
	if step != "" {
		query += fmt.Sprintf(` |> aggregateWindow(every: %s, fn: mean)`, step)
	}
	
	return query
}

func (s *TelemetryService) createDefaultDashboards() {
	// System Overview Dashboard
	systemDashboard := &Dashboard{
		ID:          "system-overview",
		Name:        "System Overview",
		Description: "Overall system health and performance",
		Panels: []DashboardPanel{
			{
				ID:    "cpu-usage",
				Title: "CPU Usage",
				Type:  "graph",
				Query: `from(bucket: "metrics") |> range(start: -1h) |> filter(fn: (r) => r["_measurement"] == "cpu_usage")`,
				Position: Position{X: 0, Y: 0, Width: 6, Height: 4},
			},
			{
				ID:    "memory-usage",
				Title: "Memory Usage",
				Type:  "graph",
				Query: `from(bucket: "metrics") |> range(start: -1h) |> filter(fn: (r) => r["_measurement"] == "memory_usage")`,
				Position: Position{X: 6, Y: 0, Width: 6, Height: 4},
			},
			{
				ID:    "active-agents",
				Title: "Active Agents",
				Type:  "gauge",
				Query: `from(bucket: "metrics") |> range(start: -5m) |> filter(fn: (r) => r["_measurement"] == "agent_status") |> last()`,
				Position: Position{X: 0, Y: 4, Width: 3, Height: 3},
			},
			{
				ID:    "job-rate",
				Title: "Job Processing Rate",
				Type:  "graph",
				Query: `from(bucket: "metrics") |> range(start: -1h) |> filter(fn: (r) => r["_measurement"] == "job_events")`,
				Position: Position{X: 3, Y: 4, Width: 9, Height: 4},
			},
		},
		CreatedBy: "system",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Agent Performance Dashboard
	agentDashboard := &Dashboard{
		ID:          "agent-performance",
		Name:        "Agent Performance",
		Description: "Detailed agent metrics and performance",
		Panels: []DashboardPanel{
			{
				ID:    "agent-cpu-by-id",
				Title: "CPU by Agent",
				Type:  "table",
				Query: `from(bucket: "metrics") |> range(start: -5m) |> filter(fn: (r) => r["_measurement"] == "cpu_usage") |> group(columns: ["agent_id"]) |> mean()`,
				Position: Position{X: 0, Y: 0, Width: 12, Height: 4},
			},
			{
				ID:    "agent-job-success",
				Title: "Job Success Rate",
				Type:  "heatmap",
				Query: `from(bucket: "metrics") |> range(start: -1h) |> filter(fn: (r) => r["_measurement"] == "job_success_rate")`,
				Position: Position{X: 0, Y: 4, Width: 6, Height: 4},
			},
		},
		CreatedBy: "system",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	s.mu.Lock()
	s.dashboards[systemDashboard.ID] = systemDashboard
	s.dashboards[agentDashboard.ID] = agentDashboard
	s.mu.Unlock()
}

=======
		// Add to buffer
		s.bufferMu.Lock()
		s.metricBuffer = append(s.metricBuffer, metrics...)
		s.bufferMu.Unlock()
		
		// Stream to WebSocket clients
		go s.streamMetrics(metrics)
	})
	
	// Subscribe to system events for metrics
	s.nats.Subscribe("job.started", func(msg *nats.Msg) {
		// Create metric for job start
		metric := MetricPoint{
			Name:       "job.events",
			Value:      1,
			Tags:       map[string]string{"event": "started"},
			Timestamp:  time.Now(),
			MetricType: "counter",
		}
		
		s.bufferMu.Lock()
		s.metricBuffer = append(s.metricBuffer, &metric)
		s.bufferMu.Unlock()
	})
}

func (s *TelemetryService) loadAlerts() error {
	rows, err := s.db.Query(`
		SELECT id, name, condition, threshold, metric_name, tags, severity,
			state, last_triggered, notify_webhook, notify_email, metadata
		FROM alerts WHERE active = true
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var alert Alert
		var tagsJSON, emailJSON, metadataJSON []byte
		var lastTriggered sql.NullTime
		
		err := rows.Scan(&alert.ID, &alert.Name, &alert.Condition, &alert.Threshold,
			&alert.MetricName, &tagsJSON, &alert.Severity, &alert.State,
			&lastTriggered, &alert.NotifyWebhook, &emailJSON, &metadataJSON)
		if err != nil {
			continue
		}
		
		if lastTriggered.Valid {
			alert.LastTriggered = &lastTriggered.Time
		}
		
		json.Unmarshal(tagsJSON, &alert.Tags)
		json.Unmarshal(emailJSON, &alert.NotifyEmail)
		json.Unmarshal(metadataJSON, &alert.Metadata)
		
		s.alertMu.Lock()
		s.alerts[alert.ID] = &alert
		s.alertMu.Unlock()
	}
	
	return nil
}

func (s *TelemetryService) saveAlert(alert *Alert) error {
	tagsJSON, _ := json.Marshal(alert.Tags)
	emailJSON, _ := json.Marshal(alert.NotifyEmail)
	metadataJSON, _ := json.Marshal(alert.Metadata)
	
	_, err := s.db.Exec(`
		INSERT INTO alerts (id, name, condition, threshold, metric_name, tags,
			severity, state, notify_webhook, notify_email, metadata, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, true)
		ON CONFLICT (id) DO UPDATE SET
			name = $2, condition = $3, threshold = $4, metric_name = $5,
			tags = $6, severity = $7, notify_webhook = $9,
			notify_email = $10, metadata = $11
	`, alert.ID, alert.Name, alert.Condition, alert.Threshold, alert.MetricName,
		tagsJSON, alert.Severity, alert.State, alert.NotifyWebhook,
		emailJSON, metadataJSON)
	
	return err
}

func (s *TelemetryService) updateAlertState(alert *Alert) error {
	_, err := s.db.Exec(`
		UPDATE alerts 
		SET state = $1, last_triggered = $2
		WHERE id = $3
	`, alert.State, alert.LastTriggered, alert.ID)
	
	return err
}

// Database schema initialization
func initSchema(db *sql.DB) error {
	schema := `
	-- Enable TimescaleDB extension
	CREATE EXTENSION IF NOT EXISTS timescaledb;
	
	-- Metrics table
	CREATE TABLE IF NOT EXISTS metrics (
		time        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		name        TEXT NOT NULL,
		value       DOUBLE PRECISION NOT NULL,
		tags        JSONB,
		fields      JSONB,
		timestamp   TIMESTAMPTZ NOT NULL,
		agent_id    TEXT,
		metric_type TEXT,
		unit        TEXT
	);
	
	-- Create hypertable
	SELECT create_hypertable('metrics', 'time', if_not_exists => TRUE);
	
	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_metrics_name_time ON metrics (name, time DESC);
	CREATE INDEX IF NOT EXISTS idx_metrics_agent_time ON metrics (agent_id, time DESC);
	CREATE INDEX IF NOT EXISTS idx_metrics_tags ON metrics USING GIN (tags);
	
	-- Aggregated metrics table
	CREATE TABLE IF NOT EXISTS metrics_aggregated (
		name       TEXT NOT NULL,
		agent_id   TEXT,
		tags       JSONB,
		period     TEXT NOT NULL,
		start_time TIMESTAMPTZ NOT NULL,
		end_time   TIMESTAMPTZ NOT NULL,
		count      BIGINT,
		sum        DOUBLE PRECISION,
		min        DOUBLE PRECISION,
		max        DOUBLE PRECISION,
		avg        DOUBLE PRECISION,
		p50        DOUBLE PRECISION,
		p95        DOUBLE PRECISION,
		p99        DOUBLE PRECISION,
		std_dev    DOUBLE PRECISION,
		PRIMARY KEY (name, agent_id, tags, period, start_time)
	);
	
	-- Alerts table
	CREATE TABLE IF NOT EXISTS alerts (
		id             TEXT PRIMARY KEY,
		name           TEXT NOT NULL,
		condition      TEXT NOT NULL,
		threshold      DOUBLE PRECISION NOT NULL,
		metric_name    TEXT NOT NULL,
		tags           JSONB,
		severity       TEXT NOT NULL,
		state          TEXT NOT NULL DEFAULT 'inactive',
		last_triggered TIMESTAMPTZ,
		notify_webhook TEXT,
		notify_email   JSONB,
		metadata       JSONB,
		active         BOOLEAN DEFAULT true,
		created_at     TIMESTAMPTZ DEFAULT NOW()
	);
	
	-- Continuous aggregates for real-time analytics
	CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_1min
	WITH (timescaledb.continuous) AS
	SELECT 
		time_bucket('1 minute', time) AS bucket,
		name,
		agent_id,
		tags,
		COUNT(*) as count,
		AVG(value) as avg,
		MAX(value) as max,
		MIN(value) as min
	FROM metrics
	GROUP BY bucket, name, agent_id, tags
	WITH NO DATA;
	
	-- Refresh policy
	SELECT add_continuous_aggregate_policy('metrics_1min',
		start_offset => INTERVAL '10 minutes',
		end_offset => INTERVAL '1 minute',
		schedule_interval => INTERVAL '1 minute',
		if_not_exists => TRUE);
	`
	
	_, err := db.Exec(schema)
	return err
}

// JWT Claims
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
		// For metrics ingestion, allow agent authentication
		if r.URL.Path == "/api/v1/metrics" && r.Method == "POST" {
			// Check for agent API key
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" {
				// Validate API key (simplified for now)
				next(w, r)
				return
			}
		}
		
		// Standard JWT authentication for other endpoints
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}
		
		tokenString = tokenString[7:] // Remove "Bearer "
		
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		claims := token.Claims.(*Claims)
		ctx := context.WithValue(r.Context(), "claims", claims)
		next(w, r.WithContext(ctx))
	}
}

>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
<<<<<<< HEAD
	// Create telemetry service
=======
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	telemetryService, err := NewTelemetryService()
	if err != nil {
		log.Fatalf("Failed to create telemetry service: %v", err)
	}
<<<<<<< HEAD
	defer telemetryService.influxClient.Close()
	
	// Setup routes
=======
	
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	router := mux.NewRouter()
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
<<<<<<< HEAD
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	// Metrics endpoint for Prometheus
	router.Handle("/metrics", promhttp.Handler())
	
	// Telemetry API endpoints
	router.HandleFunc("/api/v1/metrics", telemetryService.IngestMetrics).Methods("POST")
	router.HandleFunc("/api/v1/metrics/query", telemetryService.QueryMetrics).Methods("GET")
	router.HandleFunc("/api/v1/logs", telemetryService.IngestLogs).Methods("POST")
	router.HandleFunc("/api/v1/logs/query", telemetryService.QueryLogs).Methods("GET", "POST")
	router.HandleFunc("/api/v1/alerts", telemetryService.GetAlerts).Methods("GET")
	router.HandleFunc("/api/v1/alerts", telemetryService.CreateAlert).Methods("POST")
	router.HandleFunc("/api/v1/dashboards", telemetryService.GetDashboards).Methods("GET")
	router.HandleFunc("/api/v1/dashboards", telemetryService.CreateDashboard).Methods("POST")
	
	// WebSocket endpoint
	router.HandleFunc("/ws", telemetryService.HandleWebSocket)
	
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://computehive.io"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
=======
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")
	
	// Prometheus metrics
	router.Handle("/metrics", promhttp.Handler())
	
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Metrics endpoints
	api.HandleFunc("/metrics", telemetryService.IngestMetrics).Methods("POST")
	api.HandleFunc("/metrics/query", authMiddleware(telemetryService.QueryMetrics)).Methods("GET")
	api.HandleFunc("/agents/{agent_id}/metrics", authMiddleware(telemetryService.GetAgentMetrics)).Methods("GET")
	
	// Alert endpoints
	api.HandleFunc("/alerts", authMiddleware(telemetryService.CreateAlert)).Methods("POST")
	api.HandleFunc("/alerts", authMiddleware(telemetryService.GetAlerts)).Methods("GET")
	
	// WebSocket endpoint
	api.HandleFunc("/stream", telemetryService.StreamMetricsWS)
	
	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://computehive.io"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-API-Key"},
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
		AllowCredentials: true,
	})
	
	handler := c.Handler(router)
	
<<<<<<< HEAD
	// Start server
=======
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}
	
	log.Printf("Telemetry service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
<<<<<<< HEAD
} 
=======
}

>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
