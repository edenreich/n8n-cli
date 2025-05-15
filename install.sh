#!/bin/sh

set -e

GITHUB_REPO="edenreich/n8n-cli"
INSTALL_DIR="$HOME/.local/bin"

info() {
  echo "\033[0;34m==>\033[0m $1"
}

success() {
  echo "\033[0;32m==>\033[0m $1"
}

error() {
  echo "\033[0;31mERROR:\033[0m $1" >&2
  exit 1
}

detect_os() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$OS" in
    linux)
      OS="linux"
      ;;
    darwin)
      OS="darwin"
      ;;
    *)
      error "Unsupported OS: $OS. This tool only supports Linux and macOS."
      ;;
  esac
  echo $OS
}

detect_arch() {
  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64|amd64)
      ARCH="amd64"
      ;;
    armv7*|armv6*|armv5*)
      ARCH="arm"
      ;;
    aarch64|arm64)
      ARCH="arm64"
      ;;
    *)
      error "Unsupported architecture: $ARCH"
      ;;
  esac
  echo $ARCH
}

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

get_latest_release() {
  if command_exists "curl"; then
    VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  else
    error "curl is required for installation"
  fi
  
  if [ -z "$VERSION" ]; then
    error "Could not determine the latest version"
  fi
  
  echo $VERSION
}

prepare_install_dir() {
  mkdir -p "$INSTALL_DIR" 2>/dev/null || error "Failed to create directory $INSTALL_DIR"
}

do_install() {
  OS=$(detect_os)
  ARCH=$(detect_arch)
  VERSION=$(get_latest_release)
  
  prepare_install_dir
  
  info "Installing n8n $VERSION for $OS/$ARCH to $INSTALL_DIR..."
  
  VERSION_NUM=$(echo "$VERSION" | sed 's/^v//')
  
  BINARY="n8n"
  
  DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/n8n_${OS}_${ARCH}"
  
  info "Downloading from $DOWNLOAD_URL"

  if ! curl -sL "$DOWNLOAD_URL" -o "$INSTALL_DIR/$BINARY"; then
    error "Failed to download n8n"
  fi
  
  success "Download complete"

  chmod +x "$INSTALL_DIR/$BINARY" || error "Failed to make binary executable"

  echo "\033[1;32m
          ____      
   ____  ( __ )____ 
  / __ \\/ __  / __ \\
 / / / / /_/ / / / /
/_/ /_/\\____/_/ /_/ CLI
\033[0m"
  echo "\033[1;32m================================\033[0m"
  echo "\033[1;32m✓ Successfully installed n8n!\033[0m"
  echo "\033[1;32m================================\033[0m"
  info "n8n has been installed to $INSTALL_DIR/$BINARY"

  if command_exists "$BINARY"; then
    info "Run 'n8n --help' to get started"
  else
    if [ -x "$INSTALL_DIR/$BINARY" ]; then
      info "Run '$INSTALL_DIR/$BINARY --help' to get started"
      
      if ! echo "$PATH" | tr ':' '\n' | grep -q "^$INSTALL_DIR$"; then
        echo "\033[1;33m⚠️  IMPORTANT: Your installation directory is not in your PATH\033[0m"
        
        SHELL_NAME=$(basename "$SHELL")
        case "$SHELL_NAME" in
          bash)
            SHELL_CONFIG="$HOME/.bashrc"
            ;;
          zsh)
            SHELL_CONFIG="$HOME/.zshrc"
            ;;
          fish)
            SHELL_CONFIG="$HOME/.config/fish/config.fish"
            ;;
          *)
            SHELL_CONFIG="your shell configuration file"
            ;;
        esac
        
        echo "\033[1;36m➡️  To use n8n from anywhere, add it to your PATH:\033[0m"
        case "$SHELL_NAME" in
          fish)
            echo "   \033[1m echo 'set -x PATH \$PATH $INSTALL_DIR' >> $SHELL_CONFIG\033[0m"
            ;;
          *)
            echo "   \033[1m echo 'export PATH=\$PATH:$INSTALL_DIR' >> $SHELL_CONFIG\033[0m"
            ;;
        esac
        echo "   \033[1m source $SHELL_CONFIG\033[0m"
        
        echo ""
        echo "\033[1;36m➡️  Or run this command directly:\033[0m"
        case "$SHELL_NAME" in
          fish)
            echo "   \033[1m set -x PATH \$PATH $INSTALL_DIR\033[0m"
            ;;
          *)
            echo "   \033[1m export PATH=\$PATH:$INSTALL_DIR\033[0m"
            ;;
        esac
      fi
    else
      error "Installation failed"
    fi
  fi
}

do_install
