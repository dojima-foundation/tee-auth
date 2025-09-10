package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivityService provides activity management functionality
type ActivityService struct {
	*GAuthService
}

// CreateActivity creates a new activity
func (s *ActivityService) CreateActivity(ctx context.Context, organizationID, activityType, parameters, createdBy string) (*models.Activity, error) {
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	createdByID, err := uuid.Parse(createdBy)
	if err != nil {
		return nil, fmt.Errorf("invalid created_by user ID: %w", err)
	}

	// Validate and convert JSON parameters
	var parametersRaw json.RawMessage
	if parameters != "" {
		if err := json.Unmarshal([]byte(parameters), &parametersRaw); err != nil {
			return nil, fmt.Errorf("invalid JSON parameters: %w", err)
		}
	}

	activity := &models.Activity{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Type:           activityType,
		Status:         "PENDING",
		Parameters:     parametersRaw,
		CreatedBy:      createdByID,
	}

	if err := s.db.GetDB().WithContext(ctx).Create(activity).Error; err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	s.logger.LogActivity(activityType, activity.ID.String(), createdBy, organizationID)

	// Record metrics
	telemetry.RecordActivityCreated(activityType)

	return activity, nil
}

// GetActivity retrieves an activity by ID
func (s *ActivityService) GetActivity(ctx context.Context, id string) (*models.Activity, error) {
	activityID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid activity ID: %w", err)
	}

	var activity models.Activity
	if err := s.db.GetDB().WithContext(ctx).First(&activity, "id = ?", activityID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("activity not found")
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
}

// ListActivities lists activities in an organization with pagination and filtering
func (s *ActivityService) ListActivities(ctx context.Context, organizationID string, activityType, status *string, pageSize int, pageToken string) ([]models.Activity, string, error) {
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

	query := s.db.GetDB().WithContext(ctx).
		Where("organization_id = ?", orgID).
		Limit(pageSize + 1)

	if activityType != nil {
		query = query.Where("type = ?", *activityType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if pageToken != "" {
		if tokenID, err := uuid.Parse(pageToken); err == nil {
			query = query.Where("id > ?", tokenID)
		}
	}

	var activities []models.Activity
	if err := query.Find(&activities).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list activities: %w", err)
	}

	var nextToken string
	if len(activities) > pageSize {
		nextToken = activities[pageSize-1].ID.String()
		activities = activities[:pageSize]
	}

	return activities, nextToken, nil
}
