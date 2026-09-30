package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Manifest is the speckit.manifest.json structure.
type Manifest struct {
	SchemaVersion  string `json:"schema_version"`
	SpeckitVersion string `json:"speckit_version"`
	GeneratedAt   string `json:"generated_at"`
}

// WriteManifest writes speckit.manifest.json atomically to projectRoot/.specify/.
func WriteManifest(projectRoot string, m *Manifest) error {
	path := filepath.Join(projectRoot, ".specify", "speckit.manifest.json")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".manifest.json.tmp.*")
	if err != nil {
		return err
	}
	removeTmp := func() { os.Remove(f.Name()) }
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		f.Close()
		removeTmp()
		return err
	}
	if err := f.Close(); err != nil {
		removeTmp()
		return err
	}
	if err := os.Rename(f.Name(), path); err != nil {
		removeTmp()
		return err
	}
	return nil
}

// NewManifest returns a Manifest with current time and given speckit version.
func NewManifest(speckitVersion string) *Manifest {
	return &Manifest{
		SchemaVersion:   "1.0",
		SpeckitVersion: speckitVersion,
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
	}
}
