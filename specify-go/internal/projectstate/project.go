package projectstate

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNotAProject = errors.New("not a Spec Kit project (no .specify/ directory)")

// ResolveProjectRoot returns the project root.
// Checks SPECIFY_INIT_DIR env var first, then cwd for .specify/.
func ResolveProjectRoot() (string, error) {
	if dir := os.Getenv("SPECIFY_INIT_DIR"); dir != "" {
		root := filepath.Join(dir, ".specify")
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(dir)
			return abs, nil
		}
		return "", ErrNotAProject
	}
	cwd, _ := os.Getwd()
	if _, err := os.Stat(filepath.Join(cwd, ".specify")); os.IsNotExist(err) {
		return "", ErrNotAProject
	} else if err != nil {
		return "", err
	}
	return cwd, nil
}
