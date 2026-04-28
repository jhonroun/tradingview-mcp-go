---
name: regression-smoke
description: Run focused regression checks for tradingview-mcp-go and save evidence under research/.
---

# Regression Smoke

Use this before closing a stabilization pass.

## Static Checks

1. `go test ./...`
2. `go vet ./...`

## MCP Golden Checks

1. `initialize`
2. `tools/list` and count current tools (`85`)
3. `tools/call` unknown tool error shape

## Live Checks

When TradingView Desktop is available:

- `chart_get_state`
- `quote_get`
- `data_get_indicator`
- `data_get_indicator_history`
- strategy results/trades/orders if a test strategy is safely loaded
- equity plot extraction if `Strategy Equity` plot exists

Always save outputs in a dated `research/` folder.
## Release v1.2.0 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.


