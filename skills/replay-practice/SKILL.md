---
name: replay-practice
description: Practice trading in TradingView replay mode — step through historical bars, take trades, track P&L. Use when the user wants to practice or backtest manually.
---

# Replay Practice Trading

You are guiding the user through replay-mode practice trading on TradingView.

## Step 1: Setup

1. `chart_set_symbol` — set the desired symbol
2. `chart_set_timeframe` — set the trading timeframe
3. `replay_start` with a date — enter replay mode at the starting point

## Step 2: Pre-Trade Analysis

Before stepping through bars:
1. `data_get_ohlcv` — get the historical context leading up to the replay point
2. Add any indicators the user wants: `chart_manage_indicator`
3. `capture_screenshot` — show the starting chart state

## Step 3: Step Through Bars

Use `replay_step` to advance one bar at a time, or `replay_autoplay` for continuous play.

After each significant move:
1. `replay_status` — check current date, position, and P&L
2. Announce what happened (breakout, support test, etc.)

Valid autoplay speeds (ms): 100, 143, 200, 300, 1000, 2000, 3000, 5000, 10000

## Step 4: Execute Trades

When the user identifies an entry:
- `replay_trade` with action "buy" or "sell"
- `replay_status` to confirm the position was opened

When the user wants to exit:
- `replay_trade` with action "close"
- `replay_status` to show the P&L

## Step 5: Review

After the practice session:
1. `replay_status` — final P&L summary
2. `capture_screenshot` — capture the final chart state
3. `replay_stop` — exit replay mode

Report:
- Total trades taken
- Win/loss record
- Net P&L
- Key lessons from the session

## Tips

- Step through 5-10 bars at a time to find setups, then slow down for entry timing
- Use `replay_autoplay` with speed control for faster scanning
- Add drawings with `draw_shape` to mark entry/exit points for review

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- Replay outputs are UI-state dependent and should be treated as control/practice state, not broker truth.
- If replay status cannot be verified live, report it as partial or unverified.
- Use screenshots for visual confirmation when replay state is ambiguous.
## Release v1.2.0 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.


