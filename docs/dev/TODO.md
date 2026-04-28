# TODO — tradingview-mcp-go stabilization

## Status baseline (2026-04-25)

```text
tradingview-mcp-go: parity baseline passed, stabilization extensions added
historical Node parity: 78/78 tools implemented
current Go registry: 85 MCP tools
CLI implemented
MCP manual implementation accepted
critical blockers: none
```

---

## Phase 1 — Golden tests (MCP protocol layer)

Tests that do NOT require TradingView running. Pure protocol coverage.

- [x] `initialize` → correct protocolVersion, serverInfo, capabilities
- [x] `tools/list` → historical parity was 78 tools; current stabilized Go registry returns 85 tools, each has name/description/inputSchema
- [x] `tools/call` known tool → `{"content":[{"type":"text","text":"..."}],"isError":false}`
- [x] `tools/call` unknown tool → `{"content":[...],"isError":true}`
- [x] `tools/call` with bad JSON args → `{"content":[...],"isError":true}`
- [x] `ping` → `{}`
- [x] unknown method → JSON-RPC error -32601
- [x] parse error (malformed JSON) → JSON-RPC error -32700
- [x] multiline JSON input → each partial line produces -32700 (NDJSON protocol violation)
- [x] large response (>64KB) — regression for old Scanner limit
- [x] notifications/initialized → no response emitted
- [x] sequential requests → 3 in, 3 out, all succeed

---

## Phase 2 — Smoke tests (TradingView live)

Require: TradingView Desktop running with `--remote-debugging-port=9222`.
Run manually: `go test ./tests/smoke/... -v -timeout 60s`

> **Windows MSIX note**: TradingView v3.1.0 (after auto-update from 3.0.0)
> accepts `--remote-debugging-port=9222` when launched directly from its
> installation directory (`cmd.Dir = filepath.Dir(execPath)`). The launcher
> now uses this approach for all installs including MSIX.
> Use `tv launch` (or `tv launch --kill` to restart) to enable CDP.

- [x] Smoke test suite created in `tests/smoke/smoke_test.go` with skip logic
- [x] `TestHealthCheckShape` passes without CDP (shape validation only)
- [x] Launcher: `cmd.Dir` set to TradingView install dir — CDP accepted by MSIX v3.1.0
- [x] Launcher: already-running check before attempting new launch (2s probe on port)
- [x] `launchDirect` used for all installs — clean, minimal implementation
- [x] **MSIX launch limitation documented**: `tv launch` works from an interactive user
  terminal (TTY present). MSIX GUI apps cannot be started from a non-interactive
  subprocess (MCP server context) — Windows/Electron exits cleanly with code 0.
  Workaround: user runs `tv launch` or starts TradingView manually, then uses MCP.
- [x] `TestCDPConnect` — connected=true, targetUrl present
- [x] `TestChartGetState` — symbol=RUS:NG1! timeframe=1D type=1 (fixed: added `timeframe`/`type` aliases, stringify chartType)
- [x] `TestQuoteGet` — symbol/last/close present
- [x] `TestDataGetOHLCV` — 5 bars returned with OHLCV fields
- [x] `TestDataGetStudyValues` — 3 studies returned
- [x] `TestCaptureScreenshot` — saved to screenshots/ (fixed: added `path` alias for `file_path`)
- [x] `TestPineGetSource` — Pine source returned (33 KB)

---

## Phase 3 — Windows doctor

`tv doctor` should provide actionable diagnostics on Windows.

- [x] Check `localhost:9222` reachable — clear message if not
- [x] List running processes, detect TradingView.exe without `--remote-debugging-port`
- [x] Probe WindowsApps/MSIX path (`Get-AppxPackage *TradingView*`)
- [x] Probe `%LOCALAPPDATA%\TradingView\` and `%APPDATA%\TradingView\`
- [x] Output exact command to restart with CDP flag
- [x] Detect if port 9222 is used by a different process (Chrome, etc.)
- [x] Structured JSON output + human-readable hint block

---

## Phase 4 — HTS-ready tools

New tools to support the HTS integration layer.

- [x] `chart_context_for_llm` — combines chart_get_state + quote_get + top-N study values into one call; returns a single structured object ready for LLM prompt injection
- [x] `indicator_state` — current value + signal direction (above/below zero, overbought/oversold) for a named indicator; partial name match; RSI/Stoch/CCI-aware thresholds
- [x] `market_summary` — symbol, timeframe, last bar OHLCV, change%, volume vs 20-bar avg, all active indicators; one call for full context
- [x] `continuous_contract_context` — for futures: detect continuous contract (NG1!, ES1!, CL2!), parse base symbol + roll number, enrich with symbolExt() description/exchange/type

---

## Phase 5 — JSON contracts (HTS integration)

Formal response schemas for tools the HTS layer consumes.
See: [JSON_CONTRACTS.md](JSON_CONTRACTS.md)
Final readiness audit: [FINAL_AUDIT_REPORT.md](FINAL_AUDIT_REPORT.md)
See also: [HTS_MARKET_SUMMARY_CONTRACT.md](HTS_MARKET_SUMMARY_CONTRACT.md)
for the source-aware external HTS MCP -> LLM summary contract.
See also: [LLM_MARKET_CONTEXT_CONTRACT.md](LLM_MARKET_CONTEXT_CONTRACT.md)
for compact LLM payloads without raw candles.
See also: [INSTRUMENT_RESOLVER_CONTRACT.md](INSTRUMENT_RESOLVER_CONTRACT.md)
for TradingView analysis symbol -> Tinkoff execution instrument mapping.

Priority tools:

- [x] `data_get_study_values` — `entity_id`, `plot_count`, `plots` array (numeric values); `studies` always `[]`
- [x] `data_get_indicator` — `inputs` now a key→value map; `plots` array; `name` from metaInfo
- [x] `chart_get_state` — added `exchange`, `ticker`, `pane_count`; `indicators` canonical alias for `studies`
- [x] `quote_get` — `bid`, `ask`, `change`, `change_pct` always present (0 sentinel)
- [x] `symbol_search` / `symbol_info` — `type`, `exchange`, `description` always present (empty string)
- [x] Define retryable errors (`CDP`/`connect`/`timeout`/`websocket`) vs permanent (`unknown tool`/`unmarshal`/`is required`) — `mcp.IsRetryable()` + `mcp.ClassifyError()`

---

## Technical debt

- [x] Replace `bufio.Scanner` with `bufio.Reader.ReadBytes` in mcp/server.go (done 2026-04-25)
- [x] MCP server: add max message size guard (16 MB per-line check, done 2026-04-25)
- [x] `.github/workflows/test.yml` — run `go test ./...` on every push (done 2026-04-25)
- [ ] CDP client: configurable reconnect backoff (currently hardcoded)
- [ ] `go test ./...` in CI release.yml (add before package step)
