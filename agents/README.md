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

```text
agents/
  market-analyst.md             ← Claude Code   (Claude Agents SDK)
  futures-analyst.md            ← Claude Code   (Claude Agents SDK)
  performance-analyst.md        ← Claude Code   (Claude Agents SDK)
  cursor/
    market-analyst.mdc          ← Cursor        (.cursor/rules/)
    futures-analyst.mdc         ← Cursor        (.cursor/rules/)
    performance-analyst.mdc     ← Cursor        (.cursor/rules/)
  cline/
    market-analyst.md           ← Cline         (.clinerules/)
    futures-analyst.md          ← Cline         (.clinerules/)
    performance-analyst.md      ← Cline         (.clinerules/)
  windsurf/
    market-analyst.md           ← Windsurf      (.windsurfrules)
    futures-analyst.md          ← Windsurf      (.windsurfrules)
    performance-analyst.md      ← Windsurf      (.windsurfrules)
  continue/
    market-analyst.prompt       ← Continue      (.continue/prompts/)
    futures-analyst.prompt      ← Continue      (.continue/prompts/)
    performance-analyst.prompt  ← Continue      (.continue/prompts/)
  codex/
    market-analyst.md           ← OpenAI Codex  (AGENTS.md / --instructions)
    futures-analyst.md          ← OpenAI Codex  (AGENTS.md / --instructions)
    performance-analyst.md      ← OpenAI Codex  (AGENTS.md / --instructions)
  gemini/
    market-analyst.md           ← Gemini CLI    (GEMINI.md / --system)
    futures-analyst.md          ← Gemini CLI    (GEMINI.md / --system)
    performance-analyst.md      ← Gemini CLI    (GEMINI.md / --system)
```

Sources of truth:

- [`prompts/market-analyst.md`](../prompts/market-analyst.md)
- [`prompts/futures-analyst.md`](../prompts/futures-analyst.md)
- [`prompts/performance-analyst.md`](../prompts/performance-analyst.md)

---

## Agents

| Agent | Primary tools | Use for |
| ----- | ------------- | ------- |
| `market-analyst` | `chart_context_for_llm`, `market_summary`, `indicator_state` | Live chart reads, indicator briefs, bias |
| `futures-analyst` | `continuous_contract_context`, `market_summary`, `indicator_state` | Continuous contracts, roll timing, spreads |
| `performance-analyst` | `data_get_strategy_results`, `data_get_trades`, `data_get_equity` | Strategy backtests, equity curve, trade stats |

All agents are Phase 5-aware: they expect `plots` arrays from `data_get_study_values`, always-present `bid`/`ask`/`change`/`change_pct` in `quote_get`, and `exchange`/`ticker`/`pane_count` in `chart_get_state`.

---

## Skills

Skills extend agent capabilities with specific workflows. Place them in `skills/` and reference them from agent prompts or call them directly in the AI client.

| Skill | File | Use for |
| ----- | ---- | ------- |
| `llm-context` | `skills/llm-context/SKILL.md` | Single-call LLM context snapshot before analysis |
| `market-brief` | `skills/market-brief/SKILL.md` | Structured market briefing (price action, volume, indicators) |
| `indicator-scan` | `skills/indicator-scan/SKILL.md` | Scan multiple indicators across a watchlist |
| `futures-roll` | `skills/futures-roll/SKILL.md` | Detect and analyze futures roll timing |
| `chart-analysis` | `skills/chart-analysis/SKILL.md` | Full chart setup analysis with screenshot |
| `multi-symbol-scan` | `skills/multi-symbol-scan/SKILL.md` | Batch scan across symbols with `batch_run` |
| `pine-develop` | `skills/pine-develop/SKILL.md` | Pine Script development workflow |
| `replay-practice` | `skills/replay-practice/SKILL.md` | Historical replay practice sessions |
| `strategy-report` | `skills/strategy-report/SKILL.md` | Strategy backtest reporting |
| `json-contracts` | `skills/json-contracts/SKILL.md` | Verify Phase 5 JSON response contracts |
| `error-handling` | `skills/error-handling/SKILL.md` | Classify and recover from MCP tool errors |

---

## Claude Code

**Format:** Claude Agents SDK (YAML frontmatter + markdown body)

```bash
claude --agent agents/market-analyst.md
claude --agent agents/futures-analyst.md
claude --agent agents/performance-analyst.md
```

---

## Cursor

**Format:** Cursor Rules (`.mdc` with YAML frontmatter)

**Install** (project-level):

```bash
mkdir -p .cursor/rules
cp agents/cursor/market-analyst.mdc .cursor/rules/
cp agents/cursor/futures-analyst.mdc .cursor/rules/
cp agents/cursor/performance-analyst.mdc .cursor/rules/
```

**Use:** Type `@market-analyst`, `@futures-analyst`, or `@performance-analyst` in Cursor chat,
or Cursor applies the rule automatically when the description matches your request (`alwaysApply: false`).

**Global install:** Cursor → Settings → Rules for AI → paste the file body.

---

## Cline

**Format:** Cline rules (plain markdown, no frontmatter)

**Install** (project-level):

```bash
mkdir -p .clinerules
cp agents/cline/market-analyst.md .clinerules/
cp agents/cline/futures-analyst.md .clinerules/
cp agents/cline/performance-analyst.md .clinerules/
```

Cline loads all `.md` files from `.clinerules/` automatically.

**Global install:** VS Code → Extensions → Cline → Settings → Custom Instructions → paste the file content.

---

## Windsurf

**Format:** Windsurf rules (plain markdown)

**Install** (project-level, append to rules file):

```bash
cat agents/windsurf/market-analyst.md >> .windsurfrules
cat agents/windsurf/futures-analyst.md >> .windsurfrules
cat agents/windsurf/performance-analyst.md >> .windsurfrules
```

**Global install:** Windsurf → Settings → AI Rules → paste the file content.

---

## Continue

**Format:** Continue prompt template (`.prompt` file)

**Install:**

```bash
mkdir -p .continue/prompts
cp agents/continue/market-analyst.prompt .continue/prompts/
cp agents/continue/futures-analyst.prompt .continue/prompts/
cp agents/continue/performance-analyst.prompt .continue/prompts/
```

**Use:** In Continue chat, type `/market-analyst`, `/futures-analyst`, or `/performance-analyst`.

---

## OpenAI Codex CLI

**Format:** Markdown instruction file (`AGENTS.md` convention)

**Option 1 — project AGENTS.md** (one agent at a time):

```bash
cp agents/codex/market-analyst.md AGENTS.md
codex "What is the current chart setup?"
```

**Option 2 — inline flag:**

```bash
codex --instructions "$(cat agents/codex/futures-analyst.md)" \
      "Analyze NG1! roll timing"
```

**Option 3 — environment variable:**

```bash
export CODEX_SYSTEM_PROMPT="$(cat agents/codex/performance-analyst.md)"
codex "Analyze the current TradingView strategy"
```

---

## Gemini CLI

**Format:** Markdown instruction file (`GEMINI.md` convention)

**Option 1 — project GEMINI.md** (one agent at a time):

```bash
cp agents/gemini/market-analyst.md GEMINI.md
gemini "What is AAPL doing on the daily?"
```

**Option 2 — inline flag:**

```bash
gemini --system "$(cat agents/gemini/futures-analyst.md)" \
       "Check ES1! roll status"
```

---

## Any other MCP client

1. Open the relevant file under `prompts/`
2. Copy the body (starting from "You are a …")
3. Paste it as the **system prompt** / **instructions** / **custom rules** in your client's settings

The MCP tools are provided by the running `tvmcp` server — the agents work regardless of which AI model or client you use, as long as it is connected to `tvmcp` via MCP.
