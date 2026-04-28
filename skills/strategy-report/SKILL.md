---
name: strategy-report
description: Generate a comprehensive strategy performance report — metrics, trade analysis, equity curve, and recommendations. Use after backtesting a Pine Script strategy.
---

# Strategy Performance Report

You are generating a detailed performance report for a Pine Script strategy on TradingView.

## Step 1: Gather Data

Collect all available performance data:
1. `data_get_strategy_results` — overall metrics (net profit, win rate, profit factor, etc.)
2. `data_get_trades` — individual trade list (up to 20)
3. `data_get_equity` — equity curve data points
4. `chart_get_state` — current symbol, timeframe, and studies on chart
5. `symbol_info` — symbol metadata for context

## Step 2: Capture Visuals

1. `capture_screenshot` with region "chart" — the chart with strategy overlay
2. `capture_screenshot` with region "strategy_tester" — the Strategy Tester panel

## Step 3: Analyze

### Key Metrics
Report these if available:
- **Net Profit** and **% return**
- **Total Trades** and **Win Rate**
- **Profit Factor** (target > 1.5)
- **Max Drawdown** ($ and %)
- **Average Trade** ($ and %)
- **Sharpe Ratio** if available
- **Max Consecutive Losses**

### Trade Analysis
From the trade list:
- Largest winner and largest loser
- Average winner vs average loser (reward:risk)
- Long vs short performance breakdown
- Time in market

### Equity Curve Assessment
- Is it smooth and upward-sloping?
- Any extended drawdown periods?
- Does it show consistency or was profit front/back-loaded?

## Step 4: Generate Report

Format as a structured report:

```
## Strategy Report: [Strategy Name]
**Symbol:** [symbol] | **Timeframe:** [tf] | **Period:** [date range]

### Summary
[1-2 sentence overview of performance]

### Key Metrics
| Metric | Value |
|--------|-------|
| Net Profit | ... |
| Win Rate | ... |
| Profit Factor | ... |
| Max Drawdown | ... |

### Strengths
- [bullet points]

### Weaknesses
- [bullet points]

### Recommendations
- [specific actionable improvements]
```

## Step 5: Suggest Improvements

Based on the analysis:
- If win rate < 50% but profit factor > 1: suggest tighter entries
- If max drawdown > 20%: suggest position sizing or stop loss adjustments
- If profit factor < 1.2: suggest the strategy may need fundamental changes
- If few trades: suggest widening the lookback or loosening entry criteria

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- Use `data_get_orders` in addition to `data_get_trades` when order fill details matter.
- Strategy tools are reliable only when `status: ok` and `source: tradingview_backtesting_api`.
- If status is `no_strategy_loaded`, do not produce a fake empty performance report.
- `data_get_equity` is reliable for trading logic only when `source: tradingview_strategy_plot` and `status: ok`.
- Equity with `coverage: loaded_chart_bars` is partial chart coverage, not guaranteed full backtest history.
- If `needs_equity_plot`, suggest `plot(strategy.equity, "Strategy Equity", display=display.data_window)`.
## Release v1.2.0 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.


