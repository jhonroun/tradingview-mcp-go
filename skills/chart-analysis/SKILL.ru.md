---
name: chart-analysis
description: Технический анализ текущего графика TradingView с проверкой символа, индикаторов, данных и скриншота.
---

# Анализ графика

## Workflow

1. `chart_get_state` — проверь symbol, timeframe, studies/entity IDs.
2. `quote_get` — получи текущий OHLCV snapshot; если `bidAskAvailable:false`, не считай bid/ask spread.
3. `data_get_ohlcv` с `summary:true` — получи компактный контекст баров.
4. Для важных индикаторов используй `data_get_indicator` по `entity_id`.
5. Для истории индикаторов используй `data_get_indicator_history`; coverage = `loaded_chart_bars`.
6. `capture_screenshot` — только визуальное подтверждение.

## Правила качества

- Для trading logic доверяй только значениям с `source: tradingview_study_model` и `reliableForTradingLogic:true`.
- Не выводи numeric indicator values из pixels/canvas.
- Явно отделяй visual observations от numeric data.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


