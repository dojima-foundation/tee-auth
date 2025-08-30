package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	restServer "github.com/dojima-foundation/tee-auth/gauth/api/rest"
	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOAuthIntegration_Simple tests basic OAuth integration functionality
func TestOAuthIntegration_Simple(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	// Setup
	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	// Create test logger
	testLogger, err := logger.New(&config.LoggingConfig{
		Level:  "info",
		Output: "stdout",
		Format: "json",
	})
	require.NoError(t, err)

	// Create mock Redis interface
	mockRedis := &MockRedisInterface{
		sessions: make(map[string]string),
	}

	// Create test server
	server := restServer.NewServer(&config.Config{}, testLogger, nil)
	server.SetRedis(mockRedis)

	// Create session manager
	sessionManager := restServer.NewSessionManager(server)

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

	// Test 1: OAuth callback creates session
	t.Run("OAuthCallbackCreatesSession", func(t *testing.T) {
		// Simulate OAuth callback creating a session
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)
		assert.NotEmpty(t, sessionID)

		// Verify session was created with correct data
		sessionData, err := sessionManager.ValidateSession(ctx, sessionID)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), sessionData.UserID)
		assert.Equal(t, orgID.String(), sessionData.OrganizationID)
		assert.Equal(t, "test@example.com", sessionData.Email)
		assert.Equal(t, "google", sessionData.OAuthProvider)
		assert.Equal(t, "root_user", sessionData.Role)
	})

	// Test 2: Session persists across requests
	t.Run("SessionPersistence", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		// Validate session multiple times
		for i := 0; i < 3; i++ {
			sessionData, err := sessionManager.ValidateSession(ctx, sessionID)
			require.NoError(t, err)
			assert.Equal(t, userID.String(), sessionData.UserID)
			assert.Equal(t, orgID.String(), sessionData.OrganizationID)
		}
	})

	// Test 3: Multiple sessions for same user
	t.Run("MultipleSessions", func(t *testing.T) {
		// Create multiple sessions for the same user
		sessionIDs := make([]string, 3)
		for i := 0; i < 3; i++ {
			sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
			require.NoError(t, err)
			sessionIDs[i] = sessionID
		}

		// Verify all sessions are valid
		for _, sessionID := range sessionIDs {
			sessionData, err := sessionManager.ValidateSession(ctx, sessionID)
			require.NoError(t, err)
			assert.Equal(t, userID.String(), sessionData.UserID)
		}

		// Destroy one session
		err = sessionManager.DestroySession(ctx, sessionIDs[0])
		require.NoError(t, err)

		// Verify destroyed session is invalid
		_, err = sessionManager.ValidateSession(ctx, sessionIDs[0])
		assert.Error(t, err)

		// Verify other sessions still work
		for i := 1; i < 3; i++ {
			sessionData, err := sessionManager.ValidateSession(ctx, sessionIDs[i])
			require.NoError(t, err)
			assert.Equal(t, userID.String(), sessionData.UserID)
		}
	})

	// Test 4: Session with different OAuth providers
	t.Run("DifferentOAuthProviders", func(t *testing.T) {
		providers := []string{"google", "github", "microsoft"}

		for _, provider := range providers {
			sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, provider)
			require.NoError(t, err)

			sessionData, err := sessionManager.ValidateSession(ctx, sessionID)
			require.NoError(t, err)
			assert.Equal(t, provider, sessionData.OAuthProvider)
		}
	})

	// Test 5: Session middleware with OAuth session
	t.Run("SessionMiddlewareWithOAuth", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		// Create test router
		router := gin.New()
		router.Use(sessionManager.SessionMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			sessionData, exists := c.Get("session")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
				return
			}

			session := sessionData.(*restServer.SessionData)
			c.JSON(http.StatusOK, gin.H{
				"user_id":         session.UserID,
				"organization_id": session.OrganizationID,
				"email":           session.Email,
				"oauth_provider":  session.OAuthProvider,
				"role":            session.Role,
			})
		})

		// Test with valid OAuth session
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "gauth_session",
			Value: sessionID,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify response contains OAuth data
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), response["user_id"])
		assert.Equal(t, orgID.String(), response["organization_id"])
		assert.Equal(t, "test@example.com", response["email"])
		assert.Equal(t, "google", response["oauth_provider"])
		assert.Equal(t, "root_user", response["role"])
	})

	// Test 6: Session refresh
	t.Run("SessionRefresh", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		// Get initial session data
		initialSession, err := sessionManager.ValidateSession(ctx, sessionID)
		require.NoError(t, err)
		initialLastActivity := initialSession.LastActivity

		// Wait a bit
		time.Sleep(10 * time.Millisecond)

		// Refresh session (simulate activity)
		refreshedSession, err := sessionManager.ValidateSession(ctx, sessionID)
		require.NoError(t, err)

		// LastActivity should be updated
		assert.True(t, refreshedSession.LastActivity.After(initialLastActivity))
	})
}

// TestOAuthFlow_EndToEnd tests the complete OAuth flow
func TestOAuthFlow_EndToEnd(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	// Setup
	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	// Create test logger
	testLogger, err := logger.New(&config.LoggingConfig{
		Level:  "info",
		Output: "stdout",
		Format: "json",
	})
	require.NoError(t, err)

	// Create mock Redis interface
	mockRedis := &MockRedisInterface{
		sessions: make(map[string]string),
	}

	// Create test server
	server := restServer.NewServer(&config.Config{}, testLogger, nil)
	server.SetRedis(mockRedis)

	// Create session manager
	sessionManager := restServer.NewSessionManager(server)

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

	// Test complete OAuth flow
	t.Run("CompleteOAuthFlow", func(t *testing.T) {
		// Step 1: User initiates OAuth (simulated by creating session)
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		// Step 2: Verify session is created and valid
		sessionData, err := sessionManager.ValidateSession(ctx, sessionID)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), sessionData.UserID)
		assert.Equal(t, "google", sessionData.OAuthProvider)

		// Step 3: User accesses protected resource
		router := gin.New()
		router.Use(sessionManager.SessionMiddleware())
		router.GET("/dashboard", func(c *gin.Context) {
			sessionData, exists := c.Get("session")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Welcome to dashboard", "session": sessionData})
		})

		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.AddCookie(&http.Cookie{
			Name:  "gauth_session",
			Value: sessionID,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Step 4: User logs out
		err = sessionManager.DestroySession(ctx, sessionID)
		require.NoError(t, err)

		// Step 5: Verify session is destroyed
		_, err = sessionManager.ValidateSession(ctx, sessionID)
		assert.Error(t, err)

		// Step 6: Verify protected resource is no longer accessible
		req2, _ := http.NewRequest("GET", "/dashboard", nil)
		req2.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID,
		})

		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusUnauthorized, w2.Code)
	})
}
