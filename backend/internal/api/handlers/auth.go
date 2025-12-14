package handlers

import (
	"net/http"

	"home-run-backend/internal/auth"
	"home-run-backend/internal/config"
	"home-run-backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		}).Warn("Invalid login request")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: username and password are required",
		})
		return
	}

	// Plain text comparison as per requirements
	if req.Username != h.cfg.Auth.Username || req.Password != h.cfg.Auth.Password {
		logger.WithFields(logrus.Fields{
			"username": req.Username,
			"ip":       c.ClientIP(),
		}).Warn("Failed login attempt")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid credentials",
		})
		return
	}

	if err := auth.SetUser(c, req.Username); err != nil {
		logger.WithFields(logrus.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Error("Failed to create session")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create session",
		})
		return
	}

	logger.WithFields(logrus.Fields{
		"username": req.Username,
		"ip":       c.ClientIP(),
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"username": req.Username,
		},
	})
}

// Logout clears the user session
func (h *AuthHandler) Logout(c *gin.Context) {
	username := auth.GetUser(c)
	if err := auth.ClearUser(c); err != nil {
		logger.WithFields(logrus.Fields{
			"username": username,
			"error":    err.Error(),
		}).Error("Failed to clear session")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to clear session",
		})
		return
	}

	logger.WithFields(logrus.Fields{
		"username": username,
		"ip":       c.ClientIP(),
	}).Info("User logged out")

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
