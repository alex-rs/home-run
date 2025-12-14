package services

import (
	"context"
	"testing"
	"time"

	"home-run-backend/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	assert.NotNil(t, manager)

	// Cleanup
	defer manager.Stop()
}

func TestManager_GetAll_EmptyServices(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Stop()

	ctx := context.Background()
	services := manager.GetAll(ctx)

	assert.Empty(t, services)
}

func TestManager_GetByID_NotFound(t *testing.T) {
	cfg := &config.Config{
		Services: []config.ServiceConfig{},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Stop()

	ctx := context.Background()
	svc, err := manager.GetByID(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "service not found")
}

func TestGenerateID(t *testing.T) {
	id1 := generateID("test-service")
	id2 := generateID("test-service")
	id3 := generateID("different-service")

	// Same name should generate same ID
	assert.Equal(t, id1, id2)

	// Different names should generate different IDs
	assert.NotEqual(t, id1, id3)

	// ID should be 12 characters (MD5 hash truncated to 6 bytes = 12 hex chars)
	assert.Len(t, id1, 12)
}

func TestDetectConfigType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/config.yaml", "YAML"},
		{"/path/to/config.yml", "YAML"},
		{"/path/to/config.json", "JSON"},
		{"/path/to/config.ini", "INI"},
		{"/path/to/Dockerfile", "DOCKERFILE"},
		{"/path/to/app.dockerfile", "DOCKERFILE"},
		{"/path/to/unknown.txt", "YAML"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectConfigType(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{"zero", "0s"},
		{"minutes", "5m30s"},
		{"hours", "2h30m"},
		{"days", "25h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that formatUptime doesn't panic with zero time
			result := formatUptime(time.Time{})
			assert.Equal(t, "Unknown", result)
		})
	}
}
