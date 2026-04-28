# MCP Tools (85 current Go tools)

> [← Back to docs](README.md)

The original Node.js parity baseline was 78 tools. The current Go registry has 85 tools: the 78 parity tools plus Go extensions for study history, orders, Pine restore safety, and aggregate LLM context helpers.

---

### Health & Connection (4)

| Tool | Description |
| --- | --- |
| `tv_health_check` | Check CDP connection |
| `tv_discover` | Report available TradingView API paths |
| `tv_ui_state` | Get UI panel state |
| `tv_launch` | Launch TradingView with CDP |

### Chart State (2)

| Tool | Description |
| --- | --- |
| `chart_get_state` | Symbol, timeframe, type, indicators |
| `chart_get_visible_range` | Current visible date range |

### Chart Control (6)

| Tool | Description |
| --- | --- |
| `chart_set_symbol` | Change symbol |
| `chart_set_timeframe` | Change timeframe |
| `chart_set_type` | Chart type (Candles, HeikinAshi, Line…) |
| `chart_manage_indicator` | Add/remove indicator; supports explicit `allow_remove_any` for limit recovery |
| `chart_scroll_to_date` | Navigate to date |
| `chart_set_visible_range` | Set visible range |

### Symbols (2)

| Tool | Description |
| --- | --- |
| `symbol_info` | Symbol metadata |
| `symbol_search` | Search symbols; empty results include `status: no_results` |

### Data (10)

| Tool | Description |
| --- | --- |
| `quote_get` | Real-time OHLCV quote; unavailable bid/ask is marked with `bidAskAvailable:false` |
| `data_get_ohlcv` | Historical bars |
| `data_get_study_values` | Current numeric values for visible studies from the TradingView study model when available |
| `data_get_indicator` | Current numeric values for one study by entity ID/name |
| `data_get_indicator_history` | Loaded-bar study history from `fullRangeIterator()` |
| `data_get_strategy_results` | Strategy performance through `TradingViewApi.backtestingStrategyApi()` |
| `data_get_trades` | Strategy trades from the backtesting report |
| `data_get_orders` | Strategy filled orders from the backtesting report |
| `data_get_equity` | Equity from explicit `Strategy Equity` plot or documented fallback/status |
| `depth_get` | Order book (Level 2) |

### Pine Graphics (4)

| Tool | Description |
| --- | --- |
| `data_get_pine_lines` | Horizontal levels from `line.new()` |
| `data_get_pine_labels` | Labels from `label.new()` |
| `data_get_pine_tables` | Tables from `table.new()` |
| `data_get_pine_boxes` | Boxes from `box.new()` |

### Screenshot (1)

| Tool | Description |
| --- | --- |
| `capture_screenshot` | Screenshot (full / chart / strategy_tester); `.png` extension is normalized |

### Pine Script (13)

| Tool | Description |
| --- | --- |
| `pine_get_source` | Read source, hash, script name/type from editor |
| `pine_set_source` | Backup current source, then write source to editor |
| `pine_restore_source` | Restore a backup and verify SHA256 |
| `pine_compile` | Compile/add to chart; supports English and Russian Add-to-chart labels |
| `pine_smart_compile` | Compile with diagnostics and study-added check |
| `pine_get_errors` | Structured Monaco error list |
| `pine_get_console` | Pine console output |
| `pine_save` | Save script (Ctrl+S) |
| `pine_new` | New script (indicator/strategy/library) |
| `pine_open` | Open script by name |
| `pine_list_scripts` | List saved scripts |
| `pine_analyze` | Offline static analysis |
| `pine_check` | Check via pine-facade API |

### Drawing (5)

| Tool | Description |
| --- | --- |
| `draw_shape` | Draw a shape |
| `draw_list` | List all shapes |
| `draw_get_properties` | Shape properties |
| `draw_remove_one` | Remove a shape |
| `draw_clear` | Clear all shapes |

### Alerts (3)

| Tool | Description |
| --- | --- |
| `alert_create` | Create a price alert |
| `alert_list` | List alerts |
| `alert_delete` | Delete alerts |

### Watchlist (2)

| Tool | Description |
| --- | --- |
| `watchlist_get` | Read watchlist |
| `watchlist_add` | Add symbol to watchlist |

### Indicators (2)

| Tool | Description |
| --- | --- |
| `indicator_set_inputs` | Set indicator inputs |
| `indicator_toggle_visibility` | Show/hide indicator |

### Replay (6)

| Tool | Description |
| --- | --- |
| `replay_start` | Start replay from date |
| `replay_step` | Step forward one bar |
| `replay_stop` | Stop replay |
| `replay_status` | Status (date, position, P&L) |
| `replay_autoplay` | Auto-play |
| `replay_trade` | Trade in replay (buy/sell/close) |

### Panes (4)

| Tool | Description |
| --- | --- |
| `pane_list` | List panes and layout |
| `pane_set_layout` | Change layout |
| `pane_focus` | Focus a pane |
| `pane_set_symbol` | Set pane symbol |

### Tabs (4)

| Tool | Description |
| --- | --- |
| `tab_list` | List tabs |
| `tab_new` | Open new tab |
| `tab_close` | Close tab |
| `tab_switch` | Switch to tab |

### UI Automation (10)

| Tool | Description |
| --- | --- |
| `ui_click` | Click an element |
| `ui_open_panel` | Open/close panel |
| `ui_fullscreen` | Toggle fullscreen |
| `ui_keyboard` | Send key press |
| `ui_type_text` | Type text |
| `ui_hover` | Hover over element |
| `ui_scroll` | Scroll |
| `ui_mouse_click` | Click by coordinates |
| `ui_find_element` | Find element on page |
| `ui_evaluate` | Execute JS expression; public behavior unchanged |

### Layouts (2)

| Tool | Description |
| --- | --- |
| `layout_list` | List saved layouts |
| `layout_switch` | Switch layout |

### Batch (1)

| Tool | Description |
| --- | --- |
| `batch_run` | Iterate symbols × timeframes with actions |

### LLM/Context Helpers (4)

| Tool | Description |
| --- | --- |
| `chart_context_for_llm` | Compact chart state + price + top-N study values |
| `indicator_state` | Named indicator signal/direction summary |
| `market_summary` | OHLCV summary + volume context + active studies |
| `continuous_contract_context` | Continuous futures metadata and roll-number parsing |

## Data Reliability Policy

Values intended for trading logic must be checked for `source`, `reliability`, and `reliableForTradingLogic`.

- `tradingview_study_model`: numeric Pine runtime values from TradingView study internals; reliable but unstable internal path.
- `tradingview_backtesting_api`: Strategy Tester report data; reliable when `status: ok`, unstable internal path.
- `tradingview_strategy_plot`: explicit Pine `Strategy Equity` plot; reliable for `coverage: loaded_chart_bars` only.
- `tradingview_ui_data_window`: localized display string fallback; not reliable for trading logic.
- `derived_from_ohlcv_and_trades`: derived fallback; conditional, `reliableForTradingLogic:false` unless a caller independently guarantees complete OHLCV/trades/settings coverage, and not equivalent to native TradingView equity.

## Compatibility Probes

`tv_discover` keeps the legacy `paths` object and also returns `compatibility_probes`.
Each probe is non-mutating and includes:

- `compatible`: the internal path or method exists in this TradingView build.
- `available`: useful data is present in the current chart state.
- `status`: examples include `ok`, `no_strategy_loaded`, `needs_equity_plot`, `strategy_report_unavailable`, `unavailable`, and `error`.
- `stability`: always `unstable_internal_path` for undocumented TradingView internals.
- `reliability`: the reliability class to propagate into dependent tool responses.

Run `tv discover` after TradingView Desktop updates or when data tools begin returning unavailable statuses.

## Equity Coverage

`data_get_equity` is not a full Strategy Tester equity export. The reliable path is:

```pine
plot(strategy.equity, "Strategy Equity", display=display.data_window)
```

When the plot is present, the tool reads the loaded chart bars through the strategy source model and returns `coverage: loaded_chart_bars`. That can be the full requested range only if TradingView has actually loaded it.

Optional history-load workflow:

1. Use `chart_set_visible_range` or `chart_scroll_to_date` to move/expand the chart range.
2. Wait for TradingView to load more bars.
3. Re-run `data_get_equity` or `data_get_indicator_history`.
4. Compare `loaded_bar_count`, `data_points`, `total_data_points`, and `coverage`.
5. Keep the result marked as loaded-bars coverage, not native full backtest history.

Do not spend implementation time on "full native bar-by-bar Strategy Tester equity" unless TradingView exposes a stable report field for it.
