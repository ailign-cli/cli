"use strict";

const { existsSync, mkdirSync, unlinkSync, createWriteStream, chmodSync } = require("fs");
const { join } = require("path");
const { execSync } = require("child_process");
const https = require("https");
const { platformKey, packageName, binaryName, fallbackBinaryPath } = require("../lib/platform");

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

// Download into .cache/ directory to avoid colliding with the bin/ailign shim.
const cacheDir = join(__dirname, "..", ".cache");
const binPath = join(cacheDir, binaryName());

if (existsSync(binPath)) {
  process.exit(0);
}

mkdirSync(cacheDir, { recursive: true });

console.log(`ailign: downloading ${url}`);

downloadAndExtract(url).catch((err) => {
  console.error(`ailign: failed to download binary: ${err.message}`);
  console.error("ailign: you can install manually from https://github.com/ailign-cli/cli/releases");
  process.exit(0); // Don't fail the install.
});

function downloadAndExtract(downloadUrl, redirectCount) {
  redirectCount = redirectCount || 0;
  const MAX_REDIRECTS = 5;

  return new Promise((resolve, reject) => {
    https.get(downloadUrl, (response) => {
      const statusCode = response.statusCode || 0;

      // Follow redirects (GitHub releases redirect to S3).
      if (statusCode >= 300 && statusCode < 400 && response.headers.location) {
        response.resume();
        if (redirectCount >= MAX_REDIRECTS) {
          return reject(new Error("Too many redirects"));
        }
        return resolve(downloadAndExtract(response.headers.location, redirectCount + 1));
      }

      if (statusCode < 200 || statusCode >= 300) {
        response.resume();
        return reject(new Error(`HTTP ${statusCode}`));
      }

      const tmpArchive = join(cacheDir, archiveName);
      const fileStream = createWriteStream(tmpArchive);
      response.pipe(fileStream);

      fileStream.on("finish", () => {
        fileStream.close(() => {
          try {
            if (ext === "tar.gz") {
              execSync(`tar xzf "${tmpArchive}" -C "${cacheDir}" ${binaryName()}`, { stdio: "pipe" });
            } else {
              execSync(
                `powershell -Command "Expand-Archive -Path '${tmpArchive}' -DestinationPath '${cacheDir}' -Force"`,
                { stdio: "pipe" }
              );
            }
            chmodSync(binPath, 0o755);
            console.log(`ailign: installed to ${binPath}`);
          } catch (err) {
            reject(err);
            return;
          } finally {
            try { unlinkSync(tmpArchive); } catch {}
          }
          resolve();
        });
      });

      fileStream.on("error", (err) => {
        response.resume();
        try { unlinkSync(tmpArchive); } catch {}
        reject(err);
      });
    }).on("error", reject);
  });
}
