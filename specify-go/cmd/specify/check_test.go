package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aiaikit/speckit/internal/ui"
)

func TestCheckCommand_Exists(t *testing.T) {
	cmd := NewRootCmd()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Name() == "check" {
			found = true
			break
		}
	}
	if !found {
		t.Error("check subcommand not registered")
	}
}

func TestCheck_RendersTracker(t *testing.T) {
	var buf bytes.Buffer
	c := ui.NewConsole(&buf, &buf)
	tr := ui.NewTracker("Check Available Tools")
	if c == nil || tr == nil {
		t.Fatal("nil ui")
	}
	tr.Add("claude", "Claude")
	tr.Mark("claude", ui.Done, "available")
	out := tr.Render()
	if !strings.Contains(out, "Claude") {
		t.Errorf("render missing Claude:\n%s", out)
	}
}
