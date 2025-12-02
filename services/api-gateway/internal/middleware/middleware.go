package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware logs incoming requests
func LoggingMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		logger.Infof("%s %s | Status: %d | Duration: %v", method, path, statusCode, duration)
	}
}

// CORSMiddleware enables CORS
func CORSMiddleware(allowOrigins string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Admin-Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware (basic implementation - can be enhanced with actual rate limiting library)
func RateLimitMiddleware(requestsPerSecond int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Placeholder for rate limiting logic
		// In production, use a library like "go-baku/rate"
		c.Next()
	}
}

// AuthenticationMiddleware checks for authorization headers on protected routes
func AuthenticationMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for public routes
		publicPaths := []string{"/health", "/api/v1/products", "/api/v1/auth/login"}
		for _, path := range publicPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// Check for token in header or skip if not required
		token := c.GetHeader("Authorization")
		if token == "" {
			// For now, continue without token (can be made stricter)
			logger.Debugf("No authorization token provided for %s", c.Request.URL.Path)
		}

		c.Next()
	}
}

// TraefikLabelMiddleware sets Traefik-specific headers (informational, optional)
func TraefikLabelMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add custom headers that may be useful for monitoring
		c.Header("X-Gateway-Version", "1.0.0")
		c.Header("X-Gateway-Service", "Permia-API-Gateway")
		c.Next()
	}
}
