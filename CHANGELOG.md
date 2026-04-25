# CHANGELOG.md

Формат основан на Keep a Changelog, но адаптирован под процесс портирования Node.js → Go.

Каждый этап портирования должен добавлять датированную запись. Типы записей: **Added**, **Changed**, **Fixed**, **Compatibility**, **Pending**, **Breaking** (Breaking — только с обоснованием).

---

## 2026-04-24

### Added

- Завершена инвентаризация исходного Node.js репозитория `tradingview-mcp` (P0, P0S).
- Установлено точное число MCP tools: **78**, распределённых по 15 группам.
- Установлено точное число CLI-команд: **83**, распределённых по 16 групп.
- Выявлены 8 CDP методов: `Runtime.evaluate`, `Runtime.enable`, `Page.enable`, `DOM.enable`, `Page.captureScreenshot`, `Input.dispatchKeyEvent`, `Input.insertText`, `Input.dispatchMouseEvent`.
- Задокументированы 6 scripts: `pine_pull.js`, `pine_push.js`, 4 платформенных launch scripts.
- Задокументированы 5 skills/workflows: chart-analysis, multi-symbol-scan, pine-develop, replay-practice, strategy-report.
- Заполнен `COMPATIBILITY_MATRIX.md`: полная таблица tool-by-tool, CLI команды, CDP вызовы, scripts, skills.
- Создан `PORTING_NOTES.md` с архитектурными деталями, паттернами Node.js-кода и примечаниями для Go-порта.
- Обновлён `TODO.md`: исправлены имена tools в P8 (`draw_remove_one`, `draw_get_properties`), добавлены пропущенные tools в P4, P5, P7, P12.
- Добавлено требование поддержки Windows Microsoft Store / WindowsApps установки TradingView Desktop.
- Добавлено требование портировать skills, scripts и utilities, а не только MCP tools.
- Добавлен регламент оркестрации портирования в лимитах Claude Pro.
- Добавлены отдельные документы `WINDOWS_TRADINGVIEW_DISCOVERY.md`, `ORCHESTRATION_AND_LIMITS.md`, `PROMPTS.md`.
- Подготовлен план 1:1 переноса `tradingview-mcp` с Node.js на Go.
- Зафиксирована целевая архитектура Go-проекта.

### Compatibility

- MCP server: `src/server.js` экспортирует 78 tools через `@modelcontextprotocol/sdk`.
- CLI binary: `tv` → `src/cli/index.js`, router → `src/cli/router.js`.
- CDP port default: `localhost:9222`.
- TradingView API root: `window.TradingViewApi` (объект инжектируется TradingView Desktop).
- Exit codes: 0 (success), 1 (error), 2 (connection failure).

### Pending

- P0.02 Зафиксировать commit hash оригинального репозитория.
- P0.03–P0.04 Запустить `npm install` и `npm test`.
- P2 MCP stdio server (JSON-RPC types уже созданы; нужны доп. тесты).
- P3 CDP client (WebSocket + Runtime.evaluate реализованы; нужны тесты с mock-сервером).
- P4 E2E: `tv_health_check`, `tv_discover`, `tv_ui_state`, `tv_launch`, CLI `tv status`.
- P2W.04–P2W.12 CLI flag --tv-path, doctor windows, unit-тесты discovery.

---

## 2026-04-24 (P1 — Go skeleton)

### Added — Go skeleton files

- `go.mod` — модуль `github.com/jhonroun/tradingview-mcp-go`, Go 1.21, зависимость `github.com/gorilla/websocket v1.5.3`.
- `internal/mcp/types.go` — JSON-RPC 2.0 и MCP protocol типы: Request, Response, RPCError, Tool, InputSchema, ListToolsResult, CallToolResult.
- `internal/mcp/registry.go` — ToolDef и Registry: Register, Get, List, Call.
- `internal/mcp/server.go` — MCP stdio server: bufio read loop, handleInitialize, handleListTools, handleCallTool, handlePing.
- `internal/mcp/server_test.go` — 5 unit-тестов: initialize, list tools, call tool, unknown method, registry call unknown.
- `internal/cdp/types.go` — Target, Message, CDPError, EvaluateParams, RemoteObject, EvaluateResult, ExceptionDetails.
- `internal/cdp/discovery.go` — ListTargets (GET /json/list), FindChartTarget (prefer chart URL, fallback to tradingview URL).
- `internal/cdp/discovery_test.go` — 5 unit-тестов: exact match, fallback, none, skip non-page, prefer chart over root.
- `internal/cdp/client.go` — Client: Connect, ConnectWithRetry (exponential backoff), EnableDomains, Evaluate, LivenessCheck, Close; goroutine-safe request/response matching.
- `internal/tools/health/health.go` — HealthCheck, Discover, UIState, Launch; RegisterTools → регистрирует tv_health_check, tv_discover, tv_ui_state, tv_launch.
- `internal/discovery/discovery.go` — Find: TRADINGVIEW_PATH env → LOCALAPPDATA/PROGRAMFILES → Microsoft Store via PowerShell Get-AppxPackage → /Applications (macOS) → PATH (Linux).
- `internal/launcher/launcher.go` — Launch: killRunning, exec TradingView с --remote-debugging-port, ожидание CDP с таймаутом 15 с.
- `internal/cli/router.go` — Register, Dispatch, parseFlags (--key=value, --key value, --bool-flag).
- `cmd/tvmcp/main.go` — MCP stdio сервер: регистрирует health tools, запускает Server.Run().
- `cmd/tv/main.go` — CLI: команды `status` и `launch`.

### Verified

- `go build ./...` — успешно, нет ошибок.
- `go test ./...` — 10/10 тестов PASS (5 cdp/discovery, 5 mcp/server).

---

## 2026-04-24 (P2/P3/P4 — MCP server, CDP client, first E2E tools)

### Added — P3 CDP enhancements

- `internal/cdp/client.go` — добавлен `CaptureScreenshot` (Page.captureScreenshot → base64 PNG).
- `internal/cdp/client_test.go` — 5 тестов через mock WebSocket CDP сервер: Evaluate (число), EnableDomains, LivenessCheck, JS-ошибка, Screenshot.

### Added — CLI commands

- `cmd/tv/main.go` — команды `discover` (tv_discover), `ui-state` (tv_ui_state), `doctor` (диагностика CDP + installation).
- `tv doctor` выводит JSON с ключами `cdp` (ok, targets, chart, targetId, url) и `install` (ok, path, source, platform).

### Added — unit tests

- `internal/cli/router_test.go` — 5 тестов parseFlags: equals-form, space-form, bool-flag, empty, mixed.
- `internal/discovery/discovery_test.go` — 3 теста: TRADINGVIEW_PATH env, missing env path, fileExists helper.

### Verified — P2/P3/P4

---

## 2026-04-24 (P5 — Read-only chart tools)

### Added — P5 packages

- `internal/cdp/session.go` — `WithSession` helper: connect → enable domains → run fn → close. Eliminates boilerplate in every tool.
- `internal/cdp/client.go` — `ScreenshotClip` type, `CaptureScreenshotClip(clip)`, refactored `screenshot()` private method.
- `internal/tradingview/js.go` — `ChartAPI`, `BarsPath`, `ChartWidget` constants; `SafeString()` (mirrors connection.js safeString).
- `internal/tools/chart/chart.go` — `chart_get_state` (symbol, resolution, chartType, studies list), `chart_get_visible_range`.
- `internal/tools/data/data.go` — 12 tools: `data_get_ohlcv` (with Go-side summary computation), `quote_get`, `data_get_study_values`, `data_get_pine_lines` (horizontal-level dedup + verbose), `data_get_pine_labels` (max_labels + verbose), `data_get_pine_tables` (row formatting), `data_get_pine_boxes` (zone dedup + verbose), `data_get_indicator`, `data_get_strategy_results`, `data_get_trades`, `data_get_equity`, `depth_get`.
- `internal/tools/capture/capture.go` — `capture_screenshot` (regions: full / chart / strategy_tester; JS clip detection; saves to screenshots/).
- `cmd/tvmcp/main.go` — registers chart, data, capture tool groups (total ~18 new MCP tools).
- `cmd/tv/main.go` — CLI commands: `tv quote [SYMBOL]`, `tv ohlcv [--count N] [--summary]`, `tv screenshot [--region X] [--filename F]`, `tv chart-state`.

### Added — P5 tests

- `internal/tools/data/data_test.go` — TestRegisteredToolNames (12 tool names vs compatibility matrix), TestRound2, TestBuildGraphicsJSContainsFilter, TestJoinStr, TestCoalesce.
- `internal/tradingview/js_test.go` — TestSafeString (5 cases), TestSafeStringNoInjection (3 dangerous inputs).

### Pending — P5

- P5.16 `batch_run` — complex multi-symbol iteration, deferred to later session.

### Verified — P5 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **30/30 тестов PASS** (5 cdp/client, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 5 tools/data, 2 tradingview/js).

- `go build ./...` — успешно.
- `go test ./...` — **23/23 тестов PASS** (5 cdp/client, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server).

---

## 2026-04-24 (P6 — Chart control + indicator tools)

### Added — P6 packages

- `internal/tools/chart/wait.go` — `waitForChartReady`: polls every 200 ms; checks loading spinner, symbol match (case-insensitive), bar count stable ×2; timeout 10 s.
- `internal/tools/chart/control.go` — `SetSymbol` (Promise + waitForChartReady), `SetTimeframe` (setResolution), `SetType` (chartTypeMap 0–9), `ManageIndicator` (add: createStudy + diff, remove: removeEntity), `SetVisibleRange` (bar-index scan + zoomToBarsRange), `ScrollToDate` (±25 bar window); `registerControlTools` adds 6 MCP tools.
- `internal/tools/chart/symbol.go` — `SymbolInfo` (chart.symbolExt()), `SymbolSearch` (REST GET symbol-search.tradingview.com, strips `<em>`, max 15 results); `registerSymbolTools` adds 2 MCP tools.
- `internal/tools/indicators/indicators.go` — `SetInputs` (getInputValues → override by id → setInputValues), `ToggleVisibility` (setVisible + isVisible); `RegisterTools` adds 2 MCP tools.
- `internal/tools/chart/chart.go` — `RegisterTools` now calls `registerControlTools` + `registerSymbolTools` (total 10 chart tools).
- `cmd/tvmcp/main.go` — registers `indicators.RegisterTools` (12 total indicator + 10 chart tools).
- `cmd/tv/main.go` — CLI commands: `tv set-symbol`, `tv set-timeframe`, `tv set-type`, `tv symbol-info`, `tv symbol-search [--type] [--exchange]`, `tv indicator-toggle ENTITY_ID [--visible]`.

### Added — P6 tests

- `internal/tools/chart/chart_p6_test.go` — TestChartTypeMap (9 entries), TestSetTypeUnknown, TestSetTypeNormalization, TestRegisterToolsP6Names (10 tool names).
- `internal/tools/indicators/indicators_test.go` — TestRegisterIndicatorTools (2 tool names).

### Verified — P6 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **38/38 тестов PASS** (5 cdp/client, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 5 tools/data, 4 tools/chart, 1 tools/indicators, 2 tradingview/js).

---

## 2026-04-24 (P7 — Pine Script tools)

### Added — P7 packages

- `internal/cdp/client.go` — `KeyEventParams`, `DispatchKeyEvent(ctx, params)` for Ctrl+Enter / Ctrl+S dispatch.
- `internal/tools/pine/pine.go` — all CDP-dependent Pine tools:
  - `findMonaco` constant — React Fiber traversal to locate Monaco editor instance.
  - `ensurePineEditorOpen(ctx, client)` — opens Pine Editor panel, polls until Monaco ready (10 s).
  - `GetSource()` — `editor.getValue()`.
  - `SetSource(source)` — `editor.setValue(escaped)`.
  - `Compile()` — clicks "Save and add to chart" / "Add to chart" / "Update on chart" buttons; fallback Ctrl+Enter; waits 2 s.
  - `SmartCompile()` — counts studies before/after, clicks compile, reads Monaco markers, reports study_added.
  - `GetErrors()` — `getModelMarkers({resource: model.uri})`.
  - `GetConsole()` — DOM scraping for console rows with timestamp/type classification.
  - `Save()` — Ctrl+S + handles "Save Script" name dialog.
  - `NewScript(type)` — injects indicator/strategy/library template via `editor.setValue`.
  - `OpenScript(name)` — fetch pine-facade list (credentials:include) + fuzzy name match + get source + Monaco setValue.
  - `ListScripts()` — fetch pine-facade list, returns id/name/title/version/modified.
  - `Check(source)` — HTTP POST to `pine-facade.tradingview.com/pine-facade/translate_light` (Guest, public endpoint); returns errors/warnings.
- `internal/tools/pine/analyze.go` — offline static analyzer (no CDP):
  - Detects Pine version from `//@version=N`.
  - Tracks `array.from()` and `array.new*()` declarations with sizes.
  - Flags `array.get/set` calls with literal out-of-bounds indices.
  - Flags `array.first/last()` on zero-size arrays.
  - Flags `strategy.entry/close` without `strategy()` declaration.
  - Warns about Pine versions < 5.
- `cmd/tvmcp/main.go` — registers `pine.RegisterTools(reg)` (12 new MCP tools; total ~42).
- `cmd/tv/main.go` — `tv pine <get|set|compile|smart-compile|raw-compile|errors|console|save|new|open|list|analyze|check>` dispatcher.

### Added — P7 tests

- `internal/tools/pine/pine_test.go` — TestRegisterPineToolNames (12 names), TestAnalyzeCleanScript, TestAnalyzeArrayOutOfBounds, TestAnalyzeStrategyWithoutDecl, TestAnalyzeOldVersion, TestAnalyzeStrategyWithDecl.

### Verified — P7 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **43/43 тестов PASS** (5 cdp/client, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 5 tools/data, 4 tools/chart, 1 tools/indicators, 5 tools/pine, 2 tradingview/js).

---

## 2026-04-24 (P8 — Drawing tools)

### Added — P8 packages

- `internal/tools/drawing/drawing.go` — 5 MCP tools:
  - `DrawShape` — `createShape` (1-point) or `createMultipointShape` (2-point); waits 200 ms; diffs `getAllShapes()` to extract new entity ID. `fmtNum` uses `strconv.FormatFloat('f')` to avoid scientific notation in JS. `requireFinite` validates point coordinates.
  - `ListDrawings` — `getAllShapes()` → `[{id, name}]`.
  - `GetProperties` — `getShapeById(eid)` → points, properties (with `.properties()` fallback), visibility, lock, selection state, available methods.
  - `RemoveOne` — pre-checks shape exists, calls `removeEntity(eid)`, post-verifies removal; returns `remaining_shapes`.
  - `ClearAll` — `removeAllShapes()`.
- `cmd/tvmcp/main.go` — registers `drawing.RegisterTools(reg)` (5 new MCP tools; total ~47).
- `cmd/tv/main.go` — `tv draw <shape|list|get|remove|clear>` dispatcher; `tv draw shape` accepts `--time/--price/--time2/--price2/--text` flags.

### Added — P8 tests

- `internal/tools/drawing/drawing_test.go` — TestRegisterDrawingToolNames (5 names), TestRequireFinite (4 cases), TestFmtNum (4 cases), TestDrawShapeValidation (NaN guard).

### Verified — P8 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **47/47 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pine, 2 tradingview/js).

---

## 2026-04-24 (P9 — Alerts + Watchlist)

### Added — P9 packages

- `internal/cdp/client.go` — `InsertText(ctx, text)` via `Input.insertText`.
- `internal/tools/alerts/alerts.go` — 5 MCP tools (alert + watchlist group):
  - `CreateAlert` — clicks "Create Alert" button or Shift+A fallback; sets price via React synthetic event override (`Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set`); sets message via textarea; clicks "Create" button.
  - `ListAlerts` — fetch `pricealerts.tradingview.com/list_alerts` (credentials:include) via CDP; returns alert_id, symbol, type, condition, active, timestamps.
  - `DeleteAlerts` — `delete_all:true`: opens context menu (manual confirmation required); `delete_all:false`: returns "not yet supported" error (matches Node.js behavior).
  - `GetWatchlist` — three-tier DOM scraping: `data-symbol-full` attributes → `symbolName/tickerName` text scan; returns symbol + last/change/change_percent.
  - `AddToWatchlist` — opens watchlist panel, clicks add-symbol button, `InsertText(symbol)`, Enter to confirm, Escape to close; Escape cleanup on error.
- `cmd/tvmcp/main.go` — registers `alerts.RegisterTools(reg)` (5 new tools; total ~52).
- `cmd/tv/main.go` — `tv alert <list|create|delete>`, `tv watchlist <get|add SYMBOL>`.

### Added — P9 tests

- `internal/tools/alerts/alerts_test.go` — TestRegisterAlertToolNames (5 names), TestDeleteAlertsIndividualError.

### Verified — P9 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **49/49 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pine, 2 tradingview/js).

---

## 2026-04-24 (P10 — Panes + Tabs)

### Added — P10 packages

- `internal/tools/pane/pane.go` — 4 MCP tools:
  - `ListPanes` — evaluates `_chartWidgetCollection.getAll()`, returns layout code/name, chart_count, active_index, per-pane symbol/resolution.
  - `SetLayout` — `resolveLayout` normalises aliases (single→s, 2x2→4, quad→4, grid→4, 2x1→2h, 1x2→2v) and friendly names; calls `_chartWidgetCollection.setLayout(code)`; waits 500 ms; returns updated pane list.
  - `FocusPane` — finds pane by 0-based index in `getAll()`, calls `_mainDiv.click()`; returns focused_index and total_panes.
  - `SetPaneSymbol` — focuses pane first (FocusPane + 300 ms), then calls `chart.setSymbol(symbol, {})` via Promise (500 ms internal delay).
- `internal/tools/tab/tab.go` — 4 MCP tools:
  - `ListTabs` — GET `/json/list`; filters TradingView pages; returns id, title, url per tab.
  - `NewTab` — sends Ctrl+T (modifiers=2, keyCode=84) to active window; waits 1 s; returns updated tab list.
  - `CloseTab` — guards against closing the last tab (≥2 required); sends Ctrl+W; waits 500 ms; returns updated tab list.
  - `SwitchTab(tabID)` — GET `http://localhost:9222/json/activate/{id}`; returns activated_tab_id.
- `cmd/tvmcp/main.go` — registers `pane.RegisterTools` + `tab.RegisterTools` (8 new tools; total ~60).
- `cmd/tv/main.go` — `tv pane <list|set-layout|focus|set-symbol>` and `tv tab <list|new|close|switch ID>` dispatchers.

### Fixed — P10

- `internal/tools/pane/pane.go` — typo `mpc.PropertySchema` → `mcp.PropertySchema` in `pane_focus` registration.

### Added — P10 tests

- `internal/tools/pane/pane_test.go` — TestResolveLayoutKnownCodes (4 codes), TestResolveLayoutAliases (8 aliases), TestResolveLayoutCaseInsensitive, TestResolveLayoutUnknown, TestRegisterPaneToolNames (4 names).
- `internal/tools/tab/tab_test.go` — TestRegisterTabToolNames (4 names), TestSwitchTabEmptyID.

### Verified — P10 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **56/56 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pane, 5 tools/pine, 2 tools/tab, 2 tradingview/js).

---

## 2026-04-24 (P11 — Replay)

### Added — P11 packages

- `internal/tools/replay/replay.go` — 6 MCP tools (1:1 port of `src/core/replay.js`):
  - `Start(date)` — checks `isReplayAvailable()`, shows toolbar, calls `selectDate(tsMs)` (awaited) or `selectFirstAvailableDate()`, polls 30×250 ms until `isReplayStarted && currentDate != null`; on failure calls `stopReplay()` and returns descriptive error.
  - `Step()` — guards `isReplayStarted`, reads `currentDate` before, calls `doStep()`, polls 12×250 ms until date changes.
  - `Stop()` — idempotent: returns `already_stopped` if not running, else calls `stopReplay()`.
  - `Status()` — single JS block reads `isReplayAvailable/Started/AutoplayStarted, replayMode, currentDate, autoplayDelay`; appends `position` and `realizedPL`.
  - `Autoplay(speedMs)` — validates speed against `VALID_AUTOPLAY_DELAYS` (100,143,200,300,1000,2000,3000,5000,10000) **before** CDP calls; calls `changeAutoplayDelay` if non-zero, then `toggleAutoplay`; returns `autoplay_active` + `delay_ms`.
  - `Trade(action)` — dispatches `buy()`, `sell()`, `closePosition()`; returns `position` + `realized_pnl`.
  - `wv(path)` helper — unwraps TradingView observable (mirrors Node.js `wv()` in core/replay.js).
- `cmd/tvmcp/main.go` — registers `replay.RegisterTools(reg)` (6 new tools; total ~66).
- `cmd/tv/main.go` — `tv replay <start [--date YYYY-MM-DD]|step|stop|status|autoplay [--speed MS]|trade buy|sell|close>` dispatcher.

### Added — P11 tests

- `internal/tools/replay/replay_test.go` — TestRegisterReplayToolNames (6 names + count), TestAutoplayInvalidSpeed, TestAutoplayValidSpeeds (9 valid delays), TestTradeInvalidAction, TestWvHelper.

### Verified — P11 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **60/60 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pane, 5 tools/pine, 4 tools/replay, 2 tools/tab, 2 tradingview/js).

---

## 2026-04-25 (P12 — UI automation + Layouts)

### Added — P12 packages

- `internal/cdp/client.go` — `MouseEventParams` struct and `DispatchMouseEvent(ctx, p)` via `Input.dispatchMouseEvent`.
- `internal/tools/ui/ui.go` — 12 MCP tools (1:1 port of `src/core/ui.js`):
  - `Click(by, value)` — finds element by aria-label / data-name / text / class-contains; calls `.click()`; returns tag, text, aria_label, data_name.
  - `OpenPanel(panel, action)` — bottom panels (pine-editor, strategy-tester) use `bottomWidgetBar.activateScriptEditorTab/showWidget/hideWidget`; side panels (watchlist, alerts, trading) use data-name button with aria-pressed state detection; action: open/close/toggle.
  - `Fullscreen()` — clicks `[data-name="header-toolbar-fullscreen"]`.
  - `LayoutList()` — `getSavedCharts` Promise (awaitPromise=true, 5 s timeout); returns id/name/symbol/resolution/modified per layout.
  - `LayoutSwitch(name)` — numeric ID → `loadChartFromServer(id)` directly; name → `getSavedCharts` exact match then substring match → `loadChartFromServer`; waits 500 ms; dismisses "unsaved changes" dialog (open anyway / don't save / discard) if present; waits 1 s after dismiss.
  - `Keyboard(key, modifiers)` — keyMap for 17 named keys; fallback to `Key<UPPER>` + charCodeAt(0); modifiers: alt=1, ctrl=2, meta=4, shift=8; dispatches keyDown + keyUp.
  - `TypeText(text)` — `Input.insertText`; returns typed (capped 100 chars) + length.
  - `Hover(by, value)` — finds element coords via `getBoundingClientRect()` centre; dispatches `mouseMoved`.
  - `Scroll(direction, amount)` — finds chart canvas centre; dispatches `mouseWheel` with deltaX/deltaY; default 300 px.
  - `MouseClick(x, y, button, doubleClick)` — `mouseMoved` → `mousePressed` → `mouseReleased`; optional 50 ms + second press/release for double click.
  - `FindElement(query, strategy)` — text: textContent scan on interactive elements (max 20, visible only); aria-label: `[aria-label*=query]`; css: `querySelectorAll(query)`; returns tag/text/aria_label/data_name/x/y/width/height/visible per element.
  - `Evaluate(expression)` — raw `Runtime.evaluate` passthrough; returns `{ success, result }`.
- `cmd/tvmcp/main.go` — registers `ui.RegisterTools(reg)` (12 new tools; total ~78).
- `cmd/tv/main.go` — `tv ui <click|open-panel|fullscreen|keyboard|type|hover|scroll|mouse|find|eval>` and `tv layout <list|switch NAME>` dispatchers.

### Added — P12 tests

- `internal/tools/ui/ui_test.go` — TestRegisterUIToolNames (12 names + count), TestKeyMapEntries (17 keys), TestKeyboardModifierBitfield (7 cases), TestScrollDefaultAmount, TestMouseClickButtonNormalise (5 cases), TestFindElementDefaultStrategy.

### Verified — P12 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **66/66 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pane, 5 tools/pine, 4 tools/replay, 2 tools/tab, 6 tools/ui, 2 tradingview/js).

---

## 2026-04-25 (P13 — Streaming)

### Added — P13 packages

- `internal/stream/stream.go` — JSONL streaming engine (1:1 port of `src/core/stream.js`):
  - `pollLoop(ctx, w, errW, label, intervalMs, dedupe, fetcher)` — connects once; reuses single CDP client; on CDP/WebSocket error reconnects after 2 s; dedup via `JSON.stringify` comparison; appends `_ts` (UnixMilli) and `_stream` fields to every emitted line; exits cleanly on `ctx.Done()`.
  - `isCDPError(err)` — detects CDP/websocket/connection errors for silent reconnect.
  - `StreamQuote(ctx, w, errW, intervalMs)` — last bar OHLCV; default 300 ms.
  - `StreamBars(ctx, w, errW, intervalMs)` — last bar with symbol/resolution/bar_index; default 500 ms.
  - `StreamValues(ctx, w, errW, intervalMs)` — all visible indicator `_lastBarValues`; default 500 ms.
  - `StreamLines(ctx, w, errW, intervalMs, filter)` — Pine `line.new()` price levels, deduped + sorted desc; default 1000 ms.
  - `StreamLabels(ctx, w, errW, intervalMs, filter)` — Pine `label.new()` text+price, max 50; default 1000 ms.
  - `StreamTables(ctx, w, errW, intervalMs, filter)` — Pine `table.new()` row data; default 2000 ms.
  - `StreamAllPanes(ctx, w, errW, intervalMs)` — all chart panes OHLCV in one tick; default 500 ms.
- `cmd/tv/main.go` — `tv stream` handled before `cli.Dispatch` (streams never return):
  - `signal.NotifyContext` for SIGINT/SIGTERM graceful shutdown.
  - Subcommands: `quote bars values lines labels tables all`.
  - `--interval MS` and `--filter NAME` flags parsed inline.
  - Compliance notice printed to stderr before any stream starts.

### Added — P13 tests

- `internal/stream/stream_test.go` — TestDefaultIntervals (7 streams × default + explicit), TestBuildLinesExprContainsFilter, TestBuildLabelsExprContainsFilter, TestBuildTablesExprContainsFilter, TestIsCDPError (7 cases), TestPollLoopCancelledContext (exit on pre-cancelled ctx + "stopped" in stderr), TestJSONLTimestampFields (`_ts` + `_stream` present), TestConstantsContainTradingViewAPI.

### Verified — P13 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **73/73 тестов PASS** (5 cdp, 5 cdp/discovery, 5 cli/router, 3 discovery, 5 mcp/server, 2 tools/alerts, 5 tools/data, 4 tools/chart, 4 tools/drawing, 1 tools/indicators, 5 tools/pane, 5 tools/pine, 4 tools/replay, 7 stream, 2 tools/tab, 6 tools/ui, 2 tradingview/js).

---

## 2026-04-25 (P14 — Compatibility audit + batch_run + README)

### Added — batch_run (P5.16 backlog)

- `internal/tools/batch/batch.go` — `batch_run` MCP tool (1:1 port of `src/core/batch.js`):
  - Iterates every `symbol × timeframe` combination.
  - Per iteration: `chart.SetSymbol` (includes waitForChartReady), optional `chart.SetTimeframe` + 500 ms settle, user `delay_ms` (default 2000 ms).
  - Actions: `screenshot` (capture.CaptureScreenshot "chart" region, safe filename), `get_ohlcv` (exportData Promise, cap 500 bars), `get_strategy_results` (DOM scrape backtesting panel + 1 s settle).
  - Returns `{ success, total_iterations, successful, failed, results[] }`.
- `cmd/tvmcp/main.go` — registers `batch.RegisterTools(reg)` (total **78 MCP tools**).
- `cmd/tv/main.go` — `tv batch --symbols SYM1,SYM2 --action ACTION [--timeframes TF1,TF2] [--delay MS] [--count N]`.

### Compatibility audit results (P14.01–P14.03)

- **MCP tools**: Go 78 / Node.js 78 — exact match, zero missing, zero extra.
- **CLI groups**: all 15 Node.js groups covered (health, chart, data, capture, indicators, pine, drawing, pane, replay, tab, alerts, ui, layout, watchlist, stream); `batch` added as bonus CLI command.
- **Tool names**: verified via live `tools/list` RPC against sorted Node.js grep — 100% match.
- **JSON output structure**: `{ success: bool, ...fields }` pattern preserved across all tools.

### Added — README.md (P14.08)

- `README.md` — user-facing documentation: requirements, build, launch, Claude Code MCP config, full tool table (78 tools grouped), CLI reference, architecture diagram, compatibility notes, disclaimer.

### Added — P14 tests

- `internal/tools/batch/batch_test.go` — TestRegisterBatchToolName, TestBatchRunEmptySymbols, TestBatchRunDefaultDelayAndCount, TestBatchRunOhlcvCountCap, TestBatchRunTimeframeDefault.

### Verified — P14 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **78/78 тестов PASS** (5 tools/batch + all prior).

---

## 2026-04-25 (P15 — Build system, Skills, Documentation)

### Added — P2W.04: --tv-path CLI flag

- `internal/launcher/launcher.go` — `Launch(port, killExisting, tvPath string)`: when `tvPath` is non-empty, skips auto-discovery and uses the provided path directly; validates file existence; reports `source: "cli-flag"` in response.
- `internal/tools/health/health.go` — `LaunchArgs.TvPath *string` field; MCP `tv_launch` schema updated with `tv_path` property; `Launch()` threads `tvPath` to launcher.
- `cmd/tv/main.go` — `tv launch --tv-path=PATH` parses `opts["tv-path"]` and populates `LaunchArgs.TvPath`.

### Added — Build system (P15.01–P15.04)

- `Makefile` — targets: `build` (current platform → `bin/`), `build-all` (6 platforms: windows/linux/darwin × amd64/arm64), `install` (`go install`), `test`, `test-verbose`, `clean`, `release` (build-all + ZIP/tar.gz archives in `bin/releases/`).
- `scripts/build.sh` — shell build script; respects `GOOS`/`GOARCH` env vars.
- `scripts/build.bat` — Windows batch build script.
- `scripts/install.sh` — installs to `/usr/local/bin` (or `PREFIX`).
- `scripts/install.bat` — installs to `%SystemRoot%\System32` (or first argument).

### Added — Scripts (P0S.05–P0S.07)

- `scripts/pine_pull.sh` + `scripts/pine_pull.bat` — wraps `tv pine get > scripts/current.pine`; drop-in replacement for original `pine_pull.js`.
- `scripts/pine_push.sh` + `scripts/pine_push.bat` — wraps `tv pine set` + `tv pine smart-compile`; drop-in replacement for original `pine_push.js`.
- `scripts/launch_tv_debug.bat`, `scripts/launch_tv_debug.vbs`, `scripts/launch_tv_debug_linux.sh`, `scripts/launch_tv_debug_mac.sh` — copied from original Node.js project; still functional as standalone launchers alongside `tv launch`.

### Added — Skills (P0S.08)

- `skills/chart-analysis/SKILL.md` — technical analysis workflow (unchanged from original; no Node.js script references).
- `skills/multi-symbol-scan/SKILL.md` — multi-symbol scan using `batch_run`; JSON input examples instead of code blocks.
- `skills/pine-develop/SKILL.md` — Pine development loop updated to reference `tv pine` CLI and `scripts/pine_pull.sh` / `pine_push.sh` instead of `node scripts/pine_pull.js`.
- `skills/replay-practice/SKILL.md` — replay practice workflow (unchanged from original).
- `skills/strategy-report/SKILL.md` — strategy report workflow (unchanged from original).

### Added — Documentation (P15.05–P15.08)

- `docs/ru/README.md` — primary full Russian documentation: история портирования (Node.js → Go с Claude Code), требования, сборка, запуск, MCP-конфиг, таблица 78 инструментов, CLI-справка, скрипты, навыки, архитектура, совместимость, дисклеймер.
- `docs/en/README.md` — full English documentation: same structure including Origin Story section (porting process, Claude Code role).
- `README.md` (корень) — обновлён: навигационная таблица → docs/ru + docs/en, раздел поддерживаемых MCP-клиентов (не привязан к Claude Code), история портирования, layout структуры проекта.

### Changed — README.md root

- Rewritten as a concise navigation hub pointing to `docs/ru/README.md` and `docs/en/README.md` for full content.
- Added "Supported MCP Clients" section: Claude Code, Cursor, Cline, Continue, Windsurf, any stdio MCP client.
- Added project history note: Go port performed with Claude Code assistance, April 2026.

### Verified — P15 build and tests

- `go build ./...` — успешно.
- `go test ./...` — **78/78 тестов PASS** (все предыдущие; новые файлы не имеют тестируемого Go-кода).
- `bash scripts/build.sh` — `bin/tvmcp.exe` + `bin/tv.exe` — успешно.
