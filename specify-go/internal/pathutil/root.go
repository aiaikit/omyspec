// Package pathutil enforces project-root containment for all file I/O.
//
// All writes in the CLI MUST go through Root methods, never directly
// through os.WriteFile / os.Create. Phase 1 establishes the constraint;
// Phase 2 (init) and Phase 3+ (integration setup) rely on it.
package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Root wraps os.Root with extra symlink-component rejection.
//
// os.Root (Go 1.24) handles path containment (rejecting ../ and absolute
// paths), but it does NOT reject symlinks in intermediate components.
// This wrapper adds it because the CLI must defend against malicious
// symlinks in user-supplied project directories.
type Root struct {
	abs    string
	osRoot *os.Root
}

// NewRoot validates absPath is an existing directory and returns a Root.
func NewRoot(absPath string) (*Root, error) {
	abs, err := filepath.Abs(absPath)
	if err != nil {
		return nil, fmt.Errorf("pathutil: resolve %q: %w", absPath, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotDirectory, abs)
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrNotDirectory, abs)
	}
	or, err := os.OpenRoot(abs)
	if err != nil {
		return nil, err
	}
	return &Root{abs: abs, osRoot: or}, nil
}

// Abs returns the absolute path of the root.
func (r *Root) Abs() string { return r.abs }

// validate checks the relative path for safety: no absolute, no Windows
// drive-relative, no .. segments, no symlinks in any component.
func (r *Root) validate(rel string) error {
	if rel == "" {
		return fmt.Errorf("%w: empty path", ErrAbsolute)
	}
	if filepath.IsAbs(rel) {
		return fmt.Errorf("%w: %s", ErrAbsolute, rel)
	}
	// Windows drive-relative: "C:foo" parses as relative, must reject.
	if len(rel) >= 2 && rel[1] == ':' {
		return fmt.Errorf("%w: %s", ErrWindowsDriveRelative, rel)
	}
	cleaned := filepath.Clean(rel)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%w: %s", ErrPathEscape, rel)
	}
	// Symlink-component check: walk each component from root.
	cur := r.abs
	parts := strings.Split(cleaned, string(filepath.Separator))
	for _, p := range parts {
		if p == "" || p == "." {
			continue
		}
		next := filepath.Join(cur, p)
		info, err := os.Lstat(next)
		if err != nil {
			if os.IsNotExist(err) {
				// Component doesn't exist yet — that's fine for WriteFile.
				return nil
			}
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: %s", ErrSymlinkComponent, next)
		}
		cur = next
	}
	return nil
}

// Open opens a file relative to the root.
func (r *Root) Open(rel string) (*os.File, error) {
	if err := r.validate(rel); err != nil {
		return nil, err
	}
	return r.osRoot.Open(rel)
}

// WriteFile writes data to a relative path, creating parent dirs.
func (r *Root) WriteFile(rel string, data []byte, perm os.FileMode) error {
	if err := r.validate(rel); err != nil {
		return err
	}
	if dir := filepath.Dir(rel); dir != "." && dir != "" {
		if err := r.osRoot.Mkdir(dir, 0o755); err != nil && !os.IsExist(err) {
			return err
		}
	}
	return r.osRoot.WriteFile(rel, data, perm)
}

// MkdirAll creates a directory hierarchy.
func (r *Root) MkdirAll(rel string, perm os.FileMode) error {
	if err := r.validate(rel); err != nil {
		return err
	}
	return r.osRoot.MkdirAll(rel, perm)
}

// Chmod changes the mode of a file. On Windows, only the read-only bit
// is meaningful; other bits are silently ignored by the OS.
func (r *Root) Chmod(rel string, mode os.FileMode) error {
	if err := r.validate(rel); err != nil {
		return err
	}
	return r.osRoot.Chmod(rel, mode)
}
