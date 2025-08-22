package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/gin-gonic/gin"
)

// CreateActivityRequest represents the request payload for creating an activity
type CreateActivityRequest struct {
	OrganizationID string      `json:"organization_id" binding:"required"`
	Type           string      `json:"type" binding:"required"`
	Parameters     interface{} `json:"parameters,omitempty"`
	CreatedBy      string      `json:"created_by" binding:"required"`
}

// handleCreateActivity creates a new activity
func (s *Server) handleCreateActivity(c *gin.Context) {
	var req CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid create activity request", "error", err)
		c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request payload"))
		return
	}

	// Convert parameters to JSON string
	var parametersJSON string
	if req.Parameters != nil {
		paramBytes, err := json.Marshal(req.Parameters)
		if err != nil {
			s.logger.Error("Failed to marshal activity parameters", "error", err)
			c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid parameters format"))
			return
		}
		parametersJSON = string(paramBytes)
	}

	// Call gRPC service
	grpcReq := &pb.CreateActivityRequest{
		OrganizationId: req.OrganizationID,
		Type:           req.Type,
		Parameters:     parametersJSON,
		CreatedBy:      req.CreatedBy,
	}

	resp, err := s.grpcClient.CreateActivity(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to create activity via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to create activity"))
		return
	}

	c.JSON(http.StatusCreated, successResponse(gin.H{
		"activity": convertProtoActivityToREST(resp.Activity),
	}))
}

// handleGetActivity retrieves an activity by ID
func (s *Server) handleGetActivity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "Activity ID is required"))
		return
	}

	// Call gRPC service
	grpcReq := &pb.GetActivityRequest{
		Id: id,
	}

	resp, err := s.grpcClient.GetActivity(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to get activity via gRPC", "error", err, "id", id)
		c.JSON(http.StatusNotFound, errorResponse(err, "Activity not found"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"activity": convertProtoActivityToREST(resp.Activity),
	}))
}

// handleListActivities lists activities with pagination and filtering
func (s *Server) handleListActivities(c *gin.Context) {
	// Parse query parameters
	organizationID := c.Query("organization_id")
	if organizationID == "" {
		c.JSON(http.StatusBadRequest, errorResponse(nil, "organization_id query parameter is required"))
		return
	}

	pageSize := int32(10) // default
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.ParseInt(ps, 10, 32); err == nil {
			pageSize = int32(parsed)
		}
	}

	pageToken := c.Query("page_token")
	activityType := c.Query("type")
	status := c.Query("status")

	// Call gRPC service
	grpcReq := &pb.ListActivitiesRequest{
		OrganizationId: organizationID,
		PageSize:       pageSize,
		PageToken:      pageToken,
	}

	if activityType != "" {
		grpcReq.Type = &activityType
	}
	if status != "" {
		grpcReq.Status = &status
	}

	resp, err := s.grpcClient.ListActivities(c.Request.Context(), grpcReq)
	if err != nil {
		s.logger.Error("Failed to list activities via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Failed to list activities"))
		return
	}

	// Convert activities
	activities := make([]interface{}, len(resp.Activities))
	for i, activity := range resp.Activities {
		activities[i] = convertProtoActivityToREST(activity)
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"activities":      activities,
		"next_page_token": resp.NextPageToken,
	}))
}

// convertProtoActivityToREST converts a protobuf activity to REST format
func convertProtoActivityToREST(activity *pb.Activity) map[string]interface{} {
	result := map[string]interface{}{
		"id":              activity.Id,
		"organization_id": activity.OrganizationId,
		"type":            activity.Type,
		"status":          activity.Status,
		"created_by":      activity.CreatedBy,
	}

	// Parse JSON parameters
	if activity.Parameters != "" {
		var params interface{}
		if err := json.Unmarshal([]byte(activity.Parameters), &params); err == nil {
			result["parameters"] = params
		} else {
			result["parameters"] = activity.Parameters // fallback to raw string
		}
	}

	// Parse JSON result
	if activity.Result != nil && *activity.Result != "" {
		var resultData interface{}
		if err := json.Unmarshal([]byte(*activity.Result), &resultData); err == nil {
			result["result"] = resultData
		} else {
			result["result"] = *activity.Result // fallback to raw string
		}
	}

	// Add intent information
	if activity.Intent != nil {
		result["intent"] = map[string]interface{}{
			"fingerprint": activity.Intent.Fingerprint,
			"summary":     activity.Intent.Summary,
		}
	}

	if activity.CreatedAt != nil {
		result["created_at"] = activity.CreatedAt.AsTime()
	}

	if activity.UpdatedAt != nil {
		result["updated_at"] = activity.UpdatedAt.AsTime()
	}

	return result
}
