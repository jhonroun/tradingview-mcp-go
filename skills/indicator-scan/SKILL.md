---
name: indicator-scan
description: Read and interpret all active indicator signals on the current chart by name. Use when you need a clear bullish/bearish/neutral verdict for each indicator without parsing raw value arrays. Calls indicator_state for each study of interest.
---

# Indicator Signal Scanner

You are reading and classifying signals from active TradingView indicators on the current chart.
Use this to produce a structured signal table without dealing with raw numeric arrays.

## When to Use

- User asks "what are my indicators saying?"
- You need a quick multi-indicator confluence check
- You want to know if RSI is overbought, MACD is bullish, etc. — without calling `data_get_study_values` and parsing it yourself
- You are building a trading decision framework that needs structured signals

## Step 1: Discover What Is Loaded

Call `chart_get_state` to see which indicators are active:

```json
chart_get_state {}
```

Look at the `studies` array — each entry has `id` (entity ID) and `name`.
Note the names; you will use them in the next step.

If no indicators are loaded, add them first:
- `chart_manage_indicator` with `action: "add"` and `indicator_name: "Relative Strength Index"`

## Step 2: Query Each Indicator by Name

Call `indicator_state` for each study you care about:

```json
{ "name": "RSI" }
{ "name": "MACD" }
{ "name": "Bollinger" }
{ "name": "EMA" }
{ "name": "Volume" }
```

`name` is a partial, case-insensitive match against active study names.
You do not need the entity ID — just the indicator's common name.

Each response gives you:
- `matched_name` — the full display name TradingView uses
- `primary_value` — first numeric value on the current bar, rounded to 2 dp
- `primary_key` — what that value represents (e.g. "RSI", "Histogram", "Upper")
- `direction` — `above_zero` / `below_zero` / `at_zero`
- `signal` — `bullish` / `bearish` / `neutral` / `overbought` / `oversold`
- `near_zero` — `true` if `|value| < 0.5` (momentum near inflection)
- `values` — full data-window dict for the bar (all sub-lines)

## Step 3: Build the Signal Table

Compile results:

| Indicator | Value | Signal | Direction |
|-----------|-------|--------|-----------|
| RSI       | 67.4  | neutral | above_zero |
| MACD      | 0.23  | bullish | above_zero |
| Bollinger | 149.2 | bullish | above_zero |

## Step 4: Confluence Assessment

Count signals:
- **Bullish confluence**: majority `bullish` or `above_zero`, no `overbought`
- **Bearish confluence**: majority `bearish` or `below_zero`, no `oversold`
- **Caution — overbought**: RSI or Stochastic returning `overbought`; risk of reversal
- **Caution — oversold**: RSI or Stochastic returning `oversold`; potential bounce setup
- **Mixed / neutral**: signals conflict; no strong directional bias

## Step 5: Report

```
**Signal Scan — [symbol] [timeframe]**

| Indicator | Value | Signal |
|-----------|-------|--------|
| RSI       | 67.4  | neutral |
| MACD Histogram | 0.23 | bullish |

**Confluence:** [bullish / bearish / mixed]
**Note:** [any overbought/oversold cautions, near_zero indicators approaching crossover]
```

## Tips

- If `indicator_state` returns `success: false`, the indicator is not loaded on the chart.
  Either skip it or add it with `chart_manage_indicator`.
- `near_zero: true` on a momentum indicator (MACD histogram, ROC, MOM) is worth flagging —
  it may signal an imminent crossover.
- For indicators with multiple lines (Bollinger Bands has Upper/Mid/Lower; MACD has Line/Signal/Histogram),
  `primary_value` returns the first numeric value. Check `values` for all sub-lines if needed.
- Combine this scan with `market_summary` to get OHLCV context alongside the signal read.

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- For exact current values, prefer `data_get_indicator` by `entity_id` from `chart_get_state`.
- For loaded-bar history, use `data_get_indicator_history`; coverage is `loaded_chart_bars`.
- Trading-logic values should have `source: tradingview_study_model`, `reliability: reliable_pine_runtime_value_unstable_internal_path`, and `reliableForTradingLogic:true`.
- If a value comes from `tradingview_ui_data_window`, treat it as localized display fallback, not reliable trading data.
## Release 1.2 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.

