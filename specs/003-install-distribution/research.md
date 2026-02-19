# Research: Install & Distribution

## Decision 1: GoReleaser Feature Tiers

**Decision**: Split distribution targets into three tiers based on infrastructure complexity.

**Tier 1 — GoReleaser OSS (GitHub-only credentials)**:
- Homebrew (homebrew_casks) — needs tap repo + PAT
- Scoop (scoops) — needs bucket repo + PAT
- nFPM (nfpms) — local artifact generation (deb, rpm, apk)

**Tier 2 — GoReleaser OSS (external accounts needed)**:
- AUR (aurs) — needs AUR account + SSH key
- Nix (nix) — needs NUR repo + PAT + `nix-hash` on CI runner
- Docker (dockers) — needs registry credentials + `docker login`
- Snapcraft (snapcrafts) — needs snapcraft.io account + credentials
- Chocolatey (chocolateys) — needs chocolatey.org account + API key
- WinGet (winget) — uses DISTRIBUTION_REPO_TOKEN (staging repo), PRs to microsoft/winget-pkgs

**Tier 3 — GoReleaser Pro required**:
- DMG (dmgs) — requires Pro license + `mkisofs` on non-macOS runners

**Rationale**: Tier 1 can be implemented immediately with minimal setup. Tier 2 requires creating accounts and configuring secrets. Tier 3 requires a paid license. Implement in tier order.

## Decision 2: Homebrew — brews vs homebrew_casks

**Decision**: Use `homebrew_casks` (the current recommended section).

**Rationale**: `brews` (formulas) is deprecated in GoReleaser and will be removed in v3. `homebrew_casks` was introduced in v2.10 as the recommended replacement for pre-compiled binaries. Note: unsigned binaries may trigger macOS Gatekeeper; users need `xattr -d com.apple.quarantine` workaround until code signing is implemented.

## Decision 3: NPM Distribution Pattern

**Decision**: Use the platform-specific `optionalDependencies` pattern with a postinstall fallback.

**Package structure**:
- `@ailign/cli` — main wrapper package with CLI shim
- `@ailign/cli-{platform}-{arch}` — platform-specific packages containing the binary
- Platforms: darwin-arm64, darwin-x64, linux-x64, linux-arm64, win32-x64

**Rationale**: This is the industry-standard pattern used by esbuild, Biome, SWC, and Turbo. It works with `--ignore-scripts`, offline installs, and custom registries. The postinstall fallback handles edge cases where optional deps fail.

**Alternatives considered**:
- Postinstall-only download: Fails with `--ignore-scripts`, proxy issues, read-only filesystems
- Single package with embedded binaries: Bloated download for all platforms

## Decision 4: Install Script Design

**Decision**: Create a POSIX shell script (`install.sh`) at repository root.

**Features**:
- Auto-detect OS (linux, darwin) and architecture (amd64, arm64)
- Accept version override via `AILIGN_VERSION` env var
- Fetch latest version from GitHub API when not specified
- Download correct archive from GitHub Releases
- Verify checksum using `checksums.txt`
- Install to `$INSTALL_DIR` > `$HOME/.local/bin` > `/usr/local/bin`
- Warn if install directory is not in `$PATH`

**Rationale**: Consistent with industry standard (rustup, deno, starship all use this pattern). POSIX shell for maximum compatibility. Windows users use Scoop/Chocolatey/WinGet instead.

## Decision 5: Docker Image Strategy

**Decision**: Use `dockers` (not `dockers_v2` yet) with GHCR as the registry.

**Image**: `ghcr.io/ailign-cli/ailign`
**Tags**: `v{version}`, `latest`
**Platforms**: linux/amd64 (single platform initially)

**Rationale**: `dockers_v2` uses buildx for multi-arch but adds complexity. Start with single-platform image on GHCR (free for public repos, integrated with GitHub). Add multi-arch later.

## Decision 6: nFPM Package Formats

**Decision**: Generate deb, rpm, and apk packages.

**Distribution**: Attach as GitHub Release assets (no external package repository initially).

**Rationale**: nFPM generates packages locally with zero external accounts. Users can download and install with `dpkg -i`, `rpm -i`, or `apk add`. Setting up apt/yum/apk repositories is out of scope for v1.

## Secrets Required

| Secret | Used By | Notes |
|--------|---------|-------|
| `GITHUB_TOKEN` | GoReleaser releases | Auto-provided |
| `DISTRIBUTION_REPO_TOKEN` | Homebrew/Scoop/NUR/WinGet push | PAT with `repo` scope on `ailign-cli/distribution` |
| `CHOCOLATEY_API_KEY` | Chocolatey publish | From chocolatey.org account |
| `SNAPCRAFT_STORE_CREDENTIALS` | Snapcraft publish | From `snapcraft export-login` |
| `AUR_KEY` | AUR package push | SSH private key (no passphrase) |
| `NPM_TOKEN` | npm publish | From npmjs.com account |

## External Repositories Required

| Repository | Purpose |
|------------|---------|
| `ailign-cli/distribution` | Unified distribution repo: Homebrew formula (Formula/), Scoop manifest (scoop/), NUR package (nix/), WinGet manifest (winget/) |

## Limitations & Gotchas

- **DMG requires GoReleaser Pro** — skip unless Pro license is available
- **Snapcraft cannot be built inside Docker** — needs native snapcraft on CI runner
- **Chocolatey packages are manually reviewed** — first publish takes time
- **WinGet PRs may fail** — not a blocking issue, just a warning
- **Scoop manifests must be in repo root** — `scoop bucket list` shows 0 manifests if they are in a subdirectory. In the unified `ailign-cli/distribution` repo, Scoop manifests go in root (not a `scoop/` subdirectory)
- **AUR SSH key must be passphrase-free**
- **nix-hash must be on CI PATH** for Nix packages
- **NPM platform packages must have identical versions** — automate from single source
- **Docker: GoReleaser does NOT run `docker login`** — must be a prior CI step
