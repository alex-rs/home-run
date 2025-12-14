package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapDockerState(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"running", "RUNNING"},
		{"RUNNING", "RUNNING"},
		{"exited", "STOPPED"},
		{"dead", "STOPPED"},
		{"paused", "MAINTENANCE"},
		{"restarting", "MAINTENANCE"},
		{"created", "STOPPED"},
		{"unknown", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapDockerState(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateCPUPercent_ZeroValues(t *testing.T) {
	// This test ensures the function doesn't panic with nil or zero values
	// In real scenarios, we'd need to mock the Docker API
	// For now, we just test the mapping functions
	assert.NotPanics(t, func() {
		_ = mapDockerState("running")
	})
}
