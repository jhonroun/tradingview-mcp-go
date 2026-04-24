# CLAUDE.md — портирование tradingview-mcp с Node.js на Go

## Цель

Перенести репозиторий `tradesdontlie/tradingview-mcp` с Node.js на Go с сохранением поведения 1:1.

Новая Go-реализация должна быть функционально совместима с оригиналом:

- тот же смысл MCP tools;
- те же имена инструментов MCP;
- совместимые JSON-входы и JSON-выходы;
- совместимое поведение CLI `tv`;
- тот же принцип работы через Chrome DevTools Protocol на `localhost:9222`;
- без обхода платных функций TradingView;
- без сетевого сбора, хранения или ретрансляции рыночных данных вне локальной машины.

## Главный принцип

Не перепридумывать архитектуру. Сначала зафиксировать поведение Node.js-версии, затем повторить его на Go.

Любое изменение считается ошибкой, если оно:

- меняет имя MCP tool;
- меняет схему аргументов;
- меняет структуру JSON-ответа;
- меняет текст ошибок без необходимости;
- добавляет внешнюю сетевую передачу данных;
- ломает CLI-совместимость;
- заменяет TradingView Desktop/CDP другим источником данных.

## Исходный проект

Оригинальный репозиторий:

```text
https://github.com/tradesdontlie/tradingview-mcp
```

Из README следует, что проект:

- подключает Claude Code к локальному TradingView Desktop;
- использует Chrome DevTools Protocol;
- требует запуска TradingView с `--remote-debugging-port=9222`;
- предоставляет MCP tools и CLI-команды;
- содержит около 78 MCP-инструментов;
- работает локально и не исполняет реальные сделки.

## Целевая Go-архитектура

Рекомендуемая структура:

```text
.
├── cmd/
│   ├── tvmcp/
│   │   └── main.go              # MCP server over stdio
│   └── tv/
│       └── main.go              # CLI
├── internal/
│   ├── cdp/                     # CDP client, Runtime/Page/DOM
│   ├── tradingview/             # JS expressions and TradingView API wrappers
│   ├── mcp/                     # MCP protocol server, tool registry
│   ├── tools/                   # tool implementations grouped by domain
│   ├── cli/                     # CLI command dispatch
│   ├── launch/                  # launch TradingView debug mode
│   ├── stream/                  # JSONL polling streams
│   ├── pine/                    # Pine helpers, analysis, compile workflow
│   └── compat/                  # JSON compatibility helpers
├── pkg/
│   └── tvbridge/                # optional public Go API
├── testdata/
├── tests/
├── scripts/
├── go.mod
├── README.md
├── PORTING_GUIDE.md
├── PLAN.md
├── TODO.md
└── CHANGELOG.md
```

## Работа Claude Code

Claude Code должен работать итерационно:

1. читать `PLAN.md`;
2. брать следующий пункт из `TODO.md`;
3. сверять поведение с Node.js-версией;
4. писать небольшой законченный фрагмент Go-кода;
5. запускать `go test ./...`;
6. обновлять `TODO.md`;
7. обновлять `CHANGELOG.md`;
8. не переходить к следующему модулю, пока текущий не компилируется.

## Обязательные документы

- `PLAN.md` — стратегический план переноса.
- `TODO.md` — рабочий чек-лист по этапам.
- `CHANGELOG.md` — журнал фактически выполненных изменений.
- `PORTING_GUIDE.md` — правила соответствия Node.js → Go.
- `AGENTS.md` — роли агентов и порядок взаимодействия.

## Правила портирования

### MCP

Сначала реализовать минимальный MCP stdio-сервер на Go.

Требования:

- JSON-RPC 2.0 поверх stdin/stdout;
- поддержка `initialize`;
- поддержка `tools/list`;
- поддержка `tools/call`;
- регистрация всех инструментов через единый registry;
- строгая сериализация JSON;
- ошибки возвращаются в MCP-совместимом формате.

Не начинать портирование TradingView tools до появления работающего MCP skeleton.

### CDP

Node.js использует `chrome-remote-interface`.

Go-эквивалент должен:

- обращаться к `http://localhost:9222/json/list`;
- искать target с `tradingview.com/chart`;
- подключаться к `webSocketDebuggerUrl`;
- уметь вызывать `Runtime.evaluate`;
- включать домены Runtime/Page/DOM;
- выполнять liveness-check;
- переподключаться с retry/backoff.

Допустимо использовать Go-библиотеку для WebSocket, но слой CDP должен быть собственным и понятным.

### JavaScript expressions

Оригинальный проект активно выполняет JavaScript внутри TradingView Desktop.

Правила:

- все JS expressions сохранять как строковые константы;
- не смешивать Go-логику и JS-код в одном месте;
- для пользовательских строк использовать JSON escaping;
- для чисел проверять `finite`;
- каждую JS-операцию оборачивать в typed Go-функцию.

### CLI

CLI `tv` должен остаться pipe-friendly.

Требования:

- вывод JSON по умолчанию;
- exit code `0` при успехе;
- exit code `1` при ошибке;
- ошибки в stderr;
- команды группируются как в оригинале: `status`, `quote`, `symbol`, `ohlcv`, `pine`, `draw`, `alert`, `stream`, `ui`, `pane`, `tab`, `replay`.

### Streaming

Streaming-команды должны писать JSONL.

Требования:

- одна строка = один JSON object;
- graceful shutdown по Ctrl+C;
- interval задаётся параметром;
- ошибки не должны ломать формат потока без необходимости.

## Запреты

Нельзя:

- добавлять реальные торговые операции;
- обходить ограничения TradingView;
- извлекать данные напрямую с серверов TradingView;
- менять публичный контракт tools без записи в `CHANGELOG.md`;
- удалять Node.js-логику до завершения портирования соответствующего Go-модуля;
- писать «почти аналог» вместо 1:1 совместимости.

## Первый рабочий этап

Начать с минимального Go-каркаса:

```text
go mod init github.com/<owner>/tradingview-mcp-go
cmd/tvmcp/main.go
cmd/tv/main.go
internal/mcp
internal/cdp
internal/tools/health
```

Минимальная цель этапа:

```text
tvmcp запускается как MCP server
tv status возвращает JSON
tv_health_check доступен как MCP tool
CDP подключение к localhost:9222 проверяется
```

## Критерий готовности полного порта

Порт считается завершённым только если:

- все исходные MCP tools перенесены;
- CLI покрывает те же команды;
- Go-сервер работает в Claude Code MCP config;
- health-check проходит;
- Pine Script workflow работает;
- screenshot работает;
- chart read/control работает;
- replay/draw/alert/pane/tab/ui группы либо перенесены, либо явно помечены как pending в `TODO.md`;
- `go test ./...` проходит;
- `CHANGELOG.md` отражает все изменения.
