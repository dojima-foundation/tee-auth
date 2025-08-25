package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
)

// HealthService provides health check and status functionality
type HealthService struct {
	*GAuthService
}

// Health checks the health of all services
func (s *HealthService) Health(ctx context.Context) (*ServiceHealthResponse, error) {
	s.logger.Debug("Performing health check")

	services := []ServiceStatus{}
	overallStatus := "healthy"

	// Check database
	dbStatus := "healthy"
	if err := s.db.GetDB().WithContext(ctx).Raw("SELECT 1").Error; err != nil {
		dbStatus = "unhealthy"
		overallStatus = "unhealthy"
		s.logger.Error("Database health check failed", "error", err)
	}
	services = append(services, ServiceStatus{
		Name:   "database",
		Status: dbStatus,
	})

	// Check Redis
	redisStatus := "healthy"
	if err := s.redis.Health(ctx); err != nil {
		redisStatus = "unhealthy"
		overallStatus = "unhealthy"
		s.logger.Error("Redis health check failed", "error", err)
	}
	services = append(services, ServiceStatus{
		Name:   "redis",
		Status: redisStatus,
	})

	// Check enclave
	enclaveStatus := "healthy"
	if err := s.renclave.Health(ctx); err != nil {
		enclaveStatus = "unhealthy"
		overallStatus = "unhealthy"
		s.logger.Error("Enclave health check failed", "error", err)
	}
	services = append(services, ServiceStatus{
		Name:   "enclave",
		Status: enclaveStatus,
	})

	response := &ServiceHealthResponse{
		Status:    overallStatus,
		Services:  services,
		Timestamp: time.Now(),
	}

	s.logger.Debug("Health check completed", "status", overallStatus)
	return response, nil
}

// Status provides detailed status information
func (s *HealthService) Status(ctx context.Context) (*StatusResponse, error) {
	s.logger.Debug("Getting service status")

	// Get basic counts
	var orgCount, userCount, walletCount, privateKeyCount int64

	// Check if database is available before trying to use it
	if s.db != nil && s.db.GetDB() != nil {
		if err := s.db.GetDB().WithContext(ctx).Model(&models.Organization{}).Count(&orgCount).Error; err != nil {
			s.logger.Error("Failed to count organizations", "error", err)
		}

		if err := s.db.GetDB().WithContext(ctx).Model(&models.User{}).Count(&userCount).Error; err != nil {
			s.logger.Error("Failed to count users", "error", err)
		}

		if err := s.db.GetDB().WithContext(ctx).Model(&models.Wallet{}).Count(&walletCount).Error; err != nil {
			s.logger.Error("Failed to count wallets", "error", err)
		}

		if err := s.db.GetDB().WithContext(ctx).Model(&models.PrivateKey{}).Count(&privateKeyCount).Error; err != nil {
			s.logger.Error("Failed to count private keys", "error", err)
		}
	}

	// Get enclave info (for health check only)
	if s.renclave != nil {
		if _, err := s.renclave.GetInfo(ctx); err != nil {
			s.logger.Error("Failed to get enclave info", "error", err)
		}
	}

	response := &StatusResponse{
		Version:   "1.0.0",
		BuildTime: time.Now().Format(time.RFC3339),
		GitCommit: "development",
		Uptime:    time.Now(),
		Metrics: map[string]string{
			"organizations": fmt.Sprintf("%d", orgCount),
			"users":         fmt.Sprintf("%d", userCount),
			"wallets":       fmt.Sprintf("%d", walletCount),
			"private_keys":  fmt.Sprintf("%d", privateKeyCount),
		},
	}

	s.logger.Debug("Status check completed")
	return response, nil
}
