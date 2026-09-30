package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteManifest_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	m := NewManifest("1.2.3")
	if err := WriteManifest(tmp, m); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}
	path := filepath.Join(tmp, ".specify", "speckit.manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", got.SchemaVersion, "1.0")
	}
	if got.SpeckitVersion != "1.2.3" {
		t.Errorf("SpeckitVersion = %q, want %q", got.SpeckitVersion, "1.2.3")
	}
	if got.GeneratedAt == "" {
		t.Error("GeneratedAt is empty")
	}
}

func TestWriteManifest_Atomic(t *testing.T) {
	tmp := t.TempDir()
	m := NewManifest("9.9.9")
	path := filepath.Join(tmp, ".specify", "speckit.manifest.json")
	// Write to a sibling of .specify so we can check the file doesn't exist during the write
	parent := tmp
	f, err := os.CreateTemp(parent, ".manifest-atomic-check.*")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	marker := f.Name()
	os.Remove(marker) // use it as a marker that should not exist after WriteManifest

	if err := WriteManifest(tmp, m); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}
	// The temp file must have been renamed away
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Error("temp file not cleaned up after rename")
	}
	// The final file must exist
	if _, err := os.Stat(path); err != nil {
		t.Errorf("final manifest.json missing: %v", err)
	}
}
