# Feature Specification: Install & Distribution

**Feature Branch**: `003-install-distribution`
**Created**: 2026-02-17
**Status**: Draft
**Input**: User description: "In order to use this tool without a lot of friction, as a developer, I want to use familiar tooling to install a published version of the tool and have clear instructions on all the methods available"

## User Scenarios & Testing

### User Story 1 - Install via Homebrew (Priority: P1)

**In order to** install ailign using my existing package manager workflow
**As a** macOS or Linux developer
**I want to** run `brew install ailign` and have a working binary

A developer discovers ailign and wants to try it. They open their terminal, run a single brew command, and within seconds have `ailign` available on their PATH. They can immediately run `ailign --version` to confirm it works.

**Why this priority**: Homebrew is the dominant package manager for macOS developers and widely used on Linux. It's the lowest-friction install path for the primary audience and establishes the distribution pattern for other methods.

**Independent Test**: Run `brew install` from the tap, verify `ailign --version` prints the installed version and `ailign sync --help` shows usage.

**Acceptance Scenarios**: See [`features/install-via-homebrew.feature`](../../features/install-via-homebrew.feature)

| Scenario                    | Description                                                    |
|-----------------------------|----------------------------------------------------------------|
| Fresh install from tap      | `brew install` downloads and places binary on PATH             |
| Version matches release     | `ailign --version` output matches the installed release tag    |
| Upgrade to newer version    | `brew upgrade ailign` replaces binary with latest version      |
| Uninstall cleanly           | `brew uninstall ailign` removes the binary completely          |

---

### User Story 2 - Install via Go Toolchain (Priority: P2)

**In order to** install ailign using the Go toolchain I already have
**As a** Go developer
**I want to** run `go install` and have a working binary

A Go developer who already has the Go toolchain installed runs `go install github.com/ailign/cli/cmd/ailign@latest` and gets a working binary in their `$GOPATH/bin`. No additional package managers needed.

**Why this priority**: Go developers are a core audience. `go install` requires zero additional tooling and is the idiomatic way to distribute Go CLI tools. However, it requires a Go toolchain, which not all developers have.

**Independent Test**: Run `go install` with a released version tag, verify the binary works and reports the correct version.

**Acceptance Scenarios**: See [`features/install-via-go.feature`](../../features/install-via-go.feature)

| Scenario                       | Description                                              |
|--------------------------------|----------------------------------------------------------|
| Install latest version         | `go install ...@latest` produces a working binary        |
| Install specific version       | `go install ...@v0.2.0` installs the exact version       |
| Version output includes tag    | `ailign --version` reflects the installed version tag     |

---

### User Story 3 - Download Pre-built Binary (Priority: P3)

**In order to** install ailign on any system without a package manager
**As a** developer or CI/CD pipeline
**I want to** download a pre-built binary from GitHub Releases

A developer on a system without Homebrew or Go, or a CI/CD pipeline that needs ailign, downloads the appropriate binary directly from GitHub Releases. The release page lists binaries for all supported platforms with checksums for verification.

**Why this priority**: Direct download is the universal fallback. It works everywhere and is essential for CI/CD pipelines. However, it's more manual than package managers, so it's lower priority for interactive use.

**Independent Test**: Download a release archive, extract it, verify the binary runs and the checksum matches.

**Acceptance Scenarios**: See [`features/install-via-binary.feature`](../../features/install-via-binary.feature)

| Scenario                       | Description                                                      |
|--------------------------------|------------------------------------------------------------------|
| Download and run on macOS      | Archive for darwin/amd64 or darwin/arm64 contains working binary |
| Download and run on Linux      | Archive for linux/amd64 or linux/arm64 contains working binary   |
| Download and run on Windows    | Archive for windows/amd64 contains working binary (direct download — install script is POSIX-only) |
| Checksum verification          | Published checksums match the downloaded archives                |
| CI/CD install script           | A one-liner curl/wget command installs the binary                |

---

### User Story 4 - Installation Documentation (Priority: P4)

**In order to** find the right installation method for my environment
**As a** developer visiting the project for the first time
**I want to** see clear, concise installation instructions in the README

The project README includes an Installation section near the top listing all available methods (Homebrew, go install, binary download) with copy-pasteable commands. Each method shows the one or two commands needed. A first-time visitor can install ailign within 60 seconds of reading the README.

**Why this priority**: Documentation ties together all installation methods. Without it, users won't discover available options. It depends on the other stories being implemented first.

**Independent Test**: A new developer can follow the README instructions to install ailign using any documented method within 60 seconds.

**Acceptance Scenarios**: See [`features/install-documentation.feature`](../../features/install-documentation.feature)

| Scenario                        | Description                                                    |
|---------------------------------|----------------------------------------------------------------|
| README has install section      | Installation section appears before usage instructions         |
| All methods documented          | Homebrew, go install, install script, Scoop, and npm are all listed |
| Commands are copy-pasteable     | Each method has a single command block that can be copied       |
| Version verification documented | Shows how to verify the installation with `ailign --version`   |

---

### Edge Cases

- What happens when a user installs via Homebrew on an unsupported architecture?
- What happens when `go install` is run with a Go version below the minimum required?
- What happens when a user downloads the wrong platform archive?
- What happens when the Homebrew tap is unreachable (network error)?
- What happens when an older version is already installed via a different method?

## Requirements

### Functional Requirements

- **FR-001**: The project MUST publish a Homebrew tap that installs the `ailign` binary on macOS (Intel and Apple Silicon) and Linux (amd64 and arm64)
- **FR-002**: The project MUST be installable via `go install` with a module path that resolves to the correct binary
- **FR-003**: Every GitHub Release MUST include pre-built binaries for macOS (amd64, arm64), Linux (amd64, arm64), and Windows (amd64)
- **FR-004**: Every GitHub Release MUST include a checksums file for all archives
- **FR-005**: The `ailign --version` command MUST report the version matching the release tag (e.g., `ailign version 0.2.0`)
- **FR-006**: The project README MUST include an Installation section with instructions for all supported methods
- **FR-007**: The release process MUST be automated — tagging a version triggers building, packaging, and publishing without manual steps
- **FR-008**: The Homebrew formula MUST be automatically updated when a new release is published
- **FR-009**: The project MUST provide a one-liner install script for CI/CD environments that downloads and verifies the correct binary

### Key Entities

- **Release**: A versioned distribution of the CLI binary (tag, version, platform archives, checksums)
- **Homebrew Tap**: A repository containing the Homebrew formula for ailign, auto-updated on release
- **Platform Archive**: A compressed file containing the binary for a specific OS/architecture combination

## Success Criteria

### Measurable Outcomes

- **SC-001**: A developer can install ailign via Homebrew in under 30 seconds (excluding download time)
- **SC-002**: A developer can install ailign via `go install` in a single command
- **SC-003**: Pre-built binaries are available for all 5 platform/architecture combinations within 5 minutes of tagging a release
- **SC-004**: 100% of published archives match their published checksums
- **SC-005**: A first-time visitor can find and follow installation instructions within 60 seconds of opening the README
- **SC-006**: The Homebrew formula is automatically updated within 10 minutes of a new release being published

## Scope

### In Scope

- Universal install script (`install.sh`) for CI/CD and quick installs
- `go install` support via proper Go module publishing
- Pre-built binary distribution via GitHub Releases (already partially implemented via GoReleaser)
- Version embedding in the binary at build time (already partially implemented via ldflags)
- GoReleaser-managed package manager distribution:
  - Homebrew (macOS/Linux)
  - Scoop (Windows)
  - WinGet (Windows)
  - Chocolatey (Windows)
  - Snapcraft (Linux)
  - AUR (Arch Linux)
  - Nix/NUR (NixOS/cross-platform)
  - Docker (container image)
  - nFPM (deb, rpm, apk packages)
- NPM wrapper package for Node.js ecosystem distribution
- README installation documentation

### Out of Scope

- DMG (macOS disk image) — deferred, requires GoReleaser Pro license
- Shell completion scripts (separate feature)
- Man page generation (separate feature)
- Automatic update checking within the CLI itself

## Assumptions

- GoReleaser is already configured and the release workflow exists — this feature extends it rather than building from scratch
- GoReleaser natively supports all listed package managers via configuration
- The Homebrew tap will be a separate GitHub repository (standard pattern: `ailign-cli/homebrew-tap`)
- The Scoop bucket will be a separate GitHub repository (standard pattern: `ailign-cli/scoop-bucket`)
- `go install` builds from source and does not receive GoReleaser ldflags; the CLI MUST fall back to Go module build info (e.g., `runtime/debug.ReadBuildInfo`) when the compiled-in version is "dev", so that `ailign --version` still reports a meaningful version
- External accounts/credentials are available for: Chocolatey, Snapcraft, AUR, Docker Hub/GHCR, npm
- The install script will be a POSIX shell script hosted in the repository
- The NPM package wraps the binary — it downloads the correct platform binary during `npm install`

## Dependencies

- GoReleaser configuration (`.goreleaser.yml`) — already exists
- GitHub Release workflow (`.github/workflows/release.yml`) — already exists
- Version ldflags in build (`-X main.version={{.Version}}`) — already configured
- External repositories: Homebrew tap, Scoop bucket
- External accounts: Chocolatey, Snapcraft, AUR, Docker registry, npm registry
- GitHub Actions secrets for publishing credentials
