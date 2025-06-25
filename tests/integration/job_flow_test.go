package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestConfig holds test environment configuration
type TestConfig struct {
	APIGatewayURL string
	AuthToken     string
	TestTimeout   time.Duration
}

// JobFlowTestSuite tests the complete job submission and execution flow
type JobFlowTestSuite struct {
	suite.Suite
	config     TestConfig
	httpClient *http.Client
	userID     string
	authToken  string
}

// SetupSuite runs once before all tests
func (s *JobFlowTestSuite) SetupSuite() {
	// Initialize test configuration
	s.config = TestConfig{
		APIGatewayURL: getEnvOrDefault("TEST_API_URL", "http://localhost:8000"),
		TestTimeout:   5 * time.Minute,
	}
	
	s.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Register a test user and get auth token
	s.registerTestUser()
}

// TearDownSuite runs once after all tests
func (s *JobFlowTestSuite) TearDownSuite() {
	// Clean up test data
	s.cleanupTestData()
}

// Test complete job flow from submission to completion
func (s *JobFlowTestSuite) TestCompleteJobFlow() {
	// Step 1: Submit a job
	jobID := s.submitTestJob()
	
	// Step 2: Verify job is created
	job := s.getJob(jobID)
	assert.Equal(s.T(), "pending", job["status"])
	assert.Equal(s.T(), s.userID, job["user_id"])
	
	// Step 3: Wait for job to be scheduled
	s.waitForJobStatus(jobID, "scheduled", 30*time.Second)
	
	// Step 4: Verify job is assigned to an agent
	job = s.getJob(jobID)
	assert.NotEmpty(s.T(), job["assigned_agent_id"])
	
	// Step 5: Simulate job execution (in real test, agent would update status)
	// For integration test, we'll wait for status changes
	s.waitForJobStatus(jobID, "running", 60*time.Second)
	
	// Step 6: Wait for job completion
	finalJob := s.waitForJobCompletion(jobID, 3*time.Minute)
	
	// Step 7: Verify final job state
	assert.Equal(s.T(), "completed", finalJob["status"])
	assert.NotNil(s.T(), finalJob["completed_at"])
	assert.Greater(s.T(), finalJob["actual_cost"].(float64), 0.0)
}

// Test job cancellation
func (s *JobFlowTestSuite) TestJobCancellation() {
	// Submit a job
	jobID := s.submitTestJob()
	
	// Wait for it to be scheduled
	s.waitForJobStatus(jobID, "scheduled", 30*time.Second)
	
	// Cancel the job
	s.cancelJob(jobID)
	
	// Verify job is cancelled
	job := s.getJob(jobID)
	assert.Equal(s.T(), "cancelled", job["status"])
}

// Test job with specific resource requirements
func (s *JobFlowTestSuite) TestJobWithGPURequirements() {
	// Submit a GPU job
	jobData := map[string]interface{}{
		"type":     "docker",
		"priority": 8,
		"requirements": map[string]interface{}{
			"cpu_cores":    4,
			"memory_mb":    8192,
			"gpu_count":    2,
			"gpu_type":     "NVIDIA V100",
			"storage_mb":   10240,
			"network_mbps": 1000,
		},
		"payload": map[string]interface{}{
			"image":   "tensorflow/tensorflow:latest-gpu",
			"command": []string{"python", "-c", "import tensorflow as tf; print(tf.config.list_physical_devices('GPU'))"},
		},
		"timeout":     3600,
		"max_retries": 3,
	}
	
	resp := s.makeAPIRequest("POST", "/api/v1/scheduler/jobs", jobData)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var job map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&job)
	require.NoError(s.T(), err)
	
	jobID := job["id"].(string)
	
	// Verify GPU requirements are properly set
	requirements := job["requirements"].(map[string]interface{})
	assert.Equal(s.T(), float64(2), requirements["gpu_count"])
	assert.Equal(s.T(), "NVIDIA V100", requirements["gpu_type"])
}

// Test marketplace offer creation and matching
func (s *JobFlowTestSuite) TestMarketplaceFlow() {
	// Step 1: Create a compute offer
	offerData := map[string]interface{}{
		"resources": map[string]interface{}{
			"cpu": map[string]interface{}{
				"cores":     16,
				"model":     "Intel Xeon Gold 6248",
				"frequency": "2.5GHz",
			},
			"memory": map[string]interface{}{
				"total_mb": 65536,
				"type":     "DDR4",
				"speed":    "2933MHz",
			},
			"storage": map[string]interface{}{
				"total_mb": 1024000,
				"type":     "nvme",
				"iops":     100000,
			},
			"network": map[string]interface{}{
				"bandwidth_mbps": 10000,
				"type":           "dedicated",
			},
		},
		"price_per_hour": map[string]interface{}{
			"cpu":     "0.10",
			"memory":  "0.01",
			"storage": "0.001",
			"gpu":     "0.50",
		},
		"min_duration": 3600,
		"max_duration": 86400,
		"availability": map[string]interface{}{
			"start_time": time.Now().Format(time.RFC3339),
			"end_time":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		},
		"location": "us-east-1",
		"features": []string{"docker", "kubernetes", "sgx"},
		"sla_guarantees": map[string]interface{}{
			"uptime_percentage":     99.9,
			"max_response_time_ms": 100,
			"support_level":        "priority",
		},
	}
	
	// Create offer
	resp := s.makeAPIRequest("POST", "/api/v1/marketplace/offers", offerData)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var offer map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&offer)
	require.NoError(s.T(), err)
	
	// Step 2: Create a bid that matches the offer
	bidData := map[string]interface{}{
		"requirements": map[string]interface{}{
			"min_cpu_cores":   8,
			"min_memory_mb":   32768,
			"min_storage_mb":  100000,
			"min_network_mbps": 1000,
			"min_gpu_count":   0,
		},
		"max_price_per_hour": "5.00",
		"duration":           7200, // 2 hours
		"start_time":         time.Now().Add(10 * time.Minute).Format(time.RFC3339),
		"flexibility":        600, // 10 minutes
		"location":           "us-east-1",
	}
	
	resp = s.makeAPIRequest("POST", "/api/v1/marketplace/bids", bidData)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var bid map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&bid)
	require.NoError(s.T(), err)
	
	// Step 3: Wait for matching
	time.Sleep(15 * time.Second) // Wait for matching engine
	
	// Step 4: Check if match was created
	resp = s.makeAPIRequest("GET", "/api/v1/marketplace/matches", nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var matches []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&matches)
	require.NoError(s.T(), err)
	
	// Verify match exists
	assert.Greater(s.T(), len(matches), 0)
	
	// Find our match
	var ourMatch map[string]interface{}
	for _, match := range matches {
		if match["bid_id"] == bid["id"] {
			ourMatch = match
			break
		}
	}
	
	require.NotNil(s.T(), ourMatch)
	assert.Equal(s.T(), offer["id"], ourMatch["offer_id"])
	assert.Equal(s.T(), "pending", ourMatch["status"])
}

// Test rate limiting
func (s *JobFlowTestSuite) TestRateLimiting() {
	// Make many rapid requests
	results := make(chan int, 150)
	
	for i := 0; i < 150; i++ {
		go func() {
			resp := s.makeAPIRequest("GET", "/api/v1/scheduler/jobs", nil)
			results <- resp.StatusCode
		}()
	}
	
	// Collect results
	rateLimited := 0
	successful := 0
	
	for i := 0; i < 150; i++ {
		status := <-results
		if status == http.StatusTooManyRequests {
			rateLimited++
		} else if status == http.StatusOK {
			successful++
		}
	}
	
	// Should have some rate limited requests
	assert.Greater(s.T(), rateLimited, 0, "Expected some requests to be rate limited")
	assert.Greater(s.T(), successful, 0, "Expected some requests to succeed")
}

// Test WebSocket connection for real-time updates
func (s *JobFlowTestSuite) TestWebSocketUpdates() {
	// This is a placeholder for WebSocket testing
	// In a real implementation, you would:
	// 1. Connect to WebSocket endpoint
	// 2. Subscribe to job updates
	// 3. Submit a job
	// 4. Verify real-time updates are received
	s.T().Skip("WebSocket testing requires additional setup")
}

// Helper methods

func (s *JobFlowTestSuite) registerTestUser() {
	userData := map[string]interface{}{
		"email":    fmt.Sprintf("test-%d@computehive.io", time.Now().UnixNano()),
		"username": fmt.Sprintf("testuser%d", time.Now().UnixNano()),
		"password": "TestPassword123!",
	}
	
	resp, err := s.httpClient.Post(
		s.config.APIGatewayURL+"/api/v1/auth/register",
		"application/json",
		bytes.NewBuffer(mustMarshal(userData)),
	)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var authResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	require.NoError(s.T(), err)
	
	s.authToken = authResp["access_token"].(string)
	s.userID = authResp["user_id"].(string)
}

func (s *JobFlowTestSuite) submitTestJob() string {
	jobData := map[string]interface{}{
		"type":     "docker",
		"priority": 5,
		"requirements": map[string]interface{}{
			"cpu_cores":    2,
			"memory_mb":    2048,
			"gpu_count":    0,
			"storage_mb":   10240,
			"network_mbps": 100,
		},
		"payload": map[string]interface{}{
			"image":   "alpine:latest",
			"command": []string{"echo", "Hello from ComputeHive!"},
		},
		"timeout":     300,
		"max_retries": 3,
	}
	
	resp := s.makeAPIRequest("POST", "/api/v1/scheduler/jobs", jobData)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var job map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&job)
	require.NoError(s.T(), err)
	
	return job["id"].(string)
}

func (s *JobFlowTestSuite) getJob(jobID string) map[string]interface{} {
	resp := s.makeAPIRequest("GET", fmt.Sprintf("/api/v1/scheduler/jobs/%s", jobID), nil)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	
	var job map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&job)
	require.NoError(s.T(), err)
	
	return job
}

func (s *JobFlowTestSuite) cancelJob(jobID string) {
	resp := s.makeAPIRequest("POST", fmt.Sprintf("/api/v1/scheduler/jobs/%s/cancel", jobID), nil)
	require.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
}

func (s *JobFlowTestSuite) waitForJobStatus(jobID, expectedStatus string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			s.T().Fatalf("Timeout waiting for job %s to reach status %s", jobID, expectedStatus)
		case <-ticker.C:
			job := s.getJob(jobID)
			if job["status"] == expectedStatus {
				return
			}
		}
	}
}

func (s *JobFlowTestSuite) waitForJobCompletion(jobID string, timeout time.Duration) map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			s.T().Fatalf("Timeout waiting for job %s to complete", jobID)
		case <-ticker.C:
			job := s.getJob(jobID)
			status := job["status"].(string)
			if status == "completed" || status == "failed" || status == "cancelled" {
				return job
			}
		}
	}
}

func (s *JobFlowTestSuite) makeAPIRequest(method, path string, body interface{}) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(mustMarshal(body))
	}
	
	req, err := http.NewRequest(method, s.config.APIGatewayURL+path, bodyReader)
	require.NoError(s.T(), err)
	
	req.Header.Set("Content-Type", "application/json")
	if s.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.authToken)
	}
	
	resp, err := s.httpClient.Do(req)
	require.NoError(s.T(), err)
	
	return resp
}

func (s *JobFlowTestSuite) cleanupTestData() {
	// In a real implementation, clean up test jobs, offers, etc.
	// This could involve calling cleanup endpoints or directly accessing the database
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// TestJobFlow runs the test suite
func TestJobFlow(t *testing.T) {
	suite.Run(t, new(JobFlowTestSuite))
}

// Additional test cases for error scenarios

func (s *JobFlowTestSuite) TestJobSubmissionValidation() {
	testCases := []struct {
		name        string
		jobData     map[string]interface{}
		expectedErr string
	}{
		{
			name: "Missing CPU cores",
			jobData: map[string]interface{}{
				"type": "docker",
				"requirements": map[string]interface{}{
					"memory_mb": 1024,
				},
			},
			expectedErr: "CPU cores must be positive",
		},
		{
			name: "Invalid job type",
			jobData: map[string]interface{}{
				"type": "invalid",
				"requirements": map[string]interface{}{
					"cpu_cores": 1,
					"memory_mb": 1024,
				},
			},
			expectedErr: "Invalid job type",
		},
		{
			name: "Negative priority",
			jobData: map[string]interface{}{
				"type":     "docker",
				"priority": -1,
				"requirements": map[string]interface{}{
					"cpu_cores": 1,
					"memory_mb": 1024,
				},
			},
			expectedErr: "Priority must be between 0 and 10",
		},
	}
	
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			resp := s.makeAPIRequest("POST", "/api/v1/scheduler/jobs", tc.jobData)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			
			var errResp map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&errResp)
			require.NoError(t, err)
			
			assert.Contains(t, errResp["message"].(string), tc.expectedErr)
		})
	}
}

func (s *JobFlowTestSuite) TestConcurrentJobSubmissions() {
	numJobs := 10
	results := make(chan string, numJobs)
	errors := make(chan error, numJobs)
	
	// Submit multiple jobs concurrently
	for i := 0; i < numJobs; i++ {
		go func(index int) {
			jobData := map[string]interface{}{
				"type":     "docker",
				"priority": 5,
				"requirements": map[string]interface{}{
					"cpu_cores": 1,
					"memory_mb": 512,
				},
				"payload": map[string]interface{}{
					"image":   "alpine:latest",
					"command": []string{"echo", fmt.Sprintf("Job %d", index)},
				},
			}
			
			resp := s.makeAPIRequest("POST", "/api/v1/scheduler/jobs", jobData)
			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("job %d failed with status %d", index, resp.StatusCode)
				return
			}
			
			var job map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
				errors <- err
				return
			}
			
			results <- job["id"].(string)
		}(i)
	}
	
	// Collect results
	jobIDs := make([]string, 0, numJobs)
	for i := 0; i < numJobs; i++ {
		select {
		case jobID := <-results:
			jobIDs = append(jobIDs, jobID)
		case err := <-errors:
			s.T().Errorf("Concurrent job submission failed: %v", err)
		}
	}
	
	// Verify all jobs were created
	assert.Equal(s.T(), numJobs, len(jobIDs))
	
	// Verify all job IDs are unique
	uniqueIDs := make(map[string]bool)
	for _, id := range jobIDs {
		uniqueIDs[id] = true
	}
	assert.Equal(s.T(), numJobs, len(uniqueIDs))
} 