package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/google/uuid"
)

// AuthService provides authentication and authorization functionality
type AuthService struct {
	*GAuthService
}

// Authenticate authenticates a user and creates a session
func (s *AuthService) Authenticate(ctx context.Context, organizationID, userID, authMethodID, signature, timestamp string) (*AuthenticationResponse, error) {
	s.logger.Info("Authenticating user", "organization_id", organizationID, "user_id", userID, "auth_method_id", authMethodID)

	// Validate organization ID
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	// Validate user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate auth method ID
	if _, err := uuid.Parse(authMethodID); err != nil {
		return nil, fmt.Errorf("invalid auth method ID: %w", err)
	}

	// Get user
	var user models.User
	if err := s.db.GetDB().WithContext(ctx).First(&user, "id = ? AND organization_id = ?", userUUID, orgID).Error; err != nil {
		return nil, fmt.Errorf("user not found or not in organization")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user is not active")
	}

	// In production, validate signature and timestamp
	// For now, we'll accept any valid UUID combination
	s.logger.Warn("Signature validation not implemented - implement in production")

	// Generate session token
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour session

	// Store session in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionToken)
	sessionData := map[string]interface{}{
		"user_id":         userID,
		"organization_id": organizationID,
		"auth_method_id":  authMethodID,
		"expires_at":      expiresAt.Unix(),
	}

	// Convert session data to JSON for Redis storage
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		s.logger.Error("Failed to marshal session data", "error", err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	if err := s.redis.Set(ctx, sessionKey, string(sessionDataJSON), 24*time.Hour); err != nil {
		s.logger.Error("Failed to store session in Redis", "error", err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	response := &AuthenticationResponse{
		Authenticated: true,
		SessionToken:  sessionToken,
		ExpiresAt:     expiresAt,
		User:          &user,
	}

	s.logger.Info("User authenticated successfully", "user_id", userID, "session_token", sessionToken)
	return response, nil
}

// Authorize checks if a user is authorized for a specific activity
func (s *AuthService) Authorize(ctx context.Context, sessionToken, activityType, parameters string) (*AuthorizationResponse, error) {
	s.logger.Info("Authorizing activity", "activity_type", activityType, "session_token", sessionToken[:8]+"...")

	// Get session from Redis
	sessionKey := fmt.Sprintf("session:%s", sessionToken)
	_, err := s.redis.Get(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// In production, parse session data and validate
	// For now, we'll accept any valid session token
	s.logger.Warn("Session validation not fully implemented - implement in production")

	// Check authorization based on activity type
	authorized := true
	reason := "Authorized"
	var requiredApprovals []string

	switch activityType {
	case "CREATE_WALLET", "DELETE_WALLET":
		// These activities require quorum approval
		authorized = false
		reason = "Requires quorum approval"
		requiredApprovals = []string{"admin", "security_officer"}
	case "CREATE_PRIVATE_KEY", "DELETE_PRIVATE_KEY":
		// These activities require quorum approval
		authorized = false
		reason = "Requires quorum approval"
		requiredApprovals = []string{"admin", "security_officer"}
	case "SIGN_TRANSACTION":
		// These activities require quorum approval
		authorized = false
		reason = "Requires quorum approval"
		requiredApprovals = []string{"admin", "treasurer"}
	default:
		// Read operations are generally allowed
		authorized = true
		reason = "Read operation allowed"
	}

	response := &AuthorizationResponse{
		Authorized:        authorized,
		Reason:            reason,
		RequiredApprovals: requiredApprovals,
	}

	s.logger.Info("Authorization check completed", "activity_type", activityType, "authorized", authorized)
	return response, nil
}
