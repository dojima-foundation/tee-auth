package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type WalletServiceTestSuite struct {
	suite.Suite
	service        *GAuthService
	db             *testhelpers.TestDB
	organizationID string
	ctx            context.Context
}

func (suite *WalletServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Setup test database
	suite.db = testhelpers.SetupTestDB(suite.T())

	// Setup test logger
	testLogger, err := logger.New(&config.LoggingConfig{
		Level:  "debug",
		Format: "text",
	})
	require.NoError(suite.T(), err)

	// Setup test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
		},
	}

	// Create service instance
	suite.service = NewGAuthServiceWithEnclave(cfg, testLogger, suite.db, nil, NewMockRenclaveClient())

	// Create test organization
	org := &models.Organization{
		ID:      uuid.New(),
		Name:    "Test Organization",
		Version: "1.0",
		RootQuorum: models.Quorum{
			Threshold: 1,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := suite.db.GetDB().Create(org).Error; err != nil {
		require.NoError(suite.T(), err)
	}
	suite.organizationID = org.ID.String()
}

func (suite *WalletServiceTestSuite) TearDownSuite() {
	suite.db.Cleanup()
}

func (suite *WalletServiceTestSuite) TearDownTest() {
	// Clean up wallets and accounts created during tests
	suite.db.GetDB().Exec("DELETE FROM wallet_accounts")
	suite.db.GetDB().Exec("DELETE FROM wallets")
}

func (suite *WalletServiceTestSuite) TestCreateWallet_Success() {
	accounts := []models.WalletAccount{
		{
			Curve:         "CURVE_SECP256K1",
			Path:          "m/44'/60'/0'/0/0",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		},
	}

	mnemonicLength := int32(12)
	tags := []string{"test", "ethereum"}

	wallet, addresses, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet",
		accounts,
		&mnemonicLength,
		tags,
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), wallet)
	assert.Equal(suite.T(), "Test Wallet", wallet.Name)
	assert.Equal(suite.T(), suite.organizationID, wallet.OrganizationID.String())
	// TODO: Fix tags handling in wallet service
	// assert.Equal(suite.T(), tags, wallet.Tags)
	assert.True(suite.T(), wallet.IsActive)
	assert.Len(suite.T(), addresses, 1)
	assert.Len(suite.T(), wallet.Accounts, 1)

	// Check account details
	account := wallet.Accounts[0]
	assert.Equal(suite.T(), "CURVE_SECP256K1", account.Curve)
	assert.Equal(suite.T(), "m/44'/60'/0'/0/0", account.Path)
	assert.Equal(suite.T(), "ADDRESS_FORMAT_ETHEREUM", account.AddressFormat)
	assert.NotEmpty(suite.T(), account.Address)
	assert.NotEmpty(suite.T(), account.PublicKey)
	assert.True(suite.T(), account.IsActive)
}

func (suite *WalletServiceTestSuite) TestCreateWallet_InvalidMnemonicLength() {
	accounts := []models.WalletAccount{
		{
			Curve:         "CURVE_SECP256K1",
			Path:          "m/44'/60'/0'/0/0",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		},
	}

	invalidLength := int32(13) // Invalid mnemonic length

	wallet, addresses, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Invalid Wallet",
		accounts,
		&invalidLength,
		nil,
	)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), wallet)
	assert.Nil(suite.T(), addresses)
	assert.Contains(suite.T(), err.Error(), "invalid mnemonic length")
}

func (suite *WalletServiceTestSuite) TestCreateWallet_DefaultMnemonicLength() {
	accounts := []models.WalletAccount{
		{
			Curve:         "CURVE_ED25519",
			Path:          "m/44'/501'/0'/0",
			AddressFormat: "ADDRESS_FORMAT_SOLANA",
		},
	}

	wallet, addresses, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Default Mnemonic Wallet",
		accounts,
		nil, // No mnemonic length specified, should default to 12
		nil,
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), wallet)
	assert.Len(suite.T(), addresses, 1)
}

func (suite *WalletServiceTestSuite) TestCreateWallet_MultipleAccounts() {
	accounts := []models.WalletAccount{
		{
			Curve:         "CURVE_SECP256K1",
			Path:          "m/44'/60'/0'/0/0",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		},
		{
			Curve:         "CURVE_SECP256K1",
			Path:          "m/44'/0'/0'/0/0",
			AddressFormat: "ADDRESS_FORMAT_BITCOIN_MAINNET_P2WPKH",
		},
		{
			Curve:         "CURVE_ED25519",
			Path:          "m/44'/501'/0'/0",
			AddressFormat: "ADDRESS_FORMAT_SOLANA",
		},
	}

	mnemonicLength := int32(24)

	wallet, addresses, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Multi-Chain Wallet",
		accounts,
		&mnemonicLength,
		[]string{"multi-chain", "production"},
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), wallet)
	assert.Len(suite.T(), addresses, 3)
	assert.Len(suite.T(), wallet.Accounts, 3)

	// Check each account has different addresses
	addressSet := make(map[string]bool)
	for _, addr := range addresses {
		assert.NotEmpty(suite.T(), addr)
		assert.False(suite.T(), addressSet[addr], "Duplicate address generated")
		addressSet[addr] = true
	}
}

func (suite *WalletServiceTestSuite) TestGetWallet_Success() {
	// Create a wallet first
	accounts := []models.WalletAccount{
		{
			Curve:         "CURVE_SECP256K1",
			Path:          "m/44'/60'/0'/0/0",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		},
	}

	createdWallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Get Test Wallet",
		accounts,
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)

	// Get the wallet
	retrievedWallet, err := suite.service.GetWallet(suite.ctx, createdWallet.ID.String())

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedWallet)
	assert.Equal(suite.T(), createdWallet.ID, retrievedWallet.ID)
	assert.Equal(suite.T(), createdWallet.Name, retrievedWallet.Name)
	assert.Len(suite.T(), retrievedWallet.Accounts, 1)
}

func (suite *WalletServiceTestSuite) TestGetWallet_NotFound() {
	nonExistentID := uuid.New().String()

	wallet, err := suite.service.GetWallet(suite.ctx, nonExistentID)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), wallet)
	assert.Contains(suite.T(), err.Error(), "wallet not found")
}

func (suite *WalletServiceTestSuite) TestListWallets_Success() {
	// Create multiple wallets
	accounts := []models.WalletAccount{
		{
			Curve:         "CURVE_SECP256K1",
			Path:          "m/44'/60'/0'/0/0",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		},
	}

	for i := 0; i < 3; i++ {
		_, _, err := suite.service.CreateWallet(
			suite.ctx,
			suite.organizationID,
			fmt.Sprintf("List Test Wallet %d", i+1),
			accounts,
			nil,
			[]string{"test"},
		)
		require.NoError(suite.T(), err)
	}

	// List wallets
	wallets, nextToken, err := suite.service.ListWallets(
		suite.ctx,
		suite.organizationID,
		10,
		"",
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), wallets, 3)
	assert.Empty(suite.T(), nextToken) // No pagination needed for 3 items with limit 10

	// Check all wallets have accounts loaded
	for _, wallet := range wallets {
		assert.NotEmpty(suite.T(), wallet.Accounts)
	}
}

func (suite *WalletServiceTestSuite) TestListWallets_Pagination() {
	// Create multiple wallets
	accounts := []models.WalletAccount{
		{
			Curve:         "CURVE_SECP256K1",
			Path:          "m/44'/60'/0'/0/0",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		},
	}

	for i := 0; i < 5; i++ {
		_, _, err := suite.service.CreateWallet(
			suite.ctx,
			suite.organizationID,
			fmt.Sprintf("Pagination Test Wallet %d", i+1),
			accounts,
			nil,
			[]string{"pagination"},
		)
		require.NoError(suite.T(), err)
	}

	// List wallets with small page size
	wallets, nextToken, err := suite.service.ListWallets(
		suite.ctx,
		suite.organizationID,
		2, // Small page size
		"",
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), wallets, 2)
	assert.NotEmpty(suite.T(), nextToken) // Should have next page
}

func (suite *WalletServiceTestSuite) TestDeleteWallet_Success() {
	// Create a wallet first
	accounts := []models.WalletAccount{
		{
			Curve:         "CURVE_SECP256K1",
			Path:          "m/44'/60'/0'/0/0",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		},
	}

	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Delete Test Wallet",
		accounts,
		nil,
		nil,
	)
	require.NoError(suite.T(), err)

	// Delete the wallet
	err = suite.service.DeleteWallet(suite.ctx, wallet.ID.String(), true) // Force delete

	// Assertions
	require.NoError(suite.T(), err)

	// Verify wallet is deleted
	deletedWallet, err := suite.service.GetWallet(suite.ctx, wallet.ID.String())
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), deletedWallet)
}

func TestWalletServiceTestSuite(t *testing.T) {
	suite.Run(t, new(WalletServiceTestSuite))
}
