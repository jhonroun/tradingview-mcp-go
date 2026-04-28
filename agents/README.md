# Agents

Agent definitions for `tradingview-mcp-go`, in native formats for supported AI clients.

Current status:

- Current Go MCP registry: `85` tools.
- Historical Node.js parity baseline: `78` tools.
- English and Russian variants are provided for every agent and client wrapper.
- Source prompts live in `prompts/`.

## Agents

| Agent | EN source | RU source | Use for |
| --- | --- | --- | --- |
| `market-analyst` | `prompts/market-analyst.md` | `prompts/market-analyst.ru.md` | Live chart reads, indicator briefs, market bias |
| `futures-analyst` | `prompts/futures-analyst.md` | `prompts/futures-analyst.ru.md` | Continuous futures, roll context, bid/ask caveats |
| `performance-analyst` | `prompts/performance-analyst.md` | `prompts/performance-analyst.ru.md` | Strategy results, trades, orders, equity coverage |

All agents are source/reliability aware. They must not treat UI/canvas values, unavailable bid/ask, derived equity, or loaded-bars-only equity as fully reliable trading data.

## File Layout

```text
agents/
  market-analyst.md
  market-analyst.ru.md
  futures-analyst.md
  futures-analyst.ru.md
  performance-analyst.md
  performance-analyst.ru.md
  cursor/*.mdc / *.ru.mdc
  cline/*.md / *.ru.md
  windsurf/*.md / *.ru.md
  continue/*.prompt / *.ru.prompt
  codex/*.md / *.ru.md
  gemini/*.md / *.ru.md
```

## Skills

Skills are workflow instructions in `skills/<name>/SKILL.md`. Russian variants are in `skills/<name>/SKILL.ru.md`.

Current skills:

| Skill | Purpose |
| --- | --- |
| `chart-analysis` | Full chart setup analysis with data-quality checks |
| `data-quality` | Verify source/reliability/coverage before using data |
| `error-handling` | Classify MCP errors and structured statuses |
| `futures-roll` | Continuous futures roll context |
| `indicator-scan` | Indicator signal table with study-model validation |
| `json-contracts` | Validate response fields and contracts |
| `llm-context` | Build compact LLM-ready market context |
| `market-brief` | Market brief from price, volume, indicators |
| `multi-symbol-scan` | Scan multiple symbols |
| `pine-develop` | Pine development loop |
| `pine-safe-edit` | Backup/hash/restore-safe Pine editing |
| `regression-smoke` | Regression and smoke workflow |
| `replay-practice` | Replay-mode practice |
| `strategy-backtesting-api` | Strategy Tester report/trades/orders |
| `strategy-equity-plot` | Explicit Strategy Equity plot workflow |
| `strategy-report` | Strategy report generation |
| `study-model-values` | Reliable indicator current/history values |
| `tradingview-limit-handling` | Study-limit detection and safe recovery |

## Install Examples

### Claude Code

```bash
claude --agent agents/market-analyst.md
claude --agent agents/market-analyst.ru.md
```

### Cursor

```bash
mkdir -p .cursor/rules
cp agents/cursor/market-analyst.mdc .cursor/rules/
cp agents/cursor/market-analyst.ru.mdc .cursor/rules/
```

### Cline

```bash
mkdir -p .clinerules
cp agents/cline/market-analyst.md .clinerules/
cp agents/cline/market-analyst.ru.md .clinerules/
```

### Windsurf

```bash
cat agents/windsurf/market-analyst.md >> .windsurfrules
cat agents/windsurf/market-analyst.ru.md >> .windsurfrules
```

### Continue

```bash
mkdir -p .continue/prompts
cp agents/continue/market-analyst.prompt .continue/prompts/
cp agents/continue/market-analyst.ru.prompt .continue/prompts/
```

### Codex

```bash
codex --instructions "$(cat agents/codex/market-analyst.md)" "Analyze the current chart"
codex --instructions "$(cat agents/codex/market-analyst.ru.md)" "Разбери текущий график"
```

### Gemini

```bash
gemini --system "$(cat agents/gemini/market-analyst.md)" "Analyze the current chart"
gemini --system "$(cat agents/gemini/market-analyst.ru.md)" "Разбери текущий график"
```

## Data Reliability Rules

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- `tradingview_study_model`: reliable numeric indicator values, unstable internal path.
- `tradingview_backtesting_api`: reliable strategy report when `status: ok`, unstable internal path.
- `tradingview_strategy_plot`: reliable equity plot values for `coverage: loaded_chart_bars`.
- `tradingview_ui_data_window`: localized display fallback, not reliable for trading logic.
- `bidAskAvailable:false`: bid/ask spread unavailable.
- Derived equity is conditional and must not be presented as native full Strategy Tester equity.
- Optional history loading is best-effort: expand/scroll chart range, repeat the data call, compare `loaded_bar_count` / `data_points`, and keep `coverage: loaded_chart_bars`.
- Full native bar-by-bar Strategy Tester equity is not a release target until TradingView exposes a stable report field.
