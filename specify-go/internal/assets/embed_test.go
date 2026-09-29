package assets

import (
	"io/fs"
	"testing"
)

func TestRootFSContainsTemplates(t *testing.T) {
	if _, err := fs.ReadFile(rootFS, "templates/spec-template.md"); err != nil {
		t.Errorf("read templates/spec-template.md: %v", err)
	}
}

func TestRootFSContainsExtensions(t *testing.T) {
	if _, err := fs.ReadFile(rootFS, "extensions/git/extension.yml"); err != nil {
		t.Errorf("read extensions/git/extension.yml: %v", err)
	}
}

func TestTemplatesSubFS(t *testing.T) {
	if _, err := fs.ReadFile(Templates(), "spec-template.md"); err != nil {
		t.Errorf("read templates/spec-template: %v", err)
	}
}

func TestExtensionsSubFS(t *testing.T) {
	if _, err := fs.ReadFile(Extensions(), "git/extension.yml"); err != nil {
		t.Errorf("read extensions/git/extension.yml: %v", err)
	}
}
