package core

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"crypto/x509"
)

// ControlPlaneClient handles communication with the control plane
type ControlPlaneClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
	authToken  string
}

// NewControlPlaneClient creates a new control plane client
func NewControlPlaneClient(baseURL string, securityConfig SecurityConfig, logger *zap.Logger) (*ControlPlaneClient, error) {
	// Configure HTTP client
	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 10,
	}

	// Configure TLS if enabled
	if securityConfig.EnableTLS {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		// Load client certificates if provided
		if securityConfig.CertFile != "" && securityConfig.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(securityConfig.CertFile, securityConfig.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificates: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		// Load CA certificate if provided
		if securityConfig.CAFile != "" {
			caCert, err := os.ReadFile(securityConfig.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load CA certificate: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		}

		transport.TLSClientConfig = tlsConfig
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &ControlPlaneClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// Register registers the agent with the control plane
func (c *ControlPlaneClient) Register(ctx context.Context, req *RegisterRequest) error {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/agents/register", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	var registerResp RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		return fmt.Errorf("failed to decode register response: %w", err)
	}

	if !registerResp.Success {
		return fmt.Errorf("registration failed: %s", registerResp.Message)
	}

	// Store auth token if provided
	if registerResp.Token != "" {
		c.authToken = registerResp.Token
	}

	return nil
}

// Deregister removes the agent from the control plane
func (c *ControlPlaneClient) Deregister(ctx context.Context, req *DeregisterRequest) error {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/agents/deregister", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// Heartbeat sends a heartbeat to the control plane
func (c *ControlPlaneClient) Heartbeat(ctx context.Context, req *HeartbeatRequest) error {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/agents/heartbeat", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	var heartbeatResp HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&heartbeatResp); err != nil {
		return fmt.Errorf("failed to decode heartbeat response: %w", err)
	}

	// Process any commands from the control plane
	for _, cmd := range heartbeatResp.Commands {
		c.logger.Info("Received command from control plane", zap.String("command", cmd))
		// Process commands - this would typically be handled by a command processor
		// For now, we just log them. In a full implementation, commands might include:
		// - Update agent configuration
		// - Restart agent
		// - Clear cache
		// - Update security credentials
		// The actual processing would be delegated to the agent's command handler
	}

	return nil
}

// PollJobs requests new jobs from the control plane
func (c *ControlPlaneClient) PollJobs(ctx context.Context, req *JobPollRequest) ([]*Job, error) {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/jobs/poll", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		// No jobs available
		return []*Job{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var pollResp JobPollResponse
	if err := json.NewDecoder(resp.Body).Decode(&pollResp); err != nil {
		return nil, fmt.Errorf("failed to decode job poll response: %w", err)
	}

	return pollResp.Jobs, nil
}

// ReportJobResult reports job execution results
func (c *ControlPlaneClient) ReportJobResult(ctx context.Context, req *JobResultRequest) error {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/jobs/result", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// doRequest performs an HTTP request
func (c *ControlPlaneClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ComputeHive-Agent/1.0")

	// Add authentication if available
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	// Log request
	c.logger.Debug("Making request to control plane",
		zap.String("method", method),
		zap.String("url", url))

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// handleErrorResponse handles error responses from the API
func (c *ControlPlaneClient) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}

	var apiError AgentError
	if err := json.Unmarshal(body, &apiError); err != nil {
		// If we can't parse as AgentError, return generic error
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return &apiError
}

// DownloadJobArtifacts downloads job artifacts from the control plane
func (c *ControlPlaneClient) DownloadJobArtifacts(ctx context.Context, jobID string, artifactPath string) ([]byte, error) {
	url := fmt.Sprintf("/api/v1/jobs/%s/artifacts/%s", jobID, artifactPath)
	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact data: %w", err)
	}

	return data, nil
}

// UploadJobResults uploads job results to the control plane
func (c *ControlPlaneClient) UploadJobResults(ctx context.Context, jobID string, results []byte) error {
	url := fmt.Sprintf("/api/v1/jobs/%s/results", jobID)
	
	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+url, bytes.NewReader(results))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(results)))
	
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// GetMetrics retrieves agent metrics
func (c *ControlPlaneClient) GetMetrics(ctx context.Context, agentID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("/api/v1/agents/%s/metrics", agentID)
	resp, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var metrics map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode metrics: %w", err)
	}

	return metrics, nil
} 