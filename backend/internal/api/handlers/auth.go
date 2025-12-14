package handlers

import (
	"net/http"

	"home-run-backend/internal/auth"
	"home-run-backend/internal/config"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: username and password are required",
		})
		return
	}

	// Plain text comparison as per requirements
	if req.Username != h.cfg.Auth.Username || req.Password != h.cfg.Auth.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid credentials",
		})
		return
	}

	if err := auth.SetUser(c, req.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"username": req.Username,
		},
	})
}

// Logout clears the user session
func (h *AuthHandler) Logout(c *gin.Context) {
	if err := auth.ClearUser(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to clear session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// Check verifies if the session is valid
func (h *AuthHandler) Check(c *gin.Context) {
	username := auth.GetUser(c)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"username": username,
		},
	})
}
