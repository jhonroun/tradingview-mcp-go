---
name: pine-develop
description: Full Pine Script development loop — write code, compile, fix errors, iterate. Use when building a new indicator or strategy in TradingView.
---

# Pine Script Development Loop

You are developing a Pine Script indicator or strategy in TradingView. Follow this loop precisely.

## Step 1: Understand the Goal

If not already clear, ask the user:
- What type? (indicator, strategy, library)
- What does it do? (entry/exit logic, overlay, oscillator, etc.)
- Overlay or separate pane?
- Any specific inputs or visual elements?

## Step 2: Pull Current Source (if modifying)

If modifying an existing script, read the current source:

```bash
tv pine get > scripts/current.pine
# or use the wrapper:
bash scripts/pine_pull.sh
```

Then read `scripts/current.pine` to understand what's there.

Alternatively, use the MCP tool directly:
- `pine_get_source` — returns the current editor source as a string

If creating new: use `pine_new` (type: indicator/strategy/library) for a clean template.

## Step 3: Write the Pine Script

Write the complete script. Every script MUST include:
- `//@version=6` header
- Proper `indicator()` or `strategy()` declaration
- All user inputs with `input.*()` functions and groups
- Clear comments for each logical section

For strategies, include:
- `strategy.entry()` and `strategy.exit()` calls
- Position sizing via `strategy()` declaration
- Default commission and slippage settings

## Step 4: Push and Compile

Push the source into TradingView and compile:

```bash
tv pine set "$(cat scripts/current.pine)"
tv pine smart-compile
# or use the wrapper:
bash scripts/pine_push.sh
```

Alternatively, use MCP tools:
- `pine_set_source` — inject source into Monaco editor
- `pine_smart_compile` — compile and report errors + study_added flag

## Step 5: Fix Errors

Check errors:
```bash
tv pine errors
```

Or via MCP: `pine_get_errors`

If errors are reported:
1. Read the error messages (line number + description)
2. Fix the specific lines
3. Push again and recompile
4. Repeat until 0 errors

Common Pine Script errors:
- **"Mismatched input"** — indentation issue (Pine uses 4-space, not braces)
- **"Could not find function or function reference"** — typo or wrong version
- **"Undeclared identifier"** — variable used before declaration
- **"Cannot call X with argument type Y"** — wrong parameter type

You can also run a pre-flight check without opening TradingView:
```bash
tv pine check "$(cat scripts/current.pine)"
```
Or: `pine_check` (MCP) — sends source to TradingView's public compile endpoint.

And run an offline static analysis:
```bash
tv pine analyze "$(cat scripts/current.pine)"
```
Or: `pine_analyze` (MCP) — detects array out-of-bounds, missing strategy declaration, old version.

## Step 6: Verify on Chart

After clean compilation:
1. `capture_screenshot` — verify the chart looks right
2. `data_get_strategy_results` — if it's a strategy, check performance
3. Show the user the results

## Step 7: Save

```bash
tv pine save
```
Or: `pine_save` (MCP) — Ctrl+S + handles the save dialog.

## Step 8: Iterate

If the user wants changes:
1. Pull fresh: `bash scripts/pine_pull.sh`
2. Edit locally
3. Push + compile: `bash scripts/pine_push.sh`
4. Screenshot to verify

IMPORTANT: Always compile after every change. Never claim "done" without a clean compile.

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- `pine_get_source` returns source hash, script name, and script type when available.
- `pine_set_source` creates a backup before writing and supports an `expected_current_sha256` guard.
- `pine_restore_source` restores from backup and verifies SHA256.
- Do not silently overwrite user Pine code. Use get source -> backup/hash -> set source -> compile -> verify -> restore.
- `pine_compile` and `pine_smart_compile` support English and Russian Add-to-chart labels.
## Release v1.2.0 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.


