package service

import (
	"context"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// MockDatabase implements the database interface for testing
type MockDatabase struct{}

func (m *MockDatabase) GetDB() *gorm.DB {
	return nil // Return nil for simple tests - Status method now handles this gracefully
}

func (m *MockDatabase) Close() error {
	return nil
}

func (m *MockDatabase) Health(ctx context.Context) error {
	return nil
}

func (m *MockDatabase) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"max_open_connections": 0,
		"open_connections":     0,
		"in_use":               0,
		"idle":                 0,
	}
}

func (m *MockDatabase) Transaction(fn func(*gorm.DB) error) error {
	return nil
}

func (m *MockDatabase) BeginTx(ctx context.Context) *gorm.DB {
	return nil
}

// MockRedis implements the Redis interface for testing
type MockRedis struct{}

func (m *MockRedis) GetClient() *redis.Client {
	return nil
}

func (m *MockRedis) Close() error {
	return nil
}

func (m *MockRedis) Health(ctx context.Context) error {
	return nil
}

func (m *MockRedis) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"connected_clients": 0,
		"used_memory":       0,
	}
}

func (m *MockRedis) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	return nil
}

func (m *MockRedis) GetSession(ctx context.Context, sessionID string) (string, error) {
	return "", nil
}

func (m *MockRedis) DeleteSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *MockRedis) ExtendSession(ctx context.Context, sessionID string, expiration time.Duration) error {
	return nil
}

func (m *MockRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (m *MockRedis) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *MockRedis) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *MockRedis) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *MockRedis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return false, nil
}

func (m *MockRedis) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return 0, nil
}

func (m *MockRedis) GetCounter(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *MockRedis) AcquireLock(ctx context.Context, lockKey string, expiration time.Duration) (bool, error) {
	return false, nil
}

func (m *MockRedis) ReleaseLock(ctx context.Context, lockKey string) error {
	return nil
}

func (m *MockRedis) ExtendLock(ctx context.Context, lockKey string, expiration time.Duration) error {
	return nil
}

func TestGAuthService_Status(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
	}

	logger := logger.NewDefault()
	renclave := NewRenclaveClient("http://mock-server", 1*time.Second)

	// Create a mock database and Redis for testing
	mockDB := &MockDatabase{}
	mockRedis := &MockRedis{}

	service := &GAuthService{
		config:   cfg,
		logger:   logger,
		renclave: renclave,
		db:       mockDB,
		redis:    mockRedis,
	}

	ctx := context.Background()
	response, err := service.Status(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "1.0.0", response.Version)
	assert.NotEmpty(t, response.BuildTime)
	assert.NotEmpty(t, response.GitCommit)
	assert.NotZero(t, response.Uptime)
	assert.NotEmpty(t, response.Metrics)
}

func TestGAuthService_RequestSeedGeneration_Error(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
	}

	logger := logger.NewDefault()
	renclave := NewRenclaveClient("http://invalid-server", 1*time.Second)

	service := &GAuthService{
		config:   cfg,
		logger:   logger,
		renclave: renclave,
	}

	ctx := context.Background()
	response, err := service.RequestSeedGeneration(ctx, uuid.New().String(), uuid.New().String(), 256, nil)

	// Should error since the server doesn't exist
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGAuthService_ValidateSeed_Error(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
	}

	logger := logger.NewDefault()
	renclave := NewRenclaveClient("http://invalid-server", 1*time.Second)

	service := &GAuthService{
		config:   cfg,
		logger:   logger,
		renclave: renclave,
	}

	ctx := context.Background()
	response, err := service.ValidateSeed(ctx, "test seed phrase")

	// Should error since the server doesn't exist
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGAuthService_GetEnclaveInfo_Error(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
	}

	logger := logger.NewDefault()
	renclave := NewRenclaveClient("http://invalid-server", 1*time.Second)

	service := &GAuthService{
		config:   cfg,
		logger:   logger,
		renclave: renclave,
	}

	ctx := context.Background()
	response, err := service.GetEnclaveInfo(ctx)

	// Should error since the server doesn't exist
	assert.Error(t, err)
	assert.Nil(t, response)
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}
