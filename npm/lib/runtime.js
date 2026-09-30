"use strict";

/**
 * Materialise and launch the unpacked Python runtime.
 *
 * ``ensure`` delegates to scripts/install_runtime.py because that script needs
 * the interpreter's own view of its version and platform tag; doing the
 * selection in Node would mean guessing CPython internals from outside.
 */

const path = require("node:path");
const { spawnSync } = require("node:child_process");

const PACKAGE_ROOT = path.join(__dirname, "..");
const INSTALLER = path.join(PACKAGE_ROOT, "scripts", "install_runtime.py");

function pythonArgv(python, tail) {
  return [...python.args, ...tail];
}

/**
 * Install the runtime if it is missing or stale.
 *
 * @returns {{ok: boolean, siteDir?: string, message?: string}}
 */
function ensure(python) {
  const result = spawnSync(
    python.command,
    pythonArgv(python, [INSTALLER, "ensure"]),
    { encoding: "utf8" },
  );
  if (result.error) {
    return { ok: false, message: `${result.error.message} (while running the runtime installer)` };
  }
  if (result.status !== 0) {
    const detail = (result.stderr || result.stdout || "").trim();
    return { ok: false, message: detail || `runtime installer exited with status ${result.status}` };
  }
  const lines = (result.stdout || "").split(/\r?\n/).filter((line) => line.trim());
  const siteDir = lines.pop();
  if (!siteDir) {
    return { ok: false, message: "runtime installer did not report a site directory" };
  }
  return { ok: true, siteDir: siteDir.trim() };
}

/**
 * Run the CLI as a module so no console-script shim is required.
 *
 * @returns {number} the exit status to propagate
 */
function run(python, siteDir, argv) {
  const env = { ...process.env };
  env.PYTHONPATH = env.PYTHONPATH ? `${siteDir}${path.delimiter}${env.PYTHONPATH}` : siteDir;
  // Keep the child off stale .pyc files from a previous version of the runtime.
  env.PYTHONDONTWRITEBYTECODE = env.PYTHONDONTWRITEBYTECODE || "1";

  const child = spawnSync(
    python.command,
    pythonArgv(python, ["-m", "specify_cli", ...argv]),
    { env, stdio: "inherit" },
  );

  if (child.error) {
    process.stderr.write(`specify: ${child.error.message}\n`);
    return 1;
  }
  if (child.status === null) {
    // Killed by a signal: mirror shell convention (128 + signal number).
    return 128 + (child.signal ? Number(child.signal.replace(/\D+/g, "")) || 1 : 1);
  }
  return child.status;
}

module.exports = { ensure, run, INSTALLER, PACKAGE_ROOT };
