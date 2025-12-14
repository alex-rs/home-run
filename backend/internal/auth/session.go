package auth

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const UserKey = "user"

// SessionRequired is middleware that requires a valid session
func SessionRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(UserKey)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized - please log in",
			})
			c.Abort()
			return
		}
		c.Set(UserKey, user)
		c.Next()
	}
}

// SetUser sets the user in the session
func SetUser(c *gin.Context, username string) error {
	session := sessions.Default(c)
	session.Set(UserKey, username)
	return session.Save()
}

// ClearUser removes the user from the session
func ClearUser(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}

// GetUser returns the current user from context
func GetUser(c *gin.Context) string {
	if user, exists := c.Get(UserKey); exists {
		if username, ok := user.(string); ok {
			return username
		}
	}
	return ""
}
