#!/usr/bin/env bash
# Create release archives for all supported platforms.
# Each archive contains: binaries + agents + skills + prompts + scripts + installers.
#
# Usage:
#   bash scripts/package.sh               # builds all platforms
#   VERSION=v1.2.0 bash scripts/package.sh
#
# Output: bin/releases/<name>.tar.gz (unix) / .zip (windows)
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
LDFLAGS="-ldflags=-s -w"
DIST="bin/dist"
RELEASES="bin/releases"

PLATFORMS=(
  "linux   amd64"
  "linux   arm64"
  "darwin  amd64"
  "darwin  arm64"
  "windows amd64"
  "windows arm64"
)

echo "=== tradingview-mcp-go release packager ==="
echo "Version : $VERSION"
echo "Output  : $RELEASES/"
echo

rm -rf "$DIST" && mkdir -p "$DIST" "$RELEASES"

for entry in "${PLATFORMS[@]}"; do
  GOOS=$(echo "$entry" | awk '{print $1}')
  GOARCH=$(echo "$entry" | awk '{print $2}')
  PLATFORM="${GOOS}-${GOARCH}"
  PKG="tradingview-mcp-go_${VERSION}_${PLATFORM}"
  PKGDIR="$DIST/$PKG"

  echo "→ $PLATFORM"

  EXT=""
  [ "$GOOS" = "windows" ] && EXT=".exe"

  # ── 1. Build binaries ───────────────────────────────────────────────────
  mkdir -p "$PKGDIR"
  GOOS=$GOOS GOARCH=$GOARCH go build $LDFLAGS \
      -o "$PKGDIR/tvmcp${EXT}" ./cmd/tvmcp
  GOOS=$GOOS GOARCH=$GOARCH go build $LDFLAGS \
      -o "$PKGDIR/tv${EXT}"    ./cmd/tv

  # ── 2. Agents (all client formats) ─────────────────────────────────────
  cp -r agents "$PKGDIR/agents"

  # ── 3. Skills ───────────────────────────────────────────────────────────
  cp -r skills "$PKGDIR/skills"

  # ── 4. Prompts ──────────────────────────────────────────────────────────
  cp -r prompts "$PKGDIR/prompts"

  # ── 5. Platform-appropriate scripts ─────────────────────────────────────
  mkdir -p "$PKGDIR/scripts"
  if [ "$GOOS" = "windows" ]; then
    cp scripts/pine_pull.bat   "$PKGDIR/scripts/"
    cp scripts/pine_push.bat   "$PKGDIR/scripts/"
    cp scripts/launch_tv_debug.bat  "$PKGDIR/scripts/"
    cp scripts/launch_tv_debug.vbs  "$PKGDIR/scripts/"
    cp scripts/configure-mcp.ps1    "$PKGDIR/scripts/" 2>/dev/null || true
    cp scripts/bootstrap.ps1        "$PKGDIR/scripts/" 2>/dev/null || true
  else
    cp scripts/pine_pull.sh    "$PKGDIR/scripts/"
    cp scripts/pine_push.sh    "$PKGDIR/scripts/"
    cp scripts/configure-mcp.sh    "$PKGDIR/scripts/" 2>/dev/null || true
    cp scripts/bootstrap.sh        "$PKGDIR/scripts/" 2>/dev/null || true
    if [ "$GOOS" = "darwin" ]; then
      cp scripts/launch_tv_debug_mac.sh  "$PKGDIR/scripts/"
    else
      cp scripts/launch_tv_debug_linux.sh "$PKGDIR/scripts/"
    fi
    chmod +x "$PKGDIR/scripts/"*.sh 2>/dev/null || true
  fi

  # ── 6. Local installer inside the archive ───────────────────────────────
  if [ "$GOOS" = "windows" ]; then
    # install.bat — copies tvmcp.exe + tv.exe to %PREFIX% (default: %LOCALAPPDATA%\tvmcp)
    cat > "$PKGDIR/install.bat" <<'BAT'
@echo off
setlocal
set "DEST=%~1"
if "%DEST%"=="" set "DEST=%LOCALAPPDATA%\tvmcp"
if not exist "%DEST%" mkdir "%DEST%"
copy /Y tvmcp.exe "%DEST%\tvmcp.exe" >nul
copy /Y tv.exe    "%DEST%\tv.exe"    >nul
echo Installed to %DEST%
echo Add %DEST% to your PATH if not already present.
echo Configure MCP:  scripts\configure-mcp.ps1 -Client claude -BinDir "%DEST%"
endlocal
BAT
    # install.ps1 — full PowerShell local installer
    cat > "$PKGDIR/install.ps1" <<'PS1'
param(
  [string]$Prefix = "$env:LOCALAPPDATA\tvmcp",
  [string]$Client = ""
)
if (-not (Test-Path $Prefix)) { New-Item -ItemType Directory -Path $Prefix -Force | Out-Null }
Copy-Item tvmcp.exe "$Prefix\tvmcp.exe" -Force
Copy-Item tv.exe    "$Prefix\tv.exe"    -Force

# Add to user PATH if not present
$currentPath = [Environment]::GetEnvironmentVariable("PATH","User")
if ($currentPath -notlike "*$Prefix*") {
  [Environment]::SetEnvironmentVariable("PATH","$currentPath;$Prefix","User")
  Write-Host "Added $Prefix to user PATH"
}
Write-Host "Installed: $Prefix\tvmcp.exe  $Prefix\tv.exe"

if ($Client -ne "") {
  & "$PSScriptRoot\scripts\configure-mcp.ps1" -Client $Client -BinDir $Prefix
}
PS1
  else
    # install.sh — copies binaries to PREFIX (default: /usr/local/bin)
    cat > "$PKGDIR/install.sh" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
PREFIX="${1:-/usr/local/bin}"
CLIENT="${CLIENT:-}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

if [ ! -d "$PREFIX" ]; then
  echo "Creating $PREFIX"
  mkdir -p "$PREFIX"
fi

install -m 755 "$SCRIPT_DIR/tvmcp" "$PREFIX/tvmcp"
install -m 755 "$SCRIPT_DIR/tv"    "$PREFIX/tv"
echo "Installed: $PREFIX/tvmcp  $PREFIX/tv"

if [ -n "$CLIENT" ]; then
  bash "$SCRIPT_DIR/scripts/configure-mcp.sh" --client "$CLIENT" --bin-dir "$PREFIX"
fi
SH
    chmod +x "$PKGDIR/install.sh"
  fi

  # ── 7. README ───────────────────────────────────────────────────────────
  cp README.md "$PKGDIR/README.md"

  # ── 8. Pack ─────────────────────────────────────────────────────────────
  if [ "$GOOS" = "windows" ]; then
    (cd "$DIST" && zip -qr "$REPO_ROOT/$RELEASES/${PKG}.zip" "$PKG/")
    echo "   → $RELEASES/${PKG}.zip"
  else
    tar -czf "$RELEASES/${PKG}.tar.gz" -C "$DIST" "$PKG/"
    echo "   → $RELEASES/${PKG}.tar.gz"
  fi
done

# ── Checksums ────────────────────────────────────────────────────────────────
echo
echo "Generating checksums..."
cd "$RELEASES"
if command -v sha256sum &>/dev/null; then
  sha256sum ./* > "checksums.txt"
elif command -v shasum &>/dev/null; then
  shasum -a 256 ./* > "checksums.txt"
fi
echo "Done. Archives in $RELEASES/"
ls -lh "$REPO_ROOT/$RELEASES/"
