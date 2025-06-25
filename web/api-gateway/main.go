package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
)

// Service represents a backend service
type Service struct {
	Name        string
	URL         *url.URL
	HealthCheck string
	Proxy       *httputil.ReverseProxy
}

// RateLimiter manages rate limiting per IP
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// APIGateway manages routing and middleware
type APIGateway struct {
	services    map[string]*Service
	rateLimiter *RateLimiter
	jwtSecret   []byte
	
	// Metrics
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	activeRequests  *prometheus.GaugeVec
	rateLimitHits   prometheus.Counter
}

// NewAPIGateway creates a new API Gateway instance
func NewAPIGateway() (*APIGateway, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "development-secret-change-in-production"
	}
	
	gateway := &APIGateway{
		services:    make(map[string]*Service),
		jwtSecret:   []byte(jwtSecret),
		rateLimiter: NewRateLimiter(100, 200), // 100 requests/second with burst of 200
		
		// Initialize metrics
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_gateway_requests_total",
				Help: "Total number of requests by service and status",
			},
			[]string{"service", "method", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_gateway_request_duration_seconds",
				Help:    "Request duration by service",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		),
		activeRequests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "api_gateway_active_requests",
				Help: "Number of active requests by service",
			},
			[]string{"service"},
		),
		rateLimitHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "api_gateway_rate_limit_hits_total",
				Help: "Total number of rate limit hits",
			},
		),
	}
	
	// Register metrics
	prometheus.MustRegister(
		gateway.requestsTotal,
		gateway.requestDuration,
		gateway.activeRequests,
		gateway.rateLimitHits,
	)
	
	// Initialize services
	gateway.initializeServices()
	
	// Start health check routine
	go gateway.healthCheckRoutine()
	
	return gateway, nil
}

// initializeServices sets up backend service configurations
func (g *APIGateway) initializeServices() {
	// Service configurations
	serviceConfigs := []struct {
		name        string
		envVar      string
		defaultURL  string
		healthCheck string
	}{
		{"auth", "AUTH_SERVICE_URL", "http://localhost:8001", "/health"},
		{"scheduler", "SCHEDULER_SERVICE_URL", "http://localhost:8002", "/health"},
		{"marketplace", "MARKETPLACE_SERVICE_URL", "http://localhost:8003", "/health"},
		{"payment", "PAYMENT_SERVICE_URL", "http://localhost:8004", "/health"},
		{"telemetry", "TELEMETRY_SERVICE_URL", "http://localhost:8005", "/health"},
		{"resource", "RESOURCE_SERVICE_URL", "http://localhost:8006", "/health"},
	}
	
	for _, config := range serviceConfigs {
		serviceURL := os.Getenv(config.envVar)
		if serviceURL == "" {
			serviceURL = config.defaultURL
		}
		
		parsedURL, err := url.Parse(serviceURL)
		if err != nil {
			log.Printf("Failed to parse URL for service %s: %v", config.name, err)
			continue
		}
		
		proxy := httputil.NewSingleHostReverseProxy(parsedURL)
		
		// Customize proxy error handling
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error for service %s: %v", config.name, err)
			http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
		}
		
		// Add custom headers
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Header.Set("X-Forwarded-Service", config.name)
			req.Header.Set("X-Request-ID", generateRequestID())
		}
		
		g.services[config.name] = &Service{
			Name:        config.name,
			URL:         parsedURL,
			HealthCheck: config.healthCheck,
			Proxy:       proxy,
		}
		
		log.Printf("Registered service: %s -> %s", config.name, serviceURL)
	}
}

// Middleware functions

// loggingMiddleware logs all requests
func (g *APIGateway) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
		
		// Extract service name from path
		serviceName := extractServiceName(r.URL.Path)
		
		// Increment active requests
		g.activeRequests.WithLabelValues(serviceName).Inc()
		defer g.activeRequests.WithLabelValues(serviceName).Dec()
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Record metrics
		duration := time.Since(start).Seconds()
		g.requestDuration.WithLabelValues(serviceName, r.Method).Observe(duration)
		g.requestsTotal.WithLabelValues(serviceName, r.Method, fmt.Sprintf("%d", wrapped.statusCode)).Inc()
		
		// Log request
		log.Printf(
			"[%s] %s %s %d %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.statusCode,
			time.Since(start),
			r.UserAgent(),
		)
	})
}

// rateLimitMiddleware implements rate limiting per IP
func (g *APIGateway) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract client IP
		clientIP := getClientIP(r)
		
		// Get or create visitor
		visitor := g.rateLimiter.GetVisitor(clientIP)
		
		// Check rate limit
		if !visitor.limiter.Allow() {
			g.rateLimitHits.Inc()
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// authMiddleware validates JWT tokens for protected routes
func (g *APIGateway) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for certain paths
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/v1/auth/login") ||
			strings.HasPrefix(path, "/api/v1/auth/register") ||
			strings.HasPrefix(path, "/api/v1/auth/refresh") ||
			strings.HasPrefix(path, "/health") ||
			strings.HasPrefix(path, "/metrics") {
			next.ServeHTTP(w, r)
			return
		}
		
		// Extract token from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}
		
		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}
		
		tokenString := parts[1]
		
		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return g.jwtSecret, nil
		})
		
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		// Add claims to request header for downstream services
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, ok := claims["user_id"].(string); ok {
				r.Header.Set("X-User-ID", userID)
			}
			if role, ok := claims["role"].(string); ok {
				r.Header.Set("X-User-Role", role)
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// Route handlers

// routeRequest routes requests to appropriate backend services
func (g *APIGateway) routeRequest(w http.ResponseWriter, r *http.Request) {
	// Extract service name from path
	serviceName := extractServiceName(r.URL.Path)
	
	service, exists := g.services[serviceName]
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}
	
	// Strip /api/v1/{service} prefix
	r.URL.Path = strings.TrimPrefix(r.URL.Path, fmt.Sprintf("/api/v1/%s", serviceName))
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}
	
	// Forward request to service
	service.Proxy.ServeHTTP(w, r)
}

// Health check endpoints

// healthCheck returns gateway health status
func (g *APIGateway) healthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().UTC(),
		"services": make(map[string]string),
	}
	
	// Check each service health
	for name, service := range g.services {
		if g.checkServiceHealth(service) {
			health["services"].(map[string]string)[name] = "healthy"
		} else {
			health["services"].(map[string]string)[name] = "unhealthy"
			health["status"] = "degraded"
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// checkServiceHealth checks if a service is healthy
func (g *APIGateway) checkServiceHealth(service *Service) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	
	healthURL := service.URL.String() + service.HealthCheck
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// healthCheckRoutine periodically checks service health
func (g *APIGateway) healthCheckRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		for name, service := range g.services {
			if g.checkServiceHealth(service) {
				log.Printf("Service %s is healthy", name)
			} else {
				log.Printf("Service %s is unhealthy", name)
			}
		}
	}
}

// Special handlers

// handleWebSocket proxies WebSocket connections
func (g *APIGateway) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract service name
	serviceName := extractServiceName(r.URL.Path)
	
	service, exists := g.services[serviceName]
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}
	
	// WebSocket proxy requires special handling
	target := service.URL
	targetURL := "ws://" + target.Host + strings.TrimPrefix(r.URL.Path, fmt.Sprintf("/api/v1/%s", serviceName))
	
	// Create WebSocket proxy
	proxyURL, _ := url.Parse(targetURL)
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	
	// Forward the request
	proxy.ServeHTTP(w, r)
}

// Helper functions

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    b,
	}
	
	// Cleanup old visitors periodically
	go rl.cleanupVisitors()
	
	return rl
}

// GetVisitor gets or creates a visitor
func (rl *RateLimiter) GetVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	} else {
		v.lastSeen = time.Now()
	}
	
	return v
}

// cleanupVisitors removes old visitors
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// extractServiceName extracts service name from path
func extractServiceName(path string) string {
	// Path format: /api/v1/{service}/...
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Get first IP in the chain
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	
	return ip
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", timestamp)))
	return hex.EncodeToString(hash[:])[:16]
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Admin endpoints

// reloadConfig reloads gateway configuration
func (g *APIGateway) reloadConfig(w http.ResponseWriter, r *http.Request) {
	// Verify admin role
	if r.Header.Get("X-User-Role") != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	// Reinitialize services
	g.initializeServices()
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Configuration reloaded"))
}

// getStats returns gateway statistics
func (g *APIGateway) getStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"services":     len(g.services),
		"rate_limiter": map[string]interface{}{
			"visitors": len(g.rateLimiter.visitors),
			"rate":     g.rateLimiter.rate,
			"burst":    g.rateLimiter.burst,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func main() {
	// Create API Gateway
	gateway, err := NewAPIGateway()
	if err != nil {
		log.Fatalf("Failed to create API Gateway: %v", err)
	}
	
	// Create router
	router := mux.NewRouter()
	
	// Apply global middleware
	router.Use(gateway.loggingMiddleware)
	router.Use(gateway.rateLimitMiddleware)
	
	// Health and metrics endpoints (no auth required)
	router.HandleFunc("/health", gateway.healthCheck).Methods("GET")
	router.Handle("/metrics", promhttp.Handler())
	
	// Admin endpoints
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(gateway.authMiddleware)
	adminRouter.HandleFunc("/reload", gateway.reloadConfig).Methods("POST")
	adminRouter.HandleFunc("/stats", gateway.getStats).Methods("GET")
	
	// API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(gateway.authMiddleware)
	
	// WebSocket routes (special handling)
	apiRouter.HandleFunc("/marketplace/ws", gateway.handleWebSocket)
	apiRouter.HandleFunc("/telemetry/ws", gateway.handleWebSocket)
	
	// Service routes
	apiRouter.PathPrefix("/").HandlerFunc(gateway.routeRequest)
	
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://computehive.io"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	
	handler := c.Handler(router)
	
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	
	log.Printf("API Gateway starting on port %s", port)
	log.Printf("Registered services: %d", len(gateway.services))
	
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 