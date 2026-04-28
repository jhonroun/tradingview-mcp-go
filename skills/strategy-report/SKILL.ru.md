---
name: strategy-report
description: Формировать отчёт по Strategy Tester: metrics, trades, orders, equity и ограничения.
---

# Strategy Report

## Workflow

1. `data_get_strategy_results`; если `status` не `ok`, остановись.
2. `data_get_trades` для сделок.
3. `data_get_orders` для filled orders.
4. `data_get_equity` для equity.
5. `capture_screenshot` chart/strategy_tester опционально.

## Проверки

- Strategy data reliable only with `source: tradingview_backtesting_api` and `status: ok`.
- Equity reliable only with `source: tradingview_strategy_plot` and explicit `Strategy Equity` plot.
- `coverage: loaded_chart_bars` не равен full history.
- При `needs_equity_plot` предложи Pine line.
## Release 1.2 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.

