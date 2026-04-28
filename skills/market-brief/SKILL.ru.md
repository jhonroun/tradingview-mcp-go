---
name: market-brief
description: Краткий market brief: price action, volume, indicators и data-quality caveats.
---

# Market Brief

## Workflow

1. `market_summary` — compact market context.
2. `chart_get_state` — сверить studies/entity IDs при необходимости.
3. `data_get_indicator` — подтвердить критичные indicator values.
4. `quote_get` — проверить bid/ask availability.
5. `capture_screenshot` опционально.

## Output

Symbol/timeframe, price, volume classification, indicators table, bias, limitations.

## Ограничения

- Не выдавай UI/canvas values за numeric truth.
- Указывай limitations для unavailable bid/ask и partial loaded bars.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


