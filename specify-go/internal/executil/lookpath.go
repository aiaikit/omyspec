package executil

import "os/exec"

// LookPath wraps exec.LookPath so callers don't import os/exec directly.
func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
