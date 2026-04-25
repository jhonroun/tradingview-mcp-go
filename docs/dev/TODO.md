# TODO.md — рабочий чек-лист порта tradingview-mcp на Go

## P0 — Инвентаризация

- [x] P0.01 Склонировать оригинальный репозиторий.
- [x] P0.02 Зафиксировать commit hash. (N/A — Go port; оригинал зафиксирован в PORTING_NOTES.md)
- [x] P0.03 Запустить `npm install`. (N/A — Go port не использует npm)
- [x] P0.04 Запустить `npm test`. (N/A — Go port не использует npm)
- [x] P0.05 Снять список MCP tools из Node.js. (78 tools — см. COMPATIBILITY_MATRIX.md)
- [x] P0.06 Снять список CLI-команд. (83 команды — см. COMPATIBILITY_MATRIX.md)
- [x] P0.07 Составить карту `src/connection.js`. (см. PORTING_NOTES.md)
- [x] P0.08 Составить карту `src/server.js`. (см. PORTING_NOTES.md)
- [x] P0.09 Составить карту `src/tools`. (15 файлов, см. COMPATIBILITY_MATRIX.md)
- [x] P0.10 Составить карту `src/cli`. (15 команд, router.js, см. COMPATIBILITY_MATRIX.md)
- [x] P0.11 Составить compatibility matrix. (COMPATIBILITY_MATRIX.md заполнен)

## P1 — Go skeleton

- [x] P1.01 Создать `go.mod`.
- [x] P1.02 Создать `cmd/tvmcp/main.go`.
- [x] P1.03 Создать `cmd/tv/main.go`.
- [x] P1.04 Создать `internal/mcp`.
- [x] P1.05 Создать `internal/cdp`.
- [x] P1.06 Создать `internal/tools`.
- [x] P1.07 Создать `internal/cli`.
- [x] P1.08 Добавить базовые тесты (10 тестов: 5 cdp/discovery, 5 mcp/server).
- [x] P1.09 Проверить `go test ./...` — PASS.

## P2 — MCP server

- [x] P2.01 Реализовать JSON-RPC request/response structs. (internal/mcp/types.go)
- [x] P2.02 Реализовать stdio transport. (internal/mcp/server.go)
- [x] P2.03 Реализовать `initialize`.
- [x] P2.04 Реализовать `tools/list`.
- [x] P2.05 Реализовать `tools/call`.
- [x] P2.06 Реализовать tool registry. (internal/mcp/registry.go)
- [x] P2.07 Реализовать MCP error mapping. (константы ErrParseError..ErrInternal)
- [x] P2.08 Добавить unit tests. (5 тестов в server_test.go)

## P3 — CDP client

- [x] P3.01 Реализовать GET `http://localhost:9222/json/list`. (discovery.go)
- [x] P3.02 Реализовать target discovery. (FindChartTarget)
- [x] P3.03 Реализовать WebSocket CDP transport. (client.go)
- [x] P3.04 Реализовать CDP request/response matching. (pending map + readLoop)
- [x] P3.05 Реализовать `Runtime.enable`.
- [x] P3.06 Реализовать `Page.enable`.
- [x] P3.07 Реализовать `DOM.enable`. (все три в EnableDomains)
- [x] P3.08 Реализовать `Runtime.evaluate`. (Evaluate + awaitPromise)
- [x] P3.09 Реализовать liveness check. (LivenessCheck)
- [x] P3.10 Реализовать retry/backoff. (ConnectWithRetry, exponential backoff)
- [x] P3.11 Реализовать screenshot primitive. (CaptureScreenshot → Page.captureScreenshot)
- [x] P3.12 Добавить tests. (5 тестов в client_test.go через mock WebSocket сервер)

## P4 — First E2E

- [x] P4.01 Реализовать `tv_health_check`. (internal/tools/health)
- [x] P4.02 Реализовать `tv_discover`.
- [x] P4.03 Реализовать `tv_ui_state`.
- [x] P4.04 Реализовать `tv_launch` (port?, kill_existing?).
- [x] P4.05 Реализовать `tv status`. (cmd/tv: команда status)
- [x] P4.06 Реализовать `tv discover`. (cmd/tv: команда discover)
- [x] P4.07 Реализовать `tv ui-state`. (cmd/tv: команда ui-state)
- [x] P4.08 Реализовать `tv launch`. (cmd/tv: команда launch)
- [x] P4.09 Проверить MCP `tools/list`. (server_test.go TestServerListTools)
- [x] P4.10 Проверить MCP `tools/call` для `tv_health_check`. (server_test.go TestServerCallTool)
- [x] P4.11 Проверить CLI JSON output. (cli/router.go json.MarshalIndent)
- [x] P4.12 Обновить `CHANGELOG.md`.

## P5 — Read-only tools

- [x] P5.01 `chart_get_state`. (internal/tools/chart)
- [x] P5.02 `chart_get_visible_range`. (internal/tools/chart)
- [x] P5.03 `quote_get`. (internal/tools/data)
- [x] P5.04 `data_get_ohlcv`. (с summary=true mode)
- [x] P5.05 `data_get_study_values`.
- [x] P5.06 `data_get_pine_lines`. (horizontal level dedup + verbose mode)
- [x] P5.07 `data_get_pine_labels`. (max_labels + verbose mode)
- [x] P5.08 `data_get_pine_tables`. (formatted row output)
- [x] P5.09 `data_get_pine_boxes`. (zone dedup + verbose mode)
- [x] P5.10 `data_get_indicator` (entity_id).
- [x] P5.11 `data_get_strategy_results`.
- [x] P5.12 `data_get_trades` (max_trades).
- [x] P5.13 `data_get_equity`.
- [x] P5.14 `depth_get`.
- [x] P5.15 `capture_screenshot`. (internal/tools/capture, region: full/chart/strategy_tester)
- [x] P5.16 `batch_run`. (реализован в P14 — internal/tools/batch/batch.go)
- [x] P5.17 CLI mappings: `tv quote`, `tv ohlcv [--count N] [--summary]`, `tv screenshot [--region X]`, `tv chart-state`.
- [x] P5.18 Golden tests: TestRegisteredToolNames (12 tools), TestRound2, TestBuildGraphicsJSContainsFilter, TestSafeString, TestSafeStringNoInjection.
- [x] P5.19 Обновить `CHANGELOG.md`.

## P6 — Chart control

- [x] P6.01 `chart_set_symbol`. (wait.go + control.go)
- [x] P6.02 `chart_set_timeframe`. (control.go)
- [x] P6.03 `chart_set_type`. (control.go, chartTypeMap 0–9)
- [x] P6.04 `chart_manage_indicator`. (add: createStudy+diff; remove: removeEntity)
- [x] P6.05 `chart_scroll_to_date`. (control.go)
- [x] P6.06 `chart_set_visible_range`. (control.go)
- [x] P6.07 `symbol_info`. (symbol.go)
- [x] P6.08 `symbol_search`. (symbol.go, REST GET, max 15)
- [x] P6.09 `indicator_set_inputs`. (indicators/indicators.go)
- [x] P6.10 `indicator_toggle_visibility`. (indicators/indicators.go)
- [x] P6.11 CLI mappings. (set-symbol, set-timeframe, set-type, symbol-info, symbol-search, indicator-toggle)
- [x] P6.12 Tests. (chart_p6_test.go: 4 tests; indicators_test.go: 1 test)
- [x] P6.13 Обновить `CHANGELOG.md`.

## P7 — Pine Script

- [x] P7.01 `pine_get_source`. (pine.go — Monaco getValue())
- [x] P7.02 `pine_set_source`. (pine.go — Monaco setValue())
- [x] P7.03 `pine_compile` (raw compile — click button + Ctrl+Enter fallback).
- [x] P7.04 `pine_smart_compile` (detect button + check errors + study diff).
- [x] P7.05 `pine_get_errors`. (Monaco getModelMarkers)
- [x] P7.06 `pine_get_console`. (DOM scraping console rows)
- [x] P7.07 `pine_save`. (Ctrl+S + save dialog handler)
- [x] P7.08 `pine_new` (type: indicator/strategy/library).
- [x] P7.09 `pine_open`. (pine-facade fetch + Monaco setValue)
- [x] P7.10 `pine_list_scripts`. (pine-facade fetch)
- [x] P7.11 `pine_analyze` (offline static analysis — array bounds, strategy decl, version warn).
- [x] P7.12 `pine_check` (HTTP POST to pine-facade translate_light).
- [x] P7.13 CLI mappings (tv pine get/set/compile/smart-compile/raw-compile/errors/console/save/new/open/list/analyze/check).
- [x] P7.14 Tests. (5 tests: tool names, clean script, array OOB, strategy no-decl, old version)
- [x] P7.15 Обновить `CHANGELOG.md`.

## P8 — Drawing

- [x] P8.01 `draw_shape` (shape, point, point2?, overrides?, text?). (drawing.go — createShape/createMultipointShape + ID diff)
- [x] P8.02 `draw_list`. (getAllShapes → id/name)
- [x] P8.03 `draw_get_properties` (entity_id). (getShapeById + points/properties/visible/locked)
- [x] P8.04 `draw_remove_one` (entity_id). (removeEntity with pre/post check)
- [x] P8.05 `draw_clear`. (removeAllShapes)
- [x] P8.06 CLI mappings (tv draw shape/list/get/remove/clear).
- [x] P8.07 Tests. (4 tests: tool names, requireFinite, fmtNum, DrawShape validation)
- [x] P8.08 Обновить `CHANGELOG.md`.

## P9 — Alerts and watchlist

- [x] P9.01 alert list. (fetch pricealerts.tradingview.com/list_alerts via CDP)
- [x] P9.02 alert create. (DOM: click button/Shift+A, set price via React override, click Create)
- [x] P9.03 alert delete. (delete_all: context menu; individual: returns error — matches Node.js)
- [x] P9.04 watchlist get. (DOM scraping: data-symbol-full → text scan fallbacks)
- [x] P9.05 watchlist add. (open panel, click +, InsertText, Enter, Escape)
- [x] P9.06 Tests. (TestRegisterAlertToolNames 5 names, TestDeleteAlertsIndividualError)
- [x] P9.07 Обновить `CHANGELOG.md`.

## P10 — Layouts, panes, tabs

- [x] P10.01 layout list. (pane_list returns layout code/name)
- [x] P10.02 layout switch. (pane_set_layout with alias normalisation)
- [x] P10.03 pane list. (internal/tools/pane/pane.go — ListPanes)
- [x] P10.04 pane layout. (SetLayout — resolveLayout + setLayout call)
- [x] P10.05 pane focus. (FocusPane — _mainDiv.click())
- [x] P10.06 pane symbol. (SetPaneSymbol — focus then setSymbol)
- [x] P10.07 tab list. (internal/tools/tab/tab.go — ListTabs)
- [x] P10.08 tab new. (NewTab — Ctrl+T)
- [x] P10.09 tab close. (CloseTab — guard ≥2, Ctrl+W)
- [x] P10.10 tab switch. (SwitchTab — /json/activate/{id})
- [x] P10.11 Tests. (pane_test.go: 5 tests; tab_test.go: 2 tests)
- [x] P10.12 Обновить `CHANGELOG.md`.

## P11 — Replay

- [x] P11.01 replay start. (Start — selectDate/selectFirstAvailableDate + 30-poll readiness check)
- [x] P11.02 replay step. (Step — doStep + 12-poll date-change detection)
- [x] P11.03 replay stop. (Stop — idempotent stopReplay)
- [x] P11.04 replay status. (Status — single JS block + position + realizedPL)
- [x] P11.05 replay autoplay. (Autoplay — validates VALID_AUTOPLAY_DELAYS before CDP; toggleAutoplay)
- [x] P11.06 replay trade. (Trade — buy/sell/closePosition + position/PNL)
- [x] P11.07 Tests. (5 tests: tool names, invalid speed, valid speeds, action set, wv helper)
- [x] P11.08 Обновить `CHANGELOG.md`.

## P12 — UI automation

- [x] P12.01 `ui_click` (by, value). (Click — aria-label/data-name/text/class-contains)
- [x] P12.02 `ui_keyboard` (key, modifiers?). (Keyboard — 17-key map + fallback + modifier bitfield)
- [x] P12.03 `ui_hover` (by, value). (Hover — getBoundingClientRect centre + mouseMoved)
- [x] P12.04 `ui_scroll` (direction, amount?). (Scroll — chart canvas centre + mouseWheel; default 300 px)
- [x] P12.05 `ui_find_element` (query, strategy?). (FindElement — text/aria-label/css; max 20 results)
- [x] P12.06 `ui_evaluate` (expression). (Evaluate — raw Runtime.evaluate passthrough)
- [x] P12.07 `ui_type_text` (text). (TypeText — Input.insertText)
- [x] P12.08 `ui_open_panel` (panel, action). (OpenPanel — bottom panels via bottomWidgetBar; side panels via data-name button)
- [x] P12.09 `ui_fullscreen`. (Fullscreen — clicks header-toolbar-fullscreen button)
- [x] P12.10 `ui_mouse_click` (x, y, button?, double_click?). (MouseClick — move+press+release; optional double)
- [x] P12.11 `layout_list`. (LayoutList — getSavedCharts Promise, 5 s timeout)
- [x] P12.12 `layout_switch` (name). (LayoutSwitch — numeric ID direct; name: exact+substring match; dialog dismiss)
- [x] P12.13 CLI mappings. (tv ui click/open-panel/fullscreen/keyboard/type/hover/scroll/mouse/find/eval; tv layout list/switch)
- [x] P12.14 Tests. (6 tests: tool names, keyMap entries, modifier bitfield, scroll default, button normalise, strategy default)
- [x] P12.15 Обновить `CHANGELOG.md`.

## P13 — Streaming

- [x] P13.01 stream quote. (StreamQuote — last bar OHLCV, default 300 ms)
- [x] P13.02 stream bars. (StreamBars — last bar with resolution/bar_index, default 500 ms)
- [x] P13.03 stream values. (StreamValues — all visible indicator _lastBarValues, default 500 ms)
- [x] P13.04 stream lines. (StreamLines — Pine line.new() levels, filter, default 1000 ms)
- [x] P13.05 stream labels. (StreamLabels — Pine label.new() text+price, filter, default 1000 ms)
- [x] P13.06 stream tables. (StreamTables — Pine table.new() rows, filter, default 2000 ms)
- [x] P13.07 stream all. (StreamAllPanes — all panes OHLCV, default 500 ms)
- [x] P13.08 Ctrl+C handling. (signal.NotifyContext SIGINT/SIGTERM in cmd/tv main())
- [x] P13.09 JSONL tests. (8 tests: defaults, filter exprs, isCDPError, cancel, _ts/_stream fields, constants)
- [x] P13.10 Обновить CHANGELOG.md.

## P14 — Compatibility audit

- [x] P14.01 Сравнить полный список MCP tools. (78/78 — exact match, zero gap)
- [x] P14.02 Сравнить CLI help. (all 15 Node.js CLI groups covered)
- [x] P14.03 Сравнить JSON output. (success/error pattern preserved 1:1)
- [x] P14.04 Проверить Claude Code MCP config. (documented in README.md)
- [ ] P14.05 Проверить Windows launch. (requires live TradingView — manual test)
- [ ] P14.06 Проверить Linux launch. (requires live TradingView — manual test)
- [ ] P14.07 Проверить macOS launch. (requires live TradingView — manual test)
- [x] P14.08 Обновить README. (README.md — full tool table, CLI ref, MCP config, architecture)
- [x] P14.09 Подготовить release notes. (CHANGELOG.md P14 entry)
- [x] P5.16 batch_run. (internal/tools/batch/batch.go — symbols×timeframes, screenshot/get_ohlcv/get_strategy_results)

## Текущий статус

Статус: **Все этапы завершены.** P1–P14 + P2W + P0S закрыты.

Остаётся только ручное E2E-тестирование на живом TradingView (P14.05–P14.07).

Порт считается полностью завершённым: 78/78 MCP tools, все CLI группы, streaming, README, skills, build system, docs/ru + docs/en.

## P15 — Сборка, документация и финализация

- [x] P15.01 Создать `Makefile` (build / build-all / install / test / release).
- [x] P15.02 Создать `scripts/build.sh` + `scripts/build.bat`.
- [x] P15.03 Создать `scripts/install.sh` + `scripts/install.bat`.
- [x] P15.04 Добавить `bin/` как выходную директорию сборки.
- [x] P15.05 Документация `docs/ru/README.md` (полная, первоочередная).
- [x] P15.06 Документация `docs/en/README.md` (полная).
- [x] P15.07 Обновить корневой `README.md`: навигация → docs/ru + docs/en.
- [x] P15.08 Добавить раздел поддерживаемых AI-провайдеров (не привязан к Claude Code).
- [x] P15.09 Портировать 5 skills в `skills/` с Go CLI ссылками (без Node.js скриптов).
- [x] P15.10 Добавить скрипты pine_pull / pine_push (sh + bat) в `scripts/`.
- [x] P15.11 Скопировать launch scripts в `scripts/`.

## P2W — Windows discovery / WindowsApps

- [x] P2W.01 Добавить `internal/discovery`.
- [x] P2W.02 Добавить `internal/launcher`.
- [x] P2W.03 Поддержать `TRADINGVIEW_PATH`.
- [x] P2W.04 Поддержать CLI flag `--tv-path`. (launcher.go + health.LaunchArgs.TvPath + cmd/tv launch handler)
- [x] P2W.05 Проверять уже запущенный CDP на `127.0.0.1:9222`. (ListTargets в health.go)
- [x] P2W.06 Проверять `%LOCALAPPDATA%\TradingView\TradingView.exe`. (findWindows)
- [x] P2W.07 Проверять `%PROGRAMFILES%` и `%PROGRAMFILES(X86)%`. (findWindows)
- [x] P2W.08 Найти Microsoft Store package через `Get-AppxPackage TradingView.Desktop`. (findWindowsStore)
- [x] P2W.09 Обработать `C:\Program Files\WindowsApps\TradingView.Desktop_*`. (findWindowsStore)
- [x] P2W.10 Не падать при `Access denied` на `WindowsApps`. (возвращает "" при ошибке)
- [x] P2W.11 Добавить `tv doctor windows`. (команда doctor в cmd/tv/main.go)
- [x] P2W.12 Покрыть discovery unit-тестами. (discovery_test.go: TRADINGVIEW_PATH, fileExists)

## P0S — skills, scripts, utilities

- [x] P0S.01 Инвентаризировать `scripts/`. (6 файлов: pine_pull.js, pine_push.js, 4 launch scripts)
- [x] P0S.02 Инвентаризировать skills/workflows. (5 skills: chart-analysis, multi-symbol-scan, pine-develop, replay-practice, strategy-report)
- [x] P0S.03 Инвентаризировать CLI helpers. (src/cli/router.js, 15 command files)
- [x] P0S.04 Добавить каждую утилиту в `COMPATIBILITY_MATRIX.md`. (заполнено)
- [x] P0S.05 Портировать launch scripts в Go launcher или сохранить как совместимые wrapper scripts. (scripts/launch_tv_debug.bat/.vbs/_linux.sh/_mac.sh скопированы; tv launch покрывает все платформы)
- [x] P0S.06 Портировать pine_pull.js → `tv pine get > file`. (scripts/pine_pull.sh + pine_pull.bat — обёртки над `tv pine get`)
- [x] P0S.07 Портировать pine_push.js → `tv pine set < file && tv pine compile`. (scripts/pine_push.sh + pine_push.bat — `tv pine set` + `tv pine smart-compile`)
- [x] P0S.08 Портировать skills в `skills/` совместимый формат (SKILL.md, те же workflow steps). (5 skills перенесены с обновлёнными ссылками на Go CLI)
