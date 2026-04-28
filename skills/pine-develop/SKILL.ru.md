---
name: pine-develop
description: Полный цикл разработки Pine Script с backup, compile diagnostics и restore.
---

# Разработка Pine Script

## Safe Workflow

1. `pine_get_source` — сохрани source, hash, script name/type.
2. `pine_set_source` с `expected_current_sha256`.
3. Убедись, что создан backup.
4. `pine_compile` или `pine_smart_compile`.
5. Прочитай `errors`, `warnings`, `diagnostics`.
6. Проверь chart/strategy result.
7. При необходимости `pine_restore_source`.

## Правила

- Не перезаписывай код без backup.
- Не заявляй done при `error_count > 0`.
- RU/EN Add-to-chart labels поддерживаются compile helpers.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


