---
name: tradingview-limit-handling
description: Обрабатывать лимит TradingView на studies/indicators без автоматического удаления, если пользователь явно не разрешил.
---

# Обработка лимитов TradingView

Используй при добавлении indicators или strategies.

## Workflow

1. Вызови `chart_manage_indicator` с `action: add`.
2. Если response содержит `status: study_limit_reached`, остановись и покажи:
   - `currentStudies`
   - `limit`, если доступен
   - `suggestion`
3. Не удаляй studies автоматически.
4. Передавай `allow_remove_any:true` только если пользователь явно разрешил удалить любой study.
5. После explicit removal проверь research removal log.

## Правило

Удаление study меняет UI-состояние. Требуется явное намерение пользователя.
## Release v1.2.0 Data Guards

- После обновлений TradingView Desktop или unavailable statuses у internal-path tools запускай `tv discover` и проверяй `compatibility_probes`.
- Считай `coverage: loaded_chart_bars` только chart-loaded coverage, включая strategy equity из `data_get_equity`.
- Optional history-load workflow — только best effort: расширить/проскроллить chart range, дождаться догрузки баров, повторить data call, сравнить `loaded_bar_count` / `data_points`.
- Derived equity оставляй conditional; не выдавай её за native Strategy Tester equity или безусловный `reliableForTradingLogic:true` источник.
- Не искать full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.


