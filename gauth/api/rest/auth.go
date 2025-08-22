package rest

import (
	"encoding/json"
	"net/http"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/gin-gonic/gin"
)

// AuthenticateRequest represents the request payload for authentication
type AuthenticateRequest struct {
	OrganizationID string `json:"organization_id" binding:"required"`
	UserID         string `json:"user_id" binding:"required"`
	AuthMethodID   string `json:"auth_method_id" binding:"required"`
	Signature      string `json:"signature" binding:"required"`
	Timestamp      string `json:"timestamp" binding:"required"`
}

// AuthorizeRequest represents the request payload for authorization
type AuthorizeRequest struct {
	SessionToken string      `json:"session_token" binding:"required"`
	ActivityType string      `json:"activity_type" binding:"required"`
	Parameters   interface{} `json:"parameters,omitempty"`
}

// handleAuthenticate handles user authentication
func (s *Server) handleAuthenticate(c *gin.Context) {
	var req AuthenticateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid authenticate request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.AuthenticateRequest{
		OrganizationId: req.OrganizationID,
		UserId:         req.UserID,
		AuthMethodId:   req.AuthMethodID,
		Signature:      req.Signature,
		Timestamp:      req.Timestamp,
	}

	resp, err := s.grpcClient.Authenticate(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to authenticate via gRPC", "error", err)
		c.JSON(http.StatusUnauthorized, errorResponse(err, "Authentication failed"))
		return
	}

	result := map[string]interface{}{
		"authenticated": resp.Authenticated,
		"session_token": resp.SessionToken,
		"expires_at":    resp.ExpiresAt.AsTime(),
	}

	if resp.User != nil {
		result["user"] = convertProtoUserToREST(resp.User)
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// handleAuthorize handles authorization checks
func (s *Server) handleAuthorize(c *gin.Context) {
	var req AuthorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid authorize request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Convert parameters to JSON string if provided
	var parametersJSON string
	if req.Parameters != nil {
		if paramBytes, err := json.Marshal(req.Parameters); err == nil {
			parametersJSON = string(paramBytes)
		}
	}

	// Call gRPC service
	grpcReq := &pb.AuthorizeRequest{
		SessionToken: req.SessionToken,
		ActivityType: req.ActivityType,
		Parameters:   parametersJSON,
	}

	resp, err := s.grpcClient.Authorize(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to authorize via gRPC", "error", err)
		c.JSON(http.StatusForbidden, errorResponse(err, "Authorization failed"))
		return
	}

	result := map[string]interface{}{
		"authorized":         resp.Authorized,
		"reason":             resp.Reason,
		"required_approvals": resp.RequiredApprovals,
	}

	httpStatus := http.StatusOK
	if !resp.Authorized {
		httpStatus = http.StatusForbidden
	}

	c.JSON(httpStatus, successResponse(result))
}
