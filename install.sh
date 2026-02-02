#!/bin/sh
set -e

# x CLI installer
# Works on macOS, Linux, and Windows (Git Bash/WSL)

REPO="priyanshu-shubham/x"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="x"

# Colors (if terminal supports it)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    printf "${GREEN}info${NC}: %s\n" "$1"
}

warn() {
    printf "${YELLOW}warn${NC}: %s\n" "$1"
}

error() {
    printf "${RED}error${NC}: %s\n" "$1"
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)       error "Unsupported operating system: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)            error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    curl -sL "https://api.github.com/repos/${REPO}/releases/latest" |
        grep '"tag_name":' |
        sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install
install() {
    OS=$(detect_os)
    ARCH=$(detect_arch)

    info "Detected OS: $OS, Architecture: $ARCH"

    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Check your internet connection."
    fi

    info "Latest version: $VERSION"

    # Determine file extension
    if [ "$OS" = "windows" ]; then
        EXT="zip"
        BINARY_NAME="x.exe"
    else
        EXT="tar.gz"
    fi

    FILENAME="x_${OS}_${ARCH}.${EXT}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    info "Downloading from $DOWNLOAD_URL"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download
    if command -v curl > /dev/null; then
        curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/$FILENAME"
    elif command -v wget > /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/$FILENAME"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi

    # Extract
    cd "$TMP_DIR"
    if [ "$EXT" = "zip" ]; then
        unzip -q "$FILENAME"
    else
        tar -xzf "$FILENAME"
    fi

    # Install
    if [ "$OS" = "windows" ]; then
        # On Windows (Git Bash), install to user's local bin
        INSTALL_DIR="$HOME/bin"
        mkdir -p "$INSTALL_DIR"
        cp "$BINARY_NAME" "$INSTALL_DIR/"
        info "Installed to $INSTALL_DIR/$BINARY_NAME"
        warn "Make sure $INSTALL_DIR is in your PATH"
    else
        # Check if we can write to /usr/local/bin
        if [ -w "$INSTALL_DIR" ]; then
            cp "$BINARY_NAME" "$INSTALL_DIR/"
            info "Installed to $INSTALL_DIR/$BINARY_NAME"
        else
            info "Need sudo to install to $INSTALL_DIR"
            sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
            info "Installed to $INSTALL_DIR/$BINARY_NAME"
        fi
    fi

    info "Done! Run 'x configure' to set up authentication."
}

install
