package service

import (
	"context"
	"testing"

	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OptionalFieldsTestSuite struct {
	suite.Suite
	service *GAuthService
	db      *testhelpers.TestDB
	ctx     context.Context
}

func (suite *OptionalFieldsTestSuite) SetupSuite() {
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
}

func (suite *OptionalFieldsTestSuite) TearDownSuite() {
	suite.db.Cleanup()
}

func (suite *OptionalFieldsTestSuite) TearDownTest() {
	// Clean up created organizations and users
	suite.db.GetDB().Exec("DELETE FROM users")
	suite.db.GetDB().Exec("DELETE FROM organizations")
}

func (suite *OptionalFieldsTestSuite) TestCreateOrganization_WithPublicKey() {
	publicKey := "0x1234567890abcdef1234567890abcdef12345678"

	org, err := suite.service.CreateOrganization(
		suite.ctx,
		"Test Org With Key",
		"admin@example.com",
		publicKey,
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), org)
	assert.Equal(suite.T(), "Test Org With Key", org.Name)
	assert.Len(suite.T(), org.Users, 1)

	// Check initial user has public key
	initialUser := org.Users[0]
	assert.Equal(suite.T(), "admin@example.com", initialUser.Email)
	assert.Equal(suite.T(), "admin", initialUser.Username)
	assert.Equal(suite.T(), publicKey, initialUser.PublicKey)
	assert.True(suite.T(), initialUser.IsActive)
}

func (suite *OptionalFieldsTestSuite) TestCreateOrganization_WithoutPublicKey() {
	org, err := suite.service.CreateOrganization(
		suite.ctx,
		"Test Org Without Key",
		"admin@example.com",
		"", // Empty public key
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), org)
	assert.Equal(suite.T(), "Test Org Without Key", org.Name)
	assert.Len(suite.T(), org.Users, 1)

	// Check initial user has no public key
	initialUser := org.Users[0]
	assert.Equal(suite.T(), "admin@example.com", initialUser.Email)
	assert.Equal(suite.T(), "admin", initialUser.Username)
	assert.Empty(suite.T(), initialUser.PublicKey) // Should be empty
	assert.True(suite.T(), initialUser.IsActive)
}

func (suite *OptionalFieldsTestSuite) TestCreateUser_WithPublicKey() {
	// Create organization first
	org, err := suite.service.CreateOrganization(
		suite.ctx,
		"Test Org for User",
		"admin@example.com",
		"",
	)
	require.NoError(suite.T(), err)

	publicKey := "0xabcdef1234567890abcdef1234567890abcdef12"

	user, err := suite.service.CreateUser(
		suite.ctx,
		org.ID.String(),
		"testuser",
		"user@example.com",
		publicKey,
		[]string{"developer"},
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "testuser", user.Username)
	assert.Equal(suite.T(), "user@example.com", user.Email)
	assert.Equal(suite.T(), publicKey, user.PublicKey)
	assert.Equal(suite.T(), []string{"developer"}, user.Tags)
	assert.True(suite.T(), user.IsActive)
}

func (suite *OptionalFieldsTestSuite) TestCreateUser_WithoutPublicKey() {
	// Create organization first
	org, err := suite.service.CreateOrganization(
		suite.ctx,
		"Test Org for User Without Key",
		"admin@example.com",
		"",
	)
	require.NoError(suite.T(), err)

	user, err := suite.service.CreateUser(
		suite.ctx,
		org.ID.String(),
		"testusernokey",
		"usernokey@example.com",
		"", // Empty public key
		[]string{"reviewer"},
	)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "testusernokey", user.Username)
	assert.Equal(suite.T(), "usernokey@example.com", user.Email)
	assert.Empty(suite.T(), user.PublicKey) // Should be empty
	assert.Equal(suite.T(), []string{"reviewer"}, user.Tags)
	assert.True(suite.T(), user.IsActive)
}

func (suite *OptionalFieldsTestSuite) TestCreateUser_InvalidOrganizationID() {
	invalidOrgID := "not-a-valid-uuid"

	user, err := suite.service.CreateUser(
		suite.ctx,
		invalidOrgID,
		"testuser",
		"user@example.com",
		"some-key",
		nil,
	)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Contains(suite.T(), err.Error(), "invalid organization ID")
}

func (suite *OptionalFieldsTestSuite) TestOrganizationBackwardCompatibility() {
	// Test that existing behavior still works with non-empty public keys
	publicKey := "0xfedcba0987654321fedcba0987654321fedcba09"

	org, err := suite.service.CreateOrganization(
		suite.ctx,
		"Backward Compat Org",
		"legacy@example.com",
		publicKey,
	)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), org)
	assert.Len(suite.T(), org.Users, 1)
	assert.Equal(suite.T(), publicKey, org.Users[0].PublicKey)
}

func (suite *OptionalFieldsTestSuite) TestUserBackwardCompatibility() {
	// Create organization first
	org, err := suite.service.CreateOrganization(
		suite.ctx,
		"User Compat Org",
		"admin@example.com",
		"admin-key",
	)
	require.NoError(suite.T(), err)

	// Test that existing behavior still works with non-empty public keys
	publicKey := "0x1111222233334444555566667777888899990000"

	user, err := suite.service.CreateUser(
		suite.ctx,
		org.ID.String(),
		"legacyuser",
		"legacy@example.com",
		publicKey,
		[]string{"legacy"},
	)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), publicKey, user.PublicKey)
}

func TestOptionalFieldsTestSuite(t *testing.T) {
	suite.Run(t, new(OptionalFieldsTestSuite))
}
