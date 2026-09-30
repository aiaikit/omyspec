package dlsec

import (
	"strings"
	"testing"
)

func TestReadLimited_UnderCap(t *testing.T) {
	data, truncated, err := ReadLimited(strings.NewReader("hello"), 100)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if truncated {
		t.Error("truncated = true, want false")
	}
	if string(data) != "hello" {
		t.Errorf("got %q, want hello", string(data))
	}
}

func TestReadLimited_AtCap(t *testing.T) {
	data, truncated, err := ReadLimited(strings.NewReader("hello"), 5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if truncated {
		t.Error("truncated = true at exact cap")
	}
	if string(data) != "hello" {
		t.Errorf("got %q, want hello", string(data))
	}
}

func TestReadLimited_OverCap(t *testing.T) {
	data, truncated, err := ReadLimited(strings.NewReader("hello world"), 5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !truncated {
		t.Error("truncated = false, want true")
	}
	if string(data) != "hello" {
		t.Errorf("got %q, want hello (truncated prefix)", string(data))
	}
}
