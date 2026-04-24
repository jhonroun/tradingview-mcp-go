# TODO.md — рабочий чек-лист порта tradingview-mcp на Go

## P0 — Инвентаризация

- [ ] P0.01 Склонировать оригинальный репозиторий.
- [ ] P0.02 Зафиксировать commit hash.
- [ ] P0.03 Запустить `npm install`.
- [ ] P0.04 Запустить `npm test`.
- [ ] P0.05 Снять список MCP tools из Node.js.
- [ ] P0.06 Снять список CLI-команд.
- [ ] P0.07 Составить карту `src/connection.js`.
- [ ] P0.08 Составить карту `src/server.js`.
- [ ] P0.09 Составить карту `src/tools`.
- [ ] P0.10 Составить карту `src/cli`.
- [ ] P0.11 Составить compatibility matrix.

## P1 — Go skeleton

- [ ] P1.01 Создать `go.mod`.
- [ ] P1.02 Создать `cmd/tvmcp/main.go`.
- [ ] P1.03 Создать `cmd/tv/main.go`.
- [ ] P1.04 Создать `internal/mcp`.
- [ ] P1.05 Создать `internal/cdp`.
- [ ] P1.06 Создать `internal/tools`.
- [ ] P1.07 Создать `internal/cli`.
- [ ] P1.08 Добавить базовые тесты.
- [ ] P1.09 Проверить `go test ./...`.

## P2 — MCP server

- [ ] P2.01 Реализовать JSON-RPC request/response structs.
- [ ] P2.02 Реализовать stdio transport.
- [ ] P2.03 Реализовать `initialize`.
- [ ] P2.04 Реализовать `tools/list`.
- [ ] P2.05 Реализовать `tools/call`.
- [ ] P2.06 Реализовать tool registry.
- [ ] P2.07 Реализовать MCP error mapping.
- [ ] P2.08 Добавить unit tests.

## P3 — CDP client

- [ ] P3.01 Реализовать GET `http://localhost:9222/json/list`.
- [ ] P3.02 Реализовать target discovery.
- [ ] P3.03 Реализовать WebSocket CDP transport.
- [ ] P3.04 Реализовать CDP request/response matching.
- [ ] P3.05 Реализовать `Runtime.enable`.
- [ ] P3.06 Реализовать `Page.enable`.
- [ ] P3.07 Реализовать `DOM.enable`.
- [ ] P3.08 Реализовать `Runtime.evaluate`.
- [ ] P3.09 Реализовать liveness check.
- [ ] P3.10 Реализовать retry/backoff.
- [ ] P3.11 Реализовать screenshot primitive.
- [ ] P3.12 Добавить tests.

## P4 — First E2E

- [ ] P4.01 Реализовать `tv_health_check`.
- [ ] P4.02 Реализовать `tv status`.
- [ ] P4.03 Проверить MCP `tools/list`.
- [ ] P4.04 Проверить MCP `tools/call` для `tv_health_check`.
- [ ] P4.05 Проверить CLI JSON output.
- [ ] P4.06 Обновить `CHANGELOG.md`.

## P5 — Read-only tools

- [ ] P5.01 `chart_get_state`.
- [ ] P5.02 `quote_get`.
- [ ] P5.03 `data_get_ohlcv`.
- [ ] P5.04 `data_get_study_values`.
- [ ] P5.05 `data_get_pine_lines`.
- [ ] P5.06 `data_get_pine_labels`.
- [ ] P5.07 `data_get_pine_tables`.
- [ ] P5.08 `data_get_pine_boxes`.
- [ ] P5.09 `capture_screenshot`.
- [ ] P5.10 CLI mappings.
- [ ] P5.11 Golden tests.
- [ ] P5.12 Обновить `CHANGELOG.md`.

## P6 — Chart control

- [ ] P6.01 `chart_set_symbol`.
- [ ] P6.02 `chart_set_timeframe`.
- [ ] P6.03 `chart_set_type`.
- [ ] P6.04 `chart_manage_indicator`.
- [ ] P6.05 `chart_scroll_to_date`.
- [ ] P6.06 `chart_set_visible_range`.
- [ ] P6.07 `symbol_info`.
- [ ] P6.08 `symbol_search`.
- [ ] P6.09 `indicator_set_inputs`.
- [ ] P6.10 `indicator_toggle_visibility`.
- [ ] P6.11 CLI mappings.
- [ ] P6.12 Tests.
- [ ] P6.13 Обновить `CHANGELOG.md`.

## P7 — Pine Script

- [ ] P7.01 `pine_get_source`.
- [ ] P7.02 `pine_set_source`.
- [ ] P7.03 `pine_smart_compile`.
- [ ] P7.04 `pine_get_errors`.
- [ ] P7.05 `pine_get_console`.
- [ ] P7.06 `pine_save`.
- [ ] P7.07 `pine_new`.
- [ ] P7.08 `pine_open`.
- [ ] P7.09 `pine_list_scripts`.
- [ ] P7.10 `pine_analyze`.
- [ ] P7.11 `pine_check`.
- [ ] P7.12 CLI mappings.
- [ ] P7.13 Tests.
- [ ] P7.14 Обновить `CHANGELOG.md`.

## P8 — Drawing

- [ ] P8.01 `draw_shape`.
- [ ] P8.02 draw list.
- [ ] P8.03 draw get.
- [ ] P8.04 draw remove.
- [ ] P8.05 draw clear.
- [ ] P8.06 Tests.
- [ ] P8.07 Обновить `CHANGELOG.md`.

## P9 — Alerts and watchlist

- [ ] P9.01 alert list.
- [ ] P9.02 alert create.
- [ ] P9.03 alert delete.
- [ ] P9.04 watchlist get.
- [ ] P9.05 watchlist add.
- [ ] P9.06 Tests.
- [ ] P9.07 Обновить `CHANGELOG.md`.

## P10 — Layouts, panes, tabs

- [ ] P10.01 layout list.
- [ ] P10.02 layout switch.
- [ ] P10.03 pane list.
- [ ] P10.04 pane layout.
- [ ] P10.05 pane focus.
- [ ] P10.06 pane symbol.
- [ ] P10.07 tab list.
- [ ] P10.08 tab new.
- [ ] P10.09 tab close.
- [ ] P10.10 tab switch.
- [ ] P10.11 Tests.
- [ ] P10.12 Обновить `CHANGELOG.md`.

## P11 — Replay

- [ ] P11.01 replay start.
- [ ] P11.02 replay step.
- [ ] P11.03 replay stop.
- [ ] P11.04 replay status.
- [ ] P11.05 replay autoplay.
- [ ] P11.06 replay trade.
- [ ] P11.07 Tests.
- [ ] P11.08 Обновить `CHANGELOG.md`.

## P12 — UI automation

- [ ] P12.01 ui click.
- [ ] P12.02 ui keyboard.
- [ ] P12.03 ui hover.
- [ ] P12.04 ui scroll.
- [ ] P12.05 ui find.
- [ ] P12.06 ui eval.
- [ ] P12.07 ui type.
- [ ] P12.08 ui panel.
- [ ] P12.09 ui fullscreen.
- [ ] P12.10 ui mouse.
- [ ] P12.11 Tests.
- [ ] P12.12 Обновить `CHANGELOG.md`.

## P13 — Streaming

- [ ] P13.01 stream quote.
- [ ] P13.02 stream bars.
- [ ] P13.03 stream values.
- [ ] P13.04 stream lines.
- [ ] P13.05 stream labels.
- [ ] P13.06 stream tables.
- [ ] P13.07 stream all.
- [ ] P13.08 Ctrl+C handling.
- [ ] P13.09 JSONL tests.
- [ ] P13.10 Обновить `CHANGELOG.md`.

## P14 — Compatibility audit

- [ ] P14.01 Сравнить полный список MCP tools.
- [ ] P14.02 Сравнить CLI help.
- [ ] P14.03 Сравнить JSON output.
- [ ] P14.04 Проверить Claude Code MCP config.
- [ ] P14.05 Проверить Windows launch.
- [ ] P14.06 Проверить Linux launch.
- [ ] P14.07 Проверить macOS launch.
- [ ] P14.08 Обновить README.
- [ ] P14.09 Подготовить release notes.

## Текущий статус

Статус: planning.

Следующий шаг: P0.01.


## P2W — Windows discovery / WindowsApps

- [ ] P2W.01 Добавить `internal/discovery`.
- [ ] P2W.02 Добавить `internal/launcher`.
- [ ] P2W.03 Поддержать `TRADINGVIEW_PATH`.
- [ ] P2W.04 Поддержать CLI flag `--tv-path`.
- [ ] P2W.05 Проверять уже запущенный CDP на `127.0.0.1:9222`.
- [ ] P2W.06 Проверять `%LOCALAPPDATA%\TradingView\TradingView.exe`.
- [ ] P2W.07 Проверять `%PROGRAMFILES%` и `%PROGRAMFILES(X86)%`.
- [ ] P2W.08 Найти Microsoft Store package через `Get-AppxPackage TradingView.Desktop`.
- [ ] P2W.09 Обработать `C:\Program Files\WindowsApps\TradingView.Desktop_*`.
- [ ] P2W.10 Не падать при `Access denied` на `WindowsApps`.
- [ ] P2W.11 Добавить `tv doctor windows`.
- [ ] P2W.12 Покрыть discovery unit-тестами через fake filesystem/fake command runner.

## P0S — skills, scripts, utilities

- [ ] P0S.01 Инвентаризировать `scripts/`.
- [ ] P0S.02 Инвентаризировать skills/workflows, если они есть в репозитории.
- [ ] P0S.03 Инвентаризировать CLI helpers.
- [ ] P0S.04 Добавить каждую утилиту в `COMPATIBILITY_MATRIX.md`.
- [ ] P0S.05 Портировать launch scripts в Go launcher или сохранить как совместимые wrapper scripts.
