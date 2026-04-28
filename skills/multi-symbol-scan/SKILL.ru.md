---
name: multi-symbol-scan
description: Скан нескольких символов через TradingView MCP с компактными outputs и data-quality checks.
---

# Multi-Symbol Scan

## Workflow

1. Определи список symbols.
2. Для каждого symbol: `chart_set_symbol`, затем wait/retry при loading.
3. `data_get_ohlcv` с `summary:true`.
4. `market_summary` или `data_get_indicator` для active studies.
5. Сохрани results table.

## Правила

- Не сравнивай bid/ask, если `bidAskAvailable:false`.
- Indicator comparisons требуют `tradingview_study_model`.
- Batch scan зависит от скорости загрузки TradingView chart.
## Release 1.2 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.

