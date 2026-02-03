package cmd

import (
	"testing"
)

func TestValidateDescriptionFlag(t *testing.T) {
	// Save original flag values
	originalShowDescription := showDescription
	originalShowStatus := showStatus
	originalToolFilter := toolFilter
	originalAllTools := allTools

	// Restore original values after test
	defer func() {
		showDescription = originalShowDescription
		showStatus = originalShowStatus
		toolFilter = originalToolFilter
		allTools = originalAllTools
	}()

	tests := []struct {
		name            string
		showDescription bool
		showStatus      bool
		toolFilter      string
		allTools        bool
		expectError     bool
	}{
		{
			name:            "description only",
			showDescription: true,
			showStatus:      false,
			toolFilter:      "",
			allTools:        false,
			expectError:     false,
		},
		{
			name:            "description with status",
			showDescription: true,
			showStatus:      true,
			toolFilter:      "",
			allTools:        false,
			expectError:     true,
		},
		{
			name:            "description with tool filter",
			showDescription: true,
			showStatus:      false,
			toolFilter:      "kiro",
			allTools:        false,
			expectError:     true,
		},
		{
			name:            "description with all tools",
			showDescription: true,
			showStatus:      false,
			toolFilter:      "",
			allTools:        true,
			expectError:     true,
		},
		{
			name:            "no description flag",
			showDescription: false,
			showStatus:      true,
			toolFilter:      "kiro",
			allTools:        true,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flag values for this test
			showDescription = tt.showDescription
			showStatus = tt.showStatus
			toolFilter = tt.toolFilter
			allTools = tt.allTools

			err := validateDescriptionFlag()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with spaces",
			input:    "hello world",
			expected: "\"hello world\"",
		},
		{
			name:     "string with tab",
			input:    "hello\tworld",
			expected: "\"hello\tworld\"",
		},
		{
			name:     "string with newline",
			input:    "hello\nworld",
			expected: "\"hello\nworld\"",
		},
		{
			name:     "string with double quotes",
			input:    "hello \"world\"",
			expected: "\"hello \\\"world\\\"\"",
		},
		{
			name:     "string with single quotes",
			input:    "hello 'world'",
			expected: "\"hello 'world'\"",
		},
		{
			name:     "string with backslash",
			input:    "hello\\world",
			expected: "\"hello\\\\world\"",
		},
		{
			name:     "string with backtick",
			input:    "hello`world",
			expected: "\"hello\\`world\"",
		},
		{
			name:     "string with exclamation",
			input:    "hello!world",
			expected: "\"hello!world\"",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with dollar sign",
			input:    "hello$world",
			expected: "hello$world",
		},
		{
			name:     "complex string",
			input:    "hello \"world\" with\\backslash and `backtick`",
			expected: "\"hello \\\"world\\\" with\\\\backslash and \\`backtick\\`\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
