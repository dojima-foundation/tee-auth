package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WalletService provides wallet management functionality
type WalletService struct {
	*GAuthService
}

// CreateWallet creates a new wallet with accounts using enclave for seed generation
func (s *WalletService) CreateWallet(ctx context.Context, organizationID, name string, accounts []models.WalletAccount, mnemonicLength *int32, tags []string) (*models.Wallet, []string, error) {
	s.logger.Info("Creating wallet", "organization_id", organizationID, "name", name)

	// Validate organization ID
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	// Generate seed using enclave
	strength := 256 // default to 256-bit
	if mnemonicLength != nil {
		switch *mnemonicLength {
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
			return nil, nil, fmt.Errorf("invalid mnemonic length: %d. Must be 12, 15, 18, 21, or 24", *mnemonicLength)
		}
	}

	// Request seed generation from enclave
	seedResp, err := s.renclave.GenerateSeed(ctx, strength, nil)
	if err != nil {
		s.logger.Error("Failed to generate seed from enclave", "error", err)
		return nil, nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	s.logger.Info("Seed generated successfully", "strength", seedResp.Strength, "word_count", seedResp.WordCount)

	// Create wallet with the generated seed phrase
	wallet := &models.Wallet{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		SeedPhrase:     seedResp.SeedPhrase, // Store the seed phrase
		PublicKey:      "",                  // Will be set after deriving the first account
		Tags:           tags,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Generate addresses for each account using the enclave
	addresses := make([]string, len(accounts))

	for i := range accounts {
		accounts[i].ID = uuid.New()
		accounts[i].WalletID = wallet.ID
		accounts[i].IsActive = true
		accounts[i].CreatedAt = time.Now()
		accounts[i].UpdatedAt = time.Now()

		// Derive address using enclave
		// Use standard BIP44 paths: m/44'/60'/0'/0/{index} for Ethereum, m/44'/501'/0'/0/{index} for Solana
		var derivationPath string
		switch accounts[i].Curve {
		case "CURVE_SECP256K1": // Ethereum
			derivationPath = fmt.Sprintf("m/44'/60'/0'/0/%d", i)
		case "CURVE_ED25519": // Solana
			derivationPath = fmt.Sprintf("m/44'/501'/0'/0/%d", i)
		default:
			derivationPath = fmt.Sprintf("m/44'/60'/0'/0/%d", i) // Default to Ethereum
		}

		// Derive address from seed using enclave
		addressResp, err := s.renclave.DeriveAddress(ctx, seedResp.SeedPhrase, derivationPath, accounts[i].Curve)
		if err != nil {
			s.logger.Error("Failed to derive address from enclave", "error", err, "path", derivationPath)
			return nil, nil, fmt.Errorf("failed to derive address: %w", err)
		}

		accounts[i].Path = derivationPath
		accounts[i].Address = addressResp.Address
		accounts[i].PublicKey = addressResp.Address // For now, use address as public key
		// Keep the original AddressFormat from the input
		if accounts[i].AddressFormat == "" {
			accounts[i].AddressFormat = "standard"
		}
		addresses[i] = addressResp.Address

		// Set wallet's public key to the first account's public key
		if i == 0 {
			wallet.PublicKey = addressResp.Address
		}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create wallet
		if err := tx.Create(wallet).Error; err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}

		// Create accounts
		for i := range accounts {
			if err := tx.Create(&accounts[i]).Error; err != nil {
				return fmt.Errorf("failed to create wallet account: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create wallet", "error", err)
		return nil, nil, err
	}

	// Reload the wallet with accounts
	var reloadedWallet models.Wallet
	if err := s.db.GetDB().WithContext(ctx).Preload("Accounts").First(&reloadedWallet, "id = ?", wallet.ID).Error; err != nil {
		s.logger.Error("Failed to reload wallet with accounts", "error", err)
		return nil, nil, fmt.Errorf("failed to reload wallet: %w", err)
	}

	s.logger.Info("Wallet created successfully", "wallet_id", wallet.ID.String(), "name", name)

	// Record metrics
	telemetry.RecordWalletCreated()

	return &reloadedWallet, addresses, nil
}

// GetWallet retrieves a wallet by ID
func (s *WalletService) GetWallet(ctx context.Context, walletID string) (*models.Wallet, error) {
	walletUUID, err := uuid.Parse(walletID)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet ID: %w", err)
	}

	var wallet models.Wallet
	if err := s.db.GetDB().WithContext(ctx).Preload("Accounts").First(&wallet, "id = ?", walletUUID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

// ListWallets lists wallets in an organization with pagination
func (s *WalletService) ListWallets(ctx context.Context, organizationID string, pageSize int, pageToken string) ([]models.Wallet, string, error) {
	s.logger.Debug("Listing wallets", "organization_id", organizationID)

	var wallets []models.Wallet
	query := s.db.GetDB().WithContext(ctx).Preload("Accounts").Where("organization_id = ?", organizationID)

	// Handle pagination
	if pageToken != "" {
		// Validate pageToken as UUID
		if _, err := uuid.Parse(pageToken); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		query = query.Where("id > ?", pageToken)
	}

	query = query.Order("created_at ASC").Limit(pageSize + 1) // Get one extra to check if there's a next page

	if err := query.Find(&wallets).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list wallets: %w", err)
	}

	var nextToken string
	if len(wallets) > pageSize {
		// There are more items available, set nextToken to the last item of current page
		nextToken = wallets[pageSize-1].ID.String()
		wallets = wallets[:pageSize]
	}
	// If len(wallets) <= pageSize, nextToken remains empty string (no more pages)

	return wallets, nextToken, nil
}

// DeleteWallet deletes a wallet
func (s *WalletService) DeleteWallet(ctx context.Context, walletID string, deleteWithoutExport bool) error {
	s.logger.Info("Deleting wallet", "wallet_id", walletID, "delete_without_export", deleteWithoutExport)

	// For production: Check if wallet has been exported unless deleteWithoutExport is true
	if !deleteWithoutExport {
		// Check export status - for now, we'll allow deletion
		s.logger.Warn("Wallet deletion without export check - implement export verification in production")
	}

	// Delete wallet accounts first (due to foreign key constraints)
	if err := s.db.GetDB().WithContext(ctx).Where("wallet_id = ?", walletID).Delete(&models.WalletAccount{}).Error; err != nil {
		return fmt.Errorf("failed to delete wallet accounts: %w", err)
	}

	// Delete the wallet
	if err := s.db.GetDB().WithContext(ctx).Where("id = ?", walletID).Delete(&models.Wallet{}).Error; err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	s.logger.Info("Wallet deleted successfully", "wallet_id", walletID)
	return nil
}
