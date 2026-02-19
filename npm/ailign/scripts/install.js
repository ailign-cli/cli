"use strict";

const { existsSync, mkdirSync, createWriteStream, chmodSync } = require("fs");
const { join } = require("path");
const { execSync } = require("child_process");
const https = require("https");
const { platformKey, packageName, binaryName } = require("../lib/platform");

// If the platform package was already installed via optionalDependencies, skip.
const pkg = packageName();
if (pkg) {
  try {
    require.resolve(`${pkg}/package.json`);
    // Platform package exists — nothing to download.
    process.exit(0);
  } catch {
    // Platform package missing — fall through to download.
  }
}

const key = platformKey();
const version = require("../package.json").version;

if (version === "0.0.0") {
  // Placeholder version — nothing to download yet.
  process.exit(0);
}

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const OS_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const os = OS_MAP[process.platform];
const arch = ARCH_MAP[process.arch];

if (!os || !arch) {
  console.error(`ailign: unsupported platform ${key}`);
  process.exit(0); // Don't fail the install — just warn.
}

const ext = process.platform === "win32" ? "zip" : "tar.gz";
const archiveName = `ailign_${version}_${os}_${arch}.${ext}`;
const url = `https://github.com/ailign-cli/cli/releases/download/v${version}/${archiveName}`;

const binDir = join(__dirname, "..", "bin");
const binPath = join(binDir, binaryName());

if (existsSync(binPath)) {
  process.exit(0);
}

mkdirSync(binDir, { recursive: true });

console.log(`ailign: downloading ${url}`);

fetch(url).then(downloadAndExtract).catch((err) => {
  console.error(`ailign: failed to download binary: ${err.message}`);
  console.error("ailign: you can install manually from https://github.com/ailign-cli/cli/releases");
  process.exit(0); // Don't fail the install.
});

async function downloadAndExtract(response) {
  if (!response.ok) {
    // Follow redirects (GitHub releases redirect to S3).
    if (response.status >= 300 && response.status < 400 && response.headers.get("location")) {
      return downloadAndExtract(await fetch(response.headers.get("location")));
    }
    throw new Error(`HTTP ${response.status}`);
  }

  const buffer = Buffer.from(await response.arrayBuffer());
  const tmpArchive = join(binDir, archiveName);

  require("fs").writeFileSync(tmpArchive, buffer);

  try {
    if (ext === "tar.gz") {
      execSync(`tar xzf "${tmpArchive}" -C "${binDir}" ${binaryName()}`, { stdio: "pipe" });
    } else {
      // Windows zip — use PowerShell.
      execSync(
        `powershell -Command "Expand-Archive -Path '${tmpArchive}' -DestinationPath '${binDir}' -Force"`,
        { stdio: "pipe" }
      );
    }

    chmodSync(binPath, 0o755);
    console.log(`ailign: installed to ${binPath}`);
  } finally {
    try {
      require("fs").unlinkSync(tmpArchive);
    } catch {}
  }
}
