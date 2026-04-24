# CHANGELOG.md

Формат основан на Keep a Changelog, но адаптирован под процесс портирования Node.js → Go.

## Unreleased

### Planning

- Подготовлен план 1:1 переноса `tradingview-mcp` с Node.js на Go.
- Зафиксирована целевая архитектура Go-проекта.
- Зафиксированы основные группы MCP tools и CLI-команд.
- Зафиксировано требование сохранить CDP-подключение к TradingView Desktop через `localhost:9222`.
- Зафиксировано требование сохранить JSON-совместимость MCP tools и CLI output.

### Pending

- Инвентаризация исходного репозитория.
- Go skeleton.
- MCP stdio server.
- CDP client.
- `tv_health_check`.
- CLI `tv status`.

## Правила ведения

Каждый этап портирования должен добавлять запись в этот файл.

Пример:

```md
## 2026-04-24

### Added
- Added Go MCP stdio skeleton.
- Added CDP target discovery.
- Added `tv_health_check`.

### Changed
- None.

### Fixed
- None.

### Compatibility
- MCP `tools/list` returns one tool: `tv_health_check`.
- CLI `tv status` returns JSON.

### Pending
- Chart read tools.
- Pine Script tools.
```

## Типы записей

### Added

Новый Go-код, новый tool, новая CLI-команда, новый тест.

### Changed

Изменение существующей Go-реализации.

### Fixed

Исправление ошибки в Go-порте.

### Compatibility

Сведения о совпадении или расхождении с Node.js-версией.

### Pending

Что осталось незавершённым.

### Breaking

Использовать только если совместимость с Node.js сознательно нарушена.

Breaking changes запрещены без отдельного обоснования.


### Added

- Добавлено требование поддержки Windows Microsoft Store / WindowsApps установки TradingView Desktop.
- Добавлено требование портировать skills, scripts и utilities, а не только MCP tools.
- Добавлен регламент оркестрации портирования в лимитах Claude Pro.
- Добавлены отдельные документы `WINDOWS_TRADINGVIEW_DISCOVERY.md`, `ORCHESTRATION_AND_LIMITS.md`, `PROMPTS.md`.
