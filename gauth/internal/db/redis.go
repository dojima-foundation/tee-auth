package db

import (
	"context"
	"fmt"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/redis/go-redis/v9"
)

// RedisClient represents a Redis client connection
type RedisClient struct {
	client *redis.Client
	config *config.RedisConfig
}

// NewRedisClient creates a new Redis client connection
func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: client,
		config: cfg,
	}, nil
}

// GetClient returns the Redis client instance
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Health checks the Redis connection health
func (r *RedisClient) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// GetStats returns Redis connection statistics
func (r *RedisClient) GetStats() map[string]interface{} {
	stats := r.client.PoolStats()
	return map[string]interface{}{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
	}
}

// Session management methods

// SetSession stores a session in Redis with expiration
func (r *RedisClient) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, fmt.Sprintf("session:%s", sessionID), data, expiration).Err()
}

// GetSession retrieves a session from Redis
func (r *RedisClient) GetSession(ctx context.Context, sessionID string) (string, error) {
	return r.client.Get(ctx, fmt.Sprintf("session:%s", sessionID)).Result()
}

// DeleteSession removes a session from Redis
func (r *RedisClient) DeleteSession(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, fmt.Sprintf("session:%s", sessionID)).Err()
}

// ExtendSession extends the expiration of a session
func (r *RedisClient) ExtendSession(ctx context.Context, sessionID string, expiration time.Duration) error {
	return r.client.Expire(ctx, fmt.Sprintf("session:%s", sessionID), expiration).Err()
}

// Cache management methods

// Set stores a key-value pair in Redis with expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value from Redis
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Delete removes a key from Redis
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Rate limiting methods

// IncrementCounter increments a counter with expiration (for rate limiting)
func (r *RedisClient) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// GetCounter gets the current value of a counter
func (r *RedisClient) GetCounter(ctx context.Context, key string) (int64, error) {
	result, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return result, err
}

// Lock management methods (for distributed locking)

// AcquireLock attempts to acquire a distributed lock
func (r *RedisClient) AcquireLock(ctx context.Context, lockKey string, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, fmt.Sprintf("lock:%s", lockKey), "locked", expiration).Result()
}

// ReleaseLock releases a distributed lock
func (r *RedisClient) ReleaseLock(ctx context.Context, lockKey string) error {
	return r.client.Del(ctx, fmt.Sprintf("lock:%s", lockKey)).Err()
}

// ExtendLock extends the expiration of a lock
func (r *RedisClient) ExtendLock(ctx context.Context, lockKey string, expiration time.Duration) error {
	return r.client.Expire(ctx, fmt.Sprintf("lock:%s", lockKey), expiration).Err()
}

// List operations for queues and notifications

// LPush pushes elements to the left of a list
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPop pops an element from the right of a list
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// BRPop blocks and pops an element from the right of a list with timeout
func (r *RedisClient) BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return r.client.BRPop(ctx, timeout, keys...).Result()
}

// LLen returns the length of a list
func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// Hash operations for complex data structures

// HSet sets field-value pairs in a hash
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet gets a field value from a hash
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all field-value pairs from a hash
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel deletes fields from a hash
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// Pub/Sub operations for real-time notifications

// Publish publishes a message to a channel
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// PSubscribe subscribes to channels matching patterns
func (r *RedisClient) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return r.client.PSubscribe(ctx, patterns...)
}

// RedisInterface defines the interface for Redis operations
type RedisInterface interface {
	GetClient() *redis.Client
	Close() error
	Health(ctx context.Context) error
	GetStats() map[string]interface{}

	// Session management
	SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error
	GetSession(ctx context.Context, sessionID string) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	ExtendSession(ctx context.Context, sessionID string, expiration time.Duration) error

	// Basic operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)

	// Rate limiting
	IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error)
	GetCounter(ctx context.Context, key string) (int64, error)

	// Distributed locking
	AcquireLock(ctx context.Context, lockKey string, expiration time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, lockKey string) error
	ExtendLock(ctx context.Context, lockKey string, expiration time.Duration) error
}
