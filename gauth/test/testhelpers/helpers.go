package testhelpers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestDatabase provides helper functions for test database operations
type TestDatabase struct {
	DB     *db.PostgresDB
	Config *config.DatabaseConfig
}

// TestRedis provides helper functions for test Redis operations
type TestRedis struct {
	Client *db.RedisClient
	Config *config.RedisConfig
}

// NewTestDatabase creates a test database connection
func NewTestDatabase(t *testing.T) *TestDatabase {
	if os.Getenv("INTEGRATION_TESTS") != "true" && os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping database tests. Set INTEGRATION_TESTS=true or E2E_TESTS=true to run.")
	}

	config := &config.DatabaseConfig{
		Host:         GetEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:         5432,
		Username:     GetEnvOrDefault("TEST_DB_USER", "gauth"),
		Password:     GetEnvOrDefault("TEST_DB_PASSWORD", "password"),
		Database:     GetEnvOrDefault("TEST_DB_NAME", "gauth_test"),
		SSLMode:      "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		MaxLifetime:  5 * time.Minute,
	}

	database, err := db.NewPostgresDB(config)
	require.NoError(t, err)

	// Wait for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, database.Health(ctx))

	return &TestDatabase{
		DB:     database,
		Config: config,
	}
}

// NewTestRedis creates a test Redis connection
func NewTestRedis(t *testing.T) *TestRedis {
	if os.Getenv("INTEGRATION_TESTS") != "true" && os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping Redis tests. Set INTEGRATION_TESTS=true or E2E_TESTS=true to run.")
	}

	config := &config.RedisConfig{
		Host:         GetEnvOrDefault("TEST_REDIS_HOST", "localhost"),
		Port:         6379,
		Password:     GetEnvOrDefault("TEST_REDIS_PASSWORD", ""),
		Database:     getTestRedisDB(),
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	client, err := db.NewRedisClient(config)
	require.NoError(t, err)

	// Wait for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, client.Health(ctx))

	return &TestRedis{
		Client: client,
		Config: config,
	}
}

// Close closes the test database connection
func (td *TestDatabase) Close() {
	if td.DB != nil {
		td.DB.Close()
	}
}

// Close closes the test Redis connection
func (tr *TestRedis) Close() {
	if tr.Client != nil {
		tr.Client.Close()
	}
}

// Cleanup removes all test data from the database
func (td *TestDatabase) Cleanup(ctx context.Context) {
	db := td.DB.GetDB()

	// Delete in reverse dependency order
	tables := []string{
		"proofs", "activities", "wallet_accounts", "wallets",
		"private_keys", "tags", "policies", "invitations",
		"auth_methods", "users", "organizations", "sessions",
		"quorum_members", "user_tags", "private_key_tags", "wallet_tags",
	}

	for _, table := range tables {
		db.WithContext(ctx).Exec("DELETE FROM " + table)
	}
}

// Cleanup removes all test data from Redis
func (tr *TestRedis) Cleanup(ctx context.Context) {
	tr.Client.GetClient().FlushDB(ctx)
}

// CreateTestOrganization creates a test organization with optional users
func (td *TestDatabase) CreateTestOrganization(ctx context.Context, name string, userCount int) (*models.Organization, []models.User, error) {
	db := td.DB.GetDB()

	org := &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    name,
		RootQuorum: models.Quorum{
			Threshold: 1,
		},
	}

	if err := db.WithContext(ctx).Create(org).Error; err != nil {
		return nil, nil, err
	}

	var users []models.User
	for i := 0; i < userCount; i++ {
		user := models.User{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			Username:       fmt.Sprintf("testuser%d", i),
			Email:          fmt.Sprintf("testuser%d@example.com", i),
			PublicKey:      fmt.Sprintf("public-key-%d", i),
			IsActive:       true,
		}

		if err := db.WithContext(ctx).Create(&user).Error; err != nil {
			return nil, nil, err
		}

		users = append(users, user)
	}

	return org, users, nil
}

// CreateTestActivity creates a test activity
func (td *TestDatabase) CreateTestActivity(ctx context.Context, orgID, userID uuid.UUID, activityType string) (*models.Activity, error) {
	db := td.DB.GetDB()

	activity := &models.Activity{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Type:           activityType,
		Status:         "PENDING",
		Parameters:     json.RawMessage(`{"test": "data"}`),
		Intent: models.ActivityIntent{
			Fingerprint: "test-fingerprint",
			Summary:     "Test activity",
		},
		CreatedBy: userID,
	}

	if err := db.WithContext(ctx).Create(activity).Error; err != nil {
		return nil, err
	}

	return activity, nil
}

// CreateTestWallet creates a test wallet with accounts
func (td *TestDatabase) CreateTestWallet(ctx context.Context, orgID uuid.UUID, accountCount int) (*models.Wallet, []models.WalletAccount, error) {
	db := td.DB.GetDB()

	wallet := &models.Wallet{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "Test Wallet",
		PublicKey:      "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
		IsActive:       true,
	}

	if err := db.WithContext(ctx).Create(wallet).Error; err != nil {
		return nil, nil, err
	}

	var accounts []models.WalletAccount
	for i := 0; i < accountCount; i++ {
		account := models.WalletAccount{
			ID:            uuid.New(),
			WalletID:      wallet.ID,
			Name:          fmt.Sprintf("Account %d", i),
			Path:          fmt.Sprintf("m/44'/0'/0'/0/%d", i),
			PublicKey:     fmt.Sprintf("03%064x", rand.Int63()),
			Address:       generateTestAddress(i),
			Curve:         "SECP256K1",
			AddressFormat: "P2PKH",
			IsActive:      true,
		}

		if err := db.WithContext(ctx).Create(&account).Error; err != nil {
			return nil, nil, err
		}

		accounts = append(accounts, account)
	}

	return wallet, accounts, nil
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, timeout time.Duration, condition func() bool, message string) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Timeout waiting for condition: %s", message)
}

// AssertEventuallyConsistent checks that a condition becomes true within a timeout
func AssertEventuallyConsistent(t *testing.T, timeout time.Duration, check func() error) {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		if err := check(); err == nil {
			return
		} else {
			lastErr = err
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("Condition never became true within %v. Last error: %v", timeout, lastErr)
}

// GenerateTestData generates random test data for performance testing
type TestDataGenerator struct {
	rand *rand.Rand
}

func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *TestDataGenerator) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[g.rand.Intn(len(charset))]
	}
	return string(b)
}

func (g *TestDataGenerator) RandomEmail() string {
	return fmt.Sprintf("%s@%s.com", g.RandomString(8), g.RandomString(6))
}

func (g *TestDataGenerator) RandomOrganization() *models.Organization {
	return &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    "Test Org " + g.RandomString(8),
		RootQuorum: models.Quorum{
			Threshold: g.rand.Intn(3) + 1,
		},
	}
}

func (g *TestDataGenerator) RandomUser(orgID uuid.UUID) *models.User {
	return &models.User{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Username:       "user" + g.RandomString(8),
		Email:          g.RandomEmail(),
		PublicKey:      g.RandomString(64),
		IsActive:       g.rand.Float32() > 0.1, // 90% active
	}
}

// Database transaction helpers
func (td *TestDatabase) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return td.DB.Transaction(fn)
}

// Assertion helpers for database operations
func AssertRecordExists(t *testing.T, db *gorm.DB, model interface{}, condition string, args ...interface{}) {
	var count int64
	err := db.Model(model).Where(condition, args...).Count(&count).Error
	require.NoError(t, err)
	require.Greater(t, count, int64(0), "Expected record to exist but it doesn't")
}

func AssertRecordNotExists(t *testing.T, db *gorm.DB, model interface{}, condition string, args ...interface{}) {
	var count int64
	err := db.Model(model).Where(condition, args...).Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(0), count, "Expected record to not exist but it does")
}

func AssertRecordCount(t *testing.T, db *gorm.DB, model interface{}, expectedCount int64, condition string, args ...interface{}) {
	var count int64
	err := db.Model(model).Where(condition, args...).Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, expectedCount, count, "Record count mismatch")
}

// Redis test helpers
func (tr *TestRedis) SetTestSession(ctx context.Context, sessionID, userID string, expiration time.Duration) error {
	return tr.Client.SetSession(ctx, sessionID, userID, expiration)
}

func (tr *TestRedis) GetTestSession(ctx context.Context, sessionID string) (string, error) {
	return tr.Client.GetSession(ctx, sessionID)
}

// Performance measurement helpers
type PerformanceMeasurement struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

func StartMeasurement(name string) *PerformanceMeasurement {
	return &PerformanceMeasurement{
		Name:      name,
		StartTime: time.Now(),
	}
}

func (pm *PerformanceMeasurement) End() {
	pm.EndTime = time.Now()
	pm.Duration = pm.EndTime.Sub(pm.StartTime)
}

func (pm *PerformanceMeasurement) AssertFasterThan(t *testing.T, maxDuration time.Duration) {
	require.Less(t, pm.Duration, maxDuration,
		"Operation %s took %v, expected less than %v", pm.Name, pm.Duration, maxDuration)
}

// Test environment helpers
func IsIntegrationTest() bool {
	return os.Getenv("INTEGRATION_TESTS") == "true"
}

func IsE2ETest() bool {
	return os.Getenv("E2E_TESTS") == "true"
}

func SkipIfNotIntegration(t *testing.T) {
	if !IsIntegrationTest() {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}
}

func SkipIfNotE2E(t *testing.T) {
	if !IsE2ETest() {
		t.Skip("Skipping E2E test. Set E2E_TESTS=true to run.")
	}
}

// Helper functions
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getTestRedisDB() int {
	// Use different Redis databases for different test types
	if IsE2ETest() {
		return 4
	} else if IsIntegrationTest() {
		return 3
	}
	return 2
}

func generateTestAddress(index int) string {
	// Generate deterministic test addresses for consistency
	addresses := []string{
		"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
		"bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4",
		"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",
		"bc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qccfmv3",
		"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2",
	}
	return addresses[index%len(addresses)]
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// Int32Ptr returns a pointer to the given int32
func Int32Ptr(i int32) *int32 {
	return &i
}

// BoolPtr returns a pointer to the given bool
func BoolPtr(b bool) *bool {
	return &b
}

// Database migration helpers for tests
func (td *TestDatabase) RunMigrations(ctx context.Context) error {
	db := td.DB.GetDB()

	// Auto-migrate for testing (not recommended for production)
	return db.WithContext(ctx).AutoMigrate(
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
}

// Check if database is ready for testing
func (td *TestDatabase) IsReady(ctx context.Context) bool {
	return td.DB.Health(ctx) == nil
}

// Check if Redis is ready for testing
func (tr *TestRedis) IsReady(ctx context.Context) bool {
	return tr.Client.Health(ctx) == nil
}

// Cleanup function for use in defer statements
func (td *TestDatabase) DeferCleanup(ctx context.Context) func() {
	return func() {
		td.Cleanup(ctx)
		td.Close()
	}
}

func (tr *TestRedis) DeferCleanup(ctx context.Context) func() {
	return func() {
		tr.Cleanup(ctx)
		tr.Close()
	}
}
