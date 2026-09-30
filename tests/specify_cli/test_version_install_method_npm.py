"""Tests for the npm install channel: detection, upgrade argv, and version specs.

The npm distribution runs `python -m specify_cli` against a per-user unpacked
runtime, so `specify self upgrade` has to recognise that layout instead of
reporting an unsupported install, and it has to emit an npm command rather than
a git-source install.
"""

import importlib.util
from pathlib import Path
from unittest.mock import patch

import pytest

from tests.specify_cli.self_upgrade_helpers import (
    _InstallMethod,
    _UpgradePlan,
    _assemble_installer_argv,
    _detect_install_method,
)
from specify_cli import _version

REPO_ROOT = Path(__file__).resolve().parents[2]


def _load_build_script():
    """Load .github/scripts/build_npm_package.py without making it importable."""
    spec = importlib.util.spec_from_file_location(
        "build_npm_package", REPO_ROOT / ".github" / "scripts" / "build_npm_package.py"
    )
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class TestNpmDetection:
    """Tier-1 path-prefix detection for npm installs."""

    def test_npm_runtime_path_matches(self, npm_argv0):
        method, signals = _detect_install_method(include_signals=True)
        assert method == _InstallMethod.NPM
        assert signals.matched_tier == 1
        assert "specify-cli-npm" in signals.matched_prefix.replace("\\", "/")

    def test_detection_is_deterministic(self, npm_argv0):
        assert _detect_install_method() == _detect_install_method() == _InstallMethod.NPM

    def test_xdg_data_home_is_honored(self, monkeypatch, tmp_path):
        monkeypatch.setenv("XDG_DATA_HOME", str(tmp_path / "xdg"))
        entry = (
            tmp_path
            / "xdg"
            / "specify-cli-npm"
            / "runtime"
            / "cp313-1.0.13"
            / "site-packages"
            / "specify_cli"
            / "__main__.py"
        )
        entry.parent.mkdir(parents=True)
        entry.write_text("from . import main\nmain()\n")
        monkeypatch.setattr("sys.argv", [str(entry)])
        assert _detect_install_method() == _InstallMethod.NPM

    def test_prefix_match_does_not_accept_sibling_directory(self, monkeypatch, tmp_path):
        monkeypatch.delenv("XDG_DATA_HOME", raising=False)
        monkeypatch.setenv("HOME", str(tmp_path))
        entry = (
            tmp_path
            / ".local"
            / "share"
            / "specify-cli-npm-experimental"
            / "runtime"
            / "cp311-1.0.13"
            / "site-packages"
            / "specify_cli"
            / "__main__.py"
        )
        entry.parent.mkdir(parents=True)
        entry.write_text("from . import main\nmain()\n")
        monkeypatch.setattr("sys.argv", [str(entry)])
        with patch("specify_cli._version.shutil.which", return_value=None), patch(
            "specify_cli._version._editable_marker_seen", return_value=False
        ):
            assert _detect_install_method() == _InstallMethod.UNSUPPORTED

    def test_python_prefixes_still_win_over_npm(self, uv_tool_argv0):
        # A machine can host both channels; the active entrypoint decides.
        assert _detect_install_method() == _InstallMethod.UV_TOOL


class TestNpmUpgradePlan:
    """`specify self upgrade` routing once the channel is recognised."""

    def test_npm_is_upgradable(self):
        assert _InstallMethod.NPM in _version._upgradable_methods()

    def test_installer_binary_name(self):
        assert _version._installer_binary_name(_InstallMethod.NPM) == "npm"

    def test_method_label(self):
        assert _version._method_label(_InstallMethod.NPM) == "npm"

    def test_upgrade_argv_targets_registry_package(self):
        with patch("specify_cli._version.shutil.which", return_value="/usr/bin/npm"):
            argv = _assemble_installer_argv(_InstallMethod.NPM, "v1.0.13")
        assert argv == [
            "/usr/bin/npm",
            "install",
            "-g",
            f"{_version._NPM_PACKAGE_NAME}@1.0.13",
        ]

    def test_upgrade_argv_without_tag_is_latest(self):
        with patch("specify_cli._version.shutil.which", return_value="/usr/bin/npm"):
            argv = _assemble_installer_argv(_InstallMethod.NPM, None)
        assert argv[-1] == f"{_version._NPM_PACKAGE_NAME}@latest"

    def test_missing_npm_binary_yields_no_argv(self):
        with patch("specify_cli._version.shutil.which", return_value=None):
            assert _assemble_installer_argv(_InstallMethod.NPM, "v1.0.13") is None

    def test_upgrade_argv_does_not_use_git_source(self):
        with patch("specify_cli._version.shutil.which", return_value="/usr/bin/npm"):
            argv = _assemble_installer_argv(_InstallMethod.NPM, "v1.0.13")
        # A source URL would appear as `git+https://…`; rejecting every URL also
        # rejects registries other than github.com.
        assert [part for part in argv if "://" in part] == []

    def test_rollback_hint_uses_npm(self):
        plan = _UpgradePlan(
            method=_InstallMethod.NPM,
            current_version="1.0.12",
            target_tag="v1.0.13",
            installer_argv=None,
            preview_summary="",
            pre_upgrade_snapshot="1.0.12",
        )
        assert _version._rollback_hint(plan) == (
            "To pin back to the previous version: npm install -g "
            f"{_version._NPM_PACKAGE_NAME}@1.0.12"
        )

    def test_unsupported_guidance_offers_npm(self):
        printed = []
        with patch("specify_cli._version.console") as console:
            console.print.side_effect = lambda text, **kw: printed.append(str(text))
            _version._emit_guidance(_InstallMethod.UNSUPPORTED, "v1.0.13")
        text = "\n".join(printed)
        assert f"npm install -g {_version._NPM_PACKAGE_NAME}@1.0.13" in text
        assert "uv tool install specify-cli" in text
        assert "pipx install" in text


class TestNpmVersionSpec:
    """Tag → npm version translation, including the build-script drift guard."""

    @pytest.mark.parametrize(
        "tag,expected",
        [
            ("v1.0.13", "1.0.13"),
            ("v0.7.6", "0.7.6"),
            ("v1.0.13rc1", "1.0.13-rc.1"),
            ("v1.0.13a2", "1.0.13-alpha.2"),
            ("v1.0.13b3", "1.0.13-beta.3"),
            ("v1.0.13.dev0", "1.0.13-dev.0"),
            ("v1.0.13.post2", "1.0.13-post.2"),
            (None, "latest"),
        ],
    )
    def test_tag_mapping(self, tag, expected):
        assert _version._npm_version_spec(tag) == expected

    @pytest.mark.parametrize(
        "pep440",
        ["1.0.13", "0.7.6", "1.0.13rc1", "1.0.13a2", "1.0.13b3", "1.0.13.dev0", "1.0.13.post2"],
    )
    def test_agrees_with_build_script_semver(self, pep440):
        """The CLI and the packager must not disagree on one version's identity."""
        builder = _load_build_script()
        assert _version._npm_version_spec(f"v{pep440}") == builder.to_semver(pep440)

    def test_unparseable_tag_passes_through(self):
        assert _version._npm_version_spec("vX.Y.Z") == "X.Y.Z"
