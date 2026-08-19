package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureDescriptionListOutput(t *testing.T, display func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	display()
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout pipe: %v", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	return string(output)
}

func TestDisplayServersWithDescription(t *testing.T) {
	originalComposeFile := composeFile
	originalLongFormat := longFormat
	originalCommandFormat := commandFormat
	originalShowDescription := showDescription
	defer func() {
		composeFile = originalComposeFile
		longFormat = originalLongFormat
		commandFormat = originalCommandFormat
		showDescription = originalShowDescription
	}()

	longDescription := strings.Repeat("x", MaxDescriptionLength+1)
	composeFile = filepath.Join(t.TempDir(), "mcp-compose.yml")
	if err := os.WriteFile(composeFile, []byte(`services:
  described:
    command: uvx described-server
    labels:
      mcp.description: description
`), 0644); err != nil {
		t.Fatalf("write compose file: %v", err)
	}

	servers := map[string]Service{
		"described": {
			Command: "uvx described-server",
			Labels: map[string]string{
				"mcp.description": longDescription,
			},
		},
	}

	tests := []struct {
		name            string
		longFormat      bool
		commandFormat   bool
		expectedColumns []string
		expectedDesc    string
	}{
		{
			name:            "default format truncates description",
			expectedColumns: []string{"NAME", "PROFILES", "DESCRIPTION"},
			expectedDesc:    strings.Repeat("x", MaxDescriptionLength-3) + "...",
		},
		{
			name:            "long format truncates description",
			longFormat:      true,
			expectedColumns: []string{"NAME", "PROFILES", "COMMAND", "ENVVARS", "DESCRIPTION"},
			expectedDesc:    strings.Repeat("x", MaxDescriptionLength-3) + "...",
		},
		{
			name:            "command format shows full description",
			commandFormat:   true,
			expectedColumns: []string{"NAME", "COMMAND", "DESCRIPTION"},
			expectedDesc:    longDescription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			longFormat = tt.longFormat
			commandFormat = tt.commandFormat
			showDescription = true

			output := captureDescriptionListOutput(t, func() {
				displayServers(servers)
			})

			header := strings.Fields(strings.SplitN(output, "\n", 2)[0])
			if strings.Join(header, "|") != strings.Join(tt.expectedColumns, "|") {
				t.Errorf("output header = %q, want %q:\n%s", header, tt.expectedColumns, output)
			}
			if !strings.Contains(output, tt.expectedDesc) {
				t.Errorf("output missing description %q:\n%s", tt.expectedDesc, output)
			}
		})
	}
}
