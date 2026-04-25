#!/usr/bin/env bash
# Install tvmcp and tv into /usr/local/bin (or PREFIX).
# Usage: sudo bash scripts/install.sh
#        PREFIX=$HOME/.local/bin bash scripts/install.sh
set -euo pipefail

PREFIX=${PREFIX:-/usr/local/bin}

if [ ! -d "$PREFIX" ]; then
  echo "Creating $PREFIX"
  mkdir -p "$PREFIX"
fi

echo "Building..."
bash "$(dirname "$0")/build.sh"

echo "Installing to $PREFIX..."
install -m 755 bin/tvmcp "$PREFIX/tvmcp"
install -m 755 bin/tv    "$PREFIX/tv"

echo "Done. Verify with: tvmcp --help && tv status"
