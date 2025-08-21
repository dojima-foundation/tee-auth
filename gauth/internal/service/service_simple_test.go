package service

import (
	"context"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGAuthService_Status(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
	}

	logger := logger.NewDefault()
	renclave := NewRenclaveClient("http://mock-server", 1*time.Second)

	service := &GAuthService{
		config:   cfg,
		logger:   logger,
		renclave: renclave,
	}

	ctx := context.Background()
	response, err := service.Status(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "1.0.0", response.Version)
	assert.NotEmpty(t, response.BuildTime)
	assert.NotEmpty(t, response.GitCommit)
	assert.NotZero(t, response.Uptime)
	assert.NotEmpty(t, response.Metrics)
}

func TestGAuthService_RequestSeedGeneration_Error(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
	}

	logger := logger.NewDefault()
	renclave := NewRenclaveClient("http://invalid-server", 1*time.Second)

	service := &GAuthService{
		config:   cfg,
		logger:   logger,
		renclave: renclave,
	}

	ctx := context.Background()
	response, err := service.RequestSeedGeneration(ctx, uuid.New().String(), uuid.New().String(), 256, nil)

	// Should error since the server doesn't exist
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGAuthService_ValidateSeed_Error(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
	}

	logger := logger.NewDefault()
	renclave := NewRenclaveClient("http://invalid-server", 1*time.Second)

	service := &GAuthService{
		config:   cfg,
		logger:   logger,
		renclave: renclave,
	}

	ctx := context.Background()
	response, err := service.ValidateSeed(ctx, "test seed phrase")

	// Should error since the server doesn't exist
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGAuthService_GetEnclaveInfo_Error(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
	}

	logger := logger.NewDefault()
	renclave := NewRenclaveClient("http://invalid-server", 1*time.Second)

	service := &GAuthService{
		config:   cfg,
		logger:   logger,
		renclave: renclave,
	}

	ctx := context.Background()
	response, err := service.GetEnclaveInfo(ctx)

	// Should error since the server doesn't exist
	assert.Error(t, err)
	assert.Nil(t, response)
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}
