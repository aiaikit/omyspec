#!/usr/bin/env node
"use strict";

/**
 * Unpack the runtime during `npm install` so the first `specify` invocation
 * does not pay for it, and so a machine that cannot host the runtime fails at
 * install time with an actionable message instead of at first use.
 *
 * `npm install --ignore-scripts` skips this on purpose; the launcher then
 * unpacks lazily on first run.
 */

const { findPython, MIN_PYTHON } = require("../lib/python");
const runtime = require("../lib/runtime");

const python = findPython();
if (!python) {
  process.stderr.write(
    [
      "specify-cli postinstall: no usable Python found, so the runtime was not unpacked.",
      `Install CPython ${MIN_PYTHON.major}.${MIN_PYTHON.minor}+ (or set SPECIFY_PYTHON),`,
      "then run any `specify` command; the runtime is unpacked on first use.",
      "",
    ].join("\n"),
  );
  process.exitCode = 1;
} else {
  const ensured = runtime.ensure(python);
  if (!ensured.ok) {
    process.stderr.write(`specify-cli postinstall: ${ensured.message}\n`);
    process.exitCode = 1;
  } else {
    process.stdout.write(`specify-cli: runtime unpacked to ${ensured.siteDir}\n`);
  }
}
