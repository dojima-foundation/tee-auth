package rest

import (
	"net/http"
	"strconv"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/gin-gonic/gin"
)

// CreateOrganizationRequest represents the request payload for creating an organization
type CreateOrganizationRequest struct {
	Name                 string  `json:"name" binding:"required"`
	InitialUserEmail     string  `json:"initial_user_email" binding:"required,email"`
	InitialUserPublicKey *string `json:"initial_user_public_key,omitempty"`
}

// UpdateOrganizationRequest represents the request payload for updating an organization
type UpdateOrganizationRequest struct {
	Name *string `json:"name,omitempty"`
}

// handleCreateOrganization creates a new organization
func (s *Server) handleCreateOrganization(c *gin.Context) {
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid create organization request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.CreateOrganizationRequest{
		Name:             req.Name,
		InitialUserEmail: req.InitialUserEmail,
	}

	if req.InitialUserPublicKey != nil {
		grpcReq.InitialUserPublicKey = *req.InitialUserPublicKey
	}

	resp, err := s.grpcClient.CreateOrganization(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to create organization via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to create organization"))
		return
	}

	c.JSON(http.StatusCreated, successResponse(gin.H{
		"organization": ConvertProtoOrganizationToREST(resp.Organization),
		"status":       resp.Status,
		"user_id": func() string {
			if len(resp.Organization.Users) > 0 {
				return resp.Organization.Users[0].Id
			}
			return ""
		}(),
	}))
}

// handleGetOrganization retrieves an organization by ID
func (s *Server) handleGetOrganization(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "Organization ID is required"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.GetOrganizationRequest{
		Id: id,
	}

	resp, err := s.grpcClient.GetOrganization(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to get organization via gRPC", "error", err, "id", id)
		c.JSON(http.StatusNotFound, errorResponse(err, "Organization not found"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"organization": ConvertProtoOrganizationToREST(resp.Organization),
	}))
}

// handleUpdateOrganization updates an organization
func (s *Server) handleUpdateOrganization(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "Organization ID is required"))
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid update organization request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.UpdateOrganizationRequest{
		Id:   id,
		Name: req.Name,
	}

	resp, err := s.grpcClient.UpdateOrganization(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to update organization via gRPC", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to update organization"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"organization": ConvertProtoOrganizationToREST(resp.Organization),
	}))
}

// handleListOrganizations lists organizations with pagination
func (s *Server) handleListOrganizations(c *gin.Context) {
	// Parse query parameters
	pageSize := int32(10) // default
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.ParseInt(ps, 10, 32); err == nil {
			pageSize = int32(parsed)
		}
	}

	pageToken := c.Query("page_token")

	// Call gRPC service
	grpcReq := &pb.ListOrganizationsRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
	}

	resp, err := s.grpcClient.ListOrganizations(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to list organizations via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to list organizations"))
		return
	}

	// Convert organizations
	organizations := make([]interface{}, len(resp.Organizations))
	for i, org := range resp.Organizations {
		organizations[i] = ConvertProtoOrganizationToREST(org)
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"organizations":   organizations,
		"next_page_token": resp.NextPageToken,
	}))
}

// ConvertProtoOrganizationToREST converts a protobuf organization to REST format
func ConvertProtoOrganizationToREST(org *pb.Organization) map[string]interface{} {
	result := map[string]interface{}{
		"id":      org.Id,
		"version": org.Version,
		"name":    org.Name,
	}

	if org.CreatedAt != nil {
		result["created_at"] = org.CreatedAt.AsTime()
	}

	if org.UpdatedAt != nil {
		result["updated_at"] = org.UpdatedAt.AsTime()
	}

	if org.RootQuorum != nil {
		result["root_quorum"] = map[string]interface{}{
			"threshold": org.RootQuorum.Threshold,
		}
	}

	// Include users if present
	if len(org.Users) > 0 {
		users := make([]map[string]interface{}, len(org.Users))
		for i, user := range org.Users {
			users[i] = map[string]interface{}{
				"id":              user.Id,
				"organization_id": user.OrganizationId,
				"username":        user.Username,
				"email":           user.Email,
				"public_key":      user.PublicKey,
				"tags":            user.Tags,
				"is_active":       user.IsActive,
			}
			if user.CreatedAt != nil {
				users[i]["created_at"] = user.CreatedAt.AsTime()
			}
			if user.UpdatedAt != nil {
				users[i]["updated_at"] = user.UpdatedAt.AsTime()
			}
		}
		result["users"] = users
	}

	return result
}
