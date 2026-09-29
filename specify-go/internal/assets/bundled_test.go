package assets

import (
	"io/fs"
	"testing"
)

func TestBundledExtension_KnownID(t *testing.T) {
	got, err := BundledExtension("git")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := fs.ReadFile(got, "extension.yml"); err != nil {
		t.Errorf("read extension.yml: %v", err)
	}
}

func TestBundledExtension_BadID(t *testing.T) {
	for _, id := range []string{"", "../etc", "GIT", "git space", "git/foo", "git.yml"} {
		if _, err := BundledExtension(id); err == nil {
			t.Errorf("BundledExtension(%q): expected error, got nil", id)
		}
	}
}

func TestBundledExtension_UnknownID(t *testing.T) {
	_, err := BundledExtension("does-not-exist")
	if err == nil {
		t.Error("expected error for unknown extension")
	}
}

func TestBundledWorkflow_KnownID(t *testing.T) {
	got, err := BundledWorkflow("speckit")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := fs.ReadFile(got, "workflow.yml"); err != nil {
		t.Errorf("read workflow.yml: %v", err)
	}
}

func TestBundledWorkflow_BadID(t *testing.T) {
	for _, id := range []string{"", "speckit:foo", "-speckit", "speckit-", "speckit/foo"} {
		if _, err := BundledWorkflow(id); err == nil {
			t.Errorf("BundledWorkflow(%q): expected error, got nil", id)
		}
	}
}

func TestBundledPreset_KnownID(t *testing.T) {
	got, err := BundledPreset("lean")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := fs.ReadFile(got, "preset.yml"); err != nil {
		t.Errorf("read preset.yml: %v", err)
	}
}

func TestBundledPreset_BadID(t *testing.T) {
	for _, id := range []string{"", "../etc", "LEAN", "lean space"} {
		if _, err := BundledPreset(id); err == nil {
			t.Errorf("BundledPreset(%q): expected error, got nil", id)
		}
	}
}
