#!/bin/sh
# install.sh — Universal installer for the AIlign CLI.
# Usage: curl -fsSL https://raw.githubusercontent.com/ailign-cli/cli/main/install.sh | sh
#
# Environment variables:
#   AILIGN_VERSION    — Install a specific version (e.g., v0.2.0). Default: latest.
#   INSTALL_DIR       — Override install directory. Default: ~/.local/bin or /usr/local/bin.
#   AILIGN_OS_OVERRIDE  — Override OS detection (for testing).
#   AILIGN_ARCH_OVERRIDE — Override arch detection (for testing).
#   AILIGN_TEST_MODE  — When set, skip download and use mock binaries (for testing).

set -e

GITHUB_REPO="ailign-cli/cli"

# --- Logging ---

log_info() {
    printf '%s\n' "$1"
}

log_error() {
    printf 'Error: %s\n' "$1" >&2
}

log_warn() {
    printf 'Warning: %s\n' "$1" >&2
}

# --- OS and architecture detection ---

detect_os() {
    if [ -n "$AILIGN_OS_OVERRIDE" ]; then
        echo "$AILIGN_OS_OVERRIDE"
        return
    fi
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux)  echo "linux" ;;
        darwin) echo "darwin" ;;
        mingw*|msys*|cygwin*) echo "windows" ;;
        *)      echo "$os" ;;
    esac
}

detect_arch() {
    if [ -n "$AILIGN_ARCH_OVERRIDE" ]; then
        echo "$AILIGN_ARCH_OVERRIDE"
        return
    fi
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)              echo "$arch" ;;
    esac
}

# --- Version resolution ---

get_latest_version() {
    url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    if command -v curl >/dev/null 2>&1; then
        version="$(curl -fsSL "$url" | grep '"tag_name"' | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')"
    elif command -v wget >/dev/null 2>&1; then
        version="$(wget -qO- "$url" | grep '"tag_name"' | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')"
    else
        log_error "Neither curl nor wget found. Please install one and try again."
        exit 1
    fi

    if [ -z "$version" ]; then
        log_error "Could not determine latest version from GitHub API."
        exit 1
    fi
    echo "$version"
}

# --- Install directory resolution ---

resolve_install_dir() {
    if [ -n "$INSTALL_DIR" ]; then
        echo "$INSTALL_DIR"
        return
    fi

    # Prefer ~/.local/bin if it exists or XDG_DATA_HOME is set
    if [ -d "$HOME/.local/bin" ] || [ -n "$XDG_DATA_HOME" ]; then
        echo "$HOME/.local/bin"
        return
    fi

    echo "/usr/local/bin"
}

# --- Download and install ---

download_and_install() {
    os="$1"
    arch="$2"
    version="$3"
    install_dir="$4"

    # Strip leading 'v' for archive name (GoReleaser uses version without v prefix)
    version_num="${version#v}"

    ext="tar.gz"
    if [ "$os" = "windows" ]; then
        ext="zip"
    fi

    archive_name="ailign_${version_num}_${os}_${arch}.${ext}"
    checksums_name="checksums.txt"
    base_url="https://github.com/${GITHUB_REPO}/releases/download/${version}"

    tmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t 'ailign_install')"
    trap 'rm -rf "$tmpdir"' EXIT

    log_info "Downloading ailign ${version} for ${os}/${arch}..."

    # Download archive and checksums
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "${tmpdir}/${archive_name}" "${base_url}/${archive_name}"
        curl -fsSL -o "${tmpdir}/${checksums_name}" "${base_url}/${checksums_name}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "${tmpdir}/${archive_name}" "${base_url}/${archive_name}"
        wget -q -O "${tmpdir}/${checksums_name}" "${base_url}/${checksums_name}"
    else
        log_error "Neither curl nor wget found. Please install curl or wget and re-run this installer."
        exit 1
    fi

    # Verify checksum
    log_info "Verifying checksum..."
    expected_checksum="$(grep -F "  ${archive_name}" "${tmpdir}/${checksums_name}" | awk '{print $1}')"
    if [ -z "$expected_checksum" ]; then
        log_error "Could not find checksum for ${archive_name} in checksums.txt."
        exit 1
    fi

    if command -v sha256sum >/dev/null 2>&1; then
        actual_checksum="$(sha256sum "${tmpdir}/${archive_name}" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        actual_checksum="$(shasum -a 256 "${tmpdir}/${archive_name}" | awk '{print $1}')"
    else
        log_warn "Neither sha256sum nor shasum found. Skipping checksum verification."
        actual_checksum="$expected_checksum"
    fi

    if [ "$actual_checksum" != "$expected_checksum" ]; then
        log_error "Checksum verification failed."
        log_error "Expected: ${expected_checksum}"
        log_error "Actual:   ${actual_checksum}"
        exit 1
    fi
    log_info "Checksum verified."

    # Extract
    log_info "Extracting..."
    if [ "$ext" = "tar.gz" ]; then
        tar xzf "${tmpdir}/${archive_name}" -C "$tmpdir"
    else
        unzip -qo "${tmpdir}/${archive_name}" -d "$tmpdir"
    fi

    # Install binary
    mkdir -p "$install_dir"

    binary_name="ailign"
    if [ "$os" = "windows" ]; then
        binary_name="ailign.exe"
    fi

    if [ -w "$install_dir" ]; then
        cp "${tmpdir}/${binary_name}" "${install_dir}/${binary_name}"
        chmod +x "${install_dir}/${binary_name}"
    else
        log_info "Install directory ${install_dir} requires elevated permissions."
        sudo cp "${tmpdir}/${binary_name}" "${install_dir}/${binary_name}"
        sudo chmod +x "${install_dir}/${binary_name}"
    fi
}

# --- PATH check ---

check_path() {
    install_dir="$1"
    case ":${PATH}:" in
        *":${install_dir}:"*) ;;
        *)
            log_warn "${install_dir} is not in your PATH."
            log_warn "Add it to your shell profile:"
            log_warn "  export PATH=\"${install_dir}:\$PATH\""
            ;;
    esac
}

# --- Main ---

main() {
    os="$(detect_os)"
    arch="$(detect_arch)"

    # Validate platform
    case "$os" in
        linux|darwin) ;;
        windows)
            log_error "This install script is for POSIX systems (Linux/macOS)."
            log_error "On Windows, download the .zip archive from:"
            log_error "  https://github.com/${GITHUB_REPO}/releases"
            exit 1
            ;;
        *)
            log_error "Unsupported operating system: ${os}"
            log_error "Supported platforms: linux, darwin (macOS)"
            exit 1
            ;;
    esac

    case "$arch" in
        amd64|arm64) ;;
        *)
            log_error "Unsupported architecture: ${arch}"
            log_error "Supported architectures: amd64, arm64"
            exit 1
            ;;
    esac

    # Resolve version
    if [ -n "$AILIGN_VERSION" ]; then
        version="$AILIGN_VERSION"
    else
        version="$(get_latest_version)"
    fi

    install_dir="$(resolve_install_dir)"

    # Test mode: skip download, create a mock binary
    if [ -n "$AILIGN_TEST_MODE" ]; then
        mkdir -p "$install_dir"
        cat > "${install_dir}/ailign" << MOCK_EOF
#!/bin/sh
echo "ailign version ${version}"
MOCK_EOF
        chmod +x "${install_dir}/ailign"
    else
        download_and_install "$os" "$arch" "$version" "$install_dir"
    fi

    check_path "$install_dir"

    installed_version="$("${install_dir}/ailign" --version 2>/dev/null || echo "unknown")"
    log_info ""
    log_info "ailign installed successfully!"
    log_info "  Version: ${installed_version}"
    log_info "  Path:    ${install_dir}/ailign"
}

main "$@"
