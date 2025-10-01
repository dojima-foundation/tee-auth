package service

import "context"

// EnclaveClientInterface defines the interface for enclave client operations
type EnclaveClientInterface interface {
	GenerateSeed(ctx context.Context, strength int, passphrase *string) (*GenerateSeedResponse, error)
	ValidateSeed(ctx context.Context, seedPhrase string, encryptedEntropy *string) (*ValidateSeedResponse, error)
	GetInfo(ctx context.Context) (*InfoResponse, error)
	GetEnclaveInfo(ctx context.Context) (*EnclaveInfoResponse, error)
	Health(ctx context.Context) error
	DeriveKey(ctx context.Context, encryptedSeedPhrase, path, curve string) (*DeriveKeyResponse, error)
	DeriveAddress(ctx context.Context, encryptedSeedPhrase, path, curve string) (*DeriveAddressResponse, error)
	GetNetworkStatus(ctx context.Context) (*NetworkStatusResponse, error)
	TestNetworkConnectivity(ctx context.Context, targetHost string, targetPort int, timeoutSeconds int) (*NetworkTestResponse, error)
}
