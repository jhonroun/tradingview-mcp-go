---
name: error-handling
description: Handle MCP tool errors correctly — distinguish retryable (CDP disconnect, timeout) from permanent (unknown tool, bad args) errors. Use when a tool returns success=false or an unexpected result.
---

# MCP Error Handling

You are deciding how to respond to a failed MCP tool call from tradingview-mcp-go.

## Error Envelope

Every tool call returns one of two shapes:

```json
{ "success": true,  ...tool fields... }
{ "success": false, "error": "<human-readable message>" }
```

Always check `success` before accessing other fields.

## Error Classification

### Retryable errors — wait and retry

| Condition | `error` contains | Action |
|-----------|-----------------|--------|
| CDP not connected | `"CDP"` or `"connect"` | Wait 3–5 s, retry. If still failing, ask user to check TradingView is running with `--remote-debugging-port=9222`. |
| TradingView tab not found | `"no TradingView"` | Ask user to open TradingView in the browser or use `tv launch`. |
| JS evaluation timeout | `"timeout"` | Retry once. If still failing, the chart may be loading — wait 5 s. |
| WebSocket closed | `"websocket"` or `"WebSocket"` | Retry once — the CDP connection will auto-reconnect. |

### Permanent errors — do not retry

| Condition | `error` contains | Action |
|-----------|-----------------|--------|
| Unknown tool name | `"unknown tool"` | Verify the tool name exactly. Do not retry. |
| Bad argument type | `"unmarshal"` or `"invalid"` | Fix the argument shape. Check the tool's inputSchema. |
| Missing required field | `"is required"` | Add the missing argument. |

## Decision Tree

```
tool returns success=false?
├── error contains "CDP" / "connect" / "websocket"
│   └── RETRYABLE: wait 3–5 s, retry up to 3 times
│       └── still failing? → prompt user: is TradingView running with CDP?
├── error contains "no TradingView"
│   └── RETRYABLE: prompt user to open TradingView or run `tv launch`
├── error contains "timeout"
│   └── RETRYABLE: retry once after 5 s
├── error contains "unknown tool"
│   └── PERMANENT: check tool name, do not retry
├── error contains "unmarshal" / "invalid" / "is required"
│   └── PERMANENT: fix arguments, do not retry
└── other
    └── UNKNOWN: retry once; if still failing, report error to user
```

## Retry Pattern

For retryable errors:

```
attempt = 1
max_attempts = 3
while attempt <= max_attempts:
    result = call_tool(...)
    if result.success: break
    if is_permanent_error(result.error): report and stop
    wait(3s * attempt)
    attempt++
report("tool failed after 3 attempts: " + result.error)
```

## Diagnosing CDP Disconnect

If you see `"CDP"` or `"connect"` errors:

1. Call `tv_health_check` — returns `{"connected": true/false, "target_url": "..."}`
2. If `connected: false`:
   - Ask user to run `tv launch` or start TradingView manually
   - After launch, retry after 5 s
3. If `connected: true` but tool still fails:
   - The specific tool's JS expression may have failed
   - Try `tv doctor` for detailed diagnostics

## Common Error Messages

| Message | Meaning | Fix |
|---------|---------|-----|
| `"CDP connection refused at localhost:9222"` | TradingView not running in debug mode | Run `tv launch` or start TV with `--remote-debugging-port=9222` |
| `"no TradingView chart tab found"` | TV running but wrong tab | Open a chart in TradingView |
| `"Study not found: Study_RSI_0"` | Entity ID no longer valid | Call `chart_get_state` to refresh indicator list |
| `"could not retrieve quote; the chart may still be loading"` | Chart not fully loaded | Wait 2–3 s and retry |
| `"could not extract OHLCV data"` | Same as above | Wait and retry |
| `"entity_id is required"` | Missing argument | Add `entity_id` to the call |

## Windows / MSIX Note

On Windows, TradingView MSIX cannot be launched from a non-interactive subprocess (MCP server context). The server will start but Chrome exits immediately.

**Workaround:** The user must run `tv launch` from a terminal (or start TradingView manually) before using the MCP server. Once TradingView is running with CDP, all tools work normally.

Detect this condition:
- `tv_health_check` returns `connected: false`
- Error message contains `"connection refused"` on port 9222

Response to user:
```
TradingView is not running with the remote debugging port.
Please run: tv launch
Or start TradingView manually. Once it's running, the tools will connect automatically.
```

## Current MCP Contract Notes

- Current Go registry: 85 MCP tools; original Node parity baseline: 78 tools.
- Data unavailability is often expressed as structured `status`, not only `success:false`.
- Strategy statuses include `ok`, `no_strategy_loaded`, `strategy_report_unavailable`, `strategy_report_shape_unverified`, and `tradingview_backtesting_api_unavailable`.
- Equity-specific unavailable status can be `needs_equity_plot`.
- Study-limit status is `study_limit_reached`; automatic deletion is forbidden unless `allow_remove_any=true` is explicit.
- `bidAskAvailable:false` means bid/ask are unavailable, even if compatibility fields contain `0`.
## Release v1.2.0 Data Guards

- Run `tv discover` and inspect `compatibility_probes` after TradingView Desktop updates or when an internal-path-dependent tool returns unavailable statuses.
- Treat `coverage: loaded_chart_bars` as chart-loaded coverage only, including strategy equity from `data_get_equity`.
- Use the optional history-load workflow only as best effort: expand/scroll the chart range, wait for bars to load, repeat the data call, and compare `loaded_bar_count` / `data_points`.
- Keep derived equity conditional; do not present it as native Strategy Tester equity or as unqualified `reliableForTradingLogic:true` data.
- Do not pursue full native bar-by-bar Strategy Tester equity until TradingView exposes a stable report field.


