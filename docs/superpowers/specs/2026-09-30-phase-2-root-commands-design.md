# Phase 2 设计 — Specify Go Root Commands

> Go 重构项目 `github.com/aiaikit/speckit` 的 Phase 2 设计
> 2026-09-30 · 来源: `src/specify_cli/command_init.py` + `command_check.py` + `command_version.py` + `_agent_config.py` + `_init_options.py` + `_project.py` + `integration_state.py` + `integrations/base.py`

## 1. 范围

### 1.1 Phase 2 交付物 (3 周)

| # | 路径 | 对应 Python | 关键点 |
|---|---|---|---|
| 1 | `cmd/specify/init.go` | `command_init.py` | `--here`/`--integration`/`--preset`/`--extension` flags, 原子写, tracker 渲染 |
| 2 | `cmd/specify/version.go` | `command_version.py` | version + install method + GitHub check |
| 3 | `internal/projectstate/` | `_init_options.py` + `_project.py` | `.specify/` 路径解析, init-options.json 读写 |
| 4 | `internal/integration/` | `integration_state.py` | `integration.json` 读写, IntegrationState normalize |
| 5 | `internal/integration/check.go` | `command_check.py` + `integrations/base.py` | `specify check` 读 integration.json, 对每个 integration 检测工具 |
| 6 | `internal/render/` | `integrations/base.py` | render TOML/MD/YAML integration manifest |

### 1.2 Phase 2 不做

- ❌ 42 个集成适配器 (Phase 3)
- ❌ `extension add` / `preset add` 命令 (Phase 4)
- ❌ workflow run / 其他 workflow 命令 (Phase 5)
- ❌ 交互式选择器 (`select_with_arrows`) — 推迟到 Phase 6+

### 1.3 Go 模块

```
module github.com/aiaikit/speckit

go 1.24

require (
    github.com/spf13/cobra v1.8+
    github.com/charmbracelet/lipgloss v1.0+
    golang.org/x/term v0.27+
    gopkg.in/yaml.v3 v3.0+
)
```

## 2. 架构决策

### 2.1 `internal/projectstate` 是项目路径唯一入口

`resolve_specify_project_root()` 和 `SPECIFY_INIT_DIR` 环境变量解析只在 `projectstate/` 里实现。其他包不直接 `os.Getwd()`。

### 2.2 `specify init` 写出的目录结构

```
.specify/
├── speckit.manifest.json   # schema_version: "1.0"
├── integration.json        # integration_state_schema: 1
├── init-options.json       # ai, integration, script, speckit_version, here, feature_numbering
├── memory/
│   └── constitution.md     # 从 embedded template 生成
├── scripts/
│   └── (从 embedded scripts/ 复制)
├── templates/
│   └── (从 embedded templates/ 复制)
└── workflows/
    └── speckit/
        └── workflow.yml   # 从 embedded workflows/speckit/ 复制
```

### 2.3 原子写

所有 `.specify/` 文件写入使用 `os.CreateTemp` + `os.Rename` (Phase 1 `pathutil` 提供 `WriteFile`，内部已用此模式)。

### 2.4 `specify check` 读取 integration.json

Phase 1 的 `specify check` 是骨架（hardcoded 4 分支）。Phase 2 从 `.specify/integration.json` 读取 installed_integrations，对每个 integration key：
1. 从 INTEGRATION_REGISTRY 查 config
2. 对 `requires_cli=true` 的 integration，调用 `executil.Run` 检测工具是否存在
3. 用 StepTracker 渲染结果

若 `.specify/` 不存在，显示提示"Run `specify init` first"。

### 2.5 `specify version` 子命令

`specify version` 输出与 `--version` flag 不同：
- `--version`: 只输出版本字符串
- `version`: 输出版本 + 安装来源 + GitHub 最新版检查（可选网络调用）

## 3. 组件接口契约

### 3.1 `cmd/specify/init.go`

```go
package cmd

// InitCmd returns the init cobra.Command.
func InitCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "init [project-name]",
        Short: "Initialize a new Specify project",
        Args:  cobra.RangeArgs(0, 1),
    }
    cmd.Flags().Bool("here", false, "Initialize in current directory")
    cmd.Flags().Bool("force", false, "Merge into non-empty directory")
    cmd.Flags().Bool("non-interactive", false, "Skip prompts, use defaults")
    cmd.Flags().String("script", "", "Script type: sh|ps|py (default: sh on unix, ps on windows)")
    cmd.Flags().String("integration", "", "Integration key (e.g. claude, copilot, codex)")
    cmd.Flags().String("integration-options", "", "Options for the integration")
    cmd.Flags().String("preset", "", "Preset to install during init")
    cmd.Flags().StringSlice("extension", nil, "Extensions to install (bundled name, local path, or HTTPS URL)")
    cmd.Flags().Bool("ignore-agent-tools", false, "Skip agent CLI detection")
    // RunE: 解析参数 → resolve project path → StepTracker → 写文件
    return cmd
}
```

### 3.2 `cmd/specify/version.go`

```go
package cmd

func VersionCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "version",
        Short: "Show version and install information",
        RunE:  runVersion,
    }
    cmd.Flags().Bool("json", false, "Output as JSON")
    return cmd
}
```

输出格式（无 `--json`）:
```
specify version 1.2.3
Installed via: uvx (uvx-ephemeral)
Latest release: 1.3.0 (upgrade available)
```
或
```
specify version 1.2.3
Installed via: source checkout
Latest release: 1.2.3 (up to date)
```

### 3.3 `internal/projectstate/`

```go
package projectstate

// ResolveProjectRoot returns the project root.
// Checks SPECIFY_INIT_DIR env var first, then cwd for .specify/.
func ResolveProjectRoot() (string, error)

// ErrNotAPrject = errors.New("not a Spec Kit project")
```

### 3.4 `internal/integration/state.go`

```go
package integration

// ReadState reads .specify/integration.json from project root.
// Returns nil if file does not exist (not an error).
func ReadState(projectRoot string) (*IntegrationState, error)

// WriteState writes .specify/integration.json.
func WriteState(projectRoot string, state *IntegrationState) error

type IntegrationState struct {
    Version                string   `json:"version"`
    IntegrationStateSchema int      `json:"integration_state_schema"` // = 1
    Integration            string   `json:"integration,omitempty"`    // default integration key
    DefaultIntegration     string   `json:"default_integration,omitempty"`
    InstalledIntegrations []string `json:"installed_integrations"`
    IntegrationSettings    map[string]IntegrationSettings `json:"integration_settings,omitempty"`
}

type IntegrationSettings struct {
    Script          string `json:"script,omitempty"`
    RawOptions      string `json:"raw_options,omitempty"`
    ParsedOptions   map[string]any `json:"parsed_options,omitempty"`
    InvokeSeparator string `json:"invoke_separator,omitempty"`
}
```

### 3.5 `internal/integration/check.go`

```go
package integration

// CheckTools checks all installed integrations from integration.json.
// Uses tracker to report each tool's status.
// Returns error only on catastrophic failure (not on missing tools).
func CheckTools(projectRoot string, tracker *ui.Tracker) error
```

### 3.6 `internal/render/`

```go
package render

// WriteManifest writes speckit.manifest.json.
func WriteManifest(projectRoot string, manifest *Manifest) error

type Manifest struct {
    SchemaVersion string `json:"schema_version"` // = "1.0"
    SpecKitVersion string `json:"speckit_version"`
    GeneratedAt   string `json:"generated_at"` // RFC3339
}
```

## 4. 数据流

### 4.1 `specify init --here --integration claude`

```
InitCmd.RunE
  ├─ parse flags
  ├─ resolveProjectRoot() → projectPath
  ├─ ui.ShowBanner()
  ├─ ui.NewTracker("Initialize Specify Project")
  ├─ validate integration key (resolveIntegration)
  ├─ resolve script type (default: sh/ps)
  │
  ├─ IntegrationManifest.setup() → write .specify/integration.json
  │   └─ integration.WriteState()
  ├─ installSharedInfra() → copy scripts/ + templates/
  │   └─ pathutil.Root.WriteFile (atomic)
  ├─ copy bundled workflow.yml → .specify/workflows/speckit/
  ├─ materialize constitution.md
  ├─ saveInitOptions() → .specify/init-options.json
  ├─ ensureExecutableScripts()
  │
  └─ render "Next Steps" panel
```

### 4.2 `specify check` (after init)

```
CheckTools(projectRoot, tracker)
  ├─ ReadState() → IntegrationState
  │   └─ if nil: tracker.skip(all) + return nil (not an error)
  ├─ for each integration key in state.InstalledIntegrations:
  │   ├─ getIntegration(key) → IntegrationConfig
  │   ├─ tracker.Add(key, config.Name)
  │   ├─ if !config.RequiresCLI:
  │   │   └─ tracker.skip(key, "IDE-based")
  │   └─ checkCLI(key) → executil.Run/LookPath
  │       ├─ found: tracker.complete(key, "found")
  │       └─ not found: tracker.error(key, "not found")
  └─ tracker.Render() → print
```

### 4.3 `specify version`

```
runVersion
  ├─ version.DetectInstallMethod() → InstallMethod
  ├─ print: "specify version <Version>"
  ├─ print: "Installed via: <method>"
  ├─ (optional) dlsec.Client.Get(github/latest)
  │   ├─ compare tags
  │   └─ print upgrade notice
  └─ (--json) output JSON {version, installMethod, latest}
```

## 5. 错误处理约定

- `projectstate.ErrNotAProject` — cwd 或 SPECIFY_INIT_DIR 下没有 `.specify/`
- `integration.ErrStateRead` / `ErrStateWrite` — integration.json 读写出错
- `integration.ErrUnknownIntegration` — 未知 integration key
- 所有写操作用 `pathutil` 的原子 WriteFile
- `specify init` 失败时若项目目录是新建的，删除该目录（回滚）

## 6. 测试策略

### 6.1 单元测试

| 包 | 测试内容 |
|---|---|
| `internal/projectstate/` | `SPECIFY_INIT_DIR` override, cwd 无 `.specify/` → ErrNotAProject |
| `internal/integration/` | integration.json round-trip (normal + legacy + missing), normalize state |
| `internal/integration/check.go` | 无 integration.json → skip all; integration.json 有无效 JSON → error |

### 6.2 集成测试（双轨）

`specify init` 在一个临时目录跑，验证：
- 生成的 `.specify/` 目录结构完整
- `integration.json` 内容与 Python 版等价
- `init-options.json` 内容正确

`specify check` 在该目录跑，验证输出包含正确的 integration 名称。

这些是 `//go:build integration` 测试，在 Phase 1 的 `//go:build python_integration` 旁边。

## 7. 任务分解 (3 周)

| 周 | 任务 |
|---|---|
| 1 | `internal/projectstate/` + `internal/integration/state.go` + init 的目录结构部分 |
| 1 | `cmd/specify/init.go` 骨架 + tracker 渲染 |
| 2 | `specify check` 完整版（读 integration.json） |
| 2 | `internal/render/manifest.go` + constitution materialize |
| 3 | `specify version` 子命令 |
| 3 | 集成测试 + CI 更新 |

## 8. 退出标准

```
[1] specify init --here --integration claude    生成的 .specify/ 结构完整
[2] specify init --here --integration claude    integration.json 与 Python 版等价
[3] specify init --here                        目录已存在时报错
[4] specify init nonexistent --integration claude  创建新目录
[5] specify check (已有 .specify/)             正确显示 installed integrations
[6] specify check (无 .specify/)               提示先运行 init
[7] specify version                            显示版本 + 安装来源
[8] specify version --json                     JSON 输出
[9] go test ./...                              所有单元测试通过
[10] 双轨 init 测试 (integration build tag)     .specify/ 文件等价 Python 版
```

## 9. 不变量（Phase 1 继承）

1. **无 shell 注入** — `executil.Run` 仍为 argv-only
2. **路径安全** — `pathutil.Root` 仍拒绝组件级 symlink
3. **原子写** — 所有 `.specify/` 文件用 `pathutil` 写入
4. **`schema_version: "1.0"`** — `speckit.manifest.json` 根字段
5. **Go 1.24** — 保持 `go.mod` 中的 `go 1.24`

## 10. Phase 2 不变量（新增）

10. **init 回滚** — `specify init` 失败时若项目目录为新建，删除该目录
11. **SPECIFY_INIT_DIR** — 环境变量优先于 cwd
12. **integration.json schema 1** — 写入时必须设置 `integration_state_schema: 1`

## 11. 文件清单（Phase 2 结束）

```
specify-go/
├── cmd/specify/
│   ├── main.go
│   ├── root.go
│   ├── check.go          (Phase 1: 骨架 → Phase 2: 完整版)
│   ├── init.go           (Phase 2: 新增)
│   ├── version.go        (Phase 2: 完整版替代 Phase 1 简单版)
│   └── integration_test.go
├── internal/
│   ├── ui/
│   ├── executil/
│   ├── pathutil/
│   ├── assets/
│   ├── version/
│   ├── dlsec/
│   ├── projectstate/     (Phase 2: 新增)
│   │   ├── project.go
│   │   └── project_test.go
│   ├── integration/      (Phase 2: 新增)
│   │   ├── state.go
│   │   ├── state_test.go
│   │   ├── check.go
│   │   ├── check_test.go
│   │   └── INTEGRATION_REGISTRY.go  (硬编码静态 map)
│   └── render/           (Phase 2: 新增)
│       ├── manifest.go
│       └── render_test.go
```

## 12. INTEGRATION_REGISTRY

Phase 2 的 INTEGRATION_REGISTRY 是**硬编码静态 map**（不读 Python 的 catalog.json）。Phase 3 才从 `extensions/catalog.json` 动态加载。

Phase 2 支持的 integrations（对应 Python AGENT_CONFIG keys）：
- `claude` — Claude Code
- `copilot` — GitHub Copilot
- `codex` — OpenAI Codex
- `generic` — 通用 agent（需要 --commands-dir）

其他 integrations 的检测推迟到 Phase 3。
