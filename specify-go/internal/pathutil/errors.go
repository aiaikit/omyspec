package pathutil

import "errors"

// Sentinel errors returned by Root methods. Callers should use
// errors.Is to classify; fmt.Errorf wrapping at call sites preserves
// the chain.
var (
	ErrPathEscape           = errors.New("pathutil: path escapes root")
	ErrSymlinkComponent     = errors.New("pathutil: symlink in path component")
	ErrAbsolute             = errors.New("pathutil: absolute path not allowed")
	ErrWindowsDriveRelative = errors.New("pathutil: Windows drive-relative path not allowed")
	ErrNotDirectory         = errors.New("pathutil: not a directory")
)
