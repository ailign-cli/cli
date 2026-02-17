# Quickstart: Install & Distribution

## Installing ailign

### Option 1: Homebrew (macOS / Linux)

```bash
brew install ailign-cli/tap/ailign
```

### Option 2: Go Install

```bash
go install github.com/ailign/cli/cmd/ailign@latest
```

### Option 3: Universal Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/ailign-cli/cli/main/install.sh | sh
```

With a specific version:

```bash
AILIGN_VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/ailign-cli/cli/main/install.sh | sh
```

With a custom install directory:

```bash
INSTALL_DIR=/opt/bin curl -fsSL https://raw.githubusercontent.com/ailign-cli/cli/main/install.sh | sh
```

### Option 4: Scoop (Windows)

```powershell
scoop bucket add ailign https://github.com/ailign-cli/scoop-bucket
scoop install ailign
```

### Option 5: NPM / npx

```bash
# Install globally
npm install -g @ailign/cli

# Or run without installing
npx @ailign/cli sync
```

### Option 6: Docker

```bash
docker run --rm -v "$(pwd):/repo" -w /repo ghcr.io/ailign-cli/ailign sync
```

### Option 7: Direct Download

Download the latest release from [GitHub Releases](https://github.com/ailign-cli/cli/releases):

```bash
# macOS (Apple Silicon)
curl -Lo ailign.tar.gz https://github.com/ailign-cli/cli/releases/latest/download/ailign_0.2.0_darwin_arm64.tar.gz
tar xzf ailign.tar.gz
./ailign --version
```

### Option 8: Linux Packages

```bash
# Debian/Ubuntu
sudo dpkg -i ailign_0.2.0_linux_amd64.deb

# RHEL/Fedora
sudo rpm -i ailign_0.2.0_linux_amd64.rpm

# Alpine
sudo apk add --allow-untrusted ailign_0.2.0_linux_amd64.apk
```

## Verify Installation

```bash
ailign --version
# ailign version 0.2.0 (abc1234)
```

## Post-Install: Quick Test

```bash
# Create a minimal config
cat > .ailign.yml << 'EOF'
targets:
  - claude
  - cursor
local_overlays:
  - instructions.md
EOF

# Create an instruction file
echo "# Project Instructions" > instructions.md

# Run sync
ailign sync
```
