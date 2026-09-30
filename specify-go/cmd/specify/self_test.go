package main

import (
	"testing"
)

func TestSelfCheckCmd_Exists(t *testing.T) {
	cmd := SelfCheckCmd()
	if cmd == nil {
		t.Fatal("SelfCheckCmd() returned nil")
	}
	if cmd.Use != "check" {
		t.Errorf("Use = %q, want %q", cmd.Use, "check")
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
}
