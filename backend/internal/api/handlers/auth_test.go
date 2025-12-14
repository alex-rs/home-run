package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"home-run-backend/internal/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup session middleware
	store := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("test_session", store))

	return r
}

func TestAuthHandler_Login_Success(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Username: "testuser",
			Password: "testpass",
		},
	}

	handler := NewAuthHandler(cfg)
	router := setupTestRouter(cfg)
	router.POST("/login", handler.Login)

	loginReq := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["user"])
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Username: "testuser",
			Password: "testpass",
		},
	}

	handler := NewAuthHandler(cfg)
	router := setupTestRouter(cfg)
	router.POST("/login", handler.Login)

	loginReq := LoginRequest{
		Username: "wronguser",
		Password: "wrongpass",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.Equal(t, "Invalid credentials", response["error"])
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Username: "testuser",
			Password: "testpass",
		},
	}

	handler := NewAuthHandler(cfg)
	router := setupTestRouter(cfg)
	router.POST("/login", handler.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Logout(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Username: "testuser",
			Password: "testpass",
		},
	}

	handler := NewAuthHandler(cfg)
	router := setupTestRouter(cfg)
	router.POST("/logout", handler.Logout)

	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
}

func TestAuthHandler_Check(t *testing.T) {
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Username: "testuser",
			Password: "testpass",
		},
	}

	handler := NewAuthHandler(cfg)
	router := setupTestRouter(cfg)
	router.GET("/check", handler.Check)

	req := httptest.NewRequest("GET", "/check", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
}
