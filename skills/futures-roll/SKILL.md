---
name: futures-roll
description: Analyze a continuous futures contract — detect symbol type, parse base/roll, read price and indicators, and flag roll timing considerations. Use when the user is trading or monitoring futures contracts like NG1!, ES1!, CL2!, NQ1!.
---

# Futures Contract Context Workflow

You are analyzing a continuous futures contract on TradingView.
Use this when the current chart shows a symbol ending in `!` (e.g. `NG1!`, `ES1!`, `CL2!`).

## When to Use

- User mentions a futures symbol with `!` suffix
- User asks about contract rolls, expiry, front month
- You need to confirm whether the current chart is a continuous or single-expiry contract
- You are comparing front-month vs back-month spread

## Step 1: Detect the Contract

Call `continuous_contract_context`:

```json
continuous_contract_context {}
```

This returns:
- `symbol` — full symbol with exchange prefix (e.g. `"NYMEX:NG1!"`)
- `is_continuous` — `true` if the symbol ends in `!`
- `base_symbol` — root commodity (e.g. `"NG"`, `"ES"`, `"CL"`)
- `roll_number` — 1 = front month, 2 = second month, etc.
- `description` — TradingView's human-readable name (e.g. `"Natural Gas Futures"`)
- `exchange` — exchange (e.g. `"NYMEX"`, `"CME"`)
- `type` — instrument type (usually `"futures"`)
- `currency_code` — settlement currency
- `note` — reminder that expiry/roll dates require external data

If `is_continuous: false`, the chart shows a single expiry contract (e.g. `NGZ2024`).
Inform the user and suggest switching to `NG1!` for continuous data.

## Step 2: Get Current Price and Context

Call `market_summary` to get the full current snapshot:

```json
market_summary {}
```

Key fields for futures:
- `last_bar.close` — current futures price
- `change` and `change_pct` — session move
- `volume_vs_avg` — participation vs norm (important for roll periods — volume drops as expiry nears)
- `indicators` — any loaded studies (ATR is useful for futures volatility)

## Step 3: Check Indicators (if relevant)

If the user has indicators loaded, call `indicator_state` for relevant ones:

```json
{ "name": "ATR" }
{ "name": "RSI" }
{ "name": "Volume" }
```

For futures, ATR is especially useful — it tells you the typical daily range in price units,
which directly translates to contract dollar value (ATR × contract multiplier).

## Step 4: Roll Timing Awareness

TradingView's `!` contracts auto-roll — the chart data stitches contracts together.
Key considerations:

**Volume as roll proxy**: Volume on the front-month contract drops sharply in the
week before expiry as traders roll to the next contract. If `volume_vs_avg < 0.5`
on what should be a liquid market, check whether roll is imminent.

**Basis / front-back spread**: TradingView does not expose the front-back spread
through the chart JS API (noted in `continuous_contract_context.note`).
To check the spread:
1. Open a second pane: `pane_set_symbol` to set `NG2!` in pane 1 while `NG1!` is in pane 0
2. Note both prices from `quote_get` for each pane
3. Calculate spread manually: `NG2! close − NG1! close`

**Typical roll windows** (approximate — verify against exchange calendar):
- Energy (NG, CL): ~3rd business day before contract expiry, usually mid-month
- Equity index (ES, NQ): quarterly (March, June, September, December), roll ~1 week before expiry
- Metals (GC, SI): roll varies; watch open interest shift

## Step 5: Build the Report

```
## Futures Context — [symbol] | [timeframe]

**Contract**
Type: Continuous (roll_number: [N]) | Base: [base_symbol]
Exchange: [exchange] | Currency: [currency_code]
Description: [description]

**Price Action**
Close: [close] | Change: [change] ([change_pct])
Volume: [volume_vs_avg]× avg — [high/normal/low: flag if < 0.6 — possible roll period]

**Indicators**
[signal table from indicator_state calls]

**Roll Status**
[Comment on volume vs avg — normal activity or signs of roll approaching]
[Note: exact expiry/roll dates require exchange calendar; TradingView JS API does not expose them]
```

## Step 6: Multi-Contract Comparison (Optional)

To compare front vs second month:
1. `pane_list` — see current panes
2. `pane_set_symbol` for pane 1 → set to `[base]2!` (e.g. `NG2!`)
3. `quote_get` for each pane using `chart_set_symbol` + `quote_get` in sequence
4. Report the spread and its direction (contango = NG2! > NG1!, backwardation = NG2! < NG1!)

## Notes

- All `!` symbols in TradingView are continuous contracts using back-adjusted prices.
  Individual contract prices will differ from quoted settlement prices on the exchange.
- `continuous_contract_context` reads the current active chart — always call it first to confirm
  which contract is actually displayed before pulling other data.
- For single-expiry contracts (e.g. `ESH2025`), `is_continuous: false`; use `symbol_info`
  instead for contract details.

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- `continuous_contract_context` is a local TradingView context helper, not an exchange-calendar API.
- For MOEX futures and other feeds, `quote_get` can return `bidAskAvailable:false`; do not use `bid=0` or `ask=0` as a tradable spread.
- Indicator values used in roll or momentum analysis should come from `tradingview_study_model` with `reliableForTradingLogic:true`.
## Release 1.2 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.

