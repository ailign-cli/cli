# AIlign CLI Specifications

**Instruction governance & distribution for engineering organizations**

One source of truth for AI coding assistant instructions across tools and repositories.

## Installation

### Homebrew (macOS / Linux)

```bash
brew install ailign-cli/distribution/ailign
```

### Go Install

```bash
go install github.com/ailign/cli/cmd/ailign@latest
```

### Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/ailign-cli/cli/main/install.sh | sh
```

To install a specific version or to a custom directory:

```bash
AILIGN_VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/ailign-cli/cli/main/install.sh | sh
```

```bash
INSTALL_DIR=/opt/bin curl -fsSL https://raw.githubusercontent.com/ailign-cli/cli/main/install.sh | sh
```

### Scoop (Windows)

```powershell
scoop bucket add ailign https://github.com/ailign-cli/distribution
scoop install ailign
```

### NPM / npx

```bash
npm install -g @ailign/cli
```

Or run without installing:

```bash
npx @ailign/cli --version
```

### Docker

```bash
docker run --rm -v "$(pwd):/repo" -w /repo ghcr.io/ailign-cli/ailign sync
```

### Direct Download

Download the latest release from [GitHub Releases](https://github.com/ailign-cli/cli/releases):

```bash
curl -Lo ailign.tar.gz https://github.com/ailign-cli/cli/releases/download/v0.2.0/ailign_0.2.0_darwin_arm64.tar.gz
tar xzf ailign.tar.gz
chmod +x ailign
```

### Linux Packages (deb/rpm/apk)

Download packages from [GitHub Releases](https://github.com/ailign-cli/cli/releases):

```bash
sudo dpkg -i ailign_0.2.0_linux_amd64.deb
```

```bash
sudo rpm -i ailign_0.2.0_linux_amd64.rpm
```

```bash
sudo apk add --allow-untrusted ailign_0.2.0_linux_amd64.apk
```

### Verify Installation

```bash
ailign --version
```

## Quick Navigation

- **[Vision](vision.md)** - Problem statement, solution approach, and business value
- **[Scope](scope.md)** - What's in/out for MVP, key decisions
- **[Constitution](constitution.md)** - Design principles and values
- **[Features](features/)** - Individual feature specifications

## What is AIlign CLI?

AIlign manages org-wide baselines and repo-specific overlays for AI assistant instructions, rendering them to different tool formats (GitHub Copilot, Cursor, Claude Code, Windsurf).

**Core concept:** `central baseline + repo overlay → rendered formats`

## Target Users

- **Primary:** Developers working in organizations with 20+ repositories
- **Secondary:** Security/Compliance teams enforcing standards

## MVP Features

1. **Package Management** - Versioned instruction packages (`company/security@1.3.0`)
2. **CLI Tool** - `init`, `pull`, `status`, `diff`, `explain` commands
3. **Multi-Tool Rendering** - Output to 2-3 AI tool formats
4. **Deterministic Composition** - Central baseline + local overlays

## Development Approach

This project uses [spec-kit](https://github.com/github/spec-kit) for specification-driven development:

1. Define features in `/features/[feature-name]/`
2. Each feature has: spec, tasks, implementation notes
3. Constitution provides guiding principles
4. Small, incremental changes based on feature boundaries

## Status

🚧 **In Design Phase** - Specifications being written