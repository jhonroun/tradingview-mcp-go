# AGENTS.md — роли агентов для портирования tradingview-mcp на Go

## Цель

Организовать перенос Node.js-проекта `tradingview-mcp` на Go так, чтобы не потерять поведение, совместимость MCP tools и CLI.

## Роли

### 1. Architect Agent

Отвечает за архитектурные решения.

Задачи:

- поддерживать структуру Go-проекта;
- следить за разделением MCP/CDP/TradingView/CLI;
- запрещать смешивание JS expressions с Go-бизнес-логикой;
- обновлять `PLAN.md`;
- принимать решения о библиотеках.

Правило: Architect не пишет большие куски кода, если не зафиксирована целевая структура.

### 2. Compatibility Agent

Отвечает за 1:1 совместимость с Node.js.

Задачи:

- сверять имена tools;
- сверять JSON schemas;
- сверять CLI-команды;
- сверять ошибки;
- вести таблицу соответствия Node.js → Go;
- блокировать изменения публичного контракта.

### 3. CDP Agent

Отвечает за Chrome DevTools Protocol.

Задачи:

- реализовать подключение к `localhost:9222`;
- найти TradingView target;
- реализовать WebSocket JSON-RPC CDP client;
- реализовать `Runtime.evaluate`;
- реализовать screenshot;
- реализовать retry/backoff;
- покрыть тестами сериализацию CDP сообщений.

### 4. TradingView Tools Agent

Отвечает за перенос tool-логики.

Группы:

- health/status;
- chart;
- quote/ohlcv/values;
- Pine;
- drawing;
- alert;
- watchlist;
- indicator;
- layout/pane/tab;
- replay;
- stream;
- UI automation;
- screenshot/discover/range/scroll.

Правило: одна группа tools — один законченный этап.

### 5. CLI Agent

Отвечает за команду `tv`.

Задачи:

- реализовать CLI dispatcher;
- сохранить JSON-вывод;
- сохранить pipe-friendly поведение;
- реализовать групповые команды;
- добавить help;
- синхронизировать CLI с MCP registry.

### 6. Test Agent

Отвечает за тестирование.

Задачи:

- unit-тесты для MCP;
- unit-тесты для CDP serialization;
- golden tests для JSON-ответов;
- smoke tests для CLI;
- optional e2e tests с реальным TradingView Desktop.

### 7. Documentation Agent

Отвечает за документы.

Задачи:

- обновлять `README.md`;
- обновлять `PORTING_GUIDE.md`;
- обновлять `TODO.md`;
- обновлять `CHANGELOG.md`;
- фиксировать известные расхождения.

## Процесс работы

Каждый этап выполняется так:

1. Architect уточняет границы этапа.
2. Compatibility Agent фиксирует ожидаемое поведение.
3. Implementation Agent пишет код.
4. Test Agent добавляет тесты.
5. Documentation Agent обновляет документы.
6. Claude запускает проверку.
7. Только после зелёного состояния этап закрывается.

## Формат записи в TODO.md

```md
- [ ] P2.03 `quote_get`
  - Source: `src/tools/...`
  - Target: `internal/tools/quote/...`
  - MCP: `quote_get`
  - CLI: `tv quote`
  - Status: pending
  - Tests: required
```

## Формат записи в CHANGELOG.md

```md
## 2026-04-24

### Added
- Added Go MCP stdio skeleton.
- Added CDP `/json/list` target discovery.
- Added `tv_health_check` tool.

### Changed
- None.

### Pending
- Pine tools are not ported yet.
```

## Запреты для агентов

Агентам запрещено:

- пропускать тесты;
- портировать «по памяти» без сверки с исходником;
- объединять несколько крупных групп tools в один коммит;
- менять MCP имена;
- использовать YAML;
- добавлять торговые функции;
- отправлять данные наружу;
- писать «TODO потом» в коде без записи в `TODO.md`.
