package config

import (
	"testing"
)

func TestIsContextAllowed(t *testing.T) {
	// Reset the global options between tests
	defer func() {
		GlobalOptions = &Options{}
	}()

	tests := []struct {
		name            string
		allowedContexts []string
		contextName     string
		expected        bool
	}{
		{
			name:            "all contexts allowed when no restrictions",
			allowedContexts: []string{},
			contextName:     "any-context",
			expected:        true,
		},
		{
			name:            "context explicitly allowed",
			allowedContexts: []string{"context-1", "context-2"},
			contextName:     "context-1",
			expected:        true,
		},
		{
			name:            "context not allowed",
			allowedContexts: []string{"context-1", "context-2"},
			contextName:     "context-3",
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			GlobalOptions.AllowedContexts = tt.allowedContexts

			// Test
			result := IsContextAllowed(tt.contextName)

			// Verify
			if result != tt.expected {
				t.Errorf("IsContextAllowed(%q) = %v, want %v",
					tt.contextName, result, tt.expected)
			}
		})
	}
}
