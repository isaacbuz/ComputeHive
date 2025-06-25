package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/shopspring/decimal"
)

// Offer represents a compute resource offer
type Offer struct {
	ID              string                 `json:"id"`
	ProviderID      string                 `json:"provider_id"`
	AgentID         string                 `json:"agent_id"`
	Resources       ResourceSpecification  `json:"resources"`
	PricePerHour    map[string]decimal.Decimal `json:"price_per_hour"`
	MinDuration     time.Duration          `json:"min_duration"`
	MaxDuration     time.Duration          `json:"max_duration"`
	Availability    AvailabilityWindow     `json:"availability"`
	Location        string                 `json:"location"`
	Features        []string               `json:"features"`
	SLAGuarantees   SLAGuarantees          `json:"sla_guarantees"`
	Status          string                 `json:"status"` // active, reserved, expired
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	ReservationID   string                 `json:"reservation_id,omitempty"`
}

// Bid represents a request for compute resources
type Bid struct {
	ID               string                 `json:"id"`
	ConsumerID       string                 `json:"consumer_id"`
	Requirements     ResourceRequirements   `json:"requirements"`
	MaxPricePerHour  decimal.Decimal        `json:"max_price_per_hour"`
	Duration         time.Duration          `json:"duration"`
	StartTime        time.Time              `json:"start_time"`
	Flexibility      time.Duration          `json:"flexibility"` // How flexible the start time is
	Location         string                 `json:"location,omitempty"`
	PreferredRegions []string               `json:"preferred_regions,omitempty"`
	Status           string                 `json:"status"` // pending, matched, expired, cancelled
	CreatedAt        time.Time              `json:"created_at"`
	ExpiresAt        time.Time              `json:"expires_at"`
	MatchedOfferID   string                 `json:"matched_offer_id,omitempty"`
}

// Match represents a matched bid and offer
type Match struct {
	ID             string          `json:"id"`
	BidID          string          `json:"bid_id"`
	OfferID        string          `json:"offer_id"`
	ConsumerID     string          `json:"consumer_id"`
	ProviderID     string          `json:"provider_id"`
	AgreedPrice    decimal.Decimal `json:"agreed_price"`
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time"`
	Status         string          `json:"status"` // pending, confirmed, active, completed, disputed
	ContractHash   string          `json:"contract_hash,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	ConfirmedAt    *time.Time      `json:"confirmed_at,omitempty"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
}

// ResourceSpecification details what resources are available
type ResourceSpecification struct {
	CPU         CPUSpec         `json:"cpu"`
	Memory      MemorySpec      `json:"memory"`
	GPU         []GPUSpec       `json:"gpu,omitempty"`
	Storage     StorageSpec     `json:"storage"`
	Network     NetworkSpec     `json:"network"`
}

// ResourceRequirements details what resources are needed
type ResourceRequirements struct {
	MinCPU      int      `json:"min_cpu_cores"`
	MinMemory   int      `json:"min_memory_mb"`
	MinGPU      int      `json:"min_gpu_count"`
	GPUTypes    []string `json:"gpu_types,omitempty"`
	MinStorage  int      `json:"min_storage_mb"`
	MinNetwork  int      `json:"min_network_mbps"`
	Features    []string `json:"required_features,omitempty"`
}

// Resource specification types
type CPUSpec struct {
	Cores     int    `json:"cores"`
	Model     string `json:"model"`
	Frequency string `json:"frequency"`
}

type MemorySpec struct {
	TotalMB int    `json:"total_mb"`
	Type    string `json:"type"`
	Speed   string `json:"speed"`
}

type GPUSpec struct {
	Model    string `json:"model"`
	MemoryMB int    `json:"memory_mb"`
	Count    int    `json:"count"`
}

type StorageSpec struct {
	TotalMB int    `json:"total_mb"`
	Type    string `json:"type"` // ssd, hdd, nvme
	IOPS    int    `json:"iops"`
}

type NetworkSpec struct {
	BandwidthMbps int    `json:"bandwidth_mbps"`
	Type          string `json:"type"` // dedicated, shared
}

// AvailabilityWindow represents when resources are available
type AvailabilityWindow struct {
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Recurring      bool      `json:"recurring"`
	RecurrenceRule string    `json:"recurrence_rule,omitempty"` // RFC5545 RRULE
}

// SLAGuarantees represents service level guarantees
type SLAGuarantees struct {
	Uptime       float64 `json:"uptime_percentage"`
	ResponseTime int     `json:"max_response_time_ms"`
	Support      string  `json:"support_level"` // basic, priority, enterprise
}

// MarketplaceService handles resource trading
type MarketplaceService struct {
	offers      map[string]*Offer
	bids        map[string]*Bid
	matches     map[string]*Match
	mu          sync.RWMutex
	nats        *nats.Conn
	matcher     *MatchingEngine
	wsUpgrader  websocket.Upgrader
	subscribers map[string]map[*websocket.Conn]bool // topic -> connections
	subMu       sync.RWMutex
	
	// Metrics
	offersCreated   prometheus.Counter
	bidsCreated     prometheus.Counter
	matchesCreated  prometheus.Counter
	matchingTime    prometheus.Histogram
	activeOffers    prometheus.Gauge
	activeBids      prometheus.Gauge
}

// MatchingEngine handles bid-offer matching
type MatchingEngine struct {
	service *MarketplaceService
	ticker  *time.Ticker
}

// NewMarketplaceService creates a new marketplace service
func NewMarketplaceService() (*MarketplaceService, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	
	s := &MarketplaceService{
		offers:      make(map[string]*Offer),
		bids:        make(map[string]*Bid),
		matches:     make(map[string]*Match),
		nats:        nc,
		subscribers: make(map[string]map[*websocket.Conn]bool),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Configure this properly in production
				return true
			},
		},
		
		// Initialize metrics
		offersCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "marketplace_offers_created_total",
			Help: "Total number of offers created",
		}),
		bidsCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "marketplace_bids_created_total",
			Help: "Total number of bids created",
		}),
		matchesCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "marketplace_matches_created_total",
			Help: "Total number of matches created",
		}),
		matchingTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "marketplace_matching_duration_seconds",
			Help:    "Time taken to match bids and offers",
			Buckets: prometheus.DefBuckets,
		}),
		activeOffers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "marketplace_active_offers",
			Help: "Current number of active offers",
		}),
		activeBids: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "marketplace_active_bids",
			Help: "Current number of active bids",
		}),
	}
	
	// Register metrics
	prometheus.MustRegister(
		s.offersCreated, s.bidsCreated, s.matchesCreated,
		s.matchingTime, s.activeOffers, s.activeBids,
	)
	
	// Create matching engine
	s.matcher = &MatchingEngine{
		service: s,
		ticker:  time.NewTicker(10 * time.Second), // Run matching every 10 seconds
	}
	
	// Start matching engine
	go s.matcher.run()
	
	// Subscribe to events
	s.subscribeToEvents()
	
	return s, nil
}

// HTTP Handlers

// CreateOffer handles offer creation
func (s *MarketplaceService) CreateOffer(w http.ResponseWriter, r *http.Request) {
	var offer Offer
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Extract provider ID from JWT
	claims := r.Context().Value("claims").(*Claims)
	offer.ProviderID = claims.UserID
	
	// Generate offer ID
	offer.ID = generateID()
	offer.Status = "active"
	offer.CreatedAt = time.Now()
	offer.UpdatedAt = time.Now()
	
	// Validate offer
	if err := s.validateOffer(&offer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Store offer
	s.mu.Lock()
	s.offers[offer.ID] = &offer
	s.mu.Unlock()
	
	// Update metrics
	s.offersCreated.Inc()
	s.updateActiveMetrics()
	
	// Publish event
	s.publishEvent("offer.created", &offer)
	
	// Broadcast to WebSocket subscribers
	s.broadcastUpdate("offers", map[string]interface{}{
		"type": "offer_created",
		"data": offer,
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(offer)
}

// CreateBid handles bid creation
func (s *MarketplaceService) CreateBid(w http.ResponseWriter, r *http.Request) {
	var bid Bid
	if err := json.NewDecoder(r.Body).Decode(&bid); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Extract consumer ID from JWT
	claims := r.Context().Value("claims").(*Claims)
	bid.ConsumerID = claims.UserID
	
	// Generate bid ID
	bid.ID = generateID()
	bid.Status = "pending"
	bid.CreatedAt = time.Now()
	
	// Validate bid
	if err := s.validateBid(&bid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Store bid
	s.mu.Lock()
	s.bids[bid.ID] = &bid
	s.mu.Unlock()
	
	// Update metrics
	s.bidsCreated.Inc()
	s.updateActiveMetrics()
	
	// Publish event
	s.publishEvent("bid.created", &bid)
	
	// Broadcast to WebSocket subscribers
	s.broadcastUpdate("bids", map[string]interface{}{
		"type": "bid_created",
		"data": bid,
	})
	
	// Trigger immediate matching attempt
	go s.matcher.matchBid(&bid)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bid)
}

// ListOffers returns available offers
func (s *MarketplaceService) ListOffers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	minCPU := r.URL.Query().Get("min_cpu")
	minMemory := r.URL.Query().Get("min_memory")
	maxPrice := r.URL.Query().Get("max_price")
	location := r.URL.Query().Get("location")
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var filteredOffers []*Offer
	for _, offer := range s.offers {
		// Apply filters
		if offer.Status != "active" {
			continue
		}
		
		if minCPU != "" {
			// Filter by CPU (implement actual comparison)
		}
		
		if location != "" && offer.Location != location {
			continue
		}
		
		filteredOffers = append(filteredOffers, offer)
	}
	
	// Sort by price
	sort.Slice(filteredOffers, func(i, j int) bool {
		// Compare CPU prices as example
		priceI := filteredOffers[i].PricePerHour["cpu"]
		priceJ := filteredOffers[j].PricePerHour["cpu"]
		return priceI.LessThan(priceJ)
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredOffers)
}

// GetMatch retrieves match details
func (s *MarketplaceService) GetMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["id"]
	
	s.mu.RLock()
	match, exists := s.matches[matchID]
	s.mu.RUnlock()
	
	if !exists {
		http.Error(w, "Match not found", http.StatusNotFound)
		return
	}
	
	// Check authorization
	claims := r.Context().Value("claims").(*Claims)
	if match.ConsumerID != claims.UserID && match.ProviderID != claims.UserID && claims.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(match)
}

// ConfirmMatch confirms a match (both parties must confirm)
func (s *MarketplaceService) ConfirmMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["id"]
	
	s.mu.Lock()
	match, exists := s.matches[matchID]
	if !exists {
		s.mu.Unlock()
		http.Error(w, "Match not found", http.StatusNotFound)
		return
	}
	
	// Check authorization
	claims := r.Context().Value("claims").(*Claims)
	isConsumer := match.ConsumerID == claims.UserID
	isProvider := match.ProviderID == claims.UserID
	
	if !isConsumer && !isProvider {
		s.mu.Unlock()
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}
	
	// Update confirmation status
	if match.Status == "pending" {
		match.Status = "confirmed"
		now := time.Now()
		match.ConfirmedAt = &now
		
		// Update offer and bid status
		if offer, exists := s.offers[match.OfferID]; exists {
			offer.Status = "reserved"
			offer.ReservationID = matchID
		}
		if bid, exists := s.bids[match.BidID]; exists {
			bid.Status = "matched"
			bid.MatchedOfferID = match.OfferID
		}
	}
	
	s.mu.Unlock()
	
	// Publish confirmation event
	s.publishEvent("match.confirmed", match)
	
	// Broadcast update
	s.broadcastUpdate("matches", map[string]interface{}{
		"type": "match_confirmed",
		"data": match,
	})
	
	w.WriteHeader(http.StatusNoContent)
}

// WebSocket handler for real-time updates
func (s *MarketplaceService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	
	// Subscribe to topics based on query parameters
	topics := r.URL.Query()["topic"]
	if len(topics) == 0 {
		topics = []string{"offers", "bids", "matches"} // Subscribe to all by default
	}
	
	// Register connection
	s.subMu.Lock()
	for _, topic := range topics {
		if s.subscribers[topic] == nil {
			s.subscribers[topic] = make(map[*websocket.Conn]bool)
		}
		s.subscribers[topic][conn] = true
	}
	s.subMu.Unlock()
	
	// Unregister on disconnect
	defer func() {
		s.subMu.Lock()
		for _, topic := range topics {
			delete(s.subscribers[topic], conn)
		}
		s.subMu.Unlock()
	}()
	
	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// Matching Engine implementation

func (me *MatchingEngine) run() {
	for range me.ticker.C {
		me.performMatching()
	}
}

func (me *MatchingEngine) performMatching() {
	timer := prometheus.NewTimer(me.service.matchingTime)
	defer timer.ObserveDuration()
	
	me.service.mu.RLock()
	
	// Get active bids and offers
	var activeBids []*Bid
	for _, bid := range me.service.bids {
		if bid.Status == "pending" && time.Now().Before(bid.ExpiresAt) {
			activeBids = append(activeBids, bid)
		}
	}
	
	var activeOffers []*Offer
	for _, offer := range me.service.offers {
		if offer.Status == "active" && time.Now().Before(offer.ExpiresAt) {
			activeOffers = append(activeOffers, offer)
		}
	}
	
	me.service.mu.RUnlock()
	
	// Sort bids by price (highest first)
	sort.Slice(activeBids, func(i, j int) bool {
		return activeBids[i].MaxPricePerHour.GreaterThan(activeBids[j].MaxPricePerHour)
	})
	
	// Match bids with offers
	for _, bid := range activeBids {
		me.matchBid(bid)
	}
}

func (me *MatchingEngine) matchBid(bid *Bid) {
	me.service.mu.Lock()
	defer me.service.mu.Unlock()
	
	// Skip if already matched
	if bid.Status != "pending" {
		return
	}
	
	var bestOffer *Offer
	var bestScore float64
	
	for _, offer := range me.service.offers {
		if offer.Status != "active" {
			continue
		}
		
		// Check if offer meets requirements
		if !me.offerMeetsRequirements(offer, bid) {
			continue
		}
		
		// Calculate match score
		score := me.calculateMatchScore(offer, bid)
		if score > bestScore {
			bestScore = score
			bestOffer = offer
		}
	}
	
	if bestOffer != nil {
		// Create match
		match := &Match{
			ID:          generateID(),
			BidID:       bid.ID,
			OfferID:     bestOffer.ID,
			ConsumerID:  bid.ConsumerID,
			ProviderID:  bestOffer.ProviderID,
			AgreedPrice: me.calculateAgreedPrice(bestOffer, bid),
			StartTime:   bid.StartTime,
			EndTime:     bid.StartTime.Add(bid.Duration),
			Status:      "pending",
			CreatedAt:   time.Now(),
		}
		
		me.service.matches[match.ID] = match
		
		// Update bid and offer status
		bid.Status = "matched"
		bid.MatchedOfferID = bestOffer.ID
		bestOffer.Status = "reserved"
		bestOffer.ReservationID = match.ID
		
		// Update metrics
		me.service.matchesCreated.Inc()
		me.service.updateActiveMetrics()
		
		// Publish match event
		me.service.publishEvent("match.created", match)
		
		// Broadcast update
		me.service.broadcastUpdate("matches", map[string]interface{}{
			"type": "match_created",
			"data": match,
		})
		
		log.Printf("Created match %s: bid %s with offer %s", match.ID, bid.ID, bestOffer.ID)
	}
}

func (me *MatchingEngine) offerMeetsRequirements(offer *Offer, bid *Bid) bool {
	// Check CPU requirements
	if offer.Resources.CPU.Cores < bid.Requirements.MinCPU {
		return false
	}
	
	// Check memory requirements
	if offer.Resources.Memory.TotalMB < bid.Requirements.MinMemory {
		return false
	}
	
	// Check GPU requirements
	totalGPUs := 0
	for _, gpu := range offer.Resources.GPU {
		totalGPUs += gpu.Count
	}
	if totalGPUs < bid.Requirements.MinGPU {
		return false
	}
	
	// Check storage requirements
	if offer.Resources.Storage.TotalMB < bid.Requirements.MinStorage {
		return false
	}
	
	// Check network requirements
	if offer.Resources.Network.BandwidthMbps < bid.Requirements.MinNetwork {
		return false
	}
	
	// Check price
	offerPrice := offer.PricePerHour["cpu"].Mul(decimal.NewFromInt(int64(bid.Requirements.MinCPU)))
	if bid.Requirements.MinGPU > 0 {
		gpuPrice := offer.PricePerHour["gpu"].Mul(decimal.NewFromInt(int64(bid.Requirements.MinGPU)))
		offerPrice = offerPrice.Add(gpuPrice)
	}
	
	if offerPrice.GreaterThan(bid.MaxPricePerHour) {
		return false
	}
	
	// Check availability
	bidEnd := bid.StartTime.Add(bid.Duration)
	if offer.Availability.StartTime.After(bid.StartTime) || offer.Availability.EndTime.Before(bidEnd) {
		return false
	}
	
	// Check location preferences
	if len(bid.PreferredRegions) > 0 {
		found := false
		for _, region := range bid.PreferredRegions {
			if offer.Location == region {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check required features
	for _, required := range bid.Requirements.Features {
		found := false
		for _, feature := range offer.Features {
			if feature == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

func (me *MatchingEngine) calculateMatchScore(offer *Offer, bid *Bid) float64 {
	score := 100.0
	
	// Price score (lower is better)
	offerPrice := me.calculateOfferPrice(offer, bid)
	priceRatio := offerPrice.Div(bid.MaxPricePerHour).InexactFloat64()
	score *= (2.0 - priceRatio) // Price factor: 1.0 at max price, 2.0 at free
	
	// Location score
	if offer.Location == bid.Location {
		score *= 1.2 // 20% bonus for same location
	}
	
	// Over-provisioning penalty (slight penalty for too much excess resources)
	cpuExcess := float64(offer.Resources.CPU.Cores-bid.Requirements.MinCPU) / float64(bid.Requirements.MinCPU)
	if cpuExcess > 0.5 {
		score *= (1.5 - cpuExcess*0.2) // Up to 10% penalty for 50%+ excess
	}
	
	// Feature bonus
	featureCount := 0
	for _, feature := range offer.Features {
		for _, req := range bid.Requirements.Features {
			if feature == req {
				featureCount++
			}
		}
	}
	score *= (1.0 + float64(featureCount)*0.05) // 5% bonus per matching feature
	
	// SLA bonus
	if offer.SLAGuarantees.Uptime >= 99.9 {
		score *= 1.1 // 10% bonus for high SLA
	}
	
	return score
}

func (me *MatchingEngine) calculateOfferPrice(offer *Offer, bid *Bid) decimal.Decimal {
	cpuPrice := offer.PricePerHour["cpu"].Mul(decimal.NewFromInt(int64(bid.Requirements.MinCPU)))
	memPrice := offer.PricePerHour["memory"].Mul(decimal.NewFromInt(int64(bid.Requirements.MinMemory))).Div(decimal.NewFromInt(1024))
	
	totalPrice := cpuPrice.Add(memPrice)
	
	if bid.Requirements.MinGPU > 0 {
		gpuPrice := offer.PricePerHour["gpu"].Mul(decimal.NewFromInt(int64(bid.Requirements.MinGPU)))
		totalPrice = totalPrice.Add(gpuPrice)
	}
	
	return totalPrice
}

func (me *MatchingEngine) calculateAgreedPrice(offer *Offer, bid *Bid) decimal.Decimal {
	// Simple implementation: use offer price
	// In production, could implement more sophisticated pricing strategies
	return me.calculateOfferPrice(offer, bid)
}

// Helper methods

func (s *MarketplaceService) validateOffer(offer *Offer) error {
	if offer.Resources.CPU.Cores <= 0 {
		return fmt.Errorf("CPU cores must be positive")
	}
	if offer.Resources.Memory.TotalMB <= 0 {
		return fmt.Errorf("memory must be positive")
	}
	if offer.ExpiresAt.IsZero() {
		offer.ExpiresAt = time.Now().Add(24 * time.Hour) // Default 24h expiry
	}
	if offer.MinDuration <= 0 {
		offer.MinDuration = 1 * time.Hour // Default 1h minimum
	}
	if offer.MaxDuration <= 0 {
		offer.MaxDuration = 24 * time.Hour // Default 24h maximum
	}
	return nil
}

func (s *MarketplaceService) validateBid(bid *Bid) error {
	if bid.Requirements.MinCPU <= 0 {
		return fmt.Errorf("minimum CPU cores must be positive")
	}
	if bid.Requirements.MinMemory <= 0 {
		return fmt.Errorf("minimum memory must be positive")
	}
	if bid.MaxPricePerHour.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("max price must be positive")
	}
	if bid.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	if bid.ExpiresAt.IsZero() {
		bid.ExpiresAt = time.Now().Add(1 * time.Hour) // Default 1h expiry
	}
	if bid.StartTime.IsZero() {
		bid.StartTime = time.Now() // Default to immediate start
	}
	return nil
}

func (s *MarketplaceService) updateActiveMetrics() {
	activeOffers := 0
	activeBids := 0
	
	for _, offer := range s.offers {
		if offer.Status == "active" {
			activeOffers++
		}
	}
	
	for _, bid := range s.bids {
		if bid.Status == "pending" {
			activeBids++
		}
	}
	
	s.activeOffers.Set(float64(activeOffers))
	s.activeBids.Set(float64(activeBids))
}

func (s *MarketplaceService) broadcastUpdate(topic string, data interface{}) {
	s.subMu.RLock()
	connections := s.subscribers[topic]
	s.subMu.RUnlock()
	
	if len(connections) == 0 {
		return
	}
	
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal update: %v", err)
		return
	}
	
	s.subMu.RLock()
	defer s.subMu.RUnlock()
	
	for conn := range connections {
		go func(c *websocket.Conn) {
			if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
			}
		}(conn)
	}
}

func (s *MarketplaceService) publishEvent(event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	s.nats.Publish(event, jsonData)
}

func (s *MarketplaceService) subscribeToEvents() {
	// Subscribe to agent updates to update offers
	s.nats.Subscribe("agent.status", func(msg *nats.Msg) {
		var status map[string]interface{}
		if err := json.Unmarshal(msg.Data, &status); err != nil {
			return
		}
		
		agentID := status["agent_id"].(string)
		agentStatus := status["status"].(string)
		
		// Update offers from this agent
		s.mu.Lock()
		for _, offer := range s.offers {
			if offer.AgentID == agentID && agentStatus == "offline" {
				offer.Status = "expired"
			}
		}
		s.mu.Unlock()
	})
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
	// Create marketplace service
	marketplace, err := NewMarketplaceService()
	if err != nil {
		log.Fatalf("Failed to create marketplace service: %v", err)
	}
	
	// Setup routes
	router := mux.NewRouter()
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler())
	
	// Marketplace endpoints
	router.HandleFunc("/api/v1/offers", authMiddleware(marketplace.CreateOffer)).Methods("POST")
	router.HandleFunc("/api/v1/offers", marketplace.ListOffers).Methods("GET")
	router.HandleFunc("/api/v1/bids", authMiddleware(marketplace.CreateBid)).Methods("POST")
	router.HandleFunc("/api/v1/matches/{id}", authMiddleware(marketplace.GetMatch)).Methods("GET")
	router.HandleFunc("/api/v1/matches/{id}/confirm", authMiddleware(marketplace.ConfirmMatch)).Methods("POST")
	
	// WebSocket endpoint
	router.HandleFunc("/ws", marketplace.HandleWebSocket)
	
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
		port = "8003"
	}
	
	log.Printf("Marketplace service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 