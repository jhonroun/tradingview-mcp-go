---
name: strategy-equity-plot
description: Извлекать equity стратегии из explicit Pine Strategy Equity plot и корректно обрабатывать coverage loaded bars.
---

# Strategy Equity Plot

Используй, когда пользователю нужна bar-by-bar equity из TradingView.

## Требование

Надёжная equity требует строку в Pine strategy:

```pine
plot(strategy.equity, "Strategy Equity", display=display.data_window)
```

## Workflow

1. Вызови `data_get_equity`.
2. Если `status: ok`, проверь:
   - `source: tradingview_strategy_plot`
   - `coverage: loaded_chart_bars`
   - `reliableForTradingLogic: true`
3. Если `status: needs_equity_plot`, верни suggested Pine line.
4. Если source derived, пометь результат conditional и не называй его native Strategy Tester equity.

## Ограничение

`loaded_chart_bars` не равен полной backtest history, если TradingView не загрузил весь диапазон.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


