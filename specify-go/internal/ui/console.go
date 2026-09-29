package ui

import (
	"fmt"
	"io"
)

// Style is a placeholder for visual styling. Phase 1 uses defaults;
// Phase 2+ will add color, border style, padding knobs here.
type Style struct{}

// Console separates stdout and stderr writes so --json consumers can
// reliably parse stdout without contamination from progress messages.
type Console interface {
	Print(msg string)
	PrintErr(msg string)
	PrintPanel(title, body string, style Style)
}

// consoleImpl is the default Console backed by two io.Writers.
type consoleImpl struct {
	out io.Writer
	err io.Writer
}

// NewConsole returns a Console that writes to out (stdout) and err (stderr).
// Pass the same writer to both for tests.
func NewConsole(out, err io.Writer) Console {
	return &consoleImpl{out: out, err: err}
}

func (c *consoleImpl) Print(msg string) {
	fmt.Fprintln(c.out, msg)
}

func (c *consoleImpl) PrintErr(msg string) {
	fmt.Fprintln(c.err, msg)
}

func (c *consoleImpl) PrintPanel(title, body string, _ Style) {
	fmt.Fprintf(c.out, "=== %s ===\n%s\n=== end ===\n", title, body)
}
