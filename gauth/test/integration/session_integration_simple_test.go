package integration

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionIntegration_Simple tests basic session management functionality
func TestSessionIntegration_Simple(t *testing.T) {
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

	// Test 1: Create session
	t.Run("CreateSession", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)
		assert.NotEmpty(t, sessionID)

		// Verify session was stored
		sessionKey := fmt.Sprintf("session:%s", sessionID)
		sessionData, err := mockRedis.Get(ctx, sessionKey)
		require.NoError(t, err)
		assert.NotEmpty(t, sessionData)

		// Verify session data structure
		var sessionDataObj restServer.SessionData
		err = json.Unmarshal([]byte(sessionData), &sessionDataObj)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), sessionDataObj.UserID)
		assert.Equal(t, orgID.String(), sessionDataObj.OrganizationID)
		assert.Equal(t, "test@example.com", sessionDataObj.Email)
		assert.Equal(t, "google", sessionDataObj.OAuthProvider)
	})

	// Test 2: Validate session
	t.Run("ValidateSession", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		sessionData, err := sessionManager.ValidateSession(ctx, sessionID)
		require.NoError(t, err)
		assert.NotNil(t, sessionData)
		assert.Equal(t, userID.String(), sessionData.UserID)
		assert.Equal(t, orgID.String(), sessionData.OrganizationID)
		assert.Equal(t, "test@example.com", sessionData.Email)
	})

	// Test 3: Session middleware
	t.Run("SessionMiddleware", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		// Create test router
		router := gin.New()
		router.Use(sessionManager.SessionMiddleware())
		router.GET("/test", func(c *gin.Context) {
			sessionData, exists := c.Get("session")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"session": sessionData})
		})

		// Test with valid session cookie
		req, _ := http.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "gauth_session",
			Value: sessionID,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Test without session cookie
		req2, _ := http.NewRequest("GET", "/test", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusUnauthorized, w2.Code)
	})

	// Test 4: Destroy session
	t.Run("DestroySession", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		// Verify session exists
		_, err = sessionManager.ValidateSession(ctx, sessionID)
		require.NoError(t, err)

		// Destroy session
		err = sessionManager.DestroySession(ctx, sessionID)
		require.NoError(t, err)

		// Verify session no longer exists
		_, err = sessionManager.ValidateSession(ctx, sessionID)
		assert.Error(t, err)
	})

	// Test 5: Session expiration
	t.Run("SessionExpiration", func(t *testing.T) {
		// Create a session with short expiration
		sessionID := uuid.New().String()
		now := time.Now()
		expiresAt := now.Add(1 * time.Millisecond) // Very short expiration

		sessionData := restServer.SessionData{
			UserID:         userID.String(),
			OrganizationID: orgID.String(),
			AuthMethodID:   authMethod.ID.String(),
			OAuthProvider:  "google",
			Email:          user.Email,
			Role:           "root_user",
			CreatedAt:      now,
			LastActivity:   now,
			ExpiresAt:      expiresAt,
		}

		// Store session in Redis
		sessionKey := fmt.Sprintf("session:%s", sessionID)
		sessionDataJSON, err := json.Marshal(sessionData)
		require.NoError(t, err)

		err = mockRedis.Set(ctx, sessionKey, string(sessionDataJSON), 1*time.Millisecond)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Try to validate expired session
		_, err = sessionManager.ValidateSession(ctx, sessionID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

// MockRedisInterface is a simple mock for Redis operations
type MockRedisInterface struct {
	sessions map[string]string
}

func (m *MockRedisInterface) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	if str, ok := data.(string); ok {
		m.sessions[sessionID] = str
	} else {
		m.sessions[sessionID] = "mock_value"
	}
	return nil
}

func (m *MockRedisInterface) GetSession(ctx context.Context, sessionID string) (string, error) {
	value, exists := m.sessions[sessionID]
	if !exists {
		return "", nil
	}
	return value, nil
}

func (m *MockRedisInterface) DeleteSession(ctx context.Context, sessionID string) error {
	delete(m.sessions, sessionID)
	return nil
}

func (m *MockRedisInterface) ExtendSession(ctx context.Context, sessionID string, expiration time.Duration) error {
	return nil
}

func (m *MockRedisInterface) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if str, ok := value.(string); ok {
		m.sessions[key] = str
	} else {
		m.sessions[key] = "mock_value"
	}
	return nil
}

func (m *MockRedisInterface) Get(ctx context.Context, key string) (string, error) {
	value, exists := m.sessions[key]
	if !exists {
		return "", nil
	}
	return value, nil
}

func (m *MockRedisInterface) Delete(ctx context.Context, key string) error {
	delete(m.sessions, key)
	return nil
}

func (m *MockRedisInterface) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.sessions[key]
	return exists, nil
}

func (m *MockRedisInterface) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if _, exists := m.sessions[key]; exists {
		return false, nil
	}
	if str, ok := value.(string); ok {
		m.sessions[key] = str
	} else {
		m.sessions[key] = "mock_value"
	}
	return true, nil
}

func (m *MockRedisInterface) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return 1, nil
}

func (m *MockRedisInterface) GetCounter(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

func (m *MockRedisInterface) AcquireLock(ctx context.Context, lockKey string, expiration time.Duration) (bool, error) {
	return true, nil
}

func (m *MockRedisInterface) ReleaseLock(ctx context.Context, lockKey string) error {
	return nil
}

func (m *MockRedisInterface) ExtendLock(ctx context.Context, lockKey string, expiration time.Duration) error {
	return nil
}

func (m *MockRedisInterface) GetClient() *redis.Client {
	return nil
}

func (m *MockRedisInterface) Close() error {
	return nil
}

func (m *MockRedisInterface) Health(ctx context.Context) error {
	return nil
}

func (m *MockRedisInterface) GetStats() map[string]interface{} {
	return make(map[string]interface{})
}
