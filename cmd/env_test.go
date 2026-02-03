package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvVars(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-env-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	composePath := filepath.Join(tempDir, "mcp-compose.yml")

	t.Run("no env file", func(t *testing.T) {
		// Create compose file without .env file
		if err := os.WriteFile(composePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create compose file: %v", err)
		}

		envVars, err := loadEnvVars(composePath)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should contain system environment variables
		if len(envVars) == 0 {
			t.Error("Expected system environment variables to be loaded")
		}

		// Check that PATH exists (should be in system env)
		if _, exists := envVars["PATH"]; !exists {
			t.Error("Expected PATH to be in environment variables")
		}
	})

	t.Run("with env file", func(t *testing.T) {
		// Create .env file
		envPath := filepath.Join(tempDir, ".env")
		envContent := `# This is a comment
TEST_VAR=test_value
QUOTED_VAR="quoted value"
SINGLE_QUOTED='single quoted'
EMPTY_VAR=

# Another comment
OVERRIDE_VAR=from_env_file`

		if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}

		// Set an environment variable that should be overridden
		os.Setenv("OVERRIDE_VAR", "from_system")
		defer os.Unsetenv("OVERRIDE_VAR")

		envVars, err := loadEnvVars(composePath)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check that .env variables are loaded
		if envVars["TEST_VAR"] != "test_value" {
			t.Errorf("Expected TEST_VAR=test_value, got %s", envVars["TEST_VAR"])
		}

		if envVars["QUOTED_VAR"] != "quoted value" {
			t.Errorf("Expected QUOTED_VAR=quoted value, got %s", envVars["QUOTED_VAR"])
		}

		if envVars["SINGLE_QUOTED"] != "single quoted" {
			t.Errorf("Expected SINGLE_QUOTED=single quoted, got %s", envVars["SINGLE_QUOTED"])
		}

		if envVars["EMPTY_VAR"] != "" {
			t.Errorf("Expected EMPTY_VAR to be empty, got %s", envVars["EMPTY_VAR"])
		}

		// System environment should take precedence over .env file
		if envVars["OVERRIDE_VAR"] != "from_system" {
			t.Errorf("Expected OVERRIDE_VAR=from_system, got %s", envVars["OVERRIDE_VAR"])
		}
	})

	t.Run("malformed env file", func(t *testing.T) {
		// Create malformed .env file
		envPath := filepath.Join(tempDir, ".env")
		envContent := `VALID_VAR=valid
INVALID_LINE_NO_EQUALS
ANOTHER_VALID=value`

		if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}

		envVars, err := loadEnvVars(composePath)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Valid variables should still be loaded
		if envVars["VALID_VAR"] != "valid" {
			t.Errorf("Expected VALID_VAR=valid, got %s", envVars["VALID_VAR"])
		}

		if envVars["ANOTHER_VALID"] != "value" {
			t.Errorf("Expected ANOTHER_VALID=value, got %s", envVars["ANOTHER_VALID"])
		}
	})
}

func TestExpandEnvVars(t *testing.T) {
	envVars := map[string]string{
		"HOME":    "/home/user",
		"USER":    "testuser",
		"API_KEY": "secret123",
		"EMPTY":   "",
		"SPECIAL": "value with spaces",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no variables",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "single variable with braces",
			input:    "${HOME}/config",
			expected: "/home/user/config",
		},
		{
			name:     "single variable without braces",
			input:    "$USER is logged in",
			expected: "testuser is logged in",
		},
		{
			name:     "multiple variables",
			input:    "${HOME}/.config/${USER}/settings",
			expected: "/home/user/.config/testuser/settings",
		},
		{
			name:     "mixed formats",
			input:    "$HOME and ${USER} and $API_KEY",
			expected: "/home/user and testuser and secret123",
		},
		{
			name:     "empty variable",
			input:    "prefix${EMPTY}suffix",
			expected: "prefixsuffix",
		},
		{
			name:     "undefined variable",
			input:    "${UNDEFINED_VAR}",
			expected: "${UNDEFINED_VAR}",
		},
		{
			name:     "variable with special characters",
			input:    "Value: ${SPECIAL}",
			expected: "Value: value with spaces",
		},
		{
			name:     "escaped dollar sign",
			input:    "This is a literal $HOME",
			expected: "This is a literal /home/user",
		},
		{
			name:     "complex path",
			input:    "${HOME}/.config/app/${USER}/data/$API_KEY.json",
			expected: "/home/user/.config/app/testuser/data/secret123.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEnvVars(tt.input, envVars)
			if result != tt.expected {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestLoadEnvVarsErrorCases tests error scenarios
func TestLoadEnvVarsErrorCases(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-env-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("permission denied on env file", func(t *testing.T) {
		composePath := filepath.Join(tempDir, "mcp-compose.yml")
		envPath := filepath.Join(tempDir, ".env")

		// Create compose file
		if err := os.WriteFile(composePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create compose file: %v", err)
		}

		// Create .env file with restricted permissions
		if err := os.WriteFile(envPath, []byte("TEST=value"), 0000); err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}
		defer os.Chmod(envPath, 0644) // Restore permissions for cleanup

		_, err := loadEnvVars(composePath)
		if err == nil {
			t.Error("Expected error for permission denied")
		}
	})

	t.Run("env file with various formats", func(t *testing.T) {
		composePath := filepath.Join(tempDir, "mcp-compose-2.yml")
		envPath := filepath.Join(tempDir, ".env")

		// Create compose file
		if err := os.WriteFile(composePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create compose file: %v", err)
		}

		// Create .env file with various edge cases
		envContent := `# Comment at start
NORMAL_VAR=normal_value
EMPTY_VAR=
QUOTED_DOUBLE="double quoted value"
QUOTED_SINGLE='single quoted value'
SPACES_AROUND = value with spaces 
NO_VALUE_VAR
EQUALS_IN_VALUE=key=value=more
SPECIAL_CHARS=!@#$%^&*()
MULTILINE_START="this is a
# This line should be ignored
AFTER_COMMENT=after_comment_value`

		if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}

		envVars, err := loadEnvVars(composePath)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Test various parsed values
		if envVars["NORMAL_VAR"] != "normal_value" {
			t.Errorf("Expected NORMAL_VAR=normal_value, got %s", envVars["NORMAL_VAR"])
		}

		if envVars["EMPTY_VAR"] != "" {
			t.Errorf("Expected EMPTY_VAR to be empty, got %s", envVars["EMPTY_VAR"])
		}

		if envVars["QUOTED_DOUBLE"] != "double quoted value" {
			t.Errorf("Expected QUOTED_DOUBLE=double quoted value, got %s", envVars["QUOTED_DOUBLE"])
		}

		if envVars["QUOTED_SINGLE"] != "single quoted value" {
			t.Errorf("Expected QUOTED_SINGLE=single quoted value, got %s", envVars["QUOTED_SINGLE"])
		}

		if envVars["SPACES_AROUND"] != "value with spaces" {
			t.Errorf("Expected SPACES_AROUND=value with spaces, got %s", envVars["SPACES_AROUND"])
		}

		if envVars["EQUALS_IN_VALUE"] != "key=value=more" {
			t.Errorf("Expected EQUALS_IN_VALUE=key=value=more, got %s", envVars["EQUALS_IN_VALUE"])
		}

		if envVars["SPECIAL_CHARS"] != "!@#$%^&*()" {
			t.Errorf("Expected SPECIAL_CHARS=!@#$%%^&*(), got %s", envVars["SPECIAL_CHARS"])
		}

		if envVars["AFTER_COMMENT"] != "after_comment_value" {
			t.Errorf("Expected AFTER_COMMENT=after_comment_value, got %s", envVars["AFTER_COMMENT"])
		}

		// Variables that shouldn't exist
		if _, exists := envVars["NO_VALUE_VAR"]; exists {
			t.Error("NO_VALUE_VAR should not exist")
		}
	})
}
