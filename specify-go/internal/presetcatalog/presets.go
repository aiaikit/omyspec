package presetcatalog

import (
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/aiaikit/speckit/internal/assets"
	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("preset not found")

type Preset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Bundled     bool   `json:"bundled"`
}

type presetYAML struct {
	Preset struct {
		ID          string `yaml:"id"`
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
	} `yaml:"preset"`
}

func ListBundledPresets() []*Preset {
	presetsFS := assets.Presets()
	entries, err := fs.ReadDir(presetsFS, ".")
	if err != nil {
		return nil
	}
	var out []*Preset
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		p, err := loadPresetFromFS(presetsFS, id)
		if err != nil {
			continue
		}
		p.Bundled = true
		out = append(out, p)
	}
	return out
}

func GetPreset(id string) (*Preset, error) {
	presetsFS := assets.Presets()
	p, err := loadPresetFromFS(presetsFS, id)
	if err != nil {
		return nil, err
	}
	p.Bundled = true
	return p, nil
}

func loadPresetFromFS(presetsFS fs.FS, id string) (*Preset, error) {
	yamlPath := filepath.Join(id, "preset.yml")
	data, err := fs.ReadFile(presetsFS, yamlPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var py presetYAML
	if err := yaml.Unmarshal(data, &py); err != nil {
		return nil, err
	}
	return &Preset{
		ID:          py.Preset.ID,
		Name:        py.Preset.Name,
		Description: py.Preset.Description,
		Version:     py.Preset.Version,
	}, nil
}
