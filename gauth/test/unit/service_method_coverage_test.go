package unit

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// MockLogger implements the logger interface for testing
type MockLogger struct {
	*slog.Logger
}

func NewMockLogger() *MockLogger {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	return &MockLogger{Logger: logger}
}

// testDBWrapper wraps TestDB to implement DatabaseInterface
type testDBWrapper struct {
	testDB *testhelpers.TestDB
}

func (w *testDBWrapper) GetDB() *gorm.DB {
	return w.testDB.GetDB()
}

func (w *testDBWrapper) Close() error {
	return w.testDB.Close()
}

func (w *testDBWrapper) Health(ctx context.Context) error {
	return w.testDB.Health(ctx)
}

func (w *testDBWrapper) GetStats() map[string]interface{} {
	return w.testDB.GetStats()
}

func (w *testDBWrapper) Transaction(fn func(*gorm.DB) error) error {
	return w.testDB.Transaction(fn)
}

func (w *testDBWrapper) BeginTx(ctx context.Context) *gorm.DB {
	return w.testDB.BeginTx(ctx)
}

// mockRedisClient implements RedisInterface for testing
type mockRedisClient struct{}

// mockEnclaveClient implements EnclaveClientInterface for testing
type mockEnclaveClient struct{}

func (m *mockEnclaveClient) GenerateSeed(ctx context.Context, strength int, passphrase *string) (*service.GenerateSeedResponse, error) {
	return &service.GenerateSeedResponse{
		SeedPhrase: "encrypted_seed_hex_data_placeholder", // Mock encrypted seed data
		Entropy:    "0000000000000000",
		Strength:   strength,
		WordCount:  12,
	}, nil
}

func (m *mockEnclaveClient) ValidateSeed(ctx context.Context, seedPhrase string, encryptedEntropy *string) (*service.ValidateSeedResponse, error) {
	return &service.ValidateSeedResponse{
		IsValid:   true,
		Strength:  128,
		WordCount: 12,
	}, nil
}

func (m *mockEnclaveClient) GetInfo(ctx context.Context) (*service.InfoResponse, error) {
	return &service.InfoResponse{
		Version: "1.0.0",
	}, nil
}

func (m *mockEnclaveClient) Health(ctx context.Context) error {
	return nil
}

func (m *mockEnclaveClient) DeriveKey(ctx context.Context, seedPhrase, path, curve string) (*service.DeriveKeyResponse, error) {
	return &service.DeriveKeyResponse{
		PrivateKey: "mock-private-key",
		PublicKey:  "mock-public-key",
	}, nil
}

func (m *mockEnclaveClient) DeriveAddress(ctx context.Context, seedPhrase, path, curve string) (*service.DeriveAddressResponse, error) {
	return &service.DeriveAddressResponse{
		Address: "0x742D35CC6Bf8B8E0b8F8F8F8F8F8F8F8F8F8F8F8",
	}, nil
}

func (m *mockRedisClient) GetClient() *redis.Client         { return nil }
func (m *mockRedisClient) Close() error                     { return nil }
func (m *mockRedisClient) Health(ctx context.Context) error { return nil }
func (m *mockRedisClient) GetStats() map[string]interface{} { return map[string]interface{}{} }

// Session management
func (m *mockRedisClient) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	return nil
}
func (m *mockRedisClient) GetSession(ctx context.Context, sessionID string) (string, error) {
	return "{}", nil
}
func (m *mockRedisClient) DeleteSession(ctx context.Context, sessionID string) error { return nil }
func (m *mockRedisClient) ExtendSession(ctx context.Context, sessionID string, expiration time.Duration) error {
	return nil
}

// Basic operations
func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	return "test-value", nil
}
func (m *mockRedisClient) Delete(ctx context.Context, key string) error         { return nil }
func (m *mockRedisClient) Exists(ctx context.Context, key string) (bool, error) { return true, nil }
func (m *mockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return true, nil
}

// Rate limiting
func (m *mockRedisClient) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return 1, nil
}
func (m *mockRedisClient) GetCounter(ctx context.Context, key string) (int64, error) { return 0, nil }

// Locking
func (m *mockRedisClient) AcquireLock(ctx context.Context, lockKey string, expiration time.Duration) (bool, error) {
	return true, nil
}
func (m *mockRedisClient) ReleaseLock(ctx context.Context, lockKey string) error { return nil }
func (m *mockRedisClient) ExtendLock(ctx context.Context, lockKey string, expiration time.Duration) error {
	return nil
}

// Queues
func (m *mockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return nil
}
func (m *mockRedisClient) RPop(ctx context.Context, key string) (string, error) { return "test", nil }
func (m *mockRedisClient) BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return []string{"test"}, nil
}
func (m *mockRedisClient) LLen(ctx context.Context, key string) (int64, error) { return 0, nil }

// Hash operations
func (m *mockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return nil
}
func (m *mockRedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return "test", nil
}
func (m *mockRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (m *mockRedisClient) HDel(ctx context.Context, key string, fields ...string) error { return nil }

// Pub/Sub
func (m *mockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return nil
}
func (m *mockRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return nil
}
func (m *mockRedisClient) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return nil
}

// TestServiceMethodCoverage_ActivityService tests actual ActivityService method calls
func TestServiceMethodCoverage_ActivityService(t *testing.T) {
	// Setup
	testDB := testhelpers.SetupTestDB(t)
	defer testDB.Close()

	// Create test organization and user
	orgID := uuid.New()
	userID := uuid.New()

	org := &models.Organization{
		ID:      orgID,
		Name:    "Test Org",
		Version: "1.0",
	}
	require.NoError(t, testDB.Create(org).Error)

	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Username:       "testuser",
		Email:          "test@example.com",
		PublicKey:      "test-public-key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, testDB.Create(user).Error)

	// Create a mock logger and config
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "test",
			Password: "test",
			Database: "test",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	log := &logger.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

	// Create database wrapper that implements DatabaseInterface
	dbWrapper := &testDBWrapper{testDB: testDB}

	// Create service instance
	redisClient := &mockRedisClient{}
	enclaveClient := &mockEnclaveClient{}
	svc := service.NewGAuthServiceWithEnclave(cfg, log, dbWrapper, redisClient, enclaveClient)

	ctx := context.Background()

	// Test CreateActivity - this should increase coverage
	activity, err := svc.CreateActivity(ctx, orgID.String(), "TEST_ACTIVITY", `{"test": "data"}`, userID.String())
	require.NoError(t, err)
	assert.NotNil(t, activity)
	assert.Equal(t, "TEST_ACTIVITY", activity.Type)
	assert.Equal(t, orgID.String(), activity.OrganizationID.String())
	assert.Equal(t, userID.String(), activity.CreatedBy.String())

	// Test GetActivity - this should increase coverage
	getActivity, err := svc.GetActivity(ctx, activity.ID.String())
	require.NoError(t, err)
	assert.Equal(t, activity.ID, getActivity.ID)
	assert.Equal(t, "TEST_ACTIVITY", getActivity.Type)

	// Test ListActivities - this should increase coverage
	activities, nextToken, err := svc.ListActivities(ctx, orgID.String(), nil, nil, 10, "")
	require.NoError(t, err)
	assert.NotEmpty(t, activities)
	assert.Equal(t, activity.ID, activities[0].ID)
	assert.Empty(t, nextToken) // Should be empty for single item
}

// TestServiceMethodCoverage_AuthService tests actual AuthService method calls
func TestServiceMethodCoverage_AuthService(t *testing.T) {
	// Setup
	testDB := testhelpers.SetupTestDB(t)
	defer testDB.Close()

	// Create test organization and user
	orgID := uuid.New()
	userID := uuid.New()
	authMethodID := uuid.New()

	org := &models.Organization{
		ID:      orgID,
		Name:    "Test Org",
		Version: "1.0",
	}
	require.NoError(t, testDB.Create(org).Error)

	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Username:       "testuser",
		Email:          "test@example.com",
		PublicKey:      "test-public-key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, testDB.Create(user).Error)

	// Create service instance
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "test",
			Password: "test",
			Database: "test",
			SSLMode:  "disable",
		},
	}

	log := &logger.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	dbWrapper := &testDBWrapper{testDB: testDB}
	redisClient := &mockRedisClient{}
	enclaveClient := &mockEnclaveClient{}
	svc := service.NewGAuthServiceWithEnclave(cfg, log, dbWrapper, redisClient, enclaveClient)

	ctx := context.Background()

	// Test Authenticate - this should increase coverage
	signature := "test-signature"
	timestamp := time.Now().Format(time.RFC3339)

	authResponse, err := svc.Authenticate(ctx, orgID.String(), userID.String(), authMethodID.String(), signature, timestamp)
	require.NoError(t, err)
	assert.NotNil(t, authResponse)
	assert.NotEmpty(t, authResponse.SessionToken)
	assert.True(t, authResponse.ExpiresAt.After(time.Now()))

	// Test Authorize - this should increase coverage
	activityType := "CREATE_WALLET"
	parameters := `{"walletName": "test-wallet"}`

	authzResponse, err := svc.Authorize(ctx, authResponse.SessionToken, activityType, parameters)
	require.NoError(t, err)
	assert.NotNil(t, authzResponse)
	assert.False(t, authzResponse.Authorized) // Should require quorum approval
	assert.Equal(t, "Requires quorum approval", authzResponse.Reason)
}

// TestServiceMethodCoverage_EnclaveService tests actual EnclaveService method calls
func TestServiceMethodCoverage_EnclaveService(t *testing.T) {
	// Setup
	testDB := testhelpers.SetupTestDB(t)
	defer testDB.Close()

	// Create test organization and user
	orgID := uuid.New()
	userID := uuid.New()

	org := &models.Organization{
		ID:      orgID,
		Name:    "Test Org",
		Version: "1.0",
	}
	require.NoError(t, testDB.Create(org).Error)

	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Username:       "testuser",
		Email:          "test@example.com",
		PublicKey:      "test-public-key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, testDB.Create(user).Error)

	// Create service instance
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "test",
			Password: "test",
			Database: "test",
			SSLMode:  "disable",
		},
	}

	log := &logger.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	dbWrapper := &testDBWrapper{testDB: testDB}
	redisClient := &mockRedisClient{}
	enclaveClient := &mockEnclaveClient{}
	svc := service.NewGAuthServiceWithEnclave(cfg, log, dbWrapper, redisClient, enclaveClient)

	ctx := context.Background()

	// Test RequestSeedGeneration - this should increase coverage
	strength := 256

	seedResponse, err := svc.RequestSeedGeneration(ctx, orgID.String(), userID.String(), strength, nil)
	require.NoError(t, err)
	assert.NotNil(t, seedResponse)
	assert.NotEmpty(t, seedResponse.RequestID)

	// Test ValidateSeed - this should increase coverage
	validSeedPhrase := "encrypted_seed_hex_data_placeholder" // Mock encrypted seed data

	validateResponse, err := svc.ValidateSeed(ctx, validSeedPhrase)
	require.NoError(t, err)
	assert.NotNil(t, validateResponse)
	assert.True(t, validateResponse.IsValid)
	assert.Equal(t, 128, validateResponse.Strength)
}

// TestServiceMethodCoverage_PrivateKeyService tests actual PrivateKeyService method calls
func TestServiceMethodCoverage_PrivateKeyService(t *testing.T) {
	// Setup
	testDB := testhelpers.SetupTestDB(t)
	defer testDB.Close()

	// Create test organization and wallet
	orgID := uuid.New()
	walletID := uuid.New()

	org := &models.Organization{
		ID:      orgID,
		Name:    "Test Org",
		Version: "1.0",
	}
	require.NoError(t, testDB.Create(org).Error)

	wallet := &models.Wallet{
		ID:             walletID,
		OrganizationID: orgID,
		Name:           "Test Wallet",
		SeedPhrase:     "encrypted_seed_hex_data_placeholder", // Mock encrypted seed data
		PublicKey:      "test-public-key",
		Tags:           []string{"test"},
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, testDB.Create(wallet).Error)

	// Create service instance
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "test",
			Password: "test",
			Database: "test",
			SSLMode:  "disable",
		},
	}

	log := &logger.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	dbWrapper := &testDBWrapper{testDB: testDB}
	redisClient := &mockRedisClient{}
	enclaveClient := &mockEnclaveClient{}
	svc := service.NewGAuthServiceWithEnclave(cfg, log, dbWrapper, redisClient, enclaveClient)

	ctx := context.Background()

	// Test CreatePrivateKey - this should increase coverage
	curve := "CURVE_SECP256K1"
	var privateKeyMaterial *string

	createResponse, err := svc.CreatePrivateKey(ctx, orgID.String(), walletID.String(), "Test Private Key", curve, privateKeyMaterial, []string{"test"})
	require.NoError(t, err)
	assert.NotNil(t, createResponse)
	assert.NotEmpty(t, createResponse.ID)

	// Test GetPrivateKey - this should increase coverage
	privateKeyID := createResponse.ID.String()

	getResponse, err := svc.GetPrivateKey(ctx, privateKeyID)
	require.NoError(t, err)
	assert.NotNil(t, getResponse)
	assert.Equal(t, createResponse.ID, getResponse.ID)
	assert.Equal(t, "Test Private Key", getResponse.Name)

	// Test ListPrivateKeys - this should increase coverage
	privateKeys, nextToken, err := svc.ListPrivateKeys(ctx, orgID.String(), 10, "")
	require.NoError(t, err)
	assert.NotEmpty(t, privateKeys)
	assert.Equal(t, createResponse.ID, privateKeys[0].ID)
	assert.Empty(t, nextToken)

	// Test DeletePrivateKey - this should increase coverage
	err = svc.DeletePrivateKey(ctx, privateKeyID, false)
	require.NoError(t, err)

	// Verify private key was deleted
	var deletedPrivateKey models.PrivateKey
	err = testDB.GetDB().Where("id = ?", privateKeyID).First(&deletedPrivateKey).Error
	assert.Error(t, err) // Should not find the deleted private key
}

// TestServiceMethodCoverage_WalletService tests actual WalletService method calls
func TestServiceMethodCoverage_WalletService(t *testing.T) {
	// Setup
	testDB := testhelpers.SetupTestDB(t)
	defer testDB.Close()

	// Create test organization
	orgID := uuid.New()

	org := &models.Organization{
		ID:      orgID,
		Name:    "Test Org",
		Version: "1.0",
	}
	require.NoError(t, testDB.Create(org).Error)

	// Create service instance
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "test",
			Password: "test",
			Database: "test",
			SSLMode:  "disable",
		},
	}

	log := &logger.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	dbWrapper := &testDBWrapper{testDB: testDB}
	redisClient := &mockRedisClient{}
	enclaveClient := &mockEnclaveClient{}
	svc := service.NewGAuthServiceWithEnclave(cfg, log, dbWrapper, redisClient, enclaveClient)

	ctx := context.Background()

	// Test CreateWallet - this should increase coverage
	accounts := []models.WalletAccount{
		{
			Name:          "Test Account",
			Curve:         "CURVE_SECP256K1",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		},
	}
	mnemonicLength := int32(12)

	createResponse, seedPhrases, err := svc.CreateWallet(ctx, orgID.String(), "Test Wallet", accounts, &mnemonicLength, []string{"test"})
	require.NoError(t, err)
	assert.NotNil(t, createResponse)
	assert.NotEmpty(t, createResponse.ID)
	assert.NotEmpty(t, seedPhrases)
	assert.Len(t, seedPhrases, 1)

	// Test GetWallet - this should increase coverage
	walletID := createResponse.ID.String()

	getResponse, err := svc.GetWallet(ctx, walletID)
	require.NoError(t, err)
	assert.NotNil(t, getResponse)
	assert.Equal(t, createResponse.ID, getResponse.ID)
	assert.Equal(t, "Test Wallet", getResponse.Name)

	// Test ListWallets - this should increase coverage
	wallets, nextToken, err := svc.ListWallets(ctx, orgID.String(), 10, "")
	require.NoError(t, err)
	assert.NotEmpty(t, wallets)
	assert.Equal(t, createResponse.ID, wallets[0].ID)
	assert.Empty(t, nextToken)

	// Test DeleteWallet - this should increase coverage
	err = svc.DeleteWallet(ctx, walletID, false)
	require.NoError(t, err)

	// Verify wallet was deleted
	var deletedWallet models.Wallet
	err = testDB.GetDB().Where("id = ?", walletID).First(&deletedWallet).Error
	assert.Error(t, err) // Should not find the deleted wallet
}
