package testhelpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB provides a simplified database interface for unit tests using SQLite in memory
type TestDB struct {
	db *gorm.DB
}

// SetupTestDB creates an in-memory SQLite database for unit tests
func SetupTestDB(t *testing.T) *TestDB {
	// Use SQLite in-memory database for fast unit tests
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent for tests
	})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
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
	require.NoError(t, err)

	// Create the quorum_members table (many-to-many relationship)
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS quorum_members (
			organization_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			PRIMARY KEY (organization_id, user_id)
		)
	`).Error
	require.NoError(t, err)

	return &TestDB{db: db}
}

// GetDB returns the underlying GORM database instance
func (tdb *TestDB) GetDB() *gorm.DB {
	return tdb.db
}

// Cleanup removes all data from the test database
func (tdb *TestDB) Cleanup() {
	// Truncate all tables in reverse dependency order
	tables := []string{
		"proofs", "activities", "wallet_accounts", "wallets",
		"private_keys", "auth_methods", "users", "organizations",
		"invitations", "policies", "tags",
	}

	for _, table := range tables {
		tdb.db.Exec("DELETE FROM " + table)
	}
}

// Implement db.DatabaseInterface for compatibility
func (tdb *TestDB) Create(value interface{}) *gorm.DB {
	return tdb.db.Create(value)
}

func (tdb *TestDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return tdb.db.First(dest, conds...)
}

func (tdb *TestDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	return tdb.db.Find(dest, conds...)
}

func (tdb *TestDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	return tdb.db.Where(query, args...)
}

func (tdb *TestDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	return tdb.db.Delete(value, conds...)
}

func (tdb *TestDB) Preload(query string, args ...interface{}) *gorm.DB {
	return tdb.db.Preload(query, args...)
}

func (tdb *TestDB) Order(value interface{}) *gorm.DB {
	return tdb.db.Order(value)
}

func (tdb *TestDB) Limit(limit int) *gorm.DB {
	return tdb.db.Limit(limit)
}

func (tdb *TestDB) Transaction(fn func(tx *gorm.DB) error) error {
	return tdb.db.Transaction(fn)
}

func (tdb *TestDB) BeginTx(ctx context.Context) *gorm.DB {
	return tdb.db.Begin()
}

func (tdb *TestDB) Health(ctx context.Context) error {
	sqlDB, err := tdb.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (tdb *TestDB) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"driver": "sqlite",
		"status": "test",
	}
}

func (tdb *TestDB) Close() error {
	sqlDB, err := tdb.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// MockDB for unit tests that don't need actual database persistence
type MockDB struct {
	CreateFn      func(value interface{}) *gorm.DB
	FirstFn       func(dest interface{}, conds ...interface{}) *gorm.DB
	FindFn        func(dest interface{}, conds ...interface{}) *gorm.DB
	WhereFn       func(query interface{}, args ...interface{}) *gorm.DB
	DeleteFn      func(value interface{}, conds ...interface{}) *gorm.DB
	PreloadFn     func(query string, args ...interface{}) *gorm.DB
	OrderFn       func(value interface{}) *gorm.DB
	LimitFn       func(limit int) *gorm.DB
	TransactionFn func(fn func(tx *gorm.DB) error) error
	CloseFn       func() error
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	if m.CreateFn != nil {
		return m.CreateFn(value)
	}
	return &gorm.DB{}
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	if m.FirstFn != nil {
		return m.FirstFn(dest, conds...)
	}
	return &gorm.DB{}
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	if m.FindFn != nil {
		return m.FindFn(dest, conds...)
	}
	return &gorm.DB{}
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	if m.WhereFn != nil {
		return m.WhereFn(query, args...)
	}
	return &gorm.DB{}
}

func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	if m.DeleteFn != nil {
		return m.DeleteFn(value, conds...)
	}
	return &gorm.DB{}
}

func (m *MockDB) Preload(query string, args ...interface{}) *gorm.DB {
	if m.PreloadFn != nil {
		return m.PreloadFn(query, args...)
	}
	return &gorm.DB{}
}

func (m *MockDB) Order(value interface{}) *gorm.DB {
	if m.OrderFn != nil {
		return m.OrderFn(value)
	}
	return &gorm.DB{}
}

func (m *MockDB) Limit(limit int) *gorm.DB {
	if m.LimitFn != nil {
		return m.LimitFn(limit)
	}
	return &gorm.DB{}
}

func (m *MockDB) Transaction(fn func(tx *gorm.DB) error) error {
	if m.TransactionFn != nil {
		return m.TransactionFn(fn)
	}
	return nil
}

func (m *MockDB) Close() error {
	if m.CloseFn != nil {
		return m.CloseFn()
	}
	return nil
}

func (m *MockDB) GetDB() *gorm.DB {
	return &gorm.DB{}
}

// Wallet and Private Key Test Fixtures
type WalletFixtures struct {
	TestDB *TestDB
}

func NewWalletFixtures(testDB *TestDB) *WalletFixtures {
	return &WalletFixtures{TestDB: testDB}
}

// CreateTestWallet creates a test wallet with specified parameters
func (wf *WalletFixtures) CreateTestWallet(orgID, name string, accountCount int) (*models.Wallet, []models.WalletAccount, error) {
	wallet := &models.Wallet{
		ID:             generateUUID(),
		OrganizationID: parseUUID(orgID),
		Name:           name,
		PublicKey:      "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
		Tags:           []string{"test"},
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := wf.TestDB.Create(wallet).Error; err != nil {
		return nil, nil, err
	}

	var accounts []models.WalletAccount
	for i := 0; i < accountCount; i++ {
		account := models.WalletAccount{
			ID:            generateUUID(),
			WalletID:      wallet.ID,
			Name:          fmt.Sprintf("Account %d", i),
			Path:          fmt.Sprintf("m/44'/60'/0'/0/%d", i),
			PublicKey:     fmt.Sprintf("03%064x", i+1),
			Address:       generateTestEthAddress(i),
			Curve:         "CURVE_SECP256K1",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := wf.TestDB.Create(&account).Error; err != nil {
			return nil, nil, err
		}

		accounts = append(accounts, account)
	}

	wallet.Accounts = accounts
	return wallet, accounts, nil
}

// CreateTestPrivateKey creates a test private key
func (wf *WalletFixtures) CreateTestPrivateKey(orgID, name, curve string, tags []string) (*models.PrivateKey, error) {
	privateKey := &models.PrivateKey{
		ID:             generateUUID(),
		OrganizationID: parseUUID(orgID),
		Name:           name,
		PublicKey:      fmt.Sprintf("pub_%s_%d", curve, time.Now().Unix()),
		Curve:          curve,
		Tags:           tags,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := wf.TestDB.Create(privateKey).Error; err != nil {
		return nil, err
	}

	return privateKey, nil
}

// Helper functions for test data generation
func generateTestEthAddress(index int) string {
	addresses := []string{
		"0x742D35CC6Bf8B8E0b8F8F8F8F8F8F8F8F8F8F8F8",
		"0x8ba1f109551bD432803012645Hac136c22C8C8C8",
		"0xdD2FD4581271e230360230F9337D5c0434068C8C",
		"0xDe30040413b26d7Aa2B6Fc4E5e5e5e5e5e5e5e5e",
		"0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
	}
	return addresses[index%len(addresses)]
}

func generateUUID() uuid.UUID {
	return uuid.New()
}

func parseUUID(s string) uuid.UUID {
	parsed, _ := uuid.Parse(s)
	return parsed
}

// Test data sets for comprehensive testing
var (
	TestCurves = []string{
		"CURVE_SECP256K1",
		"CURVE_ED25519",
	}

	TestAddressFormats = []string{
		"ADDRESS_FORMAT_ETHEREUM",
		"ADDRESS_FORMAT_BITCOIN_MAINNET_P2WPKH",
		"ADDRESS_FORMAT_SOLANA",
		"ADDRESS_FORMAT_COSMOS",
	}

	TestMnemonicLengths = []int32{12, 15, 18, 21, 24}

	InvalidCurves = []string{
		"CURVE_INVALID",
		"INVALID_CURVE",
		"",
		"SECP256K1", // Missing CURVE_ prefix
	}

	InvalidMnemonicLengths = []int32{11, 13, 16, 20, 25, 0, -1}
)

// BenchmarkHelper provides utilities for performance testing
type BenchmarkHelper struct {
	TestDB *TestDB
}

func NewBenchmarkHelper(testDB *TestDB) *BenchmarkHelper {
	return &BenchmarkHelper{TestDB: testDB}
}

// CreateBulkWallets creates multiple wallets for performance testing
func (bh *BenchmarkHelper) CreateBulkWallets(count int, accountsPerWallet int) error {
	for i := 0; i < count; i++ {
		orgID := generateUUID().String()

		// Create organization first
		org := &models.Organization{
			ID:      parseUUID(orgID),
			Name:    fmt.Sprintf("Bench Org %d", i),
			Version: "1.0",
		}

		if err := bh.TestDB.Create(org).Error; err != nil {
			return err
		}

		fixtures := NewWalletFixtures(bh.TestDB)
		_, _, err := fixtures.CreateTestWallet(orgID, fmt.Sprintf("Bench Wallet %d", i), accountsPerWallet)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateBulkPrivateKeys creates multiple private keys for performance testing
func (bh *BenchmarkHelper) CreateBulkPrivateKeys(count int) error {
	orgID := generateUUID().String()

	// Create organization first
	org := &models.Organization{
		ID:      parseUUID(orgID),
		Name:    "Bench Org for Keys",
		Version: "1.0",
	}

	if err := bh.TestDB.Create(org).Error; err != nil {
		return err
	}

	fixtures := NewWalletFixtures(bh.TestDB)
	for i := 0; i < count; i++ {
		curve := TestCurves[i%len(TestCurves)]
		_, err := fixtures.CreateTestPrivateKey(orgID, fmt.Sprintf("Bench Key %d", i), curve, []string{"benchmark"})
		if err != nil {
			return err
		}
	}
	return nil
}
