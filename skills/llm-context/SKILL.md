---
name: llm-context
description: Build a complete LLM-ready market context snapshot before reasoning about a chart. Use when you need a single structured object to anchor analysis — avoids multiple round-trips. Calls chart_context_for_llm as the primary entry point.
---

# LLM Context Builder

You are building a complete, LLM-ready market context snapshot from a live TradingView chart.
Use this before any reasoning step that requires current market state.

## When to Use

- User asks "what is the current market doing?"
- You need to anchor a trading hypothesis to current data before proceeding
- You are about to generate a chart analysis without a prior `chart_get_state` call
- You want a single compact string to embed in a subsequent prompt or chain

## Step 1: Single-Call Snapshot

Call `chart_context_for_llm` with however many indicators matter for this analysis:

```json
{ "top_n": 5 }
```

This returns in one request:
- `symbol`, `timeframe`, `chart_type`
- `price`: last, open, high, low, close, volume
- `indicators`: top-N active study objects with names and current values
- `context_text`: compact pipe-delimited string, e.g.
  `"Symbol: NASDAQ:AAPL | TF: D | Price: 175.2 | RSI(RSI): 62.3 | MACD(Histogram): 0.45"`

## Step 2: Validate Completeness

Check the response:
- `indicator_count > 0` — if zero, the chart has no indicators loaded; ask the user or add them via `chart_manage_indicator`
- `price.last` is present and non-zero — if missing, the chart may still be loading; retry once after 2 s
- `symbol` matches what the user expects — if wrong, use `chart_set_symbol` to switch

## Step 3: Deepen If Needed

If the snapshot is sufficient, proceed to analysis.

If you need more detail on a specific indicator:
- `indicator_state` with `name: "RSI"` — returns `signal`, `direction`, `primary_value`, `near_zero`
- `data_get_study_values` — full values for all active studies (use when `top_n` was too low)
- `data_get_ohlcv` with `summary: true` — multi-bar summary if one bar is not enough

If you need visuals:
- `capture_screenshot` with `region: "chart"` — attach to the reasoning chain

## Step 4: Embed context_text

Use the `context_text` field as a one-line prefix when calling another model or constructing a summary:

```
Current market state: Symbol: ES1! | TF: 15 | Price: 5312.50 | RSI(RSI): 44.1 | MACD(Histogram): -2.3
```

This single line gives the model enough anchor to produce grounded output without large context.

## Common Patterns

### Quick pre-analysis check
```json
chart_context_for_llm { "top_n": 3 }
```
Fast — returns in one CDP round-trip. Use when you only need price + a few signals.

### Full indicator inventory
```json
chart_context_for_llm { "top_n": 20 }
```
Pulls all visible study values (capped at actual indicator count). Use when you don't know which indicators are loaded.

### Refresh after symbol change
After `chart_set_symbol` or `chart_set_timeframe`, always call `chart_context_for_llm` again — the previous snapshot is stale.

## Output Format

Report the context to the user as:

```
**Chart:** [symbol] | [timeframe] | Type: [chart_type]
**Price:** [last] (O: [open] H: [high] L: [low] V: [volume])
**Indicators:** [name]: [primary_value] ([signal]) for each indicator in the response
```

Then proceed with analysis or await user instructions.
