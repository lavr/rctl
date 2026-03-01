#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-dev}"

cd "$(dirname "$0")/.."

echo "Building rctl (version: $VERSION)..."
go build -ldflags "-X main.version=$VERSION" -o rctl ./cmd/rctl

mkdir -p "$INSTALL_DIR"
TARGET="$INSTALL_DIR/rctl"

if [ -L "$TARGET" ]; then
    echo "Removing existing symlink $TARGET"
    rm "$TARGET"
fi

cp rctl "$TARGET"
chmod +x "$TARGET"

echo "Installed rctl to $TARGET"

if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
    echo ""
    echo "Warning: $INSTALL_DIR is not in PATH."
    echo "Add to your shell profile:"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi
