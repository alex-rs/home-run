package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenRequired is middleware that requires a valid API token for federation
func TokenRequired(validToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing Authorization header",
			})
			c.Abort()
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid Authorization format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		if parts[1] != validToken {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Invalid API token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
