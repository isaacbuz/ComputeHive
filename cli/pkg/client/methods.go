package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// User methods

// GetCurrentUser gets the current authenticated user
func (c *Client) GetCurrentUser() (*User, error) {
	resp, err := c.do("GET", "/auth/user", nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := c.handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Agent methods

// ListAgents lists agents
func (c *Client) ListAgents(opts ListAgentsOptions) ([]Agent, error) {
	resp, err := c.do("GET", "/agents", nil)
	if err != nil {
		return nil, err
	}

	var agents []Agent
	if err := c.handleResponse(resp, &agents); err != nil {
		return nil, err
	}

	return agents, nil
}

// GetAgent gets a specific agent
func (c *Client) GetAgent(id string) (*Agent, error) {
	resp, err := c.do("GET", fmt.Sprintf("/agents/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var agent Agent
	if err := c.handleResponse(resp, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

// StartAgent starts an agent
func (c *Client) StartAgent(opts StartAgentOptions) (*Agent, error) {
	resp, err := c.do("POST", "/agents", opts)
	if err != nil {
		return nil, err
	}

	var agent Agent
	if err := c.handleResponse(resp, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

// StopAgent stops an agent
func (c *Client) StopAgent(id string) error {
	resp, err := c.do("DELETE", fmt.Sprintf("/agents/%s", id), nil)
	if err != nil {
		return err
	}
	return c.handleResponse(resp, nil)
}

// GetAgentConfig gets agent configuration
func (c *Client) GetAgentConfig(id string) (*AgentConfig, error) {
	resp, err := c.do("GET", fmt.Sprintf("/agents/%s/config", id), nil)
	if err != nil {
		return nil, err
	}

	var config AgentConfig
	if err := c.handleResponse(resp, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// UpdateAgentConfig updates agent configuration
func (c *Client) UpdateAgentConfig(id string, config AgentConfig) error {
	resp, err := c.do("PUT", fmt.Sprintf("/agents/%s/config", id), config)
	if err != nil {
		return err
	}
	return c.handleResponse(resp, nil)
}

// GetAgentLogs gets agent logs
func (c *Client) GetAgentLogs(id string, opts LogOptions) (<-chan LogEntry, error) {
	// In production, this would implement WebSocket or streaming
	ch := make(chan LogEntry)
	go func() {
		defer close(ch)
		// Simulate log entries
		for i := 0; i < 10; i++ {
			ch <- LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Line:      fmt.Sprintf("Log line %d", i),
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	return ch, nil
}

// Job methods

// SubmitJob submits a new job
func (c *Client) SubmitJob(spec JobSpec) (*Job, error) {
	resp, err := c.do("POST", "/jobs", spec)
	if err != nil {
		return nil, err
	}

	var job Job
	if err := c.handleResponse(resp, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// ListJobs lists jobs
func (c *Client) ListJobs(opts ListJobsOptions) ([]Job, error) {
	resp, err := c.do("GET", "/jobs", nil)
	if err != nil {
		return nil, err
	}

	var jobs []Job
	if err := c.handleResponse(resp, &jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetJob gets a specific job
func (c *Client) GetJob(id string) (*Job, error) {
	resp, err := c.do("GET", fmt.Sprintf("/jobs/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var job Job
	if err := c.handleResponse(resp, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// GetJobLogs gets job logs
func (c *Client) GetJobLogs(id string, opts LogOptions) (<-chan LogEntry, error) {
	// In production, this would implement WebSocket or streaming
	ch := make(chan LogEntry)
	go func() {
		defer close(ch)
		// Simulate log entries
		for i := 0; i < 20; i++ {
			ch <- LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Line:      fmt.Sprintf("Job log line %d", i),
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()
	return ch, nil
}

// CancelJob cancels a job
func (c *Client) CancelJob(id string, force bool) error {
	body := map[string]bool{"force": force}
	resp, err := c.do("POST", fmt.Sprintf("/jobs/%s/cancel", id), body)
	if err != nil {
		return err
	}
	return c.handleResponse(resp, nil)
}

// WaitForJob waits for job completion
func (c *Client) WaitForJob(id string, timeout time.Duration) (*Job, error) {
	deadline := time.Now().Add(timeout)
	if timeout == 0 {
		deadline = time.Now().Add(24 * time.Hour)
	}

	for time.Now().Before(deadline) {
		job, err := c.GetJob(id)
		if err != nil {
			return nil, err
		}

		if job.Status == "completed" || job.Status == "failed" || job.Status == "cancelled" {
			return job, nil
		}

		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("timeout waiting for job")
}

// ListJobResults lists job result files
func (c *Client) ListJobResults(id string) ([]ResultFile, error) {
	resp, err := c.do("GET", fmt.Sprintf("/jobs/%s/results", id), nil)
	if err != nil {
		return nil, err
	}

	var files []ResultFile
	if err := c.handleResponse(resp, &files); err != nil {
		return nil, err
	}

	return files, nil
}

// DownloadJobResults downloads job results
func (c *Client) DownloadJobResults(id, outputDir string) error {
	files, err := c.ListJobResults(id)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for _, file := range files {
		if err := c.downloadFile(fmt.Sprintf("/jobs/%s/results/%s", id, file.Path), 
			filepath.Join(outputDir, file.Name)); err != nil {
			return err
		}
	}

	return nil
}

// Marketplace methods

// ListOffers lists marketplace offers
func (c *Client) ListOffers(opts ListOffersOptions) ([]Offer, error) {
	resp, err := c.do("GET", "/marketplace/offers", nil)
	if err != nil {
		return nil, err
	}

	var offers []Offer
	if err := c.handleResponse(resp, &offers); err != nil {
		return nil, err
	}

	return offers, nil
}

// CreateOffer creates a new offer
func (c *Client) CreateOffer(offer CreateOfferRequest) (*Offer, error) {
	resp, err := c.do("POST", "/marketplace/offers", offer)
	if err != nil {
		return nil, err
	}

	var created Offer
	if err := c.handleResponse(resp, &created); err != nil {
		return nil, err
	}

	return &created, nil
}

// ListBids lists marketplace bids
func (c *Client) ListBids(opts ListBidsOptions) ([]Bid, error) {
	resp, err := c.do("GET", "/marketplace/bids", nil)
	if err != nil {
		return nil, err
	}

	var bids []Bid
	if err := c.handleResponse(resp, &bids); err != nil {
		return nil, err
	}

	return bids, nil
}

// CreateBid creates a new bid
func (c *Client) CreateBid(bid CreateBidRequest) (*Bid, error) {
	resp, err := c.do("POST", "/marketplace/bids", bid)
	if err != nil {
		return nil, err
	}

	var created Bid
	if err := c.handleResponse(resp, &created); err != nil {
		return nil, err
	}

	return &created, nil
}

// GetMarketPrices gets market prices
func (c *Client) GetMarketPrices(opts MarketPricesOptions) (*MarketPrices, error) {
	resp, err := c.do("GET", "/marketplace/prices", nil)
	if err != nil {
		return nil, err
	}

	var prices MarketPrices
	if err := c.handleResponse(resp, &prices); err != nil {
		return nil, err
	}

	return &prices, nil
}

// Status methods

// GetSystemStatus gets system status
func (c *Client) GetSystemStatus() (*SystemStatus, error) {
	resp, err := c.do("GET", "/status", nil)
	if err != nil {
		return nil, err
	}

	var status SystemStatus
	if err := c.handleResponse(resp, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// GetServiceHealth gets service health
func (c *Client) GetServiceHealth() ([]ServiceHealth, error) {
	resp, err := c.do("GET", "/status/services", nil)
	if err != nil {
		return nil, err
	}

	var services []ServiceHealth
	if err := c.handleResponse(resp, &services); err != nil {
		return nil, err
	}

	return services, nil
}

// GetJobStatistics gets job statistics
func (c *Client) GetJobStatistics(period string) (*JobStatistics, error) {
	resp, err := c.do("GET", fmt.Sprintf("/status/jobs?period=%s", period), nil)
	if err != nil {
		return nil, err
	}

	var stats JobStatistics
	if err := c.handleResponse(resp, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetAccountStatus gets account status
func (c *Client) GetAccountStatus() (*AccountStatus, error) {
	resp, err := c.do("GET", "/account/status", nil)
	if err != nil {
		return nil, err
	}

	var status AccountStatus
	if err := c.handleResponse(resp, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// Billing methods

// GetUsage gets usage report
func (c *Client) GetUsage(opts UsageOptions) (*UsageReport, error) {
	resp, err := c.do("GET", fmt.Sprintf("/billing/usage?period=%s", opts.Period), nil)
	if err != nil {
		return nil, err
	}

	var usage UsageReport
	if err := c.handleResponse(resp, &usage); err != nil {
		return nil, err
	}

	return &usage, nil
}

// ListInvoices lists invoices
func (c *Client) ListInvoices(opts ListInvoicesOptions) ([]Invoice, error) {
	resp, err := c.do("GET", "/billing/invoices", nil)
	if err != nil {
		return nil, err
	}

	var invoices []Invoice
	if err := c.handleResponse(resp, &invoices); err != nil {
		return nil, err
	}

	return invoices, nil
}

// DownloadInvoice downloads an invoice
func (c *Client) DownloadInvoice(invoiceNumber, filename string) error {
	return c.downloadFile(fmt.Sprintf("/billing/invoices/%s/download", invoiceNumber), filename)
}

// ListPaymentMethods lists payment methods
func (c *Client) ListPaymentMethods() ([]PaymentMethod, error) {
	resp, err := c.do("GET", "/billing/payment-methods", nil)
	if err != nil {
		return nil, err
	}

	var methods []PaymentMethod
	if err := c.handleResponse(resp, &methods); err != nil {
		return nil, err
	}

	return methods, nil
}

// GetPaymentMethodAddURL gets URL to add payment method
func (c *Client) GetPaymentMethodAddURL() (string, error) {
	resp, err := c.do("GET", "/billing/payment-methods/add-url", nil)
	if err != nil {
		return "", err
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return "", err
	}

	return result.URL, nil
}

// RemovePaymentMethod removes a payment method
func (c *Client) RemovePaymentMethod(id string) error {
	resp, err := c.do("DELETE", fmt.Sprintf("/billing/payment-methods/%s", id), nil)
	if err != nil {
		return err
	}
	return c.handleResponse(resp, nil)
}

// SetDefaultPaymentMethod sets default payment method
func (c *Client) SetDefaultPaymentMethod(id string) error {
	resp, err := c.do("POST", fmt.Sprintf("/billing/payment-methods/%s/default", id), nil)
	if err != nil {
		return err
	}
	return c.handleResponse(resp, nil)
}

// GetPaymentHistory gets payment history
func (c *Client) GetPaymentHistory(opts PaymentHistoryOptions) ([]Transaction, error) {
	resp, err := c.do("GET", "/billing/transactions", nil)
	if err != nil {
		return nil, err
	}

	var transactions []Transaction
	if err := c.handleResponse(resp, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetBalance gets account balance
func (c *Client) GetBalance() (*Balance, error) {
	resp, err := c.do("GET", "/billing/balance", nil)
	if err != nil {
		return nil, err
	}

	var balance Balance
	if err := c.handleResponse(resp, &balance); err != nil {
		return nil, err
	}

	return &balance, nil
}

// AddFunds adds funds to account
func (c *Client) AddFunds(req AddFundsRequest) (*Transaction, error) {
	resp, err := c.do("POST", "/billing/add-funds", req)
	if err != nil {
		return nil, err
	}

	var transaction Transaction
	if err := c.handleResponse(resp, &transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}

// ListBillingAlerts lists billing alerts
func (c *Client) ListBillingAlerts() ([]BillingAlert, error) {
	resp, err := c.do("GET", "/billing/alerts", nil)
	if err != nil {
		return nil, err
	}

	var alerts []BillingAlert
	if err := c.handleResponse(resp, &alerts); err != nil {
		return nil, err
	}

	return alerts, nil
}

// CreateBillingAlert creates a billing alert
func (c *Client) CreateBillingAlert(alert BillingAlert) (*BillingAlert, error) {
	resp, err := c.do("POST", "/billing/alerts", alert)
	if err != nil {
		return nil, err
	}

	var created BillingAlert
	if err := c.handleResponse(resp, &created); err != nil {
		return nil, err
	}

	return &created, nil
}

// RemoveBillingAlert removes a billing alert
func (c *Client) RemoveBillingAlert(id string) error {
	resp, err := c.do("DELETE", fmt.Sprintf("/billing/alerts/%s", id), nil)
	if err != nil {
		return err
	}
	return c.handleResponse(resp, nil)
}

// Token methods

// ListAPITokens lists API tokens
func (c *Client) ListAPITokens() ([]APIToken, error) {
	resp, err := c.do("GET", "/auth/tokens", nil)
	if err != nil {
		return nil, err
	}

	var tokens []APIToken
	if err := c.handleResponse(resp, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// CreateAPIToken creates an API token
func (c *Client) CreateAPIToken(req CreateTokenRequest) (*APIToken, error) {
	resp, err := c.do("POST", "/auth/tokens", req)
	if err != nil {
		return nil, err
	}

	var token APIToken
	if err := c.handleResponse(resp, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// RevokeAPIToken revokes an API token
func (c *Client) RevokeAPIToken(id string) error {
	resp, err := c.do("DELETE", fmt.Sprintf("/auth/tokens/%s", id), nil)
	if err != nil {
		return err
	}
	return c.handleResponse(resp, nil)
}

// Helper methods

// downloadFile downloads a file from the API
func (c *Client) downloadFile(path, filename string) error {
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
} 