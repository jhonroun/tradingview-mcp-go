---
name: tradingview-limit-handling
description: Handle TradingView study/indicator limits without automatic deletion unless explicitly allowed.
---

# TradingView Limit Handling

Use this when adding indicators or strategies.

## Workflow

1. Call `chart_manage_indicator` with `action: add`.
2. If the response has `status: study_limit_reached`, stop and show:
   - `currentStudies`
   - `limit` when available
   - `suggestion`
3. Do not remove studies automatically.
4. Only pass `allow_remove_any:true` if the user explicitly approved removing any study.
5. After explicit removal, check the research removal log.

## Rule

Study deletion is destructive UI state mutation. Treat it as requiring explicit user intent.
## Release 1.2 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.

