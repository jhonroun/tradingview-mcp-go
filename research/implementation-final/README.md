# Implementation Final Regression Report

Date: 2026-04-27  
OS: Windows  
TradingView target: TradingView Desktop via CDP  
Live chart: `RUS:NG1!`, `1D`

## Summary

Final regression suite passed.

- `go test ./...`: passed, exit code `0`.
- `go vet ./...`: passed, exit code `0`.
- MCP initialize/tools-list/tools-call error shape: passed.
- Current MCP registry count: `85` tools.
- Live study-model indicator values: passed.
- Live indicator history: passed.
- Live strategy results/trades/orders: passed with a disposable test strategy.
- Live equity extraction: passed from explicit `Strategy Equity` plot with `coverage: loaded_chart_bars`.

Primary machine-readable summary:

- `research/implementation-final/regression-summary.json`

## Regression Coverage

Unit tests:

- Locale parser: `internal/tools/data/data_test.go`
- Indicator plot mapper / study model JS contract: `internal/tools/data/data_test.go`
- Strategy report normalizer / backtesting API JS contract: `internal/tools/data/data_test.go`
- Equity response builder / Strategy Equity plot extraction contract: `internal/tools/data/data_test.go`
- Screenshot filename normalizer: `internal/tools/capture/capture_test.go`

MCP golden tests:

- `initialize`: `internal/mcp/server_test.go`
- `tools/list` count: `internal/mcp/server_test.go` (`TestToolsListExact85`)
- `tools/call` error shape: `internal/mcp/server_test.go`

Live smoke tests:

- Indicator current values: `live-mcp-indicator-smoke.output.jsonl`
- Indicator history: `live-mcp-indicator-smoke.output.jsonl`
- Strategy results: `live-mcp-strategy-smoke.output.jsonl`
- Trades: `live-mcp-strategy-smoke.output.jsonl`
- Orders: `live-mcp-strategy-smoke.output.jsonl`
- Equity plot extraction: `live-mcp-strategy-smoke.output.jsonl`

## Live Evidence

Indicator current values:

- Tool: `data_get_indicator`
- Entity: `Vvzmzg`
- Study: `Помошник RSI - True ADX`
- Result: `success: true`
- Source: `tradingview_study_model`
- Reliability: `reliable_pine_runtime_value_unstable_internal_path`
- `reliableForTradingLogic: true`
- Sample plots: `DI+`, `DI-`, `ADX`, `RSI`
- Loaded bars: `400`

Indicator history:

- Tool: `data_get_indicator_history`
- Result: `success: true`
- Source: `tradingview_study_model`
- Coverage: `loaded_chart_bars`
- Returned sample bars: `5`

Strategy smoke:

- Test strategy: `MCP Test SMA Strategy`
- Temporary entity: `rqX2ro`
- Strategy results status: `ok`
- Source: `tradingview_backtesting_api`
- Trades: `43` total, `5` sampled
- Orders: `86` total, `5` sampled
- Metrics: `47`

Equity smoke:

- Tool: `data_get_equity`
- Result: `success: true`
- Status: `ok`
- Source: `tradingview_strategy_plot`
- Plot: `Strategy Equity`
- Coverage: `loaded_chart_bars`
- Points: `400`
- Warning preserved: loaded chart bars are not guaranteed to be the full backtest history.

## State Restoration

The live strategy smoke temporarily changed the Pine editor and chart. It was restored.

- Original Pine hash: `e675eb546b9f96fb79c1b7cd179d8908e7a77c7fcb4ca798124166042d350765`
- Backup manifest: `research/pine-source-safety/session-20260427T095805.302320300Z/backup.json`
- Restore result: `success: true`, `verified: true`
- Test strategy removed: `rqX2ro`, status `ok`
- Final chart studies returned to:
  - `Vvzmzg` / `Помошник RSI - True ADX`
  - `cfyrsD` / `Volume`

## Ready

- `data_get_indicator` returns numeric study model values, not canvas/UI coordinates.
- `data_get_indicator_history` returns loaded-bar history from the TradingView study model.
- `data_get_study_values` uses the same study model path for real values.
- `data_get_strategy_results`, `data_get_trades`, and `data_get_orders` use `TradingViewApi.backtestingStrategyApi()` with `awaitPromise=true`.
- `data_get_equity` extracts explicit `Strategy Equity` Pine plot values when available.
- Pine source tools create backups, return hashes, and restore with SHA256 verification.
- `symbol_search` no longer returns silent empty results.
- MOEX futures bid/ask zero/unavailable values are explicitly marked unavailable.
- Screenshot filenames no longer become `.png.png`.
- Study limit detection returns structured `study_limit_reached` without auto-delete unless explicitly allowed.

## Reliable For Trading Logic

Conditionally reliable tools:

- `data_get_indicator`: `reliableForTradingLogic=true` when `source=tradingview_study_model`.
- `data_get_indicator_history`: `reliableForTradingLogic=true` when `source=tradingview_study_model`.
- `data_get_study_values`: reliable only for values returned from `tradingview_study_model`.
- `data_get_strategy_results`: reliable only when `status=ok` and `source=tradingview_backtesting_api`.
- `data_get_trades`: reliable only when `status=ok` and `source=tradingview_backtesting_api`.
- `data_get_orders`: reliable only when `status=ok` and `source=tradingview_backtesting_api`.
- `data_get_equity`: reliable only when `status=ok`, `source=tradingview_strategy_plot`, and an explicit `Strategy Equity` Pine plot exists.

All of the above still depend on undocumented TradingView internals, so the reliability string intentionally includes `unstable_internal_path`.

## UI / Control Only

These tools are automation/control or visual-context tools, not reliable trading data sources:

- Chart controls: `chart_set_symbol`, `chart_set_timeframe`, `chart_set_type`, `chart_set_visible_range`, `chart_scroll_to_date`, `chart_manage_indicator`.
- Indicator controls: `indicator_set_inputs`, `indicator_toggle_visibility`.
- UI automation: `ui_click`, `ui_open_panel`, `ui_keyboard`, `ui_mouse_click`, `ui_type_text`, `ui_hover`, `ui_scroll`, `ui_find_element`, `ui_evaluate`.
- Capture: `capture_screenshot`.
- Pine editing/management: `pine_set_source`, `pine_restore_source`, `pine_compile`, `pine_smart_compile`, `pine_save`, `pine_new`, `pine_open`, `pine_list_scripts`, `pine_analyze`, `pine_check`.
- Replay controls: `replay_start`, `replay_step`, `replay_stop`, `replay_status`, `replay_autoplay`, `replay_trade`.
- Drawings/panes/tabs/layouts/watchlist/alerts tools are UI/control unless a response explicitly marks a reliable data source.

## Unstable TradingView Internals

The following paths are verified but undocumented and may break after TradingView updates:

- `window.TradingViewApi._activeChartWidgetWV`
- `chart.model().model().dataSources()`
- `study.data().valueAt(index)`
- `study.data().fullRangeIterator()`
- `study.metaInfo().plots`
- `study.metaInfo().styles`
- `model.strategySources()`
- `model.activeStrategySource()`
- `window.TradingViewApi.backtestingStrategyApi()`
- Pine editor/Monaco DOM and React-fiber paths

## Remaining Limitations

- Equity plot extraction covers loaded chart bars only, not necessarily full Strategy Tester history.
- Full native bar-by-bar equity was not found in the backtesting report; explicit Pine `plot(strategy.equity, "Strategy Equity", display=display.data_window)` remains required for reliable equity bars.
- Derived/reconstructed equity must remain marked `reliableForTradingLogic=false` or conditional.
- TradingView locale/UI text can affect control actions. This was later closed in the release-hardening pass: `pine_compile` / `pine_smart_compile` now recognize English and Russian Add-to-chart labels, with smoke evidence in `research/pine-localized-compile/`.
- MOEX futures bid/ask may remain unavailable from TradingView; returned `0`/missing bid or ask is not treated as a valid quote.
- Replay trade workflows remain UI-state dependent and should be live-tested separately when replay mode is intentionally active.

## Release v1.2.0 Follow-up

- `tv_discover` now includes structured `compatibility_probes` for unstable TradingView internals.
- Equity remains explicitly loaded-bars-only with `coverage: loaded_chart_bars`.
- Optional chart history loading is documented as best-effort: expand/scroll chart range, wait for bars, repeat data calls, compare `loaded_bar_count` / `data_points`.
- Derived equity remains conditional and is not native Strategy Tester equity.
- Full native bar-by-bar Strategy Tester equity is not pursued until TradingView exposes a stable report field.

## Artifacts

- `go-test.output.txt`, `go-test.exitcode.txt`
- `go-vet.output.txt`, `go-vet.exitcode.txt`
- `live-mcp-indicator-smoke.output.jsonl`
- `live-mcp-strategy-smoke.output.jsonl`
- `live-chart-before.json`
- `live-chart-after-strategy-add.json`
- `live-chart-after-cleanup.json`
- `live-pine-before.json`
- `live-pine-set-equity-strategy.json`
- `live-pine-smart-compile-equity-strategy.json`
- `live-pine-compile-equity-strategy.json`
- `live-pine-restore-original.json`
- `live-pine-after-restore.json`
- `regression-summary.json`

