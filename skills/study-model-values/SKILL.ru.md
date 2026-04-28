---
name: study-model-values
description: Читать надёжные значения индикаторов TradingView из внутренней study model, включая current values и историю loaded bars.
---

# Значения из Study Model

Используй этот workflow, когда numeric indicator values влияют на анализ или торговую логику.

## Workflow

1. Вызови `chart_get_state` и сохрани `id`/`name` study.
2. Вызови `data_get_indicator` с `entity_id` для текущих значений.
3. Вызови `data_get_indicator_history` с `entity_id` и `max_bars`, если нужна история.
4. Доверяй numeric values только если:
   - `source: tradingview_study_model`
   - `reliability: reliable_pine_runtime_value_unstable_internal_path`
   - `reliableForTradingLogic: true`
5. `coverage: loaded_chart_bars` означает только историю, загруженную на графике.

## Нельзя

- Нельзя выводить RSI/ADX/EMA/MACD из пикселей или canvas coordinates.
- Нельзя считать `tradingview_ui_data_window` localized strings надёжными для trading logic.
- Нельзя работать со старым `entity_id`; сначала обнови `chart_get_state`.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


