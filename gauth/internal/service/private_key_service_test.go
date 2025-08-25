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

type PrivateKeyServiceTestSuite struct {
	suite.Suite
	service        *GAuthService
	db             *testhelpers.TestDB
	organizationID string
	ctx            context.Context
}

func (suite *PrivateKeyServiceTestSuite) SetupSuite() {
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

func (suite *PrivateKeyServiceTestSuite) TearDownSuite() {
	suite.db.Cleanup()
}

func (suite *PrivateKeyServiceTestSuite) TearDownTest() {
	// Clean up private keys created during tests
	suite.db.GetDB().Exec("DELETE FROM private_keys")
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_SECP256K1_Success() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	tags := []string{"test", "ethereum"}

	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		wallet.ID.String(),
		"Test SECP256K1 Key",
		"CURVE_SECP256K1",
		nil, // Generate new key
		tags,
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), privateKey)
	assert.Equal(suite.T(), "Test SECP256K1 Key", privateKey.Name)
	assert.Equal(suite.T(), suite.organizationID, privateKey.OrganizationID.String())
	assert.Equal(suite.T(), wallet.ID, privateKey.WalletID)
	assert.Equal(suite.T(), "CURVE_SECP256K1", privateKey.Curve)
	assert.Equal(suite.T(), tags, privateKey.Tags)
	assert.True(suite.T(), privateKey.IsActive)
	assert.NotEmpty(suite.T(), privateKey.PublicKey)
	assert.NotEmpty(suite.T(), privateKey.Path)
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_ED25519_Success() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet ED25519",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	tags := []string{"test", "solana"}

	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		wallet.ID.String(),
		"Test ED25519 Key",
		"CURVE_ED25519",
		nil, // Generate new key
		tags,
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), privateKey)
	assert.Equal(suite.T(), "Test ED25519 Key", privateKey.Name)
	assert.Equal(suite.T(), "CURVE_ED25519", privateKey.Curve)
	assert.Equal(suite.T(), tags, privateKey.Tags)
	assert.NotEmpty(suite.T(), privateKey.PublicKey)
	assert.NotEmpty(suite.T(), privateKey.Path)
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_WithMaterial() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet Material",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	keyMaterial := "test_private_key_material_12345"

	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		wallet.ID.String(),
		"Imported Key",
		"CURVE_SECP256K1",
		&keyMaterial,
		[]string{"imported"},
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), privateKey)
	assert.Equal(suite.T(), "Imported Key", privateKey.Name)
	assert.NotEmpty(suite.T(), privateKey.PublicKey)
	assert.NotEmpty(suite.T(), privateKey.Path)
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_InvalidCurve() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet Invalid",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	_, err = suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		wallet.ID.String(),
		"Invalid Key",
		"INVALID_CURVE",
		nil,
		[]string{"test"},
	)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid curve")
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_InvalidOrganizationID() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet Invalid Org",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	_, err = suite.service.CreatePrivateKey(
		suite.ctx,
		"invalid-uuid",
		wallet.ID.String(),
		"Invalid Org Key",
		"CURVE_SECP256K1",
		nil,
		[]string{"test"},
	)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid organization ID")
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_InvalidWalletID() {
	_, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		"invalid-uuid",
		"Invalid Wallet Key",
		"CURVE_SECP256K1",
		nil,
		[]string{"test"},
	)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid wallet ID")
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_WalletNotFound() {
	_, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		uuid.New().String(), // Non-existent wallet ID
		"Non-existent Wallet Key",
		"CURVE_SECP256K1",
		nil,
		[]string{"test"},
	)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "wallet not found")
}

func (suite *PrivateKeyServiceTestSuite) TestGetPrivateKey_Success() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet Get",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	// Create a private key
	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		wallet.ID.String(),
		"Test Get Key",
		"CURVE_SECP256K1",
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), privateKey)

	// Get the private key
	retrievedKey, err := suite.service.GetPrivateKey(suite.ctx, privateKey.ID.String())

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedKey)
	assert.Equal(suite.T(), privateKey.ID, retrievedKey.ID)
	assert.Equal(suite.T(), privateKey.Name, retrievedKey.Name)
}

func (suite *PrivateKeyServiceTestSuite) TestGetPrivateKey_NotFound() {
	_, err := suite.service.GetPrivateKey(suite.ctx, uuid.New().String())

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "private key not found")
}

func (suite *PrivateKeyServiceTestSuite) TestListPrivateKeys_Success() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet List",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	// Create multiple private keys
	for i := 0; i < 3; i++ {
		_, err := suite.service.CreatePrivateKey(
			suite.ctx,
			suite.organizationID,
			wallet.ID.String(),
			fmt.Sprintf("Test Key %d", i),
			"CURVE_SECP256K1",
			nil,
			[]string{"test"},
		)
		require.NoError(suite.T(), err)
	}

	// List private keys
	privateKeys, nextToken, err := suite.service.ListPrivateKeys(suite.ctx, suite.organizationID, 10, "")

	// Assertions
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), privateKeys, 3)
	assert.Empty(suite.T(), nextToken)
}

func (suite *PrivateKeyServiceTestSuite) TestListPrivateKeys_Pagination() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet Pagination",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	// Create multiple private keys
	for i := 0; i < 5; i++ {
		_, err := suite.service.CreatePrivateKey(
			suite.ctx,
			suite.organizationID,
			wallet.ID.String(),
			fmt.Sprintf("Pagination Test Key %d", i+1),
			"CURVE_SECP256K1",
			nil,
			[]string{"pagination"},
		)
		require.NoError(suite.T(), err)
	}

	// List private keys with small page size
	privateKeys, nextToken, err := suite.service.ListPrivateKeys(
		suite.ctx,
		suite.organizationID,
		2, // Small page size
		"",
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), privateKeys, 2)
	assert.NotEmpty(suite.T(), nextToken) // Should have next page
}

func (suite *PrivateKeyServiceTestSuite) TestDeletePrivateKey_Success() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet Delete",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	// Create a private key
	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		wallet.ID.String(),
		"Test Delete Key",
		"CURVE_SECP256K1",
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), privateKey)

	// Delete the private key
	err = suite.service.DeletePrivateKey(suite.ctx, privateKey.ID.String(), true)

	// Assertions
	require.NoError(suite.T(), err)

	// Verify it's deleted
	_, err = suite.service.GetPrivateKey(suite.ctx, privateKey.ID.String())
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "private key not found")
}

func (suite *PrivateKeyServiceTestSuite) TestDeletePrivateKey_ExportWarning() {
	// First create a test wallet
	wallet, _, err := suite.service.CreateWallet(
		suite.ctx,
		suite.organizationID,
		"Test Wallet Export Warning",
		[]models.WalletAccount{},
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), wallet)

	// Create a private key
	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		wallet.ID.String(),
		"Export Warning Test Key",
		"CURVE_SECP256K1",
		nil,
		nil,
	)
	require.NoError(suite.T(), err)

	// Delete without export (should warn but still proceed in test)
	err = suite.service.DeletePrivateKey(suite.ctx, privateKey.ID.String(), false)

	// Assertions
	require.NoError(suite.T(), err) // Should succeed but log warning

	// Verify private key is deleted
	_, err = suite.service.GetPrivateKey(suite.ctx, privateKey.ID.String())
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "private key not found")
}

func TestPrivateKeyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PrivateKeyServiceTestSuite))
}
