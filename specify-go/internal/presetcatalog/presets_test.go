package presetcatalog

import (
	"errors"
	"testing"
)

func TestListBundledPresets(t *testing.T) {
	presets := ListBundledPresets()
	if len(presets) == 0 {
		t.Fatal("ListBundledPresets() returned empty")
	}
	for _, p := range presets {
		if !p.Bundled {
			t.Errorf("ListBundledPresets() returned non-bundled preset %q", p.ID)
		}
	}
}

func TestGetPreset_KnownLean(t *testing.T) {
	p, err := GetPreset("lean")
	if err != nil {
		t.Fatalf("GetPreset() error = %v", err)
	}
	if p == nil {
		t.Fatal("GetPreset() returned nil")
	}
	if p.ID != "lean" {
		t.Errorf("p.ID = %q, want %q", p.ID, "lean")
	}
	if p.Name == "" {
		t.Error("p.Name is empty")
	}
	if p.Version == "" {
		t.Error("p.Version is empty")
	}
	if !p.Bundled {
		t.Error("lean preset should be bundled")
	}
}

func TestGetPreset_Unknown(t *testing.T) {
	_, err := GetPreset("does-not-exist")
	if err == nil {
		t.Fatal("GetPreset() expected error for unknown preset")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetPreset() error = %v, want ErrNotFound", err)
	}
}
