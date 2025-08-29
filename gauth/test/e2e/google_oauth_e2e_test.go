package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
)

type GoogleOAuthE2ETestSuite struct {
	suite.Suite
	config      *config.Config
	database    *db.PostgresDB
	redis       *db.RedisClient
	service     *service.GAuthService
	googleOAuth *service.GoogleOAuthService
	mockServer  *httptest.Server
}

func (suite *GoogleOAuthE2ETestSuite) SetupSuite() {
	// Skip E2E tests if not in E2E test mode
	if testhelpers.GetEnvOrDefault("E2E_TESTS", "") != "true" {
		suite.T().Skip("Skipping E2E tests. Set E2E_TESTS=true to run.")
	}

	// Setup test configuration
	suite.config = &config.Config{
		Database: config.DatabaseConfig{
			Host:         testhelpers.GetEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:         5432,
			Username:     testhelpers.GetEnvOrDefault("TEST_DB_USER", "gauth"),
			Password:     testhelpers.GetEnvOrDefault("TEST_DB_PASSWORD", "password"),
			Database:     testhelpers.GetEnvOrDefault("TEST_DB_NAME", "gauth_test"),
			SSLMode:      "disable",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
			MaxLifetime:  5 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host:         testhelpers.GetEnvOrDefault("TEST_REDIS_HOST", "localhost"),
			Port:         6379,
			Password:     testhelpers.GetEnvOrDefault("TEST_REDIS_PASSWORD", ""),
			Database:     4, // Different database for E2E tests
			PoolSize:     10,
			MinIdleConns: 5,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
			GoogleOAuth: config.GoogleOAuthConfig{
				ClientID:     "e2e-test-client-id",
				ClientSecret: "e2e-test-client-secret",
				RedirectURL:  "http://localhost:8080/callback",
			},
		},
		Renclave: config.RenclaveConfig{
			Host:    testhelpers.GetEnvOrDefault("TEST_RENCLAVE_HOST", "localhost"),
			Port:    3000,
			UseTLS:  false,
			Timeout: 30 * time.Second,
		},
	}

	// Initialize dependencies
	var err error
	suite.database, err = db.NewPostgresDB(&suite.config.Database)
	require.NoError(suite.T(), err)

	suite.redis, err = db.NewRedisClient(&suite.config.Redis)
	require.NoError(suite.T(), err)

	logger := logger.NewDefault()

	// Setup mock Google OAuth server first
	suite.setupMockGoogleServer()

	// Initialize service with mock endpoint
	suite.service = service.NewGAuthService(suite.config, logger, suite.database, suite.redis)

	// Create custom OAuth endpoint pointing to mock server
	mockEndpoint := oauth2.Endpoint{
		AuthURL:  suite.mockServer.URL + "/oauth2/authorize",
		TokenURL: suite.mockServer.URL + "/oauth2/token",
	}

	suite.googleOAuth = service.NewGoogleOAuthServiceWithEndpointAndAPIBase(suite.service, suite.config, mockEndpoint, suite.mockServer.URL)
}

func (suite *GoogleOAuthE2ETestSuite) TearDownSuite() {
	if suite.mockServer != nil {
		suite.mockServer.Close()
	}
	if suite.database != nil {
		suite.database.Close()
	}
	if suite.redis != nil {
		suite.redis.Close()
	}
}

func (suite *GoogleOAuthE2ETestSuite) SetupTest() {
	// Clean up database before each test
	suite.cleanupDatabase()
}

func (suite *GoogleOAuthE2ETestSuite) setupMockGoogleServer() {
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/authorize":
			// Mock OAuth2 authorization endpoint - redirect with code
			code := r.URL.Query().Get("code")
			if code == "" {
				code = "e2e-mock-auth-code"
			}
			state := r.URL.Query().Get("state")
			redirectURI := r.URL.Query().Get("redirect_uri")

			// Redirect to callback with code and state
			http.Redirect(w, r, redirectURI+"?code="+code+"&state="+state, http.StatusFound)

		case "/oauth2/token":
			// Mock OAuth2 token exchange
			if err := r.ParseForm(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			code := r.FormValue("code")
			if code == "" {
				// Return error for empty code
				w.WriteHeader(http.StatusBadRequest)
				errorResponse := map[string]interface{}{
					"error":             "invalid_grant",
					"error_description": "Invalid authorization code",
				}
				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			tokenResponse := map[string]interface{}{
				"access_token":  "e2e-mock-access-token",
				"refresh_token": "e2e-mock-refresh-token",
				"token_type":    "Bearer",
				"expires_in":    3600,
			}
			json.NewEncoder(w).Encode(tokenResponse)

		case "/oauth2/v2/userinfo":
			// Mock Google user info endpoint
			w.Header().Set("Content-Type", "application/json")
			userInfo := map[string]interface{}{
				"id":             "e2e-mock-google-user-123",
				"email":          "e2e-test@example.com",
				"name":           "E2E Test User",
				"given_name":     "E2E",
				"family_name":    "Test",
				"picture":        "https://example.com/e2e-avatar.jpg",
				"verified_email": true,
			}
			json.NewEncoder(w).Encode(userInfo)

		default:
			http.NotFound(w, r)
		}
	}))
}

func (suite *GoogleOAuthE2ETestSuite) cleanupDatabase() {
	ctx := context.Background()

	// Clean up in reverse order of dependencies
	suite.database.GetDB().WithContext(ctx).Exec("DELETE FROM quorum_members")
	suite.database.GetDB().WithContext(ctx).Exec("DELETE FROM auth_methods")
	suite.database.GetDB().WithContext(ctx).Exec("DELETE FROM users")
	suite.database.GetDB().WithContext(ctx).Exec("DELETE FROM organizations")
}

func (suite *GoogleOAuthE2ETestSuite) TestGoogleOAuthCompleteFlow() {
	ctx := context.Background()

	// Step 1: Generate OAuth URL for new user
	authURL, err := suite.googleOAuth.GetAuthURL(ctx, "", "e2e-test-state")
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), authURL)
	assert.Contains(suite.T(), authURL, "oauth2/authorize")

	// Step 2: Simulate OAuth callback
	state := suite.createStateWithoutOrgID()
	code := "e2e-mock-auth-code"

	response, err := suite.googleOAuth.HandleCallback(ctx, code, state)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.True(suite.T(), response.Success)

	// Step 3: Verify complete user setup
	user := response.User
	assert.Equal(suite.T(), "Root user", user.Username)
	assert.Equal(suite.T(), "e2e-test@example.com", user.Email)
	assert.True(suite.T(), user.IsActive)

	// Step 4: Verify organization creation
	org, err := suite.service.GetOrganization(ctx, user.OrganizationID.String())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "E2E Test User's Organization", org.Name)
	assert.Equal(suite.T(), "1.0", org.Version)

	// Step 5: Verify auth method creation
	authMethod := response.AuthMethod
	assert.Equal(suite.T(), "OAUTH", authMethod.Type)
	assert.Equal(suite.T(), "Google OAuth", authMethod.Name)
	assert.True(suite.T(), authMethod.IsActive)

	// Step 6: Verify session creation
	sessionData, err := suite.redis.Get(ctx, fmt.Sprintf("session:%s", response.SessionToken))
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), sessionData)

	var session map[string]interface{}
	err = json.Unmarshal([]byte(sessionData), &session)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID.String(), session["user_id"])
	assert.Equal(suite.T(), user.OrganizationID.String(), session["organization_id"])
	assert.Equal(suite.T(), "google", session["oauth_provider"])

	// Step 7: Verify quorum membership
	var quorumMember service.QuorumMember
	err = suite.database.GetDB().WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", user.OrganizationID, user.ID).
		First(&quorumMember).Error
	assert.NoError(suite.T(), err)
}

func (suite *GoogleOAuthE2ETestSuite) TestGoogleOAuthReturningUserFlow() {
	ctx := context.Background()

	// Step 1: Create initial user through OAuth
	state1 := suite.createStateWithoutOrgID()
	code1 := "e2e-mock-auth-code-1"

	response1, err := suite.googleOAuth.HandleCallback(ctx, code1, state1)
	require.NoError(suite.T(), err)

	// Step 2: Simulate user returning (same Google account)
	state2 := suite.createStateWithOrgID(response1.User.OrganizationID.String())
	code2 := "e2e-mock-auth-code-2"

	response2, err := suite.googleOAuth.HandleCallback(ctx, code2, state2)
	require.NoError(suite.T(), err)

	// Step 3: Verify same user is returned
	assert.Equal(suite.T(), response1.User.ID, response2.User.ID)
	assert.Equal(suite.T(), response1.User.OrganizationID, response2.User.OrganizationID)
	assert.Equal(suite.T(), response1.User.Email, response2.User.Email)

	// Step 4: Verify new session is created
	assert.NotEqual(suite.T(), response1.SessionToken, response2.SessionToken)
	assert.True(suite.T(), response2.Success)
}

func (suite *GoogleOAuthE2ETestSuite) TestGoogleOAuthMultipleOrganizations() {
	ctx := context.Background()

	// Step 1: Create first organization and user
	state1 := suite.createStateWithoutOrgID()
	code1 := "e2e-mock-auth-code-1"

	response1, err := suite.googleOAuth.HandleCallback(ctx, code1, state1)
	require.NoError(suite.T(), err)

	// Step 2: Create second organization and user with different Google ID
	state2 := suite.createStateWithoutOrgID()
	code2 := "e2e-mock-auth-code-2"

	// Modify mock server to return different Google ID
	originalHandler := suite.mockServer.Config.Handler
	suite.mockServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/v2/userinfo" {
			w.Header().Set("Content-Type", "application/json")
			userInfo := map[string]interface{}{
				"id":             "e2e-mock-google-user-456", // Different ID
				"email":          "e2e-test-2@example.com",   // Different email
				"name":           "E2E Test User 2",
				"given_name":     "E2E",
				"family_name":    "Test2",
				"picture":        "https://example.com/e2e-avatar-2.jpg",
				"verified_email": true,
			}
			json.NewEncoder(w).Encode(userInfo)
		} else {
			originalHandler.ServeHTTP(w, r)
		}
	})
	defer func() {
		suite.mockServer.Config.Handler = originalHandler
	}()

	response2, err := suite.googleOAuth.HandleCallback(ctx, code2, state2)
	require.NoError(suite.T(), err)

	// Step 3: Verify different organizations and users
	assert.NotEqual(suite.T(), response1.User.ID, response2.User.ID)
	assert.NotEqual(suite.T(), response1.User.OrganizationID, response2.User.OrganizationID)
	assert.NotEqual(suite.T(), response1.User.Email, response2.User.Email)

	// Step 4: Verify both organizations exist
	org1, err := suite.service.GetOrganization(ctx, response1.User.OrganizationID.String())
	require.NoError(suite.T(), err)
	org2, err := suite.service.GetOrganization(ctx, response2.User.OrganizationID.String())
	require.NoError(suite.T(), err)

	assert.NotEqual(suite.T(), org1.ID, org2.ID)
	assert.NotEqual(suite.T(), org1.Name, org2.Name)
}

func (suite *GoogleOAuthE2ETestSuite) TestGoogleOAuthSessionManagement() {
	ctx := context.Background()

	// Step 1: Create user and session
	state := suite.createStateWithoutOrgID()
	code := "e2e-mock-auth-code"

	response, err := suite.googleOAuth.HandleCallback(ctx, code, state)
	require.NoError(suite.T(), err)

	// Step 2: Verify session exists and is valid
	sessionData, err := suite.redis.Get(ctx, fmt.Sprintf("session:%s", response.SessionToken))
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), sessionData)

	// Step 3: Verify session contains correct data
	var session map[string]interface{}
	err = json.Unmarshal([]byte(sessionData), &session)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), response.User.ID.String(), session["user_id"])
	assert.Equal(suite.T(), response.User.OrganizationID.String(), session["organization_id"])
	assert.Equal(suite.T(), response.AuthMethod.ID.String(), session["auth_method_id"])
	assert.Equal(suite.T(), "google", session["oauth_provider"])

	// Step 4: Verify session expiration
	expiresAt, ok := session["expires_at"].(float64)
	assert.True(suite.T(), ok)
	assert.Greater(suite.T(), expiresAt, float64(time.Now().Unix()))
}

func (suite *GoogleOAuthE2ETestSuite) TestGoogleOAuthDataPersistence() {
	ctx := context.Background()

	// Step 1: Create user through OAuth
	state := suite.createStateWithoutOrgID()
	code := "e2e-mock-auth-code"

	response, err := suite.googleOAuth.HandleCallback(ctx, code, state)
	require.NoError(suite.T(), err)

	// Step 2: Verify user data is persisted
	user, err := suite.service.GetUser(ctx, response.User.ID.String())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Root user", user.Username)
	assert.Equal(suite.T(), "e2e-test@example.com", user.Email)
	assert.True(suite.T(), user.IsActive)

	// Step 3: Verify organization data is persisted
	org, err := suite.service.GetOrganization(ctx, user.OrganizationID.String())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "E2E Test User's Organization", org.Name)
	assert.Equal(suite.T(), "1.0", org.Version)

	// Step 4: Verify auth method data is persisted
	// Note: We would need to add a method to retrieve auth methods by user ID
	// For now, we'll verify the auth method exists in the response
	assert.NotNil(suite.T(), response.AuthMethod)
	assert.Equal(suite.T(), "OAUTH", response.AuthMethod.Type)
	assert.Equal(suite.T(), "Google OAuth", response.AuthMethod.Name)
}

func (suite *GoogleOAuthE2ETestSuite) TestGoogleOAuthErrorHandling() {
	ctx := context.Background()

	tests := []struct {
		name          string
		code          string
		state         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "invalid state format",
			code:          "e2e-mock-auth-code",
			state:         "invalid-state",
			expectError:   true,
			errorContains: "invalid state parameter",
		},
		{
			name:          "empty state",
			code:          "e2e-mock-auth-code",
			state:         "",
			expectError:   true,
			errorContains: "invalid state parameter",
		},
		{
			name:          "empty code",
			code:          "",
			state:         suite.createStateWithoutOrgID(),
			expectError:   true,
			errorContains: "",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			response, err := suite.googleOAuth.HandleCallback(ctx, tt.code, tt.state)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}
		})
	}
}

func (suite *GoogleOAuthE2ETestSuite) TestGoogleOAuthConcurrentAccess() {
	ctx := context.Background()

	// Test concurrent OAuth callbacks
	const numConcurrent = 5
	results := make(chan *service.GoogleOAuthResponse, numConcurrent)
	errors := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(index int) {
			state := suite.createStateWithoutOrgID()
			code := fmt.Sprintf("e2e-mock-auth-code-%d", index)

			response, err := suite.googleOAuth.HandleCallback(ctx, code, state)
			if err != nil {
				errors <- err
			} else {
				results <- response
			}
		}(i)
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < numConcurrent; i++ {
		select {
		case response := <-results:
			assert.NotNil(suite.T(), response)
			assert.True(suite.T(), response.Success)
			successCount++
		case err := <-errors:
			assert.Error(suite.T(), err)
			errorCount++
		case <-time.After(10 * time.Second):
			suite.T().Fatal("Timeout waiting for concurrent OAuth callbacks")
		}
	}

	// Verify all requests succeeded
	assert.Equal(suite.T(), numConcurrent, successCount)
	assert.Equal(suite.T(), 0, errorCount)
}

// Helper functions

func (suite *GoogleOAuthE2ETestSuite) createStateWithOrgID(orgID string) string {
	stateData := map[string]string{
		"organization_id": orgID,
		"state":           "e2e-test-state",
		"timestamp":       fmt.Sprintf("%d", time.Now().Unix()),
	}
	stateJSON, _ := json.Marshal(stateData)
	return string(stateJSON)
}

func (suite *GoogleOAuthE2ETestSuite) createStateWithoutOrgID() string {
	stateData := map[string]string{
		"state":     "e2e-test-state",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}
	stateJSON, _ := json.Marshal(stateData)
	return string(stateJSON)
}

func TestGoogleOAuthE2ETestSuite(t *testing.T) {
	suite.Run(t, new(GoogleOAuthE2ETestSuite))
}
