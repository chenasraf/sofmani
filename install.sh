#!/usr/bin/env bash

REPO="chenasraf/sofmani"
INSTALL_DIR="$HOME/.local/bin"
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep "tag_name" | cut -d '"' -f 4)
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/sofmani-linux-amd64.tar.gz"

mkdir -p "$INSTALL_DIR"
curl -L "$DOWNLOAD_URL" | tar -xz -C "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/sofmani"

echo "sofmani installed successfully!"
