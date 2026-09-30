# Install from an npm registry

The npm distribution of the Specify CLI is published on npm as
`@aiaikit/specify-cli`. It exists for environments where Python tooling cannot
reach a public index: it ships as an ordinary npm package, and **installation
performs no network access beyond the npm registry you configure**.

Each GitHub release publishes it automatically through the
`Publish npm distribution` workflow; the tarball is built from the same source
as the PyPI wheel by `.github/scripts/build_npm_package.py`. The Python package
itself stays [`specify-cli` on PyPI](https://pypi.org/project/specify-cli/).

## How it stays offline

`npm install` normally resolves transitive dependencies from a registry. This
package has **none** — the CLI wheel and every Python dependency are vendored as
`.whl` files inside the tarball under `npm/vendor/`. The launcher unpacks them
with a stdlib-only Python script, so the target machine needs neither `pip`, nor
`python3-venv`, nor a compiler, nor root.

## Prerequisites

- Node.js ≥ 18 (runs the launcher only)
- CPython **3.11, 3.12 or 3.13** — the interpreter the vendored wheels target
- One of Linux x86_64, macOS arm64, Windows x64
- PowerShell 7+ (`pwsh`) on Windows, as with any install route, for scaffolding

## Install

From the public registry:

```bash
npm install -g @aiaikit/specify-cli
specify init my-project --integration copilot
```

From an internal mirror that replicates, or hosts, the package:

```bash
npm config set registry https://nexus.example.com/repository/npm-all/
npm install -g @aiaikit/specify-cli
```

From a transferred tarball, with no registry reachable at all:

```bash
npm install -g ./aiaikit-specify-cli-1.0.13.tgz
```

The global prefix may be root-owned; that only affects writing the `specify`
shim. The Python runtime is always unpacked into the invoking user's directory,
so `postinstall` does not need elevated rights.

Point the launcher at a specific interpreter when several are present:

```bash
SPECIFY_PYTHON=/usr/bin/python3.12 specify init my-project --integration copilot
```

On Windows, set the same variable for the session:
`$env:SPECIFY_PYTHON = "C:\Python312\python.exe"`.

## Where the runtime lives

| OS | Location |
| --- | --- |
| Linux / macOS | `${XDG_DATA_HOME:-~/.local/share}/specify-cli-npm/runtime/<cptag>-<version>/` |
| Windows | `%LOCALAPPDATA%\specify-cli-npm\runtime\<cptag>-<version>\` |

That directory is derived entirely from `vendor/`, so removing it is safe; the
next `specify` invocation rebuilds it. Each Python version and CLI version gets
its own subdirectory, and switching `SPECIFY_PYTHON` does not mix runtimes.

## Upgrade

`specify self upgrade` detects this channel and issues the npm command instead
of a `uv`/`pipx` one:

```bash
specify self upgrade --dry-run     # shows the npm command it would run
specify self upgrade               # npm install -g @aiaikit/specify-cli@latest
specify self upgrade --tag v1.0.13 # explicit pin
```

The target version is resolved from GitHub Releases. On a machine that cannot
reach GitHub, pass `--tag` explicitly, or skip the CLI and use npm directly:
`npm view @aiaikit/specify-cli version` reports what your registry holds.

## Publishing

A published GitHub release runs the whole pipeline: vendor wheels on nine
(OS, CPython) cells, assemble, `check`, pack, publish. Publishing is the
default for release events; a manual dispatch defaults to a dry run and only
reports the packed contents. Two settings decide where it goes:

| Setting | Purpose |
| --- | --- |
| `NPM_REGISTRY` (repository variable) | Target registry URL; unset means `https://registry.npmjs.org` |
| `NPM_AUTH_TOKEN` (secret, used by the `npm` environment) | CI/CD automation token for that registry |

`check` is the gate: if any matrix cell is missing a platform, a wheel set
drifted from the manifest, or the package would be unpublishable, the run fails
before anything reaches the registry.

A version with a prerelease suffix (`1.0.13-rc.1`) is published under the `next`
dist-tag so `latest` keeps pointing at the last stable release. Every non-dry
run also uploads the tarball and its `npm-tarball.sha256` as the `npm-tarball`
artifact, which is what air-gapped sites install from directly.

## Building the tarball by hand

Run on a connected machine, once per (OS, CPython) pair — wheels carrying
native extensions are resolved for the machine doing the download:

```bash
python3 .github/scripts/build_npm_package.py download-vendor   # per platform
python3 .github/scripts/build_npm_package.py assemble          # merge + manifest
python3 .github/scripts/build_npm_package.py check             # matrix + integrity
python3 .github/scripts/build_npm_package.py pack              # dist/*.tgz
```

`assemble` rewrites `npm/package.json`'s version from `pyproject.toml`, so the
npm and PyPI identities of a release cannot diverge.

## Troubleshooting

| Symptom | Cause |
| --- | --- |
| `no Python 3.11+ interpreter found` | No usable CPython on PATH; set `SPECIFY_PYTHON` |
| `no vendored wheels for <platform>/<cptag>` | Python newer than the vendored matrix — install 3.11–3.13 or use the pip/uv route |
| `vendored wheel failed checksum` | The tarball was altered in transit; re-fetch and compare `npm-tarball.sha256` |
| `vendor/manifest.json is missing` | The tarball was packed without running `assemble` |
| Install fails in `postinstall` | Expected when no Python is present — the failure is deliberate, so a broken CLI is never left on PATH |
| `402 Payment Required` on publish | Scoped packages are private unless access is declared; `publishConfig.access` in `npm/package.json` covers it |
| `E404 ... scope does not exist` | The `aiaikit` organization has not been created on the target registry yet |

To install without running the unpacker at all, use
`npm install -g --ignore-scripts ./aiaikit-specify-cli-1.0.13.tgz`; the runtime
is then unpacked lazily on the first `specify` command.
