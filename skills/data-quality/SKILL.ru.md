---
name: data-quality
description: Проверять source, reliability, coverage и unavailable-value semantics перед использованием TradingView MCP data.
---

# Качество данных

Используй перед тем, как опираться на MCP data в trading logic.

## Checklist

- `success` true или non-error status явно документирован.
- `source` подходит для use case.
- `reliability` присутствует для TradingView internal paths.
- `reliableForTradingLogic` true для trading-logic conclusions.
- `coverage` понятен, особенно `loaded_chart_bars`.
- `bidAskAvailable` true перед использованием bid/ask spread.
- Derived data помечен conditional и не считается native TradingView output.

## Надёжные источники

- `tradingview_study_model`
- `tradingview_backtesting_api` при `status: ok`
- `tradingview_strategy_plot` при explicit equity plot и с оговоркой loaded bars

## Ненадёжные или conditional

- `tradingview_ui_data_window`
- canvas/pixel/visual coordinates
- `derived_from_ohlcv_and_trades`
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


