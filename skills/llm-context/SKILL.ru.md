---
name: llm-context
description: Собирать LLM-ready snapshot текущего графика перед анализом.
---

# LLM Context

## Workflow

1. `chart_context_for_llm` с `top_n`.
2. Проверь symbol/timeframe/price.
3. Для важных indicator values подтверди `data_get_indicator`.
4. Если нужен history, вызови `data_get_indicator_history`.
5. Если нужен визуальный контекст, добавь `capture_screenshot`.

## Правила

- Aggregate helpers удобны, но не заменяют source/reliability checks.
- Не делай trading-logic выводы без reliable source.
- При `bidAskAvailable:false` не рассчитывай spread.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


