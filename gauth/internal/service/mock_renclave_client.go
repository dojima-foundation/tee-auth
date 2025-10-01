package service

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// MockRenclaveClient is a mock implementation of the enclave client for testing
type MockRenclaveClient struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

// NewMockRenclaveClient creates a new mock enclave client
func NewMockRenclaveClient() *MockRenclaveClient {
	return &MockRenclaveClient{
		tracer:     otel.Tracer("mock-renclave-client"),
		propagator: otel.GetTextMapPropagator(),
	}
}

// GenerateSeed returns a mock seed generation response
func (m *MockRenclaveClient) GenerateSeed(ctx context.Context, strength int, passphrase *string) (*GenerateSeedResponse, error) {
	return &GenerateSeedResponse{
		SeedPhrase: "encrypted_seed_hex_data_placeholder", // Mock encrypted seed data
		Entropy:    "00000000000000000000000000000000",
		Strength:   strength,
		WordCount:  strength / 32 * 3, // 12 words for 128-bit, 24 words for 256-bit
	}, nil
}

// ValidateSeed returns a mock seed validation response
func (m *MockRenclaveClient) ValidateSeed(ctx context.Context, seedPhrase string, encryptedEntropy *string) (*ValidateSeedResponse, error) {
	// Simple validation - consider valid if it contains "abandon"
	isValid := len(seedPhrase) > 0 && len(seedPhrase) < 1000 // Basic sanity check

	// Mock entropy validation if encrypted entropy is provided
	var entropyMatch *bool
	var derivedEntropy *string
	if encryptedEntropy != nil {
		match := true // Mock successful entropy match
		entropyMatch = &match
		derived := "mock_derived_entropy_1234567890abcdef"
		derivedEntropy = &derived
	}

	return &ValidateSeedResponse{
		IsValid:        isValid,
		Strength:       256,
		WordCount:      24,
		Errors:         []string{},
		EntropyMatch:   entropyMatch,
		DerivedEntropy: derivedEntropy,
	}, nil
}

// GetInfo returns mock enclave information
func (m *MockRenclaveClient) GetInfo(ctx context.Context) (*InfoResponse, error) {
	return &InfoResponse{
		Version:      "1.0.0-mock",
		EnclaveID:    "mock-enclave-id",
		Capabilities: []string{"seed_generation", "bip39_compliance"},
		Healthy:      true,
	}, nil
}

// Health returns a mock health check response
func (m *MockRenclaveClient) Health(ctx context.Context) error {
	return nil // Always healthy
}

// GetEnclaveInfo returns mock detailed enclave information
func (m *MockRenclaveClient) GetEnclaveInfo(ctx context.Context) (*EnclaveInfoResponse, error) {
	return &EnclaveInfoResponse{
		Version:      "1.0.0-mock",
		EnclaveID:    "mock-enclave-id",
		Capabilities: []string{"seed_generation", "bip39_compliance", "network_testing"},
		Healthy:      true,
		Status:       "active",
		Details: map[string]interface{}{
			"mock_field": "mock_value",
			"uptime":     "24h",
		},
	}, nil
}

// DeriveKey returns a mock key derivation response
func (m *MockRenclaveClient) DeriveKey(ctx context.Context, encryptedSeedPhrase, path, curve string) (*DeriveKeyResponse, error) {
	return &DeriveKeyResponse{
		PrivateKey: "mock_private_key_1234567890abcdef",
		PublicKey:  "mock_public_key_1234567890abcdef",
		Address:    "mock_address_1234567890abcdef",
		Path:       path,
		Curve:      curve,
	}, nil
}

// DeriveAddress returns a mock address derivation response
func (m *MockRenclaveClient) DeriveAddress(ctx context.Context, encryptedSeedPhrase, path, curve string) (*DeriveAddressResponse, error) {
	// Generate a deterministic but unique address based on the path
	// This ensures different paths get different addresses
	addressSuffix := ""
	for _, char := range path {
		if char >= '0' && char <= '9' {
			addressSuffix += string(char)
		}
	}
	if addressSuffix == "" {
		addressSuffix = "0"
	}

	// Create a mock address that varies based on the path
	address := fmt.Sprintf("mock_address_%s_%s_%s", curve, path, addressSuffix)

	return &DeriveAddressResponse{
		Address: address,
		Path:    path,
		Curve:   curve,
	}, nil
}

// GetNetworkStatus returns mock network status
func (m *MockRenclaveClient) GetNetworkStatus(ctx context.Context) (*NetworkStatusResponse, error) {
	return &NetworkStatusResponse{
		Status: "connected",
		Connectivity: map[string]bool{
			"internet": true,
			"dns":      true,
		},
		Configuration: map[string]interface{}{
			"timeout": 30,
			"retries": 3,
		},
	}, nil
}

// TestNetworkConnectivity returns mock network connectivity test result
func (m *MockRenclaveClient) TestNetworkConnectivity(ctx context.Context, targetHost string, targetPort int, timeoutSeconds int) (*NetworkTestResponse, error) {
	return &NetworkTestResponse{
		Success:      true,
		ResponseTime: 150.5,
		Error:        nil,
	}, nil
}
