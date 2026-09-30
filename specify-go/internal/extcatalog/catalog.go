package extcatalog

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed catalog.json
var embeddedCatalogJSON []byte

var ErrNotFound = errors.New("extension not found")

type Extension struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Repository  string   `json:"repository"`
	Tags        []string `json:"tags"`
	Bundled     bool     `json:"bundled"`
}

type Catalog struct {
	SchemaVersion string              `json:"schema_version"`
	Extensions    map[string]Extension `json:"extensions"`
}

func LoadCatalog() (*Catalog, error) {
	var cat Catalog
	if err := json.Unmarshal(embeddedCatalogJSON, &cat); err != nil {
		return nil, err
	}
	return &cat, nil
}

func GetExtension(id string) (*Extension, error) {
	cat, err := LoadCatalog()
	if err != nil {
		return nil, err
	}
	ext, ok := cat.Extensions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &ext, nil
}

func ListBundledExtensions() []*Extension {
	cat, _ := LoadCatalog()
	var out []*Extension
	for _, ext := range cat.Extensions {
		if ext.Bundled {
			copy := ext
			out = append(out, &copy)
		}
	}
	return out
}

// ponytail: ListLocalExtensions reads .specify/extensions/<id>/extension.yml for each subdir.
// yaml struct matches Extension (ID, Name, Version, Description, Author, Repository, Tags, Bundled)
func ListLocalExtensions(projectRoot string) []*Extension {
	dir := filepath.Join(projectRoot, ".specify", "extensions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []*Extension
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		ymlPath := filepath.Join(dir, id, "extension.yml")
		data, err := os.ReadFile(ymlPath)
		if err != nil {
			continue
		}
		var ext Extension
		if err := yaml.Unmarshal(data, &ext); err != nil {
			continue
		}
		ext.ID = id
		out = append(out, &ext)
	}
	return out
}
