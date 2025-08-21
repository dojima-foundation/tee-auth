package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RenclaveClient handles communication with the renclave-v2 service
type RenclaveClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRenclaveClient creates a new RenclaveClient instance
func NewRenclaveClient(baseURL string, timeout time.Duration) *RenclaveClient {
	return &RenclaveClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Request and response types for renclave-v2 communication

type GenerateSeedRequest struct {
	Strength   int     `json:"strength"`
	Passphrase *string `json:"passphrase,omitempty"`
}

type GenerateSeedResponse struct {
	SeedPhrase string `json:"seed_phrase"`
	Entropy    string `json:"entropy"`
	Strength   int    `json:"strength"`
	WordCount  int    `json:"word_count"`
}

type ValidateSeedRequest struct {
	SeedPhrase string `json:"seed_phrase"`
}

type ValidateSeedResponse struct {
	IsValid   bool     `json:"is_valid"`
	Strength  int      `json:"strength"`
	WordCount int      `json:"word_count"`
	Errors    []string `json:"errors"`
}

type InfoResponse struct {
	Version      string   `json:"version"`
	EnclaveID    string   `json:"enclave_id"`
	Capabilities []string `json:"capabilities"`
	Healthy      bool     `json:"healthy"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

// GenerateSeed requests seed generation from renclave-v2
func (c *RenclaveClient) GenerateSeed(ctx context.Context, strength int, passphrase *string) (*GenerateSeedResponse, error) {
	req := GenerateSeedRequest{
		Strength:   strength,
		Passphrase: passphrase,
	}

	var resp GenerateSeedResponse
	if err := c.makeRequest(ctx, "POST", "/generate-seed", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	return &resp, nil
}

// ValidateSeed requests seed validation from renclave-v2
func (c *RenclaveClient) ValidateSeed(ctx context.Context, seedPhrase string) (*ValidateSeedResponse, error) {
	req := ValidateSeedRequest{
		SeedPhrase: seedPhrase,
	}

	var resp ValidateSeedResponse
	if err := c.makeRequest(ctx, "POST", "/validate-seed", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to validate seed: %w", err)
	}

	return &resp, nil
}

// GetInfo requests information from renclave-v2
func (c *RenclaveClient) GetInfo(ctx context.Context) (*InfoResponse, error) {
	var resp InfoResponse
	if err := c.makeRequest(ctx, "GET", "/info", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	return &resp, nil
}

// Health checks the health of renclave-v2
func (c *RenclaveClient) Health(ctx context.Context) error {
	var resp HealthResponse
	if err := c.makeRequest(ctx, "GET", "/health", nil, &resp); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != "healthy" && resp.Status != "ok" {
		return fmt.Errorf("renclave is unhealthy: %s", resp.Status)
	}

	return nil
}

// makeRequest is a helper method to make HTTP requests to renclave-v2
func (c *RenclaveClient) makeRequest(ctx context.Context, method, path string, reqBody interface{}, respBody interface{}) error {
	url := c.baseURL + path

	var body *bytes.Buffer
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, body)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	return nil
}
