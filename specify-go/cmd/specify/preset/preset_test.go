package preset

import (
	"strings"
	"testing"
)

func TestPresetCmd_Exists(t *testing.T) {
	cmd := PresetCmd()
	if cmd == nil {
		t.Fatal("PresetCmd() returned nil")
	}
	if cmd.Name() != "preset" {
		t.Errorf("name = %q, want %q", cmd.Name(), "preset")
	}
	if len(cmd.Commands()) == 0 {
		t.Error("PresetCmd should have subcommands")
	}
}

func TestPresetListCmd(t *testing.T) {
	cmd := PresetListCmd()
	if cmd == nil {
		t.Fatal("PresetListCmd() returned nil")
	}
	if cmd.Name() != "list" {
		t.Errorf("name = %q, want %q", cmd.Name(), "list")
	}
}

func TestPresetInfoCmd_NotFound(t *testing.T) {
	cmd := PresetInfoCmd()
	if cmd == nil {
		t.Fatal("PresetInfoCmd() returned nil")
	}
	if cmd.Name() != "info" {
		t.Errorf("name = %q, want %q", cmd.Name(), "info")
	}
	if cmd.Use != "info [id]" {
		t.Errorf("use = %q, want %q", cmd.Use, "info [id]")
	}

	buf := &strings.Builder{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"nonexistent-id"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent preset id")
	}
	if !strings.Contains(err.Error(), "nonexistent-id") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "nonexistent-id")
	}
}
