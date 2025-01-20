#!/usr/bin/env bash

REPO="chenasraf/sofmani"
if [[ -z "$INSTALL_DIR" ]]; then
  INSTALL_DIR="$HOME/.local/bin"
fi
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep "tag_name" | cut -d '"' -f 4)
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/sofmani-linux-amd64.tar.gz"
TEMP_DIR=$(mktemp -d)

echo "Installing sofmani $LATEST_RELEASE"
mkdir -p "$INSTALL_DIR"

echo "Downloading $DOWNLOAD_URL..."
echo

if ! curl -L "$DOWNLOAD_URL" | tar -xz -C "$TEMP_DIR"; then
  echo "Failed to download $DOWNLOAD_URL"
  exit 1
fi

echo
echo "Installing binary to $INSTALL_DIR..."

if ! mv "$TEMP_DIR/sofmani" "$INSTALL_DIR/sofmani"; then
  echo "Failed to move sofmani binary to $INSTALL_DIR"
  exit 1
fi

chmod +x "$INSTALL_DIR/sofmani"

echo "sofmani installed successfully!"
