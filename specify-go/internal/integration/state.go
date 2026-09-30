package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	IntegrationStateSchema = 1
	IntegrationJSONPath    = ".specify/integration.json"
)

var (
	ErrStateRead  = errors.New("failed to read integration state")
	ErrStateWrite = errors.New("failed to write integration state")
)

type IntegrationSettings struct {
	Script          string         `json:"script,omitempty"`
	RawOptions      string         `json:"raw_options,omitempty"`
	ParsedOptions   map[string]any `json:"parsed_options,omitempty"`
	InvokeSeparator string         `json:"invoke_separator,omitempty"`
}

type IntegrationState struct {
	Version                string                     `json:"version"`
	IntegrationStateSchema int                        `json:"integration_state_schema"`
	Integration           string                     `json:"integration,omitempty"`
	DefaultIntegration    string                     `json:"default_integration,omitempty"`
	InstalledIntegrations []string                   `json:"installed_integrations"`
	IntegrationSettings   map[string]IntegrationSettings `json:"integration_settings,omitempty"`
}

// ReadState reads .specify/integration.json from project root.
// Returns nil,nil if the file does not exist.
func ReadState(projectRoot string) (*IntegrationState, error) {
	if projectRoot == "" {
		return nil, ErrStateRead
	}
	path := filepath.Join(projectRoot, IntegrationJSONPath)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStateRead, err)
	}
	var st IntegrationState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStateRead, err)
	}
	return &st, nil
}

// WriteState writes .specify/integration.json atomically.
func WriteState(projectRoot string, st *IntegrationState) error {
	if projectRoot == "" {
		return ErrStateWrite
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStateWrite, err)
	}
	path := filepath.Join(projectRoot, IntegrationJSONPath)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: %v", ErrStateWrite, err)
	}
	tmp, err := os.CreateTemp(dir, ".integration.json.tmp.*")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStateWrite, err)
	}
	_, err = tmp.Write(append(data, '\n'))
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("%w: %v", ErrStateWrite, err)
	}
	if renameErr := os.Rename(tmp.Name(), path); renameErr != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("%w: %v", ErrStateWrite, renameErr)
	}
	return nil
}
