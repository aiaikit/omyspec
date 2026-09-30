package projectstate

import (
	"errors"
	"os"
	"testing"
)

func TestResolveProjectRoot_CWDNotProject(t *testing.T) {
	tmp, _ := os.MkdirTemp("", "notproject")
	defer os.RemoveAll(tmp)
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(tmp)
	os.Unsetenv("SPECIFY_INIT_DIR")
	_, err := ResolveProjectRoot()
	if !errors.Is(err, ErrNotAProject) {
		t.Errorf("got %v, want ErrNotAProject", err)
	}
}

func TestResolveProjectRoot_WithDotSpecify(t *testing.T) {
	tmp, _ := os.MkdirTemp("", "project")
	defer os.RemoveAll(tmp)
	os.MkdirAll(tmp+"/.specify", 0755)
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(tmp)
	root, err := ResolveProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root != tmp {
		t.Errorf("got %q, want %q", root, tmp)
	}
}

func TestResolveProjectRoot_SPECIFY_INIT_DIR(t *testing.T) {
	tmp, _ := os.MkdirTemp("", "envproject")
	defer os.RemoveAll(tmp)
	os.MkdirAll(tmp+"/.specify", 0755)
	os.Setenv("SPECIFY_INIT_DIR", tmp)
	defer os.Unsetenv("SPECIFY_INIT_DIR")
	root, err := ResolveProjectRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root != tmp {
		t.Errorf("got %q, want %q", root, tmp)
	}
}

func TestResolveProjectRoot_SPECIFY_INIT_DIR_NotProject(t *testing.T) {
	tmp, _ := os.MkdirTemp("", "notenvproject")
	defer os.RemoveAll(tmp)
	os.Setenv("SPECIFY_INIT_DIR", tmp)
	defer os.Unsetenv("SPECIFY_INIT_DIR")
	_, err := ResolveProjectRoot()
	if !errors.Is(err, ErrNotAProject) {
		t.Errorf("got %v, want ErrNotAProject", err)
	}
}
