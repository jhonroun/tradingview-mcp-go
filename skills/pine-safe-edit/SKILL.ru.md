---
name: pine-safe-edit
description: Безопасно читать, изменять, компилировать, проверять и восстанавливать Pine source через backup и SHA256 guards.
---

# Безопасное редактирование Pine

Используй перед любым изменением Pine source.

## Workflow

1. `pine_get_source`: сохрани source, `source_sha256`, script name/type.
2. `pine_set_source` с `expected_current_sha256`, если заменяешь существующий код.
3. Проверь, что response содержит backup path и backup hash.
4. `pine_compile` или `pine_smart_compile`: прочитай structured diagnostics.
5. Проверь chart state или strategy/data output.
6. При необходимости вызови `pine_restore_source` с backup path и проверь SHA256.

## Правила

- Нельзя silently overwrite пользовательский код.
- Нельзя пропускать backup verification.
- Нельзя заявлять compile success при `error_count > 0`.
- Для equity strategies добавляй explicit `Strategy Equity` plot, если он нужен.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


