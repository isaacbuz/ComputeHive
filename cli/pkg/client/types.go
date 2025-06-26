package client

import "time"

// User represents a ComputeHive user
type User struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	Organization   string     `json:"organization,omitempty"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
}

// Agent represents a compute agent
type Agent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Status      string         `json:"status"`
	Version     string         `json:"version"`
	OS          string         `json:"os"`
	Arch        string         `json:"arch"`
	LastSeen    time.Time      `json:"last_seen"`
	Resources   AgentResources `json:"resources"`
	CurrentJob  string         `json:"current_job,omitempty"`
	Stats       AgentStats     `json:"stats"`
}

// AgentResources represents agent hardware resources
type AgentResources struct {
	CPUCores     int     `json:"cpu_cores"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryGB     int     `json:"memory_gb"`
	MemoryUsage  float64 `json:"memory_usage"`
	StorageGB    int     `json:"storage_gb"`
	StorageUsage float64 `json:"storage_usage"`
	GPUCount     int     `json:"gpu_count"`
	GPUModel     string  `json:"gpu_model,omitempty"`
	GPUs         []GPU   `json:"gpus,omitempty"`
}

// GPU represents a single GPU
type GPU struct {
	Index       int     `json:"index"`
	Model       string  `json:"model"`
	MemoryMB    int     `json:"memory_mb"`
	Usage       float64 `json:"usage"`
	Temperature float64 `json:"temperature"`
}

// AgentStats represents agent statistics
type AgentStats struct {
	JobsCompleted int           `json:"jobs_completed"`
	SuccessRate   float64       `json:"success_rate"`
	TotalRuntime  time.Duration `json:"total_runtime"`
}

// Job represents a compute job
type Job struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Status          string                `json:"status"`
	Type            string                `json:"type"`
	Priority        string                `json:"priority"`
	DockerImage     string                `json:"docker_image,omitempty"`
	Script          string                `json:"script,omitempty"`
	ScriptName      string                `json:"script_name,omitempty"`
	Command         []string              `json:"command,omitempty"`
	Environment     map[string]string     `json:"environment,omitempty"`
	Resources       ResourceRequirements  `json:"resources"`
	AssignedAgentID string                `json:"assigned_agent_id,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	StartedAt       *time.Time            `json:"started_at,omitempty"`
	CompletedAt     *time.Time            `json:"completed_at,omitempty"`
	ExitCode        int                   `json:"exit_code"`
	Error           string                `json:"error,omitempty"`
	Progress        float64               `json:"progress"`
	EstimatedCost   float64               `json:"estimated_cost"`
	ActualCost      float64               `json:"actual_cost"`
}

// JobSpec represents job submission specification
type JobSpec struct {
	Name        string                `json:"name"`
	Type        string                `json:"type"`
	DockerImage string                `json:"docker_image,omitempty"`
	Script      string                `json:"script,omitempty"`
	ScriptName  string                `json:"script_name,omitempty"`
	Command     []string              `json:"command,omitempty"`
	Environment map[string]string     `json:"environment,omitempty"`
	Volumes     []VolumeMount         `json:"volumes,omitempty"`
	Resources   ResourceRequirements  `json:"resources"`
	MaxRuntime  string                `json:"max_runtime,omitempty"`
	Priority    string                `json:"priority,omitempty"`
}

// ResourceRequirements represents compute resource requirements
type ResourceRequirements struct {
	CPUCores  int    `json:"cpu_cores"`
	MemoryGB  int    `json:"memory_gb"`
	GPUCount  int    `json:"gpu_count,omitempty"`
	GPUModel  string `json:"gpu_model,omitempty"`
	StorageGB int    `json:"storage_gb"`
}

// VolumeMount represents a volume mount
type VolumeMount struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly      bool   `json:"read_only,omitempty"`
}

// Offer represents a marketplace offer
type Offer struct {
	ID               string               `json:"id"`
	ProviderID       string               `json:"provider_id"`
	ProviderName     string               `json:"provider_name"`
	ResourceType     string               `json:"resource_type"`
	CPUCores         int                  `json:"cpu_cores"`
	MemoryGB         int                  `json:"memory_gb"`
	GPUCount         int                  `json:"gpu_count"`
	GPUModel         string               `json:"gpu_model,omitempty"`
	StorageGB        int                  `json:"storage_gb"`
	NetworkBandwidth float64              `json:"network_bandwidth"`
	PricePerHour     float64              `json:"price_per_hour"`
	Location         string               `json:"location"`
	ReputationScore  float64              `json:"reputation_score"`
	Status           string               `json:"status"`
	ExpiresAt        time.Time            `json:"expires_at"`
}

// Bid represents a marketplace bid
type Bid struct {
	ID              string               `json:"id"`
	ConsumerID      string               `json:"consumer_id"`
	ConsumerName    string               `json:"consumer_name"`
	Requirements    ResourceRequirements `json:"requirements"`
	MaxPricePerHour float64              `json:"max_price_per_hour"`
	DurationHours   int                  `json:"duration_hours"`
	Deadline        time.Time            `json:"deadline"`
	Status          string               `json:"status"`
}

// MarketPrices represents market pricing information
type MarketPrices struct {
	CPU     PriceStats             `json:"cpu"`
	GPU     map[string]PriceStats  `json:"gpu"`
	Memory  PriceStats             `json:"memory"`
	Storage PriceStats             `json:"storage"`
}

// PriceStats represents price statistics
type PriceStats struct {
	Average      float64 `json:"average"`
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	TrendPercent float64 `json:"trend_percent"`
}

// SystemStatus represents overall system status
type SystemStatus struct {
	Overall              string                `json:"overall"`
	APIVersion           string                `json:"api_version"`
	LastUpdated          time.Time             `json:"last_updated"`
	Services             []ServiceHealth       `json:"services"`
	Stats                SystemStats           `json:"stats"`
	RecentIncidents      []Incident            `json:"recent_incidents"`
	ScheduledMaintenance []MaintenanceWindow   `json:"scheduled_maintenance"`
}

// ServiceHealth represents health of a service
type ServiceHealth struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	ResponseTime float64   `json:"response_time"`
	Version      string    `json:"version"`
	StartTime    time.Time `json:"start_time"`
}

// SystemStats represents system statistics
type SystemStats struct {
	ActiveAgents   int     `json:"active_agents"`
	RunningJobs    int     `json:"running_jobs"`
	AvailableGPUs  int     `json:"available_gpus"`
	TotalCapacity  float64 `json:"total_capacity"`
}

// Incident represents a system incident
type Incident struct {
	Time    time.Time `json:"time"`
	Service string    `json:"service"`
	Message string    `json:"message"`
}

// MaintenanceWindow represents scheduled maintenance
type MaintenanceWindow struct {
	StartTime   time.Time `json:"start_time"`
	Duration    string    `json:"duration"`
	Description string    `json:"description"`
}

// UsageReport represents resource usage
type UsageReport struct {
	TotalCost        float64                `json:"total_cost"`
	ComputeCost      float64                `json:"compute_cost"`
	ComputePercent   float64                `json:"compute_percent"`
	StorageCost      float64                `json:"storage_cost"`
	StoragePercent   float64                `json:"storage_percent"`
	NetworkCost      float64                `json:"network_cost"`
	NetworkPercent   float64                `json:"network_percent"`
	OtherCost        float64                `json:"other_cost"`
	CPUHours         float64                `json:"cpu_hours"`
	GPUHours         float64                `json:"gpu_hours"`
	StorageGBHours   float64                `json:"storage_gb_hours"`
	NetworkGB        float64                `json:"network_gb"`
	DailyBreakdown   []DailyUsage           `json:"daily_breakdown"`
	JobTypeCosts     map[string]float64     `json:"job_type_costs"`
	ProjectedMonthly float64                `json:"projected_monthly"`
	ProjectedAnnual  float64                `json:"projected_annual"`
	TrendPercent     float64                `json:"trend_percent"`
}

// DailyUsage represents daily usage breakdown
type DailyUsage struct {
	Date    time.Time `json:"date"`
	Compute float64   `json:"compute"`
	Storage float64   `json:"storage"`
	Network float64   `json:"network"`
	Total   float64   `json:"total"`
}

// Invoice represents a billing invoice
type Invoice struct {
	ID       string     `json:"id"`
	Number   string     `json:"number"`
	Date     time.Time  `json:"date"`
	DueDate  *time.Time `json:"due_date,omitempty"`
	Total    float64    `json:"total"`
	Status   string     `json:"status"`
	Items    []InvoiceItem `json:"items"`
}

// InvoiceItem represents an invoice line item
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID          string              `json:"id"`
	Type        string              `json:"type"`
	IsDefault   bool                `json:"is_default"`
	Card        *CardInfo           `json:"card,omitempty"`
	BankAccount *BankAccountInfo    `json:"bank_account,omitempty"`
	Crypto      *CryptoInfo         `json:"crypto,omitempty"`
}

// CardInfo represents credit card information
type CardInfo struct {
	Brand    string `json:"brand"`
	Last4    string `json:"last4"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
}

// BankAccountInfo represents bank account information
type BankAccountInfo struct {
	BankName string `json:"bank_name"`
	Last4    string `json:"last4"`
}

// CryptoInfo represents cryptocurrency information
type CryptoInfo struct {
	Currency string `json:"currency"`
	Address  string `json:"address"`
}

// Transaction represents a payment transaction
type Transaction struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	NewBalance  float64   `json:"new_balance,omitempty"`
}

// Balance represents account balance
type Balance struct {
	Available    float64 `json:"available"`
	Pending      float64 `json:"pending"`
	CreditLimit  float64 `json:"credit_limit"`
}

// AccountStatus represents account status
type AccountStatus struct {
	ID           string         `json:"id"`
	Email        string         `json:"email"`
	Plan         string         `json:"plan"`
	Status       string         `json:"status"`
	Organization string         `json:"organization,omitempty"`
	Balance      Balance        `json:"balance"`
	Usage        AccountUsage   `json:"usage"`
	Quotas       AccountQuotas  `json:"quotas"`
	Warnings     []string       `json:"warnings,omitempty"`
}

// AccountUsage represents account usage
type AccountUsage struct {
	Compute float64 `json:"compute"`
	Storage float64 `json:"storage"`
	Network float64 `json:"network"`
	Total   float64 `json:"total"`
}

// AccountQuotas represents account quotas
type AccountQuotas struct {
	MaxAgents     int `json:"max_agents"`
	AgentsUsed    int `json:"agents_used"`
	MaxJobsPerDay int `json:"max_jobs_per_day"`
	JobsToday     int `json:"jobs_today"`
	MaxGPUs       int `json:"max_gpus"`
	GPUsUsed      int `json:"gpus_used"`
}

// APIToken represents an API token
type APIToken struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Token     string     `json:"token,omitempty"`
	Scopes    []string   `json:"scopes"`
	CreatedAt time.Time  `json:"created_at"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
}

// JobStatistics represents job statistics
type JobStatistics struct {
	TotalJobs      int                 `json:"total_jobs"`
	Completed      int                 `json:"completed"`
	CompletionRate float64             `json:"completion_rate"`
	Failed         int                 `json:"failed"`
	FailureRate    float64             `json:"failure_rate"`
	Running        int                 `json:"running"`
	Pending        int                 `json:"pending"`
	AvgQueueTime   time.Duration       `json:"avg_queue_time"`
	AvgRuntime     time.Duration       `json:"avg_runtime"`
	SuccessRate    float64             `json:"success_rate"`
	TotalCPUHours  float64             `json:"total_cpu_hours"`
	TotalGPUHours  float64             `json:"total_gpu_hours"`
	TotalCost      float64             `json:"total_cost"`
	AvgCostPerJob  float64             `json:"avg_cost_per_job"`
	TopErrors      []ErrorCount        `json:"top_errors"`
	JobTypes       map[string]int      `json:"job_types"`
}

// ErrorCount represents error frequency
type ErrorCount struct {
	Error string `json:"error"`
	Count int    `json:"count"`
}

// BillingAlert represents a billing alert
type BillingAlert struct {
	ID            string     `json:"id"`
	Type          string     `json:"type"`
	Threshold     float64    `json:"threshold"`
	CurrentValue  float64    `json:"current_value"`
	Enabled       bool       `json:"enabled"`
	LastTriggered *time.Time `json:"last_triggered,omitempty"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Line      string    `json:"line"`
}

// ResultFile represents a job result file
type ResultFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Path string `json:"path"`
} 