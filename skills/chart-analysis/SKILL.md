---
name: chart-analysis
description: Analyze a chart — set up symbol/timeframe, add indicators, scroll to key dates, annotate, and screenshot. Use when the user wants technical analysis or chart review.
---

# Chart Analysis Workflow

You are performing technical analysis on a TradingView chart.

## Step 1: Set Up the Chart

1. `chart_set_symbol` — switch to the requested symbol
2. `chart_set_timeframe` — set the appropriate timeframe
3. Wait for the chart to load (the tool handles this automatically)

## Step 2: Add Indicators

Use `chart_manage_indicator` to add studies. Use full indicator names:
- "Relative Strength Index" (not RSI)
- "Moving Average Exponential" (not EMA)
- "Moving Average" (for SMA)
- "MACD"
- "Bollinger Bands"
- "Volume"
- "VWAP"
- "Average True Range"

After adding, use `indicator_set_inputs` to customize settings (e.g., change EMA length to 200).

## Step 3: Navigate to Key Areas

- `chart_scroll_to_date` — jump to a specific date of interest
- `chart_set_visible_range` — zoom to a specific date window
- `chart_get_visible_range` — check what's currently visible

## Step 4: Annotate

Use drawing tools to mark up the chart:
- `draw_shape` with `horizontal_line` for support/resistance
- `draw_shape` with `trend_line` for trend channels (needs two points)
- `draw_shape` with `text` for annotations

## Step 5: Capture and Analyze

1. `capture_screenshot` — screenshot the annotated chart
2. `data_get_ohlcv` — pull recent price data for quantitative analysis
3. `quote_get` — get the current real-time price
4. `symbol_info` — get symbol metadata (exchange, type, session)

## Step 6: Report

Provide the analysis:
- Current price and recent range
- Key support/resistance levels identified
- Indicator readings (RSI overbought/oversold, MACD crossover, etc.)
- Overall bias (bullish/bearish/neutral) with reasoning

## Cleanup

If you added indicators the user didn't ask for, remove them:
- `chart_manage_indicator` with action "remove" and the entity_id
- `draw_clear` to remove all drawings if they were temporary

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- Prefer `data_get_indicator` / `data_get_indicator_history` when a trading-logic conclusion depends on numeric indicator values.
- Treat values as trading-reliable only when `source: tradingview_study_model` and `reliableForTradingLogic: true`.
- UI screenshots and canvas observations are visual context only; never infer indicator values from pixels.
- `quote_get` may return `bidAskAvailable:false`; do not treat `bid`/`ask` of `0` as real bid/ask.
## Release 1.2 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.

