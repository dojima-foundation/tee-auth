package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationService provides organization management functionality
type OrganizationService struct {
	*GAuthService
}

// CreateOrganization creates a new organization with an initial user
func (s *OrganizationService) CreateOrganization(ctx context.Context, name, initialUserEmail, initialUserPublicKey string) (*models.Organization, error) {
	s.logger.Info("Creating organization", "name", name, "initial_user_email", initialUserEmail)

	org := &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    name,
		RootQuorum: models.Quorum{
			Threshold: s.config.Auth.DefaultQuorumThreshold,
		},
	}

	// Create initial user
	initialUser := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Username:       "admin",
		Email:          initialUserEmail,
		IsActive:       true,
	}

	if initialUserPublicKey != "" {
		initialUser.PublicKey = initialUserPublicKey
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create organization
		if err := tx.Create(org).Error; err != nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}

		// Create initial user
		if err := tx.Create(&initialUser).Error; err != nil {
			return fmt.Errorf("failed to create initial user: %w", err)
		}

		// Add user to root quorum
		if err := tx.Create(&QuorumMember{
			OrganizationID: org.ID,
			UserID:         initialUser.ID,
		}).Error; err != nil {
			return fmt.Errorf("failed to add user to quorum: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create organization", "error", err)
		return nil, err
	}

	// Load users for the organization before returning
	if err := s.db.GetDB().WithContext(ctx).Preload("Users").First(org, org.ID).Error; err != nil {
		s.logger.Error("Failed to load organization with users", "error", err)
		return nil, fmt.Errorf("failed to load organization: %w", err)
	}

	s.logger.Info("Organization created successfully", "organization_id", org.ID.String(), "name", name)
	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	orgID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var org models.Organization
	if err := s.db.GetDB().WithContext(ctx).Preload("Users").First(&org, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// UpdateOrganization updates an organization's name
func (s *OrganizationService) UpdateOrganization(ctx context.Context, id string, name *string) (*models.Organization, error) {
	orgID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var org models.Organization
	if err := s.db.GetDB().WithContext(ctx).First(&org, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if name != nil {
		org.Name = *name
		org.UpdatedAt = time.Now()
	}

	if err := s.db.GetDB().WithContext(ctx).Save(&org).Error; err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	s.logger.Info("Organization updated successfully", "organization_id", org.ID.String())
	return &org, nil
}

// ListOrganizations lists organizations with pagination
func (s *OrganizationService) ListOrganizations(ctx context.Context, pageSize int, pageToken string) ([]models.Organization, string, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var organizations []models.Organization
	query := s.db.GetDB().WithContext(ctx).Order("created_at ASC").Limit(pageSize + 1)

	if pageToken != "" {
		if tokenID, err := uuid.Parse(pageToken); err == nil {
			query = query.Where("id > ?", tokenID)
		}
	}

	if err := query.Find(&organizations).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list organizations: %w", err)
	}

	var nextToken string
	if len(organizations) > pageSize {
		nextToken = organizations[pageSize-1].ID.String()
		organizations = organizations[:pageSize]
	}

	return organizations, nextToken, nil
}
