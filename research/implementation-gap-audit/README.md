# Implementation Gap Audit

Date: 2026-04-27

## Result

The implementation research line is covered by code, tests, smoke outputs, or documented TradingView limitations.

Current registry facts:

- Historical Node parity baseline: `78` tools.
- Current Go registry: `85` tools.
- Go extensions include `data_get_indicator_history`, `data_get_orders`, `pine_restore_source`, and aggregate LLM/context helpers.

## Closed Code Gaps

- CDP evaluate helper supports `awaitPromise=true`.
- Locale display number parser handles RU/EN formats.
- Indicator values and history use TradingView study model paths.
- Strategy results/trades/orders use `TradingViewApi.backtestingStrategyApi()`.
- Equity extraction uses explicit `Strategy Equity` plot with `coverage: loaded_chart_bars`.
- Pine source write path creates backups and restore verifies SHA256.
- Study limit detection returns structured `study_limit_reached`.
- Known issues around symbol search, MOEX bid/ask, and screenshot filenames are covered.
- Pine compile button matcher now includes Russian Add-to-chart labels.

## Closed Docs/Skills/Agents Gaps

- Active docs now distinguish current Go `85` tools from historical Node parity `78`.
- Existing skills were updated with source/reliability/status notes.
- New skills were added for study-model values, strategy API, equity plot, Pine safe edit, TradingView limits, regression smoke, and data quality.
- Russian variants were created for every skill.
- Agents were regenerated from source prompts for all supported clients.
- Russian variants were created for every agent/client wrapper.

## Documented Limitations

- TradingView internal paths remain unstable by nature.
- `data_get_equity` returns loaded chart bars unless TradingView has loaded the full required range.
- Full native Strategy Tester bar-by-bar equity was not found in the verified report path.
- Study-limit live cap was not reproduced in the current account/session, though parser/tests and structured status handling are implemented.

## Release 1.2 Resolution

- Unstable internal paths are handled by `tv_discover.compatibility_probes`; this is a compatibility/risk-control mechanism, not a guarantee that TradingView internals are stable.
- Equity remains loaded-bars-only and documented across docs, agents, and skills.
- Optional history loading is documented as best-effort chart range expansion/scrolling followed by repeated data calls and `loaded_bar_count` / `data_points` comparison.
- Derived equity remains conditional and must not be presented as native Strategy Tester equity.
- Full native bar-by-bar Strategy Tester equity remains out of scope until TradingView exposes a stable report field.

## Machine-Readable Files

- `research-to-code-matrix.json`
- `agents-skills-gap.json`
