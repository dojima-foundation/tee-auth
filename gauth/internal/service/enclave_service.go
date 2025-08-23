package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/google/uuid"
)

// EnclaveService provides enclave integration functionality
type EnclaveService struct {
	*GAuthService
}

// RequestSeedGeneration requests seed generation from the enclave
func (s *EnclaveService) RequestSeedGeneration(ctx context.Context, organizationID, userID string, strength int, passphrase *string) (*SeedGenerationResponse, error) {
	s.logger.Info("Requesting seed generation", "organization_id", organizationID, "user_id", userID, "strength", strength)

	// Validate organization ID
	if _, err := uuid.Parse(organizationID); err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	// Validate user ID
	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate strength
	if strength != 128 && strength != 256 {
		return nil, fmt.Errorf("invalid strength: %d. Must be 128 or 256", strength)
	}

	// Request seed generation from enclave
	seedResp, err := s.renclave.GenerateSeed(ctx, strength, passphrase)
	if err != nil {
		s.logger.Error("Failed to generate seed from enclave", "error", err)
		return nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	// Create activity record
	activity := &models.Activity{
		ID:             uuid.New(),
		OrganizationID: uuid.MustParse(organizationID),
		Type:           "SEED_GENERATION",
		Status:         "COMPLETED",
		CreatedBy:      uuid.MustParse(userID),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.GetDB().WithContext(ctx).Create(activity).Error; err != nil {
		s.logger.Error("Failed to create seed generation activity", "error", err)
		// Don't fail the request if activity logging fails
	}

	response := &SeedGenerationResponse{
		SeedPhrase: seedResp.SeedPhrase,
		Entropy:    seedResp.Entropy,
		Strength:   seedResp.Strength,
		WordCount:  seedResp.WordCount,
		RequestID:  activity.ID.String(),
	}

	s.logger.Info("Seed generation completed", "request_id", response.RequestID, "strength", response.Strength)
	return response, nil
}

// ValidateSeed validates a seed phrase using the enclave
func (s *EnclaveService) ValidateSeed(ctx context.Context, seedPhrase string) (*SeedValidationResponse, error) {
	s.logger.Info("Validating seed phrase")

	// Request seed validation from enclave
	validateResp, err := s.renclave.ValidateSeed(ctx, seedPhrase)
	if err != nil {
		s.logger.Error("Failed to validate seed from enclave", "error", err)
		return nil, fmt.Errorf("failed to validate seed: %w", err)
	}

	response := &SeedValidationResponse{
		IsValid:   validateResp.IsValid,
		Strength:  validateResp.Strength,
		WordCount: validateResp.WordCount,
		Errors:    validateResp.Errors,
	}

	s.logger.Info("Seed validation completed", "is_valid", response.IsValid, "word_count", response.WordCount)
	return response, nil
}

// GetEnclaveInfo retrieves information about the enclave
func (s *EnclaveService) GetEnclaveInfo(ctx context.Context) (*EnclaveInfoResponse, error) {
	s.logger.Debug("Getting enclave information")

	// Request info from enclave
	infoResp, err := s.renclave.GetInfo(ctx)
	if err != nil {
		s.logger.Error("Failed to get enclave info", "error", err)
		return nil, fmt.Errorf("failed to get enclave info: %w", err)
	}

	response := &EnclaveInfoResponse{
		Version:      infoResp.Version,
		EnclaveID:    infoResp.EnclaveID,
		Capabilities: infoResp.Capabilities,
		Healthy:      infoResp.Healthy,
	}

	return response, nil
}
