package main

import (
	"context"
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
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/shopspring/decimal"
)

// MetricPoint represents a single metric data point
type MetricPoint struct {
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
			prometheus.CounterOpts{
				Name: "telemetry_metrics_received_total",
				Help: "Total number of metrics received",
			},
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
			prometheus.CounterOpts{
				Name: "telemetry_alerts_triggered_total",
				Help: "Total number of alerts triggered",
			},
			[]string{"alert_name", "severity"},
		),
		queryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "telemetry_query_duration_seconds",
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
			},
		),
	}
	
	// Register metrics
	prometheus.MustRegister(
		s.metricsReceived,
		s.metricsStored,
		s.alertsTriggered,
		s.queryDuration,
		s.wsConnections,
		s.bufferSize,
	)
	
	// Subscribe to events
	s.subscribeToEvents()
	
	// Start background workers
	go s.metricFlusher()
	go s.alertEvaluator()
	go s.aggregator()
	go s.retentionManager()
	
	// Load alerts from database
	s.loadAlerts()
	
	return s, nil
}

// HTTP Handlers

// IngestMetrics handles metric ingestion from agents
func (s *TelemetryService) IngestMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics []MetricPoint
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
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
	
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
		return
	}
	
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

// CreateAlert creates a new alert rule
func (s *TelemetryService) CreateAlert(w http.ResponseWriter, r *http.Request) {
	var alert Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	alert.ID = generateID()
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
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	
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
	}
}

// Background Workers

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

func (s *TelemetryService) alertEvaluator() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
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
		}
	}
}

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
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
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

func (s *TelemetryService) subscribeToEvents() {
	// Subscribe to agent metrics
	s.nats.Subscribe("agent.metrics", func(msg *nats.Msg) {
		var metrics []MetricPoint
		if err := json.Unmarshal(msg.Data, &metrics); err != nil {
			return
		}
		
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

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	telemetryService, err := NewTelemetryService()
	if err != nil {
		log.Fatalf("Failed to create telemetry service: %v", err)
	}
	
	router := mux.NewRouter()
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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

