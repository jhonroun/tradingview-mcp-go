# PROMPTS.md — короткие промпты для Claude Code

## 1. Инвентаризация

Прочитай `CLAUDE.md`, `PORTING_GUIDE.md`, `PLAN.md`, `TODO.md`, `COMPATIBILITY_MATRIX.md`. Затем проанализируй исходный Node.js-репозиторий `tradingview-mcp` без изменения кода. Заполни карту MCP tools, CLI-команд, CDP-вызовов, scripts, skills и utilities. Обнови только `COMPATIBILITY_MATRIX.md`, `TODO.md`, `CHANGELOG.md`, `PORTING_NOTES.md`.

## 2. Go skeleton

Прочитай `CLAUDE.md`, актуальный `TODO.md` и `PLAN.md`. Создай минимальный компилируемый Go skeleton для MCP-сервера, CLI `tv`, CDP-клиента, discovery и launcher. Бизнес-логику не переносить. После изменений запусти `go test ./...` и обнови `CHANGELOG.md`.

## 3. Windows TradingView discovery

Прочитай `WINDOWS_TRADINGVIEW_DISCOVERY.md`. Реализуй поиск TradingView Desktop на Windows с поддержкой `%LOCALAPPDATA%`, `%PROGRAMFILES%`, Microsoft Store / WindowsApps через Appx metadata, ручного override через env/flag/config и диагностики. Реализуй `tv doctor windows`. Не меняй MCP tools, кроме health/doctor. Запусти тесты и обнови `TODO.md`, `CHANGELOG.md`, `COMPATIBILITY_MATRIX.md`.

## 4. CDP health-check

Реализуй минимальный CDP-клиент и `tv_health_check`: подключение к `127.0.0.1:9222`, чтение `/json/version`, выбор target/page, проверка доступности TradingView API. Сохрани JSON-совместимость с Node.js-версией. Не переносить остальные tools.

## 5. Перенос одной группы tools

Перенеси только группу `<GROUP_NAME>` из Node.js в Go. Сначала сравни входные/выходные JSON-контракты, затем реализуй Go-код и тесты. Не меняй соседние группы. Обнови `COMPATIBILITY_MATRIX.md`, `TODO.md`, `CHANGELOG.md`.

## 6. Ревью после локальной модели

Проверь патч, сгенерированный локальной моделью. Исправь несовместимости с Node.js-логикой, Go idioms, ошибки контракта JSON, проблемы Windows, гонки и плохую обработку ошибок. Не добавляй новые функции сверх оригинала.
