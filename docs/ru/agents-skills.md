# Навыки и агенты

> [← Назад к документации](README.md)

В репозитории есть английские и русские workflow-файлы для всех agents и skills.

Текущий Go MCP registry: `85` tools. Историческая Node parity база: `78` tools.

## Skills

У каждого skill есть:

- English: `skills/<name>/SKILL.md`
- Russian: `skills/<name>/SKILL.ru.md`

| Skill | Для чего |
| --- | --- |
| `chart-analysis` | Полный анализ графика со screenshot context |
| `data-quality` | Проверки source/reliability/coverage |
| `error-handling` | Retryability и structured statuses |
| `futures-roll` | Continuous futures и roll context |
| `indicator-scan` | Таблицы сигналов индикаторов |
| `json-contracts` | Проверка MCP JSON responses |
| `llm-context` | Компактный LLM-ready context |
| `market-brief` | Market brief |
| `multi-symbol-scan` | Скан нескольких symbols |
| `pine-develop` | Цикл разработки Pine |
| `pine-safe-edit` | Pine edits с backup/hash/restore |
| `regression-smoke` | Regression/live smoke workflow |
| `replay-practice` | Практика в Replay |
| `strategy-backtesting-api` | Strategy report/trades/orders |
| `strategy-equity-plot` | Workflow explicit Strategy Equity plot |
| `strategy-report` | Отчёт по Strategy performance |
| `study-model-values` | Надёжные study model values/history |
| `tradingview-limit-handling` | Обработка лимита studies |

## Agents

У каждого agent есть EN/RU variants для каждого клиента.

| Agent | Для чего |
| --- | --- |
| `market-analyst` | Live chart reads и indicator-aware market briefs |
| `futures-analyst` | Continuous futures, roll context, bid/ask caveats |
| `performance-analyst` | Strategy Tester metrics, trades, orders, equity |

Root Claude agent files:

- `agents/market-analyst.md`, `agents/market-analyst.ru.md`
- `agents/futures-analyst.md`, `agents/futures-analyst.ru.md`
- `agents/performance-analyst.md`, `agents/performance-analyst.ru.md`

Client variants:

- Cursor: `agents/cursor/*.mdc`, `agents/cursor/*.ru.mdc`
- Cline: `agents/cline/*.md`, `agents/cline/*.ru.md`
- Windsurf: `agents/windsurf/*.md`, `agents/windsurf/*.ru.md`
- Continue: `agents/continue/*.prompt`, `agents/continue/*.ru.prompt`
- Codex: `agents/codex/*.md`, `agents/codex/*.ru.md`
- Gemini: `agents/gemini/*.md`, `agents/gemini/*.ru.md`

## Правила надёжности

Agents и skills должны проверять `source`, `reliability`, `status`, `coverage` и `reliableForTradingLogic` перед trading-logic выводами.

- Compatibility: после обновлений TradingView Desktop или unavailable statuses запускайте `tv discover` и проверяйте `compatibility_probes`.
- Indicator values: предпочитать `tradingview_study_model`.
- Strategy metrics/trades/orders: требовать `source: tradingview_backtesting_api` и `status: ok`.
- Equity: требовать explicit Pine plot `Strategy Equity`; `coverage: loaded_chart_bars` — частичное покрытие, не full Strategy Tester history.
- Optional history loading: расширить/проскроллить диапазон графика, дождаться догрузки баров TradingView, повторить data call, затем сравнить `loaded_bar_count` и `data_points`.
- Derived equity: только conditional; не называть native TradingView equity.
- Full native bar-by-bar Strategy Tester equity: не реализовывать и не обещать, пока TradingView не exposes стабильный report field.
- Bid/ask: использовать только при `bidAskAvailable:true`.
