package kuma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapKumaStatus(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "RUNNING"}, // UP
		{0, "STOPPED"}, // DOWN
		{2, "ERROR"},   // PENDING
		{3, "MAINTENANCE"},
		{99, "ERROR"}, // Unknown value
		{-1, "ERROR"}, // Invalid value
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.input)), func(t *testing.T) {
			result := mapKumaStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
