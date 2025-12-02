package middleware

import (
	"os"

	"Permia/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminAuth creates middleware for admin authentication.
// It checks the X-Admin-Token header against the ADMIN_SECRET_TOKEN environment variable.
func AdminAuth(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the admin token from the request header
		token := c.GetHeader("X-Admin-Token")

		// Get the expected token from environment
		expectedToken := os.Getenv("ADMIN_SECRET_TOKEN")

		// Log the authentication attempt (without logging the actual tokens for security)
		logger.Debug("Admin authentication attempt",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		// Validate token exists and is correct
		if token == "" {
			logger.Warn("Admin authentication failed: missing token",
				zap.String("path", c.Request.URL.Path),
				zap.String("remote_ip", c.ClientIP()),
			)
			response.Error(c, 401, "Missing X-Admin-Token header")
			c.Abort()
			return
		}

		if expectedToken == "" {
			logger.Error("ADMIN_SECRET_TOKEN environment variable not configured")
			response.Error(c, 500, "Server configuration error")
			c.Abort()
			return
		}

		if token != expectedToken {
			logger.Warn("Admin authentication failed: invalid token",
				zap.String("path", c.Request.URL.Path),
				zap.String("remote_ip", c.ClientIP()),
			)
			response.Error(c, 403, "Invalid admin token")
			c.Abort()
			return
		}

		// Authentication successful
		logger.Debug("Admin authentication successful",
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_ip", c.ClientIP()),
		)

		// Store admin context (optional, for logging purposes)
		c.Set("is_admin", true)

		c.Next()
	}
}
