package integration

import (
	"errors"
	"testing"
)

func TestGetIntegration_Exists(t *testing.T) {
	cfg, err := GetIntegration("claude")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Key != "claude" {
		t.Errorf("expected Key claude, got %s", cfg.Key)
	}
	if cfg.Name != "Claude Code" {
		t.Errorf("expected Name 'Claude Code', got %s", cfg.Name)
	}
}

func TestGetIntegration_Unknown(t *testing.T) {
	_, err := GetIntegration("nonexistent")
	if !errors.Is(err, ErrUnknownIntegration) {
		t.Errorf("expected ErrUnknownIntegration, got %v", err)
	}
}
