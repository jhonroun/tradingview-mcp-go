# TODO_codex.md — gap-closure pass

Правило закрытия пункта:

- изменение атомарное;
- публичные MCP tool names не сломаны;
- relevant tests/smokes сохранены в `research/`;
- `CHANGELOG.md` обновлён;
- если данные идут через TradingView internals, response/docs содержат `source`, `reliability`, `status` и/или `coverage`.

---

## 0. Planning reset

- [x] Очистить старый `PLAN.md` и заменить планом gap-closure pass.
- [x] Очистить старый `TODO_codex.md` и заменить чеклистом текущего pass.
- [x] Очистить старый `prompts.md` и заменить атомарными промптами текущего pass.

---

## 1. Pine localized compile

- [x] Найти текущий JS matcher Pine compile/Add-to-chart.
- [x] Поддержать EN labels:
  - [x] `Add to chart`
  - [x] `Update on chart`
  - [x] `Save and add to chart`
- [x] Поддержать RU labels:
  - [x] `Добавить на график`
  - [x] `Обновить на графике`
  - [x] `Сохранить и добавить на график`
- [x] Учесть дублированный `textContent`.
- [x] Не нажимать `Pine Save`, если есть видимая Add/Update button.
- [x] Добавить unit/contract test.
- [x] Запустить targeted tests.
- [x] Сохранить smoke/evidence в `research/pine-localized-compile/`.

Done:

- [x] `pine_compile` распознаёт RU Add-to-chart на уровне shared matcher.
- [x] `pine_smart_compile` распознаёт RU Add-to-chart live (`button_clicked: Добавить на графикДобавить на график`, `study_added:true`).
- [x] `go test ./internal/tools/pine` проходит.

---

## 2. Research/code/docs gap audit

- [x] Создать `research/implementation-gap-audit/`.
- [x] Создать `research-to-code-matrix.json`.
- [x] Создать `agents-skills-gap.json`.
- [x] Создать `README.md`.
- [x] Классифицировать research blocks:
  - [x] baseline
  - [x] cdp-evaluate-await
  - [x] locale parser
  - [x] indicator-study-model
  - [x] strategy-backtesting-api
  - [x] strategy-equity-extraction/full
  - [x] pine-source-safety
  - [x] study-limit-detection
  - [x] known-issues-closure
  - [x] implementation-final
- [x] Зафиксировать current registry count `85`.
- [x] Зафиксировать historical Node parity count `78`.

---

## 3. Documentation sync

- [x] Обновить `docs/en/README.md`.
- [x] Обновить `docs/en/tools.md`.
- [x] Обновить `docs/en/architecture.md`.
- [x] Обновить `docs/en/agents-skills.md`.
- [x] Обновить `docs/en/cli.md` по новым response contracts.
- [x] Обновить `docs/ru/README.md`.
- [x] Обновить `docs/ru/tools.md`.
- [x] Обновить `docs/ru/architecture.md`.
- [x] Обновить `docs/ru/agents-skills.md`.
- [x] Обновить `docs/ru/cli.md` по новым response contracts.
- [x] Обновить `TEST.md` count/checklists.
- [x] Обновить `agents/README.md`.

Docs must mention:

- [x] Current Go registry: `85`.
- [x] Node parity baseline: `78`.
- [x] Go-only extensions: `data_get_indicator_history`, `data_get_orders`, `pine_restore_source`, HTS aggregate tools.
- [x] `bidAskAvailable=false` semantics.
- [x] `Strategy Equity` plot requirement.
- [x] `coverage: loaded_chart_bars`.

---

## 4. Existing skills sync

- [x] `chart-analysis`
- [x] `error-handling`
- [x] `futures-roll`
- [x] `indicator-scan`
- [x] `json-contracts`
- [x] `llm-context`
- [x] `market-brief`
- [x] `multi-symbol-scan`
- [x] `pine-develop`
- [x] `replay-practice`
- [x] `strategy-report`

Each updated skill must include:

- [x] current MCP 85 note where relevant;
- [x] source/reliability/status policy;
- [x] no fake full-history equity;
- [x] Pine backup/restore safety where relevant.

---

## 5. Missing skills

- [x] Add `study-model-values`.
- [x] Add `strategy-backtesting-api`.
- [x] Add `strategy-equity-plot`.
- [x] Add `pine-safe-edit`.
- [x] Add `tradingview-limit-handling`.
- [x] Add `regression-smoke`.
- [x] Add `data-quality`.

---

## 6. Agents sync

Base agents:

- [x] `agents/market-analyst.md`
- [x] `agents/futures-analyst.md`
- [x] `agents/performance-analyst.md`

Client variants:

- [x] `agents/codex/*`
- [x] `agents/cline/*`
- [x] `agents/windsurf/*`
- [x] `agents/gemini/*`
- [x] `agents/cursor/*`
- [x] `agents/continue/*`

Agents must:

- [x] check `source/reliability` before trading-logic conclusions;
- [x] check strategy status before reporting metrics;
- [x] treat `loaded_chart_bars` as partial coverage;
- [x] treat unavailable bid/ask as unavailable, not zero.

---

## 7. Russian variants

Skills:

- [x] Create `SKILL.ru.md` for every existing skill.
- [x] Create `SKILL.ru.md` for every new skill.

Agents:

- [x] `agents/*.ru.md`
- [x] `agents/codex/*.ru.md`
- [x] `agents/cline/*.ru.md`
- [x] `agents/windsurf/*.ru.md`
- [x] `agents/gemini/*.ru.md`
- [x] `agents/cursor/*.ru.mdc`
- [x] `agents/continue/*.ru.prompt`

---

## 8. Verification

- [x] `go test ./...`
- [x] `go vet ./...`
- [x] MCP initialize smoke.
- [x] MCP tools/list count smoke (`85`).
- [x] MCP unknown tool error shape smoke.
- [x] Live Pine localized compile smoke if safe.
- [x] Create `research/agents-skills-sync/README.md`.
- [x] Update `CHANGELOG.md`.

---

## Current remaining known limitations

- TradingView internal paths are intentionally unstable and must remain marked as such.
- `data_get_equity` covers loaded chart bars unless TradingView has loaded the full required range.
- Full native Strategy Tester bar-by-bar equity remains unavailable; explicit Pine `Strategy Equity` plot is required for reliable equity extraction.

---

## 9. Release hardening for tag 1.2

- [x] Добавить structured compatibility probes в `tv_discover` без удаления legacy `paths`.
- [x] Покрыть compatibility probe JS unit tests.
- [x] Провести live `tv discover` и сохранить JSON в `research/compatibility-probes/`.
- [x] Документировать compatibility probes в EN/RU docs.
- [x] Во всех agents зафиксировать:
  - [x] TradingView internals нужно re-probe через `tv_discover`.
  - [x] `data_get_equity` = loaded-bars-only при `coverage: loaded_chart_bars`.
  - [x] derived equity = conditional, не native.
- [x] Во всех skills зафиксировать:
  - [x] compatibility probes перед workflows, зависящими от internals;
  - [x] equity loaded-bars-only;
  - [x] derived equity conditional;
  - [x] не искать full native equity до появления стабильного TradingView report field.
- [x] Добавить optional history-load workflow в docs/skills/agents:
  - [x] `chart_set_visible_range` / `chart_scroll_to_date`;
  - [x] повторить `data_get_indicator_history` / `data_get_equity`;
  - [x] сравнить `loaded_bar_count`, `data_points`, `coverage`;
  - [x] оставить статус best-effort.
- [x] Обновить `research/implementation-final/README.md` и снять устаревшее remaining про RU Add-to-chart helper.
- [x] Обновить `CHANGELOG.md`.
- [x] Финально запустить `go test ./...`.
- [x] Финально запустить `go vet ./...`.
- [x] Финально запустить MCP initialize/tools-list/unknown-tool smoke.
- [x] Сделать commit.
- [x] Создать tag `1.2`.
- [x] Push branch и tag.
