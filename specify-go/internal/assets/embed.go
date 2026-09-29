// Package assets embeds the Spec Kit asset tree at compile time.
//
// The embed layout MUST mirror src/specify_cli/core_pack/ in the Python
// reference implementation. Phase 2 (init) depends on this byte equality.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:templates
//go:embed all:scripts
//go:embed all:extensions
//go:embed all:workflows
//go:embed all:presets
var rootFS embed.FS

// Templates returns a fs.FS rooted at the templates/ directory.
func Templates() fs.FS {
	sub, _ := fs.Sub(rootFS, "templates")
	return sub
}

// Scripts returns a fs.FS rooted at the scripts/ directory.
func Scripts() fs.FS {
	sub, _ := fs.Sub(rootFS, "scripts")
	return sub
}

// Extensions returns a fs.FS rooted at the extensions/ directory.
func Extensions() fs.FS {
	sub, _ := fs.Sub(rootFS, "extensions")
	return sub
}

// Workflows returns a fs.FS rooted at the workflows/ directory.
func Workflows() fs.FS {
	sub, _ := fs.Sub(rootFS, "workflows")
	return sub
}

// Presets returns a fs.FS rooted at the presets/ directory.
func Presets() fs.FS {
	sub, _ := fs.Sub(rootFS, "presets")
	return sub
}
