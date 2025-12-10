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

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
			Jar:     jar,
		},
		username: username,
		password: password,
	}, nil
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

// doRequest performs an HTTP POST request to the device API
func (c *Client) doRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	// Marshal payload to JSON
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set basic auth
	req.SetBasicAuth(c.username, c.password)

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
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetLocalInfo retrieves local device information (endpoint 0x15)
func (c *Client) GetLocalInfo(ctx context.Context) (*LocalInfo, error) {
	// The endpoint expects an empty data array
	payload := apiRequest{Data: []interface{}{}}

	body, err := c.doRequest(ctx, "/ms/0/0x15", payload)
	if err != nil {
		return nil, fmt.Errorf("GetLocalInfo request failed: %w", err)
	}

	// Parse response
	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse data array
	var data []int
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse data array: %w", err)
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
	// The endpoint expects the node ID as data parameter
	payload := apiRequest{Data: nodeID}

	body, err := c.doRequest(ctx, "/ms/0/0x16", payload)
	if err != nil {
		return nil, fmt.Errorf("GetNetworkNodeInfo request failed: %w", err)
	}

	// Parse response
	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse data array
	var data []int
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse data array: %w", err)
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
	// The endpoint expects node mask as 'data' and version as 'data2'
	// Since we're sending JSON, we need to structure this appropriately
	// Based on the HTML, it seems to use form data with multiple data fields

	// Create a custom payload structure
	type fmrRequest struct {
		Data  int `json:"data"`
		Data2 int `json:"data2"`
	}

	payload := fmrRequest{
		Data:  nodeMask,
		Data2: version,
	}

	body, err := c.doRequest(ctx, "/ms/0/0x1D", payload)
	if err != nil {
		return nil, fmt.Errorf("GetFMRInfo request failed: %w", err)
	}

	// Parse response
	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse data array as uint32 values
	var data []uint32
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		// Try parsing as []int and convert
		var intData []int
		if err2 := json.Unmarshal(apiResp.Data, &intData); err2 != nil {
			return nil, fmt.Errorf("failed to parse data array: %w (also tried int: %w)", err, err2)
		}
		// Convert int to uint32
		data = make([]uint32, len(intData))
		for i, v := range intData {
			data[i] = uint32(v)
		}
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
