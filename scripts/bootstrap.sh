#!/usr/bin/env bash
# tradingview-mcp-go — bootstrap installer
#
# One-line install (Linux / macOS):
#   curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | bash
#
# With options:
#   curl -fsSL .../bootstrap.sh | CLIENT=cursor PREFIX=$HOME/.local/bin bash
#   bash bootstrap.sh --client cursor --prefix $HOME/.local/bin
#   bash bootstrap.sh --client claude --version v1.2.0
#   bash bootstrap.sh --list-clients
#
# Options:
#   --client NAME     configure MCP in this client after install (optional)
#   --prefix PATH     install binaries here (default: /usr/local/bin)
#   --version TAG     install specific version (default: latest)
#   --list-clients    print supported client names and exit
#   --no-configure    skip client MCP configuration even if --client given
set -euo pipefail

REPO="jhonroun/tradingview-mcp-go"
GH_API="https://api.github.com/repos/$REPO"
GH_RAW="https://raw.githubusercontent.com/$REPO/main"

VERSION="${VERSION:-latest}"
PREFIX="${PREFIX:-/usr/local/bin}"
CLIENT="${CLIENT:-}"
NO_CONFIGURE=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --client|-c)      CLIENT="$2"; shift 2 ;;
    --prefix|-p)      PREFIX="$2"; shift 2 ;;
    --version|-v)     VERSION="$2"; shift 2 ;;
    --no-configure)   NO_CONFIGURE=true; shift ;;
    --list-clients)
      echo "claude cursor cline windsurf continue codex gemini"
      exit 0 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# ── detect OS / arch ──────────────────────────────────────────────────────────

UNAME_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
UNAME_ARCH=$(uname -m)

case "$UNAME_OS" in
  linux)  GOOS="linux"  ;;
  darwin) GOOS="darwin" ;;
  msys*|mingw*|cygwin*) GOOS="windows" ;;
  *) echo "Unsupported OS: $UNAME_OS"; exit 1 ;;
esac

case "$UNAME_ARCH" in
  x86_64|amd64) GOARCH="amd64" ;;
  aarch64|arm64) GOARCH="arm64" ;;
  *) echo "Unsupported arch: $UNAME_ARCH"; exit 1 ;;
esac

EXT=""
[ "$GOOS" = "windows" ] && EXT=".exe"

echo "=== tradingview-mcp-go installer ==="
echo "OS/Arch : $GOOS/$GOARCH"
echo "Prefix  : $PREFIX"

# ── resolve version ───────────────────────────────────────────────────────────

if [[ "$VERSION" == "latest" ]]; then
  echo -n "Fetching latest release... "
  if command -v curl &>/dev/null; then
    VERSION=$(curl -fsSL "$GH_API/releases/latest" \
      | grep '"tag_name"' | head -1 | cut -d'"' -f4)
  elif command -v wget &>/dev/null; then
    VERSION=$(wget -qO- "$GH_API/releases/latest" \
      | grep '"tag_name"' | head -1 | cut -d'"' -f4)
  else
    echo "Error: curl or wget required"; exit 1
  fi
  echo "$VERSION"
fi

# ── download ──────────────────────────────────────────────────────────────────

ARCHIVE="tradingview-mcp-go_${VERSION}_${GOOS}-${GOARCH}"
if [[ "$GOOS" == "windows" ]]; then
  ARCHIVE_FILE="${ARCHIVE}.zip"
else
  ARCHIVE_FILE="${ARCHIVE}.tar.gz"
fi

URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE_FILE"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading $ARCHIVE_FILE ..."
if command -v curl &>/dev/null; then
  curl -fsSL "$URL" -o "$TMPDIR/$ARCHIVE_FILE"
else
  wget -qO "$TMPDIR/$ARCHIVE_FILE" "$URL"
fi

# ── extract ───────────────────────────────────────────────────────────────────

echo "Extracting..."
if [[ "$GOOS" == "windows" ]]; then
  unzip -q "$TMPDIR/$ARCHIVE_FILE" -d "$TMPDIR"
else
  tar -xzf "$TMPDIR/$ARCHIVE_FILE" -C "$TMPDIR"
fi

PKGDIR="$TMPDIR/$ARCHIVE"

# ── install binaries ──────────────────────────────────────────────────────────

if [[ ! -d "$PREFIX" ]]; then
  echo "Creating $PREFIX"
  mkdir -p "$PREFIX"
fi

install -m 755 "$PKGDIR/tvmcp${EXT}" "$PREFIX/tvmcp${EXT}"
install -m 755 "$PKGDIR/tv${EXT}"    "$PREFIX/tv${EXT}"
echo "Installed: $PREFIX/tvmcp${EXT}  $PREFIX/tv${EXT}"

# ── check PATH ────────────────────────────────────────────────────────────────

if ! echo "$PATH" | grep -q "$PREFIX"; then
  echo ""
  echo "Warning: $PREFIX is not in your PATH."
  echo "Add to your shell profile:"
  echo "  export PATH=\"\$PATH:$PREFIX\""
fi

# ── install agents + skills + prompts alongside binaries (optional) ───────────

SHARE_DIR="$(dirname "$PREFIX")/share/tradingview-mcp-go"
if [[ -d "$PKGDIR/agents" ]]; then
  mkdir -p "$SHARE_DIR"
  cp -r "$PKGDIR/agents"  "$SHARE_DIR/"
  cp -r "$PKGDIR/skills"  "$SHARE_DIR/"
  cp -r "$PKGDIR/prompts" "$SHARE_DIR/"
  echo "Assets:    $SHARE_DIR/{agents,skills,prompts}"
fi

# ── configure MCP client ──────────────────────────────────────────────────────

if [[ -n "$CLIENT" && "$NO_CONFIGURE" == "false" ]]; then
  echo ""
  echo "Configuring MCP for client: $CLIENT"
  if [[ -f "$PKGDIR/scripts/configure-mcp.sh" ]]; then
    BIN_DIR="$PREFIX" bash "$PKGDIR/scripts/configure-mcp.sh" --client "$CLIENT"
  else
    # Fall back to downloading the configure script
    CONF_TMP=$(mktemp)
    curl -fsSL "$GH_RAW/scripts/configure-mcp.sh" -o "$CONF_TMP"
    BIN_DIR="$PREFIX" bash "$CONF_TMP" --client "$CLIENT"
    rm -f "$CONF_TMP"
  fi
fi

# ── done ─────────────────────────────────────────────────────────────────────

echo ""
echo "=== Installation complete ==="
echo "  tv status      — verify CDP connection"
echo "  tv launch      — start TradingView with CDP"
echo ""
if [[ -n "$CLIENT" ]]; then
  echo "MCP configured for: $CLIENT"
  echo "Restart $CLIENT to activate the tradingview MCP server."
else
  echo "To configure MCP for your client:"
  echo "  bash configure-mcp.sh --client <claude|cursor|cline|windsurf|continue|codex|gemini>"
fi
