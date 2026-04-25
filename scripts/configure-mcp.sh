#!/usr/bin/env bash
# Configure the tvmcp MCP server in an AI client's config file.
#
# Usage:
#   bash configure-mcp.sh --client claude
#   bash configure-mcp.sh --client cursor  --bin-dir /usr/local/bin
#   bash configure-mcp.sh --client windsurf
#   bash configure-mcp.sh --client continue
#   bash configure-mcp.sh --list
#
# Supported clients: claude, cursor, cline, windsurf, continue, codex, gemini
set -euo pipefail

BIN_DIR="${BIN_DIR:-/usr/local/bin}"
CLIENT=""
LIST_ONLY=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --client|-c) CLIENT="$2"; shift 2 ;;
    --bin-dir|-b) BIN_DIR="$2"; shift 2 ;;
    --list|-l) LIST_ONLY=true; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

TVMCP_PATH="$BIN_DIR/tvmcp"

if $LIST_ONLY; then
  echo "Supported clients:"
  echo "  claude    — ~/.claude.json                  (Claude Code)"
  echo "  cursor    — ~/.cursor/mcp.json              (Cursor)"
  echo "  cline     — ~/.cline/mcp_settings.json      (Cline VS Code extension)"
  echo "  windsurf  — ~/.codeium/windsurf/mcp_config.json (Windsurf)"
  echo "  continue  — ~/.continue/config.json         (Continue)"
  echo "  codex     — ~/.config/openai/codex.json     (OpenAI Codex CLI)"
  echo "  gemini    — ~/.gemini/settings.json         (Gemini CLI)"
  exit 0
fi

if [[ -z "$CLIENT" ]]; then
  echo "Usage: $0 --client <name> [--bin-dir PATH]"
  echo "       $0 --list"
  exit 1
fi

# ── helpers ──────────────────────────────────────────────────────────────────

require_jq() {
  if ! command -v jq &>/dev/null; then
    echo "Note: jq not found — writing config without pretty-printing."
    JQ_AVAILABLE=false
  else
    JQ_AVAILABLE=true
  fi
}

ensure_dir() { mkdir -p "$(dirname "$1")"; }

merge_mcp_entry() {
  local cfg="$1"
  local key="$2"      # JSON path like .mcpServers.tradingview
  local entry="$3"    # JSON object to set

  ensure_dir "$cfg"

  if [[ ! -f "$cfg" ]]; then
    echo "{}" > "$cfg"
  fi

  if $JQ_AVAILABLE; then
    local tmp
    tmp=$(mktemp)
    jq "$key = $entry" "$cfg" > "$tmp" && mv "$tmp" "$cfg"
  else
    # Fallback: append a note (user must merge manually)
    echo ""
    echo "Could not auto-merge (jq not available). Add this to $cfg manually:"
    echo "$entry"
  fi
}

ok() { echo "✓  $1"; }

# ── client handlers ───────────────────────────────────────────────────────────

configure_claude() {
  local cfg="$HOME/.claude.json"
  require_jq
  local entry="{\"command\": \"$TVMCP_PATH\"}"
  merge_mcp_entry "$cfg" '.mcpServers.tradingview' "$entry"
  ok "Claude Code: $cfg → mcpServers.tradingview"
}

configure_cursor() {
  local cfg="$HOME/.cursor/mcp.json"
  require_jq
  local entry="{\"command\": \"$TVMCP_PATH\"}"
  merge_mcp_entry "$cfg" '.mcpServers.tradingview' "$entry"
  ok "Cursor: $cfg → mcpServers.tradingview"
}

configure_cline() {
  # Cline stores MCP config in VS Code user settings or a dedicated file.
  # We write the dedicated file; VS Code settings require manual merge.
  local cfg="$HOME/.cline/mcp_settings.json"
  require_jq
  local entry="{\"command\": \"$TVMCP_PATH\"}"
  merge_mcp_entry "$cfg" '.mcpServers.tradingview' "$entry"
  ok "Cline: $cfg → mcpServers.tradingview"
  echo "   Also add to VS Code settings.json:"
  echo "   \"cline.mcpServers\": {\"tradingview\": {\"command\": \"$TVMCP_PATH\"}}"
}

configure_windsurf() {
  local cfg="$HOME/.codeium/windsurf/mcp_config.json"
  require_jq
  local entry="{\"command\": \"$TVMCP_PATH\"}"
  merge_mcp_entry "$cfg" '.mcpServers.tradingview' "$entry"
  ok "Windsurf: $cfg → mcpServers.tradingview"
}

configure_continue() {
  local cfg="$HOME/.continue/config.json"
  require_jq
  ensure_dir "$cfg"
  if [[ ! -f "$cfg" ]]; then
    echo '{"models":[],"slashCommands":[],"mcpServers":{}}' > "$cfg"
  fi
  local entry="{\"command\": \"$TVMCP_PATH\"}"
  merge_mcp_entry "$cfg" '.mcpServers.tradingview' "$entry"
  ok "Continue: $cfg → mcpServers.tradingview"
}

configure_codex() {
  local cfg="$HOME/.config/openai/codex.json"
  require_jq
  local entry="{\"command\": \"$TVMCP_PATH\"}"
  merge_mcp_entry "$cfg" '.mcpServers.tradingview' "$entry"
  ok "Codex CLI: $cfg → mcpServers.tradingview"
}

configure_gemini() {
  local cfg="$HOME/.gemini/settings.json"
  require_jq
  local entry="{\"command\": \"$TVMCP_PATH\"}"
  merge_mcp_entry "$cfg" '.mcpServers.tradingview' "$entry"
  ok "Gemini CLI: $cfg → mcpServers.tradingview"
}

# ── dispatch ──────────────────────────────────────────────────────────────────

echo "Configuring MCP server: $TVMCP_PATH"
echo "Client: $CLIENT"
echo

case "$CLIENT" in
  claude)   configure_claude   ;;
  cursor)   configure_cursor   ;;
  cline)    configure_cline    ;;
  windsurf) configure_windsurf ;;
  continue) configure_continue ;;
  codex)    configure_codex    ;;
  gemini)   configure_gemini   ;;
  *)
    echo "Unknown client: $CLIENT"
    echo "Run '$0 --list' to see supported clients."
    exit 1
    ;;
esac

echo
echo "Done. Restart $CLIENT to pick up the new MCP server."
echo "Verify: tv status"
