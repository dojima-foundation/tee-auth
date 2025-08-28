package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	restServer "github.com/dojima-foundation/tee-auth/gauth/api/rest"
	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	grpcServer "github.com/dojima-foundation/tee-auth/gauth/internal/grpc"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RESTIntegrationTestSuite struct {
	suite.Suite
	grpcServer *grpcServer.Server
	restServer *restServer.Server
	config     *config.Config
	baseURL    string
	httpClient *http.Client
}

func (suite *RESTIntegrationTestSuite) SetupSuite() {
	// Skip integration tests if not in integration test mode
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		suite.T().Skip("Skipping integration tests. Set INTEGRATION_TESTS=true to run.")
	}

	// Setup test configuration
	suite.config = &config.Config{
		Database: config.DatabaseConfig{
			Host:         testhelpers.GetEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:         5432,
			Username:     testhelpers.GetEnvOrDefault("TEST_DB_USER", "gauth"),
			Password:     testhelpers.GetEnvOrDefault("TEST_DB_PASSWORD", "password"),
			Database:     testhelpers.GetEnvOrDefault("TEST_DB_NAME", "gauth_test"),
			SSLMode:      "disable",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
			MaxLifetime:  5 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host:         testhelpers.GetEnvOrDefault("TEST_REDIS_HOST", "localhost"),
			Port:         6379,
			Password:     testhelpers.GetEnvOrDefault("TEST_REDIS_PASSWORD", ""),
			Database:     5, // Different database for REST tests
			PoolSize:     10,
			MinIdleConns: 5,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		Server: config.ServerConfig{
			Host:         "localhost",
			Port:         8083, // Different port to avoid conflicts
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		GRPC: config.GRPCConfig{
			Host:                "localhost",
			Port:                9091, // Different port to avoid conflicts
			ConnectionTimeout:   10 * time.Second,
			MaxRecvMsgSize:      4 * 1024 * 1024,
			MaxSendMsgSize:      4 * 1024 * 1024,
			KeepAliveTime:       30 * time.Second,
			KeepAliveTimeout:    5 * time.Second,
			PermitWithoutStream: true,
		},
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
		Security: config.SecurityConfig{
			CORSEnabled:      true,
			CORSOrigins:      []string{"*"},
			RateLimitEnabled: false,
		},
		Logging: config.LoggingConfig{
			Level: "debug",
		},
	}

	suite.baseURL = fmt.Sprintf("http://%s/api/v1", suite.config.GetServerAddr())
	suite.httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// Initialize dependencies
	database, err := db.NewPostgresDB(&suite.config.Database)
	require.NoError(suite.T(), err)

	redis, err := db.NewRedisClient(&suite.config.Redis)
	require.NoError(suite.T(), err)

	logger := logger.NewDefault()

	// Initialize service
	svc := service.NewGAuthService(suite.config, logger, database, redis)

	// Initialize telemetry (disabled for tests)
	telemetry, err := telemetry.New(context.Background(), telemetry.Config{
		ServiceName:        "gauth-test",
		ServiceVersion:     "test",
		Environment:        "test",
		TracingEnabled:     false,
		MetricsEnabled:     false,
		OTLPEndpoint:       "",
		OTLPInsecure:       false,
		TraceSamplingRatio: 0.1,
		MetricsPort:        0,
	})
	require.NoError(suite.T(), err)

	// Initialize and start gRPC server
	suite.grpcServer = grpcServer.NewServer(suite.config, logger, svc, telemetry)
	go func() {
		suite.grpcServer.Start()
	}()

	// Wait for gRPC server to start
	time.Sleep(2 * time.Second)

	// Initialize and start REST server
	suite.restServer = restServer.NewServer(suite.config, logger, telemetry)
	go func() {
		suite.restServer.Start()
	}()

	// Wait for REST server to start
	time.Sleep(2 * time.Second)
}

func (suite *RESTIntegrationTestSuite) TearDownSuite() {
	if suite.restServer != nil {
		suite.restServer.Stop()
	}
	if suite.grpcServer != nil {
		suite.grpcServer.Stop()
	}
}

func (suite *RESTIntegrationTestSuite) TestHealthEndpoint() {
	resp, err := suite.httpClient.Get(suite.baseURL + "/health")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	// Should return 200 or 503 depending on health status
	assert.Contains(suite.T(), []int{http.StatusOK, http.StatusServiceUnavailable}, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "status")
	assert.Contains(suite.T(), data, "services")
	assert.Contains(suite.T(), data, "timestamp")
}

func (suite *RESTIntegrationTestSuite) TestStatusEndpoint() {
	resp, err := suite.httpClient.Get(suite.baseURL + "/status")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "version")
	assert.Contains(suite.T(), data, "uptime")
}

func (suite *RESTIntegrationTestSuite) TestOrganizationCRUD() {
	// Create organization
	createReq := map[string]interface{}{
		"name":                    "REST Test Organization",
		"initial_user_email":      "admin@resttest.com",
		"initial_user_public_key": "rest-admin-public-key",
	}

	createBody, _ := json.Marshal(createReq)
	resp, err := suite.httpClient.Post(
		suite.baseURL+"/organizations",
		"application/json",
		bytes.NewBuffer(createBody),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), createResponse["success"].(bool))
	data := createResponse["data"].(map[string]interface{})
	organization := data["organization"].(map[string]interface{})
	orgID := organization["id"].(string)

	assert.Equal(suite.T(), "REST Test Organization", organization["name"])
	assert.NotEmpty(suite.T(), orgID)

	// Get organization
	resp, err = suite.httpClient.Get(suite.baseURL + "/organizations/" + orgID)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var getResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&getResponse)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), getResponse["success"].(bool))
	data = getResponse["data"].(map[string]interface{})
	organization = data["organization"].(map[string]interface{})
	assert.Equal(suite.T(), orgID, organization["id"])
	assert.Equal(suite.T(), "REST Test Organization", organization["name"])
}

func (suite *RESTIntegrationTestSuite) TestInvalidRequests() {
	// Test invalid JSON
	resp, err := suite.httpClient.Post(
		suite.baseURL+"/organizations",
		"application/json",
		bytes.NewBufferString("invalid json"),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	// Test missing required fields
	invalidReq := map[string]interface{}{
		"name": "Test Org",
		// Missing required fields
	}

	invalidBody, _ := json.Marshal(invalidReq)
	resp, err = suite.httpClient.Post(
		suite.baseURL+"/organizations",
		"application/json",
		bytes.NewBuffer(invalidBody),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	// Test non-existent resource
	resp, err = suite.httpClient.Get(suite.baseURL + "/organizations/non-existent-id")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
}

func TestRESTIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RESTIntegrationTestSuite))
}
