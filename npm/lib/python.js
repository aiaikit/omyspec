#!/usr/bin/env node
"use strict";

/**
 * Locate a CPython interpreter able to run the vendored runtime.
 *
 * The wheels in vendor/ are selected for one interpreter at install time, so
 * the launcher and the installer must agree on which python they are using.
 * That agreement is guaranteed here: discovery runs once, and the returned
 * command is used for both calls.
 */

const { spawnSync } = require("node:child_process");

// Mirrors the lowest python tag present in vendor/manifest.json, which in turn
// mirrors requires-python in pyproject.toml.
const MIN_PYTHON = { major: 3, minor: 11 };

function candidateCommands() {
  const override = process.env.SPECIFY_PYTHON;
  if (override && override.trim()) {
    // Quoted values may carry arguments ("py -3.12"); split on whitespace only.
    return [override.trim().split(/\s+/)];
  }
  if (process.platform === "win32") {
    return [["py", "-3"], ["python"], ["python3"]];
  }
  return [["python3"], ["python"]];
}

function interpretVersion(stdout, stderr) {
  const text = `${stdout || ""}${stderr || ""}`;
  const match = /Python (\d+)\.(\d+)\.(\d+)/.exec(text);
  if (!match) {
    return null;
  }
  return { major: Number(match[1]), minor: Number(match[2]), micro: Number(match[3]) };
}

function isUsable(version) {
  if (!version) {
    return false;
  }
  if (version.major !== MIN_PYTHON.major) {
    return version.major > MIN_PYTHON.major;
  }
  return version.minor >= MIN_PYTHON.minor;
}

/**
 * @returns {{command: string, args: string[], version: object} | null}
 */
function findPython() {
  for (const argv of candidateCommands()) {
    const probe = spawnSync(argv[0], [...argv.slice(1), "--version"], {
      encoding: "utf8",
      timeout: 10000,
    });
    if (probe.error) {
      continue;
    }
    const version = interpretVersion(probe.stdout, probe.stderr);
    if (probe.status === 0 && isUsable(version)) {
      return { command: argv[0], args: argv.slice(1), version };
    }
  }
  return null;
}

module.exports = { findPython, MIN_PYTHON, candidateCommands };
