# TradingView Internal Compatibility Probes

Date: 2026-04-28  
Scope: `tradingview-mcp-go` release hardening for tag `v1.2.0`

## Purpose

TradingView internal paths are useful but undocumented. They cannot be made stable from this repository, so release `v1.2.0` treats them as a compatibility surface that must be probed and recorded.

`tv_discover` keeps its legacy `paths` object and now adds `compatibility_probes`.

Each probe records:

- `compatible`: the path/method exists in the current TradingView Desktop build.
- `available`: useful data exists in the current chart state.
- `status`: `ok`, `no_strategy_loaded`, `needs_equity_plot`, `strategy_report_unavailable`, `unavailable`, or `error`.
- `stability`: `unstable_internal_path`.
- `reliability`: reliability class used by dependent data tools.

## Probed Areas

- `window.TradingViewApi`
- `window.TradingViewApi._activeChartWidgetWV.value()`
- `chart.model().model()`
- `model.dataSources()`
- `study.data().valueAt()` / `study.data().fullRangeIterator()`
- `model.strategySources()`
- `model.activeStrategySource()`
- `await window.TradingViewApi.backtestingStrategyApi()`
- explicit `Strategy Equity` plot through `metaInfo().plots/styles` and `data().fullRangeIterator()`

## Equity Decision

`data_get_equity` remains loaded-bars-only:

- reliable path: explicit Pine line  
  `plot(strategy.equity, "Strategy Equity", display=display.data_window)`
- source: `tradingview_strategy_plot`
- coverage: `loaded_chart_bars`
- reliability: `reliable_pine_runtime_value_unstable_internal_path`

Derived equity remains conditional and must not be presented as native Strategy Tester equity. Full native bar-by-bar Strategy Tester equity is not pursued until TradingView exposes a stable report field.

## Optional History-Load Workflow

When more bars are required:

1. Expand visible range with `chart_set_visible_range` or scroll to older dates with `chart_scroll_to_date`.
2. Wait for TradingView Desktop to load additional chart bars.
3. Re-run `data_get_equity` or `data_get_indicator_history`.
4. Compare `loaded_bar_count`, `data_points`, `total_data_points`, and `coverage`.
5. Keep the result marked as loaded-bars coverage.

This workflow is best-effort chart loading, not a native full backtest equity export.

## Artifacts

- `tv-discover.json`: live CLI output from `tv discover`.
- `summary.json`: reduced summary for release notes.
- `go-test.output.txt`, `go-test.exitcode.txt`
- `go-vet.output.txt`, `go-vet.exitcode.txt`
- `mcp-smoke.jsonl`: MCP initialize/tools-list/unknown-tool smoke.
- `mcp-smoke.summary.json`: reduced MCP smoke summary.

## Live Result

Command:

```powershell
go run ./cmd/tv discover
```

Exit code: `0`

Probe summary:

- `tradingview_api`: `ok`
- `active_chart_widget`: `ok`
- `chart_model`: `ok`
- `data_sources`: `ok`
- `study_model_data`: `ok`
- `strategy_sources`: `no_strategy_loaded`
- `active_strategy_source`: `no_strategy_loaded`
- `backtesting_strategy_api`: `no_strategy_loaded`
- `strategy_equity_plot`: `no_strategy_loaded`

The no-strategy statuses are state-dependent, not a compatibility failure:
the paths are `compatible:true`, but `available:false` because the active chart
did not have a loaded strategy during this probe.

