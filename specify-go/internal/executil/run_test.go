package executil

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRun_Success(t *testing.T) {
	res, err := Run("echo", []string{"hello"}, RunOpts{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", res.ExitCode)
	}
	if !strings.Contains(string(res.Stdout), "hello") {
		t.Errorf("Stdout = %q, want 'hello'", string(res.Stdout))
	}
}

func TestRun_Failure(t *testing.T) {
	res, err := Run("false", nil, RunOpts{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.ExitCode == 0 {
		t.Errorf("ExitCode = 0, want non-zero")
	}
}

func TestRun_Timeout(t *testing.T) {
	res, err := Run("sleep", []string{"5"}, RunOpts{Timeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !res.Killed {
		t.Errorf("Killed = false, want true (timed out)")
	}
}

func TestRun_EnvScrub(t *testing.T) {
	var buf bytes.Buffer
	_, err := Run("env", nil, RunOpts{
		Env:    []string{"GH_TOKEN=secret", "FOO=bar"},
		Stdout: &buf,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "GH_TOKEN=secret") {
		t.Errorf("env contained GH_TOKEN after scrub:\n%s", out)
	}
	if !strings.Contains(out, "FOO=bar") {
		t.Errorf("env missing FOO=bar:\n%s", out)
	}
}

func TestRun_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	res, err := Run("sleep", []string{"5"}, RunOpts{Context: ctx})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !res.Killed {
		t.Errorf("Killed = false on context cancel")
	}
}
