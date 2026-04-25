#!/usr/bin/env bash
# Push scripts/current.pine into the TradingView Pine editor and compile.
# Mirrors original pine_push.js behaviour.
# Usage: bash scripts/pine_push.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PINE_FILE="$SCRIPT_DIR/current.pine"
TV="${TV:-tv}"

if [ ! -f "$PINE_FILE" ]; then
  echo "Error: $PINE_FILE not found. Run pine_pull.sh first or create the file manually."
  exit 1
fi

LINES=$(wc -l < "$PINE_FILE")
echo "Pushing ${LINES} lines from scripts/current.pine..."

# Inject source into Monaco editor
"$TV" pine set "$(cat "$PINE_FILE")"

echo "Compiling..."
"$TV" pine smart-compile

echo "Checking errors..."
"$TV" pine errors
