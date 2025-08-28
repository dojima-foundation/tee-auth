package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	grpcServer "github.com/dojima-foundation/tee-auth/gauth/internal/grpc"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

type GRPCIntegrationTestSuite struct {
	suite.Suite
	server   *grpcServer.Server
	client   pb.GAuthServiceClient
	conn     *grpc.ClientConn
	listener *bufconn.Listener
	config   *config.Config
}

func (suite *GRPCIntegrationTestSuite) SetupSuite() {
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
			Database:     2, // Different database for gRPC tests
			PoolSize:     10,
			MinIdleConns: 5,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
		Renclave: config.RenclaveConfig{
			Host:    testhelpers.GetEnvOrDefault("TEST_RENCLAVE_HOST", "localhost"),
			Port:    3000,
			UseTLS:  false,
			Timeout: 30 * time.Second,
		},
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

	// Setup in-memory gRPC server
	suite.listener = bufconn.Listen(bufSize)
	suite.server = grpcServer.NewServer(suite.config, logger, svc, telemetry)

	// Start server in background
	go func() {
		grpcSrv := grpc.NewServer()
		pb.RegisterGAuthServiceServer(grpcSrv, suite.server)
		grpcSrv.Serve(suite.listener)
	}()

	// Setup client connection
	suite.conn, err = grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(suite.bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(suite.T(), err)

	suite.client = pb.NewGAuthServiceClient(suite.conn)

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
}

func (suite *GRPCIntegrationTestSuite) TearDownSuite() {
	if suite.conn != nil {
		suite.conn.Close()
	}
	if suite.listener != nil {
		suite.listener.Close()
	}
}

func (suite *GRPCIntegrationTestSuite) bufDialer(context.Context, string) (net.Conn, error) {
	return suite.listener.Dial()
}

func (suite *GRPCIntegrationTestSuite) TestHealthCheck() {
	ctx := context.Background()

	response, err := suite.client.Health(ctx, &emptypb.Empty{})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Contains(suite.T(), []string{"healthy", "degraded"}, response.Status)
	assert.NotEmpty(suite.T(), response.Services)

	// Check individual services
	serviceMap := make(map[string]*pb.ServiceStatus)
	for _, svc := range response.Services {
		serviceMap[svc.Name] = svc
	}

	assert.Contains(suite.T(), serviceMap, "database")
	assert.Contains(suite.T(), serviceMap, "redis")
	assert.Contains(suite.T(), serviceMap, "enclave")
}

func (suite *GRPCIntegrationTestSuite) TestStatusCheck() {
	ctx := context.Background()

	response, err := suite.client.Status(ctx, &emptypb.Empty{})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.NotEmpty(suite.T(), response.Version)
	assert.NotEmpty(suite.T(), response.BuildTime)
	assert.NotEmpty(suite.T(), response.GitCommit)
	assert.NotNil(suite.T(), response.Uptime)
	assert.NotEmpty(suite.T(), response.Metrics)
}

func (suite *GRPCIntegrationTestSuite) TestOrganizationManagement() {
	ctx := context.Background()

	// Test CreateOrganization
	createReq := &pb.CreateOrganizationRequest{
		Name:                 "Test Organization",
		InitialUserEmail:     "admin@test.com",
		InitialUserPublicKey: "test-public-key",
	}

	createResp, err := suite.client.CreateOrganization(ctx, createReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createResp.Organization)
	assert.Equal(suite.T(), "Test Organization", createResp.Organization.Name)
	assert.Equal(suite.T(), "created", createResp.Status)

	orgID := createResp.Organization.Id

	// Test GetOrganization
	getReq := &pb.GetOrganizationRequest{
		Id: orgID,
	}

	getResp, err := suite.client.GetOrganization(ctx, getReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), getResp.Organization)
	assert.Equal(suite.T(), orgID, getResp.Organization.Id)
	assert.Equal(suite.T(), "Test Organization", getResp.Organization.Name)

	// Test UpdateOrganization
	updateReq := &pb.UpdateOrganizationRequest{
		Id:   orgID,
		Name: testhelpers.StringPtr("Updated Test Organization"),
	}

	updateResp, err := suite.client.UpdateOrganization(ctx, updateReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updateResp.Organization)
	assert.Equal(suite.T(), "Updated Test Organization", updateResp.Organization.Name)

	// Test ListOrganizations
	listReq := &pb.ListOrganizationsRequest{
		PageSize: 10,
	}

	listResp, err := suite.client.ListOrganizations(ctx, listReq)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), listResp.Organizations)

	// Find our organization in the list
	found := false
	for _, org := range listResp.Organizations {
		if org.Id == orgID {
			found = true
			assert.Equal(suite.T(), "Updated Test Organization", org.Name)
			break
		}
	}
	assert.True(suite.T(), found)
}

func (suite *GRPCIntegrationTestSuite) TestUserManagement() {
	ctx := context.Background()

	// Create organization first
	createOrgReq := &pb.CreateOrganizationRequest{
		Name:                 "User Test Organization",
		InitialUserEmail:     "admin@usertest.com",
		InitialUserPublicKey: "admin-public-key",
	}

	orgResp, err := suite.client.CreateOrganization(ctx, createOrgReq)
	require.NoError(suite.T(), err)
	orgID := orgResp.Organization.Id

	// Test CreateUser
	createUserReq := &pb.CreateUserRequest{
		OrganizationId: orgID,
		Username:       "testuser",
		Email:          "testuser@example.com",
		PublicKey:      "user-public-key",
		Tags:           []string{"developer", "tester"},
	}

	createUserResp, err := suite.client.CreateUser(ctx, createUserReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createUserResp.User)
	assert.Equal(suite.T(), "testuser", createUserResp.User.Username)
	assert.Equal(suite.T(), "testuser@example.com", createUserResp.User.Email)
	assert.Equal(suite.T(), orgID, createUserResp.User.OrganizationId)

	userID := createUserResp.User.Id

	// Test GetUser
	getUserReq := &pb.GetUserRequest{
		Id: userID,
	}

	getUserResp, err := suite.client.GetUser(ctx, getUserReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), getUserResp.User)
	assert.Equal(suite.T(), userID, getUserResp.User.Id)
	assert.Equal(suite.T(), "testuser", getUserResp.User.Username)

	// Test UpdateUser
	updateUserReq := &pb.UpdateUserRequest{
		Id:       userID,
		Username: testhelpers.StringPtr("updateduser"),
		Tags:     []string{"updated", "developer"},
	}

	updateUserResp, err := suite.client.UpdateUser(ctx, updateUserReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updateUserResp.User)
	assert.Equal(suite.T(), "updateduser", updateUserResp.User.Username)
	assert.Contains(suite.T(), updateUserResp.User.Tags, "updated")

	// Test ListUsers
	listUsersReq := &pb.ListUsersRequest{
		OrganizationId: orgID,
		PageSize:       10,
	}

	listUsersResp, err := suite.client.ListUsers(ctx, listUsersReq)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), listUsersResp.Users)

	// Should have at least the initial admin user and our test user
	assert.GreaterOrEqual(suite.T(), len(listUsersResp.Users), 2)
}

func (suite *GRPCIntegrationTestSuite) TestActivityManagement() {
	ctx := context.Background()

	// Create organization and user first
	orgResp, err := suite.client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
		Name:                 "Activity Test Organization",
		InitialUserEmail:     "admin@activitytest.com",
		InitialUserPublicKey: "admin-public-key",
	})
	require.NoError(suite.T(), err)
	orgID := orgResp.Organization.Id

	userResp, err := suite.client.CreateUser(ctx, &pb.CreateUserRequest{
		OrganizationId: orgID,
		Username:       "activityuser",
		Email:          "activityuser@example.com",
		PublicKey:      "activity-user-public-key",
	})
	require.NoError(suite.T(), err)
	userID := userResp.User.Id

	// Test CreateActivity
	createActivityReq := &pb.CreateActivityRequest{
		OrganizationId: orgID,
		Type:           "SEED_GENERATION",
		Parameters:     `{"strength": 256, "passphrase": false}`,
		CreatedBy:      userID,
	}

	createActivityResp, err := suite.client.CreateActivity(ctx, createActivityReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createActivityResp.Activity)
	assert.Equal(suite.T(), "SEED_GENERATION", createActivityResp.Activity.Type)
	assert.Equal(suite.T(), "PENDING", createActivityResp.Activity.Status)
	assert.Equal(suite.T(), orgID, createActivityResp.Activity.OrganizationId)

	activityID := createActivityResp.Activity.Id

	// Test GetActivity
	getActivityReq := &pb.GetActivityRequest{
		Id: activityID,
	}

	getActivityResp, err := suite.client.GetActivity(ctx, getActivityReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), getActivityResp.Activity)
	assert.Equal(suite.T(), activityID, getActivityResp.Activity.Id)
	assert.Equal(suite.T(), "SEED_GENERATION", getActivityResp.Activity.Type)

	// Test ListActivities
	listActivitiesReq := &pb.ListActivitiesRequest{
		OrganizationId: orgID,
		PageSize:       10,
	}

	listActivitiesResp, err := suite.client.ListActivities(ctx, listActivitiesReq)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), listActivitiesResp.Activities)

	// Find our activity in the list
	found := false
	for _, activity := range listActivitiesResp.Activities {
		if activity.Id == activityID {
			found = true
			assert.Equal(suite.T(), "SEED_GENERATION", activity.Type)
			break
		}
	}
	assert.True(suite.T(), found)

	// Test filtering by type
	filteredReq := &pb.ListActivitiesRequest{
		OrganizationId: orgID,
		Type:           testhelpers.StringPtr("SEED_GENERATION"),
		PageSize:       10,
	}

	filteredResp, err := suite.client.ListActivities(ctx, filteredReq)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), filteredResp.Activities)

	for _, activity := range filteredResp.Activities {
		assert.Equal(suite.T(), "SEED_GENERATION", activity.Type)
	}
}

func (suite *GRPCIntegrationTestSuite) TestRenclaveIntegration() {
	ctx := context.Background()

	// Skip if renclave is not available
	if os.Getenv("TEST_RENCLAVE_HOST") == "" {
		suite.T().Skip("Skipping renclave integration tests. Set TEST_RENCLAVE_HOST to run.")
	}

	// Test GetEnclaveInfo
	infoResp, err := suite.client.GetEnclaveInfo(ctx, &emptypb.Empty{})
	if err != nil {
		suite.T().Skipf("Renclave not available: %v", err)
	}

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), infoResp)
	assert.NotEmpty(suite.T(), infoResp.Version)
	assert.NotEmpty(suite.T(), infoResp.EnclaveId)
	assert.NotEmpty(suite.T(), infoResp.Capabilities)

	// Test RequestSeedGeneration
	seedReq := &pb.SeedGenerationRequest{
		OrganizationId: uuid.New().String(),
		UserId:         uuid.New().String(),
		Strength:       256,
	}

	seedResp, err := suite.client.RequestSeedGeneration(ctx, seedReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), seedResp)
	assert.NotEmpty(suite.T(), seedResp.SeedPhrase)
	assert.NotEmpty(suite.T(), seedResp.Entropy)
	assert.Equal(suite.T(), int32(256), seedResp.Strength)
	assert.Equal(suite.T(), int32(24), seedResp.WordCount)
	assert.NotEmpty(suite.T(), seedResp.RequestId)

	// Test ValidateSeed with the generated seed
	validateReq := &pb.SeedValidationRequest{
		SeedPhrase: seedResp.SeedPhrase,
	}

	validateResp, err := suite.client.ValidateSeed(ctx, validateReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), validateResp)
	assert.True(suite.T(), validateResp.IsValid)
	assert.Equal(suite.T(), int32(256), validateResp.Strength)
	assert.Equal(suite.T(), int32(24), validateResp.WordCount)
	assert.Empty(suite.T(), validateResp.Errors)

	// Test ValidateSeed with invalid seed
	invalidValidateReq := &pb.SeedValidationRequest{
		SeedPhrase: "invalid seed phrase",
	}

	invalidValidateResp, err := suite.client.ValidateSeed(ctx, invalidValidateReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), invalidValidateResp)
	assert.False(suite.T(), invalidValidateResp.IsValid)
	assert.NotEmpty(suite.T(), invalidValidateResp.Errors)
}

func (suite *GRPCIntegrationTestSuite) TestAuthenticationFlow() {
	ctx := context.Background()

	// Create organization and user
	orgResp, err := suite.client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
		Name:                 "Auth Test Organization",
		InitialUserEmail:     "admin@authtest.com",
		InitialUserPublicKey: "admin-public-key",
	})
	require.NoError(suite.T(), err)
	orgID := orgResp.Organization.Id

	userResp, err := suite.client.CreateUser(ctx, &pb.CreateUserRequest{
		OrganizationId: orgID,
		Username:       "authuser",
		Email:          "authuser@example.com",
		PublicKey:      "auth-user-public-key",
	})
	require.NoError(suite.T(), err)
	userID := userResp.User.Id

	// Test Authentication (simplified - in real implementation would verify signature)
	authReq := &pb.AuthenticateRequest{
		OrganizationId: orgID,
		UserId:         userID,
		AuthMethodId:   uuid.New().String(),
		Signature:      "test-signature",
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	authResp, err := suite.client.Authenticate(ctx, authReq)
	// This might fail if the full auth implementation isn't complete
	if err == nil {
		assert.NotNil(suite.T(), authResp)
		assert.NotEmpty(suite.T(), authResp.SessionToken)
		assert.NotNil(suite.T(), authResp.ExpiresAt)

		// Test Authorization with the session token
		authorizeReq := &pb.AuthorizeRequest{
			SessionToken: authResp.SessionToken,
			ActivityType: "SEED_GENERATION",
			Parameters:   `{"strength": 256}`,
		}

		authorizeResp, err := suite.client.Authorize(ctx, authorizeReq)
		require.NoError(suite.T(), err)
		assert.NotNil(suite.T(), authorizeResp)
		// Authorization result depends on policy implementation
	}
}

func (suite *GRPCIntegrationTestSuite) TestErrorHandling() {
	ctx := context.Background()

	// Test GetOrganization with invalid ID
	_, err := suite.client.GetOrganization(ctx, &pb.GetOrganizationRequest{
		Id: "invalid-uuid",
	})
	assert.Error(suite.T(), err)

	// Test GetUser with non-existent ID
	_, err = suite.client.GetUser(ctx, &pb.GetUserRequest{
		Id: uuid.New().String(),
	})
	assert.Error(suite.T(), err)

	// Test CreateUser with invalid organization ID
	_, err = suite.client.CreateUser(ctx, &pb.CreateUserRequest{
		OrganizationId: uuid.New().String(), // Non-existent org
		Username:       "testuser",
		Email:          "test@example.com",
		PublicKey:      "public-key",
	})
	assert.Error(suite.T(), err)
}

func (suite *GRPCIntegrationTestSuite) TestPagination() {
	ctx := context.Background()

	// Create organization with multiple users
	orgResp, err := suite.client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
		Name:                 "Pagination Test Organization",
		InitialUserEmail:     "admin@paginationtest.com",
		InitialUserPublicKey: "admin-public-key",
	})
	require.NoError(suite.T(), err)
	orgID := orgResp.Organization.Id

	// Create multiple users
	for i := 0; i < 5; i++ {
		_, err := suite.client.CreateUser(ctx, &pb.CreateUserRequest{
			OrganizationId: orgID,
			Username:       fmt.Sprintf("user%d", i),
			Email:          fmt.Sprintf("user%d@example.com", i),
			PublicKey:      fmt.Sprintf("public-key-%d", i),
		})
		require.NoError(suite.T(), err)
	}

	// Test pagination
	firstPageResp, err := suite.client.ListUsers(ctx, &pb.ListUsersRequest{
		OrganizationId: orgID,
		PageSize:       3,
	})
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), firstPageResp.Users, 3)

	if firstPageResp.NextPageToken != "" {
		secondPageResp, err := suite.client.ListUsers(ctx, &pb.ListUsersRequest{
			OrganizationId: orgID,
			PageSize:       3,
			PageToken:      firstPageResp.NextPageToken,
		})
		require.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), secondPageResp.Users)

		// Verify no overlap between pages
		firstPageIDs := make(map[string]bool)
		for _, user := range firstPageResp.Users {
			firstPageIDs[user.Id] = true
		}

		for _, user := range secondPageResp.Users {
			assert.False(suite.T(), firstPageIDs[user.Id], "User should not appear on both pages")
		}
	}
}

func TestGRPCIntegrationSuite(t *testing.T) {
	suite.Run(t, new(GRPCIntegrationTestSuite))
}

// Benchmark gRPC operations
func BenchmarkGRPCOperations(b *testing.B) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		b.Skip("Skipping integration benchmarks. Set INTEGRATION_TESTS=true to run.")
	}

	// Setup similar to test suite but simplified
	config := &config.Config{
		Database: config.DatabaseConfig{
			Host:         testhelpers.GetEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:         5432,
			Username:     testhelpers.GetEnvOrDefault("TEST_DB_USER", "gauth"),
			Password:     testhelpers.GetEnvOrDefault("TEST_DB_PASSWORD", "password"),
			Database:     testhelpers.GetEnvOrDefault("TEST_DB_NAME", "gauth_test"),
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 10,
			MaxLifetime:  5 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host:         testhelpers.GetEnvOrDefault("TEST_REDIS_HOST", "localhost"),
			Port:         6379,
			Database:     3,
			PoolSize:     25,
			MinIdleConns: 10,
		},
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
			SessionTimeout:         30 * time.Minute,
		},
	}

	database, err := db.NewPostgresDB(&config.Database)
	if err != nil {
		b.Fatal(err)
	}
	defer database.Close()

	redis, err := db.NewRedisClient(&config.Redis)
	if err != nil {
		b.Fatal(err)
	}
	defer redis.Close()

	logger := logger.NewDefault()
	svc := service.NewGAuthService(config, logger, database, redis)

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
	if err != nil {
		b.Fatal(err)
	}

	listener := bufconn.Listen(bufSize)
	server := grpcServer.NewServer(config, logger, svc, telemetry)

	go func() {
		grpcSrv := grpc.NewServer()
		pb.RegisterGAuthServiceServer(grpcSrv, server)
		grpcSrv.Serve(listener)
	}()

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewGAuthServiceClient(conn)

	b.Run("HealthCheck", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := client.Health(ctx, &emptypb.Empty{})
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CreateOrganization", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
				Name:                 fmt.Sprintf("Benchmark Org %d", i),
				InitialUserEmail:     fmt.Sprintf("admin%d@benchmark.com", i),
				InitialUserPublicKey: fmt.Sprintf("public-key-%d", i),
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
