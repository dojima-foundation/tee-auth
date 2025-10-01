package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/api/rest"
	"github.com/dojima-foundation/tee-auth/gauth/internal/grpc"
	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type WalletRESTTestSuite struct {
	suite.Suite
	router         *gin.Engine
	grpcServer     *grpc.Server
	restServer     *rest.Server
	service        *service.GAuthService
	db             *testhelpers.TestDatabase
	redis          *testhelpers.TestRedis
	organizationID string
	testUser       *models.User
	testAuthMethod *models.AuthMethod
	testSessionID  string
	ctx            context.Context
}

func (suite *WalletRESTTestSuite) SetupSuite() {
	testhelpers.SkipIfNotIntegration(suite.T())

	suite.ctx = context.Background()

	// Setup test database
	suite.db = testhelpers.NewTestDatabase(suite.T())

	// Run migrations
	err := suite.db.RunMigrations(suite.ctx)
	require.NoError(suite.T(), err)

	// Setup test Redis
	suite.redis = testhelpers.NewTestRedis(suite.T())

	// Setup test logger
	testLogger := logger.NewDefault()

	// Setup test config
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Host:                "0.0.0.0",
			Port:                9093,             // Use different port to avoid conflicts with WalletGRPCTestSuite
			MaxRecvMsgSize:      10 * 1024 * 1024, // 10MB
			MaxSendMsgSize:      10 * 1024 * 1024, // 10MB
			ConnectionTimeout:   10 * time.Second,
			KeepAliveTime:       30 * time.Second,
			KeepAliveTimeout:    5 * time.Second,
			PermitWithoutStream: true,
		},
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
		},
		Security: config.SecurityConfig{
			TLSEnabled:  false,
			CORSOrigins: []string{"*"},
		},
	}

	// Create service instance
	suite.service = service.NewGAuthServiceWithEnclave(cfg, testLogger, suite.db.DB, suite.redis.Client, service.NewMockRenclaveClient())

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

	// Create gRPC server (needed for REST server)
	suite.grpcServer = grpc.NewServer(cfg, testLogger, suite.service, telemetry)

	// Start gRPC server in background
	go func() {
		if err := suite.grpcServer.Start(); err != nil {
			suite.T().Logf("gRPC server error: %v", err)
		}
	}()

	// Wait a moment for gRPC server to start
	time.Sleep(100 * time.Millisecond)

	// Create REST server
	suite.restServer = rest.NewServer(cfg, testLogger, telemetry)

	// Set Redis client for session management
	suite.restServer.SetRedis(suite.redis.Client)

	// Enable test mode to bypass session validation
	suite.restServer.GetSessionManager().SetTestMode(true)

	// Connect REST server to gRPC server
	err = suite.restServer.ConnectToGRPCForTesting("localhost:9093")
	require.NoError(suite.T(), err)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup router for testing
	suite.router = gin.New()

	// Setup API routes normally - we'll override session validation
	suite.restServer.SetupAPIRoutes(suite.router)

	// Create test organization and user
	org, users, err := suite.db.CreateTestOrganization(suite.ctx, "Test Organization", 1)
	require.NoError(suite.T(), err)
	suite.organizationID = org.ID.String()
	suite.testUser = &users[0]

	// Create test auth method
	suite.testAuthMethod = &models.AuthMethod{
		ID:     uuid.New(),
		UserID: suite.testUser.ID,
		Type:   "OAUTH",
		Name:   "Google OAuth",
	}

	// Set test data for session manager
	suite.restServer.GetSessionManager().SetTestData(
		suite.organizationID,
		suite.testUser.ID.String(),
		suite.testAuthMethod.ID.String(),
	)

	// Create test session
	suite.testSessionID, err = suite.createTestSession()
	require.NoError(suite.T(), err)
}

func (suite *WalletRESTTestSuite) TearDownSuite() {
	if suite.grpcServer != nil {
		suite.grpcServer.Stop()
	}
	if suite.restServer != nil {
		_ = suite.restServer.Stop()
	}
	if suite.redis != nil {
		suite.redis.Cleanup(suite.ctx)
		suite.redis.Close()
	}
	if suite.db != nil {
		suite.db.Cleanup(suite.ctx)
		suite.db.Close()
	}
}

// createTestSession creates a test session using the session manager
func (suite *WalletRESTTestSuite) createTestSession() (string, error) {
	sessionManager := suite.restServer.GetSessionManager()
	return sessionManager.CreateSession(suite.ctx, suite.testUser, suite.testAuthMethod, "google")
}

func (suite *WalletRESTTestSuite) TearDownTest() {
	// Clean up wallets created during tests
	suite.db.DB.GetDB().Exec("DELETE FROM wallet_accounts")
	suite.db.DB.GetDB().Exec("DELETE FROM wallets")
}

func (suite *WalletRESTTestSuite) TestCreateWallet_Success() {
	requestBody := map[string]interface{}{
		"organization_id": suite.organizationID,
		"name":            "Test REST Wallet",
		"accounts": []map[string]interface{}{
			{
				"curve":          "CURVE_SECP256K1",
				"path_format":    "PATH_FORMAT_BIP32",
				"path":           "m/44'/60'/0'/0/0",
				"address_format": "ADDRESS_FORMAT_ETHEREUM",
			},
		},
		"mnemonic_length": 12,
		"tags":            []string{"ethereum", "rest-test"},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])

	data := response["data"].(map[string]interface{})
	wallet := data["wallet"].(map[string]interface{})
	addresses := data["addresses"].([]interface{})

	assert.Equal(suite.T(), "Test REST Wallet", wallet["name"])
	assert.Equal(suite.T(), suite.organizationID, wallet["organization_id"])
	assert.Len(suite.T(), addresses, 1)
	assert.NotEmpty(suite.T(), addresses[0])
}

func (suite *WalletRESTTestSuite) TestCreateWallet_InvalidRequest() {
	requestBody := map[string]interface{}{
		"organization_id": suite.organizationID,
		// Missing required "name" field
		"accounts": []map[string]interface{}{
			{
				"curve":          "CURVE_SECP256K1",
				"path_format":    "PATH_FORMAT_BIP32",
				"path":           "m/44'/60'/0'/0/0",
				"address_format": "ADDRESS_FORMAT_ETHEREUM",
			},
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.NotNil(suite.T(), response["error"])
}

func (suite *WalletRESTTestSuite) TestGetWallet_Success() {
	// Create a wallet first
	createBody := map[string]interface{}{
		"organization_id": suite.organizationID,
		"name":            "Get Test Wallet",
		"accounts": []map[string]interface{}{
			{
				"curve":          "CURVE_SECP256K1",
				"path_format":    "PATH_FORMAT_BIP32",
				"path":           "m/44'/60'/0'/0/0",
				"address_format": "ADDRESS_FORMAT_ETHEREUM",
			},
		},
		"tags": []string{"get-test"},
	}

	createBodyBytes, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(createBodyBytes))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	suite.router.ServeHTTP(createW, createReq)
	require.Equal(suite.T(), http.StatusCreated, createW.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(createW.Body.Bytes(), &createResponse)
	require.NoError(suite.T(), err)

	walletData := createResponse["data"].(map[string]interface{})
	wallet := walletData["wallet"].(map[string]interface{})
	walletID := wallet["id"].(string)

	// Get the wallet
	getReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/wallets/%s", walletID), nil)
	getW := httptest.NewRecorder()
	suite.router.ServeHTTP(getW, getReq)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, getW.Code)

	var getResponse map[string]interface{}
	err = json.Unmarshal(getW.Body.Bytes(), &getResponse)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), getResponse["success"].(bool))

	getData := getResponse["data"].(map[string]interface{})
	getWallet := getData["wallet"].(map[string]interface{})

	assert.Equal(suite.T(), walletID, getWallet["id"])
	assert.Equal(suite.T(), "Get Test Wallet", getWallet["name"])
}

func (suite *WalletRESTTestSuite) TestGetWallet_NotFound() {
	nonExistentID := uuid.New().String()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/wallets/%s", nonExistentID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.NotNil(suite.T(), response["error"])
}

func (suite *WalletRESTTestSuite) TestListWallets_Success() {
	// Create multiple wallets
	for i := 0; i < 3; i++ {
		createBody := map[string]interface{}{
			"organization_id": suite.organizationID,
			"name":            fmt.Sprintf("List Test Wallet %d", i+1),
			"accounts": []map[string]interface{}{
				{
					"curve":          "CURVE_SECP256K1",
					"path_format":    "PATH_FORMAT_BIP32",
					"path":           "m/44'/60'/0'/0/0",
					"address_format": "ADDRESS_FORMAT_ETHEREUM",
				},
			},
			"tags": []string{"list-test"},
		}

		body, _ := json.Marshal(createBody)
		req := httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		require.Equal(suite.T(), http.StatusCreated, w.Code)
	}

	// List wallets
	listReq := httptest.NewRequest("GET",
		fmt.Sprintf("/api/v1/wallets?organization_id=%s&page_size=10", suite.organizationID), nil)
	listW := httptest.NewRecorder()
	suite.router.ServeHTTP(listW, listReq)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, listW.Code)

	var response map[string]interface{}
	err := json.Unmarshal(listW.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	wallets := data["wallets"].([]interface{})

	assert.Len(suite.T(), wallets, 3)

	// Check each wallet has accounts
	for _, w := range wallets {
		wallet := w.(map[string]interface{})
		assert.NotNil(suite.T(), wallet["accounts"])
	}
}

func (suite *WalletRESTTestSuite) TestListWallets_MissingOrganizationID() {
	req := httptest.NewRequest("GET", "/api/v1/wallets", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), response["error"].(string), "organization_id")
}

func (suite *WalletRESTTestSuite) TestDeleteWallet_Success() {
	// Create a wallet first
	createBody := map[string]interface{}{
		"organization_id": suite.organizationID,
		"name":            "Delete Test Wallet",
		"accounts": []map[string]interface{}{
			{
				"curve":          "CURVE_SECP256K1",
				"path_format":    "PATH_FORMAT_BIP32",
				"path":           "m/44'/60'/0'/0/0",
				"address_format": "ADDRESS_FORMAT_ETHEREUM",
			},
		},
	}

	createBodyBytes, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(createBodyBytes))
	createReq.Header.Set("Content-Type", "application/json")

	createW := httptest.NewRecorder()
	suite.router.ServeHTTP(createW, createReq)
	require.Equal(suite.T(), http.StatusCreated, createW.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(createW.Body.Bytes(), &createResponse)
	require.NoError(suite.T(), err)

	walletData := createResponse["data"].(map[string]interface{})
	wallet := walletData["wallet"].(map[string]interface{})
	walletID := wallet["id"].(string)

	// Delete the wallet
	deleteBody := map[string]interface{}{
		"delete_without_export": true,
	}

	deleteBodyBytes, _ := json.Marshal(deleteBody)
	deleteReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/wallets/%s", walletID),
		bytes.NewBuffer(deleteBodyBytes))
	deleteReq.Header.Set("Content-Type", "application/json")

	deleteW := httptest.NewRecorder()
	suite.router.ServeHTTP(deleteW, deleteReq)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, deleteW.Code)

	var deleteResponse map[string]interface{}
	err = json.Unmarshal(deleteW.Body.Bytes(), &deleteResponse)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), deleteResponse["success"].(bool))

	data := deleteResponse["data"].(map[string]interface{})
	assert.True(suite.T(), data["success"].(bool))
	assert.NotEmpty(suite.T(), data["message"])

	// Verify wallet is deleted by trying to get it
	getReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/wallets/%s", walletID), nil)
	getW := httptest.NewRecorder()
	suite.router.ServeHTTP(getW, getReq)
	assert.Equal(suite.T(), http.StatusNotFound, getW.Code)
}

func (suite *WalletRESTTestSuite) TestCreateWallet_MultipleAccountsSuccess() {
	requestBody := map[string]interface{}{
		"organization_id": suite.organizationID,
		"name":            "Multi-Account Wallet",
		"accounts": []map[string]interface{}{
			{
				"curve":          "CURVE_SECP256K1",
				"path_format":    "PATH_FORMAT_BIP32",
				"path":           "m/44'/60'/0'/0/0",
				"address_format": "ADDRESS_FORMAT_ETHEREUM",
			},
			{
				"curve":          "CURVE_SECP256K1",
				"path_format":    "PATH_FORMAT_BIP32",
				"path":           "m/44'/0'/0'/0/0",
				"address_format": "ADDRESS_FORMAT_BITCOIN_MAINNET_P2WPKH",
			},
		},
		"mnemonic_length": 24,
		"tags":            []string{"multi-chain"},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	wallet := data["wallet"].(map[string]interface{})
	addresses := data["addresses"].([]interface{})
	accounts := wallet["accounts"].([]interface{})

	assert.Len(suite.T(), addresses, 2)
	assert.Len(suite.T(), accounts, 2)

	// Verify addresses are unique
	addressSet := make(map[string]bool)
	for _, addr := range addresses {
		address := addr.(string)
		assert.False(suite.T(), addressSet[address], "Duplicate address found")
		addressSet[address] = true
	}
}

func TestWalletRESTTestSuite(t *testing.T) {
	suite.Run(t, new(WalletRESTTestSuite))
}
