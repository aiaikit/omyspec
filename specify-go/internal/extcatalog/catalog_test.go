package extcatalog

import (
	"errors"
	"testing"
)

func TestLoadCatalog_Embedded(t *testing.T) {
	cat, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog() error = %v", err)
	}
	if cat == nil {
		t.Fatal("LoadCatalog() returned nil catalog")
	}
	if cat.SchemaVersion == "" {
		t.Error("LoadCatalog().SchemaVersion is empty")
	}
	if len(cat.Extensions) == 0 {
		t.Error("LoadCatalog().Extensions is empty")
	}
}

func TestLoadCatalog_KnownExtension(t *testing.T) {
	ext, err := GetExtension("agent-context")
	if err != nil {
		t.Fatalf("GetExtension() error = %v", err)
	}
	if ext == nil {
		t.Fatal("GetExtension() returned nil")
	}
	if ext.ID != "agent-context" {
		t.Errorf("ext.ID = %q, want %q", ext.ID, "agent-context")
	}
	if ext.Name == "" {
		t.Error("ext.Name is empty")
	}
	if ext.Version == "" {
		t.Error("ext.Version is empty")
	}
	if !ext.Bundled {
		t.Error("agent-context should be bundled")
	}
}

func TestLoadCatalog_UnknownExtension(t *testing.T) {
	_, err := GetExtension("does-not-exist")
	if err == nil {
		t.Fatal("GetExtension() expected error for unknown extension")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetExtension() error = %v, want ErrNotFound", err)
	}
}

func TestListBundledExtensions(t *testing.T) {
	exts := ListBundledExtensions()
	if len(exts) == 0 {
		t.Fatal("ListBundledExtensions() returned empty")
	}
	for _, ext := range exts {
		if !ext.Bundled {
			t.Errorf("ListBundledExtensions() returned non-bundled extension %q", ext.ID)
		}
	}
}
