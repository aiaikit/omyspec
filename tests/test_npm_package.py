"""Structural guarantees for the npm distribution under npm/.

These run without Node and without the generated ``npm/vendor/`` tree: they pin
the invariants that make the channel installable offline, so a change that would
require registry access at install time fails here rather than on a target box.
"""

import ast
import importlib.util
import json
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[1]
NPM_DIR = REPO_ROOT / "npm"
PACKAGE_JSON = NPM_DIR / "package.json"
INSTALLER = NPM_DIR / "scripts" / "install_runtime.py"
NPM = shutil.which("npm") or shutil.which("npm.cmd")


def _load_module(name: str, path: Path):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


@pytest.fixture(scope="module")
def package():
    return json.loads(PACKAGE_JSON.read_text(encoding="utf-8"))


@pytest.fixture(scope="module")
def builder():
    return _load_module(
        "build_npm_package", REPO_ROOT / ".github" / "scripts" / "build_npm_package.py"
    )


@pytest.fixture(scope="module")
def installer():
    return _load_module("install_runtime", INSTALLER)


@pytest.fixture(scope="module")
def pyproject_version() -> str:
    text = (REPO_ROOT / "pyproject.toml").read_text(encoding="utf-8")
    match = re.search(r'(?m)^version\s*=\s*"([^"]+)"', text)
    assert match, "could not locate project.version in pyproject.toml"
    return match.group(1)


class TestOfflineInstallInvariant:
    """The tarball has to be self-sufficient; any dependency breaks that."""

    def test_no_runtime_dependencies(self, package):
        assert package.get("dependencies") in (None, {}), (
            "npm dependencies are fetched from a registry, which the offline "
            "channel cannot do; vendor the Python wheel instead"
        )

    def test_no_dev_dependencies_ship(self, package):
        assert package.get("devDependencies") in (None, {})

    def test_bundle_dependencies_are_disabled(self, package):
        assert package.get("bundleDependencies") in (None, []), (
            "bundled npm packages would be unpacked from the tarball; the only "
            "payload allowed here is vendor/*.whl"
        )

    def test_scripts_field_only_has_postinstall(self, package):
        assert set(package["scripts"]) == {"postinstall"}

    def test_files_list_covers_every_runtime_path(self, package):
        # Exact paths, not directories: a directory allowlist also packs whatever
        # lands next to the sources, such as scripts/__pycache__/*.pyc.
        assert set(package["files"]) == {
            "bin/specify.js",
            "lib/python.js",
            "lib/runtime.js",
            "scripts/install_runtime.py",
            "scripts/postinstall.js",
            "vendor/",
            "README.md",
        }

    def test_declared_files_exist(self, package):
        for entry in package["files"]:
            if entry.rstrip("/") == "vendor":
                continue  # generated on a connected build machine
            assert (NPM_DIR / entry).exists(), f"package.json files lists {entry!r}"
        assert (NPM_DIR / "README.md").is_file()

    def test_vendor_tree_is_gitignored(self):
        ignored = (REPO_ROOT / ".gitignore").read_text(encoding="utf-8")
        assert "npm/vendor/" in ignored, (
            "generated wheels must never be committed"
        )


class TestVersionSync:
    """package.json is generated from pyproject.toml and may not drift."""

    def test_version_matches_pyproject(self, package, pyproject_version, builder):
        assert package["version"] == builder.to_semver(pyproject_version)

    def test_pep440_field_matches_pyproject(self, package, pyproject_version):
        assert package["specifyKit"]["pep440Version"] == pyproject_version

    def test_python_requires_matches_pyproject(self, package, builder):
        text = (REPO_ROOT / "pyproject.toml").read_text(encoding="utf-8")
        requires = re.search(r'(?m)^requires-python\s*=\s*"([^"]+)"', text).group(1)
        assert package["specifyKit"]["pythonRequires"] == requires
        assert builder.EXPECTED_PYTHON_REQUIRES == requires

    def test_launcher_minimum_matches_requires_python(self, package):
        source = (NPM_DIR / "lib" / "python.js").read_text(encoding="utf-8")
        match = re.search(
            r"MIN_PYTHON\s*=\s*\{[^}]*major:\s*(\d+)[^}]*minor:\s*(\d+)", source
        )
        assert match, "MIN_PYTHON not found in npm/lib/python.js"
        declared = package["specifyKit"]["pythonRequires"]
        floor = tuple(int(part) for part in re.findall(r"\d+", declared))
        assert (int(match.group(1)), int(match.group(2))) == floor

    def test_declared_matrix_matches_builder(self, package, builder):
        meta = package["specifyKit"]
        assert meta["vendoredPlatforms"] == sorted(builder.EXPECTED_PLATFORMS)
        assert meta["vendoredPythonTags"] == sorted(builder.EXPECTED_PYTHON_TAGS)

    def test_build_script_hardcodes_no_version(self, builder):
        assert builder.MANIFEST_SCHEMA_VERSION == "1"


class TestLauncherContract:
    """bin/specify.js must start the runtime the way the installer builds it."""

    def test_bin_entry_exists_with_node_shebang(self, package):
        entry = NPM_DIR / package["bin"]["specify"]
        assert entry.is_file()
        assert entry.read_text(encoding="utf-8").startswith("#!/usr/bin/env node")

    def test_cli_is_started_as_a_module(self):
        source = (NPM_DIR / "lib" / "runtime.js").read_text(encoding="utf-8")
        assert '"-m", "specify_cli"' in source, (
            "the npm runtime has no console-script shim; it must be invoked "
            "with `python -m specify_cli`"
        )

    def test_site_dir_is_passed_via_pythonpath(self):
        source = (NPM_DIR / "lib" / "runtime.js").read_text(encoding="utf-8")
        assert "PYTHONPATH" in source and "stdio" in source

    def test_specify_cli_has_module_entry_point(self):
        import inspect

        from specify_cli import main

        entry = REPO_ROOT / "src" / "specify_cli" / "__main__.py"
        assert entry.is_file(), "python -m specify_cli needs __main__.py"
        assert 'main(prog_name="specify")' in entry.read_text(encoding="utf-8"), (
            "under `-m` click builds the program name from __package__ and the "
            "executed file, so usage would say `python -m specify_cli` unless "
            "the entry point names the command explicitly"
        )
        assert "prog_name" in inspect.signature(main).parameters

    def test_engines_allows_node_without_python_features(self, package):
        assert package["engines"]["node"] == ">=18"


class TestInstallerIsStdlibOnly:
    """The unpacker runs before anything is installed; it may not need pip."""

    def _imported_roots(self, path: Path) -> set[str]:
        tree = ast.parse(path.read_text(encoding="utf-8"))
        roots: set[str] = set()
        for node in ast.walk(tree):
            if isinstance(node, ast.Import):
                roots.update(alias.name.split(".")[0] for alias in node.names)
            elif isinstance(node, ast.ImportFrom) and node.level == 0 and node.module:
                roots.add(node.module.split(".")[0])
        return roots

    def test_installer_imports_only_stdlib(self, installer):
        third_party = (
            self._imported_roots(INSTALLER) - sys.stdlib_module_names - {"install_runtime"}
        )
        assert not third_party, f"installer must be stdlib-only, imports {third_party}"

    def test_launcher_imports_only_node_builtin(self):
        for js in (NPM_DIR / "lib").glob("*.js"):
            requires = re.findall(r'require\("([^"]+)"\)', js.read_text(encoding="utf-8"))
            assert all(name.startswith("node:") for name in requires), (
                f"{js.name} requires non-builtin modules: {requires}"
            )

    def test_installer_never_shells_out_to_an_installer(self):
        source = INSTALLER.read_text(encoding="utf-8")
        assert "import subprocess" not in source, (
            "the installer runs before anything is installed; shelling out to "
            "pip or venv would fail on images that ship neither"
        )
        for forbidden in ("pip install", "python -m venv", "ensurepip"):
            assert forbidden not in source


class TestWheelExtractionSafety:
    """vendor/*.whl crosses a trust boundary before it reaches site-packages."""

    @pytest.mark.parametrize(
        "member",
        ["../escape.py", "../../escape.py", "pkg/../../escape.py", "/etc/passwd"],
    )
    def test_escaping_members_are_refused(self, installer, member):
        with pytest.raises(installer.UnsupportedError):
            installer._safe_member_name(member)

    @pytest.mark.parametrize("member", ["pkg/mod.py", "pkg/sub/deep.py"])
    def test_normal_members_pass_through(self, installer, member):
        assert installer._safe_member_name(member) == member.replace("\\", "/")

    def test_console_script_payload_is_skipped(self, installer):
        assert installer._safe_member_name("pkg-1.0.data/scripts/tool") is None

    def test_directories_and_empty_are_skipped(self, installer):
        assert installer._safe_member_name("pkg/") is None
        assert installer._safe_member_name("") is None

    def test_manifest_field_validation(self, installer, tmp_path, monkeypatch):
        monkeypatch.setattr(installer, "package_dir", lambda: tmp_path)
        manifest = tmp_path / "vendor" / "manifest.json"
        manifest.parent.mkdir(parents=True)
        manifest.write_text(json.dumps({"schema_version": "1"}))
        with pytest.raises(installer.UnsupportedError):
            installer.read_manifest()

    def test_missing_manifest_explains_the_build_step(self, installer, tmp_path, monkeypatch):
        monkeypatch.setattr(installer, "package_dir", lambda: tmp_path)
        with pytest.raises(installer.UnsupportedError, match="build_npm_package"):
            installer.read_manifest()


class TestPackagingMetadata:
    def test_package_name(self, package):
        from specify_cli import _version

        assert package["name"] == "@aiaikit/specify-cli"
        assert _version._NPM_PACKAGE_NAME == package["name"], (
            "`specify self upgrade` must install the package this directory "
            "publishes, not an unscoped one"
        )

    def test_scoped_package_is_publishable(self, package):
        # npm treats scoped packages as private unless access is declared, which
        # would make an automated publish fail with a 402 rather than publish.
        assert package["publishConfig"]["access"] == "public"

    def test_repository_points_at_the_publishing_scope(self, package):
        # A package published under @aiaikit must name the source it actually
        # came from; linking the npm page to an unrelated upstream repo is a
        # provenance claim, not cosmetics.
        scope = package["name"].split("/")[0].lstrip("@")
        repository = package["repository"]
        assert repository["url"] == f"git+https://github.com/{scope}/omyspec.git"
        assert repository["directory"] == "npm"
        assert package["homepage"].startswith(f"https://github.com/{scope}/")

    def test_license_matches_repo(self, package):
        assert package["license"] == "MIT"
        assert (REPO_ROOT / "LICENSE").is_file()

    def test_license_text_travels_with_the_package(self):
        # npm ships a LICENSE next to package.json even when `files` omits it, so
        # the redistributed MIT code carries its copyright notice. A copy that
        # drifts from the repo's own license is worse than no copy.
        shipped = NPM_DIR / "LICENSE"
        assert shipped.is_file(), "npm/LICENSE missing; the tarball has no license text"
        assert shipped.read_bytes() == (REPO_ROOT / "LICENSE").read_bytes(), (
            "npm/LICENSE differs from the repository LICENSE — refresh the copy"
        )

    def test_no_binary_or_os_constraints(self, package):
        # Restricting os/cpu would stop npm from even unpacking on a supported
        # machine whose libc it cannot classify.
        assert "os" not in package and "cpu" not in package


class TestPublishGate:
    """`build_npm_package.py check` gates the publish job in CI."""

    def test_scoped_package_with_public_access_passes(self, builder):
        package = {
            "name": "@aiaikit/specify-cli",
            "publishConfig": {"access": "public"},
        }
        assert builder.publish_problems(package) == []

    def test_scoped_package_without_declared_access_is_rejected(self, builder):
        problems = builder.publish_problems({"name": "@aiaikit/specify-cli"})
        assert any("402" in problem for problem in problems), (
            "npm defaults a scoped package to private, so the publish job would "
            "fail at the registry instead of here"
        )

    def test_private_access_is_rejected_for_this_channel(self, builder):
        package = {
            "name": "@aiaikit/specify-cli",
            "publishConfig": {"access": "restricted"},
        }
        assert builder.publish_problems(package) != []

    def test_unscoped_package_needs_no_access_field(self, builder):
        assert builder.publish_problems({"name": "specify-cli"}) == []

    def test_missing_name_is_rejected(self, builder):
        assert builder.publish_problems({}) != []

    def test_shipped_package_passes_the_gate(self, package, builder):
        assert builder.publish_problems(package) == []


@pytest.mark.skipif(NPM is None, reason="npm not available")
class TestPackedContents:
    """What ``npm pack`` actually puts in the tarball, not what we intend."""

    @pytest.fixture(scope="class")
    def listing(self, tmp_path_factory):
        # HOME is redirected so npm cannot write into the developer's real
        # ~/.npmrc; a configured credential helper would otherwise leak into the
        # notice output this test asserts on.
        env = {**os.environ, "HOME": str(tmp_path_factory.mktemp("npmhome"))}
        result = subprocess.run(
            [NPM, "pack", "--dry-run"],
            cwd=str(NPM_DIR),
            capture_output=True,
            text=True,
            env=env,
        )
        assert result.returncode == 0, result.stderr
        return result.stdout + result.stderr

    def test_no_bytecode_ships(self, listing):
        assert "__pycache__" not in listing, (
            "a stale .pyc in the tarball outranks the .py the installer loads "
            "against a different CPython and is invisible to every other test"
        )
        assert ".pyc" not in listing

    def test_every_launcher_path_is_present(self, listing):
        for entry in ("bin/specify.js", "lib/python.js", "lib/runtime.js",
                      "scripts/install_runtime.py", "scripts/postinstall.js"):
            assert entry in listing, f"{entry} is not packed; `specify` cannot start"

    def test_license_is_packed(self, listing):
        # Each entry is printed as `npm notice <size> <path>`.
        packed = {line.rsplit(" ", 1)[-1] for line in listing.splitlines()}
        assert "LICENSE" in packed, (
            "the redistributed MIT license text must reach the tarball"
        )

    def test_tarball_filename_matches_ci_glob(self, listing, package):
        # CI packs with `dist/*.tgz` and publishes `dist/*.tgz`; npm flattens a
        # scoped name to `scope-package-version.tgz` (the leading @ is dropped),
        # so assert the filename the docs tell people to install.
        flattened = package["name"].lstrip("@").replace("/", "-")
        expected = f"{flattened}-{package['version']}.tgz"
        assert expected in listing
