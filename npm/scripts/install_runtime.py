#!/usr/bin/env python3
"""Install and locate the Spec Kit Python runtime for the npm distribution.

Stdlib-only on purpose: enterprise images frequently ship a python3 with no pip
and no python3-venv, so an installer built on either would fail on exactly the
machines this channel targets. Wheels are ZIP archives with a ``RECORD``;
unpacking one is a ``zipfile`` copy, which needs no installer framework.

The interpreter running this script must be the interpreter that will run the
CLI, because the native wheels (PyYAML) are CPython-minor-version specific.
The launcher in ``bin/specify.js`` guarantees that by resolving one python and
using it for both calls.

Usage:
    python3 install_runtime.py ensure    # install if stale, print site dir
    python3 install_runtime.py status    # print detection/install info as JSON
"""
from __future__ import annotations

import hashlib
import json
import os
import platform
import shutil
import sys
import tempfile
import zipfile
from pathlib import Path

EXIT_OK = 0
EXIT_USAGE = 2
EXIT_UNSUPPORTED = 3
EXIT_BROKEN_PACKAGE = 4
EXIT_IO = 5

STAMP_NAME = "install-stamp.json"


class UnsupportedError(Exception):
    """The target machine or package cannot host an npm-managed runtime."""


def package_dir() -> Path:
    return Path(__file__).resolve().parents[1]


def manifest_path() -> Path:
    return package_dir() / "vendor" / "manifest.json"


def read_manifest() -> dict:
    path = manifest_path()
    if not path.is_file():
        raise UnsupportedError(
            f"vendor/manifest.json is missing from {path}. The npm tarball was "
            "not assembled (run .github/scripts/build_npm_package.py assemble)."
        )
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, ValueError) as exc:
        raise UnsupportedError(f"vendor/manifest.json is unreadable: {exc}") from exc
    if payload.get("schema_version") != "1":
        raise UnsupportedError(
            f"unsupported manifest schema_version {payload.get('schema_version')!r}"
        )
    for field, kind in (("speckit_version", str), ("common", list), ("platforms", dict)):
        if not isinstance(payload.get(field), kind):
            raise UnsupportedError(f"vendor/manifest.json field {field!r} is missing or malformed")
    return payload


def python_tag() -> str:
    """Return the CPython tag (e.g. cp311) of the running interpreter."""
    if sys.implementation.name != "cpython":
        raise UnsupportedError(
            f"the npm distribution supports CPython only, got {sys.implementation.name}"
        )
    return f"cp{sys.version_info[0]}{sys.version_info[1]}"


def platform_key() -> str:
    """Return the vendor directory name for the running OS/arch pair."""
    system = sys.platform
    machine = (platform.machine() or "").lower()
    if system == "linux":
        if machine in {"x86_64", "amd64"}:
            return "linux-x86_64"
    elif system == "darwin" and machine == "arm64":
        return "darwin-arm64"
    elif system == "win32" and machine in {"amd64", "x86_64"}:
        return "win32-x64"
    raise UnsupportedError(
        f"platform {system}/{platform.machine()} is not vendored in this package; "
        "supported: linux-x86_64, darwin-arm64, win32-x64"
    )


def runtime_root() -> Path:
    """Per-user root for the unpacked runtime, kept out of system site-packages."""
    if sys.platform == "win32":
        base = os.environ.get("LOCALAPPDATA")
        if not base:
            raise UnsupportedError("LOCALAPPDATA is not set; cannot locate a user data dir")
        return Path(base) / "specify-cli-npm" / "runtime"
    base = os.environ.get("XDG_DATA_HOME") or str(Path.home() / ".local" / "share")
    return Path(base) / "specify-cli-npm" / "runtime"


def site_dir(version: str, tag: str) -> Path:
    return runtime_root() / f"{tag}-{version}" / "site-packages"


def stamp_path(version: str, tag: str) -> Path:
    return site_dir(version, tag).parent / STAMP_NAME


def current_stamp(version: str, tag: str) -> dict:
    """Return the recorded install stamp, or ``{}`` when there is none."""
    path = stamp_path(version, tag)
    if not path.is_file():
        return {}
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return {}
    return payload if isinstance(payload, dict) else {}


def _validate_wheel_name(name: str) -> str:
    if not name or name != Path(name).name or Path(name).is_absolute():
        raise UnsupportedError(f"manifest references a non-flat wheel path: {name!r}")
    return name


def _sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1 << 20), b""):
            digest.update(block)
    return digest.hexdigest()


def planned_wheels(manifest: dict) -> list[Path]:
    """Return the wheel files this interpreter's platform/tag needs."""
    vendor = package_dir() / "vendor"
    entries: list[Path] = [
        vendor / "common" / _validate_wheel_name(name)
        for name in manifest.get("common", [])
    ]
    native = manifest.get("platforms", {}).get(platform_key(), {})
    tag = python_tag()
    if tag not in native:
        raise UnsupportedError(
            f"no vendored wheels for {platform_key()} / {tag}. This package vendors: "
            + ", ".join(f"{platform_key()}/{t}" for t in sorted(native))
            + ". Install a vendored CPython, or use the pip/uv channel instead."
        )
    entries += [
        vendor / platform_key() / tag / _validate_wheel_name(name)
        for name in native[tag]
    ]
    return entries


def _safe_member_name(name: str) -> str | None:
    """Return the name if it is safe to extract, or None if it should be skipped.

    A wheel is normally built by a trusted party, but the archive crosses a
    trust boundary here (it is unpacked as the invoking user), so a hostile or
    corrupt member must not escape the site directory.
    """
    if not name or name.endswith("/"):
        return None
    normalized = name.replace("\\", "/")
    if normalized.startswith("/") or normalized.startswith("../"):
        raise UnsupportedError(f"refusing unsafe archive member: {name!r}")
    parts = normalized.split("/")
    if len(parts) > 1 and parts[0].endswith(".data"):
        # Wheel metadata dirs: scripts/platlib payload we do not need; the
        # launcher invokes `python -m specify_cli`, not console scripts.
        return None
    if any(part in {"", ".", ".."} for part in parts):
        raise UnsupportedError(f"refusing unsafe archive member: {name!r}")
    drive = Path(normalized).drive
    if drive:
        raise UnsupportedError(f"refusing absolute archive member: {name!r}")
    return normalized


def extract_wheel(wheel: Path, site: Path) -> None:
    with zipfile.ZipFile(wheel) as archive:
        for info in archive.infolist():
            member = _safe_member_name(info.filename)
            if member is None:
                continue
            target = site / member
            parent = target.parent
            parent.mkdir(parents=True, exist_ok=True)
            with archive.open(info) as source, target.open("wb") as sink:
                shutil.copyfileobj(source, sink)


def plan_digest(manifest: dict) -> str:
    """Fingerprint everything that decides the unpacked contents.

    The stamp must not trust a bare version string: if the wheel set, their
    checksums, the platform or the interpreter tag change, the previously
    unpacked runtime is stale and has to be rebuilt.
    """
    fingerprint = hashlib.sha256()
    payload = {
        "integrity": manifest.get("integrity"),
        "speckit_version": manifest.get("speckit_version"),
        "platform_key": platform_key(),
        "python_tag": python_tag(),
        "wheels": [wheel.name for wheel in planned_wheels(manifest)],
        "checksums": manifest.get("checksums") or {},
    }
    fingerprint.update(json.dumps(payload, sort_keys=True).encode("utf-8"))
    return fingerprint.hexdigest()


def install(manifest: dict) -> Path:
    """Unpack the planned wheels and return the site directory.

    Extraction happens in a staging directory that is moved into place only
    once every wheel is unpacked and the stamp is written, so an interrupted
    run never leaves a half-populated runtime that a later launch would trust.
    """
    version = manifest["speckit_version"]
    tag = python_tag()
    site = site_dir(version, tag)
    stamp = stamp_path(version, tag)
    digest = plan_digest(manifest)

    if current_stamp(version, tag).get("plan_digest") == digest:
        return site

    wheels = planned_wheels(manifest)
    for wheel in wheels:
        if not wheel.is_file():
            raise UnsupportedError(f"vendored wheel is missing from the package: {wheel}")
        recorded = (manifest.get("checksums") or {}).get(wheel.name)
        if recorded and _sha256(wheel) != recorded:
            raise UnsupportedError(f"vendored wheel failed checksum: {wheel.name}")

    root = site.parent
    root.mkdir(parents=True, exist_ok=True)
    staging = Path(tempfile.mkdtemp(prefix="staging-", dir=str(root)))
    staging_site = staging / "site-packages"
    staging_site.mkdir(parents=True)
    try:
        for wheel in wheels:
            extract_wheel(wheel, staging_site)
        (staging / STAMP_NAME).write_text(
            json.dumps(
                {
                    "plan_digest": digest,
                    "integrity": manifest.get("integrity"),
                    "speckit_version": version,
                    "python_tag": tag,
                    "platform_key": platform_key(),
                    "interpreter": sys.executable,
                    "wheels": [wheel.name for wheel in wheels],
                },
                indent=2,
                sort_keys=True,
            ),
            encoding="utf-8",
        )

        rival = current_stamp(version, tag)
        if rival.get("plan_digest") == digest:
            # A concurrent launcher published this exact runtime while we were
            # unpacking; discard ours and adopt theirs.
            shutil.rmtree(staging, ignore_errors=True)
            return site

        # Reached only when the recorded plan differs: either nothing is there,
        # or a previous interrupted/older unpack is. Both are rebuilt from the
        # vendored wheels, so replacing them loses nothing.
        if site.exists():
            shutil.rmtree(site, ignore_errors=True)
        if stamp.exists():
            stamp.unlink()
        os.replace(staging_site, site)
        os.replace(staging / STAMP_NAME, stamp)
        shutil.rmtree(staging, ignore_errors=True)
    except Exception:
        shutil.rmtree(staging, ignore_errors=True)
        raise
    return site


def describe(manifest: dict) -> dict:
    version = manifest["speckit_version"]
    tag = python_tag()
    stamp = current_stamp(version, tag)
    return {
        "installed": stamp.get("plan_digest") == plan_digest(manifest),
        "speckit_version": version,
        "python": platform.python_version(),
        "python_tag": tag,
        "platform_key": platform_key(),
        "runtime_root": str(runtime_root()),
        "site_dir": str(site_dir(version, tag)),
        "interpreter": sys.executable,
        "stamp": stamp,
    }


def main(argv: list[str]) -> int:
    command = argv[0] if argv else "ensure"
    if command not in {"ensure", "status"}:
        print(f"unknown command: {command}", file=sys.stderr)
        return EXIT_USAGE
    try:
        manifest = read_manifest()
        if command == "status":
            print(json.dumps(describe(manifest), indent=2, sort_keys=True))
            return EXIT_OK
        print(install(manifest))
        return EXIT_OK
    except UnsupportedError as exc:
        print(f"specify-cli runtime: {exc}", file=sys.stderr)
        return EXIT_UNSUPPORTED
    except zipfile.BadZipFile as exc:
        print(f"specify-cli runtime: corrupt vendored wheel: {exc}", file=sys.stderr)
        return EXIT_BROKEN_PACKAGE
    except OSError as exc:
        print(f"specify-cli runtime: {exc}", file=sys.stderr)
        return EXIT_IO


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
