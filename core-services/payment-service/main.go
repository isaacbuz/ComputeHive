package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/shopspring/decimal"
)

// Payment represents a payment transaction
type Payment struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	Type            string          `json:"type"` // deposit, withdrawal, job_payment, refund
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"` // ETH, USDC, etc.
	Status          string          `json:"status"`   // pending, processing, completed, failed
	TxHash          string          `json:"tx_hash,omitempty"`
	FromAddress     string          `json:"from_address,omitempty"`
	ToAddress       string          `json:"to_address,omitempty"`
	JobID           string          `json:"job_id,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	FailureReason   string          `json:"failure_reason,omitempty"`
}

// Invoice represents a billing invoice
type Invoice struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	PeriodStart     time.Time       `json:"period_start"`
	PeriodEnd       time.Time       `json:"period_end"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"` // draft, pending, paid, overdue
	DueDate         time.Time       `json:"due_date"`
	PaidAt          *time.Time      `json:"paid_at,omitempty"`
	LineItems       []LineItem      `json:"line_items"`
	CreatedAt       time.Time       `json:"created_at"`
}

// LineItem represents an invoice line item
type LineItem struct {
	Description string          `json:"description"`
	Quantity    decimal.Decimal `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Amount      decimal.Decimal `json:"amount"`
	JobID       string          `json:"job_id,omitempty"`
}

// Balance represents user account balance
type Balance struct {
	UserID          string                       `json:"user_id"`
	Available       map[string]decimal.Decimal   `json:"available"`
	Pending         map[string]decimal.Decimal   `json:"pending"`
	Reserved        map[string]decimal.Decimal   `json:"reserved"`
	LastUpdated     time.Time                    `json:"last_updated"`
}

// PaymentMethod represents a user's payment method
type PaymentMethod struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Type            string                 `json:"type"` // crypto_wallet, credit_card, bank_account
	Details         map[string]interface{} `json:"details"`
	IsDefault       bool                   `json:"is_default"`
	CreatedAt       time.Time              `json:"created_at"`
}

// BlockchainConfig holds blockchain connection details
type BlockchainConfig struct {
	RPCURL          string
	ChainID         *big.Int
	ContractAddress common.Address
	PrivateKey      *ecdsa.PrivateKey
}

// PaymentService handles payment processing
type PaymentService struct {
	payments        map[string]*Payment
	invoices        map[string]*Invoice
	balances        map[string]*Balance
	paymentMethods  map[string][]*PaymentMethod
	mu              sync.RWMutex
	nats            *nats.Conn
	ethClient       *ethclient.Client
	blockchain      BlockchainConfig
	
	// Metrics
	paymentsProcessed   *prometheus.CounterVec
	paymentAmount       *prometheus.HistogramVec
	paymentDuration     *prometheus.HistogramVec
	balanceGauge        *prometheus.GaugeVec
	failedPayments      prometheus.Counter
}

// NewPaymentService creates a new payment service
func NewPaymentService() (*PaymentService, error) {
	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	
	// Connect to Ethereum
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:8545" // Default to local node
	}
	
	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum: %w", err)
	}
	
	// Parse private key for transactions
	privateKeyHex := os.Getenv("PAYMENT_PRIVATE_KEY")
	var privateKey *ecdsa.PrivateKey
	if privateKeyHex != "" {
		privateKey, err = crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}
	
	// Get contract address
	contractAddr := os.Getenv("ESCROW_CONTRACT_ADDRESS")
	if contractAddr == "" {
		contractAddr = "0x0000000000000000000000000000000000000000" // Placeholder
	}
	
	chainIDStr := os.Getenv("CHAIN_ID")
	chainID := big.NewInt(1) // Default to mainnet
	if chainIDStr != "" {
		chainID.SetString(chainIDStr, 10)
	}
	
	s := &PaymentService{
		payments:       make(map[string]*Payment),
		invoices:       make(map[string]*Invoice),
		balances:       make(map[string]*Balance),
		paymentMethods: make(map[string][]*PaymentMethod),
		nats:           nc,
		ethClient:      ethClient,
		blockchain: BlockchainConfig{
			RPCURL:          rpcURL,
			ChainID:         chainID,
			ContractAddress: common.HexToAddress(contractAddr),
			PrivateKey:      privateKey,
		},
		
		// Initialize metrics
		paymentsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_service_payments_total",
				Help: "Total number of payments processed",
			},
			[]string{"type", "status", "currency"},
		),
		paymentAmount: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "payment_service_amount",
				Help:    "Payment amounts in USD equivalent",
				Buckets: []float64{1, 10, 50, 100, 500, 1000, 5000, 10000},
			},
			[]string{"type", "currency"},
		),
		paymentDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "payment_service_duration_seconds",
				Help:    "Time taken to process payments",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"type"},
		),
		balanceGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "payment_service_user_balance",
				Help: "User balance by currency",
			},
			[]string{"user_id", "currency", "type"},
		),
		failedPayments: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "payment_service_failed_payments_total",
				Help: "Total number of failed payments",
			},
		),
	}
	
	// Register metrics
	prometheus.MustRegister(
		s.paymentsProcessed,
		s.paymentAmount,
		s.paymentDuration,
		s.balanceGauge,
		s.failedPayments,
	)
	
	// Subscribe to events
	s.subscribeToEvents()
	
	// Start background workers
	go s.paymentProcessor()
	go s.blockchainMonitor()
	go s.invoiceGenerator()
	
	return s, nil
}

// HTTP Handlers

// ProcessPayment handles payment processing requests
func (s *PaymentService) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type     string `json:"type"`
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
		JobID    string `json:"job_id,omitempty"`
		ToUserID string `json:"to_user_id,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Extract user ID from JWT
	claims := r.Context().Value("claims").(*Claims)
	userID := claims.UserID
	
	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	
	// Validate payment type
	if req.Type != "deposit" && req.Type != "withdrawal" && req.Type != "job_payment" {
		http.Error(w, "Invalid payment type", http.StatusBadRequest)
		return
	}
	
	// Create payment record
	payment := &Payment{
		ID:        generateID(),
		UserID:    userID,
		Type:      req.Type,
		Amount:    amount,
		Currency:  req.Currency,
		Status:    "pending",
		JobID:     req.JobID,
		CreatedAt: time.Now(),
	}
	
	// Store payment
	s.mu.Lock()
	s.payments[payment.ID] = payment
	s.mu.Unlock()
	
	// Process payment asynchronously
	go s.processPayment(payment)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

// GetBalance returns user balance
func (s *PaymentService) GetBalance(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	userID := claims.UserID
	
	s.mu.RLock()
	balance, exists := s.balances[userID]
	s.mu.RUnlock()
	
	if !exists {
		// Create default balance
		balance = &Balance{
			UserID:      userID,
			Available:   make(map[string]decimal.Decimal),
			Pending:     make(map[string]decimal.Decimal),
			Reserved:    make(map[string]decimal.Decimal),
			LastUpdated: time.Now(),
		}
		
		// Initialize with zero balances
		currencies := []string{"ETH", "USDC"}
		for _, currency := range currencies {
			balance.Available[currency] = decimal.Zero
			balance.Pending[currency] = decimal.Zero
			balance.Reserved[currency] = decimal.Zero
		}
		
		s.mu.Lock()
		s.balances[userID] = balance
		s.mu.Unlock()
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// GetPaymentHistory returns user's payment history
func (s *PaymentService) GetPaymentHistory(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	userID := claims.UserID
	
	// Get query parameters
	limit := 100 // Default limit
	offset := 0  // Default offset
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var userPayments []*Payment
	for _, payment := range s.payments {
		if payment.UserID == userID {
			userPayments = append(userPayments, payment)
		}
	}
	
	// Sort by creation time (newest first)
	// In production, this would be done in the database
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userPayments)
}

// GetInvoices returns user's invoices
func (s *PaymentService) GetInvoices(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	userID := claims.UserID
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var userInvoices []*Invoice
	for _, invoice := range s.invoices {
		if invoice.UserID == userID {
			userInvoices = append(userInvoices, invoice)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInvoices)
}

// AddPaymentMethod adds a new payment method
func (s *PaymentService) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    string                 `json:"type"`
		Details map[string]interface{} `json:"details"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	claims := r.Context().Value("claims").(*Claims)
	userID := claims.UserID
	
	// Validate payment method type
	if req.Type != "crypto_wallet" && req.Type != "credit_card" && req.Type != "bank_account" {
		http.Error(w, "Invalid payment method type", http.StatusBadRequest)
		return
	}
	
	// Create payment method
	method := &PaymentMethod{
		ID:        generateID(),
		UserID:    userID,
		Type:      req.Type,
		Details:   req.Details,
		IsDefault: false,
		CreatedAt: time.Now(),
	}
	
	// Store payment method
	s.mu.Lock()
	if s.paymentMethods[userID] == nil {
		s.paymentMethods[userID] = make([]*PaymentMethod, 0)
	}
	
	// Set as default if it's the first payment method
	if len(s.paymentMethods[userID]) == 0 {
		method.IsDefault = true
	}
	
	s.paymentMethods[userID] = append(s.paymentMethods[userID], method)
	s.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(method)
}

// Payment Processing

func (s *PaymentService) processPayment(payment *Payment) {
	timer := prometheus.NewTimer(s.paymentDuration.WithLabelValues(payment.Type))
	defer timer.ObserveDuration()
	
	// Update status to processing
	s.updatePaymentStatus(payment.ID, "processing", "")
	
	var err error
	switch payment.Type {
	case "deposit":
		err = s.processDeposit(payment)
	case "withdrawal":
		err = s.processWithdrawal(payment)
	case "job_payment":
		err = s.processJobPayment(payment)
	default:
		err = fmt.Errorf("unsupported payment type: %s", payment.Type)
	}
	
	if err != nil {
		s.updatePaymentStatus(payment.ID, "failed", err.Error())
		s.failedPayments.Inc()
		log.Printf("Payment %s failed: %v", payment.ID, err)
	} else {
		s.updatePaymentStatus(payment.ID, "completed", "")
		s.paymentsProcessed.WithLabelValues(payment.Type, "completed", payment.Currency).Inc()
		s.paymentAmount.WithLabelValues(payment.Type, payment.Currency).Observe(payment.Amount.InexactFloat64())
		
		// Update user balance
		s.updateBalance(payment)
		
		// Publish payment completed event
		s.publishPaymentEvent("payment.completed", payment)
	}
}

func (s *PaymentService) processDeposit(payment *Payment) error {
	// In production, this would:
	// 1. Monitor blockchain for incoming transaction
	// 2. Verify transaction confirmations
	// 3. Credit user account
	
	// For now, simulate deposit processing
	time.Sleep(2 * time.Second)
	
	// Generate transaction hash (mock)
	payment.TxHash = fmt.Sprintf("0x%x", time.Now().UnixNano())
	
	return nil
}

func (s *PaymentService) processWithdrawal(payment *Payment) error {
	// Check user balance
	s.mu.RLock()
	balance, exists := s.balances[payment.UserID]
	s.mu.RUnlock()
	
	if !exists || balance.Available[payment.Currency].LessThan(payment.Amount) {
		return fmt.Errorf("insufficient balance")
	}
	
	// Reserve funds
	s.mu.Lock()
	balance.Available[payment.Currency] = balance.Available[payment.Currency].Sub(payment.Amount)
	balance.Reserved[payment.Currency] = balance.Reserved[payment.Currency].Add(payment.Amount)
	s.mu.Unlock()
	
	// Process blockchain withdrawal
	if payment.Currency == "ETH" {
		txHash, err := s.sendETH(payment.ToAddress, payment.Amount)
		if err != nil {
			// Restore balance
			s.mu.Lock()
			balance.Available[payment.Currency] = balance.Available[payment.Currency].Add(payment.Amount)
			balance.Reserved[payment.Currency] = balance.Reserved[payment.Currency].Sub(payment.Amount)
			s.mu.Unlock()
			return err
		}
		payment.TxHash = txHash
	}
	
	return nil
}

func (s *PaymentService) processJobPayment(payment *Payment) error {
	// Process job payment through escrow contract
	// This would interact with the smart contract
	
	// For now, simulate processing
	time.Sleep(1 * time.Second)
	
	// Generate transaction hash (mock)
	payment.TxHash = fmt.Sprintf("0x%x", time.Now().UnixNano())
	
	return nil
}

func (s *PaymentService) sendETH(toAddress string, amount decimal.Decimal) (string, error) {
	if s.blockchain.PrivateKey == nil {
		return "", fmt.Errorf("no private key configured")
	}
	
	// Convert decimal to wei
	wei := new(big.Int)
	wei.SetString(amount.Mul(decimal.NewFromInt(1e18)).StringFixed(0), 10)
	
	// Get nonce
	fromAddress := crypto.PubkeyToAddress(s.blockchain.PrivateKey.PublicKey)
	nonce, err := s.ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}
	
	// Get gas price
	gasPrice, err := s.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}
	
	// Create transaction
	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(toAddress),
		wei,
		21000, // Standard ETH transfer gas limit
		gasPrice,
		nil,
	)
	
	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(s.blockchain.ChainID), s.blockchain.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}
	
	// Send transaction
	err = s.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}
	
	return signedTx.Hash().Hex(), nil
}

func (s *PaymentService) updateBalance(payment *Payment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	balance, exists := s.balances[payment.UserID]
	if !exists {
		balance = &Balance{
			UserID:      payment.UserID,
			Available:   make(map[string]decimal.Decimal),
			Pending:     make(map[string]decimal.Decimal),
			Reserved:    make(map[string]decimal.Decimal),
			LastUpdated: time.Now(),
		}
		s.balances[payment.UserID] = balance
	}
	
	switch payment.Type {
	case "deposit":
		balance.Available[payment.Currency] = balance.Available[payment.Currency].Add(payment.Amount)
	case "withdrawal":
		balance.Reserved[payment.Currency] = balance.Reserved[payment.Currency].Sub(payment.Amount)
	case "job_payment":
		// Handle job payment balance updates
		balance.Available[payment.Currency] = balance.Available[payment.Currency].Sub(payment.Amount)
	}
	
	balance.LastUpdated = time.Now()
	
	// Update metrics
	s.balanceGauge.WithLabelValues(payment.UserID, payment.Currency, "available").Set(balance.Available[payment.Currency].InexactFloat64())
	s.balanceGauge.WithLabelValues(payment.UserID, payment.Currency, "pending").Set(balance.Pending[payment.Currency].InexactFloat64())
	s.balanceGauge.WithLabelValues(payment.UserID, payment.Currency, "reserved").Set(balance.Reserved[payment.Currency].InexactFloat64())
}

func (s *PaymentService) blockchainMonitor() {
	// Monitor blockchain for incoming transactions
	// This would watch for deposits and update balances accordingly
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		// Check for pending deposits
		s.mu.RLock()
		for _, payment := range s.payments {
			if payment.Type == "deposit" && payment.Status == "pending" {
				// In production, check blockchain for transaction confirmation
				// For now, just simulate
			}
		}
		s.mu.RUnlock()
	}
}

func (s *PaymentService) invoiceGenerator() {
	// Generate monthly invoices
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// Check if it's the first day of the month
		now := time.Now()
		if now.Day() == 1 {
			s.generateMonthlyInvoices()
		}
	}
}

func (s *PaymentService) generateMonthlyInvoices() {
	// Generate invoices for all users
	// This would aggregate job costs and create invoices
	
	s.mu.RLock()
	users := make(map[string]bool)
	for _, payment := range s.payments {
		users[payment.UserID] = true
	}
	s.mu.RUnlock()
	
	for userID := range users {
		// Calculate monthly usage
		// Create invoice
		invoice := &Invoice{
			ID:          generateID(),
			UserID:      userID,
			PeriodStart: time.Now().AddDate(0, -1, 0),
			PeriodEnd:   time.Now(),
			TotalAmount: decimal.Zero,
			Currency:    "USD",
			Status:      "draft",
			DueDate:     time.Now().AddDate(0, 0, 30),
			LineItems:   []LineItem{},
			CreatedAt:   time.Now(),
		}
		
		s.mu.Lock()
		s.invoices[invoice.ID] = invoice
		s.mu.Unlock()
		
		// Publish invoice created event
		s.publishInvoiceEvent("invoice.created", invoice)
	}
}

func (s *PaymentService) subscribeToEvents() {
	// Subscribe to job completion events
	s.nats.Subscribe("job.completed", func(msg *nats.Msg) {
		var job map[string]interface{}
		if err := json.Unmarshal(msg.Data, &job); err != nil {
			return
		}
		
		s.handleJobCompletion(job)
	})
	
	// Subscribe to marketplace match events
	s.nats.Subscribe("marketplace.match.confirmed", func(msg *nats.Msg) {
		var match map[string]interface{}
		if err := json.Unmarshal(msg.Data, &match); err != nil {
			return
		}
		
		s.handleMatchConfirmed(match)
	})
}

func (s *PaymentService) handleJobCompletion(job map[string]interface{}) {
	// Process job payment
	jobID, _ := job["id"].(string)
	userID, _ := job["user_id"].(string)
	cost, _ := job["cost"].(float64)
	
	if jobID != "" && userID != "" && cost > 0 {
		payment := &Payment{
			ID:        generateID(),
			UserID:    userID,
			Type:      "job_payment",
			Amount:    decimal.NewFromFloat(cost),
			Currency:  "USD",
			Status:    "pending",
			JobID:     jobID,
			CreatedAt: time.Now(),
		}
		
		s.mu.Lock()
		s.payments[payment.ID] = payment
		s.mu.Unlock()
		
		// Process payment
		go s.processPayment(payment)
	}
}

func (s *PaymentService) handleMatchConfirmed(match map[string]interface{}) {
	// Handle marketplace match confirmation
	// This would trigger payment processing
}

func (s *PaymentService) updatePaymentStatus(paymentID, status, failureReason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if payment, exists := s.payments[paymentID]; exists {
		payment.Status = status
		if failureReason != "" {
			payment.FailureReason = failureReason
		}
		if status == "completed" {
			now := time.Now()
			payment.CompletedAt = &now
		}
	}
}

func (s *PaymentService) publishPaymentEvent(event string, payment *Payment) {
	data, _ := json.Marshal(payment)
	s.nats.Publish(event, data)
}

func (s *PaymentService) publishInvoiceEvent(event string, invoice *Invoice) {
	data, _ := json.Marshal(invoice)
	s.nats.Publish(event, data)
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
	paymentService, err := NewPaymentService()
	if err != nil {
		log.Fatalf("Failed to create payment service: %v", err)
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
	
	// Payment endpoints
	api.HandleFunc("/payments", authMiddleware(paymentService.ProcessPayment)).Methods("POST")
	api.HandleFunc("/payments/balance", authMiddleware(paymentService.GetBalance)).Methods("GET")
	api.HandleFunc("/payments", authMiddleware(paymentService.GetPaymentHistory)).Methods("GET")
	api.HandleFunc("/payments/invoices", authMiddleware(paymentService.GetInvoices)).Methods("GET")
	api.HandleFunc("/payments/methods", authMiddleware(paymentService.AddPaymentMethod)).Methods("POST")
	
	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://computehive.io"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	
	handler := c.Handler(router)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}
	
	log.Printf("Payment service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 