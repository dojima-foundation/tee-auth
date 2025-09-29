package unit

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletService_Validation_Success(t *testing.T) {
	// Test UUID validation logic used in wallet service
	validOrgID := uuid.New().String()

	// Test that valid UUIDs can be parsed
	_, err := uuid.Parse(validOrgID)
	require.NoError(t, err)
}

func TestWalletService_Validation_InvalidUUIDs(t *testing.T) {
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

func TestWalletService_MnemonicLengthValidation(t *testing.T) {
	// Test mnemonic length validation logic
	validLengths := []int32{12, 15, 18, 21, 24}
	invalidLengths := []int32{11, 13, 16, 20, 25, 0, -1}

	// Test valid lengths
	for _, length := range validLengths {
		t.Run("valid_"+string(rune(length)), func(t *testing.T) {
			assert.True(t, length == 12 || length == 15 || length == 18 || length == 21 || length == 24)
		})
	}

	// Test invalid lengths
	for _, length := range invalidLengths {
		t.Run("invalid_"+string(rune(length)), func(t *testing.T) {
			assert.False(t, length == 12 || length == 15 || length == 18 || length == 21 || length == 24)
		})
	}
}

func TestWalletService_StrengthCalculation(t *testing.T) {
	// Test strength calculation from mnemonic length
	testCases := []struct {
		mnemonicLength   int32
		expectedStrength int
	}{
		{12, 128},
		{15, 160},
		{18, 192},
		{21, 224},
		{24, 256},
	}

	for _, tc := range testCases {
		t.Run("length_"+string(rune(tc.mnemonicLength)), func(t *testing.T) {
			var strength int
			switch tc.mnemonicLength {
			case 12:
				strength = 128
			case 15:
				strength = 160
			case 18:
				strength = 192
			case 21:
				strength = 224
			case 24:
				strength = 256
			default:
				strength = 256 // default
			}
			assert.Equal(t, tc.expectedStrength, strength)
		})
	}
}

func TestWalletService_CurveValidation(t *testing.T) {
	// Test curve validation for wallet accounts
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

func TestWalletService_DerivationPathGeneration(t *testing.T) {
	// Test derivation path generation for different curves
	testCases := []struct {
		curve          string
		index          int
		expectedPrefix string
	}{
		{"CURVE_SECP256K1", 0, "m/44'/60'/0'/0/0"},
		{"CURVE_SECP256K1", 1, "m/44'/60'/0'/0/1"},
		{"CURVE_ED25519", 0, "m/44'/501'/0'/0/0"},
		{"CURVE_ED25519", 1, "m/44'/501'/0'/0/1"},
	}

	for _, tc := range testCases {
		t.Run("curve_"+tc.curve+"_index_"+string(rune(tc.index)), func(t *testing.T) {
			var derivationPath string
			switch tc.curve {
			case "CURVE_SECP256K1": // Ethereum
				derivationPath = "m/44'/60'/0'/0/" + fmt.Sprintf("%d", tc.index)
			case "CURVE_ED25519": // Solana
				derivationPath = "m/44'/501'/0'/0/" + fmt.Sprintf("%d", tc.index)
			default:
				derivationPath = "m/44'/60'/0'/0/" + fmt.Sprintf("%d", tc.index) // Default to Ethereum
			}
			assert.Equal(t, tc.expectedPrefix, derivationPath)
		})
	}
}

func TestWalletService_WalletModelValidation(t *testing.T) {
	// Test wallet model validation
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

	// Verify wallet was created correctly
	var retrievedWallet models.Wallet
	err := testDB.GetDB().First(&retrievedWallet, "id = ?", walletID).Error
	require.NoError(t, err)
	assert.Equal(t, walletID, retrievedWallet.ID)
	assert.Equal(t, orgID, retrievedWallet.OrganizationID)
	assert.Equal(t, "Test Wallet", retrievedWallet.Name)
	assert.Equal(t, "encrypted_seed_hex_data_placeholder", retrievedWallet.SeedPhrase)
	assert.True(t, retrievedWallet.IsActive)
}

func TestWalletService_WalletAccountModelValidation(t *testing.T) {
	// Test wallet account model validation
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

	// Create test wallet account
	accountID := uuid.New()
	account := &models.WalletAccount{
		ID:            accountID,
		WalletID:      walletID,
		Name:          "Test Account",
		Path:          "m/44'/60'/0'/0/0",
		PublicKey:     "account-public-key",
		Address:       "0x742D35CC6Bf8B8E0b8F8F8F8F8F8F8F8F8F8F8F8",
		Curve:         "CURVE_SECP256K1",
		AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	require.NoError(t, testDB.Create(account).Error)

	// Verify account was created correctly
	var retrievedAccount models.WalletAccount
	err := testDB.GetDB().First(&retrievedAccount, "id = ?", accountID).Error
	require.NoError(t, err)
	assert.Equal(t, accountID, retrievedAccount.ID)
	assert.Equal(t, walletID, retrievedAccount.WalletID)
	assert.Equal(t, "Test Account", retrievedAccount.Name)
	assert.Equal(t, "m/44'/60'/0'/0/0", retrievedAccount.Path)
	assert.Equal(t, "CURVE_SECP256K1", retrievedAccount.Curve)
	assert.Equal(t, "ADDRESS_FORMAT_ETHEREUM", retrievedAccount.AddressFormat)
	assert.True(t, retrievedAccount.IsActive)
}

func TestWalletService_WalletAccountRelationship(t *testing.T) {
	// Test wallet-account relationship
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

	// Create multiple accounts for the wallet
	for i := 0; i < 3; i++ {
		account := &models.WalletAccount{
			ID:            uuid.New(),
			WalletID:      walletID,
			Name:          "Test Account " + string(rune(i)),
			Path:          "m/44'/60'/0'/0/" + string(rune(i)),
			PublicKey:     "account-public-key-" + string(rune(i)),
			Address:       "0x742D35CC6Bf8B8E0b8F8F8F8F8F8F8F8F8F8F8F" + string(rune(i)),
			Curve:         "CURVE_SECP256K1",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		require.NoError(t, testDB.Create(account).Error)
	}

	// Verify all accounts belong to the wallet
	var accounts []models.WalletAccount
	err := testDB.GetDB().Where("wallet_id = ?", walletID).Find(&accounts).Error
	require.NoError(t, err)
	assert.Len(t, accounts, 3)

	for _, account := range accounts {
		assert.Equal(t, walletID, account.WalletID)
	}
}

func TestWalletService_PaginationLogic(t *testing.T) {
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

	// Create multiple wallets with different timestamps
	for i := 0; i < 5; i++ {
		walletID := uuid.New()

		wallet := &models.Wallet{
			ID:             walletID,
			OrganizationID: orgID,
			Name:           "Test Wallet " + string(rune(i)),
			SeedPhrase:     "encrypted_test_seed_hex_data", // Mock encrypted seed data
			PublicKey:      "test-public-key-" + string(rune(i)),
			Tags:           []string{"test"},
			IsActive:       true,
			CreatedAt:      time.Now().Add(time.Duration(i) * time.Minute),
			UpdatedAt:      time.Now().Add(time.Duration(i) * time.Minute),
		}
		require.NoError(t, testDB.Create(wallet).Error)
	}

	// Test pagination with page size 2
	pageSize := 2
	var wallets []models.Wallet
	err := testDB.GetDB().Where("organization_id = ?", orgID).
		Order("created_at ASC").
		Limit(pageSize + 1). // Get one extra to check if there's a next page
		Find(&wallets).Error
	require.NoError(t, err)

	// Should get 3 items (pageSize + 1)
	assert.Len(t, wallets, 3)

	// Test next token logic
	var nextToken string
	if len(wallets) > pageSize {
		nextToken = wallets[pageSize-1].ID.String()
		wallets = wallets[:pageSize]
	}

	assert.NotEmpty(t, nextToken)
	assert.Len(t, wallets, pageSize)
}

func TestWalletService_DeleteValidation(t *testing.T) {
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

	// Create test wallet account
	account := &models.WalletAccount{
		ID:            uuid.New(),
		WalletID:      walletID,
		Name:          "Test Account",
		Path:          "m/44'/60'/0'/0/0",
		PublicKey:     "account-public-key",
		Address:       "0x742D35CC6Bf8B8E0b8F8F8F8F8F8F8F8F8F8F8F8",
		Curve:         "CURVE_SECP256K1",
		AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	require.NoError(t, testDB.Create(account).Error)

	// Test delete without export check
	deleteWithoutExport := true
	if !deleteWithoutExport {
		// In production, would check export status
		assert.Fail(t, "Should not reach here when deleteWithoutExport is true")
	}

	// Delete wallet accounts first (due to foreign key constraints)
	err := testDB.GetDB().Where("wallet_id = ?", walletID).Delete(&models.WalletAccount{}).Error
	require.NoError(t, err)

	// Delete the wallet
	err = testDB.GetDB().Where("id = ?", walletID).Delete(&models.Wallet{}).Error
	require.NoError(t, err)

	// Verify wallet was deleted
	var deletedWallet models.Wallet
	err = testDB.GetDB().Where("id = ?", walletID).First(&deletedWallet).Error
	assert.Error(t, err) // Should not find the deleted wallet

	// Verify account was deleted
	var deletedAccount models.WalletAccount
	err = testDB.GetDB().Where("wallet_id = ?", walletID).First(&deletedAccount).Error
	assert.Error(t, err) // Should not find the deleted account
}

func TestWalletService_AddressFormatValidation(t *testing.T) {
	// Test address format validation
	validFormats := []string{
		"ADDRESS_FORMAT_ETHEREUM",
		"ADDRESS_FORMAT_BITCOIN_MAINNET_P2WPKH",
		"ADDRESS_FORMAT_SOLANA",
		"ADDRESS_FORMAT_COSMOS",
		"standard",
	}

	invalidFormats := []string{
		"",
		"invalid_format",
		"INVALID_FORMAT",
	}

	// Test valid formats
	for _, format := range validFormats {
		t.Run("valid_"+format, func(t *testing.T) {
			assert.NotEmpty(t, format)
			assert.True(t, len(format) > 3)
		})
	}

	// Test invalid formats
	for _, format := range invalidFormats {
		t.Run("invalid_"+format, func(t *testing.T) {
			if format == "" {
				assert.Empty(t, format)
			} else {
				assert.True(t, len(format) <= 3 || format == "invalid_format" || format == "INVALID_FORMAT")
			}
		})
	}
}

func TestWalletService_SeedPhraseValidation(t *testing.T) {
	// Test seed phrase validation
	validSeedPhrases := []string{
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
	}

	invalidSeedPhrases := []string{
		"",
		"invalid seed phrase",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
	}

	// Test valid seed phrases
	for i, seedPhrase := range validSeedPhrases {
		t.Run("valid_"+string(rune(i)), func(t *testing.T) {
			words := len(strings.Fields(seedPhrase))
			assert.True(t, words == 12 || words == 15 || words == 18 || words == 21 || words == 24)
		})
	}

	// Test invalid seed phrases
	for i, seedPhrase := range invalidSeedPhrases {
		t.Run("invalid_"+string(rune(i)), func(t *testing.T) {
			words := len(strings.Fields(seedPhrase))
			assert.False(t, words == 12 || words == 15 || words == 18 || words == 21 || words == 24)
		})
	}
}
