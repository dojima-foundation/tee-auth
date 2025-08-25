package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		GRPC: config.GRPCConfig{
			Host:              "localhost",
			Port:              9090,
			ConnectionTimeout: 10 * time.Second,
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

	lgr := logger.NewDefault()
	server := NewServer(cfg, lgr)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.Equal(t, lgr, server.logger)
}

func TestMiddleware(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			CORSEnabled:      true,
			CORSOrigins:      []string{"http://localhost:3000", "https://app.example.com"},
			RateLimitEnabled: false,
		},
		Logging: config.LoggingConfig{
			Level: "debug",
		},
	}

	lgr := logger.NewDefault()
	server := NewServer(cfg, lgr)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	t.Run("CORS Middleware", func(t *testing.T) {
		router.Use(server.corsMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})

		// Test allowed origin
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))

		// Test disallowed origin
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.Header.Set("Origin", "http://malicious.com")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Empty(t, w2.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS Request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestErrorResponse(t *testing.T) {
	err := assert.AnError
	message := "Test error"

	response := errorResponse(err, message)

	expected := gin.H{
		"error":   message,
		"details": err.Error(),
	}

	assert.Equal(t, expected, response)
}

func TestSuccessResponse(t *testing.T) {
	data := map[string]interface{}{
		"id":   "123",
		"name": "test",
	}

	response := successResponse(data)

	expected := gin.H{
		"success": true,
		"data":    data,
	}

	assert.Equal(t, expected, response)
}

func TestJSONBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", func(c *gin.Context) {
		var req CreateOrganizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse(err, "Invalid request"))
			return
		}
		c.JSON(http.StatusOK, successResponse(req))
	})

	t.Run("Valid JSON", func(t *testing.T) {
		payload := CreateOrganizationRequest{
			Name:                 "Test Org",
			InitialUserEmail:     "admin@test.com",
			InitialUserPublicKey: stringPtr("public-key-123"),
		}

		jsonData, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Test Org", data["name"])
		assert.Equal(t, "admin@test.com", data["initial_user_email"])
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Invalid request", response["error"])
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Test Org",
			// Missing required fields
		}

		jsonData, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
