# COMPATIBILITY_MATRIX.md — матрица совместимости Node.js → Go

## Назначение

Файл фиксирует соответствие исходной Node.js-реализации и Go-порта.
Статусы заполняются по мере реализации каждого этапа.

Источник инвентаризации: `tradingview-mcp` commit зафиксирован в `PORTING_NOTES.md`.

---

## Статусы

- `pending` — не начато
- `in-progress` — переносится
- `compatible` — совпадает с Node.js
- `partial` — частично совпадает
- `changed` — отличается, нужно обоснование в CHANGELOG.md
- `blocked` — заблокировано внешним фактором

---

## Общие компоненты

| Компонент | Node.js | Go | Статус |
|---|---|---|---|
| MCP server | `src/server.js` | `cmd/tvmcp`, `internal/mcp` | pending |
| CLI entry | `src/cli/index.js` | `cmd/tv`, `internal/cli` | pending |
| CLI router | `src/cli/router.js` | `internal/cli/router.go` | pending |
| CDP connection | `src/connection.js` | `internal/cdp` | pending |
| Runtime.evaluate | `chrome-remote-interface` | Go CDP client | pending |
| Tools registry | Node MCP SDK | Go tool registry | pending |
| Chart readiness poll | `src/wait.js` | `internal/tradingview/wait.go` | pending |
| JSON formatting | `src/tools/_format.js` | `internal/compat` | pending |
| Streaming (JSONL) | `src/core/stream.js` + CLI | Go JSONL stream | pending |
| safeString helper | `connection.js::safeString()` | `internal/cdp/helpers.go::SafeJSString()` | pending |
| requireFinite helper | `connection.js::requireFinite()` | `internal/cdp/helpers.go::RequireFinite()` | pending |

---

## MCP Tools — Health & Connection (4 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `tv_health_check` | `src/tools/health.js` | none | pending |
| `tv_discover` | `src/tools/health.js` | none | pending |
| `tv_ui_state` | `src/tools/health.js` | none | pending |
| `tv_launch` | `src/tools/health.js` | `port?`, `kill_existing?` | pending |

## MCP Tools — Chart State & Control (8 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `chart_get_state` | `src/tools/chart.js` | none | pending |
| `chart_set_symbol` | `src/tools/chart.js` | `symbol` | pending |
| `chart_set_timeframe` | `src/tools/chart.js` | `timeframe` | pending |
| `chart_set_type` | `src/tools/chart.js` | `chart_type` | pending |
| `chart_manage_indicator` | `src/tools/chart.js` | `action`, `indicator`, `entity_id?`, `inputs?` | pending |
| `chart_get_visible_range` | `src/tools/chart.js` | none | pending |
| `chart_set_visible_range` | `src/tools/chart.js` | `from`, `to` | pending |
| `chart_scroll_to_date` | `src/tools/chart.js` | `date` | pending |

## MCP Tools — Symbol Information (2 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `symbol_info` | `src/tools/chart.js` | none | pending |
| `symbol_search` | `src/tools/chart.js` | `query`, `type?` | pending |

## MCP Tools — Data Reading (12 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `quote_get` | `src/tools/data.js` | `symbol?` | pending |
| `data_get_ohlcv` | `src/tools/data.js` | `count?`, `summary?` | pending |
| `data_get_study_values` | `src/tools/data.js` | none | pending |
| `data_get_pine_lines` | `src/tools/data.js` | `study_filter?`, `verbose?` | pending |
| `data_get_pine_labels` | `src/tools/data.js` | `study_filter?`, `max_labels?`, `verbose?` | pending |
| `data_get_pine_tables` | `src/tools/data.js` | `study_filter?` | pending |
| `data_get_pine_boxes` | `src/tools/data.js` | `study_filter?`, `verbose?` | pending |
| `data_get_indicator` | `src/tools/data.js` | `entity_id` | pending |
| `data_get_strategy_results` | `src/tools/data.js` | none | pending |
| `data_get_trades` | `src/tools/data.js` | `max_trades?` | pending |
| `data_get_equity` | `src/tools/data.js` | none | pending |
| `depth_get` | `src/tools/data.js` | none | pending |

## MCP Tools — Screenshot (1 tool)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `capture_screenshot` | `src/tools/capture.js` | `region?`, `filename?`, `method?` | pending |

## MCP Tools — Pine Script (12 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `pine_get_source` | `src/tools/pine.js` | none | pending |
| `pine_set_source` | `src/tools/pine.js` | `source` | pending |
| `pine_compile` | `src/tools/pine.js` | none | pending |
| `pine_smart_compile` | `src/tools/pine.js` | none | pending |
| `pine_get_errors` | `src/tools/pine.js` | none | pending |
| `pine_get_console` | `src/tools/pine.js` | none | pending |
| `pine_save` | `src/tools/pine.js` | none | pending |
| `pine_new` | `src/tools/pine.js` | `type` (indicator/strategy/library) | pending |
| `pine_open` | `src/tools/pine.js` | `name` | pending |
| `pine_list_scripts` | `src/tools/pine.js` | none | pending |
| `pine_analyze` | `src/tools/pine.js` | `source` | pending |
| `pine_check` | `src/tools/pine.js` | `source` | pending |

## MCP Tools — Drawing (5 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `draw_shape` | `src/tools/drawing.js` | `shape`, `point`, `point2?`, `overrides?`, `text?` | pending |
| `draw_list` | `src/tools/drawing.js` | none | pending |
| `draw_clear` | `src/tools/drawing.js` | none | pending |
| `draw_remove_one` | `src/tools/drawing.js` | `entity_id` | pending |
| `draw_get_properties` | `src/tools/drawing.js` | `entity_id` | pending |

## MCP Tools — Alerts (3 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `alert_create` | `src/tools/alerts.js` | `condition`, `price`, `message?` | pending |
| `alert_list` | `src/tools/alerts.js` | none | pending |
| `alert_delete` | `src/tools/alerts.js` | `delete_all?` | pending |

## MCP Tools — Batch (1 tool)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `batch_run` | `src/tools/batch.js` | `symbols`, `timeframes?`, `action`, `delay_ms?`, `ohlcv_count?` | pending |

## MCP Tools — Replay (6 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `replay_start` | `src/tools/replay.js` | `date?` | pending |
| `replay_step` | `src/tools/replay.js` | none | pending |
| `replay_autoplay` | `src/tools/replay.js` | `speed?` | pending |
| `replay_stop` | `src/tools/replay.js` | none | pending |
| `replay_trade` | `src/tools/replay.js` | `action` (buy/sell/close) | pending |
| `replay_status` | `src/tools/replay.js` | none | pending |

## MCP Tools — Indicator Control (2 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `indicator_set_inputs` | `src/tools/indicators.js` | `entity_id`, `inputs` (JSON) | pending |
| `indicator_toggle_visibility` | `src/tools/indicators.js` | `entity_id`, `visible` | pending |

## MCP Tools — Watchlist (2 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `watchlist_get` | `src/tools/watchlist.js` | none | pending |
| `watchlist_add` | `src/tools/watchlist.js` | `symbol` | pending |

## MCP Tools — UI Automation (12 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `ui_click` | `src/tools/ui.js` | `by` (aria-label/data-name/text/class-contains), `value` | pending |
| `ui_open_panel` | `src/tools/ui.js` | `panel`, `action` (open/close/toggle) | pending |
| `ui_fullscreen` | `src/tools/ui.js` | none | pending |
| `layout_list` | `src/tools/ui.js` | none | pending |
| `layout_switch` | `src/tools/ui.js` | `name` | pending |
| `ui_keyboard` | `src/tools/ui.js` | `key`, `modifiers?` | pending |
| `ui_type_text` | `src/tools/ui.js` | `text` | pending |
| `ui_hover` | `src/tools/ui.js` | `by`, `value` | pending |
| `ui_scroll` | `src/tools/ui.js` | `direction`, `amount?` | pending |
| `ui_mouse_click` | `src/tools/ui.js` | `x`, `y`, `button?`, `double_click?` | pending |
| `ui_find_element` | `src/tools/ui.js` | `query`, `strategy?` | pending |
| `ui_evaluate` | `src/tools/ui.js` | `expression` | pending |

## MCP Tools — Pane Management (4 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `pane_list` | `src/tools/pane.js` | none | pending |
| `pane_set_layout` | `src/tools/pane.js` | `layout` (s/2h/2v/2x2/…) | pending |
| `pane_focus` | `src/tools/pane.js` | `index` | pending |
| `pane_set_symbol` | `src/tools/pane.js` | `index`, `symbol` | pending |

## MCP Tools — Tab Management (4 tools)

| Tool Name | Source file | Input params | Go status |
|---|---|---|---|
| `tab_list` | `src/tools/tab.js` | none | pending |
| `tab_new` | `src/tools/tab.js` | none | pending |
| `tab_close` | `src/tools/tab.js` | none | pending |
| `tab_switch` | `src/tools/tab.js` | `index` | pending |

---

## Итоговый счёт MCP tools

| Группа | Кол-во | Go статус |
|---|---|---|
| Health & Connection | 4 | pending |
| Chart State & Control | 8 | pending |
| Symbol | 2 | pending |
| Data Reading | 12 | pending |
| Screenshot | 1 | pending |
| Pine Script | 12 | pending |
| Drawing | 5 | pending |
| Alerts | 3 | pending |
| Batch | 1 | pending |
| Replay | 6 | pending |
| Indicator Control | 2 | pending |
| Watchlist | 2 | pending |
| UI Automation | 12 | pending |
| Pane Management | 4 | pending |
| Tab Management | 4 | pending |
| **Итого** | **78** | **0 / 78** |

---

## CLI Commands

| Command | Node.js source | Go status | Notes |
|---|---|---|---|
| `tv status` | `src/cli/commands/health.js` | pending | первая команда |
| `tv launch` | `src/cli/commands/health.js` | pending | platform-specific |
| `tv discover` | `src/cli/commands/health.js` | pending | |
| `tv ui-state` | `src/cli/commands/health.js` | pending | |
| `tv state` | `src/cli/commands/chart.js` | pending | |
| `tv symbol [SYM]` | `src/cli/commands/chart.js` | pending | get or set |
| `tv timeframe [TF]` | `src/cli/commands/chart.js` | pending | get or set |
| `tv type [TYPE]` | `src/cli/commands/chart.js` | pending | get or set |
| `tv info` | `src/cli/commands/chart.js` | pending | symbol metadata |
| `tv search QUERY` | `src/cli/commands/chart.js` | pending | |
| `tv range` | `src/cli/commands/chart.js` | pending | --from, --to |
| `tv scroll DATE` | `src/cli/commands/chart.js` | pending | |
| `tv quote [SYM]` | `src/cli/commands/data.js` | pending | |
| `tv ohlcv` | `src/cli/commands/data.js` | pending | -n, --summary |
| `tv values` | `src/cli/commands/data.js` | pending | |
| `tv data lines` | `src/cli/commands/data.js` | pending | |
| `tv data labels` | `src/cli/commands/data.js` | pending | |
| `tv data tables` | `src/cli/commands/data.js` | pending | |
| `tv data boxes` | `src/cli/commands/data.js` | pending | |
| `tv data strategy` | `src/cli/commands/data.js` | pending | |
| `tv data trades` | `src/cli/commands/data.js` | pending | -n/--max |
| `tv data equity` | `src/cli/commands/data.js` | pending | |
| `tv data depth` | `src/cli/commands/data.js` | pending | |
| `tv data indicator ENTITY_ID` | `src/cli/commands/data.js` | pending | |
| `tv pine get` | `src/cli/commands/pine.js` | pending | |
| `tv pine set` | `src/cli/commands/pine.js` | pending | stdin or --file |
| `tv pine compile` | `src/cli/commands/pine.js` | pending | smart compile |
| `tv pine raw-compile` | `src/cli/commands/pine.js` | pending | click button only |
| `tv pine analyze` | `src/cli/commands/pine.js` | pending | offline, -f/--file |
| `tv pine check` | `src/cli/commands/pine.js` | pending | server-side, -f/--file |
| `tv pine save` | `src/cli/commands/pine.js` | pending | |
| `tv pine new [TYPE]` | `src/cli/commands/pine.js` | pending | indicator/strategy/library |
| `tv pine open NAME` | `src/cli/commands/pine.js` | pending | |
| `tv pine list` | `src/cli/commands/pine.js` | pending | |
| `tv pine errors` | `src/cli/commands/pine.js` | pending | |
| `tv pine console` | `src/cli/commands/pine.js` | pending | |
| `tv screenshot` | `src/cli/commands/capture.js` | pending | -r/--region, -o/--output |
| `tv replay start` | `src/cli/commands/replay.js` | pending | -d/--date |
| `tv replay step` | `src/cli/commands/replay.js` | pending | |
| `tv replay stop` | `src/cli/commands/replay.js` | pending | |
| `tv replay status` | `src/cli/commands/replay.js` | pending | |
| `tv replay autoplay` | `src/cli/commands/replay.js` | pending | -s/--speed |
| `tv replay trade ACTION` | `src/cli/commands/replay.js` | pending | buy/sell/close |
| `tv draw shape` | `src/cli/commands/drawing.js` | pending | -t/--type, -p/--price, ... |
| `tv draw list` | `src/cli/commands/drawing.js` | pending | |
| `tv draw get ENTITY_ID` | `src/cli/commands/drawing.js` | pending | |
| `tv draw remove ENTITY_ID` | `src/cli/commands/drawing.js` | pending | |
| `tv draw clear` | `src/cli/commands/drawing.js` | pending | |
| `tv alert list` | `src/cli/commands/alerts.js` | pending | |
| `tv alert create` | `src/cli/commands/alerts.js` | pending | -p/--price, -c/--condition, -m/--message |
| `tv alert delete` | `src/cli/commands/alerts.js` | pending | --all |
| `tv watchlist get` | `src/cli/commands/watchlist.js` | pending | |
| `tv watchlist add SYM` | `src/cli/commands/watchlist.js` | pending | |
| `tv layout list` | `src/cli/commands/ui.js` | pending | |
| `tv layout switch NAME` | `src/cli/commands/ui.js` | pending | |
| `tv indicator add NAME` | `src/cli/commands/indicator.js` | pending | -i/--inputs |
| `tv indicator remove ENTITY_ID` | `src/cli/commands/indicator.js` | pending | |
| `tv indicator toggle ENTITY_ID` | `src/cli/commands/indicator.js` | pending | --visible/--hidden |
| `tv indicator set ENTITY_ID` | `src/cli/commands/indicator.js` | pending | -i/--inputs |
| `tv indicator get ENTITY_ID` | `src/cli/commands/indicator.js` | pending | |
| `tv ui click` | `src/cli/commands/ui.js` | pending | |
| `tv ui keyboard KEY` | `src/cli/commands/ui.js` | pending | --ctrl/--shift/--alt/--meta |
| `tv ui hover` | `src/cli/commands/ui.js` | pending | |
| `tv ui scroll [DIR]` | `src/cli/commands/ui.js` | pending | |
| `tv ui find QUERY` | `src/cli/commands/ui.js` | pending | -s/--strategy |
| `tv ui eval EXPR` | `src/cli/commands/ui.js` | pending | |
| `tv ui type TEXT` | `src/cli/commands/ui.js` | pending | |
| `tv ui panel PANEL ACTION` | `src/cli/commands/ui.js` | pending | open/close/toggle |
| `tv ui fullscreen` | `src/cli/commands/ui.js` | pending | |
| `tv ui mouse X Y` | `src/cli/commands/ui.js` | pending | --right/--double |
| `tv pane list` | `src/cli/commands/pane.js` | pending | |
| `tv pane layout LAYOUT` | `src/cli/commands/pane.js` | pending | |
| `tv pane focus INDEX` | `src/cli/commands/pane.js` | pending | |
| `tv pane symbol INDEX SYM` | `src/cli/commands/pane.js` | pending | |
| `tv tab list` | `src/cli/commands/tab.js` | pending | |
| `tv tab new` | `src/cli/commands/tab.js` | pending | |
| `tv tab close` | `src/cli/commands/tab.js` | pending | |
| `tv tab switch INDEX` | `src/cli/commands/tab.js` | pending | |
| `tv stream quote` | `src/cli/commands/stream.js` | pending | -i/--interval, JSONL |
| `tv stream bars` | `src/cli/commands/stream.js` | pending | -i/--interval, JSONL |
| `tv stream values` | `src/cli/commands/stream.js` | pending | -i/--interval, JSONL |
| `tv stream lines` | `src/cli/commands/stream.js` | pending | -f/--filter, -i/--interval, JSONL |
| `tv stream labels` | `src/cli/commands/stream.js` | pending | -f/--filter, -i/--interval, JSONL |
| `tv stream tables` | `src/cli/commands/stream.js` | pending | -f/--filter, -i/--interval, JSONL |
| `tv stream all` | `src/cli/commands/stream.js` | pending | -i/--interval, JSONL |

---

## CDP Calls

| CDP Method | Используется в | Параметры | Go status |
|---|---|---|---|
| `Runtime.enable` | `src/connection.js` | — | pending |
| `Page.enable` | `src/connection.js` | — | pending |
| `DOM.enable` | `src/connection.js` | — | pending |
| `Runtime.evaluate` | все core-модули | `expression`, `returnByValue`, `awaitPromise` | pending |
| `Input.dispatchKeyEvent` | `src/core/ui.js`, `src/core/pine.js`, `src/core/tab.js`, `src/core/watchlist.js`, `src/core/alerts.js` | `type`, `key`, `code`, `modifiers`, `windowsVirtualKeyCode` | pending |
| `Input.insertText` | `src/core/ui.js`, `src/core/watchlist.js` | `text` | pending |
| `Input.dispatchMouseEvent` | `src/core/ui.js` | `type`, `x`, `y`, `button`, `buttons`, `clickCount`, `deltaX`, `deltaY` | pending |
| `Page.captureScreenshot` | `src/core/capture.js`, `src/core/batch.js` | `format`, `quality`, `clip` | pending |

---

## Scripts

| Script | Платформа | Назначение | Go equivalent | Статус |
|---|---|---|---|---|
| `scripts/launch_tv_debug.bat` | Windows | Запуск TradingView с `--remote-debugging-port=9222` | `internal/launch/windows.go` | pending |
| `scripts/launch_tv_debug.vbs` | Windows (VBS) | Альтернативный запуск TradingView | `internal/launch/windows.go` | pending |
| `scripts/launch_tv_debug_linux.sh` | Linux | Запуск TradingView с CDP | `internal/launch/linux.go` | pending |
| `scripts/launch_tv_debug_mac.sh` | macOS | Запуск TradingView с CDP | `internal/launch/macos.go` | pending |
| `scripts/pine_pull.js` | Cross-platform | Извлечь Pine source из редактора → `scripts/current.pine` | `tv pine get > file` или Go util | pending |
| `scripts/pine_push.js` | Cross-platform | Вставить Pine source из файла и скомпилировать | `tv pine set < file && tv pine compile` | pending |

---

## Skills / Workflows

| Skill | Назначение | CLI tools used | Go status |
|---|---|---|---|
| `skills/chart-analysis` | Анализ графика: индикаторы, уровни, скриншот, отчёт | chart, data, draw, capture | pending |
| `skills/multi-symbol-scan` | Скан нескольких символов, сравнение | batch_run, watchlist | pending |
| `skills/pine-develop` | Цикл разработки Pine Script: write → compile → fix → verify | pine_*, pine_pull, pine_push | pending |
| `skills/replay-practice` | Торговля в режиме replay, пошаговый анализ | replay_*, chart, capture | pending |
| `skills/strategy-report` | Генерация отчёта по стратегии: метрики, сделки, equity | data_get_strategy_results, data_get_trades, data_get_equity, capture | pending |

---

## Windows discovery / launcher (P2W)

| Компонент | Node.js | Go | Статус | Примечание |
|---|---|---|---|---|
| Standard Windows install | launch script / auto-detect | `internal/discovery/windows.go` | pending | `%LOCALAPPDATA%`, `%PROGRAMFILES%` |
| Microsoft Store install | часто проблемно | `Get-AppxPackage` + WindowsApps fallback | pending | `C:\Program Files\WindowsApps\TradingView.Desktop_*` |
| Manual override | частично | env/flag/config | pending | `TRADINGVIEW_PATH`, `--tv-path` |
| Diagnostics | частично | `tv doctor windows` | pending | troubleshooting tool |

---

## TradingView API paths (из connection.js)

| JS path | Назначение |
|---|---|
| `window.TradingViewApi._activeChartWidgetWV.value()` | Активный chart widget |
| `window.TradingViewApi._chartWidgetCollection` | Коллекция chart widgets (все panes) |
| `window.TradingViewApi._replayApi` | Replay API |
| `window.TradingViewApi._alertService` | Alert service |
| `window.TradingViewApi._activeChartWidgetWV.value()._chartWidget.model().mainSeries().bars()` | Bars главной серии |
