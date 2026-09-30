package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aiaikit/speckit/internal/ui"
)

func TestEnsureConstitutionFromTemplate_Materializes(t *testing.T) {
	tmp := t.TempDir()
	// Pre-create .specify/memory so MkdirAll doesn't fail
	memDir := filepath.Join(tmp, ".specify", "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatal(err)
	}

	tr := ui.NewTracker("test")
	err := ensureConstitutionFromTemplate(tmp, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dst := filepath.Join(memDir, "constitution.md")
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("constitution.md not created: %v", err)
	}
}

func TestEnsureConstitutionFromTemplate_SkipsIfExists(t *testing.T) {
	tmp := t.TempDir()
	memDir := filepath.Join(tmp, ".specify", "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(memDir, "constitution.md")
	if err := os.WriteFile(existing, []byte("already there"), 0644); err != nil {
		t.Fatal(err)
	}

	tr := ui.NewTracker("test")
	err := ensureConstitutionFromTemplate(tmp, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "already there" {
		t.Fatalf("constitution was overwritten, want 'already there', got %q", string(got))
	}
}

func TestEnsureExecutableScripts_ChmodsShFiles(t *testing.T) {
	tmp := t.TempDir()
	scriptsDir := filepath.Join(tmp, ".specify", "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write a non-executable .sh file
	scriptPath := filepath.Join(scriptsDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho hi"), 0644); err != nil {
		t.Fatal(err)
	}

	tr := ui.NewTracker("test")
	err := ensureExecutableScripts(tmp, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat script: %v", err)
	}
	if mode := info.Mode(); mode&0111 == 0 {
		t.Fatalf("script not executable, mode=%v", mode)
	}
}

func TestEnsureExecutableScripts_SkipsIfNoScriptsDir(t *testing.T) {
	tmp := t.TempDir()
	// .specify exists but no scripts/
	if err := os.MkdirAll(filepath.Join(tmp, ".specify"), 0755); err != nil {
		t.Fatal(err)
	}

	tr := ui.NewTracker("test")
	err := ensureExecutableScripts(tmp, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitCmd_RunE_UnknownIntegration(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(oldCwd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	cmd := InitCmd()
	cmd.SetArgs([]string{"--here", "--integration", "not-a-real-integration"})
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for unknown integration, got nil")
	}
}

func TestInitCmd_RunE_HereWithForce(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "existing.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(oldCwd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	cmd := InitCmd()
	cmd.SetArgs([]string{"--here", "--force", "--integration", "claude"})
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error with --force: %v", err)
	}

	specDir := filepath.Join(tmp, ".specify")
	if _, err := os.Stat(specDir); err != nil {
		t.Errorf(".specify directory not created: %v", err)
	}

	for _, f := range []string{
		"speckit.manifest.json",
		"integration.json",
		"init-options.json",
	} {
		path := filepath.Join(specDir, f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s not created: %v", f, err)
		}
	}
}

func TestInitCmd_RunE_HereNonEmptyWithoutForce(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "existing.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(oldCwd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	cmd := InitCmd()
	cmd.SetArgs([]string{"--here"})
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for non-empty directory without --force, got nil")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Errorf("error message should mention 'not empty', got: %v", err)
	}
}
