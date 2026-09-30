# Phase 2 Implementation Plan — Root Commands

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `specify init`, `specify check` (full), and `specify version` subcommand.

**Architecture:** Write `.specify/` structure using `pathutil` for atomic writes. Read integration.json in `specify check`. `specify version` shows install source and optional upgrade notice.

**Tech Stack:** Go 1.24, cobra, lipgloss, `gopkg.in/yaml.v3`, Phase 1 packages (pathutil, executil, assets, ui, version, dlsec).

---

## Global Constraints

- `go 1.24` in `go.mod` (Phase 1 constraint)
- `module github.com/aiaikit/speckit`
- `executil.Run` signature: no shell parameter (invariant)
- `pathutil.Root` is the only write entry (Phase 1 invariant)
- All writes: `pathutil.Root.WriteFile` (atomic via os.CreateTemp + os.Rename)
- `schema_version: "1.0"` in `speckit.manifest.json`
- `integration_state_schema: 1` in `integration.json`
- `SPECIFY_INIT_DIR` env var overrides cwd

---

## Task 1: internal/projectstate/

**Files:**
- Create: `specify-go/internal/projectstate/project.go`
- Test: `specify-go/internal/projectstate/project_test.go`

**Interfaces:**
- Produces: `ResolveProjectRoot() (string, error)`, `ErrNotAProject`

- [ ] **Step 1: Write the failing test**

```go
// project_test.go
func TestResolveProjectRoot_CWDNotProject(t *testing.T) {
    // Create temp dir without .specify/
    tmp, _ := os.MkdirTemp("", "notproject")
    defer os.RemoveAll(tmp)
    orig, _ := os.Getwd()
    defer os.Chdir(orig)
    os.Chdir(tmp)
    os.Unsetenv("SPECIFY_INIT_DIR")
    _, err := projectstate.ResolveProjectRoot()
    if !errors.Is(err, projectstate.ErrNotAProject) {
        t.Errorf("got %v, want ErrNotAProject", err)
    }
}

func TestResolveProjectRoot_WithDotSpecify(t *testing.T) {
    tmp, _ := os.MkdirTemp("", "project")
    defer os.RemoveAll(tmp)
    os.MkdirAll(tmp+"/.specify", 0755)
    orig, _ := os.Getwd()
    defer os.Chdir(orig)
    os.Chdir(tmp)
    root, err := projectstate.ResolveProjectRoot()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if root != tmp {
        t.Errorf("got %q, want %q", root, tmp)
    }
}

func TestResolveProjectRoot_SPECIFY_INIT_DIR(t *testing.T) {
    tmp, _ := os.MkdirTemp("", "envproject")
    defer os.RemoveAll(tmp)
    os.MkdirAll(tmp+"/.specify", 0755)
    os.Setenv("SPECIFY_INIT_DIR", tmp)
    defer os.Unsetenv("SPECIFY_INIT_DIR")
    root, err := projectstate.ResolveProjectRoot()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if root != tmp {
        t.Errorf("got %q, want %q", root, tmp)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/projectstate/... -v`
Expected: FAIL — "cannot find package"

- [ ] **Step 3: Write minimal implementation**

```go
// project.go
package projectstate

import (
    "errors"
    "os"
    "path/filepath"
)

var ErrNotAProject = errors.New("not a Spec Kit project (no .specify/ directory)")

// ResolveProjectRoot returns the project root.
// Checks SPECIFY_INIT_DIR first, then cwd for .specify/.
func ResolveProjectRoot() (string, error) {
    if dir := os.Getenv("SPECIFY_INIT_DIR"); dir != "" {
        root := filepath.Join(dir, ".specify")
        if info, err := os.Stat(root); err == nil && info.IsDir() {
            abs, _ := filepath.Abs(dir)
            return abs, nil
        }
        return "", ErrNotAProject
    }
    cwd, _ := os.Getwd()
    if _, err := os.Stat(filepath.Join(cwd, ".specify")); os.IsNotExist(err) {
        return "", ErrNotAProject
    } else if err != nil {
        return "", err
    }
    return cwd, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/projectstate/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add specify-go/internal/projectstate/
git commit -m "feat(go): add internal/projectstate with ResolveProjectRoot

Phase 2 task 1.

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

## Task 2: internal/integration/state.go

**Files:**
- Create: `specify-go/internal/integration/state.go`
- Create: `specify-go/internal/integration/state_test.go`

**Interfaces:**
- Consumes: `projectstate.ResolveProjectRoot`
- Produces: `ReadState(projectRoot string) (*IntegrationState, error)`, `WriteState(projectRoot string, state *IntegrationState) error`

```go
// state.go
package integration

import (
    "encoding/json"
    "errors"
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
    Version                string                           `json:"version"`
    IntegrationStateSchema int                              `json:"integration_state_schema"`
    Integration           string                           `json:"integration,omitempty"`
    DefaultIntegration    string                           `json:"default_integration,omitempty"`
    InstalledIntegrations []string                         `json:"installed_integrations"`
    IntegrationSettings   map[string]IntegrationSettings   `json:"integration_settings,omitempty"`
}

// ReadState reads .specify/integration.json. Returns nil,nil if absent.
func ReadState(projectRoot string) (*IntegrationState, error) {
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

// WriteState atomically writes .specify/integration.json.
func WriteState(projectRoot string, st *IntegrationState) error {
    data, err := json.MarshalIndent(st, "", "  ")
    if err != nil {
        return fmt.Errorf("%w: %v", ErrStateWrite, err)
    }
    path := filepath.Join(projectRoot, IntegrationJSONPath)
    tmp, err := os.CreateTemp(filepath.Dir(path), ".integration.json.tmp.*")
    if err != nil {
        return fmt.Errorf("%w: %v", ErrStateWrite, err)
    }
    defer os.Remove(tmp.Name())
    if _, err := tmp.Write(append(data, '\n')); err != nil {
        tmp.Close()
        return fmt.Errorf("%w: %v", ErrStateWrite, err)
    }
    tmp.Close()
    if err := os.Rename(tmp.Name(), path); err != nil {
        return fmt.Errorf("%w: %v", ErrStateWrite, err)
    }
    return nil
}
```

- [ ] Write `state_test.go` with 5+ cases: round-trip, missing file, invalid JSON, legacy format, normalize
- [ ] Run tests, fix, commit

---

## Task 3: internal/integration/INTEGRATION_REGISTRY.go

**Files:**
- Create: `specify-go/internal/integration/registry.go`
- Test: `specify-go/internal/integration/registry_test.go`

**Interfaces:**
- Produces: `INTEGRATION_REGISTRY map[string]IntegrationConfig`, `GetIntegration(key string) IntegrationConfig`, `ErrUnknownIntegration`

```go
// registry.go
package integration

import "errors"

var ErrUnknownIntegration = errors.New("unknown integration")

type IntegrationConfig struct {
    Key         string
    Name        string
    RequiresCLI bool
    InstallURL  string
}

var INTEGRATION_REGISTRY = map[string]IntegrationConfig{
    "claude":   {Key: "claude", Name: "Claude Code", RequiresCLI: true, InstallURL: "https://docs.anthropic.com/en/docs/claude-code"},
    "copilot":  {Key: "copilot", Name: "GitHub Copilot", RequiresCLI: false, InstallURL: ""},
    "codex":    {Key: "codex", Name: "OpenAI Codex", RequiresCLI: true, InstallURL: "https://openai.com/index/openai-codex/"},
    "generic":  {Key: "generic", Name: "Generic Agent", RequiresCLI: false, InstallURL: ""},
}

func GetIntegration(key string) (IntegrationConfig, error) {
    cfg, ok := INTEGRATION_REGISTRY[key]
    if !ok {
        return IntegrationConfig{}, ErrUnknownIntegration
    }
    return cfg, nil
}
```

- [ ] Write registry_test.go: GetIntegration exists, GetIntegration unknown
- [ ] Run tests, commit

---

## Task 4: cmd/specify/init.go (skeleton)

**Files:**
- Create: `specify-go/cmd/specify/init.go`
- Modify: `specify-go/cmd/specify/root.go` — register InitCmd

**Interfaces:**
- Consumes: `projectstate.ResolveProjectRoot`, `ui.Show`, `ui.NewTracker`
- Produces: `InitCmd() *cobra.Command`

- [ ] Write `init.go` with flag definitions (--here, --force, --non-interactive, --script, --integration, --integration-options, --preset, --extension, --ignore-agent-tools)
- [ ] In `root.go`, add `cmd.AddCommand(InitCmd())`
- [ ] `InitCmd.RunE`: parse flags → resolve project root → show banner → basic error check (unknown integration key → error)
- [ ] Write init_test.go: InitCmd exists, flags registered
- [ ] `go build ./cmd/specify` compiles
- [ ] Commit

---

## Task 5: specify init — write .specify/ structure

**Files:**
- Modify: `specify-go/cmd/specify/init.go`
- Create: `specify-go/internal/render/manifest.go`

**Interfaces:**
- Consumes: `assets.Templates()`, `assets.Scripts()`, `pathutil.NewRoot`
- Produces: `render.WriteManifest(projectRoot string, manifest *Manifest) error`

```go
// render/manifest.go
package render

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

type Manifest struct {
    SchemaVersion   string `json:"schema_version"`
    SpeckitVersion  string `json:"speckit_version"`
    GeneratedAt     string `json:"generated_at"`
}

func WriteManifest(projectRoot string, m *Manifest) error {
    path := filepath.Join(projectRoot, ".specify", "speckit.manifest.json")
    f, err := os.CreateTemp(filepath.Dir(path), ".manifest.json.tmp.*")
    if err != nil {
        return err
    }
    defer os.Remove(f.Name())
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    if err := enc.Encode(m); err != nil {
        f.Close()
        return err
    }
    f.Close()
    return os.Rename(f.Name(), path)
}
```

**init flow to implement in init.go RunE:**
1. Parse args → projectPath
2. Create StepTracker "Initialize Specify Project"
3. Tracker steps: precheck, ai-select, script-select, integration, shared-infra, scripts, constitution, workflow, final
4. `integration.WriteState(projectPath, &IntegrationState{Version: version.Version, IntegrationStateSchema: 1, Integration: integrationKey, InstalledIntegrations: [integrationKey]})`
5. Copy scripts from `assets.Scripts()` to `projectPath/.specify/scripts/` using `pathutil`
6. Copy templates from `assets.Templates()` to `projectPath/.specify/templates/`
7. Write `speckit.manifest.json` via `render.WriteManifest`
8. Write `init-options.json` via atomic write
9. Copy bundled workflow: `assets.Workflows()` → `projectPath/.specify/workflows/speckit/workflow.yml`
10. Ensure scripts executable
11. On error: if project dir was newly created, remove it (rollback)

- [ ] Write render/manifest.go + manifest_test.go
- [ ] Extend init.go to write all .specify/ files
- [ ] Write init_test.go: full integration test using temp dir
- [ ] All tests pass
- [ ] Commit

---

## Task 6: specify init — constitution materialize + scripts executable

**Files:**
- Modify: `specify-go/cmd/specify/init.go`

**Interfaces:**
- Consumes: `assets.Templates()`, `internal/projectstate`

- [ ] `ensureConstitutionFromTemplate`: read `assets.Templates()` → `templates/constitution-template.md`; if exists, copy to `projectPath/.specify/memory/constitution.md`; if already exists, skip
- [ ] `ensureExecutableScripts`: for each `.sh` file in `projectPath/.specify/scripts/`, chmod +x
- [ ] Add constitution + chmod steps to tracker
- [ ] Tests pass, commit

---

## Task 7: specify check (full version)

**Files:**
- Create: `specify-go/internal/integration/check.go`
- Create: `specify-go/internal/integration/check_test.go`
- Modify: `specify-go/cmd/specify/check.go`

**Interfaces:**
- Consumes: `integration.ReadState`, `integration.GetIntegration`, `executil.Run`, `ui.NewTracker`
- Produces: `CheckTools(projectRoot string, tracker *ui.Tracker) error`

```go
// check.go
package integration

// CheckTools checks all installed integrations from integration.json.
// Uses tracker to report each tool's status.
// Returns error only on catastrophic failure (not on missing tools).
func CheckTools(projectRoot string, tracker *ui.Tracker) error {
    state, err := ReadState(projectRoot)
    if err != nil {
        return err
    }
    if state == nil {
        // No integration.json — not an error, just nothing to check
        return nil
    }
    for _, key := range state.InstalledIntegrations {
        cfg, err := GetIntegration(key)
        if err != nil {
            tracker.Add(key, key)
            tracker.Error(key, "unknown integration")
            continue
        }
        tracker.Add(key, cfg.Name)
        if !cfg.RequiresCLI {
            tracker.Skip(key, "IDE-based")
            continue
        }
        // Check if tool exists via executil.LookPath or executil.Run
        found, _ := executil.LookPath(key)
        if found {
            tracker.Complete(key, "found")
        } else {
            tracker.Error(key, "not found")
        }
    }
    return nil
}
```

- [ ] Rewrite check.go to use CheckTools
- [ ] check.go: if projectstate.ErrNotAProject, print "Run 'specify init' first" and exit 0
- [ ] Write check_test.go: no integration.json → skip; invalid integration → error
- [ ] Tests pass, commit

---

## Task 8: cmd/specify/version.go (full version)

**Files:**
- Modify: `specify-go/cmd/specify/version.go`

**Interfaces:**
- Consumes: `version.DetectInstallMethod`, `version.InstallMethod.String()`, `version.FetchLatestRelease`, `dlsec.NewClient`
- Produces: VersionCmd with --json flag

```
specify version
# Output (no --json):
specify version {Version}
Installed via: {install method}
Latest release: {tag} ({up-to-date|upgrade available})

# With --json:
{"version":"{Version}","installMethod":"{method}","latestRelease":"{tag}"}
```

- [ ] Rewrite version.go to implement full subcommand
- [ ] Write version_test.go: JSON flag output
- [ ] Tests pass, commit

---

## Task 9: Integration tests

**Files:**
- Modify: `specify-go/cmd/specify/integration_test.go` (add new tests)
- Create: `specify-go/cmd/specify/init_integration_test.go` (build tag: `integration`)

**Build tag:** `//go:build integration` (separate from Phase 1's `python_integration`)

```go
//go:build integration

func TestInit_CreatesDotSpecify(t *testing.T) {
    tmp, _ := os.MkdirTemp("", "init-test")
    defer os.RemoveAll(tmp)
    os.Chdir(tmp)
    
    cmd := InitCmd()
    cmd.SetArgs([]string{"--here", "--integration", "claude"})
    err := cmd.Execute()
    if err != nil {
        t.Fatalf("init failed: %v", err)
    }
    
    // Verify structure
    dirs := []string{".specify", ".specify/scripts", ".specify/templates", ".specify/workflows/speckit", ".specify/memory"}
    for _, d := range dirs {
        if _, err := os.Stat(filepath.Join(tmp, d)); os.IsNotExist(err) {
            t.Errorf("missing directory: %s", d)
        }
    }
    
    files := []string{".specify/speckit.manifest.json", ".specify/integration.json", ".specify/init-options.json"}
    for _, f := range files {
        if _, err := os.Stat(filepath.Join(tmp, f)); os.IsNotExist(err) {
            t.Errorf("missing file: %s", f)
        }
    }
    
    // Verify integration.json content
    state, err := integration.ReadState(tmp)
    if err != nil {
        t.Fatalf("ReadState failed: %v", err)
    }
    if state == nil {
        t.Fatal("integration.json was not created")
    }
    if state.Integration != "claude" {
        t.Errorf("Integration = %q, want %q", state.Integration, "claude")
    }
    if state.IntegrationStateSchema != 1 {
        t.Errorf("IntegrationStateSchema = %d, want %d", state.IntegrationStateSchema, 1)
    }
}

func TestInit_RollsBackOnError(t *testing.T) {
    // TODO: test that if write fails midway, the project dir is removed
    t.Skip("defer to Phase 3 when more file types exist")
}
```

- [ ] Run: `go test -tags=integration ./cmd/specify/...`
- [ ] Commit

---

## Task 10: CI workflow update

**Files:**
- Modify: `specify-go/.github/workflows/ci.yml`

- [ ] Verify CI already covers `go test ./...` — if yes, nothing to do
- [ ] Add `go vet ./...` to CI steps
- [ ] Commit if changed

---

## Dependencies Between Tasks

```
Task 1 (projectstate)          ← base
Task 2 (integration/state)     ← needs Task 1
Task 3 (integration/registry)   ← needs Task 2
Task 4 (init skeleton)         ← needs Tasks 1,2,3
Task 5 (init write files)      ← needs Task 4
Task 6 (init constitution)    ← needs Task 5
Task 7 (check full)           ← needs Tasks 2,3
Task 8 (version full)          ← needs nothing (standalone)
Task 9 (integration tests)    ← needs Tasks 5,7,8
Task 10 (CI update)            ← needs nothing special
```
