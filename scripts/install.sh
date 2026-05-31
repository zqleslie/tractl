#!/usr/bin/env bash
# traCtl installer
#
# Usage:
#   curl -fsSL https://install.tractl.softwits.com | sh
#
#   # Specific version:
#   TRACTL_VERSION=v0.1.0 curl -fsSL https://install.tractl.softwits.com | sh
#
#   # Custom install directory:
#   TRACTL_INSTALL_DIR=/usr/local/bin curl -fsSL https://install.tractl.softwits.com | sh

set -euo pipefail

REPO="tractl/tractl"
BINARY="tractl"
INSTALL_DIR="${TRACTL_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${TRACTL_VERSION:-}"   # empty = latest

# ── Detect platform ──────────────────────────────────────────────────────────

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  OS="linux"  ;;
  Darwin) OS="darwin" ;;
  *)
    echo "error: unsupported OS: $OS"
    echo ""
    echo "Windows: download the .zip from https://github.com/${REPO}/releases"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64 | amd64)  ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "error: unsupported architecture: $ARCH"
    exit 1
    ;;
esac

ARCHIVE="tractl_${OS}_${ARCH}.tar.gz"

# ── Resolve version ───────────────────────────────────────────────────────────

if [ -z "$VERSION" ]; then
  echo "→ Fetching latest version..."
  VERSION="$(curl -fsSL \
    "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
fi

if [ -z "$VERSION" ]; then
  echo "error: could not determine latest version"
  echo "  Set TRACTL_VERSION=v0.1.0 and retry"
  exit 1
fi

echo "→ Installing tractl ${VERSION} (${OS}/${ARCH})"

# ── Download ──────────────────────────────────────────────────────────────────

BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
ARCHIVE_URL="${BASE_URL}/${ARCHIVE}"
CHECKSUM_URL="${BASE_URL}/tractl_${VERSION}_checksums.txt"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "→ Downloading ${ARCHIVE}..."
curl -fsSL --progress-bar "${ARCHIVE_URL}" -o "${TMP}/${ARCHIVE}"

# ── Verify checksum ───────────────────────────────────────────────────────────

echo "→ Verifying checksum..."
curl -fsSL "${CHECKSUM_URL}" -o "${TMP}/checksums.txt"

pushd "$TMP" > /dev/null
if command -v sha256sum &>/dev/null; then
  grep "${ARCHIVE}" checksums.txt | sha256sum --check --status
elif command -v shasum &>/dev/null; then
  grep "${ARCHIVE}" checksums.txt | shasum -a 256 --check --status
else
  echo "  warning: no sha256sum or shasum found, skipping verification"
fi
popd > /dev/null

echo "  ✓ Checksum verified"

# ── Extract and install ───────────────────────────────────────────────────────

echo "→ Extracting..."
tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP"

echo "→ Installing to ${INSTALL_DIR}/${BINARY}..."
mkdir -p "$INSTALL_DIR"
install -m 755 "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

# ── Verify ───────────────────────────────────────────────────────────────────

INSTALLED="$("${INSTALL_DIR}/${BINARY}" version 2>/dev/null || echo 'installed')"
echo "  ✓ ${INSTALLED}"

# ── PATH guidance ─────────────────────────────────────────────────────────────

echo ""
echo "✓ tractl ${VERSION} installed"
echo ""

if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
  SHELL_NAME="$(basename "${SHELL:-bash}")"
  case "$SHELL_NAME" in
    zsh)  RC="$HOME/.zshrc"  ;;
    bash) RC="$HOME/.bashrc" ;;
    *)    RC="your shell config" ;;
  esac

  echo "  ${INSTALL_DIR} is not in your PATH."
  echo "  Add it:"
  echo ""
  echo "    echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ${RC}"
  echo "    source ${RC}"
  echo ""
fi

# ── macOS quarantine note ─────────────────────────────────────────────────────

if [ "$OS" = "darwin" ]; then
  echo "  macOS note: if you see a Gatekeeper warning, run:"
  echo "    xattr -d com.apple.quarantine ${INSTALL_DIR}/${BINARY}"
  echo ""
fi
