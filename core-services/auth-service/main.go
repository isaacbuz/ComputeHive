package main

import (
<<<<<<< HEAD
=======
	"context"
	"crypto/rand"
	"encoding/base64"
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	"encoding/json"
	"fmt"
	"log"
	"net/http"
<<<<<<< HEAD
	"time"
	
	"github.com/gorilla/mux"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Service:   "auth-service",
		Timestamp: time.Now(),
	}
	
=======
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user account
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
}

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Role     string   `json:"role"`
	Scopes   []string `json:"scopes"`
	jwt.RegisteredClaims
}

// AuthService handles authentication operations
type AuthService struct {
	jwtSecret     []byte
	tokenDuration time.Duration
	users         map[string]*User // In production, use a database
	refreshTokens map[string]string // Maps refresh tokens to user IDs
}

// NewAuthService creates a new authentication service
func NewAuthService() *AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Generate a random secret in development
		b := make([]byte, 32)
		rand.Read(b)
		secret = base64.URLEncoding.EncodeToString(b)
		log.Printf("WARNING: Using generated JWT secret. Set JWT_SECRET environment variable in production.")
	}

	return &AuthService{
		jwtSecret:     []byte(secret),
		tokenDuration: 24 * time.Hour,
		users:         make(map[string]*User),
		refreshTokens: make(map[string]string),
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Register handles user registration
func (s *AuthService) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		http.Error(w, "Email, username, and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	for _, user := range s.users {
		if user.Email == req.Email {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}
		if user.Username == req.Username {
			http.Error(w, "Username already taken", http.StatusConflict)
			return
		}
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Create user
	user := &User{
		ID:           generateID(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	s.users[user.ID] = user

	// Generate tokens
	tokenResp, err := s.generateTokens(user)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResp)
}

// Login handles user login
func (s *AuthService) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find user by email
	var user *User
	for _, u := range s.users {
		if u.Email == req.Email {
			user = u
			break
		}
	}

	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if user is active
	if !user.IsActive {
		http.Error(w, "Account is disabled", http.StatusForbidden)
		return
	}

	// Generate tokens
	tokenResp, err := s.generateTokens(user)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResp)
}

// RefreshToken handles token refresh
func (s *AuthService) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	if refreshToken == "" {
		// Try to get from Authorization header
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			refreshToken = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	if refreshToken == "" {
		http.Error(w, "Refresh token required", http.StatusBadRequest)
		return
	}

	// Find user ID from refresh token
	userID, exists := s.refreshTokens[refreshToken]
	if !exists {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Get user
	user, exists := s.users[userID]
	if !exists || !user.IsActive {
		http.Error(w, "User not found or inactive", http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	tokenResp, err := s.generateTokens(user)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Revoke old refresh token
	delete(s.refreshTokens, refreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResp)
}

// Validate validates a token
func (s *AuthService) Validate(w http.ResponseWriter, r *http.Request) {
	tokenString := extractToken(r)
	if tokenString == "" {
		http.Error(w, "Token required", http.StatusBadRequest)
		return
	}

	claims, err := s.validateToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"valid":    true,
		"user_id":  claims.UserID,
		"email":    claims.Email,
		"username": claims.Username,
		"role":     claims.Role,
		"scopes":   claims.Scopes,
		"exp":      claims.ExpiresAt.Unix(),
	}

>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

<<<<<<< HEAD
func main() {
	fmt.Println("🔐 ComputeHive Auth Service v1.0.0")
	
	router := mux.NewRouter()
	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.HandleFunc("/api/v1/agents/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Agent registered successfully",
			"agentId": "agent-" + fmt.Sprintf("%d", time.Now().Unix()),
		})
	}).Methods("POST")
	
	fmt.Println("Auth service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}
=======
// generateTokens generates access and refresh tokens
func (s *AuthService) generateTokens(user *User) (*TokenResponse, error) {
	// Generate access token
	expiresAt := time.Now().Add(s.tokenDuration)
	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		Scopes:   s.getUserScopes(user),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "computehive-auth",
			Subject:   user.ID,
			ID:        generateID(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken := generateRefreshToken()
	s.refreshTokens[refreshToken] = user.ID

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.tokenDuration.Seconds()),
		ExpiresAt:    expiresAt,
	}, nil
}

// validateToken validates a JWT token
func (s *AuthService) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// getUserScopes returns the scopes for a user based on their role
func (s *AuthService) getUserScopes(user *User) []string {
	switch user.Role {
	case "admin":
		return []string{"read:all", "write:all", "admin:all"}
	case "provider":
		return []string{"read:jobs", "write:jobs", "read:agents", "write:agents"}
	case "consumer":
		return []string{"read:jobs", "write:jobs", "read:results"}
	default:
		return []string{"read:profile", "write:profile"}
	}
}

// Middleware provides JWT authentication middleware
func (s *AuthService) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := extractToken(r)
		if tokenString == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		claims, err := s.validateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), "claims", claims)
		next(w, r.WithContext(ctx))
	}
}

// Helper functions

func extractToken(r *http.Request) string {
	// Try Authorization header first
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Try query parameter
	return r.URL.Query().Get("token")
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func generateRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func main() {
	// Initialize service
	authService := NewAuthService()

	// Create demo users
	demoPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	authService.users["demo-admin"] = &User{
		ID:           "demo-admin",
		Email:        "admin@computehive.io",
		Username:     "admin",
		PasswordHash: string(demoPasswordHash),
		Role:         "admin",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	// Setup routes
	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Auth routes
	router.HandleFunc("/api/v1/auth/register", authService.Register).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", authService.Login).Methods("POST")
	router.HandleFunc("/api/v1/auth/refresh", authService.RefreshToken).Methods("POST")
	router.HandleFunc("/api/v1/auth/validate", authService.Validate).Methods("GET")

	// Protected route example
	router.HandleFunc("/api/v1/auth/profile", authService.Middleware(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value("claims").(*Claims)
		response := map[string]interface{}{
			"user_id":  claims.UserID,
			"email":    claims.Email,
			"username": claims.Username,
			"role":     claims.Role,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://computehive.io"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            os.Getenv("ENV") == "development",
	})

	handler := c.Handler(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	log.Printf("Auth service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
