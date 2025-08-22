package e2e

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
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type WalletWorkflowE2ETestSuite struct {
	suite.Suite
	router     *gin.Engine
	grpcServer *grpc.Server
	restServer *rest.Server
	service    *service.GAuthService
	db         *testhelpers.TestDatabase
	redis      *testhelpers.TestRedis
	ctx        context.Context
}

func (suite *WalletWorkflowE2ETestSuite) SetupSuite() {
	testhelpers.SkipIfNotE2E(suite.T())

	suite.ctx = context.Background()

	// Setup test database
	suite.db = testhelpers.NewTestDatabase(suite.T())

	// Setup test Redis
	suite.redis = testhelpers.NewTestRedis(suite.T())

	// Run migrations
	err := suite.db.RunMigrations(suite.ctx)
	require.NoError(suite.T(), err)

	// Setup test logger
	testLogger := logger.NewDefault()

	// Setup test config
	cfg := &config.Config{
		GRPC: config.GRPCConfig{
			Host:                "0.0.0.0",
			Port:                9091,
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
	}

	// Create service instance
	suite.service = service.NewGAuthService(cfg, testLogger, suite.db.DB, suite.redis.Client)

	// Create gRPC server
	suite.grpcServer = grpc.NewServer(cfg, testLogger, suite.service)

	// Start gRPC server in background
	go func() {
		if err := suite.grpcServer.Start(); err != nil {
			suite.T().Logf("gRPC server error: %v", err)
		}
	}()

	// Wait a moment for gRPC server to start
	time.Sleep(100 * time.Millisecond)

	// Create REST server
	suite.restServer = rest.NewServer(cfg, testLogger)

	// Connect REST server to gRPC server
	err = suite.restServer.ConnectToGRPCForTesting("localhost:9091")
	require.NoError(suite.T(), err)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup router for testing
	suite.router = gin.New()
	suite.restServer.SetupAPIRoutes(suite.router)
}

func (suite *WalletWorkflowE2ETestSuite) TearDownSuite() {
	if suite.grpcServer != nil {
		suite.grpcServer.Stop()
	}
	if suite.restServer != nil {
		suite.restServer.Stop()
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

func (suite *WalletWorkflowE2ETestSuite) TearDownTest() {
	// Clean up between tests
	suite.redis.Cleanup(suite.ctx)
	suite.db.Cleanup(suite.ctx)
}

// Helper method to make HTTP requests and return the response
func (suite *WalletWorkflowE2ETestSuite) makeRequest(method, url string, body interface{}) (*httptest.ResponseRecorder, map[string]interface{}) {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var response map[string]interface{}
	if w.Body.Len() > 0 {
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)
	}

	return w, response
}

func (suite *WalletWorkflowE2ETestSuite) TestCompleteWalletWorkflow() {
	// Step 1: Create an organization (with optional public key)
	suite.T().Log("Step 1: Creating organization without public key")
	orgBody := map[string]interface{}{
		"name":               "E2E Test Corporation",
		"initial_user_email": "admin@e2etest.com",
		// No initial_user_public_key - testing optional field
	}

	w, orgResp := suite.makeRequest("POST", "/api/v1/organizations", orgBody)
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	require.True(suite.T(), orgResp["success"].(bool))

	orgData := orgResp["data"].(map[string]interface{})
	org := orgData["organization"].(map[string]interface{})
	organizationID := org["id"].(string)

	suite.T().Logf("Created organization with ID: %s", organizationID)
	assert.NotEmpty(suite.T(), organizationID)
	assert.Equal(suite.T(), "E2E Test Corporation", org["name"])

	// Check that organization was created successfully
	assert.Equal(suite.T(), "E2E Test Corporation", org["name"])
	assert.NotEmpty(suite.T(), org["id"])

	// Step 2: Create additional users (with and without public keys)
	suite.T().Log("Step 2: Creating additional users")

	// User with public key
	userWithKeyBody := map[string]interface{}{
		"organization_id": organizationID,
		"username":        "alice",
		"email":           "alice@e2etest.com",
		"public_key":      "0x1234567890abcdef1234567890abcdef12345678",
		"tags":            []string{"developer", "admin"},
	}

	w, userResp := suite.makeRequest("POST", "/api/v1/users", userWithKeyBody)
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	require.True(suite.T(), userResp["success"].(bool))

	// User without public key
	userWithoutKeyBody := map[string]interface{}{
		"organization_id": organizationID,
		"username":        "bob",
		"email":           "bob@e2etest.com",
		// No public_key - testing optional field
		"tags": []string{"reviewer"},
	}

	w, userResp2 := suite.makeRequest("POST", "/api/v1/users", userWithoutKeyBody)
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	require.True(suite.T(), userResp2["success"].(bool))

	suite.T().Log("Created users successfully")

	// Step 3: Create a simple wallet
	suite.T().Log("Step 3: Creating simple Ethereum wallet")
	simpleWalletBody := map[string]interface{}{
		"organization_id": organizationID,
		"name":            "Simple Ethereum Wallet",
		"accounts": []map[string]interface{}{
			{
				"curve":          "CURVE_SECP256K1",
				"path_format":    "PATH_FORMAT_BIP32",
				"path":           "m/44'/60'/0'/0/0",
				"address_format": "ADDRESS_FORMAT_ETHEREUM",
			},
		},
		"mnemonic_length": 12,
		"tags":            []string{"ethereum", "primary"},
	}

	w, walletResp := suite.makeRequest("POST", "/api/v1/wallets", simpleWalletBody)
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	require.True(suite.T(), walletResp["success"].(bool))

	walletData := walletResp["data"].(map[string]interface{})
	wallet := walletData["wallet"].(map[string]interface{})
	addresses := walletData["addresses"].([]interface{})

	simpleWalletID := wallet["id"].(string)
	suite.T().Logf("Created simple wallet with ID: %s", simpleWalletID)

	assert.Equal(suite.T(), "Simple Ethereum Wallet", wallet["name"])
	assert.Len(suite.T(), addresses, 1)
	assert.NotEmpty(suite.T(), addresses[0])

	// Step 4: Create a multi-chain wallet
	suite.T().Log("Step 4: Creating multi-chain wallet")
	multiChainWalletBody := map[string]interface{}{
		"organization_id": organizationID,
		"name":            "Multi-Chain Portfolio",
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
			{
				"curve":          "CURVE_ED25519",
				"path_format":    "PATH_FORMAT_BIP32",
				"path":           "m/44'/501'/0'/0",
				"address_format": "ADDRESS_FORMAT_SOLANA",
			},
		},
		"mnemonic_length": 24,
		"tags":            []string{"multi-chain", "production", "diversified"},
	}

	w, multiWalletResp := suite.makeRequest("POST", "/api/v1/wallets", multiChainWalletBody)
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	require.True(suite.T(), multiWalletResp["success"].(bool))

	multiWalletData := multiWalletResp["data"].(map[string]interface{})
	multiWallet := multiWalletData["wallet"].(map[string]interface{})
	multiAddresses := multiWalletData["addresses"].([]interface{})

	multiWalletID := multiWallet["id"].(string)
	suite.T().Logf("Created multi-chain wallet with ID: %s", multiWalletID)

	assert.Equal(suite.T(), "Multi-Chain Portfolio", multiWallet["name"])
	assert.Len(suite.T(), multiAddresses, 3)

	// Verify all addresses are unique
	addressSet := make(map[string]bool)
	for _, addr := range multiAddresses {
		address := addr.(string)
		assert.False(suite.T(), addressSet[address], "Duplicate address found")
		addressSet[address] = true
	}

	// Step 5: Create private keys
	suite.T().Log("Step 5: Creating private keys")

	// SECP256K1 key for Ethereum
	ethKeyBody := map[string]interface{}{
		"organization_id": organizationID,
		"name":            "Ethereum Signing Key",
		"curve":           "CURVE_SECP256K1",
		"tags":            []string{"ethereum", "signing", "hot"},
	}

	w, ethKeyResp := suite.makeRequest("POST", "/api/v1/private-keys", ethKeyBody)
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	require.True(suite.T(), ethKeyResp["success"].(bool))

	ethKeyData := ethKeyResp["data"].(map[string]interface{})
	ethKey := ethKeyData["private_key"].(map[string]interface{})
	ethKeyID := ethKey["id"].(string)

	suite.T().Logf("Created Ethereum private key with ID: %s", ethKeyID)
	assert.Equal(suite.T(), "Ethereum Signing Key", ethKey["name"])
	assert.Equal(suite.T(), "CURVE_SECP256K1", ethKey["curve"])

	// ED25519 key for Solana
	solKeyBody := map[string]interface{}{
		"organization_id": organizationID,
		"name":            "Solana Validator Key",
		"curve":           "CURVE_ED25519",
		"tags":            []string{"solana", "validator", "staking"},
	}

	w, solKeyResp := suite.makeRequest("POST", "/api/v1/private-keys", solKeyBody)
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	require.True(suite.T(), solKeyResp["success"].(bool))

	solKeyData := solKeyResp["data"].(map[string]interface{})
	solKey := solKeyData["private_key"].(map[string]interface{})
	solKeyID := solKey["id"].(string)

	suite.T().Logf("Created Solana private key with ID: %s", solKeyID)
	assert.Equal(suite.T(), "Solana Validator Key", solKey["name"])
	assert.Equal(suite.T(), "CURVE_ED25519", solKey["curve"])

	// Step 6: List and verify all created resources
	suite.T().Log("Step 6: Listing and verifying all resources")

	// List wallets
	w, listWalletsResp := suite.makeRequest("GET",
		fmt.Sprintf("/api/v1/wallets?organization_id=%s&page_size=10", organizationID), nil)
	require.Equal(suite.T(), http.StatusOK, w.Code)
	require.True(suite.T(), listWalletsResp["success"].(bool))

	walletsData := listWalletsResp["data"].(map[string]interface{})
	wallets := walletsData["wallets"].([]interface{})

	assert.Len(suite.T(), wallets, 2)
	suite.T().Logf("Listed %d wallets", len(wallets))

	// List private keys
	w, listKeysResp := suite.makeRequest("GET",
		fmt.Sprintf("/api/v1/private-keys?organization_id=%s&page_size=10", organizationID), nil)
	require.Equal(suite.T(), http.StatusOK, w.Code)
	require.True(suite.T(), listKeysResp["success"].(bool))

	keysData := listKeysResp["data"].(map[string]interface{})
	keys := keysData["private_keys"].([]interface{})

	assert.Len(suite.T(), keys, 2)
	suite.T().Logf("Listed %d private keys", len(keys))

	// Step 7: Get individual resources
	suite.T().Log("Step 7: Getting individual resources")

	// Get simple wallet
	w, getWalletResp := suite.makeRequest("GET",
		fmt.Sprintf("/api/v1/wallets/%s", simpleWalletID), nil)
	require.Equal(suite.T(), http.StatusOK, w.Code)
	require.True(suite.T(), getWalletResp["success"].(bool))

	getWalletData := getWalletResp["data"].(map[string]interface{})
	getWallet := getWalletData["wallet"].(map[string]interface{})

	assert.Equal(suite.T(), simpleWalletID, getWallet["id"])
	assert.Equal(suite.T(), "Simple Ethereum Wallet", getWallet["name"])
	assert.NotNil(suite.T(), getWallet["accounts"])

	// Get private key
	w, getKeyResp := suite.makeRequest("GET",
		fmt.Sprintf("/api/v1/private-keys/%s", ethKeyID), nil)
	require.Equal(suite.T(), http.StatusOK, w.Code)
	require.True(suite.T(), getKeyResp["success"].(bool))

	getKeyData := getKeyResp["data"].(map[string]interface{})
	getKey := getKeyData["private_key"].(map[string]interface{})

	assert.Equal(suite.T(), ethKeyID, getKey["id"])
	assert.Equal(suite.T(), "Ethereum Signing Key", getKey["name"])

	// Step 8: Update organization with additional user (testing quorum)
	suite.T().Log("Step 8: Testing quorum scenarios")

	// Create another user for quorum testing
	quorumUserBody := map[string]interface{}{
		"organization_id": organizationID,
		"username":        "charlie",
		"email":           "charlie@e2etest.com",
		"public_key":      "0xabcdef1234567890abcdef1234567890abcdef12",
		"tags":            []string{"quorum", "security"},
	}

	w, quorumUserResp := suite.makeRequest("POST", "/api/v1/users", quorumUserBody)
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	require.True(suite.T(), quorumUserResp["success"].(bool))

	suite.T().Log("Created quorum user successfully")

	// Step 9: Performance test with multiple wallets
	suite.T().Log("Step 9: Performance testing with multiple wallet creation")

	startTime := time.Now()
	walletCount := 5

	for i := 0; i < walletCount; i++ {
		perfWalletBody := map[string]interface{}{
			"organization_id": organizationID,
			"name":            fmt.Sprintf("Performance Test Wallet %d", i+1),
			"accounts": []map[string]interface{}{
				{
					"curve":          "CURVE_SECP256K1",
					"path_format":    "PATH_FORMAT_BIP32",
					"path":           fmt.Sprintf("m/44'/60'/0'/0/%d", i),
					"address_format": "ADDRESS_FORMAT_ETHEREUM",
				},
			},
			"mnemonic_length": 12,
			"tags":            []string{"performance", "test"},
		}

		w, perfResp := suite.makeRequest("POST", "/api/v1/wallets", perfWalletBody)
		require.Equal(suite.T(), http.StatusCreated, w.Code)
		require.True(suite.T(), perfResp["success"].(bool))
	}

	duration := time.Since(startTime)
	suite.T().Logf("Created %d wallets in %v (avg: %v per wallet)",
		walletCount, duration, duration/time.Duration(walletCount))

	// Performance should be reasonable
	assert.Less(suite.T(), duration, 10*time.Second, "Wallet creation took too long")

	// Step 10: Cleanup test - Delete a wallet
	suite.T().Log("Step 10: Testing wallet deletion")

	deleteBody := map[string]interface{}{
		"delete_without_export": true, // Force delete for testing
	}

	w, deleteResp := suite.makeRequest("DELETE",
		fmt.Sprintf("/api/v1/wallets/%s", simpleWalletID), deleteBody)
	require.Equal(suite.T(), http.StatusOK, w.Code)
	require.True(suite.T(), deleteResp["success"].(bool))

	// Verify wallet is deleted
	w, _ = suite.makeRequest("GET", fmt.Sprintf("/api/v1/wallets/%s", simpleWalletID), nil)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	suite.T().Log("Wallet deletion successful")

	// Step 11: Final verification
	suite.T().Log("Step 11: Final verification")

	// List wallets again to verify deletion
	w, finalListResp := suite.makeRequest("GET",
		fmt.Sprintf("/api/v1/wallets?organization_id=%s", organizationID), nil)
	require.Equal(suite.T(), http.StatusOK, w.Code)
	require.True(suite.T(), finalListResp["success"].(bool))

	finalWalletsData := finalListResp["data"].(map[string]interface{})
	finalWallets := finalWalletsData["wallets"].([]interface{})

	// Should have multi-chain wallet + 5 performance test wallets
	assert.Len(suite.T(), finalWallets, 6)

	suite.T().Log("Complete wallet workflow test passed!")
}

func (suite *WalletWorkflowE2ETestSuite) TestErrorHandlingWorkflow() {
	suite.T().Log("Testing error handling scenarios")

	// Test invalid organization creation
	invalidOrgBody := map[string]interface{}{
		"name":               "",              // Empty name should fail
		"initial_user_email": "invalid-email", // Invalid email
	}

	w, resp := suite.makeRequest("POST", "/api/v1/organizations", invalidOrgBody)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.NotNil(suite.T(), resp["error"])

	// Test invalid wallet creation
	invalidWalletBody := map[string]interface{}{
		"organization_id": "invalid-uuid",
		"name":            "Test Wallet",
		"accounts": []map[string]interface{}{
			{
				"curve":          "INVALID_CURVE",
				"path_format":    "PATH_FORMAT_BIP32",
				"path":           "invalid-path",
				"address_format": "INVALID_FORMAT",
			},
		},
		"mnemonic_length": 13, // Invalid length
	}

	w, resp = suite.makeRequest("POST", "/api/v1/wallets", invalidWalletBody)
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.NotNil(suite.T(), resp["error"])

	// Test invalid private key creation
	invalidKeyBody := map[string]interface{}{
		"organization_id": "invalid-uuid",
		"name":            "Test Key",
		"curve":           "INVALID_CURVE",
	}

	w, resp = suite.makeRequest("POST", "/api/v1/private-keys", invalidKeyBody)
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.NotNil(suite.T(), resp["error"])

	suite.T().Log("Error handling tests passed")
}

func TestWalletWorkflowE2ETestSuite(t *testing.T) {
	suite.Run(t, new(WalletWorkflowE2ETestSuite))
}
