package service

import "context"

// EnclaveClientInterface defines the interface for enclave client operations
type EnclaveClientInterface interface {
	GenerateSeed(ctx context.Context, strength int, passphrase *string) (*GenerateSeedResponse, error)
	ValidateSeed(ctx context.Context, seedPhrase string) (*ValidateSeedResponse, error)
	GetInfo(ctx context.Context) (*InfoResponse, error)
	Health(ctx context.Context) error
	DeriveKey(ctx context.Context, seedPhrase, path, curve string) (*DeriveKeyResponse, error)
	DeriveAddress(ctx context.Context, seedPhrase, path, curve string) (*DeriveAddressResponse, error)
}
