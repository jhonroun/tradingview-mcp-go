# PORTING_NOTES.md — заметки по переносу Node.js → Go

Этот файл содержит конкретные наблюдения о Node.js-реализации, необходимые для точного Go-порта.
Источник: анализ `tradingview-mcp/` репозитория (апрель 2026).

---

## Инвентаризация

- **Репозиторий:** `tradingview-mcp` (клон в `tradingview-mcp/`)
- **Commit hash:** зафиксировать командой `git -C tradingview-mcp rev-parse HEAD` (TODO P0.02)
- **Node.js версия:** 18+ (ES modules, `"type": "module"` в package.json)
- **MCP SDK:** `@modelcontextprotocol/sdk ^1.12.1`
- **CDP library:** `chrome-remote-interface ^0.33.2`
- **MCP server name:** `tradingview`, version `2.0.0`

---

## Точный счёт tools

Итого: **78 MCP tools** (в CLAUDE.md было написано ~78, это подтверждено).

| Группа | Кол-во |
| --- | --- |
| Health & Connection | 4 |
| Chart State & Control | 8 |
| Symbol | 2 |
| Data Reading | 12 |
| Screenshot | 1 |
| Pine Script | 12 |
| Drawing | 5 |
| Alerts | 3 |
| Batch | 1 |
| Replay | 6 |
| Indicator Control | 2 |
| Watchlist | 2 |
| UI Automation | 12 |
| Pane Management | 4 |
| Tab Management | 4 |
| **Итого** | **78** |

---

## Архитектура Node.js-проекта

```
src/
├── server.js          # MCP entry, регистрирует tools через registerXxxTools(server)
├── connection.js      # CDP клиент, safeString, requireFinite, API-path helpers
├── wait.js            # waitForChartReady() — polling до стабилизации графика
├── tools/             # 15 файлов — только MCP-обёртки (вызывают core/*)
│   └── _format.js     # jsonResult(data, isError)
├── core/              # 15 файлов — бизнес-логика, вызывают evaluate()
└── cli/
    ├── index.js       # CLI entry
    ├── router.js      # dispatch команд
    └── commands/      # 15 файлов команд
```

**Паттерн регистрации tools:**

```js
// src/tools/health.js
export function registerHealthTools(server) {
    server.tool("tv_health_check", "...", {}, handler);
}

// src/server.js
import { registerHealthTools } from "./tools/health.js";
registerHealthTools(server);
```

**Go-эквивалент:**

```go
// internal/tools/health/health.go
func Register(r *mcp.Registry) {
    r.Register(mcp.Tool{Name: "tv_health_check", ...})
}
```

---

## connection.js — ключевые детали

### Алгоритм подключения

1. GET `http://localhost:9222/json/list`
2. Найти target: `type == "page"` && URL содержит `tradingview.com/chart`
3. Fallback: любой target с `tradingview` в URL
4. Подключиться к `webSocketDebuggerUrl`
5. Вызвать `Runtime.enable`, `Page.enable`, `DOM.enable`
6. Retry: 5 попыток, exponential backoff

### Helpers из connection.js

```js
safeString(str)       // JSON.stringify(String(str))
requireFinite(v, name) // throws если NaN или Inf
getChartApi()          // возвращает строку-путь, проверяет доступность
getChartCollection()
getBottomBar()
getReplayApi()
getMainSeriesBars()
```

**Go-эквиваленты** нужны в `internal/cdp/helpers.go`:

```go
func SafeJSString(s string) string        // json.Marshal(s)
func RequireFinite(v float64, name string) (float64, error)
```

**Go-эквиваленты path-helpers** нужны в `internal/tradingview/paths.go`:

```go
func ChartAPIExpr() string   // "window.TradingViewApi._activeChartWidgetWV.value()"
func ChartCollectionExpr() string
func ReplayAPIExpr() string
func AlertServiceExpr() string
func MainSeriesBarsExpr() string
```

---

## wait.js — polling chart readiness

Функция `waitForChartReady(expectedSymbol, expectedTf, timeout)`:

- Timeout: 10 секунд
- Poll interval: 200 мс
- Условия готовности:
  1. Нет loading spinners (`document.querySelector('.tv-spinner')`)
  2. Symbol совпадает с ожидаемым (если передан)
  3. Количество bars стабильно 2 подряд (не меняется между итерациями)

**Go-эквивалент** в `internal/tradingview/wait.go`:

```go
func WaitForChartReady(ctx context.Context, cdp *cdp.Client, symbol, tf string) error
```

---

## CDP Input calls — точные параметры

### Input.dispatchKeyEvent

```js
// Пример: Ctrl+Enter (compile)
client.Input.dispatchKeyEvent({
    type: "keyDown",
    key: "Enter",
    code: "Enter",
    modifiers: 2,  // ctrl=2, shift=1, alt=4, meta=8
    windowsVirtualKeyCode: 13,
})
```

Используется в:
- `src/core/pine.js`: Ctrl+Enter (compile), Ctrl+S (save)
- `src/core/tab.js`: Ctrl+T (new tab), Ctrl+W (close tab)
- `src/core/alerts.js`: Ctrl+A (select all)
- `src/core/watchlist.js`: Enter (confirm), Escape (cancel)
- `src/core/ui.js`: общий keyboard dispatch

### Input.insertText

```js
client.Input.insertText({ text: "AAPL" })
```

### Input.dispatchMouseEvent

```js
// Move
{ type: "mouseMoved", x, y }
// Click
{ type: "mousePressed", x, y, button: "left", buttons: 1, clickCount: 1 }
{ type: "mouseReleased", x, y, button: "left", buttons: 0, clickCount: 1 }
// Scroll (wheel)
{ type: "mouseWheel", x, y, deltaX: 0, deltaY: amount }
```

---

## Pine Script — особенности

### pine_compile vs pine_smart_compile

- `pine_compile` — просто кликает кнопку "Add to chart" / "Update on chart"
- `pine_smart_compile` — определяет тип кнопки, кликает, затем читает ошибки и возвращает расширенный результат

### Кнопки компиляции (в порядке приоритета)

```
"Save and add to chart"
"Add to chart"
"Update on chart"
```

Если кнопка не найдена — fallback: `Input.dispatchKeyEvent(Ctrl+Enter)`.

### pine_analyze

Работает **offline**, без TradingView Desktop. Выполняет статический анализ Pine Script через собственный парсер (не CDP).

### pine_check

Отправляет запрос на **сервер TradingView** для проверки компиляции. Требует интернет, не требует Desktop.

---

## Replay — важное ограничение

`replay_trade` — торговля только в **режиме replay** (симуляция исторических баров). Это НЕ реальные сделки через брокера. Реализация через кнопки UI в TradingView Replay mode.

---

## UI Automation — React Fiber introspection

Некоторые core-модули используют React Fiber для доступа к внутренним компонентам:

```js
const fiber = Object.keys(el).find(k => k.startsWith("__reactFiber$"));
```

Go-порт должен воспроизводить те же JS-выражения через `Runtime.evaluate`.

---

## Streaming — паттерн

```js
// src/core/stream.js
// Команды stream никогда не возвращают — бесконечный цикл до SIGINT
while (true) {
    const data = await getQuote();
    process.stdout.write(JSON.stringify(data) + "\n");
    await sleep(intervalMs);
}
```

**Go-эквивалент** в `internal/stream/stream.go`:

```go
func StreamQuote(ctx context.Context, w io.Writer, interval time.Duration) error {
    for {
        select {
        case <-ctx.Done():
            return nil
        case <-time.After(interval):
            // fetch and write JSONL
        }
    }
}
```

---

## CLI exit codes

| Код | Значение |
| --- | --- |
| 0 | success |
| 1 | error |
| 2 | connection failure (CDP не доступен) |

---

## Tests в Node.js-проекте

| Файл | Что тестирует |
| --- | --- |
| `tests/cli.test.js` | CLI command router, arg parsing |
| `tests/e2e.test.js` | E2E с живым TradingView |
| `tests/pine_analyze.test.js` | Pine Script static analyzer |
| `tests/replay.test.js` | Replay mode |
| `tests/sanitization.test.js` | safeString, requireFinite |

Go golden test файлы должны располагаться в `tests/golden/<tool-name>.input.json` и `tests/golden/<tool-name>.output.json`.

---

## Известные проблемы Node.js-версии

1. **Windows TradingView discovery** — auto-detect плохо работает с Microsoft Store установкой. Подробности: `WINDOWS_TRADINGVIEW_DISCOVERY.md`.
2. **pine_analyze** — работает offline, но качество анализа ограничено внутренним парсером.
3. **batch_run** — может быть медленным без `delay_ms`; TradingView может throttle переключения символа.
