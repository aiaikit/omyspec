package extension

import (
	"strings"
	"testing"
)

func TestExtensionCmd_Exists(t *testing.T) {
	cmd := ExtensionCmd()
	if cmd == nil {
		t.Fatal("ExtensionCmd() returned nil")
	}
	if cmd.Name() != "extension" {
		t.Errorf("name = %q, want %q", cmd.Name(), "extension")
	}
	if len(cmd.Commands()) == 0 {
		t.Error("ExtensionCmd should have subcommands")
	}
}

func TestExtensionListCmd(t *testing.T) {
	cmd := ExtensionListCmd()
	if cmd == nil {
		t.Fatal("ExtensionListCmd() returned nil")
	}
	if cmd.Name() != "list" {
		t.Errorf("name = %q, want %q", cmd.Name(), "list")
	}
}

func TestExtensionRemoveCmd_BundledError(t *testing.T) {
	cmd := ExtensionRemoveCmd()
	if cmd == nil {
		t.Fatal("ExtensionRemoveCmd() returned nil")
	}
	if cmd.Name() != "remove" {
		t.Errorf("name = %q, want %q", cmd.Name(), "remove")
	}
	if cmd.Use != "remove [id]" {
		t.Errorf("use = %q, want %q", cmd.Use, "remove [id]")
	}
}

func TestExtensionRemoveCmd_NotFound(t *testing.T) {
	cmd := ExtensionRemoveCmd()
	if cmd == nil {
		t.Fatal("ExtensionRemoveCmd() returned nil")
	}

	buf := &strings.Builder{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"nonexistent-id"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent extension id")
	}
	if !strings.Contains(err.Error(), "nonexistent-id") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "nonexistent-id")
	}
}

func TestExtensionInfoCmd_NotFound(t *testing.T) {
	cmd := ExtensionInfoCmd()
	if cmd == nil {
		t.Fatal("ExtensionInfoCmd() returned nil")
	}
	if cmd.Name() != "info" {
		t.Errorf("name = %q, want %q", cmd.Name(), "info")
	}
	if cmd.Use != "info [id]" {
		t.Errorf("use = %q, want %q", cmd.Use, "info [id]")
	}

	// Execute should return error for unknown id
	buf := &strings.Builder{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"nonexistent-id"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent extension id")
	}
	if !strings.Contains(err.Error(), "nonexistent-id") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "nonexistent-id")
	}
}
