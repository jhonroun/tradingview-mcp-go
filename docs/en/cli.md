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
