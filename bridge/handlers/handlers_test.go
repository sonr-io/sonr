package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetQueueFromPriority(t *testing.T) {
	tests := []struct {
		priority string
		expected string
	}{
		{"critical", "critical"},
		{"high", "critical"},
		{"low", "low"},
		{"", "default"},
		{"unknown", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			result := GetQueueFromPriority(tt.priority)
			assert.Equal(t, tt.expected, result)
		})
	}
}
