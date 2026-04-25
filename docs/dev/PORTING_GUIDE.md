# PORTING_GUIDE.md — руководство по переносу tradingview-mcp с Node.js на Go

## 1. Назначение

Этот документ описывает правила 1:1 портирования проекта `tradingview-mcp` с Node.js на Go.

Цель — получить полнофункциональный Go-код с той же логикой работы:

- MCP-сервер;
- CLI `tv`;
- подключение к TradingView Desktop через CDP;
- чтение состояния графика;
- управление графиком;
- Pine Script workflow;
- screenshots;
- streaming JSONL;
- drawing/alerts/replay/ui automation.

## 2. Исходная архитектура

Оригинальный проект содержит:

```text
src/
├── cli/
├── core/
├── tools/
├── connection.js
├── server.js
└── wait.js
```

Ключевые зависимости:

```text
@modelcontextprotocol/sdk
chrome-remote-interface
```

Из этого следует:

- `server.js` — точка входа MCP;
- `connection.js` — CDP-подключение и `Runtime.evaluate`;
- `tools/` — реализация MCP tools;
- `cli/` — CLI-обёртка над теми же возможностями;
- `core/` — общая логика.

## 3. Целевая архитектура Go

```text
cmd/
├── tvmcp/
│   └── main.go
└── tv/
    └── main.go

internal/
├── cdp/
├── mcp/
├── tradingview/
├── tools/
├── cli/
├── stream/
├── launch/
├── pine/
└── compat/
```

## 4. Таблица соответствия

| Node.js | Go |
|---|---|
| `src/server.js` | `cmd/tvmcp/main.go`, `internal/mcp` |
| `src/connection.js` | `internal/cdp`, `internal/tradingview/eval.go` |
| `src/cli/index.js` | `cmd/tv/main.go`, `internal/cli` |
| `src/tools/*` | `internal/tools/*` |
| `src/core/*` | `internal/tradingview`, `internal/compat`, `internal/pine` |
| `scripts/*` | `scripts/*`, `internal/launch` |
| `tests/*` | `tests/*`, Go unit/golden/e2e tests |

## 5. MCP protocol

Go-сервер должен работать через stdio.

Минимальные методы:

- `initialize`;
- `notifications/initialized`;
- `tools/list`;
- `tools/call`.

### Tool registry

Все tools регистрируются централизованно:

```go
type Tool struct {
    Name        string
    Description string
    InputSchema any
    Handler     ToolHandler
}
```

### Handler contract

```go
type ToolHandler func(ctx context.Context, args json.RawMessage) (any, error)
```

Ошибки должны конвертироваться в MCP error response.

## 6. CDP layer

### Подключение

Алгоритм:

1. GET `http://localhost:9222/json/list`.
2. Распарсить список targets.
3. Найти target:
   - сначала `type == "page"` и URL содержит `tradingview.com/chart`;
   - затем любой target с `tradingview`;
   - иначе ошибка.
4. Подключиться к `webSocketDebuggerUrl`.
5. Вызвать:
   - `Runtime.enable`;
   - `Page.enable`;
   - `DOM.enable`.

### Runtime.evaluate

Go API:

```go
func (c *Client) Evaluate(ctx context.Context, expression string, opts EvaluateOptions) (*EvaluateResult, error)
```

Обязательные опции:

- `ReturnByValue`;
- `AwaitPromise`;
- timeout через context.

### Безопасность строк

Node.js использует `JSON.stringify(String(str))`.

Go-эквивалент:

```go
func SafeJSString(s string) string {
    b, _ := json.Marshal(s)
    return string(b)
}
```

### Проверка чисел

Node.js использует finite validation.

Go-эквивалент:

```go
func RequireFinite(v float64, name string) (float64, error) {
    if math.IsNaN(v) || math.IsInf(v, 0) {
        return 0, fmt.Errorf("%s must be a finite number", name)
    }
    return v, nil
}
```

## 7. TradingView internal API wrappers

Не вставлять JS прямо в handlers.

Правильно:

```go
func (tv *Bridge) GetQuote(ctx context.Context) (*Quote, error) {
    return tv.evalQuote(ctx)
}
```

Неправильно:

```go
func quoteGetHandler(...) {
    client.Evaluate("window.TradingViewApi....")
}
```

## 8. Группы инструментов

### P0 — Foundation

- MCP skeleton;
- CLI skeleton;
- CDP client;
- health check;
- JSON helpers.

### P1 — Read-only chart data

- `chart_get_state`;
- `quote_get`;
- `data_get_ohlcv`;
- `data_get_study_values`;
- `capture_screenshot`.

### P2 — Chart control

- symbol;
- timeframe;
- chart type;
- visible range;
- scroll to date;
- indicators.

### P3 — Pine Script

- get source;
- set source;
- smart compile;
- get errors;
- console;
- save/new/open/list;
- static analyze/check.

### P4 — Drawings and annotations

- lines;
- labels;
- tables;
- boxes;
- shape draw/list/get/remove/clear.

### P5 — Layouts, panes, tabs

- layout list/switch;
- pane list/layout/focus/symbol;
- tab list/new/close/switch.

### P6 — Alerts and watchlist

- alert list/create/delete;
- watchlist get/add.

### P7 — Replay

- replay start/step/stop/status/autoplay/trade.

Important: replay trade is simulated TradingView replay workflow only. It must not become real brokerage trading.

### P8 — UI automation

- click;
- keyboard;
- hover;
- scroll;
- find;
- eval;
- type;
- panel;
- fullscreen;
- mouse.

### P9 — Streaming

- quote;
- bars;
- values;
- lines;
- labels;
- tables;
- all.

## 9. CLI compatibility

CLI должна быть thin-wrapper над тем же registry.

Пример:

```text
tv status
tv quote
tv symbol AAPL
tv ohlcv --summary
tv screenshot -r chart
tv pine compile
tv pane layout 2x2
tv stream quote
```

Вывод:

- stdout: JSON;
- stderr: ошибки;
- JSONL для stream;
- без лишнего текста в stdout.

## 10. Тестирование

### Unit

- JSON-RPC parser;
- MCP request/response;
- CDP message encoding;
- SafeJSString;
- RequireFinite;
- CLI arg parsing.

### Golden

Для каждого tool:

```text
tests/golden/<tool-name>.input.json
tests/golden/<tool-name>.output.json
```

### E2E

E2E запускается только при наличии TradingView Desktop debug port.

```text
TV_E2E=1 go test ./tests/e2e
```

Если `TV_E2E` не задан, тесты пропускаются.

## 11. Порядок переноса

1. Зафиксировать список tools из Node.js.
2. Создать Go skeleton.
3. Реализовать MCP stdio.
4. Реализовать CDP.
5. Реализовать `tv_health_check`.
6. Реализовать CLI `tv status`.
7. Переносить группы tools по одной.
8. После каждой группы обновлять `TODO.md` и `CHANGELOG.md`.
9. После переноса всех tools провести compatibility audit.
10. Обновить README и инструкции установки.

## 12. Definition of Done

Порт готов, когда:

- `go test ./...` проходит;
- `tvmcp` работает в Claude Code MCP config;
- `tv status` и `tv quote` работают;
- все tools перечислены в `tools/list`;
- имена tools совпадают с оригиналом;
- CLI-команды покрывают оригинальные use cases;
- JSON-схемы совместимы;
- известные расхождения перечислены в `CHANGELOG.md`.
