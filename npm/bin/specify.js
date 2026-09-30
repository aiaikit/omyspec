#!/usr/bin/env node
"use strict";

/**
 * `specify` launcher for the npm distribution.
 *
 * Node's only job here is to find a usable python, make sure the vendored
 * wheels are unpacked, and hand the command line over. All Spec Kit behaviour
 * stays in Python so the two distribution channels cannot drift apart.
 */

const { findPython, MIN_PYTHON } = require("../lib/python");
const runtime = require("../lib/runtime");

function main() {
  const python = findPython();
  if (!python) {
    process.stderr.write(
      [
        `specify: no Python ${MIN_PYTHON.major}.${MIN_PYTHON.minor}+ interpreter found.`,
        "The npm package ships Python wheels, so it needs CPython on the machine.",
        "Point the launcher at one explicitly:  SPECIFY_PYTHON=/path/to/python3.12 specify --version",
        "",
      ].join("\n"),
    );
    return 1;
  }

  const ensured = runtime.ensure(python);
  if (!ensured.ok) {
    process.stderr.write(`specify: ${ensured.message}\n`);
    return 1;
  }

  return runtime.run(python, ensured.siteDir, process.argv.slice(2));
}

process.exitCode = main();
