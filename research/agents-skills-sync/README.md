# Agents / Skills Sync Report

Date: 2026-04-27

## Result

Gap-closure pass completed.

- Current Go MCP registry: `85` tools.
- Historical Node parity baseline: `78` tools.
- Skills: `18`
- Skills with Russian variants: `18`
- Russian agent client variants: `21`

## Code Fix

Pine compile/Add-to-chart matcher now supports English and Russian TradingView UI labels:

- `Add to chart`
- `Update on chart`
- `Save and add to chart`
- `Добавить на график`
- `Обновить на графике`
- `Сохранить и добавить на график`

It also handles duplicated DOM text such as `Добавить на графикДобавить на график`.

Live smoke evidence:

- `research/pine-localized-compile/smart-compile.json`
- `button_clicked: Добавить на графикДобавить на график`
- `study_added: true`
- original Pine source restored and SHA256 verified

## Docs / Agents / Skills

Updated:

- `docs/en/*`
- `docs/ru/*`
- `TEST.md`
- `agents/README.md`
- `prompts/*.md`
- `agents/*`
- `skills/*`

Added missing skills:

- `study-model-values`
- `strategy-backtesting-api`
- `strategy-equity-plot`
- `pine-safe-edit`
- `tradingview-limit-handling`
- `regression-smoke`
- `data-quality`

Created Russian variants:

- `skills/<name>/SKILL.ru.md` for every skill
- `agents/<agent>.ru.md`
- `agents/codex/<agent>.ru.md`
- `agents/cline/<agent>.ru.md`
- `agents/windsurf/<agent>.ru.md`
- `agents/gemini/<agent>.ru.md`
- `agents/cursor/<agent>.ru.mdc`
- `agents/continue/<agent>.ru.prompt`

## Verification

- `go test ./...`: passed, exit code `0`
- `go vet ./...`: passed, exit code `0`
- MCP initialize: passed
- MCP `tools/list`: `85`
- MCP unknown tool error shape: `isError: true`

Machine-readable summary:

- `summary.json`

## Remaining Limitations

- TradingView internal paths remain undocumented and unstable by nature.
- Equity from `data_get_equity` remains loaded-chart-bars coverage unless TradingView has loaded the full required range.
- Native full Strategy Tester bar-by-bar equity was not found in verified report paths; explicit Pine `Strategy Equity` plot remains required.
- Live TradingView study limit cap was not reproduced in this session, though detection and tests are implemented.

## Release 1.2 Hardening Notes

- `tv_discover` now exposes structured `compatibility_probes` so agents/skills can check unstable internal paths before relying on study model, backtesting API, or strategy equity plot workflows.
- All agent variants and all EN/RU skills now include release `1.2` data guards:
  - `coverage: loaded_chart_bars` is chart-loaded coverage only;
  - optional history loading is best-effort and must compare `loaded_bar_count` / `data_points`;
  - derived equity remains conditional and is not native Strategy Tester equity;
  - full native bar-by-bar Strategy Tester equity is not a target until TradingView exposes a stable report field.
