# Agents

Agent definitions for the `tradingview-mcp-go` MCP server,  
in native formats for each supported AI client.

## Quick install (bootstrap)

The bootstrap installer downloads pre-built binaries and optionally configures
the MCP server for your AI client — no cloning or building required.

### Linux / macOS

```bash
# Install only
curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | bash

# Install + configure Claude Code
curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | CLIENT=claude bash

# Install + configure Cursor
curl -fsSL https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.sh | CLIENT=cursor bash
```

### Windows (PowerShell)

```powershell
# Install only
iwr -useb https://raw.githubusercontent.com/jhonroun/tradingview-mcp-go/main/scripts/bootstrap.ps1 | iex

# Install + configure Claude Code
.\bootstrap.ps1 -Client claude

# Install + configure Cursor with custom path
.\bootstrap.ps1 -Client cursor -Prefix "C:\tools\tvmcp"
```

Supported clients: `claude` · `cursor` · `cline` · `windsurf` · `continue` · `codex` · `gemini`

---

```
agents/
  performance-analyst.md        ← Claude Code   (Claude Agents SDK)
  cursor/
    performance-analyst.mdc     ← Cursor        (.cursor/rules/)
  cline/
    performance-analyst.md      ← Cline         (.clinerules/)
  windsurf/
    performance-analyst.md      ← Windsurf      (.windsurfrules)
  continue/
    performance-analyst.prompt  ← Continue      (.continue/prompts/)
  codex/
    performance-analyst.md      ← OpenAI Codex  (AGENTS.md / --instructions)
  gemini/
    performance-analyst.md      ← Gemini CLI    (GEMINI.md / --system)
```

The underlying system prompt is the same across all clients.  
Source of truth: [`prompts/performance-analyst.md`](../prompts/performance-analyst.md)

---

## Claude Code

**File:** `agents/performance-analyst.md`  
**Format:** Claude Agents SDK (YAML frontmatter + markdown body)

```bash
# Run agent directly
claude --agent agents/performance-analyst.md

# Or start a session and reference it
claude "Analyze the current TradingView strategy" --agent agents/performance-analyst.md
```

---

## Cursor

**File:** `agents/cursor/performance-analyst.mdc`  
**Format:** Cursor Rules (`.mdc` with YAML frontmatter)

**Install** (project-level, recommended):
```bash
mkdir -p .cursor/rules
cp agents/cursor/performance-analyst.mdc .cursor/rules/performance-analyst.mdc
```

**Use:** In Cursor chat, type `@performance-analyst` or Cursor will apply it automatically when the description matches your request (`alwaysApply: false`).

**Global install** (all projects):  
Cursor → Settings → Rules for AI → paste the file body.

---

## Cline

**File:** `agents/cline/performance-analyst.md`  
**Format:** Cline rules (plain markdown, no frontmatter)

**Install** (project-level, per-rule directory):
```bash
mkdir -p .clinerules
cp agents/cline/performance-analyst.md .clinerules/performance-analyst.md
```

Cline loads all `.md` files from `.clinerules/` automatically.

**Single-file install:**
```bash
cp agents/cline/performance-analyst.md .clinerules
```

**Global install:**  
VS Code → Extensions → Cline → Settings → Custom Instructions → paste the file content.

---

## Windsurf

**File:** `agents/windsurf/performance-analyst.md`  
**Format:** Windsurf rules (plain markdown)

**Install** (project-level, append to rules file):
```bash
cat agents/windsurf/performance-analyst.md >> .windsurfrules
```

**Global install:**  
Windsurf → Settings → AI Rules → paste the file content.

---

## Continue

**File:** `agents/continue/performance-analyst.prompt`  
**Format:** Continue prompt template (`.prompt` file)

**Install:**
```bash
mkdir -p .continue/prompts
cp agents/continue/performance-analyst.prompt .continue/prompts/performance-analyst.prompt
```

**Use:** In Continue chat, type `/performance-analyst` to invoke the prompt.

---

## OpenAI Codex CLI

**File:** `agents/codex/performance-analyst.md`  
**Format:** Markdown instruction file (`AGENTS.md` convention)

**Option 1 — project AGENTS.md:**
```bash
cp agents/codex/performance-analyst.md AGENTS.md
# Codex CLI reads AGENTS.md automatically from the working directory
codex "Analyze the current TradingView strategy"
```

**Option 2 — inline flag:**
```bash
codex --instructions "$(cat agents/codex/performance-analyst.md)" \
      "Analyze the current TradingView strategy"
```

**Option 3 — environment variable:**
```bash
export CODEX_SYSTEM_PROMPT="$(cat agents/codex/performance-analyst.md)"
codex "Analyze the current TradingView strategy"
```

---

## Gemini CLI

**File:** `agents/gemini/performance-analyst.md`  
**Format:** Markdown instruction file (`GEMINI.md` convention)

**Option 1 — project GEMINI.md:**
```bash
cp agents/gemini/performance-analyst.md GEMINI.md
# Gemini CLI reads GEMINI.md automatically from the working directory
gemini "Analyze the current TradingView strategy"
```

**Option 2 — inline flag:**
```bash
gemini --system "$(cat agents/gemini/performance-analyst.md)" \
       "Analyze the current TradingView strategy"
```

---

## Any other MCP client

1. Open `prompts/performance-analyst.md`
2. Copy the body (starting from "You are a trading strategy…")
3. Paste it as the **system prompt** / **instructions** / **custom rules** in your client's settings

The MCP tools are provided by the running `tvmcp` server — the agent works regardless of which AI model or client you use, as long as it's connected to `tvmcp` via MCP.
