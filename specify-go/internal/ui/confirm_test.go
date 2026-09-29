package ui

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestConfirm_Yes(t *testing.T) {
	got, err := Confirm("continue?", false, strings.NewReader("y\n"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !got {
		t.Error("got false, want true")
	}
}

func TestConfirm_No(t *testing.T) {
	got, err := Confirm("continue?", true, strings.NewReader("n\n"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got {
		t.Error("got true, want false")
	}
}

func TestConfirm_DefaultYes(t *testing.T) {
	got, err := Confirm("continue?", true, strings.NewReader("\n"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !got {
		t.Error("got false, want default true")
	}
}

func TestConfirm_DefaultNo(t *testing.T) {
	got, err := Confirm("continue?", false, strings.NewReader("\n"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got {
		t.Error("got true, want default false")
	}
}

func TestConfirm_EOFReturnsErr(t *testing.T) {
	_, err := Confirm("continue?", false, strings.NewReader(""))
	if err == nil {
		t.Error("expected error on EOF, got nil")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("err = %v, want io.EOF", err)
	}
}
