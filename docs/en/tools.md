# MCP Tools (78 total)

> [← Back to docs](README.md)

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
| `chart_manage_indicator` | Add / remove indicator |
| `chart_scroll_to_date` | Navigate to date |
| `chart_set_visible_range` | Set visible range |

### Symbols (2)

| Tool | Description |
| --- | --- |
| `symbol_info` | Symbol metadata |
| `symbol_search` | Search symbols (up to 15 results) |

### Data (8)

| Tool | Description |
| --- | --- |
| `quote_get` | Real-time OHLCV quote |
| `data_get_ohlcv` | Historical bars |
| `data_get_study_values` | All indicator values |
| `data_get_indicator` | Specific indicator values |
| `data_get_strategy_results` | Strategy backtest results |
| `data_get_trades` | Strategy trade list |
| `data_get_equity` | Equity curve |
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
| `capture_screenshot` | Screenshot (full / chart / strategy_tester) |

### Pine Script (12)

| Tool | Description |
| --- | --- |
| `pine_get_source` | Read source from editor |
| `pine_set_source` | Write source to editor |
| `pine_compile` | Compile (click button) |
| `pine_smart_compile` | Compile with error check |
| `pine_get_errors` | Monaco error list |
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
| `indicator_toggle_visibility` | Show / hide indicator |

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
| `ui_evaluate` | Execute JS expression |

### Layouts (2)

| Tool | Description |
| --- | --- |
| `layout_list` | List saved layouts |
| `layout_switch` | Switch layout |

### Batch (1)

| Tool | Description |
| --- | --- |
| `batch_run` | Iterate symbols × timeframes with actions |
