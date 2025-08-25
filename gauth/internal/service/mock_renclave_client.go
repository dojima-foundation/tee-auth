package service

import (
	"context"
	"fmt"
)

// MockRenclaveClient is a mock implementation of the enclave client for testing
type MockRenclaveClient struct{}

// NewMockRenclaveClient creates a new mock enclave client
func NewMockRenclaveClient() *MockRenclaveClient {
	return &MockRenclaveClient{}
}

// GenerateSeed returns a mock seed generation response
func (m *MockRenclaveClient) GenerateSeed(ctx context.Context, strength int, passphrase *string) (*GenerateSeedResponse, error) {
	return &GenerateSeedResponse{
		SeedPhrase: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		Entropy:    "00000000000000000000000000000000",
		Strength:   strength,
		WordCount:  strength / 32 * 3, // 12 words for 128-bit, 24 words for 256-bit
	}, nil
}

// ValidateSeed returns a mock seed validation response
func (m *MockRenclaveClient) ValidateSeed(ctx context.Context, seedPhrase string) (*ValidateSeedResponse, error) {
	// Simple validation - consider valid if it contains "abandon"
	isValid := len(seedPhrase) > 0 && len(seedPhrase) < 1000 // Basic sanity check
	return &ValidateSeedResponse{
		IsValid:   isValid,
		Strength:  256,
		WordCount: 24,
		Errors:    []string{},
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

// DeriveKey returns a mock key derivation response
func (m *MockRenclaveClient) DeriveKey(ctx context.Context, seedPhrase, path, curve string) (*DeriveKeyResponse, error) {
	return &DeriveKeyResponse{
		PrivateKey: "mock_private_key_1234567890abcdef",
		PublicKey:  "mock_public_key_1234567890abcdef",
		Address:    "mock_address_1234567890abcdef",
		Path:       path,
		Curve:      curve,
	}, nil
}

// DeriveAddress returns a mock address derivation response
func (m *MockRenclaveClient) DeriveAddress(ctx context.Context, seedPhrase, path, curve string) (*DeriveAddressResponse, error) {
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
