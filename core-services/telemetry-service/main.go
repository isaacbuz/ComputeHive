package main

import (
	"context"
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
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/shopspring/decimal"
)

// MetricPoint represents a single metric data point
type MetricPoint struct {
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
			prometheus.CounterOpts{
				Name: "telemetry_metrics_received_total",
				Help: "Total number of metrics received",
			},
		),
		logsReceived: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "telemetry_logs_received_total",
				Help: "Total number of logs received",
			},
		),
		alertsTriggered: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "telemetry_alerts_triggered_total",
				Help: "Total number of alerts triggered",
			},
		),
		queryExecutions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "telemetry_query_executions_total",
				Help: "Total number of queries executed",
			},
			[]string{"type"},
		),
		queryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "telemetry_query_duration_seconds",
				Help:    "Query execution duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"type"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "telemetry_active_connections",
				Help: "Number of active WebSocket connections",
			},
		),
	}
	
	// Register internal metrics
	prometheus.MustRegister(
		s.metricsReceived,
		s.logsReceived,
		s.alertsTriggered,
		s.queryExecutions,
		s.queryDuration,
		s.activeConnections,
	)
	
	// Subscribe to events
	s.subscribeToEvents()
	
	// Start background workers
	go s.alertEvaluator()
	go s.metricAggregator()
	go s.logRotator()
	
	// Create default dashboards
	s.createDefaultDashboards()
	
	return s, nil
}

// HTTP Handlers

// IngestMetrics handles metric ingestion
func (s *TelemetryService) IngestMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics []MetricPoint
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
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
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	
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

// CreateAlert creates a new alert rule
func (s *TelemetryService) CreateAlert(w http.ResponseWriter, r *http.Request) {
	var alert Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	alert.ID = generateID()
	alert.Status = "active"
	
	s.mu.Lock()
	s.alerts[alert.ID] = &alert
	s.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

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
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
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
	}
}

// Background Workers

func (s *TelemetryService) alertEvaluator() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
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
		}
	}
}

func (s *TelemetryService) metricAggregator() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
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

func (s *TelemetryService) subscribeToEvents() {
	// Subscribe to agent metrics
	s.nats.Subscribe("agent.metrics", func(msg *nats.Msg) {
		var metrics []MetricPoint
		if err := json.Unmarshal(msg.Data, &metrics); err != nil {
			return
		}
		
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

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	// Create telemetry service
	telemetryService, err := NewTelemetryService()
	if err != nil {
		log.Fatalf("Failed to create telemetry service: %v", err)
	}
	defer telemetryService.influxClient.Close()
	
	// Setup routes
	router := mux.NewRouter()
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
		AllowCredentials: true,
	})
	
	handler := c.Handler(router)
	
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}
	
	log.Printf("Telemetry service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 