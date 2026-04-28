---
name: strategy-equity-plot
description: Extract strategy equity from an explicit Pine Strategy Equity plot and handle loaded-bar coverage correctly.
---

# Strategy Equity Plot

Use this when the user needs bar-by-bar equity from TradingView.

## Requirement

Reliable equity requires the Pine strategy to include:

```pine
plot(strategy.equity, "Strategy Equity", display=display.data_window)
```

## Workflow

1. Call `data_get_equity`.
2. If `status: ok`, verify:
   - `source: tradingview_strategy_plot`
   - `coverage: loaded_chart_bars`
   - `reliableForTradingLogic: true`
3. If `status: needs_equity_plot`, return the suggested Pine line.
4. If source is derived, mark the result conditional and not native Strategy Tester equity.

## Limitation

`loaded_chart_bars` is not full backtest history unless TradingView has loaded the full range.
## Release 1.2 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.

