"use strict";

const { join } = require("path");

const PLATFORMS = {
  "darwin-arm64": "@ailign/cli-darwin-arm64",
  "darwin-x64": "@ailign/cli-darwin-x64",
  "linux-x64": "@ailign/cli-linux-x64",
  "linux-arm64": "@ailign/cli-linux-arm64",
  "win32-x64": "@ailign/cli-win32-x64",
};

function platformKey() {
  return `${process.platform}-${process.arch}`;
}

function packageName() {
  return PLATFORMS[platformKey()] || null;
}

function binaryName() {
  return process.platform === "win32" ? "ailign.exe" : "ailign";
}

function binaryPath() {
  const pkg = packageName();
  if (!pkg) return null;

  try {
    const pkgDir = require.resolve(`${pkg}/package.json`);
    return join(pkgDir, "..", "bin", binaryName());
  } catch {
    return null;
  }
}

module.exports = { platformKey, packageName, binaryName, binaryPath, PLATFORMS };
