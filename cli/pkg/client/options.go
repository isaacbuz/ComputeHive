package client

import "time"

// ListAgentsOptions represents options for listing agents
type ListAgentsOptions struct {
	All    bool
	Status string
	Limit  int
}

// ListJobsOptions represents options for listing jobs
type ListJobsOptions struct {
	Status string
	Limit  int
	Since  *time.Time
	UserID string
	All    bool
}

// ListOffersOptions represents options for listing offers
type ListOffersOptions struct {
	ResourceType string
	Location     string
	MinCPU       int
	MinMemory    int
	MinGPU       int
	MaxPrice     float64
	Limit        int
}

// ListBidsOptions represents options for listing bids
type ListBidsOptions struct {
	Status string
	Limit  int
	Mine   bool
}

// CreateOfferRequest represents a request to create an offer
type CreateOfferRequest struct {
	CPUCores         int           `json:"cpu_cores"`
	MemoryGB         int           `json:"memory_gb"`
	GPUCount         int           `json:"gpu_count"`
	GPUModel         string        `json:"gpu_model,omitempty"`
	StorageGB        int           `json:"storage_gb"`
	NetworkBandwidth float64       `json:"network_bandwidth"`
	PricePerHour     float64       `json:"price_per_hour"`
	Location         string        `json:"location"`
	Duration         time.Duration `json:"duration"`
	AutoAccept       bool          `json:"auto_accept"`
}

// CreateBidRequest represents a request to create a bid
type CreateBidRequest struct {
	Requirements    ResourceRequirements `json:"requirements"`
	MaxPricePerHour float64              `json:"max_price_per_hour"`
	DurationHours   int                  `json:"duration_hours"`
	Deadline        time.Time            `json:"deadline"`
	Location        string               `json:"location,omitempty"`
}

// MarketPricesOptions represents options for getting market prices
type MarketPricesOptions struct {
	ResourceType string
	Location     string
	Period       string
}

// LogOptions represents options for getting logs
type LogOptions struct {
	Follow bool
	Tail   int
	Since  string
}

// UsageOptions represents options for getting usage
type UsageOptions struct {
	Period  string
	Details bool
}

// ListInvoicesOptions represents options for listing invoices
type ListInvoicesOptions struct {
	Limit  int
	Status string
}

// PaymentHistoryOptions represents options for payment history
type PaymentHistoryOptions struct {
	Limit  int
	Filter string
}

// AddFundsRequest represents a request to add funds
type AddFundsRequest struct {
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method,omitempty"`
}

// CreateTokenRequest represents a request to create an API token
type CreateTokenRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// StartAgentOptions represents options for starting an agent
type StartAgentOptions struct {
	Name       string            `json:"name"`
	Tags       []string          `json:"tags"`
	Labels     map[string]string `json:"labels"`
	Datacenter string            `json:"datacenter"`
}

// AgentConfig represents agent configuration
type AgentConfig struct {
	Enabled    bool              `json:"enabled"`
	MaxJobs    int               `json:"max_jobs"`
	MaxPrice   float64           `json:"max_price"`
	Tags       []string          `json:"tags"`
	Labels     map[string]string `json:"labels"`
	Datacenter string            `json:"datacenter"`
} 