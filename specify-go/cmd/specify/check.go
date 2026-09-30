package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/aiaikit/speckit/internal/executil"
	"github.com/aiaikit/speckit/internal/ui"
	"github.com/spf13/cobra"
)

// newCheckCmd implements `specify check` — verify AI agent CLIs are
// available. Phase 1 covers the 4 special-agent branches that don't
// follow standard PATH lookup (Claude, Kiro, Rovodev, Docker Agent).
func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check that all required tools are installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := ui.NewConsole(cmd.OutOrStdout(), cmd.ErrOrStderr())
			tr := ui.NewTracker("Check Available Tools")

			tr.Add("claude", "Claude Code")
			if checkClaude() {
				tr.Mark("claude", ui.Done, "available")
			} else {
				tr.Mark("claude", ui.Error, "not found")
			}

			tr.Add("kiro-cli", "Kiro CLI")
			if checkKiroCLI() {
				tr.Mark("kiro-cli", ui.Done, "available")
			} else {
				tr.Mark("kiro-cli", ui.Error, "not found")
			}

			tr.Add("rovodev", "Rovo Dev")
			if checkRovoDev() {
				tr.Mark("rovodev", ui.Done, "available")
			} else {
				tr.Mark("rovodev", ui.Error, "not found")
			}

			tr.Add("docker-agent", "Docker Agent")
			if checkDockerAgent() {
				tr.Mark("docker-agent", ui.Done, "available")
			} else {
				tr.Mark("docker-agent", ui.Error, "not found")
			}

			c.Print(tr.Render())
			return nil
		},
	}
}

// checkClaude looks at ~/.claude/local/{claude, node_modules/.bin/claude}.
// Mirrors Python `_utils.check_tool` "claude" special case.
func checkClaude() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	candidates := []string{
		filepath.Join(home, ".claude", "local", "claude"),
		filepath.Join(home, ".claude", "local", "node_modules", ".bin", "claude"),
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

// checkKiroCLI checks for `kiro-cli` or `kiro` on PATH.
func checkKiroCLI() bool {
	if _, err := executil.LookPath("kiro-cli"); err == nil {
		return true
	}
	if _, err := executil.LookPath("kiro"); err == nil {
		return true
	}
	return false
}

// checkRovoDev checks for `acli` (Atlassian CLI), which is what
// Rovodev ships as a plugin.
func checkRovoDev() bool {
	_, err := executil.LookPath("acli")
	return err == nil
}

// checkDockerAgent runs `docker agent version` with a 5s timeout and
// returns true on exit 0.
func checkDockerAgent() bool {
	res, err := executil.Run("docker", []string{"agent", "version"}, executil.RunOpts{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return false
	}
	return res.ExitCode == 0
}
