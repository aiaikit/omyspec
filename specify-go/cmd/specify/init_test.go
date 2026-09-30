package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
