//go:build integration

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aiaikit/speckit/internal/integration"
)

func TestInit_CreatesDotSpecify(t *testing.T) {
	tmp, err := os.MkdirTemp("", "init-test")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmp)

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	cmd := InitCmd()
	cmd.SetArgs([]string{"--here", "--integration", "claude"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify directories exist
	dirs := []string{".specify", ".specify/scripts", ".specify/templates", ".specify/workflows/speckit", ".specify/memory"}
	for _, d := range dirs {
		path := filepath.Join(tmp, d)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("missing directory: %s", d)
		}
	}

	// Verify files exist
	files := []string{".specify/speckit.manifest.json", ".specify/integration.json", ".specify/init-options.json"}
	for _, f := range files {
		path := filepath.Join(tmp, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("missing file: %s", f)
		}
	}

	// Verify integration.json content
	state, err := integration.ReadState(tmp)
	if err != nil {
		t.Fatalf("ReadState failed: %v", err)
	}
	if state == nil {
		t.Fatal("integration.json was not created")
	}
	if state.Integration != "claude" {
		t.Errorf("Integration = %q, want %q", state.Integration, "claude")
	}
	if state.IntegrationStateSchema != 1 {
		t.Errorf("IntegrationStateSchema = %d, want %d", state.IntegrationStateSchema, 1)
	}
}

func TestInit_RollsBackOnError(t *testing.T) {
	t.Skip("defer to Phase 3 when more file types exist")
}
