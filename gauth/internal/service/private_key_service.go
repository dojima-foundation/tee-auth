package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PrivateKeyService provides private key management functionality
type PrivateKeyService struct {
	*GAuthService
}

// CreatePrivateKey creates a new private key using enclave for key generation
func (s *PrivateKeyService) CreatePrivateKey(ctx context.Context, organizationID, walletID, name, curve string, privateKeyMaterial *string, tags []string) (*models.PrivateKey, error) {
	s.logger.Info("Creating private key", "organization_id", organizationID, "wallet_id", walletID, "name", name, "curve", curve)

	// Validate organization ID
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	// Validate wallet ID
	walletUUID, err := uuid.Parse(walletID)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet ID: %w", err)
	}

	// Validate curve
	validCurves := []string{"CURVE_SECP256K1", "CURVE_ED25519"}
	valid := false
	for _, vc := range validCurves {
		if curve == vc {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("invalid curve: %s. Must be one of: CURVE_SECP256K1, CURVE_ED25519", curve)
	}

	// Get the wallet to access its seed phrase
	var wallet models.Wallet
	if err := s.db.GetDB().WithContext(ctx).Where("id = ? AND organization_id = ?", walletUUID, orgID).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("wallet not found or does not belong to organization")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Determine derivation path based on curve
	var derivationPath string
	switch curve {
	case "CURVE_SECP256K1": // Ethereum
		derivationPath = fmt.Sprintf("m/44'/60'/0'/0/%d", time.Now().UnixNano()%1000) // Use timestamp for uniqueness
	case "CURVE_ED25519": // Solana
		derivationPath = fmt.Sprintf("m/44'/501'/0'/0/%d", time.Now().UnixNano()%1000)
	default:
		derivationPath = fmt.Sprintf("m/44'/60'/0'/0/%d", time.Now().UnixNano()%1000)
	}

	// Derive private key from wallet's seed phrase using enclave
	keyResp, err := s.renclave.DeriveKey(ctx, wallet.SeedPhrase, derivationPath, curve)
	if err != nil {
		s.logger.Error("Failed to derive private key from enclave", "error", err, "path", derivationPath)
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	privateKey := &models.PrivateKey{
		ID:             uuid.New(),
		OrganizationID: orgID,
		WalletID:       walletUUID,
		Name:           name,
		PublicKey:      keyResp.PublicKey,
		Curve:          curve,
		Path:           derivationPath,
		Tags:           tags,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save to database
	if err := s.db.GetDB().WithContext(ctx).Create(privateKey).Error; err != nil {
		s.logger.Error("Failed to create private key", "error", err)
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	s.logger.Info("Private key created successfully", "private_key_id", privateKey.ID.String(), "path", derivationPath)
	return privateKey, nil
}

// GetPrivateKey retrieves a private key by ID
func (s *PrivateKeyService) GetPrivateKey(ctx context.Context, privateKeyID string) (*models.PrivateKey, error) {
	s.logger.Debug("Getting private key", "private_key_id", privateKeyID)

	var privateKey models.PrivateKey
	if err := s.db.GetDB().WithContext(ctx).Where("id = ?", privateKeyID).First(&privateKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("private key not found: %s", privateKeyID)
		}
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	return &privateKey, nil
}

// ListPrivateKeys lists private keys in an organization with pagination
func (s *PrivateKeyService) ListPrivateKeys(ctx context.Context, organizationID string, pageSize int, pageToken string) ([]models.PrivateKey, string, error) {
	s.logger.Debug("Listing private keys", "organization_id", organizationID)

	var privateKeys []models.PrivateKey
	query := s.db.GetDB().WithContext(ctx).Where("organization_id = ?", organizationID)

	// Handle pagination
	if pageToken != "" {
		// Validate pageToken as UUID
		if _, err := uuid.Parse(pageToken); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		query = query.Where("id > ?", pageToken)
	}

	query = query.Order("created_at ASC").Limit(pageSize + 1) // Get one extra to check if there's a next page

	if err := query.Find(&privateKeys).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list private keys: %w", err)
	}

	var nextToken string
	if len(privateKeys) > pageSize {
		nextToken = privateKeys[pageSize-1].ID.String()
		privateKeys = privateKeys[:pageSize]
	}

	return privateKeys, nextToken, nil
}

// DeletePrivateKey deletes a private key
func (s *PrivateKeyService) DeletePrivateKey(ctx context.Context, privateKeyID string, deleteWithoutExport bool) error {
	s.logger.Info("Deleting private key", "private_key_id", privateKeyID, "delete_without_export", deleteWithoutExport)

	// For production: Check if private key has been exported unless deleteWithoutExport is true
	if !deleteWithoutExport {
		// Check export status - for now, we'll allow deletion
		s.logger.Warn("Private key deletion without export check - implement export verification in production")
	}

	// Delete the private key
	if err := s.db.GetDB().WithContext(ctx).Where("id = ?", privateKeyID).Delete(&models.PrivateKey{}).Error; err != nil {
		return fmt.Errorf("failed to delete private key: %w", err)
	}

	s.logger.Info("Private key deleted successfully", "private_key_id", privateKeyID)
	return nil
}
