package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

// GoogleOAuthService handles Google OAuth authentication
type GoogleOAuthService struct {
	*GAuthService
	config     *oauth2.Config
	apiBaseURL string // For testing, can be overridden to point to mock server
}

// GoogleUserInfo represents user information from Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

// GoogleOAuthData represents the data stored in AuthMethod for Google OAuth
type GoogleOAuthData struct {
	GoogleUserID string    `json:"google_user_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Picture      string    `json:"picture"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

// NewGoogleOAuthService creates a new Google OAuth service
func NewGoogleOAuthService(gauthService *GAuthService, cfg *config.Config) *GoogleOAuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.Auth.GoogleOAuth.ClientID,
		ClientSecret: cfg.Auth.GoogleOAuth.ClientSecret,
		RedirectURL:  cfg.Auth.GoogleOAuth.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOAuthService{
		GAuthService: gauthService,
		config:       oauthConfig,
		apiBaseURL:   "https://www.googleapis.com",
	}
}

// NewGoogleOAuthServiceWithEndpoint creates a new Google OAuth service with custom endpoint (for testing)
func NewGoogleOAuthServiceWithEndpoint(gauthService *GAuthService, cfg *config.Config, endpoint oauth2.Endpoint) *GoogleOAuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.Auth.GoogleOAuth.ClientID,
		ClientSecret: cfg.Auth.GoogleOAuth.ClientSecret,
		RedirectURL:  cfg.Auth.GoogleOAuth.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: endpoint,
	}

	return &GoogleOAuthService{
		GAuthService: gauthService,
		config:       oauthConfig,
		apiBaseURL:   "https://www.googleapis.com",
	}
}

// NewGoogleOAuthServiceWithEndpointAndAPIBase creates a new Google OAuth service with custom endpoint and API base URL (for testing)
func NewGoogleOAuthServiceWithEndpointAndAPIBase(gauthService *GAuthService, cfg *config.Config, endpoint oauth2.Endpoint, apiBaseURL string) *GoogleOAuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.Auth.GoogleOAuth.ClientID,
		ClientSecret: cfg.Auth.GoogleOAuth.ClientSecret,
		RedirectURL:  cfg.Auth.GoogleOAuth.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: endpoint,
	}

	return &GoogleOAuthService{
		GAuthService: gauthService,
		config:       oauthConfig,
		apiBaseURL:   apiBaseURL,
	}
}

// GetAuthURL generates the Google OAuth authorization URL
func (s *GoogleOAuthService) GetAuthURL(ctx context.Context, organizationID, state string) (string, error) {
	s.logger.Info("Generating Google OAuth auth URL", "organization_id", organizationID)

	// Create state parameter that includes organization ID (optional for new users)
	stateData := map[string]string{
		"state":     state,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}

	// Only include organization_id if it's provided and valid
	if organizationID != "" {
		if _, err := uuid.Parse(organizationID); err != nil {
			return "", fmt.Errorf("invalid organization ID: %w", err)
		}
		stateData["organization_id"] = organizationID
	}

	stateJSON, err := json.Marshal(stateData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state: %w", err)
	}

	authURL := s.config.AuthCodeURL(string(stateJSON), oauth2.AccessTypeOffline)
	return authURL, nil
}

// HandleCallback processes the Google OAuth callback
func (s *GoogleOAuthService) HandleCallback(ctx context.Context, code, state string) (*GoogleOAuthResponse, error) {
	stateLog := state
	if len(state) > 8 {
		stateLog = state[:8] + "..."
	}
	s.logger.Info("Processing Google OAuth callback", "state", stateLog)

	// Parse state parameter
	var stateData map[string]string
	if err := json.Unmarshal([]byte(state), &stateData); err != nil {
		return nil, fmt.Errorf("invalid state parameter: %w", err)
	}

	// For new users, organization_id might not be provided as we'll create one
	organizationID, exists := stateData["organization_id"]
	if !exists {
		organizationID = "" // Will be created during user creation
	}

	// Exchange code for token
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		s.logger.Error("Failed to exchange code for token", "error", err)
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info from Google
	userInfo, err := s.getGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		s.logger.Error("Failed to get Google user info", "error", err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, authMethod, err := s.findOrCreateUser(ctx, organizationID, userInfo, token)
	if err != nil {
		s.logger.Error("Failed to find or create user", "error", err)
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Generate session token
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Store session in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionToken)
	sessionData := map[string]interface{}{
		"user_id":         user.ID.String(),
		"organization_id": user.OrganizationID.String(),
		"auth_method_id":  authMethod.ID.String(),
		"expires_at":      expiresAt.Unix(),
		"oauth_provider":  "google",
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

	response := &GoogleOAuthResponse{
		Success:      true,
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt,
		User:         user,
		AuthMethod:   authMethod,
	}

	s.logger.Info("Google OAuth authentication successful",
		"user_id", user.ID,
		"email", user.Email,
		"organization_id", organizationID)

	return response, nil
}

// getGoogleUserInfo fetches user information from Google
func (s *GoogleOAuthService) getGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET",
		s.apiBaseURL+"/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API returned status: %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &userInfo, nil
}

// findOrCreateUser finds an existing user or creates a new one
func (s *GoogleOAuthService) findOrCreateUser(ctx context.Context, organizationID string,
	googleUser *GoogleUserInfo, token *oauth2.Token) (*models.User, *models.AuthMethod, error) {

	// Check if user already exists with this Google ID
	var existingAuthMethod models.AuthMethod
	err := s.db.GetDB().WithContext(ctx).
		Where("type = ? AND data->>'google_user_id' = ?", "OAUTH", googleUser.ID).
		First(&existingAuthMethod).Error

	if err == nil {
		// User exists, update token information
		oauthData := GoogleOAuthData{
			GoogleUserID: googleUser.ID,
			Email:        googleUser.Email,
			Name:         googleUser.Name,
			Picture:      googleUser.Picture,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.Expiry,
		}

		dataJSON, _ := json.Marshal(oauthData)
		existingAuthMethod.Data = string(dataJSON)

		if err := s.db.GetDB().WithContext(ctx).Save(&existingAuthMethod).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to update auth method: %w", err)
		}

		// Get user
		var user models.User
		if err := s.db.GetDB().WithContext(ctx).First(&user, existingAuthMethod.UserID).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to get user: %w", err)
		}

		return &user, &existingAuthMethod, nil
	}

	// User doesn't exist, create new organization and user
	// Handle potential race conditions by implementing a simple retry mechanism
	var orgUUID uuid.UUID
	var user models.User
	var authMethod models.AuthMethod

	// Try up to 3 times to handle race conditions
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = s.db.Transaction(func(tx *gorm.DB) error {
			// Check if organization ID is provided and valid
			if organizationID != "" {
				parsedOrgID, err := uuid.Parse(organizationID)
				if err != nil {
					return fmt.Errorf("invalid organization ID: %w", err)
				}

				// Verify organization exists
				var existingOrg models.Organization
				if err := tx.First(&existingOrg, parsedOrgID).Error; err != nil {
					return fmt.Errorf("organization not found: %w", err)
				}

				orgUUID = parsedOrgID
			} else {
				// Create organization for new user
				org := &models.Organization{
					ID:      uuid.New(),
					Version: "1.0",
					Name:    fmt.Sprintf("%s's Organization", googleUser.Name),
					RootQuorum: models.Quorum{
						Threshold: s.GAuthService.config.Auth.DefaultQuorumThreshold,
					},
				}

				if err := tx.Create(org).Error; err != nil {
					return fmt.Errorf("failed to create organization: %w", err)
				}

				orgUUID = org.ID
			}

			// Create Root user
			user = models.User{
				ID:             uuid.New(),
				OrganizationID: orgUUID,
				Username:       "Root user",
				Email:          googleUser.Email,
				PublicKey:      "", // Will be generated or set later
				IsActive:       true,
			}

			// Check if email is already taken globally
			var existingUser models.User
			err := tx.Where("email = ?", googleUser.Email).First(&existingUser).Error
			if err == nil {
				// Email exists globally, generate a unique email and username
				baseUsername := "Root user"
				counter := 1
				for {
					newUsername := fmt.Sprintf("%s_%d", baseUsername, counter)
					newEmail := fmt.Sprintf("%s+%d@%s",
						strings.Split(googleUser.Email, "@")[0],
						counter,
						strings.Split(googleUser.Email, "@")[1])

					// Check if both username and email are available
					var existingUserWithUsername models.User
					var existingUserWithEmail models.User
					err1 := tx.Where("organization_id = ? AND username = ?", orgUUID, newUsername).First(&existingUserWithUsername).Error
					err2 := tx.Where("email = ?", newEmail).First(&existingUserWithEmail).Error

					if err1 != nil && err2 != nil {
						user.Username = newUsername
						user.Email = newEmail
						break
					}
					counter++
				}
			}

			// Create user
			if err := tx.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}

			// Add user to root quorum
			if err := tx.Create(&QuorumMember{
				OrganizationID: orgUUID,
				UserID:         user.ID,
			}).Error; err != nil {
				return fmt.Errorf("failed to add user to quorum: %w", err)
			}

			// Create auth method
			oauthData := GoogleOAuthData{
				GoogleUserID: googleUser.ID,
				Email:        googleUser.Email,
				Name:         googleUser.Name,
				Picture:      googleUser.Picture,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				ExpiresAt:    token.Expiry,
			}

			dataJSON, _ := json.Marshal(oauthData)
			authMethod = models.AuthMethod{
				ID:       uuid.New(),
				UserID:   user.ID,
				Type:     "OAUTH",
				Name:     "Google OAuth",
				Data:     string(dataJSON),
				IsActive: true,
			}

			if err := tx.Create(&authMethod).Error; err != nil {
				return fmt.Errorf("failed to create auth method: %w", err)
			}

			return nil
		})

		if err != nil {
			// Check if this is a duplicate key error (race condition)
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") ||
				strings.Contains(err.Error(), "UNIQUE constraint failed") {
				// Another process might have created the user, check again
				var existingAuthMethod models.AuthMethod
				checkErr := s.db.GetDB().WithContext(ctx).
					Where("type = ? AND data->>'google_user_id' = ?", "OAUTH", googleUser.ID).
					First(&existingAuthMethod).Error

				if checkErr == nil {
					// User was created by another process, return it
					var existingUser models.User
					if getErr := s.db.GetDB().WithContext(ctx).First(&existingUser, existingAuthMethod.UserID).Error; getErr == nil {
						return &existingUser, &existingAuthMethod, nil
					}
				}

				// If we can't find the existing user, return the original error
				if attempt == maxRetries-1 {
					return nil, nil, err
				}
				// Wait a bit before retrying
				time.Sleep(time.Duration(attempt+1) * 50 * time.Millisecond)
				continue
			}
			return nil, nil, err
		}

		break // Success, exit retry loop
	}

	if err != nil {
		return nil, nil, err
	}

	s.logger.Info("Created new organization and Root user for Google OAuth signin",
		"organization_id", orgUUID.String(),
		"user_id", user.ID.String(),
		"email", googleUser.Email)

	return &user, &authMethod, nil
}

// RefreshToken refreshes the Google OAuth token
func (s *GoogleOAuthService) RefreshToken(ctx context.Context, authMethodID string) error {
	s.logger.Info("Refreshing Google OAuth token", "auth_method_id", authMethodID)

	authMethodUUID, err := uuid.Parse(authMethodID)
	if err != nil {
		return fmt.Errorf("invalid auth method ID: %w", err)
	}

	var authMethod models.AuthMethod
	if err := s.db.GetDB().WithContext(ctx).First(&authMethod, authMethodUUID).Error; err != nil {
		return fmt.Errorf("auth method not found: %w", err)
	}

	if authMethod.Type != "OAUTH" {
		return fmt.Errorf("auth method is not OAuth type")
	}

	var oauthData GoogleOAuthData
	if err := json.Unmarshal([]byte(authMethod.Data), &oauthData); err != nil {
		return fmt.Errorf("invalid OAuth data: %w", err)
	}

	if oauthData.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Create token source
	token := &oauth2.Token{
		RefreshToken: oauthData.RefreshToken,
	}

	tokenSource := s.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update stored token
	oauthData.AccessToken = newToken.AccessToken
	oauthData.ExpiresAt = newToken.Expiry
	if newToken.RefreshToken != "" {
		oauthData.RefreshToken = newToken.RefreshToken
	}

	dataJSON, _ := json.Marshal(oauthData)
	authMethod.Data = string(dataJSON)

	if err := s.db.GetDB().WithContext(ctx).Save(&authMethod).Error; err != nil {
		return fmt.Errorf("failed to update auth method: %w", err)
	}

	s.logger.Info("Google OAuth token refreshed successfully", "auth_method_id", authMethodID)
	return nil
}

// GoogleOAuthResponse represents the response from Google OAuth authentication
type GoogleOAuthResponse struct {
	Success      bool               `json:"success"`
	SessionToken string             `json:"session_token"`
	ExpiresAt    time.Time          `json:"expires_at"`
	User         *models.User       `json:"user"`
	AuthMethod   *models.AuthMethod `json:"auth_method"`
}
