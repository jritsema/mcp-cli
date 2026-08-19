package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// writeSelectTestCompose writes a compose file with a mix of default and
// profiled servers and returns the loaded config.
func writeSelectTestCompose(t *testing.T) *ComposeConfig {
	t.Helper()

	composeContent := `services:
  default-server:
    command: python -m default_server

  time:
    command: uvx mcp-server-time
    labels:
      mcp.profile: time

  github:
    command: npx -y @modelcontextprotocol/server-github
    labels:
      mcp.profile: development, programming
`

	dir := t.TempDir()
	path := filepath.Join(dir, "mcp-compose.yml")
	if err := os.WriteFile(path, []byte(composeContent), 0644); err != nil {
		t.Fatalf("write compose: %v", err)
	}

	config, err := loadComposeFile(path)
	if err != nil {
		t.Fatalf("load compose: %v", err)
	}
	return config
}

// TestSelectServersSingleServerIgnoresProfile is the regression test for the
// bug where `mcp set -s <name>` (no profile arg) could not find a server that
// belonged to a profile, because the profile filter ran before the -s lookup.
func TestSelectServersSingleServerIgnoresProfile(t *testing.T) {
	config := writeSelectTestCompose(t)

	cases := []string{"time", "github", "default-server"}
	for _, name := range cases {
		// No profile provided — this used to fail for profiled servers.
		servers, err := selectServers(config, "", name)
		if err != nil {
			t.Fatalf("selectServers(-s %q) returned error: %v", name, err)
		}
		if len(servers) != 1 {
			t.Fatalf("selectServers(-s %q): expected 1 server, got %d", name, len(servers))
		}
		if _, ok := servers[name]; !ok {
			t.Fatalf("selectServers(-s %q): expected server %q in result", name, name)
		}
	}
}

func TestSelectServersSingleServerNotFound(t *testing.T) {
	config := writeSelectTestCompose(t)

	if _, err := selectServers(config, "", "does-not-exist"); err == nil {
		t.Fatal("expected error for unknown server, got nil")
	}
}

func TestSelectServersProfile(t *testing.T) {
	config := writeSelectTestCompose(t)

	// Profile selection includes defaults plus servers in the profile.
	servers, err := selectServers(config, "development", "")
	if err != nil {
		t.Fatalf("selectServers(profile): %v", err)
	}
	if _, ok := servers["github"]; !ok {
		t.Fatal("expected 'github' in development profile selection")
	}
	if _, ok := servers["default-server"]; !ok {
		t.Fatal("expected default server included in profile selection")
	}
	if _, ok := servers["time"]; ok {
		t.Fatal("did not expect 'time' in development profile selection")
	}
}

func TestSelectServersMultipleServersIgnoreProfile(t *testing.T) {
	config := writeSelectTestCompose(t)

	servers, err := selectServers(config, "development", "time", "github", "time")
	if err != nil {
		t.Fatalf("selectServers(multiple -s): %v", err)
	}

	expectedServers := []string{"time", "github"}
	if len(servers) != len(expectedServers) {
		t.Fatalf("selectServers(multiple -s): expected %d servers, got %d", len(expectedServers), len(servers))
	}
	for _, name := range expectedServers {
		if _, ok := servers[name]; !ok {
			t.Fatalf("selectServers(multiple -s): expected server %q in result", name)
		}
	}
}
