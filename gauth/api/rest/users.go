package rest

import (
	"net/http"
	"strconv"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/gin-gonic/gin"
)

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Username  string   `json:"username" binding:"required"`
	Email     string   `json:"email" binding:"required,email"`
	PublicKey *string  `json:"public_key,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Username *string  `json:"username,omitempty"`
	Email    *string  `json:"email,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// handleCreateUser creates a new user
func (s *Server) handleCreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid create user request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Get organization ID from session context (set by session middleware)
	organizationID, exists := GetOrganizationIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse(nil, "Organization ID not found in session"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.CreateUserRequest{
		OrganizationId: organizationID,
		Username:       req.Username,
		Email:          req.Email,
		Tags:           req.Tags,
	}

	if req.PublicKey != nil {
		grpcReq.PublicKey = *req.PublicKey
	}

	resp, err := s.grpcClient.CreateUser(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to create user via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to create user"))
		return
	}

	c.JSON(http.StatusCreated, successResponse(gin.H{
		"user": convertProtoUserToREST(resp.User),
	}))
}

// handleGetUser retrieves a user by ID
func (s *Server) handleGetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "User ID is required"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.GetUserRequest{
		Id: id,
	}

	resp, err := s.grpcClient.GetUser(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to get user via gRPC", "error", err, "id", id)
		c.JSON(http.StatusNotFound, errorResponse(err, "User not found"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"user": convertProtoUserToREST(resp.User),
	}))
}

// handleUpdateUser updates a user
func (s *Server) handleUpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "User ID is required"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid update user request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.UpdateUserRequest{
		Id:       id,
		Username: req.Username,
		Email:    req.Email,
		Tags:     req.Tags,
		IsActive: req.IsActive,
	}

	resp, err := s.grpcClient.UpdateUser(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to update user via gRPC", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to update user"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"user": convertProtoUserToREST(resp.User),
	}))
}

// handleListUsers lists users with pagination and filtering
func (s *Server) handleListUsers(c *gin.Context) {
	// Get organization ID from session context (set by session middleware)
	organizationID, exists := GetOrganizationIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse(nil, "Organization ID not found in session"))
		return
	}

	pageSize := int32(10) // default
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.ParseInt(ps, 10, 32); err == nil {
			pageSize = int32(parsed)
		}
	}

	pageToken := c.Query("page_token")

	// Call gRPC service
	grpcReq := &pb.ListUsersRequest{
		OrganizationId: organizationID,
		PageSize:       pageSize,
		PageToken:      pageToken,
	}

	resp, err := s.grpcClient.ListUsers(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to list users via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to list users"))
		return
	}

	// Convert users
	users := make([]interface{}, len(resp.Users))
	for i, user := range resp.Users {
		users[i] = convertProtoUserToREST(user)
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"users":           users,
		"next_page_token": resp.NextPageToken,
	}))
}

// convertProtoUserToREST converts a protobuf user to REST format
func convertProtoUserToREST(user *pb.User) map[string]interface{} {
	result := map[string]interface{}{
		"id":              user.Id,
		"organization_id": user.OrganizationId,
		"username":        user.Username,
		"email":           user.Email,
		"public_key":      user.PublicKey,
		"tags":            user.Tags,
		"is_active":       user.IsActive,
	}

	if user.CreatedAt != nil {
		result["created_at"] = user.CreatedAt.AsTime()
	}

	if user.UpdatedAt != nil {
		result["updated_at"] = user.UpdatedAt.AsTime()
	}

	// Convert auth methods
	if len(user.AuthMethods) > 0 {
		authMethods := make([]interface{}, len(user.AuthMethods))
		for i, method := range user.AuthMethods {
			authMethods[i] = map[string]interface{}{
				"id":        method.Id,
				"type":      method.Type,
				"is_active": method.IsActive,
			}
			if method.CreatedAt != nil {
				authMethods[i].(map[string]interface{})["created_at"] = method.CreatedAt.AsTime()
			}
		}
		result["auth_methods"] = authMethods
	}

	return result
}
