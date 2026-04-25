# PLAN.md — план портирования tradingview-mcp на Go

## Цель проекта

Создать Go-порт `tradingview-mcp`, совместимый с оригинальной Node.js-реализацией.

Нужно сохранить:

- MCP tools;
- CLI `tv`;
- Chrome DevTools Protocol workflow;
- локальный режим работы;
- JSON-совместимость;
- Pine Script workflow;
- chart analysis workflow.

## Исходные факты

Оригинальный проект:

- Node.js 18+;
- ES modules;
- main: `src/server.js`;
- CLI binary: `tv -> src/cli/index.js`;
- dependencies:
  - `@modelcontextprotocol/sdk`;
  - `chrome-remote-interface`;
- TradingView Desktop запускается с `--remote-debugging-port=9222`;
- CDP endpoint: `localhost:9222`;
- основной механизм: `Runtime.evaluate` JavaScript внутри TradingView Desktop.

## Этап P0 — Инвентаризация

### Цель

Получить точную карту исходного проекта.

### Действия

- [ ] Склонировать оригинальный репозиторий.
- [ ] Зафиксировать commit hash.
- [ ] Выполнить `npm install`.
- [ ] Запустить Node.js-тесты.
- [ ] Составить список MCP tools.
- [ ] Составить список CLI-команд.
- [ ] Составить карту файлов `src/*`.
- [ ] Зафиксировать known behavior.

### Результат

- Таблица tools.
- Таблица CLI-команд.
- Начальный compatibility matrix.

## Этап P1 — Go skeleton

### Цель

Создать минимально компилируемый Go-проект.

### Действия

- [ ] `go mod init`.
- [ ] Создать `cmd/tvmcp`.
- [ ] Создать `cmd/tv`.
- [ ] Создать `internal/mcp`.
- [ ] Создать `internal/cdp`.
- [ ] Создать `internal/tools`.
- [ ] Добавить Makefile или scripts без YAML.
- [ ] Добавить `go test ./...`.

### Результат

- Go-проект компилируется.
- Есть пустой MCP server.
- Есть пустой CLI.

## Этап P2 — MCP stdio server

### Цель

Реализовать MCP protocol без TradingView-зависимостей.

### Действия

- [ ] JSON-RPC 2.0 parser.
- [ ] `initialize`.
- [ ] `tools/list`.
- [ ] `tools/call`.
- [ ] Tool registry.
- [ ] Error mapping.
- [ ] Unit tests.

### Результат

- Claude Code может увидеть Go MCP server.
- `tools/list` возвращает минимум `tv_health_check`.

## Этап P3 — CDP client

### Цель

Заменить `chrome-remote-interface`.

### Действия

- [ ] GET `/json/list`.
- [ ] Target discovery.
- [ ] WebSocket connection.
- [ ] CDP request id generator.
- [ ] `Runtime.enable`.
- [ ] `Page.enable`.
- [ ] `DOM.enable`.
- [ ] `Runtime.evaluate`.
- [ ] Liveness check.
- [ ] Retry/backoff.
- [ ] Screenshot primitive.

### Результат

- Go-код выполняет JS внутри TradingView Desktop.
- `tv_health_check` проверяет debug port и chart target.

## Этап P4 — Health/status tools

### Цель

Получить первый end-to-end результат.

### Tools

- [ ] `tv_health_check`
- [ ] status/connection info
- [ ] target info

### CLI

- [ ] `tv status`
- [ ] `tv launch` placeholder или полноценный launch

### Результат

- `tv status` возвращает JSON.
- MCP `tv_health_check` работает.

## Этап P5 — Read-only chart tools

### Цель

Перенести безопасные read-only операции.

### Tools

- [ ] `chart_get_state`
- [ ] `quote_get`
- [ ] `data_get_ohlcv`
- [ ] `data_get_study_values`
- [ ] `data_get_pine_lines`
- [ ] `data_get_pine_labels`
- [ ] `data_get_pine_tables`
- [ ] `data_get_pine_boxes`
- [ ] `capture_screenshot`

### Результат

- Claude может читать график и делать screenshot.

## Этап P6 — Chart control

### Цель

Перенести управление графиком.

### Tools

- [ ] `chart_set_symbol`
- [ ] `chart_set_timeframe`
- [ ] `chart_set_type`
- [ ] `chart_manage_indicator`
- [ ] `chart_scroll_to_date`
- [ ] `chart_set_visible_range`
- [ ] `symbol_info`
- [ ] `symbol_search`
- [ ] `indicator_set_inputs`
- [ ] `indicator_toggle_visibility`

## Этап P7 — Pine Script workflow

### Цель

Перенести workflow разработки Pine Script.

### Tools

- [ ] `pine_get_source`
- [ ] `pine_set_source`
- [ ] `pine_smart_compile`
- [ ] `pine_get_errors`
- [ ] `pine_get_console`
- [ ] `pine_save`
- [ ] `pine_new`
- [ ] `pine_open`
- [ ] `pine_list_scripts`
- [ ] `pine_analyze`
- [ ] `pine_check`

### Результат

- Claude может вставить Pine Script, скомпилировать, прочитать ошибки и исправить.

## Этап P8 — Drawing, alerts, watchlist

### Drawing

- [ ] `draw_shape`
- [ ] draw list
- [ ] draw get
- [ ] draw remove
- [ ] draw clear

### Alerts

- [ ] alert list
- [ ] alert create
- [ ] alert delete

### Watchlist

- [ ] watchlist get
- [ ] watchlist add

## Этап P9 — Layouts, panes, tabs

### Tools

- [ ] layout list
- [ ] layout switch
- [ ] pane list
- [ ] pane layout
- [ ] pane focus
- [ ] pane symbol
- [ ] tab list
- [ ] tab new
- [ ] tab close
- [ ] tab switch

## Этап P10 — Replay

### Tools

- [ ] replay start
- [ ] replay step
- [ ] replay stop
- [ ] replay status
- [ ] replay autoplay
- [ ] replay trade

## Этап P11 — UI automation

### Tools

- [ ] ui click
- [ ] ui keyboard
- [ ] ui hover
- [ ] ui scroll
- [ ] ui find
- [ ] ui eval
- [ ] ui type
- [ ] ui panel
- [ ] ui fullscreen
- [ ] ui mouse

## Этап P12 — Streaming

### Commands

- [ ] `tv stream quote`
- [ ] `tv stream bars`
- [ ] `tv stream values`
- [ ] `tv stream lines`
- [ ] `tv stream labels`
- [ ] `tv stream tables`
- [ ] `tv stream all`

### Результат

- JSONL output.
- Ctrl+C handled correctly.
- Нет лишнего текста в stdout.

## Этап P13 — Compatibility audit

### Действия

- [ ] Сравнить tools/list Node.js vs Go.
- [ ] Сравнить CLI help Node.js vs Go.
- [ ] Сравнить JSON output для ключевых команд.
- [ ] Проверить README install flow.
- [ ] Проверить Claude Code MCP config.
- [ ] Проверить Windows/macOS/Linux launch scripts.
- [ ] Обновить `CHANGELOG.md`.

## Этап P14 — Release candidate

### Действия

- [ ] Tag `v0.1.0-go-port`.
- [ ] Собрать бинарники.
- [ ] Протестировать на Windows.
- [ ] Протестировать на Linux.
- [ ] Протестировать на macOS, если доступно.
- [ ] Написать migration notes Node.js → Go.

## Готовность

Проект считается перенесённым только после закрытия P13.


## Дополнение: WindowsApps и packaged TradingView

В план добавлен обязательный этап `P2W`: корректный поиск TradingView Desktop на Windows, включая Microsoft Store / WindowsApps. Без этого Go-порт будет повторять одну из главных практических проблем оригинала: MCP не всегда находит TradingView на Windows. Подробности: `WINDOWS_TRADINGVIEW_DISCOVERY.md`.

## Дополнение: skills/utilities/scripts

Портировать нужно не только MCP tools, но и все вспомогательные scripts, skills, утилиты запуска, проверки, диагностики и CLI-команды. Каждая утилита должна быть отражена в `COMPATIBILITY_MATRIX.md`.

## Дополнение: экономия лимитов Claude Pro

Работа ведётся малыми сессиями по одной группе задач. Детальный регламент: `ORCHESTRATION_AND_LIMITS.md`.
