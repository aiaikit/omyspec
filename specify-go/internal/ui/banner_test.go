package ui

import (
	"strings"
	"testing"
)

func TestShowContainsSpecKit(t *testing.T) {
	s := Show()
	if !strings.Contains(s, "Spec Kit") {
		t.Errorf("banner missing 'Spec Kit':\n%s", s)
	}
}

func TestShowNonEmpty(t *testing.T) {
	if Show() == "" {
		t.Error("banner is empty")
	}
}