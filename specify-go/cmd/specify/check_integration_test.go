//go:build integration

package main

import (
	"bytes"
	"os"
	"testing"
)

func TestCheck_NoIntegrationJSON(t *testing.T) {
	tmp, err := os.MkdirTemp("", "check-test")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmp)

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	// Run check command in a directory with no .specify/
	var stdout, stderr bytes.Buffer
	cmd := newCheckCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("check command returned error: %v", err)
	}

	output := stdout.String()
	if output == "" && stderr.String() == "" {
		t.Fatal("expected some output when .specify/ is missing")
	}
}
