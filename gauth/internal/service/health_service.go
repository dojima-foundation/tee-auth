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
		Name:    "database",
		Status:  dbStatus,
		Details: fmt.Sprintf("PostgreSQL connection"),
	})

	// Check Redis
	redisStatus := "healthy"
	if err := s.redis.Ping(ctx); err != nil {
		redisStatus = "unhealthy"
		overallStatus = "unhealthy"
		s.logger.Error("Redis health check failed", "error", err)
	}
	services = append(services, ServiceStatus{
		Name:    "redis",
		Status:  redisStatus,
		Details: fmt.Sprintf("Redis cache connection"),
	})

	// Check enclave
	enclaveStatus := "healthy"
	if err := s.renclave.Health(ctx); err != nil {
		enclaveStatus = "unhealthy"
		overallStatus = "unhealthy"
		s.logger.Error("Enclave health check failed", "error", err)
	}
	services = append(services, ServiceStatus{
		Name:    "enclave",
		Status:  enclaveStatus,
		Details: fmt.Sprintf("Renclave-v2 service"),
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

	// Get enclave info
	var enclaveInfo *EnclaveInfoResponse
	if info, err := s.renclave.GetInfo(ctx); err == nil {
		enclaveInfo = &EnclaveInfoResponse{
			Version:      info.Version,
			EnclaveID:    info.EnclaveID,
			Capabilities: info.Capabilities,
			Healthy:      info.Healthy,
		}
	} else {
		s.logger.Error("Failed to get enclave info", "error", err)
	}

	response := &StatusResponse{
		Service:        "gauth",
		Version:        "1.0.0",
		Status:         "running",
		Uptime:         time.Since(s.startTime),
		OrganizationCount: orgCount,
		UserCount:      userCount,
		WalletCount:    walletCount,
		PrivateKeyCount: privateKeyCount,
		EnclaveInfo:    enclaveInfo,
		Timestamp:      time.Now(),
	}

	s.logger.Debug("Status check completed")
	return response, nil
}
