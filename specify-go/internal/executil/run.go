// Package executil is the argv-only subprocess wrapper.
//
// IMPORTANT: Run's signature intentionally has NO `shell bool` parameter.
// This makes shell injection impossible at compile time. Any caller wanting
// to run a shell pipeline must construct argv explicitly.
package executil

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"time"
)

// RunOpts configures Run. Zero value runs `name args` with no env
// additions, no timeout, no stdin, output to /dev/null, env scrubbed.
type RunOpts struct {
	Dir      string
	Env      []string      // appended to (scrubbed) os.Environ()
	Timeout  time.Duration // 0 = no timeout
	Stdin    io.Reader
	Stdout   io.Writer       // default: io.Discard
	Stderr   io.Writer       // default: io.Discard
	ScrubEnv *bool           // default: true (set false to disable)
	Context  context.Context // optional; Timeout wins if both set
}

func (o RunOpts) scrub() bool {
	if o.ScrubEnv == nil {
		return true
	}
	return *o.ScrubEnv
}

// Result is the outcome of Run. Killed is true if Timeout fired or
// Context was canceled before the process exited naturally.
type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Killed   bool
}

// Run executes name with args, capturing output and enforcing the timeout.
//
// Returns *Result with ExitCode set even on non-zero exit; only returns
// error for process-spawn failures (e.g., binary not found) or context
// cancellation. This mirrors Python's subprocess.run semantics.
func Run(name string, args []string, opts RunOpts) (*Result, error) {
	// Build env: start from scratch so we don't accidentally pull in GH creds.
	env := make([]string, 0, len(os.Environ())+len(opts.Env))
	if opts.scrub() {
		env = append(env, ScrubEnv(os.Environ())...)
	} else {
		env = append(env, os.Environ()...)
	}
	if opts.scrub() {
		env = append(env, ScrubEnv(opts.Env)...)
	} else {
		env = append(env, opts.Env...)
	}

	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	var stdoutBuf, stderrBuf bytes.Buffer

	ctx := opts.Context
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), opts.Timeout)
	}

	var cmd *exec.Cmd
	if ctx != nil {
		cmd = exec.CommandContext(ctx, name, args...)
	} else {
		cmd = exec.Command(name, args...)
	}
	cmd.Dir = opts.Dir
	cmd.Env = env
	cmd.Stdout = io.MultiWriter(stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(stderr, &stderrBuf)
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}

	if cancel != nil {
		defer cancel()
	}

	err := cmd.Run()
	res := &Result{
		Stdout: stdoutBuf.Bytes(),
		Stderr: stderrBuf.Bytes(),
	}
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			res.ExitCode = ee.ExitCode()
			// ExitError due to context cancel or timeout
			if ctx != nil && ctx.Err() != nil {
				res.Killed = true
			}
			return res, nil
		}
		return nil, err
	}
	res.ExitCode = 0
	return res, nil
}
