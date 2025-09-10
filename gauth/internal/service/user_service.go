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

// UserService provides user management functionality
type UserService struct {
	*GAuthService
}

// CreateUser creates a new user in an organization
func (s *UserService) CreateUser(ctx context.Context, organizationID, username, email, publicKey string, tags []string) (*models.User, error) {
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	user := &models.User{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Username:       username,
		Email:          email,
		PublicKey:      publicKey,
		Tags:           tags,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.GetDB().WithContext(ctx).Create(user).Error; err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User created successfully", "user_id", user.ID.String(), "username", username)

	// Record metrics
	telemetry.RecordUserCreated()

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user models.User
	if err := s.db.GetDB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(ctx context.Context, id string, username, email *string, tags []string, isActive *bool) (*models.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user models.User
	if err := s.db.GetDB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if username != nil {
		user.Username = *username
	}
	if email != nil {
		user.Email = *email
	}
	if tags != nil {
		user.Tags = tags
	}
	if isActive != nil {
		user.IsActive = *isActive
	}

	user.UpdatedAt = time.Now()

	if err := s.db.GetDB().WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("User updated successfully", "user_id", user.ID.String())
	return &user, nil
}

// ListUsers lists users in an organization with pagination
func (s *UserService) ListUsers(ctx context.Context, organizationID string, pageSize int, pageToken string) ([]models.User, string, error) {
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid organization ID: %w", err)
	}

	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var users []models.User
	query := s.db.GetDB().WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("id ASC").
		Limit(pageSize + 1)

	if pageToken != "" {
		if tokenID, err := uuid.Parse(pageToken); err == nil {
			query = query.Where("id > ?", tokenID)
		}
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list users: %w", err)
	}

	var nextToken string
	if len(users) > pageSize {
		nextToken = users[pageSize-1].ID.String()
		users = users[:pageSize]
	}

	return users, nextToken, nil
}
