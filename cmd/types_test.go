package cmd

import (
	"testing"
)

// TestGetDescription tests the GetDescription function
func TestGetDescription(t *testing.T) {
	tests := []struct {
		name     string
		service  Service
		expected string
	}{
		{
			name: "service with description label",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.profile":     "default",
					"mcp.description": "A time server for MCP",
				},
			},
			expected: "A time server for MCP",
		},
		{
			name: "service without description label",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.profile": "default",
				},
			},
			expected: "",
		},
		{
			name: "service with nil labels",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels:  nil,
			},
			expected: "",
		},
		{
			name: "service with empty labels map",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels:  map[string]string{},
			},
			expected: "",
		},
		{
			name: "service with empty description",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.description": "",
				},
			},
			expected: "",
		},
		{
			name: "service with description containing special characters",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.description": "Server with special chars: @#$%^&*()!",
				},
			},
			expected: "Server with special chars: @#$%^&*()!",
		},
		{
			name: "service with description containing spaces and punctuation",
			service: Service{
				Command: "uvx mcp-server-time",
				Labels: map[string]string{
					"mcp.description": "This is a server, with punctuation. And spaces!",
				},
			},
			expected: "This is a server, with punctuation. And spaces!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDescription(tt.service)
			if result != tt.expected {
				t.Errorf("GetDescription() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestTruncateDescription tests the TruncateDescription function
func TestTruncateDescription(t *testing.T) {
	tests := []struct {
		name     string
		desc     string
		maxLen   int
		expected string
	}{
		{
			name:     "empty string",
			desc:     "",
			maxLen:   60,
			expected: "",
		},
		{
			name:     "string shorter than max length",
			desc:     "A short description",
			maxLen:   60,
			expected: "A short description",
		},
		{
			name:     "string exactly at max length",
			desc:     "This is exactly sixty characters long, no more and no less!!", // 60 chars
			maxLen:   60,
			expected: "This is exactly sixty characters long, no more and no less!!",
		},
		{
			name:     "string one character over max length",
			desc:     "This is exactly sixty-one characters long, no more no less!!!", // 61 chars
			maxLen:   60,
			expected: "This is exactly sixty-one characters long, no more no les...",
		},
		{
			name:     "long string requiring truncation",
			desc:     "This is a very long description that exceeds the maximum allowed length and should be truncated with ellipsis",
			maxLen:   60,
			expected: "This is a very long description that exceeds the maximum ...",
		},
		{
			name:     "string with special characters truncated",
			desc:     "Server with special chars: @#$%^&*()! and more text that makes it too long to display",
			maxLen:   60,
			expected: "Server with special chars: @#$%^&*()! and more text that ...",
		},
		{
			name:     "using MaxDescriptionLength constant",
			desc:     "This description is longer than sixty characters and will be truncated at the default max",
			maxLen:   MaxDescriptionLength,
			expected: "This description is longer than sixty characters and will...",
		},
		{
			name:     "custom max length shorter than default",
			desc:     "A medium length description",
			maxLen:   20,
			expected: "A medium length d...",
		},
		{
			name:     "string exactly at custom max length",
			desc:     "Exactly twenty chars",
			maxLen:   20,
			expected: "Exactly twenty chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateDescription(tt.desc, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateDescription(%q, %d) = %q, want %q", tt.desc, tt.maxLen, result, tt.expected)
			}
			// Verify truncated strings don't exceed maxLen
			if len(result) > tt.maxLen {
				t.Errorf("TruncateDescription(%q, %d) returned string of length %d, exceeds max %d", tt.desc, tt.maxLen, len(result), tt.maxLen)
			}
			// Verify truncated strings end with "..." when truncation occurred
			if len(tt.desc) > tt.maxLen && len(result) >= 3 {
				if result[len(result)-3:] != "..." {
					t.Errorf("TruncateDescription(%q, %d) = %q, expected to end with '...'", tt.desc, tt.maxLen, result)
				}
			}
		})
	}
}

// TestMaxDescriptionLengthConstant verifies the constant value
func TestMaxDescriptionLengthConstant(t *testing.T) {
	if MaxDescriptionLength != 60 {
		t.Errorf("MaxDescriptionLength = %d, want 60", MaxDescriptionLength)
	}
}
