package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Confirm reads one line from stdin and returns true for "y"/"yes" (case
// insensitive), false for "n"/"no". An empty line returns defaultYes.
// Returns io.EOF on closed/empty stdin so callers can distinguish
// "user said no" from "no input available".
//
// The third parameter is the input Reader (defaults to os.Stdin in main.go
// wrappers; tests inject strings.NewReader). This signature lets tests
// avoid faking a TTY — phase 2 will add a TTY-aware variant.
func Confirm(prompt string, defaultYes bool, in io.Reader) (bool, error) {
	suffix := "[Y/n]"
	if !defaultYes {
		suffix = "[y/N]"
	}
	fmt.Fprintf(defaultWriter(), "%s %s ", prompt, suffix)

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, io.EOF
	}
	line := strings.ToLower(strings.TrimSpace(scanner.Text()))
	switch line {
	case "":
		return defaultYes, nil
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return defaultYes, fmt.Errorf("unrecognized response %q", scanner.Text())
	}
}

func defaultWriter() io.Writer { return os.Stderr }
