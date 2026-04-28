---
name: replay-practice
description: Практика ручной торговли в TradingView Replay через UI/control tools.
---

# Replay Practice

## Workflow

1. `replay_status` — проверь текущее состояние.
2. `replay_start` при необходимости.
3. `replay_step` или `replay_autoplay`.
4. `replay_trade` для buy/sell/close в режиме replay.
5. `capture_screenshot` для визуального подтверждения.
6. `replay_stop` после практики.

## Правила

- Replay tools — UI/control, не broker integration.
- Если состояние не проверено live, пометь как partial/unverified.
- Не заявляй real execution.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


