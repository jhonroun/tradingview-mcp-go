#!/usr/bin/env bash
# Pull the current Pine Script source from the TradingView editor.
# Saves to scripts/current.pine (mirrors original pine_pull.js).
# Usage: bash scripts/pine_pull.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TV="${TV:-tv}"

"$TV" pine get > "$SCRIPT_DIR/current.pine"
LINES=$(wc -l < "$SCRIPT_DIR/current.pine")
echo "Pulled ${LINES} lines → scripts/current.pine"
