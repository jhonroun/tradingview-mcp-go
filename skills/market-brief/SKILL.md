---
name: market-brief
description: Generate a concise market briefing — price action, volume profile, and indicator snapshot — for any symbol and timeframe. Use for morning briefs, end-of-day reviews, or any "give me a quick read" request. Calls market_summary as the primary entry point.
---

# Market Brief

You are generating a concise, structured market briefing for a TradingView chart.
This covers price action, volume, and all active indicator signals in a single readable output.

## When to Use

- User says "give me a quick read on [symbol]"
- Morning or end-of-day review
- Pre-trade context check before placing an order
- Regular scheduled briefing across a watchlist

## Step 1: Optional — Navigate to Symbol

If the user specified a symbol different from what is currently on the chart:
1. `chart_set_symbol` — switch to the requested symbol
2. `chart_set_timeframe` — set the requested timeframe (if specified)

Wait for the chart to load (the tools handle this automatically).

## Step 2: One-Call Snapshot

Call `market_summary`:

```json
market_summary {}
```

This returns in one request:
- `symbol`, `timeframe`, `chart_type`
- `last_bar`: `{time, open, high, low, close, volume}` — most recent completed bar
- `change`: close − previous bar's close (absolute)
- `change_pct`: e.g. `"1.35%"` or `"-0.72%"`
- `volume_vs_avg`: ratio to 20-bar average (e.g. `1.8` = 80% above average)
- `indicators`: all active study objects with `name` and `values`

## Step 3: Interpret Volume

Use `volume_vs_avg` to classify volume:
- `> 2.0` — exceptionally high volume; significant move or event
- `1.5 – 2.0` — above-average volume; confirms the move
- `0.8 – 1.5` — normal volume; no unusual activity
- `< 0.8` — below-average volume; move may lack conviction

## Step 4: Read Indicator Values

From the `indicators` array, for each study call `indicator_state` to get a structured signal:

```json
{ "name": "[indicator name from the array]" }
```

Or read `values` directly from the `market_summary` response if you only need raw numbers.

Key indicators to look for and their interpretation:
- **RSI**: overbought ≥ 70 (bearish risk), oversold ≤ 30 (bullish opportunity), 40–60 = neutral trend
- **MACD Histogram**: positive and rising = bullish momentum, negative and falling = bearish
- **Bollinger Bands**: price near Upper = extended, price near Lower = oversold, at Mid = mean reversion zone
- **EMA / SMA**: price above = bullish bias, price below = bearish bias
- **Volume**: compare `volume_vs_avg` to confirm or question the directional move

## Step 5: Capture Visual (Optional)

For a richer brief:
1. `capture_screenshot` with `region: "chart"` — attach chart image
2. Present the screenshot alongside the written brief

## Step 6: Deliver the Brief

Format:

```
## Market Brief — [symbol] | [timeframe] | [date/time of last bar]

**Price Action**
Close: [close] | Change: [change] ([change_pct])
Range: H [high] / L [low] | Open: [open]

**Volume**
[volume] — [classification: exceptional / above-avg / normal / light]
([volume_vs_avg]× 20-bar average)

**Indicators**
| Indicator | Value | Signal |
|-----------|-------|--------|
| [name]    | [val] | [signal] |
...

**Summary**
[2-3 sentences: overall bias, key levels, volume confirms or questions the move]
```

## Multi-Symbol Briefing

To brief multiple symbols, iterate:
1. `chart_set_symbol` → `market_summary` → record → repeat
2. Or use `batch_run` with `action: "screenshot"` for a visual overview, then `market_summary` per symbol

## Notes

- `market_summary` fetches 21 bars internally; the 20-bar volume average excludes the current bar.
- `change` and `change_pct` compare the last completed bar to the one before it — not to a session open.
- If `last_bar` is missing from the response, the chart is still loading; retry after 2 s.
- For after-hours or pre-market symbols, `volume` may be zero or low — note this in the brief.

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- Use aggregate tools for speed, then verify critical indicator values with study-model tools.
- Trading-logic conclusions require reliable values from `tradingview_study_model` or a clearly marked backtesting/equity source.
- For unavailable bid/ask, report the limitation instead of calculating a spread from zero values.
## Release v1.2.0 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.


