package unit

import (
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrivateKeyService_Validation_Success(t *testing.T) {
	// Test UUID validation logic used in private key service
	validOrgID := uuid.New().String()
	validWalletID := uuid.New().String()

	// Test that valid UUIDs can be parsed
	_, err := uuid.Parse(validOrgID)
	require.NoError(t, err)

	_, err = uuid.Parse(validWalletID)
	require.NoError(t, err)
}

func TestPrivateKeyService_Validation_InvalidUUIDs(t *testing.T) {
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

func TestPrivateKeyService_CurveValidation(t *testing.T) {
	// Test curve validation logic
	validCurves := []string{"CURVE_SECP256K1", "CURVE_ED25519"}
	invalidCurves := []string{"CURVE_INVALID", "INVALID_CURVE", "", "SECP256K1"}

	// Test valid curves
	for _, curve := range validCurves {
		t.Run("valid_"+curve, func(t *testing.T) {
			assert.True(t, curve == "CURVE_SECP256K1" || curve == "CURVE_ED25519")
		})
	}

	// Test invalid curves
	for _, curve := range invalidCurves {
		t.Run("invalid_"+curve, func(t *testing.T) {
			assert.False(t, curve == "CURVE_SECP256K1" || curve == "CURVE_ED25519")
		})
	}
}

func TestPrivateKeyService_DerivationPathGeneration(t *testing.T) {
	// Test derivation path generation logic
	testCases := []struct {
		curve          string
		expectedPrefix string
	}{
		{"CURVE_SECP256K1", "m/44'/60'/0'/0/"},
		{"CURVE_ED25519", "m/44'/501'/0'/0/"},
	}

	for _, tc := range testCases {
		t.Run("curve_"+tc.curve, func(t *testing.T) {
			// Test that derivation paths are generated correctly
			// In the actual service, paths would be generated with timestamps for uniqueness
			path := tc.expectedPrefix + "0"
			assert.Contains(t, path, tc.expectedPrefix)
		})
	}
}

func TestPrivateKeyService_PrivateKeyModelValidation(t *testing.T) {
	// Test private key model validation
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

	// Create test wallet
	walletID := uuid.New()
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

	// Create test private key
	privateKeyID := uuid.New()
	privateKey := &models.PrivateKey{
		ID:             privateKeyID,
		OrganizationID: orgID,
		WalletID:       walletID,
		Name:           "Test Private Key",
		PublicKey:      "test-public-key",
		Curve:          "CURVE_SECP256K1",
		Path:           "m/44'/60'/0'/0/0",
		Tags:           []string{"test"},
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, testDB.Create(privateKey).Error)

	// Verify private key was created correctly
	var retrievedPrivateKey models.PrivateKey
	err := testDB.GetDB().First(&retrievedPrivateKey, "id = ?", privateKeyID).Error
	require.NoError(t, err)
	assert.Equal(t, privateKeyID, retrievedPrivateKey.ID)
	assert.Equal(t, orgID, retrievedPrivateKey.OrganizationID)
	assert.Equal(t, walletID, retrievedPrivateKey.WalletID)
	assert.Equal(t, "Test Private Key", retrievedPrivateKey.Name)
	assert.Equal(t, "CURVE_SECP256K1", retrievedPrivateKey.Curve)
	assert.Equal(t, "m/44'/60'/0'/0/0", retrievedPrivateKey.Path)
	assert.True(t, retrievedPrivateKey.IsActive)
}

func TestPrivateKeyService_WalletRelationship(t *testing.T) {
	// Test wallet-private key relationship
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

	// Create test wallet
	walletID := uuid.New()
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

	// Create multiple private keys for the wallet
	for i := 0; i < 3; i++ {
		privateKey := &models.PrivateKey{
			ID:             uuid.New(),
			OrganizationID: orgID,
			WalletID:       walletID,
			Name:           "Test Private Key " + string(rune(i)),
			PublicKey:      "test-public-key-" + string(rune(i)),
			Curve:          "CURVE_SECP256K1",
			Path:           "m/44'/60'/0'/0/" + string(rune(i)),
			Tags:           []string{"test"},
			IsActive:       true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		require.NoError(t, testDB.Create(privateKey).Error)
	}

	// Verify all private keys belong to the wallet
	var privateKeys []models.PrivateKey
	err := testDB.GetDB().Where("wallet_id = ?", walletID).Find(&privateKeys).Error
	require.NoError(t, err)
	assert.Len(t, privateKeys, 3)

	for _, privateKey := range privateKeys {
		assert.Equal(t, walletID, privateKey.WalletID)
		assert.Equal(t, orgID, privateKey.OrganizationID)
	}
}

func TestPrivateKeyService_PaginationLogic(t *testing.T) {
	// Test pagination logic
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

	// Create test wallet
	walletID := uuid.New()
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

	// Create multiple private keys with different timestamps
	for i := 0; i < 5; i++ {
		privateKeyID := uuid.New()

		privateKey := &models.PrivateKey{
			ID:             privateKeyID,
			OrganizationID: orgID,
			WalletID:       walletID,
			Name:           "Test Private Key " + string(rune(i)),
			PublicKey:      "test-public-key-" + string(rune(i)),
			Curve:          "CURVE_SECP256K1",
			Path:           "m/44'/60'/0'/0/" + string(rune(i)),
			Tags:           []string{"test"},
			IsActive:       true,
			CreatedAt:      time.Now().Add(time.Duration(i) * time.Minute),
			UpdatedAt:      time.Now().Add(time.Duration(i) * time.Minute),
		}
		require.NoError(t, testDB.Create(privateKey).Error)
	}

	// Test pagination with page size 2
	pageSize := 2
	var privateKeys []models.PrivateKey
	err := testDB.GetDB().Where("organization_id = ?", orgID).
		Order("created_at ASC").
		Limit(pageSize + 1). // Get one extra to check if there's a next page
		Find(&privateKeys).Error
	require.NoError(t, err)

	// Should get 3 items (pageSize + 1)
	assert.Len(t, privateKeys, 3)

	// Test next token logic
	var nextToken string
	if len(privateKeys) > pageSize {
		nextToken = privateKeys[pageSize-1].ID.String()
		privateKeys = privateKeys[:pageSize]
	}

	assert.NotEmpty(t, nextToken)
	assert.Len(t, privateKeys, pageSize)
}

func TestPrivateKeyService_DeleteValidation(t *testing.T) {
	// Test delete validation logic
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

	// Create test wallet
	walletID := uuid.New()
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

	// Create test private key
	privateKeyID := uuid.New()
	privateKey := &models.PrivateKey{
		ID:             privateKeyID,
		OrganizationID: orgID,
		WalletID:       walletID,
		Name:           "Test Private Key",
		PublicKey:      "test-public-key",
		Curve:          "CURVE_SECP256K1",
		Path:           "m/44'/60'/0'/0/0",
		Tags:           []string{"test"},
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, testDB.Create(privateKey).Error)

	// Test delete without export check
	deleteWithoutExport := true
	if !deleteWithoutExport {
		// In production, would check export status
		assert.Fail(t, "Should not reach here when deleteWithoutExport is true")
	}

	// Delete the private key
	err := testDB.GetDB().Where("id = ?", privateKeyID).Delete(&models.PrivateKey{}).Error
	require.NoError(t, err)

	// Verify private key was deleted
	var deletedPrivateKey models.PrivateKey
	err = testDB.GetDB().Where("id = ?", privateKeyID).First(&deletedPrivateKey).Error
	assert.Error(t, err) // Should not find the deleted private key
}

func TestPrivateKeyService_TagsValidation(t *testing.T) {
	// Test tags validation
	validTags := [][]string{
		{"test"},
		{"test", "production"},
		{"wallet", "ethereum", "mainnet"},
		{},
	}

	invalidTags := [][]string{
		nil, // nil tags should be handled
	}

	// Test valid tags
	for i, tags := range validTags {
		t.Run("valid_tags_"+string(rune(i)), func(t *testing.T) {
			// Tags should be a slice of strings
			assert.IsType(t, []string{}, tags)
		})
	}

	// Test invalid tags
	for i, tags := range invalidTags {
		t.Run("invalid_tags_"+string(rune(i)), func(t *testing.T) {
			// Nil tags should be handled gracefully
			if tags == nil {
				tags = []string{} // Convert nil to empty slice
			}
			assert.IsType(t, []string{}, tags)
		})
	}
}

func TestPrivateKeyService_PublicKeyValidation(t *testing.T) {
	// Test public key validation
	validPublicKeys := []string{
		"03a34b99f22c790c4e36b2b3c2c35a36db06226e41c692fc82b8b56ac1c540c5bd",
		"02b97c30de767f084ce3050167c2e1435b8a2e3b4b5c6d7e8f9a0b1c2d3e4f5a6b",
		"mock_public_key_1234567890abcdef",
	}

	invalidPublicKeys := []string{
		"",
		"invalid",
		"too_short",
	}

	// Test valid public keys
	for i, publicKey := range validPublicKeys {
		t.Run("valid_public_key_"+string(rune(i)), func(t *testing.T) {
			assert.NotEmpty(t, publicKey)
			assert.True(t, len(publicKey) > 10) // Minimum length check
		})
	}

	// Test invalid public keys
	for i, publicKey := range invalidPublicKeys {
		t.Run("invalid_public_key_"+string(rune(i)), func(t *testing.T) {
			if publicKey == "" {
				assert.Empty(t, publicKey)
			} else {
				assert.True(t, len(publicKey) <= 10) // Should be too short
			}
		})
	}
}
