---
name: futures-roll
description: Анализ continuous futures contracts, roll context и ограничений TradingView data.
---

# Futures Roll

## Workflow

1. `continuous_contract_context` — проверь `is_continuous`, `base_symbol`, `roll_number`, exchange.
2. `market_summary` — price action, volume vs average, active studies.
3. `indicator_state` или `data_get_indicator` — momentum/volatility studies.
4. `quote_get` — только если bid/ask доступны; проверь `bidAskAvailable`.

## Ограничения

- TradingView JS path не даёт точные expiry/roll calendars.
- Continuous futures prices могут быть back-adjusted.
- Для точного календаря экспирации нужен внешний exchange calendar, вне скоупа этого repo.
## Release 1.2 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.

