---
name: indicator-scan
description: Сканировать активные индикаторы и формировать signal table с проверкой source/reliability.
---

# Скан индикаторов

## Workflow

1. `chart_get_state` — получи studies и `entity_id`.
2. Для каждого важного study вызови `data_get_indicator`.
3. Для истории вызови `data_get_indicator_history`.
4. Используй `indicator_state` только как компактный signal helper.

## Проверки

- Trading-logic values: `source: tradingview_study_model`, `reliableForTradingLogic:true`.
- `coverage: loaded_chart_bars` означает только загруженную историю.
- UI/Data Window fallback не является reliable trading data.

## Output

Таблица: indicator, plot, current, signal, source, reliability.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


