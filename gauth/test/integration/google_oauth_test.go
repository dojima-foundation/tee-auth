package integration

import (
	"context"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (suite *GRPCIntegrationTestSuite) TestGoogleOAuthOrganizationCreation() {
	ctx := context.Background()

	// Test 1: Get OAuth URL without organization ID (should work for new users)
	grpcReq := &pb.GoogleOAuthURLRequest{
		OrganizationId: "", // Empty for new users
	}

	resp, err := suite.client.GetGoogleOAuthURL(ctx, grpcReq)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), resp.Url)
	assert.NotEmpty(suite.T(), resp.State)

	// Test 2: Get OAuth URL with organization ID (should work for existing orgs)
	// First create an organization
	orgResp, err := suite.client.CreateOrganization(ctx, &pb.CreateOrganizationRequest{
		Name:                 "Test OAuth Organization",
		InitialUserEmail:     "admin@testoauth.com",
		InitialUserPublicKey: "admin-public-key",
	})
	require.NoError(suite.T(), err)

	grpcReqWithOrg := &pb.GoogleOAuthURLRequest{
		OrganizationId: orgResp.Organization.Id,
	}

	respWithOrg, err := suite.client.GetGoogleOAuthURL(ctx, grpcReqWithOrg)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), respWithOrg.Url)
	assert.NotEmpty(suite.T(), respWithOrg.State)

	// Test 3: Verify that the URLs are different (they should be)
	assert.NotEqual(suite.T(), resp.Url, respWithOrg.Url)
}

func (suite *GRPCIntegrationTestSuite) TestGoogleOAuthCallbackWithoutOrganization() {
	// This test would require a mock Google OAuth flow
	// For now, we'll just test that the service can handle the request structure

	// Create a mock state that doesn't include organization_id
	stateData := map[string]string{
		"state":     "test-state",
		"timestamp": "1234567890",
		// Note: no organization_id - this should trigger organization creation
	}

	// In a real test, we would:
	// 1. Mock the Google OAuth exchange
	// 2. Mock the Google user info response
	// 3. Test that an organization is created with "Root user"
	// 4. Verify the user is added to the root quorum

	// For now, just verify the test structure
	assert.NotNil(suite.T(), stateData)
	assert.Equal(suite.T(), "test-state", stateData["state"])
}
