#!/bin/bash
#
# install.sh
# Copyright © 2025 Ben Sapp
# Licensed under the MIT License
#
# Install the latest version of rlgl CLI tool.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/benwsapp/rlgl/main/install.sh | bash
#
# Options:
#   --version <version>  Install a specific version (e.g., v1.0.0)
#   --dir <directory>    Install to a custom directory (default: ~/.local/bin)
#   --no-path            Skip adding to PATH
#   --force              Force reinstall even if already installed
#   --help               Show this help message

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

INSTALL_DIR="${HOME}/.local/bin"
ADD_TO_PATH=true
FORCE_INSTALL=false
VERSION="latest"

GITHUB_REPO="benwsapp/rlgl"

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

show_help() {
    cat << EOF
rlgl Installer

Install the latest version of rlgl CLI tool.

Usage:
  curl -sSL https://raw.githubusercontent.com/benwsapp/rlgl/main/install.sh | bash

  # Install specific version
  curl -sSL https://raw.githubusercontent.com/benwsapp/rlgl/main/install.sh | bash -s -- --version v1.0.0

  # Install to custom directory
  curl -sSL https://raw.githubusercontent.com/benwsapp/rlgl/main/install.sh | bash -s -- --dir /usr/local/bin

Options:
  --version <version>  Install a specific version (e.g., v1.0.0)
  --dir <directory>    Install to a custom directory (default: ~/.local/bin)
  --no-path            Skip adding to PATH
  --force              Force reinstall even if already installed
  --help               Show this help message

EOF
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --no-path)
            ADD_TO_PATH=false
            shift
            ;;
        --force)
            FORCE_INSTALL=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

detect_platform() {
    local os
    local arch

    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        aarch64|arm64)  arch="arm64" ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac

    echo "${os}-${arch}"
}

get_latest_version() {
    local latest
    latest=$(curl -sSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$latest" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi

    echo "$latest"
}

check_existing_installation() {
    if command -v rlgl &> /dev/null; then
        local current_version
        current_version=$(rlgl serve --version 2>/dev/null | head -1 || echo "unknown")

        if [ "$FORCE_INSTALL" = false ]; then
            log_warning "rlgl is already installed (version: ${current_version})"
            read -p "Do you want to reinstall? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_info "Installation cancelled"
                exit 0
            fi
        fi
    fi
}

install_rlgl() {
    local platform
    local version
    local download_url
    local archive_name
    local temp_dir

    platform=$(detect_platform)

    if [ "$VERSION" = "latest" ]; then
        version=$(get_latest_version)
        log_info "Installing latest version: ${version}"
    else
        version="$VERSION"
        log_info "Installing version: ${version}"
    fi

    archive_name="rlgl-${version}-${platform}.tar.gz"
    download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${archive_name}"

    log_info "Downloading from: ${download_url}"

    temp_dir=$(mktemp -d)
    trap 'rm -rf "$temp_dir"' EXIT

    if command -v curl &> /dev/null; then
        if ! curl -sSL -f -o "${temp_dir}/${archive_name}" "$download_url"; then
            log_error "Failed to download rlgl. Check if version ${version} exists."
            exit 1
        fi
    elif command -v wget &> /dev/null; then
        if ! wget -q -O "${temp_dir}/${archive_name}" "$download_url"; then
            log_error "Failed to download rlgl. Check if version ${version} exists."
            exit 1
        fi
    else
        log_error "curl or wget is required to download rlgl"
        exit 1
    fi

    # Extract archive
    log_info "Extracting archive..."
    if ! tar -xzf "${temp_dir}/${archive_name}" -C "$temp_dir"; then
        log_error "Failed to extract archive"
        exit 1
    fi

    # Install binary
    mkdir -p "$INSTALL_DIR"

    local extracted_dir="${temp_dir}/rlgl-${version}-${platform}"
    local binary_path="${extracted_dir}/rlgl-${version}"

    if [ -f "$binary_path" ]; then
        mv "$binary_path" "${INSTALL_DIR}/rlgl"
        chmod +x "${INSTALL_DIR}/rlgl"
        log_success "rlgl installed to ${INSTALL_DIR}/rlgl"
    else
        log_error "Binary not found in archive (expected: ${binary_path})"
        exit 1
    fi

    if "${INSTALL_DIR}/rlgl" serve --help &> /dev/null; then
        log_success "Installation verified"
    else
        log_warning "Binary installed but verification failed"
    fi
}

add_to_path() {
    if [ "$ADD_TO_PATH" = false ]; then
        return
    fi

    if echo "$PATH" | grep -q "$INSTALL_DIR"; then
        log_info "Install directory already in PATH"
        return
    fi

    log_info "Adding ${INSTALL_DIR} to PATH"

    local shell_config=""

    if [ -n "$BASH_VERSION" ]; then
        if [ -f "$HOME/.bashrc" ]; then
            shell_config="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            shell_config="$HOME/.bash_profile"
        fi
    elif [ -n "$ZSH_VERSION" ]; then
        shell_config="$HOME/.zshrc"
    else
        case "$SHELL" in
            */bash)
                if [ -f "$HOME/.bashrc" ]; then
                    shell_config="$HOME/.bashrc"
                elif [ -f "$HOME/.bash_profile" ]; then
                    shell_config="$HOME/.bash_profile"
                fi
                ;;
            */zsh)
                shell_config="$HOME/.zshrc"
                ;;
        esac
    fi

    if [ -n "$shell_config" ]; then
        if ! grep -q "export PATH=\"\$PATH:${INSTALL_DIR}\"" "$shell_config"; then
            echo "" >> "$shell_config"
            echo "# Added by rlgl installer" >> "$shell_config"
            echo "export PATH=\"\$PATH:${INSTALL_DIR}\"" >> "$shell_config"
            log_success "Added to ${shell_config}"
            log_warning "Please restart your shell or run: source ${shell_config}"
        else
            log_info "PATH already configured in ${shell_config}"
        fi
    else
        log_warning "Could not detect shell config file. Please add ${INSTALL_DIR} to your PATH manually."
    fi
}

main() {
    log_info "Starting rlgl installation..."

    check_existing_installation
    install_rlgl
    add_to_path

    echo ""
    log_success "Installation complete!"
    echo ""
    log_info "Get started with:"
    echo "  $ rlgl serve           # Start the server"
    echo "  $ rlgl client --help   # See client options"
    echo ""
    log_info "Documentation: https://github.com/${GITHUB_REPO}"
}

main
