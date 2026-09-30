# Phase 3 设计 — Specify Go Self / Extension / Preset Commands

> Go 重构项目 `github.com/aiaikit/speckit` 的 Phase 3 设计
> 2026-09-30 · 来源: `src/specify_cli/selfs/` + `src/specify_cli/extensions/` + `src/specify_cli/presets/`

## 1. 范围

### 1.1 Phase 3 交付物

| # | 路径 | 对应 Python | 关键点 |
|---|---|---|---|
| 1 | `cmd/specify/self.go` | `selfs/command_check.py` | `specify self check` |
| 2 | `cmd/specify/upgrade.go` | `selfs/command_upgrade.py` | `specify self upgrade [--dry-run] [--tag vX.Y.Z]` |
| 3 | `cmd/specify/extension/` | `extensions/` | `extension list` / `info <id>` / `remove <id>` |
| 4 | `cmd/specify/preset/` | `presets/` | `preset list` / `info <id>` / `remove <id>` |
| 5 | `internal/extcatalog/` | `extensions/catalog.json` | 解析 catalog.json，区分 bundled/local/installed |
| 6 | `internal/presetcatalog/` | `presets/` | 解析 preset 清单 |

### 1.2 Phase 3 不做

- ❌ `extension add` / `preset add` 命令（Phase 4）
- ❌ `extension enable` / `extension disable`
- ❌ `extension search` / `extension update`
- ❌ `preset add` / `preset update`
- ❌ workflow run / workflow add（Phase 5）

## 2. 架构决策

### 2.1 `specify self` 命令组

```
specify self check      # 只读：显示当前版本 vs 最新版本
specify self upgrade    # 升级 CLI，支持 --dry-run 和 --tag
```

`self check` 复用 Phase 1 `version.DetectInstallMethod` 和 `FetchLatestRelease`，输出当前版本、安装方式、GitHub 最新版本。

`self upgrade` 根据检测到的安装方式决定行为：
- **uv-tool**: `uv tool upgrade specify`
- **pipx**: `pipx upgrade specify-cli`
- **npm**: `npm install -g specify-cli`
- **uvx ephemeral / source checkout / other**: 打印引导信息，退出 0

`--dry-run` 预览命令但不执行。`--tag vX.Y.Z` 锁定目标版本。

### 2.2 `specify extension` 命令组

```
specify extension list              # 列出所有扩展（bundled + 本地安装）
specify extension info <id>         # 显示扩展详情
specify extension remove <id>       # 移除本地安装的扩展（bundled 不可移除）
```

**扩展来源：**
- **Bundled**: `extensions/catalog.json` 中 `bundled: true` 的扩展，embedded FS 提供
- **Local**: `.specify/extensions/` 目录下用户安装的扩展

**扩展元数据读取优先级：**
1. Bundled 扩展 → 从 embedded `extensions/catalog/` 目录读 `extension.yml`
2. Local 扩展 → 从 `.specify/extensions/<id>/extension.yml` 读

### 2.3 `specify preset` 命令组

```
specify preset list                # 列出所有 preset
specify preset info <id>           # 显示 preset 详情
specify preset remove <id>         # 移除 preset
```

**Preset 来源：**
- **Bundled**: Python presets 在 `src/specify_cli/presets/` — 转为 Go embedded FS 或 catalog.json
- **Local**: `.specify/presets/` 目录下用户安装的 preset

## 3. 数据模型

### 3.1 Extension Catalog (`extensions/catalog.json`)

已在 repo 中：
```json
{
  "schema_version": "1.0",
  "extensions": {
    "agent-context": { "bundled": true, ... },
    "assess": { "bundled": true, ... },
    "bug": { "bundled": true, ... },
    ...
  }
}
```

### 3.2 Extension State (`integration.json`)

`integration.json` 已有 `installed_extensions: []string` 字段（Phase 2 未写入）。Phase 3 init 命令扩展时写入。

### 3.3 Extension Descriptor

```go
type Extension struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description"`
    Author      string   `json:"author"`
    Repository  string   `json:"repository"`
    Tags        []string `json:"tags"`
    Bundled     bool     `json:"bundled"`
    Local       bool     `json:"-"` // not from catalog
}
```

### 3.4 Preset Descriptor

```go
type Preset struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Version     string   `json:"version"`
    Bundled     bool     `json:"bundled"`
    Local       bool     `json:"-"`
}
```

## 4. 组件接口

### 4.1 `cmd/specify/self.go`

```go
func SelfCheckCmd() *cobra.Command
// 复用 version.DetectInstallMethod + version.FetchLatestRelease
// 输出: 当前版本, 安装方式, GitHub 最新版本, 升级建议

func SelfUpgradeCmd() *cobra.Command
// --dry-run: 只打印命令，不执行
// --tag: 锁定目标版本
// RunE: 解析安装方式 → 构建 argv → exec → verify
```

### 4.2 `cmd/specify/extension/`

```go
func ExtensionListCmd() *cobra.Command
func ExtensionInfoCmd() *cobra.Command
func ExtensionRemoveCmd() *cobra.Command
```

### 4.3 `cmd/specify/preset/`

```go
func PresetListCmd() *cobra.Command
func PresetInfoCmd() *cobra.Command
func PresetRemoveCmd() *cobra.Command
```

### 4.4 `internal/extcatalog/`

```go
func LoadCatalog() (*Catalog, error) // 解析 embedded catalog.json
func GetExtension(id string) (*Extension, error)
func ListBundledExtensions() []*Extension
func ListLocalExtensions(projectRoot string) []*Extension
func ListInstalledExtensions(projectRoot string) []*Extension // 来自 integration.json
```

### 4.5 `internal/presetcatalog/`

```go
func ListBundledPresets() []*Preset
func ListLocalPresets(projectRoot string) []*Preset
func GetPreset(id string) (*Preset, error)
```

## 5. 关键文件清单

```
specify-go/
├── cmd/specify/
│   ├── self.go              (Phase 3: 新增)
│   ├── upgrade.go           (Phase 3: 新增)
│   ├── extension/           (Phase 3: 新增)
│   │   ├── cmd.go
│   │   ├── list.go
│   │   ├── info.go
│   │   └── remove.go
│   └── preset/              (Phase 3: 新增)
│       ├── cmd.go
│       ├── list.go
│       ├── info.go
│       └── remove.go
└── internal/
    ├── extcatalog/          (Phase 3: 新增)
    │   ├── catalog.go
    │   └── catalog_test.go
    └── presetcatalog/       (Phase 3: 新增)
        ├── presets.go
        └── presets_test.go
```

## 6. 测试策略

- 单元测试：catalog 解析、extension preset descriptor 字段
- 集成测试：`self check` / `extension list` / `preset list` 在临时目录跑
- Self upgrade 测试：`--dry-run` 验证命令打印正确，`upgrade` 路径检测在 CI mock

## 7. 退出标准

```
[1] specify self check              显示版本信息
[2] specify self upgrade --dry-run  打印升级命令但不执行
[3] specify extension list          列出 bundled + local 扩展
[4] specify extension info <id>     显示扩展详情
[5] specify extension remove <id>   移除本地扩展（Bundled 报错）
[6] specify preset list             列出所有 preset
[7] specify preset info <id>        显示 preset 详情
[8] specify preset remove <id>      移除 preset
[9] go test ./...                  所有单元测试通过
```

## 8. Phase 2 继承的不变量

1. **go 1.24** — `go.mod` 保持 `go 1.24`
2. **module github.com/aiaikit/speckit**
3. **executil.Run argv-only** — 无 shell 注入
4. **pathutil.Root** — 唯一写入口
5. **原子写** — 所有 `.specify/` 写操作用 atomic
6. **dlsec.ValidateURL** — URL 验证用 Phase 1 的 dlsec 包
