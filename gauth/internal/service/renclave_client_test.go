package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenclaveClient_GenerateSeed(t *testing.T) {
	tests := []struct {
		name           string
		strength       int
		passphrase     *string
		serverResponse GenerateSeedResponse
		serverStatus   int
		expectError    bool
	}{
		{
			name:       "successful seed generation",
			strength:   256,
			passphrase: nil,
			serverResponse: GenerateSeedResponse{
				SeedPhrase: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
				Entropy:    "00000000000000000000000000000000",
				Strength:   256,
				WordCount:  24,
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:       "seed generation with passphrase",
			strength:   128,
			passphrase: stringPtr("test passphrase"),
			serverResponse: GenerateSeedResponse{
				SeedPhrase: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
				Entropy:    "00000000000000000000000000000000",
				Strength:   128,
				WordCount:  12,
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "server error",
			strength:       256,
			passphrase:     nil,
			serverResponse: GenerateSeedResponse{},
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "invalid strength",
			strength:       999,
			passphrase:     nil,
			serverResponse: GenerateSeedResponse{},
			serverStatus:   http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/generate-seed", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				// Verify request body
				var req GenerateSeedRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.strength, req.Strength)
				assert.Equal(t, tt.passphrase, req.Passphrase)

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			// Create client
			client := NewRenclaveClient(server.URL, 5*time.Second)
			ctx := context.Background()

			// Make request
			response, err := client.GenerateSeed(ctx, tt.strength, tt.passphrase)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.serverResponse.SeedPhrase, response.SeedPhrase)
				assert.Equal(t, tt.serverResponse.Entropy, response.Entropy)
				assert.Equal(t, tt.serverResponse.Strength, response.Strength)
				assert.Equal(t, tt.serverResponse.WordCount, response.WordCount)
			}
		})
	}
}

func TestRenclaveClient_ValidateSeed(t *testing.T) {
	tests := []struct {
		name           string
		seedPhrase     string
		serverResponse ValidateSeedResponse
		serverStatus   int
		expectError    bool
	}{
		{
			name:       "valid seed phrase",
			seedPhrase: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			serverResponse: ValidateSeedResponse{
				IsValid:   true,
				Strength:  128,
				WordCount: 12,
				Errors:    []string{},
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:       "invalid seed phrase",
			seedPhrase: "invalid seed phrase",
			serverResponse: ValidateSeedResponse{
				IsValid:   false,
				Strength:  0,
				WordCount: 3,
				Errors:    []string{"Invalid word count", "Invalid checksum"},
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "server error",
			seedPhrase:     "test phrase",
			serverResponse: ValidateSeedResponse{},
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "empty seed phrase",
			seedPhrase:     "",
			serverResponse: ValidateSeedResponse{},
			serverStatus:   http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/validate-seed", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Verify request body
				var req ValidateSeedRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.seedPhrase, req.SeedPhrase)

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			// Create client
			client := NewRenclaveClient(server.URL, 5*time.Second)
			ctx := context.Background()

			// Make request
			response, err := client.ValidateSeed(ctx, tt.seedPhrase)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.serverResponse.IsValid, response.IsValid)
				assert.Equal(t, tt.serverResponse.Strength, response.Strength)
				assert.Equal(t, tt.serverResponse.WordCount, response.WordCount)
				assert.Equal(t, tt.serverResponse.Errors, response.Errors)
			}
		})
	}
}

func TestRenclaveClient_GetInfo(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse InfoResponse
		serverStatus   int
		expectError    bool
	}{
		{
			name: "successful info retrieval",
			serverResponse: InfoResponse{
				Version:      "1.0.0",
				EnclaveID:    "test-enclave-id",
				Capabilities: []string{"seed_generation", "bip39_compliance"},
				Healthy:      true,
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "enclave unhealthy",
			serverResponse: InfoResponse{
				Version:      "1.0.0",
				EnclaveID:    "test-enclave-id",
				Capabilities: []string{"seed_generation"},
				Healthy:      false,
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "server error",
			serverResponse: InfoResponse{},
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/info", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			// Create client
			client := NewRenclaveClient(server.URL, 5*time.Second)
			ctx := context.Background()

			// Make request
			response, err := client.GetInfo(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.serverResponse.Version, response.Version)
				assert.Equal(t, tt.serverResponse.EnclaveID, response.EnclaveID)
				assert.Equal(t, tt.serverResponse.Capabilities, response.Capabilities)
				assert.Equal(t, tt.serverResponse.Healthy, response.Healthy)
			}
		})
	}
}

func TestRenclaveClient_Health(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse HealthResponse
		serverStatus   int
		expectError    bool
	}{
		{
			name: "healthy service",
			serverResponse: HealthResponse{
				Status: "healthy",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "ok status",
			serverResponse: HealthResponse{
				Status: "ok",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "unhealthy service",
			serverResponse: HealthResponse{
				Status: "unhealthy",
			},
			serverStatus: http.StatusOK,
			expectError:  true,
		},
		{
			name: "degraded service",
			serverResponse: HealthResponse{
				Status: "degraded",
			},
			serverStatus: http.StatusOK,
			expectError:  true,
		},
		{
			name:           "server error",
			serverResponse: HealthResponse{},
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/health", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			// Create client
			client := NewRenclaveClient(server.URL, 5*time.Second)
			ctx := context.Background()

			// Make request
			err := client.Health(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRenclaveClient_Timeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}))
	defer server.Close()

	// Create client with very short timeout
	client := NewRenclaveClient(server.URL, 50*time.Millisecond)
	ctx := context.Background()

	// Request should timeout
	err := client.Health(ctx)
	assert.Error(t, err)
	// Check for timeout-related error messages
	assert.True(t,
		contains(err.Error(), "timeout") ||
			contains(err.Error(), "deadline exceeded") ||
			contains(err.Error(), "Client.Timeout exceeded"),
		"Expected timeout-related error, got: %v", err)
}

func TestRenclaveClient_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}))
	defer server.Close()

	// Create client
	client := NewRenclaveClient(server.URL, 5*time.Second)

	// Create context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Request should be cancelled
	err := client.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestRenclaveClient_InvalidJSON(t *testing.T) {
	// Create server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	// Create client
	client := NewRenclaveClient(server.URL, 5*time.Second)
	ctx := context.Background()

	// Request should fail due to invalid JSON
	response, err := client.GetInfo(ctx)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "decode")
}

func TestRenclaveClient_NetworkError(t *testing.T) {
	// Create client with invalid URL
	client := NewRenclaveClient("http://invalid-host:9999", 5*time.Second)
	ctx := context.Background()

	// Request should fail due to network error
	err := client.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make request")
}

// Benchmark tests
func BenchmarkRenclaveClient_Health(b *testing.B) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}))
	defer server.Close()

	// Create client
	client := NewRenclaveClient(server.URL, 5*time.Second)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := client.Health(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenclaveClient_GenerateSeed(b *testing.B) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GenerateSeedResponse{
			SeedPhrase: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			Entropy:    "00000000000000000000000000000000",
			Strength:   256,
			WordCount:  24,
		})
	}))
	defer server.Close()

	// Create client
	client := NewRenclaveClient(server.URL, 5*time.Second)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GenerateSeed(ctx, 256, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test concurrent requests
func TestRenclaveClient_ConcurrentRequests(t *testing.T) {
	var requestCount int32 // Use atomic operations for thread safety
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}))
	defer server.Close()

	client := NewRenclaveClient(server.URL, 5*time.Second)
	ctx := context.Background()

	// Run concurrent requests
	concurrency := 10
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			err := client.Health(ctx)
			errChan <- err
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < concurrency; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}

	assert.Equal(t, int32(concurrency), atomic.LoadInt32(&requestCount))
}

// Helper function for contains check
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
