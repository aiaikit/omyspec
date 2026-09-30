package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectInstallMethod_SourceCheckout(t *testing.T) {
	// Create a temp dir with a go.mod so DetectInstallMethod treats it as
	// a source checkout via runtime/debug.ReadBuildInfo's module path match.
	// Phase 1 simplification: when running under `go test`, ReadBuildInfo's
	// Main.Path is the test binary's module; we can't easily fake this. Test
	// only the fallback paths.
	t.Skip("source checkout detection requires test-binary setup; covered in Phase 7")
}

func TestDetectInstallMethod_Unsupported(t *testing.T) {
	// Save and restore PATH to ensure no installer matches.
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", "")

	// Save and restore argv0 — we use os.Executable which is the test binary.
	got := DetectInstallMethod()
	if got != InstallUnsupported && got != InstallSourceCheckout {
		t.Errorf("DetectInstallMethod() = %d, want Unsupported or SourceCheckout", got)
	}
}

func TestDetectInstallMethod_PathMatching(t *testing.T) {
	// Create a fake executable path that matches one of the installer prefixes.
	tmp := t.TempDir()
	fake := filepath.Join(tmp, "uv-tool-bin", "specify")
	if err := os.MkdirAll(filepath.Dir(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fake, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Inject fake via re-exec is not feasible; skip — covered by unit testing
	// path-match logic directly.
	t.Skip("path matching covered by direct unit test of matchInstallerPath")
}
