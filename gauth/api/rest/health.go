package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
)

// handleHealth returns the health status of the service
func (s *Server) handleHealth(c *gin.Context) {
	// Call gRPC service
	resp, err := s.grpcClient.Health(c.Request.Context(), &emptypb.Empty{})
	if err != nil {
		s.logger.Error("Failed to get health status via gRPC", "error", err)
		c.JSON(http.StatusServiceUnavailable, errorResponse(err, "Health check failed"))
		return
	}

	// Convert services
	services := make([]interface{}, len(resp.Services))
	for i, service := range resp.Services {
		services[i] = map[string]interface{}{
			"name":   service.Name,
			"status": service.Status,
			"error":  service.Error,
		}
	}

	result := map[string]interface{}{
		"status":    resp.Status,
		"services":  services,
		"timestamp": resp.Timestamp.AsTime(),
	}

	// Set appropriate HTTP status based on health status
	httpStatus := http.StatusOK
	if resp.Status != "healthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, successResponse(result))
}

// handleStatus returns detailed status information about the service
func (s *Server) handleStatus(c *gin.Context) {
	// Call gRPC service
	resp, err := s.grpcClient.Status(c.Request.Context(), &emptypb.Empty{})
	if err != nil {
		s.logger.Error("Failed to get status via gRPC", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse(err, "Status check failed"))
		return
	}

	result := map[string]interface{}{
		"version":    resp.Version,
		"build_time": resp.BuildTime,
		"git_commit": resp.GitCommit,
		"uptime":     resp.Uptime.AsTime(),
		"metrics":    resp.Metrics,
	}

	c.JSON(http.StatusOK, successResponse(result))
}
