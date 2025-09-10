package rest

import (
	"net/http"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		s.logger.Info("HTTP request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)
		return ""
	})
}

// recoveryMiddleware handles panics
func (s *Server) recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		s.logger.Error("HTTP request panic", "error", recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
}

// httpMetricsMiddleware records HTTP request metrics
func (s *Server) httpMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process the request
		c.Next()

		// Record metrics
		duration := time.Since(start)
		telemetry.RecordHTTPRequest(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}

// corsMiddleware handles CORS headers for cross-domain sessions
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		var allowedOrigin string

		for _, allowedOrig := range s.config.Security.CORSOrigins {
			if allowedOrig == "*" {
				allowed = true
				allowedOrigin = origin
				break
			} else if allowedOrig == origin {
				allowed = true
				allowedOrigin = origin
				break
			}
		}

		// Set CORS headers
		if allowed {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Expose-Headers", "Set-Cookie, X-Session-Token")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// rateLimitMiddleware implements rate limiting
func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(s.config.Security.RateLimitRPS), s.config.Security.RateLimitBurst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			s.logger.Warn("Rate limit exceeded", "ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// errorResponse creates a standardized error response
func errorResponse(err error, message string) gin.H {
	response := gin.H{
		"error": message,
	}

	if err != nil {
		response["details"] = err.Error()
	}

	return response
}

// successResponse creates a standardized success response
func successResponse(data interface{}) gin.H {
	return gin.H{
		"success": true,
		"data":    data,
	}
}
