---
name: error-handling
description: Корректно обрабатывать ошибки MCP tools: retryable, permanent, unavailable statuses и TradingView limitations.
---

# Обработка ошибок MCP

## Классификация

- Retryable: CDP disconnect, websocket, timeout, chart loading.
- Permanent: unknown tool, bad args, missing required fields.
- Structured unavailable: `no_strategy_loaded`, `needs_equity_plot`, `study_limit_reached`.

## Правила

- Не скрывай unavailable data за fake success.
- Если `study_limit_reached`, не удаляй studies без explicit `allow_remove_any:true`.
- Если `bidAskAvailable:false`, bid/ask spread недоступен.
- Если `strategy_report_shape_unverified`, не делай выводы по metrics.

## Recovery

1. Retry once for transient CDP/loading failures.
2. Refresh `chart_get_state` for stale entity IDs.
3. Ask the user to open TradingView/CDP if connection is unavailable.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


