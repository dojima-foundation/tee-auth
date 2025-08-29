package rest

import (
	"fmt"
	"net/http"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GoogleOAuthRequest represents the request payload for Google OAuth
type GoogleOAuthRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	State          string `json:"state,omitempty"`
}

// GoogleOAuthCallbackRequest represents the callback request
type GoogleOAuthCallbackRequest struct {
	Code  string `form:"code" binding:"required"`
	State string `form:"state" binding:"required"`
}

// handleGoogleOAuthLogin initiates Google OAuth login
func (s *Server) handleGoogleOAuthLogin(c *gin.Context) {
	var req GoogleOAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid Google OAuth request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Validate organization ID if provided
	if req.OrganizationID != "" {
		if _, err := uuid.Parse(req.OrganizationID); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid organization ID"))
			return
		}
	}

	// Call gRPC service
	grpcReq := &pb.GoogleOAuthURLRequest{
		OrganizationId: req.OrganizationID,
	}

	resp, err := s.grpcClient.GetGoogleOAuthURL(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to generate Google OAuth URL", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to generate OAuth URL"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"auth_url": resp.Url,
		"state":    resp.State,
	}))
}

// handleGoogleOAuthCallback handles the Google OAuth callback
func (s *Server) handleGoogleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		// Redirect to frontend error page
		c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3001/auth/error?error=missing_params")
		return
	}

	// Call gRPC service
	grpcReq := &pb.GoogleOAuthCallbackRequest{
		Code:  code,
		State: state,
	}

	resp, err := s.grpcClient.HandleGoogleOAuthCallback(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to process Google OAuth callback", "error", err)
		// Redirect to frontend error page
		c.Redirect(http.StatusTemporaryRedirect, "http://localhost:3001/auth/error?error=callback_failed")
		return
	}

	// Redirect to frontend callback page with session data
	redirectURL := fmt.Sprintf("http://localhost:3001/auth/callback?session_token=%s&expires_at=%d&user_id=%s&email=%s&organization_id=%s",
		resp.SessionToken,
		resp.ExpiresAt.AsTime().Unix(),
		resp.User.Id,
		resp.User.Email,
		resp.User.OrganizationId)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// handleGoogleOAuthRefresh refreshes Google OAuth token
func (s *Server) handleGoogleOAuthRefresh(c *gin.Context) {
	authMethodID := c.Param("id")
	if authMethodID == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "Auth method ID is required"))
		return
	}

	// Validate auth method ID
	if _, err := uuid.Parse(authMethodID); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid auth method ID"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.RefreshGoogleOAuthTokenRequest{
		RefreshToken: authMethodID,
	}

	resp, err := s.grpcClient.RefreshGoogleOAuthToken(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to refresh Google OAuth token", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to refresh token"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"message":    "Token refreshed successfully",
		"token_type": resp.TokenType,
	}))
}
