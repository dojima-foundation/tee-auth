package rest

import (
	"net/http"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SeedGenerationRequest represents the request payload for seed generation
type SeedGenerationRequest struct {
	OrganizationID string  `json:"organization_id" binding:"required"`
	UserID         string  `json:"user_id" binding:"required"`
	Strength       int32   `json:"strength" binding:"required"`
	Passphrase     *string `json:"passphrase,omitempty"`
}

// SeedValidationRequest represents the request payload for seed validation
type SeedValidationRequest struct {
	SeedPhrase      string  `json:"seed_phrase" binding:"required"`
	EncryptedEntropy *string `json:"encrypted_entropy,omitempty"`
}

// handleGetEnclaveInfo returns information about the enclave
func (s *Server) handleGetEnclaveInfo(c *gin.Context) {
	// Call gRPC service
	resp, err := s.grpcClient.GetEnclaveInfo(c.Request.Context(), &emptypb.Empty{})
	if err != nil {
		s.logger.Error("Failed to get enclave info via gRPC", "error", err)
		c.JSON(http.StatusServiceUnavailable, errorResponse(err, "Failed to get enclave information"))
		return
	}

	result := map[string]interface{}{
		"version":      resp.Version,
		"enclave_id":   resp.EnclaveId,
		"capabilities": resp.Capabilities,
		"healthy":      resp.Healthy,
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// handleRequestSeedGeneration requests seed phrase generation from the enclave
func (s *Server) handleRequestSeedGeneration(c *gin.Context) {
	var req SeedGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid seed generation request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Validate strength
	if req.Strength != 128 && req.Strength != 256 {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "Strength must be 128 or 256"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.SeedGenerationRequest{
		OrganizationId: req.OrganizationID,
		UserId:         req.UserID,
		Strength:       req.Strength,
	}

	if req.Passphrase != nil {
		grpcReq.Passphrase = req.Passphrase
	}

	resp, err := s.grpcClient.RequestSeedGeneration(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to request seed generation via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to generate seed"))
		return
	}

	result := map[string]interface{}{
		"seed_phrase": resp.SeedPhrase,
		"entropy":     resp.Entropy,
		"strength":    resp.Strength,
		"word_count":  resp.WordCount,
		"request_id":  resp.RequestId,
	}

	c.JSON(http.StatusOK, successResponse(result))
}

// handleValidateSeed validates a seed phrase
func (s *Server) handleValidateSeed(c *gin.Context) {
	var req SeedValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid seed validation request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.SeedValidationRequest{
		SeedPhrase: req.SeedPhrase,
	}

	resp, err := s.grpcClient.ValidateSeed(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to validate seed via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to validate seed"))
		return
	}

	result := map[string]interface{}{
		"is_valid":   resp.IsValid,
		"strength":   resp.Strength,
		"word_count": resp.WordCount,
		"errors":     resp.Errors,
	}

	httpStatus := http.StatusOK
	if !resp.IsValid {
		httpStatus = http.StatusBadRequest
	}

	c.JSON(httpStatus, successResponse(result))
}
