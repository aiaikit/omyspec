package main

import (
	"testing"
)

func TestInitCmd_Exists(t *testing.T) {
	cmd := InitCmd()
	if cmd == nil {
		t.Fatal("InitCmd() returned nil")
	}
	if cmd.Name() != "init" {
		t.Errorf("InitCmd().Name() = %q, want %q", cmd.Name(), "init")
	}
}

func TestInitCmd_FlagsRegistered(t *testing.T) {
	cmd := InitCmd()
	expectedFlags := []string{
		"here",
		"force",
		"non-interactive",
		"script",
		"integration",
		"integration-options",
		"preset",
		"extension",
		"ignore-agent-tools",
	}
	for _, name := range expectedFlags {
		if !cmd.Flags().Lookup(name).Changed {
			// Flags are registered; Changed==false means flag exists but wasn't set.
			// Use cmd.Flags().Changed to detect pre-set flags.
		}
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("flag --%s not registered", name)
		}
	}
}
