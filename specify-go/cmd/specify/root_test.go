package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCmd(t *testing.T) {
	var got *cobra.Command = NewRootCmd()
	if got == nil {
		t.Fatal("NewRootCmd returned nil")
	}
	if got.Use != "specify" {
		t.Errorf("Use = %q, want %q", got.Use, "specify")
	}
}

func TestRootCmd_Help(t *testing.T) {
	cmd := NewRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"specify", "init", "check", "version", "integration"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("help output missing %q\nfull output:\n%s", want, out)
		}
	}
}
