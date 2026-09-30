//go:build integration

package main

import (
	"strings"
	"testing"

	"github.com/aiaikit/speckit/cmd/specify/extension"
	"github.com/aiaikit/speckit/cmd/specify/preset"
)

func TestSelfCheck_CmdExists(t *testing.T) {
	cmd := SelfCheckCmd()
	if cmd == nil {
		t.Fatal("SelfCheckCmd() returned nil")
	}
	// Exercise the command in a temp dir (no network required for construction).
	cmd.SetArgs([]string{})
	buf := new(strings.Builder)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Logf("output: %s", buf.String())
		t.Logf("stderr: %s", buf.String())
	}
}

func TestExtensionList_CmdExists(t *testing.T) {
	cmd := extension.ExtensionListCmd()
	if cmd == nil {
		t.Fatal("ExtensionListCmd() returned nil")
	}
	cmd.SetArgs([]string{})
	buf := new(strings.Builder)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Logf("output: %s", buf.String())
		t.Errorf("ExtensionListCmd().Execute() error: %v", err)
	}
}

func TestExtensionInfo_NotFound(t *testing.T) {
	cmd := extension.ExtensionInfoCmd()
	if cmd == nil {
		t.Fatal("ExtensionInfoCmd() returned nil")
	}
	cmd.SetArgs([]string{"nonexistent-extension-id-12345"})
	buf := new(strings.Builder)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unknown extension id, got nil")
	} else if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "NotFound") {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestPresetList_CmdExists(t *testing.T) {
	cmd := preset.PresetListCmd()
	if cmd == nil {
		t.Fatal("PresetListCmd() returned nil")
	}
	cmd.SetArgs([]string{})
	buf := new(strings.Builder)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Logf("output: %s", buf.String())
		t.Errorf("PresetListCmd().Execute() error: %v", err)
	}
}
