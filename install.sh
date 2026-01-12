#!/bin/sh
set -e

# Configuration
REPO="x950827/jv"
APP_NAME="jv"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() { printf "${BLUE}▸${NC} %s\n" "$1"; }
success() { printf "${GREEN}✓${NC} %s\n" "$1"; }
error() { printf "${RED}✗${NC} %s\n" "$1" >&2; exit 1; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*) echo "darwin" ;;
        Linux*)  echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        armv7l) echo "arm" ;;
        i386|i686) echo "386" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "curl or wget is required"
    fi
}

# Download file
download() {
    url="$1"
    output="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -sL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "curl or wget is required"
    fi
}

main() {
    echo ""
    echo "  ╭─────────────────────────────╮"
    echo "  │     JSON Viewer Installer   │"
    echo "  ╰─────────────────────────────╯"
    echo ""

    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected: ${OS}/${ARCH}"

    info "Fetching latest version..."
    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        error "Failed to get latest version"
    fi
    info "Latest version: ${VERSION}"

    # Construct download URL
    if [ "$OS" = "windows" ]; then
        FILENAME="${APP_NAME}_${OS}_${ARCH}.zip"
    else
        FILENAME="${APP_NAME}_${OS}_${ARCH}.tar.gz"
    fi
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    info "Downloading ${FILENAME}..."
    download "$DOWNLOAD_URL" "${TMP_DIR}/${FILENAME}"

    info "Extracting..."
    cd "$TMP_DIR"
    if [ "$OS" = "windows" ]; then
        unzip -q "$FILENAME"
    else
        tar -xzf "$FILENAME"
    fi

    info "Installing to ${INSTALL_DIR}..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$APP_NAME" "$INSTALL_DIR/$APP_NAME"
    else
        sudo mv "$APP_NAME" "$INSTALL_DIR/$APP_NAME"
    fi
    chmod +x "$INSTALL_DIR/$APP_NAME"

    echo ""
    success "Installed ${APP_NAME} ${VERSION} successfully!"
    echo ""
    echo "  Usage:"
    echo "    ${APP_NAME}                 # Interactive mode"
    echo "    ${APP_NAME} file.json       # View file"
    echo "    cat data.json | ${APP_NAME} # Pipe JSON"
    echo ""
}

main
