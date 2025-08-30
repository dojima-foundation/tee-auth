package e2e

import (
	"context"
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

// TestOAuthSessionsE2E_Simple tests the complete OAuth + Sessions flow
func TestOAuthSessionsE2E_Simple(t *testing.T) {
	// Skip if not running E2E tests
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test. Set E2E_TESTS=true to run.")
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

	// Test complete OAuth + Sessions flow
	t.Run("CompleteOAuthSessionsFlow", func(t *testing.T) {
		// Step 1: User initiates OAuth login
		// This would normally redirect to Google, but we simulate the callback
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)
		assert.NotEmpty(t, sessionID)

		// Step 2: Verify session is created with correct data
		sessionData, err := sessionManager.ValidateSession(ctx, sessionID)
		require.NoError(t, err)
		assert.Equal(t, userID.String(), sessionData.UserID)
		assert.Equal(t, orgID.String(), sessionData.OrganizationID)
		assert.Equal(t, "test@example.com", sessionData.Email)
		assert.Equal(t, "google", sessionData.OAuthProvider)
		assert.Equal(t, "root_user", sessionData.Role)

		// Step 3: User accesses protected resources
		router := gin.New()
		router.Use(sessionManager.SessionMiddleware())

		// Dashboard endpoint
		router.GET("/dashboard", func(c *gin.Context) {
			sessionData, exists := c.Get("session")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Welcome to dashboard",
				"user_id": sessionData.(*restServer.SessionData).UserID,
			})
		})

		// Users endpoint
		router.GET("/users", func(c *gin.Context) {
			sessionData, exists := c.Get("session")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"users": []gin.H{
					{"id": sessionData.(*restServer.SessionData).UserID, "email": "test@example.com"},
				},
			})
		})

		// Wallets endpoint
		router.GET("/wallets", func(c *gin.Context) {
			_, exists := c.Get("session")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"wallets": []gin.H{
					{"id": "wallet-1", "name": "Test Wallet"},
				},
			})
		})

		// Test accessing dashboard
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.AddCookie(&http.Cookie{
			Name:  "gauth_session",
			Value: sessionID,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Test accessing users
		req2, _ := http.NewRequest("GET", "/users", nil)
		req2.AddCookie(&http.Cookie{
			Name:  "gauth_session",
			Value: sessionID,
		})

		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)

		// Test accessing wallets
		req3, _ := http.NewRequest("GET", "/wallets", nil)
		req3.AddCookie(&http.Cookie{
			Name:  "gauth_session",
			Value: sessionID,
		})

		w3 := httptest.NewRecorder()
		router.ServeHTTP(w3, req3)

		assert.Equal(t, http.StatusOK, w3.Code)

		// Step 4: User logs out
		err = sessionManager.DestroySession(ctx, sessionID)
		require.NoError(t, err)

		// Step 5: Verify session is destroyed
		_, err = sessionManager.ValidateSession(ctx, sessionID)
		assert.Error(t, err)

		// Step 6: Verify protected resources are no longer accessible
		req4, _ := http.NewRequest("GET", "/dashboard", nil)
		req4.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID,
		})

		w4 := httptest.NewRecorder()
		router.ServeHTTP(w4, req4)

		assert.Equal(t, http.StatusUnauthorized, w4.Code)
	})

	// Test session management functionality
	t.Run("SessionManagement", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		// Test session validation
		_, err = sessionManager.ValidateSession(ctx, sessionID)
		require.NoError(t, err)

		// Test session destruction
		err = sessionManager.DestroySession(ctx, sessionID)
		require.NoError(t, err)

		// Verify session is destroyed
		_, err = sessionManager.ValidateSession(ctx, sessionID)
		assert.Error(t, err)
	})
}

// TestCrossDomainSessionsE2E_Simple tests cross-domain session scenarios
func TestCrossDomainSessionsE2E_Simple(t *testing.T) {
	// Skip if not running E2E tests
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test. Set E2E_TESTS=true to run.")
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

	// Test cross-domain session scenarios
	t.Run("CrossDomainSessionScenarios", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		// Create router with CORS middleware
		router := gin.New()

		// Add CORS middleware first
		router.Use(func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		})

		// Then add session middleware
		router.Use(sessionManager.SessionMiddleware())

		// API endpoint
		router.GET("/api/v1/user", func(c *gin.Context) {
			sessionData, exists := c.Get("session")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"user_id": sessionData.(*restServer.SessionData).UserID,
				"email":   sessionData.(*restServer.SessionData).Email,
			})
		})

		// Test with session cookie
		req, _ := http.NewRequest("GET", "/api/v1/user", nil)
		req.AddCookie(&http.Cookie{
			Name:  "gauth_session",
			Value: sessionID,
		})
		req.Header.Set("Origin", "https://dashboard.dojima.foundation")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

		// Test OPTIONS request (CORS preflight)
		req2, _ := http.NewRequest("OPTIONS", "/api/v1/user", nil)
		req2.Header.Set("Origin", "https://dashboard.dojima.foundation")
		req2.Header.Set("Access-Control-Request-Method", "GET")

		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusNoContent, w2.Code)
		assert.Equal(t, "true", w2.Header().Get("Access-Control-Allow-Credentials"))
	})

	// Test session persistence across domains
	t.Run("SessionPersistenceAcrossDomains", func(t *testing.T) {
		sessionID, err := sessionManager.CreateSession(ctx, user, authMethod, "google")
		require.NoError(t, err)

		// Simulate requests from different domains
		domains := []string{
			"https://dashboard.dojima.foundation",
			"https://api.dojima.foundation",
			"http://localhost:3001",
		}

		router := gin.New()
		router.Use(sessionManager.SessionMiddleware())
		router.GET("/api/v1/session/validate", func(c *gin.Context) {
			sessionData, exists := c.Get("session")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"valid": true, "session": sessionData})
		})

		for _, domain := range domains {
			req, _ := http.NewRequest("GET", "/api/v1/session/validate", nil)
			req.AddCookie(&http.Cookie{
				Name:  "gauth_session",
				Value: sessionID,
			})
			req.Header.Set("Origin", domain)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Session should be valid for domain: %s", domain)
		}
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
