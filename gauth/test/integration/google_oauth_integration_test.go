package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
)

type GoogleOAuthIntegrationTestSuite struct {
	suite.Suite
	config      *config.Config
	database    *db.PostgresDB
	redis       *db.RedisClient
	service     *service.GAuthService
	googleOAuth *service.GoogleOAuthService
	mockServer  *httptest.Server
}

func (suite *GoogleOAuthIntegrationTestSuite) SetupSuite() {
	// Skip integration tests if not in integration test mode
	if testhelpers.GetEnvOrDefault("INTEGRATION_TESTS", "") != "true" {
		suite.T().Skip("Skipping integration tests. Set INTEGRATION_TESTS=true to run.")
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
			Database:     3, // Different database for Google OAuth tests
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
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
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

func (suite *GoogleOAuthIntegrationTestSuite) TearDownSuite() {
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

func (suite *GoogleOAuthIntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.cleanupDatabase()
}

func (suite *GoogleOAuthIntegrationTestSuite) setupMockGoogleServer() {
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/authorize":
			// Mock OAuth2 authorization endpoint - redirect with code
			code := r.URL.Query().Get("code")
			if code == "" {
				code = "mock-auth-code"
			}
			state := r.URL.Query().Get("state")
			redirectURI := r.URL.Query().Get("redirect_uri")

			// Redirect to callback with code and state
			http.Redirect(w, r, redirectURI+"?code="+code+"&state="+state, http.StatusFound)

		case "/oauth2/token":
			// Mock OAuth2 token exchange
			w.Header().Set("Content-Type", "application/json")
			tokenResponse := map[string]interface{}{
				"access_token":  "mock-access-token",
				"refresh_token": "mock-refresh-token",
				"token_type":    "Bearer",
				"expires_in":    3600,
			}
			json.NewEncoder(w).Encode(tokenResponse)

		case "/oauth2/v2/userinfo":
			// Mock Google user info endpoint
			w.Header().Set("Content-Type", "application/json")
			userInfo := map[string]interface{}{
				"id":             "mock-google-user-123",
				"email":          "test@example.com",
				"name":           "Test User",
				"given_name":     "Test",
				"family_name":    "User",
				"picture":        "https://example.com/avatar.jpg",
				"verified_email": true,
			}
			json.NewEncoder(w).Encode(userInfo)

		default:
			http.NotFound(w, r)
		}
	}))
}

func (suite *GoogleOAuthIntegrationTestSuite) cleanupDatabase() {
	ctx := context.Background()

	// Clean up in reverse order of dependencies
	suite.database.GetDB().WithContext(ctx).Exec("DELETE FROM quorum_members")
	suite.database.GetDB().WithContext(ctx).Exec("DELETE FROM auth_methods")
	suite.database.GetDB().WithContext(ctx).Exec("DELETE FROM users")
	suite.database.GetDB().WithContext(ctx).Exec("DELETE FROM organizations")
}

func (suite *GoogleOAuthIntegrationTestSuite) TestGoogleOAuthURLGeneration() {
	ctx := context.Background()

	tests := []struct {
		name           string
		organizationID string
		state          string
		expectError    bool
	}{
		{
			name:           "with valid organization ID",
			organizationID: uuid.New().String(),
			state:          "test-state",
			expectError:    false,
		},
		{
			name:           "without organization ID (new user)",
			organizationID: "",
			state:          "test-state",
			expectError:    false,
		},
		{
			name:           "with invalid organization ID",
			organizationID: "invalid-uuid",
			state:          "test-state",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			authURL, err := suite.googleOAuth.GetAuthURL(ctx, tt.organizationID, tt.state)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, authURL)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, authURL)
				assert.Contains(t, authURL, "oauth2/auth")
				assert.Contains(t, authURL, "test-client-id")
			}
		})
	}
}

func (suite *GoogleOAuthIntegrationTestSuite) TestGoogleOAuthCallback_NewUser() {
	ctx := context.Background()

	// Test callback for new user (no organization ID)
	state := suite.createStateWithoutOrgID()
	code := "mock-auth-code"

	response, err := suite.googleOAuth.HandleCallback(ctx, code, state)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.SessionToken)
	assert.NotNil(suite.T(), response.User)
	assert.NotNil(suite.T(), response.AuthMethod)

	// Verify user was created with correct details
	assert.Equal(suite.T(), "Root user", response.User.Username)
	assert.Equal(suite.T(), "test@example.com", response.User.Email)
	assert.True(suite.T(), response.User.IsActive)

	// Verify organization was created
	assert.NotEmpty(suite.T(), response.User.OrganizationID)

	// Verify auth method was created
	assert.Equal(suite.T(), "OAUTH", response.AuthMethod.Type)
	assert.Equal(suite.T(), "Google OAuth", response.AuthMethod.Name)
	assert.True(suite.T(), response.AuthMethod.IsActive)

	// Verify user is in root quorum
	var quorumMember service.QuorumMember
	err = suite.database.GetDB().WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", response.User.OrganizationID, response.User.ID).
		First(&quorumMember).Error
	assert.NoError(suite.T(), err)
}

func (suite *GoogleOAuthIntegrationTestSuite) TestGoogleOAuthCallback_ExistingUser() {
	ctx := context.Background()

	// First, create a user through OAuth
	state1 := suite.createStateWithoutOrgID()
	code1 := "mock-auth-code-1"

	response1, err := suite.googleOAuth.HandleCallback(ctx, code1, state1)
	require.NoError(suite.T(), err)

	// Now test callback for the same user (should find existing user)
	state2 := suite.createStateWithOrgID(response1.User.OrganizationID.String())
	code2 := "mock-auth-code-2"

	response2, err := suite.googleOAuth.HandleCallback(ctx, code2, state2)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response2)
	assert.True(suite.T(), response2.Success)
	assert.NotEmpty(suite.T(), response2.SessionToken)

	// Should be the same user
	assert.Equal(suite.T(), response1.User.ID, response2.User.ID)
	assert.Equal(suite.T(), response1.User.OrganizationID, response2.User.OrganizationID)
}

func (suite *GoogleOAuthIntegrationTestSuite) TestGoogleOAuthCallback_InvalidState() {
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
			code:          "mock-auth-code",
			state:         "invalid-state",
			expectError:   true,
			errorContains: "invalid state parameter",
		},
		{
			name:          "empty state",
			code:          "mock-auth-code",
			state:         "",
			expectError:   true,
			errorContains: "invalid state parameter",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			response, err := suite.googleOAuth.HandleCallback(ctx, tt.code, tt.state)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
			assert.Nil(t, response)
		})
	}
}

func (suite *GoogleOAuthIntegrationTestSuite) TestGoogleOAuthCallback_OrganizationCreation() {
	ctx := context.Background()

	// Test that organization is created with correct details
	state := suite.createStateWithoutOrgID()
	code := "mock-auth-code"

	response, err := suite.googleOAuth.HandleCallback(ctx, code, state)
	require.NoError(suite.T(), err)

	// Verify organization was created
	var org models.Organization
	err = suite.database.GetDB().WithContext(ctx).
		Where("id = ?", response.User.OrganizationID).
		First(&org).Error
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Test User's Organization", org.Name)
	assert.Equal(suite.T(), "1.0", org.Version)
	assert.Equal(suite.T(), 1, org.RootQuorum.Threshold)
}

func (suite *GoogleOAuthIntegrationTestSuite) TestGoogleOAuthCallback_DuplicateEmail() {
	ctx := context.Background()

	// Create first user
	state1 := suite.createStateWithoutOrgID()
	code1 := "mock-auth-code-1"

	response1, err := suite.googleOAuth.HandleCallback(ctx, code1, state1)
	require.NoError(suite.T(), err)

	// Try to create another user with the same email but different Google ID
	// This should create a new user with a modified username
	state2 := suite.createStateWithoutOrgID()
	code2 := "mock-auth-code-2"

	// Modify the mock server to return different Google ID but same email
	originalHandler := suite.mockServer.Config.Handler
	suite.mockServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/v2/userinfo" {
			w.Header().Set("Content-Type", "application/json")
			userInfo := map[string]interface{}{
				"id":             "mock-google-user-456", // Different ID
				"email":          "test@example.com",     // Same email
				"name":           "Test User",
				"given_name":     "Test",
				"family_name":    "User",
				"picture":        "https://example.com/avatar.jpg",
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

	// Should be different users
	assert.NotEqual(suite.T(), response1.User.ID, response2.User.ID)
	assert.NotEqual(suite.T(), response1.User.Username, response2.User.Username)
	// Email should be modified to avoid global uniqueness constraint
	assert.Equal(suite.T(), "test+1@example.com", response2.User.Email)
	assert.Contains(suite.T(), response2.User.Username, "Root user")
}

func (suite *GoogleOAuthIntegrationTestSuite) TestGoogleOAuthCallback_SessionCreation() {
	ctx := context.Background()

	state := suite.createStateWithoutOrgID()
	code := "mock-auth-code"

	response, err := suite.googleOAuth.HandleCallback(ctx, code, state)
	require.NoError(suite.T(), err)

	// Verify session was created in Redis
	sessionData, err := suite.redis.Get(ctx, fmt.Sprintf("session:%s", response.SessionToken))
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), sessionData)

	// Parse session data
	var session map[string]interface{}
	err = json.Unmarshal([]byte(sessionData), &session)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), response.User.ID.String(), session["user_id"])
	assert.Equal(suite.T(), response.User.OrganizationID.String(), session["organization_id"])
	assert.Equal(suite.T(), response.AuthMethod.ID.String(), session["auth_method_id"])
	assert.Equal(suite.T(), "google", session["oauth_provider"])
}

func (suite *GoogleOAuthIntegrationTestSuite) TestGoogleOAuthCallback_ExistingOrganization() {
	ctx := context.Background()

	// Create an organization first
	org, err := suite.service.CreateOrganization(ctx, "Existing Org", "admin@example.com", "admin-public-key")
	require.NoError(suite.T(), err)

	// Test callback with existing organization ID
	state := suite.createStateWithOrgID(org.ID.String())
	code := "mock-auth-code"

	response, err := suite.googleOAuth.HandleCallback(ctx, code, state)
	require.NoError(suite.T(), err)

	// Should create user in existing organization
	assert.Equal(suite.T(), org.ID, response.User.OrganizationID)
	assert.Equal(suite.T(), "Root user", response.User.Username)
}

func (suite *GoogleOAuthIntegrationTestSuite) TestGoogleOAuthData_Serialization() {
	// Test that GoogleOAuthData can be properly serialized and deserialized
	originalData := &service.GoogleOAuthData{
		GoogleUserID: "mock-google-user-123",
		Email:        "test@example.com",
		Name:         "Test User",
		Picture:      "https://example.com/avatar.jpg",
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	// Serialize
	dataJSON, err := json.Marshal(originalData)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), dataJSON)

	// Deserialize
	var deserializedData service.GoogleOAuthData
	err = json.Unmarshal(dataJSON, &deserializedData)
	assert.NoError(suite.T(), err)

	// Verify all fields
	assert.Equal(suite.T(), originalData.GoogleUserID, deserializedData.GoogleUserID)
	assert.Equal(suite.T(), originalData.Email, deserializedData.Email)
	assert.Equal(suite.T(), originalData.Name, deserializedData.Name)
	assert.Equal(suite.T(), originalData.Picture, deserializedData.Picture)
	assert.Equal(suite.T(), originalData.AccessToken, deserializedData.AccessToken)
	assert.Equal(suite.T(), originalData.RefreshToken, deserializedData.RefreshToken)
	assert.WithinDuration(suite.T(), originalData.ExpiresAt, deserializedData.ExpiresAt, time.Second)
}

// Helper functions

func (suite *GoogleOAuthIntegrationTestSuite) createStateWithOrgID(orgID string) string {
	stateData := map[string]string{
		"organization_id": orgID,
		"state":           "test-state",
		"timestamp":       fmt.Sprintf("%d", time.Now().Unix()),
	}
	stateJSON, _ := json.Marshal(stateData)
	return string(stateJSON)
}

func (suite *GoogleOAuthIntegrationTestSuite) createStateWithoutOrgID() string {
	stateData := map[string]string{
		"state":     "test-state",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}
	stateJSON, _ := json.Marshal(stateData)
	return string(stateJSON)
}

func TestGoogleOAuthIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleOAuthIntegrationTestSuite))
}
