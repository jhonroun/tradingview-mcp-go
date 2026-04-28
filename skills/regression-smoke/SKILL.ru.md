---
name: regression-smoke
description: Запускать focused regression checks для tradingview-mcp-go и сохранять evidence в research/.
---

# Regression Smoke

Используй перед закрытием stabilization pass.

## Static Checks

1. `go test ./...`
2. `go vet ./...`

## MCP Golden Checks

1. `initialize`
2. `tools/list` и current count (`85`)
3. `tools/call` unknown tool error shape

## Live Checks

Если TradingView Desktop доступен:

- `chart_get_state`
- `quote_get`
- `data_get_indicator`
- `data_get_indicator_history`
- strategy results/trades/orders, если test strategy безопасно загружена
- equity plot extraction, если есть `Strategy Equity` plot

Всегда сохраняй outputs в датированную папку `research/`.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


