package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type DatabaseIntegrationTestSuite struct {
	suite.Suite
	db     *db.PostgresDB
	redis  *db.RedisClient
	config *config.Config
}

func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
	// Skip integration tests if not in integration test mode
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		suite.T().Skip("Skipping integration tests. Set INTEGRATION_TESTS=true to run.")
	}

	// Setup test configuration
	suite.config = &config.Config{
		Database: config.DatabaseConfig{
			Host:         getEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:         5432,
			Username:     getEnvOrDefault("TEST_DB_USER", "gauth"),
			Password:     getEnvOrDefault("TEST_DB_PASSWORD", "password"),
			Database:     getEnvOrDefault("TEST_DB_NAME", "gauth_test"),
			SSLMode:      "disable",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
			MaxLifetime:  5 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host:         getEnvOrDefault("TEST_REDIS_HOST", "localhost"),
			Port:         6379,
			Password:     getEnvOrDefault("TEST_REDIS_PASSWORD", ""),
			Database:     1, // Use different database for tests
			PoolSize:     10,
			MinIdleConns: 5,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
	}

	// Initialize database
	var err error
	suite.db, err = db.NewPostgresDB(&suite.config.Database)
	require.NoError(suite.T(), err)

	// Initialize Redis
	suite.redis, err = db.NewRedisClient(&suite.config.Redis)
	require.NoError(suite.T(), err)

	// Wait for connections to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(suite.T(), suite.db.Health(ctx))
	require.NoError(suite.T(), suite.redis.Health(ctx))

	// Run migrations (in a real setup, this would be done separately)
	suite.runTestMigrations()
}

func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.redis != nil {
		suite.redis.Close()
	}
}

func (suite *DatabaseIntegrationTestSuite) SetupTest() {
	// Clean up test data before each test
	suite.cleanupTestData()
}

func (suite *DatabaseIntegrationTestSuite) TearDownTest() {
	// Clean up test data after each test
	suite.cleanupTestData()
}

func (suite *DatabaseIntegrationTestSuite) TestOrganizationCRUD() {
	ctx := context.Background()
	db := suite.db.GetDB()

	// Create organization
	org := &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    "Test Organization",
		RootQuorum: models.Quorum{
			Threshold: 1,
		},
	}

	err := db.WithContext(ctx).Create(org).Error
	require.NoError(suite.T(), err)
	assert.NotZero(suite.T(), org.CreatedAt)
	assert.NotZero(suite.T(), org.UpdatedAt)

	// Read organization
	var retrievedOrg models.Organization
	err = db.WithContext(ctx).First(&retrievedOrg, "id = ?", org.ID).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), org.ID, retrievedOrg.ID)
	assert.Equal(suite.T(), org.Name, retrievedOrg.Name)
	assert.Equal(suite.T(), org.Version, retrievedOrg.Version)

	// Update organization
	retrievedOrg.Name = "Updated Organization"
	err = db.WithContext(ctx).Save(&retrievedOrg).Error
	require.NoError(suite.T(), err)

	// Verify update
	var updatedOrg models.Organization
	err = db.WithContext(ctx).First(&updatedOrg, "id = ?", org.ID).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Organization", updatedOrg.Name)
	assert.True(suite.T(), updatedOrg.UpdatedAt.After(updatedOrg.CreatedAt))

	// Delete organization
	err = db.WithContext(ctx).Delete(&updatedOrg).Error
	require.NoError(suite.T(), err)

	// Verify deletion
	var deletedOrg models.Organization
	err = db.WithContext(ctx).First(&deletedOrg, "id = ?", org.ID).Error
	assert.Error(suite.T(), err)
}

func (suite *DatabaseIntegrationTestSuite) TestUserManagement() {
	ctx := context.Background()
	db := suite.db.GetDB()

	// Create organization first
	org := &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    "Test Organization",
		RootQuorum: models.Quorum{
			Threshold: 1,
		},
	}
	require.NoError(suite.T(), db.WithContext(ctx).Create(org).Error)

	// Create user
	user := &models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Username:       "testuser",
		Email:          "test@example.com",
		PublicKey:      "test-public-key",
		Tags:           []string{"admin", "test"},
		IsActive:       true,
	}

	err := db.WithContext(ctx).Create(user).Error
	require.NoError(suite.T(), err)

	// Create auth method for user
	authMethod := &models.AuthMethod{
		ID:       uuid.New(),
		UserID:   user.ID,
		Type:     "API_KEY",
		Name:     "Test API Key",
		Data:     `{"key": "test-key", "created": "2024-01-01"}`,
		IsActive: true,
	}

	err = db.WithContext(ctx).Create(authMethod).Error
	require.NoError(suite.T(), err)

	// Retrieve user with auth methods
	var retrievedUser models.User
	err = db.WithContext(ctx).Preload("AuthMethods").First(&retrievedUser, "id = ?", user.ID).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Username, retrievedUser.Username)
	assert.Equal(suite.T(), user.Email, retrievedUser.Email)
	assert.Len(suite.T(), retrievedUser.AuthMethods, 1)
	assert.Equal(suite.T(), "API_KEY", retrievedUser.AuthMethods[0].Type)

	// Test unique constraints
	duplicateUser := &models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Username:       "testuser", // Same username
		Email:          "different@example.com",
		PublicKey:      "different-public-key",
		IsActive:       true,
	}

	err = db.WithContext(ctx).Create(duplicateUser).Error
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

func (suite *DatabaseIntegrationTestSuite) TestActivityAuditTrail() {
	ctx := context.Background()
	db := suite.db.GetDB()

	// Setup organization and user
	org := &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    "Test Organization",
		RootQuorum: models.Quorum{
			Threshold: 1,
		},
	}
	require.NoError(suite.T(), db.WithContext(ctx).Create(org).Error)

	user := &models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Username:       "testuser",
		Email:          "test@example.com",
		PublicKey:      "test-public-key",
		IsActive:       true,
	}
	require.NoError(suite.T(), db.WithContext(ctx).Create(user).Error)

	// Create activity
	activity := &models.Activity{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Type:           "SEED_GENERATION",
		Status:         "PENDING",
		Parameters:     json.RawMessage(`{"strength": 256, "passphrase": false}`),
		Intent: models.ActivityIntent{
			Fingerprint: "test-fingerprint",
			Summary:     "Generate 256-bit seed phrase",
		},
		CreatedBy: user.ID,
	}

	err := db.WithContext(ctx).Create(activity).Error
	require.NoError(suite.T(), err)

	// Create proof for activity
	proof := &models.Proof{
		ID:         uuid.New(),
		ActivityID: activity.ID,
		Type:       "SIGNATURE",
		Data:       "test-proof-data",
	}

	err = db.WithContext(ctx).Create(proof).Error
	require.NoError(suite.T(), err)

	// Retrieve activity
	var retrievedActivity models.Activity
	err = db.WithContext(ctx).First(&retrievedActivity, "id = ?", activity.ID).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), activity.Type, retrievedActivity.Type)
	assert.Equal(suite.T(), activity.Status, retrievedActivity.Status)

	// Update activity status
	retrievedActivity.Status = "COMPLETED"
	resultData := json.RawMessage(`{"seed_phrase": "test phrase", "entropy": "test entropy"}`)
	retrievedActivity.Result = resultData

	err = db.WithContext(ctx).Save(&retrievedActivity).Error
	require.NoError(suite.T(), err)

	// Verify status update
	var completedActivity models.Activity
	err = db.WithContext(ctx).First(&completedActivity, "id = ?", activity.ID).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "COMPLETED", completedActivity.Status)
	assert.NotEmpty(suite.T(), completedActivity.Result)
}

func (suite *DatabaseIntegrationTestSuite) TestWalletManagement() {
	ctx := context.Background()
	db := suite.db.GetDB()

	// Setup organization
	org := &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    "Test Organization",
		RootQuorum: models.Quorum{
			Threshold: 1,
		},
	}
	require.NoError(suite.T(), db.WithContext(ctx).Create(org).Error)

	// Create wallet
	wallet := &models.Wallet{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "Test Wallet",
		PublicKey:      "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
		Tags:           []string{"bitcoin", "mainnet"},
		IsActive:       true,
	}

	err := db.WithContext(ctx).Create(wallet).Error
	require.NoError(suite.T(), err)

	// Create wallet accounts
	accounts := []models.WalletAccount{
		{
			ID:            uuid.New(),
			WalletID:      wallet.ID,
			Name:          "Account 0",
			Path:          "m/44'/0'/0'/0/0",
			PublicKey:     "03a34b99f22c790c4e36b2b3c2c35a36db06226e41c692fc82b8b56ac1c540c5bd",
			Address:       "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			Curve:         "SECP256K1",
			AddressFormat: "P2PKH",
			IsActive:      true,
		},
		{
			ID:            uuid.New(),
			WalletID:      wallet.ID,
			Name:          "Account 1",
			Path:          "m/44'/0'/0'/0/1",
			PublicKey:     "02f9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9",
			Address:       "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4",
			Curve:         "SECP256K1",
			AddressFormat: "P2WPKH",
			IsActive:      true,
		},
	}

	for _, account := range accounts {
		err := db.WithContext(ctx).Create(&account).Error
		require.NoError(suite.T(), err)
	}

	// Retrieve wallet with accounts
	var retrievedWallet models.Wallet
	err = db.WithContext(ctx).Preload("Accounts").First(&retrievedWallet, "id = ?", wallet.ID).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), wallet.Name, retrievedWallet.Name)
	assert.Len(suite.T(), retrievedWallet.Accounts, 2)

	// Verify BIP44 paths
	for _, account := range retrievedWallet.Accounts {
		assert.Regexp(suite.T(), `^m/44'/\d+'/\d+'/\d+/\d+$`, account.Path)
		assert.NotEmpty(suite.T(), account.Address)
		assert.NotEmpty(suite.T(), account.PublicKey)
	}
}

func (suite *DatabaseIntegrationTestSuite) TestRedisOperations() {
	ctx := context.Background()

	// Test basic key-value operations
	key := "test:key:" + uuid.New().String()
	value := "test-value"

	err := suite.redis.Set(ctx, key, value, time.Hour)
	require.NoError(suite.T(), err)

	retrievedValue, err := suite.redis.Get(ctx, key)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), value, retrievedValue)

	// Test session management
	sessionID := uuid.New().String()
	userID := uuid.New().String()

	err = suite.redis.SetSession(ctx, sessionID, userID, 30*time.Minute)
	require.NoError(suite.T(), err)

	retrievedUserID, err := suite.redis.GetSession(ctx, sessionID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), userID, retrievedUserID)

	// Test session expiration extension
	err = suite.redis.ExtendSession(ctx, sessionID, time.Hour)
	require.NoError(suite.T(), err)

	// Test session deletion
	err = suite.redis.DeleteSession(ctx, sessionID)
	require.NoError(suite.T(), err)

	_, err = suite.redis.GetSession(ctx, sessionID)
	assert.Error(suite.T(), err) // Should not exist

	// Test distributed locking
	lockKey := "test-lock:" + uuid.New().String()

	acquired, err := suite.redis.AcquireLock(ctx, lockKey, time.Minute)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), acquired)

	// Try to acquire same lock again
	acquired2, err := suite.redis.AcquireLock(ctx, lockKey, time.Minute)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), acquired2)

	// Release lock
	err = suite.redis.ReleaseLock(ctx, lockKey)
	require.NoError(suite.T(), err)

	// Should be able to acquire again
	acquired3, err := suite.redis.AcquireLock(ctx, lockKey, time.Minute)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), acquired3)

	// Clean up
	suite.redis.ReleaseLock(ctx, lockKey)
}

func (suite *DatabaseIntegrationTestSuite) TestTransactionRollback() {
	ctx := context.Background()

	// Create organization
	org := &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    "Test Organization",
		RootQuorum: models.Quorum{
			Threshold: 1,
		},
	}

	// Test transaction rollback on error
	err := suite.db.Transaction(func(tx *gorm.DB) error {
		// Create organization
		if err := tx.WithContext(ctx).Create(org).Error; err != nil {
			return err
		}

		// Create user
		user := &models.User{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			Username:       "testuser",
			Email:          "test@example.com",
			PublicKey:      "test-public-key",
			IsActive:       true,
		}

		if err := tx.WithContext(ctx).Create(user).Error; err != nil {
			return err
		}

		// Force an error to trigger rollback
		return fmt.Errorf("forced error for rollback test")
	})

	assert.Error(suite.T(), err)

	// Verify that nothing was committed
	var count int64
	suite.db.GetDB().WithContext(ctx).Model(&models.Organization{}).Where("id = ?", org.ID).Count(&count)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *DatabaseIntegrationTestSuite) runTestMigrations() {
	// This is a simplified version - in production, you'd use proper migrations
	db := suite.db.GetDB()

	// Auto-migrate for testing (not recommended for production)
	err := db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.AuthMethod{},
		&models.Invitation{},
		&models.Policy{},
		&models.Tag{},
		&models.PrivateKey{},
		&models.Wallet{},
		&models.WalletAccount{},
		&models.Activity{},
		&models.Proof{},
	)
	require.NoError(suite.T(), err)
}

func (suite *DatabaseIntegrationTestSuite) cleanupTestData() {
	db := suite.db.GetDB()
	ctx := context.Background()

	// Delete in reverse dependency order
	tables := []string{
		"proofs", "activities", "wallet_accounts", "wallets",
		"private_keys", "tags", "policies", "invitations",
		"auth_methods", "users", "organizations",
	}

	for _, table := range tables {
		db.WithContext(ctx).Exec("DELETE FROM " + table)
	}

	// Clear Redis test database
	suite.redis.GetClient().FlushDB(context.Background())
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestDatabaseIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DatabaseIntegrationTestSuite))
}

// Benchmark integration tests
func BenchmarkDatabaseOperations(b *testing.B) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		b.Skip("Skipping integration benchmarks. Set INTEGRATION_TESTS=true to run.")
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:         getEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:         5432,
			Username:     getEnvOrDefault("TEST_DB_USER", "gauth"),
			Password:     getEnvOrDefault("TEST_DB_PASSWORD", "password"),
			Database:     getEnvOrDefault("TEST_DB_NAME", "gauth_test"),
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
			MaxLifetime:  5 * time.Minute,
		},
	}

	database, err := db.NewPostgresDB(&cfg.Database)
	if err != nil {
		b.Fatal(err)
	}
	defer database.Close()

	ctx := context.Background()
	db := database.GetDB()

	// Setup test data
	org := &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    "Benchmark Organization",
		RootQuorum: models.Quorum{
			Threshold: 1,
		},
	}
	db.WithContext(ctx).Create(org)

	b.ResetTimer()

	b.Run("CreateUser", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			user := &models.User{
				ID:             uuid.New(),
				OrganizationID: org.ID,
				Username:       fmt.Sprintf("user%d", i),
				Email:          fmt.Sprintf("user%d@example.com", i),
				PublicKey:      fmt.Sprintf("public-key-%d", i),
				IsActive:       true,
			}
			db.WithContext(ctx).Create(user)
		}
	})

	b.Run("QueryUsers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var users []models.User
			db.WithContext(ctx).Where("organization_id = ?", org.ID).Limit(10).Find(&users)
		}
	})
}
