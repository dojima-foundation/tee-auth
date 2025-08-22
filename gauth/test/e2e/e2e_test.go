package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type E2ETestSuite struct {
	suite.Suite
	client          pb.GAuthServiceClient
	conn            *grpc.ClientConn
	serverProcess   *exec.Cmd
	renclaveProcess *exec.Cmd
	config          E2EConfig
}

type E2EConfig struct {
	GAuthHost    string
	GAuthPort    string
	RenclaveHost string
	RenclavePort string
	DatabaseURL  string
	RedisURL     string
	TestTimeout  time.Duration
	StartupDelay time.Duration
}

func (suite *E2ETestSuite) SetupSuite() {
	// Skip E2E tests if not in E2E test mode
	if os.Getenv("E2E_TESTS") != "true" {
		suite.T().Skip("Skipping E2E tests. Set E2E_TESTS=true to run.")
	}

	// // Skip if external services are not available and not starting them
	// if os.Getenv("E2E_START_SERVICES") != "true" {
	// 	suite.T().Skip("Skipping external service E2E tests. Set E2E_START_SERVICES=true to run with external services.")
	// }

	suite.config = E2EConfig{
		GAuthHost:    getEnvOrDefault("E2E_GAUTH_HOST", "localhost"),
		GAuthPort:    getEnvOrDefault("E2E_GAUTH_PORT", "9091"),
		RenclaveHost: getEnvOrDefault("E2E_RENCLAVE_HOST", "localhost"),
		RenclavePort: getEnvOrDefault("E2E_RENCLAVE_PORT", "3000"),
		DatabaseURL:  getEnvOrDefault("E2E_DATABASE_URL", "postgres://gauth:password@localhost:5432/gauth_e2e?sslmode=disable"),
		RedisURL:     getEnvOrDefault("E2E_REDIS_URL", "redis://localhost:6379/4"),
		TestTimeout:  60 * time.Second,
		StartupDelay: 10 * time.Second,
	}

	// Start dependencies if needed
	if os.Getenv("E2E_START_SERVICES") == "true" {
		suite.startServices()
	}

	// Connect to gauth service
	suite.connectToGAuth()

	// Wait for services to be ready
	suite.waitForServices()
}

func (suite *E2ETestSuite) TearDownSuite() {
	if suite.conn != nil {
		suite.conn.Close()
	}

	// Stop services if we started them
	if os.Getenv("E2E_START_SERVICES") == "true" {
		suite.stopServices()
	}
}

func (suite *E2ETestSuite) startServices() {
	suite.T().Log("Starting services for E2E tests...")

	// Start renclave-v2 service
	if suite.config.RenclaveHost == "localhost" {
		renclaveCmd := exec.Command("docker", "run", "--rm", "-d",
			"--name", "gauth-e2e-renclave",
			"-p", fmt.Sprintf("%s:3000", suite.config.RenclavePort),
			"renclave-v2:latest")

		err := renclaveCmd.Run()
		if err != nil {
			suite.T().Logf("Warning: Could not start renclave container: %v", err)
		} else {
			suite.T().Log("Started renclave-v2 container")
		}
	}

	// Start gauth service
	gauthBinary := getEnvOrDefault("E2E_GAUTH_BINARY", "./bin/gauth")

	suite.serverProcess = exec.Command(gauthBinary)
	suite.serverProcess.Env = append(os.Environ(),
		"GRPC_PORT="+suite.config.GAuthPort,
		"SERVER_PORT=8082",
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_USERNAME=gauth",
		"DB_PASSWORD=password",
		"DB_DATABASE=gauth_e2e",
		"REDIS_HOST=localhost",
		"REDIS_PORT=6379",
		"REDIS_DATABASE=4",
		"RENCLAVE_HOST="+suite.config.RenclaveHost,
		"RENCLAVE_PORT="+suite.config.RenclavePort,
		"LOG_LEVEL=info",
		"JWT_SECRET=e2e-test-secret-key-32-bytes-long",
		"ENCRYPTION_KEY=e2e-test-encryption-key-32-bytes",
	)

	err := suite.serverProcess.Start()
	require.NoError(suite.T(), err, "Failed to start gauth server")

	suite.T().Log("Started gauth server")

	// Wait for startup
	time.Sleep(suite.config.StartupDelay)
}

func (suite *E2ETestSuite) stopServices() {
	suite.T().Log("Stopping services...")

	if suite.serverProcess != nil {
		suite.serverProcess.Process.Kill()
		suite.serverProcess.Wait()
		suite.T().Log("Stopped gauth server")
	}

	// Stop renclave container if we started it
	if suite.config.RenclaveHost == "localhost" {
		stopCmd := exec.Command("docker", "stop", "gauth-e2e-renclave")
		stopCmd.Run() // Ignore errors
		suite.T().Log("Stopped renclave-v2 container")
	}
}

func (suite *E2ETestSuite) connectToGAuth() {
	address := fmt.Sprintf("%s:%s", suite.config.GAuthHost, suite.config.GAuthPort)

	// Don't use WithBlock() to avoid hanging if service isn't available
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Skip external service tests unless explicitly requested
	if os.Getenv("E2E_START_SERVICES") != "true" {
		suite.T().Skip("Skipping external service E2E tests: E2E_START_SERVICES is not set to true")
		return
	}

	// Use a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		suite.T().Fatalf("Failed to establish connection to gRPC server: %v", err)
		return
	}

	// Test the connection with a quick health check
	client := pb.NewGAuthServiceClient(conn)
	healthCtx, healthCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer healthCancel()
	
	_, err = client.Health(healthCtx, &emptypb.Empty{})
	if err != nil {
		conn.Close()
		suite.T().Fatalf("Service is not responding to health checks: %v", err)
		return
	}

	suite.conn = conn
	suite.client = client
	suite.T().Logf("Connected to gauth service at %s", address)
}

func (suite *E2ETestSuite) waitForServices() {
	// We've already done a health check in connectToGAuth
	// This is just a placeholder in case we need additional setup
	suite.T().Log("Services are ready")
}

func (suite *E2ETestSuite) TestCompleteWorkflow() {
	ctx := context.Background()

	suite.T().Log("=== Testing Complete E2E Workflow ===")

	// Step 1: Check service health
	suite.T().Log("Step 1: Checking service health")
	healthResp, err := suite.client.Health(ctx, &emptypb.Empty{})
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), []string{"healthy", "degraded"}, healthResp.Status)
	suite.T().Logf("Health status: %s", healthResp.Status)

	// Step 2: Create organization
	suite.T().Log("Step 2: Creating organization")
	orgResp, err := suite.client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
		Name:                 "E2E Test Organization",
		InitialUserEmail:     "admin@e2etest.com",
		InitialUserPublicKey: "e2e-admin-public-key",
	})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), orgResp.Organization)
	orgID := orgResp.Organization.Id
	suite.T().Logf("Created organization: %s", orgID)

	// Step 3: Create additional user
	suite.T().Log("Step 3: Creating additional user")
	userResp, err := suite.client.CreateUser(ctx, &pb.CreateUserRequest{
		OrganizationId: orgID,
		Username:       "e2euser",
		Email:          "e2euser@example.com",
		PublicKey:      "e2e-user-public-key",
		Tags:           []string{"tester", "e2e"},
	})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), userResp.User)
	userID := userResp.User.Id
	suite.T().Logf("Created user: %s", userID)

	// Step 4: Create activity
	suite.T().Log("Step 4: Creating activity")
	activityResp, err := suite.client.CreateActivity(ctx, &pb.CreateActivityRequest{
		OrganizationId: orgID,
		Type:           "SEED_GENERATION",
		Parameters:     `{"strength": 256, "passphrase": false}`,
		CreatedBy:      userID,
	})
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), activityResp.Activity)
	activityID := activityResp.Activity.Id
	suite.T().Logf("Created activity: %s", activityID)

	// Step 5: Test seed generation (if renclave is available)
	suite.T().Log("Step 5: Testing seed generation")
	seedResp, err := suite.client.RequestSeedGeneration(ctx, &pb.SeedGenerationRequest{
		OrganizationId: orgID,
		UserId:         userID,
		Strength:       256,
	})

	if err != nil {
		suite.T().Logf("Seed generation failed (renclave may not be available): %v", err)
	} else {
		assert.NotNil(suite.T(), seedResp)
		assert.NotEmpty(suite.T(), seedResp.SeedPhrase)
		assert.Equal(suite.T(), int32(256), seedResp.Strength)
		suite.T().Logf("Generated seed phrase: %s", seedResp.SeedPhrase[:50]+"...")

		// Step 6: Validate the generated seed
		suite.T().Log("Step 6: Validating generated seed")
		validateResp, err := suite.client.ValidateSeed(ctx, &pb.SeedValidationRequest{
			SeedPhrase: seedResp.SeedPhrase,
		})
		require.NoError(suite.T(), err)
		assert.True(suite.T(), validateResp.IsValid)
		assert.Empty(suite.T(), validateResp.Errors)
		suite.T().Log("Seed validation successful")
	}

	// Step 7: List and verify data
	suite.T().Log("Step 7: Verifying created data")

	// List organizations
	orgsResp, err := suite.client.ListOrganizations(ctx, &pb.ListOrganizationsRequest{
		PageSize: 10,
	})
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), orgsResp.Organizations)
	suite.T().Logf("Found %d organizations", len(orgsResp.Organizations))

	// List users
	usersResp, err := suite.client.ListUsers(ctx, &pb.ListUsersRequest{
		OrganizationId: orgID,
		PageSize:       10,
	})
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(usersResp.Users), 2) // Admin + our user
	suite.T().Logf("Found %d users in organization", len(usersResp.Users))

	// List activities
	activitiesResp, err := suite.client.ListActivities(ctx, &pb.ListActivitiesRequest{
		OrganizationId: orgID,
		PageSize:       10,
	})
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), activitiesResp.Activities)
	suite.T().Logf("Found %d activities", len(activitiesResp.Activities))

	suite.T().Log("=== E2E Workflow Completed Successfully ===")
}

func (suite *E2ETestSuite) TestConcurrentOperations() {
	ctx := context.Background()
	suite.T().Log("=== Testing Concurrent Operations ===")

	// Create organization for concurrent tests
	orgResp, err := suite.client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
		Name:                 "Concurrent Test Organization",
		InitialUserEmail:     "admin@concurrent.com",
		InitialUserPublicKey: "concurrent-admin-key",
	})
	require.NoError(suite.T(), err)
	orgID := orgResp.Organization.Id

	// Test concurrent user creation
	concurrency := 5
	userChan := make(chan *pb.CreateUserResponse, concurrency)
	errorChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			userResp, err := suite.client.CreateUser(ctx, &pb.CreateUserRequest{
				OrganizationId: orgID,
				Username:       fmt.Sprintf("concurrent-user-%d", index),
				Email:          fmt.Sprintf("user%d@concurrent.com", index),
				PublicKey:      fmt.Sprintf("public-key-%d", index),
			})
			if err != nil {
				errorChan <- err
			} else {
				userChan <- userResp
			}
		}(i)
	}

	// Collect results
	successCount := 0
	errorCount := 0
	for i := 0; i < concurrency; i++ {
		select {
		case <-userChan:
			successCount++
		case err := <-errorChan:
			errorCount++
			suite.T().Logf("Concurrent user creation error: %v", err)
		case <-time.After(10 * time.Second):
			suite.T().Fatal("Timeout waiting for concurrent operations")
		}
	}

	suite.T().Logf("Concurrent operations: %d successful, %d failed", successCount, errorCount)
	assert.Greater(suite.T(), successCount, 0, "At least some concurrent operations should succeed")

	// Verify all users were created
	usersResp, err := suite.client.ListUsers(ctx, &pb.ListUsersRequest{
		OrganizationId: orgID,
		PageSize:       20,
	})
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(usersResp.Users), successCount+1) // +1 for admin user
}

func (suite *E2ETestSuite) TestErrorScenarios() {
	ctx := context.Background()
	suite.T().Log("=== Testing Error Scenarios ===")

	// Test invalid UUID
	_, err := suite.client.GetOrganization(ctx, &pb.GetOrganizationRequest{
		Id: "invalid-uuid",
	})
	assert.Error(suite.T(), err)
	suite.T().Log("✓ Invalid UUID handling works")

	// Test non-existent resource
	_, err = suite.client.GetUser(ctx, &pb.GetUserRequest{
		Id: uuid.New().String(),
	})
	assert.Error(suite.T(), err)
	suite.T().Log("✓ Non-existent resource handling works")

	// Test invalid seed phrase validation
	validateResp, err := suite.client.ValidateSeed(ctx, &pb.SeedValidationRequest{
		SeedPhrase: "invalid seed phrase with wrong words",
	})

	if err == nil {
		// If renclave is available, it should return validation errors
		assert.False(suite.T(), validateResp.IsValid)
		assert.NotEmpty(suite.T(), validateResp.Errors)
		suite.T().Log("✓ Invalid seed validation works")
	} else {
		suite.T().Logf("Seed validation unavailable (renclave not running): %v", err)
	}

	// Test duplicate user creation
	orgResp, err := suite.client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
		Name:                 "Error Test Organization",
		InitialUserEmail:     "admin@errortest.com",
		InitialUserPublicKey: "error-admin-key",
	})
	require.NoError(suite.T(), err)
	orgID := orgResp.Organization.Id

	// Create first user
	_, err = suite.client.CreateUser(ctx, &pb.CreateUserRequest{
		OrganizationId: orgID,
		Username:       "duplicateuser",
		Email:          "duplicate@example.com",
		PublicKey:      "duplicate-key",
	})
	require.NoError(suite.T(), err)

	// Try to create duplicate user
	_, err = suite.client.CreateUser(ctx, &pb.CreateUserRequest{
		OrganizationId: orgID,
		Username:       "duplicateuser", // Same username
		Email:          "different@example.com",
		PublicKey:      "different-key",
	})
	assert.Error(suite.T(), err)
	suite.T().Log("✓ Duplicate user prevention works")
}

func (suite *E2ETestSuite) TestPerformance() {
	ctx := context.Background()
	suite.T().Log("=== Testing Performance ===")

	// Create organization for performance tests
	orgResp, err := suite.client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
		Name:                 "Performance Test Organization",
		InitialUserEmail:     "admin@perftest.com",
		InitialUserPublicKey: "perf-admin-key",
	})
	require.NoError(suite.T(), err)
	orgID := orgResp.Organization.Id

	// Test health check performance
	start := time.Now()
	healthChecks := 10
	for i := 0; i < healthChecks; i++ {
		_, err := suite.client.Health(ctx, &emptypb.Empty{})
		require.NoError(suite.T(), err)
	}
	healthDuration := time.Since(start)
	avgHealthTime := healthDuration / time.Duration(healthChecks)
	suite.T().Logf("Average health check time: %v", avgHealthTime)
	assert.Less(suite.T(), avgHealthTime, 100*time.Millisecond, "Health checks should be fast")

	// Test user creation performance
	start = time.Now()
	userCreations := 5
	for i := 0; i < userCreations; i++ {
		_, err := suite.client.CreateUser(ctx, &pb.CreateUserRequest{
			OrganizationId: orgID,
			Username:       fmt.Sprintf("perfuser-%d", i),
			Email:          fmt.Sprintf("perfuser%d@example.com", i),
			PublicKey:      fmt.Sprintf("perf-key-%d", i),
		})
		require.NoError(suite.T(), err)
	}
	userCreationDuration := time.Since(start)
	avgUserCreationTime := userCreationDuration / time.Duration(userCreations)
	suite.T().Logf("Average user creation time: %v", avgUserCreationTime)
	assert.Less(suite.T(), avgUserCreationTime, 500*time.Millisecond, "User creation should be reasonably fast")

	// Test list operations performance
	start = time.Now()
	_, err = suite.client.ListUsers(ctx, &pb.ListUsersRequest{
		OrganizationId: orgID,
		PageSize:       100,
	})
	require.NoError(suite.T(), err)
	listDuration := time.Since(start)
	suite.T().Logf("List users time: %v", listDuration)
	assert.Less(suite.T(), listDuration, 200*time.Millisecond, "List operations should be fast")
}

func (suite *E2ETestSuite) TestDataConsistency() {
	ctx := context.Background()
	suite.T().Log("=== Testing Data Consistency ===")

	// Create organization
	orgResp, err := suite.client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
		Name:                 "Consistency Test Organization",
		InitialUserEmail:     "admin@consistency.com",
		InitialUserPublicKey: "consistency-admin-key",
	})
	require.NoError(suite.T(), err)
	orgID := orgResp.Organization.Id

	// Create user
	userResp, err := suite.client.CreateUser(ctx, &pb.CreateUserRequest{
		OrganizationId: orgID,
		Username:       "consistencyuser",
		Email:          "consistency@example.com",
		PublicKey:      "consistency-key",
	})
	require.NoError(suite.T(), err)
	userID := userResp.User.Id

	// Update user
	updateResp, err := suite.client.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:       userID,
		Username: stringPtr("updated-consistency-user"),
		Tags:     []string{"updated", "consistent"},
	})
	require.NoError(suite.T(), err)

	// Verify update consistency
	getResp, err := suite.client.GetUser(ctx, &pb.GetUserRequest{
		Id: userID,
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), updateResp.User.Username, getResp.User.Username)
	assert.Equal(suite.T(), updateResp.User.Tags, getResp.User.Tags)
	suite.T().Log("✓ Data consistency verified")

	// Test activity-user relationship consistency
	activityResp, err := suite.client.CreateActivity(ctx, &pb.CreateActivityRequest{
		OrganizationId: orgID,
		Type:           "CONSISTENCY_TEST",
		Parameters:     `{"test": "consistency"}`,
		CreatedBy:      userID,
	})
	require.NoError(suite.T(), err)

	// Verify activity references correct user
	getActivityResp, err := suite.client.GetActivity(ctx, &pb.GetActivityRequest{
		Id: activityResp.Activity.Id,
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), userID, getActivityResp.Activity.CreatedBy)
	assert.Equal(suite.T(), orgID, getActivityResp.Activity.OrganizationId)
	suite.T().Log("✓ Relationship consistency verified")
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

// Helper function
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func stringPtr(s string) *string {
	return &s
}
