# Skills and Agents

> [← Back to docs](README.md)

The repository ships English and Russian workflow files for all agents and skills.

Current Go MCP registry: `85` tools. Original Node parity baseline: `78` tools.

## Skills

Each skill has:

- English: `skills/<name>/SKILL.md`
- Russian: `skills/<name>/SKILL.ru.md`

| Skill | Use for |
| --- | --- |
| `chart-analysis` | Full chart analysis with screenshot context |
| `data-quality` | Source/reliability/coverage checks |
| `error-handling` | Retryability and structured statuses |
| `futures-roll` | Continuous futures and roll context |
| `indicator-scan` | Indicator signal tables |
| `json-contracts` | MCP JSON response validation |
| `llm-context` | Compact LLM-ready context |
| `market-brief` | Market brief |
| `multi-symbol-scan` | Multiple-symbol scans |
| `pine-develop` | Pine development loop |
| `pine-safe-edit` | Backup/hash/restore-safe Pine edits |
| `regression-smoke` | Regression and live smoke workflow |
| `replay-practice` | Replay practice |
| `strategy-backtesting-api` | Strategy report/trades/orders |
| `strategy-equity-plot` | Explicit Strategy Equity plot workflow |
| `strategy-report` | Strategy performance report |
| `study-model-values` | Reliable study model values/history |
| `tradingview-limit-handling` | Study-limit handling |

## Agents

Each agent has English and Russian variants for every client wrapper.

| Agent | Use for |
| --- | --- |
| `market-analyst` | Live chart reads and indicator-aware market briefs |
| `futures-analyst` | Continuous futures, roll context, bid/ask caveats |
| `performance-analyst` | Strategy Tester metrics, trades, orders, equity |

Root Claude agent files:

- `agents/market-analyst.md`, `agents/market-analyst.ru.md`
- `agents/futures-analyst.md`, `agents/futures-analyst.ru.md`
- `agents/performance-analyst.md`, `agents/performance-analyst.ru.md`

Client variants:

- Cursor: `agents/cursor/*.mdc`, `agents/cursor/*.ru.mdc`
- Cline: `agents/cline/*.md`, `agents/cline/*.ru.md`
- Windsurf: `agents/windsurf/*.md`, `agents/windsurf/*.ru.md`
- Continue: `agents/continue/*.prompt`, `agents/continue/*.ru.prompt`
- Codex: `agents/codex/*.md`, `agents/codex/*.ru.md`
- Gemini: `agents/gemini/*.md`, `agents/gemini/*.ru.md`

## Reliability Rules

Agents and skills must verify `source`, `reliability`, `status`, `coverage`, and `reliableForTradingLogic` before making trading-logic claims.

- Compatibility: run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path tool returns unavailable statuses.
- Indicator values: prefer `tradingview_study_model`.
- Strategy metrics/trades/orders: require `source: tradingview_backtesting_api` and `status: ok`.
- Equity: require explicit `Strategy Equity` Pine plot; `coverage: loaded_chart_bars` is partial and not full Strategy Tester history.
- Optional history loading: expand/scroll the chart range, wait for TradingView to load bars, repeat the data call, then compare `loaded_bar_count` and `data_points`.
- Derived equity: conditional only; do not call it native TradingView equity.
- Full native bar-by-bar Strategy Tester equity: do not implement or promise it until TradingView exposes a stable report field.
- Bid/ask: use only when `bidAskAvailable:true`.
