package ui

import (
	"strings"
	"testing"
)

func TestTracker_AddThenMark(t *testing.T) {
	tr := NewTracker("Test Title")
	tr.Add("k1", "label 1")
	tr.Add("k2", "label 2")
	tr.Mark("k1", Running, "")
	tr.Mark("k1", Done, "completed")
	tr.Mark("k2", Error, "failed")
	out := tr.Render()
	if !strings.Contains(out, "Test Title") {
		t.Errorf("render missing title:\n%s", out)
	}
	if !strings.Contains(out, "label 1") || !strings.Contains(out, "label 2") {
		t.Errorf("render missing labels:\n%s", out)
	}
	if !strings.Contains(out, "completed") {
		t.Errorf("render missing detail:\n%s", out)
	}
}

func TestTracker_StatusOrdinals(t *testing.T) {
	if Pending >= Running || Running >= Done || Done >= Error || Error >= Skipped {
		t.Errorf("status ordinals not in expected order: %d %d %d %d %d",
			Pending, Running, Done, Error, Skipped)
	}
}

func TestTracker_AttachFiresOnMark(t *testing.T) {
	tr := NewTracker("X")
	tr.Add("k", "l")
	calls := 0
	tr.Attach(func() { calls++ })
	tr.Mark("k", Done, "ok")
	tr.Mark("k", Error, "boom")
	if calls != 2 {
		t.Errorf("refresh called %d times, want 2", calls)
	}
}

func TestTracker_MarkUnknownKeyNoCrash(t *testing.T) {
	tr := NewTracker("X")
	tr.Mark("missing", Done, "")
	// No assertion — test passes if no panic.
}
