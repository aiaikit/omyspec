package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aiaikit/speckit/internal/assets"
	"github.com/aiaikit/speckit/internal/integration"
	"github.com/aiaikit/speckit/internal/render"
	"github.com/aiaikit/speckit/internal/ui"
	"github.com/aiaikit/speckit/internal/version"
	"github.com/spf13/cobra"
)

// InitCmd returns the `specify init` command.
func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Spec Kit project",
		Long:  "Initialize a new Spec Kit project in the current directory or a named path.",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  runInit,
	}
	cmd.Flags().Bool("here", false, "Initialize in the current directory")
	cmd.Flags().Bool("force", false, "Force init even if already initialized")
	cmd.Flags().Bool("non-interactive", false, "Run without prompting")
	cmd.Flags().String("script", "", "Output the init script instead of running it")
	cmd.Flags().String("integration", "", "Integration key to configure (claude, copilot, codex, generic)")
	cmd.Flags().String("integration-options", "", "JSON options for the integration")
	cmd.Flags().String("preset", "", "Preset name to apply")
	cmd.Flags().StringSlice("extension", nil, "Extension names to install")
	cmd.Flags().Bool("ignore-agent-tools", false, "Skip agent tool checks")
	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	flags := parseInitFlags(cmd, args)

	// Determine project path
	var projectPath string
	var createdDir bool
	if flags.here {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd: %w", err)
		}
		projectPath = cwd
		// Check non-empty directory
		if !flags.force {
			entries, err := os.ReadDir(projectPath)
			if err != nil {
				return fmt.Errorf("read current directory: %w", err)
			}
			// Filter out . and ..
			hasEntries := false
			for _, e := range entries {
				if e.Name() != "." && e.Name() != ".." {
					hasEntries = true
					break
				}
			}
			if hasEntries {
				return fmt.Errorf("current directory is not empty (use --force to merge)")
			}
		}
	} else {
		// Project name or path as first arg, default to "."
		name := "."
		if len(args) > 0 {
			name = args[0]
		}
		abs, err := filepath.Abs(name)
		if err != nil {
			return fmt.Errorf("abs: %w", err)
		}
		projectPath = abs
		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			if err := os.MkdirAll(projectPath, 0755); err != nil {
				return fmt.Errorf("create project directory: %w", err)
			}
			createdDir = true
		} else if err != nil {
			return fmt.Errorf("stat project directory: %w", err)
		}
	}

	// Rollback: remove project dir if we created it and an error occurs
	removeOnError := func() {
		if createdDir {
			os.RemoveAll(projectPath)
		}
	}

	// Banner
	banner := ui.Show()
	fmt.Fprintln(cmd.OutOrStdout(), banner)

	// StepTracker
	tr := ui.NewTracker("Initialize Specify Project")
	tr.Add("project", "Initialize project")
	tr.Add("integration", "Configure integration")
	tr.Add("shared-infra", "Install shared infrastructure")
	tr.Add("scripts", "Copy scripts")
	tr.Add("templates", "Copy templates")
	tr.Add("constitution", "Materialize constitution")
	tr.Add("chmod-scripts", "Make scripts executable")
	tr.Add("manifest", "Write manifest")
	tr.Add("workflow", "Install bundled workflow")
	tr.Add("final", "Finalize")
	tr.Mark("project", ui.Running, "")
	fmt.Fprintln(cmd.OutOrStdout(), tr.Render())

	// Resolve integration
	integrationKey := flags.integration
	if integrationKey == "" {
		integrationKey = "claude" // default
	}
	_, err := integration.GetIntegration(integrationKey)
	if err != nil {
		return fmt.Errorf("unknown integration key %q: %w", integrationKey, err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "integration: %s\n", integrationKey)

	// Write integration state
	tr.Mark("integration", ui.Running, integrationKey)
	state := &integration.IntegrationState{
		Version:                version.Version,
		IntegrationStateSchema: integration.IntegrationStateSchema,
		Integration:           integrationKey,
		InstalledIntegrations:  []string{integrationKey},
	}
	if err := integration.WriteState(projectPath, state); err != nil {
		tr.Mark("integration", ui.Error, "failed")
		removeOnError()
		return fmt.Errorf("write integration state: %w", err)
	}
	tr.Mark("integration", ui.Done, integrationKey)

	// Install shared infra: scripts + templates
	tr.Mark("shared-infra", ui.Running, "")
	if err := copyDir(assets.Scripts(), projectPath, ".specify", "scripts"); err != nil {
		tr.Mark("shared-infra", ui.Error, "failed")
		removeOnError()
		return fmt.Errorf("copy scripts: %w", err)
	}
	tr.Mark("scripts", ui.Done, "done")
	if err := copyDir(assets.Templates(), projectPath, ".specify", "templates"); err != nil {
		tr.Mark("shared-infra", ui.Error, "failed")
		removeOnError()
		return fmt.Errorf("copy templates: %w", err)
	}
	tr.Mark("templates", ui.Done, "done")
	tr.Mark("shared-infra", ui.Done, "done")

	// Materialize constitution
	tr.Mark("constitution", ui.Running, "")
	if err := ensureConstitutionFromTemplate(projectPath, tr); err != nil {
		tr.Mark("constitution", ui.Error, "failed")
		removeOnError()
		return fmt.Errorf("ensure constitution: %w", err)
	}

	// chmod scripts
	tr.Mark("chmod-scripts", ui.Running, "")
	if err := ensureExecutableScripts(projectPath, tr); err != nil {
		tr.Mark("chmod-scripts", ui.Error, "failed")
		removeOnError()
		return fmt.Errorf("ensure executable scripts: %w", err)
	}

	// Write manifest
	tr.Mark("manifest", ui.Running, "")
	manifest := render.NewManifest(version.Version)
	if err := render.WriteManifest(projectPath, manifest); err != nil {
		tr.Mark("manifest", ui.Error, "failed")
		removeOnError()
		return fmt.Errorf("write manifest: %w", err)
	}
	tr.Mark("manifest", ui.Done, "done")

	// Write init-options.json
	tr.Mark("final", ui.Running, "")
	opts := initOptions{
		AI:                integrationKey,
		Integration:       integrationKey,
		Script:            detectScriptType(),
		SpeckitVersion:    version.Version,
		Here:              flags.here,
		FeatureNumbering:  true,
	}
	if flags.preset != "" {
		opts.Preset = &flags.preset
	}
	if len(flags.extensions) > 0 {
		opts.Extensions = flags.extensions
	}
	optsData, err := json.MarshalIndent(opts, "", "  ")
	if err != nil {
		tr.Mark("final", ui.Error, "failed")
		removeOnError()
		return fmt.Errorf("marshal init-options: %w", err)
	}
	optsData = append(optsData, '\n')
	optsPath := filepath.Join(projectPath, ".specify", "init-options.json")
	dir := filepath.Dir(optsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		tr.Mark("final", ui.Error, "failed")
		removeOnError()
		return err
	}
	tmp, err := os.CreateTemp(dir, ".init-options.json.tmp.*")
	if err != nil {
		tr.Mark("final", ui.Error, "failed")
		removeOnError()
		return err
	}
	if _, err := tmp.Write(optsData); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		tr.Mark("final", ui.Error, "failed")
		removeOnError()
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		tr.Mark("final", ui.Error, "failed")
		removeOnError()
		return err
	}
	if err := os.Rename(tmp.Name(), optsPath); err != nil {
		os.Remove(tmp.Name())
		tr.Mark("final", ui.Error, "failed")
		removeOnError()
		return err
	}

	// Copy bundled workflow: workflows/speckit/workflow.yml
	tr.Mark("workflow", ui.Running, "")
	wfSrc, err := assets.BundledWorkflow("speckit")
	if err != nil {
		tr.Mark("workflow", ui.Error, "failed")
		removeOnError()
		return fmt.Errorf("bundled workflow speckit: %w", err)
	}
	wfDstDir := filepath.Join(projectPath, ".specify", "workflows", "speckit")
	if err := os.MkdirAll(wfDstDir, 0755); err != nil {
		tr.Mark("workflow", ui.Error, "failed")
		removeOnError()
		return err
	}
	wfData, err := fs.ReadFile(wfSrc, "workflow.yml")
	if err != nil {
		tr.Mark("workflow", ui.Error, "failed")
		removeOnError()
		return fmt.Errorf("read bundled workflow.yml: %w", err)
	}
	wfDstPath := filepath.Join(wfDstDir, "workflow.yml")
	if err := os.WriteFile(wfDstPath, wfData, 0644); err != nil {
		tr.Mark("workflow", ui.Error, "failed")
		removeOnError()
		return err
	}
	tr.Mark("workflow", ui.Done, "done")

	tr.Mark("final", ui.Done, "done")
	fmt.Fprintln(cmd.OutOrStdout(), tr.Render())
	return nil
}

// copyDir copies all files from srcFS into projectPath/rootRelBase/relRoot.
// It preserves directory structure.
func copyDir(srcFS fs.FS, projectPath, rootRelBase, relRoot string) error {
	dstDir := filepath.Join(projectPath, rootRelBase, relRoot)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}
	return fs.WalkDir(srcFS, ".", func(srcPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		srcData, readErr := fs.ReadFile(srcFS, srcPath)
		if readErr != nil {
			return readErr
		}
		dstPath := filepath.Join(dstDir, srcPath)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(dstPath, srcData, 0644)
	})
}

// ensureConstitutionFromTemplate copies constitution-template.md from assets to
// projectPath/.specify/memory/constitution.md if the target doesn't already exist.
// If the template is missing, it reports an error via tracker.
func ensureConstitutionFromTemplate(projectPath string, tracker ui.Tracker) error {
	srcData, err := fs.ReadFile(assets.Templates(), "constitution-template.md")
	if err != nil {
		tracker.Mark("constitution", ui.Error, "template not found")
		return fmt.Errorf("read constitution-template.md: %w", err)
	}
	dir := filepath.Join(projectPath, ".specify", "memory")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir memory dir: %w", err)
	}
	dstPath := filepath.Join(dir, "constitution.md")
	var statErr error
	if info, err := os.Stat(dstPath); err == nil && info.Mode().IsRegular() {
		tracker.Mark("constitution", ui.Skipped, "already exists")
		return nil
	} else {
		statErr = err
	}
	// Permission denied or other non-NotExist error — don't overwrite
	if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
		tracker.Mark("constitution", ui.Error, fmt.Sprintf("cannot check existing file: %v", statErr))
		return nil
	}
	if writeErr := os.WriteFile(dstPath, srcData, 0644); writeErr != nil {
		return fmt.Errorf("write constitution: %w", writeErr)
	}
	tracker.Mark("constitution", ui.Done, "materialized")
	return nil
}

// ensureExecutableScripts chmods +x every .sh file under
// projectPath/.specify/scripts/.
func ensureExecutableScripts(projectPath string, tracker ui.Tracker) error {
	scriptsDir := filepath.Join(projectPath, ".specify", "scripts")
	entries, err := os.ReadDir(scriptsDir)
	if os.IsNotExist(err) {
		tracker.Mark("chmod-scripts", ui.Skipped, "no scripts directory")
		return nil
	}
	if err != nil {
		return fmt.Errorf("read scripts dir: %w", err)
	}
	var made, missing int
	for _, e := range entries {
		if e.IsDir() {
			if err := walkShFiles(filepath.Join(scriptsDir, e.Name()), &made, &missing); err != nil {
				return err
			}
		} else if strings.HasSuffix(e.Name(), ".sh") {
			path := filepath.Join(scriptsDir, e.Name())
			if err := os.Chmod(path, 0755); err != nil {
				missing++
				continue
			}
			made++
		}
	}
	if missing > 0 {
		tracker.Mark("chmod-scripts", ui.Done, fmt.Sprintf("%d made executable, %d failed", made, missing))
	} else {
		tracker.Mark("chmod-scripts", ui.Done, fmt.Sprintf("%d made executable", made))
	}
	return nil
}

// walkShFiles recursively chmods +x all .sh files under dir.
func walkShFiles(dir string, made, missing *int) error {
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".sh") {
			if err := os.Chmod(path, 0755); err != nil {
				(*missing)++
				return nil
			}
			(*made)++
		}
		return nil
	})
}

// detectScriptType returns "sh" on Unix, "ps" on Windows.
func detectScriptType() string {
	return "sh" // ponytail: detect OS at runtime; add windows/powershell support later
}

type initOptions struct {
	AI               string   `json:"ai"`
	Integration      string   `json:"integration"`
	Script           string   `json:"script"`
	SpeckitVersion   string   `json:"speckit_version"`
	Here             bool     `json:"here"`
	FeatureNumbering bool     `json:"feature_numbering"`
	Preset           *string  `json:"preset,omitempty"`
	Extensions       []string `json:"extensions,omitempty"`
	// ponytail: add integration-options, ignore-agent-tools when Phase 3 adds CLI detection
}

type initFlags struct {
	here             bool
	force            bool
	nonInteractive   bool
	script           string
	integration      string
	integrationOpts  string
	preset           string
	extensions       []string
	ignoreAgentTools bool
}

func parseInitFlags(cmd *cobra.Command, args []string) initFlags {
	flags := initFlags{}
	flags.here, _ = cmd.Flags().GetBool("here")
	flags.force, _ = cmd.Flags().GetBool("force")
	flags.nonInteractive, _ = cmd.Flags().GetBool("non-interactive")
	flags.script, _ = cmd.Flags().GetString("script")
	flags.integration, _ = cmd.Flags().GetString("integration")
	flags.integrationOpts, _ = cmd.Flags().GetString("integration-options")
	flags.preset, _ = cmd.Flags().GetString("preset")
	flags.extensions, _ = cmd.Flags().GetStringSlice("extension")
	flags.ignoreAgentTools, _ = cmd.Flags().GetBool("ignore-agent-tools")
	return flags
}
