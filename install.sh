#!/usr/bin/env sh

set -eu

REPO="${REPO:-chenasraf/sofmani}"
INSTALL_DIR="${INSTALL_DIR:-"$HOME/.local/bin"}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Error: required command '$1' not found in PATH" >&2
    exit 1
  }
}

need curl
need tar
need uname
need mktemp

os="$(uname -s 2>/dev/null | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m 2>/dev/null)"
case "$os" in
linux) os="linux" ;;
darwin) os="darwin" ;; # adjust if your asset names use "darwin" instead
*)
  echo "Unsupported OS: $os" >&2
  exit 1
  ;;
esac

case "$arch" in
x86_64 | amd64) arch="amd64" ;;
aarch64 | arm64) arch="arm64" ;;
*)
  echo "Unsupported architecture: $arch" >&2
  exit 1
  ;;
esac

asset="sofmani-${os}-${arch}.tar.gz"

download_url="https://github.com/${REPO}/releases/latest/download/${asset}"

# temp dir + cleanup
tmpdir="$(mktemp -d)"
cleanup() { [ -n "${tmpdir:-}" ] && rm -rf "$tmpdir"; }
trap cleanup EXIT INT HUP TERM

echo "Installing sofmani (latest) for ${os}/${arch}"
mkdir -p "$INSTALL_DIR"

echo "Downloading ${download_url} ..."

if ! curl -fsSL "$download_url" | tar -xzf - -C "$tmpdir"; then
  echo "Failed to download or extract ${download_url}" >&2
  exit 1
fi

if [ ! -f "$tmpdir/sofmani" ]; then
  echo "Extracted archive did not contain 'sofmani' binary" >&2
  exit 1
fi

echo "Installing to ${INSTALL_DIR} ..."

if ! install -m 0755 "$tmpdir/sofmani" "$INSTALL_DIR/sofmani"; then
  echo "Failed to install sofmani to ${INSTALL_DIR}" >&2
  exit 1
fi

echo "sofmani installed successfully at ${INSTALL_DIR}/sofmani"

case ":$PATH:" in
*":$INSTALL_DIR:"*) : ;; # already in PATH
*)
  echo
  printf "\033[33mNote: %s is not in PATH. Add this to your shell config:\n" "$INSTALL_DIR"
  printf "  export PATH=\"\$PATH:%s\"\033[0m\n" "${INSTALL_DIR}"
  ;;
esac
