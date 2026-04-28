---
name: pine-safe-edit
description: Safely read, edit, compile, verify, and restore Pine source with backup and SHA256 guards.
---

# Pine Safe Edit

Use this before any Pine source mutation.

## Workflow

1. `pine_get_source`: capture source, `source_sha256`, script name/type.
2. `pine_set_source` with `expected_current_sha256` when replacing existing code.
3. Confirm response includes backup path and backup hash.
4. `pine_compile` or `pine_smart_compile`: read structured diagnostics.
5. Verify chart state or strategy/data output.
6. If needed, `pine_restore_source` with backup path and verify SHA256.

## Rules

- Never silently overwrite user code.
- Never skip backup verification.
- Do not claim compile success if `error_count > 0`.
- For equity strategies, include the explicit `Strategy Equity` plot when required.
## Release 1.2 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.

