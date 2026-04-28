---
name: multi-symbol-scan
description: Scan multiple symbols for setups, patterns, or strategy performance. Use when comparing across instruments or screening for opportunities.
---

# Multi-Symbol Scanner

You are scanning multiple symbols for trading setups or comparing performance.

## Step 1: Define the Scan

Determine:
- **Symbols**: Which instruments to scan (user-provided or watchlist via `watchlist_get`)
- **Timeframe**: Which timeframe to analyze
- **Criteria**: What to look for (indicator values, strategy results, visual patterns)

## Step 2: Run the Scan

### For Strategy Performance Comparison
Use `batch_run` with action `get_strategy_results`:
```json
{
  "symbols": ["ES1!", "NQ1!", "YM1!", "RTY1!"],
  "timeframes": ["15"],
  "action": "get_strategy_results"
}
```

### For Screenshot Comparison
Use `batch_run` with action `screenshot`:
```json
{
  "symbols": ["AAPL", "MSFT", "GOOGL", "AMZN"],
  "timeframes": ["D"],
  "action": "screenshot"
}
```

### For Custom Analysis (per-symbol)
Loop through symbols manually:
1. `chart_set_symbol` + `chart_set_timeframe`
2. `chart_manage_indicator` — add the study
3. `data_get_ohlcv` — pull price data
4. `data_get_indicator` — read indicator values
5. Analyze and record findings

## Step 3: Compile Results

Build a comparison table:
| Symbol | Key Metric 1 | Key Metric 2 | Signal |
|--------|-------------|-------------|--------|

Sort by the most relevant metric.

## Step 4: Report

Present findings:
- Ranked list of symbols by the scan criteria
- Highlight the strongest setups
- Note any divergences or anomalies
- Screenshot the top 1-2 charts for visual confirmation

## Watchlist Integration

To scan the user's watchlist:
1. `watchlist_get` — read all symbols
2. Use the symbol list for the scan
3. `watchlist_add` — add new finds to the watchlist

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- Prefer `data_get_ohlcv` with `summary:true` for compact multi-symbol scans.
- Indicator values should be treated as reliable only when sourced from `tradingview_study_model`.
- If a symbol returns `bidAskAvailable:false`, omit bid/ask spread comparisons for that symbol.
## Release 1.2 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.

