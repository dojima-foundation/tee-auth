package rest

import (
	"net/http"
	"strconv"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/gin-gonic/gin"
)

// CreateWalletRequest represents the request payload for creating a wallet
type CreateWalletRequest struct {
	Name           string                   `json:"name" binding:"required"`
	Accounts       []CreateWalletAccountReq `json:"accounts" binding:"required,min=1"`
	MnemonicLength *int32                   `json:"mnemonic_length,omitempty"` // 12, 15, 18, 21, 24
	Tags           []string                 `json:"tags,omitempty"`
}

// CreateWalletAccountReq represents a wallet account in the request
type CreateWalletAccountReq struct {
	Curve         string `json:"curve" binding:"required"`          // CURVE_SECP256K1, CURVE_ED25519
	PathFormat    string `json:"path_format" binding:"required"`    // PATH_FORMAT_BIP32
	Path          string `json:"path" binding:"required"`           // e.g., m/44'/60'/0'/0/0
	AddressFormat string `json:"address_format" binding:"required"` // ADDRESS_FORMAT_ETHEREUM, etc.
}

// DeleteWalletRequest represents the request payload for deleting a wallet
type DeleteWalletRequest struct {
	DeleteWithoutExport *bool `json:"delete_without_export,omitempty"`
}

// handleCreateWallet creates a new wallet
func (s *Server) handleCreateWallet(c *gin.Context) {
	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid create wallet request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Get organization ID from session context (set by session middleware)
	organizationID, exists := GetOrganizationIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse(nil, "Organization ID not found in session"))
		return
	}

	// Convert accounts
	accounts := make([]*pb.CreateWalletAccount, len(req.Accounts))
	for i, acc := range req.Accounts {
		accounts[i] = &pb.CreateWalletAccount{
			Curve:         acc.Curve,
			PathFormat:    acc.PathFormat,
			Path:          acc.Path,
			AddressFormat: acc.AddressFormat,
		}
	}

	// Call gRPC service
	grpcReq := &pb.CreateWalletRequest{
		OrganizationId: organizationID,
		Name:           req.Name,
		Accounts:       accounts,
		Tags:           req.Tags,
	}

	if req.MnemonicLength != nil {
		grpcReq.MnemonicLength = req.MnemonicLength
	}

	resp, err := s.grpcClient.CreateWallet(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to create wallet via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to create wallet"))
		return
	}

	c.JSON(http.StatusCreated, successResponse(gin.H{
		"wallet":    convertProtoWalletToREST(resp.Wallet),
		"addresses": resp.Addresses,
	}))
}

// handleGetWallet retrieves a wallet by ID
func (s *Server) handleGetWallet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "Wallet ID is required"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.GetWalletRequest{
		Id: id,
	}

	resp, err := s.grpcClient.GetWallet(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to get wallet via gRPC", "error", err, "id", id)
		c.JSON(http.StatusNotFound, errorResponse(err, "Wallet not found"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"wallet": convertProtoWalletToREST(resp.Wallet),
	}))
}

// handleListWallets lists wallets with pagination and filtering
func (s *Server) handleListWallets(c *gin.Context) {
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
	grpcReq := &pb.ListWalletsRequest{
		OrganizationId: organizationID,
		PageSize:       pageSize,
		PageToken:      pageToken,
	}

	resp, err := s.grpcClient.ListWallets(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to list wallets via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to list wallets"))
		return
	}

	// Convert wallets
	wallets := make([]interface{}, len(resp.Wallets))
	for i, wallet := range resp.Wallets {
		wallets[i] = convertProtoWalletToREST(wallet)
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"wallets":         wallets,
		"next_page_token": resp.NextPageToken,
	}))
}

// handleDeleteWallet deletes a wallet
func (s *Server) handleDeleteWallet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Wallet ID is required",
		})
		return
	}

	var req DeleteWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, use defaults
		req = DeleteWalletRequest{}
	}

	// Call gRPC service
	grpcReq := &pb.DeleteWalletRequest{
		Id: id,
	}

	if req.DeleteWithoutExport != nil {
		grpcReq.DeleteWithoutExport = req.DeleteWithoutExport
	}

	resp, err := s.grpcClient.DeleteWallet(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to delete wallet via gRPC", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to delete wallet"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"success": resp.Success,
		"message": resp.Message,
	}))
}

// convertProtoWalletToREST converts a protobuf wallet to REST format
func convertProtoWalletToREST(wallet *pb.Wallet) map[string]interface{} {
	result := map[string]interface{}{
		"id":              wallet.Id,
		"organization_id": wallet.OrganizationId,
		"name":            wallet.Name,
		"public_key":      wallet.PublicKey,
		"tags":            wallet.Tags,
		"is_active":       wallet.IsActive,
	}

	if wallet.CreatedAt != nil {
		result["created_at"] = wallet.CreatedAt.AsTime()
	}

	if wallet.UpdatedAt != nil {
		result["updated_at"] = wallet.UpdatedAt.AsTime()
	}

	// Convert wallet accounts
	if len(wallet.Accounts) > 0 {
		accounts := make([]interface{}, len(wallet.Accounts))
		for i, account := range wallet.Accounts {
			accounts[i] = map[string]interface{}{
				"id":             account.Id,
				"wallet_id":      account.WalletId,
				"name":           account.Name,
				"path":           account.Path,
				"public_key":     account.PublicKey,
				"address":        account.Address,
				"curve":          account.Curve,
				"address_format": account.AddressFormat,
				"is_active":      account.IsActive,
			}
			if account.CreatedAt != nil {
				accounts[i].(map[string]interface{})["created_at"] = account.CreatedAt.AsTime()
			}
			if account.UpdatedAt != nil {
				accounts[i].(map[string]interface{})["updated_at"] = account.UpdatedAt.AsTime()
			}
		}
		result["accounts"] = accounts
	}

	return result
}
