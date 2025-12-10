package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// Client represents a goCoax device HTTP client
type Client struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}

// NewClient creates a new goCoax device client
func NewClient(address, username, password string, timeout time.Duration) (*Client, error) {
	// Create cookie jar for session management (CSRF tokens, etc.)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	baseURL := fmt.Sprintf("http://%s", address)

	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
			Jar:     jar,
		},
		username: username,
		password: password,
	}

	// Initialize session by fetching the main page to get CSRF token
	if err := client.initSession(); err != nil {
		return nil, fmt.Errorf("failed to initialize session: %w", err)
	}

	return client, nil
}

// initSession initializes the session by fetching a page to get CSRF token
func (c *Client) initSession() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.httpClient.Timeout)
	defer cancel()

	// GET the phyRates page to get CSRF token cookie
	url := fmt.Sprintf("%s/phyRates.html", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create init request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)

	fmt.Printf("[DEBUG] Initializing session with GET %s\n", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initialize session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("session init failed with status %d", resp.StatusCode)
	}

	// Check if we got a CSRF token cookie
	cookies := c.httpClient.Jar.Cookies(req.URL)
	fmt.Printf("[DEBUG] Got %d cookies after init\n", len(cookies))
	for _, cookie := range cookies {
		fmt.Printf("[DEBUG] Cookie: %s=%s\n", cookie.Name, cookie.Value)
	}

	return nil
}

// apiResponse represents the JSON response structure from the device
type apiResponse struct {
	Data json.RawMessage `json:"data"`
}

// apiRequest represents the JSON request structure to the device
type apiRequest struct {
	Data interface{} `json:"data"`
}

// LocalInfo represents local device information
type LocalInfo struct {
	MyNodeID       int
	NCNodeID       int
	NCMocaVersion  int
	MocaNetVersion int
	NodeBitMask    int
	RawData        []int
}

// NetworkNodeInfo represents information about a network node
type NetworkNodeInfo struct {
	NodeID      int
	MocaVersion int
	RawData     []int
}

// FMRInfo represents Frame Management Request information
type FMRInfo struct {
	Data []uint32
}

// retryableError checks if an error is retryable
func retryableError(err error) bool {
	// Network errors, timeouts, and 5xx status codes are retryable
	if err == nil {
		return false
	}
	// Check for common transient errors
	errStr := err.Error()
	return contains(errStr, "connection refused") ||
		contains(errStr, "connection reset") ||
		contains(errStr, "timeout") ||
		contains(errStr, "temporary failure")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		 findSubstr(s, substr)))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// doRequestWithRetry performs an HTTP POST request with retry logic
func (c *Client) doRequestWithRetry(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms
			backoff := time.Duration(100*(1<<uint(attempt-1))) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		body, err := c.doRequest(ctx, endpoint, payload)
		if err == nil {
			return body, nil
		}

		lastErr = err

		// Don't retry if error is not retryable
		if !retryableError(err) {
			break
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// doRequest performs an HTTP POST request to the device API
func (c *Client) doRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	// Marshal payload to JSON
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug logging
	fmt.Printf("[DEBUG] Request to %s: %s\n", url, string(reqBody))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - device expects form-encoded, not JSON
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html, */*")

	// Debug: Log headers
	fmt.Printf("[DEBUG] Content-Type: %s\n", req.Header.Get("Content-Type"))

	// Set basic auth
	req.SetBasicAuth(c.username, c.password)
	fmt.Printf("[DEBUG] Using Basic Auth with username: %s\n", c.username)

	// Extract CSRF token from cookies if present
	for _, cookie := range c.httpClient.Jar.Cookies(req.URL) {
		if cookie.Name == "XSRF-TOKEN" || cookie.Name == "csrf_token" {
			req.Header.Set("X-CSRF-TOKEN", cookie.Value)
			break
		}
	}

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[DEBUG] Error response: status=%d, body=%s\n", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Debug logging
	fmt.Printf("[DEBUG] Response (first 200 chars): %s\n", string(body[:min(200, len(body))]))

	return body, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetLocalInfo retrieves local device information (endpoint 0x15)
func (c *Client) GetLocalInfo(ctx context.Context) (*LocalInfo, error) {
	// The endpoint expects {"data":[]} format
	payload := map[string]interface{}{
		"data": []interface{}{},
	}

	body, err := c.doRequestWithRetry(ctx, "/ms/0/0x15", payload)
	if err != nil {
		return nil, fmt.Errorf("GetLocalInfo request failed: %w", err)
	}

	// Parse response
	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse data array - device returns hex strings like "0x00000001"
	var dataStrings []string
	if err := json.Unmarshal(apiResp.Data, &dataStrings); err != nil {
		return nil, fmt.Errorf("failed to parse data array: %w", err)
	}

	// Convert hex strings to integers
	data := make([]int, len(dataStrings))
	for i, str := range dataStrings {
		var val int64
		_, err := fmt.Sscanf(str, "0x%x", &val)
		if err != nil {
			return nil, fmt.Errorf("failed to parse hex value %s: %w", str, err)
		}
		data[i] = int(val)
	}

	// Validate data length (should have at least 13 elements based on JavaScript)
	if len(data) < 13 {
		return nil, fmt.Errorf("insufficient data elements: got %d, expected at least 13", len(data))
	}

	// Parse according to JavaScript: LocalInfo[0]=myNodeID, [1]=NCNodeID, [11]=mocaNetVer, [12]=nodeBitMask
	localInfo := &LocalInfo{
		MyNodeID:       data[0],
		NCNodeID:       data[1] & 0xFF, // NC node ID is in lower byte
		MocaNetVersion: data[11],
		NodeBitMask:    data[12],
		RawData:        data,
	}

	// Get NC MoCA version from the NC node ID (will be fetched separately)
	// For now, we store the network version

	return localInfo, nil
}

// GetNetworkNodeInfo retrieves information about a specific network node (endpoint 0x16)
func (c *Client) GetNetworkNodeInfo(ctx context.Context, nodeID int) (*NetworkNodeInfo, error) {
	// The endpoint expects the node ID as data array: {"data":[nodeID]}
	payload := map[string]interface{}{
		"data": []interface{}{nodeID},
	}

	body, err := c.doRequestWithRetry(ctx, "/ms/0/0x16", payload)
	if err != nil {
		return nil, fmt.Errorf("GetNetworkNodeInfo request failed: %w", err)
	}

	// Parse response
	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse data array - device returns hex strings
	var dataStrings []string
	if err := json.Unmarshal(apiResp.Data, &dataStrings); err != nil {
		return nil, fmt.Errorf("failed to parse data array: %w", err)
	}

	// Convert hex strings to integers
	data := make([]int, len(dataStrings))
	for i, str := range dataStrings {
		var val int64
		_, err := fmt.Sscanf(str, "0x%x", &val)
		if err != nil {
			return nil, fmt.Errorf("failed to parse hex value %s: %w", str, err)
		}
		data[i] = int(val)
	}

	// Validate data length (should have at least 5 elements based on JavaScript usage)
	if len(data) < 5 {
		return nil, fmt.Errorf("insufficient data elements: got %d, expected at least 5", len(data))
	}

	// Parse according to JavaScript: netInfo[nodeId][4] contains MoCA version
	nodeInfo := &NetworkNodeInfo{
		NodeID:      nodeID,
		MocaVersion: data[4] & 0xFF, // MoCA version is in lower byte
		RawData:     data,
	}

	return nodeInfo, nil
}

// GetFMRInfo retrieves Frame Management Request information (endpoint 0x1D)
func (c *Client) GetFMRInfo(ctx context.Context, nodeMask, version int) (*FMRInfo, error) {
	// The endpoint expects both values in data array: {"data":[nodeMask, version]}
	payload := map[string]interface{}{
		"data": []interface{}{nodeMask, version},
	}

	body, err := c.doRequestWithRetry(ctx, "/ms/0/0x1D", payload)
	if err != nil {
		return nil, fmt.Errorf("GetFMRInfo request failed: %w", err)
	}

	// Parse response
	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse data array - device returns hex strings
	var dataStrings []string
	if err := json.Unmarshal(apiResp.Data, &dataStrings); err != nil {
		return nil, fmt.Errorf("failed to parse data array: %w", err)
	}

	// Convert hex strings to uint32
	data := make([]uint32, len(dataStrings))
	for i, str := range dataStrings {
		var val uint64
		_, err := fmt.Sscanf(str, "0x%x", &val)
		if err != nil {
			return nil, fmt.Errorf("failed to parse hex value %s: %w", str, err)
		}
		data[i] = uint32(val)
	}

	fmrInfo := &FMRInfo{
		Data: data,
	}

	return fmrInfo, nil
}

// Close closes the HTTP client and releases resources
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
