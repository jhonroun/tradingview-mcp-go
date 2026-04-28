---
name: strategy-backtesting-api
description: Read Strategy Tester performance, trades, and filled orders through TradingView backtestingStrategyApi with explicit status handling.
---

# Strategy Backtesting API

Use this when a Pine strategy is loaded on the chart and the user needs Strategy Tester data.

## Workflow

1. Call `data_get_strategy_results`.
2. If `status` is not `ok`, stop and report the status.
3. Call `data_get_trades` for trades.
4. Call `data_get_orders` for filled orders.
5. Treat data as trading-reliable only when:
   - `source: tradingview_backtesting_api`
   - `status: ok`
   - `reliableForTradingLogic: true`

## Status Handling

- `no_strategy_loaded`: ask the user to add/load a Pine strategy.
- `tradingview_backtesting_api_unavailable`: TradingView internal API is unavailable.
- `strategy_report_unavailable`: strategy exists but report is not ready.
- `strategy_report_shape_unverified`: response shape changed; do not infer metrics.

## Output

Report status, strategy name/id, metric count, total trades, total orders, and any limitations.
## Release 1.2 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.

