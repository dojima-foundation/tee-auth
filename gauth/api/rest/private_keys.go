package rest

import (
	"net/http"
	"strconv"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/gin-gonic/gin"
)

// CreatePrivateKeyRequest represents the request payload for creating a private key
type CreatePrivateKeyRequest struct {
	WalletID           string   `json:"wallet_id" binding:"required"` // Link to wallet
	Name               string   `json:"name" binding:"required"`
	Curve              string   `json:"curve" binding:"required"`       // CURVE_SECP256K1, CURVE_ED25519
	PrivateKeyMaterial *string  `json:"private_key_material,omitempty"` // Optional: provide key material, otherwise generate
	Tags               []string `json:"tags,omitempty"`
}

// DeletePrivateKeyRequest represents the request payload for deleting a private key
type DeletePrivateKeyRequest struct {
	DeleteWithoutExport *bool `json:"delete_without_export,omitempty"`
}

// handleCreatePrivateKey creates a new private key
func (s *Server) handleCreatePrivateKey(c *gin.Context) {
	var req CreatePrivateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid create private key request", "error", err)
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
	grpcReq := &pb.CreatePrivateKeyRequest{
		OrganizationId: organizationID,
		WalletId:       req.WalletID,
		Name:           req.Name,
		Curve:          req.Curve,
		Tags:           req.Tags,
	}

	if req.PrivateKeyMaterial != nil {
		grpcReq.PrivateKeyMaterial = req.PrivateKeyMaterial
	}

	resp, err := s.grpcClient.CreatePrivateKey(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to create private key via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to create private key"))
		return
	}

	c.JSON(http.StatusCreated, successResponse(gin.H{
		"private_key": convertProtoPrivateKeyToREST(resp.PrivateKey),
	}))
}

// handleGetPrivateKey retrieves a private key by ID
func (s *Server) handleGetPrivateKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "Private key ID is required"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.GetPrivateKeyRequest{
		Id: id,
	}

	resp, err := s.grpcClient.GetPrivateKey(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to get private key via gRPC", "error", err, "id", id)
		c.JSON(http.StatusNotFound, errorResponse(err, "Private key not found"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"private_key": convertProtoPrivateKeyToREST(resp.PrivateKey),
	}))
}

// handleListPrivateKeys lists private keys with pagination and filtering
func (s *Server) handleListPrivateKeys(c *gin.Context) {
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
	grpcReq := &pb.ListPrivateKeysRequest{
		OrganizationId: organizationID,
		PageSize:       pageSize,
		PageToken:      pageToken,
	}

	resp, err := s.grpcClient.ListPrivateKeys(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to list private keys via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to list private keys"))
		return
	}

	// Convert private keys
	privateKeys := make([]interface{}, len(resp.PrivateKeys))
	for i, pk := range resp.PrivateKeys {
		privateKeys[i] = convertProtoPrivateKeyToREST(pk)
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"private_keys":    privateKeys,
		"next_page_token": resp.NextPageToken,
	}))
}

// handleDeletePrivateKey deletes a private key
func (s *Server) handleDeletePrivateKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "Private key ID is required"))
		return
	}

	var req DeletePrivateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, use defaults
		req = DeletePrivateKeyRequest{}
	}

	// Call gRPC service
	grpcReq := &pb.DeletePrivateKeyRequest{
		Id: id,
	}

	if req.DeleteWithoutExport != nil {
		grpcReq.DeleteWithoutExport = req.DeleteWithoutExport
	}

	resp, err := s.grpcClient.DeletePrivateKey(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to delete private key via gRPC", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to delete private key"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"success": resp.Success,
		"message": resp.Message,
	}))
}

// convertProtoPrivateKeyToREST converts a protobuf private key to REST format
func convertProtoPrivateKeyToREST(pk *pb.PrivateKey) map[string]interface{} {
	result := map[string]interface{}{
		"id":              pk.Id,
		"organization_id": pk.OrganizationId,
		"name":            pk.Name,
		"public_key":      pk.PublicKey,
		"curve":           pk.Curve,
		"tags":            pk.Tags,
		"is_active":       pk.IsActive,
	}

	if pk.CreatedAt != nil {
		result["created_at"] = pk.CreatedAt.AsTime()
	}

	if pk.UpdatedAt != nil {
		result["updated_at"] = pk.UpdatedAt.AsTime()
	}

	return result
}
