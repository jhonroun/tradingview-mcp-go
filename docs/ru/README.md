# tradingview-mcp-go — документация (RU)

> [English version](../en/README.md) | [Корень проекта](../../README.md)

AI Go-порт проекта [tradesdontlie/tradingview-mcp](https://github.com/tradesdontlie/tradingview-mcp).

Подключает любой MCP-клиент к живому чарту **TradingView Desktop** через Chrome DevTools Protocol.

> Не является официальным продуктом TradingView Inc. или Anthropic.  
> Использование должно соответствовать Условиям использования TradingView.

---

## Навигация

| Раздел | Файл |
| --- | --- |
| О проекте, история создания, требования | эта страница |
| Установка, запуск CDP, настройка MCP, провайдеры | [install.md](install.md) |
| MCP-инструменты (78 штук) | [tools.md](tools.md) |
| CLI-команды и скрипты | [cli.md](cli.md) |
| Навыки и агенты | [agents-skills.md](agents-skills.md) |
| Архитектура, совместимость, дисклеймер | [architecture.md](architecture.md) |

---

## О проекте

`tradingview-mcp-go` — MCP-сервер и CLI-утилита, которые позволяют AI-ассистентам (Claude Code, Cursor, Cline и другим MCP-клиентам) взаимодействовать с запущенным TradingView Desktop:

- читать и управлять графиком (символ, таймфрейм, индикаторы);
- получать рыночные данные в реальном времени (OHLCV, котировки, стратегии);
- работать с Pine Script (чтение, редактирование, компиляция, анализ);
- рисовать фигуры, управлять алертами, watchlist, panes, tabs;
- запускать backtest в режиме replay;
- управлять интерфейсом TradingView программно;
- стримить данные в формате JSONL.

Всё взаимодействие — **только локально**, через Chrome DevTools Protocol на `localhost:9222`. Данные не передаются на внешние серверы.

---

## История создания

### Оригинальный проект

Репозиторий [`tradesdontlie/tradingview-mcp`](https://github.com/tradesdontlie/tradingview-mcp) — реализация MCP-сервера на **Node.js**, предоставляющая 78 инструментов для работы с TradingView Desktop через CDP.

### Портирование на Go с помощью Claude Code

В апреле 2026 года оригинальный Node.js-проект был полностью перенесён на **Go** с сохранением поведения 1:1.

Весь процесс портирования выполнялся при непосредственном участии **Claude Code** (CLI от Anthropic) — AI-ассистента для разработчиков. Claude Code:

- проводил инвентаризацию Node.js-кода и составлял матрицу совместимости;
- последовательно реализовывал каждый модуль (MCP, CDP, tools, CLI, stream);
- писал unit-тесты после каждого этапа и проверял `go test ./...`;
- вёл `CHANGELOG.md` и `TODO.md` по ходу работы;
- обеспечивал 1:1 совместимость по именам инструментов, схемам аргументов и структуре JSON-ответов.

Итог: **78/78 MCP tools, 83+ CLI-команды, 78 unit-тестов** — за одну сессию.

**Результат:** полностью функциональный Go-бинарник без зависимостей от Node.js, `npm` или `chrome-remote-interface`.

---

## Требования

| Компонент | Версия |
| --- | --- |
| Go | 1.21+ |
| TradingView Desktop | Windows / macOS / Linux |
| TradingView запущен с | `--remote-debugging-port=9222` |

Зависимостей от Node.js нет. Единственная Go-зависимость: `github.com/gorilla/websocket`.
