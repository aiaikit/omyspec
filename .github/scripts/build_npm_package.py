#!/usr/bin/env python3
"""Build the npm distribution of the Specify CLI.

Runs on a connected build machine, never on the air-gapped target. It produces
the vendored wheel set that ``npm/scripts/install_runtime.py`` unpacks, keeps
``npm/package.json`` version-locked to ``pyproject.toml``, and writes the
manifest the installer reads.

Subcommands
    download-vendor  fetch wheels for the machine running this script
    assemble         write vendor/manifest.json and sync npm/package.json
    check            validate an already-assembled package without writing
    pack             assemble, then run `npm pack`

One `download-vendor` run only covers the OS/CPython pair executing it, because
wheels with native extensions resolve for the local platform. Publish jobs must
run it once per (platform, python tag) cell and merge the results.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import re
import shutil
import subprocess
import sys
import tomllib
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
NPM_DIR = REPO_ROOT / "npm"
VENDOR_DIR = NPM_DIR / "vendor"
PACKAGE_JSON = NPM_DIR / "package.json"
MANIFEST_JSON = VENDOR_DIR / "manifest.json"

MANIFEST_SCHEMA_VERSION = "1"
# Keep in sync with npm/lib/python.js MIN_PYTHON.
EXPECTED_PYTHON_REQUIRES = ">=3.11"
# The support matrix is declared here rather than inferred from whatever a
# publish run happened to upload: a matrix cell that fails would otherwise
# shrink the package silently and still pass a self-consistent check.
EXPECTED_PLATFORMS = ("darwin-arm64", "linux-x86_64", "win32-x64")
EXPECTED_PYTHON_TAGS = ("cp311", "cp312", "cp313")
WHEEL_NAME_RE = re.compile(
    r"^(?P<name>[A-Za-z0-9][A-Za-z0-9._+]*)-"
    r"(?P<version>[A-Za-z0-9][A-Za-z0-9._+!]*)-"
    r"(?P<python>[A-Za-z0-9_.]+)-(?P<abi>[A-Za-z0-9_.]+)-(?P<platform>[A-Za-z0-9_.]+)\.whl$"
)
PURE_PLATFORM_TAGS = {"any"}
PURE_PYTHON_TAGS = {"py2.py3", "py3"}
CP_TAG_RE = re.compile(r"^cp(3\d\d)$")
LOCAL_PLATFORM_KEY_RE = re.compile(r"^([a-z0-9_]+-[a-z0-9_]+)/(cp3\d\d)/.+$")


class BuildError(Exception):
    """The package cannot be built or validated as configured."""


def platform_key() -> str:
    """Return the vendor directory name for the machine running the build."""
    import platform as _platform

    system = sys.platform
    machine = (_platform.machine() or "").lower()
    if system == "linux" and machine in {"x86_64", "amd64"}:
        return "linux-x86_64"
    if system == "darwin" and machine == "arm64":
        return "darwin-arm64"
    if system == "win32" and machine in {"amd64", "x86_64"}:
        return "win32-x64"
    raise BuildError(
        f"{system}/{_platform.machine()} is not a supported npm target; refusing to "
        "vendor wheels that no launcher will select."
    )


def python_tag() -> str:
    if sys.implementation.name != "cpython":
        raise BuildError("vendor wheels must be downloaded with CPython")
    return f"cp{sys.version_info[0]}{sys.version_info[1]}"


def pyproject_versions() -> tuple[str, str]:
    data = tomllib.loads((REPO_ROOT / "pyproject.toml").read_text(encoding="utf-8"))
    project = data["project"]
    return project["version"], project.get("requires-python", "")


def to_semver(pep440: str) -> str:
    """Translate a PEP 440 version into the npm semver equivalent.

    npm rejects `1.0.13.dev0`; the release keeps its identity and the pre/post
    part becomes a semver prerelease, so PyPI and npm stay order-comparable.
    """
    match = re.match(
        r"^(?P<release>\d+\.\d+\.\d+)"
        # PEP 440 attaches release candidates without a dot (`1.0.13rc1`) but
        # writes dev/post releases with one (`1.0.13.dev0`); npm wants a hyphen
        # either way.
        r"(?:\.?(?P<kind>a|b|rc|dev|post)\.?(?P<num>\d+))?$",
        pep440,
    )
    if not match:
        raise BuildError(f"version {pep440!r} is not expressible as npm semver")
    kind, num = match.group("kind"), match.group("num")
    if not kind:
        return match.group("release")
    label = {"a": "alpha", "b": "beta", "rc": "rc", "dev": "dev", "post": "post"}[kind]
    return f"{match.group('release')}-{label}.{num}"


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1 << 20), b""):
            digest.update(block)
    return digest.hexdigest()


def build_cli_wheel(dist_dir: Path) -> Path:
    """Build the specify-cli wheel, preferring uv when it is available."""
    existing = sorted(dist_dir.glob("specify_cli-*.whl"))
    if existing:
        return existing[-1]
    dist_dir.mkdir(parents=True, exist_ok=True)
    if shutil.which("uv"):
        argv = ["uv", "build", "--wheel", "--out-dir", str(dist_dir)]
    else:
        argv = [sys.executable, "-m", "build", "--wheel", "--outdir", str(dist_dir)]
    run(argv, cwd=REPO_ROOT)
    produced = sorted(dist_dir.glob("specify_cli-*.whl"))
    if not produced:
        raise BuildError(f"wheel build produced no artifact in {dist_dir}")
    return produced[-1]


def run(argv: list[str], cwd: Path) -> None:
    print("+ " + " ".join(argv))
    result = subprocess.run(argv, cwd=str(cwd))
    if result.returncode != 0:
        raise BuildError(f"command failed ({result.returncode}): {' '.join(argv)}")


def classify(wheel: Path) -> tuple[str, str]:
    """Return (bucket, destination-dir-name) for one wheel file."""
    match = WHEEL_NAME_RE.match(wheel.name)
    if not match:
        raise BuildError(f"unrecognised wheel filename: {wheel.name}")
    python, _, plat = match.group("python"), match.group("abi"), match.group("platform")
    if plat in PURE_PLATFORM_TAGS and python in PURE_PYTHON_TAGS:
        return "common", "common"
    tag_match = CP_TAG_RE.match(python)
    if not tag_match:
        raise BuildError(
            f"{wheel.name} is neither a pure wheel nor CPython-tagged; the npm "
            "installer cannot select it, so it must not be vendored"
        )
    return "platform", tag_match.group(0)


def download_vendor() -> int:
    """Vendor the wheels required by the local platform and interpreter."""
    dist_dir = REPO_ROOT / "dist"
    wheel = build_cli_wheel(dist_dir)
    staging = VENDOR_DIR / "_staging"
    shutil.rmtree(staging, ignore_errors=True)
    staging.mkdir(parents=True)

    run([sys.executable, "-m", "pip", "download", "--dest", str(staging), str(wheel)])

    key, tag = platform_key(), python_tag()
    common = VENDOR_DIR / "common"
    native = VENDOR_DIR / key / tag
    common.mkdir(parents=True, exist_ok=True)
    native.mkdir(parents=True, exist_ok=True)

    moved = 0
    for source in sorted(staging.glob("*.whl")):
        bucket, dirname = classify(source)
        target_dir = common if bucket == "common" else native
        target = target_dir / source.name
        shutil.copy2(source, target)
        moved += 1
    shutil.rmtree(staging, ignore_errors=True)
    print(f"vendored {moved} wheels for {key}/{tag}")
    return moved


def scan_vendor(vendor_root: Path) -> tuple[list[str], dict[str, dict[str, list[str]]], dict[str, str]]:
    """Read the vendor tree into manifest-shaped structures."""
    common = sorted(p.name for p in (vendor_root / "common").glob("*.whl")) if (
        vendor_root / "common"
    ).is_dir() else []
    platforms: dict[str, dict[str, list[str]]] = {}
    checksums: dict[str, str] = {}

    for wheel in sorted(vendor_root.rglob("*.whl")):
        relative = wheel.relative_to(vendor_root).as_posix()
        match = LOCAL_PLATFORM_KEY_RE.match(relative)
        if match:
            key, tag = match.group(1), match.group(2)
            platforms.setdefault(key, {}).setdefault(tag, []).append(wheel.name)
        elif not relative.startswith("common/"):
            raise BuildError(f"wheel is outside the vendor layout: {relative}")
        digest = sha256_file(wheel)
        known = checksums.get(wheel.name)
        if known is not None and known != digest:
            # The manifest and the installer both key checksums by filename, so
            # two same-named wheels with different bytes would make one of them
            # fail verification on the target.
            raise BuildError(f"wheel name collision with different bytes: {wheel.name}")
        checksums[wheel.name] = digest

    return common, {k: {t: sorted(v) for t, v in sorted(tags.items())} for k, tags in sorted(platforms.items())}, checksums


def matrix_problems(platforms: dict) -> list[str]:
    """Compare the vendored matrix against the declared support matrix."""
    problems: list[str] = []
    covered = {key: set(tags) for key, tags in platforms.items()}
    for key in EXPECTED_PLATFORMS:
        tags = covered.get(key)
        if tags is None:
            problems.append(f"platform {key} has no vendored wheels")
            continue
        for tag in EXPECTED_PYTHON_TAGS:
            if tag not in tags:
                problems.append(f"{key} is missing python tag {tag}")
    for key in sorted(set(covered) - set(EXPECTED_PLATFORMS)):
        problems.append(f"platform {key} is vendored but not declared in EXPECTED_PLATFORMS")
    return problems


def manifest_integrity(common: list[str], platforms: dict, checksums: dict) -> str:
    payload = json.dumps(
        {"common": common, "platforms": platforms, "checksums": checksums},
        sort_keys=True,
    )
    return hashlib.sha256(payload.encode("utf-8")).hexdigest()


def write_package_json(pep440: str, platforms: dict) -> None:
    """Rewrite the fields the build owns so npm and PyPI cannot drift."""
    package = json.loads(PACKAGE_JSON.read_text(encoding="utf-8"))
    package["version"] = to_semver(pep440)
    meta = package.setdefault("specifyKit", {})
    meta["pep440Version"] = pep440
    meta["vendoredPlatforms"] = sorted(platforms)
    meta["vendoredPythonTags"] = sorted({tag for tags in platforms.values() for tag in tags})
    PACKAGE_JSON.write_text(json.dumps(package, indent=2) + "\n", encoding="utf-8")


def assemble() -> dict:
    common, platforms, checksums = scan_vendor(VENDOR_DIR)
    pep440, requires_python = pyproject_versions()
    if requires_python != EXPECTED_PYTHON_REQUIRES:
        raise BuildError(
            f"requires-python is {requires_python!r} but the launcher hard-codes "
            f"{EXPECTED_PYTHON_REQUIRES!r} (npm/lib/python.js) — update both together"
        )
    if not any("specify_cli-" in name for name in common):
        raise BuildError("the specify_cli wheel itself is not vendored in common/")
    problems = matrix_problems(platforms)
    if problems:
        raise BuildError("incomplete support matrix: " + "; ".join(problems))

    integrity = manifest_integrity(common, platforms, checksums)
    manifest = {
        "schema_version": MANIFEST_SCHEMA_VERSION,
        "speckit_version": pep440,
        "python_requires": requires_python,
        "integrity": integrity,
        "common": common,
        "platforms": platforms,
        "checksums": checksums,
    }
    VENDOR_DIR.mkdir(parents=True, exist_ok=True)
    MANIFEST_JSON.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    write_package_json(pep440, platforms)
    print(
        f"assembled npm package for {pep440}: {len(common)} common wheels, "
        f"platforms={ {k: sorted(v) for k, v in platforms.items()} }"
    )
    return manifest


def publish_problems(package: dict) -> list[str]:
    """Reject a package.json that the registry would refuse to accept."""
    problems: list[str] = []
    name = package.get("name") or ""
    if not name:
        problems.append("package.json declares no name")
    elif name.startswith("@") and package.get("publishConfig", {}).get("access") != "public":
        # npm creates scoped packages as private, and publishing to a private
        # package on the public registry is a 402 rather than an actionable error.
        problems.append(
            f"scoped package {name} needs publishConfig.access='public' or "
            "`npm publish` fails with 402 Payment Required"
        )
    return problems


def check() -> int:
    """Verify an assembled package matches the repo's declared contract."""
    errors: list[str] = []
    if not MANIFEST_JSON.is_file():
        print("npm package is not assembled: vendor/manifest.json is missing")
        return 1
    manifest = json.loads(MANIFEST_JSON.read_text(encoding="utf-8"))
    package = json.loads(PACKAGE_JSON.read_text(encoding="utf-8"))
    pep440, requires_python = pyproject_versions()

    if manifest.get("speckit_version") != pep440:
        errors.append(
            f"manifest version {manifest.get('speckit_version')} != pyproject {pep440}"
        )
    if package.get("version") != to_semver(pep440):
        errors.append(
            f"package.json version {package.get('version')} != {to_semver(pep440)}"
        )
    if package.get("dependencies") or package.get("devDependencies"):
        errors.append("npm package must declare no dependencies; installs must stay offline")
    errors.extend(publish_problems(package))
    if requires_python != EXPECTED_PYTHON_REQUIRES:
        errors.append(f"requires-python {requires_python!r} != {EXPECTED_PYTHON_REQUIRES!r}")

    platforms = manifest.get("platforms", {})
    errors.extend(matrix_problems(platforms))
    meta = package.get("specifyKit", {})
    if sorted(platforms) != list(meta.get("vendoredPlatforms", [])):
        errors.append(
            f"package.json advertises platforms {meta.get('vendoredPlatforms')} "
            f"but the manifest vendors {sorted(platforms)}"
        )

    common, rescanned, checksums = scan_vendor(VENDOR_DIR)
    if common != manifest.get("common"):
        errors.append("vendor/common no longer matches the manifest")
    if rescanned != platforms:
        errors.append("vendor platform tree no longer matches the manifest")
    if manifest_integrity(common, rescanned, checksums) != manifest.get("integrity"):
        errors.append("manifest integrity hash does not match the vendored bytes")

    for message in errors:
        print(f"FAIL: {message}")
    print("npm package check: OK" if not errors else f"npm package check: {len(errors)} problem(s)")
    return 0 if not errors else 1


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument(
        "command", choices=("download-vendor", "assemble", "check", "pack")
    )
    parser.add_argument("--out-dir", type=Path, default=REPO_ROOT / "dist")
    args = parser.parse_args(argv)

    try:
        if args.command == "download-vendor":
            download_vendor()
        elif args.command == "assemble":
            assemble()
        elif args.command == "check":
            return check()
        elif args.command == "pack":
            assemble()
            npm = shutil.which("npm") or shutil.which("npm.cmd")
            if not npm:
                raise BuildError("npm is not on PATH; cannot pack the tarball")
            args.out_dir.mkdir(parents=True, exist_ok=True)
            run([npm, "pack", "--pack-destination", str(args.out_dir)], cwd=NPM_DIR)
    except BuildError as exc:
        print(f"build_npm_package: {exc}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
