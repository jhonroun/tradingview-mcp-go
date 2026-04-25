#!/usr/bin/env bash
# Build tradingview-mcp-go for the current platform.
# Run from repository root: bash scripts/build.sh
set -euo pipefail

mkdir -p bin

GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}
EXT=$( [ "$GOOS" = "windows" ] && echo ".exe" || echo "" )

echo "Building tvmcp${EXT} (${GOOS}/${GOARCH})..."
GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "bin/tvmcp${EXT}" ./cmd/tvmcp

echo "Building tv${EXT} (${GOOS}/${GOARCH})..."
GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "bin/tv${EXT}"    ./cmd/tv

echo "Done: bin/tvmcp${EXT}  bin/tv${EXT}"
