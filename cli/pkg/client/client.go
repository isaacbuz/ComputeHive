package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents the ComputeHive API client
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// New creates a new API client
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken updates the client's authentication token
func (c *Client) SetToken(token string) {
	c.Token = token
}

// do performs an HTTP request
func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ComputeHive-CLI/1.0")

	return c.HTTPClient.Do(req)
}

// Error represents an API error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// handleResponse processes API responses
func (c *Client) handleResponse(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr Error
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return fmt.Errorf("API error: %s", resp.Status)
		}
		return apiErr
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}

	return nil
}

// Authentication methods

// Login authenticates with email and password
func (c *Client) Login(email, password string) (string, error) {
	body := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, err := c.do("POST", "/auth/login", body)
	if err != nil {
		return "", err
	}

	var result struct {
		Token string `json:"token"`
		User  User   `json:"user"`
	}

	if err := c.handleResponse(resp, &result); err != nil {
		return "", err
	}

	return result.Token, nil
}

// VerifyToken verifies an authentication token
func (c *Client) VerifyToken(token string) error {
	c.Token = token
	resp, err := c.do("GET", "/auth/verify", nil)
	if err != nil {
		return err
	}
	return c.handleResponse(resp, nil)
}

// GetOAuthURL gets the OAuth authentication URL
func (c *Client) GetOAuthURL(provider string) (string, error) {
	resp, err := c.do("GET", fmt.Sprintf("/auth/oauth/%s", provider), nil)
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

// ExchangeOAuthCode exchanges an OAuth code for a token
func (c *Client) ExchangeOAuthCode(provider, code string) (string, error) {
	body := map[string]string{
		"code": code,
	}

	resp, err := c.do("POST", fmt.Sprintf("/auth/oauth/%s/callback", provider), body)
	if err != nil {
		return "", err
	}

	var result struct {
		Token string `json:"token"`
	}

	if err := c.handleResponse(resp, &result); err != nil {
		return "", err
	}

	return result.Token, nil
} 