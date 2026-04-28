---
name: strategy-backtesting-api
description: Читать performance, trades и filled orders из Strategy Tester через TradingView backtestingStrategyApi с явной обработкой статусов.
---

# Strategy Backtesting API

Используй, когда на графике загружена Pine strategy и нужны данные Strategy Tester.

## Workflow

1. Вызови `data_get_strategy_results`.
2. Если `status` не `ok`, остановись и сообщи статус.
3. Вызови `data_get_trades` для сделок.
4. Вызови `data_get_orders` для filled orders.
5. Считай данные надёжными для trading logic только если:
   - `source: tradingview_backtesting_api`
   - `status: ok`
   - `reliableForTradingLogic: true`

## Статусы

- `no_strategy_loaded`: попроси пользователя добавить/загрузить Pine strategy.
- `tradingview_backtesting_api_unavailable`: внутренний API TradingView недоступен.
- `strategy_report_unavailable`: strategy есть, но report не готов.
- `strategy_report_shape_unverified`: shape изменился; не делай выводы по metrics.

## Output

Укажи status, strategy name/id, metric count, total trades, total orders и ограничения.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


