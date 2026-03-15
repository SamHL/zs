package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/SamHL/zs/internal/auth"
	"github.com/SamHL/zs/internal/config"
)

// Client is the Zoho Sprints API client
type Client struct {
	httpClient  *http.Client
	baseURL     string
	rateLimiter *RateLimiter
	debug       bool
}

// RateLimiter implements a simple rate limiter (100 requests per 2 minutes)
type RateLimiter struct {
	mu          sync.Mutex
	requests    []time.Time
	maxRequests int
	window      time.Duration
}

// APIError represents an error response from the API
type APIError struct {
	Code    string `json:"errorCode"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// APIResponse wraps all API responses
type APIResponse struct {
	Status       string          `json:"status"`
	ErrorCode    string          `json:"errorCode"`
	ErrorMessage string          `json:"message"`
	Data         json.RawMessage `json:"-"`
}

// NewClient creates a new API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: config.GetBaseURL(),
		rateLimiter: &RateLimiter{
			maxRequests: 100,
			window:      2 * time.Minute,
		},
		debug: false,
	}
}

// SetDebug enables or disables debug mode
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

// Wait blocks until a request can be made within rate limits
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Remove old requests outside the window
	var validRequests []time.Time
	for _, t := range rl.requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	rl.requests = validRequests

	// If at limit, wait until oldest request expires
	if len(rl.requests) >= rl.maxRequests {
		waitTime := rl.requests[0].Add(rl.window).Sub(now)
		if waitTime > 0 {
			rl.mu.Unlock()
			time.Sleep(waitTime)
			rl.mu.Lock()
		}
		// Remove the oldest request
		rl.requests = rl.requests[1:]
	}

	// Record this request
	rl.requests = append(rl.requests, now)
}

// doRequest performs an HTTP request with auth and rate limiting
func (c *Client) doRequest(method, path string, body io.Reader, params url.Values) ([]byte, error) {
	// Ensure valid token
	if err := auth.EnsureValidToken(); err != nil {
		return nil, err
	}

	// Apply rate limiting
	c.rateLimiter.Wait()

	// Build URL
	reqURL := fmt.Sprintf("%s/zsapi%s", c.baseURL, path)
	if params != nil && len(params) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	// Create request
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	cfg := config.Get()
	req.Header.Set("Authorization", fmt.Sprintf("Zoho-oauthtoken %s", cfg.Auth.AccessToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if c.debug {
		fmt.Printf("DEBUG: %s %s\n", method, reqURL)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.debug {
		fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
		if len(respBody) < 1000 {
			fmt.Printf("DEBUG: Response body: %s\n", string(respBody))
		}
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Code != "" {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get performs a GET request
func (c *Client) Get(path string, params url.Values) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, nil, params)
}

// Post performs a POST request
func (c *Client) Post(path string, data url.Values) ([]byte, error) {
	var body io.Reader
	if data != nil {
		body = strings.NewReader(data.Encode())
	}
	return c.doRequest(http.MethodPost, path, body, nil)
}

// Put performs a PUT request
func (c *Client) Put(path string, data url.Values) ([]byte, error) {
	var body io.Reader
	if data != nil {
		body = strings.NewReader(data.Encode())
	}
	return c.doRequest(http.MethodPut, path, body, nil)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string) ([]byte, error) {
	return c.doRequest(http.MethodDelete, path, nil, nil)
}

// GetJSON performs a GET request and unmarshals the response
func (c *Client) GetJSON(path string, params url.Values, result interface{}) error {
	data, err := c.Get(path, params)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

// PostJSON performs a POST request and unmarshals the response
func (c *Client) PostJSON(path string, data url.Values, result interface{}) error {
	respData, err := c.Post(path, data)
	if err != nil {
		return err
	}
	return json.Unmarshal(respData, result)
}

// PutJSON performs a PUT request and unmarshals the response
func (c *Client) PutJSON(path string, data url.Values, result interface{}) error {
	respData, err := c.Put(path, data)
	if err != nil {
		return err
	}
	return json.Unmarshal(respData, result)
}

// GetTeamPath returns the API path prefix for a team
func (c *Client) GetTeamPath(teamID string) string {
	if teamID == "" {
		teamID = config.Get().Defaults.TeamID
	}
	return fmt.Sprintf("/team/%s", teamID)
}

// GetProjectPath returns the API path prefix for a project
func (c *Client) GetProjectPath(teamID, projectID string) string {
	if teamID == "" {
		teamID = config.Get().Defaults.TeamID
	}
	if projectID == "" {
		projectID = config.Get().Defaults.ProjectID
	}
	return fmt.Sprintf("/team/%s/projects/%s", teamID, projectID)
}
