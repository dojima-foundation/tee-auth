package rest

import (
	"context"
	"fmt"
	"testing"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSessionManager_CreateSession_Simple(t *testing.T) {
	// Setup
	mockRedis := NewMockRedisInterface()

	// Create a real logger for testing
	testLogger, err := logger.New(&config.LoggingConfig{
		Level:  "info",
		Output: "stdout",
		Format: "json",
	})
	assert.NoError(t, err)

	// Create a minimal server for testing
	server := &Server{
		redis:  mockRedis,
		logger: testLogger,
	}
	sessionManager := NewSessionManager(server)

	// Test data
	userID := uuid.New()
	orgID := uuid.New()
	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Email:          "test@example.com",
		Username:       "testuser",
	}
	authMethod := &models.AuthMethod{
		ID:     uuid.New(),
		UserID: userID,
		Type:   "OAUTH",
		Name:   "Google OAuth",
	}

	// Test session creation
	sessionID, err := sessionManager.CreateSession(context.Background(), user, authMethod, "google")

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, sessionID)

	// Verify session was stored in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	sessionData, err := mockRedis.Get(context.Background(), sessionKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, sessionData)
}

func TestSessionManager_ValidateSession_Simple(t *testing.T) {
	// Setup
	mockRedis := NewMockRedisInterface()

	// Create a real logger for testing
	testLogger, err := logger.New(&config.LoggingConfig{
		Level:  "info",
		Output: "stdout",
		Format: "json",
	})
	assert.NoError(t, err)

	// Create a minimal server for testing
	server := &Server{
		redis:  mockRedis,
		logger: testLogger,
	}
	sessionManager := NewSessionManager(server)

	// Test data
	userID := uuid.New()
	orgID := uuid.New()
	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Email:          "test@example.com",
		Username:       "testuser",
	}
	authMethod := &models.AuthMethod{
		ID:     uuid.New(),
		UserID: userID,
		Type:   "OAUTH",
		Name:   "Google OAuth",
	}

	// Create a session first
	sessionID, err := sessionManager.CreateSession(context.Background(), user, authMethod, "google")
	assert.NoError(t, err)

	// Test getting the session
	sessionData, err := sessionManager.ValidateSession(context.Background(), sessionID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, sessionData)
	assert.Equal(t, userID.String(), sessionData.UserID)
	assert.Equal(t, orgID.String(), sessionData.OrganizationID)
	assert.Equal(t, "test@example.com", sessionData.Email)
}

func TestSessionManager_DestroySession_Simple(t *testing.T) {
	// Setup
	mockRedis := NewMockRedisInterface()

	// Create a real logger for testing
	testLogger, err := logger.New(&config.LoggingConfig{
		Level:  "info",
		Output: "stdout",
		Format: "json",
	})
	assert.NoError(t, err)

	// Create a minimal server for testing
	server := &Server{
		redis:  mockRedis,
		logger: testLogger,
	}
	sessionManager := NewSessionManager(server)

	// Test data
	userID := uuid.New()
	orgID := uuid.New()
	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Email:          "test@example.com",
		Username:       "testuser",
	}
	authMethod := &models.AuthMethod{
		ID:     uuid.New(),
		UserID: userID,
		Type:   "OAUTH",
		Name:   "Google OAuth",
	}

	// Create a session first
	sessionID, err := sessionManager.CreateSession(context.Background(), user, authMethod, "google")
	assert.NoError(t, err)

	// Verify session exists
	_, err = sessionManager.ValidateSession(context.Background(), sessionID)
	assert.NoError(t, err)

	// Test destroying the session
	err = sessionManager.DestroySession(context.Background(), sessionID)
	assert.NoError(t, err)

	// Verify session no longer exists
	_, err = sessionManager.ValidateSession(context.Background(), sessionID)
	assert.Error(t, err)
}
