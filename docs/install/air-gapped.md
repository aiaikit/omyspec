# Enterprise / Air-Gapped Installation

If your environment blocks access to PyPI or GitHub, you can create a portable wheel bundle on a connected machine and transfer it to the air-gapped target.

## Step 1: Build the wheel on a connected machine

> **Important:** `pip download` resolves platform-specific wheels (e.g., PyYAML includes native extensions). You must run this step on a machine with the **same OS and Python version** as the air-gapped target. If you need to support multiple platforms, repeat this step on each target OS (Linux, macOS, Windows) and Python version.

```bash
# Clone the repository
git clone https://github.com/github/spec-kit.git
cd spec-kit

# Build the wheel
pip install build
python -m build --wheel --outdir dist/

# Download the wheel and all its runtime dependencies
pip download -d dist/ dist/specify_cli-*.whl
```

## Step 2: Transfer the `dist/` directory

Copy the entire `dist/` directory (which contains the `specify-cli` wheel and all dependency wheels) to the target machine via USB, network share, or other approved transfer method.

## Step 3: Install on the air-gapped machine

```bash
pip install --no-index --find-links=./dist specify-cli
```

## Step 4: Initialize a project

No network access is required — bundled assets are used by default:

```bash
specify init my-project --integration copilot
```

> **Note:** Python 3.11+ is required.
>
> **Windows note:** Offline scaffolding requires PowerShell 7+ (`pwsh`), not Windows PowerShell 5.x (`powershell.exe`). Install from <https://aka.ms/powershell>.

## Offline install through npm

If your organisation mirrors npm internally, there is a second offline route
that needs no Python packaging tooling on the target at all:
`@aiaikit/specify-cli` vendors the `specify-cli` wheel and every runtime
dependency inside a single tarball and unpacks them with a stdlib-only Python
script. It requires no `pip`, no `python3-venv`, no compiler, and no root.
Because the package declares no npm dependencies, mirroring that one package is
enough — once your mirror has it, `npm install -g @aiaikit/specify-cli` works
with no egress at all.

To build the tarball on a connected machine instead of relying on a mirror:

```bash
# On a connected build machine:
python3 .github/scripts/build_npm_package.py download-vendor   # once per OS + Python
python3 .github/scripts/build_npm_package.py assemble
python3 .github/scripts/build_npm_package.py pack              # dist/*.tgz

# Transfer dist/*.tgz to the target, then:
npm install -g ./aiaikit-specify-cli-1.0.13.tgz
```

See [Install from npm](npm.md) for requirements, the runtime location, and
troubleshooting.

## Git Credential Manager on Linux

If you're having issues with Git authentication on Linux, you can install Git Credential Manager:

```bash
#!/usr/bin/env bash
set -e
echo "Downloading Git Credential Manager v2.6.1..."
wget https://github.com/git-ecosystem/git-credential-manager/releases/download/v2.6.1/gcm-linux_amd64.2.6.1.deb
echo "Installing Git Credential Manager..."
sudo dpkg -i gcm-linux_amd64.2.6.1.deb
echo "Configuring Git to use GCM..."
git config --global credential.helper manager
echo "Cleaning up..."
rm gcm-linux_amd64.2.6.1.deb
```
