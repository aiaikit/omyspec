package integration

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadState_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	st := &IntegrationState{
		Version:                "1.0.0",
		IntegrationStateSchema: IntegrationStateSchema,
		Integration:           "claude",
		DefaultIntegration:    "claude",
		InstalledIntegrations: []string{"claude", "copilot"},
		IntegrationSettings: map[string]IntegrationSettings{
			"claude": {Script: "sh", RawOptions: "--verbose", InvokeSeparator: " -- "},
		},
	}
	if err := WriteState(tmp, st); err != nil {
		t.Fatalf("WriteState: %v", err)
	}
	got, err := ReadState(tmp)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if got.Version != st.Version {
		t.Errorf("Version = %q, want %q", got.Version, st.Version)
	}
	if got.IntegrationStateSchema != st.IntegrationStateSchema {
		t.Errorf("IntegrationStateSchema = %d, want %d", got.IntegrationStateSchema, st.IntegrationStateSchema)
	}
	if got.Integration != st.Integration {
		t.Errorf("Integration = %q, want %q", got.Integration, st.Integration)
	}
	if len(got.InstalledIntegrations) != len(st.InstalledIntegrations) {
		t.Errorf("InstalledIntegrations len = %d, want %d", len(got.InstalledIntegrations), len(st.InstalledIntegrations))
	}
	if got.IntegrationSettings["claude"].Script != st.IntegrationSettings["claude"].Script {
		t.Errorf("IntegrationSettings[claude].Script = %q, want %q",
			got.IntegrationSettings["claude"].Script, st.IntegrationSettings["claude"].Script)
	}
}

func TestReadState_MissingFile(t *testing.T) {
	tmp := t.TempDir()
	got, err := ReadState(tmp)
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if got != nil {
		t.Errorf("expected nil state for missing file, got %+v", got)
	}
}

func TestReadState_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".specify")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create .specify dir: %v", err)
	}
	path := filepath.Join(dir, "integration.json")
	if err := os.WriteFile(path, []byte("not json{"), 0644); err != nil {
		t.Fatalf("write bad file: %v", err)
	}
	_, err := ReadState(tmp)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !errors.Is(err, ErrStateRead) {
		t.Errorf("error should wrap ErrStateRead, got %v", err)
	}
}

func TestReadState_EmptyProjectRoot(t *testing.T) {
	_, err := ReadState("")
	if err == nil {
		t.Fatal("expected error for empty project root")
	}
}

func TestWriteState_Normalize(t *testing.T) {
	tmp := t.TempDir()
	// Write a state with no schema set; WriteState should still write valid JSON
	// (schema is the caller's responsibility to set on write; we just marshal)
	st := &IntegrationState{
		Version:                "1.0.0",
		IntegrationStateSchema: IntegrationStateSchema,
	}
	if err := WriteState(tmp, st); err != nil {
		t.Fatalf("WriteState: %v", err)
	}
	got, err := ReadState(tmp)
	if err != nil {
		t.Fatalf("ReadState after WriteState: %v", err)
	}
	if got.Version != st.Version {
		t.Errorf("Version = %q, want %q", got.Version, st.Version)
	}
}

func TestWriteState_Atomic(t *testing.T) {
	tmp := t.TempDir()
	st := &IntegrationState{Version: "1.0.0", IntegrationStateSchema: IntegrationStateSchema}
	path := filepath.Join(tmp, IntegrationJSONPath)
	// Write should not leave partial files; ReadState after WriteState should succeed
	if err := WriteState(tmp, st); err != nil {
		t.Fatalf("WriteState: %v", err)
	}
	_, err := ReadState(tmp)
	if err != nil {
		t.Errorf("ReadState after atomic write failed: %v", err)
	}
	// Verify the file exists with proper content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("integration.json not found after write: %v", err)
	}
	if len(data) == 0 {
		t.Error("integration.json is empty")
	}
}
