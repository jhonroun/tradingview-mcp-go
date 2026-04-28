# Known Issues Closure

Date: 2026-04-27
OS: Windows
TradingView access: local TradingView Desktop through CDP
Chart during live checks: `RUS:NG1!`, timeframe `1D`

## Summary

This pass closes the fixed-scope issues listed in `PLAN.md` / `TODO_codex.md`
without adding non-TradingView systems.

## Issue Results

| Issue | Result | Evidence |
| --- | --- | --- |
| `symbol_search` returned silent `[]` | Fixed. Empty result now returns `status: no_results` and `reason`. | `before-symbol-search-empty.stdout.json`, `after-symbol-search-empty.stdout.json` |
| `data_get_indicator` partial | Verified. Current values and history come from `tradingview_study_model` with `reliableForTradingLogic: true`. | `after-mcp-indicator-strategy.output.jsonl` ids 2-3 |
| Pine/strategy/replay not fully classified | Status matrix added: Pine `live_tested`, strategy `partial`, replay `partial`, replay trade `unverified`. | `tool-verification-status.json` |
| indicator values could be canvas/display values | Closed for `data_get_indicator` / `data_get_study_values`: numeric values use `study.data().valueAt()` and `fullRangeIterator()`, not canvas/UI coordinates. | `after-mcp-indicator-strategy.output.jsonl`, unit test `TestBuildStudyModelJSUsesModelPaths` |
| MOEX futures `bid/ask=0` | Fixed. `quote_get` keeps numeric sentinel fields but adds `bidAskAvailable:false`, per-side availability, `sourceLimitation`, and warning. | `before-quote-current.stdout.json`, `after-quote-current.stdout.json` |
| screenshot filename `.png.png` | Fixed and verified. Existing old file shows the prior bug; new captures with `.png` filename do not duplicate extension. | `screenshot-filename-evidence.json`, `after-screenshot-filename.stdout.json` |

## Before / After Highlights

### `symbol_search`

Before:

```json
[]
```

After:

```json
{
  "success": true,
  "status": "no_results",
  "count": 0,
  "reason": "TradingView symbol search API returned no results after applying requested filters."
}
```

### MOEX `quote_get`

Before:

```json
{
  "symbol": "RUS:NG1!",
  "bid": 0,
  "ask": 0,
  "success": true
}
```

After:

```json
{
  "symbol": "RUS:NG1!",
  "bid": 0,
  "ask": 0,
  "bidAskAvailable": false,
  "sourceLimitation": "tradingview_moex_futures_bid_ask_unavailable",
  "warning": "TradingView did not expose usable bid/ask for this MOEX futures symbol; bid and ask are zero sentinels and must not be treated as executable quotes."
}
```

### `data_get_indicator`

Live MCP output for entity `Vvzmzg` returned RSI/ADX/DI as floats from:

```text
source: tradingview_study_model
reliability: reliable_pine_runtime_value_unstable_internal_path
reliableForTradingLogic: true
coverage: loaded_chart_bars
```

### Pine / Strategy / Replay Statuses

- Pine: `live_tested` for `pine_get_source`.
- Strategy: `partial`; current chart returned structured `no_strategy_loaded`
  for results/trades/orders/equity. Loaded-strategy mutation was not repeated
  in this pass.
- Replay: `partial`; `replay_status` returned live state, but replay trade
  workflow remains `unverified`.

## Verification

- `go test ./internal/tools/chart ./internal/tools/data ./internal/cli ./internal/tools/capture`
- `go test ./...`
- `go vet ./...`

Full command logs:

- `go-test.output.txt`
- `go-vet.output.txt`
