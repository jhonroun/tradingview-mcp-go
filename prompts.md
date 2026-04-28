# prompts.md — атомарные задания gap-closure pass

Каждый промпт предполагает:

1. прочитать `AGENTS.md`;
2. прочитать `PLAN.md`;
3. работать только в скоупе `tradingview-mcp-go`;
4. не добавлять HTS, brokers, Telegram, orchestration, автоторговлю;
5. делать маленькие изменения;
6. обновлять `TODO_codex.md`, `CHANGELOG.md` и research notes;
7. запускать релевантные тесты.

---

## PROMPT 01 — Pine localized compile

```md
Задача: исправить Pine compile/Add-to-chart matcher для русской и английской локали TradingView.

Проблема:
В RU UI кнопка может называться `Добавить на график` и иметь дублированный textContent:
`Добавить на графикДобавить на график`. Текущий helper ищет английские labels и может нажать Pine Save.

Требования:
1. Поддержать EN:
   - Add to chart
   - Update on chart
   - Save and add to chart
2. Поддержать RU:
   - Добавить на график
   - Обновить на графике
   - Сохранить и добавить на график
3. Не ломать compile result shape.
4. Добавить unit/contract test на JS matcher.
5. Сохранить smoke/evidence в `research/pine-localized-compile/`.

Done criteria:
- `pine_compile` и `pine_smart_compile` распознают RU Add-to-chart;
- targeted pine tests проходят;
- `go test ./...` проходит.
```

---

## PROMPT 02 — Research/code/docs gap audit

```md
Задача: повторно оценить покрытие research/results текущим кодом и документацией.

Сделать:
1. Создать `research/implementation-gap-audit/`.
2. Создать `research-to-code-matrix.json`.
3. Создать `agents-skills-gap.json`.
4. Создать README с выводами.
5. Для каждого блока поставить статус:
   - implemented_and_tested
   - implemented_live_smoked
   - documented_limitation
   - stale_docs_only
   - real_code_gap
   - out_of_scope

Done criteria:
- ясно видно, что покрыто кодом;
- ясно видно, что является limitation TradingView;
- ясно видно, где docs/agents/skills устарели.
```

---

## PROMPT 03 — Documentation sync

```md
Задача: синхронизировать документацию с текущим MCP registry и response contracts.

Контекст:
- Историческая Node parity база: 78 tools.
- Текущий Go registry: 85 tools.

Обновить:
- docs/en/*
- docs/ru/*
- TEST.md
- agents/README.md

Документировать:
- data_get_indicator_history
- data_get_orders
- pine_restore_source
- source/reliability/reliableForTradingLogic policy
- bidAskAvailable=false
- Strategy Equity plot requirement
- loaded_chart_bars coverage

Done criteria:
- больше нет актуальных docs, которые утверждают current registry = 78/82;
- docs различают historical parity и current Go extensions.
```

---

## PROMPT 04 — Existing skills sync

```md
Задача: обновить существующие skills под текущие MCP contracts.

Skills:
- chart-analysis
- error-handling
- futures-roll
- indicator-scan
- json-contracts
- llm-context
- market-brief
- multi-symbol-scan
- pine-develop
- replay-practice
- strategy-report

Требования:
- учитывать 85 current tools;
- проверять source/reliability/status;
- не считать loaded_chart_bars full history;
- Pine workflow должен использовать backup/hash/restore;
- strategy workflow должен учитывать no_strategy_loaded и needs_equity_plot.

Done criteria:
- каждый существующий skill обновлён;
- устаревшие 78/82 assumptions удалены.
```

---

## PROMPT 05 — Missing skills

```md
Задача: добавить недостающие skills для новых reliable workflows.

Добавить:
- study-model-values
- strategy-backtesting-api
- strategy-equity-plot
- pine-safe-edit
- tradingview-limit-handling
- regression-smoke
- data-quality

Done criteria:
- у каждого нового skill есть SKILL.md;
- каждый skill отражает текущие MCP status/source/reliability contracts.
```

---

## PROMPT 06 — Agents sync

```md
Задача: обновить market/futures/performance agents и все client variants.

Обновить:
- root agents
- codex
- cline
- windsurf
- gemini
- cursor
- continue

Требования:
- agents проверяют source/reliability перед trading logic;
- strategy agent проверяет status и equity coverage;
- futures agent учитывает bidAskAvailable=false;
- output не должен выдавать partial/derived data как reliable.

Done criteria:
- все agent variants синхронизированы с текущим MCP contract.
```

---

## PROMPT 07 — Russian variants

```md
Задача: создать русские варианты для всех agents и skills.

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

Done criteria:
- у каждого existing/new skill есть RU variant;
- у каждого agent/client variant есть RU variant;
- README объясняет, какие файлы использовать.
```

---

## PROMPT 08 — Final verification

```md
Задача: финальная проверка gap-closure pass.

Сделать:
1. `go test ./...`
2. `go vet ./...`
3. MCP initialize smoke
4. MCP tools/list count smoke
5. MCP unknown tool error shape smoke
6. Live Pine localized compile smoke, если безопасно
7. Создать `research/agents-skills-sync/README.md`
8. Обновить `CHANGELOG.md` и `TODO_codex.md`

Done criteria:
- все проверки сохранены в research;
- TODO отражает done/remaining;
- changelog дописан.
```

---

## PROMPT 09 — Release hardening: unstable paths, equity coverage, tag v1.2.0

```md
Задача: перед релизом v1.2.0 явно провести и зафиксировать compatibility/risk pass.

Сделать:
1. Добавить compatibility probes для TradingView unstable internal paths в `tv_discover`.
   - сохранить legacy `paths`;
   - добавить `compatibility_probes`;
   - для каждого probe вернуть `compatible`, `available`, `status`, `stability`, `reliability`.
2. Сохранить live output в `research/compatibility-probes/`.
3. Во всех docs/agents/skills зафиксировать:
   - TradingView internal paths unstable и требуют re-probe после обновлений TradingView;
   - `data_get_equity` с `source: tradingview_strategy_plot` покрывает только `loaded_chart_bars`;
   - derived equity является conditional и не native Strategy Tester equity;
   - full native bar-by-bar Strategy Tester equity не реализовывать, пока TradingView не exposes стабильный report field.
4. Добавить optional workflow для догрузки истории графика:
   - расширить visible range / scroll to older date;
   - повторить data tools;
   - сравнить `loaded_bar_count`, `data_points`, `coverage`;
   - оставить best-effort статус.
5. Обновить `TODO_codex.md`, `CHANGELOG.md`, research notes.
6. Запустить `go test ./...`, `go vet ./...`, MCP smoke.
7. Сделать commit, tag `v1.2.0`, push branch и tag.

Done criteria:
- compatibility probes есть в `tv_discover` и сохранены в research;
- docs/agents/skills не обещают full equity;
- derived equity не выдана за reliable native equity;
- финальные проверки прошли;
- tag `v1.2.0` опубликован.
```

---

## PROMPT 09 — Release hardening: unstable paths, equity coverage, tag v1.2.0

```md
Задача: перед релизом v1.2.0 явно провести и зафиксировать compatibility/risk pass.

Сделать:
1. Добавить compatibility probes для TradingView unstable internal paths в `tv_discover`.
   - сохранить legacy `paths`;
   - добавить `compatibility_probes`;
   - для каждого probe вернуть `compatible`, `available`, `status`, `stability`, `reliability`.
2. Сохранить live output в `research/compatibility-probes/`.
3. Во всех docs/agents/skills зафиксировать:
   - TradingView internal paths unstable и требуют re-probe после обновлений TradingView;
   - `data_get_equity` с `source: tradingview_strategy_plot` покрывает только `loaded_chart_bars`;
   - derived equity является conditional и не native Strategy Tester equity;
   - full native bar-by-bar Strategy Tester equity не реализовывать, пока TradingView не exposes стабильный report field.
4. Добавить optional workflow для догрузки истории графика:
   - расширить visible range / scroll to older date;
   - повторить data tools;
   - сравнить `loaded_bar_count`, `data_points`, `coverage`;
   - оставить best-effort статус.
5. Обновить `TODO_codex.md`, `CHANGELOG.md`, research notes.
6. Запустить `go test ./...`, `go vet ./...`, MCP smoke.
7. Сделать commit, tag `v1.2.0`, push branch и tag.

Done criteria:
- compatibility probes есть в `tv_discover` и сохранены в research;
- docs/agents/skills не обещают full equity;
- derived equity не выдана за reliable native equity;
- финальные проверки прошли;
- tag `v1.2.0` опубликован.
```


