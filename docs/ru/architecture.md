# Архитектура, совместимость и дисклеймер

> [← Назад к документации](README.md)

---

## Архитектура

```
AI Client (Claude Code, Cursor, Cline...)
    │
    │  MCP stdio (JSON-RPC 2.0)
    ▼
tvmcp (MCP server)
    │
    │  Chrome DevTools Protocol (WebSocket)
    ▼
TradingView Desktop  ←→  localhost:9222
```

### Структура пакетов

```
cmd/tvmcp/        MCP stdio сервер
cmd/tv/           CLI утилита
internal/
  cdp/            WebSocket CDP клиент (Runtime, Page, DOM, Input)
  mcp/            JSON-RPC 2.0 сервер + реестр инструментов
  cli/            Диспетчер CLI-команд
  stream/         JSONL poll-and-dedup стриминг
  tradingview/    Константы JS-выражений + SafeString
  discovery/      Поиск TradingView (Win Store / AppData / macOS / Linux)
  launcher/       Запуск TradingView с --remote-debugging-port
  tools/
    health/       Проверка здоровья и запуск
    chart/        Состояние и управление графиком
    data/         OHLCV, котировки, индикаторы, стратегии
    capture/      Скриншот
    indicators/   Входные данные и видимость индикаторов
    pine/         Работа с Pine Script
    drawing/      Рисование фигур
    alerts/       Алерты и watchlist
    replay/       Replay-торговля
    pane/         Управление панелями
    tab/          Управление вкладками
    ui/           UI-автоматизация и лейауты
    batch/        Пакетная обработка
```

---

## Совместимость

Проект сохраняет исходную **Node.js parity базу из 78 tools** и добавляет Go-only stabilization helpers.

| Аспект | Node.js | Go-порт |
| --- | --- | --- |
| MCP tools | 78 | 85 текущих Go tools: 78 parity tools + 7 расширений |
| CLI-группы | 15 | 15+ |
| JSON-схемы аргументов | оригинал | идентичны |
| JSON-структура ответов | `{success, ...}` | идентична |
| CDP endpoint | `localhost:9222` | идентичен |
| JS-выражения | `chrome-remote-interface` | идентичны |
| Платформы | Win / macOS / Linux | Win / macOS / Linux |
| Windows Store | да | да (Get-AppxPackage) |

Детальная матрица совместимости: [docs/dev/COMPATIBILITY_MATRIX.md](../dev/COMPATIBILITY_MATRIX.md)

Go-only extensions включают `data_get_indicator_history`, `data_get_orders`, `pine_restore_source` и четыре агрегированных LLM/context tools. Data tools при необходимости возвращают `source`, `reliability`, `coverage`, `status` и `reliableForTradingLogic`, потому что часть TradingView paths является undocumented internal API.

`tv_discover` является entry point для compatibility probes этих internals. Он сохраняет legacy объект `paths` и добавляет `compatibility_probes` с `compatible`, `available`, `status`, `stability`, `reliability`. Запускайте его после обновлений TradingView Desktop до доверия workflows на study model, backtesting API или strategy equity plot.

Strategy equity намеренно описана как loaded chart data, а не полный native Strategy Tester export. Надёжный runtime path — explicit Pine `Strategy Equity` plot, прочитанный из loaded bars; derived equity остаётся conditional, а full native bar-by-bar equity out of scope, пока TradingView не exposes стабильный report field.

---

## Дисклеймер

Этот инструмент взаимодействует **только** с локально запущенным TradingView Desktop через Chrome DevTools Protocol на `localhost:9222`.

- Не подключается к серверам TradingView.
- Не исполняет реальные торговые операции.
- Не собирает и не передаёт рыночные данные за пределы локальной машины.
- Не обходит платные функции TradingView.
- Не аффилирован с TradingView Inc. или Anthropic.
