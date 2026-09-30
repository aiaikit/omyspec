# Phase 3 Implementation Plan — Self / Extension / Preset Commands

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `specify self check`, `specify self upgrade`, `specify extension list/info/remove`, and `specify preset list/info/remove`.

**Architecture:** `self check` reuses Phase 1 `version` package. `self upgrade` detects install method and runs the appropriate installer. Extension/Preset discovery reads from `extensions/catalog.json` and local `.specify/extensions/` / `.specify/presets/` directories.

**Tech Stack:** Go 1.24, cobra, lipgloss, `gopkg.in/yaml.v3`, Phase 1 packages (pathutil, executil, assets, ui, version, dlsec).

---

## Global Constraints

- `go 1.24` in `go.mod`
- `module github.com/aiaikit/speckit`
- `executil.Run` signature: no shell parameter (invariant)
- `pathutil.Root` is the only write entry (Phase 1 invariant)
- All writes: `pathutil.Root.WriteFile` (atomic via os.CreateTemp + os.Rename)
- `dlsec.ValidateURL` for URL validation

---

## Task 1: `internal/extcatalog/` — Extension Catalog Reader

**Files:**
- Create: `specify-go/internal/extcatalog/catalog.go`
- Create: `specify-go/internal/extcatalog/catalog_test.go`

**Interfaces:**
- Produces: `LoadCatalog() (*Catalog, error)`, `GetExtension(id string) (*Extension, error)`, `ListBundledExtensions() []*Extension`, `ListLocalExtensions(projectRoot string) []*Extension`

```go
// catalog.go
package extcatalog

import (
    _ "embed"
    "encoding/json"
    "os"
    "path/filepath"
)

//go:embed catalog.json
var embeddedCatalogJSON []byte

type Extension struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description"`
    Author      string   `json:"author"`
    Repository  string   `json:"repository"`
    Tags        []string `json:"tags"`
    Bundled     bool     `json:"bundled"`
}

type Catalog struct {
    SchemaVersion string              `json:"schema_version"`
    Extensions    map[string]Extension `json:"extensions"`
}

func LoadCatalog() (*Catalog, error) {
    var cat Catalog
    if err := json.Unmarshal(embeddedCatalogJSON, &cat); err != nil {
        return nil, err
    }
    return &cat, nil
}

func GetExtension(id string) (*Extension, error) {
    cat, err := LoadCatalog()
    if err != nil {
        return nil, err
    }
    ext, ok := cat.Extensions[id]
    if !ok {
        return nil, ErrNotFound
    }
    return &ext, nil
}

func ListBundledExtensions() []*Extension {
    cat, _ := LoadCatalog()
    var out []*Extension
    for _, ext := range cat.Extensions {
        if ext.Bundled {
            copy := ext
            out = append(out, &copy)
        }
    }
    return out
}

func ListLocalExtensions(projectRoot string) []*Extension {
    // Read .specify/extensions/ directory
    // Each subdir is an extension ID; read extension.yml
}
```

- [ ] **Step 1: Write the failing test**

Write `catalog_test.go` with cases:
- `TestLoadCatalog_Embedded`: `LoadCatalog()` parses the embedded `catalog.json` without error
- `TestLoadCatalog_KnownExtension`: `GetExtension("agent-context")` returns correct fields
- `TestLoadCatalog_UnknownExtension`: `GetExtension("does-not-exist")` returns `ErrNotFound`
- `TestListBundledExtensions`: returns only `bundled: true` extensions

- [ ] **Step 2: Run test to verify it fails**
Run: `go test ./internal/extcatalog/... -v`
Expected: FAIL — package does not exist

- [ ] **Step 3: Write minimal implementation**
Create `catalog.go` with `Extension` struct, `Catalog` struct, `LoadCatalog`, `GetExtension`, `ListBundledExtensions`. Copy `extensions/catalog.json` to `specify-go/internal/extcatalog/catalog.json`.

- [ ] **Step 4: Run test to verify it passes**
Run: `go test ./internal/extcatalog/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add specify-go/internal/extcatalog/
git commit -m "feat(go): add internal/extcatalog with LoadCatalog and GetExtension

Phase 3 task 1.

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

## Task 2: `internal/presetcatalog/` — Preset Catalog Reader

**Files:**
- Create: `specify-go/internal/presetcatalog/presets.go`
- Create: `specify-go/internal/presetcatalog/presets_test.go`

**Interfaces:**
- Produces: `ListBundledPresets() []*Preset`, `GetPreset(id string) (*Preset, error)`

Phase 1 assets already embed `presets/` directory. `BundledPreset(id)` from Task 10 brief already validates id and returns `fs.FS`. Phase 3 needs a higher-level reader.

```go
// presets.go
package presetcatalog

import (
    "errors"
    "io/fs"
    "path/filepath"
)

var ErrNotFound = errors.New("preset not found")

type Preset struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Version     string `json:"version"`
    Bundled     bool   `json:"bundled"`
}

// Bundled presets are subdirs of the embedded presets/ FS.
// Read preset.yml from each subdir.
func ListBundledPresets() []*Preset { ... }
func GetPreset(id string) (*Preset, error) { ... }
```

- [ ] Write `presets_test.go` with: `TestListBundledPresets`, `TestGetPreset_Lean` (known preset), `TestGetPreset_Unknown`
- [ ] Run tests, fix, commit

---

## Task 3: `specify self check` command

**Files:**
- Create: `specify-go/cmd/specify/self.go`

**Interfaces:**
- Consumes: `version.DetectInstallMethod`, `version.FetchLatestRelease`, `dlsec.NewClient`

```go
func SelfCheckCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "check",
        Short: "Check for newer specify releases",
        RunE: func(cmd *cobra.Command, args []string) error {
            v := version.Version
            method := version.DetectInstallMethod()
            fmt.Fprintf(cmd.OutOrStdout(), "Installed: %s\n", v)
            fmt.Fprintf(cmd.OutOrStdout(), "Via: %s\n", method.String())
            client := dlsec.NewClient()
            rel, err := version.FetchLatestRelease(ctx, client, "https://github.com/aiaikit/speckit/releases/latest")
            if err != nil {
                fmt.Fprintf(cmd.OutOrStdout(), "[yellow]Could not check latest release[/yellow]\n")
                return nil
            }
            current := strings.TrimPrefix(v, "v")
            latest := strings.TrimPrefix(rel.TagName, "v")
            if latest != current {
                fmt.Fprintf(cmd.OutOrStdout(), "[green]Update available: %s → %s[/green]\n", v, rel.TagName)
            } else {
                fmt.Fprintf(cmd.OutOrStdout(), "[green]Up to date[/green]\n")
            }
            return nil
        },
    }
    return cmd
}
```

- [ ] Write `self.go` with `SelfCheckCmd()`
- [ ] In `root.go`, add: `cmd.AddCommand(SelfCheckCmd())`
- [ ] Write `self_test.go`: `TestSelfCheckCmd_Exists`
- [ ] `go build ./cmd/specify` compiles
- [ ] Commit

---

## Task 4: `specify self upgrade` command

**Files:**
- Create: `specify-go/cmd/specify/upgrade.go`

**Interfaces:**
- Consumes: `version.DetectInstallMethod`, `executil.Run`

```go
func SelfUpgradeCmd() *cobra.Command {
    var dryRun bool
    var tag string

    cmd := &cobra.Command{
        Use:   "upgrade",
        Short: "Upgrade the specify CLI",
        RunE: func(cmd *cobra.Command, args []string) error {
            method := version.DetectInstallMethod()
            // Build argv based on method + tag
            // Run dry-run or actual upgrade
        },
    }
    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview upgrade command without running it")
    cmd.Flags().StringVar(&tag, "tag", "", "Pin target version (e.g. v1.2.3)")
    return cmd
}
```

**Install method → argv mapping:**
- `uv-tool`: `["uv", "tool", "upgrade", "specify"]`
- `pipx`: `["pipx", "upgrade", "specify-cli"]`
- `npm`: `["npm", "install", "-g", "specify-cli"]`
- `uvx`, `source-checkout`, `unknown`: print guidance, exit 0

- [ ] Write `upgrade.go`
- [ ] Add `cmd.AddCommand(SelfUpgradeCmd())` in `root.go`
- [ ] Write `upgrade_test.go`: `TestUpgradeCmd_DryRun` (verify correct argv printed)
- [ ] Tests pass, commit

---

## Task 5: `specify extension list` and `extension info`

**Files:**
- Create: `specify-go/cmd/specify/extension/cmd.go`
- Create: `specify-go/cmd/specify/extension/list.go`
- Create: `specify-go/cmd/specify/extension/info.go`
- Create: `specify-go/cmd/specify/extension/extension_test.go`

```go
// cmd.go
func ExtensionCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "extension", Short: "Manage extensions"}
    cmd.AddCommand(ExtensionListCmd(), ExtensionInfoCmd(), ExtensionRemoveCmd())
    return cmd
}
```

- [ ] Write `extension/cmd.go`, `list.go`, `info.go`
- [ ] `list.go`: `ExtensionListCmd()` — table with Name, ID, Version, Source (bundled/local) columns
- [ ] `info.go`: `ExtensionInfoCmd()` — `cobra.Command{Use: "info [id]"}` — show all fields for one extension
- [ ] In `root.go`, add: `cmd.AddCommand(extension.ExtensionCmd())`
- [ ] Write `extension_test.go`: `TestExtensionCmd_Exists`, `TestExtensionListCmd`
- [ ] Tests pass, commit

---

## Task 6: `specify extension remove`

**Files:**
- Modify: `specify-go/cmd/specify/extension/remove.go`
- Modify: `specify-go/cmd/specify/extension/extension_test.go`

- [ ] Write `remove.go`: `ExtensionRemoveCmd()` — remove local extension from `.specify/extensions/<id>/`
- [ ] Bundled extensions: error "bundled extensions cannot be removed"
- [ ] Unknown extensions: error "extension not found"
- [ ] Update `extension_test.go` with `TestExtensionRemoveCmd`
- [ ] Tests pass, commit

---

## Task 7: `specify preset list`, `info`, `remove`

**Files:**
- Create: `specify-go/cmd/specify/preset/cmd.go`
- Create: `specify-go/cmd/specify/preset/list.go`
- Create: `specify-go/cmd/specify/preset/info.go`
- Create: `specify-go/cmd/specify/preset/remove.go`
- Create: `specify-go/cmd/specify/preset/preset_test.go`

- [ ] Mirror Task 5+6 structure for presets
- [ ] `root.go`: add `cmd.AddCommand(preset.PresetCmd())`
- [ ] Tests pass, commit

---

## Task 8: Integration tests

**Files:**
- Modify: `specify-go/cmd/specify/integration_test.go`
- Create: `specify-go/cmd/specify/self_integration_test.go` (build tag: `integration`)

- [ ] `TestSelfCheck`: run `SelfCheckCmd()`, verify output contains version string
- [ ] `TestExtensionList`: run `ExtensionListCmd()` in project with bundled extensions
- [ ] `TestPresetList`: run `PresetListCmd()`
- [ ] Run: `go test -tags=integration ./cmd/specify/...`
- [ ] Commit

---

## Task 9: Final CI / go vet fix

**Files:**
- Modify: `specify-go/internal/pathutil/root.go`

The pre-existing `go vet` failures use go1.25 APIs on a go1.24 module. Phase 3 closes the project, so fix this now:
- Replace `os.WriteFile(path, data, 0644)` with `ioutil.WriteFile` (go1.16, available in go1.24)
- Or use `os.Create` + `Write` pattern

- [ ] Run `go vet ./...` to see all failures
- [ ] Fix pathutil/root.go go1.25 API usage
- [ ] Verify `go build ./...` and `go test ./...` pass
- [ ] Commit

---

## Dependencies Between Tasks

```
Task 1 (extcatalog)         ← base for extension commands
Task 2 (presetcatalog)      ← base for preset commands
Task 3 (self check)         ← needs nothing special (standalone)
Task 4 (self upgrade)       ← needs Task 3
Task 5 (extension list/info) ← needs Task 1
Task 6 (extension remove)   ← needs Task 5
Task 7 (preset list/info/remove) ← needs Task 2
Task 8 (integration tests)  ← needs Tasks 4,6,7
Task 9 (go vet fix)          ← needs nothing special
```
