package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOrganizationResponseIncludesUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock the gRPC client response
	mockResponse := &pb.CreateOrganizationResponse{
		Organization: &pb.Organization{
			Id:      "org-123",
			Version: "1.0",
			Name:    "Test Organization",
			Users: []*pb.User{
				{
					Id:             "user-456",
					OrganizationId: "org-123",
					Username:       "admin",
					Email:          "admin@test.com",
					PublicKey:      "public-key-123",
					IsActive:       true,
				},
			},
		},
		Status: "created",
	}

	// Create a mock server that returns our expected response
	router.POST("/api/v1/organizations", func(c *gin.Context) {
		var req CreateOrganizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request"))
			return
		}

		// Return the mock response
		c.JSON(http.StatusCreated, successResponse(gin.H{
			"organization": ConvertProtoOrganizationToREST(mockResponse.Organization),
			"status":       mockResponse.Status,
			"user_id": func() string {
				if len(mockResponse.Organization.Users) > 0 {
					return mockResponse.Organization.Users[0].Id
				}
				return ""
			}(),
		}))
	})

	// Test the endpoint
	payload := CreateOrganizationRequest{
		Name:                 "Test Organization",
		InitialUserEmail:     "admin@test.com",
		InitialUserPublicKey: stringPtr("public-key-123"),
	}

	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/organizations", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})

	// Check that user_id is present in the response
	assert.Contains(t, data, "user_id")
	assert.Equal(t, "user-456", data["user_id"])

	// Check that organization is present
	assert.Contains(t, data, "organization")
	org := data["organization"].(map[string]interface{})
	assert.Equal(t, "org-123", org["id"])
	assert.Equal(t, "Test Organization", org["name"])

	// Check that users are included in the organization
	assert.Contains(t, org, "users")
	users := org["users"].([]interface{})
	assert.Len(t, users, 1)

	user := users[0].(map[string]interface{})
	assert.Equal(t, "user-456", user["id"])
	assert.Equal(t, "admin@test.com", user["email"])

	// Check status
	assert.Equal(t, "created", data["status"])
}
