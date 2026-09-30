package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestUpgradeCmd_DryRun_NoPanic(t *testing.T) {
	cmd := SelfUpgradeCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.Flags().Set("dry-run", "true")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
}

func TestUpgradeCmd_DryRunOutputFormat(t *testing.T) {
	cmd := SelfUpgradeCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.Flags().Set("dry-run", "true")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[dry run]") {
		t.Errorf("expected [dry run] in output, got: %s", out)
	}
}

func TestUpgradeCmd_TagFlagExists(t *testing.T) {
	cmd := SelfUpgradeCmd()
	flag := cmd.Flags().Lookup("tag")
	if flag == nil {
		t.Fatal("--tag flag not found")
	}
	if flag.DefValue != "" {
		t.Errorf("tag default = %q, want empty", flag.DefValue)
	}
}

func TestUpgradeCmd_DryRunFlagExists(t *testing.T) {
	cmd := SelfUpgradeCmd()
	flag := cmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Fatal("--dry-run flag not found")
	}
}

func TestUpgradeCmd_Use(t *testing.T) {
	cmd := SelfUpgradeCmd()
	if cmd.Use != "upgrade" {
		t.Errorf("Use = %q, want %q", cmd.Use, "upgrade")
	}
}
