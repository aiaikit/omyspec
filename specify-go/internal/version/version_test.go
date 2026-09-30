package version

import (
	"runtime/debug"
	"testing"
)

func TestDefaultVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("Version is empty")
	}
	// Without ldflags injection, must read from build info or be "dev".
	if Version != "dev" {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			t.Fatalf("Version = %q but ReadBuildInfo unavailable", Version)
		}
		if Version != info.Main.Version {
			t.Errorf("Version = %q, want %q (build info)", Version, info.Main.Version)
		}
	}
}

func TestSet(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()
	Set("1.2.3")
	if Version != "1.2.3" {
		t.Errorf("after Set, Version = %q, want %q", Version, "1.2.3")
	}
}
