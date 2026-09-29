package assets

import (
	"fmt"
	"io/fs"
	"regexp"
)

var (
	extIDRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
	wfIDRegex  = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)
	presetIDRe = extIDRegex
)

// BundledExtension returns a fs.FS rooted at extensions/<id>/, validating
// the id matches Python's `^[a-z0-9-]+$` and the manifest is bundled.
func BundledExtension(id string) (fs.FS, error) {
	if !extIDRegex.MatchString(id) {
		return nil, fmt.Errorf("invalid extension id %q", id)
	}
	sub, err := fs.Sub(Extensions(), id)
	if err != nil {
		return nil, fmt.Errorf("extension %q not bundled", id)
	}
	if _, err := fs.ReadFile(sub, "extension.yml"); err != nil {
		return nil, fmt.Errorf("extension %q has no extension.yml", id)
	}
	return sub, nil
}

// BundledWorkflow returns a fs.FS rooted at workflows/<id>/, validating
// the id matches Python's `^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`.
func BundledWorkflow(id string) (fs.FS, error) {
	if !wfIDRegex.MatchString(id) {
		return nil, fmt.Errorf("invalid workflow id %q", id)
	}
	sub, err := fs.Sub(Workflows(), id)
	if err != nil {
		return nil, fmt.Errorf("workflow %q not bundled", id)
	}
	if _, err := fs.ReadFile(sub, "workflow.yml"); err != nil {
		return nil, fmt.Errorf("workflow %q has no workflow.yml", id)
	}
	return sub, nil
}

// BundledPreset returns a fs.FS rooted at presets/<id>/.
func BundledPreset(id string) (fs.FS, error) {
	if !presetIDRe.MatchString(id) {
		return nil, fmt.Errorf("invalid preset id %q", id)
	}
	sub, err := fs.Sub(Presets(), id)
	if err != nil {
		return nil, fmt.Errorf("preset %q not bundled", id)
	}
	if _, err := fs.ReadFile(sub, "preset.yml"); err != nil {
		return nil, fmt.Errorf("preset %q has no preset.yml", id)
	}
	return sub, nil
}
