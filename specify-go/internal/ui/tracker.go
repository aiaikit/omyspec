package ui

import (
	"fmt"
	"strings"
)

// Status tracks the state of a single step in a StepTracker.
// Python version uses these as ordinals only, not as a state machine;
// the Go version mirrors that — Mark accepts any transition.
type Status int

const (
	Pending Status = iota
	Running
	Done
	Error
	Skipped
)

// Tracker records per-step statuses and renders them as a tree.
type Tracker interface {
	Add(key, label string)
	Mark(key string, status Status, detail string)
	Attach(refresh func())
	Render() string
}

type step struct {
	key, label, pending string
	status              Status
}

type trackerImpl struct {
	title   string
	steps   []step
	refresh func()
}

// NewTracker creates a Tracker with the given title.
func NewTracker(title string) Tracker {
	return &trackerImpl{title: title}
}

func (t *trackerImpl) Add(key, label string) {
	for _, s := range t.steps {
		if s.key == key {
			return
		}
	}
	t.steps = append(t.steps, step{key: key, label: label, pending: label, status: Pending})
	t.notify()
}

func (t *trackerImpl) Mark(key string, status Status, detail string) {
	for i := range t.steps {
		if t.steps[i].key == key {
			t.steps[i].status = status
			if detail != "" {
				t.steps[i].pending = detail
			}
			t.notify()
			return
		}
	}
	// Mark with unknown key: silently append to surface the error in Render.
	t.steps = append(t.steps, step{key: key, label: key, pending: detail, status: status})
	t.notify()
}

func (t *trackerImpl) Attach(refresh func()) {
	t.refresh = refresh
}

func (t *trackerImpl) notify() {
	if t.refresh != nil {
		t.refresh()
	}
}

func (t *trackerImpl) Render() string {
	var b strings.Builder
	b.WriteString(t.title)
	b.WriteString("\n")
	for _, s := range t.steps {
		sym := mark(s.status)
		b.WriteString(fmt.Sprintf("  %s %s", sym, s.label))
		if s.pending != "" && s.pending != s.label {
			b.WriteString(fmt.Sprintf(" (%s)", s.pending))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func mark(s Status) string {
	switch s {
	case Done:
		return "[x]"
	case Error:
		return "[!]"
	case Running:
		return "[~]"
	case Skipped:
		return "[-]"
	default:
		return "[ ]"
	}
}
