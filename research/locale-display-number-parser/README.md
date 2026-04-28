# Locale display number parser

Date: 2026-04-27
OS: Windows
TradingView Desktop: reachable through CDP on port 9222

## Summary

- Added `data.ParseDisplayNumber` for localized TradingView UI/Data Window strings.
- Covered comma decimals, dot decimals, K/M/B/T suffixes, unicode minus, percent signs, and unavailable values.
- `data_get_study_values` and `data_get_indicator` now parse Data Window display strings in Go instead of stripping commas in JavaScript.
- DOM quote/depth display strings now use the same parser.
- UI-derived values include `source` and `reliability`; Data Window values are marked `reliableForTradingLogic: false`.
- Direct numeric chart model values such as OHLCV bars remain unchanged.

## Examples Covered By Tests

| Input | Parsed |
| --- | ---: |
| `31,51` | `31.51` |
| `14,63 K` | `14630` |
| `1.2K` | `1200` |
| `1,2 M` | `1200000` |
| `−3,45` | `-3.45` |
| `—` / `na` | unavailable |

## Live Smoke

MCP `data_get_study_values` was run against the active chart. It returned:

- `source: tradingview_ui_data_window`
- `reliability: display_value_localized_ui_string`
- `reliableForTradingLogic: false`
- localized display values such as `31,14` and `23,12 K` parsed to numeric `31.14` and `23120`.

## Artifacts

- `mcp-data-get-study-values.input.jsonl`
- `mcp-data-get-study-values.output.jsonl`
- `mcp-data-get-study-values.result.json`
- `mcp-data-get-study-values.response.json`
- `data-get-study-values.payload.json`

## Tests

- `go test ./internal/tools/data ./internal/tools/hts`
- `go test ./...`
- `go vet ./...`
