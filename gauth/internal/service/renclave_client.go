package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// RenclaveClient handles communication with the renclave-v2 service
type RenclaveClient struct {
	baseURL    string
	httpClient *http.Client
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

// NewRenclaveClient creates a new RenclaveClient instance
func NewRenclaveClient(baseURL string, timeout time.Duration) *RenclaveClient {
	// Create HTTP client with OpenTelemetry instrumentation
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)
			}),
		),
	}

	return &RenclaveClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		tracer:     otel.Tracer("renclave-client"),
		propagator: otel.GetTextMapPropagator(),
	}
}

// Request and response types for renclave-v2 communication

type GenerateSeedRequest struct {
	Strength   int     `json:"strength"`
	Passphrase *string `json:"passphrase,omitempty"`
}

type GenerateSeedResponse struct {
	SeedPhrase string `json:"seed_phrase"` // Now contains encrypted seed data (hex-encoded)
	Entropy    string `json:"entropy"`
	Strength   int    `json:"strength"`
	WordCount  int    `json:"word_count"`
}

type ValidateSeedRequest struct {
	SeedPhrase string `json:"seed_phrase"`
}

// RawValidateSeedResponse represents the actual response from renclave-v2
type RawValidateSeedResponse struct {
	Valid     bool `json:"valid"`
	WordCount int  `json:"word_count"`
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

// New request/response types for key derivation
type DeriveKeyRequest struct {
	SeedPhrase string `json:"seed_phrase"` // Now contains encrypted seed data (hex-encoded)
	Path       string `json:"path"`        // BIP32 derivation path (e.g., "m/44'/60'/0'/0/0")
	Curve      string `json:"curve"`       // CURVE_SECP256K1, CURVE_ED25519
}

type DeriveKeyResponse struct {
	PrivateKey string `json:"private_key"` // Hex-encoded private key
	PublicKey  string `json:"public_key"`  // Hex-encoded public key
	Address    string `json:"address"`     // Derived address
	Path       string `json:"path"`        // The derivation path used
	Curve      string `json:"curve"`       // The curve used
}

type DeriveAddressRequest struct {
	SeedPhrase string `json:"seed_phrase"` // Now contains encrypted seed data (hex-encoded)
	Path       string `json:"path"`        // BIP32 derivation path
	Curve      string `json:"curve"`       // CURVE_SECP256K1, CURVE_ED25519
}

type DeriveAddressResponse struct {
	Address string `json:"address"` // Derived address
	Path    string `json:"path"`    // The derivation path used
	Curve   string `json:"curve"`   // The curve used
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
// Handles both encrypted seeds and plaintext mnemonics
func (c *RenclaveClient) ValidateSeed(ctx context.Context, seedPhrase string) (*ValidateSeedResponse, error) {
	// Check if the seed phrase is encrypted (hex-encoded) or plaintext mnemonic
	// Encrypted seeds are typically long hex strings, while mnemonics are words separated by spaces
	isEncrypted := len(seedPhrase) > 100 && !strings.Contains(seedPhrase, " ")

	if isEncrypted {
		// For encrypted seeds, we cannot validate directly as they need to be decrypted first
		// Return a meaningful response indicating this limitation
		return &ValidateSeedResponse{
			IsValid:   false,
			Strength:  0,
			WordCount: 0,
			Errors:    []string{},
		}, fmt.Errorf("encrypted seed validation not supported: seed appears to be encrypted and requires decryption in renclave TEE before validation")
	}

	// For plaintext mnemonics, proceed with normal validation
	req := ValidateSeedRequest{
		SeedPhrase: seedPhrase,
	}

	var rawResp RawValidateSeedResponse
	if err := c.makeRequest(ctx, "POST", "/validate-seed", req, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to validate seed: %w", err)
	}

	// Map the raw response to the expected format
	resp := &ValidateSeedResponse{
		IsValid:   rawResp.Valid,
		Strength:  256, // Default strength for 24-word phrases
		WordCount: rawResp.WordCount,
		Errors:    []string{},
	}

	return resp, nil
}

// GetInfo requests information from renclave-v2
func (c *RenclaveClient) GetInfo(ctx context.Context) (*InfoResponse, error) {
	var resp InfoResponse
	if err := c.makeRequest(ctx, "GET", "/info", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	return &resp, nil
}

// DeriveKey derives a private key and public key from a seed phrase using BIP32
func (c *RenclaveClient) DeriveKey(ctx context.Context, seedPhrase, path, curve string) (*DeriveKeyResponse, error) {
	req := DeriveKeyRequest{
		SeedPhrase: seedPhrase,
		Path:       path,
		Curve:      curve,
	}

	var resp DeriveKeyResponse
	if err := c.makeRequest(ctx, "POST", "/derive-key", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	return &resp, nil
}

// DeriveAddress derives an address from a seed phrase using BIP32
func (c *RenclaveClient) DeriveAddress(ctx context.Context, seedPhrase, path, curve string) (*DeriveAddressResponse, error) {
	req := DeriveAddressRequest{
		SeedPhrase: seedPhrase,
		Path:       path,
		Curve:      curve,
	}

	var resp DeriveAddressResponse
	if err := c.makeRequest(ctx, "POST", "/derive-address", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}

	return &resp, nil
}

// Health checks the health of renclave-v2
func (c *RenclaveClient) Health(ctx context.Context) error {
	// Start a new span for the health check
	ctx, span := c.tracer.Start(ctx, "RenclaveClient.Health",
		trace.WithAttributes(
			attribute.String("rpc.system", "http"),
			attribute.String("rpc.service", "renclave"),
			attribute.String("rpc.method", "GET"),
			attribute.String("http.url", c.baseURL+"/health"),
		),
	)
	defer span.End()

	// For health check, we only care about the status code, not the response body
	// The renclave service returns 200 OK without a response body
	url := c.baseURL + "/health"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create health request")
		return fmt.Errorf("failed to create health request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Inject trace context into request headers
	c.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Health check failed")
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Add response details to span
	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
	)

	// For health check, any 2xx status code is considered healthy
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("health check failed with status %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// makeRequest is a helper method to make HTTP requests to renclave-v2
func (c *RenclaveClient) makeRequest(ctx context.Context, method, path string, reqBody interface{}, respBody interface{}) error {
	// Start a new span for the request
	spanName := fmt.Sprintf("RenclaveClient.%s", path)
	ctx, span := c.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("rpc.system", "http"),
			attribute.String("rpc.service", "renclave"),
			attribute.String("rpc.method", method),
			attribute.String("http.url", c.baseURL+path),
		),
	)
	defer span.End()

	url := c.baseURL + path

	var body *bytes.Buffer
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to marshal request body")
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewBuffer(jsonData)

		// Add request body details to span
		span.SetAttributes(
			attribute.String("http.request.body.size", fmt.Sprintf("%d", len(jsonData))),
		)
	}

	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, body)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Inject trace context into request headers
	c.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Add span attributes for the request
	span.SetAttributes(
		attribute.String("http.method", method),
		attribute.String("http.path", path),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to make request")
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Add response details to span
	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("request failed with status %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return err
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to decode response body")
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
