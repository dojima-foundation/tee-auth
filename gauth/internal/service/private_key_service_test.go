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
	suite.service = NewGAuthService(cfg, testLogger, suite.db, nil)

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
	tags := []string{"test", "ethereum"}

	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
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
	assert.Equal(suite.T(), "CURVE_SECP256K1", privateKey.Curve)
	assert.Equal(suite.T(), tags, privateKey.Tags)
	assert.True(suite.T(), privateKey.IsActive)
	assert.NotEmpty(suite.T(), privateKey.PublicKey)
	assert.Contains(suite.T(), privateKey.PublicKey, "pub_")
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_ED25519_Success() {
	tags := []string{"test", "solana"}

	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
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
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_WithMaterial() {
	keyMaterial := "test_private_key_material_12345"

	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
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
	assert.Contains(suite.T(), privateKey.PublicKey, "pub_from_material_")
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_InvalidCurve() {
	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		"Invalid Curve Key",
		"CURVE_INVALID",
		nil,
		nil,
	)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), privateKey)
	assert.Contains(suite.T(), err.Error(), "invalid curve")
	assert.Contains(suite.T(), err.Error(), "CURVE_SECP256K1")
	assert.Contains(suite.T(), err.Error(), "CURVE_ED25519")
}

func (suite *PrivateKeyServiceTestSuite) TestCreatePrivateKey_InvalidOrganizationID() {
	invalidOrgID := "invalid-uuid"

	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		invalidOrgID,
		"Test Key",
		"CURVE_SECP256K1",
		nil,
		nil,
	)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), privateKey)
}

func (suite *PrivateKeyServiceTestSuite) TestGetPrivateKey_Success() {
	// Create a private key first
	createdKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		"Get Test Key",
		"CURVE_SECP256K1",
		nil,
		[]string{"test"},
	)
	require.NoError(suite.T(), err)

	// Get the private key
	retrievedKey, err := suite.service.GetPrivateKey(suite.ctx, createdKey.ID.String())

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedKey)
	assert.Equal(suite.T(), createdKey.ID, retrievedKey.ID)
	assert.Equal(suite.T(), createdKey.Name, retrievedKey.Name)
	assert.Equal(suite.T(), createdKey.Curve, retrievedKey.Curve)
	assert.Equal(suite.T(), createdKey.PublicKey, retrievedKey.PublicKey)
}

func (suite *PrivateKeyServiceTestSuite) TestGetPrivateKey_NotFound() {
	nonExistentID := uuid.New().String()

	privateKey, err := suite.service.GetPrivateKey(suite.ctx, nonExistentID)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), privateKey)
	assert.Contains(suite.T(), err.Error(), "private key not found")
}

func (suite *PrivateKeyServiceTestSuite) TestListPrivateKeys_Success() {
	// Create multiple private keys
	curves := []string{"CURVE_SECP256K1", "CURVE_ED25519", "CURVE_SECP256K1"}

	for i, curve := range curves {
		_, err := suite.service.CreatePrivateKey(
			suite.ctx,
			suite.organizationID,
			fmt.Sprintf("List Test Key %d", i+1),
			curve,
			nil,
			[]string{"test"},
		)
		require.NoError(suite.T(), err)
	}

	// List private keys
	privateKeys, nextToken, err := suite.service.ListPrivateKeys(
		suite.ctx,
		suite.organizationID,
		10,
		"",
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), privateKeys, 3)
	assert.Empty(suite.T(), nextToken) // No pagination needed for 3 items with limit 10

	// Check curves are preserved
	curveCount := make(map[string]int)
	for _, pk := range privateKeys {
		curveCount[pk.Curve]++
	}
	assert.Equal(suite.T(), 2, curveCount["CURVE_SECP256K1"])
	assert.Equal(suite.T(), 1, curveCount["CURVE_ED25519"])
}

func (suite *PrivateKeyServiceTestSuite) TestListPrivateKeys_Pagination() {
	// Create multiple private keys
	for i := 0; i < 5; i++ {
		_, err := suite.service.CreatePrivateKey(
			suite.ctx,
			suite.organizationID,
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
	// Create a private key first
	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
		"Delete Test Key",
		"CURVE_SECP256K1",
		nil,
		nil,
	)
	require.NoError(suite.T(), err)

	// Delete the private key
	err = suite.service.DeletePrivateKey(suite.ctx, privateKey.ID.String(), true) // Force delete

	// Assertions
	require.NoError(suite.T(), err)

	// Verify private key is deleted
	deletedKey, err := suite.service.GetPrivateKey(suite.ctx, privateKey.ID.String())
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), deletedKey)
}

func (suite *PrivateKeyServiceTestSuite) TestDeletePrivateKey_ExportWarning() {
	// Create a private key first
	privateKey, err := suite.service.CreatePrivateKey(
		suite.ctx,
		suite.organizationID,
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
	deletedKey, err := suite.service.GetPrivateKey(suite.ctx, privateKey.ID.String())
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), deletedKey)
}

func TestPrivateKeyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PrivateKeyServiceTestSuite))
}
