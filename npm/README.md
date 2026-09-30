# @aiaikit/specify-cli (npm distribution)

The npm packaging of the Specify CLI. It exists for environments where `uv` /
`pip` are not an option and software arrives through an npm registry:
**the installation performs no network access at all** — the CLI wheel and every
runtime dependency are vendored inside this tarball.

## What this package is

```text
bin/specify.js            launcher: find python → ensure runtime → exec `python -m specify_cli`
lib/python.js             interpreter discovery (SPECIFY_PYTHON override honoured)
lib/runtime.js            ensure / run plumbing
scripts/install_runtime.py  stdlib-only wheel unpacker (no pip, no venv, no root)
scripts/postinstall.js      unpacks the runtime during npm install
vendor/                   generated at build time: wheels + manifest.json (git-ignored)
```

There are **zero npm dependencies**, on purpose: any transitive dependency would
have to be fetched from a registry, which is exactly what this channel avoids.

## Install

From the public npm registry:

```bash
npm install -g @aiaikit/specify-cli
```

From an internal registry mirror:

```bash
npm config set registry https://nexus.example.com/repository/npm-all/
npm install -g @aiaikit/specify-cli
```

From a transferred tarball, with no registry at all:

```bash
npm install -g ./aiaikit-specify-cli-1.0.13.tgz
```

## Requirements

- Node.js ≥ 18 (only to run the launcher).
- CPython **3.11, 3.12 or 3.13** — the interpreter the vendored wheels were
  built for. There is no pip or build-tool requirement; nothing is compiled on
  the target machine.
- One of: Linux x86_64, macOS arm64, Windows x64.

`specify init` scaffolds projects with shell scripts; on Windows that needs
PowerShell 7+ (`pwsh`), same as the pip/uv channel.

## Where the runtime lives

Wheels are unpacked per user, never into system `site-packages`:

| OS | Location |
| --- | --- |
| Linux / macOS | `${XDG_DATA_HOME:-~/.local/share}/specify-cli-npm/runtime/<cptag>-<version>/` |
| Windows | `%LOCALAPPDATA%\specify-cli-npm\runtime\<cptag>-<version>\` |

Keeping it in a user directory means `npm install -g` into a root-owned prefix
does not need `sudo` to unpack. The directory is a cache derived from
`vendor/`, so deleting it is safe and it rebuilds on next launch.

## Building the tarball

Runs on a connected machine, never on the target:

```bash
python3 .github/scripts/build_npm_package.py download-vendor
python3 .github/scripts/build_npm_package.py assemble
npm pack --pack-destination dist/
```

`npm pack` of a scoped package writes `dist/aiaikit-specify-cli-<version>.tgz`.

`download-vendor` must be executed once per (OS, CPython minor) pair you intend
to support, each on that platform itself — wheels with native extensions are
resolved for the machine running the download.

Publishing is automated: the `Publish npm distribution` workflow runs when a
`vX.Y.Z` release is published and pushes the tarball to the registry in
`NPM_REGISTRY` (default: the public registry) using `NPM_AUTH_TOKEN`. Pre-release
versions are published to the `next` dist-tag so `latest` stays stable. See
[`docs/install/npm.md`](../docs/install/npm.md#publishing).

## Upgrading

`specify self upgrade` detects this channel and prints the npm command rather
than a `uv`/`pipx` one. The upgrade itself is `npm install -g
@aiaikit/specify-cli@<version>`, which resolves against whatever registry is
configured.
