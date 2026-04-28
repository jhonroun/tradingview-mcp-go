---
name: data-quality
description: Verify source, reliability, coverage, and unavailable-value semantics before using TradingView MCP data.
---

# Data Quality

Use this before relying on MCP data for trading logic.

## Checklist

- `success` is true or a non-error status is explicitly documented.
- `source` is appropriate for the use case.
- `reliability` is present for internal TradingView paths.
- `reliableForTradingLogic` is true for trading-logic conclusions.
- `coverage` is understood, especially `loaded_chart_bars`.
- `bidAskAvailable` is true before using bid/ask spread.
- Derived data is marked conditional and not treated as native TradingView output.

## Reliable Sources

- `tradingview_study_model`
- `tradingview_backtesting_api` with `status: ok`
- `tradingview_strategy_plot` with explicit equity plot and loaded-bars caveat

## Unreliable Or Conditional

- `tradingview_ui_data_window`
- canvas/pixel/visual coordinates
- `derived_from_ohlcv_and_trades`
## Release v1.2.0 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.


