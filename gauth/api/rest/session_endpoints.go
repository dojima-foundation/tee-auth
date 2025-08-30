package rest

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SessionInfoResponse represents the response for session info
type SessionInfoResponse struct {
	SessionID      string    `json:"session_id"`
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	OAuthProvider  string    `json:"oauth_provider,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivity   time.Time `json:"last_activity"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// SessionRefreshResponse represents the response for session refresh
type SessionRefreshResponse struct {
	Message   string    `json:"message"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SessionLogoutResponse represents the response for session logout
type SessionLogoutResponse struct {
	Message string `json:"message"`
}

// handleSessionInfo returns information about the current session
func (s *Server) handleSessionInfo(c *gin.Context) {
	sessionData, exists := GetSessionFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse(nil, "No active session"))
		return
	}

	response := SessionInfoResponse{
		SessionID:      c.GetString("session_id"), // This will be set by middleware
		UserID:         sessionData.UserID,
		OrganizationID: sessionData.OrganizationID,
		Email:          sessionData.Email,
		Role:           sessionData.Role,
		OAuthProvider:  sessionData.OAuthProvider,
		CreatedAt:      sessionData.CreatedAt,
		LastActivity:   sessionData.LastActivity,
		ExpiresAt:      sessionData.ExpiresAt,
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// handleSessionRefresh extends the current session
func (s *Server) handleSessionRefresh(c *gin.Context) {
	sessionID := s.sessionManager.GetSessionFromRequest(c)
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, errorResponse(nil, "Session required"))
		return
	}

	err := s.sessionManager.RefreshSession(c.Request.Context(), sessionID)
	if err != nil {
		s.logger.Error("Failed to refresh session", "error", err)
		c.JSON(http.StatusUnauthorized, errorResponse(err, "Failed to refresh session"))
		return
	}

	// Get updated session data
	sessionData, err := s.sessionManager.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to get session data"))
		return
	}

	response := SessionRefreshResponse{
		Message:   "Session refreshed successfully",
		ExpiresAt: sessionData.ExpiresAt,
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// handleSessionLogout destroys the current session
func (s *Server) handleSessionLogout(c *gin.Context) {
	sessionID := s.sessionManager.GetSessionFromRequest(c)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "No session to logout"))
		return
	}

	err := s.sessionManager.DestroySession(c.Request.Context(), sessionID)
	if err != nil {
		s.logger.Error("Failed to destroy session", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to logout"))
		return
	}

	// Clear session cookie
	s.sessionManager.ClearSessionCookie(c)

	response := SessionLogoutResponse{
		Message: "Logged out successfully",
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// handleSessionValidate validates a session without requiring authentication
func (s *Server) handleSessionValidate(c *gin.Context) {
	sessionID := s.sessionManager.GetSessionFromRequest(c)
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, errorResponse(nil, "Session required"))
		return
	}

	sessionData, err := s.sessionManager.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		s.logger.Warn("Session validation failed", "error", err)
		c.JSON(http.StatusUnauthorized, errorResponse(err, "Invalid or expired session"))
		return
	}

	response := SessionInfoResponse{
		SessionID:      sessionID,
		UserID:         sessionData.UserID,
		OrganizationID: sessionData.OrganizationID,
		Email:          sessionData.Email,
		Role:           sessionData.Role,
		OAuthProvider:  sessionData.OAuthProvider,
		CreatedAt:      sessionData.CreatedAt,
		LastActivity:   sessionData.LastActivity,
		ExpiresAt:      sessionData.ExpiresAt,
	}

	c.JSON(http.StatusOK, successResponse(response))
}

// handleSessionList returns all active sessions for the current user (admin only)
func (s *Server) handleSessionList(c *gin.Context) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse(nil, "User ID not found in session"))
		return
	}

	// Get all sessions for this user from Redis
	// This is a simplified implementation - in production you might want to
	// store user sessions in a separate index for better performance
	sessionKeys, err := s.redis.GetClient().Keys(c.Request.Context(), "session:*").Result()
	if err != nil {
		s.logger.Error("Failed to get session keys", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to list sessions"))
		return
	}

	var userSessions []SessionInfoResponse
	for _, key := range sessionKeys {
		sessionDataJSON, err := s.redis.Get(c.Request.Context(), key)
		if err != nil {
			continue // Skip invalid sessions
		}

		var sessionData SessionData
		if err := json.Unmarshal([]byte(sessionDataJSON), &sessionData); err != nil {
			continue // Skip invalid session data
		}

		// Only include sessions for the current user
		if sessionData.UserID == userID {
			sessionID := strings.TrimPrefix(key, "session:")
			userSessions = append(userSessions, SessionInfoResponse{
				SessionID:      sessionID,
				UserID:         sessionData.UserID,
				OrganizationID: sessionData.OrganizationID,
				Email:          sessionData.Email,
				Role:           sessionData.Role,
				OAuthProvider:  sessionData.OAuthProvider,
				CreatedAt:      sessionData.CreatedAt,
				LastActivity:   sessionData.LastActivity,
				ExpiresAt:      sessionData.ExpiresAt,
			})
		}
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"sessions": userSessions,
		"count":    len(userSessions),
	}))
}

// handleSessionDestroy destroys a specific session by ID (admin only)
func (s *Server) handleSessionDestroy(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "Session ID is required"))
		return
	}

	// Get current user ID for authorization
	currentUserID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse(nil, "User ID not found in session"))
		return
	}

	// Validate that the session belongs to the current user
	sessionData, err := s.sessionManager.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(err, "Session not found"))
		return
	}

	if sessionData.UserID != currentUserID {
		c.JSON(http.StatusForbidden, errorResponse(nil, "Cannot destroy another user's session"))
		return
	}

	err = s.sessionManager.DestroySession(c.Request.Context(), sessionID)
	if err != nil {
		s.logger.Error("Failed to destroy session", "session_id", sessionID, "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to destroy session"))
		return
	}

	response := SessionLogoutResponse{
		Message: "Session destroyed successfully",
	}

	c.JSON(http.StatusOK, successResponse(response))
}
