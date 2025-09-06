package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionData represents the data stored in a session
type SessionData struct {
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	AuthMethodID   string    `json:"auth_method_id"`
	OAuthProvider  string    `json:"oauth_provider,omitempty"`
	Email          string    `json:"email,omitempty"`
	Role           string    `json:"role,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivity   time.Time `json:"last_activity"`
	ExpiresAt      time.Time `json:"expires_at"`
	UserAgent      string    `json:"user_agent,omitempty"`
	ClientIP       string    `json:"client_ip,omitempty"`
}

// SessionManager handles session operations
type SessionManager struct {
	server           *Server
	testMode         bool
	testOrgID        string
	testUserID       string
	testAuthMethodID string
}

// NewSessionManager creates a new session manager
func NewSessionManager(server *Server) *SessionManager {
	return &SessionManager{
		server: server,
	}
}

// SetTestMode enables test mode which bypasses session validation
func (sm *SessionManager) SetTestMode(testMode bool) {
	sm.testMode = testMode
}

// SetTestData sets test data for session validation bypass
func (sm *SessionManager) SetTestData(orgID, userID, authMethodID string) {
	sm.testOrgID = orgID
	sm.testUserID = userID
	sm.testAuthMethodID = authMethodID
}

// CreateSession creates a new session and stores it in Redis
func (sm *SessionManager) CreateSession(ctx context.Context, user *models.User, authMethod *models.AuthMethod, oauthProvider string) (string, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // 24 hour session

	sessionData := SessionData{
		UserID:         user.ID.String(),
		OrganizationID: user.OrganizationID.String(),
		AuthMethodID:   authMethod.ID.String(),
		OAuthProvider:  oauthProvider,
		Email:          user.Email,
		Role:           "root_user", // Default role for OAuth users
		CreatedAt:      now,
		LastActivity:   now,
		ExpiresAt:      expiresAt,
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := sm.server.redis.Set(ctx, sessionKey, string(sessionDataJSON), 24*time.Hour); err != nil {
		return "", fmt.Errorf("failed to store session in Redis: %w", err)
	}

	sm.server.logger.Info("Session created successfully",
		"session_id", sessionID,
		"user_id", user.ID,
		"organization_id", user.OrganizationID,
		"oauth_provider", oauthProvider)

	return sessionID, nil
}

// ValidateSession validates a session and returns session data
func (sm *SessionManager) ValidateSession(ctx context.Context, sessionID string) (*SessionData, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	sessionDataJSON, err := sm.server.redis.Get(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("session not found or expired")
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(sessionDataJSON), &sessionData); err != nil {
		return nil, fmt.Errorf("invalid session data format")
	}

	// Check if session is expired
	if time.Now().After(sessionData.ExpiresAt) {
		// Clean up expired session
		_ = sm.server.redis.Delete(ctx, sessionKey)
		return nil, fmt.Errorf("session expired")
	}

	// Update last activity
	sessionData.LastActivity = time.Now()
	updatedSessionDataJSON, err := json.Marshal(sessionData)
	if err == nil {
		_ = sm.server.redis.Set(ctx, sessionKey, string(updatedSessionDataJSON), 24*time.Hour)
	}

	return &sessionData, nil
}

// RefreshSession extends the session expiration time
func (sm *SessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	sessionData, err := sm.ValidateSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Extend session by 24 hours
	sessionData.ExpiresAt = time.Now().Add(24 * time.Hour)
	sessionData.LastActivity = time.Now()

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := sm.server.redis.Set(ctx, sessionKey, string(sessionDataJSON), 24*time.Hour); err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	sm.server.logger.Info("Session refreshed successfully", "session_id", sessionID)
	return nil
}

// DestroySession removes a session from Redis
func (sm *SessionManager) DestroySession(ctx context.Context, sessionID string) error {
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	if err := sm.server.redis.Delete(ctx, sessionKey); err != nil {
		return fmt.Errorf("failed to destroy session: %w", err)
	}

	sm.server.logger.Info("Session destroyed successfully", "session_id", sessionID)
	return nil
}

// GetSessionFromRequest extracts session ID from request (cookie or header)
func (sm *SessionManager) GetSessionFromRequest(c *gin.Context) string {
	// First try to get from cookie
	if cookie, err := c.Cookie("gauth_session"); err == nil && cookie != "" {
		return cookie
	}

	// Then try to get from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Finally try to get from X-Session-Token header
	return c.GetHeader("X-Session-Token")
}

// SetSessionCookie sets the session cookie in the response
func (sm *SessionManager) SetSessionCookie(c *gin.Context, sessionID string) {
	// Determine cookie domain based on environment
	domain := ""
	if os.Getenv("ENVIRONMENT") == "production" {
		domain = ".dojima.foundation" // Shared parent domain for subdomains
	}

	c.SetCookie(
		"gauth_session",                      // name
		sessionID,                            // value
		24*60*60,                             // maxAge (24 hours in seconds)
		"/",                                  // path
		domain,                               // domain
		sm.server.config.Security.TLSEnabled, // secure (HTTPS only in production)
		true,                                 // httpOnly (prevent XSS)
	)
}

// ClearSessionCookie removes the session cookie
func (sm *SessionManager) ClearSessionCookie(c *gin.Context) {
	domain := ""
	if os.Getenv("ENVIRONMENT") == "production" {
		domain = ".dojima.foundation"
	}

	c.SetCookie(
		"gauth_session",                      // name
		"",                                   // value
		-1,                                   // maxAge (expire immediately)
		"/",                                  // path
		domain,                               // domain
		sm.server.config.Security.TLSEnabled, // secure
		true,                                 // httpOnly
	)
}

// SessionMiddleware validates sessions for protected routes
func (sm *SessionManager) SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In test mode, bypass session validation
		if sm.testMode {
			// Set mock session data for testing
			c.Set("session", &SessionData{
				UserID:         sm.testUserID,
				OrganizationID: sm.testOrgID,
				AuthMethodID:   sm.testAuthMethodID,
				OAuthProvider:  "google",
				Role:           "root_user",
				CreatedAt:      time.Now(),
				LastActivity:   time.Now(),
				ExpiresAt:      time.Now().Add(24 * time.Hour),
			})
			c.Set("user_id", sm.testUserID)
			c.Set("organization_id", sm.testOrgID)
			c.Set("auth_method_id", sm.testAuthMethodID)
			c.Next()
			return
		}

		sessionID := sm.GetSessionFromRequest(c)
		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, errorResponse(nil, "Session required"))
			c.Abort()
			return
		}

		sessionData, err := sm.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			sm.server.logger.Warn("Session validation failed",
				"session_id", sessionID[:8]+"...",
				"error", err,
				"ip", c.ClientIP(),
				"user_agent", c.GetHeader("User-Agent"))

			c.JSON(http.StatusUnauthorized, errorResponse(err, "Invalid or expired session"))
			c.Abort()
			return
		}

		// Store session data in context for use in handlers
		c.Set("session", sessionData)
		c.Set("user_id", sessionData.UserID)
		c.Set("organization_id", sessionData.OrganizationID)
		c.Set("auth_method_id", sessionData.AuthMethodID)

		// Update session activity
		sessionData.LastActivity = time.Now()
		sessionKey := fmt.Sprintf("session:%s", sessionID)
		sessionDataJSON, _ := json.Marshal(sessionData)
		_ = sm.server.redis.Set(c.Request.Context(), sessionKey, string(sessionDataJSON), 24*time.Hour)

		c.Next()
	}
}

// OptionalSessionMiddleware validates sessions but doesn't require them
func (sm *SessionManager) OptionalSessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := sm.GetSessionFromRequest(c)
		if sessionID != "" {
			sessionData, err := sm.ValidateSession(c.Request.Context(), sessionID)
			if err == nil {
				// Store session data in context if valid
				c.Set("session", sessionData)
				c.Set("user_id", sessionData.UserID)
				c.Set("organization_id", sessionData.OrganizationID)
				c.Set("auth_method_id", sessionData.AuthMethodID)
			}
		}
		c.Next()
	}
}

// GetSessionFromContext retrieves session data from Gin context
func GetSessionFromContext(c *gin.Context) (*SessionData, bool) {
	session, exists := c.Get("session")
	if !exists {
		return nil, false
	}
	sessionData, ok := session.(*SessionData)
	return sessionData, ok
}

// GetUserIDFromContext retrieves user ID from Gin context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// GetOrganizationIDFromContext retrieves organization ID from Gin context
func GetOrganizationIDFromContext(c *gin.Context) (string, bool) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		return "", false
	}
	orgIDStr, ok := orgID.(string)
	return orgIDStr, ok
}
