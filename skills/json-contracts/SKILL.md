---
name: json-contracts
description: Verify that a tool response conforms to the Phase 5 JSON contract. Use before passing data to a downstream system (HTS layer, LLM prompt, reporting pipeline). Checks field presence, types, and sentinel values for the six priority tools.
---

# JSON Contract Verification

You are checking that a MCP tool response matches the Phase 5 JSON contract before passing it downstream.

## When to Use

- You are about to embed tool output in an LLM prompt and need to confirm completeness
- A downstream pipeline (HTS layer, reporting, alert system) will consume the response
- You receive unexpected `null` or missing fields from a tool call
- You are debugging why an indicator value is not appearing

## Priority Tools and Their Contracts

### `data_get_study_values`

```json
{
  "success": true,
  "study_count": 2,
  "studies": [
    {
      "name": "RSI",
      "entity_id": "Study_RSI_0",
      "plot_count": 1,
      "plots": [{ "name": "RSI", "current": 55.3, "values": [55.3] }]
    }
  ]
}
```

Checks:
- `studies` is always `[]` (never `null`) — safe to iterate
- Each study has `entity_id` (string), `plot_count` (integer), `plots` (array)
- `plots[0].current === plots[0].values[0]` — current is always `values[0]`

### `chart_get_state`

```json
{
  "success": true,
  "symbol": "BINANCE:BTCUSDT",
  "exchange": "BINANCE",
  "ticker": "BTCUSDT",
  "timeframe": "60",
  "type": "1",
  "indicators": [{ "id": "Study_RSI_0", "name": "RSI" }],
  "pane_count": 2
}
```

Checks:
- `exchange` and `ticker` are always strings (empty if symbol has no `:` prefix)
- `indicators` === `studies` (alias) — both present for backward compat
- `pane_count` is integer ≥ 0

### `quote_get`

```json
{
  "success": true,
  "symbol": "BINANCE:BTCUSDT",
  "last": 67400.0,
  "bid": 0,
  "ask": 0,
  "change": 600.0,
  "change_pct": 0.9
}
```

Checks:
- `bid`, `ask`, `change`, `change_pct` are **always present** — use `0` as sentinel for unavailable
- All price fields are `float64`, never `null`

### `symbol_info`

Checks:
- `symbol`, `exchange`, `description`, `type` are always strings (empty if not returned by TradingView)

### `symbol_search`

Checks:
- Each result always has `symbol`, `exchange`, `description`, `type` (all strings)
- `count` ≤ 15

### `data_get_indicator`

```json
{
  "success": true,
  "entity_id": "Study_RSI_0",
  "name": "Relative Strength Index",
  "inputs": { "length": 14 },
  "plots": [{ "name": "RSI", "current": 55.3, "values": [55.3] }]
}
```

Checks:
- `inputs` is always an object `{}` — never an array, never null
- `name` is always a string (empty if metaInfo is unavailable)
- `plots` is always an array (empty if study has no visible outputs)

## Verification Workflow

### Step 1: Call the Tool

Call the tool normally. If `success: false`, handle the error (see error-handling skill).

### Step 2: Check Required Fields

```
if result["studies"] == null → studies not yet loaded; retry after 2s
if result["bid"] == null    → bug in tool; bid should always be 0 for unavailable
if result["exchange"] == null → bug in tool; exchange should always be ""
```

### Step 3: Check Array Safety

Before iterating:
- `data_get_study_values.studies` — always `[]`, safe
- `data_get_indicator.plots` — always `[]`, safe
- `chart_get_state.indicators` — always `[]`, safe

### Step 4: Access Plots Correctly

```
primary_value = study.plots[0].current       // current bar value
values_array  = study.plots[0].values        // [current] — only current bar available
```

Multi-output indicators (MACD, Bollinger Bands) have multiple entries in `plots`:
```
macd_histogram = plots[2].current  // index depends on indicator order
```

## Recovery Patterns

### Empty studies array
```
If study_count == 0:
  → Chart has no indicators loaded
  → Use chart_manage_indicator to add them
  → Then retry data_get_study_values
```

### Missing current in plot
```
If plot.current == null:
  → The indicator output is non-numeric (e.g. a label or shape)
  → Skip this plot for numeric analysis
```

### Wrong entity_id
```
Use chart_get_state to list all indicators with entity IDs
Then call data_get_indicator with the correct entity_id
```

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- Go-only extension tools include `data_get_indicator_history`, `data_get_orders`, and `pine_restore_source`.
- Data tools may include `source`, `reliability`, `coverage`, `status`, `warning`, and `reliableForTradingLogic`.
- Equity from `tradingview_strategy_plot` has `coverage: loaded_chart_bars` and is not automatically full Strategy Tester history.
- `quote_get` can include `bidAskAvailable:false`, `bidAvailable:false`, `askAvailable:false`, and `sourceLimitation`.
## Release 1.2 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.

