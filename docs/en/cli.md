# CLI Commands and Scripts

> [← Back to docs](README.md)

---

## CLI (`tv`)

```bash
# Health
tv status                          # CDP health check
tv launch [--port N] [--no-kill] [--tv-path PATH]
tv doctor                          # Diagnose installation + CDP
tv discover                        # Available API paths
tv ui-state                        # Open panels

# Chart
tv chart-state
tv set-symbol SYMBOL
tv set-timeframe TF                # 1 5 15 60 D W M
tv set-type TYPE                   # Candles HeikinAshi Line Area ...

# Data
tv quote [SYMBOL]
tv ohlcv [--count N] [--summary]
tv screenshot [--region full|chart|strategy_tester] [--filename F]

# Symbols
tv symbol-info
tv symbol-search QUERY [--type T] [--exchange E]

# Indicators
tv indicator-toggle ENTITY_ID [--visible=true|false]

# Pine Script
tv pine get
tv pine set "SOURCE"
tv pine compile
tv pine smart-compile
tv pine raw-compile
tv pine errors
tv pine console
tv pine save
tv pine new [indicator|strategy|library]
tv pine open NAME
tv pine list
tv pine analyze "SOURCE"
tv pine check "SOURCE"

# Drawing
tv draw shape --time T --price P [--time2 T2 --price2 P2] [--text TEXT]
tv draw list
tv draw get ENTITY_ID
tv draw remove ENTITY_ID
tv draw clear

# Panes
tv pane list
tv pane set-layout LAYOUT
tv pane focus INDEX
tv pane set-symbol SYMBOL

# Tabs
tv tab list
tv tab new
tv tab close
tv tab switch ID

# Replay
tv replay start [--date YYYY-MM-DD]
tv replay step
tv replay stop
tv replay status
tv replay autoplay [--speed MS]
tv replay trade buy|sell|close

# Alerts & Watchlist
tv alert list
tv alert create --price P [--message MSG]
tv alert delete [--all]
tv watchlist get
tv watchlist add SYMBOL

# UI Automation
tv ui click --by aria-label --value "..."
tv ui open-panel PANEL [--action open|close|toggle]
tv ui fullscreen
tv ui keyboard KEY [--modifiers ctrl,shift]
tv ui type TEXT
tv ui hover --by aria-label --value "..."
tv ui scroll up|down [--amount N]
tv ui mouse X Y [--button right] [--double]
tv ui find QUERY [--strategy text|css|aria-label]
tv ui eval "JS_EXPRESSION"

# Layouts
tv layout list
tv layout switch NAME

# Batch
tv batch --symbols SYM1,SYM2 --action screenshot|get_ohlcv|get_strategy_results \
         [--timeframes TF1,TF2] [--delay MS] [--count N]

# HTS — LLM-ready composite commands (Phase 4)
tv context [--top-n N]             # chart state + price + top-N indicators in one call
tv indicator NAME                  # current value + signal for a named indicator
tv market                          # full market summary (OHLCV, change%, volume vs avg, indicators)
tv futures-context                 # continuous contract info (NG1!, ES1!, CL2!, …)

# JSONL Streams (Ctrl+C to stop)
tv stream quote    [--interval MS]
tv stream bars     [--interval MS]
tv stream values   [--interval MS]
tv stream lines    [--interval MS] [--filter STUDY]
tv stream labels   [--interval MS] [--filter STUDY]
tv stream tables   [--interval MS] [--filter STUDY]
tv stream all      [--interval MS]
```

---

## `tv doctor` — Windows diagnostics

`tv doctor` probes the local system and returns a structured JSON report with
actionable hints. Useful when `tv status` fails and you need to know why.

```bash
tv doctor
```

### Output fields

| Field | Type | Description |
| --- | --- | --- |
| `port.reachable` | bool | `true` if `localhost:9222` responds |
| `port.cdp` | bool | `true` if the response is a valid CDP target list |
| `port.owner` | string | Process name that owns port 9222 (e.g. `"chrome.exe"`) |
| `port.error` | string | Human-readable reason if not reachable |
| `process.running` | bool | `true` if `TradingView.exe` is found in the process list |
| `process.pid` | int | PID of the running process |
| `process.has_cdp_flag` | bool | `true` if `--remote-debugging-port` is in its command line |
| `process.cmdline` | string | Full command line of the running process |
| `install.found` | bool | `true` if the TradingView executable was located |
| `install.path` | string | Absolute path to `TradingView.exe` |
| `install.source` | string | Where it was found (`LOCALAPPDATA`, `Microsoft Store`, etc.) |
| `install.is_msix` | bool | `true` for Microsoft Store (WindowsApps) installs |
| `install.local_appdata_dir` | string | Path to `%LOCALAPPDATA%\TradingView` if it exists |
| `install.appdata_dir` | string | Path to `%APPDATA%\TradingView` if it exists |
| `launch_cmd` | string | Exact shell command to start TradingView with CDP |
| `hints` | string[] | Ordered actionable messages explaining what to fix |

### Example — TradingView not running

```json
{
  "port":    { "reachable": false, "cdp": false, "error": "connection refused — port not listening" },
  "process": { "running": false },
  "install": { "found": true, "path": "C:\\Users\\you\\AppData\\Local\\TradingView\\TradingView.exe", "source": "LOCALAPPDATA" },
  "launch_cmd": "cd \"C:\\Users\\you\\AppData\\Local\\TradingView\" && \"C:\\...\\TradingView.exe\" --remote-debugging-port=9222",
  "hints": ["TradingView is not running. Start it: tv launch"]
}
```

---

## HTS-ready composite tools (Phase 4)

Four tools that reduce multi-call round-trips for LLM integration.
Each combines several underlying MCP tools into a single request.

### `chart_context_for_llm`

Aggregates `chart_get_state` + `quote_get` + top-N study values into one structured object
and produces a compact `context_text` string ready for LLM prompt injection.

#### Arguments

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `top_n` | integer | 5 | Max number of indicators to include |

#### Response fields

| Field | Type | Description |
| --- | --- | --- |
| `symbol` | string | Current chart symbol |
| `timeframe` | string | Current timeframe (e.g. `"D"`) |
| `chart_type` | string | Chart type code |
| `price` | object | `{last, open, high, low, close, volume}` — last bar snapshot |
| `indicators` | array | Top-N study objects with `name` and `values` |
| `indicator_count` | int | Number of indicators included |
| `context_text` | string | `"Symbol: X \| TF: D \| Price: 150 \| RSI(RSI): 65.3 \| …"` |

#### CLI

```bash
tv context              # top 5 indicators (default)
tv context --top-n 10   # top 10 indicators
```

---

### `indicator_state`

Finds a study by partial name match and classifies its current value as a direction + signal,
sparing the LLM from interpreting raw value arrays.

#### Arguments

| Field | Type | Description |
| --- | --- | --- |
| `name` | string | Partial, case-insensitive indicator name (e.g. `"RSI"`, `"MACD"`) |

#### Response fields

| Field | Type | Description |
| --- | --- | --- |
| `matched_name` | string | Full name of the matched study |
| `values` | object | All data-window values for the current bar |
| `primary_value` | number | First numeric value, rounded to 2 dp |
| `primary_key` | string | Name of the primary value field |
| `direction` | string | `above_zero` / `below_zero` / `at_zero` |
| `signal` | string | `bullish` / `bearish` / `neutral` / `overbought` / `oversold` |
| `near_zero` | bool | `true` if `\|value\| < 0.5` (near-crossing indicator) |

Signal rules:

- RSI / Relative Strength Index / Stochastic: overbought ≥ 70, oversold ≤ 30, neutral otherwise
- CCI: overbought ≥ 100, oversold ≤ −100, neutral otherwise
- All others: positive = bullish, negative = bearish, zero = neutral

#### CLI

```bash
tv indicator RSI
tv indicator "MACD"
tv indicator "Bollinger"
```

---

### `market_summary`

One-call full market context: symbol, timeframe, last bar OHLCV, bar-over-bar change%,
volume relative to the 20-bar average, and all active indicator values.

#### Response fields

| Field | Type | Description |
| --- | --- | --- |
| `symbol` | string | Current chart symbol |
| `timeframe` | string | Current timeframe |
| `chart_type` | string | Chart type code |
| `last_bar` | object | `{time, open, high, low, close, volume}` |
| `change` | number | Close − previous close (rounded to 2 dp) |
| `change_pct` | string | Percentage change, e.g. `"1.35%"` |
| `volume_vs_avg` | number | Last bar volume ÷ prior 20-bar average (rounded to 2 dp) |
| `indicators` | array | All active study objects with `name` and `values` |

#### CLI

```bash
tv market
```

---

### `continuous_contract_context`

Detects whether the current chart symbol is a continuous futures contract
(e.g. `NG1!`, `ES1!`, `CL2!`), parses its base symbol and roll number,
and enriches the response with description and exchange from TradingView's `symbolExt()`.

#### Response fields

| Field | Type | Description |
| --- | --- | --- |
| `symbol` | string | Full symbol including exchange prefix |
| `is_continuous` | bool | `true` if the symbol contains `!` |
| `base_symbol` | string | Root symbol (e.g. `"NG"` from `"NG1!"`) |
| `roll_number` | int | Contract roll number (1 = front month, 2 = second month, …) |
| `description` | string | Human-readable name from TradingView |
| `exchange` | string | Exchange name |
| `type` | string | Instrument type (e.g. `"futures"`) |
| `currency_code` | string | Settlement currency |
| `root_description` | string | Futures root description (if available) |
| `note` | string | Reminder that expiry/roll dates are not available via JS API |

#### CLI

```bash
tv futures-context
```

---

## JSON Contracts and Error Handling (Phase 5)

Phase 5 locks the response schemas for the six tools consumed by the HTS integration layer
and adds consistent error classification.

### Stable response contracts

#### `data_get_study_values`

```json
{
  "success": true,
  "study_count": 2,
  "studies": [
    {
      "name": "RSI",
      "entity_id": "Study_RSI_0",
      "plot_count": 1,
      "plots": [{ "name": "RSI", "current": 55.3, "values": [55.3] }]
    }
  ]
}
```

- `studies` is always `[]`, never `null`
- `entity_id` is the TradingView internal study source ID (use with `data_get_indicator`)
- `plots[0].current === plots[0].values[0]` — current bar value alias

#### `chart_get_state`

```json
{
  "success": true,
  "symbol": "BINANCE:BTCUSDT",
  "exchange": "BINANCE",
  "ticker": "BTCUSDT",
  "timeframe": "60",
  "type": "1",
  "indicators": [{ "id": "Study_RSI_0", "name": "RSI" }],
  "pane_count": 2
}
```

- `exchange` and `ticker` parsed from `symbol` — always strings (empty if no `:` prefix)
- `indicators` is the canonical field name; `studies` is kept as a backward-compat alias
- `pane_count` is the number of visible chart panes

#### `quote_get`

```json
{
  "success": true,
  "symbol": "BINANCE:BTCUSDT",
  "last": 67400.0,
  "open": 66800.0, "high": 67900.0, "low": 66500.0, "close": 67400.0,
  "volume": 12345.67,
  "bid": 0, "ask": 0,
  "change": 600.0, "change_pct": 0.9
}
```

- `bid`, `ask`, `change`, `change_pct` are **always present** — use `0` as sentinel for unavailable
- `change` = close − previous bar close; `change_pct` = percentage (not a decimal fraction)

#### `symbol_info`

- `symbol`, `exchange`, `description`, `type` are always present (empty string if not returned by TradingView)

#### `symbol_search`

- Every result always contains `symbol`, `exchange`, `description`, `type`

#### `data_get_indicator`

```json
{
  "success": true,
  "entity_id": "Study_RSI_0",
  "name": "Relative Strength Index",
  "inputs": { "length": 14, "source": "close" },
  "plots": [{ "name": "RSI", "current": 55.3, "values": [55.3] }]
}
```

- `inputs` is always a key→value **object** (not an array); oversized values are truncated
- `plots` is always an array (empty when study has no visible outputs)
- `name` is always a string (empty if metaInfo unavailable)

---

### Error classification

Every tool either returns `{ "success": true, …fields… }` or `{ "success": false, "error": "…" }`.

#### Retryable errors (transient — wait and retry)

| `error` contains | Cause | Recovery |
| --- | --- | --- |
| `"CDP"` or `"connect"` | Chrome DevTools Protocol not reachable | Run `tv launch` or start TradingView manually |
| `"no TradingView"` | Chart tab not found | Open a chart in TradingView |
| `"timeout"` | JS evaluation timed out | Retry after 5 s |
| `"websocket"` / `"WebSocket"` | WebSocket connection dropped | Auto-reconnects; retry after 3 s |

#### Permanent errors (do not retry)

| `error` contains | Cause | Recovery |
| --- | --- | --- |
| `"unknown tool"` | Tool name does not exist | Check the tool name |
| `"unmarshal"` or `"invalid"` | Bad argument type | Fix the argument |
| `"is required"` | Missing required field | Add the missing argument |

---

### Example — port 9222 in use by Chrome

```json
{
  "port":    { "reachable": false, "cdp": false, "owner": "chrome.exe", "error": "port in use by \"chrome.exe\" but not CDP" },
  "process": { "running": false },
  "install": { "found": true, "path": "...", "source": "LOCALAPPDATA" },
  "launch_cmd": "...",
  "hints": ["Port 9222 is in use by \"chrome.exe\". Close it or choose a different port."]
}
```

### Example — TradingView running without CDP

```json
{
  "port":    { "reachable": false, "cdp": false, "error": "connection refused — port not listening" },
  "process": { "running": true, "pid": 4812, "has_cdp_flag": false, "cmdline": "TradingView.exe" },
  "install": { "found": true, "path": "...", "source": "LOCALAPPDATA" },
  "launch_cmd": "...",
  "hints": [
    "TradingView.exe is running but --remote-debugging-port is not set. Restart it: tv launch --kill",
    "Or restart manually: cd \"...\" && \"TradingView.exe\" --remote-debugging-port=9222"
  ]
}
```

> `process.*` and `install.*` fields are Windows-only. On macOS/Linux they are
> omitted or empty; use `tv status` and `tv launch` on those platforms.

---

## Scripts

### Launch TradingView

| Script | Platform | Description |
| --- | --- | --- |
| `scripts/launch_tv_debug.bat` | Windows | Launch TradingView with CDP |
| `scripts/launch_tv_debug.vbs` | Windows | Silent launch (no cmd window) |
| `scripts/launch_tv_debug_mac.sh` | macOS | Launch TradingView with CDP |
| `scripts/launch_tv_debug_linux.sh` | Linux | Launch TradingView with CDP |

### Pine Script

| Script | Description |
| --- | --- |
| `scripts/pine_pull.sh` / `.bat` | Pull source from editor → `scripts/current.pine` |
| `scripts/pine_push.sh` / `.bat` | Push `scripts/current.pine` to editor + compile |

### Build and install

| Script | Description |
| --- | --- |
| `scripts/build.sh` | Build for current platform into `bin/` |
| `scripts/build.bat` | Same for Windows |
| `scripts/install.sh` | Install to `/usr/local/bin` (or `PREFIX`) |
| `scripts/install.bat` | Install to `%SystemRoot%\System32` |
| `scripts/bootstrap.sh` | Curl-pipe installer (downloads binaries from GitHub) |
| `scripts/bootstrap.ps1` | PowerShell installer for Windows |
| `scripts/configure-mcp.sh` | Configure MCP config for a given client |
| `scripts/configure-mcp.ps1` | Same for Windows |
| `scripts/package.sh` | Build release archives for all platforms |
