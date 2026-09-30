package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aiaikit/speckit/internal/ui"
)

func TestCheckTools_NoIntegrationJSON(t *testing.T) {
	tmp := t.TempDir()
	// No .specify directory at all
	tr := ui.NewTracker("Check Available Tools")
	err := CheckTools(tmp, tr)
	if err != nil {
		t.Fatalf("CheckTools on empty dir: expected nil, got %v", err)
	}
	// Nothing was added to tracker since there's nothing to check
	got := tr.Render()
	if got != "Check Available Tools\n" {
		t.Errorf("unexpected render: %q", got)
	}
}

func TestCheckTools_UnknownIntegration(t *testing.T) {
	tmp := t.TempDir()
	// Write integration.json with unknown integration
	st := &IntegrationState{
		Version:                "1.0.0",
		IntegrationStateSchema: IntegrationStateSchema,
		InstalledIntegrations: []string{"unknown-integration"},
	}
	dir := filepath.Join(tmp, ".specify")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir .specify: %v", err)
	}
	if err := WriteState(tmp, st); err != nil {
		t.Fatalf("WriteState: %v", err)
	}

	tr := ui.NewTracker("Check Available Tools")
	err := CheckTools(tmp, tr)
	if err != nil {
		t.Fatalf("CheckTools with unknown integration: expected nil, got %v", err)
	}
	rendered := tr.Render()
	if !strings.Contains(rendered, "unknown integration") {
		t.Errorf("expected 'unknown integration' in output, got: %s", rendered)
	}
}

func TestCheckTools_IDEBasedIntegration(t *testing.T) {
	tmp := t.TempDir()
	st := &IntegrationState{
		Version:                "1.0.0",
		IntegrationStateSchema: IntegrationStateSchema,
		InstalledIntegrations: []string{"copilot"},
	}
	dir := filepath.Join(tmp, ".specify")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir .specify: %v", err)
	}
	if err := WriteState(tmp, st); err != nil {
		t.Fatalf("WriteState: %v", err)
	}

	tr := ui.NewTracker("Check Available Tools")
	err := CheckTools(tmp, tr)
	if err != nil {
		t.Fatalf("CheckTools with IDE-based: expected nil, got %v", err)
	}
	rendered := tr.Render()
	if !strings.Contains(rendered, "IDE-based") {
		t.Errorf("expected 'IDE-based' in output, got: %s", rendered)
	}
}
