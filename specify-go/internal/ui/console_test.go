package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewConsoleWritesToProvidedBuffers(t *testing.T) {
	var out, errBuf bytes.Buffer
	c := NewConsole(&out, &errBuf)
	c.Print("hello stdout")
	c.PrintErr("hello stderr")
	if got := out.String(); got != "hello stdout" {
		t.Errorf("stdout = %q, want %q", got, "hello stdout")
	}
	if got := errBuf.String(); got != "hello stderr" {
		t.Errorf("stderr = %q, want %q", got, "hello stderr")
	}
}

func TestPrintPanelIncludesBoth(t *testing.T) {
	var out bytes.Buffer
	c := NewConsole(&out, &out)
	c.PrintPanel("Title", "Body", Style{})
	s := out.String()
	if !strings.Contains(s, "Title") || !strings.Contains(s, "Body") {
		t.Errorf("panel missing title/body:\n%s", s)
	}
}