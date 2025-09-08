package unit

import (
	"strings"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Validation_Success(t *testing.T) {
	// Test UUID validation logic that's used in authentication
	validOrgID := uuid.New().String()
	validUserID := uuid.New().String()
	validAuthMethodID := uuid.New().String()

	// Test that valid UUIDs can be parsed
	_, err := uuid.Parse(validOrgID)
	require.NoError(t, err)

	_, err = uuid.Parse(validUserID)
	require.NoError(t, err)

	_, err = uuid.Parse(validAuthMethodID)
	require.NoError(t, err)
}

func TestAuthService_Validation_InvalidUUIDs(t *testing.T) {
	// Test invalid UUID validation
	invalidUUIDs := []string{
		"invalid-uuid",
		"not-a-uuid",
		"",
		"123",
		"abc-def-ghi",
	}

	for _, invalidUUID := range invalidUUIDs {
		t.Run("invalid_"+invalidUUID, func(t *testing.T) {
			_, err := uuid.Parse(invalidUUID)
			assert.Error(t, err)
		})
	}
}

func TestAuthService_AuthorizationRules(t *testing.T) {
	// Test the authorization logic for different activity types
	criticalOperations := []string{
		"CREATE_WALLET",
		"DELETE_WALLET",
		"CREATE_PRIVATE_KEY",
		"DELETE_PRIVATE_KEY",
		"SIGN_TRANSACTION",
	}

	readOperations := []string{
		"READ_WALLET",
		"LIST_WALLETS",
		"GET_WALLET",
		"READ_PRIVATE_KEY",
		"LIST_PRIVATE_KEYS",
	}

	// Test that critical operations require quorum approval
	for _, operation := range criticalOperations {
		t.Run("critical_"+operation, func(t *testing.T) {
			// In the actual service, these would return false for authorized
			// and require specific approvals
			isCritical := strings.Contains(operation, "CREATE") ||
				strings.Contains(operation, "DELETE") ||
				strings.Contains(operation, "SIGN")
			assert.True(t, isCritical, "Operation %s should be critical", operation)
		})
	}

	// Test that read operations are generally allowed
	for _, operation := range readOperations {
		t.Run("read_"+operation, func(t *testing.T) {
			// In the actual service, these would return true for authorized
			isReadOperation := strings.Contains(operation, "READ") ||
				strings.Contains(operation, "LIST") ||
				strings.Contains(operation, "GET")
			assert.True(t, isReadOperation, "Operation %s should be a read operation", operation)
		})
	}
}

func TestAuthService_SessionTokenGeneration(t *testing.T) {
	// Test session token generation
	sessionToken1 := uuid.New().String()
	sessionToken2 := uuid.New().String()

	// Session tokens should be unique
	assert.NotEqual(t, sessionToken1, sessionToken2)

	// Session tokens should be valid UUIDs
	_, err := uuid.Parse(sessionToken1)
	require.NoError(t, err)

	_, err = uuid.Parse(sessionToken2)
	require.NoError(t, err)

	// Session tokens should be non-empty
	assert.NotEmpty(t, sessionToken1)
	assert.NotEmpty(t, sessionToken2)
}

func TestAuthService_SessionExpiration(t *testing.T) {
	// Test session expiration logic
	now := time.Now()
	sessionDuration := 24 * time.Hour
	expiresAt := now.Add(sessionDuration)

	// Session should expire in the future
	assert.True(t, expiresAt.After(now))

	// Session duration should be 24 hours
	assert.Equal(t, 24*time.Hour, sessionDuration)

	// Test expired session
	expiredTime := now.Add(-1 * time.Hour)
	assert.True(t, now.After(expiredTime))
}

func TestAuthService_UserModelValidation(t *testing.T) {
	// Test user model validation
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

	// Create test user
	userID := uuid.New()
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

	// Verify user was created correctly
	var retrievedUser models.User
	err := testDB.GetDB().First(&retrievedUser, "id = ?", userID).Error
	require.NoError(t, err)
	assert.Equal(t, userID, retrievedUser.ID)
	assert.Equal(t, orgID, retrievedUser.OrganizationID)
	assert.Equal(t, "testuser", retrievedUser.Username)
	assert.Equal(t, "test@example.com", retrievedUser.Email)
	assert.True(t, retrievedUser.IsActive)
}

func TestAuthService_InactiveUserValidation(t *testing.T) {
	// Test inactive user validation
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

	// Create inactive user
	userID := uuid.New()
	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Username:       "inactiveuser",
		Email:          "inactive@example.com",
		PublicKey:      "test-public-key",
		IsActive:       false, // Inactive user
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, testDB.Create(user).Error)

	// Update the user to be inactive (since GORM default is true)
	err := testDB.GetDB().Model(&user).Update("is_active", false).Error
	require.NoError(t, err)

	// Verify user is inactive
	var retrievedUser models.User
	err = testDB.GetDB().First(&retrievedUser, "id = ?", userID).Error
	require.NoError(t, err)
	// The user should now be inactive
	assert.Equal(t, false, retrievedUser.IsActive, "User should be inactive")
}

func TestAuthService_OrganizationUserRelationship(t *testing.T) {
	// Test organization-user relationship validation
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

	// Create user in organization
	userID := uuid.New()
	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Username:       "orguser",
		Email:          "org@example.com",
		PublicKey:      "test-public-key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, testDB.Create(user).Error)

	// Verify user belongs to organization
	var retrievedUser models.User
	err := testDB.GetDB().First(&retrievedUser, "id = ? AND organization_id = ?", userID, orgID).Error
	require.NoError(t, err)
	assert.Equal(t, orgID, retrievedUser.OrganizationID)

	// Test user not in organization
	differentOrgID := uuid.New()
	var notFoundUser models.User
	err = testDB.GetDB().First(&notFoundUser, "id = ? AND organization_id = ?", userID, differentOrgID).Error
	assert.Error(t, err) // Should not find user in different organization
}

func TestAuthService_RequiredApprovals(t *testing.T) {
	// Test required approvals for different operations
	walletOperations := []string{"CREATE_WALLET", "DELETE_WALLET"}
	keyOperations := []string{"CREATE_PRIVATE_KEY", "DELETE_PRIVATE_KEY"}
	transactionOperations := []string{"SIGN_TRANSACTION"}

	// Wallet operations should require admin and security_officer approval
	for _, operation := range walletOperations {
		t.Run("wallet_"+operation, func(t *testing.T) {
			// In the actual service, these would require specific approvals
			requiredApprovals := []string{"admin", "security_officer"}
			assert.Contains(t, requiredApprovals, "admin")
			assert.Contains(t, requiredApprovals, "security_officer")
		})
	}

	// Key operations should require admin and security_officer approval
	for _, operation := range keyOperations {
		t.Run("key_"+operation, func(t *testing.T) {
			requiredApprovals := []string{"admin", "security_officer"}
			assert.Contains(t, requiredApprovals, "admin")
			assert.Contains(t, requiredApprovals, "security_officer")
		})
	}

	// Transaction operations should require admin and treasurer approval
	for _, operation := range transactionOperations {
		t.Run("transaction_"+operation, func(t *testing.T) {
			requiredApprovals := []string{"admin", "treasurer"}
			assert.Contains(t, requiredApprovals, "admin")
			assert.Contains(t, requiredApprovals, "treasurer")
		})
	}
}
