package rest

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// MockRedisInterface is a simple mock for Redis operations
type MockRedisInterface struct {
	sessions map[string]string
}

func NewMockRedisInterface() *MockRedisInterface {
	return &MockRedisInterface{
		sessions: make(map[string]string),
	}
}

func (m *MockRedisInterface) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	m.sessions[sessionID] = data.(string)
	return nil
}

func (m *MockRedisInterface) GetSession(ctx context.Context, sessionID string) (string, error) {
	value, exists := m.sessions[sessionID]
	if !exists {
		return "", nil
	}
	return value, nil
}

func (m *MockRedisInterface) DeleteSession(ctx context.Context, sessionID string) error {
	delete(m.sessions, sessionID)
	return nil
}

func (m *MockRedisInterface) ExtendSession(ctx context.Context, sessionID string, expiration time.Duration) error {
	return nil
}

func (m *MockRedisInterface) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if str, ok := value.(string); ok {
		m.sessions[key] = str
	} else {
		// Handle non-string values
		m.sessions[key] = "mock_value"
	}
	return nil
}

func (m *MockRedisInterface) Get(ctx context.Context, key string) (string, error) {
	value, exists := m.sessions[key]
	if !exists {
		return "", nil
	}
	return value, nil
}

func (m *MockRedisInterface) Delete(ctx context.Context, key string) error {
	delete(m.sessions, key)
	return nil
}

func (m *MockRedisInterface) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.sessions[key]
	return exists, nil
}

func (m *MockRedisInterface) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if _, exists := m.sessions[key]; exists {
		return false, nil
	}
	m.sessions[key] = value.(string)
	return true, nil
}

func (m *MockRedisInterface) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return 1, nil
}

func (m *MockRedisInterface) GetCounter(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

func (m *MockRedisInterface) AcquireLock(ctx context.Context, lockKey string, expiration time.Duration) (bool, error) {
	return true, nil
}

func (m *MockRedisInterface) ReleaseLock(ctx context.Context, lockKey string) error {
	return nil
}

func (m *MockRedisInterface) ExtendLock(ctx context.Context, lockKey string, expiration time.Duration) error {
	return nil
}

func (m *MockRedisInterface) GetClient() *redis.Client {
	return nil
}

func (m *MockRedisInterface) Close() error {
	return nil
}

func (m *MockRedisInterface) Health(ctx context.Context) error {
	return nil
}

func (m *MockRedisInterface) GetStats() map[string]interface{} {
	return make(map[string]interface{})
}

// MockLogger is a simple mock for logger operations
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) Fatal(msg string, fields ...interface{}) {}
func (m *MockLogger) With(fields ...interface{}) interface{}  { return m }
