//go:build python_integration

package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestVersionMatchesPythonReference builds the Go binary and compares its
// --version output against the Python reference implementation.
// Requires `uvx` on PATH and the Go toolchain.
func TestVersionMatchesPythonReference(t *testing.T) {
	goBin := buildGoBinary(t)

	goOut, err := exec.Command(goBin, "--version").Output()
	if err != nil {
		t.Fatalf("go --version: %v", err)
	}

	pyOut, err := exec.Command("uvx", "specify-cli", "--version").Output()
	if err != nil {
		t.Skipf("uvx specify-cli not available: %v (run `uv tool install specify-cli` first)", err)
	}

	// Compare version strings ignoring the leading "specify " prefix,
	// since Python's format is "specify X.Y.Z" and Go's may be "specify version X.Y.Z".
	goVer := normalizeVersion(string(goOut))
	pyVer := normalizeVersion(string(pyOut))

	if !strings.Contains(goVer, pyVer) && !strings.Contains(pyVer, goVer) {
		t.Errorf("version mismatch:\n  go: %q\n  py: %q", goVer, pyVer)
	}
}

// buildGoBinary compiles the Go CLI to a temp file and returns its path.
func buildGoBinary(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	bin := tmp + "/specify"
	cmd := exec.Command("go", "build", "-o", bin, ".")
	// Test cwd is cmd/specify/ (the package directory), which contains main.go.
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

// normalizeVersion strips the "specify " prefix and trailing whitespace.
func normalizeVersion(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "specify ")
	s = strings.TrimPrefix(s, "version ")
	return s
}
