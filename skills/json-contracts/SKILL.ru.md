---
name: json-contracts
description: Проверять JSON response contracts MCP tools, включая source/reliability/status/coverage.
---

# JSON Contracts

## Проверяй

- `success`
- `error` или structured `status`
- `source`
- `reliability`
- `reliableForTradingLogic`
- `coverage`
- `warning`

## Текущие важные контракты

- Current Go registry: 85 tools.
- Go-only tools: `data_get_indicator_history`, `data_get_orders`, `pine_restore_source`.
- Equity reliable only with `source: tradingview_strategy_plot` and `status: ok`.
- `bidAskAvailable:false` запрещает bid/ask spread interpretation.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


