# Indicator study model extraction

Date: 2026-04-27
OS: Windows
TradingView Desktop: reachable through CDP on port 9222
Chart: `RUS:NG1!`, timeframe `1D`

## Summary

- `data_get_study_values` now reads current values from `study.data().valueAt(index)`.
- `data_get_indicator` reads the same study-model values by `entity_id` or name.
- Added `data_get_indicator_history` for loaded-bar history via `fullRangeIterator()`.
- Plot mapping uses `metaInfo().plots` plus `metaInfo().styles`.
- Numeric values are tagged:
  - `source: tradingview_study_model`
  - `reliability: reliable_pine_runtime_value_unstable_internal_path`
  - `reliableForTradingLogic: true`
- UI/Data Window parsing remains separate and is not used for these study-model values.

## Live Smoke

MCP smoke file: `mcp-study-model.result.json`

Observed current values for `Помошник RSI - True ADX` (`entity_id: Vvzmzg`):

- `DI+`: `15.545741865281482`
- `ADX`: `26.293855814935608`
- `RSI`: `31.141247635850746`
- `history_count`: `5` in the smoke sample
- `total_bars`: `400`
- `current_bar_index`: `299`
- MCP `tools/list`: `83` tools, including `data_get_indicator_history`

CLI smoke:

- `go run ./cmd/tv indicator-history RSI --count 3`
- Result: exit code `0`, `history_count: 3`, `source: tradingview_study_model`

## Artifacts

- `mcp-study-model.input.jsonl`
- `mcp-study-model.output.jsonl`
- `mcp-study-model.result.json`
- `mcp-study-model.stderr.txt`
- `payload-2.json` — `data_get_study_values`
- `payload-3.json` — `data_get_indicator`
- `payload-4.json` — `data_get_indicator_history`
- `tools-list.response.json`
- `cli-indicator-history.stdout.json`
- `cli-indicator-history.stderr.txt`
- `cli-indicator-history.result.json`

## Tests

- `go test ./internal/tools/data ./internal/mcp ./cmd/tv ./cmd/tvmcp`
- `go test ./...`
- `go vet ./...`
