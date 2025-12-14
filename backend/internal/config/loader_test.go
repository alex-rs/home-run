package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidConfig(t *testing.T) {
	yamlContent := `
server:
  port: 8080
  cors_allow_origin: "*"
  session_secret: "test-secret-32-characters-long"

auth:
  username: "admin"
  password: "password"
  api_token: "test-token"

services:
  - name: "Test Service"
    backend: "docker"
    container_name: "test-container"
    url: "http://localhost:8080"
    port: 8080
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "admin", cfg.Auth.Username)
	assert.Equal(t, "password", cfg.Auth.Password)
	assert.Len(t, cfg.Services, 1)
	assert.Equal(t, "Test Service", cfg.Services[0].Name)
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/config.yml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestValidate_MissingUsername(t *testing.T) {
	cfg := &Config{
		Auth: AuthConfig{
			Password: "password",
			APIToken: "token",
		},
	}

	err := validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username is required")
}

func TestValidate_MissingPassword(t *testing.T) {
	cfg := &Config{
		Auth: AuthConfig{
			Username: "admin",
			APIToken: "token",
		},
	}

	err := validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password is required")
}

func TestValidate_MissingAPIToken(t *testing.T) {
	cfg := &Config{
		Auth: AuthConfig{
			Username: "admin",
			Password: "password",
		},
	}

	err := validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_token is required")
}

func TestValidate_InvalidBackend(t *testing.T) {
	cfg := &Config{
		Auth: AuthConfig{
			Username: "admin",
			Password: "password",
			APIToken: "token",
		},
		Services: []ServiceConfig{
			{
				Name:    "Test",
				Backend: "invalid",
			},
		},
	}

	err := validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend must be")
}

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.NotEmpty(t, cfg.Server.SessionSecret)
	assert.Equal(t, "*", cfg.Server.CORSAllowOrigin)
}
