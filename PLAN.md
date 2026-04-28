# PLAN.md — gap-closure pass для tradingview-mcp-go

## Цель

Закрыть оставшиеся разрывы после внедрения research-результатов:

1. исправить локализованный Pine compile/Add-to-chart путь;
2. повторно оценить соответствие `research`/`results` текущему коду;
3. синхронизировать документацию с фактическим MCP registry (`85` tools);
4. обновить существующие skills/agents под новые MCP-контракты;
5. добавить недостающие skills;
6. создать русские варианты для всех agents и skills;
7. сохранить evidence и финальный статус в `research/`.

Скоуп остаётся только `tradingview-mcp-go`: CDP/MCP/CLI/TradingView Desktop. Не добавлять HTS, брокеров, Telegram, orchestration, автоторговлю или multi-user системы.

## Фактическая база

- Историческая Node parity база: `78` tools.
- Текущий Go registry после стабилизации: `85` tools.
- Финальный implementation smoke: `research/implementation-final/`.
- Основные reliable data paths:
  - `tradingview_study_model`
  - `tradingview_backtesting_api`
  - `tradingview_strategy_plot`
- Все TradingView internal paths должны оставаться явно помеченными как unstable.

## Принципы

1. Не ломать имена MCP tools, аргументы и базовый success/error contract.
2. Не выдавать UI/canvas/display values за торговые numeric truth.
3. Для data tools проверять `source`, `reliability`, `status`, `coverage`, `reliableForTradingLogic`.
4. Любое изменение Pine source должно иметь backup/hash/restore path.
5. Docs/skills/agents не должны обещать больше, чем подтверждено code/tests/live smoke.
6. Делать атомарные изменения и после каждого смыслового блока запускать релевантные тесты.

## Этап 1. Pine localized compile

Проблема: в русском UI TradingView кнопка `Добавить на график` не распознаётся текущим compile helper, поэтому helper может нажать Save вместо Add-to-chart.

Сделать:

- вынести JS matcher кнопок Pine compile в переиспользуемый helper;
- поддержать EN/RU labels:
  - `Add to chart`
  - `Update on chart`
  - `Save and add to chart`
  - `Добавить на график`
  - `Обновить на графике`
  - `Сохранить и добавить на график`
- учесть дублированный `textContent`, например `Добавить на графикДобавить на график`;
- добавить unit/contract test на JS matcher;
- сохранить smoke в `research/pine-localized-compile/`.

## Этап 2. Research/code/docs gap audit

Создать `research/implementation-gap-audit/`:

- `research-to-code-matrix.json`
- `agents-skills-gap.json`
- `README.md`

Классификация:

- `implemented_and_tested`
- `implemented_live_smoked`
- `documented_limitation`
- `stale_docs_only`
- `real_code_gap`
- `out_of_scope`

Проверить минимум:

- CDP awaitPromise
- locale parser
- indicator study model
- strategy backtesting API
- equity strategy plot
- Pine safety
- study limit detection
- known issues closure
- implementation-final regression
- docs/agents/skills counts and contracts

## Этап 3. Documentation sync

Обновить документацию, чтобы она не расходилась с кодом:

- `docs/en/*`
- `docs/ru/*`
- `docs/dev/*` где это актуальные рабочие документы, а не исторические записи;
- `agents/README.md`
- root planning files.

Ключевые правки:

- текущий registry: `85` tools;
- historical Node parity: `78` tools;
- Go extensions: `+7`;
- новые tools/контракты: `data_get_indicator_history`, `data_get_orders`, `pine_restore_source`;
- `bidAskAvailable=false` для unavailable bid/ask;
- `coverage: loaded_chart_bars` для equity;
- `Strategy Equity` plot requirement.

## Этап 4. Skills sync and missing skills

Обновить существующие skills:

- `chart-analysis`
- `error-handling`
- `futures-roll`
- `indicator-scan`
- `json-contracts`
- `llm-context`
- `market-brief`
- `multi-symbol-scan`
- `pine-develop`
- `replay-practice`
- `strategy-report`

Добавить missing skills:

- `study-model-values`
- `strategy-backtesting-api`
- `strategy-equity-plot`
- `pine-safe-edit`
- `tradingview-limit-handling`
- `regression-smoke`
- `data-quality`

Каждый skill должен учитывать source/reliability/status policy.

## Этап 5. Agents sync

Обновить agents:

- `market-analyst`
- `futures-analyst`
- `performance-analyst`

Их client variants:

- root Claude agent files;
- `agents/codex/`
- `agents/cline/`
- `agents/windsurf/`
- `agents/gemini/`
- `agents/cursor/`
- `agents/continue/`

Agents должны:

- проверять `source/reliability` перед trading-logic выводами;
- не считать `bid/ask=0` валидным bid/ask без `bidAskAvailable`;
- различать `ok`, `no_strategy_loaded`, `needs_equity_plot`, `strategy_report_unavailable`;
- не считать `loaded_chart_bars` полной backtest history.

## Этап 6. Russian variants

Создать русские варианты для всех agents и skills.

Skills:

- `skills/<name>/SKILL.ru.md`

Agents:

- `agents/<agent>.ru.md`
- `agents/codex/<agent>.ru.md`
- `agents/cline/<agent>.ru.md`
- `agents/windsurf/<agent>.ru.md`
- `agents/gemini/<agent>.ru.md`
- `agents/cursor/<agent>.ru.mdc`
- `agents/continue/<agent>.ru.prompt`

Если клиент загружает только `SKILL.md`, `SKILL.ru.md` считается дистрибутивным вариантом. При необходимости позже можно добавить mirror tree `skills-ru/<name>/SKILL.md`.

## Этап 7. Final verification

Минимум:

- `go test ./...`
- `go vet ./...`
- MCP `initialize`
- MCP `tools/list` count = `85`
- MCP unknown tool error shape
- targeted Pine matcher tests
- live Pine localized compile smoke, если TradingView UI позволяет безопасный backup/restore cycle

Финальный отчёт:

- `research/agents-skills-sync/README.md`
- update `CHANGELOG.md`
- update `TODO_codex.md`

## Этап 8. Release hardening before tag 1.2

Перед коммитом/пушем/тегом явно закрыть оставшиеся архитектурные решения:

1. Compatibility probes для TradingView unstable internal paths:
   - расширить `tv_discover`, сохранив старое поле `paths`;
   - добавить structured `compatibility_probes` с `compatible`, `available`, `status`, `stability`, `reliability`;
   - покрыть probe JS unit tests;
   - сохранить live output в `research/compatibility-probes/`.
2. Equity policy:
   - во всех docs/agents/skills зафиксировать, что `data_get_equity` с `source: tradingview_strategy_plot` покрывает только `coverage: loaded_chart_bars`;
   - не называть это полной Strategy Tester history.
3. Optional history-load workflow:
   - описать workflow через chart navigation/visible range/scroll, повторное чтение `loaded_bar_count` и сравнение coverage;
   - явно указать, что это best-effort догрузка TradingView chart history, не новый backtesting engine.
4. Derived equity:
   - оставить только `conditional`;
   - `reliableForTradingLogic:false` или conditional по полным OHLCV/trades/settings;
   - не выдавать reconstruction за native equity.
5. Native full equity:
   - не тратить время на поиск/реализацию full native bar-by-bar Strategy Tester equity, пока TradingView не exposes стабильный report field.
6. Release:
   - финальные `go test ./...`, `go vet ./...`, MCP smoke;
   - commit;
   - tag `1.2`;
   - push ветки и tag.
