package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/dojima-foundation/tee-auth/gauth/internal/grpc"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type WalletGRPCTestSuite struct {
	suite.Suite
	server         *grpc.Server
	service        *service.GAuthService
	db             *testhelpers.TestDatabase
	organizationID string
	ctx            context.Context
}

func (suite *WalletGRPCTestSuite) SetupSuite() {
	testhelpers.SkipIfNotIntegration(suite.T())

	suite.ctx = context.Background()

	// Setup test database
	suite.db = testhelpers.NewTestDatabase(suite.T())

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
	suite.service = service.NewGAuthServiceWithEnclave(cfg, testLogger, suite.db.DB, nil, service.NewMockRenclaveClient())

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

	// Create gRPC server
	suite.server = grpc.NewServer(cfg, testLogger, suite.service, telemetry)

	// Create test organization
	org, _, err := suite.db.CreateTestOrganization(suite.ctx, "Test Organization", 1)
	require.NoError(suite.T(), err)
	suite.organizationID = org.ID.String()
}

func (suite *WalletGRPCTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Stop()
	}
	if suite.db != nil {
		suite.db.Cleanup(suite.ctx)
		suite.db.Close()
	}
}

func (suite *WalletGRPCTestSuite) TearDownTest() {
	// Clean up wallets created during tests
	suite.db.DB.GetDB().Exec("DELETE FROM wallet_accounts")
	suite.db.DB.GetDB().Exec("DELETE FROM wallets")
}

func (suite *WalletGRPCTestSuite) TestCreateWallet_Success() {
	req := &pb.CreateWalletRequest{
		OrganizationId: suite.organizationID,
		Name:           "Test Ethereum Wallet",
		Accounts: []*pb.CreateWalletAccount{
			{
				Curve:         "CURVE_SECP256K1",
				PathFormat:    "PATH_FORMAT_BIP32",
				Path:          "m/44'/60'/0'/0/0",
				AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
			},
		},
		MnemonicLength: testhelpers.Int32Ptr(12),
		Tags:           []string{"ethereum", "test"},
	}

	resp, err := suite.server.CreateWallet(suite.ctx, req)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.NotNil(suite.T(), resp.Wallet)
	assert.Equal(suite.T(), "Test Ethereum Wallet", resp.Wallet.Name)
	assert.Equal(suite.T(), suite.organizationID, resp.Wallet.OrganizationId)
	assert.Len(suite.T(), resp.Addresses, 1)
	assert.Len(suite.T(), resp.Wallet.Accounts, 1)

	// Check account details
	account := resp.Wallet.Accounts[0]
	assert.Equal(suite.T(), "CURVE_SECP256K1", account.Curve)
	assert.Equal(suite.T(), "m/44'/60'/0'/0/0", account.Path)
	assert.Equal(suite.T(), "ADDRESS_FORMAT_ETHEREUM", account.AddressFormat)
	assert.NotEmpty(suite.T(), account.Address)
	assert.NotEmpty(suite.T(), account.PublicKey)
	assert.True(suite.T(), account.IsActive)
}

func (suite *WalletGRPCTestSuite) TestCreateWallet_MultipleAccounts() {
	req := &pb.CreateWalletRequest{
		OrganizationId: suite.organizationID,
		Name:           "Multi-Chain Wallet",
		Accounts: []*pb.CreateWalletAccount{
			{
				Curve:         "CURVE_SECP256K1",
				PathFormat:    "PATH_FORMAT_BIP32",
				Path:          "m/44'/60'/0'/0/0",
				AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
			},
			{
				Curve:         "CURVE_SECP256K1",
				PathFormat:    "PATH_FORMAT_BIP32",
				Path:          "m/44'/0'/0'/0/0",
				AddressFormat: "ADDRESS_FORMAT_BITCOIN_MAINNET_P2WPKH",
			},
			{
				Curve:         "CURVE_ED25519",
				PathFormat:    "PATH_FORMAT_BIP32",
				Path:          "m/44'/501'/0'/0",
				AddressFormat: "ADDRESS_FORMAT_SOLANA",
			},
		},
		MnemonicLength: testhelpers.Int32Ptr(24),
		Tags:           []string{"multi-chain", "production"},
	}

	resp, err := suite.server.CreateWallet(suite.ctx, req)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Len(suite.T(), resp.Addresses, 3)
	assert.Len(suite.T(), resp.Wallet.Accounts, 3)

	// Verify all addresses are unique
	addressSet := make(map[string]bool)
	for _, addr := range resp.Addresses {
		assert.False(suite.T(), addressSet[addr], "Duplicate address generated")
		addressSet[addr] = true
	}
}

func (suite *WalletGRPCTestSuite) TestCreateWallet_InvalidMnemonicLength() {
	req := &pb.CreateWalletRequest{
		OrganizationId: suite.organizationID,
		Name:           "Invalid Wallet",
		Accounts: []*pb.CreateWalletAccount{
			{
				Curve:         "CURVE_SECP256K1",
				PathFormat:    "PATH_FORMAT_BIP32",
				Path:          "m/44'/60'/0'/0/0",
				AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
			},
		},
		MnemonicLength: testhelpers.Int32Ptr(13), // Invalid
	}

	resp, err := suite.server.CreateWallet(suite.ctx, req)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), resp)
	assert.Contains(suite.T(), err.Error(), "invalid mnemonic length")
}

func (suite *WalletGRPCTestSuite) TestGetWallet_Success() {
	// Create a wallet first
	createReq := &pb.CreateWalletRequest{
		OrganizationId: suite.organizationID,
		Name:           "Get Test Wallet",
		Accounts: []*pb.CreateWalletAccount{
			{
				Curve:         "CURVE_SECP256K1",
				PathFormat:    "PATH_FORMAT_BIP32",
				Path:          "m/44'/60'/0'/0/0",
				AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
			},
		},
		Tags: []string{"test"},
	}

	createResp, err := suite.server.CreateWallet(suite.ctx, createReq)
	require.NoError(suite.T(), err)

	// Get the wallet
	getReq := &pb.GetWalletRequest{
		Id: createResp.Wallet.Id,
	}

	getResp, err := suite.server.GetWallet(suite.ctx, getReq)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), getResp)
	assert.Equal(suite.T(), createResp.Wallet.Id, getResp.Wallet.Id)
	assert.Equal(suite.T(), createResp.Wallet.Name, getResp.Wallet.Name)
	assert.Len(suite.T(), getResp.Wallet.Accounts, 1)
}

func (suite *WalletGRPCTestSuite) TestGetWallet_NotFound() {
	req := &pb.GetWalletRequest{
		Id: uuid.New().String(),
	}

	resp, err := suite.server.GetWallet(suite.ctx, req)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), resp)
	assert.Contains(suite.T(), err.Error(), "wallet not found")
}

func (suite *WalletGRPCTestSuite) TestListWallets_Success() {
	// Create multiple wallets
	for i := 0; i < 3; i++ {
		createReq := &pb.CreateWalletRequest{
			OrganizationId: suite.organizationID,
			Name:           fmt.Sprintf("List Test Wallet %d", i+1),
			Accounts: []*pb.CreateWalletAccount{
				{
					Curve:         "CURVE_SECP256K1",
					PathFormat:    "PATH_FORMAT_BIP32",
					Path:          "m/44'/60'/0'/0/0",
					AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
				},
			},
			Tags: []string{"test", "list"},
		}

		_, err := suite.server.CreateWallet(suite.ctx, createReq)
		require.NoError(suite.T(), err)
	}

	// List wallets
	req := &pb.ListWalletsRequest{
		OrganizationId: suite.organizationID,
		PageSize:       10,
		PageToken:      "",
	}

	resp, err := suite.server.ListWallets(suite.ctx, req)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Len(suite.T(), resp.Wallets, 3)
	assert.Empty(suite.T(), resp.NextPageToken) // No pagination needed

	// Check all wallets have accounts
	for _, wallet := range resp.Wallets {
		assert.NotEmpty(suite.T(), wallet.Accounts)
	}
}

func (suite *WalletGRPCTestSuite) TestListWallets_Pagination() {
	// Create multiple wallets
	for i := 0; i < 5; i++ {
		createReq := &pb.CreateWalletRequest{
			OrganizationId: suite.organizationID,
			Name:           fmt.Sprintf("Pagination Test Wallet %d", i+1),
			Accounts: []*pb.CreateWalletAccount{
				{
					Curve:         "CURVE_SECP256K1",
					PathFormat:    "PATH_FORMAT_BIP32",
					Path:          "m/44'/60'/0'/0/0",
					AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
				},
			},
			Tags: []string{"pagination"},
		}

		_, err := suite.server.CreateWallet(suite.ctx, createReq)
		require.NoError(suite.T(), err)
	}

	// List wallets with pagination
	req := &pb.ListWalletsRequest{
		OrganizationId: suite.organizationID,
		PageSize:       2, // Small page size
		PageToken:      "",
	}

	resp, err := suite.server.ListWallets(suite.ctx, req)

	// Assertions
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), resp.Wallets, 2)
	assert.NotEmpty(suite.T(), resp.NextPageToken) // Should have next page
}

func (suite *WalletGRPCTestSuite) TestDeleteWallet_Success() {
	// Create a wallet first
	createReq := &pb.CreateWalletRequest{
		OrganizationId: suite.organizationID,
		Name:           "Delete Test Wallet",
		Accounts: []*pb.CreateWalletAccount{
			{
				Curve:         "CURVE_SECP256K1",
				PathFormat:    "PATH_FORMAT_BIP32",
				Path:          "m/44'/60'/0'/0/0",
				AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
			},
		},
	}

	createResp, err := suite.server.CreateWallet(suite.ctx, createReq)
	require.NoError(suite.T(), err)

	// Delete the wallet
	deleteReq := &pb.DeleteWalletRequest{
		Id:                  createResp.Wallet.Id,
		DeleteWithoutExport: testhelpers.BoolPtr(true), // Force delete
	}

	deleteResp, err := suite.server.DeleteWallet(suite.ctx, deleteReq)

	// Assertions
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), deleteResp)
	assert.True(suite.T(), deleteResp.Success)
	assert.NotEmpty(suite.T(), deleteResp.Message)

	// Verify wallet is deleted
	getReq := &pb.GetWalletRequest{Id: createResp.Wallet.Id}
	getResp, err := suite.server.GetWallet(suite.ctx, getReq)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), getResp)
}

func (suite *WalletGRPCTestSuite) TestDeleteWallet_NotFound() {
	req := &pb.DeleteWalletRequest{
		Id: uuid.New().String(),
	}

	resp, err := suite.server.DeleteWallet(suite.ctx, req)

	// Assertions - deleting non-existent wallet should succeed (idempotent)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.Success)
}

func TestWalletGRPCTestSuite(t *testing.T) {
	suite.Run(t, new(WalletGRPCTestSuite))
}
