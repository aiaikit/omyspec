package pathutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewRoot_AcceptsExistingDir(t *testing.T) {
	dir := t.TempDir()
	r, err := NewRoot(dir)
	if err != nil {
		t.Fatalf("NewRoot: %v", err)
	}
	if r == nil {
		t.Fatal("Root is nil")
	}
}

func TestNewRoot_RejectsNonExistent(t *testing.T) {
	_, err := NewRoot(filepath.Join(t.TempDir(), "nope"))
	if !errors.Is(err, ErrNotDirectory) {
		t.Errorf("err = %v, want ErrNotDirectory", err)
	}
}

func TestNewRoot_RejectsFileInsteadOfAFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := NewRoot(f)
	if !errors.Is(err, ErrNotDirectory) {
		t.Errorf("err = %v, want ErrNotDirectory", err)
	}
}

func TestRoot_WriteFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	r, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.WriteFile("a.txt", []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q, want hello", string(got))
	}
}

func TestRoot_WriteFileRejectsAbsolute(t *testing.T) {
	dir := t.TempDir()
	r, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = r.WriteFile("/etc/passwd", []byte("x"), 0o644)
	if !errors.Is(err, ErrAbsolute) {
		t.Errorf("err = %v, want ErrAbsolute", err)
	}
}

func TestRoot_WriteFileRejectsDotDot(t *testing.T) {
	dir := t.TempDir()
	r, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = r.WriteFile("../escape.txt", []byte("x"), 0o644)
	if err == nil {
		t.Error("expected error on ../, got nil")
	}
}

func TestRoot_MkdirAllNested(t *testing.T) {
	dir := t.TempDir()
	r, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.MkdirAll("a/b/c", 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a", "b", "c")); err != nil {
		t.Errorf("stat: %v", err)
	}
}

func TestRoot_RejectsSymlinkComponent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows; tested separately")
	}
	dir := t.TempDir()
	linkTarget := t.TempDir()
	linkPath := filepath.Join(dir, "lnk")
	if err := os.Symlink(linkTarget, linkPath); err != nil {
		t.Fatal(err)
	}
	r, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = r.WriteFile("lnk/file.txt", []byte("x"), 0o644)
	if !errors.Is(err, ErrSymlinkComponent) {
		t.Errorf("err = %v, want ErrSymlinkComponent", err)
	}
}
