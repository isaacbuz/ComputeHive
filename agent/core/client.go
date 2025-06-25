package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the control plane
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a new control plane client
func NewClient(config *Config) (*Client, error) {
	return &Client{
		baseURL: config.ControlPlaneURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: config.Token,
	}, nil
}

// Register registers the agent with the control plane
func (c *Client) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	var resp RegisterResponse
	err := c.doRequest(ctx, "POST", "/api/v1/agents/register", req, &resp)
	if err != nil {
		return nil, err
	}
	
	// Update client token
	c.token = resp.Token
	
	return &resp, nil
}

// Deregister removes the agent from the control plane
func (c *Client) Deregister(ctx context.Context, agentID string) error {
	endpoint := fmt.Sprintf("/api/v1/agents/%s/deregister", agentID)
	return c.doRequest(ctx, "POST", endpoint, nil, nil)
}

// SendHeartbeat sends a heartbeat to the control plane
func (c *Client) SendHeartbeat(ctx context.Context, heartbeat *Heartbeat) error {
	return c.doRequest(ctx, "POST", "/api/v1/agents/heartbeat", heartbeat, nil)
}

// GetJobs retrieves available jobs for the agent
func (c *Client) GetJobs(ctx context.Context, agentID string) ([]*Job, error) {
	endpoint := fmt.Sprintf("/api/v1/agents/%s/jobs", agentID)
	var jobs []*Job
	err := c.doRequest(ctx, "GET", endpoint, nil, &jobs)
	return jobs, err
}

// ReportJobResult reports the result of a job execution
func (c *Client) ReportJobResult(ctx context.Context, result *JobResult) error {
	endpoint := fmt.Sprintf("/api/v1/jobs/%s/result", result.JobID)
	return c.doRequest(ctx, "POST", endpoint, result, nil)
}

// ReportMetrics sends metrics to the control plane
func (c *Client) ReportMetrics(ctx context.Context, metrics *MetricsReport) error {
	return c.doRequest(ctx, "POST", "/api/v1/agents/metrics", metrics, nil)
}

// doRequest performs an HTTP request
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body, result interface{}) error {
	url := c.baseURL + endpoint
	
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("ComputeHive-Agent/%s", Version))
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	
	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	// Decode response if needed
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	
	return nil
}

// UploadArtifact uploads a job artifact
func (c *Client) UploadArtifact(ctx context.Context, jobID string, artifact *JobArtifact, data io.Reader) error {
	endpoint := fmt.Sprintf("/api/v1/jobs/%s/artifacts", jobID)
	
	// In a real implementation, this would use multipart/form-data
	// For now, we'll use a simplified approach
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, data)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-Artifact-Name", artifact.Name)
	req.Header.Set("X-Artifact-Size", fmt.Sprintf("%d", artifact.Size))
	req.Header.Set("Content-Type", artifact.MimeType)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// DownloadJobData downloads input data for a job
func (c *Client) DownloadJobData(ctx context.Context, jobID string, dest io.Writer) error {
	endpoint := fmt.Sprintf("/api/v1/jobs/%s/data", jobID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+endpoint, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	_, err = io.Copy(dest, resp.Body)
	return err
} 