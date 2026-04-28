---
name: study-model-values
description: Read reliable TradingView indicator values from the internal study model, including current values and loaded-bar history.
---

# Study Model Values

Use this workflow when numeric indicator values may affect analysis or trading logic.

## Workflow

1. Call `chart_get_state` and record study `id`/`name`.
2. Call `data_get_indicator` with `entity_id` for current values.
3. Call `data_get_indicator_history` with `entity_id` and `max_bars` when history is needed.
4. Trust numeric values only when:
   - `source: tradingview_study_model`
   - `reliability: reliable_pine_runtime_value_unstable_internal_path`
   - `reliableForTradingLogic: true`
5. Treat `coverage: loaded_chart_bars` as chart-loaded history only.

## Do Not

- Do not infer RSI/ADX/EMA/MACD values from pixels or canvas coordinates.
- Do not treat `tradingview_ui_data_window` localized strings as trading-reliable.
- Do not continue with stale `entity_id`; refresh with `chart_get_state`.
## Release v1.2.0 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.


